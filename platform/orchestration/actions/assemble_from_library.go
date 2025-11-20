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

// BuildPlan defines the structure of the JSON we expect from the strategist.
type BuildPlan struct {
	Sections []string `json:"sections"`
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

	if len(buildPlan.Sections) == 0 {
		return nil, fmt.Errorf("build plan has no sections")
	}

	// 4. Assemble components from library
	output, err := assembleComponents(ctx, db, buildPlan, params.Logger)
	if err != nil {
		return nil, err
	}

	params.Logger.Info("Successfully assembled template",
		zap.Int("component_count", len(output.ComponentIDs)),
		zap.Int("html_length", len(output.StitchedHTMLTemplate)),
	)

	return output, nil
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
				zap.Int("json_length", len(buildPlanJSON)),
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
		if buildPlanJSON := extractJSONFromValue(val, logger); buildPlanJSON != "" {
			return buildPlanJSON, true
		}
	}

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
					zap.String("full_path", fullPath),
				)
				return buildPlanJSON, true
			}
		}
	}

	return "", false
}

// extractJSONFromValue tries to extract a JSON string from various value types.
func extractJSONFromValue(val interface{}, logger *zap.Logger) string {
	switch v := val.(type) {
	case string:
		// Direct string - might be JSON or might be markdown-wrapped
		return cleanMarkdownJSON(v)

	case map[string]interface{}:
		// If it's already a map, check for common keys
		if result, ok := v["result"].(string); ok {
			return cleanMarkdownJSON(result)
		}
		if buildPlan, ok := v["build_plan_json"].(string); ok {
			return cleanMarkdownJSON(buildPlan)
		}
		// Try marshaling the whole map as JSON
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}

	default:
		logger.Debug("Unexpected value type for build plan",
			zap.String("type", fmt.Sprintf("%T", val)),
		)
	}

	return ""
}

// cleanMarkdownJSON removes markdown code fences if present.
func cleanMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove ```json ... ``` wrappers
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")

	return strings.TrimSpace(s)
}

// parseBuildPlan unmarshals the JSON into a BuildPlan struct.
func parseBuildPlan(buildPlanJSON string, logger *zap.Logger) (*BuildPlan, error) {
	var buildPlan BuildPlan

	if err := json.Unmarshal([]byte(buildPlanJSON), &buildPlan); err != nil {
		// Try double-unquoting (common with LLM outputs)
		var unquoted string
		if unqErr := json.Unmarshal([]byte(buildPlanJSON), &unquoted); unqErr == nil {
			if err2 := json.Unmarshal([]byte(unquoted), &buildPlan); err2 == nil {
				logger.Info("Successfully parsed double-encoded build plan")
				return &buildPlan, nil
			}
		}

		logger.Error("Failed to parse build plan",
			zap.Error(err),
			zap.String("json_preview", buildPlanJSON[:min(len(buildPlanJSON), 200)]),
		)
		return nil, fmt.Errorf("failed to parse build plan: %w", err)
	}

	logger.Info("Successfully parsed build plan",
		zap.Int("section_count", len(buildPlan.Sections)),
		zap.Strings("sections", buildPlan.Sections),
	)

	return &buildPlan, nil
}

// assembleComponents queries the database and stitches together the HTML.
func assembleComponents(ctx context.Context, db *sql.DB, buildPlan *BuildPlan, logger *zap.Logger) (*AssembleOutput, error) {
	var finalHTML strings.Builder
	contentRequirements := make(map[string]interface{})
	var componentIDs []string

	for idx, function := range buildPlan.Sections {
		logger.Info("Querying for component",
			zap.Int("index", idx),
			zap.String("function", function),
		)

		// Query for the component (with fallback)
		component, err := queryComponentWithFallback(ctx, db, logger, function)
		if err != nil {
			logger.Error("Failed to get component (even with fallback)",
				zap.String("function", function),
				zap.Error(err),
			)
			continue
		}

		// Stitch the HTML with a unique component ID
		componentID := fmt.Sprintf("component_%s_%d", component.Function, idx)
		templatedHTML := strings.Replace(component.HTMLTemplate, "{{.ComponentID}}", componentID, -1)
		finalHTML.WriteString(templatedHTML + "\n")

		// Collect content requirements for this component
		var schema interface{}
		if err := json.Unmarshal(component.InputSchema, &schema); err == nil {
			contentRequirements[componentID] = schema
		}

		componentIDs = append(componentIDs, component.ID)
	}

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
