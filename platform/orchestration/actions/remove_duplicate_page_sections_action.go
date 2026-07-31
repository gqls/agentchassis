// FILE: platform/orchestration/actions/remove_duplicate_page_sections_action.go
//
// The deterministic half of the content-duplication repair (owner ruling
// 2026-07-31, recorded in gauntlet_dead_cta/SUMMARY_2026-07-31): when one page
// carries the SAME section text more than once, keep the earliest row and remove
// the rest. No LLM, no judgement, no rewriting — the near-duplicate case that
// needs judgement is never dispatched here, it becomes a capability_gap
// (check_content_duplication.go, remit.go).
//
// WHY THIS EXISTS AS AN ACTION AT ALL
// -----------------------------------
// The live instance (bugs_open/156, vonc.com about page: 12 rows that were 6
// identical pairs, rendering the whole page twice for two days) was repaired by
// hand — "6 DELETEs + renumber in one transaction". That is a fix, not a
// mechanism: nothing detects the next one and nothing repairs it. This action is
// the repair half; check_content_duplication is the detect half.
//
// THE THREE THINGS THAT MAKE THIS SAFE TO AUTOMATE
// -----------------------------------------------
//  1. IT RE-DERIVES THE VICTIMS, it does not trust the work item. The spec
//     carries remove_component_ids from detection time, but a page can change
//     between detection and repair, so acting on a stale list could delete a row
//     that is no longer a duplicate. The spec list is used only to CROSS-CHECK:
//     anything it names that is no longer duplicated is reported and skipped.
//  2. IT COMPARES CONTENT, NEVER SLOT NAMES. Fleet-wide there are 17 duplicate
//     (page_id, slot_name) groups and 11 are legitimate — repeated slots with
//     DIFFERENT content on five sites. Keying on slot name would delete real
//     content. Content identity is the discriminator (bugs_open/156).
//  3. IT REFUSES TO DELETE EVERYTHING. Per identical group it keeps the lowest
//     position, always. If a computed plan would remove every section on the
//     page it aborts — that shape means the comparison is wrong, not that the
//     page is entirely redundant.
//
// Positions are renumbered 1..n afterwards, because gaps change rendering order
// downstream and the hand-fix did the same.
//
// Config (literals): none.
// Data inputs (via ActionInputSpec):
//   - page_id (required) — the page to deduplicate
//   - work_item_id (optional) — for the audit line in the result

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RemoveDuplicatePageSectionsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"page_id"},
	Optional:   []string{"work_item_id"},
	ConfigKeys: []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("remove_duplicate_page_sections", RemoveDuplicatePageSectionsInputSpec)
}

type dupSectionRow struct {
	ID       uuid.UUID
	Position int
	Slot     string
	Text     string
	Raw      string // canonical content_data::text — see datahelpers.SectionIdentityKey
}

