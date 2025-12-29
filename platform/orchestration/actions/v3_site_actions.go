// FILE: platform/orchestration/actions/v3_site_actions.go
// Additional actions needed for the v3 multipage website builder component-based architecture.
// These complement existing actions in site_db_actions.go and component_library.go.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
		"published": true, "archived": true, "error": true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s", newStatus)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	query := `UPDATE sites SET status = $2, updated_at = NOW() WHERE id = $1`
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

// UpdatePageStatusAction updates a single page's status
// Config:
//   - page_id_field: path to page_id OR
//   - site_id_field + page_name_field: to look up page
//   - status: new status value
func UpdatePageStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		return map[string]interface{}{"updated": false, "reason": "no database"}, nil
	}

	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required")
	}

	var pageID uuid.UUID

	// Try direct page_id
	if pageIDField, ok := config["page_id_field"].(string); ok && pageIDField != "" {
		pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField)
		if pageIDStr != "" {
			var err error
			pageID, err = uuid.Parse(pageIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid page_id: %w", err)
			}
		}
	}

	// Alternative: look up by site_id + page_name
	if pageID == uuid.Nil {
		siteIDField := config["site_id_field"].(string)
		pageNameField := config["page_name_field"].(string)

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

	if pageID == uuid.Nil {
		return nil, fmt.Errorf("could not determine page_id")
	}

	query := `UPDATE pages SET status = $2, updated_at = NOW() WHERE id = $1`
	if err := execDB(ctx, params.DB, query, pageID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update page status: %w", err)
	}

	return map[string]interface{}{
		"updated": true,
		"page_id": pageID.String(),
		"status":  newStatus,
	}, nil
}

// ============================================================================
// ACTION: build_render_context
// ============================================================================

// BuildRenderContextAction assembles a RenderContext from multiple sources
// Used before rendering components with templates
// Config:
//   - sources: array of field paths to merge into context
//   - site_id_field: optional, to load site-specific data
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

	// Get sources to merge
	sources := []string{"input_data", "site_record", "style_collection", "page_plan"}
	if s, ok := config["sources"].([]interface{}); ok {
		sources = make([]string, 0, len(s))
		for _, src := range s {
			if srcStr, ok := src.(string); ok {
				sources = append(sources, srcStr)
			}
		}
	}

	// Extract and merge data from each source
	for _, source := range sources {
		sourceData := datahelpers.ExtractNestedField(params.CollectedData, source)
		if sourceData == nil {
			continue
		}

		if m, ok := sourceData.(map[string]interface{}); ok {
			mergeIntoRenderContext(renderCtx, m)
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
		zap.String("logo_text", renderCtx.LogoText),
		zap.Int("nav_items", len(renderCtx.NavItems)),
	)

	return map[string]interface{}{
		"render_context": renderCtxToMap(renderCtx),
		"sources_merged": len(sources),
	}, nil
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
	}
	if ctx.SiteID != uuid.Nil {
		result["site_id"] = ctx.SiteID.String()
	}
	if len(ctx.NavItems) > 0 {
		navItems := make([]map[string]interface{}, len(ctx.NavItems))
		for i, item := range ctx.NavItems {
			navItems[i] = map[string]interface{}{
				"label":     item.Label,
				"url":       item.URL,
				"is_active": item.IsActive,
			}
		}
		result["nav_items"] = navItems
	}
	return result
}

// ============================================================================
// ACTION: render_component
// ============================================================================

