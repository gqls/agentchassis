// internal/backend/agent-chassis/platform/orchestration/actions/site_architect_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
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

	// 2. Get the Build Plan JSON string from CollectedData
	// The 'call_strategist' step stored its output in 'build_plan_data'.
	var buildPlanData map[string]interface{}
	if bpData, ok := params.CollectedData["build_plan_data"].(map[string]interface{}); ok {
		buildPlanData = bpData
	} else {
		return nil, fmt.Errorf("'build_plan_data' not found or invalid in CollectedData")
	}

	var buildPlanJSON string
	if bpJSON, ok := buildPlanData["build_plan_json"].(string); ok {
		buildPlanJSON = bpJSON
	} else {
		return nil, fmt.Errorf("'build_plan_json' string not found in build_plan_data")
	}

	// 3. Unmarshal the Build Plan
	var buildPlan BuildPlan
	if err := json.Unmarshal([]byte(buildPlanJSON), &buildPlan); err != nil {
		params.Logger.Error("Failed to unmarshal Build Plan JSON", zap.Error(err), zap.String("json", buildPlanJSON))
		return nil, fmt.Errorf("failed to parse build plan: %w", err)
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
func queryComponent(ctx context.Context, db *pgxpool.Pool, log *zap.Logger, function string) (*ComponentTemplate, error) {
	query := `
		SELECT id, html_template, input_schema, "function"
		FROM content_components
		WHERE "function" = $1
		LIMIT 1`

	var component ComponentTemplate
	err := db.QueryRow(ctx, query, function).Scan(
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
