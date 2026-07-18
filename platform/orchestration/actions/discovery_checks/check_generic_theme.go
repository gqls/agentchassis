package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"strings"
)

func init() { Register(&GenericThemeCheck{}) }

type GenericThemeCheck struct{}

func (c *GenericThemeCheck) Name() string { return "generic_theme" }

func (c *GenericThemeCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	finding, err := findGenericTheme(dctx)
	if err != nil {
		return nil, err
	}
	if finding == nil {
		return &CheckResult{}, nil
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":   "generic_theme",
		"finding": finding,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "generic_theme",
			"finding": finding,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "generic_theme",
			Severity:     "medium",
			Summary:      "Site using default theme — no industry-specific styling",
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "generic_theme",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

type genericThemeFinding struct {
	HasWebdesignSpec bool   `json:"has_webdesign_spec"`
	HasIdentitySpec  bool   `json:"has_identity_spec"`
	CSSLength        int    `json:"css_length"`
	UsesDefaultColor bool   `json:"uses_default_color"`
	Detail           string `json:"detail"`
}

func findGenericTheme(dctx DiscoveryCheckContext) (*genericThemeFinding, error) {
	// No-collection sites are owned by missing_style_collection, which emits
	// the composition pair (needs_composition -> needs_design). Skip them here
	// so generic_theme doesn't also queue a bare webdesign-agent run that would
	// race ahead of composition. generic_theme handles only the "collection
	// exists but the theme looks default/bland" case.
	var hasCollection bool
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT style_collection_id IS NOT NULL FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&hasCollection); err != nil {
		return nil, err
	}
	if !hasCollection {
		return nil, nil
	}

	finding := &genericThemeFinding{}

	// Check for webdesign spec. Two storage conventions exist:
	//   - site_specs aspect='webdesign' — this check's original contract,
	//     which NO code path has ever written (0 rows fleet-wide, 2026-07-17);
	//   - sites.content_data.color_scheme — what the webdesign-agent's
	//     update_site step (update_site_content, merge=true,
	//     content_field=design_spec.result) actually stores.
	// Testing only the former made this check fire on every themed site
	// every discovery pass, each time dispatching webdesign-agent — whose
	// analyze_design LLM re-rolls the palette, drifting site colours run
	// over run (robot-hands R1, 2026-07-17: four CSS rewrites in a day,
	// one rolled a light background onto a dark site).
	var webdesignCount int
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM site_specs
		WHERE site_id = $1 AND aspect = 'webdesign' AND is_current = true
	`, dctx.SiteID).Scan(&webdesignCount)
	if webdesignCount == 0 {
		dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT COUNT(*) FROM sites
			WHERE id = $1 AND content_data ? 'color_scheme'
		`, dctx.SiteID).Scan(&webdesignCount)
	}
	finding.HasWebdesignSpec = webdesignCount > 0

	// Check for identity spec (needed by webdesign)
	var identityCount int
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM site_specs 
		WHERE site_id = $1 AND aspect = 'identity' AND is_current = true
	`, dctx.SiteID).Scan(&identityCount)
	finding.HasIdentitySpec = identityCount > 0

	// Check actual CSS for default values
	var cssHTML sql.NullString
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT rendered_html FROM site_components
		WHERE site_id = $1 AND slot_name = 'head'
	`, dctx.SiteID).Scan(&cssHTML)

	if cssHTML.Valid {
		finding.CSSLength = len(cssHTML.String)
		// Default fallback primary color from base stylesheet
		if strings.Contains(cssHTML.String, "--color-primary: #2c3e50") ||
			strings.Contains(cssHTML.String, "--color-primary: #333") {
			finding.UsesDefaultColor = true
		}
	}

	// Determine if this is a problem
	isGeneric := false
	if !finding.HasWebdesignSpec {
		finding.Detail = "No webdesign spec in site_specs — agent never produced themed CSS"
		isGeneric = true
	} else if finding.UsesDefaultColor {
		finding.Detail = "CSS uses default fallback colours — webdesign may have had no identity context"
		isGeneric = true
	} else if finding.CSSLength == 0 {
		finding.Detail = "Head component has no CSS content"
		isGeneric = true
	}

	if !isGeneric {
		return nil, nil
	}

	return finding, nil
}
