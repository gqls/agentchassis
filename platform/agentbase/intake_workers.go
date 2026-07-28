// FILE: platform/agentbase/intake_workers.go
//
// chassis_replica_scaling CS-2 (P1a): the claim-worker pool. Workers scan for
// serialisation keys with pending intake events, acquire a key via the
// claims-table CAS, drain that key's events IN ORDER through the existing
// processMessage, and release. Per-orchestration ordering is therefore
// structural (one holder per key), and cross-orchestration parallelism is the
// worker count — adjustable per pod via CHASSIS_INTAKE_WORKERS, and later per
// fleet via replicas (P2).
//
// What a wedged event costs HERE, honestly: one worker slot. processMessage
// takes no context, so this pool cannot impose a hard deadline on an event
// without re-plumbing the whole reviewed call chain — that is deliberately NOT
// part of CS-2. The heartbeat keeps a live-but-stuck holder's claim alive
// (letting it lapse would hand the key to another worker and wedge that one
// too, serially, until no workers remain), and INTAKE_EVENT_SLOW lines every
// minute make the wedge visible to the idle-orchestration watchdog the
// dispatch_queue_serialisation workstream owns. Today the same wedge freezes
// an entire lane until a pod roll; one slot of N, loudly logged, is the
// strictly smaller blast radius.

package agentbase

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"go.uber.org/zap"
)

const (
	intakeWorkersEnv       = "CHASSIS_INTAKE_WORKERS"
	intakeWorkersDefault   = 4
	intakeRetentionEnv     = "CHASSIS_INTAKE_RETENTION_DAYS"
	intakeRetentionDefault = 7
	intakeLeaseEnv         = "CHASSIS_CLAIM_LEASE_SECONDS"
	intakeLeaseDefault     = 180 * time.Second

	// One event may be re-popped this many times (each pop is one holder
	// death mid-event) before it is marked failed rather than run again.
	intakeMaxAttempts = 3

	intakeCandidateBatch = 5
	intakePollInterval   = 750 * time.Millisecond
)

type intakeWorkerPool struct {
	agent  *Agent
	repo   *orchestration.IntakeRepository
	logger *zap.Logger

	workers   int
	lease     time.Duration
	retention time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// startIntakeWorkers builds and starts the pool. Called from Run, only when
// setupIntake enabled the path.
func (a *Agent) startIntakeWorkers() {
	workers := intakeWorkersDefault
	if v, err := strconv.Atoi(os.Getenv(intakeWorkersEnv)); err == nil && v > 0 {
		workers = v
	}
	lease := intakeLeaseDefault
	if v, err := strconv.Atoi(os.Getenv(intakeLeaseEnv)); err == nil && v >= 30 {
		lease = time.Duration(v) * time.Second
	}
	retentionDays := intakeRetentionDefault
	if v, err := strconv.Atoi(os.Getenv(intakeRetentionEnv)); err == nil && v > 0 {
		retentionDays = v
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool := &intakeWorkerPool{
		agent:     a,
		repo:      a.intakeRepo,
		logger:    a.logger,
		workers:   workers,
		lease:     lease,
		retention: time.Duration(retentionDays) * 24 * time.Hour,
		ctx:       ctx,
		cancel:    cancel,
	}
	a.intakePool = pool

	for i := 0; i < workers; i++ {
		workerID := fmt.Sprintf("%s/worker-%d", a.PodName, i)
		pool.wg.Add(1)
		go pool.runWorker(workerID)
	}
	pool.wg.Add(1)
	go pool.runPurger()

	a.logger.Info("INTAKE_POOL_STARTED",
		zap.Int("workers", workers),
		zap.Duration("lease", lease),
		zap.Duration("retention", pool.retention))
}

// stopIntakeWorkers stops scanning and waits for in-flight events. An event
// that outlives the pod's grace is recovered by lease expiry + takeover reset,
// with processed_messages suppressing double side effects — the same recovery
// a redelivered Kafka message gets today.
func (a *Agent) stopIntakeWorkers(wait time.Duration) {
	if a.intakePool == nil {
		return
	}
	a.intakePool.cancel()
	done := make(chan struct{})
	go func() {
		a.intakePool.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		a.logger.Info("INTAKE_POOL_STOPPED: all workers drained")
	case <-time.After(wait):
		a.logger.Warn("INTAKE_POOL_STOP_TIMEOUT: an event is still running; lease expiry will recover its key")
	}
}

func (p *intakeWorkerPool) runWorker(workerID string) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		keys, err := p.repo.CandidateKeys(p.ctx, intakeCandidateBatch)
		if err != nil {
			if p.ctx.Err() == nil {
				p.logger.Error("INTAKE_SCAN_FAILED", zap.String("worker", workerID), zap.Error(err))
			}
			p.sleep(intakePollInterval)
			continue
		}

		claimedAny := false
		for _, key := range keys {
			won, err := p.repo.ClaimSerialisationKey(p.ctx, key, workerID, p.lease)
			if err != nil {
				if p.ctx.Err() == nil {
					p.logger.Error("INTAKE_CLAIM_FAILED", zap.String("worker", workerID), zap.Error(err))
				}
				break
			}
			if !won {
				continue // another worker got there; try the next candidate
			}
			claimedAny = true
			p.drainKey(key, workerID)
			break
		}

		if !claimedAny {
			p.sleep(intakePollInterval)
		}
	}
}

