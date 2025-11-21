// internal/backend/agent-chassis/platform/orchestration/actions/site_architect_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AssembleOutput is the successful return value for the action.
type AssembleOutput struct {
	StitchedHTMLTemplate string                 `json:"stitched_html_template"`
	ContentRequirements  map[string]interface{} `json:"content_requirements"`
	ComponentIDs         []string               `json:"component_ids"`
}

// BuildPlan can handle multiple formats from the strategist
type BuildPlan struct {
	Sections []json.RawMessage      `json:"sections"`           // Can be strings OR objects
	Strategy map[string]interface{} `json:"strategy,omitempty"` // Optional detailed strategy
}

// Section represents a single component in the build plan
type Section struct {
	Component            string      `json:"component"`
	MessageStrategyStage string      `json:"message_strategy_stage"`
	CopyStructure        string      `json:"copy_structure"`
	SuggestedCopy        interface{} `json:"suggested_copy"` // Can be string or []string
	GraphicsStyle        string      `json:"graphics_style"`
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

	// 2. Extract build plan using standard input_fields mechanism
	buildPlanJSON, err := extractBuildPlan(params)
	if err != nil {
		return nil, err
	}

	// 3. Parse the build plan
	buildPlan, err := parseBuildPlan(buildPlanJSON, params.Logger)
	if err != nil {
		return nil, err
	}

	// 4. Extract component names (handles both formats)
	componentNames, err := extractComponentNames(buildPlan, params.Logger)
	if err != nil {
		return nil, err
	}

	/*	if len(buildPlan.Sections) == 0 {
		return nil, fmt.Errorf("build plan has no sections")
	}*/

	// 5. Assemble components from library
	output, err := assembleComponentsByName(ctx, params.DB, componentNames, params.Logger)
	if err != nil {
		return nil, err
	}

	params.Logger.Info("In AssembleFromLibraryAction",
		zap.String("build plan (json)", buildPlanJSON),
		zap.Any("parsed build plan", buildPlan),
		zap.Any("assembledComponents, output", output),
		zap.Any("DEBUGaa: assemble - params", params),
	)

	params.Logger.Info("Successfully assembled template",
		zap.Int("component_count", len(output.ComponentIDs)),
		zap.Int("html_length", len(output.StitchedHTMLTemplate)),
	)

	return output, nil
}

// extractComponentNames handles both string and object formats
func extractComponentNames(buildPlan *BuildPlan, logger *zap.Logger) ([]string, error) {
	var componentNames []string

	logger.Info("in extractComponentNames",
		zap.Any("build plan inward is:", buildPlan),
	)

	for idx, rawSection := range buildPlan.Sections {
		var componentName string

		// Try parsing as string first (simpler format)
		var strSection string
		if err := json.Unmarshal(rawSection, &strSection); err == nil {
			componentName = strSection
			logger.Info("Parsed section as string",
				zap.Int("index", idx),
				zap.String("component", componentName))
		} else {
			// Try parsing as object (complex format)
			var objSection struct {
				Component string `json:"component"`
			}
			if err := json.Unmarshal(rawSection, &objSection); err == nil {
				componentName = objSection.Component
				logger.Info("Parsed section as object",
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
		inputFields = []string{"build_plan_data", "call_strategist"}
	}

	params.Logger.Info("Searching for build plan",
		zap.Strings("input_fields", inputFields),
		zap.Strings("available_root_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Check if data is wrapped in "input_data" and unwrap it
	dataToSearch := params.CollectedData
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		params.Logger.Info("Found input_data wrapper, searching within it",
			zap.Strings("wrapped_keys", datahelpers.GetMapKeys(inputData)),
		)
		dataToSearch = inputData
	}

	// Try each input field in order
	for _, fieldName := range inputFields {
		buildPlanJSON, found := findBuildPlanInField(fieldName, dataToSearch, params.Logger)
		if found {
			params.Logger.Info("Found build plan",
				zap.String("field", fieldName),
				zap.Any("build plan json is", buildPlanJSON),
			)
			return buildPlanJSON, nil
		}
	}

	// Log what we have for debugging
	params.Logger.Error("Build plan not found in any input_fields",
		zap.Strings("searched_fields", inputFields),
		zap.Strings("available_keys", datahelpers.GetMapKeys(dataToSearch)),
	)

	return "", fmt.Errorf("build plan not found in input_fields: %v", inputFields)
}

// findBuildPlanInFieldWithData searches for build plan JSON in a specific field within provided data.
func findBuildPlanInField(fieldName string, data map[string]interface{}, logger *zap.Logger) (string, bool) {
	// Try direct lookup with dot notation support
	if val, ok := datahelpers.GetValueByPath(data, fieldName, logger); ok {
		logger.Info("In findBuildPlanInField got some value",
			zap.String("field name", fieldName),
			zap.Any("val", val),
		)
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
				logger.Info("Found build plan in subkey",
					zap.String("full_path", fullPath),
				)
				return buildPlanJSON, true
			}
		}
	}

	logger.Info("In findBuildPlanInField didnt find build plan in field fionalattempt",
		zap.String("field name", fieldName),
		zap.Any("data", data),
	)

	return "", false
}

// extractJSONFromValue tries to extract a JSON string from various value types.
// extractJSONFromValue tries to extract a JSON string from various value types.
// It handles common LLM response patterns without hardcoding specific field names.
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
		// Common LLM response field names to try (in priority order)
		commonResultFields := []string{
			"result",          // Most common
			"output",          // Alternative
			"response",        // Alternative
			"data",            // Alternative
			"content",         // Alternative
			"build_plan_json", // Legacy/specific format
		}

		// Try direct access to common result fields
		for _, fieldName := range commonResultFields {
			if result, ok := v[fieldName].(string); ok {
				cleaned := datahelpers.CleanMarkdownJSON(result)
				logger.Debug("Found result in common field",
					zap.String("field_name", fieldName),
					zap.Int("cleaned_length", len(cleaned)),
				)
				return cleaned
			}
		}

		// Recursively search nested maps for result fields
		// This handles patterns like: {any_key: {result: "..."}}
		jsonStr := datahelpers.SearchNestedForJSON(v, commonResultFields, logger, 0)
		if jsonStr != "" {
			return jsonStr
		}

		// Check if the map itself looks like a valid BuildPlan
		// (has "sections" array at top level)
		if _, ok := v["sections"]; ok {
			if jsonBytes, err := json.Marshal(v); err == nil {
				jsonStr := string(jsonBytes)
				logger.Debug("Map has 'sections' field, using as-is",
					zap.Int("json_length", len(jsonStr)),
				)
				return jsonStr
			}
		}

		// Last resort: marshal the whole map (but log warning)
		if jsonBytes, err := json.Marshal(v); err == nil {
			jsonStr := string(jsonBytes)
			logger.Warn("Marshaling entire map as JSON (no known patterns found)",
				zap.Int("json_length", len(jsonStr)),
				zap.Strings("top_level_keys", getMapKeys(v)),
			)
			return jsonStr
		}

	default:
		logger.Warn("Unexpected value type for build plan",
			zap.String("type", fmt.Sprintf("%T", val)),
		)
	}

	return ""
}

