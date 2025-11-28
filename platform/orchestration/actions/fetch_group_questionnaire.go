// FILE: platform/orchestration/actions/fetch_group_questionnaire.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// FetchGroupQuestionnaireAction retrieves the briefing questionnaire
// from an agent_group_definition.
//
// Config:
//   - group_type: string - explicit group type
//   - group_type_field: string - dot-path to read group type from collected_data
//
// Returns:
//   - questionnaire: the briefing_questionnaire JSON from the group definition
//   - group_type: the resolved group type
//   - group_name: the group's display name
func FetchGroupQuestionnaireAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("FetchGroupQuestionnaireAction starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	config := params.StepConfig.Config

	// Resolve group type
	groupType, err := resolveGroupTypeFromConfig(config, params.CollectedData, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve group_type: %w", err)
	}

	params.Logger.Info("Fetching questionnaire for group",
		zap.String("group_type", groupType),
	)

	// Get database connection
	db := params.DB
	if db == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// Fetch questionnaire
	var groupName string
	var questionnaireJSON []byte

	err = db.QueryRowContext(ctx,
		`SELECT name, COALESCE(briefing_questionnaire, '{}'::jsonb)
		 FROM agent_group_definitions 
		 WHERE group_type = $1 
		 ORDER BY version DESC 
		 LIMIT 1`,
		groupType,
	).Scan(&groupName, &questionnaireJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group type '%s' not found", groupType)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Parse questionnaire
	var questionnaire map[string]interface{}
	if err := json.Unmarshal(questionnaireJSON, &questionnaire); err != nil {
		return nil, fmt.Errorf("failed to parse questionnaire: %w", err)
	}

	params.Logger.Info("Fetched questionnaire",
		zap.String("group_type", groupType),
		zap.String("group_name", groupName),
		zap.Int("questionnaire_sections", countSections(questionnaire)),
	)

	return map[string]interface{}{
		"questionnaire": questionnaire,
		"group_type":    groupType,
		"group_name":    groupName,
	}, nil
}

func resolveGroupTypeFromConfig(config map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) (string, error) {
	// Option 1: Explicit group_type
	if groupType, ok := config["group_type"].(string); ok && groupType != "" {
		return groupType, nil
	}

	// Option 2: From field path
	fieldPath, ok := config["group_type_field"].(string)
	if !ok || fieldPath == "" {
		return "", fmt.Errorf("either 'group_type' or 'group_type_field' must be specified")
	}

	// Navigate to field
	parts := strings.Split(fieldPath, ".")
	current := interface{}(collectedData)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return "", fmt.Errorf("key '%s' not found at position %d", part, i)
			}
			current = val
		default:
			return "", fmt.Errorf("cannot navigate into type %T at '%s'", current, part)
		}
	}

	groupType, ok := current.(string)
	if !ok {
		return "", fmt.Errorf("value at path is not a string: %T", current)
	}

	return groupType, nil
}

func countSections(questionnaire map[string]interface{}) int {
	if sections, ok := questionnaire["sections"].([]interface{}); ok {
		return len(sections)
	}
	return 0
}

// ============================================================================
// REGISTRY UPDATE
// ============================================================================
// Add to GlobalActionRegistry in registry.go:
//
// "fetch_group_questionnaire": FetchGroupQuestionnaireAction,
//
