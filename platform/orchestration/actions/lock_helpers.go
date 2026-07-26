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
// The CHROME half (site_components: header/footer/head, bugs_open/069) lives in
// site_component_lock_guard.go and shares this file's predicate, its hard/soft
// classifier (classifyComponentLock) and its work-item plumbing
// (emitLockBlockedChange).
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

	return classifyComponentLock(lockedAt, lockedBy, lockType, lockExpiresAt, agentWritable), nil
}

// classifyComponentLock turns the five columns every lock-bearing table shares
// into a ComponentLockStatus. Pure — the caller does its own SELECT (the tables
// differ in how you address a row: page_components by id, site_components by
// site_id+slot_name) and hands the scanned values here, so the hard/soft rule
// below can never drift between surfaces.
//
// Hard = never auto-expires; only a human unlock releases it. lock_type is the
// classifier (lock_policy.go, permanent <=> human). A locked row with no
// lock_type predates the policy stamp — treat as hard, conservatively. (The old
// locked_by IN ('admin',...) switch misclassified every live lock: real rows
// carry free-text reasons in locked_by, e.g. '182_legal_pages'.)
func classifyComponentLock(lockedAt sql.NullTime, lockedBy, lockType sql.NullString,
	lockExpiresAt sql.NullTime, agentWritable bool) *ComponentLockStatus {

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
	if status.IsLocked {
		status.IsHard = IsHardLockType(status.LockType) || status.LockType == ""
	}
	return status
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

	emitLockBlockedChange(ctx, db, lockBlockedChange{
		siteID:        siteID,
		pageID:        pageID,
		componentID:   componentID,
		surface:       "page_component",
		slotName:      slotName,
		lockedBy:      lockedBy,
		lockType:      lockType,
		blockedAction: blockedAction,
		sourceAction:  sourceAction,
		itemKey:       fmt.Sprintf("lock_blocked_change:%s:%s", pageName, slotName),
		summary: fmt.Sprintf("Lock held: %s wanted to %s locked section %q on page %q (locked by: %s)",
			sourceAction, blockedAction, slotName, pageName, lockedBy),
		fixText: "A human lock on this section blocked an automated change; the " +
			"locked copy was preserved (bugs_open/058). Decide: keep the lock (the " +
			"planned change stays unapplied — no action needed), or unlock via the " +
			"admin dashboard and re-run the change if the section should follow " +
			"the plan again.",
		extraSpec: map[string]interface{}{"page_name": pageName},
	}, logger)
}

// lockBlockedChange carries everything that DIFFERS between the surfaces a lock
// can block. What is shared — and therefore lives in emitLockBlockedChange
// below — is the plumbing: the idx_swi_dedup-matched ON CONFLICT via
// insertWorkItem, the no-handler_agent routing, the 250-char summary cap, and
// log-never-return. The prose is not shared: "section on page X" is meaningless
// for a site-wide chrome slot, and the consequence a human needs to read
// differs (stale nav across every page, versus one unchanged section).
//
// severity defaults to "medium" when empty.
type lockBlockedChange struct {
	siteID        uuid.UUID
	pageID        *uuid.UUID
	componentID   *uuid.UUID
	surface       string // page_component | site_component
	slotName      string
	lockedBy      string
	lockType      string
	blockedAction string // overwrite | remove | edit | relink | rerender
	sourceAction  string
	severity      string
	itemKey       string
	summary       string
	fixText       string
	extraSpec     map[string]interface{}
}

// emitLockBlockedChange is the shared body behind emitLockBlockedChangeItem
// (page sections, bugs_open/058) and emitChromeLockBlockedChangeItem (site
// chrome, bugs_open/069).
func emitLockBlockedChange(ctx context.Context, db *sql.DB, c lockBlockedChange, logger *zap.Logger) {
	spec := map[string]interface{}{
		"surface":        c.surface,
		"slot_name":      c.slotName,
		"locked_by":      c.lockedBy,
		"lock_type":      c.lockType,
		"blocked_action": c.blockedAction,
		"source":         c.sourceAction,
		"fix":            c.fixText,
	}
	if c.componentID != nil {
		spec["component_id"] = c.componentID.String()
	}
	for k, v := range c.extraSpec {
		spec[k] = v
	}
	specJSON, _ := json.Marshal(spec)

	summary := c.summary
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	severity := c.severity
	if severity == "" {
		severity = "medium"
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("emitLockBlockedChange: begin tx failed",
			zap.String("surface", c.surface), zap.String("slot", c.slotName), zap.Error(err))
		return
	}
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:      c.siteID,
		pageID:      c.pageID,
		componentID: c.componentID,
		source:      c.sourceAction,
		pipeline:    "build",
		itemType:    "lock_blocked_change",
		severity:    severity,
		summary:     summary,
		spec:        string(specJSON),
		priority:    30,
		status:      "needs_human_review",
		createdBy:   c.sourceAction,
		itemKey:     c.itemKey,
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("emitLockBlockedChange: insert failed",
			zap.String("surface", c.surface), zap.String("slot", c.slotName), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("emitLockBlockedChange: commit failed",
			zap.String("surface", c.surface), zap.String("slot", c.slotName), zap.Error(err))
	}
}
