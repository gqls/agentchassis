// FILE: platform/orchestration/actions/site_component_lock_guard.go
//
// The write-side lock gate for SITE CHROME — the site_components table
// (header / footer / head, one row per site_id + slot_name). bugs_open/069,
// the residual bugs_open/058 spun out to keep its reviewed change narrow.
//
// site_components carries the same lock columns as page_components, from the
// same applied 053/115 migration and under the same CHECK constraint
// (permanent | timed | review), and the admin dashboard sets them
// (HandleLockSiteComponent / the auto-lock on the site-component edit
// endpoint). Until this file, no chrome writer read them back: a human who
// locked a header got it silently overwritten by the next FORCED chrome
// re-render, template fix, or component relink.
//
// Everything here reuses 058's machinery rather than restating it:
// pageComponentAgentWritableSQL is table-generic (bare column names, both
// tables share the column set — see its doc comment), classifyComponentLock
// owns the hard/soft rule, and emitLockBlockedChange owns the work-item
// plumbing. What is chrome-specific and therefore lives here: addressing a row
// by (site_id, slot_name) rather than id, the fact that a chrome slot may have
// no backing component at all, and what a blocked change MEANS for a human —
// chrome carries the nav, so a permanently locked header means new pages stop
// appearing in navigation on every page of the site.
//
// Two deliberate consequences, decided rather than discovered:
//
//  1. A locked slot whose rendered_html is EMPTY stays empty site-wide until a
//     human unlocks, because the page assembler filters empty chrome. Blocking
//     uniformly is still right — a softer policy cannot be expressed in the
//     WHERE predicate, and anything decided outside it reopens the TOCTOU the
//     predicate exists to close — but the emitted item says so
//     (artefact_empty).
//  2. The gate only bites when a caller FORCES the re-render. The unforced
//     path already no-ops on a populated slot, so gating above that check would
//     file items claiming a writer "wanted to change this" for calls that were
//     never going to write.

package actions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// errSiteComponentLocked is returned by the guarded chrome writers when the
// UPDATE matched no row: the slot carries an active human lock (or vanished
// under us). Callers convert it to a skip-RESULT, never an error — an error
// would fail and retry the orchestration against a state only a human unlock
// can change. Distinct from errComponentLocked, whose text names page
// components, so a wrapped or logged message stays true.
var errSiteComponentLocked = errors.New("site component is locked or missing — automated chrome write refused (bugs_open/069)")

// SiteComponentLockStatus is ComponentLockStatus plus the three things a chrome
// writer needs from the same row read, and cannot get from the page-side
// helper: whether the row exists at all (the render fallback and the linker
// both INSERT), whether it currently holds an artefact (a lock over an empty
// slot freezes the chrome), and which content component it points at (NULL is
// a real, detected state — checkUnlinkedSiteComponents exists for it).
type SiteComponentLockStatus struct {
	ComponentLockStatus

	RowExists   bool
	HasHTML     bool
	ComponentID uuid.NullUUID
	RowID       uuid.NullUUID
}

