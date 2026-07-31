// FILE: platform/orchestration/actions/save_sections_prune_floor.go
//
// A COMPLETENESS floor for save_page_sections' reconciliation delete —
// bugs_open/165 site A, applying the rule shipped by bugs_closed/135
// (prune_floor.go, register CTXA-025).
//
// THE DEFECT. SavePageSectionsAction deletes every agent-writable row for the
// page and re-inserts the section set it was handed. That is correct when the
// run saw the whole page and catastrophic when it did not, and nothing checked
// which. The section set comes from an LLM writer (or, on the fallback path,
// from regex-parsing assembled HTML); both can return a short-but-nonzero
// result with no error at all, and the outcome is not a broken row — it is
// ABSENCE, which reads as "there was never anything there". This table is the
// one that has actually lost customer-facing content: 016b §9 cases 1 and 2, an
// A* pathfinding game destroyed by exactly this DELETE+INSERT and recurring
// independently on a second site, plus case files 001, 037, 038 and
// bugs_closed/058.
//
// WHY THE EXISTING GUARDS DO NOT COVER IT. This file already refuses an owned
// page, a text regression, a lost interactive tool and a banned claim. None of
// them is a completeness check:
//
//   - pageComponentAgentWritableSQL is an AUTHORITY guard ("may I delete this
//     row?"). A writer that returned two sections instead of twelve passes it
//     perfectly — every row it deletes is one it was entitled to delete.
//   - the text and interactivity guards read `build_status = 'deployed'` only,
//     and 142 of the 426 pages that have components have NO deployed row
//     (measured 2026-07-31). On a third of the corpus they are blind.
//   - the text guard compares CHARACTERS. Two verbose sections replacing twelve
//     concise ones clears its quarter-of-existing bar while dropping ten.
//
// WHY THE REFUSAL FAILS THE SAVE, where 135's merely skips the prune. In 135 the
// delete and the write are separable, so a refusal keeps stale rows and the next
// healthy run prunes them — self-healing. Here they are one operation: refusing
// the delete but performing the insert would write the new sections alongside
// the old ones, which is bugs_open/156's duplicate-page_components defect. So the
// guard refuses the whole save, exactly as the three sibling regression guards
// above it already do, leaving the existing good page in place.
//
// THE COHORTS WERE MEASURED, NOT ASSUMED (bugs_open/165 says this in terms, and
// 135's were defensible only because they were chosen after reading the live
// distribution). Measured against production on 2026-07-31:
//
//   - REJECTED — one cohort per slot_name. 998 of 1,009 (page, slot_name) groups
//     hold exactly ONE row, so every per-slot cohort is 1 stored: dropping a
//     single section legitimately scores that cohort 0% and refuses the save. The
//     history shows 89 such legitimate shrinkages in 4.5 months, every one of
//     which this shape would have blocked.
//   - REJECTED — one cohort per component_id. 365 of 409 pages have as many
//     distinct components as rows, so it is the row count in a costume, not an
//     independent unit.
//   - KEPT — `sections`, the row unit: what this save will insert against the
//     rows the DELETE would remove. Across 2,620 consecutive overwrite pairs in
//     page_component_history (2026-03-16 → 2026-07-31), 89 shrank at all and
//     exactly 4 fell below 0.5 — a 0.15% refusal rate on real rebuilds.
//   - KEPT — `planned sections`, the PLAN unit, and the reason there are two
//     cohorts at all. pages.sections is written by other actions entirely
//     (load_page_sections_from_spec, apply_gap_plan, apply_adoption_plan,
//     create_blog_posts, deploy_tool, create_report_page, adopt_verbatim) and
//     never by this one, so it is a genuinely independent statement of what the
//     page is meant to contain. It is what breaks the RATCHET: once a truncating
//     writer has cut a page from twelve rows to two, the row cohort scores 2/2 =
//     100% for ever and the damage becomes the new baseline. The plan cohort
//     still reads 2 of 12.
//
// THE PLAN COHORT'S DENOMINATOR EXCLUDES LOCKED ROWS, and getting this wrong is
// the trap. A locked slot is not deleted and the incoming section that matches it
// is discarded, so a locked slot can never be part of what this save writes.
// Counting it in the denominator refuses healthy rebuilds of exactly the pages a
// human cared enough about to lock: idea.uk/index.html plans 6 sections and holds
// 4 locks, so a perfect rebuild writes 2 and would have scored 2/6 = 33%. With
// locks excluded the cohort trips on 0 of the 238 reachable pages; before the
// correction it tripped on that one. (The other two pages that trip are
// rebuild_policy='owned', which this action refuses ~370 lines earlier.)

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// savePageSectionsFloorKey is the step-config knob that resolves this floor. It
// deliberately shares the name 135 uses: the floor is one concept across every
// reconciliation delete, and a second spelling is how an operator who has learnt
// the first one gets surprised. Step config is per-step, so there is no
// collision between the two actions' values.
const savePageSectionsFloorKey = "prune_floor_ratio"

