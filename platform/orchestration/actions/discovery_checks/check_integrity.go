// FILE: platform/orchestration/actions/discovery_checks/integrity_checks.go
//
// Algorithmic discovery checks for site integrity issues.
// Registered via init() into the existing discovery check registry.
// Run by completeness-discovery-agent via run_discovery_checks action.
//
// Checks:
//   cross_site_contamination   — another site's company name in rendered HTML
//   unrendered_templates       — raw {{.field}} syntax in stored HTML
//   missing_style_collection   — site has no style_collection_id
//   deactivated_site_components — site_components pointing to inactive components

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() {
	Register(&CrossSiteContaminationCheck{})
	Register(&UnrenderedTemplatesCheck{})
	Register(&MissingStyleCollectionCheck{})
	Register(&DeactivatedSiteComponentsCheck{})
}

// ============================================================================
// Check: cross_site_contamination
// ============================================================================

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
			"slot_name":       slotName,
			"location":        "site_component",
			"foreign_domain":  foreignDomain,
			"foreign_company": foreignCompany,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "structural",
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

// ============================================================================
// Check: unrendered_templates
// ============================================================================

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
			"slot_name": slotName,
			"location":  "site_component",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "structural",
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

// ============================================================================
// Check: missing_style_collection
// ============================================================================

type MissingStyleCollectionCheck struct{}

func (c *MissingStyleCollectionCheck) Name() string { return "missing_style_collection" }

func (c *MissingStyleCollectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	var hasCollection bool

	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT s.domain, s.style_collection_id IS NOT NULL
		FROM sites s WHERE s.id = $1
	`, dctx.SiteID).Scan(&domain, &hasCollection)
	if err != nil {
		return nil, fmt.Errorf("missing_style_collection: %w", err)
	}

	if !hasCollection {
		spec, _ := json.Marshal(map[string]interface{}{
			"domain": domain,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "structural",
			ItemType:     "missing_style_collection",
			Severity:     "high",
			Summary:      fmt.Sprintf("Site %s has no style collection assigned", domain),
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
		})
	}

	return result, nil
}

// ============================================================================
// Check: deactivated_site_components
// ============================================================================

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
			"slot_name":      slotName,
			"component_name": componentName,
			"component_id":   componentID,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "structural",
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
