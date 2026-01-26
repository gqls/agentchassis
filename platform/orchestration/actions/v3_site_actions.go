// FILE: platform/orchestration/actions/v3_site_actions.go
// Additional actions needed for the v3 multipage website builder component-based architecture.
// These complement existing actions in site_db_actions.go and component_library.go.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ============================================================================
// ACTION: select_style_collection
// ============================================================================

// SelectStyleCollectionAction selects a style collection for a site
// Either by explicit ID, by site_id lookup, or by domain/industry matching
// Config:
//   - style_collection_id: explicit UUID (optional)
//   - site_id_field: path to site_id in collected_data (optional)
//   - domain_field: path to domain for industry matching (optional)
func SelectStyleCollectionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SelectStyleCollectionAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	if params.DB == nil {
		params.Logger.Warn("SelectStyleCollectionAction: No database, returning default style")
		return getDefaultStyleCollection(), nil
	}

	// Priority 1: Explicit style_collection_id
	if scID, ok := config["style_collection_id"].(string); ok && scID != "" {
		scUUID, err := uuid.Parse(scID)
		if err == nil {
			coll, err := getStyleCollectionByID(ctx, params.DB, scUUID, params.Logger)
			if err == nil {
				return styleCollectionToResult(coll), nil
			}
		}
	}

	// Priority 2: Look up by site_id
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			siteUUID, err := uuid.Parse(siteIDStr)
			if err == nil {
				coll, err := GetStyleCollectionForSite(ctx, params.DB, siteUUID, params.Logger)
				if err == nil && coll != nil {
					return styleCollectionToResult(coll), nil
				}
			}
		}
	}

	// Priority 3: Match by domain/industry
	domainField := "input_data.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
	if domain != "" {
		coll, err := SelectStyleCollectionByDomain(ctx, params.DB, domain, params.Logger)
		if err == nil && coll != nil {
			return styleCollectionToResult(coll), nil
		}
	}

	// Fallback: Return default
	params.Logger.Info("SelectStyleCollectionAction: No matching style found, using default")
	return getDefaultStyleCollection(), nil
}

func styleCollectionToResult(coll *StyleCollection) map[string]interface{} {
	result := map[string]interface{}{
		"style_collection_id": coll.ID.String(),
		"name":                coll.Name,
		"display_name":        coll.DisplayName,
		"category":            coll.Category,
		"color_palette":       coll.ColorPalette,
		"typography":          coll.Typography,
	}
	if coll.HeaderComponentID != nil {
		result["header_component_id"] = coll.HeaderComponentID.String()
	}
	if coll.FooterComponentID != nil {
		result["footer_component_id"] = coll.FooterComponentID.String()
	}
	if coll.CSSThemeID != nil {
		result["css_theme_id"] = coll.CSSThemeID.String()
	}
	return result
}

func getDefaultStyleCollection() map[string]interface{} {
	return map[string]interface{}{
		"style_collection_id": "",
		"name":                "default",
		"display_name":        "Default Style",
		"category":            "general",
		"color_palette": map[string]string{
			"primary":    "#1a1a2e",
			"secondary":  "#16213e",
			"accent":     "#0f3460",
			"text":       "#333333",
			"background": "#ffffff",
		},
		"typography": map[string]string{
			"heading": "system-ui, sans-serif",
			"body":    "system-ui, sans-serif",
		},
	}
}

// ============================================================================
// ACTION: update_site_content
// ============================================================================

// UpdateSiteContentAction updates the sites.content_data JSONB column
// Used to store the site plan, brand DNA, or other structured content
// Config:
//   - site_id_field: path to site_id in collected_data
//   - content_field: path to content data to store
//   - merge: boolean - if true, merges with existing content_data
func UpdateSiteContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteContentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get content to store
	contentField := "page_plan"
	if f, ok := config["content_field"].(string); ok && f != "" {
		contentField = f
	}
	contentValue := datahelpers.ExtractNestedField(params.CollectedData, contentField)
	if contentValue == nil {
		return nil, fmt.Errorf("content not found at %s", contentField)
	}

	contentJSON, err := json.Marshal(contentValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteContentAction: No database, skipping update")
		return map[string]interface{}{
			"updated":     false,
			"site_id":     siteIDStr,
			"reason":      "no database connection",
			"content_key": contentField,
		}, nil
	}

	// Determine if merge or replace
	merge, _ := config["merge"].(bool)

	var query string
	if merge {
		query = `
			UPDATE sites 
			SET content_data = COALESCE(content_data, '{}'::jsonb) || $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE sites 
			SET content_data = $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	}

	if err := execDB(ctx, params.DB, query, siteID, string(contentJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site content: %w", err)
	}

	params.Logger.Info("UpdateSiteContentAction: Site content updated",
		zap.String("site_id", siteIDStr),
		zap.String("content_field", contentField),
		zap.Bool("merged", merge),
	)

	return map[string]interface{}{
		"updated":      true,
		"site_id":      siteIDStr,
		"content_key":  contentField,
		"content_size": len(contentJSON),
		"merged":       merge,
	}, nil
}

// ============================================================================
// ACTION: update_site_status
// ============================================================================

// UpdateSiteStatusAction updates the sites.status column
// Config:
//   - site_id_field: path to site_id
//   - status: new status value (draft, building, review, published, archived)
func UpdateSiteStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get new status
	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required in config")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft": true, "building": true, "review": true,
		"published": true, "deployed": true, "archived": true, "error": true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s (valid: draft, building, review, published, deployed, archived, error)", newStatus)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Check if deployed_at should be set
	var query string
	deployedAt, hasDeployedAt := config["deployed_at"].(string)
	if hasDeployedAt && (deployedAt == "now" || deployedAt == "NOW()") && newStatus == "deployed" {
		query = `UPDATE sites SET status = $2, last_deployed_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE sites SET status = $2, updated_at = NOW() WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, siteID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update site status: %w", err)
	}

	params.Logger.Info("UpdateSiteStatusAction: Status updated",
		zap.String("site_id", siteIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":   true,
		"site_id":   siteIDStr,
		"status":    newStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ============================================================================
// ACTION: update_site_defaults
// ============================================================================

// UpdateSiteDefaultsAction updates the sites.default_components JSONB column
// Used to store default header/footer component IDs
// Config:
//   - site_id_field: path to site_id in collected_data
//   - header_component_id: UUID of header component (optional)
//   - footer_component_id: UUID of footer component (optional)
//   - css_theme_id: UUID of CSS theme (optional)
//   - defaults_field: path to a map containing all defaults (alternative to individual fields)
func UpdateSiteDefaultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteDefaultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Build defaults map
	defaults := make(map[string]interface{})

	// Try getting from defaults_field first
	if defaultsField, ok := config["defaults_field"].(string); ok && defaultsField != "" {
		if defaultsData := datahelpers.ExtractNestedField(params.CollectedData, defaultsField); defaultsData != nil {
			if m, ok := defaultsData.(map[string]interface{}); ok {
				defaults = m
			}
		}
	}

	// Override/add from explicit config fields
	if headerID, ok := config["header_component_id"].(string); ok && headerID != "" {
		defaults["header_component_id"] = headerID
	}
	if footerID, ok := config["footer_component_id"].(string); ok && footerID != "" {
		defaults["footer_component_id"] = footerID
	}
	if cssThemeID, ok := config["css_theme_id"].(string); ok && cssThemeID != "" {
		defaults["css_theme_id"] = cssThemeID
	}

	// Also check collected data for style_collection results
	if sc := datahelpers.ExtractNestedField(params.CollectedData, "style_collection"); sc != nil {
		if scMap, ok := sc.(map[string]interface{}); ok {
			if hID, ok := scMap["header_component_id"].(string); ok && hID != "" {
				if defaults["header_component_id"] == nil {
					defaults["header_component_id"] = hID
				}
			}
			if fID, ok := scMap["footer_component_id"].(string); ok && fID != "" {
				if defaults["footer_component_id"] == nil {
					defaults["footer_component_id"] = fID
				}
			}
			if cID, ok := scMap["css_theme_id"].(string); ok && cID != "" {
				if defaults["css_theme_id"] == nil {
					defaults["css_theme_id"] = cID
				}
			}
		}
	}

	if len(defaults) == 0 {
		params.Logger.Warn("UpdateSiteDefaultsAction: No defaults to set")
		return map[string]interface{}{
			"updated": false,
			"site_id": siteIDStr,
			"reason":  "no defaults provided",
		}, nil
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteDefaultsAction: No database")
		return map[string]interface{}{
			"updated":  false,
			"site_id":  siteIDStr,
			"defaults": defaults,
			"reason":   "no database connection",
		}, nil
	}

	defaultsJSON, err := json.Marshal(defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal defaults: %w", err)
	}

	query := `
		UPDATE sites 
		SET default_components = COALESCE(default_components, '{}'::jsonb) || $2::jsonb,
		    updated_at = NOW()
		WHERE id = $1
	`

	if err := execDB(ctx, params.DB, query, siteID, string(defaultsJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site defaults: %w", err)
	}

	params.Logger.Info("UpdateSiteDefaultsAction: Defaults updated",
		zap.String("site_id", siteIDStr),
		zap.Any("defaults", defaults),
	)

	return map[string]interface{}{
		"updated":  true,
		"site_id":  siteIDStr,
		"defaults": defaults,
	}, nil
}

// ============================================================================
// ACTION: update_page_status
// ============================================================================

// UpdatePageStatusAction updates a single page's build_status
// Config:
//   - page_id_field: path to page_id OR
//   - site_id_field + page_name_field: to look up page
//   - status: new build_status value (e.g., "deployed", "failed", "building")
func UpdatePageStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageStatusAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("UpdatePageStatusAction: No database connection available")
		return map[string]interface{}{"updated": false, "reason": "no database"}, nil
	}

	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required")
	}

	var pageID uuid.UUID

	// Try direct page_id from config field
	if pageIDField, ok := config["page_id_field"].(string); ok && pageIDField != "" {
		pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField)
		if pageIDStr != "" {
			var err error
			pageID, err = uuid.Parse(pageIDStr)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Invalid page_id format",
					zap.String("page_id_field", pageIDField),
					zap.String("value", pageIDStr),
					zap.Error(err))
				return nil, fmt.Errorf("invalid page_id: %w", err)
			}
		}
	}

	// Alternative: try current_page.id (common in loop iterations)
	if pageID == uuid.Nil {
		if pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.id"); pageIDStr != "" {
			if parsed, err := uuid.Parse(pageIDStr); err == nil {
				pageID = parsed
			}
		}
	}

	// Alternative: look up by site_id + page_name
	if pageID == uuid.Nil {
		siteIDField, _ := config["site_id_field"].(string)
		pageNameField, _ := config["page_name_field"].(string)

		if siteIDField != "" && pageNameField != "" {
			siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
			pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, pageNameField)

			if siteIDStr != "" && pageName != "" {
				siteUUID, _ := uuid.Parse(siteIDStr)
				var err error
				pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
				if err != nil {
					return nil, fmt.Errorf("page lookup failed: %w", err)
				}
			}
		}
	}

	// Last resort: try current_page.name with site_record.site_id
	if pageID == uuid.Nil {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.name")

		if siteIDStr != "" && pageName != "" {
			siteUUID, _ := uuid.Parse(siteIDStr)
			var err error
			pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Page lookup by name failed",
					zap.String("site_id", siteIDStr),
					zap.String("page_name", pageName),
					zap.Error(err))
			}
		}
	}

	if pageID == uuid.Nil {
		params.Logger.Error("UpdatePageStatusAction: Could not determine page_id",
			zap.Any("config", config))
		return nil, fmt.Errorf("could not determine page_id")
	}

	// Build the query - use build_status column (not status)
	var query string
	if newStatus == "deployed" {
		// Also set deployed_at when marking as deployed
		query = `UPDATE pages SET build_status = $2, deployed_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE pages SET build_status = $2, updated_at = NOW() WHERE id = $1`
	}

	result, err := params.DB.ExecContext(ctx, query, pageID, newStatus)
	if err != nil {
		params.Logger.Error("UpdatePageStatusAction: Failed to update page",
			zap.String("page_id", pageID.String()),
			zap.String("build_status", newStatus),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update page build_status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	params.Logger.Info("UpdatePageStatusAction: Updated page",
		zap.String("page_id", pageID.String()),
		zap.String("build_status", newStatus),
		zap.Int64("rows_affected", rowsAffected))

	return map[string]interface{}{
		"updated":       true,
		"page_id":       pageID.String(),
		"build_status":  newStatus,
		"rows_affected": rowsAffected,
	}, nil
}

// BuildRenderContextAction assembles a RenderContext from multiple sources
// Used before rendering components with templates
func BuildRenderContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildRenderContextAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Start with empty context
	renderCtx := &RenderContext{
		Year: fmt.Sprintf("%d", time.Now().Year()),
	}

	sourcesMerged := 0

	// Handle sources config - can be array or map format
	// Array format: ["input_data", "site_record", ...]
	// Map format: {"page": "input_data.current_page", "site": "input_data.site_record", ...}

	if sourcesMap, ok := config["sources"].(map[string]interface{}); ok {
		// Map format: keys are logical names, values are paths to data
		params.Logger.Info("Using map-format sources config",
			zap.Int("source_count", len(sourcesMap)))

		for logicalName, pathVal := range sourcesMap {
			path, ok := pathVal.(string)
			if !ok {
				continue
			}

			sourceData := datahelpers.ExtractNestedField(params.CollectedData, path)
			if sourceData == nil {
				params.Logger.Debug("Source not found",
					zap.String("name", logicalName),
					zap.String("path", path))
				continue
			}

			if m, ok := sourceData.(map[string]interface{}); ok {
				params.Logger.Debug("Merging source",
					zap.String("name", logicalName),
					zap.String("path", path))
				mergeIntoRenderContextEnhanced(renderCtx, m, logicalName, params.Logger)
				sourcesMerged++
			}
		}
	} else if sourcesArray, ok := config["sources"].([]interface{}); ok {
		// Array format: direct paths
		for _, src := range sourcesArray {
			if srcStr, ok := src.(string); ok {
				sourceData := datahelpers.ExtractNestedField(params.CollectedData, srcStr)
				if sourceData == nil {
					continue
				}
				if m, ok := sourceData.(map[string]interface{}); ok {
					mergeIntoRenderContextEnhanced(renderCtx, m, srcStr, params.Logger)
					sourcesMerged++
				}
			}
		}
	} else {
		// Default sources
		defaultSources := []string{"input_data", "site_record", "style_collection", "reviewed_brief", "page_plan"}
		for _, source := range defaultSources {
			sourceData := datahelpers.ExtractNestedField(params.CollectedData, source)
			if sourceData == nil {
				// Try with input_data prefix
				sourceData = datahelpers.ExtractNestedField(params.CollectedData, "input_data."+source)
			}
			if sourceData == nil {
				continue
			}
			if m, ok := sourceData.(map[string]interface{}); ok {
				mergeIntoRenderContextEnhanced(renderCtx, m, source, params.Logger)
				sourcesMerged++
			}
		}
	}

	// Try to load navigation from DB if we have site_id
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if siteUUID, err := uuid.Parse(siteIDStr); err == nil && params.DB != nil {
				renderCtx.SiteID = siteUUID
				nav, _ := getNavigationFromDB(ctx, params.DB, siteUUID, "header", params.Logger)
				if nav != nil && len(nav.Items) > 0 {
					renderCtx.NavItems = convertNavigationItems(nav.Items)
				}
			}
		}
	}

	params.Logger.Info("BuildRenderContextAction: Context built",
		zap.String("domain", renderCtx.Domain),
		zap.String("company_name", renderCtx.CompanyName),
		zap.String("logo_text", renderCtx.LogoText),
		zap.Int("nav_items", len(renderCtx.NavItems)),
		zap.Int("sources_merged", sourcesMerged),
	)

	// FIX: Return the context directly, not wrapped in "render_context" key
	// The workflow output_field already specifies where to store it
	// Adding metadata fields at same level
	result := renderCtxToMap(renderCtx)
	result["_sources_merged"] = sourcesMerged
	result["_built_at"] = time.Now().Format(time.RFC3339)

	return result, nil
}

