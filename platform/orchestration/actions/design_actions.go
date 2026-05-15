// LoadSiteForDesignAction loads site context for design agents
//
// File: platform/orchestration/actions/design_actions.go
//
// Config:
//   - site_id_field: path to site UUID
//   - domain_field: alternative path to domain
//   - include_pages: load pages and extract component functions (default true)
//   - include_style_collection: load linked style collection (default true)
//
// Returns:
//   - site_id, domain, company_name, industry, tagline
//   - style_collection_id: string UUID of the site's linked collection, or
//     nil if the site has no collection yet. Surfaced so downstream steps
//     (e.g. webdesign-agent's install_theme conditional) can test whether
//     a theme is already installed.
//   - pages: [{title, name, component_functions, built_component_count, planned_component_count}]
//   - all_component_functions: deduplicated list across all pages
//   - color_palette, typography (from style_collection or content_data)
//   - source: "database"

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

func LoadSiteForDesignAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadSiteForDesignAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Get site_id or domain
	var siteID uuid.UUID
	var domain string

	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		if siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField); siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = parsed
			}
		}
	}

	if siteID == uuid.Nil {
		if domainField, ok := config["domain_field"].(string); ok && domainField != "" {
			domain = datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
		}
	}

	if siteID == uuid.Nil && domain == "" {
		return nil, fmt.Errorf("either site_id or domain is required")
	}

	// Load site data
	var query string
	var args []interface{}

	if siteID != uuid.Nil {
		query = `SELECT id, domain, style_collection_id, content_data FROM sites WHERE id = $1`
		args = []interface{}{siteID}
	} else {
		query = `SELECT id, domain, style_collection_id, content_data FROM sites WHERE domain = $1`
		args = []interface{}{domain}
	}

	var id uuid.UUID
	var siteDomain string
	var styleCollectionID *uuid.UUID
	var contentDataJSON []byte

	err := params.DB.QueryRowContext(ctx, query, args...).Scan(&id, &siteDomain, &styleCollectionID, &contentDataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load site: %w", err)
	}

	// Parse content_data
	contentData := make(map[string]interface{})
	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &contentData)
	}

	// Build result
	result := map[string]interface{}{
		"site_id":      id.String(),
		"domain":       siteDomain,
		"company_name": getStringFromPaths(contentData, "company_name", "reviewed_brief.company_name"),
		"industry":     getStringFromPaths(contentData, "industry", "reviewed_brief.industry"),
		"tagline":      getStringFromPaths(contentData, "tagline", "reviewed_brief.tagline"),
		"source":       "database",
	}

	// Surface style_collection_id so downstream steps can conditionally
	// install a theme or skip. Explicit nil when absent makes the field
	// present in the map — conditionals can test `== null` rather than
	// the undefined "missing key" case.
	if styleCollectionID != nil {
		result["style_collection_id"] = styleCollectionID.String()
	} else {
		result["style_collection_id"] = nil
	}

	// --- Enrich from site_specs if content_data was sparse ---
	// The new build pipeline writes identity/classification to site_specs,
	// not sites.content_data. Fall back to site_specs for missing fields.
	needsEnrichment := result["company_name"] == "" || result["industry"] == "" || result["tagline"] == ""
	if needsEnrichment {
		params.Logger.Info("LoadSiteForDesignAction: content_data sparse, checking site_specs",
			zap.String("company_name", fmt.Sprintf("%v", result["company_name"])),
			zap.String("industry", fmt.Sprintf("%v", result["industry"])),
		)

		specRows, err := params.DB.QueryContext(ctx, `
			SELECT aspect, data
			FROM site_specs
			WHERE site_id = $1
			  AND aspect IN ('identity', 'briefing', 'classification')
			  AND is_current = true
			ORDER BY CASE aspect
				WHEN 'identity' THEN 1
				WHEN 'briefing' THEN 2
				WHEN 'classification' THEN 3
			END
		`, id)
		if err == nil {
			defer specRows.Close()
			for specRows.Next() {
				var aspect string
				var specDataJSON []byte
				if err := specRows.Scan(&aspect, &specDataJSON); err != nil {
					continue
				}
				var specData map[string]interface{}
				if json.Unmarshal(specDataJSON, &specData) != nil {
					continue
				}

				switch aspect {
				case "identity":
					if result["industry"] == "" || result["industry"] == nil {
						if v, _ := specData["industry"].(string); v != "" {
							result["industry"] = v
						}
					}
					if result["tagline"] == "" || result["tagline"] == nil {
						if v, _ := specData["tagline"].(string); v != "" {
							result["tagline"] = v
						}
					}
					if result["company_name"] == "" || result["company_name"] == nil {
						if v, _ := specData["company_name"].(string); v != "" {
							result["company_name"] = v
						}
					}
					if svcs, ok := specData["services"]; ok && svcs != nil {
						result["services"] = svcs
					}

				case "briefing":
					if result["tagline"] == "" || result["tagline"] == nil {
						if v, _ := specData["tagline"].(string); v != "" {
							result["tagline"] = v
						}
					}
					if result["company_name"] == "" || result["company_name"] == nil {
						if v, _ := specData["company_name"].(string); v != "" {
							result["company_name"] = v
						}
					}
					if tone, _ := specData["tone"].(string); tone != "" {
						result["brand_tone"] = tone
					}
					if about, _ := specData["about_us"].(string); about != "" {
						result["about_us"] = about
					}

				case "classification":
					if siteType, _ := specData["site_type"].(string); siteType != "" {
						result["site_type"] = siteType
					}
				}
			}
		}

		params.Logger.Info("LoadSiteForDesignAction: After site_specs enrichment",
			zap.String("company_name", fmt.Sprintf("%v", result["company_name"])),
			zap.String("industry", fmt.Sprintf("%v", result["industry"])),
			zap.String("brand_tone", fmt.Sprintf("%v", result["brand_tone"])),
		)
	}

	// Load pages if requested (default true)
	includePages := true
	if ip, ok := config["include_pages"].(bool); ok {
		includePages = ip
	}

	allComponents := make(map[string]bool)
	pagesLoaded := 0
	builtTotal := 0
	plannedTotal := 0

	if includePages {
		pages, err := loadPagesWithComponents(ctx, params.DB, id, params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to load pages", zap.Error(err))
			pages = []map[string]interface{}{}
		}
		result["pages"] = pages
		pagesLoaded = len(pages)

		for _, page := range pages {
			if funcs, ok := page["component_functions"].([]string); ok {
				for _, f := range funcs {
					allComponents[f] = true
				}
			}
			if n, ok := page["built_component_count"].(int); ok {
				builtTotal += n
			}
			if n, ok := page["planned_component_count"].(int); ok {
				plannedTotal += n
			}
		}
	}

	// Loud warning when the fallback fires. Previously this was a silent
	// no-op which masked the load_site_for_design bug for months — empty
	// allComponents meant every webdesign-agent run got the same 5-item
	// hardcoded list as the "real" component inventory, no matter what the
	// site actually contained. If this Warn fires on an established site
	// (one with page_components rows), something is broken upstream.
	if len(allComponents) == 0 {
		params.Logger.Warn("LoadSiteForDesignAction: NO COMPONENTS FOUND — using hardcoded 5-item fallback list. "+
			"For a site with built pages this indicates page_components is empty or the query is broken.",
			zap.String("site_id", id.String()),
			zap.Bool("queried_pages", includePages),
			zap.Int("pages_loaded", pagesLoaded),
			zap.Int("built_components_total", builtTotal),
			zap.Int("planned_components_total", plannedTotal))
		for _, f := range []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"} {
			allComponents[f] = true
		}
	}

	funcSlice := make([]string, 0, len(allComponents))
	for f := range allComponents {
		funcSlice = append(funcSlice, f)
	}
	result["all_component_functions"] = funcSlice

	// Get colors from content_data first
	if cp := getMapFromPath(contentData, "color_palette"); cp != nil {
		result["color_palette"] = cp
	}
	if tp := getMapFromPath(contentData, "typography"); tp != nil {
		result["typography"] = tp
	}

	// Load style collection if available and colors not in content_data
	includeStyleCollection := true
	if isc, ok := config["include_style_collection"].(bool); ok {
		includeStyleCollection = isc
	}

	if includeStyleCollection && styleCollectionID != nil && result["color_palette"] == nil {
		var colorPaletteJSON, typographyJSON []byte
		err := params.DB.QueryRowContext(ctx,
			`SELECT color_palette, typography FROM style_collections WHERE id = $1`,
			*styleCollectionID,
		).Scan(&colorPaletteJSON, &typographyJSON)

		if err == nil {
			if len(colorPaletteJSON) > 0 {
				var cp map[string]interface{}
				if json.Unmarshal(colorPaletteJSON, &cp) == nil {
					result["color_palette"] = cp
				}
			}
			if len(typographyJSON) > 0 {
				var tp map[string]interface{}
				if json.Unmarshal(typographyJSON, &tp) == nil {
					result["typography"] = tp
				}
			}
		}
	}

	params.Logger.Info("LoadSiteForDesignAction: Complete",
		zap.String("site_id", id.String()),
		zap.Int("components", len(funcSlice)),
		zap.Int("pages_loaded", pagesLoaded),
		zap.Int("built_components_total", builtTotal),
		zap.Int("planned_components_total", plannedTotal))

	return result, nil
}

