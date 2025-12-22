// ===========================================================================
// ASSEMBLE FROM LIBRARY - Updated to use component_library.go
// FILE: platform/orchestration/actions/assemble_from_library.go
// ===========================================================================
// Assembles full pages from component lists.
// Now uses shared component_library.go for queries and rendering.
// ===========================================================================

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

// AssembleOutput is the successful return value for the action.
type AssembleOutput struct {
	StitchedHTMLTemplate string                 `json:"stitched_html_template"`
	ContentRequirements  map[string]interface{} `json:"content_requirements"`
	ComponentIDs         []string               `json:"component_ids"`
	StyleCollection      string                 `json:"style_collection"`
}

// BuildPlan defines the structure of the JSON we expect from the strategist.
type BuildPlan struct {
	Sections []json.RawMessage      `json:"sections"`
	Strategy map[string]interface{} `json:"strategy,omitempty"`
}

// AssembleFromLibraryAction uses standard input_fields to find the build plan.
func AssembleFromLibraryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing AssembleFromLibraryAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// 1. Get DB connection
	db := params.DB
	if db == nil {
		params.Logger.Error("Database pool (params.DB) is nil")
		return nil, fmt.Errorf("database connection is not available")
	}

	// 2. Extract domain and site ID
	domain := extractDomainFromParams(params)
	siteID := extractSiteIDFromParams(params)
	params.Logger.Info("Extracted site info",
		zap.String("domain", domain),
		zap.String("site_id", siteID.String()),
	)

	// 3. Get style collection (from site or by domain)
	var styleCollection *StyleCollection
	var err error

	if siteID != uuid.Nil {
		styleCollection, err = GetStyleCollectionForSite(ctx, db, siteID, params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to get style collection for site", zap.Error(err))
		}
	}

	if styleCollection == nil && domain != "" {
		styleCollection, err = SelectStyleCollectionByDomain(ctx, db, domain, params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to select style collection by domain", zap.Error(err))
		}
	}

	// 4. Get theme CSS
	var themeCSS string
	var themeName string

	if styleCollection != nil && styleCollection.CSSThemeID != nil {
		// Get theme by ID from collection
		theme, err := getThemeByID(ctx, db, *styleCollection.CSSThemeID, params.Logger)
		if err == nil {
			themeCSS = theme.CSSContent
			themeName = theme.Name
		}
	}

	if themeCSS == "" {
		// Fallback: get theme by collection name or default
		if styleCollection != nil {
			themeName = styleCollection.Name
		} else {
			themeName = "default"
		}
		theme, err := GetThemeByName(ctx, db, themeName, params.Logger)
		if err == nil {
			themeCSS = theme.CSSContent
		}
	}

	// 5. Extract build plan
	buildPlanJSON, err := extractBuildPlan(params)
	if err != nil {
		return nil, err
	}

	buildPlan, err := parseBuildPlan(buildPlanJSON, params.Logger)
	if err != nil {
		return nil, err
	}

	componentNames, err := extractComponentNames(buildPlan, params.Logger)
	if err != nil {
		return nil, err
	}

	// 6. Build render context
	renderCtx := buildRenderContextFromParams(params, styleCollection)
	renderCtx.ThemeCSS = themeCSS

	// 7. Assemble components
	output, err := assembleComponents(ctx, db, componentNames, renderCtx, themeName, params.Logger)
	if err != nil {
		return nil, err
	}

	if styleCollection != nil {
		output.StyleCollection = styleCollection.Name
	}

	params.Logger.Info("Successfully assembled template",
		zap.Strings("components", output.ComponentIDs),
		zap.String("style_collection", output.StyleCollection),
	)

	return output, nil
}

// extractDomainFromParams gets the domain from params
func extractDomainFromParams(params ActionParams) string {
	// Check input_data wrapper
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			return domain
		}
	}

	// Try direct path
	if domain, ok := datahelpers.GetValueByPath(params.CollectedData, "input_data.domain", params.Logger); ok {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}

	return ""
}

