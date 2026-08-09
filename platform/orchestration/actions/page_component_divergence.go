// FILE: platform/orchestration/actions/page_component_divergence.go
//
// The LOUD half of the bugs_open/229 fix (owner ruling 2026-08-09: extend the
// 344 shape to pages): classify the page components a rebuild is about to
// destroy against their render stamps, and file a work item for each one whose
// bytes were put there outside the render path.
//
// The RECOVERY half is not here and not in any Go file — it is the DB trigger
// pair trg_page_component_artefact_archive_upd/_del (sql_for_agents/357),
// which archives outgoing rendered_html into page_component_history atomically
// with every differing overwrite AND every delete, from every writer (these
// actions, the colour-fix artefact rewriters, core-manager admin edits, raw
// psql). Page-side the DELETE arm is the one that matters: DELETE+INSERT is
// the dominant lifecycle (19,054 deletes vs 4,928 updates all-time, measured
// 2026-08-09), and it is how STY-025's interactive tools were destroyed.
//
// The stamp contract, page-side: rendered_html_digest = md5(html) is written
// in the SAME statement as the bytes by the render/save paths ONLY
// (save_page_sections, rebuild_blog_listing, section_editor,
// create_report_page). adopt_verbatim deliberately does NOT stamp — ported
// bytes are not reproducible from content_data, and a stamp that says
// "reproducible" beside bytes that are not is what silences this detector.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// pageComponentDivergence is one about-to-be-destroyed component whose stored
// bytes do not match their render stamp.
type pageComponentDivergence struct {
	ComponentID   uuid.UUID
	SiteID        uuid.UUID
	SlotName      string // "" when the row has no slot
	Position      int
	StampedDigest string
	CurrentDigest string
	Bytes         int
	OwnedPage     bool // pages.rebuild_policy = 'owned': artefact-primary, severity high
}

