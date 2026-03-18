// FILE: platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go
//
// CHANGE: Added pc.locked_at IS NULL to skip locked components.

package discovery_checks

import (
	"encoding/json"
	"fmt"
)

func init() { Register(&HardcodedSectionColorsCheck{}) }

type HardcodedSectionColorsCheck struct{}

func (c *HardcodedSectionColorsCheck) Name() string { return "hardcoded_section_colors" }

func (c *HardcodedSectionColorsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	count, err := countHardcodedColorComponents(dctx)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return &CheckResult{}, nil
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":            "hardcoded_section_colors",
		"components_found": count,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "hardcoded_section_colors",
			"components_found": count,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "design",
			ItemType:     "hardcoded_section_colors",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Found %d components with hardcoded hex colors in inline styles instead of CSS variables", count),
			SpecJSON:     string(specJSON),
			Priority:     55,
			HandlerAgent: "color-variable-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "hardcoded_section_colors",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

func countHardcodedColorComponents(dctx DiscoveryCheckContext) (int, error) {
	var count int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
		  AND pc.rendered_html LIKE '%<style%'
	`, dctx.SiteID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("hardcoded color count query failed: %w", err)
	}
	return count, nil
}
