package discovery_checks

import (
	"encoding/json"
	"fmt"
)

func init() { Register(&ForcedTextColorsCheck{}) }

type ForcedTextColorsCheck struct{}

func (c *ForcedTextColorsCheck) Name() string { return "forced_text_colors" }

func (c *ForcedTextColorsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	count, err := countForcedTextColorIssues(dctx)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return &CheckResult{}, nil
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":            "forced_text_colors",
		"components_found": count,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "forced_text_colors",
			"components_found": count,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "design",
			ItemType:     "forced_text_colors",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Found %d sections with dark backgrounds missing light text color overrides", count),
			SpecJSON:     string(specJSON),
			Priority:     56,
			HandlerAgent: "forced-text-color-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "forced_text_colors",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// countForcedTextColorIssues finds page_components that have dark background
// colors (in <style> blocks) but no corresponding light text color override.
// These sections render as dark-on-dark text — unreadable.
func countForcedTextColorIssues(dctx DiscoveryCheckContext) (int, error) {
	var count int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html LIKE '%<style%'
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-4][0-9a-fA-F]{5}'
		  AND pc.rendered_html NOT SIMILAR TO '%color:\s*(#[f-fF-F][0-9a-fA-F]{5}|#[e-fE-F][0-9a-fA-F]{5}|white|#fff)%'
	`, dctx.SiteID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("forced_text_colors count query failed: %w", err)
	}
	return count, nil
}
