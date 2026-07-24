// FILE: platform/orchestration/actions/lock_helpers.go
//
// Shared helpers for checking component lock status (bugs_open/058).
// Used by execution agents that write to page_components and need
// to respect human locks.
//
// Discovery agents don't need this — they filter via SQL (AND pc.locked_at IS NULL).
// Writers use pageComponentAgentWritableSQL directly in their UPDATE/DELETE
// predicates (TOCTOU-free), and CheckComponentLock when they need the lock's
// details for reporting.
//
// Lock semantics come from the applied 053/115 schema migration
// (docs/agent_docs/sql_for_agents/115_locks.sql) and lock_policy.go: a row is
// agent-writable iff it carries no lock, or only a timed lock whose expiry has
// passed. Permanent locks (human) never expire; a NULL lock_type on a locked
// row is treated as permanent (conservative — never silently overwrite what we
// can't classify).

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// pageComponentAgentWritableSQL returns the canonical "may automation write
// this row?" predicate for page_components (and site_components — same
// columns). alias is the table alias including the trailing dot ("pc.") or ""
// when the statement has no alias. Single source of truth: every automated
// writer's UPDATE/DELETE must use this fragment, not its own lock test.
func pageComponentAgentWritableSQL(alias string) string {
	return fmt.Sprintf(
		"(%[1]slocked_at IS NULL OR (%[1]slock_type = 'timed' AND %[1]slock_expires_at IS NOT NULL AND %[1]slock_expires_at < NOW()))",
		alias)
}

// ComponentLockStatus represents the lock state of a page component.
type ComponentLockStatus struct {
	IsLocked      bool // an ACTIVE lock (expiry-aware): automation must not overwrite
	LockedBy      string
	LockedAt      *time.Time
	LockType      string // permanent | timed | review | "" (unset)
	LockExpiresAt *time.Time
	IsHard        bool // permanent (or unclassified) = hard; only a human unlock releases it
}

// CheckComponentLock checks whether a page_component carries an active lock.
// Returns the lock status. Callers decide whether to skip based on the status;
// the write itself should still be guarded with pageComponentAgentWritableSQL
// in its WHERE clause so the decision is race-free.
func CheckComponentLock(ctx context.Context, db *sql.DB, componentID uuid.UUID, logger *zap.Logger) (*ComponentLockStatus, error) {
	var lockedAt, lockExpiresAt sql.NullTime
	var lockedBy, lockType sql.NullString
	var agentWritable bool

	err := db.QueryRowContext(ctx, `
		SELECT locked_at, locked_by, lock_type, lock_expires_at, `+pageComponentAgentWritableSQL("")+`
		FROM page_components WHERE id = $1
	`, componentID).Scan(&lockedAt, &lockedBy, &lockType, &lockExpiresAt, &agentWritable)

	if err == sql.ErrNoRows {
		return &ComponentLockStatus{IsLocked: false}, nil
	}
	if err != nil {
		return nil, err
	}

	status := &ComponentLockStatus{
		IsLocked: lockedAt.Valid && !agentWritable,
		LockedBy: lockedBy.String,
		LockType: lockType.String,
	}
	if lockedAt.Valid {
		status.LockedAt = &lockedAt.Time
	}
	if lockExpiresAt.Valid {
		status.LockExpiresAt = &lockExpiresAt.Time
	}

	// Hard = never auto-expires; only a human unlock releases it. lock_type is
	// the classifier (lock_policy.go, permanent <=> human). A locked row with
	// no lock_type predates the policy stamp — treat as hard, conservatively.
	// (The old locked_by IN ('admin',...) switch misclassified every live lock:
	// real rows carry free-text reasons in locked_by, e.g. '182_legal_pages'.)
	if status.IsLocked {
		status.IsHard = IsHardLockType(status.LockType) || status.LockType == ""
	}

	return status, nil
}

// CheckPageHasHardLocks counts components on a page carrying a hard (human)
// lock. Classification matches CheckComponentLock: lock_type='permanent', or a
// locked row with no lock_type stamped.
func CheckPageHasHardLocks(ctx context.Context, db *sql.DB, pageID uuid.UUID) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM page_components
		WHERE page_id = $1
		  AND locked_at IS NOT NULL
		  AND (lock_type = 'permanent' OR lock_type IS NULL)
	`, pageID).Scan(&count)
	return count, err
}

// emitLockBlockedChangeItem files ONE work item when an active component lock
// blocked a change an automated writer wanted to make (bugs_open/058 candidate
// 3): a silent skip trades one silent failure for another, so the blocked
// intent must surface. Routing mirrors the dead-control family
// (emitChromeDeadControlItem / check_dead_controls): status
// 'needs_human_review' with NO handler_agent — whether the lock should yield
// is a human decision (unlock, or accept that the plan's change does not
// apply), and the 033 dashboard surfaces the queue. Persistence goes through
// the shared insertWorkItem helper so it inherits the idx_swi_dedup-matched ON
// CONFLICT (no 42P10 drift) — repeat rebuilds against the same lock collapse
// into one open item. A failure is logged, never returned: the signal must not
// block the write path that raised it.
//
// blockedAction is one of "overwrite" (rebuild produced fresh copy for a
// locked slot), "remove" (recomposition no longer includes the locked slot),
// "edit" (a targeted section edit hit the lock).
func emitLockBlockedChangeItem(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	pageID, componentID *uuid.UUID, pageName, slotName, lockedBy, lockType,
	blockedAction, sourceAction string, logger *zap.Logger) {

	spec := map[string]interface{}{
		"surface":        "page_component",
		"page_name":      pageName,
		"slot_name":      slotName,
		"locked_by":      lockedBy,
		"lock_type":      lockType,
		"blocked_action": blockedAction,
		"source":         sourceAction,
		"fix": "A human lock on this section blocked an automated change; the " +
			"locked copy was preserved (bugs_open/058). Decide: keep the lock (the " +
			"planned change stays unapplied — no action needed), or unlock via the " +
			"admin dashboard and re-run the change if the section should follow " +
			"the plan again.",
	}
	if componentID != nil {
		spec["component_id"] = componentID.String()
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("Lock held: %s wanted to %s locked section %q on page %q (locked by: %s)",
		sourceAction, blockedAction, slotName, pageName, lockedBy)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}
	itemKey := fmt.Sprintf("lock_blocked_change:%s:%s", pageName, slotName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("emitLockBlockedChangeItem: begin tx failed",
			zap.String("slot", slotName), zap.Error(err))
		return
	}
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:      siteID,
		pageID:      pageID,
		componentID: componentID,
		source:      sourceAction,
		pipeline:    "build",
		itemType:    "lock_blocked_change",
		severity:    "medium",
		summary:     summary,
		spec:        string(specJSON),
		priority:    30,
		status:      "needs_human_review",
		createdBy:   sourceAction,
		itemKey:     itemKey,
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("emitLockBlockedChangeItem: insert failed",
			zap.String("slot", slotName), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("emitLockBlockedChangeItem: commit failed",
			zap.String("slot", slotName), zap.Error(err))
	}
}