// classifyPageComponentArtefacts returns the hand-patched rows among those a
// page rebuild will destroy. The WHERE mirrors the destructive statements
// exactly: same page, same non-empty-artefact test, and the SAME
// agent-writable predicate — a locked row survives the rebuild, so counting it
// as destroyed would be a false alarm. Unstamped rows are deliberately not
// returned: they cannot be told apart (they converge as the fleet re-renders),
// and the 357 trigger archives them regardless.
func classifyPageComponentArtefacts(ctx context.Context, db *sql.DB,
	pageID uuid.UUID) ([]pageComponentDivergence, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT pc.id, p.site_id, COALESCE(pc.slot_name, ''), pc.position,
		       pc.rendered_html_digest, md5(pc.rendered_html), length(pc.rendered_html),
		       COALESCE(p.rebuild_policy, '') = 'owned'
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.page_id = $1
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
		  AND pc.rendered_html_digest IS NOT NULL
		  AND pc.rendered_html_digest <> md5(pc.rendered_html)
		  AND `+pageComponentAgentWritableSQL("pc.")+`
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []pageComponentDivergence
	for rows.Next() {
		var d pageComponentDivergence
		if err := rows.Scan(&d.ComponentID, &d.SiteID, &d.SlotName, &d.Position,
			&d.StampedDigest, &d.CurrentDigest, &d.Bytes, &d.OwnedPage); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// readBackPageDivergenceFromLedger recovers the verdicts when the pre-destroy
// classification failed: the 357 trigger computes the same judgement DB-side
// on every row it archives, so the fallback source of truth is the ledger
// entries the destruction just caused. No rows in the window means nothing
// divergent was destroyed. (The chrome sibling earned this fallback in council
// round 2 of trail cffbfec4; built here from the start.)
func readBackPageDivergenceFromLedger(ctx context.Context, db *sql.DB,
	pageID uuid.UUID) ([]pageComponentDivergence, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(component_id, '00000000-0000-0000-0000-000000000000'::uuid),
		       site_id, COALESCE(slot_name, ''), COALESCE(position, 0),
		       COALESCE(rendered_html_digest, ''), md5(rendered_html), length(rendered_html),
		       false
		FROM page_component_history
		WHERE page_id = $1
		  AND source = 'artefact_archive_trigger'
		  AND divergence = 'hand_patched'
		  AND created_at >= now() - interval '15 seconds'
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []pageComponentDivergence
	for rows.Next() {
		var d pageComponentDivergence
		if err := rows.Scan(&d.ComponentID, &d.SiteID, &d.SlotName, &d.Position,
			&d.StampedDigest, &d.CurrentDigest, &d.Bytes, &d.OwnedPage); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// pageDivergenceItemKey: page fragment + position locate the slot; the CURRENT
// (patched) digest makes distinct patches distinct findings while an identical
// repeat dedupes (idx_swi_dedup: UNIQUE on site_id + item_key over
// non-terminal rows). Same key design the chrome sibling settled in council.
func pageDivergenceItemKey(pageID uuid.UUID, position int, currentDigest string) string {
	if len(currentDigest) > 12 {
		currentDigest = currentDigest[:12]
	}
	page := pageID.String()
	if len(page) > 8 {
		page = page[:8]
	}
	return fmt.Sprintf("page_divergence_overwritten:page_component:%s:%d:%s", page, position, currentDigest)
}

// emitPageDivergenceItems files one deduped work item per destroyed
// hand-patched component. Routing mirrors the chrome emitter: needs_human_review,
// no handler — whether the lost content should be re-declared in content_data,
// carried by config, or locked is a human decision. Severity high only for
// rebuild_policy='owned' pages, whose real value lives in the artefact.
//
// Failures are logged, never returned: the signal must not break the write
// path that raised it.
func emitPageDivergenceItems(ctx context.Context, db *sql.DB, pageID uuid.UUID,
	pageName string, divergent []pageComponentDivergence, sourceAction string,
	logger *zap.Logger) {

	for _, d := range divergent {
		d := d
		logger.Warn("page components: artefact diverged from its render stamp — hand-patched bytes were overwritten and archived (bugs_open/229)",
			zap.String("page_id", pageID.String()),
			zap.String("page", pageName),
			zap.String("slot", d.SlotName),
			zap.Int("position", d.Position),
			zap.String("stamped_digest", d.StampedDigest),
			zap.String("current_digest", d.CurrentDigest),
			zap.Int("bytes", d.Bytes),
			zap.String("source", sourceAction),
		)

		severity := "medium"
		if d.OwnedPage {
			severity = "high"
		}

		spec := map[string]interface{}{
			"surface":        "page_component",
			"page_id":        pageID.String(),
			"page":           pageName,
			"slot_name":      d.SlotName,
			"position":       d.Position,
			"stamped_digest": d.StampedDigest,
			"current_digest": d.CurrentDigest,
			"artefact_bytes": d.Bytes,
			"source":         sourceAction,
			"fix": "A page rebuild destroyed this component's stored artefact, and the outgoing " +
				"bytes were NOT what the render path last wrote — a writer outside the render " +
				"path (hand patch, admin edit, colour-fix rewrite) had changed them " +
				"(bugs_open/229). Nothing is lost: the outgoing HTML is archived in " +
				"page_component_history (newest source='artefact_archive_trigger', " +
				"divergence='hand_patched' row for this page_id + position). Diff it against " +
				"the component's current rendered_html; if the patched content should exist, " +
				"re-declare it in content_data or lock the component (058) — do not paste it " +
				"back into rendered_html, which only re-arms this same loss.",
		}
		specJSON, _ := json.Marshal(spec)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			logger.Warn("emitPageDivergenceItems: begin tx failed",
				zap.String("page_id", pageID.String()), zap.Error(err))
			continue
		}
		pid := pageID
		var cid *uuid.UUID
		if d.ComponentID != uuid.Nil {
			c := d.ComponentID
			cid = &c
		}
		_, err = insertWorkItem(ctx, tx, workItem{
			siteID:   d.SiteID,
			source:   sourceAction,
			pipeline: "build",
			itemType: "page_divergence_overwritten",
			severity: severity,
			summary: fmt.Sprintf("Page rebuild overwrote hand-patched artefact on %q position %d (%d bytes archived to page_component_history)",
				pageName, d.Position, d.Bytes),
			spec:        string(specJSON),
			pageID:      &pid,
			componentID: cid,
			priority:    30,
			status:      "needs_human_review",
			createdBy:   sourceAction,
			itemKey:     pageDivergenceItemKey(pageID, d.Position, d.CurrentDigest),
			// Each divergence is a new genuine event (a different patch was
			// destroyed); without this the two-strike guard silently drops the
			// third same-key finding (chrome round 1, editquality seat).
			recurrenceExpected: true,
		}, logger)
		if err != nil {
			_ = tx.Rollback()
			logger.Warn("emitPageDivergenceItems: insert failed",
				zap.String("page_id", pageID.String()), zap.Int("position", d.Position), zap.Error(err))
			continue
		}
		if err := tx.Commit(); err != nil {
			logger.Warn("emitPageDivergenceItems: commit failed",
				zap.String("page_id", pageID.String()), zap.Error(err))
		}
	}
}