// drainKey runs the key's events oldest-first until none remain, the claim is
// lost, or the pool stops. Exactly one goroutine per key can be here — that is
// the ordering guarantee.
func (p *intakeWorkerPool) drainKey(key, workerID string) {
	// Any 'running' event under a claim we have just WON belonged to a dead
	// holder — return it to pending so it re-runs (dedupe suppresses repeated
	// side effects one level down).
	if reset, err := p.repo.ResetRunningEvents(p.ctx, key); err == nil && reset > 0 {
		p.logger.Warn("INTAKE_TAKEOVER_RESET: recovered events from an expired holder",
			zap.String("serialisation_key", key),
			zap.Int64("events", reset),
			zap.String("worker", workerID))
	}

	// Heartbeat until released. If the heartbeat ever reports the claim gone,
	// the drain loop must stop before popping another event.
	var claimLost atomic.Bool
	hbCtx, hbCancel := context.WithCancel(context.Background())
	defer hbCancel()
	go func() {
		ticker := time.NewTicker(p.lease / 3)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return
			case <-ticker.C:
				held, err := p.repo.HeartbeatClaim(hbCtx, key, workerID, p.lease)
				if err == nil && !held {
					claimLost.Store(true)
					p.logger.Error("INTAKE_CLAIM_LOST: lease lapsed and was taken — abandoning this key",
						zap.String("serialisation_key", key),
						zap.String("worker", workerID))
					return
				}
			}
		}
	}()

	for {
		// Pool shutdown: stop between events, release, let another pod finish.
		if p.ctx.Err() != nil || claimLost.Load() {
			break
		}

		ev, err := p.repo.NextPendingEvent(p.ctx, key)
		if err != nil {
			if p.ctx.Err() == nil {
				p.logger.Error("INTAKE_POP_FAILED", zap.String("serialisation_key", key), zap.Error(err))
			}
			break
		}
		if ev == nil {
			break // drained
		}

		if ev.Attempts > intakeMaxAttempts {
			reason := fmt.Sprintf("abandoned after %d attempts (holders died mid-event)", ev.Attempts-1)
			if err := p.repo.MarkEventFailed(p.ctx, ev.ID, reason); err == nil {
				p.logger.Error("INTAKE_EVENT_FAILED: attempts exhausted",
					zap.Int64("event_id", ev.ID),
					zap.String("serialisation_key", key),
					zap.String("reason", reason))
			}
			continue
		}

		p.runEvent(ev, workerID)
	}

	hbCancel()
	if !claimLost.Load() {
		if err := p.repo.ReleaseClaim(context.Background(), key, workerID); err != nil {
			p.logger.Warn("INTAKE_RELEASE_FAILED: lease expiry will clear it",
				zap.String("serialisation_key", key), zap.Error(err))
		}
	}
}

// runEvent reconstructs the original Kafka message and hands it to the SAME
// processMessage the inline path uses — validation, bug-034 rejection
// recording, dedupe and error routing all unchanged. Done is unconditional on
// the in-process outcome, mirroring commitConsumed (bugs_open/003 F3).
func (p *intakeWorkerPool) runEvent(ev *orchestration.IntakeEvent, workerID string) {
	msg := kafka.Message{
		Topic:     ev.Topic,
		Partition: ev.Partition,
		Offset:    ev.Offset,
		Headers:   kafka.MapToHeaders(ev.Headers),
		Value:     ev.Payload,
	}

	p.logger.Info("INTAKE_WORKER_CLAIMED",
		zap.Int64("event_id", ev.ID),
		zap.String("kind", ev.Kind),
		zap.String("serialisation_key", ev.SerialisationKey),
		zap.String("worker", workerID),
		zap.Int("attempt", ev.Attempts))

	// A once-a-minute pulse while the event runs: the visibility that lets a
	// watchdog (or an operator's grep) tell a slow event from a dead pool.
	start := time.Now()
	pulseCtx, pulseCancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-pulseCtx.Done():
				return
			case <-ticker.C:
				p.logger.Warn("INTAKE_EVENT_SLOW",
					zap.Int64("event_id", ev.ID),
					zap.String("serialisation_key", ev.SerialisationKey),
					zap.String("worker", workerID),
					zap.Duration("running_for", time.Since(start)))
			}
		}
	}()

	p.agent.processMessage(msg, ev.Kind)

	pulseCancel()
	if err := p.repo.MarkEventDone(context.Background(), ev.ID); err != nil {
		// The event RAN; only the bookkeeping failed. Leaving it 'running'
		// means a future takeover re-runs it, and dedupe suppresses the side
		// effects — prefer that to inventing a second status write path.
		p.logger.Error("INTAKE_MARK_DONE_FAILED: event ran, row stays running until takeover",
			zap.Int64("event_id", ev.ID), zap.Error(err))
		return
	}
	p.logger.Info("INTAKE_EVENT_DONE",
		zap.Int64("event_id", ev.ID),
		zap.String("serialisation_key", ev.SerialisationKey),
		zap.Duration("took", time.Since(start)))
}

func (p *intakeWorkerPool) runPurger() {
	defer p.wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if purged, err := p.repo.PurgeDoneEvents(p.ctx, p.retention); err == nil && purged > 0 {
				p.logger.Info("INTAKE_PURGED", zap.Int64("rows", purged))
			}
		}
	}
}

func (p *intakeWorkerPool) sleep(d time.Duration) {
	select {
	case <-p.ctx.Done():
	case <-time.After(d):
	}
}
