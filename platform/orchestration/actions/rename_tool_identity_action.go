// FILE: platform/orchestration/actions/rename_tool_identity_action.go
//
// RenameToolIdentityAction renames a tool's identity ATOMICALLY: the
// content_components.function, the page_components.slot_name that mirrors it,
// and the travelling docs (doc_plans/doc_notes subject_key) — then records the
// rename as a doc note under the NEW key. Guard rail 2 of the experience loop
// (RUNBOOK T2.2): a rename that moves only one of the three keys strands the
// acceptance sweep. The vonc arena is the proof: the page became tool-arena
// while function + docs stayed tool-arena-interface, so criteria loaded but
// the URL resolution (pages.name = function) missed the live page and the
// broken widget was invisible to the ladder.
//
// Deliberately does NOT rename pages rows — page identity belongs to the
// plan/reconcile side (CanonicalisePage). Instead the result REPORTS whether a
// page named new_function exists (page_name_match), so the caller can see the
// acceptance coupling is closed, plus counts of js_snippets / site_nav_items
// rows still referencing the old key (the "stale nav phantom" class) for
// follow-up — those are surfaced, not blindly rewritten.
//
// Inputs:
//   - old_function, new_function  (required; component_level='tool')
//   - target_site_id              (optional; scopes the page/nav reports)

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenameToolIdentityInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"old_function", "new_function"},
	Optional:   []string{"target_site_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("rename_tool_identity", RenameToolIdentityInputSpec)
}

// RenameToolIdentityAction handler.
func RenameToolIdentityAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "rename_tool_identity"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RenameToolIdentityInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	oldFn := inputs.Get("old_function")
	newFn := inputs.Get("new_function")
	siteID := inputs.Get("target_site_id")
	if oldFn == "" || newFn == "" || oldFn == newFn {
		return nil, fmt.Errorf("old_function and new_function must be distinct and non-empty (old=%q new=%q)", oldFn, newFn)
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. The component itself. Refuse when the target function is already
	// taken by another active tool — that is a merge, not a rename.
	var clash int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM content_components
		WHERE function = $1 AND component_level = 'tool' AND is_active
	`, newFn).Scan(&clash); err != nil {
		return nil, fmt.Errorf("target-function probe: %w", err)
	}
	if clash > 0 {
		return nil, fmt.Errorf("an active tool component already has function %q — this would be a merge, not a rename", newFn)
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE content_components SET function = $2, updated_at = NOW()
		WHERE function = $1 AND component_level = 'tool'
	`, oldFn, newFn)
	if err != nil {
		return nil, fmt.Errorf("rename component function: %w", err)
	}
	componentsRenamed, _ := res.RowsAffected()
	if componentsRenamed == 0 {
		return nil, fmt.Errorf("no tool component has function %q — nothing to rename", oldFn)
	}

	// 2. slot_name mirrors the function (component naming contract; chrome-fix
	// routing keys on it).
	res, err = tx.ExecContext(ctx, `
		UPDATE page_components pc SET slot_name = $2
		FROM content_components cc
		WHERE pc.component_id = cc.id AND cc.function = $2 AND pc.slot_name = $1
	`, oldFn, newFn)
	if err != nil {
		return nil, fmt.Errorf("rename slot_name: %w", err)
	}
	slotsRenamed, _ := res.RowsAffected()

	// 3. The travelling docs.
	plansMoved, notesMoved, err := datahelpers.RekeyTravellingDocs(ctx, tx, "tool", oldFn, newFn)
	if err != nil {
		return nil, err
	}

	// 4. Record the rename in the docs themselves, under the NEW key.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, source_agent, created_by)
		VALUES ('tool', $1,
		        '## Tool identity renamed: ' || $2 || ' -> ' || $1 || E'\n' ||
		        'Component function, page_components.slot_name and all travelling docs moved atomically (rename_tool_identity). ' ||
		        'Older notes under this subject predating this entry were written under the old key.',
		        '["rekey"]'::jsonb, 'rename_tool_identity', $3, 'rename_tool_identity')
	`, newFn, oldFn, params.ExecutionContext.Sender.AgentType)
	if err != nil {
		return nil, fmt.Errorf("record rename note: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// --- Post-commit reports (read-only; failures are logged, not fatal) ---
	pageNameMatch := false
	pageQuery := `SELECT COUNT(*) FROM pages WHERE name = $1`
	pageArgs := []interface{}{newFn}
	if siteID != "" {
		pageQuery += ` AND site_id = $2`
		pageArgs = append(pageArgs, siteID)
	}
	var pageCount int
	if err := params.DB.QueryRowContext(ctx, pageQuery, pageArgs...).Scan(&pageCount); err == nil {
		pageNameMatch = pageCount > 0
	}

	var jsRefs, navRefs int
	_ = params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM js_snippets WHERE applies_to::text ILIKE '%' || $1 || '%'
	`, oldFn).Scan(&jsRefs)
	navQuery := `SELECT COUNT(*) FROM site_nav_items WHERE (label ILIKE '%' || $1 || '%' OR url ILIKE '%' || $1 || '%')`
	navArgs := []interface{}{oldFn}
	if siteID != "" {
		navQuery += ` AND site_id = $2`
		navArgs = append(navArgs, siteID)
	}
	_ = params.DB.QueryRowContext(ctx, navQuery, navArgs...).Scan(&navRefs)

	logger.Info("RenameToolIdentityAction: Complete",
		zap.String("old_function", oldFn),
		zap.String("new_function", newFn),
		zap.Int64("components_renamed", componentsRenamed),
		zap.Int64("slots_renamed", slotsRenamed),
		zap.Int64("plans_moved", plansMoved),
		zap.Int64("notes_moved", notesMoved),
		zap.Bool("page_name_match", pageNameMatch),
		zap.Int("js_snippets_referencing_old", jsRefs),
		zap.Int("nav_items_referencing_old", navRefs),
	)

	return map[string]interface{}{
		"renamed":                     true,
		"old_function":                oldFn,
		"new_function":                newFn,
		"components_renamed":          componentsRenamed,
		"slots_renamed":               slotsRenamed,
		"plans_moved":                 plansMoved,
		"notes_moved":                 notesMoved,
		"page_name_match":             pageNameMatch,
		"js_snippets_referencing_old": jsRefs,
		"nav_items_referencing_old":   navRefs,
	}, nil
}
