// FILE: platform/orchestration/actions/fix_component_template_action.go
//
// FixComponentTemplateAction applies targeted fixes. Routes on spec.fix_type.
//
// Grouped by what each fix_type actually touches — the distinction matters, and
// an earlier version of this header blurred it:
//
//	site_components.rendered_html   (header/footer/head ONLY, keyed by site_id+slot_name)
//	  - inject_nav_flex_css: adds display:flex CSS for stacked nav lists
//	  - responsive_fix:      adds mobile media queries
//	  - remove_element:      removes HTML elements matching a pattern (e.g. search icon)
//
//	page_components METADATA (never its rendered_html)
//	  - align_slot_name:               slot_name := the data-component value
//	  - repair_page_component_status:  build_status := 'deployed' for a live section
//	                                   carrying a status no reader honours
//
//	content_components.html_template
//	  - repair_template_slots: rewrites <no value>field</no> -> {{.field}}
//
// NOTE: remove_element operates on site_components ONLY. It cannot reach a
// content_components template or a page section — a 2026-07-09 handoff planned a
// section trim around it on the strength of this header and had to be re-planned.
//
// Per the source-of-truth principle (003b), rewriting rendered_html is acceptable
// for site_components because they are re-rendered from templates. page_components
// content changes go through the section-editor workflow (apply_section_edit);
// this action only ever touches their metadata columns.
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
	"database/sql"
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
	case "repair_template_slots":
		return fixRepairTemplateSlots(ctx, params, logger)
	case "inject_nav_flex_css", "spacing_fix", "spacing":
		// "spacing" is a raw category value from SQL-patched items
		return fixInjectNavFlexCSS(ctx, params, siteID, logger)
	case "remove_element":
		return fixRemoveElement(ctx, params, siteID, logger)
	case "align_slot_name":
		return fixAlignSlotName(ctx, params, logger)
	case "repair_page_component_status":
		return fixPageComponentStatus(ctx, params, logger)
	case "chrome_overflow_fix":
		// Targeted overflow repair from Tier-4 acceptance: slot + offending
		// selector come from the browser that measured it. Never guesses.
		return fixChromeOverflow(ctx, params, siteID, logger)
	case "responsive_fix", "responsive":
		// "responsive" is a raw category value from SQL-patched items.
		// NOTE: this path defaults to the HEADER slot and injects canned
		// header-nav CSS — it is not a general responsive fixer. See
		// chrome_overflow_fix above for the targeted version.
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
// chrome_overflow_fix — a TARGETED overflow repair, driven by Tier-4 acceptance.
//
// Deliberately NOT part of the legacy responsive_fix path, which defaults to the
// header slot and injects a canned header-nav CSS block regardless of what is
// actually broken. On 2026-07-14 that path was handed a FOOTER overflow, patched
// the HEADER, and returned fixed=true — the failure mode this one exists to
// avoid. Here, the slot and the offending selector both come from the browser
// that measured the defect, and the fix REFUSES to run without them rather than
// guess at a target.

