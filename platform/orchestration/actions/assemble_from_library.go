// internal/backend/agent-chassis/platform/orchestration/actions/site_architect_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// AssembleOutput is the successful return value for the action.
type AssembleOutput struct {
	StitchedHTMLTemplate string                 `json:"stitched_html_template"`
	ContentRequirements  map[string]interface{} `json:"content_requirements"` // This tells the Content Creator what to do
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

// AssembleFromLibraryAction is the core "Intelligent Fallback" action.
func AssembleFromLibraryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing AssembleFromLibraryAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// 1. Get DB connection from ActionParams
	db := params.DB
	if db == nil {
		params.Logger.Error("Database pool (params.DB) is nil")
		return nil, fmt.Errorf("database connection is not available")
	}

	var buildPlanJSON string

	// 1. Check if a specific path is defined in the Step Configuration
	// Example config in YAML:
	// config:
	//   build_plan_path: "planner_step.result.build_plan_json"
	if pathRaw, ok := params.CollectedData["build_plan_path"]; ok {
		if pathStr, ok := pathRaw.(string); ok && pathStr != "" {
			params.Logger.Debug("Using configured path for build plan", zap.String("path", pathStr))
			if val, found := getValueByPath(params.CollectedData, pathStr); found {
				if strVal, ok := val.(string); ok {
					buildPlanJSON = strVal
				}
			}
		}
	}

	// 2. Fallback: If not found via config, use the Heuristic Search (Auto-discovery)
	if buildPlanJSON == "" {
		params.Logger.Debug("build_plan_path not configured or not found, attempting heuristic search")
		buildPlanJSON = findBuildPlanHeuristically(params.CollectedData)
	}

	// --- LOGIC END: Generic Key Extraction ---

	if buildPlanJSON == "" {
		params.Logger.Error("Build Plan JSON not found. Checked specific config path and standard locations.")
		return nil, fmt.Errorf("build_plan_json not found in collected data")
	}

	// 3. Unmarshal the Build Plan
	var buildPlan BuildPlan
	if err := json.Unmarshal([]byte(buildPlanJSON), &buildPlan); err != nil {
		// Attempt to clean string if it was double-encoded (common issue with LLM outputs)
		var unquoted string
		if unqErr := json.Unmarshal([]byte(buildPlanJSON), &unquoted); unqErr == nil {
			// If we successfully unquoted it, try unmarshalling into struct again
			if err2 := json.Unmarshal([]byte(unquoted), &buildPlan); err2 != nil {
				return nil, fmt.Errorf("failed to parse build plan (nested): %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse build plan: %w", err)
		}
	}

	if len(buildPlan.Sections) == 0 {
		return nil, fmt.Errorf("build plan has no sections")
	}

	// 4. Loop through sections and query DB
	var finalHTML strings.Builder
	contentRequirements := make(map[string]interface{})
	var componentIDs []string

	for _, function := range buildPlan.Sections {
		params.Logger.Info("Querying for component", zap.String("function", function))

		// 5. Query for the component (P1: Perfect Match)
		component, err := queryComponent(ctx, db, params.Logger, function)
		if err != nil {
			params.Logger.Warn("P1 query failed, trying fallback", zap.String("function", function), zap.Error(err))

			// 6. Fallback Logic (P3: Base Fallback)
			component, err = queryComponent(ctx, db, params.Logger, "generic-text-block")
			if err != nil {
				params.Logger.Error("Fallback query for 'generic-text-block' failed", zap.String("function", function), zap.Error(err))
				continue // Skip this component
			}
		}

		// 7. Stitch the HTML
		// We templatize the HTML to add placeholders for the Content Creator
		componentID := fmt.Sprintf("component_%s", component.ID[:8])
		templatedHTML := strings.Replace(component.HTMLTemplate, "{{.ComponentID}}", componentID, -1)
		finalHTML.WriteString(templatedHTML + "\n")

		// 8. Collect the content requirements
		var schema interface{}
		if err := json.Unmarshal(component.InputSchema, &schema); err == nil {
			contentRequirements[componentID] = schema
		}
		componentIDs = append(componentIDs, component.ID)
	}

	// 9. Create and return the output struct
	output := AssembleOutput{
		StitchedHTMLTemplate: finalHTML.String(),
		ContentRequirements:  contentRequirements,
		ComponentIDs:         componentIDs,
	}

	return output, nil
}

// queryComponent is a local helper function for our P1/P3 logic.
// It's not an "action" itself.
func queryComponent(ctx context.Context, db *sql.DB, log *zap.Logger, function string) (*ComponentTemplate, error) {
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
		log.Warn("queryComponent failed", zap.String("function", function), zap.Error(err))
		return nil, err
	}
	return &component, nil
}

// getValueByPath traverses a map using dot notation (e.g. "step1.result.data")
func getValueByPath(data map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = data

	for _, key := range keys {
		// Ensure current is a map
		currMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		// Check if key exists
		val, exists := currMap[key]
		if !exists {
			return nil, false
		}
		current = val
	}
	return current, true
}

// findBuildPlanHeuristically contains the original "hardcoded" logic to ensure
// backward compatibility with existing workflows that don't use the config.
func findBuildPlanHeuristically(data map[string]interface{}) string {
	// Helper to check a specific map layer
	checkMap := func(m map[string]interface{}) string {
		// Check direct key
		if val, ok := m["build_plan_json"].(string); ok {
			return val
		}

		// Check inside "result" (common wrapper)
		if res, ok := m["result"].(map[string]interface{}); ok {
			if val, ok := res["build_plan_json"].(string); ok {
				return val
			}
		}
		return ""
	}

	// A. Check Top Level
	if val := checkMap(data); val != "" {
		return val
	}

	// B. Check "generate_build_plan" (Common previous step name)
	if sub, ok := data["generate_build_plan"].(map[string]interface{}); ok {
		if val := checkMap(sub); val != "" {
			return val
		}
	}

	// C. Check "input_data" (Passed from parent)
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		if val := checkMap(inputData); val != "" {
			return val
		}
	}

	// D. Legacy fallback
	if bpData, ok := data["build_plan_data"].(map[string]interface{}); ok {
		if val, ok := bpData["build_plan_json"].(string); ok {
			return val
		}
	}

	return ""
}