// RemoveDuplicatePageSectionsAction removes content-identical duplicate sections
// from one page inside a single transaction, then renumbers positions.
func RemoveDuplicatePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "remove_duplicate_page_sections"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RemoveDuplicatePageSectionsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	pageID, err := uuid.Parse(inputs.Get("page_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, position, COALESCE(slot_name, ''), COALESCE(content_data::text, '{}')
		FROM page_components
		WHERE page_id = $1
		ORDER BY position
		FOR UPDATE
	`, pageID)
	if err != nil {
		return nil, fmt.Errorf("load sections: %w", err)
	}

	var sections []dupSectionRow
	for rows.Next() {
		var r dupSectionRow
		var raw string
		if err := rows.Scan(&r.ID, &r.Position, &r.Slot, &raw); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan section: %w", err)
		}
		r.Text = datahelpers.NormaliseSectionText(raw)
		r.Raw = raw
		sections = append(sections, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sections: %w", err)
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("page %s has no page_components rows — refusing to act on an empty page", pageID)
	}

	// --- re-derive the removal set from CURRENT content (see header note 1) ---
	byText := map[string][]dupSectionRow{}
	for _, s := range sections {
		if len(s.Text) < 80 {
			continue
		}
		// Same identity key as the detector, from datahelpers, so the two halves
		// cannot disagree about what "identical" means.
		k := datahelpers.SectionIdentityKey(s.Slot, s.Raw)
		byText[k] = append(byText[k], s)
	}

	var remove []dupSectionRow
	keptPerGroup := 0
	for _, grp := range byText {
		if len(grp) < 2 {
			continue
		}
		sort.Slice(grp, func(i, j int) bool { return grp[i].Position < grp[j].Position })
		keptPerGroup++
		remove = append(remove, grp[1:]...)
	}

	if len(remove) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit (no-op): %w", err)
		}
		logger.Info("remove_duplicate_page_sections: nothing to remove",
			zap.String("page_id", pageID.String()), zap.Int("sections", len(sections)))
		return map[string]interface{}{
			"page_id":          pageID.String(),
			"sections_before":  len(sections),
			"removed":          0,
			"duplicate_groups": 0,
			"changed":          false,
			"detail":           "no content-identical duplicate sections on this page",
		}, nil
	}

	// --- refusal guard (see header note 3) ---
	if len(remove) >= len(sections) {
		return nil, fmt.Errorf(
			"computed removal of %d of %d sections on page %s — refusing: that shape means the comparison is wrong, not that the page is redundant",
			len(remove), len(sections), pageID)
	}

	// --- cross-check against the work item's stale list, report both directions ---
	specNamed := specRemoveIDs(params.CollectedData)
	nowSet := map[string]bool{}
	for _, r := range remove {
		nowSet[r.ID.String()] = true
	}
	var staleFromSpec []string
	for _, id := range specNamed {
		if !nowSet[id] {
			staleFromSpec = append(staleFromSpec, id)
		}
	}

	ids := make([]uuid.UUID, 0, len(remove))
	removedDetail := make([]map[string]interface{}, 0, len(remove))
	for _, r := range remove {
		ids = append(ids, r.ID)
		removedDetail = append(removedDetail, map[string]interface{}{
			"component_id": r.ID.String(),
			"slot_name":    r.Slot,
			"position":     r.Position,
		})
	}

	res, err := tx.ExecContext(ctx,
		`DELETE FROM page_components WHERE page_id = $1 AND id = ANY($2::uuid[])`,
		pageID, uuidArray(ids))
	if err != nil {
		return nil, fmt.Errorf("delete duplicates: %w", err)
	}
	deleted, _ := res.RowsAffected()
	if int(deleted) != len(ids) {
		return nil, fmt.Errorf("expected to delete %d rows, deleted %d — aborting", len(ids), deleted)
	}

	// --- renumber 1..n so downstream ordering has no gaps ---
	if _, err := tx.ExecContext(ctx, `
		WITH ordered AS (
			SELECT id, row_number() OVER (ORDER BY position, id) AS rn
			FROM page_components WHERE page_id = $1
		)
		UPDATE page_components pc SET position = o.rn
		FROM ordered o WHERE pc.id = o.id AND pc.position <> o.rn
	`, pageID); err != nil {
		return nil, fmt.Errorf("renumber positions: %w", err)
	}

	var remaining int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM page_components WHERE page_id = $1`, pageID).Scan(&remaining); err != nil {
		return nil, fmt.Errorf("recount: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("remove_duplicate_page_sections: complete",
		zap.String("page_id", pageID.String()),
		zap.Int("sections_before", len(sections)),
		zap.Int64("removed", deleted),
		zap.Int("sections_after", remaining),
		zap.Int("duplicate_groups", keptPerGroup),
		zap.Int("stale_ids_in_spec", len(staleFromSpec)))

	return map[string]interface{}{
		"page_id":            pageID.String(),
		"sections_before":    len(sections),
		"sections_after":     remaining,
		"removed":            int(deleted),
		"duplicate_groups":   keptPerGroup,
		"removed_components": removedDetail,
		"stale_spec_ids":     staleFromSpec,
		"changed":            true,
		"needs_rerender":     true,
	}, nil
}

// specRemoveIDs pulls the detection-time removal list out of the work item spec,
// if one travelled with the request. Used ONLY to report drift, never to decide.
func specRemoveIDs(collected map[string]interface{}) []string {
	raw, ok := collected["work_item"]
	if !ok {
		raw, ok = collected["item"]
	}
	if !ok {
		return nil
	}
	blob, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var wi struct {
		Spec struct {
			RemoveComponentIDs []string `json:"remove_component_ids"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(blob, &wi); err != nil {
		return nil
	}
	return wi.Spec.RemoveComponentIDs
}

func uuidArray(ids []uuid.UUID) interface{} {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}
	return "{" + strings.Join(parts, ",") + "}"
}

var _ = sql.ErrNoRows // keep database/sql imported for future row-level guards
