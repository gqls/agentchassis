// FILE: platform/orchestration/actions/discovery_checks.go
//
// Discovery check functions and action for the unified build/maintenance system.
// Writes findings to site_work_items (not maintenance_queue).
//
// Action:
//   RunDiscoveryChecksAction — runs configurable checks per site,
//                              inserts findings as work items with status='detected'
//
// Check functions (pure queries, no side effects):
//   findEmptySections      — deployed page_components with null/empty rendered_html
//   findUndeployedAssets   — assets in DB not referenced in any deployed page HTML
//   findMissingCSS         — no style_collection or css_theme linked
//   findDuplicatePalette   — same colour palette as another active site
//
// Discovery agents write items. They do not fix anything. They do not call
// other agents. Findings arrive as status='detected', enriched by triage.
//
// Integration:
//   - site-work-orchestrator step 4 (run_discovery) spawns discovery agents
//   - maintenance-batch-scheduler can also trigger standalone
//   - Existing maintenance-triage + maintenance_queue continue working unchanged

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var RunDiscoveryChecksInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("run_discovery_checks", RunDiscoveryChecksInputSpec)
}

// ============================================================================
// ACTION: run_discovery_checks
// Used by: design-discovery-agent, completeness-discovery-agent,
//          site-work-orchestrator (discovery phase)
// ============================================================================
//
// Runs configurable checks for a single site and inserts findings into
// site_work_items with source='discovery', status='detected'.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — resolved from collectedData
//
// Config literals (read directly from step config):
//   - checks: []string — which checks to run
//     Available: "empty_sections", "undeployed_assets", "missing_css", "duplicate_palette"
//   - check_domain: string — the work item domain for findings (e.g. "design", "content")
//     Named check_domain to avoid collision with site_record.domain

func RunDiscoveryChecksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("RunDiscoveryChecksAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- site_id needs path resolution from collectedData ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		RunDiscoveryChecksInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// --- Config literals (not paths) ---
	config := params.StepConfig.Config
	checkDomain, _ := config["check_domain"].(string)
	if checkDomain == "" {
		checkDomain = "design"
	}

	checks := []string{"empty_sections", "undeployed_assets", "missing_css", "duplicate_palette"}
	if configChecks, ok := config["checks"].([]interface{}); ok && len(configChecks) > 0 {
		checks = make([]string, len(configChecks))
		for i, c := range configChecks {
			checks[i] = fmt.Sprintf("%v", c)
		}
	}

	// --- Look up domain name for logging ---
	var siteDomain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID,
	).Scan(&siteDomain)

	logger.Info("RunDiscoveryChecksAction: Running checks",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.Strings("checks", checks),
	)

	// --- Run checks and collect findings ---
	batchID := uuid.New()
	agentType := params.ExecutionContext.Sender.AgentType
	if agentType == "" {
		agentType = "discovery-agent"
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var allFindings []interface{}
	inserted := 0
	skipped := 0

	// --- empty_sections → handler: page-content-writer ---
	if containsCheck(checks, "empty_sections") {
		emptySections, err := findEmptySections(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("empty_sections check failed", zap.Error(err))
		} else if len(emptySections) > 0 {
			allFindings = append(allFindings, map[string]interface{}{
				"check":      "empty_sections",
				"count":      len(emptySections),
				"components": emptySections,
			})

			// One work item per empty section (granular — per v6 "per-section for maintenance")
			for _, section := range emptySections {
				specJSON, _ := json.Marshal(map[string]interface{}{
					"check":              "empty_sections",
					"component_id":       section.ComponentID,
					"page_id":            section.PageID,
					"page_name":          section.PageName,
					"slot_name":          section.SlotName,
					"component_function": section.ComponentFunction,
					"empty_pattern":      section.EmptyPattern,
				})

				var pageIDPtr *uuid.UUID
				if parsed, err := uuid.Parse(section.PageID); err == nil {
					pageIDPtr = &parsed
				}

				ok, err := insertWorkItem(ctx, tx, workItem{
					siteID:       siteID,
					source:       "discovery",
					domain:       "content",
					itemType:     "empty_section",
					severity:     "medium",
					summary:      fmt.Sprintf("Empty section '%s' on page %s", section.SlotName, section.PageName),
					spec:         string(specJSON),
					pageID:       pageIDPtr,
					priority:     100, // default — triage adjusts
					handlerAgent: "page-content-writer",
					status:       "detected",
					createdBy:    agentType,
					itemKey:      fmt.Sprintf("empty_section:%s:%s", section.PageID, section.SlotName),
					batchID:      batchID,
				}, logger)
				if err != nil {
					logger.Warn("Failed to insert empty_section item", zap.Error(err))
				} else if ok {
					inserted++
				} else {
					skipped++
				}
			}
		}
	}

	// --- undeployed_assets → handler: asset-deployer ---
	if containsCheck(checks, "undeployed_assets") {
		undeployed, err := findUndeployedAssets(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("undeployed_assets check failed", zap.Error(err))
		} else if len(undeployed) > 0 {
			allFindings = append(allFindings, map[string]interface{}{
				"check":  "undeployed_assets",
				"count":  len(undeployed),
				"assets": undeployed,
			})

			// One item per asset
			for _, asset := range undeployed {
				specJSON, _ := json.Marshal(map[string]interface{}{
					"check":      "undeployed_assets",
					"asset_id":   asset.AssetID,
					"purpose":    asset.Purpose,
					"asset_type": asset.AssetType,
					"url":        asset.URL,
				})

				ok, err := insertWorkItem(ctx, tx, workItem{
					siteID:       siteID,
					source:       "discovery",
					domain:       "design",
					itemType:     "undeployed_asset",
					severity:     "high",
					summary:      fmt.Sprintf("Asset '%s' generated but not deployed to site", asset.Purpose),
					spec:         string(specJSON),
					priority:     60, // visible issue — higher than cosmetic
					handlerAgent: "asset-deployer",
					status:       "detected",
					createdBy:    agentType,
					itemKey:      fmt.Sprintf("undeployed_asset:%s", asset.AssetID),
					batchID:      batchID,
				}, logger)
				if err != nil {
					logger.Warn("Failed to insert undeployed_asset item", zap.Error(err))
				} else if ok {
					inserted++
				} else {
					skipped++
				}
			}
		}
	}

	// --- missing_css → handler: webdesign-agent ---
	if containsCheck(checks, "missing_css") {
		hasMissing, detail, err := findMissingCSS(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("missing_css check failed", zap.Error(err))
		} else if hasMissing {
			allFindings = append(allFindings, map[string]interface{}{
				"check":  "missing_css",
				"detail": detail,
			})

			specJSON, _ := json.Marshal(map[string]interface{}{
				"check":  "missing_css",
				"detail": detail,
			})

			// Site-wide item (no page_id)
			ok, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "discovery",
				domain:       "design",
				itemType:     "missing_css",
				severity:     "high",
				summary:      "Site has no custom stylesheet — using raw defaults",
				spec:         string(specJSON),
				priority:     50, // site looks wrong without CSS
				handlerAgent: "webdesign-agent",
				status:       "detected",
				createdBy:    agentType,
				itemKey:      "missing_css",
				batchID:      batchID,
			}, logger)
			if err != nil {
				logger.Warn("Failed to insert missing_css item", zap.Error(err))
			} else if ok {
				inserted++
			} else {
				skipped++
			}
		}
	}

	// --- duplicate_palette → handler: webdesign-agent ---
	if containsCheck(checks, "duplicate_palette") {
		dupes, err := findDuplicatePalette(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("duplicate_palette check failed", zap.Error(err))
		} else if len(dupes) > 0 {
			matchingDomains := make([]string, len(dupes))
			for i, d := range dupes {
				matchingDomains[i] = d.MatchingDomain
			}
			allFindings = append(allFindings, map[string]interface{}{
				"check":            "duplicate_palette",
				"matching_domains": matchingDomains,
				"matches":          dupes,
			})

			specJSON, _ := json.Marshal(map[string]interface{}{
				"check":            "duplicate_palette",
				"matching_domains": matchingDomains,
				"matches":          dupes,
			})

			// Site-wide, cosmetic
			ok, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "discovery",
				domain:       "design",
				itemType:     "duplicate_palette",
				severity:     "low",
				summary:      fmt.Sprintf("Colour palette identical to %d other site(s)", len(dupes)),
				spec:         string(specJSON),
				priority:     150, // cosmetic — low urgency
				handlerAgent: "webdesign-agent",
				status:       "detected",
				createdBy:    agentType,
				itemKey:      "duplicate_palette",
				batchID:      batchID,
			}, logger)
			if err != nil {
				logger.Warn("Failed to insert duplicate_palette item", zap.Error(err))
			} else if ok {
				inserted++
			} else {
				skipped++
			}
		}
	}

	// --- hardcoded_section_colors → handler: color-variable-fixer ---
	if containsCheck(checks, "hardcoded_section_colors") {
		count, err := countHardcodedColorComponents(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("hardcoded_section_colors check failed", zap.Error(err))
		} else if count > 0 {
			allFindings = append(allFindings, map[string]interface{}{
				"check":            "hardcoded_section_colors",
				"components_found": count,
			})

			specJSON, _ := json.Marshal(map[string]interface{}{
				"check":            "hardcoded_section_colors",
				"components_found": count,
			})

			ok, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "discovery",
				domain:       "design",
				itemType:     "hardcoded_section_colors",
				severity:     "medium",
				summary:      fmt.Sprintf("Found %d components with hardcoded hex colors in inline styles instead of CSS variables", count),
				spec:         string(specJSON),
				priority:     55, // after missing_css (50) but before cosmetic fixes
				handlerAgent: "color-variable-fixer",
				status:       "detected",
				createdBy:    agentType,
				itemKey:      "hardcoded_section_colors",
				batchID:      batchID,
			}, logger)
			if err != nil {
				logger.Warn("Failed to insert hardcoded_section_colors item", zap.Error(err))
			} else if ok {
				inserted++
			} else {
				skipped++
			}
		}
	}

	if containsCheck(checks, "broken_nav_links") {
		broken, err := findBrokenNavLinks(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("broken_nav_links check failed", zap.Error(err))
		} else if len(broken) > 0 {
			allFindings = append(allFindings, map[string]interface{}{
				"check":    "broken_nav_links",
				"count":    len(broken),
				"findings": broken,
			})

			// One item per slot (header, footer)
			for _, finding := range broken {
				specJSON, _ := json.Marshal(map[string]interface{}{
					"check":        "broken_nav_links",
					"slot_name":    finding.SlotName,
					"link_count":   finding.LinkCount,
					"example_href": finding.ExampleHref,
					"fix": "Template uses #{{.slug}} — should use {{.url}}. " +
						"Fix template in content_components, then force re-render site_components.",
				})

				ok, err := insertWorkItem(ctx, tx, workItem{
					siteID:       siteID,
					source:       "discovery",
					domain:       "build",
					itemType:     "broken_nav_links",
					severity:     "high",
					summary:      fmt.Sprintf("Navigation in %s uses anchor links (#slug) instead of page URLs", finding.SlotName),
					spec:         string(specJSON),
					priority:     40, // navigation is broken — high priority
					handlerAgent: "nav-link-fixer",
					status:       "detected",
					createdBy:    agentType,
					itemKey:      fmt.Sprintf("broken_nav_links:%s", finding.SlotName),
					batchID:      batchID,
				}, logger)
				if err != nil {
					logger.Warn("Failed to insert broken_nav_links item", zap.Error(err))
				} else if ok {
					inserted++
				} else {
					skipped++
				}
			}
		}
	}

	if containsCheck(checks, "placeholder_contact") {
		placeholders, err := findPlaceholderContact(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("placeholder_contact check failed", zap.Error(err))
		} else if len(placeholders) > 0 {
			allFindings = append(allFindings, map[string]interface{}{
				"check":    "placeholder_contact",
				"count":    len(placeholders),
				"findings": placeholders,
			})

			// Group by page — one work item per affected page
			pageFindings := map[string][]placeholderContactFinding{}
			for _, f := range placeholders {
				pageFindings[f.PageID] = append(pageFindings[f.PageID], f)
			}

			for pageID, findings := range pageFindings {
				patterns := make([]string, len(findings))
				for i, f := range findings {
					patterns[i] = f.Pattern
				}

				specJSON, _ := json.Marshal(map[string]interface{}{
					"check":     "placeholder_contact",
					"page_id":   pageID,
					"page_name": findings[0].PageName,
					"patterns":  patterns,
					"findings":  findings,
				})

				var pageIDPtr *uuid.UUID
				if parsed, err := uuid.Parse(pageID); err == nil {
					pageIDPtr = &parsed
				}

				ok, err := insertWorkItem(ctx, tx, workItem{
					siteID:       siteID,
					source:       "discovery",
					domain:       "content",
					itemType:     "placeholder_contact",
					severity:     "high",
					summary:      fmt.Sprintf("Fabricated contact info on page %s (%d patterns)", findings[0].PageName, len(findings)),
					spec:         string(specJSON),
					pageID:       pageIDPtr,
					priority:     30, // fake contact info is a trust issue — very high
					handlerAgent: "page-content-writer",
					status:       "detected",
					createdBy:    agentType,
					itemKey:      fmt.Sprintf("placeholder_contact:%s", pageID),
					batchID:      batchID,
				}, logger)
				if err != nil {
					logger.Warn("Failed to insert placeholder_contact item", zap.Error(err))
				} else if ok {
					inserted++
				} else {
					skipped++
				}
			}
		}
	}

	if containsCheck(checks, "generic_theme") {
		finding, err := findGenericTheme(ctx, params.DB, siteID, logger)
		if err != nil {
			logger.Warn("generic_theme check failed", zap.Error(err))
		} else if finding != nil {
			allFindings = append(allFindings, map[string]interface{}{
				"check":   "generic_theme",
				"finding": finding,
			})

			specJSON, _ := json.Marshal(map[string]interface{}{
				"check":   "generic_theme",
				"finding": finding,
			})

			ok, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "discovery",
				domain:       "build",
				itemType:     "generic_theme",
				severity:     "medium",
				summary:      "Site using default theme — no industry-specific styling",
				spec:         string(specJSON),
				priority:     60,
				handlerAgent: "webdesign-agent",
				status:       "detected",
				createdBy:    agentType,
				itemKey:      "generic_theme",
				batchID:      batchID,
			}, logger)
			if err != nil {
				logger.Warn("Failed to insert generic_theme item", zap.Error(err))
			} else if ok {
				inserted++
			} else {
				skipped++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit discovery items: %w", err)
	}

	logger.Info("RunDiscoveryChecksAction: Complete",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.Int("items_inserted", inserted),
		zap.Int("items_skipped", skipped),
		zap.Int("findings", len(allFindings)),
	)

	return map[string]interface{}{
		"site_id":        siteIDStr,
		"domain":         siteDomain,
		"items_inserted": inserted,
		"items_skipped":  skipped,
		"findings":       allFindings,
		"batch_id":       batchID.String(),
		"checks_run":     checks,
	}, nil
}