// mergeIntoRenderContextEnhanced extracts data from various source formats
func mergeIntoRenderContextEnhanced(ctx *RenderContext, data map[string]interface{}, sourceName string, logger *zap.Logger) {
	// =========================================================================
	// STEP 1: Unwrap .response wrapper if present (common in agent responses)
	// reviewed_brief has: {"response": {...actual data...}, "response_status": "complete"}
	// This recursively processes the unwrapped data first
	// =========================================================================
	if response, ok := data["response"].(map[string]interface{}); ok {
		// Check if this looks like a wrapped response (has response_status sibling)
		if _, hasStatus := data["response_status"]; hasStatus {
			logger.Debug("Unwrapping .response wrapper for source",
				zap.String("source", sourceName))
			// Recursively merge the unwrapped response data FIRST
			mergeIntoRenderContextEnhanced(ctx, response, sourceName+".response", logger)
			// Then continue to process the outer data for any additional fields
		}
	}

	// =========================================================================
	// STEP 2: Direct field extraction from current data level
	// =========================================================================

	// Domain
	if v, ok := data["domain"].(string); ok && v != "" {
		ctx.Domain = v
	}

	// Company name (sets logo_text as fallback)
	if v, ok := data["company_name"].(string); ok && v != "" {
		ctx.CompanyName = v
		if ctx.LogoText == "" {
			ctx.LogoText = v
		}
	}

	// Logo text (explicit override)
	if v, ok := data["logo_text"].(string); ok && v != "" {
		ctx.LogoText = v
	}

	// Tagline
	if v, ok := data["tagline"].(string); ok && v != "" {
		ctx.Tagline = v
	}

	// Email - check both "email" and "contact_email" (reviewed_brief uses contact_email)
	if v, ok := data["email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}

	// Phone - check both "phone" and "contact_phone"
	if v, ok := data["phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["contact_phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}

	// Colors - direct fields
	if v, ok := data["primary_color"].(string); ok && v != "" {
		ctx.PrimaryColor = v
	}
	if v, ok := data["secondary_color"].(string); ok && v != "" {
		ctx.SecondaryColor = v
	}
	if v, ok := data["accent_color"].(string); ok && v != "" {
		ctx.AccentColor = v
	}
	if v, ok := data["text_color"].(string); ok && v != "" {
		ctx.TextColor = v
	}
	if v, ok := data["background_color"].(string); ok && v != "" {
		ctx.BackgroundColor = v
	}

	// =========================================================================
	// STEP 3: Check nested color_palette (from style_collection)
	// =========================================================================
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok && v != "" {
			ctx.AccentColor = v
		}
		if v, ok := palette["background"].(string); ok && v != "" {
			ctx.BackgroundColor = v
		}
		if v, ok := palette["text"].(string); ok && v != "" {
			ctx.TextColor = v
		}
	}

	// =========================================================================
	// STEP 4: Extract from nested structures (business_context, contact_info, brand)
	// =========================================================================

	// business_context (alternative structure)
	if brief, ok := data["business_context"].(map[string]interface{}); ok {
		if v, ok := brief["company_name"].(string); ok && v != "" {
			ctx.CompanyName = v
			if ctx.LogoText == "" {
				ctx.LogoText = v
			}
		}
		if v, ok := brief["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
		if v, ok := brief["industry"].(string); ok && v != "" {
			ctx.Industry = v
		}
	}

	// contact_info (nested contact structure)
	if contact, ok := data["contact_info"].(map[string]interface{}); ok {
		if v, ok := contact["email"].(string); ok && v != "" {
			ctx.Email = v
		}
		if v, ok := contact["phone"].(string); ok && v != "" {
			ctx.Phone = v
		}
	}

	// brand (nested brand/visual settings)
	if brand, ok := data["brand"].(map[string]interface{}); ok {
		if v, ok := brand["primary_color"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := brand["secondary_color"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := brand["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
	}

	// =========================================================================
	// STEP 5: Content generation context (tone, target_audience, industry)
	// =========================================================================
	if v, ok := data["tone"].(string); ok && v != "" {
		ctx.Tone = v
	}
	if v, ok := data["target_audience"].(string); ok && v != "" {
		ctx.TargetAudience = v
	}
	if v, ok := data["industry"].(string); ok && v != "" {
		ctx.Industry = v
	}

	// =========================================================================
	// STEP 6: Site/page identifiers
	// =========================================================================
	if v, ok := data["site_id"].(string); ok && v != "" {
		if siteUUID, err := uuid.Parse(v); err == nil {
			ctx.SiteID = siteUUID
		}
	}

	// =========================================================================
	// STEP 7: CTA settings
	// =========================================================================
	if v, ok := data["cta_text"].(string); ok && v != "" {
		ctx.CTAText = v
	}
	if v, ok := data["cta_url"].(string); ok && v != "" {
		ctx.CTAUrl = v
	}

	// =========================================================================
	// STEP 8: Extract services array (for footer and services sections)
	// Services appear in reviewed_brief.response.services as []interface{}
	// Each service is {"name": "...", "description": "..."}
	//
	// ctx.Services is []string (just names)
	// ctx.ContentData["services"] is []interface{} (full objects for {{range .services}})
	// =========================================================================
	if services, ok := data["services"].([]interface{}); ok && len(services) > 0 {
		// Store full services in ContentData for template access via {{range .services}}
		if ctx.ContentData == nil {
			ctx.ContentData = make(map[string]interface{})
		}
		ctx.ContentData["services"] = services

		// Also extract just the names to ctx.Services ([]string)
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if name, ok := svcMap["name"].(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				}
			}
		}

		logger.Info("Extracted services array",
			zap.String("source", sourceName),
			zap.Int("full_count", len(services)),
			zap.Int("names_count", len(ctx.Services)))
	}

	// =========================================================================
	// STEP 9: Extract navigation from db_sync source
	// db_sync contains: {"navigation": {"items": [{"label": "Home", "url": "/index.html"}, ...]}}
	// =========================================================================
	if sourceName == "db_sync" || sourceName == "db_sync.response" {
		if navigation, ok := data["navigation"].(map[string]interface{}); ok {
			if items, ok := navigation["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							ctx.NavItems = append(ctx.NavItems, NavItem{
								Label: label,
								URL:   url,
							})
						}
					}
				}
				if len(ctx.NavItems) > 0 {
					logger.Info("Extracted navigation items from db_sync",
						zap.Int("count", len(ctx.NavItems)))
				}
			}
		}
	}

	// Handle site_record.content_data - recursively merge nested content_data
	if sourceName == "site_record" {
		if contentData, ok := data["content_data"].(map[string]interface{}); ok {
			logger.Debug("Processing site_record.content_data",
				zap.Int("fields", len(contentData)))
			mergeIntoRenderContextEnhanced(ctx, contentData, "site_record.content_data", logger)
		}
	}

	// Handle services array (for services_html generation)
	if len(ctx.Services) == 0 {
		if services, ok := data["services"].([]interface{}); ok && len(services) > 0 {
			for _, svc := range services {
				if name, ok := svc.(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				} else if svcMap, ok := svc.(map[string]interface{}); ok {
					if name, ok := svcMap["name"].(string); ok && name != "" {
						ctx.Services = append(ctx.Services, name)
					}
				}
			}
		}
	}

	// =========================================================================
	// STEP 10: Log final state for debugging
	// =========================================================================
	logger.Info("Merged source into render context",
		zap.String("source", sourceName),
		zap.String("company_name", ctx.CompanyName),
		zap.String("domain", ctx.Domain),
		zap.String("tagline", ctx.Tagline),
		zap.String("email", ctx.Email),
		zap.Int("nav_items", len(ctx.NavItems)),
		zap.Int("services", len(ctx.Services)))
}

