// FILE: platform/orchestration/actions/site_component_divergence.go
//
// The LOUD half of the bugs_open/226 fix: classify a chrome slot's stored
// artefact against its render stamp just before a rebuild overwrites it, and
// file a work item when the bytes were put there outside the render path.
//
// The RECOVERY half is not here and not in any Go file — it is the DB trigger
// trg_site_component_archive (sql_for_agents/344), which archives the outgoing
// rendered_html into site_component_history atomically with every differing
// overwrite, from every writer (this action, relink's erase arm, the chrome-fix
// artefact patches, core-manager admin edits, raw psql). That placement is
// deliberate: the writer class that caused 226 was raw psql, which no Go guard
// can see. Because the trigger owns recovery, a race between this file's read
// and the store costs at most a mislabelled log line, never the artefact.
//
// The stamp contract: renderAndStoreSiteComponent sets
// rendered_html_digest = md5($1) in the SAME statement that stores the bytes,
// so digest = md5(rendered_html) thereafter means "these are exactly the bytes
// the render path wrote". The digest belongs to the render path ONLY — no
// other writer may set it, because a stamp written beside patched bytes is
// what silences this detector.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type siteComponentArtefactState string

const (
	// artefactEmpty: no row, or no stored HTML — nothing to lose.
	artefactEmpty siteComponentArtefactState = "empty"
	// artefactUnstamped: pre-344 row, or a row whose last write predates the
	// Go half of the fix. Cannot distinguish machine from hand; converges as
	// the fleet re-renders.
	artefactUnstamped siteComponentArtefactState = "unstamped"
	// artefactMachineMade: bytes match the stamp — reproducible from inputs,
	// overwriting destroys nothing that cannot be rebuilt.
	artefactMachineMade siteComponentArtefactState = "machine_made"
	// artefactHandPatched: bytes do NOT match the stamp — a writer outside the
	// render path (hand patch, admin edit, chrome-fix artefact patch) changed
	// the artefact since the render path last wrote it.
	artefactHandPatched siteComponentArtefactState = "hand_patched"
)

type siteComponentArtefactVerdict struct {
	State         siteComponentArtefactState
	StampedDigest string // "" when unstamped
	CurrentDigest string
	Bytes         int
}

// classifySiteComponentArtefact reads the slot's stored artefact and judges it
// against its render stamp. Advisory: callers proceed regardless of the answer
// (refusal is the 069 locks' job); the answer decides only how loudly.
func classifySiteComponentArtefact(ctx context.Context, db *sql.DB,
	siteID uuid.UUID, slot string) (*siteComponentArtefactVerdict, error) {

	var stamped sql.NullString
	var current string
	var byteCount int
	err := db.QueryRowContext(ctx, `
		SELECT rendered_html_digest, md5(rendered_html), length(rendered_html)
		FROM site_components
		WHERE site_id = $1 AND slot_name = $2
		  AND rendered_html IS NOT NULL AND rendered_html <> ''
	`, siteID, slot).Scan(&stamped, &current, &byteCount)
	if err == sql.ErrNoRows {
		return &siteComponentArtefactVerdict{State: artefactEmpty}, nil
	}
	if err != nil {
		return nil, err
	}

	v := &siteComponentArtefactVerdict{CurrentDigest: current, Bytes: byteCount}
	switch {
	case !stamped.Valid || stamped.String == "":
		v.State = artefactUnstamped
	case stamped.String == current:
		v.State = artefactMachineMade
		v.StampedDigest = stamped.String
	default:
		v.State = artefactHandPatched
		v.StampedDigest = stamped.String
	}
	return v, nil
}

func chromeDivergenceItemKey(slot string) string {
	return fmt.Sprintf("chrome_divergence_overwritten:site_component:%s", slot)
}

// emitChromeDivergenceItem files ONE deduped work item recording that a chrome
// rebuild overwrote hand-patched bytes. Routing mirrors the 069 lock emitter
// (needs_human_review, no handler_agent — whether the lost content should be
// re-declared through STY-050 carriage or a 069 lock is a human decision) and
// the severity split mirrors it too: chrome is site-wide, so header/footer are
// high.
//
// A failure is logged, never returned: the signal must not break the write
// path that raised it.
func emitChromeDivergenceItem(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	slot string, v *siteComponentArtefactVerdict, sourceAction string,
	logger *zap.Logger) {

	severity := "medium"
	if slot == "header" || slot == "footer" {
		severity = "high"
	}

	spec := map[string]interface{}{
		"surface":        "site_component",
		"slot_name":      slot,
		"stamped_digest": v.StampedDigest,
		"current_digest": v.CurrentDigest,
		"artefact_bytes": v.Bytes,
		"source":         sourceAction,
		"fix": "A chrome rebuild replaced this slot's stored artefact, and the outgoing bytes " +
			"were NOT what the render path last wrote — a writer outside the render path " +
			"(hand patch, admin edit, or chrome-fix artefact patch) had changed them, and " +
			"whatever it changed is now off the live pages (bugs_open/226). Nothing is lost: " +
			"the outgoing HTML is archived in site_component_history (newest " +
			"divergence='hand_patched' row for this site_id + slot_name). Diff it against the " +
			"slot's current rendered_html; if the patched content should exist, re-declare it " +
			"durably — STY-050 config carriage if it should survive and evolve with rebuilds, " +
			"a 069 lock if the slot must freeze — and do not paste it back into rendered_html, " +
			"which only re-arms this same loss.",
	}
	specJSON, _ := json.Marshal(spec)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("emitChromeDivergenceItem: begin tx failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:   siteID,
		source:   sourceAction,
		pipeline: "build",
		itemType: "chrome_divergence_overwritten",
		severity: severity,
		summary: fmt.Sprintf("Chrome rebuild overwrote hand-patched %s artefact (%d bytes archived to site_component_history)",
			slot, v.Bytes),
		spec:      string(specJSON),
		priority:  30,
		status:    "needs_human_review",
		createdBy: sourceAction,
		itemKey:   chromeDivergenceItemKey(slot),
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("emitChromeDivergenceItem: insert failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("emitChromeDivergenceItem: commit failed",
			zap.String("slot", slot), zap.Error(err))
	}
}
