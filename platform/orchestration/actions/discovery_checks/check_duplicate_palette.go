package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&DuplicatePaletteCheck{}) }

type DuplicatePaletteCheck struct{}

func (c *DuplicatePaletteCheck) Name() string { return "duplicate_palette" }

func (c *DuplicatePaletteCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	dupes, err := findDuplicatePalette(dctx)
	if err != nil {
		return nil, err
	}
	if len(dupes) == 0 {
		return &CheckResult{}, nil
	}

	matchingDomains := make([]string, len(dupes))
	for i, d := range dupes {
		matchingDomains[i] = d.MatchingDomain
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":            "duplicate_palette",
		"matching_domains": matchingDomains,
		"matches":          dupes,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "duplicate_palette",
			"matching_domains": matchingDomains,
			"matches":          dupes,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "design",
			ItemType:     "duplicate_palette",
			Severity:     "low",
			Summary:      fmt.Sprintf("Colour palette identical to %d other site(s)", len(dupes)),
			SpecJSON:     string(specJSON),
			Priority:     150,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "duplicate_palette",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

type duplicatePaletteFinding struct {
	MatchingDomain string `json:"matching_domain"`
	Primary        string `json:"primary"`
	Secondary      string `json:"secondary"`
	Accent         string `json:"accent"`
}

func findDuplicatePalette(dctx DiscoveryCheckContext) ([]duplicatePaletteFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT b.domain,
		       ct_a.color_palette->>'primary',
		       ct_a.color_palette->>'secondary',
		       ct_a.color_palette->>'accent'
		FROM sites a
		JOIN style_collections sc_a ON a.style_collection_id = sc_a.id
		JOIN css_themes ct_a ON sc_a.css_theme_id = ct_a.id
		JOIN sites b ON b.id != a.id AND b.status IN ('active', 'deployed')
		JOIN style_collections sc_b ON b.style_collection_id = sc_b.id
		JOIN css_themes ct_b ON sc_b.css_theme_id = ct_b.id
		WHERE a.id = $1
		  AND ct_a.color_palette->>'primary' = ct_b.color_palette->>'primary'
		  AND ct_a.color_palette->>'secondary' = ct_b.color_palette->>'secondary'
		  AND ct_a.color_palette->>'accent' = ct_b.color_palette->>'accent'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("duplicate_palette query failed: %w", err)
	}
	defer rows.Close()

	var findings []duplicatePaletteFinding
	for rows.Next() {
		var f duplicatePaletteFinding
		if err := rows.Scan(&f.MatchingDomain, &f.Primary, &f.Secondary, &f.Accent); err != nil {
			dctx.Logger.Warn("Failed to scan duplicate palette", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