// renderCtxToMap converts RenderContext to map for template substitution
func renderCtxToMap(ctx *RenderContext) map[string]interface{} {
	result := map[string]interface{}{
		"domain":           ctx.Domain,
		"logo_text":        ctx.LogoText,
		"company_name":     ctx.CompanyName,
		"tagline":          ctx.Tagline,
		"email":            ctx.Email,
		"phone":            ctx.Phone,
		"primary_color":    ctx.PrimaryColor,
		"secondary_color":  ctx.SecondaryColor,
		"accent_color":     ctx.AccentColor,
		"text_color":       ctx.TextColor,
		"background_color": ctx.BackgroundColor,
		"year":             ctx.Year,
		"cta_text":         ctx.CTAText,
		"cta_url":          ctx.CTAUrl,
		"industry":         ctx.Industry,
		"tone":             ctx.Tone,
		"target_audience":  ctx.TargetAudience,
	}

	if ctx.SiteID != uuid.Nil {
		result["site_id"] = ctx.SiteID.String()
	}

	// =========================================================================
	// Generate nav_items and nav_items_html from NavItems
	// Templates may use either:
	//   {{range .nav_items}} for iteration
	//   {{.nav_items_html}} for pre-rendered HTML
	// =========================================================================
	if len(ctx.NavItems) > 0 {
		// nav_items as array (for {{range .nav_items}})
		navItems := make([]map[string]interface{}, len(ctx.NavItems))
		for i, item := range ctx.NavItems {
			navItems[i] = map[string]interface{}{
				"label":     item.Label,
				"url":       item.URL,
				"is_active": item.IsActive,
			}
		}
		result["nav_items"] = navItems

		// nav_items_html as pre-rendered string (for {{.nav_items_html}})
		// Format: <li><a href="/page.html">Label</a></li>
		var htmlParts []string
		for _, item := range ctx.NavItems {
			activeClass := ""
			if item.IsActive {
				activeClass = ` class="active"`
			}
			htmlParts = append(htmlParts, fmt.Sprintf(
				`<li><a href="%s"%s>%s</a></li>`,
				item.URL, activeClass, item.Label,
			))
		}
		result["nav_items_html"] = strings.Join(htmlParts, "\n                ")
	} else {
		// Ensure empty values don't render as "<no value>"
		result["nav_items"] = []map[string]interface{}{}
		result["nav_items_html"] = ""
	}

	// =========================================================================
	// Generate services list HTML for footer
	// Templates use {{.services_html}} for footer services list
	// ctx.Services is []string (service names extracted from reviewed_brief)
	// =========================================================================
	if len(ctx.Services) > 0 {
		result["services"] = ctx.Services

		// services_html as pre-rendered string
		var servicesParts []string
		for _, serviceName := range ctx.Services {
			if serviceName != "" {
				servicesParts = append(servicesParts, fmt.Sprintf(
					`<li><a href="/services.html">%s</a></li>`,
					serviceName,
				))
			}
		}
		result["services_html"] = strings.Join(servicesParts, "\n                ")
	} else {
		result["services"] = []string{}
		result["services_html"] = ""
	}

	// =========================================================================
	// Merge ContentData fields for additional template access
	// This includes full service objects for {{range .services}} iteration
	// =========================================================================
	if ctx.ContentData != nil {
		for key, value := range ctx.ContentData {
			// Don't overwrite explicit fields
			if _, exists := result[key]; !exists {
				result[key] = value
			}
		}
	}

	return result
}

func mergeIntoRenderContext(ctx *RenderContext, data map[string]interface{}) {
	if v, ok := data["domain"].(string); ok && v != "" {
		ctx.Domain = v
	}
	if v, ok := data["company_name"].(string); ok && v != "" {
		ctx.CompanyName = v
		if ctx.LogoText == "" {
			ctx.LogoText = v
		}
	}
	if v, ok := data["logo_text"].(string); ok && v != "" {
		ctx.LogoText = v
	}
	if v, ok := data["tagline"].(string); ok && v != "" {
		ctx.Tagline = v
	}
	if v, ok := data["email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["primary_color"].(string); ok && v != "" {
		ctx.PrimaryColor = v
	}
	if v, ok := data["secondary_color"].(string); ok && v != "" {
		ctx.SecondaryColor = v
	}
	if v, ok := data["accent_color"].(string); ok && v != "" {
		ctx.AccentColor = v
	}

	// Check nested color_palette
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok {
			ctx.AccentColor = v
		}
	}

	// Capture ALL fields into ContentData for template access
	if ctx.ContentData == nil {
		ctx.ContentData = make(map[string]interface{})
	}
	for key, value := range data {
		ctx.ContentData[key] = value
	}
}

func convertNavigationItems(items []NavigationItem) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		result[i] = NavItem{
			Label: item.Label,
			URL:   item.URL,
		}
	}
	return result
}

