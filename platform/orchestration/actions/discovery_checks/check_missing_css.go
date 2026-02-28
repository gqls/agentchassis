package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func init() { Register(&MissingCSSCheck{}) }

type MissingCSSCheck struct{}

func (c *MissingCSSCheck) Name() string { return "missing_css" }

func (c *MissingCSSCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	hasMissing, detail, err := findMissingCSS(dctx)
	if err != nil {
		return nil, err
	}
	if !hasMissing {
		return &CheckResult{}, nil
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":  "missing_css",
		"detail": detail,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":  "missing_css",
			"detail": detail,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Domain:       "design",
			ItemType:     "missing_css",
			Severity:     "high",
			Summary:      "Site has no custom stylesheet — using raw defaults",
			SpecJSON:     string(specJSON),
			Priority:     50,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "missing_css",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

func findMissingCSS(dctx DiscoveryCheckContext) (bool, map[string]interface{}, error) {
	var noStyleCollection, noCSSTheme bool
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT
		    s.style_collection_id IS NULL,
		    ct.id IS NULL
		FROM sites s
		LEFT JOIN style_collections sc ON s.style_collection_id = sc.id
		LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id
		WHERE s.id = $1
	`, dctx.SiteID).Scan(&noStyleCollection, &noCSSTheme)

	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("missing_css query failed: %w", err)
	}

	if !noStyleCollection && !noCSSTheme {
		return false, nil, nil
	}

	detail := map[string]interface{}{
		"no_style_collection": noStyleCollection,
		"no_css_theme":        noCSSTheme,
	}
	return true, detail, nil
}
