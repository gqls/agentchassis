// internal/backend/agent-chassis/platform/orchestration/actions/assemble_from_library.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AssembleOutput is the successful return value for the action.
type AssembleOutput struct {
	StitchedHTMLTemplate string                 `json:"stitched_html_template"`
	ContentRequirements  map[string]interface{} `json:"content_requirements"`
	ComponentIDs         []string               `json:"component_ids"`
}

// BuildPlan defines the structure of the JSON we expect from the strategist.
type BuildPlan struct {
	Sections []json.RawMessage      `json:"sections"`           // Can be strings OR objects
	Strategy map[string]interface{} `json:"strategy,omitempty"` // Optional detailed strategy
}

// ComponentTemplate is a helper struct for our DB query.
type ComponentTemplate struct {
	ID           string          `db:"id"`
	HTMLTemplate string          `db:"html_template"`
	InputSchema  json.RawMessage `db:"input_schema"`
	Function     string          `db:"function"`
}

// AssembleFromLibraryAction uses standard input_fields to find the build plan.
func AssembleFromLibraryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing AssembleFromLibraryAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("DEBUGaa: assemble - params", params),
	)

	// 1. Get DB connection
	db := params.DB
	if db == nil {
		params.Logger.Error("Database pool (params.DB) is nil")
		return nil, fmt.Errorf("database connection is not available")
	}

	// 2. Extract domain from input_data
	domain := extractDomain(params)
	params.Logger.Info("Extracted domain", zap.String("domain", domain))

	// 3. Select theme based on domain
	themeName := selectTheme(domain, params.Logger)

	// 4. Fetch theme CSS from database
	themeCSS, err := fetchThemeCSS(ctx, db, themeName, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to fetch theme CSS, proceeding without theme",
			zap.Error(err))
		themeCSS = "" // Continue without theme
	}

	// 5. Extract build plan using standard input_fields mechanism
	buildPlanJSON, err := extractBuildPlan(params)
	if err != nil {
		return nil, err
	}

	// 6. Parse the build plan
	buildPlan, err := parseBuildPlan(buildPlanJSON, params.Logger)
	if err != nil {
		return nil, err
	}

	// 7. Extract component names (handles both formats)
	componentNames, err := extractComponentNames(buildPlan, params.Logger)
	if err != nil {
		return nil, err
	}

	// 8. Assemble components from library
	output, err := assembleComponentsByName(ctx, db, componentNames, domain, themeName, themeCSS, params.Logger)
	if err != nil {
		return nil, err
	}

	params.Logger.Info("Successfully assembled template",
		zap.Any("components", output.ComponentIDs),
		zap.Any("stitched html template", output.StitchedHTMLTemplate),
		zap.String("theme", themeName),
	)

	return output, nil
}

// extractDomain gets the domain from input_data
func extractDomain(params ActionParams) string {
	// Check input_data wrapper first
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if buildPlanData, ok := inputData["build_plan_data"].(map[string]interface{}); ok {
			if inputDataNested, ok := buildPlanData["input_data"].(map[string]interface{}); ok {
				if inputDataFinal, ok := inputDataNested["input_data"].(map[string]interface{}); ok {
					if domain, ok := inputDataFinal["domain"].(string); ok {
						return domain
					}
				}
			}
		} else if buildPlanData, ok = inputData["build_plan"].(map[string]interface{}); ok {
			if inputDataNested, ok := buildPlanData["input_data"].(map[string]interface{}); ok {
				if inputDataFinal, ok := inputDataNested["input_data"].(map[string]interface{}); ok {
					if domain, ok := inputDataFinal["domain"].(string); ok {
						return domain
					}
				}
			}
		}
	}

	// Fallback: try direct path
	if domain, ok := datahelpers.GetValueByPath(params.CollectedData, "input_data.domain", params.Logger); ok {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}

	params.Logger.Warn("Could not extract domain, using empty string")
	return ""
}