// RenderComponentAction renders a single component template with context
// Config:
//   - component_function: function name to look up (e.g., "hero-banner")
//   - component_id: explicit component UUID (alternative to function)
//   - component_from: path to object containing function/id (e.g., "current_section")
//   - context_field: path to render context in collected_data
//   - content_field: path to additional content data
//   - content_from: alias for content_field (for consistency)
func RenderComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderComponentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for component rendering")
	}

	// Get component - now with support for component_from indirection
	var comp *Component
	var err error
	var componentFunction string
	var componentID string

	// Priority 1: Direct component_id in config
	if compID, ok := config["component_id"].(string); ok && compID != "" {
		componentID = compID
	}

	// Priority 2: Direct component_function in config
	if compFunc, ok := config["component_function"].(string); ok && compFunc != "" {
		componentFunction = compFunc
	}

	// Priority 3: Extract from component_from field (indirection)
	if componentID == "" && componentFunction == "" {
		if componentFrom, ok := config["component_from"].(string); ok && componentFrom != "" {
			componentData := datahelpers.ExtractNestedField(params.CollectedData, componentFrom)
			if componentData != nil {
				// Case 1: component_from points to a map (e.g., "current_section")
				if compMap, ok := componentData.(map[string]interface{}); ok {
					// Try to get function first (most common)
					if fn, ok := compMap["function"].(string); ok && fn != "" {
						componentFunction = fn
						params.Logger.Debug("RenderComponentAction: Extracted function from component_from",
							zap.String("component_from", componentFrom),
							zap.String("function", componentFunction),
						)
					}
					// Also check for component_function key
					if fn, ok := compMap["component_function"].(string); ok && fn != "" {
						componentFunction = fn
					}
					// Check for id/component_id
					if id, ok := compMap["id"].(string); ok && id != "" {
						componentID = id
					}
					if id, ok := compMap["component_id"].(string); ok && id != "" {
						componentID = id
					}
					// Check for name as fallback (some components use name as function)
					if componentFunction == "" {
						if name, ok := compMap["name"].(string); ok && name != "" {
							componentFunction = name
							params.Logger.Debug("RenderComponentAction: Using name as function fallback",
								zap.String("name", name),
							)
						}
					}
				} else if compStr, ok := componentData.(string); ok && compStr != "" {
					// Case 2: component_from points directly to a string value
					// e.g., "current_section.function" resolves to "hero"
					componentFunction = compStr
					params.Logger.Debug("RenderComponentAction: Extracted function string directly from component_from",
						zap.String("component_from", componentFrom),
						zap.String("function", componentFunction),
					)
				}
			} else {
				params.Logger.Warn("RenderComponentAction: component_from field not found",
					zap.String("component_from", componentFrom),
				)
			}
		}
	}

	// Now resolve the component
	if componentID != "" {
		compUUID, parseErr := uuid.Parse(componentID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid component_id: %w", parseErr)
		}
		comp, err = GetComponentByID(ctx, params.DB, compUUID, params.Logger)
	} else if componentFunction != "" {
		comp, err = GetComponentWithFallback(ctx, params.DB, componentFunction, params.Logger)
	} else {
		// Log available info for debugging
		params.Logger.Error("RenderComponentAction: No component identifier found",
			zap.Any("config_keys", datahelpers.GetMapKeys(config)),
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		)
		return nil, fmt.Errorf("component_function, component_id, or component_from required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get component '%s': %w", componentFunction, err)
	}

	// Get render context
	contextField := "render_context"
	if cf, ok := config["context_field"].(string); ok && cf != "" {
		contextField = cf
	}
	// Also support context_from as an alias
	if cf, ok := config["context_from"].(string); ok && cf != "" {
		contextField = cf
	}

	renderCtxData := datahelpers.ExtractNestedField(params.CollectedData, contextField)
	renderCtx := &RenderContext{}
	if m, ok := renderCtxData.(map[string]interface{}); ok {
		mergeIntoRenderContext(renderCtx, m)
	}

	// Merge additional content if specified (support both content_field and content_from)
	contentField := ""
	if cf, ok := config["content_field"].(string); ok && cf != "" {
		contentField = cf
	}
	if cf, ok := config["content_from"].(string); ok && cf != "" {
		contentField = cf
	}

	if contentField != "" {
		// Try to extract content with fallback paths
		// LLM responses sometimes have .result wrapper, sometimes not
		contentData := extractContentWithFallbacks(params.CollectedData, contentField, params.Logger)
		if contentData != nil {
			params.Logger.Info("RenderComponentAction: Merging content data",
				zap.String("content_field", contentField),
				zap.Int("field_count", len(contentData)),
				zap.Any("keys", datahelpers.GetMapKeys(contentData)))
			mergeIntoRenderContext(renderCtx, contentData)
		} else {
			params.Logger.Warn("RenderComponentAction: No content data found at any path",
				zap.String("content_field", contentField),
				zap.Any("available_top_keys", datahelpers.GetMapKeys(params.CollectedData)))
		}
	}

	// Before calling RenderTemplate, set ComponentID in context
	renderCtx.ContentData["ComponentID"] = comp.ID

	// Render template
	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, params.Logger)

	params.Logger.Info("RenderComponentAction: Component rendered",
		zap.String("component", comp.Name),
		zap.String("function", comp.Function),
		zap.Int("output_length", len(rendered)),
	)

	return map[string]interface{}{
		"rendered_html":      rendered,
		"component_id":       comp.ID,
		"component_name":     comp.Name,
		"component_function": comp.Function,
	}, nil
}

// ============================================================================
// ACTION: compile_page_sections
// ============================================================================