// CheckSiteComponentLock reads the lock state of one chrome slot. A missing row
// is not an error: it reports RowExists=false and IsLocked=false, because a row
// that does not exist yet cannot be locked and the caller is about to create
// it.
//
// The check is advisory — it exists to produce an accurate message and to skip
// the work before it is done. The race-free enforcement is
// pageComponentAgentWritableSQL in the writer's own WHERE clause.
func CheckSiteComponentLock(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	slot string, logger *zap.Logger) (*SiteComponentLockStatus, error) {

	var lockedAt, lockExpiresAt sql.NullTime
	var lockedBy, lockType sql.NullString
	var agentWritable, hasHTML bool
	var componentID, rowID uuid.NullUUID

	err := db.QueryRowContext(ctx, `
		SELECT locked_at, locked_by, lock_type, lock_expires_at,
		       `+pageComponentAgentWritableSQL("")+`,
		       component_id, id, COALESCE(rendered_html, '') <> ''
		FROM site_components
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slot).Scan(&lockedAt, &lockedBy, &lockType, &lockExpiresAt,
		&agentWritable, &componentID, &rowID, &hasHTML)

	if err == sql.ErrNoRows {
		return &SiteComponentLockStatus{}, nil
	}
	if err != nil {
		return nil, err
	}

	return &SiteComponentLockStatus{
		ComponentLockStatus: *classifyComponentLock(lockedAt, lockedBy, lockType, lockExpiresAt, agentWritable),
		RowExists:           true,
		HasHTML:             hasHTML,
		ComponentID:         componentID,
		RowID:               rowID,
	}, nil
}

// setSiteComponentHTML replaces a chrome slot's rendered HTML, refusing when the
// slot carries an active human lock. The lock predicate is in the WHERE, so the
// refusal is race-free; zero rows affected means locked (the callers all read
// the row moments earlier, so "row missing" is a concurrent delete, and refusing
// is right in that case too).
func setSiteComponentHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID, slot, html string) error {
	res, err := db.ExecContext(ctx, `
		UPDATE site_components SET rendered_html = $1, updated_at = NOW()
		WHERE site_id = $2 AND slot_name = $3 AND `+pageComponentAgentWritableSQL("")+`
	`, html, siteID, slot)
	if err != nil {
		return err
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		return errSiteComponentLocked
	}
	return nil
}

// appendSiteComponentHTML appends to a chrome slot's rendered HTML under the
// same gate. Same shape as setSiteComponentHTML; kept separate rather than
// parameterised because the SQL differs and a boolean argument at the call site
// would read as noise.
func appendSiteComponentHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID, slot, suffix string) error {
	res, err := db.ExecContext(ctx, `
		UPDATE site_components SET rendered_html = rendered_html || $1, updated_at = NOW()
		WHERE site_id = $2 AND slot_name = $3 AND `+pageComponentAgentWritableSQL("")+`
	`, suffix, siteID, slot)
	if err != nil {
		return err
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		return errSiteComponentLocked
	}
	return nil
}

// relinkSiteComponent points a chrome slot at a content component, creating the
// row if it is missing. Reports whether anything changed.
//
// Two guards on the DO UPDATE arm, doing different jobs: the pre-existing
// IS DISTINCT FROM makes the action idempotent (re-running links nothing), and
// the lock predicate closes the race between the caller's lock read and this
// write (bugs_open/069). Because the first guard means "0 rows" is the normal
// outcome, this function does NOT interpret 0 as "locked" — only the caller's
// read can tell those apart, which is why it does one.
//
// The INSERT arm needs no gate: a row that does not exist yet cannot be locked.
func relinkSiteComponent(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	slot string, componentID uuid.UUID) (bool, error) {

	res, err := db.ExecContext(ctx, `
		INSERT INTO site_components (site_id, slot_name, component_id, build_status, rendered_html)
		VALUES ($1, $2, $3, 'pending', NULL)
		ON CONFLICT (site_id, slot_name) DO UPDATE SET
			component_id = EXCLUDED.component_id,
			rendered_html = NULL,
			build_status = 'pending',
			updated_at = NOW()
		WHERE site_components.component_id IS DISTINCT FROM EXCLUDED.component_id
		  AND `+pageComponentAgentWritableSQL("site_components.")+`
	`, siteID, slot, componentID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// chromeLockItemKey is the dedup key for a blocked chrome change. Dedup is
// UNIQUE (site_id, item_key) for non-terminal statuses, so the key is
// site-scoped already and needs no site id. It names the SURFACE rather than
// "chrome" because the page-side key is lock_blocked_change:<pageName>:<slot>
// and a page called "chrome" would otherwise collide with the header slot.
func chromeLockItemKey(slot string) string {
	return fmt.Sprintf("lock_blocked_change:site_component:%s", slot)
}

// emitChromeLockBlockedChangeItem files ONE deduped work item when an active
// lock stopped an automated chrome write. Routing matches the page-side
// emitter and the chrome dead-control family: needs_human_review with NO
// handler_agent — whether the lock should yield is a human decision, and the
// 033 dashboard is the queue.
//
// header and footer are filed at HIGH severity (the emitChromeDeadControlItem
// precedent): chrome is site-wide, and a locked header is the bugs_open/049
// mechanism — new pages never appear in the navigation of any page. head is
// medium: it carries metadata, not navigation.
//
// A failure is logged, never returned: the signal must not break the write path
// that raised it.
func emitChromeLockBlockedChangeItem(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	slot string, status *SiteComponentLockStatus, blockedAction, sourceAction string,
	logger *zap.Logger) {

	severity := "medium"
	consequence := ""
	switch slot {
	case "header", "footer":
		severity = "high"
		consequence = " Chrome is site-wide: while this lock holds, nav changes (including newly " +
			"built pages appearing in the menu) will not reach ANY page of this site."
	}

	extra := map[string]interface{}{}
	var componentID *uuid.UUID
	if status != nil {
		if !status.HasHTML {
			extra["artefact_empty"] = true
			consequence += " This slot currently holds NO rendered HTML, so the lock is preserving " +
				"nothing and every page will ship without it until the lock is released."
		}
		if status.RowID.Valid {
			extra["site_component_id"] = status.RowID.UUID.String()
		}
		if status.ComponentID.Valid {
			id := status.ComponentID.UUID
			componentID = &id
		}
	}

	lockedBy, lockType := "", ""
	if status != nil {
		lockedBy, lockType = status.LockedBy, status.LockType
	}

	emitLockBlockedChange(ctx, db, lockBlockedChange{
		siteID:        siteID,
		componentID:   componentID,
		surface:       "site_component",
		slotName:      slot,
		lockedBy:      lockedBy,
		lockType:      lockType,
		blockedAction: blockedAction,
		sourceAction:  sourceAction,
		severity:      severity,
		itemKey:       chromeLockItemKey(slot),
		summary: fmt.Sprintf("Lock held: %s wanted to %s the locked %s chrome slot (locked by: %s)",
			sourceAction, blockedAction, slot, lockedBy),
		fixText: "A human lock on this site-chrome slot blocked an automated change; the locked " +
			"artefact was preserved (bugs_open/069)." + consequence +
			" Decide: keep the lock (the planned change stays unapplied — no action needed), or " +
			"unlock the slot via the admin dashboard and re-run the change. If you have just " +
			"edited this slot in the dashboard, this item is EXPECTED: the edit auto-locks the " +
			"slot, and this records that the follow-up re-render was refused rather than " +
			"discarding your edit.",
		extraSpec: extra,
	}, logger)
}
