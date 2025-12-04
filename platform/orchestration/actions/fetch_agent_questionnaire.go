// FILE: platform/orchestration/actions/fetch_agent_questionnaire.go
// New action to replace fetch_group_questionnaire when we eliminate agent_group_definitions

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// FetchAgentQuestionnaireAction retrieves the briefing_questionnaire from an agent definition
// This replaces fetch_group_questionnaire now that groups are eliminated
//
// Config (accepts both old and new naming):
//
//	agent_type OR group_type: static agent type to fetch questionnaire for
//	agent_type_field OR group_type_field: field path to get agent type dynamically
//
// Returns:
//
//	questionnaire: the briefing_questionnaire JSON from agent_definitions
//	agent_type: the resolved agent type
func FetchAgentQuestionnaireAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	logger.Info("FetchAgentQuestionnaireAction starting",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.Any("config", config),
	)

	// Debug: log what we're working with
	logger.Info("FetchAgentQuestionnaireAction debug",
		zap.Any("step_config", params.StepConfig),
		zap.Any("collected_data_keys", getMapKeys(params.CollectedData)),
		zap.Any("DEBUGaa: collected data", params.CollectedData),
	)

	// Resolve agent type - check both new and old naming for backward compat
	var agentType string

	// Try static config first (new naming, then old)
	if at, ok := config["agent_type"].(string); ok && at != "" {
		agentType = at
		logger.Info("in fetchagentquestionnaire action Found static agent_type", zap.String("agent_type", at))
	} else if at, ok := config["group_type"].(string); ok && at != "" {
		agentType = at
		logger.Info("Found static group_type", zap.String("group_type", at))
	}

	// Try dynamic field (new naming, then old)
	if agentType == "" {
		var fieldPath string
		if fp, ok := config["agent_type_field"].(string); ok && fp != "" {
			fieldPath = fp
			logger.Info("Found agent_type_field", zap.String("field_path", fp))
		} else if fp, ok := config["group_type_field"].(string); ok && fp != "" {
			fieldPath = fp
			logger.Info("Found group_type_field", zap.String("field_path", fp))
		}

		if fieldPath != "" {
			value := resolveFieldPathQuestionnaire(fieldPath, params.CollectedData)
			logger.Info("in fetchagentquestionnaire action Resolved field path",
				zap.String("field_path", fieldPath),
				zap.Any("resolved_value", value),
			)
			if at, ok := value.(string); ok && at != "" {
				agentType = at
			}
		} else {
			logger.Warn("fetchagentquestionnaire No field path found in config",
				zap.Any("config_keys", getMapKeys(config)),
			)
		}
	}

	if agentType == "" {
		return nil, fmt.Errorf("could not resolve agent_type - provide 'agent_type', 'agent_type_field', 'group_type', or 'group_type_field' in config")
	}

	logger.Info("Resolved agent type", zap.String("agent_type", agentType))

	// Query agent_definitions for the briefing_questionnaire
	questionnaireJSON, err := queryBriefingQuestionnaire(ctx, params.DB, agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch questionnaire for agent type %s: %w", agentType, err)
	}

	// Parse the questionnaire
	var questionnaire interface{}
	if err := json.Unmarshal(questionnaireJSON, &questionnaire); err != nil {
		return nil, fmt.Errorf("failed to parse questionnaire JSON: %w", err)
	}

	logger.Info("Fetched agent questionnaire",
		zap.String("agent_type", agentType),
		zap.Int("json_length", len(questionnaireJSON)),
	)

	return map[string]interface{}{
		"questionnaire": questionnaire,
		"agent_type":    agentType,
	}, nil
}

// resolveFieldPathQuestionnaire extracts a nested value from a map using dot notation
// Named to avoid conflict with similar helpers in other action files
// e.g., "confirmed_type.recommended_group" from {"confirmed_type": {"recommended_group": "landing-page-builder"}}
func resolveFieldPathQuestionnaire(path string, data map[string]interface{}) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// queryBriefingQuestionnaire handles both sql.DB and pgxpool.Pool
func queryBriefingQuestionnaire(ctx context.Context, db interface{}, agentType string) ([]byte, error) {
	query := `
		SELECT COALESCE(briefing_questionnaire, '{}'::jsonb)
		FROM agent_definitions 
		WHERE type = $1 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	var questionnaireJSON []byte

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, agentType).Scan(&questionnaireJSON)
		return questionnaireJSON, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(&questionnaireJSON)
		return questionnaireJSON, err
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}
}