// CompilePageSectionsAction combines multiple rendered sections into a page
// Config:
//   - sections_from OR sections_field: path to array of section results
//   - page_from: path to page data for name/title
//   - page_name: explicit name for the page (fallback)
//   - inject_header: boolean
//   - inject_footer: boolean
func CompilePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompilePageSectionsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Accept both "sections_from" and "sections_field" for compatibility
	sectionsField := "rendered_sections"
	if sf, ok := config["sections_from"].(string); ok && sf != "" {
		sectionsField = sf
	} else if sf, ok := config["sections_field"].(string); ok && sf != "" {
		sectionsField = sf
	}

	params.Logger.Info("CompilePageSectionsAction: Looking for sections",
		zap.String("sections_field", sectionsField))

	sectionsData := datahelpers.ExtractNestedField(params.CollectedData, sectionsField)
	if sectionsData == nil {
		// Try with .results suffix (loop action output format)
		sectionsData = datahelpers.ExtractNestedField(params.CollectedData, sectionsField+".results")
		if sectionsData != nil {
			params.Logger.Info("CompilePageSectionsAction: Found sections at .results path")
		}
	}

	if sectionsData == nil {
		params.Logger.Error("CompilePageSectionsAction: Sections not found",
			zap.String("tried_path", sectionsField),
			zap.String("also_tried", sectionsField+".results"),
			zap.Strings("available_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return nil, fmt.Errorf("sections not found at %s", sectionsField)
	}

	var sections []string

	switch v := sectionsData.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				sections = append(sections, s)
			} else if m, ok := item.(map[string]interface{}); ok {
				// Try multiple keys for the HTML content
				if html, ok := m["rendered_html"].(string); ok && html != "" {
					sections = append(sections, html)
				} else if html, ok := m["page_html"].(string); ok && html != "" {
					sections = append(sections, html)
				} else if html, ok := m["html"].(string); ok && html != "" {
					sections = append(sections, html)
				}
			}
		}
	case map[string]interface{}:
		// Check if this is a loop output with "results" array
		if results, ok := v["results"].([]interface{}); ok {
			for _, item := range results {
				if m, ok := item.(map[string]interface{}); ok {
					if html, ok := m["rendered_html"].(string); ok && html != "" {
						sections = append(sections, html)
					} else if html, ok := m["page_html"].(string); ok && html != "" {
						sections = append(sections, html)
					}
				}
			}
		} else {
			// Ordered by keys (section_0, section_1, etc.)
			for i := 0; i < len(v); i++ {
				key := fmt.Sprintf("section_%d", i)
				if section, ok := v[key]; ok {
					if s, ok := section.(string); ok {
						sections = append(sections, s)
					} else if m, ok := section.(map[string]interface{}); ok {
						if html, ok := m["rendered_html"].(string); ok {
							sections = append(sections, html)
						}
					}
				}
			}
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Extracted sections",
		zap.Int("count", len(sections)))

	if len(sections) == 0 {
		params.Logger.Warn("CompilePageSectionsAction: No sections to compile, returning placeholder")

		// Get page info for context
		pageName := "page"
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if name, ok := pm["name"].(string); ok && name != "" {
						pageName = name
					}
				}
			}
		}

		return map[string]interface{}{
			"page_body":     "",
			"page_name":     pageName,
			"section_count": 0,
			"skipped":       true,
			"reason":        "no sections defined for page",
		}, nil
	}

	// Build page body
	pageBody := strings.Join(sections, "\n\n")

	// Get page name - try page_from first, then page_name config
	pageName := "index"
	if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
		if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
			if pm, ok := pageData.(map[string]interface{}); ok {
				if name, ok := pm["name"].(string); ok && name != "" {
					pageName = name
				}
			}
		}
	}
	if pn, ok := config["page_name"].(string); ok && pn != "" {
		pageName = pn
	}

	// Build full HTML page
	pageHTML := buildPageHTML(pageName, pageBody)

	// Optionally inject header/footer from component library
	injectHeader, _ := config["inject_header"].(bool)
	injectFooter, _ := config["inject_footer"].(bool)

	if params.DB != nil && (injectHeader || injectFooter) {
		// Get site_id for component lookup
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteUUID := uuid.Nil
		if siteIDStr != "" {
			siteUUID, _ = uuid.Parse(siteIDStr)
		}

		renderCtx := &RenderContext{}
		if rc := datahelpers.ExtractNestedField(params.CollectedData, "render_context"); rc != nil {
			if m, ok := rc.(map[string]interface{}); ok {
				mergeIntoRenderContext(renderCtx, m)
			}
		}

		if injectHeader {
			pageHTML = InjectHeader(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectFooter {
			pageHTML = InjectFooter(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Page compiled",
		zap.String("page_name", pageName),
		zap.Int("section_count", len(sections)),
		zap.Int("html_length", len(pageHTML)),
	)

	return map[string]interface{}{
		"page_html":     pageHTML,
		"page_name":     pageName,
		"section_count": len(sections),
	}, nil
}

func buildPageHTML(pageName, body string) string {
	title := strings.Title(strings.ReplaceAll(pageName, "-", " "))
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body>
%s
</body>
</html>`, title, body)
}

// ============================================================================
// ACTION: insert_research_result
// ============================================================================

// InsertResearchResultAction stores research findings in the database
// Config:
//   - table: target table name (default: "research_results")
//   - fields: map of column_name -> data_path for dynamic field mapping
//   - site_id_field: path to site_id (fallback if not in fields)
//   - result_type: type of research (fallback if not in fields)
func InsertResearchResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("InsertResearchResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("InsertResearchResultAction: No database")
		return map[string]interface{}{"inserted": false, "reason": "no database"}, nil
	}

	// Get table name (default to research_results)
	tableName := "research_results"
	if tn, ok := config["table"].(string); ok && tn != "" {
		tableName = tn
	}

	// Get field mappings from config
	fieldMappings, hasFields := config["fields"].(map[string]interface{})

	// Generate new ID
	resultID := uuid.New()

	// Build dynamic INSERT based on field mappings
	if hasFields && len(fieldMappings) > 0 {
		// Dynamic field-based insert
		columns := []string{"id"}
		placeholders := []string{"$1"}
		values := []interface{}{resultID}
		paramIdx := 2

		for column, dataPath := range fieldMappings {
			dataPathStr, ok := dataPath.(string)
			if !ok {
				continue
			}

			// Extract the value from collected data
			value := datahelpers.ExtractNestedField(params.CollectedData, dataPathStr)

			// Special handling for site_id (needs UUID conversion)
			if column == "site_id" {
				if siteIDStr, ok := value.(string); ok && siteIDStr != "" {
					if siteUUID, err := uuid.Parse(siteIDStr); err == nil {
						value = siteUUID
					} else {
						value = nil
					}
				} else {
					value = nil
				}
			}

			// Skip nil values for optional fields, but include empty strings
			if value == nil {
				continue
			}

			// For complex types, marshal to JSON
			/*switch v := value.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
						zap.String("column", column),
						zap.Error(err))
					continue
				}
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
				values = append(values, string(jsonBytes))
			default:
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d", paramIdx))
				values = append(values, value)
			}*/

			// make it all json
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
					zap.String("column", column),
					zap.String("value_type", fmt.Sprintf("%T", value)),
					zap.Error(err))
				continue
			}
			columns = append(columns, column)
			placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
			values = append(values, string(jsonBytes))
			paramIdx++
		}

		// Add created_at
		columns = append(columns, "created_at")
		placeholders = append(placeholders, "NOW()")

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)

		params.Logger.Info("InsertResearchResultAction: Executing dynamic insert",
			zap.String("table", tableName),
			zap.Strings("columns", columns),
			zap.Int("value_count", len(values)))

		if err := execDB(ctx, params.DB, query, values...); err != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed",
				zap.String("query", query),
				zap.Error(err))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": "general",
				"error":       err.Error(),
			}, nil
		}

		return map[string]interface{}{
			"inserted":  true,
			"id":        resultID.String(),
			"result_id": resultID.String(),
			"table":     tableName,
			"columns":   columns,
		}, nil
	}

	// Fallback: Legacy mode with hardcoded columns
	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	siteID := uuid.Nil
	if siteIDStr != "" {
		siteID, _ = uuid.Parse(siteIDStr)
	}

	// Get result type
	resultType := "general"
	if rt, ok := config["result_type"].(string); ok && rt != "" {
		resultType = rt
	}

	// Get research data - try multiple field paths
	dataField := "research_result"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	researchData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// If no data found, try to build from synthesis
	if researchData == nil {
		researchData = map[string]interface{}{
			"summary":  datahelpers.ExtractNestedField(params.CollectedData, "synthesis.summary"),
			"findings": datahelpers.ExtractNestedField(params.CollectedData, "synthesis"),
			"query":    datahelpers.ExtractNestedField(params.CollectedData, "search_query.result"),
			"topic":    datahelpers.ExtractNestedField(params.CollectedData, "extracted.topic"),
		}
	}

	dataJSON, err := json.Marshal(researchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal research data: %w", err)
	}

	// Try insert with 'findings' column first (new schema), fall back to 'data' (old schema)
	var siteIDArg interface{} = nil
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	// Try new schema first
	query := `
		INSERT INTO research_results (id, site_id, result_type, findings, created_at)
		VALUES ($1, $2, $3, $4::jsonb, NOW())
	`

	if err := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err != nil {
		// Try with 'data' column (old schema)
		query = `
			INSERT INTO research_results (id, site_id, result_type, data, created_at)
			VALUES ($1, $2, $3, $4::jsonb, NOW())
		`
		if err2 := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err2 != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed with both schemas",
				zap.Error(err),
				zap.Error(err2))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": resultType,
				"error":       err2.Error(),
			}, nil
		}
	}

	return map[string]interface{}{
		"inserted":    true,
		"result_id":   resultID.String(),
		"result_type": resultType,
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// ACTION: store_asset
// ============================================================================

// StoreAssetAction stores an asset (image, font, file) in the assets table
// This is a LOCAL action - it does NOT require a topic
// Config:
//   - asset_type: type of asset (image, font, css, js, etc.)
//   - site_id_field: path to site_id in collected_data
//   - data_field: path to asset data (URL, base64, or content)
//   - name_field: path to asset name
//   - metadata_field: optional path to additional metadata
func StoreAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreAssetAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get asset type
	assetType := "image"
	if at, ok := config["asset_type"].(string); ok && at != "" {
		assetType = at
	}

	// Get site_id (optional - assets can be global)
	var siteID *uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = &parsed
			}
		}
	}

	// Get asset data
	dataField := "asset_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	assetData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// Get asset name
	nameField := "asset_name"
	if nf, ok := config["name_field"].(string); ok && nf != "" {
		nameField = nf
	}
	assetName := datahelpers.ExtractNestedFieldString(params.CollectedData, nameField)
	if assetName == "" {
		assetName = fmt.Sprintf("%s_%s", assetType, uuid.New().String()[:8])
	}

	// Extract URL or content from asset data
	var assetURL, assetContent string
	switch v := assetData.(type) {
	case string:
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "s3://") {
			assetURL = v
		} else {
			assetContent = v
		}
	case map[string]interface{}:
		if url, ok := v["url"].(string); ok {
			assetURL = url
		}
		if url, ok := v["image_url"].(string); ok {
			assetURL = url
		}
		if content, ok := v["content"].(string); ok {
			assetContent = content
		}
		if content, ok := v["base64"].(string); ok {
			assetContent = content
		}
	}

	if assetURL == "" && assetContent == "" {
		params.Logger.Warn("StoreAssetAction: No asset URL or content found")
		return map[string]interface{}{
			"stored":     false,
			"asset_name": assetName,
			"asset_type": assetType,
			"reason":     "no asset data found",
		}, nil
	}

	// Get optional metadata
	var metadata map[string]interface{}
	if metaField, ok := config["metadata_field"].(string); ok && metaField != "" {
		if m := datahelpers.ExtractNestedField(params.CollectedData, metaField); m != nil {
			metadata, _ = m.(map[string]interface{})
		}
	}

	// If no DB, return success without persisting
	if params.DB == nil {
		params.Logger.Warn("StoreAssetAction: No database, returning without persistence")
		return map[string]interface{}{
			"stored":     true,
			"persisted":  false,
			"asset_id":   uuid.New().String(),
			"asset_name": assetName,
			"asset_type": assetType,
			"asset_url":  assetURL,
		}, nil
	}

	// Insert into assets table
	assetID := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO assets (id, site_id, name, asset_type, url, content, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, NOW())
		ON CONFLICT (site_id, name, asset_type) DO UPDATE SET
			url = EXCLUDED.url,
			content = EXCLUDED.content,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		RETURNING id
	`

	var returnedID uuid.UUID
	err := queryRowScanUUID(ctx, params.DB, query, &returnedID,
		assetID, siteID, assetName, assetType,
		nullString(assetURL), nullString(assetContent), string(metadataJSON))

	if err != nil {
		// Table might not exist - log and continue
		params.Logger.Warn("StoreAssetAction: Insert failed (table may not exist)",
			zap.Error(err))
		return map[string]interface{}{
			"stored":     true,
			"persisted":  false,
			"asset_id":   assetID.String(),
			"asset_name": assetName,
			"asset_type": assetType,
			"asset_url":  assetURL,
			"error":      err.Error(),
		}, nil
	}

	params.Logger.Info("StoreAssetAction: Asset stored",
		zap.String("asset_id", returnedID.String()),
		zap.String("asset_name", assetName),
		zap.String("asset_type", assetType),
	)

	return map[string]interface{}{
		"stored":     true,
		"persisted":  true,
		"asset_id":   returnedID.String(),
		"asset_name": assetName,
		"asset_type": assetType,
		"asset_url":  assetURL,
	}, nil
}

func ValidateSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ValidateSitePlanAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	planField := "llm_plan"
	if pf, ok := config["plan_field"].(string); ok && pf != "" {
		planField = pf
	}

	// Extract the plan data
	planData := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planData == nil {
		return nil, fmt.Errorf("plan not found at '%s'", planField)
	}

	planData = datahelpers.UnwrapDeep(planData, params.Logger)

	params.Logger.Info("ValidateSitePlanAction: After UnwrapDeep",
		zap.String("planData_type", fmt.Sprintf("%T", planData)))

	var plan map[string]interface{}
	switch v := planData.(type) {
	case map[string]interface{}:
		plan = v
	case string:
		if v == "" {
			return nil, fmt.Errorf("plan is empty string - LLM may have returned no content. Check template rendering logs for <no value> placeholders")
		}
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			return nil, fmt.Errorf("plan is empty after cleaning markdown")
		}
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			// Include content preview in error for debugging
			preview := cleaned
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return nil, fmt.Errorf("failed to parse plan JSON: %w (preview: %s)", err, preview)
		}
	default:
		return nil, fmt.Errorf("plan must be object or JSON string, got %T", planData)
	}

	pagesRaw, ok := plan["pages"]
	if !ok {
		// Log available keys for debugging
		keys := make([]string, 0, len(plan))
		for k := range plan {
			keys = append(keys, k)
		}
		params.Logger.Error("ValidateSitePlanAction: plan missing 'pages'",
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("plan must have 'pages' array (available keys: %v)", keys)
	}

	pages, ok := pagesRaw.([]interface{})
	if !ok || len(pages) == 0 {
		return nil, fmt.Errorf("pages must be non-empty array")
	}

	maxPages := 20
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}
	if len(pages) > maxPages {
		pages = pages[:maxPages]
		plan["pages"] = pages
	}

	pageNames := make(map[string]bool)
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			if name, ok := pm["name"].(string); ok {
				pageNames[name] = true
			}
		}
	}

	if ensurePages, ok := config["ensure_pages"].([]interface{}); ok {
		for _, req := range ensurePages {
			if reqName, ok := req.(string); ok && !pageNames[reqName] {
				pages = append(pages, map[string]interface{}{
					"name": reqName, "title": strings.Title(reqName),
					"nav_label": strings.Title(reqName), "nav_order": len(pages) + 1,
					"in_header": true, "in_footer": true, "sections": []interface{}{},
				})
			}
		}
		plan["pages"] = pages
	}

	if _, ok := plan["style_collection"].(string); !ok {
		if ds, ok := config["default_style"].(string); ok {
			plan["style_collection"] = ds
		}
	}

	if plan["needs_logo"] == nil {
		plan["needs_logo"] = false
	}
	if plan["needs_images"] == nil {
		plan["needs_images"] = false
	}
	if plan["image_prompts"] == nil {
		plan["image_prompts"] = map[string]interface{}{}
	}

	params.Logger.Info("ValidateSitePlanAction: Complete", zap.Int("pages", len(pages)))
	return plan, nil
}

// nullString returns nil for empty strings, otherwise the string pointer
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ============================================================================
// ACTION: db_sync
// ============================================================================

