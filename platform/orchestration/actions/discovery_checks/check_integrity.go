// FILE: platform/orchestration/actions/discovery_checks/integrity_checks.go
//
// Structural and content integrity discovery checks.
// Registered via init() into the existing discovery check registry.
//
// Check placement per 009b improvement loop:
//   design-discovery-agent:        missing_style_collection, deactivated_site_components,
//                                  stale_site_components, shared_css_theme
//   completeness-discovery-agent:  cross_site_contamination, unrendered_templates
//
// All are Layer 1 (algorithmic, no LLM). They run before LLM audits in the
// improvement loop so the auditors see corrected state.
//
// Work item Domain is "build" for all items — the dispatch loop filters on
// domain = 'build'. The check_domain in the discovery agent config is a
// grouping label for the checks themselves, not the work item domain.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() {
	// Design-discovery checks (structural linkage)
	Register(&MissingStyleCollectionCheck{})
	Register(&DeactivatedSiteComponentsCheck{})
	Register(&StaleSiteComponentsCheck{})
	Register(&SharedCSSThemeCheck{})

	// Completeness-discovery checks (content integrity)
	Register(&CrossSiteContaminationCheck{})
	Register(&UnrenderedTemplatesCheck{})
}

// ============================================================================
// DESIGN-DISCOVERY CHECKS
// Wire into design-discovery-agent alongside undeployed_assets, missing_css, etc.
// ============================================================================

// --- missing_style_collection ---
// Detects sites with no style_collection_id. This causes the rerender to skip
// template resolution entirely, producing raw {{}} in rendered output.
// Also catches the name-not-resolved case: content_data.style_collection has
// a name but style_collection_id is NULL.

type MissingStyleCollectionCheck struct{}

func (c *MissingStyleCollectionCheck) Name() string { return "missing_style_collection" }

