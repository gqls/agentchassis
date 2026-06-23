// FILE: platform/orchestration/actions/discovery_checks/check_component_standards.go
//
// Discovery check that audits site components against the contracts in 003b.
// Detects: unlinked site_components, missing data-component attributes,
// slot_name mismatches, missing site metadata, missing asset refs in rendered HTML,
// stacked nav (missing flex CSS), unwanted elements (search icon),
// empty page sections, and broken template slots (<no value> artifacts in
// html_template that prevent content substitution during rendering).
//
// Each sub-check produces WorkItemSpec entries routed to the appropriate handler.
// Register with: checks = ["validate_component_standards"] in workflow config.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&ComponentStandardsCheck{}) }

type ComponentStandardsCheck struct{}

func (c *ComponentStandardsCheck) Name() string { return "validate_component_standards" }

func (c *ComponentStandardsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	checkUnlinkedSiteComponents(dctx, result)
	checkSlotNameMismatch(dctx, result)
	checkMissingSiteMetadata(dctx, result)
	checkMissingAssetRefs(dctx, result)
	checkNavLayout(dctx, result)
	checkUnwantedElements(dctx, result)
	checkEmptyPageSections(dctx, result)
	checkBrokenTemplateSlots(dctx, result)

	dctx.Logger.Info("ComponentStandardsCheck complete",
		zap.Int("findings", len(result.Findings)),
		zap.Int("work_items", len(result.WorkItems)),
	)

	return result, nil
}

// ---------------------------------------------------------------------------
// Sub-check: unlinked site_components
// site_components rows with NULL component_id for header/footer/head.
// Without linkage, renderAndStoreSiteComponent falls through to a hardcoded
// fallback that ignores the style collection's template.
// ---------------------------------------------------------------------------

func checkUnlinkedSiteComponents(dctx DiscoveryCheckContext, result *CheckResult) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.slot_name IN ('header', 'footer', 'head')
		  AND sc.component_id IS NULL
	`, dctx.SiteID)
	if err != nil {
		dctx.Logger.Warn("checkUnlinkedSiteComponents: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var slotName string
		if err := rows.Scan(&slotName); err != nil {
			continue
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "unlinked_site_component",
			"slot_name": slotName,
		})

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "unlinked_site_component",
			"slot_name": slotName,
			"detail":    fmt.Sprintf("%s component_id is NULL — using fallback rendering", slotName),
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "unlinked_site_component",
			Severity:     "high",
			Summary:      fmt.Sprintf("%s component not linked to template — using fallback rendering", slotName),
			SpecJSON:     string(specJSON),
			Priority:     5,
			HandlerAgent: "site-component-linker",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("unlinked_%s_%s", slotName, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}
}

// ---------------------------------------------------------------------------
// Sub-check: slot_name ↔ data-component mismatch
// page_components where slot_name doesn't match the data-component attribute
// in rendered_html. This breaks component lookup during section-editor edits.
// ---------------------------------------------------------------------------

func checkSlotNameMismatch(dctx DiscoveryCheckContext, result *CheckResult) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.id, pc.slot_name, p.name as page_name,
		       substring(pc.rendered_html from 'data-component="([^"]*)"') as data_component
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.slot_name IS NOT NULL AND pc.slot_name != ''
		  AND pc.rendered_html LIKE '%data-component=%'
		  AND pc.slot_name != substring(pc.rendered_html from 'data-component="([^"]*)"')
	`, dctx.SiteID)
	if err != nil {
		dctx.Logger.Warn("checkSlotNameMismatch: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var pcID, slotName, pageName string
		var dataComponent sql.NullString
		if err := rows.Scan(&pcID, &slotName, &pageName, &dataComponent); err != nil {
			continue
		}
		if !dataComponent.Valid {
			continue
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":             "slot_name_mismatch",
			"page_component_id": pcID,
			"page_name":         pageName,
			"current_slot_name": slotName,
			"data_component":    dataComponent.String,
			"fix_type":          "align_slot_name",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "slot_name_mismatch",
			Severity:     "medium",
			Summary:      fmt.Sprintf("page %s: slot_name '%s' != data-component '%s'", pageName, slotName, dataComponent.String),
			SpecJSON:     string(specJSON),
			Priority:     15,
			HandlerAgent: "component-template-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("slot_mismatch_%s", pcID),
			BatchID:      dctx.BatchID,
		})
		count++
	}

	if count > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "slot_name_mismatch",
			"count": count,
		})
	}
}

// ---------------------------------------------------------------------------
// Sub-check: missing site metadata
// Sites where company_name, tagline, logo_url, or email are empty.
// ---------------------------------------------------------------------------