// pageSectionCompleteness holds one page's measurement. Kept as a struct rather
// than returned as bare cohorts so the caller can report the raw numbers on a
// PASSING save too — a bare "sections_saved: 2" is the alarm presented as
// output, which is candidate (3) of 135 restated for this table.
type pageSectionCompleteness struct {
	Projected   int // sections this save will actually INSERT
	WritableRow int // rows the DELETE would remove (its exact complement)
	LockedRows  int // rows the DELETE excludes, and the plan cohort must too
	Planned     int // pages.sections, less suppressed, less locked
	Cohorts     []pruneCohort
}

// measurePageSectionCompleteness takes both denominators in one round trip.
//
// The writable-row count uses pageComponentAgentWritableSQL — the SAME helper the
// DELETE's own predicate is built from, not a hand-rolled equivalent. Any other
// spelling is a second predicate free to drift from the one being guarded, and a
// guard measuring a different population from the one at risk is worse than none:
// it reports a ratio for rows nobody was going to delete.
func measurePageSectionCompleteness(ctx context.Context, db *sql.DB, pageID uuid.UUID, projected int) (pageSectionCompleteness, error) {
	m := pageSectionCompleteness{Projected: projected}
	if db == nil {
		return m, fmt.Errorf("no DB handle")
	}

	var planned, suppressed int
	err := db.QueryRowContext(ctx, `
		SELECT
			(SELECT count(*) FROM page_components pc
			  WHERE pc.page_id = p.id AND `+pageComponentAgentWritableSQL("pc.")+`),
			(SELECT count(*) FROM page_components pc
			  WHERE pc.page_id = p.id AND NOT `+pageComponentAgentWritableSQL("pc.")+`),
			CASE WHEN jsonb_typeof(p.sections) = 'array'
			     THEN jsonb_array_length(p.sections) ELSE 0 END,
			CASE WHEN jsonb_typeof(p.suppressed_sections) = 'array'
			     THEN jsonb_array_length(p.suppressed_sections) ELSE 0 END
		FROM pages p WHERE p.id = $1
	`, pageID).Scan(&m.WritableRow, &m.LockedRows, &planned, &suppressed)
	if err != nil {
		return m, fmt.Errorf("completeness measurement: %w", err)
	}

	// Planned, reduced to what this save could actually write: a suppressed
	// section is one the planner decided not to build, and a locked slot is one
	// this save may not touch. Only 4 pages carry suppressed_sections today
	// (max 2 each), but the mechanism is live and would otherwise inflate the
	// denominator on exactly the pages an operator has already curated.
	m.Planned = planned - suppressed - m.LockedRows
	if m.Planned < 0 {
		m.Planned = 0
	}

	m.Cohorts = []pruneCohort{
		{Label: "sections", Confirmed: projected, Stored: m.WritableRow},
		{Label: "planned sections", Confirmed: projected, Stored: m.Planned},
	}
	return m, nil
}

// projectedSectionInserts counts the sections this save will actually write, and
// it is the numerator BOTH cohorts divide by. len(sections) is the wrong number:
// a section whose slot is held by an active lock is discarded (the locked copy
// stands — bugs_closed/058), and an unresolvable empty stub is refused
// (bugs_open/039). Counting either as confirmed content would inflate the
// numerator and let a short run through, which is the one direction this guard
// must not be wrong in.
//
// The locked-row matching is SIMULATED on copies: matchLockedRow itself is
// read-only, but one locked row may swallow only one incoming section, and that
// bookkeeping lives in the caller's `consumed` flag. Mutating the real rows here
// would make every locked slot look consumed before the insert loop ran.
func projectedSectionInserts(sections []SectionData, lockedRows []*lockedPageRow) int {
	sim := make([]*lockedPageRow, 0, len(lockedRows))
	for _, lr := range lockedRows {
		clone := *lr
		sim = append(sim, &clone)
	}

	n := 0
	for _, s := range sections {
		if lr := matchLockedRow(sim, s.ComponentName); lr != nil {
			lr.consumed = true
			continue
		}
		if sectionIsUnresolvableStub(s) {
			continue
		}
		n++
	}
	return n
}