// selectTheme chooses a CSS theme based on domain keywords
func selectTheme(domain string, logger *zap.Logger) string {
	domain = strings.ToLower(domain)

	// Sports & Competition
	if strings.Contains(domain, "box") || strings.Contains(domain, "fight") ||
		strings.Contains(domain, "sport") || strings.Contains(domain, "gym") ||
		strings.Contains(domain, "fitness") || strings.Contains(domain, "martial") {
		logger.Info("Selected boxing theme", zap.String("domain", domain))
		return "boxing"
	}

	// Food & Hospitality
	if strings.Contains(domain, "bak") || strings.Contains(domain, "food") ||
		strings.Contains(domain, "cafe") || strings.Contains(domain, "restaurant") ||
		strings.Contains(domain, "cook") || strings.Contains(domain, "chef") ||
		strings.Contains(domain, "bistro") {
		logger.Info("Selected bakery theme", zap.String("domain", domain))
		return "bakery"
	}

	// Tech & SaaS
	if strings.Contains(domain, "tech") || strings.Contains(domain, "software") ||
		strings.Contains(domain, "app") || strings.Contains(domain, "ai") ||
		strings.Contains(domain, "cloud") || strings.Contains(domain, "dev") ||
		strings.Contains(domain, "code") || strings.Contains(domain, "data") ||
		strings.Contains(domain, "cyber") {
		logger.Info("Selected tech theme", zap.String("domain", domain))
		return "tech"
	}

	// Law & Finance
	if strings.Contains(domain, "law") || strings.Contains(domain, "legal") ||
		strings.Contains(domain, "finance") || strings.Contains(domain, "invest") ||
		strings.Contains(domain, "consult") || strings.Contains(domain, "advisor") ||
		strings.Contains(domain, "capital") {
		logger.Info("Selected professional-dark theme", zap.String("domain", domain))
		return "professional-dark"
	}

	// Default fallback
	logger.Info("Selected default theme", zap.String("domain", domain))
	return "default"
}

// fetchThemeCSS retrieves CSS content from the database
func fetchThemeCSS(ctx context.Context, db *sql.DB, themeName string, logger *zap.Logger) (string, error) {
	query := `
		SELECT css_content 
		FROM css_themes 
		WHERE name = $1 AND is_active = true
		LIMIT 1`

	var cssContent string
	err := db.QueryRowContext(ctx, query, themeName).Scan(&cssContent)

	if err != nil {
		if err == sql.ErrNoRows {
			// Theme not found, try default
			logger.Warn("Theme not found, falling back to default",
				zap.String("requested_theme", themeName))

			err = db.QueryRowContext(ctx, query, "default").Scan(&cssContent)
			if err != nil {
				return "", fmt.Errorf("default theme not found: %w", err)
			}
			return cssContent, nil
		}
		return "", fmt.Errorf("failed to fetch theme CSS: %w", err)
	}

	logger.Info("Successfully fetched theme CSS",
		zap.String("theme", themeName),
		zap.Int("css_length", len(cssContent)))

	return cssContent, nil
}

// extractBuildPlan uses the standard input_fields mechanism to find the build plan JSON.
func extractBuildPlan(params ActionParams) (string, error) {
	// Get input_fields from config (standard approach)
	var inputFields []string
	if fields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		for _, fieldInterface := range fields {
			if field, ok := fieldInterface.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	}

	// Default: look for common field names if not specified
	if len(inputFields) == 0 {
		params.Logger.Warn("No input_fields specified, using defaults",
			zap.Strings("defaults", []string{"build_plan_data", "call_strategist"}),
		)
		inputFields = []string{"build_plan", "build_plan_data", "call_strategist"}
	}

	params.Logger.Info("Searching for build plan",
		zap.Strings("input_fields", inputFields),
		zap.Strings("available_root_keys", datahelpers.GetMapKeys(params.CollectedData)))

	// Check if data is wrapped in "input_data" and unwrap it
	dataToSearch := params.CollectedData
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		params.Logger.Info("Found input_data wrapper, searching within it",
			zap.Strings("wrapped_keys", datahelpers.GetMapKeys(inputData)))
		dataToSearch = inputData
	}

	// Try each input field in order
	for _, fieldName := range inputFields {
		buildPlanJSON, found := findBuildPlanInFieldWithData(fieldName, dataToSearch, params.Logger)
		if found {
			params.Logger.Info("Found build plan",
				zap.String("field", fieldName),
				zap.String("build plan json is:", buildPlanJSON))
			return buildPlanJSON, nil
		}
	}

	// Log what we have for debugging
	params.Logger.Error("Build plan not found in any input_fields",
		zap.Strings("searched_fields", inputFields),
		zap.Strings("available_keys", datahelpers.GetMapKeys(dataToSearch)))

	return "", fmt.Errorf("build plan not found in input_fields: %v", inputFields)
}

