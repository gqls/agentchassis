// FILE: platform/orchestration/actions/fix_hardcoded_colors_action.go
//
// FixHardcodedColorsAction replaces hardcoded hex color values in inline <style>
// blocks with CSS variable references (var(--color-primary) etc).
//
// Fixes two levels:
//   1. content_components.html_template — permanent fix for future renders
//   2. page_components.rendered_html    — immediate fix for current stored HTML
//
// The browser resolves var(--color-primary) at runtime from :root in styles.css,
// so this works in both template source and final rendered HTML.
//
// Used by: color-variable-fixer agent (dispatched by improvement loop)
//
// Config:
//   - fix_templates: bool (default true) — fix content_components.html_template
//   - fix_rendered:  bool (default true) — fix page_components.rendered_html
//
// Data inputs (via ActionInputSpec):
//   - site_id (required)

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var FixHardcodedColorsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fix_hardcoded_colors", FixHardcodedColorsInputSpec)
}

// ============================================================================
// ACTION: fix_hardcoded_colors
// ============================================================================

func FixHardcodedColorsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "fix_hardcoded_colors"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("FixHardcodedColorsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve site_id ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		FixHardcodedColorsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	config := params.StepConfig.Config
	fixTemplates := true
	fixRendered := true
	if v, ok := config["fix_templates"].(bool); ok {
		fixTemplates = v
	}
	if v, ok := config["fix_rendered"].(bool); ok {
		fixRendered = v
	}

	templatesFixed := 0
	renderedFixed := 0
	var details []map[string]interface{}

	// =========================================================================
	// 1. Fix content_components.html_template for dark-section components
	//    used by this site
	// =========================================================================
	if fixTemplates {
		tf, td := fixTemplateColors(ctx, params.DB, siteID, logger)
		templatesFixed = tf
		details = append(details, td...)
	}

	// =========================================================================
	// 2. Fix page_components.rendered_html for this site's pages
	// =========================================================================
	if fixRendered {
		rf, rd := fixRenderedColors(ctx, params.DB, siteID, logger)
		renderedFixed = rf
		details = append(details, rd...)
	}

	totalFixed := templatesFixed + renderedFixed

	logger.Info("FixHardcodedColorsAction: Complete",
		zap.Int("templates_fixed", templatesFixed),
		zap.Int("rendered_fixed", renderedFixed),
	)

	return map[string]interface{}{
		"site_id":         siteIDStr,
		"templates_fixed": templatesFixed,
		"rendered_fixed":  renderedFixed,
		"total_fixed":     totalFixed,
		"details":         details,
		"needs_rerender":  totalFixed > 0,
	}, nil
}

// fixTemplateColors fixes content_components.html_template for components
// assigned to this site (via site_components or page_components).
func fixTemplateColors(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (int, []map[string]interface{}) {
	var details []map[string]interface{}
	fixed := 0

	// Find all content_components used by this site that have dark sections
	// with hardcoded hex backgrounds in their templates
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT cc.id, cc.function, cc.html_template
		FROM content_components cc
		WHERE cc.is_active = true
		  AND cc.html_template ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
		  AND (
		    -- Used as site component (header/footer)
		    EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id = $1 AND sc.component_id = cc.id)
		    OR
		    -- Used as page section
		    EXISTS (
		      SELECT 1 FROM page_components pc
		      JOIN pages p ON pc.page_id = p.id
		      WHERE p.site_id = $1 AND pc.component_id = cc.id
		    )
		  )
	`, siteID)
	if err != nil {
		logger.Error("fixTemplateColors: query failed", zap.Error(err))
		return 0, details
	}
	defer rows.Close()

	for rows.Next() {
		var compID uuid.UUID
		var function, template string
		if err := rows.Scan(&compID, &function, &template); err != nil {
			continue
		}

		newTemplate := checks.ReplaceHardcodedColors(template)
		if newTemplate == template {
			continue
		}

		_, err := db.ExecContext(ctx, `
			UPDATE content_components SET html_template = $1 WHERE id = $2
		`, newTemplate, compID)
		if err != nil {
			logger.Error("fixTemplateColors: update failed",
				zap.String("function", function), zap.Error(err))
			continue
		}

		fixed++
		details = append(details, map[string]interface{}{
			"layer":    "template",
			"function": function,
			"id":       compID.String(),
			"changed":  true,
		})

		logger.Info("fixTemplateColors: Fixed component template",
			zap.String("function", function),
			zap.String("component_id", compID.String()),
		)
	}

	return fixed, details
}

// fixRenderedColors fixes page_components.rendered_html for this site's pages.
func fixRenderedColors(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (int, []map[string]interface{}) {
	var details []map[string]interface{}
	fixed := 0

	rows, err := db.QueryContext(ctx, `
		SELECT pc.id, pc.slot_name, pc.rendered_html, p.name as page_name
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
	`, siteID)
	if err != nil {
		logger.Error("fixRenderedColors: query failed", zap.Error(err))
		return 0, details
	}
	defer rows.Close()

	for rows.Next() {
		var pcID uuid.UUID
		var slotName, renderedHTML, pageName string
		if err := rows.Scan(&pcID, &slotName, &renderedHTML, &pageName); err != nil {
			continue
		}

		newHTML := checks.ReplaceHardcodedColors(renderedHTML)
		if newHTML == renderedHTML {
			continue
		}

		res, err := db.ExecContext(ctx, `
			UPDATE page_components SET rendered_html = $1 WHERE id = $2 AND `+pageComponentAgentWritableSQL(""),
			newHTML, pcID)
		if err != nil {
			logger.Error("fixRenderedColors: update failed",
				zap.String("slot", slotName), zap.String("page", pageName), zap.Error(err))
			continue
		}
		if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
			logger.Warn("fixRenderedColors: component is human-locked — colour fix skipped (bugs_open/058)",
				zap.String("slot", slotName), zap.String("page", pageName))
			continue
		}

		fixed++
		details = append(details, map[string]interface{}{
			"layer":     "rendered",
			"page":      pageName,
			"slot_name": slotName,
			"id":        pcID.String(),
			"changed":   true,
		})

		logger.Info("fixRenderedColors: Fixed rendered component",
			zap.String("page", pageName),
			zap.String("slot", slotName),
		)
	}

	return fixed, details
}

// The replacement transform (ReplaceHardcodedColors) lives in
// discovery_checks/check_hardcoded_section_colors.go — the completion-time
// verifier for this item type uses it as the remit predicate, and that package
// cannot import this one (actions imports discovery_checks). Keep it the single
// copy; do not fork a private one back here.
