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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// chromeFixLockSkip is the shared lock gate for every fix_type in this file
// that rewrites site_components.rendered_html (bugs_open/069). It returns a
// skip-RESULT when the slot carries an active human lock, or nil to proceed.
//
// The result speaks this file's existing vocabulary: fixed:false plus
// action:"needs_review", which is what stops the dispatch loop recording the
// work item as done. Without it the handler reports success, discovery
// re-detects the same defect, and the two-strike rule parks the item
// 'unresolved' two cycles later — a silent skip traded for a mislabelled one.
//
// A check failure is non-fatal: the writes below carry the lock predicate
// themselves, so enforcement never depends on this read.
func chromeFixLockSkip(ctx context.Context, params ActionParams, siteID uuid.UUID,
	slot, fixType string, logger *zap.Logger) map[string]interface{} {

	lock, err := CheckSiteComponentLock(ctx, params.DB, siteID, slot, logger)
	if err != nil {
		logger.Warn("fix_component_template: chrome lock check failed — relying on the guarded update",
			zap.String("slot", slot), zap.Error(err))
		return nil
	}
	if !lock.IsLocked {
		return nil
	}

	logger.Warn("fix_component_template: refusing to patch human-locked chrome slot (bugs_open/069)",
		zap.String("slot", slot),
		zap.String("fix_type", fixType),
		zap.String("locked_by", lock.LockedBy),
	)
	emitChromeLockBlockedChangeItem(ctx, params.DB, siteID, slot, lock,
		"overwrite", "fix_component_template", logger)

	return map[string]interface{}{
		"fixed":     false,
		"locked":    true,
		"action":    "needs_review",
		"fix_type":  fixType,
		"slot_name": slot,
		"reason": fmt.Sprintf("chrome slot %q is locked by %q — automated fix refused; "+
			"unlock via the admin dashboard to apply it", slot, lock.LockedBy),
	}
}