func checkMissingSiteMetadata(dctx DiscoveryCheckContext, result *CheckResult) {
	var companyName, tagline, logoURL, email sql.NullString
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT
			NULLIF(COALESCE(company_name, ''), ''),
			NULLIF(COALESCE(tagline, ''), ''),
			NULLIF(COALESCE(logo_url, ''), ''),
			NULLIF(COALESCE(email, ''), '')
		FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&companyName, &tagline, &logoURL, &email)
	if err != nil {
		dctx.Logger.Warn("checkMissingSiteMetadata: query failed", zap.Error(err))
		return
	}

	var missing []string
	if !companyName.Valid {
		missing = append(missing, "company_name")
	}
	if !tagline.Valid {
		missing = append(missing, "tagline")
	}
	if !logoURL.Valid {
		missing = append(missing, "logo_url")
	}
	if !email.Valid {
		missing = append(missing, "email")
	}

	if len(missing) == 0 {
		return
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":          "missing_site_metadata",
		"missing_fields": missing,
		"derive_from":    "content_data",
	})

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":          "missing_site_metadata",
		"missing_fields": missing,
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "missing_site_metadata",
		Severity:     "high",
		Summary:      fmt.Sprintf("Missing site metadata: %s", strings.Join(missing, ", ")),
		SpecJSON:     string(specJSON),
		Priority:     3,
		HandlerAgent: "site-metadata-fixer",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("missing_metadata_%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})
}

// ---------------------------------------------------------------------------
// Sub-check: missing asset refs in rendered HTML
// Sites with logo_url set but header HTML has no <img> tag.
// ---------------------------------------------------------------------------

func checkMissingAssetRefs(dctx DiscoveryCheckContext, result *CheckResult) {
	var logoURL sql.NullString
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT NULLIF(COALESCE(logo_url, ''), '') FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&logoURL)

	if !logoURL.Valid {
		return // no logo_url set, nothing to check
	}

	var headerHTML sql.NullString
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT rendered_html FROM site_components
		WHERE site_id = $1 AND slot_name = 'header'
	`, dctx.SiteID).Scan(&headerHTML)

	if !headerHTML.Valid {
		return
	}

	if strings.Contains(headerHTML.String, "<img") {
		return // header already has an image tag
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":        "missing_logo_in_header",
		"logo_url":     logoURL.String,
		"likely_cause": "site_components.component_id is NULL or template doesn't use {{.logo_url}}",
	})

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":  "missing_logo_in_header",
		"detail": "logo_url is set but header has no <img> tag",
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "missing_logo_in_header",
		Severity:     "high",
		Summary:      "Logo URL set but header doesn't render <img> tag — likely unlinked component",
		SpecJSON:     string(specJSON),
		Priority:     5,
		HandlerAgent: "site-component-linker",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("missing_logo_header_%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})
}

// ---------------------------------------------------------------------------
// Sub-check: nav layout (stacked nav)
// Header HTML has a <ul> but no display:flex in accompanying CSS.
// ---------------------------------------------------------------------------

func checkNavLayout(dctx DiscoveryCheckContext, result *CheckResult) {
	var headerHTML sql.NullString
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT rendered_html FROM site_components
		WHERE site_id = $1 AND slot_name = 'header'
	`, dctx.SiteID).Scan(&headerHTML)

	if !headerHTML.Valid || headerHTML.String == "" {
		return
	}

	html := headerHTML.String
	hasNavList := strings.Contains(html, "<ul")
	hasFlexCSS := strings.Contains(html, "display: flex") || strings.Contains(html, "display:flex")

	if !hasNavList || hasFlexCSS {
		return // nav is either absent or already has flex
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":     "stacked_nav",
		"fix_type":  "inject_nav_flex_css",
		"slot_name": "header",
	})

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":  "stacked_nav",
		"detail": "Header nav <ul> has no display:flex — links stack vertically",
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "stacked_nav",
		Severity:     "high",
		Summary:      "Header nav links stack vertically — missing flex CSS",
		SpecJSON:     string(specJSON),
		Priority:     5,
		HandlerAgent: "component-template-fixer",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("stacked_nav_%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})
}

// ---------------------------------------------------------------------------
// Sub-check: unwanted nav elements (search icon)
// Header contains a search toggle button when no search functionality exists.
// ---------------------------------------------------------------------------

