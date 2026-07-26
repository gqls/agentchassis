// FILE: platform/orchestration/actions/fix_nav_link_templates_action.go
//
// FixNavLinkTemplatesAction repairs broken navigation links in site header/footer
// templates. Finds content_component templates assigned to a site's header and
// footer slots, applies find/replace patterns (e.g. href="#{{.slug}}" → href="{{.url}}"),
// and writes the corrected templates back.
//
// This fixes the template source — render_site_components (with force_rerender)
// must run afterwards to regenerate the rendered HTML in site_components.
//
// Used by: nav-link-fixer agent
//
// Config (literals, not paths):
//   - patterns: []{"find": "...", "replace": "..."} — replacements to apply
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — which site's templates to fix

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var FixNavLinkTemplatesInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fix_nav_link_templates", FixNavLinkTemplatesInputSpec)
}

// ============================================================================
// ACTION: fix_nav_link_templates
// ============================================================================

func FixNavLinkTemplatesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "fix_nav_link_templates"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("FixNavLinkTemplatesAction: Starting")

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
		FixNavLinkTemplatesInputSpec,
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

	// --- Parse patterns from config ---
	//
	// The parser, the defaults and the transform below all live in
	// discovery_checks alongside check_broken_nav_links, which partitions its
	// findings by whether these patterns would change the template (bugs_open/077).
	// One copy, so the check cannot credit this handler with a replacement it does
	// not make. Do not inline a private copy back here.
	config := params.StepConfig.Config
	patterns := checks.ParseNavLinkPatterns(config)

	// If no patterns configured, use sensible defaults
	if len(patterns) == 0 {
		patterns = checks.DefaultNavLinkPatterns
	}

	logger.Info("FixNavLinkTemplatesAction: Loaded patterns",
		zap.String("site_id", siteIDStr),
		zap.Int("pattern_count", len(patterns)),
	)

	// --- Load templates for header and footer slots ---
	rows, err := params.DB.QueryContext(ctx, `
		SELECT sc.slot_name, sc.component_id, cc.html_template
		FROM site_components sc
		JOIN content_components cc ON sc.component_id = cc.id
		WHERE sc.site_id = $1 AND sc.slot_name IN ('header', 'footer')
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query site component templates: %w", err)
	}
	defer rows.Close()

	type slotTemplate struct {
		SlotName    string
		ComponentID uuid.UUID
		Template    string
	}

	var templates []slotTemplate
	for rows.Next() {
		var st slotTemplate
		if err := rows.Scan(&st.SlotName, &st.ComponentID, &st.Template); err != nil {
			logger.Warn("Failed to scan template row", zap.Error(err))
			continue
		}
		templates = append(templates, st)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template rows: %w", err)
	}

	if len(templates) == 0 {
		logger.Info("FixNavLinkTemplatesAction: No header/footer templates found for site")
		return map[string]interface{}{
			"updated":    0,
			"slot_names": []string{},
			"site_id":    siteIDStr,
			"reason":     "no header/footer component templates assigned to site",
		}, nil
	}

	// --- Apply patterns ---
	updated := 0
	var updatedSlots []string
	var details []map[string]interface{}

	for _, st := range templates {
		newTemplate, replacementsApplied := checks.ApplyNavLinkPatterns(st.Template, patterns)

		if newTemplate == st.Template {
			logger.Info("FixNavLinkTemplatesAction: No changes needed",
				zap.String("slot", st.SlotName),
				zap.String("component_id", st.ComponentID.String()),
			)
			details = append(details, map[string]interface{}{
				"slot_name":    st.SlotName,
				"component_id": st.ComponentID.String(),
				"changed":      false,
			})
			continue
		}

		// Write the corrected template back
		_, err := params.DB.ExecContext(ctx, `
			UPDATE content_components 
			SET html_template = $1, updated_at = now()
			WHERE id = $2
		`, newTemplate, st.ComponentID)
		if err != nil {
			logger.Error("FixNavLinkTemplatesAction: Failed to update template",
				zap.String("slot", st.SlotName),
				zap.String("component_id", st.ComponentID.String()),
				zap.Error(err),
			)
			details = append(details, map[string]interface{}{
				"slot_name":    st.SlotName,
				"component_id": st.ComponentID.String(),
				"changed":      false,
				"error":        err.Error(),
			})
			continue
		}

		updated++
		updatedSlots = append(updatedSlots, st.SlotName)

		logger.Info("FixNavLinkTemplatesAction: Template updated",
			zap.String("slot", st.SlotName),
			zap.String("component_id", st.ComponentID.String()),
			zap.Int("replacements", replacementsApplied),
		)

		details = append(details, map[string]interface{}{
			"slot_name":    st.SlotName,
			"component_id": st.ComponentID.String(),
			"changed":      true,
			"replacements": replacementsApplied,
		})
	}

	logger.Info("FixNavLinkTemplatesAction: Complete",
		zap.Int("updated", updated),
		zap.Strings("updated_slots", updatedSlots),
	)

	return map[string]interface{}{
		"updated":        updated,
		"slot_names":     updatedSlots,
		"site_id":        siteIDStr,
		"details":        details,
		"needs_rerender": updated > 0,
	}, nil
}
