// FILE: platform/orchestration/actions/fix_component_template_action.go
//
// FixComponentTemplateAction applies targeted fixes to site_components and
// page_components rendered HTML. Routes on spec.fix_type:
//
//   - inject_nav_flex_css: adds display:flex CSS for stacked nav lists
//   - remove_element: removes HTML elements matching a pattern (e.g. search icon)
//   - align_slot_name: updates page_components.slot_name to match data-component
//
// This action modifies rendered_html directly. Per the source-of-truth principle
// (003b), this is acceptable for site_components (header/footer/head) because
// they are re-rendered from templates. For page_components, we update slot_name
// (metadata) but NOT rendered_html content — content changes go through the
// section-editor workflow.
//
// Registration:
//   "fix_component_template": {
//       Handler:     FixComponentTemplateAction,
//       Category:    "site",
//       Description: "Apply targeted fixes to component HTML/CSS",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FixComponentTemplateInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"fix_type", "slot_name", "page_component_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fix_component_template", FixComponentTemplateInputSpec)
}

// inferFixTypeFromCategory maps audit category names to actionable fix_type values.
func inferFixTypeFromCategory(category string) string {
	category = strings.ToLower(strings.TrimSpace(category))
	switch category {
	case "spacing":
		return "inject_nav_flex_css"
	case "responsive":
		return "responsive_fix"
	case "cta":
		return "cta_improvement"
	case "nav_restructure":
		return "nav_restructure"
	default:
		return ""
	}
}

// inferFixTypeFromItemType maps work item item_type values to fix_type as a last resort.
func inferFixTypeFromItemType(itemType string) string {
	switch itemType {
	case "spacing_fix":
		return "inject_nav_flex_css"
	case "responsive_fix":
		return "responsive_fix"
	case "cta_improvement":
		return "cta_improvement"
	case "nav_restructure":
		return "nav_restructure"
	case "header_footer_fix":
		return "inject_nav_flex_css"
	default:
		return ""
	}
}

func FixComponentTemplateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fix_component_template"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// Read fix_type from config (literal, not a path)
	config := params.StepConfig.Config
	fixType, _ := config["fix_type"].(string)

	// Also check spec in input_data (when dispatched via work item)
	if fixType == "" {
		fixType = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.fix_type")
	}
	if fixType == "" {
		fixType = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.fix_type")
	}

	// Fallback: derive from category
	if fixType == "" {
		category := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.category")
		if category != "" {
			fixType = inferFixTypeFromCategory(category)
			if fixType != "" {
				logger.Info("Derived fix_type from category",
					zap.String("category", category),
					zap.String("fix_type", fixType))
			}
		}
	}

	// Fallback: derive from item_type
	if fixType == "" {
		itemType := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.item_type")
		if itemType != "" {
			fixType = inferFixTypeFromItemType(itemType)
			if fixType != "" {
				logger.Info("Derived fix_type from item_type",
					zap.String("item_type", itemType),
					zap.String("fix_type", fixType))
			}
		}
	}

	if fixType == "" {
		return nil, fmt.Errorf("fix_type is required (config, input_data.spec.fix_type, spec.category, or input_data.item_type)")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		config,
		FixComponentTemplateInputSpec,
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

	logger.Info("FixComponentTemplateAction: Starting",
		zap.String("fix_type", fixType),
		zap.String("site_id", siteIDStr),
	)

	switch fixType {
	case "inject_nav_flex_css", "spacing_fix", "spacing":
		// "spacing" is a raw category value from SQL-patched items
		return fixInjectNavFlexCSS(ctx, params, siteID, logger)
	case "remove_element":
		return fixRemoveElement(ctx, params, siteID, logger)
	case "align_slot_name":
		return fixAlignSlotName(ctx, params, logger)
	case "responsive_fix", "responsive":
		// "responsive" is a raw category value from SQL-patched items
		return fixInjectResponsiveCSS(ctx, params, siteID, logger)
	case "cta_improvement", "cta", "nav_restructure":
		// "cta" is a raw category value from SQL-patched items.
		// These need LLM-driven content changes, not programmatic HTML edits.
		// Return skipped so the item doesn't fail repeatedly.
		logger.Info("Fix type requires LLM involvement, marking for review",
			zap.String("fix_type", fixType))
		return map[string]interface{}{
			"fixed":    false,
			"fix_type": fixType,
			"reason":   "fix_type requires LLM-driven changes, not programmatic HTML edits",
			"action":   "needs_review",
		}, nil
	default:
		logger.Warn("Unrecognised fix_type, marking for review",
			zap.String("fix_type", fixType))
		return map[string]interface{}{
			"fixed":    false,
			"fix_type": fixType,
			"reason":   fmt.Sprintf("unrecognised fix_type: %s", fixType),
			"action":   "needs_review",
		}, nil
	}
}