func (c *MissingStyleCollectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	var hasCollection bool
	var plannedName *string

	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT s.domain,
		       s.style_collection_id IS NOT NULL,
		       s.content_data->>'style_collection'
		FROM sites s WHERE s.id = $1
	`, dctx.SiteID).Scan(&domain, &hasCollection, &plannedName)
	if err != nil {
		return nil, fmt.Errorf("missing_style_collection: %w", err)
	}

	if !hasCollection {
		summary := fmt.Sprintf("Site %s has no style collection assigned", domain)
		reason := "no_style_collection"
		if plannedName != nil && *plannedName != "" {
			summary = fmt.Sprintf("Site %s has planned style '%s' but style_collection_id is NULL — name was never resolved to ID", domain, *plannedName)
			reason = "style_name_not_resolved"
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"domain":       domain,
			"planned_name": plannedName,
			"reason":       reason,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "missing_style_collection",
			Severity:     "high",
			Summary:      summary,
			SpecJSON:     string(spec),
			Priority:     1,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("no_style_%s", domain),
			BatchID:      dctx.BatchID,
		})

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":  "missing_style_collection",
			"domain": domain,
			"reason": reason,
		})
	}

	return result, nil
}

// --- deactivated_site_components ---
// Detects site_components pointing to content_components with is_active = false.
// These produce stale or broken rendering — the template may have been replaced
// by a newer active version.

type DeactivatedSiteComponentsCheck struct{}

func (c *DeactivatedSiteComponentsCheck) Name() string { return "deactivated_site_components" }

func (c *DeactivatedSiteComponentsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name, cc.name, cc.id::text
		FROM site_components sc
		JOIN content_components cc ON cc.id = sc.component_id
		WHERE sc.site_id = $1 AND cc.is_active = false
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("deactivated_site_components: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotName, componentName, componentID string
		if err := rows.Scan(&slotName, &componentName, &componentID); err != nil {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"slot_name":               slotName,
			"component_name":          componentName,
			"component_id":            componentID,
			"refresh_site_components": true,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "deactivated_component",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site component %s points to deactivated component '%s'", slotName, componentName),
			SpecJSON:     string(spec),
			Priority:     5,
			HandlerAgent: "rerender-pages",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("deactivated_%s", slotName),
			BatchID:      dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "deactivated_site_components",
			"count": len(result.WorkItems),
		})
	}

	return result, nil
}

// --- stale_site_components ---
// Detects site_components.rendered_html that hasn't been updated since page
// content was last generated. This means the header/footer is out of date —
// nav items may have changed, company name may have changed, template may
// have been upgraded.
// Threshold: site_component older than newest page_component by > 24 hours.

type StaleSiteComponentsCheck struct{}

func (c *StaleSiteComponentsCheck) Name() string { return "stale_site_components" }

func (c *StaleSiteComponentsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name, sc.updated_at, max_pc.latest_update
		FROM site_components sc
		CROSS JOIN LATERAL (
			SELECT MAX(pc.updated_at) as latest_update
			FROM page_components pc
			JOIN pages p ON p.id = pc.page_id
			WHERE p.site_id = $1 AND pc.rendered_html IS NOT NULL
		) max_pc
		WHERE sc.site_id = $1
		  AND sc.rendered_html IS NOT NULL
		  AND max_pc.latest_update IS NOT NULL
		  AND sc.updated_at < max_pc.latest_update - INTERVAL '24 hours'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("stale_site_components: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotName string
		var scUpdated, pcUpdated interface{}
		if err := rows.Scan(&slotName, &scUpdated, &pcUpdated); err != nil {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"slot_name":               slotName,
			"reason":                  "stale_render",
			"refresh_site_components": true,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "needs_rerender",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site component %s is stale — last rendered before page content was updated", slotName),
			SpecJSON:     string(spec),
			Priority:     8,
			HandlerAgent: "rerender-pages",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("stale_sc_%s", slotName),
			BatchID:      dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "stale_site_components",
			"count": len(result.WorkItems),
		})
	}

	return result, nil
}

// --- shared_css_theme ---
// Detects when multiple sites share the same style_collection_id.
// Each site should have its own style collection — shared collections mean a
// design change for one site affects all others. This is different from using
// the same *template* (which is fine) — it's sharing the same *instance* with
// the same CSS theme, colours, and component references.

type SharedCSSThemeCheck struct{}

func (c *SharedCSSThemeCheck) Name() string { return "shared_css_theme" }

func (c *SharedCSSThemeCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	var themeCount int

	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT s.domain, COUNT(s2.id)
		FROM sites s
		JOIN sites s2 ON s2.style_collection_id = s.style_collection_id
		                AND s2.id != s.id
		WHERE s.id = $1
		  AND s.style_collection_id IS NOT NULL
		GROUP BY s.domain
	`, dctx.SiteID).Scan(&domain, &themeCount)
	if err != nil {
		// No rows = not shared — fine
		return result, nil
	}

	if themeCount > 0 {
		spec, _ := json.Marshal(map[string]interface{}{
			"domain":      domain,
			"shared_with": themeCount,
			"reason":      "shared_style_collection",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "needs_design_review",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site %s shares its style collection with %d other site(s) — needs its own collection", domain, themeCount),
			SpecJSON:     string(spec),
			Priority:     7,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("shared_style_%s", domain),
			BatchID:      dctx.BatchID,
		})

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":       "shared_css_theme",
			"domain":      domain,
			"shared_with": themeCount,
		})
	}

	return result, nil
}

// ============================================================================
// COMPLETENESS-DISCOVERY CHECKS
// Wire into completeness-discovery-agent alongside empty_sections, empty_blog.
// ============================================================================

// --- cross_site_contamination ---
// Detects when rendered HTML contains another site's company name.
// Checks both site_components (header/footer) and page_components (sections).
// This catches LLM context bleed during concurrent content generation.

type CrossSiteContaminationCheck struct{}

func (c *CrossSiteContaminationCheck) Name() string { return "cross_site_contamination" }