// safeCSSSelector limits what can be interpolated into a stylesheet: a tag,
// id and/or class chain, e.g. "div.footer-legal", "#legal", ".a.b". Anything
// else (braces, angle brackets, at-rules, whitespace) is rejected — the value
// crosses a process boundary from a browser into an HTML <style> block.
var safeCSSSelector = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*(?:[.#][A-Za-z_-][A-Za-z0-9_-]*)*$|^[.#][A-Za-z_-][A-Za-z0-9_-]*(?:[.#][A-Za-z_-][A-Za-z0-9_-]*)*$`)

// buildOverflowCSS writes the media query that constrains the offender. Appended
// AFTER the slot's existing <style>, so it wins ties on document order without
// resorting to !important. flex-wrap is the fix for the common cause (a nowrap
// flex row); max-width is a harmless belt-and-braces for the non-flex case.
func buildOverflowCSS(selector, marker string) string {
	return "\n<style>\n" + marker + "\n" +
		"@media (max-width: 768px) {\n" +
		"    " + selector + " {\n" +
		"        flex-wrap: wrap;\n" +
		"        justify-content: center;\n" +
		"        max-width: 100%;\n" +
		"    }\n" +
		"    " + selector + " > * {\n" +
		"        max-width: 100%;\n" +
		"    }\n" +
		"}\n</style>\n"
}

func overflowMarker(selector string) string {
	return "/* Overflow fix (Tier-4 acceptance) — " + selector + " */"
}

func fixChromeOverflow(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) (interface{}, error) {
	slot := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.slot_name")
	selector := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.overflow_selector")

	// No silent defaults. A fix aimed at a guessed target is worse than no fix:
	// it reports success and closes the finding.
	switch slot {
	case "header", "footer", "head":
	case "":
		logger.Warn("chrome_overflow_fix: no slot_name in spec — refusing to guess a slot")
		return map[string]interface{}{
			"fixed": false, "fix_type": "chrome_overflow_fix",
			"reason": "spec.slot_name is required (header|footer|head) — refusing to guess; the legacy path's header default is exactly how a footer defect got 'fixed' in the header",
			"action": "needs_review",
		}, nil
	default:
		return map[string]interface{}{
			"fixed": false, "fix_type": "chrome_overflow_fix",
			"reason": fmt.Sprintf("slot_name %q is not a site_components slot (header|footer|head)", slot),
			"action": "needs_review",
		}, nil
	}

	if selector == "" || !safeCSSSelector.MatchString(selector) {
		logger.Warn("chrome_overflow_fix: missing or unsafe overflow_selector",
			zap.String("selector", selector))
		return map[string]interface{}{
			"fixed": false, "fix_type": "chrome_overflow_fix",
			"reason": fmt.Sprintf("spec.overflow_selector missing or not a simple tag/id/class selector: %q", selector),
			"action": "needs_review",
		}, nil
	}

	marker := overflowMarker(selector)
	overflowCSS := buildOverflowCSS(selector, marker)

	// The DURABLE source is the content_component template, NOT the site_component's
	// rendered_html: refresh_site_components regenerates rendered_html FROM the
	// template, so a patch to the artifact is wiped by the next refresh (observed
	// on vonc.com 2026-07-15 — a footer fix survived exactly until the refresh
	// that deployed it). Resolve the slot's backing component and patch there.
	var componentID sql.NullString
	var renderedHTML string
	err := params.DB.QueryRowContext(ctx, `
		SELECT component_id::text, COALESCE(rendered_html, '')
		FROM site_components WHERE site_id = $1 AND slot_name = $2
	`, siteID, slot).Scan(&componentID, &renderedHTML)
	if err == sql.ErrNoRows {
		return map[string]interface{}{
			"fixed": false, "fix_type": "chrome_overflow_fix",
			"slot_name": slot, "reason": "no site_component in that slot",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("chrome_overflow_fix: load %s slot: %w", slot, err)
	}

	if componentID.Valid && componentID.String != "" {
		var template string
		if err := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(html_template, '') FROM content_components WHERE id = $1
		`, componentID.String).Scan(&template); err != nil {
			return nil, fmt.Errorf("chrome_overflow_fix: load component %s: %w", componentID.String, err)
		}
		if strings.Contains(template, marker) {
			return map[string]interface{}{
				"fixed": false, "fix_type": "chrome_overflow_fix", "slot_name": slot,
				"overflow_selector": selector, "layer": "content_component",
				"component_id": componentID.String, "reason": "template already patched for this selector",
			}, nil
		}

		// Blast radius: this template is shared — the fix reaches every site that
		// uses it. That is correct for a genuine template CSS defect (a shared bug
		// gets a shared fix), and recorded so it is never a surprise. Other sites'
		// rendered_html self-heals on their next refresh; only THIS site is
		// rerendered by the item that triggered us.
		var sharedSites int
		_ = params.DB.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT site_id) FROM site_components WHERE component_id = $1
		`, componentID.String).Scan(&sharedSites)

		if _, err := params.DB.ExecContext(ctx, `
			UPDATE content_components SET html_template = html_template || $2, updated_at = now()
			WHERE id = $1
		`, componentID.String, overflowCSS); err != nil {
			return nil, fmt.Errorf("chrome_overflow_fix: patch component %s: %w", componentID.String, err)
		}

		logger.Info("chrome_overflow_fix: patched the DURABLE content_component template",
			zap.String("slot", slot), zap.String("selector", selector),
			zap.String("component_id", componentID.String), zap.Int("shared_sites", sharedSites))

		return map[string]interface{}{
			"fixed": true, "fix_type": "chrome_overflow_fix", "slot_name": slot,
			"overflow_selector": selector, "layer": "content_component",
			"component_id": componentID.String, "shared_sites": sharedSites,
			"detail": fmt.Sprintf("injected a mobile media query into the %s template (shared by %d site(s)) so %s wraps; a rerender deploys it and survives refresh", slot, sharedSites, selector),
		}, nil
	}

	// No backing component (inline/legacy slot): patch rendered_html as the only
	// available layer, and say plainly that it is transient — a refresh_site_components
	// would wipe it, so a durable fix needs the slot to be component-backed.
	if renderedHTML == "" {
		return map[string]interface{}{
			"fixed": false, "fix_type": "chrome_overflow_fix",
			"slot_name": slot, "reason": "no component_id and no rendered HTML in that slot",
		}, nil
	}
	if strings.Contains(renderedHTML, marker) {
		return map[string]interface{}{
			"fixed": false, "fix_type": "chrome_overflow_fix", "slot_name": slot,
			"overflow_selector": selector, "layer": "rendered_html",
			"reason": "already patched for this selector",
		}, nil
	}
	if _, err := params.DB.ExecContext(ctx, `
		UPDATE site_components SET rendered_html = rendered_html || $3, updated_at = now()
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slot, overflowCSS); err != nil {
		return nil, fmt.Errorf("chrome_overflow_fix: update %s rendered_html: %w", slot, err)
	}
	logger.Warn("chrome_overflow_fix: no backing component — patched rendered_html (TRANSIENT; a refresh will wipe it)",
		zap.String("slot", slot), zap.String("selector", selector))
	return map[string]interface{}{
		"fixed": true, "fix_type": "chrome_overflow_fix", "slot_name": slot,
		"overflow_selector": selector, "layer": "rendered_html", "durable": false,
		"detail": "patched rendered_html only (slot has no backing component); TRANSIENT — a refresh_site_components will wipe it",
	}, nil
}

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

// ---------------------------------------------------------------------------
// repair_page_component_status: flips a live page_component's build_status back
// to 'deployed'.
//
// Raised by the page_component_status_drift discovery check. build_status has no
// CHECK constraint, so a writer can invent a value; every discovery check filters
// `pc.build_status = 'deployed'`, so such a row ships live yet is invisible to the
// whole audit surface (apply_section_edit's 'approved', 2026-07-09).
//
// Metadata-only, like align_slot_name above — this does NOT touch rendered_html,
// so 003's source-of-truth principle is not in play.
//
// Two guards, both refusing rather than guessing:
//   - the parent page must itself be 'deployed' (the HTML really is live);
//   - the component must carry non-empty rendered_html (positive evidence, the
//     same rule pageHasComponents applies before marking a page deployed).
//
// It deliberately refuses to touch 'pending' / 'needs_rebuild' / 'removed': those
// are honest states whose repair is a rebuild, not a status flip. The check does
// not raise them, but a hand-written work item might.
// ---------------------------------------------------------------------------

// knownRebuildStatuses are honest non-deployed states: a rebuild resolves them,
// a status flip would only hide them. Mirrors knownComponentStatuses in
// discovery_checks/check_page_component_status_drift.go, minus 'deployed'.
var knownRebuildStatuses = map[string]bool{
	"pending":       true,
	"removed":       true,
	"needs_rebuild": true,
}

func fixPageComponentStatus(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	pcIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_component_id")
	if pcIDStr == "" {
		pcIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.page_component_id")
	}
	if pcIDStr == "" {
		return nil, fmt.Errorf("page_component_id is required for repair_page_component_status")
	}
	pcID, err := uuid.Parse(pcIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_component_id %q: %w", pcIDStr, err)
	}

	var observed, slotName, pageStatus string
	var hasHTML bool
	err = params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(pc.build_status, ''), COALESCE(pc.slot_name, ''),
		       COALESCE(p.build_status, ''),
		       (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.id = $1
	`, pcID).Scan(&observed, &slotName, &pageStatus, &hasHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to load page_component %s: %w", pcIDStr, err)
	}

	switch {
	case observed == "deployed":
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason": "already deployed",
		}, nil
	case knownRebuildStatuses[observed]:
		logger.Info("repair_page_component_status: refusing to flip an honest status",
			zap.String("slot", slotName), zap.String("build_status", observed))
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"observed_status": observed,
			"reason":          "status is a legitimate awaiting-rebuild state — needs a rebuild, not a status flip",
			"action":          "needs_review",
		}, nil
	case pageStatus != "deployed":
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason": fmt.Sprintf("parent page build_status is %q, not deployed", pageStatus),
			"action": "needs_review",
		}, nil
	case !hasHTML:
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason": "component has no rendered_html — cannot claim it is deployed",
			"action": "needs_review",
		}, nil
	}

	if _, err = params.DB.ExecContext(ctx, `
		UPDATE page_components SET build_status = 'deployed', updated_at = NOW() WHERE id = $1
	`, pcID); err != nil {
		return nil, fmt.Errorf("failed to repair build_status for %s: %w", pcIDStr, err)
	}

	logger.Info("repair_page_component_status: build_status repaired",
		zap.String("page_component_id", pcIDStr),
		zap.String("slot_name", slotName),
		zap.String("observed_status", observed))

	return map[string]interface{}{
		"fixed":             true,
		"fix_type":          "repair_page_component_status",
		"page_component_id": pcIDStr,
		"slot_name":         slotName,
		"observed_status":   observed,
		"new_status":        "deployed",
	}, nil
}