// loadPagesWithComponents loads pages and their component functions for a site.
//
// PRIMARY source: page_components JOIN content_components — the live page
// sections actually built by page-content-writer. This is what every
// downstream CSS/JS pipeline step needs to see.
//
// SECONDARY source: pages.sections JSONB — used as a fallback for pages
// PLANNED but not yet BUILT (page-content-writer hasn't run, so
// page_components has no rows for them). Site-planner writes planned
// section descriptors to pages.sections.
//
// Previously this function relied ONLY on extractComponentsFromSections
// over pages.sections, which returned empty for every built site (planning-
// time data, stale once content-writer runs). That caused the 5-item
// hardcoded fallback in LoadSiteForDesignAction to fire on every webdesign-
// agent invocation, masking the entire css_snippets library — only snippets
// with applies_to overlapping those 5 names ever shipped to any site.
func loadPagesWithComponents(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			p.id,
			COALESCE(p.title, ''),
			p.name,
			p.url,
			COALESCE(p.sections, '[]'::jsonb),
			COALESCE(p.status, 'unknown'),
			COALESCE(
				jsonb_agg(DISTINCT cc.function ORDER BY cc.function) FILTER (
					WHERE cc.function IS NOT NULL AND cc.function != ''
				),
				'[]'::jsonb
			) AS built_component_functions
		FROM pages p
		LEFT JOIN page_components pc ON pc.page_id = p.id
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND p.status IN ('deployed', 'published', 'draft', 'planned')
		GROUP BY p.id
		ORDER BY CASE WHEN p.name = 'index' OR p.name = 'home' THEN 0 ELSE 1 END,
		         COALESCE(p.nav_order, 100)
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []map[string]interface{}
	for rows.Next() {
		var (
			id             uuid.UUID
			title          string
			name           string
			url            *string
			sectionsJSON   []byte
			status         string
			builtFuncsJSON []byte
		)
		if err := rows.Scan(&id, &title, &name, &url,
			&sectionsJSON, &status, &builtFuncsJSON); err != nil {
			logger.Warn("loadPagesWithComponents: scan error", zap.Error(err))
			continue
		}

		page := map[string]interface{}{
			"id":     id.String(),
			"title":  title,
			"name":   name,
			"status": status,
		}
		if url != nil {
			page["url"] = *url
		}

		// PRIMARY: built component functions from page_components ⋈ content_components
		var builtFuncs []string
		if err := json.Unmarshal(builtFuncsJSON, &builtFuncs); err != nil {
			logger.Warn("loadPagesWithComponents: built_component_functions JSON parse failed",
				zap.String("page_name", name),
				zap.Error(err))
			builtFuncs = nil
		}

		// SECONDARY: planned section descriptors from pages.sections JSONB.
		// Used when page-content-writer hasn't run yet (page_components empty).
		plannedFuncs := extractComponentsFromSections(sectionsJSON)

		// Merge — both sources contribute, deduped
		merged := make(map[string]bool)
		for _, f := range builtFuncs {
			merged[f] = true
		}
		for _, f := range plannedFuncs {
			merged[f] = true
		}
		funcs := make([]string, 0, len(merged))
		for f := range merged {
			funcs = append(funcs, f)
		}

		page["component_functions"] = funcs
		page["built_component_count"] = len(builtFuncs)
		page["planned_component_count"] = len(plannedFuncs)

		pages = append(pages, page)
	}
	return pages, nil
}