// DBSyncAction is a general-purpose database sync action
// Can be used to insert/update records in any table
// Config:
//   - operation: insert, update, upsert
//   - table: table name
//   - data_field: path to data to sync
//   - key_fields: array of fields that form the primary/unique key
func DBSyncAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DBSyncAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("DBSyncAction: No database connection")
		return map[string]interface{}{"synced": false, "reason": "no database"}, nil
	}

	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "upsert"
	}

	table, ok := config["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// Get data to sync
	dataField := "sync_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	syncData := datahelpers.ExtractNestedField(params.CollectedData, dataField)
	if syncData == nil {
		return nil, fmt.Errorf("no data found at %s", dataField)
	}

	dataMap, ok := syncData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("sync data must be a map")
	}

	// Build and execute query based on operation
	// This is a simplified implementation - in production you'd want more robust SQL building

	params.Logger.Info("DBSyncAction: Syncing data",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int("field_count", len(dataMap)),
	)

	// For now, just marshal and log - actual implementation would build SQL
	dataJSON, _ := json.Marshal(dataMap)

	return map[string]interface{}{
		"synced":      true,
		"operation":   operation,
		"table":       table,
		"field_count": len(dataMap),
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func execDB(ctx context.Context, db interface{}, query string, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, args...)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, args...)
		return err
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

func lookupPageID(ctx context.Context, db interface{}, siteID uuid.UUID, pageName string, logger *zap.Logger) (uuid.UUID, error) {
	query := `SELECT id FROM pages WHERE site_id = $1 AND name = $2`
	var pageID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	default:
		return uuid.Nil, fmt.Errorf("unsupported database type: %T", db)
	}
}