// ---------------------------------------------------------------------------
// inject_nav_flex_css: adds CSS for horizontal nav layout
// ---------------------------------------------------------------------------

func fixInjectNavFlexCSS(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) (interface{}, error) {
	slotName := "header"
	if s := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.slot_name"); s != "" {
		slotName = s
	}

	var html string
	err := params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(rendered_html, '') FROM site_components
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slotName).Scan(&html)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s component: %w", slotName, err)
	}

	if html == "" {
		return map[string]interface{}{"fixed": false, "reason": "no rendered HTML"}, nil
	}

	// Check if already has flex
	if strings.Contains(html, "display: flex") || strings.Contains(html, "display:flex") {
		return map[string]interface{}{"fixed": false, "reason": "already has flex CSS"}, nil
	}

	// Detect the nav list class name from the HTML
	navCSS := `
<style>
/* Nav flex fix — injected by component-template-fixer */
.site-header__categories,
.main-nav ul,
.site-header__menu {
    display: flex;
    gap: 1.5rem;
    list-style: none;
    align-items: center;
    flex-wrap: wrap;
    margin: 0;
    padding: 0;
}
.site-header__category,
.main-nav a {
    text-decoration: none;
    font-size: 0.95rem;
    white-space: nowrap;
}
</style>`

	// Inject before closing </header> tag
	if strings.Contains(html, "</header>") {
		html = strings.Replace(html, "</header>", navCSS+"\n</header>", 1)
	} else {
		html = html + navCSS
	}

	_, err = params.DB.ExecContext(ctx, `
		UPDATE site_components SET rendered_html = $1, updated_at = NOW()
		WHERE site_id = $2 AND slot_name = $3
	`, html, siteID, slotName)
	if err != nil {
		return nil, fmt.Errorf("failed to update %s: %w", slotName, err)
	}

	logger.Info("Injected nav flex CSS",
		zap.String("slot", slotName))

	return map[string]interface{}{
		"fixed":     true,
		"fix_type":  "inject_nav_flex_css",
		"slot_name": slotName,
	}, nil
}

// ---------------------------------------------------------------------------
// responsive_fix: adds responsive CSS media queries for mobile layout
// ---------------------------------------------------------------------------

func fixInjectResponsiveCSS(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) (interface{}, error) {
	slotName := "header"
	if s := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.slot_name"); s != "" {
		slotName = s
	}

	var html string
	err := params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(rendered_html, '') FROM site_components
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slotName).Scan(&html)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s component: %w", slotName, err)
	}

	if html == "" {
		return map[string]interface{}{"fixed": false, "reason": "no rendered HTML"}, nil
	}

	// Check if already has responsive media query for this slot
	if strings.Contains(strings.ToLower(html), "responsive fix") {
		return map[string]interface{}{"fixed": false, "reason": "already has responsive CSS"}, nil
	}

	responsiveCSS := `
<style>
/* Responsive fix — injected by component-template-fixer */
@media (max-width: 768px) {
    .site-header__categories,
    .main-nav ul,
    .site-header__menu {
        flex-direction: column;
        gap: 0.5rem;
        text-align: center;
    }
    .site-header {
        flex-direction: column;
        padding: 1rem;
    }
    .site-header__actions {
        justify-content: center;
        margin-top: 0.5rem;
    }
}
@media (max-width: 480px) {
    .site-header__brand {
        font-size: 1.2rem;
    }
}
</style>`

	// Inject before closing tag for the slot
	closingTag := fmt.Sprintf("</%s>", slotName)
	if strings.Contains(html, closingTag) {
		html = strings.Replace(html, closingTag, responsiveCSS+"\n"+closingTag, 1)
	} else {
		html = html + responsiveCSS
	}

	_, err = params.DB.ExecContext(ctx, `
		UPDATE site_components SET rendered_html = $1, updated_at = NOW()
		WHERE site_id = $2 AND slot_name = $3
	`, html, siteID, slotName)
	if err != nil {
		return nil, fmt.Errorf("failed to update %s: %w", slotName, err)
	}

	logger.Info("Injected responsive CSS",
		zap.String("slot", slotName))

	return map[string]interface{}{
		"fixed":     true,
		"fix_type":  "responsive_fix",
		"slot_name": slotName,
	}, nil
}