// ---------------------------------------------------------------------------
// repair_template_slots: rewrites <no value>field_name</no> → {{.field_name}}
//
// Targets content_components.html_template. The pattern arises when a template
// is rendered against an empty context and the output is stored as the source.
// StoreGeneratedComponentAction rejects this at creation time, but older
// components may carry the artifact.
//
// The fix uses a regexp replace: captures the field name between <no value>
// and </no>, rewrites as {{.field_name}}. A snapshot is written to
// component_versions before the UPDATE so the repair is reversible.
// ---------------------------------------------------------------------------

func fixRepairTemplateSlots(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	componentIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.component_id")
	if componentIDStr == "" {
		componentIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.component_id")
	}
	if componentIDStr == "" {
		return nil, fmt.Errorf("component_id is required for repair_template_slots (expected in spec.component_id)")
	}

	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid component_id %q: %w", componentIDStr, err)
	}

	// Read current template
	var function, currentTemplate string
	err = params.DB.QueryRowContext(ctx, `
		SELECT function, html_template
		FROM content_components
		WHERE id = $1
	`, componentID).Scan(&function, &currentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to load component %s: %w", componentIDStr, err)
	}

	if !strings.Contains(currentTemplate, "<no value>") {
		logger.Info("repair_template_slots: no artifacts found, nothing to do",
			zap.String("function", function))
		return map[string]interface{}{
			"fixed":    false,
			"function": function,
			"reason":   "no <no value> artifacts found",
		}, nil
	}

	artifactCount := strings.Count(currentTemplate, "<no value>")

	// Snapshot before modifying — write to component_versions so the repair
	// is reversible. Non-fatal if this fails; log and continue.
	var maxVersion int
	_ = params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version_number), 0) FROM component_versions WHERE component_id = $1
	`, componentID).Scan(&maxVersion)

	_, snapErr := params.DB.ExecContext(ctx, `
		INSERT INTO component_versions (component_id, version_number, html_template, change_source)
		VALUES ($1, $2, $3, 'repair_template_slots')
	`, componentID, maxVersion+1, currentTemplate)
	if snapErr != nil {
		logger.Warn("repair_template_slots: snapshot failed, continuing with repair",
			zap.String("function", function),
			zap.Error(snapErr))
	}

	// Before attempting repair, check whether the pattern is repairable.
	// repairNoValueSlots requires </no> closing tags to identify field names.
	// If none are present, the rendered output was stored without field name
	// fallbacks — the slots are unrecoverable and the component must be
	// regenerated via needs_component_regeneration, not repaired.
	closingTagCount := strings.Count(currentTemplate, "</no>")
	if closingTagCount == 0 {
		logger.Info("repair_template_slots: no </no> closing tags — needs regeneration",
			zap.String("function", function),
			zap.Int("artifact_count", artifactCount))
		return map[string]interface{}{
			"fixed":        false,
			"fix_type":     "repair_template_slots",
			"function":     function,
			"component_id": componentIDStr,
			"reason":       "template slots unrecoverable — field names absent, component needs regeneration",
			"action":       "needs_regeneration",
		}, nil
	}

	// Replace <no value>FIELDNAME</no> with {{.FIELDNAME}} using Go strings.
	// We iterate rather than using a single regex to avoid importing regexp
	// and to handle the pattern precisely.
	repairedTemplate := repairNoValueSlots(currentTemplate)

	// Verify the repair was clean — no residual artifacts
	residual := strings.Count(repairedTemplate, "<no value>")
	if residual > 0 {
		return nil, fmt.Errorf(
			"repair_template_slots: %d artifacts remain after repair for %q — mixed or unexpected slot format",
			residual, function)
	}

	_, err = params.DB.ExecContext(ctx, `
		UPDATE content_components
		SET html_template = $1, updated_at = now()
		WHERE id = $2
	`, repairedTemplate, componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to update template for %s: %w", function, err)
	}

	logger.Info("repair_template_slots: template repaired",
		zap.String("function", function),
		zap.String("component_id", componentIDStr),
		zap.Int("artifacts_fixed", artifactCount),
		zap.Int("snapshot_version", maxVersion+1))

	return map[string]interface{}{
		"fixed":            true,
		"fix_type":         "repair_template_slots",
		"function":         function,
		"component_id":     componentIDStr,
		"artifacts_fixed":  artifactCount,
		"snapshot_version": maxVersion + 1,
	}, nil
}

// repairNoValueSlots replaces all occurrences of <no value>FIELDNAME</no>
// with {{.FIELDNAME}} in a template string.
func repairNoValueSlots(template string) string {
	const prefix = "<no value>"
	const suffix = "</no>"

	var b strings.Builder
	remaining := template
	for {
		start := strings.Index(remaining, prefix)
		if start == -1 {
			b.WriteString(remaining)
			break
		}
		// Write everything before the artifact
		b.WriteString(remaining[:start])

		// Find the closing tag
		afterPrefix := remaining[start+len(prefix):]
		end := strings.Index(afterPrefix, suffix)
		if end == -1 {
			// Malformed: no closing tag — write the prefix literally and advance past it
			b.WriteString(prefix)
			remaining = afterPrefix
			continue
		}

		fieldName := afterPrefix[:end]
		b.WriteString("{{.")
		b.WriteString(fieldName)
		b.WriteString("}}")

		remaining = afterPrefix[end+len(suffix):]
	}
	return b.String()
}
