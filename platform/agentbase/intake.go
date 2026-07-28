// FILE: platform/agentbase/intake.go
//
// chassis_replica_scaling CS-2 (P1a): thin ingest. When CHASSIS_INTAKE_MODE is
// set, the consume loops stop executing orchestrations inline and instead
// persist each message as a chassis_intake_events row (milliseconds), commit
// the offset, and fetch again — the claim-worker pool (intake_workers.go) does
// the executing. Cross-orchestration head-of-line blocking ends here: a
// council that runs for nine minutes occupies one worker's claim, not a lane
// (bugs_open/030 residual, bugs_open/096 mechanism).
//
// Modes:
//   ""                — (default) inline processing, byte-identical to today.
//   "worker_pool"     — requests through the pool; responses stay inline.
//   "worker_pool_all" — requests AND responses through the pool (Phase 3;
//                       ships dark in the same image, enabled separately).
//
// The guards mirror setupExtraRequestLanes, for the same reason: spawned Job
// pods inherit personae-prod-config wholesale, and a spawned pod running
// chassis intake would execute chassis orchestrations under the wrong agent
// identity. >>> NEVER put CHASSIS_INTAKE_MODE in personae-prod-config. <<<

package agentbase

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"go.uber.org/zap"
)

const (
	intakeModeEnv           = "CHASSIS_INTAKE_MODE"
	intakeModeWorkerPool    = "worker_pool"
	intakeModeWorkerPoolAll = "worker_pool_all"

	// Backpressure (CS-2c, guardian objection on corr 9f0499b9): thin ingest
	// removes Kafka's implicit throttling — the blocking consume loop WAS the
	// backpressure — so the table must impose its own bound. When the not-done
	// backlog reaches the cap, persistIntake stalls BEFORE inserting; the lane
	// stops advancing, messages accumulate in Kafka exactly as they do today,
	// and every existing LAG monitor sees it. The bound is structural, not a
	// runbook instruction.
	intakeMaxPendingEnv     = "CHASSIS_INTAKE_MAX_PENDING"
	intakeMaxPendingDefault = 5000
)

// intakeKeyNamespace makes degenerate serialisation keys deterministic: the
// same unresolvable request id always lands in the same order domain, so a
// redelivered copy cannot run concurrently with the original.
var intakeKeyNamespace = uuid.MustParse("8f0e2f5c-5b8a-4c2e-9d3a-1f6b7a9c0d21")

// setupIntake decides whether this process runs the intake path, and wires the
// repository if so. Called from setupConsumers with the same mainTopic the
// lane guard uses, AFTER the DB is initialised.
func (a *Agent) setupIntake(mainTopic string) {
	mode := strings.TrimSpace(os.Getenv(intakeModeEnv))
	if mode == "" {
		return
	}
	if mode != intakeModeWorkerPool && mode != intakeModeWorkerPoolAll {
		a.logger.Warn("CHASSIS_INTAKE_MODE set to an unknown value — ignoring it (inline processing)",
			zap.String("value", mode))
		return
	}

	// Statically deployed agents only — structurally, not by deployment
	// discipline (see the header and setupExtraRequestLanes' identical guard).
	if !strings.HasPrefix(mainTopic, "system.agent.") || a.spawned {
		a.logger.Warn("CHASSIS_INTAKE_MODE is set but this agent is not statically deployed — ignoring it",
			zap.String("mode", mode),
			zap.String("main_topic", mainTopic),
			zap.Bool("spawned", a.spawned))
		return
	}
	if a.db == nil {
		a.logger.Error("CHASSIS_INTAKE_MODE is set but this agent has no database — ignoring it (inline processing)",
			zap.String("mode", mode))
		return
	}

	a.intakeRepo = orchestration.NewIntakeRepository(a.db, a.logger)
	a.intakeRequests = true
	a.intakeResponses = mode == intakeModeWorkerPoolAll
	a.intakeMaxPending = intakeMaxPendingDefault
	if v, err := strconv.Atoi(os.Getenv(intakeMaxPendingEnv)); err == nil && v > 0 {
		a.intakeMaxPending = v
	}

	// The pod-grep discriminator for deploy verification: this exact token is
	// new with CS-2 and logged once at startup.
	a.logger.Info("INTAKE_MODE_ACTIVE: consume loops persist, claim-workers execute (chassis_replica_scaling CS-2)",
		zap.String("mode", mode),
		zap.Bool("requests_via_pool", a.intakeRequests),
		zap.Bool("responses_via_pool", a.intakeResponses))
}

func (a *Agent) intakeForRequests() bool  { return a.intakeRepo != nil && a.intakeRequests }
func (a *Agent) intakeForResponses() bool { return a.intakeRepo != nil && a.intakeResponses }