// ---------------------------------------------------------------------------
// remove_element: removes HTML elements matching a pattern
// ---------------------------------------------------------------------------

func fixRemoveElement(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) (interface{}, error) {
	slotName := "header"
	if s := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.slot_name"); s != "" {
		slotName = s
	}

	pattern := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.pattern")
	if pattern == "" {
		pattern, _ = params.StepConfig.Config["pattern"].(string)
	}
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required for remove_element")
	}

	var html string
	err := params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(rendered_html, '') FROM site_components
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slotName).Scan(&html)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", slotName, err)
	}

	if !strings.Contains(html, pattern) {
		return map[string]interface{}{"fixed": false, "reason": "pattern not found"}, nil
	}

	originalLen := len(html)

	// Remove the element containing the pattern
	switch pattern {
	case "search-toggle":
		// Remove the search toggle button
		re := regexp.MustCompile(`(?s)<button[^>]*search-toggle[^>]*>.*?</button>`)
		html = re.ReplaceAllString(html, "")
		// Remove the empty actions div if it now contains only whitespace
		re = regexp.MustCompile(`(?s)<div[^>]*site-header__actions[^>]*>\s*</div>`)
		html = re.ReplaceAllString(html, "")
	default:
		// Generic: remove elements with the pattern as a class or attribute
		re := regexp.MustCompile(fmt.Sprintf(`(?s)<[^>]*%s[^>]*>.*?</[^>]+>`, regexp.QuoteMeta(pattern)))
		html = re.ReplaceAllString(html, "")
	}

	if len(html) == originalLen {
		return map[string]interface{}{"fixed": false, "reason": "regex didn't match"}, nil
	}

	_, err = params.DB.ExecContext(ctx, `
		UPDATE site_components SET rendered_html = $1, updated_at = NOW()
		WHERE site_id = $2 AND slot_name = $3
	`, html, siteID, slotName)
	if err != nil {
		return nil, fmt.Errorf("failed to update %s: %w", slotName, err)
	}

	logger.Info("Removed element",
		zap.String("pattern", pattern),
		zap.String("slot", slotName),
		zap.Int("bytes_removed", originalLen-len(html)))

	return map[string]interface{}{
		"fixed":         true,
		"fix_type":      "remove_element",
		"pattern":       pattern,
		"bytes_removed": originalLen - len(html),
	}, nil
}

// ---------------------------------------------------------------------------
// align_slot_name: updates page_components.slot_name to match data-component
// ---------------------------------------------------------------------------

func fixAlignSlotName(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	pcIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_component_id")
	if pcIDStr == "" {
		pcIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.page_component_id")
	}
	if pcIDStr == "" {
		return nil, fmt.Errorf("page_component_id is required for align_slot_name")
	}

	pcID, err := uuid.Parse(pcIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_component_id: %w", err)
	}

	dataComponent := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.data_component")
	if dataComponent == "" {
		return nil, fmt.Errorf("data_component is required for align_slot_name")
	}

	result, err := params.DB.ExecContext(ctx, `
		UPDATE page_components SET slot_name = $1, updated_at = NOW()
		WHERE id = $2
	`, dataComponent, pcID)
	if err != nil {
		return nil, fmt.Errorf("failed to update slot_name: %w", err)
	}

	rows, _ := result.RowsAffected()
	logger.Info("Aligned slot_name",
		zap.String("page_component_id", pcIDStr),
		zap.String("new_slot_name", dataComponent),
		zap.Int64("rows_affected", rows))

	return map[string]interface{}{
		"fixed":             true,
		"fix_type":          "align_slot_name",
		"page_component_id": pcIDStr,
		"new_slot_name":     dataComponent,
	}, nil
}