// findBuildPlanInFieldWithData searches for build plan JSON in a specific field within provided data.
func findBuildPlanInFieldWithData(fieldName string, data map[string]interface{}, logger *zap.Logger) (string, bool) {
	// Try direct lookup with dot notation support
	if val, ok := datahelpers.GetValueByPath(data, fieldName, logger); ok {
		logger.Info("In findBuildPlanInField got some value",
			zap.String("field name", fieldName),
			zap.Any("val", val))
		if buildPlanJSON := extractJSONFromValue(val, logger); buildPlanJSON != "" {
			return buildPlanJSON, true
		}
	}

	logger.Info("In findBuildPlanInField didnt find build plan in field first attempt",
		zap.String("field name", fieldName),
		zap.Any("data", data),
	)

	// Try looking inside the field for common sub-keys
	commonSubKeys := []string{
		"result",
		"generate_build_plan.result",
		"build_plan_json",
	}

	for _, subKey := range commonSubKeys {
		fullPath := fieldName + "." + subKey
		if val, ok := datahelpers.GetValueByPath(data, fullPath, logger); ok {
			if buildPlanJSON := extractJSONFromValue(val, logger); buildPlanJSON != "" {
				logger.Debug("Found build plan in subkey",
					zap.String("full_path", fullPath))
				return buildPlanJSON, true
			}
		}
	}

	logger.Error("In findBuildPlanInField didnt find build plan in field final attempt",
		zap.String("field name", fieldName),
		zap.Any("data", data),
	)

	return "", false
}

// Keep old function for backward compatibility (can be removed later)
func findBuildPlanInField(fieldName string, params ActionParams) (string, bool) {
	return findBuildPlanInFieldWithData(fieldName, params.CollectedData, params.Logger)
}

// extractJSONFromValue tries to extract a JSON string from various value types.
func extractJSONFromValue(val interface{}, logger *zap.Logger) string {
	switch v := val.(type) {
	case string:
		// Direct string - might be JSON or might be markdown-wrapped
		cleaned := datahelpers.CleanMarkdownJSON(v)
		logger.Debug("Extracted string value",
			zap.Int("original_length", len(v)),
			zap.Int("cleaned_length", len(cleaned)),
		)
		return cleaned

	case map[string]interface{}:
		// Check for nested generate_build_plan.result pattern (common from strategist)
		if genBuildPlan, ok := v["generate_build_plan"].(map[string]interface{}); ok {
			if result, ok := genBuildPlan["result"].(string); ok {
				logger.Debug("Found build plan in generate_build_plan.result")
				return datahelpers.CleanMarkdownJSON(result)
			}
		}

		// Check for nested call_strategist.result pattern
		if callStrategist, ok := v["call_strategist"].(map[string]interface{}); ok {
			if result, ok := callStrategist["result"].(string); ok {
				logger.Debug("Found build plan in call_strategist.result")
				return datahelpers.CleanMarkdownJSON(result)
			}
		}

		// Direct result key
		if result, ok := v["result"].(string); ok {
			logger.Debug("Found build plan in direct result key")
			return datahelpers.CleanMarkdownJSON(result)
		}

		// build_plan_json key
		if buildPlan, ok := v["build_plan_json"].(string); ok {
			logger.Debug("Found build plan in build_plan_json key")
			return datahelpers.CleanMarkdownJSON(buildPlan)
		}

		// Log available keys for debugging before fallback
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		logger.Warn("No known build plan pattern found, marshaling entire map",
			zap.Strings("available_keys", keys))

		// Try marshaling the whole map as JSON
		if jsonBytes, err := json.Marshal(v); err == nil {
			jsonStr := string(jsonBytes)
			logger.Warn("Marshaling entire map as JSON (no known patterns found)",
				zap.Int("json_length", len(jsonStr)),
				zap.Strings("top_level_keys", datahelpers.GetMapKeys(v)),
			)
			return jsonStr
		}

	default:
		logger.Debug("Unexpected value type for build plan",
			zap.String("type", fmt.Sprintf("%T", val)))
	}

	return ""
}