// persistIntake makes one consumed message durable, then returns so the caller
// can commit the offset. It does NOT return on persistent DB failure — it
// retries with backoff until the insert lands or the agent stops. That is
// deliberate: kafka-go's FetchMessage advances the session position whether or
// not we commit, so "give up and continue the loop" would SKIP the message for
// the rest of the session. A stalled lane during a DB outage is exactly what
// the inline path gives today (processMessage cannot run without the DB
// either); a silently skipped dispatch is not.
func (a *Agent) persistIntake(msg kafka.Message, kind string) bool {
	// Backpressure gate FIRST: while the not-done backlog is at the cap, this
	// loop stalls and the lane stops advancing — Kafka accumulates and every
	// existing LAG monitor sees it, exactly the bound the blocking consume
	// loop used to provide implicitly. One indexed count per message; ingest
	// is milliseconds either way at this fleet's dispatch rates.
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		backlog, err := a.intakeRepo.PendingBacklog(ctx)
		cancel()
		if err == nil && backlog < a.intakeMaxPending {
			break
		}
		if err == nil {
			a.logger.Warn("INTAKE_BACKPRESSURE: backlog at cap, lane holds",
				zap.Int("backlog", backlog),
				zap.Int("cap", a.intakeMaxPending),
				zap.String("topic", msg.Topic))
		}
		select {
		case <-a.ctx.Done():
			return false
		case <-time.After(2 * time.Second):
		}
	}

	headers := kafka.HeadersToMap(msg.Headers)
	key, orchestrationID, correlationID, requestID := a.intakeSerialisationKey(headers, kind)

	ev := &orchestration.IntakeEvent{
		Topic:            msg.Topic,
		Partition:        msg.Partition,
		Offset:           msg.Offset,
		Kind:             kind,
		SerialisationKey: key,
		Headers:          headers,
		Payload:          msg.Value,
	}

	backoff := 100 * time.Millisecond
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		inserted, err := a.intakeRepo.InsertIntakeEvent(ctx, ev, orchestrationID, correlationID, requestID)
		cancel()

		if err == nil {
			if inserted {
				a.logger.Info("INTAKE_PERSISTED",
					zap.String("kind", kind),
					zap.String("topic", msg.Topic),
					zap.Int64("kafka_offset", msg.Offset),
					zap.String("serialisation_key", key),
					zap.String("correlation_id", correlationID))
			} else {
				// Transport redelivery (crash between insert and commit last
				// time). The row is already there; re-committing is the fix.
				a.logger.Info("INTAKE_DUPLICATE_DELIVERY: row exists, re-committing offset",
					zap.String("topic", msg.Topic),
					zap.Int64("kafka_offset", msg.Offset))
			}
			return true
		}

		select {
		case <-a.ctx.Done():
			// Shutting down without the insert landing: do NOT commit. The
			// redelivery on next start hits ON CONFLICT or inserts cleanly.
			a.logger.Warn("INTAKE_PERSIST_ABANDONED: shutdown before insert landed — offset stays uncommitted",
				zap.String("topic", msg.Topic),
				zap.Int64("kafka_offset", msg.Offset),
				zap.Error(err))
			return false
		default:
		}

		a.logger.Error("INTAKE_PERSIST_RETRY: insert failed, lane holds until it lands",
			zap.String("topic", msg.Topic),
			zap.Int64("kafka_offset", msg.Offset),
			zap.Duration("backoff", backoff),
			zap.Error(err))
		time.Sleep(backoff)
		if backoff < 5*time.Second {
			backoff *= 2
		}
	}
}

// intakeSerialisationKey derives the order domain for a message.
//
// Requests: the orchestration_id header — every dispatch carries one, and it
// is the exact unit UpdateStateWithVersion serialises on. Fallbacks keep a
// malformed message flowing (processMessage owns rejecting it): correlation_id
// if it parses, else a fresh key (alone in its own domain).
//
// Responses: the request id resolved through awaited_requests to the PARENT
// orchestration. A response's own orchestration_id header names the CHILD's
// run — keying on it would serialise against the wrong domain. An unresolvable
// request id gets a DETERMINISTIC degenerate key (uuid5 of the id) so a
// redelivered copy shares the original's domain; ClaimAwaitedRequest remains
// the semantic duplicate arbiter either way.
func (a *Agent) intakeSerialisationKey(headers map[string]string, kind string) (key, orchestrationID, correlationID, requestID string) {
	if id := headers["correlation_id"]; uuidParses(id) {
		correlationID = id
	}

	if kind == "request" {
		requestID = headers["request_id"]
		if id := headers["orchestration_id"]; uuidParses(id) {
			return id, id, correlationID, requestID
		}
		if correlationID != "" {
			a.logger.Warn("INTAKE_KEY_FALLBACK: request without a parseable orchestration_id — keying on correlation_id",
				zap.String("correlation_id", correlationID))
			return correlationID, "", correlationID, requestID
		}
		fresh := uuid.New().String()
		a.logger.Warn("INTAKE_KEY_FALLBACK: request without orchestration_id or correlation_id — fresh key, own order domain",
			zap.String("key", fresh))
		return fresh, "", "", requestID
	}

	// kind == "response"
	requestID = headers["in_response_to_request_id"]
	if requestID == "" {
		requestID = headers["request_id"]
	}
	if requestID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		parentOrch, err := a.intakeRepo.ResolveResponseOrchestration(ctx, requestID)
		cancel()
		if err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Warn("INTAKE_KEY_LOOKUP_FAILED: awaited_requests unreachable — degenerate key",
				zap.String("request_id", requestID), zap.Error(err))
		}
		if parentOrch != "" && uuidParses(parentOrch) {
			return parentOrch, parentOrch, correlationID, requestID
		}
		return uuid.NewSHA1(intakeKeyNamespace, []byte(requestID)).String(), "", correlationID, requestID
	}

	fresh := uuid.New().String()
	a.logger.Warn("INTAKE_KEY_FALLBACK: response without any request id — fresh key, own order domain",
		zap.String("key", fresh))
	return fresh, "", correlationID, ""
}

func uuidParses(s string) bool {
	if s == "" {
		return false
	}
	_, err := uuid.Parse(s)
	return err == nil
}
