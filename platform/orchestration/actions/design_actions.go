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
//   - pages: [{title, name, component_functions}]
//   - all_component_functions: deduplicated list
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
					// identity.industry, identity.tagline, identity.company_name
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
					// Also extract services for richer context
					if svcs, ok := specData["services"]; ok && svcs != nil {
						result["services"] = svcs
					}

				case "briefing":
					// briefing.tone, briefing.about_us, briefing.tagline
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
					// Pass tone and about_us for richer design context
					if tone, _ := specData["tone"].(string); tone != "" {
						result["brand_tone"] = tone
					}
					if about, _ := specData["about_us"].(string); about != "" {
						result["about_us"] = about
					}

				case "classification":
					// classification.site_type, classification.recommended_builder
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

	if includePages {
		pages, err := loadPagesWithComponents(ctx, params.DB, id)
		if err != nil {
			params.Logger.Warn("Failed to load pages", zap.Error(err))
			pages = []map[string]interface{}{}
		}
		result["pages"] = pages

		for _, page := range pages {
			if funcs, ok := page["component_functions"].([]string); ok {
				for _, f := range funcs {
					allComponents[f] = true
				}
			}
		}
	}

	// Default components if none found
	if len(allComponents) == 0 {
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
		zap.Int("components", len(funcSlice)))

	return result, nil
}

// loadPagesWithComponents loads pages and extracts component functions from sections jsonb
func loadPagesWithComponents(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, title, name, url, sections, status
		FROM pages 
		WHERE site_id = $1 AND status IN ('deployed', 'published', 'draft', 'planned')
		ORDER BY CASE WHEN name = 'index' OR name = 'home' THEN 0 ELSE 1 END, nav_order
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var title, name, status string
		var url *string
		var sectionsJSON []byte

		if err := rows.Scan(&id, &title, &name, &url, &sectionsJSON, &status); err != nil {
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

		// Extract component functions from sections jsonb
		page["component_functions"] = extractComponentsFromSections(sectionsJSON)

		pages = append(pages, page)
	}

	return pages, nil
}

// extractComponentsFromSections extracts component function names from the sections jsonb column.
// Sections is an array like: [{"component_name": "hero-split-image", "function": "hero", ...}, ...]
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