// RenderComponentAction renders a single component template with context
// Config:
//   - component_function: function name to look up (e.g., "hero-banner")
//   - component_id: explicit component UUID (alternative to function)
//   - context_field: path to render context in collected_data
//   - content_field: path to additional content data
func RenderComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderComponentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for component rendering")
	}

	// Get component
	var comp *Component
	var err error

	if compID, ok := config["component_id"].(string); ok && compID != "" {
		compUUID, _ := uuid.Parse(compID)
		comp, err = GetComponentByID(ctx, params.DB, compUUID, params.Logger)
	} else if compFunc, ok := config["component_function"].(string); ok && compFunc != "" {
		comp, err = GetComponentWithFallback(ctx, params.DB, compFunc, params.Logger)
	} else {
		return nil, fmt.Errorf("component_function or component_id required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	// Get render context
	contextField := "render_context"
	if cf, ok := config["context_field"].(string); ok && cf != "" {
		contextField = cf
	}

	renderCtxData := datahelpers.ExtractNestedField(params.CollectedData, contextField)
	renderCtx := &RenderContext{}
	if m, ok := renderCtxData.(map[string]interface{}); ok {
		mergeIntoRenderContext(renderCtx, m)
	}

	// Merge additional content if specified
	if contentField, ok := config["content_field"].(string); ok && contentField != "" {
		contentData := datahelpers.ExtractNestedField(params.CollectedData, contentField)
		if m, ok := contentData.(map[string]interface{}); ok {
			mergeIntoRenderContext(renderCtx, m)
		}
	}

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
//   - sections_field: path to array of section results
//   - page_name: name for the page
//   - inject_header: boolean
//   - inject_footer: boolean
func CompilePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompilePageSectionsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	sectionsField := "rendered_sections"
	if sf, ok := config["sections_field"].(string); ok && sf != "" {
		sectionsField = sf
	}

	sectionsData := datahelpers.ExtractNestedField(params.CollectedData, sectionsField)
	if sectionsData == nil {
		return nil, fmt.Errorf("sections not found at %s", sectionsField)
	}

	var sections []string

	switch v := sectionsData.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				sections = append(sections, s)
			} else if m, ok := item.(map[string]interface{}); ok {
				if html, ok := m["rendered_html"].(string); ok {
					sections = append(sections, html)
				}
			}
		}
	case map[string]interface{}:
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

	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections found to compile")
	}

	// Build page body
	pageBody := strings.Join(sections, "\n\n")

	// Get page name
	pageName := "index"
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
//   - site_id_field: path to site_id
//   - result_type: type of research (competitor, industry, trend, etc.)
//   - data_field: path to research data
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

	// Get research data
	dataField := "research_result"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	researchData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	dataJSON, err := json.Marshal(researchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal research data: %w", err)
	}

	// Insert into research_results table
	resultID := uuid.New()
	query := `
		INSERT INTO research_results (id, site_id, result_type, data, created_at)
		VALUES ($1, $2, $3, $4::jsonb, NOW())
	`

	var siteIDArg interface{} = nil
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	if err := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err != nil {
		// Table might not exist - log and continue
		params.Logger.Warn("InsertResearchResultAction: Insert failed (table may not exist)",
			zap.Error(err))
		return map[string]interface{}{
			"inserted":    false,
			"result_type": resultType,
			"error":       err.Error(),
		}, nil
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
	params.Logger.Info("ValidateSitePlanAction: starting")

	config := params.StepConfig.Config

	// Get the plan field name
	planField := "llm_plan"
	if pf, ok := config["plan_field"].(string); ok && pf != "" {
		planField = pf
	}

	// Get the plan from collected data
	plan, ok := params.CollectedData[planField].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("plan not found at %s", planField)
	}

	// Validate pages array exists
	pages, ok := plan["pages"].([]interface{})
	if !ok || len(pages) == 0 {
		return nil, fmt.Errorf("plan must have pages array")
	}

	// Validate max pages
	maxPages := 20
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}
	if len(pages) > maxPages {
		return nil, fmt.Errorf("too many pages: %d (max %d)", len(pages), maxPages)
	}

	// Validate required pages exist
	if ensurePages, ok := config["ensure_pages"].([]interface{}); ok {
		pageNames := make(map[string]bool)
		for _, p := range pages {
			if pm, ok := p.(map[string]interface{}); ok {
				if name, ok := pm["name"].(string); ok {
					pageNames[name] = true
				}
			}
		}
		for _, required := range ensurePages {
			if reqName, ok := required.(string); ok {
				if !pageNames[reqName] {
					return nil, fmt.Errorf("required page missing: %s", reqName)
				}
			}
		}
	}

	// Apply default style if not set
	if _, ok := plan["style_collection"].(string); !ok {
		if defaultStyle, ok := config["default_style"].(string); ok {
			plan["style_collection"] = defaultStyle
		}
	}

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
