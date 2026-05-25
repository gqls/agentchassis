package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func init() { Register(&MissingCSSCheck{}) }

type MissingCSSCheck struct{}

func (c *MissingCSSCheck) Name() string { return "missing_css" }

// Run emits design work for sites whose stylesheet is missing.
//
// CHANGE (loop now mirrors the build path): when a site has NO composition
// installed (style_collection_id IS NULL), the check emits the same pair the
// planner's emit_design_items step emits — needs_composition (→ site-design-
// planner) ahead of needs_design (→ webdesign-agent) — so a stale site gets a
// resolved palette/layout/typography composition, not just emergency-fallback
// CSS. site-design-planner's install_site_composition sets style_collection_id,
// which is what stops missing_css re-firing on the next loop.
//
// When a collection EXISTS but its css_theme is missing (collection present,
// theme NULL), re-resolving composition would fail — install_site_composition
// refuses when style_collection_id is already set — so we keep the original
// behaviour and route straight to webdesign-agent to regenerate CSS against
// the existing collection.
//
// Ordering: discovery items are emitted at status 'detected' (triaged before
// dispatch). Composition is priority 7, design priority 8; LoadWorkItemsAction
// orders priority ASC and find_dispatchable_site guarantees a single dispatch
// loop per site, so composition is claimed and completed before design. The
// discovery WorkItemSpec path has no depends_on field, so ordering relies on
// priority rather than an explicit dependency — if composition fails, design
// runs and falls back, and missing_css re-fires next loop to retry. Adding a
// DependsOnItemKey to WorkItemSpec (resolved post-insert in run_discovery_checks)
// would give the same hard gate the planner's emit_design_items has; deferred.
func (c *MissingCSSCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	hasMissing, detail, err := findMissingCSS(dctx)
	if err != nil {
		return nil, err
	}
	if !hasMissing {
		return &CheckResult{}, nil
	}

	noStyleCollection, _ := detail["no_style_collection"].(bool)

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":  "missing_css",
			"detail": detail,
		}},
	}

	if noStyleCollection {
		// No composition installed → emit the planner's trigger pair.
		compSpec, _ := json.Marshal(map[string]interface{}{
			"check":  "missing_css",
			"stage":  "composition",
			"detail": detail,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "design",
			ItemType:     "needs_composition",
			Severity:     "high",
			Summary:      "No composition installed — resolve palette/layout/typography",
			SpecJSON:     string(compSpec),
			Priority:     7,
			HandlerAgent: "site-design-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "needs_composition",
			BatchID:      dctx.BatchID,
		})

		designSpec, _ := json.Marshal(map[string]interface{}{
			"check": "missing_css",
			"stage": "design",
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "design",
			ItemType:     "needs_design",
			Severity:     "high",
			Summary:      "Generate site stylesheet (after composition)",
			SpecJSON:     string(designSpec),
			Priority:     8,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "needs_design",
			BatchID:      dctx.BatchID,
		})

		return result, nil
	}

	// Collection exists but its css_theme is missing — regenerate CSS against
	// the existing collection. Routing composition here would fail the install
	// guard, so go straight to webdesign-agent (original behaviour).
	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":  "missing_css",
		"detail": detail,
	})
	result.WorkItems = append(result.WorkItems, WorkItemSpec{
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
	})

	return result, nil
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