// parseBuildPlan unmarshals the JSON into a BuildPlan struct.
func parseBuildPlan(buildPlanJSON string, logger *zap.Logger) (*BuildPlan, error) {
	logger.Info("Parsing build plan JSON",
		zap.String("json_preview", buildPlanJSON[:min(len(buildPlanJSON), 200)]),
		zap.Int("json_length", len(buildPlanJSON)))

	var buildPlan BuildPlan

	if err := json.Unmarshal([]byte(buildPlanJSON), &buildPlan); err != nil {
		// Try double-unquoting (common with LLM outputs)
		var unquoted string
		if unqErr := json.Unmarshal([]byte(buildPlanJSON), &unquoted); unqErr == nil {
			if err2 := json.Unmarshal([]byte(unquoted), &buildPlan); err2 == nil {
				logger.Info("Successfully parsed double-encoded build plan",
					zap.Any("build plan", buildPlan),
				)
				return &buildPlan, nil
			}
		}

		logger.Error("Failed to parse build plan",
			zap.Error(err),
			zap.String("json_preview", buildPlanJSON[:min(len(buildPlanJSON), 200)]))
		return nil, fmt.Errorf("failed to parse build plan: %w", err)
	}

	logger.Info("In parseBuildPlan didnot parse build plan",
		zap.Any("build plan, look for section_count", buildPlan),
		zap.Any("build plan sections (objects)", buildPlan.Sections))

	return &buildPlan, nil
}

// extractComponentNames handles both string and object formats
func extractComponentNames(buildPlan *BuildPlan, logger *zap.Logger) ([]string, error) {
	logger.Info("in extractComponentNames",
		zap.Any("build plan inward is:", buildPlan))

	var componentNames []string

	for idx, rawSection := range buildPlan.Sections {
		var componentName string

		// Try parsing as string first (simpler format)
		var strSection string
		if err := json.Unmarshal(rawSection, &strSection); err == nil {
			componentName = strSection
			logger.Debug("Parsed section as string",
				zap.Int("index", idx),
				zap.String("component", componentName))
		} else {
			// Try parsing as object (complex format)
			var objSection struct {
				Component string `json:"component"`
			}
			if err := json.Unmarshal(rawSection, &objSection); err == nil {
				componentName = objSection.Component
				logger.Debug("Parsed section as object",
					zap.Int("index", idx),
					zap.String("component", componentName))
			} else {
				logger.Error("Could not parse section as string or object",
					zap.Int("index", idx),
					zap.String("raw", string(rawSection)))
				continue
			}
		}

		if componentName == "" {
			logger.Warn("Empty component name", zap.Int("index", idx))
			continue
		}

		componentNames = append(componentNames, componentName)
	}

	if len(componentNames) == 0 {
		return nil, fmt.Errorf("no valid component names found in build plan")
	}

	logger.Info("Extracted component names",
		zap.Strings("components", componentNames),
		zap.Int("count", len(componentNames)))

	return componentNames, nil
}

// buildThemeMetadataComment creates a CSS comment with theme and component info for traceability
func buildThemeMetadataComment(themeName string, componentFunctions []string, domain string) string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	components := strings.Join(componentFunctions, ", ")

	return fmt.Sprintf(`/*
 * ============================================
 * SITE BUILD METADATA
 * ============================================
 * Theme: %s
 * Domain: %s
 * Components: %s
 * Generated: %s
 * Source: component-library (assemble_from_library)
 * ============================================
 */
`, themeName, domain, components, timestamp)
}

