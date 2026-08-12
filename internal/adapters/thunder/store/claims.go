// FILE: internal/adapters/thunder/store/claims.go
//
// Idempotency + audit for provision_instance, keyed on correlation_id.
// Pure transport over *sql.DB — no business logic beyond the atomicity the
// dedup depends on.
//
// WHY THIS EXISTS (bugs_open/259 — resolve that number by SLUG,
// "..._one_provision_request_builds_several_billable_gpus"):
//
//	The dispatch_provision step awaits a response for 600s. When that await
//	expires, the chassis retry driver re-executes the step
//	(coordinator.go retryExpiredAwaitedRequest, budget RetryVersion < 3).
//	Each re-execution mints a NEW request_id and publishes a NEW provision
//	message, so one logical request can build up to FOUR billable GPUs.
//	Measured: orchestration 8c5bf926 produced 4 awaited_requests rows with 4
//	distinct request_ids, each sent ~1s after the previous one's timeout_at.
//
//	correlation_id is the ONLY identifier stable across those attempts.
//	Keying on request_id would never fire — and request_id is exactly what
//	the dispatch code makes look canonical, which is the trap.
//
// The claim is taken BEFORE the vendor call. That ordering is the whole
// mechanism: a claim taken afterwards would leave the create/claim window
// open, and that window is where the money is spent.

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrClaimHeld is returned by TakeProvisionClaim when a claim for this
// correlation already exists — i.e. this is a repeat attempt at a request
// that has already reached the vendor once. The caller MUST NOT provision.
var ErrClaimHeld = errors.New("provision claim already held for this correlation")

// Claim statuses. 'claimed' is the pre-vendor state, so a row left in it
// means the adapter died between taking the claim and the create returning.
const (
	ClaimStatusClaimed   = "claimed"
	ClaimStatusCreated   = "created"
	ClaimStatusSucceeded = "succeeded"
	ClaimStatusFailed    = "failed"
)

// ClaimInsertSQL is the claim statement. Exported so a test can assert its
// SHAPE, which is load-bearing: claiming and counting must happen in ONE
// statement. A read-then-insert pair would let two retry messages arriving
// close together both read "no claim" and both provision — the race this
// table exists to close.
//
// `xmax = 0` is Postgres's own answer to "did this INSERT insert, or did the
// ON CONFLICT arm fire?" — it is zero only for a freshly inserted tuple. That
// is the whole dedup verdict, decided by the database rather than by a
// count we compute and could get wrong.
const ClaimInsertSQL = `
		INSERT INTO thunder_provision_claims (
			correlation_id, orchestration_id, first_request_id,
			training_run_id, requested_by, attempts, status
		) VALUES ($1, $2, $3, $4, $5, 1, 'claimed')
		ON CONFLICT (correlation_id) DO UPDATE
			SET attempts = thunder_provision_claims.attempts + 1
		RETURNING correlation_id, attempts, status,
		          thunder_instance_id, provisioning_id, last_error,
		          (xmax = 0) AS inserted
	`

// ProvisionClaim mirrors a thunder_provision_claims row.
type ProvisionClaim struct {
	CorrelationID     string
	OrchestrationID   sql.NullString
	FirstRequestID    sql.NullString
	Attempts          int
	Status            string
	ThunderInstanceID sql.NullString
	ProvisioningID    sql.NullString
	LastError         sql.NullString
}

// NewClaim is the set of fields a first attempt records.
type NewClaim struct {
	CorrelationID   string
	OrchestrationID string // uuid-as-string; empty → NULL
	RequestID       string
	TrainingRunID   string // uuid-as-string; empty → NULL
	RequestedBy     string
}

// TakeProvisionClaim atomically claims the right to provision for one
// correlation. Exactly one caller can win, ever.
//
// Returns nil if the claim was taken (proceed to the vendor). Returns
// ErrClaimHeld — wrapped, with the existing row for the caller to report —
// if another attempt already holds it, in which case the caller must NOT
// create an instance. Any other error is infrastructural and also means
// "do not provision": this fails CLOSED by construction, because the failure
// mode it exists to prevent costs real money.
//
// The INSERT ... ON CONFLICT DO UPDATE bumps `attempts` on the losing path,
// so the row counts how hard the retry driver leaned on the door. Reading
// then inserting would reintroduce exactly the race this closes, so the
// count and the claim are one statement.
func TakeProvisionClaim(ctx context.Context, db *sql.DB, c NewClaim) (*ProvisionClaim, error) {
	if c.CorrelationID == "" {
		// No key means no dedup is possible. Refuse rather than provision
		// unguarded — an unkeyed request on this path is the exact shape
		// that spends money nobody can attribute.
		return nil, fmt.Errorf("provision request carries no correlation_id: refusing to provision without an idempotency key")
	}

	var (
		row      ProvisionClaim
		inserted bool
	)
	err := db.QueryRowContext(ctx, ClaimInsertSQL,
		c.CorrelationID,
		nullableText(c.OrchestrationID),
		nullableText(c.RequestID),
		nullableText(c.TrainingRunID),
		nullableText(c.RequestedBy),
	).Scan(
		&row.CorrelationID,
		&row.Attempts,
		&row.Status,
		&row.ThunderInstanceID,
		&row.ProvisioningID,
		&row.LastError,
		&inserted,
	)
	if err != nil {
		return nil, fmt.Errorf("take provision claim: %w", err)
	}

	if !inserted {
		return &row, ErrClaimHeld
	}
	return &row, nil
}

// MarkClaimCreated records the vendor identifier the instant CreateInstance
// returns, so a crash after that point still leaves the box attributable to
// the request that built it.
func MarkClaimCreated(ctx context.Context, db *sql.DB, correlationID, thunderInstanceID string) error {
	const q = `
		UPDATE thunder_provision_claims
		SET status = 'created', thunder_instance_id = $2
		WHERE correlation_id = $1
	`
	_, err := db.ExecContext(ctx, q, correlationID, thunderInstanceID)
	return err
}

// MarkClaimSucceeded records the finished provision and its DB row.
func MarkClaimSucceeded(ctx context.Context, db *sql.DB, correlationID, provisioningID string) error {
	const q = `
		UPDATE thunder_provision_claims
		SET status = 'succeeded', provisioning_id = $2::uuid
		WHERE correlation_id = $1
	`
	_, err := db.ExecContext(ctx, q, correlationID, provisioningID)
	return err
}

// MarkClaimFailed records why a provision failed. This is the durable record
// bugs_open/258 defect 3 says does not exist: today a failed provision writes
// no thunder_instances row and no agent_error_log row, so once the pod log
// rotates there is nothing left to audit.
//
// The claim deliberately STAYS held after a failure. A later attempt under the
// same correlation is refused rather than retried, because the retry driver's
// re-dispatch is not a considered decision to spend money again — it is a
// timeout firing.
func MarkClaimFailed(ctx context.Context, db *sql.DB, correlationID, errMsg string) error {
	const q = `
		UPDATE thunder_provision_claims
		SET status = 'failed', last_error = $2
		WHERE correlation_id = $1
	`
	_, err := db.ExecContext(ctx, q, correlationID, truncateForColumn(errMsg, 2000))
	return err
}

// nullableText maps "" → NULL so audit columns stay honestly empty rather
// than holding a zero-length string that reads as "recorded".
func nullableText(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// truncateForColumn bounds an error string so a pathological vendor payload
// cannot bloat the row.
func truncateForColumn(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}