// extractComponentsFromSections extracts component function names from the sections jsonb column.
// Sections is an array like: [{"component_name": "hero-split-image", "function": "hero", ...}, ...]
//
// Used as SECONDARY source by loadPagesWithComponents for pages that have
// been planned but not yet built. For built pages, page_components is the
// authoritative source.
func extractComponentsFromSections(sectionsJSON []byte) []string {
	if len(sectionsJSON) == 0 {
		return []string{}
	}

	var sections []map[string]interface{}
	if err := json.Unmarshal(sectionsJSON, &sections); err != nil {
		return []string{}
	}

	funcs := make(map[string]bool)
	for _, section := range sections {
		if f, ok := section["function"].(string); ok && f != "" {
			funcs[f] = true
		}
		if f, ok := section["category"].(string); ok && f != "" {
			funcs[f] = true
		}
	}

	result := make([]string, 0, len(funcs))
	for f := range funcs {
		result = append(result, f)
	}
	return result
}

// getStringFromPaths tries multiple paths to find a string
func getStringFromPaths(data map[string]interface{}, paths ...string) string {
	for _, path := range paths {
		if strings.Contains(path, ".") {
			if v := datahelpers.ExtractNestedFieldString(data, path); v != "" {
				return v
			}
		} else {
			if v, ok := data[path].(string); ok && v != "" {
				return v
			}
		}
	}
	return ""
}

// getMapFromPath extracts a nested map
func getMapFromPath(data map[string]interface{}, key string) map[string]interface{} {
	if v, ok := data[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}