// assembleComponentsByName queries the database and stitches together HTML from component names.
func assembleComponentsByName(ctx context.Context, db *sql.DB, componentNames []string, domain string, themeName string, themeCSS string, logger *zap.Logger) (*AssembleOutput, error) {
	var finalHTML strings.Builder
	contentRequirements := make(map[string]interface{})
	var componentIDs []string
	var componentFunctions []string // Track function names for metadata

	for idx, componentName := range componentNames {
		logger.Info("Querying for component",
			zap.Int("index", idx),
			zap.String("function", componentName))

		// Query for the component (with fallback)
		component, err := queryComponentWithFallback(ctx, db, logger, componentName)
		if err != nil {
			logger.Error("Failed to get component (even with fallback)",
				zap.String("function", componentName),
				zap.Error(err))
			continue
		}

		// Track component function for metadata
		componentFunctions = append(componentFunctions, component.Function)

		// Stitch the HTML with a unique component ID
		componentID := fmt.Sprintf("component_%s_%d", component.Function, idx)
		templatedHTML := strings.Replace(component.HTMLTemplate, "{{.ComponentID}}", componentID, -1)

		// Special handling for HEAD component - inject theme CSS with metadata and other values
		if component.Function == "head" {
			// Build metadata comment and prepend to theme CSS
			metadataComment := buildThemeMetadataComment(themeName, componentNames, domain)
			labeledThemeCSS := metadataComment + themeCSS

			templatedHTML = strings.Replace(templatedHTML, "{{.theme_css}}", labeledThemeCSS, -1)
			// Also inject title and description if available
			templatedHTML = strings.Replace(templatedHTML, "{{.title}}", domain, -1)
			templatedHTML = strings.Replace(templatedHTML, "{{.description}}", "Welcome to "+domain, -1)
		}

		finalHTML.WriteString(templatedHTML + "\n")

		logger.Info("Stitched HTML for component",
			zap.String("component_function", component.Function),
			zap.String("component_id", componentID),
			zap.Int("html_length", len(templatedHTML)))

		// Collect content requirements for this component
		var schema interface{}
		if err := json.Unmarshal(component.InputSchema, &schema); err == nil {
			contentRequirements[componentID] = schema
		} else {
			logger.Warn("Failed to unmarshal input schema for component",
				zap.String("component_id", componentID),
				zap.Error(err))
		}

		componentIDs = append(componentIDs, component.ID)
	}

	if len(componentIDs) == 0 {
		return nil, fmt.Errorf("no components were successfully assembled")
	}

	logger.Info("Successfully assembled all components",
		zap.Int("total_components", len(componentIDs)),
		zap.Int("total_html_length", finalHTML.Len()),
		zap.String("theme_used", themeName),
		zap.Strings("component_functions", componentFunctions))

	return &AssembleOutput{
		StitchedHTMLTemplate: finalHTML.String(),
		ContentRequirements:  contentRequirements,
		ComponentIDs:         componentIDs,
	}, nil
}

// queryComponentWithFallback tries to get a component, falling back to generic if needed.
func queryComponentWithFallback(ctx context.Context, db *sql.DB, logger *zap.Logger, function string) (*ComponentTemplate, error) {
	// Try primary query
	component, err := queryComponent(ctx, db, logger, function)
	if err == nil {
		return component, nil
	}

	logger.Warn("Component not found, using fallback",
		zap.String("requested_function", function),
		zap.String("fallback", "generic-text-block"))

	// Fallback to generic component
	component, err = queryComponent(ctx, db, logger, "generic-text-block")
	if err != nil {
		return nil, fmt.Errorf("fallback component 'generic-text-block' not found: %w", err)
	}

	return component, nil
}

// queryComponent fetches a single component from the database.
func queryComponent(ctx context.Context, db *sql.DB, logger *zap.Logger, function string) (*ComponentTemplate, error) {
	query := `
		SELECT id, html_template, input_schema, "function"
		FROM content_components
		WHERE "function" = $1
		LIMIT 1`

	var component ComponentTemplate
	err := db.QueryRowContext(ctx, query, function).Scan(
		&component.ID,
		&component.HTMLTemplate,
		&component.InputSchema,
		&component.Function,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("component not found: %s", function)
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	return &component, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
