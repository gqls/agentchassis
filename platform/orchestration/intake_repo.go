// FILE: platform/orchestration/intake_repo.go
//
// chassis_replica_scaling CS-2 (P1a): the Postgres side of "Kafka delivers,
// Postgres decides". The chassis's consume loops persist each message here in
// milliseconds and commit the offset; a pool of claim-workers executes events
// with per-orchestration ordering enforced by chassis_orchestration_claims.
// Schema: docs/agent_docs/sql_for_agents/249_chassis_intake_events.sql.
//
// Concurrency contract, in one place:
//   - (topic, partition, kafka_offset) UNIQUE  → transport-level exactly-once
//     intake. A crash between insert and offset-commit redelivers the message;
//     the redelivery hits ON CONFLICT DO NOTHING and is re-committed.
//   - chassis_orchestration_claims PK          → at most one worker per
//     serialisation key at any instant, whatever the interleaving. The claim
//     is the same ON-CONFLICT lease-takeover shape as RecordMessageProcessing
//     (state.go), which bugs_open/003 already paid to review.
//   - a single-query "claim the next event whose key has no running peer"
//     (NOT EXISTS + FOR UPDATE SKIP LOCKED) was considered and REJECTED: under
//     READ COMMITTED two workers can skip each other's uncommitted locks onto
//     two events of the SAME key. The claims-table CAS cannot be raced that
//     way, at the cost of one extra round-trip per key acquisition.
//   - only the claim holder pops events for its key, so NextPendingEvent needs
//     no row locks — its guarded UPDATE is belt-and-braces.

package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// IntakeRepository mediates every access to chassis_intake_events and
// chassis_orchestration_claims.
type IntakeRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewIntakeRepository(db *sql.DB, logger *zap.Logger) *IntakeRepository {
	return &IntakeRepository{db: db, logger: logger}
}

// IntakeEvent is one persisted Kafka message awaiting (or under) execution.
type IntakeEvent struct {
	ID               int64
	Topic            string
	Partition        int
	Offset           int64
	Kind             string // "request" | "response" — processMessage's messageType
	SerialisationKey string
	Headers          map[string]string
	Payload          []byte
	Attempts         int
}

