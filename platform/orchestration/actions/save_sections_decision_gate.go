// FILE: platform/orchestration/actions/save_sections_decision_gate.go
//
// RFC_015 §5b, the SECOND write seam — the "rebuild door", gentle version
// (OWNER RULING 2026-08-10: "please guard the rebuild door now, gentler
// version", choosing option 1 of the three costed in RFC_015 §5b).
//
// WHY THIS EXISTS. The citation gate at apply_section_edit guards single-section
// edits. Whole-page rebuilds go through save_page_sections, which DELETEs the
// page's agent-writable rows and re-inserts the new composition — so a rebuild
// could overwrite decision-protected content with no citation and no signal.
// That is not hypothetical: on 2026-08-10 a section_data_resolved rerender
// restored a section the owner had had removed on idea.uk/index, and the
// mechanism caught it ~7 hours later by async discovery rather than preventing
// it. The council's bug_historian seat gated RFC_015 on exactly this, twice
// (corr cb547e0a rounds 1 and 2), citing 016b §9's "one call site of a shared
// judgement gets the rigorous fix; the sibling stays heuristic".
//
// GENTLE, AND WHAT THAT MEANS PRECISELY. A covered slot arriving WITHOUT a
// citation does not fail the rebuild and does not abort the save. The stored row
// STANDS (kept out of the DELETE, repositioned to follow the new composition),
// the freshly generated copy for that slot is DISCARDED, and the blocked
// overwrite is filed as a decision_blocked_change work item naming the
// decision(s). Everything else in the rebuild proceeds normally. So the failure
// mode of this gate is "one slot kept its content and someone was told", never
// "the page did not build".
//
// THIS IS THE LOCK GATE'S PATTERN, NOT A NEW ONE. bugs_open/058 already does
// exactly this for human-locked rows in save_page_sections: preload the rows
// ahead of the DELETE, exclude them from it, discard the incoming section that
// would have replaced one, reposition, and emit lock_blocked_change. Reusing a
// proven shape in the same function is the whole reason this option was
// recommended over a blocking gate.
//
// WHY THERE IS NO BYPASS FIELD, since the 2026-08-02 ruling on shared seams will
// be the reviewer's first question. That ruling applies when a seam's widest
// branch is licensed by a claim about callers ("callers must all be X") — make X
// a field with the unsafe default OFF. Here the widest branch (preserve the
// stored row) is licensed by DATA that is per-site, explicit and readable: a
// decision record with a ```covers fence. Nothing rests on an assumption about
// callers, so there is no claim to turn into a field. The UNSAFE capability —
// overwriting decision-protected content — is what requires the explicit field,
// and it already does: acknowledges_decision / supersedes_decision, absent by
// default. Adding a "skip the gate" flag would hand exactly that capability back
// with no name attached, which is the thing the ruling exists to prevent.
//
// A caller that cannot yet plumb a citation (page-content-writer today) is not
// broken by this: its rebuilds succeed, protected slots keep their content, and
// each blocked overwrite is visible as a work item. That is the deferral this
// gate replaces, made safe rather than postponed.
//
// FAILS OPEN, deliberately and consistently. A coverage-query error leaves the
// save exactly as it was before this gate existed, with a warning — the same
// posture as the lock check beside it and the edit-seam gate. A DB blip must not
// start discarding freshly generated sections.
//
// SELF-SCOPING: a site with no decision records, or a slot no ```covers fence
// names, is untouched. Measured 2026-08-10: four decision rows exist fleet-wide,
// all on idea.uk, three carrying a covers fence.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// decisionProtectedRow is a page_components row whose slot is covered by a
// decision this save did not cite, preserved through the rebuild.
//
// Deliberately a sibling of lockedPageRow rather than a reuse of it: the two
// carry different provenance (a lock names a locker; this names decisions) and
// emit different work items, and collapsing them would make one struct answer
// two questions — the asymmetry that 016b §9 keeps recording. The matching
// logic below is the same shape as matchLockedRow and must stay so.
type decisionProtectedRow struct {
	id        uuid.UUID
	slot      string
	position  int
	decisions []string // decision keys covering this slot, for the work item
	consumed  bool     // matched (and thereby blocked) an incoming section
}

// decisionCitationPaths are the collected-data paths a citation may arrive on,
// tried in order. A rebuild's envelope is not the edit seam's: apply_section_edit
// reads acknowledges_decision through its ActionInputSpec, while
// save_page_sections has no spec and reads everything out of CollectedData, so
// the citation has to be looked for where a work item's spec actually lands.
var decisionCitationPaths = []string{
	"input_data.spec.acknowledges_decision",
	"input_data.spec.supersedes_decision",
	"spec.acknowledges_decision",
	"spec.supersedes_decision",
	"input_data.acknowledges_decision",
	"input_data.supersedes_decision",
	"acknowledges_decision",
	"supersedes_decision",
}