// enforcePageSectionFloor is the guard itself, called immediately before the
// DELETE. It returns the numbers for the action's result on a pass, and an error
// on a refusal — at which point the caller must return without deleting or
// inserting anything.
//
// FAILS CLOSED on an unmeasurable floor. The measurement is the only thing
// standing between a half-blind writer and a stripped page, so when it cannot be
// taken the destructive half must not proceed. That costs a failed build step,
// which is loud and recoverable; the alternative is silent absence, which is not.
func enforcePageSectionFloor(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID,
	pageName string, sections []SectionData, lockedRows []*lockedPageRow) (map[string]interface{}, error) {

	config := params.StepConfig.Config
	floor, fromConfig := pruneFloorFromConfig(config, savePageSectionsFloorKey, defaultPruneFloorRatio)
	projected := projectedSectionInserts(sections, lockedRows)

	m, err := measurePageSectionCompleteness(ctx, params.DB, pageID, projected)
	if err != nil {
		reason := fmt.Sprintf("save_page_sections: REFUSED for page %q — the completeness floor could not be measured (%v), so nothing was deleted or written", pageName, err)
		params.Logger.Error(reason)
		emitIncompleteSaveRefusalItem(ctx, params.DB, siteID, pageID, pageName, reason, params.Logger)
		return nil, fmt.Errorf("%s", reason)
	}

	verdict := evaluatePruneFloor(floor, m.Cohorts)
	reason := verdict.Reason("save_page_sections: overwrite",
		fmt.Sprintf("page %q", pageName), savePageSectionsFloorKey)

	if verdict.Clamped {
		params.Logger.Warn("save_page_sections: prune_floor_ratio out of range — clamped",
			zap.String("page_name", pageName),
			zap.Float64("configured", verdict.Asked),
			zap.Float64("applied", verdict.Floor))
	}

	detail := map[string]interface{}{
		"completeness_floor":       verdict.Floor,
		"completeness_from_config": fromConfig,
		"completeness_status":      "passed",
		"completeness_reason":      reason,
		"completeness_cohorts":     verdict.Detail(),
		"sections_projected":       projected,
		"writable_rows":            m.WritableRow,
		"locked_rows":              m.LockedRows,
		"planned_sections":         m.Planned,
	}

	switch {
	case !verdict.Allowed:
		params.Logger.Error(reason,
			zap.String("page_name", pageName),
			zap.Int("sections_projected", projected),
			zap.Int("writable_rows", m.WritableRow),
			zap.Int("planned_sections", m.Planned))
		emitIncompleteSaveRefusalItem(ctx, params.DB, siteID, pageID, pageName, reason, params.Logger)
		return nil, fmt.Errorf("%s", reason)
	case verdict.Disabled:
		detail["completeness_status"] = "floor_disabled"
		params.Logger.Warn(reason, zap.String("page_name", pageName))
	default:
		params.Logger.Info(reason, zap.String("page_name", pageName))
	}
	return detail, nil
}

// emitIncompleteSaveRefusalItem is the DURABLE surface for a refusal.
//
// site_work_items rather than 135's doc_notes, because this writer IS
// site-scoped and the columns site_work_items requires (site_id, pipeline) are
// all in hand — 135 fell back to doc_notes only because a repo-wide indexer has
// no site. Routing mirrors the lock-blocked family: status 'needs_human_review'
// with NO handler_agent, because whether a page genuinely shrank by half is a
// human judgement (accept it and lower the floor for that step, or find the
// writer that truncated), and no automated handler can make it. Persistence goes
// through insertWorkItem so it inherits the idx_swi_dedup-matched ON CONFLICT and
// the two-strike rule; the item_key is per page, so a page failing every rebuild
// collapses into one open item instead of one per attempt.
//
// Best-effort by design: a failure here is logged and swallowed. The refusal
// itself is already carried by the returned error, and losing the note must not
// change the decision that produced it.
func emitIncompleteSaveRefusalItem(ctx context.Context, db *sql.DB, siteID, pageID uuid.UUID,
	pageName, reason string, logger *zap.Logger) {

	if db == nil {
		return
	}

	spec := map[string]interface{}{
		"page_name": pageName,
		"source":    "save_page_sections",
		"reason":    reason,
		"fix": "A page rebuild produced too few sections to be allowed to replace what is stored, " +
			"so NOTHING was deleted and the existing page still stands (bugs_open/165). Decide: if the " +
			"page genuinely shrank, lower prune_floor_ratio on the save_page_sections step (0 disables " +
			"the floor); otherwise find why the writer returned a short section set — a truncated LLM " +
			"completion, a partial plan read, or an upstream step that failed without erroring.",
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("Page rebuild refused: %q returned too few sections to replace what is stored", pageName)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("save_page_sections: begin tx failed for refusal item",
			zap.String("page_name", pageName), zap.Error(err))
		return
	}
	pID := pageID
	if _, err := insertWorkItem(ctx, tx, workItem{
		siteID:    siteID,
		pageID:    &pID,
		source:    "save_page_sections",
		pipeline:  "build",
		itemType:  "save_refused_incomplete",
		severity:  "high",
		summary:   summary,
		spec:      string(specJSON),
		priority:  20,
		status:    "needs_human_review",
		createdBy: "save_page_sections",
		itemKey:   fmt.Sprintf("save_refused_incomplete:%s", pageName),
	}, logger); err != nil {
		_ = tx.Rollback()
		logger.Warn("save_page_sections: could not record the refusal as a work item",
			zap.String("page_name", pageName), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("save_page_sections: commit failed for refusal item",
			zap.String("page_name", pageName), zap.Error(err))
	}
}
