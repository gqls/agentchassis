// FILE: platform/orchestration/actions/lock_helpers.go
//
// Shared helper for checking component lock status.
// Used by execution agents that write to page_components and need
// to respect human locks on stale work items from before the lock was set.
//
// Discovery agents don't need this — they filter via SQL (AND pc.locked_at IS NULL).
// This is for execution agents that receive a work item referencing a specific
// component and need to check if it was locked after the item was created.

package actions

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ComponentLockStatus represents the lock state of a page component.
type ComponentLockStatus struct {
	IsLocked bool
	LockedBy string
	LockedAt *time.Time
	IsHard   bool // admin, admin-removed, checkpoint = hard; deploy = soft
}

// CheckComponentLock checks whether a page_component is locked by a human.
// Returns the lock status. Callers decide whether to skip based on lock type.
//
// Usage:
//
//	lock, err := CheckComponentLock(ctx, db, componentID, logger)
//	if err != nil { ... }
//	if lock.IsHard {
//	    // Component is human-locked, skip unless work item has force=true
//	    return map[string]interface{}{"skipped": true, "reason": "component locked by " + lock.LockedBy}, nil
//	}
func CheckComponentLock(ctx context.Context, db *sql.DB, componentID uuid.UUID, logger *zap.Logger) (*ComponentLockStatus, error) {
	var lockedAt sql.NullTime
	var lockedBy sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT locked_at, locked_by FROM page_components WHERE id = $1
	`, componentID).Scan(&lockedAt, &lockedBy)

	if err == sql.ErrNoRows {
		return &ComponentLockStatus{IsLocked: false}, nil
	}
	if err != nil {
		return nil, err
	}

	if !lockedAt.Valid {
		return &ComponentLockStatus{IsLocked: false}, nil
	}

	status := &ComponentLockStatus{
		IsLocked: true,
		LockedBy: lockedBy.String,
		LockedAt: &lockedAt.Time,
	}

	// Hard locks: set by humans, only humans can remove
	switch lockedBy.String {
	case "admin", "admin-removed", "checkpoint":
		status.IsHard = true
	default:
		// "deploy" or anything else is a soft lock
		status.IsHard = false
	}

	return status, nil
}

// CheckPageHasHardLocks checks if any component on a page has a human lock.
// Used by write_audit_findings to skip component-level findings for locked pages.
func CheckPageHasHardLocks(ctx context.Context, db *sql.DB, pageID uuid.UUID) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM page_components
		WHERE page_id = $1
		  AND locked_at IS NOT NULL
		  AND locked_by IN ('admin', 'admin-removed', 'checkpoint')
	`, pageID).Scan(&count)
	return count, err
}
