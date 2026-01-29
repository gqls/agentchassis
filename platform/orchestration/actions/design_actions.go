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
//   - pages: [{title, slug, component_functions}]
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

// loadPagesWithComponents loads pages and extracts component functions from HTML
func loadPagesWithComponents(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, title, slug, html_content, status
		FROM pages 
		WHERE site_id = $1 AND status IN ('deployed', 'published', 'draft', 'planned')
		ORDER BY CASE WHEN slug = '/' OR slug = '/index.html' THEN 0 ELSE 1 END, sort_order
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var title, slug, status string
		var htmlContent *string

		if err := rows.Scan(&id, &title, &slug, &htmlContent, &status); err != nil {
			continue
		}

		page := map[string]interface{}{
			"id":     id.String(),
			"title":  title,
			"slug":   slug,
			"status": status,
		}

		// Extract component functions from data-component attributes
		if htmlContent != nil {
			page["component_functions"] = extractDataComponents(*htmlContent)
		} else {
			page["component_functions"] = []string{}
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// extractDataComponents finds data-component="xxx" in HTML
func extractDataComponents(html string) []string {
	funcs := make(map[string]bool)
	search := `data-component="`
	idx := 0

	for {
		pos := strings.Index(html[idx:], search)
		if pos == -1 {
			break
		}
		start := idx + pos + len(search)
		end := strings.Index(html[start:], `"`)
		if end == -1 {
			break
		}
		name := html[start : start+end]
		if name != "" {
			funcs[name] = true
		}
		idx = start + end
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