func (c *CrossSiteContaminationCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// Check site_components for other sites' company names
	scRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name, s2.domain, s2.company_name
		FROM site_components sc
		JOIN sites s1 ON s1.id = sc.site_id
		CROSS JOIN sites s2
		WHERE s1.id = $1
		  AND s1.id != s2.id
		  AND s2.company_name IS NOT NULL AND s2.company_name != ''
		  AND sc.rendered_html IS NOT NULL
		  AND sc.rendered_html LIKE '%' || s2.company_name || '%'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("cross_site_contamination site_components: %w", err)
	}
	defer scRows.Close()

	for scRows.Next() {
		var slotName, foreignDomain, foreignCompany string
		if err := scRows.Scan(&slotName, &foreignDomain, &foreignCompany); err != nil {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"slot_name":               slotName,
			"location":                "site_component",
			"foreign_domain":          foreignDomain,
			"foreign_company":         foreignCompany,
			"refresh_site_components": true,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "cross_site_contamination",
			Severity:     "high",
			Summary:      fmt.Sprintf("Site component %s contains '%s' from %s", slotName, foreignCompany, foreignDomain),
			SpecJSON:     string(spec),
			Priority:     2,
			HandlerAgent: "rerender-pages",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("contamination_sc_%s_%s", slotName, foreignDomain),
			BatchID:      dctx.BatchID,
		})
	}

	// Check page_components
	pcRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT p.name, pc.slot_name, s2.domain, s2.company_name
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		JOIN sites s1 ON s1.id = p.site_id
		CROSS JOIN sites s2
		WHERE s1.id = $1
		  AND s1.id != s2.id
		  AND s2.company_name IS NOT NULL AND s2.company_name != ''
		  AND pc.rendered_html IS NOT NULL
		  AND pc.rendered_html LIKE '%' || s2.company_name || '%'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("cross_site_contamination page_components: %w", err)
	}
	defer pcRows.Close()

	for pcRows.Next() {
		var pageName, slotName, foreignDomain, foreignCompany string
		if err := pcRows.Scan(&pageName, &slotName, &foreignDomain, &foreignCompany); err != nil {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"page_name":       pageName,
			"slot_name":       slotName,
			"reason":          "cross_site_contamination",
			"foreign_domain":  foreignDomain,
			"foreign_company": foreignCompany,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "content_rewrite",
			Severity:     "high",
			Summary:      fmt.Sprintf("Page %s section %s contains '%s' from %s", pageName, slotName, foreignCompany, foreignDomain),
			SpecJSON:     string(spec),
			Priority:     3,
			HandlerAgent: "page-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("contamination_pc_%s_%s_%s", pageName, slotName, foreignDomain),
			BatchID:      dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "cross_site_contamination",
			"count": len(result.WorkItems),
		})
	}

	dctx.Logger.Info("CrossSiteContaminationCheck complete",
		zap.Int("findings", len(result.WorkItems)))

	return result, nil
}

// --- unrendered_templates ---
// Detects raw {{.field}} or {{range}} syntax in stored rendered HTML.
// In site_components this means the template was stored without rendering
// (usually because style_collection_id was NULL when the render ran).
// In page_components it means a template variable wasn't resolved.

type UnrenderedTemplatesCheck struct{}

func (c *UnrenderedTemplatesCheck) Name() string { return "unrendered_templates" }

func (c *UnrenderedTemplatesCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// site_components with raw templates
	scRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name
		FROM site_components sc
		WHERE sc.site_id = $1 AND sc.rendered_html LIKE '%{{%'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("unrendered_templates site_components: %w", err)
	}
	defer scRows.Close()

	for scRows.Next() {
		var slotName string
		if err := scRows.Scan(&slotName); err != nil {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"slot_name":               slotName,
			"location":                "site_component",
			"refresh_site_components": true,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "unrendered_template",
			Severity:     "high",
			Summary:      fmt.Sprintf("Site component %s has unrendered Go template syntax", slotName),
			SpecJSON:     string(spec),
			Priority:     2,
			HandlerAgent: "rerender-pages",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("unrendered_sc_%s", slotName),
			BatchID:      dctx.BatchID,
		})
	}

	// page_components with raw templates
	pcRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, pc.slot_name
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1 AND pc.rendered_html LIKE '%{{%'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("unrendered_templates page_components: %w", err)
	}
	defer pcRows.Close()

	for pcRows.Next() {
		var pageName, slotName string
		if err := pcRows.Scan(&pageName, &slotName); err != nil {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"page_name": pageName,
			"slot_name": slotName,
			"reason":    "unrendered_template",
			"location":  "page_component",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "content_rewrite",
			Severity:     "high",
			Summary:      fmt.Sprintf("Page %s section %s has unrendered template syntax", pageName, slotName),
			SpecJSON:     string(spec),
			Priority:     3,
			HandlerAgent: "page-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("unrendered_pc_%s_%s", pageName, slotName),
			BatchID:      dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "unrendered_templates",
			"count": len(result.WorkItems),
		})
	}

	dctx.Logger.Info("UnrenderedTemplatesCheck complete",
		zap.Int("findings", len(result.WorkItems)))

	return result, nil
}