// readDecisionCitation collects every citation value present in the envelope,
// comma-joined for CitationSatisfies (which splits on commas and trims).
//
// Both verbs are read the same way and pooled: acknowledges vs supersedes is a
// distinction the DECISION TRAIL cares about, not the gate — either one names
// the decision, which is the whole test ("you may change anything you can
// name"). config may override the search with decision_citation_field.
func readDecisionCitation(collected map[string]interface{}, config map[string]interface{}) string {
	var found []string
	if f, ok := config["decision_citation_field"].(string); ok && f != "" {
		if v := datahelpers.ExtractNestedFieldString(collected, f); v != "" {
			found = append(found, v)
		}
	}
	for _, p := range decisionCitationPaths {
		if v := datahelpers.ExtractNestedFieldString(collected, p); v != "" {
			found = append(found, v)
		}
	}
	return strings.Join(found, ",")
}

// loadDecisionProtectedRows returns the page's rows whose slot is covered by a
// decision the citation does not name, in position order.
//
// alreadyLocked rows are EXCLUDED: they are protected by the lock path, which
// preserves them and emits its own item, and a row preserved twice would be
// repositioned twice and filed twice.
//
// Best-effort by design: any error returns nil, so the save behaves exactly as
// it did before this gate existed.
func loadDecisionProtectedRows(
	ctx context.Context,
	db *sql.DB,
	siteID, pageID uuid.UUID,
	pageName, citation string,
	alreadyLocked []*lockedPageRow,
	logger *zap.Logger,
) []*decisionProtectedRow {
	if db == nil || pageName == "" {
		return nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(slot_name, ''), position
		FROM page_components
		WHERE page_id = $1 AND `+pageComponentAgentWritableSQL("")+`
		ORDER BY position ASC
	`, pageID)
	if err != nil {
		logger.Warn("SavePageSectionsAction: decision-protected preload failed — save proceeds ungated (RFC_015)",
			zap.Error(err))
		return nil
	}
	defer rows.Close()

	type candidate struct {
		id       uuid.UUID
		slot     string
		position int
	}
	var candidates []candidate
	lockedSlots := map[string]bool{}
	for _, lr := range alreadyLocked {
		lockedSlots[lr.slot] = true
	}
	for rows.Next() {
		var c candidate
		if scanErr := rows.Scan(&c.id, &c.slot, &c.position); scanErr != nil {
			logger.Warn("SavePageSectionsAction: decision-protected scan failed", zap.Error(scanErr))
			continue
		}
		if c.slot == "" || lockedSlots[c.slot] {
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Warn("SavePageSectionsAction: decision-protected row iteration failed — save proceeds ungated",
			zap.Error(err))
		return nil
	}
	if len(candidates) == 0 {
		return nil
	}

	// Coverage is asked per slot through the SAME helper the edit-seam gate
	// uses, rather than reimplemented here: a second copy of covers-fence
	// matching would drift from the first, and the two seams disagreeing about
	// what a decision covers is worse than either being wrong alone.
	protected := make([]*decisionProtectedRow, 0, len(candidates))
	coverageCache := map[string][]DecisionCoverage{}
	for _, c := range candidates {
		covered, ok := coverageCache[c.slot]
		if !ok {
			var covErr error
			covered, covErr = CheckDecisionCoverage(ctx, db, siteID, pageName, c.slot)
			if covErr != nil {
				logger.Warn("SavePageSectionsAction: decision coverage check failed — save proceeds ungated for this page (RFC_015)",
					zap.String("slot_name", c.slot), zap.Error(covErr))
				return nil
			}
			coverageCache[c.slot] = covered
		}
		if len(covered) == 0 {
			continue
		}
		if CitationSatisfies(citation, covered) {
			continue // named it — the change is allowed, and the trail records who
		}
		protected = append(protected, &decisionProtectedRow{
			id:        c.id,
			slot:      c.slot,
			position:  c.position,
			decisions: CoveredKeySlice(covered),
		})
	}
	if len(protected) == 0 {
		return nil
	}
	return protected
}

// matchDecisionProtectedRow finds the first unconsumed protected row whose slot
// matches the incoming section — exact first, then kebab-normalised. Same shape
// and same reasons as matchLockedRow (the 041 naming landmine: the library
// stores kebab-case but older rows and plans may carry snake_case or CamelCase
// variants of one slot), and each protected row matches at most one incoming
// section so a duplicated slot cannot have one decision swallow several.
func matchDecisionProtectedRow(protected []*decisionProtectedRow, sectionName string) *decisionProtectedRow {
	if sectionName == "" {
		return nil
	}
	for _, dr := range protected {
		if !dr.consumed && dr.slot == sectionName {
			return dr
		}
	}
	norm := NormalizeComponentFunction(sectionName)
	if norm == "" {
		return nil
	}
	for _, dr := range protected {
		if !dr.consumed && dr.slot != "" && NormalizeComponentFunction(dr.slot) == norm {
			return dr
		}
	}
	return nil
}

// decisionProtectedIDArrayLiteral renders the ids to exclude from the rebuild's
// DELETE as a PostgreSQL array literal, for use with an explicit ::uuid[] cast.
//
// A LITERAL, and NEVER a nil parameter, because the NULL semantics here are
// catastrophic rather than merely wrong. The DELETE reads
// `AND NOT (id = ANY($2::uuid[]))`; with a NULL parameter `id = ANY(NULL)` is
// NULL, `NOT NULL` is NULL, the WHERE excludes every row, and the save DELETEs
// NOTHING — every rebuild on every site silently stops clearing old sections,
// duplicating the whole page. An empty slice must therefore render as '{}' (a
// real empty array, over which ANY is false and NOT-ANY is true), which is what
// this returns and what the no-protected-rows case — i.e. almost every save
// fleet-wide — depends on.
//
// Follows the house idiom (toPGTextArrayLiteral, resolve_composition_helpers.go):
// this codebase deliberately does not use lib/pq's pq.Array. No escaping is
// needed or attempted: every element comes from uuid.UUID.String(), which cannot
// produce anything but hex and hyphens, so there is no injection surface — and if
// that ever stops being true, the ::uuid[] cast rejects it rather than executing it.
func decisionProtectedIDArrayLiteral(protected []*decisionProtectedRow) string {
	if len(protected) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(protected))
	for _, dr := range protected {
		parts = append(parts, dr.id.String())
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// emitDecisionBlockedChangeItem files the blocked overwrite.
//
// item_type decision_blocked_change is DISTINCT from decision_regression on
// purpose: a regression is "the outcome a decision pins is no longer true and
// someone must restore it"; this is "an overwrite was prevented, and the page is
// intact". Same status, opposite news — folding them together would put a
// prevented change and a live defect in one queue.
//
// KEYED BY PAGE AND SLOT, not by decision: one decision may cover many slots,
// and dedup on a coarser key would drop the second slot's block silently — the
// fleet's "idx_swi_dedup: KEY coarser than FINDING" mode, which this lane
// already got wrong once on decision_regression (council corr cb547e0a).
func emitDecisionBlockedChangeItem(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageID, componentID *uuid.UUID,
	pageName, slotName string,
	decisions []string,
	blockedAction, sourceAction string,
	logger *zap.Logger,
) {
	specMap := map[string]interface{}{
		"slot_name":      slotName,
		"decisions":      decisions,
		"blocked_action": blockedAction,
		"source":         sourceAction,
		"fix": "The rebuild kept this slot's stored content. To change it, re-dispatch " +
			"with acknowledges_decision (work within the decision) or supersedes_decision " +
			"(replace it) naming one of the decision keys above — read the decision row in " +
			"doc_notes first.",
	}
	if componentID != nil {
		specMap["component_id"] = componentID.String()
	}
	specJSON, _ := json.Marshal(specMap)

	summary := fmt.Sprintf("Rebuild would have overwritten %q on page %q, which decision(s) %s protect — stored content kept, no citation given",
		slotName, pageName, strings.Join(decisions, ", "))
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("emitDecisionBlockedChangeItem: begin tx failed",
			zap.String("slot", slotName), zap.Error(err))
		return
	}
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:      siteID,
		pageID:      pageID,
		componentID: componentID,
		source:      sourceAction,
		pipeline:    "build",
		itemType:    "decision_blocked_change",
		severity:    "medium",
		summary:     summary,
		spec:        string(specJSON),
		priority:    30,
		status:      "needs_human_review",
		createdBy:   sourceAction,
		itemKey:     fmt.Sprintf("decision_blocked_change:%s:%s", pageName, slotName),
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("emitDecisionBlockedChangeItem: insert failed",
			zap.String("slot", slotName), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("emitDecisionBlockedChangeItem: commit failed",
			zap.String("slot", slotName), zap.Error(err))
	}
}
