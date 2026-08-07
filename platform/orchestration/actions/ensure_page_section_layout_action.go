// FILE: platform/orchestration/actions/ensure_page_section_layout_action.go
//
// EnsurePageSectionLayoutAction fills in a default section layout for a page
// that currently has NONE — and refuses (no-op) if it has any, from any
// source. Built for bugs_open/206: `directory-index` and `guides-index`
// existed as page rows (scaffolded by an earlier plan) with an empty
// `sections` array, so page-build-handler correctly no-op'd every time it
// was asked to build them — there was nothing to fill.
//
// This is deliberately NOT a re-planner. It only ever writes site_plan_sections
// for the ONE named page, and only when that page has zero sections from
// EVERY source load_page_sections_from_spec_action.go's own priority order
// checks (site_plan_sections for the current plan; pages.sections). It never
// touches another page's plan and never creates a new plan version — the
// guard this repo's "never re-plan this site" rule (bugs_closed/001) asks
// for, satisfied structurally rather than by caller discipline.
//
// The layout itself comes from defaultSectionsForPage (apply_gap_plan_action.go)
// — the same type/name-keyed default-layout chooser the gap-planner already
// uses, so there is one source of truth for "what's the sensible layout for
// a page like this," not two.
//
// Registration:
//   "ensure_page_section_layout": {
//       Handler:     EnsurePageSectionLayoutAction,
//       Category:    "site",
//       Description: "Fill in a default section layout for a page with none — refuses if it already has one",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var EnsurePageSectionLayoutInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id", "page_name"},
}

func init() {
	datahelpers.RegisterActionInputSpec("ensure_page_section_layout", EnsurePageSectionLayoutInputSpec)
}

func EnsurePageSectionLayoutAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "ensure_page_section_layout"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		EnsurePageSectionLayoutInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	pageName := inputs.Get("page_name")
	if pageName == "" {
		return nil, fmt.Errorf("page_name is required")
	}

	return ensurePageSectionLayout(ctx, params.DB, siteID, pageName, logger)
}

func ensurePageSectionLayout(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string, logger *zap.Logger) (interface{}, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin ensure_page_section_layout for %q: %w", pageName, err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var pageType string
	var pageSectionsJSON []byte
	if qerr := tx.QueryRowContext(ctx, `
		SELECT COALESCE(page_type, ''), COALESCE(sections, '[]'::jsonb)
		FROM pages WHERE site_id = $1 AND name = $2
		FOR UPDATE
	`, siteID, pageName).Scan(&pageType, &pageSectionsJSON); qerr != nil {
		if qerr == sql.ErrNoRows {
			return map[string]interface{}{
				"applied": false,
				"reason":  fmt.Sprintf("no page named %q on this site", pageName),
			}, nil
		}
		return nil, fmt.Errorf("load page %q: %w", pageName, qerr)
	}

	var pageSections []interface{}
	if len(pageSectionsJSON) > 0 {
		_ = json.Unmarshal(pageSectionsJSON, &pageSections)
	}
	if len(pageSections) > 0 {
		return map[string]interface{}{
			"applied": false,
			"reason":  "pages.sections already non-empty — refusing (this action never overwrites an existing layout)",
		}, nil
	}

	var currentPlanID uuid.UUID
	if qerr := tx.QueryRowContext(ctx, `
		SELECT id FROM site_plans WHERE site_id = $1 AND is_current = true
	`, siteID).Scan(&currentPlanID); qerr != nil {
		if qerr == sql.ErrNoRows {
			return map[string]interface{}{
				"applied": false,
				"reason":  "site has no current site_plans row",
			}, nil
		}
		return nil, fmt.Errorf("load current plan for site %s: %w", siteID, qerr)
	}

	var existingSectionCount int
	if qerr := tx.QueryRowContext(ctx, `
		SELECT count(*) FROM site_plan_sections WHERE plan_id = $1 AND page_name = $2
	`, currentPlanID, pageName).Scan(&existingSectionCount); qerr != nil {
		return nil, fmt.Errorf("check existing site_plan_sections for %q: %w", pageName, qerr)
	}
	if existingSectionCount > 0 {
		return map[string]interface{}{
			"applied": false,
			"reason":  fmt.Sprintf("current plan already carries %d section row(s) for this page — refusing", existingSectionCount),
		}, nil
	}

	sections := defaultSectionsForPage(pageName, pageType)

	if err := insertSitePlanSectionRows(ctx, tx, currentPlanID, pageName, sections); err != nil {
		return nil, err
	}

	sectionsJSON, jerr := json.Marshal(sections)
	if jerr != nil {
		return nil, fmt.Errorf("marshal sections for %q: %w", pageName, jerr)
	}
	if _, uerr := tx.ExecContext(ctx, `
		UPDATE pages SET sections = $3::jsonb, updated_at = NOW()
		WHERE site_id = $1 AND name = $2
	`, siteID, pageName, string(sectionsJSON)); uerr != nil {
		return nil, fmt.Errorf("sync pages.sections cache for %q: %w", pageName, uerr)
	}

	if cerr := tx.Commit(); cerr != nil {
		return nil, fmt.Errorf("commit ensure_page_section_layout for %q: %w", pageName, cerr)
	}
	committed = true

	logger.Info("ensure_page_section_layout: applied default layout",
		zap.String("page_name", pageName),
		zap.String("page_type", pageType),
		zap.Strings("sections", sections),
		zap.String("plan_id", currentPlanID.String()),
	)

	return map[string]interface{}{
		"applied":   true,
		"page_name": pageName,
		"page_type": pageType,
		"sections":  sections,
		"plan_id":   currentPlanID.String(),
	}, nil
}