// parseBuildPlan unmarshals the JSON into a BuildPlan struct.
func parseBuildPlan(buildPlanJSON string, logger *zap.Logger) (*BuildPlan, error) {
	var buildPlan BuildPlan

	logger.Info("Parsing build plan JSON",
		zap.String("json_preview", buildPlanJSON[:min(len(buildPlanJSON), 500)]),
		zap.Int("json_length", len(buildPlanJSON)),
	)

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

		logger.Error("In parseBuildPlan Failed to parse build plan",
			zap.Error(err),
			zap.String("json_preview", buildPlanJSON[:min(len(buildPlanJSON), 200)]),
		)
		return nil, fmt.Errorf("failed to parse build plan: %w", err)
	}

	logger.Info("In parseBuildPlan didnot parse build plan",
		zap.Any("build plan, look for section_count", buildPlan),
		zap.Any("build plan sections (objects)", buildPlan.Sections),
	)

	return &buildPlan, nil
}

// assembleComponentsByName queries the database and stitches together HTML from component names.
func assembleComponentsByName(ctx context.Context, db *sql.DB, componentNames []string, logger *zap.Logger) (*AssembleOutput, error) {
	var finalHTML strings.Builder
	contentRequirements := make(map[string]interface{})
	var componentIDs []string

	for idx, componentName := range componentNames {
		logger.Info("Querying for component",
			zap.Int("index", idx),
			zap.String("function", componentName),
		)

		// Query for the component (with fallback)
		component, err := queryComponentWithFallback(ctx, db, logger, componentName)
		if err != nil {
			logger.Error("Failed to get component (even with fallback)",
				zap.String("function", componentName),
				zap.Error(err),
			)
			continue
		}

		// Stitch the HTML with a unique component ID
		componentID := fmt.Sprintf("component_%s_%d", component.Function, idx)
		templatedHTML := strings.Replace(component.HTMLTemplate, "{{.ComponentID}}", componentID, -1)
		finalHTML.WriteString(templatedHTML + "\n")

		logger.Info("Stitched HTML for component",
			zap.String("component_function", component.Function),
			zap.String("component_id", componentID),
			zap.Int("html_length", len(templatedHTML)),
		)

		// Collect content requirements for this component
		var schema interface{}
		if err := json.Unmarshal(component.InputSchema, &schema); err == nil {
			contentRequirements[componentID] = schema
		} else {
			logger.Warn("Failed to unmarshal input schema for component",
				zap.String("component_id", componentID),
				zap.Error(err),
			)
		}

		componentIDs = append(componentIDs, component.ID)
	}

	if len(componentIDs) == 0 {
		return nil, fmt.Errorf("no components were successfully assembled")
	}

	logger.Info("Successfully assembled all components",
		zap.Int("total_components", len(componentIDs)),
		zap.Int("total_html_length", finalHTML.Len()),
	)

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
		zap.String("fallback", "generic-text-block"),
	)

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