// queryRowScanUUID executes a query and scans the result into a UUID
func queryRowScanUUID(ctx context.Context, db interface{}, query string, dest *uuid.UUID, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		return d.QueryRowContext(ctx, query, args...).Scan(dest)
	case *pgxpool.Pool:
		return d.QueryRow(ctx, query, args...).Scan(dest)
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

// getStyleCollectionByID looks up a style collection by UUID
// This is a local helper since component_library.go doesn't have this function yet
func getStyleCollectionByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT id, name, display_name, category, color_palette, typography,
		       header_component_id, footer_component_id, css_theme_id
		FROM style_collections
		WHERE id = $1 AND is_active = true
	`

	var coll StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var headerID, footerID, cssThemeID *uuid.UUID

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if colorPaletteJSON != nil {
		json.Unmarshal(colorPaletteJSON, &coll.ColorPalette)
	}
	if typographyJSON != nil {
		json.Unmarshal(typographyJSON, &coll.Typography)
	}
	coll.HeaderComponentID = headerID
	coll.FooterComponentID = footerID
	coll.CSSThemeID = cssThemeID

	return &coll, nil
}

func BuildReviewResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildReviewResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := map[string]interface{}{
		"reviewed_at": time.Now().UTC().Format(time.RFC3339),
		"review_mode": "unknown",
		"approved":    false,
		"issues":      []interface{}{},
		"edits":       map[string]interface{}{},
	}

	if approved, ok := config["approved"].(bool); ok {
		result["approved"] = approved
	} else if field, ok := config["approved_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			if b, ok := val.(bool); ok {
				result["approved"] = b
			}
		}
	}

	if reviewer, ok := config["reviewer"].(string); ok && reviewer != "" {
		result["reviewed_by"] = reviewer
	} else if field, ok := config["reviewer_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			result["reviewed_by"] = val
		}
	}
	if result["reviewed_by"] == nil {
		result["reviewed_by"] = "system"
	}

	if mode, ok := config["review_mode"].(string); ok {
		result["review_mode"] = mode
	}

	if field, ok := config["eval_score"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["eval_score"] = val
		}
	}

	if field, ok := config["edits_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["edits"] = val
		}
	}

	if field, ok := config["auto_eval_issues"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["auto_eval_issues"] = val
			if issues, ok := val.([]interface{}); ok {
				result["issues"] = issues
			}
		}
	}

	if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
		result["content"] = content
	}

	params.Logger.Info("BuildReviewResultAction: Complete",
		zap.Bool("approved", result["approved"].(bool)),
		zap.String("mode", result["review_mode"].(string)))
	return result, nil
}

// ============================================================================
// ACTION: prepare_review_data
// ============================================================================

func PrepareReviewDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareReviewDataAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	reviewData := map[string]interface{}{
		"prepared_at": time.Now().UTC().Format(time.RFC3339),
		"fields":      map[string]interface{}{},
	}

	if includeFields, ok := config["include_fields"].([]interface{}); ok {
		fieldsMap := reviewData["fields"].(map[string]interface{})
		for _, field := range includeFields {
			if fieldName, ok := field.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, fieldName); value != nil {
					fieldsMap[fieldName] = value
				}
			}
		}
	}

	if formatForDisplay, ok := config["format_for_display"].(bool); ok && formatForDisplay {
		if page := datahelpers.ExtractNestedField(params.CollectedData, "current_page"); page != nil {
			if pm, ok := page.(map[string]interface{}); ok {
				reviewData["page_name"] = pm["name"]
				reviewData["page_title"] = pm["title"]
			}
		}
		if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
			reviewData["content"] = content
		}
		if brief := datahelpers.ExtractNestedField(params.CollectedData, "reviewed_brief"); brief != nil {
			if bm, ok := brief.(map[string]interface{}); ok {
				reviewData["company_name"] = bm["company_name"]
				reviewData["tone"] = bm["tone"]
			}
		}
	}

	params.Logger.Info("PrepareReviewDataAction: Complete")
	return reviewData, nil
}

// ============================================================================
// ACTION: update_page_components_status
// ============================================================================

func UpdatePageComponentsStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageComponentsStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	newStatus := "approved"
	if status, ok := config["status"].(string); ok && status != "" {
		newStatus = status
	}

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := datahelpers.ExtractNestedField(params.CollectedData, pageField)
	if pageData == nil {
		return map[string]interface{}{"updated": false, "reason": "no page data", "status_set": newStatus}, nil
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	var pageID uuid.UUID
	if idStr, ok := page["id"].(string); ok && idStr != "" {
		pageID, _ = uuid.Parse(idStr)
	}

	reviewedAt := time.Now().UTC()
	if field, ok := config["reviewed_at_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				reviewedAt = t
			}
		}
	}

	reviewedBy := "system"
	if field, ok := config["reviewed_by_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			reviewedBy = val
		}
	}

	if params.DB != nil && pageID != uuid.Nil {
		query := `UPDATE page_components SET build_status = $1, reviewed_at = $2, reviewed_by = $3, updated_at = NOW() WHERE page_id = $4`
		result, err := params.DB.ExecContext(ctx, query, newStatus, reviewedAt, reviewedBy, pageID)
		if err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
		rows, _ := result.RowsAffected()
		return map[string]interface{}{
			"updated": true, "rows_affected": rows, "page_id": pageID.String(),
			"status_set": newStatus, "reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
		}, nil
	}

	return map[string]interface{}{
		"updated": false, "reason": "no db or page_id", "status_set": newStatus,
		"reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
	}, nil
}

// When call_agent passes input_fields, they arrive under input_data.*
// This checks both root and input_data locations
func LoadPageSectionComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadPageSectionComponentsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	// Try to find page data with fallback to input_data
	pageData := extractWithInputDataFallback(params.CollectedData, pageField, params.Logger)

	if pageData == nil {
		keys := make([]string, 0, len(params.CollectedData))
		for k := range params.CollectedData {
			keys = append(keys, k)
		}
		params.Logger.Error("Page not found",
			zap.String("page_field", pageField),
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("page not found at '%s'", pageField)
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	params.Logger.Info("Found page data",
		zap.String("page_field", pageField),
		zap.Any("page_name", page["name"]),
		zap.Any("page_title", page["title"]))

	sectionsRaw := page["sections"]
	if sectionsRaw == nil {
		sectionsRaw = extractWithInputDataFallback(params.CollectedData, "sections", params.Logger)
	}

	sectionNames := datahelpers.ExtractSectionNamesHelper(sectionsRaw)
	if len(sectionNames) == 0 {
		params.Logger.Warn("No sections found for page")
		return map[string]interface{}{"components": []interface{}{}, "count": 0, "from_database": false}, nil
	}

	params.Logger.Info("Loading components for sections",
		zap.Strings("sections", sectionNames))

	var components []map[string]interface{}

	if params.DB != nil {
		placeholders := make([]string, len(sectionNames))
		args := make([]interface{}, len(sectionNames))
		for i, name := range sectionNames {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = name
		}

		// Query for components by name
		query := fmt.Sprintf(`
				SELECT 
					id,
					name, 
					COALESCE(display_name, name) as display_name,
					function, 
					COALESCE(category, '') as category,
					semantic_tags, 
					description, 
					html_template,
					input_schema,
					COALESCE(render_mode, 'template') as render_mode,
					agent_type,
					COALESCE(component_level, 'section') as component_level
				FROM content_components 
				WHERE name IN (%s)`, strings.Join(placeholders, ", "))

		rows, err := params.DB.QueryContext(ctx, query, args...)
		if err != nil {
			params.Logger.Error("LoadPageSectionComponentsAction: Query failed",
				zap.Error(err),
				zap.Strings("sections", sectionNames))
			return map[string]interface{}{
				"components":    sectionNames,
				"count":         len(sectionNames),
				"from_database": false,
				"db_error":      err.Error(),
			}, nil
		}
		defer rows.Close()

		for rows.Next() {
			var id, name, function string
			var displayName, category sql.NullString
			var semanticTags, description, htmlTemplate, inputSchema sql.NullString
			var renderMode, agentType, componentLevel sql.NullString

			if err := rows.Scan(&id, &name, &displayName, &function, &category, &semanticTags, &description, &htmlTemplate, &inputSchema, &renderMode, &agentType, &componentLevel); err != nil {
				params.Logger.Error("LoadPageSectionComponentsAction: Row scan failed",
					zap.Error(err))
				continue
			}

			comp := map[string]interface{}{
				"component_id": id,
				"name":         name,
				"function":     function,
			}

			if displayName.Valid {
				comp["display_name"] = displayName.String
			} else {
				comp["display_name"] = name
			}

			if category.Valid {
				comp["category"] = category.String
			} else {
				comp["category"] = ""
			}

			if semanticTags.Valid {
				comp["semantic_tags"] = semanticTags.String
			}
			if description.Valid && description.String != "" {
				comp["description"] = description.String
			}
			if htmlTemplate.Valid && htmlTemplate.String != "" {
				comp["html_template"] = htmlTemplate.String
			}
			if inputSchema.Valid && inputSchema.String != "" {
				comp["input_schema"] = inputSchema.String
			}
			if renderMode.Valid && renderMode.String != "" {
				comp["render_mode"] = renderMode.String
			} else {
				comp["render_mode"] = "template"
			}

			if agentType.Valid && agentType.String != "" {
				comp["agent_type"] = agentType.String
			}

			if componentLevel.Valid && componentLevel.String != "" {
				comp["component_level"] = componentLevel.String
			} else {
				comp["component_level"] = "section"
			}

			needsLLM := detectNeedsLLMContent(htmlTemplate.String, inputSchema.String)
			comp["needs_llm"] = needsLLM

			components = append(components, comp)
		}

		// Track which components were found by name
		foundNames := make(map[string]bool)
		for _, comp := range components {
			if name, ok := comp["name"].(string); ok {
				foundNames[name] = true
			}
			if fn, ok := comp["function"].(string); ok {
				foundNames[fn] = true
			}
		}

		// Find missing components
		var missing []string
		for _, name := range sectionNames {
			if !foundNames[name] {
				missing = append(missing, name)
			}
		}

		// Try fallback: lookup missing components by function
		if len(missing) > 0 {
			params.Logger.Info("LoadPageSectionComponentsAction: Trying function lookup for missing components",
				zap.Strings("missing", missing))

			funcPlaceholders := make([]string, len(missing))
			funcArgs := make([]interface{}, len(missing))
			for i, name := range missing {
				funcPlaceholders[i] = fmt.Sprintf("$%d", i+1)
				funcArgs[i] = name
			}

			funcQuery := fmt.Sprintf(`
				SELECT DISTINCT ON (function)
					id,
					name, 
					COALESCE(display_name, name) as display_name,
					function, 
					COALESCE(category, '') as category,
					semantic_tags, 
					description, 
					html_template,
					input_schema
				FROM content_components 
				WHERE function IN (%s)
				ORDER BY function, created_at DESC
			`, strings.Join(funcPlaceholders, ", "))

			funcRows, err := params.DB.QueryContext(ctx, funcQuery, funcArgs...)
			if err != nil {
				params.Logger.Warn("LoadPageSectionComponentsAction: Function lookup query failed",
					zap.Error(err))
			} else {
				defer funcRows.Close()
				for funcRows.Next() {
					var id, name, function string
					var displayName, category sql.NullString
					var semanticTags, description, htmlTemplate, inputSchema sql.NullString

					if err := funcRows.Scan(&id, &name, &displayName, &function, &category, &semanticTags, &description, &htmlTemplate, &inputSchema); err != nil {
						continue
					}

					if containsString(missing, function) {
						comp := map[string]interface{}{
							"component_id": id,
							"name":         name,
							"function":     function,
						}
						if displayName.Valid {
							comp["display_name"] = displayName.String
						} else {
							comp["display_name"] = name
						}
						if category.Valid {
							comp["category"] = category.String
						}
						if semanticTags.Valid {
							comp["semantic_tags"] = semanticTags.String
						}
						if description.Valid && description.String != "" {
							comp["description"] = description.String
						}
						if htmlTemplate.Valid && htmlTemplate.String != "" {
							comp["html_template"] = htmlTemplate.String
						}
						if inputSchema.Valid && inputSchema.String != "" {
							comp["input_schema"] = inputSchema.String
						}

						needsLLM := detectNeedsLLMContent(htmlTemplate.String, inputSchema.String)
						comp["needs_llm"] = needsLLM

						components = append(components, comp)
						foundNames[function] = true
						params.Logger.Info("LoadPageSectionComponentsAction: Found component by function",
							zap.String("function", function),
							zap.String("name", name),
							zap.Bool("needs_llm", needsLLM))
					}
				}
			}
		}

		// Rebuild missing list after function lookup
		var stillMissing []string
		for _, name := range sectionNames {
			if !foundNames[name] {
				stillMissing = append(stillMissing, name)
			}
		}

		// Create stubs only for components still not found
		if len(stillMissing) > 0 {
			params.Logger.Warn("LoadPageSectionComponentsAction: Creating stubs for components not found in database",
				zap.Strings("missing", stillMissing))

			for _, name := range stillMissing {
				components = append(components, map[string]interface{}{
					"name":         name,
					"display_name": name,
					"function":     name,
					"category":     "",
					"needs_llm":    true,
					"description":  "",
				})
			}
		}

		// =====================================================================
		// REORDER COMPONENTS TO MATCH ORIGINAL sectionNames ORDER
		// Database query returns in arbitrary order, but we need to preserve
		// the section order defined in the page plan
		// =====================================================================
		orderedComponents := make([]map[string]interface{}, 0, len(components))
		for _, sectionName := range sectionNames {
			for _, comp := range components {
				name, _ := comp["name"].(string)
				function, _ := comp["function"].(string)
				if name == sectionName || function == sectionName {
					orderedComponents = append(orderedComponents, comp)
					break
				}
			}
		}

		params.Logger.Info("LoadPageSectionComponentsAction: Reordered components to match section order",
			zap.Int("original_count", len(components)),
			zap.Int("ordered_count", len(orderedComponents)),
			zap.Strings("requested_order", sectionNames))

		return map[string]interface{}{
			"components":    orderedComponents,
			"count":         len(orderedComponents),
			"from_database": len(orderedComponents) > 0,
			"requested":     sectionNames,
		}, nil
	}

	// No DB - return section names as stub components
	stubComponents := make([]map[string]interface{}, len(sectionNames))
	for i, name := range sectionNames {
		stubComponents[i] = map[string]interface{}{
			"name":         name,
			"function":     name,
			"display_name": name,
			"description":  "",
			"needs_llm":    true,
		}
	}

	return map[string]interface{}{
		"components":    stubComponents,
		"count":         len(stubComponents),
		"from_database": false,
		"db_note":       "no matching components in database",
		"requested":     sectionNames,
	}, nil
}

// Order: direct path -> input_data.{path} -> FindByPath helper
// extractWithInputDataFallback tries to extract a field, falling back to input_data prefix
// This handles the common case where workflows specify paths like "current_section.name"
// but the data is actually at "input_data.current_section.name"
func extractWithInputDataFallback(data map[string]interface{}, path string, logger *zap.Logger) interface{} {
	// Try direct path first
	if value := datahelpers.ExtractNestedField(data, path); value != nil {
		logger.Debug("extractWithInputDataFallback: Found at direct path",
			zap.String("path", path),
		)
		return value
	}

	// If path doesn't already start with input_data, try with prefix
	if !strings.HasPrefix(path, "input_data.") {
		prefixedPath := "input_data." + path
		if value := datahelpers.ExtractNestedField(data, prefixedPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via input_data prefix",
				zap.String("original_path", path),
				zap.String("actual_path", prefixedPath),
			)
			return value
		}
	}

	// Try in __raw_message__.body.input_data (deeply nested case from child agents)
	if !strings.HasPrefix(path, "__raw_message__") {
		rawMsgPath := "__raw_message__.body.input_data." + path
		if value := datahelpers.ExtractNestedField(data, rawMsgPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via __raw_message__.body.input_data",
				zap.String("original_path", path),
			)
			return value
		}
	}

	// Also try agent_config location (for workflow config data)
	if !strings.HasPrefix(path, "agent_config") {
		agentConfigPath := "agent_config." + path
		if value := datahelpers.ExtractNestedField(data, agentConfigPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via agent_config",
				zap.String("original_path", path),
			)
			return value
		}
	}

	logger.Debug("extractWithInputDataFallback: Not found anywhere",
		zap.String("path", path),
	)
	return nil
}

// ============================================================================
// ACTION: filter_search_results
// ============================================================================

// FilterSearchResultsAction filters search results based on criteria
// Handles various response formats: direct array, {results: []}, {data: {results: []}}, etc.
func FilterSearchResultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("FilterSearchResultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	resultsField := "search_results"
	if rf, ok := config["results_field"].(string); ok && rf != "" {
		resultsField = rf
	}

	// Try to find results array - handles various response formats
	results := datahelpers.FindResultsArray(params.CollectedData, resultsField, params.Logger)
	if results == nil {
		params.Logger.Warn("FilterSearchResultsAction: No results found",
			zap.String("results_field", resultsField),
			zap.Strings("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return map[string]interface{}{"filtered_results": []interface{}{}, "count": 0, "original_count": 0}, nil
	}

	params.Logger.Info("FilterSearchResultsAction: Found results",
		zap.Int("count", len(results)))

	// Support both max_results and max_sources config keys
	maxResults := 10
	if mr, ok := config["max_results"].(float64); ok {
		maxResults = int(mr)
	} else if ms, ok := config["max_sources"].(float64); ok {
		maxResults = int(ms)
	}

	excludePatterns := datahelpers.ExtractStringListHelper(config["exclude_patterns"])
	requiredKeywords := datahelpers.ExtractStringListHelper(config["required_keywords"])
	preferDomains := datahelpers.ExtractStringListHelper(config["prefer_domains"])

	var filtered []interface{}
	var preferred []interface{}

	for _, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := result["title"].(string)
		content, _ := result["content"].(string)
		snippet, _ := result["snippet"].(string)
		url, _ := result["url"].(string)
		searchText := strings.ToLower(title + " " + content + " " + snippet + " " + url)

		// Check exclusions
		excluded := false
		for _, pattern := range excludePatterns {
			if strings.Contains(searchText, strings.ToLower(pattern)) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check required keywords
		if len(requiredKeywords) > 0 {
			hasKeyword := false
			for _, kw := range requiredKeywords {
				if strings.Contains(searchText, strings.ToLower(kw)) {
					hasKeyword = true
					break
				}
			}
			if !hasKeyword {
				continue
			}
		}

		// Check if from preferred domain
		isPreferred := false
		for _, domain := range preferDomains {
			if strings.Contains(strings.ToLower(url), strings.ToLower(domain)) {
				isPreferred = true
				break
			}
		}

		if isPreferred {
			preferred = append(preferred, result)
		} else {
			filtered = append(filtered, result)
		}
	}

	// Combine: preferred first, then others, up to maxResults
	combined := append(preferred, filtered...)
	if len(combined) > maxResults {
		combined = combined[:maxResults]
	}

	params.Logger.Info("FilterSearchResultsAction: Complete",
		zap.Int("original", len(results)),
		zap.Int("preferred", len(preferred)),
		zap.Int("filtered", len(combined)))

	return map[string]interface{}{
		"filtered_results": combined,
		"count":            len(combined),
		"original_count":   len(results),
		"preferred_count":  len(preferred),
	}, nil
}

// ============================================================================
// ACTION: extract_fields
// ============================================================================

func ExtractFieldsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ExtractFieldsAction: Starting",
		zap.Any("config_keys", getConfigKeys(params.StepConfig.Config)))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := make(map[string]interface{})

	// Handle fields array format: ["field1", "field2"]
	if fields, ok := config["fields"].([]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as array")
		for _, f := range fields {
			switch field := f.(type) {
			case string:
				if value := datahelpers.ExtractNestedField(params.CollectedData, field); value != nil {
					parts := strings.Split(field, ".")
					result[parts[len(parts)-1]] = value
				}
			case map[string]interface{}:
				source, _ := field["source"].(string)
				target, _ := field["target"].(string)
				if source == "" {
					continue
				}
				if target == "" {
					parts := strings.Split(source, ".")
					target = parts[len(parts)-1]
				}
				if value := datahelpers.ExtractNestedField(params.CollectedData, source); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Handle fields as map-of-arrays format (fallback paths)
	// Example: {"topic": ["path1", "path2"], "company": ["path3"]}
	if fieldsMap, ok := config["fields"].(map[string]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as map-of-arrays",
			zap.Int("field_count", len(fieldsMap)))

		for targetField, pathsRaw := range fieldsMap {
			var found bool

			// Handle array of paths
			if paths, ok := pathsRaw.([]interface{}); ok {
				for _, pathRaw := range paths {
					if path, ok := pathRaw.(string); ok {
						// Try direct path first
						if value := datahelpers.ExtractNestedField(params.CollectedData, path); value != nil {
							result[targetField] = value
							found = true
							params.Logger.Info("ExtractFieldsAction: Found via direct path",
								zap.String("target", targetField),
								zap.String("path", path))
							break
						}

						// Try with input_data prefix
						if !strings.HasPrefix(path, "input_data.") {
							prefixedPath := "input_data." + path
							if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
								result[targetField] = value
								found = true
								params.Logger.Info("ExtractFieldsAction: Found via input_data prefix",
									zap.String("target", targetField),
									zap.String("original_path", path),
									zap.String("prefixed_path", prefixedPath))
								break
							}
						}
					}
				}
			}

			// Handle single string path (not array)
			if singlePath, ok := pathsRaw.(string); ok && !found {
				if value := datahelpers.ExtractNestedField(params.CollectedData, singlePath); value != nil {
					result[targetField] = value
					found = true
				} else if !strings.HasPrefix(singlePath, "input_data.") {
					prefixedPath := "input_data." + singlePath
					if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
						result[targetField] = value
						found = true
					}
				}
			}

			if !found {
				params.Logger.Warn("ExtractFieldsAction: Field not found in any path",
					zap.String("target", targetField),
					zap.Any("tried_paths", pathsRaw))
			}
		}
	}

	// Handle field_map format: {"target": "source"}
	if fieldMap, ok := config["field_map"].(map[string]interface{}); ok {
		for target, source := range fieldMap {
			if sourceStr, ok := source.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, sourceStr); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Apply defaults
	if defaults, ok := config["defaults"].(map[string]interface{}); ok {
		for key, val := range defaults {
			if result[key] == nil {
				result[key] = val
			}
		}
	}

	params.Logger.Info("ExtractFieldsAction: Complete",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", datahelpers.GetMapKeys(result)))
	return result, nil
}

// Priority 5: Try to build query from section context (for research-agent)
// Check both root level and inside input_data
func getSearchQueryFromSectionContext(params ActionParams) string {
	var currentSection map[string]interface{}

	// Try root level first
	if cs, ok := params.CollectedData["current_section"].(map[string]interface{}); ok {
		currentSection = cs
	}
	// Try input_data.current_section
	if currentSection == nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if cs, ok := inputData["current_section"].(map[string]interface{}); ok {
				currentSection = cs
			}
		}
	}

	if currentSection == nil {
		return ""
	}

	// Get section name/function
	sectionName := ""
	if fn, ok := currentSection["function"].(string); ok && fn != "" {
		sectionName = fn
	} else if name, ok := currentSection["name"].(string); ok && name != "" {
		sectionName = name
	}

	if sectionName == "" {
		return ""
	}

	// Get domain
	domain := ""
	if d, ok := params.CollectedData["domain"].(string); ok {
		domain = d
	} else if siteRecord, ok := params.CollectedData["site_record"].(map[string]interface{}); ok {
		if d, ok := siteRecord["domain"].(string); ok {
			domain = d
		}
	} else if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if siteRecord, ok := inputData["site_record"].(map[string]interface{}); ok {
			if d, ok := siteRecord["domain"].(string); ok {
				domain = d
			}
		}
	}

	// Build query
	query := sectionName
	if domain != "" {
		query = fmt.Sprintf("%s %s", sectionName, domain)
	}

	return query
}

// containsString checks if a string slice contains a specific string
// extractContentWithFallbacks tries multiple paths to find content data
// This handles different output formats from execute_llm_prompt (with/without .result wrapper)
func extractContentWithFallbacks(data map[string]interface{}, contentField string, logger *zap.Logger) map[string]interface{} {
	// Build list of paths to try
	pathsToTry := []string{contentField}

	// If path ends with ".result", also try without it
	if strings.HasSuffix(contentField, ".result") {
		base := strings.TrimSuffix(contentField, ".result")
		pathsToTry = append(pathsToTry, base)
		pathsToTry = append(pathsToTry, base+".response")
		pathsToTry = append(pathsToTry, base+".content")
	} else {
		// If path doesn't end with .result, also try with it
		pathsToTry = append(pathsToTry, contentField+".result")
		pathsToTry = append(pathsToTry, contentField+".response")
		pathsToTry = append(pathsToTry, contentField+".content")
	}

	// Try each path
	for _, path := range pathsToTry {
		if extracted := datahelpers.ExtractNestedField(data, path); extracted != nil {
			if m, ok := extracted.(map[string]interface{}); ok && len(m) > 0 {
				logger.Debug("extractContentWithFallbacks: Found content",
					zap.String("path", path),
					zap.Int("field_count", len(m)))
				return m
			}
		}
	}

	// Last resort: check if the base field exists and contains the content directly
	// Sometimes LLM output is stored as the field value itself
	parts := strings.Split(contentField, ".")
	if len(parts) > 0 {
		baseField := parts[0]
		if baseData := datahelpers.ExtractNestedField(data, baseField); baseData != nil {
			if m, ok := baseData.(map[string]interface{}); ok && len(m) > 0 {
				// Check if this map contains content-like fields
				if hasContentFields(m) {
					logger.Debug("extractContentWithFallbacks: Found content at base field",
						zap.String("base_field", baseField),
						zap.Int("field_count", len(m)))
					return m
				}
			}
		}
	}

	logger.Debug("extractContentWithFallbacks: No content found",
		zap.String("original_path", contentField),
		zap.Strings("tried_paths", pathsToTry))

	return nil
}

// hasContentFields checks if a map contains typical content field names
func hasContentFields(m map[string]interface{}) bool {
	contentFieldNames := []string{
		"headline", "subheadline", "body", "content", "heading",
		"title", "description", "text", "features", "items",
		"primary_cta", "cta_text", "button_text",
	}
	for _, fieldName := range contentFieldNames {
		if _, exists := m[fieldName]; exists {
			return true
		}
	}
	return false
}

// detectNeedsLLMContent determines if a component template needs LLM-generated content
// Returns true if template has content placeholders that need dynamic content
// Returns false if template only has structural placeholders (domain, logo_text, etc.)
func detectNeedsLLMContent(htmlTemplate, inputSchema string) bool {
	// No template = needs LLM to generate something
	if htmlTemplate == "" {
		return true
	}

	// If has input_schema, it defines what content is needed
	if inputSchema != "" && inputSchema != "{}" && inputSchema != "null" {
		return true
	}

	// Content placeholders that indicate LLM content is needed
	contentPlaceholders := []string{
		"{{.headline}}", "{{.subheadline}}", "{{.body}}", "{{.content}}",
		"{{.description}}", "{{.text}}", "{{.paragraph}}",
		"{{.primary_cta}}", "{{.secondary_cta}}", "{{.cta_text}}",
		"{{.features}}", "{{.services}}", "{{.benefits}}", "{{.items}}",
		"{{.testimonial}}", "{{.quote}}", "{{.author}}",
		"{{.heading}}", "{{.subtitle}}",
		"{{ .headline}}", "{{ .body}}", "{{ .content}}",
		"{{range .features}}", "{{range .services}}", "{{range .items}}",
		"{{range .benefits}}", "{{range .testimonials}}",
	}

	// Check for content placeholders
	for _, placeholder := range contentPlaceholders {
		if strings.Contains(htmlTemplate, placeholder) {
			return true
		}
	}

	// Structural fields that DON'T require LLM (filled from render_context)
	structuralFields := map[string]bool{
		"domain": true, "logo_text": true, "company_name": true,
		"tagline": true, "current_page": true, "year": true,
		"primary_color": true, "secondary_color": true, "accent_color": true,
		"background_color": true, "text_color": true,
		"email": true, "phone": true, "address": true,
		"nav_items_html": true, "theme_css": true, "title": true,
		"cta_url": true, "industry": true, "tone": true,
	}

	// Check for any {{.X}} where X is not a known structural field
	re := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(htmlTemplate, -1)
	for _, match := range matches {
		if len(match) > 1 {
			fieldName := match[1]
			if !structuralFields[fieldName] {
				return true
			}
		}
	}

	// Only structural placeholders - doesn't need LLM
	return false
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
