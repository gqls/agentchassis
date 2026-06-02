package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func init() { Register(&MissingCSSCheck{}) }

type MissingCSSCheck struct{}

func (c *MissingCSSCheck) Name() string { return "missing_css" }

// Run now handles ONLY the case where a style collection is installed but its
// css_theme is missing — regenerate the stylesheet against the existing
// collection via webdesign-agent.
//
// The no-collection case (style_collection_id IS NULL) is owned by
// missing_style_collection, which emits the composition pair
// (needs_composition → site-design-planner, then needs_design →
// webdesign-agent). missing_css used to also emit that pair here, which raced
// missing_style_collection (priority 1) and webdesign-agent ahead of
// composition; it now defers the no-collection case entirely.
func (c *MissingCSSCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	hasMissing, detail, err := findMissingCSS(dctx)
	if err != nil {
		return nil, err
	}
	if !hasMissing {
		return &CheckResult{}, nil
	}

	// Defer the no-collection case to missing_style_collection.
	if noStyleCollection, _ := detail["no_style_collection"].(bool); noStyleCollection {
		return &CheckResult{}, nil
	}

	// Collection exists but its css_theme is missing — regenerate CSS against
	// the existing collection. Routing composition here would fail the install
	// guard (install_site_composition refuses when style_collection_id is set),
	// so go straight to webdesign-agent.
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
			Pipeline:     "design",
			ItemType:     "missing_css",
			Severity:     "high",
			Summary:      "Style collection has no css_theme — regenerate stylesheet",
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