// chromeFixLockRefused handles a guarded write that matched no row. Almost
// always the lock arrived between the pre-check and the write; if the re-check
// disagrees the row was deleted concurrently, which is still not a success.
func chromeFixLockRefused(ctx context.Context, params ActionParams, siteID uuid.UUID,
	slot, fixType string, logger *zap.Logger) map[string]interface{} {

	if skip := chromeFixLockSkip(ctx, params, siteID, slot, fixType, logger); skip != nil {
		return skip
	}
	return map[string]interface{}{
		"fixed":     false,
		"action":    "needs_review",
		"fix_type":  fixType,
		"slot_name": slot,
		"reason":    fmt.Sprintf("chrome slot %q matched no row at write time — locked or removed concurrently", slot),
	}
}

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
	case "scope_component_instance", "instance_scope_conversion":
		// bugs_open/283 / RFC_034: the deterministic half of the per-instance
		// conversion programme. "instance_scope_conversion" is the item_type
		// spelling so a work item routes without a spec.fix_type.
		return fixScopeComponentInstance(ctx, params, logger)
	case "repair_instance_scope_bindings":
		// bugs_open/283 §14: pass 5 applied to a row the mechanical batch
		// already converted before pass 5 existed. Same gate, same snapshot
		// contract, separate change_source so the repair is its own audit row.
		return fixRepairInstanceScopeBindings(ctx, params, logger)
	case "scope_component_instance_judged":
		// bugs_open/283 / RFC_034: the JUDGED half — gate + write for an
		// LLM-rewritten script, reached only after the arm above refused with
		// needs_script_scoping. Never routed from an item_type: it needs the
		// rewrite in hand.
		return fixScopeComponentInstanceJudged(ctx, params, logger)
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

	if skip := chromeFixLockSkip(ctx, params, siteID, slotName, "inject_nav_flex_css", logger); skip != nil {
		return skip, nil
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

	if err := setSiteComponentHTML(ctx, params.DB, siteID, slotName, html); err != nil {
		if errors.Is(err, errSiteComponentLocked) {
			return chromeFixLockRefused(ctx, params, siteID, slotName, "inject_nav_flex_css", logger), nil
		}
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

	// Human lock state for THIS site's slot (bugs_open/069). It deliberately
	// does NOT stop the shared template patch below — a shared CSS defect still
	// deserves the shared fix, and the other sites on that template need it —
	// but it does stop this action claiming the fix reaches THIS site, because
	// the re-render that would deploy it is now refused. Reporting
	// fixed:true/durable:true there would be exactly the "reports success and
	// closes the finding" failure this function's own header warns about.
	chromeLock, lockErr := CheckSiteComponentLock(ctx, params.DB, siteID, slot, logger)
	if lockErr != nil {
		logger.Warn("chrome_overflow_fix: chrome lock check failed — relying on the guarded update",
			zap.String("slot", slot), zap.Error(lockErr))
		chromeLock = &SiteComponentLockStatus{}
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

		result := map[string]interface{}{
			"fixed": true, "fix_type": "chrome_overflow_fix", "slot_name": slot,
			"overflow_selector": selector, "layer": "content_component",
			"component_id": componentID.String, "shared_sites": sharedSites,
			"detail": fmt.Sprintf("injected a mobile media query into the %s template (shared by %d site(s)) so %s wraps; a rerender deploys it and survives refresh", slot, sharedSites, selector),
		}
		if chromeLock.IsLocked {
			logger.Warn("chrome_overflow_fix: template patched, but THIS site's slot is human-locked — it will not reach this site until unlocked (bugs_open/069)",
				zap.String("slot", slot), zap.String("locked_by", chromeLock.LockedBy))
			emitChromeLockBlockedChangeItem(ctx, params.DB, siteID, slot, chromeLock,
				"rerender", "fix_component_template", logger)
			result["site_slot_locked"] = true
			result["durable_for_this_site"] = false
			result["action"] = "needs_review"
			result["detail"] = fmt.Sprintf("%s — but this site's %s slot is locked by %q, so the "+
				"re-render that would deploy it is refused; the other %d site(s) on this template still get the fix",
				result["detail"], slot, chromeLock.LockedBy, sharedSites-1)
		}
		return result, nil
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
	if skip := chromeFixLockSkip(ctx, params, siteID, slot, "chrome_overflow_fix", logger); skip != nil {
		return skip, nil
	}
	if err := appendSiteComponentHTML(ctx, params.DB, siteID, slot, overflowCSS); err != nil {
		if errors.Is(err, errSiteComponentLocked) {
			return chromeFixLockRefused(ctx, params, siteID, slot, "chrome_overflow_fix", logger), nil
		}
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

	if skip := chromeFixLockSkip(ctx, params, siteID, slotName, "responsive_fix", logger); skip != nil {
		return skip, nil
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

	if err := setSiteComponentHTML(ctx, params.DB, siteID, slotName, html); err != nil {
		if errors.Is(err, errSiteComponentLocked) {
			return chromeFixLockRefused(ctx, params, siteID, slotName, "responsive_fix", logger), nil
		}
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

	if skip := chromeFixLockSkip(ctx, params, siteID, slotName, "remove_element", logger); skip != nil {
		return skip, nil
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

	if err := setSiteComponentHTML(ctx, params.DB, siteID, slotName, html); err != nil {
		if errors.Is(err, errSiteComponentLocked) {
			return chromeFixLockRefused(ctx, params, siteID, slotName, "remove_element", logger), nil
		}
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

// resolveStatusRepairComponent finds the page_component this finding is really
// about, preferring the STABLE key over the stored id.
//
// bugs_open/300. `page_components.id` is not stable across re-renders — the
// estate's own rule, written in 016b and obeyed by revalidate_review_queue_action
// and create_report_page_action. The page_component_status_drift check stores the
// id as the finding's only handle, so an ordinary re-render between filing and
// dispatch turns a true finding into `sql.ErrNoRows`, which this action turned
// into a hard error and the item into `failed`. That is not merely a lost repair:
// detected-item-promoter's 25% floor (migration 444/454) cannot tell an
// artefact failure from an incompetent handler, so enough of them switch the
// whole item_type off — including the findings that are still true.
//
// [MEASURED 2026-08-18, all 82 lifetime page_component_status_drift rows]
// `spec.page_component_id` resolves for 70; `(page_id, slot_name)` resolves for
// 82 of 82. The 12 dead ids are the ageing this fixes. On 2026-08-17 the same
// query gave 16 of 16 deferred rows resolving by id; today 11 do — five died in a
// day, in a queue nobody touched, which is the mechanism rather than an anecdote.
//
// THE TIEBREAK IS NOT DECORATION. `(page_id, slot_name)` is NOT unique fleet-wide
// — [MEASURED 2026-08-18] 17 (page_id, slot_name) pairs carry more than one
// component, worst case 4. None of them is a drift row today, but resolving by
// the pair alone would silently pick an arbitrary component on the day one is.
// So: the stored id wins WITHIN the pair's matches when it is still alive, a lone
// match is taken, and genuine ambiguity is REFUSED rather than guessed — the same
// posture as the two guards below it.
//
// Returns the resolved id and how it was resolved (recorded on the result so a
// census can tell repaired-by-stable-key from repaired-by-stored-id), or ok=false
// with a caller-returnable reason when the subject cannot be identified.
func resolveStatusRepairComponent(
	ctx context.Context,
	params ActionParams,
	specIDStr, slotName, workItemIDStr string,
	logger *zap.Logger,
) (pcID uuid.UUID, resolvedBy string, reason string, ok bool) {
	var specID uuid.UUID
	specIDValid := false
	if specIDStr != "" {
		if parsed, err := uuid.Parse(specIDStr); err == nil {
			specID, specIDValid = parsed, true
		} else {
			logger.Warn("repair_page_component_status: spec.page_component_id is not a uuid",
				zap.String("page_component_id", specIDStr), zap.Error(err))
		}
	}

	// The stable pair. page_id lives on the work item ROW, not in the dispatch
	// payload — [VERIFIED 2026-08-18 against a live component-template-fixer
	// orchestration] input_data carries spec/domain/site_id/item_type/
	// component_id/current_page/work_item_id and NO page_id or page_name. Joining
	// through work_item_id is therefore what makes this fix possible without
	// touching the shared dispatch mapping, which every handler reads.
	if workItemIDStr != "" && slotName != "" {
		if workItemID, err := uuid.Parse(workItemIDStr); err == nil {
			rows, qErr := params.DB.QueryContext(ctx, `
				SELECT pc.id
				FROM page_components pc
				JOIN site_work_items wi ON wi.page_id = pc.page_id
				WHERE wi.id = $1 AND pc.slot_name = $2
				ORDER BY pc.id
			`, workItemID, slotName)
			if qErr != nil {
				// Fail open to the stored id: a lookup error must not be worse
				// than the behaviour this replaces.
				logger.Warn("repair_page_component_status: stable-key lookup failed, falling back to the stored id",
					zap.String("slot_name", slotName), zap.Error(qErr))
			} else {
				var matches []uuid.UUID
				for rows.Next() {
					var id uuid.UUID
					if scanErr := rows.Scan(&id); scanErr != nil {
						logger.Warn("repair_page_component_status: scan failed on the stable-key lookup", zap.Error(scanErr))
						break
					}
					matches = append(matches, id)
				}
				rows.Close()

				switch {
				case len(matches) == 1:
					if specIDValid && matches[0] != specID {
						logger.Info("repair_page_component_status: the stored component id is stale — resolved by (page_id, slot_name)",
							zap.String("stale_page_component_id", specIDStr),
							zap.String("resolved_page_component_id", matches[0].String()),
							zap.String("slot_name", slotName))
					}
					return matches[0], "page_id+slot_name", "", true
				case len(matches) > 1:
					// The stored id disambiguates when it is one of them.
					if specIDValid {
						for _, m := range matches {
							if m == specID {
								return m, "page_id+slot_name+stored_id_tiebreak", "", true
							}
						}
					}
					logger.Warn("repair_page_component_status: (page_id, slot_name) is ambiguous and the stored id is not among the matches — refusing to guess",
						zap.String("slot_name", slotName), zap.Int("matches", len(matches)),
						zap.String("stored_page_component_id", specIDStr))
					return uuid.Nil, "", fmt.Sprintf(
						"slot %q resolves to %d components on this page and the stored page_component_id is not one of them — refusing to guess which",
						slotName, len(matches)), false
				}
				// len(matches) == 0 falls through to the stored id below.
			}
		} else {
			logger.Warn("repair_page_component_status: work_item_id is not a uuid",
				zap.String("work_item_id", workItemIDStr), zap.Error(err))
		}
	}

	if specIDValid {
		return specID, "spec.page_component_id", "", true
	}
	return uuid.Nil, "", "no usable subject: (page_id, slot_name) resolved nothing and spec.page_component_id is absent or unparseable", false
}

func fixPageComponentStatus(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	pcIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_component_id")
	if pcIDStr == "" {
		pcIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.page_component_id")
	}
	specSlotName := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.slot_name")
	workItemIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")

	pcID, resolvedBy, resolveReason, resolved := resolveStatusRepairComponent(
		ctx, params, pcIDStr, specSlotName, workItemIDStr, logger)
	if !resolved {
		if resolveReason == "" {
			return nil, fmt.Errorf("page_component_id is required for repair_page_component_status")
		}
		// A subject we cannot identify is not a repair we attempted. Refusing
		// softly keeps it off the promoter's failure ledger while still leaving
		// the reason on the row for a human — the same shape as the guards below.
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason": resolveReason,
			"action": "needs_review",
		}, nil
	}

	var observed, slotName, pageStatus string
	var hasHTML bool
	err := params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(pc.build_status, ''), COALESCE(pc.slot_name, ''),
		       COALESCE(p.build_status, ''),
		       (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.id = $1
	`, pcID).Scan(&observed, &slotName, &pageStatus, &hasHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to load page_component %s (resolved by %s): %w", pcID, resolvedBy, err)
	}
	pcIDStr = pcID.String()

	switch {
	case observed == "deployed":
		// This is the arm bugs_open/300's measured instance lands in once the
		// subject is resolved by the stable key: the drift was real when filed
		// and an ordinary re-render fixed it days later. Under the old lookup
		// the same row was a hard error, and so a vote against the pair.
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason":            "already deployed",
			"page_component_id": pcIDStr,
			"resolved_by":       resolvedBy,
		}, nil
	case knownRebuildStatuses[observed]:
		logger.Info("repair_page_component_status: refusing to flip an honest status",
			zap.String("slot", slotName), zap.String("build_status", observed))
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"observed_status":   observed,
			"reason":            "status is a legitimate awaiting-rebuild state — needs a rebuild, not a status flip",
			"action":            "needs_review",
			"page_component_id": pcIDStr,
			"resolved_by":       resolvedBy,
		}, nil
	case pageStatus != "deployed":
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason":            fmt.Sprintf("parent page build_status is %q, not deployed", pageStatus),
			"action":            "needs_review",
			"page_component_id": pcIDStr,
			"resolved_by":       resolvedBy,
		}, nil
	case !hasHTML:
		return map[string]interface{}{
			"fixed": false, "fix_type": "repair_page_component_status",
			"reason":            "component has no rendered_html — cannot claim it is deployed",
			"action":            "needs_review",
			"page_component_id": pcIDStr,
			"resolved_by":       resolvedBy,
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
		"resolved_by":       resolvedBy,
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

// ---------------------------------------------------------------------------
// scope_component_instance: the deterministic half of bugs_open/283's
// conversion programme (RFC_034, owner ruling 2026-08-17 — hybrid shape,
// through the framework).
//
// Targets content_components.html_template, keyed by spec.component_id — the
// ROW, never the function: four functions carry forks, and a function-keyed
// conversion silently skips nine rows (RFC_034 §1).
//
// The transform and the acceptance gate live in
// component_instance_conversion.go; this wrapper is the framework seam: load,
// convert, GATE, snapshot to component_versions, write. Three exits and each
// says which pool the component is in:
//
//   fixed:true                       — converted, gated clean, WRITTEN. The
//                                      pages carrying it still serve stored
//                                      rendered_html, so the result reaches
//                                      visitors only on the next
//                                      rerender+deploy — the caller's plan
//                                      owns that step, and the result names it.
//   fixed:false action:"needs_script_scoping"
//                                    — ids convert cleanly but the script
//                                      genuinely declares into global scope;
//                                      RFC_034 §2.1 forbids shipping that
//                                      half-state, so NOTHING was written.
//                                      Route to the judged (LLM) pipeline.
//   fixed:false + reason             — refused or nothing to do (already
//                                      converted, no ids, hex-ambiguous id,
//                                      unrecognised binding construction).
//                                      NOTHING was written.
// ---------------------------------------------------------------------------

func fixScopeComponentInstance(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	componentIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.component_id")
	if componentIDStr == "" {
		componentIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.component_id")
	}
	if componentIDStr == "" {
		return nil, fmt.Errorf("component_id is required for scope_component_instance (expected in spec.component_id)")
	}
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid component_id %q: %w", componentIDStr, err)
	}

	var function, currentTemplate string
	err = params.DB.QueryRowContext(ctx, `
		SELECT function, html_template
		FROM content_components
		WHERE id = $1
	`, componentID).Scan(&function, &currentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to load component %s: %w", componentIDStr, err)
	}

	converted, rep, ok := ConvertTemplateToInstanceScope(currentTemplate)
	if !ok {
		logger.Info("scope_component_instance: refused/no-op",
			zap.String("function", function),
			zap.String("reason", rep.RefusedReason),
			zap.Bool("already_converted", rep.AlreadyConverted))
		return map[string]interface{}{
			"fixed":        false,
			"fix_type":     "scope_component_instance",
			"function":     function,
			"component_id": componentIDStr,
			"reason":       rep.RefusedReason,
		}, nil
	}

	needsJudged, gateErr := GateConvertedTemplate(function, converted, logger)
	if gateErr != nil {
		// A gate ERROR is a transform defect on this template, not a judged
		// case — fail the step loudly so the defect is diagnosed, never
		// written around.
		return nil, fmt.Errorf("scope_component_instance gate failed for %q: %w", function, gateErr)
	}
	if needsJudged {
		logger.Info("scope_component_instance: ids convert cleanly but the script declares into global scope — refusing the half-state (RFC_034 §2.1), routing to the judged pool",
			zap.String("function", function),
			zap.Strings("ids", rep.IDsDeclared))
		// The judged branch of the workflow picks these up: the IDS-CONVERTED
		// template is what the LLM is given (so it never touches the surfaces
		// the deterministic pass is proven on), and the handler inventory is
		// what the brief names. Nothing is written here.
		handlers := InlineHandlerInventory(converted)
		if handlers == nil {
			handlers = []string{}
		}
		ub := rep.UnprefixedBindings
		if ub == nil {
			ub = []string{}
		}
		reason := "script declares into global scope; deterministic conversion alone would ship a page that reads clean on ids while every button runs the last instance's logic"
		if len(ub) > 0 {
			reason = "deterministic conversion left " + fmt.Sprint(len(ub)) + " binding(s) unprefixed (" + strings.Join(ub, "; ") + ") — and/or the script declares into global scope; shipping would dangle those lookups at runtime"
		}
		return map[string]interface{}{
			"fixed":               false,
			"fix_type":            "scope_component_instance",
			"function":            function,
			"component_id":        componentIDStr,
			"action":              "needs_script_scoping",
			"reason":              reason,
			"ids_declared":        len(rep.IDsDeclared),
			"converted_template":  converted,
			"inline_handlers":     handlers,
			"inline_handler_n":    len(handlers),
			"window_onload_n":     WindowOnloadCount(converted),
			"unprefixed_bindings": ub,
		}, nil
	}

	maxVersion, err := writeScopedTemplate(ctx, params, componentID, function, currentTemplate, converted, "scope_component_instance", logger)
	if err != nil {
		return nil, err
	}

	logger.Info("scope_component_instance: converted and written",
		zap.String("function", function),
		zap.String("component_id", componentIDStr),
		zap.Int("ids", len(rep.IDsDeclared)),
		zap.Int("id_attrs", rep.IDAttrsRenamed),
		zap.Int("get_element_by_id", rep.GetElementByID),
		zap.Int("id_ref_attrs", rep.IDRefAttrs),
		zap.Int("hash_refs", rep.HashRefs),
		zap.Int("binding_literals", rep.Bindings.LiteralIDsRenamed),
		zap.Int("binding_concat_sites", rep.Bindings.ConcatSitesRenamed),
		zap.Int("snapshot_version", maxVersion+1))

	return map[string]interface{}{
		"fixed":                true,
		"fix_type":             "scope_component_instance",
		"function":             function,
		"component_id":         componentIDStr,
		"ids_declared":         len(rep.IDsDeclared),
		"id_attrs_renamed":     rep.IDAttrsRenamed,
		"get_element_by_id":    rep.GetElementByID,
		"id_ref_attrs":         rep.IDRefAttrs,
		"hash_refs":            rep.HashRefs,
		"binding_literals":     rep.Bindings.LiteralIDsRenamed,
		"binding_concat_sites": rep.Bindings.ConcatSitesRenamed,
		"snapshot_version":     maxVersion + 1,
		"note":                 "pages carrying this component still serve stored rendered_html; the conversion reaches visitors on their next rerender+deploy",
	}, nil
}

// writeScopedTemplate is the one write both halves of the 283 programme share:
// snapshot the CURRENT template to component_versions under the given
// change_source (reversibility contract, same as repair_template_slots;
// non-fatal if the snapshot fails), then overwrite html_template. Returns the
// pre-snapshot max version so the caller can report snapshot_version = max+1.
func writeScopedTemplate(ctx context.Context, params ActionParams, componentID uuid.UUID, function, currentTemplate, newTemplate, changeSource string, logger *zap.Logger) (int, error) {
	var maxVersion int
	_ = params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version_number), 0) FROM component_versions WHERE component_id = $1
	`, componentID).Scan(&maxVersion)
	if _, snapErr := params.DB.ExecContext(ctx, `
		INSERT INTO component_versions (component_id, version_number, html_template, change_source)
		VALUES ($1, $2, $3, $4)
	`, componentID, maxVersion+1, currentTemplate, changeSource); snapErr != nil {
		logger.Warn(changeSource+": snapshot failed, continuing",
			zap.String("function", function), zap.Error(snapErr))
	}
	if _, err := params.DB.ExecContext(ctx, `
		UPDATE content_components
		SET html_template = $1, updated_at = now()
		WHERE id = $2
	`, newTemplate, componentID); err != nil {
		return maxVersion, fmt.Errorf("failed to update template for %s: %w", function, err)
	}
	return maxVersion, nil
}

// repair_instance_scope_bindings: bugs_open/283 §14 (2026-08-19). The
// mechanical batch converted 69 templates with a converter whose completeness
// check could not see a binding that does not CONTAIN the id literal — an
// array of ids, a {id:'x'} config, a helper call el('x'), a concatenated
// lookup 'step-' + n. 32 rows carry at least one such dangling binding; 14 were
// serving it live. This arm applies pass 5 (component_instance_bindings.go) to
// an ALREADY-CONVERTED row: load, repair, require UnprefixedBindings empty and
// the two-instance gate clean, snapshot under its own change_source, write.
// The fixer then files the page-scoped template_changed rerenders as for any
// template fix. Three exits, same contract as the mechanical arm:
//
//	fixed:true                         — repaired, gated, WRITTEN.
//	fixed:false action:"needs_script_scoping"
//	                                   — pass 5 could not place every binding
//	                                     (a refuse-context literal, a dynamic
//	                                     declaration with no static prefix);
//	                                     nothing written; judged pool.
//	fixed:false + reason               — not converted (wrong arm), or
//	                                     nothing to repair. Nothing written.
func fixRepairInstanceScopeBindings(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	componentIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.component_id")
	if componentIDStr == "" {
		componentIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.component_id")
	}
	if componentIDStr == "" {
		return nil, fmt.Errorf("component_id is required for repair_instance_scope_bindings (expected in spec.component_id)")
	}
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid component_id %q: %w", componentIDStr, err)
	}

	var function, currentTemplate string
	err = params.DB.QueryRowContext(ctx, `
		SELECT function, html_template
		FROM content_components
		WHERE id = $1
	`, componentID).Scan(&function, &currentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to load component %s: %w", componentIDStr, err)
	}

	base := map[string]interface{}{
		"fixed":        false,
		"fix_type":     "repair_instance_scope_bindings",
		"function":     function,
		"component_id": componentIDStr,
	}
	if !strings.Contains(currentTemplate, "{{.InstanceID}}") {
		base["reason"] = "template is not converted — nothing to repair; use scope_component_instance"
		return base, nil
	}

	before := UnprefixedBindings(currentTemplate)
	repaired, brep, rejects, ok := RepairConvertedTemplateBindings(currentTemplate)
	if !ok {
		base["action"] = "needs_script_scoping"
		base["reason"] = fmt.Sprintf("dynamic id declaration %q has no static `-`/`_` prefix to carry the token — judged pool", rejects[0])
		base["converted_template"] = currentTemplate
		base["unprefixed_bindings"] = before
		return base, nil
	}
	after := UnprefixedBindings(repaired)
	if len(after) > 0 {
		// Pass 5 placed what it recognises; something remains. Refuse to the
		// judged pool with the repaired-so-far template as the baseline the
		// LLM will see (the judged arm re-derives exactly this).
		handlers := InlineHandlerInventory(repaired)
		if handlers == nil {
			handlers = []string{}
		}
		base["action"] = "needs_script_scoping"
		base["reason"] = fmt.Sprintf("pass 5 left %d binding(s) unprefixed (%s) — judged pool", len(after), strings.Join(after, "; "))
		base["converted_template"] = repaired
		base["inline_handlers"] = handlers
		base["inline_handler_n"] = len(handlers)
		base["window_onload_n"] = WindowOnloadCount(repaired)
		base["unprefixed_bindings"] = after
		base["bindings_placed"] = brep.LiteralIDsRenamed + brep.ConcatSitesRenamed
		return base, nil
	}
	if repaired == currentTemplate {
		base["reason"] = "no unprefixed bindings found and nothing changed — already sound"
		base["unprefixed_before"] = len(before)
		return base, nil
	}

	needsJudged, gateErr := GateConvertedTemplate(function, repaired, logger)
	if gateErr != nil {
		return nil, fmt.Errorf("repair_instance_scope_bindings gate failed for %q: %w", function, gateErr)
	}
	if needsJudged {
		// Bindings are placed but the script itself is unscoped — a row
		// that should never have been written by the mechanical arm; route.
		base["action"] = "needs_script_scoping"
		base["reason"] = "bindings repaired but the script declares into global scope — judged pool"
		base["converted_template"] = repaired
		return base, nil
	}
	if issues := componentRegressionIssues(currentTemplate, repaired); len(issues) > 0 {
		return nil, fmt.Errorf("repair_instance_scope_bindings: write guard refused %q: %s", function, strings.Join(issues, " | "))
	}

	maxVersion, err := writeScopedTemplate(ctx, params, componentID, function, currentTemplate, repaired, "repair_instance_scope_bindings", logger)
	if err != nil {
		return nil, err
	}
	logger.Info("repair_instance_scope_bindings: repaired and written",
		zap.String("function", function),
		zap.String("component_id", componentIDStr),
		zap.Int("unprefixed_before", len(before)),
		zap.Int("binding_literals", brep.LiteralIDsRenamed),
		zap.Int("binding_concat_sites", brep.ConcatSitesRenamed),
		zap.Strings("concat_prefixes", brep.ConcatPrefixes),
		zap.Int("snapshot_version", maxVersion+1))

	return map[string]interface{}{
		"fixed":                  true,
		"fix_type":               "repair_instance_scope_bindings",
		"function":               function,
		"component_id":           componentIDStr,
		"unprefixed_before":      len(before),
		"unprefixed_before_list": before,
		"binding_literals":       brep.LiteralIDsRenamed,
		"binding_concat_sites":   brep.ConcatSitesRenamed,
		"concat_prefixes":        brep.ConcatPrefixes,
		"snapshot_version":       maxVersion + 1,
		"note":                   "pages carrying this component still serve stored rendered_html; the repair reaches visitors on their next rerender+deploy",
	}, nil
}

// scope_component_instance_judged: the JUDGED half of bugs_open/283 (RFC_034
// shape C; design PLAN_2026-08-18_judged_pipeline.md). Reached by the
// component-template-fixer workflow only after scope_component_instance
// refused with needs_script_scoping and an execute_llm_prompt step rewrote the
// script. This arm is the GATE and the WRITE, fused so nothing can sit between
// them:
//
//  1. re-load the row and RE-DERIVE the ids-converted baseline itself — the
//     workflow-carried copy is never trusted (another session may have
//     converted, deactivated or edited the row between steps);
//  2. if the baseline now gates clean (corpus drift: someone scoped the
//     script meanwhile), converge on the mechanical contract and write the
//     baseline — the LLM output is discarded;
//  3. otherwise JudgedConversionIssues (two-instance gate fully clean, markup
//     parity outside <script>, id-set parity, no surviving on*=) then the
//     comparative write guard (collapse ratio, ends-mid-token — the
//     bugs_open/012 class, second layer behind execute_llm_prompt's own
//     capped-completion refusal);
//  4. snapshot under change_source='scope_component_instance_judged', write.
//
// Any refusal returns fixed:false with the failing checks named and writes
// NOTHING; the workflow routes that to needs_human_review. No automatic retry
// — 25 components; a handful of refusals is triage, not a queue.
//
// The rewrite is read from the step-config `html_field` path (default
// "scoped_script.result", the execute_llm_prompt output shape). The component
// is keyed by spec.component_id, the ROW — same as the mechanical arm.
func fixScopeComponentInstanceJudged(ctx context.Context, params ActionParams, logger *zap.Logger) (interface{}, error) {
	componentIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.component_id")
	if componentIDStr == "" {
		componentIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.component_id")
	}
	if componentIDStr == "" {
		return nil, fmt.Errorf("component_id is required for scope_component_instance_judged (expected in spec.component_id)")
	}
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid component_id %q: %w", componentIDStr, err)
	}

	htmlField, _ := params.StepConfig.Config["html_field"].(string)
	if htmlField == "" {
		htmlField = "scoped_script.result"
	}
	rewrite := datahelpers.ExtractNestedFieldString(params.CollectedData, htmlField)
	if strings.TrimSpace(rewrite) == "" {
		return nil, fmt.Errorf("scope_component_instance_judged: no rewrite found at %q — the LLM step produced nothing, or html_field names the wrong path", htmlField)
	}
	rewrite = StripMarkdownFence(rewrite)

	var function, currentTemplate string
	err = params.DB.QueryRowContext(ctx, `
		SELECT function, html_template
		FROM content_components
		WHERE id = $1
	`, componentID).Scan(&function, &currentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to load component %s: %w", componentIDStr, err)
	}

	refuse := func(reason string, extra map[string]interface{}) (interface{}, error) {
		logger.Warn("scope_component_instance_judged: REFUSED, nothing written",
			zap.String("function", function),
			zap.String("component_id", componentIDStr),
			zap.String("reason", reason))
		out := map[string]interface{}{
			"fixed":        false,
			"fix_type":     "scope_component_instance_judged",
			"function":     function,
			"component_id": componentIDStr,
			"action":       "needs_human_review",
			"reason":       reason,
		}
		for k, v := range extra {
			out[k] = v
		}
		return out, nil
	}

	// Step 1 — the baseline, re-derived. Deterministic, so identical to what
	// the earlier step handed the LLM unless the row moved underneath us.
	baseline, rep, ok := ConvertTemplateToInstanceScope(currentTemplate)
	if !ok && rep.AlreadyConverted {
		// A row that was converted BEFORE pass 5 and routed here by the
		// repair arm: the baseline the LLM saw is the current row with the
		// binding passes applied (RepairConvertedTemplateBindings is
		// deterministic, so this re-derives it exactly).
		repaired, _, rejects, rok := RepairConvertedTemplateBindings(currentTemplate)
		if !rok {
			return refuse(fmt.Sprintf("baseline could not be re-derived: dynamic id declaration %q has no static prefix", rejects[0]), nil)
		}
		baseline, ok = repaired, true
		rep.IDsDeclared, _, _ = BindingIDSets(currentTemplate)
	}
	if !ok {
		return refuse("deterministic baseline could not be re-derived: "+rep.RefusedReason, nil)
	}

	// Step 2 — converge if the judged pool no longer applies.
	needsJudged, gateErr := GateConvertedTemplate(function, baseline, logger)
	if gateErr != nil {
		return nil, fmt.Errorf("scope_component_instance_judged: baseline gate failed for %q: %w", function, gateErr)
	}
	if !needsJudged {
		maxVersion, werr := writeScopedTemplate(ctx, params, componentID, function, currentTemplate, baseline, "scope_component_instance", logger)
		if werr != nil {
			return nil, werr
		}
		logger.Info("scope_component_instance_judged: baseline gates clean (script scoped since routing) — wrote the MECHANICAL conversion, LLM output discarded",
			zap.String("function", function), zap.Int("snapshot_version", maxVersion+1))
		return map[string]interface{}{
			"fixed":            true,
			"fix_type":         "scope_component_instance",
			"converged_from":   "scope_component_instance_judged",
			"function":         function,
			"component_id":     componentIDStr,
			"ids_declared":     len(rep.IDsDeclared),
			"snapshot_version": maxVersion + 1,
			"note":             "script was already scoped by the time the judged arm ran; mechanical conversion written, rewrite discarded",
		}, nil
	}

	// Step 3 — the gate on the rewrite, then the comparative write guard.
	if issues := JudgedConversionIssues(function, baseline, rewrite, logger); len(issues) > 0 {
		return refuse("rewrite refused by the judged gate: "+strings.Join(issues, " | "), map[string]interface{}{
			"gate_issues": issues,
			"rewrite_len": len(rewrite),
		})
	}
	if issues := componentRegressionIssues(currentTemplate, rewrite); len(issues) > 0 {
		return refuse("rewrite refused by the comparative write guard: "+strings.Join(issues, " | "), map[string]interface{}{
			"guard_issues": issues,
			"rewrite_len":  len(rewrite),
			"current_len":  len(currentTemplate),
		})
	}

	// Step 4 — write.
	maxVersion, err := writeScopedTemplate(ctx, params, componentID, function, currentTemplate, rewrite, "scope_component_instance_judged", logger)
	if err != nil {
		return nil, err
	}

	handlersBefore := InlineHandlerInventory(baseline)
	logger.Info("scope_component_instance_judged: rewrite gated clean and written",
		zap.String("function", function),
		zap.String("component_id", componentIDStr),
		zap.Int("ids", len(rep.IDsDeclared)),
		zap.Int("inline_handlers_rewired", len(handlersBefore)),
		zap.Int("window_onload_replaced", WindowOnloadCount(baseline)-WindowOnloadCount(rewrite)),
		zap.Int("rewrite_len", len(rewrite)),
		zap.Int("current_len", len(currentTemplate)),
		zap.Int("snapshot_version", maxVersion+1))

	return map[string]interface{}{
		"fixed":                   true,
		"fix_type":                "scope_component_instance_judged",
		"function":                function,
		"component_id":            componentIDStr,
		"ids_declared":            len(rep.IDsDeclared),
		"inline_handlers_rewired": len(handlersBefore),
		"window_onload_before":    WindowOnloadCount(baseline),
		"window_onload_after":     WindowOnloadCount(rewrite),
		"rewrite_len":             len(rewrite),
		"current_len":             len(currentTemplate),
		"snapshot_version":        maxVersion + 1,
		"note":                    "pages carrying this component still serve stored rendered_html; generic pages take the conversion on their template_changed rerender, owned pages on their section_edit delivery",
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
