package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&BrokenNavLinksCheck{}) }

type BrokenNavLinksCheck struct{}

func (c *BrokenNavLinksCheck) Name() string { return "broken_nav_links" }

func (c *BrokenNavLinksCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	broken, err := findBrokenNavLinks(dctx)
	if err != nil {
		return nil, err
	}
	if len(broken) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":    "broken_nav_links",
			"count":    len(broken),
			"findings": broken,
		}},
	}

	for _, finding := range broken {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":        "broken_nav_links",
			"slot_name":    finding.SlotName,
			"link_count":   finding.LinkCount,
			"example_href": finding.ExampleHref,
			"fix": "Template uses #{{.slug}} — should use {{.url}}. " +
				"Fix template in content_components, then force re-render site_components.",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "build",
			ItemType:     "broken_nav_links",
			Severity:     "high",
			Summary:      fmt.Sprintf("Navigation in %s uses anchor links (#slug) instead of page URLs", finding.SlotName),
			SpecJSON:     string(specJSON),
			Priority:     40,
			HandlerAgent: "nav-link-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("broken_nav_links:%s", finding.SlotName),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

type brokenNavLinkFinding struct {
	SlotName    string `json:"slot_name"`
	LinkCount   int    `json:"link_count"`
	ExampleHref string `json:"example_href"`
}

func findBrokenNavLinks(dctx DiscoveryCheckContext) ([]brokenNavLinkFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name,
		       (LENGTH(sc.rendered_html) - LENGTH(REPLACE(sc.rendered_html, 'href="#', ''))) 
		           / LENGTH('href="#') as link_count,
		       SUBSTRING(sc.rendered_html FROM 'href="(#[a-zA-Z][^"]*)"') as example_href
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.slot_name IN ('header', 'footer')
		  AND sc.rendered_html IS NOT NULL
		  AND sc.rendered_html ~ 'href="#[a-zA-Z]'
		ORDER BY sc.slot_name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("broken_nav_links query failed: %w", err)
	}
	defer rows.Close()

	var findings []brokenNavLinkFinding
	for rows.Next() {
		var f brokenNavLinkFinding
		var exampleHref sql.NullString
		if err := rows.Scan(&f.SlotName, &f.LinkCount, &exampleHref); err != nil {
			dctx.Logger.Warn("Failed to scan broken nav link", zap.Error(err))
			continue
		}
		if exampleHref.Valid {
			f.ExampleHref = exampleHref.String
		}
		findings = append(findings, f)
	}
	return findings, nil
}