func checkUnwantedElements(dctx DiscoveryCheckContext, result *CheckResult) {
	var headerHTML sql.NullString
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT rendered_html FROM site_components
		WHERE site_id = $1 AND slot_name = 'header'
	`, dctx.SiteID).Scan(&headerHTML)

	if !headerHTML.Valid || headerHTML.String == "" {
		return
	}

	if !strings.Contains(headerHTML.String, "search-toggle") &&
		!strings.Contains(headerHTML.String, `aria-label="Search"`) {
		return
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":     "unwanted_nav_element",
		"fix_type":  "remove_element",
		"slot_name": "header",
		"pattern":   "search-toggle",
	})

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":  "unwanted_nav_element",
		"detail": "Header contains search icon — site has no search functionality",
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "unwanted_nav_element",
		Severity:     "low",
		Summary:      "Header contains search icon — no search functionality on site",
		SpecJSON:     string(specJSON),
		Priority:     50,
		HandlerAgent: "component-template-fixer",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("search_icon_%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})
}

// ---------------------------------------------------------------------------
// Sub-check: empty page sections
// Deployed pages with no rendered page_components.
// ---------------------------------------------------------------------------

func checkEmptyPageSections(dctx DiscoveryCheckContext, result *CheckResult) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, p.id::text
		FROM pages p
		LEFT JOIN page_components pc ON pc.page_id = p.id
			AND pc.rendered_html IS NOT NULL AND pc.rendered_html != ''
		WHERE p.site_id = $1
		  AND p.build_status IN ('deployed', 'active')
		GROUP BY p.id, p.name
		HAVING COUNT(pc.id) = 0
	`, dctx.SiteID)
	if err != nil {
		dctx.Logger.Warn("checkEmptyPageSections: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pageName, pageIDStr string
		if err := rows.Scan(&pageName, &pageIDStr); err != nil {
			continue
		}

		pageID, _ := uuid.Parse(pageIDStr)

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "empty_page_sections",
			"page_name": pageName,
			"page_id":   pageIDStr,
		})

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "empty_page_sections",
			"page_name": pageName,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       &pageID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "needs_content_page",
			Severity:     "high",
			Summary:      fmt.Sprintf("Page '%s' has no rendered sections", pageName),
			SpecJSON:     string(specJSON),
			Priority:     50,
			HandlerAgent: "page-content-writer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("empty_page_%s_%s", pageName, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}
}

// ---------------------------------------------------------------------------
// Sub-check: broken template slots (<no value> artifacts in html_template)
//
// Detects content_components whose html_template contains literal "<no value>"
// strings — a Go text/template render artifact that occurs when a template is
// executed against an empty data context and the rendered output is mistakenly
// stored back as the source template. Such a template cannot have its slots
// substituted at render time; all affected fields render as empty strings after
// RenderTemplate's cleanup pass, producing pages with blank headings, labels,
// and CTAs.
//
// StoreGeneratedComponentAction rejects templates with this pattern at creation
// time (via blockingIssues). This check catches components that pre-date that
// gate or were written through a path that bypasses it.
//
// Scoped to components used by this site's pages to avoid false positives on
// unrelated components in the shared library.
// ---------------------------------------------------------------------------

func checkBrokenTemplateSlots(dctx DiscoveryCheckContext, result *CheckResult) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT cc.id::text, cc.function,
		       (LENGTH(cc.html_template) - LENGTH(REPLACE(cc.html_template, '<no value>', '')))
		           / LENGTH('<no value>') AS artifact_count
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND cc.html_template LIKE '%<no value>%'
		  AND cc.is_active = true
	`, dctx.SiteID)
	if err != nil {
		dctx.Logger.Warn("checkBrokenTemplateSlots: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var componentID, function string
		var artifactCount int
		if err := rows.Scan(&componentID, &function, &artifactCount); err != nil {
			continue
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":          "broken_template_slots",
			"component_id":   componentID,
			"function":       function,
			"artifact_count": artifactCount,
			"fix_type":       "repair_template_slots",
		})

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":          "broken_template_slots",
			"function":       function,
			"artifact_count": artifactCount,
			"detail": fmt.Sprintf(
				"component %q html_template contains %d '<no value>' artifacts — slots cannot be substituted",
				function, artifactCount),
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "broken_template_slots",
			Severity:     "high",
			Summary:      fmt.Sprintf("Component %q has %d broken template slots — content cannot be injected", function, artifactCount),
			SpecJSON:     string(specJSON),
			Priority:     5,
			HandlerAgent: "component-template-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("broken_slots_%s_%s", function, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})

		dctx.Logger.Info("checkBrokenTemplateSlots: found broken component",
			zap.String("function", function),
			zap.String("component_id", componentID),
			zap.Int("artifact_count", artifactCount))
	}
}