// InsertIntakeEvent persists one consumed message. Reports false when the row
// already existed (transport redelivery) — the caller commits the offset either
// way, which is what makes intake exactly-once.
func (r *IntakeRepository) InsertIntakeEvent(ctx context.Context, ev *IntakeEvent, orchestrationID, correlationID, requestID string) (bool, error) {
	headersJSON, err := json.Marshal(ev.Headers)
	if err != nil {
		return false, fmt.Errorf("marshal intake headers: %w", err)
	}

	// Empty ids become NULL rather than '' so the uuid columns stay honest.
	query := `
		INSERT INTO chassis_intake_events
			(topic, partition, kafka_offset, kind, serialisation_key,
			 orchestration_id, correlation_id, request_id, headers, payload)
		VALUES ($1, $2, $3, $4, $5::uuid,
		        NULLIF($6,'')::uuid, NULLIF($7,'')::uuid, NULLIF($8,''), $9, $10)
		ON CONFLICT (topic, partition, kafka_offset) DO NOTHING
	`
	result, err := r.db.ExecContext(ctx, query,
		ev.Topic, ev.Partition, ev.Offset, ev.Kind, ev.SerialisationKey,
		orchestrationID, correlationID, requestID, headersJSON, ev.Payload)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// CandidateKeys returns up to limit serialisation keys that have claimable
// work and no live claim, oldest first. Read-only and deliberately
// approximate — two workers may see the same key; ClaimSerialisationKey
// decides the winner.
//
// 'running' events count as claimable work (CS-2d, found live 2026-07-28): a
// running event under an EXPIRED claim belongs to a dead holder, and if the
// key has no pending siblings a pending-only scan never surfaces it — the
// takeover reset that would recover it only runs on a claim, so the event is
// orphaned forever (two real dispatches sat that way for 90 minutes). A
// running event under a LIVE claim never reaches a worker: the NOT EXISTS
// excludes its key, and the lease heartbeat keeps that true while the holder
// works.
func (r *IntakeRepository) CandidateKeys(ctx context.Context, limit int) ([]string, error) {
	query := `
		SELECT e.serialisation_key
		FROM chassis_intake_events e
		WHERE e.status IN ('pending','running')
		  AND NOT EXISTS (
		        SELECT 1 FROM chassis_orchestration_claims c
		        WHERE c.serialisation_key = e.serialisation_key
		          AND c.lease_expires_at > NOW())
		GROUP BY e.serialisation_key
		ORDER BY MIN(e.id)
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// ClaimSerialisationKey atomically acquires the key for claimedBy, taking over
// an expired lease in the same statement. Reports whether the claim was won.
// The PK makes a double-claim unrepresentable; there is no interleaving in
// which two callers both see rows==1 while a lease is live.
func (r *IntakeRepository) ClaimSerialisationKey(ctx context.Context, key, claimedBy string, lease time.Duration) (bool, error) {
	query := `
		INSERT INTO chassis_orchestration_claims (serialisation_key, claimed_by, lease_expires_at)
		VALUES ($1::uuid, $2, NOW() + make_interval(secs => $3))
		ON CONFLICT (serialisation_key) DO UPDATE
		   SET claimed_by       = EXCLUDED.claimed_by,
		       claimed_at       = NOW(),
		       lease_expires_at = EXCLUDED.lease_expires_at
		 WHERE chassis_orchestration_claims.lease_expires_at <= NOW()
	`
	result, err := r.db.ExecContext(ctx, query, key, claimedBy, lease.Seconds())
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// HeartbeatClaim extends the lease. Reports false when the claim is no longer
// held by claimedBy — the lease lapsed and another worker took the key. The
// holder MUST stop draining the key when that happens: continuing would run
// events concurrently with the new holder, which is the one thing the claims
// table exists to prevent.
func (r *IntakeRepository) HeartbeatClaim(ctx context.Context, key, claimedBy string, lease time.Duration) (bool, error) {
	query := `
		UPDATE chassis_orchestration_claims
		SET lease_expires_at = NOW() + make_interval(secs => $3)
		WHERE serialisation_key = $1::uuid AND claimed_by = $2
	`
	result, err := r.db.ExecContext(ctx, query, key, claimedBy, lease.Seconds())
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// ReleaseClaim drops the claim if claimedBy still holds it. Losing the race
// (lease lapsed, someone took over) is not an error — the new holder owns it.
func (r *IntakeRepository) ReleaseClaim(ctx context.Context, key, claimedBy string) error {
	query := `DELETE FROM chassis_orchestration_claims WHERE serialisation_key = $1::uuid AND claimed_by = $2`
	_, err := r.db.ExecContext(ctx, query, key, claimedBy)
	return err
}

// ResetRunningEvents returns a dead holder's in-flight events to pending.
// Called by a worker immediately after WINNING a claim: any 'running' row for
// a key you have just claimed belonged to a holder whose lease expired —
// double-execution of the re-run is suppressed one level down by the
// processed_messages two-phase dedupe, exactly as it is for a redelivered
// Kafka message today (bugs_open/003 F3).
func (r *IntakeRepository) ResetRunningEvents(ctx context.Context, key string) (int64, error) {
	query := `
		UPDATE chassis_intake_events
		SET status = 'pending', started_at = NULL
		WHERE serialisation_key = $1::uuid AND status = 'running'
	`
	result, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// NextPendingEvent pops the oldest pending event for the key (pending →
// running, attempts+1). Returns nil when the key is drained. Only the claim
// holder calls this, so the guarded UPDATE is protection in depth, not the
// ordering mechanism.
func (r *IntakeRepository) NextPendingEvent(ctx context.Context, key string) (*IntakeEvent, error) {
	query := `
		UPDATE chassis_intake_events
		SET status = 'running', started_at = NOW(), attempts = attempts + 1
		WHERE id = (SELECT id FROM chassis_intake_events
		            WHERE serialisation_key = $1::uuid AND status = 'pending'
		            ORDER BY id LIMIT 1)
		  AND status = 'pending'
		RETURNING id, topic, partition, kafka_offset, kind, serialisation_key,
		          headers, payload, attempts
	`
	var (
		ev          IntakeEvent
		headersJSON []byte
	)
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&ev.ID, &ev.Topic, &ev.Partition, &ev.Offset, &ev.Kind,
		&ev.SerialisationKey, &headersJSON, &ev.Payload, &ev.Attempts)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(headersJSON, &ev.Headers); err != nil {
		return nil, fmt.Errorf("unmarshal intake headers for event %d: %w", ev.ID, err)
	}
	return &ev, nil
}

// MarkEventDone is UNCONDITIONAL on the in-process outcome, mirroring
// commitConsumed's contract (agentbase/agent.go): handler errors have already
// routed through handleProcessingError and the parent drives the retry.
func (r *IntakeRepository) MarkEventDone(ctx context.Context, id int64) error {
	query := `UPDATE chassis_intake_events SET status = 'done', finished_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// MarkEventFailed is for infrastructure-level inability to run the event at
// all (attempts exhausted after repeated holder deaths) — never for handler
// errors, which are the parent's to retry.
func (r *IntakeRepository) MarkEventFailed(ctx context.Context, id int64, reason string) error {
	query := `UPDATE chassis_intake_events SET status = 'failed', finished_at = NOW(), last_error = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, reason)
	return err
}

// PurgeDoneEvents deletes done rows older than the retention window. Failed
// rows are deliberately kept — they are the durable record of the only losses
// this layer can cause.
func (r *IntakeRepository) PurgeDoneEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM chassis_intake_events
		WHERE status = 'done' AND finished_at < NOW() - make_interval(secs => $1)
	`
	result, err := r.db.ExecContext(ctx, query, olderThan.Seconds())
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// PendingBacklog counts events not yet done — the figure the ingest loops
// gate on (CS-2c, council corr 9f0499b9 guardian objection: removing Kafka's
// implicit backpressure must not leave the table unbounded). Uses the partial
// index idx_cie_pending.
func (r *IntakeRepository) PendingBacklog(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM chassis_intake_events WHERE status IN ('pending','running')`).Scan(&n)
	return n, err
}

// ResolveResponseOrchestration maps a response's request id to the PARENT
// orchestration that awaits it, via awaited_requests. A response's own
// orchestration_id header names the CHILD's run — serialising on it would
// order responses against the wrong domain. Reports "" (no error) when the
// row does not exist yet; the caller falls back to a degenerate key and
// ClaimAwaitedRequest remains the semantic arbiter.
func (r *IntakeRepository) ResolveResponseOrchestration(ctx context.Context, requestID string) (string, error) {
	var orchestrationID string
	err := r.db.QueryRowContext(ctx,
		`SELECT orchestration_id FROM awaited_requests WHERE request_id = $1`,
		requestID).Scan(&orchestrationID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return orchestrationID, nil
}