// extractSiteIDFromParams gets the site ID from params
func extractSiteIDFromParams(params ActionParams) uuid.UUID {
	if siteRecord, ok := params.CollectedData["site_record"].(map[string]interface{}); ok {
		if idStr, ok := siteRecord["site_id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

// buildRenderContextFromParams creates a RenderContext from params
func buildRenderContextFromParams(params ActionParams, coll *StyleCollection) *RenderContext {
	ctx := &RenderContext{
		Domain:      extractDomainFromParams(params),
		SiteID:      extractSiteIDFromParams(params),
		LogoText:    extractLogoTextFromParams(params),
		CompanyName: extractCompanyNameFromParams(params),
	}

	// Apply colors from style collection
	if coll != nil && coll.ColorPalette != nil {
		ctx.PrimaryColor = coll.ColorPalette["primary"]
		ctx.AccentColor = coll.ColorPalette["accent"]
		ctx.SecondaryColor = coll.ColorPalette["secondary"]
	}

	// Get navigation if available
	ctx.NavItems = extractNavItemsFromParams(params)

	return ctx
}

// extractLogoTextFromParams gets logo text from params
func extractLogoTextFromParams(params ActionParams) string {
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if brief, ok := inputData["reviewed_brief"].(map[string]interface{}); ok {
			if name, ok := brief["business_name"].(string); ok && name != "" {
				return name
			}
		}
		if name, ok := inputData["business_name"].(string); ok && name != "" {
			return name
		}
		// Extract from domain
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			parts := strings.Split(domain, ".")
			if len(parts) > 0 && len(parts[0]) > 0 {
				name := parts[0]
				return strings.ToUpper(name[:1]) + name[1:]
			}
		}
	}
	return "Company"
}

// extractCompanyNameFromParams gets company name
func extractCompanyNameFromParams(params ActionParams) string {
	return extractLogoTextFromParams(params)
}

// extractNavItemsFromParams gets nav items from db_sync
func extractNavItemsFromParams(params ActionParams) []NavItem {
	var items []NavItem

	if dbSync, ok := params.CollectedData["db_sync"].(map[string]interface{}); ok {
		if nav, ok := dbSync["navigation"].(map[string]interface{}); ok {
			if navItems, ok := nav["items"].([]interface{}); ok {
				for _, item := range navItems {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							items = append(items, NavItem{Label: label, URL: url})
						}
					}
				}
			}
		}
	}

	return items
}

// getThemeByID fetches a theme by UUID
func getThemeByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*Theme, error) {
	query := `SELECT id, name, css_content, color_palette FROM css_themes WHERE id = $1`

	var theme Theme
	var colorPaletteJSON []byte

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, id).Scan(
			&theme.ID, &theme.Name, &theme.CSSContent, &colorPaletteJSON,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported database type")
	}

	if len(colorPaletteJSON) > 0 {
		json.Unmarshal(colorPaletteJSON, &theme.ColorPalette)
	}

	return &theme, nil
}