// ============================================================================
// CHECK: empty_sections
// ============================================================================

// emptySectionFinding holds a detected empty section on a deployed page.
type emptySectionFinding struct {
	ComponentID       string `json:"component_id"`
	PageID            string `json:"page_id"`
	PageName          string `json:"page_name"`
	SlotName          string `json:"slot_name"`
	ComponentFunction string `json:"component_function"`
	HTMLLength        int    `json:"html_length"`
	EmptyPattern      string `json:"empty_pattern"`
}

// findEmptySections finds deployed page_components with null/empty rendered_html,
// excluding structural slots (header, footer, head).
func findEmptySections(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]emptySectionFinding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT pc.id, pc.page_id, p.name, COALESCE(pc.slot_name, ''),
		       COALESCE(cc.function, pc.slot_name, 'unknown'),
		       LENGTH(COALESCE(pc.rendered_html, '')),
		       CASE
		           WHEN pc.rendered_html IS NULL THEN 'null_html'
		           WHEN TRIM(pc.rendered_html) = '' THEN 'empty_html'
		           WHEN LENGTH(pc.rendered_html) < 50 THEN 'minimal_html'
		           WHEN pc.rendered_html ~* '<(h[1-6])[^>]*>\s*</\1>' THEN 'empty_heading'
		           WHEN pc.rendered_html ~* 'class="section[^"]*">\s*</div>' THEN 'empty_container'
		           ELSE 'near_empty'
		       END
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components cc ON pc.component_id = cc.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')
		  AND COALESCE(cc.function, '') NOT IN ('header', 'footer', 'head-seo')
		  AND (
		      pc.rendered_html IS NULL
		      OR TRIM(pc.rendered_html) = ''
		      OR LENGTH(pc.rendered_html) < 50
		      OR pc.rendered_html ~* '<(h[1-6])[^>]*>\s*</\1>'
		  )
		ORDER BY p.name, pc.position
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("empty_sections query failed: %w", err)
	}
	defer rows.Close()

	var findings []emptySectionFinding
	for rows.Next() {
		var f emptySectionFinding
		if err := rows.Scan(&f.ComponentID, &f.PageID, &f.PageName, &f.SlotName,
			&f.ComponentFunction, &f.HTMLLength, &f.EmptyPattern); err != nil {
			logger.Warn("Failed to scan empty section", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// CHECK: undeployed_assets
// ============================================================================

// undeployedAssetFinding holds an asset in the DB not referenced in page HTML.
type undeployedAssetFinding struct {
	AssetID   string `json:"asset_id"`
	Purpose   string `json:"purpose"`
	AssetType string `json:"asset_type"`
	URL       string `json:"url"`
}

// findUndeployedAssets finds assets in the assets table that aren't referenced
// in any deployed page_component's rendered_html for the site.
func findUndeployedAssets(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]undeployedAssetFinding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT a.id, COALESCE(a.purpose, 'unknown'), a.asset_type, a.url
		FROM assets a
		WHERE a.site_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM page_components pc
		      JOIN pages p ON pc.page_id = p.id
		      WHERE p.site_id = a.site_id
		        AND pc.build_status = 'deployed'
		        AND (
		            pc.rendered_html LIKE '%/assets/images/' || COALESCE(a.purpose, '') || '.%'
		            OR pc.rendered_html LIKE '%/assets/images/' || COALESCE(a.purpose, '') || '-%'
		        )
		  )
		ORDER BY a.purpose
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("undeployed_assets query failed: %w", err)
	}
	defer rows.Close()

	var findings []undeployedAssetFinding
	for rows.Next() {
		var f undeployedAssetFinding
		if err := rows.Scan(&f.AssetID, &f.Purpose, &f.AssetType, &f.URL); err != nil {
			logger.Warn("Failed to scan undeployed asset", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// CHECK: missing_css
// ============================================================================

// findMissingCSS checks if the site has no style_collection or no css_theme.
// Returns true if there's an issue, plus detail about what's missing.
func findMissingCSS(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (bool, map[string]interface{}, error) {
	var noStyleCollection, noCSSTheme bool
	err := db.QueryRowContext(ctx, `
		SELECT
		    s.style_collection_id IS NULL,
		    ct.id IS NULL
		FROM sites s
		LEFT JOIN style_collections sc ON s.style_collection_id = sc.id
		LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id
		WHERE s.id = $1
	`, siteID).Scan(&noStyleCollection, &noCSSTheme)

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

// ============================================================================
// CHECK: duplicate_palette
// ============================================================================

// duplicatePaletteFinding holds a colour collision with another site.
type duplicatePaletteFinding struct {
	MatchingDomain string `json:"matching_domain"`
	Primary        string `json:"primary"`
	Secondary      string `json:"secondary"`
	Accent         string `json:"accent"`
}

// findDuplicatePalette checks if this site shares identical primary/secondary/accent
// colours with any other active site via css_themes color_palette.
func findDuplicatePalette(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]duplicatePaletteFinding, error) {
	rows, err := db.QueryContext(ctx, `
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
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("duplicate_palette query failed: %w", err)
	}
	defer rows.Close()

	var findings []duplicatePaletteFinding
	for rows.Next() {
		var f duplicatePaletteFinding
		if err := rows.Scan(&f.MatchingDomain, &f.Primary, &f.Secondary, &f.Accent); err != nil {
			logger.Warn("Failed to scan duplicate palette", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// CHECK: broken_nav_links
// Detects href="#slug" patterns in rendered site_components (header/footer)
// that should be href="/slug.html" proper page URLs.
// Root cause: header/footer template uses #{{.slug}} instead of {{.url}}
// Handler: site-component-fixer (re-renders after template correction)
// ============================================================================

type brokenNavLinkFinding struct {
	SlotName    string `json:"slot_name"`
	LinkCount   int    `json:"link_count"`
	ExampleHref string `json:"example_href"`
}

func findBrokenNavLinks(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]brokenNavLinkFinding, error) {
	// Look for href="#something" in site_components rendered_html
	// These are anchor-style links that should be page links
	rows, err := db.QueryContext(ctx, `
		SELECT sc.slot_name,
		       -- Count href="#word" occurrences (not href="#" alone which is valid)
		       (LENGTH(sc.rendered_html) - LENGTH(REPLACE(sc.rendered_html, 'href="#', ''))) 
		           / LENGTH('href="#') as link_count,
		       -- Extract first example
		       SUBSTRING(sc.rendered_html FROM 'href="(#[a-zA-Z][^"]*)"') as example_href
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.slot_name IN ('header', 'footer')
		  AND sc.rendered_html IS NOT NULL
		  AND sc.rendered_html ~ 'href="#[a-zA-Z]'
		ORDER BY sc.slot_name
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("broken_nav_links query failed: %w", err)
	}
	defer rows.Close()

	var findings []brokenNavLinkFinding
	for rows.Next() {
		var f brokenNavLinkFinding
		var exampleHref sql.NullString
		if err := rows.Scan(&f.SlotName, &f.LinkCount, &exampleHref); err != nil {
			logger.Warn("Failed to scan broken nav link", zap.Error(err))
			continue
		}
		if exampleHref.Valid {
			f.ExampleHref = exampleHref.String
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// CHECK: placeholder_contact
// Detects fabricated/placeholder contact info in page_components.
// Looks for common hallucination patterns: 555-xxxx phones, example.com emails,
// 123 Main St addresses, Lorem ipsum, [placeholder] markers.
// Handler: page-content-writer (re-write the section with strict no-fabrication prompt)
// ============================================================================

type placeholderContactFinding struct {
	PageID      string `json:"page_id"`
	PageName    string `json:"page_name"`
	Position    int    `json:"position"`
	Pattern     string `json:"pattern"`
	MatchedText string `json:"matched_text"`
}

func findPlaceholderContact(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]placeholderContactFinding, error) {
	// Check rendered_html for common fabrication patterns
	rows, err := db.QueryContext(ctx, `
		SELECT pc.page_id, p.name, pc.position,
		       CASE
		           WHEN pc.rendered_html ~* '555[- ]?\d{3}[- ]?\d{4}' THEN 'fake_phone_555'
		           WHEN pc.rendered_html ~* '\(555\)' THEN 'fake_phone_555'
		           WHEN pc.rendered_html ~* '@example\.(com|org|net)' THEN 'example_email'
		           WHEN pc.rendered_html ~* 'info@(company|business|yourdomain|domain)\.' THEN 'generic_email'
		           WHEN pc.rendered_html ~* '123\s+(Main|First|Business)\s+(St|Street|Ave|Road)' THEN 'fake_address'
		           WHEN pc.rendered_html ~* '\[(?:your|insert|company|phone|email|address)' THEN 'bracket_placeholder'
		           WHEN pc.rendered_html ~* 'Lorem ipsum' THEN 'lorem_ipsum'
		           WHEN pc.rendered_html ~* '\+1[- ]?\(0{3}\)' THEN 'fake_phone_000'
		           WHEN pc.rendered_html ~* '(?:John|Jane)\s+(?:Doe|Smith)\s' THEN 'placeholder_name'
		       END as pattern,
		       SUBSTRING(pc.rendered_html FROM '(?i)((?:555[- ]?\d{3}[- ]?\d{4}|\(555\)[^<]{0,20}|[\w.]+@example\.(?:com|org|net)|info@(?:company|business|yourdomain|domain)\.\w+|123\s+(?:Main|First|Business)\s+(?:St|Street|Ave|Road)[^<]{0,30}|\[(?:your|insert|company|phone|email|address)[^\]]*\]|Lorem ipsum[^<]{0,30}))') as matched
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL
		  AND (
		      pc.rendered_html ~* '555[- ]?\d{3}[- ]?\d{4}'
		      OR pc.rendered_html ~* '\(555\)'
		      OR pc.rendered_html ~* '@example\.(com|org|net)'
		      OR pc.rendered_html ~* 'info@(company|business|yourdomain|domain)\.'
		      OR pc.rendered_html ~* '123\s+(Main|First|Business)\s+(St|Street|Ave|Road)'
		      OR pc.rendered_html ~* '\[(?:your|insert|company|phone|email|address)'
		      OR pc.rendered_html ~* 'Lorem ipsum'
		      OR pc.rendered_html ~* '\+1[- ]?\(0{3}\)'
		      OR pc.rendered_html ~* '(?:John|Jane)\s+(?:Doe|Smith)\s'
		  )
		ORDER BY p.name, pc.position
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("placeholder_contact query failed: %w", err)
	}
	defer rows.Close()

	var findings []placeholderContactFinding
	for rows.Next() {
		var f placeholderContactFinding
		var pattern, matched sql.NullString
		if err := rows.Scan(&f.PageID, &f.PageName, &f.Position, &pattern, &matched); err != nil {
			logger.Warn("Failed to scan placeholder contact", zap.Error(err))
			continue
		}
		if pattern.Valid {
			f.Pattern = pattern.String
		}
		if matched.Valid {
			f.MatchedText = matched.String
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// CHECK: generic_theme
// Detects sites where the CSS is still using default/fallback values,
// meaning the webdesign agent either didn't run or had no context.
// Checks site_specs for missing 'webdesign' aspect AND checks the actual
// CSS for default color values.
// Handler: webdesign-agent (re-run with proper identity context)
// ============================================================================

type genericThemeFinding struct {
	HasWebdesignSpec bool   `json:"has_webdesign_spec"`
	HasIdentitySpec  bool   `json:"has_identity_spec"`
	CSSLength        int    `json:"css_length"`
	UsesDefaultColor bool   `json:"uses_default_color"`
	Detail           string `json:"detail"`
}

func findGenericTheme(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (*genericThemeFinding, error) {
	finding := &genericThemeFinding{}

	// Check for webdesign spec
	var webdesignCount int
	db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM site_specs 
		WHERE site_id = $1 AND aspect = 'webdesign' AND is_current = true
	`, siteID).Scan(&webdesignCount)
	finding.HasWebdesignSpec = webdesignCount > 0

	// Check for identity spec (needed by webdesign)
	var identityCount int
	db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM site_specs 
		WHERE site_id = $1 AND aspect = 'identity' AND is_current = true
	`, siteID).Scan(&identityCount)
	finding.HasIdentitySpec = identityCount > 0

	// Check actual CSS for default values
	var cssHTML sql.NullString
	db.QueryRowContext(ctx, `
		SELECT rendered_html FROM site_components
		WHERE site_id = $1 AND slot_name = 'head'
	`, siteID).Scan(&cssHTML)

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
		return nil, nil // not a problem
	}

	return finding, nil
}

// countHardcodedColorComponents counts page_components that have hardcoded hex
// background colors inside <style> blocks in their rendered_html.
func countHardcodedColorComponents(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
		  AND pc.rendered_html LIKE '%<style%'
	`, siteID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("hardcoded color count query failed: %w", err)
	}
	return count, nil
}