// assembleComponents queries and stitches components using the shared library
func assembleComponents(ctx context.Context, db interface{}, componentNames []string, renderCtx *RenderContext, themeName string, logger *zap.Logger) (*AssembleOutput, error) {
	var finalHTML strings.Builder
	contentRequirements := make(map[string]interface{})
	var componentIDs []string
	var componentFunctions []string

	for idx, componentName := range componentNames {
		logger.Info("Querying for component",
			zap.Int("index", idx),
			zap.String("function", componentName))

		// Use shared GetComponentWithFallback
		comp, err := GetComponentWithFallback(ctx, db, componentName, logger)
		if err != nil {
			logger.Error("Failed to get component", zap.Error(err))
			continue
		}

		componentFunctions = append(componentFunctions, comp.Function)

		// Generate component ID
		componentID := fmt.Sprintf("component_%s_%d", comp.Function, idx)

		// Render template using shared RenderTemplate
		// For head component, inject theme CSS
		if comp.Function == "head" {
			metadataComment := BuildThemeMetadata(themeName, componentNames, renderCtx.Domain)
			renderCtx.ThemeCSS = metadataComment + renderCtx.ThemeCSS
			renderCtx.Title = renderCtx.Domain
			renderCtx.Description = "Welcome to " + renderCtx.Domain
		}

		renderedHTML := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)

		// Also handle old-style {{.ComponentID}} placeholder
		renderedHTML = strings.ReplaceAll(renderedHTML, "{{.ComponentID}}", componentID)

		finalHTML.WriteString(renderedHTML + "\n")

		logger.Info("Stitched component",
			zap.String("function", comp.Function),
			zap.String("component_id", componentID),
			zap.Int("html_length", len(renderedHTML)))

		// Collect content requirements
		if comp.InputSchema != nil {
			contentRequirements[componentID] = comp.InputSchema
		}

		componentIDs = append(componentIDs, comp.ID)
	}

	if len(componentIDs) == 0 {
		return nil, fmt.Errorf("no components were successfully assembled")
	}

	logger.Info("Successfully assembled all components",
		zap.Int("total_components", len(componentIDs)),
		zap.Int("total_html_length", finalHTML.Len()),
		zap.Strings("component_functions", componentFunctions))

	return &AssembleOutput{
		StitchedHTMLTemplate: finalHTML.String(),
		ContentRequirements:  contentRequirements,
		ComponentIDs:         componentIDs,
	}, nil
}

// ===========================================================================
// EXISTING HELPER FUNCTIONS (unchanged)
// ===========================================================================

// extractBuildPlan - uses standard input_fields mechanism
func extractBuildPlan(params ActionParams) (string, error) {
	var inputFields []string
	if fields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		for _, f := range fields {
			if field, ok := f.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	}

	if len(inputFields) == 0 {
		inputFields = []string{"build_plan", "build_plan_data", "call_strategist"}
	}

	dataToSearch := params.CollectedData
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		dataToSearch = inputData
	}

	for _, field := range inputFields {
		if val, ok := dataToSearch[field]; ok {
			return extractJSONFromValue(val, params.Logger)
		}
	}

	return "", fmt.Errorf("build plan not found in any of: %v", inputFields)
}

// extractJSONFromValue extracts JSON string from various value types
func extractJSONFromValue(val interface{}, logger *zap.Logger) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case map[string]interface{}:
		// Look for common content fields
		for _, key := range []string{"build_plan", "content", "response", "plan"} {
			if content, ok := v[key]; ok {
				return extractJSONFromValue(content, logger)
			}
		}
		// Serialize the map
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("cannot extract JSON from type %T", v)
		}
		return string(bytes), nil
	}
}

// parseBuildPlan parses the JSON string into BuildPlan
func parseBuildPlan(buildPlanJSON string, logger *zap.Logger) (*BuildPlan, error) {
	var buildPlan BuildPlan

	if err := json.Unmarshal([]byte(buildPlanJSON), &buildPlan); err != nil {
		// Try double-unquoting
		var unquoted string
		if json.Unmarshal([]byte(buildPlanJSON), &unquoted) == nil {
			if json.Unmarshal([]byte(unquoted), &buildPlan) == nil {
				return &buildPlan, nil
			}
		}
		return nil, fmt.Errorf("failed to parse build plan: %w", err)
	}

	return &buildPlan, nil
}

// extractComponentNames handles both string and object formats
func extractComponentNames(buildPlan *BuildPlan, logger *zap.Logger) ([]string, error) {
	var componentNames []string

	for idx, rawSection := range buildPlan.Sections {
		var componentName string

		// Try as string
		var strSection string
		if json.Unmarshal(rawSection, &strSection) == nil {
			componentName = strSection
		} else {
			// Try as object
			var objSection struct {
				Component string `json:"component"`
			}
			if json.Unmarshal(rawSection, &objSection) == nil {
				componentName = objSection.Component
			} else {
				logger.Warn("Could not parse section", zap.Int("index", idx))
				continue
			}
		}

		if componentName != "" {
			componentNames = append(componentNames, componentName)
		}
	}

	if len(componentNames) == 0 {
		return nil, fmt.Errorf("no valid component names found")
	}

	return componentNames, nil
}
