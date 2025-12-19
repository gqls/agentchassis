// FILE: platform/orchestration/actions/fetch_agent_questionnaire.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// resolveFieldPathQuestionnaire extracts value from nested map using dot notation
func resolveFieldPathQuestionnaire(path string, data map[string]interface{}, logger *zap.Logger) interface{} {
	logger.Info("Resolving field path using datahelpers.FindByPath",
		zap.String("path", path),
	)

	// Use the battle-tested FindByPath which handles:
	// - Path traversal (e.g., "call_classifier.classify_site.recommended_builder")
	// - Unwrapping {result: "..."} patterns
	// - Stripping markdown fences (```json)
	// - Parsing JSON strings
	// - Recursive unwrapping
	result := datahelpers.FindByPath(data, path, logger)

	if result == nil {
		logger.Warn("FindByPath returned nil",
			zap.String("path", path),
			zap.Strings("available_top_level_keys", datahelpers.GetMapKeys(data)),
		)
	} else {
		logger.Info("FindByPath resolved successfully",
			zap.String("path", path),
			zap.String("result_type", fmt.Sprintf("%T", result)),
		)
	}

	return result
}

// Strategy 2: Try ExtractStepData first (handles step result wrappers)
func resolveFieldPathWithStepData(path string, data map[string]interface{}, logger *zap.Logger) interface{} {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}

	// First part might be a step name
	firstPart := parts[0]

	// Try to extract step data
	if stepVal, ok := data[firstPart]; ok {
		extracted := datahelpers.ExtractStepData(stepVal)
		if extracted != nil {
			// Now navigate the rest of the path through the extracted data
			if len(parts) == 1 {
				return extracted
			}

			// Continue with remaining parts
			remaining := strings.Join(parts[1:], ".")
			if extractedMap, ok := extracted.(map[string]interface{}); ok {
				return resolveFieldPathQuestionnaire(remaining, extractedMap, logger)
			}
		}
	}

	return nil
}

// Strategy 3: Check if entire path is in a step result wrapper
func tryExtractFromStepResult(path string, data map[string]interface{}, logger *zap.Logger) interface{} {
	parts := strings.Split(path, ".")

	for i := len(parts); i > 0; i-- {
		// Try treating first i parts as step name
		stepName := strings.Join(parts[:i], ".")
		if stepVal, ok := data[stepName]; ok {
			extracted := datahelpers.ExtractStepData(stepVal)
			if extracted != nil {
				// If we used all parts, return the extracted value
				if i == len(parts) {
					return extracted
				}
				// Otherwise continue with remaining path
				remaining := strings.Join(parts[i:], ".")
				if extractedMap, ok := extracted.(map[string]interface{}); ok {
					return resolveFieldPathQuestionnaire(remaining, extractedMap, logger)
				}
			}
		}
	}

	return nil
}

func cleanMarkdownJSONQuestionnaire(s string) string {
	s = strings.TrimSpace(s)

	hasJSONFence := strings.HasPrefix(s, "```json")
	hasPlainFence := strings.HasPrefix(s, "```")

	if hasJSONFence {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSpace(s)
	} else if hasPlainFence {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
	}

	if hasJSONFence || hasPlainFence {
		if idx := strings.Index(s, "```"); idx >= 0 {
			s = s[:idx]
			s = strings.TrimSpace(s)
		}
	} else if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}

	return s
}

func tryParseJSONStringQuestionnaire(s string) (map[string]interface{}, bool) {
	cleaned := cleanMarkdownJSONQuestionnaire(s)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err == nil {
		return result, true
	}

	return nil, false
}

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

func FetchAgentQuestionnaireAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	logger.Info("FetchAgentQuestionnaireAction starting",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.Any("config", config),
	)

	// Log collected data structure for debugging
	logger.Info("Collected data structure",
		zap.Strings("top_level_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Resolve agent type - check both new and old naming
	var agentType string

	// Try static config first
	if at, ok := config["agent_type"].(string); ok && at != "" {
		agentType = at
		logger.Info("Using static agent_type", zap.String("agent_type", at))
	} else if at, ok := config["group_type"].(string); ok && at != "" {
		agentType = at
		logger.Info("Using static group_type (old naming)", zap.String("agent_type", at))
	}

	// Try dynamic field if not found
	if agentType == "" {
		var fieldPath string
		if fp, ok := config["agent_type_field"].(string); ok && fp != "" {
			fieldPath = fp
			logger.Info("Using dynamic agent_type_field", zap.String("field_path", fp))
		} else if fp, ok := config["group_type_field"].(string); ok && fp != "" {
			fieldPath = fp
			logger.Info("Using dynamic group_type_field (old naming)", zap.String("field_path", fp))
		}

		if fieldPath != "" {
			// Use FindByPath - handles all unwrapping and JSON parsing
			value := resolveFieldPathQuestionnaire(fieldPath, params.CollectedData, logger)

			if value != nil {
				logger.Info("Resolved field path",
					zap.String("field_path", fieldPath),
					zap.String("value_type", fmt.Sprintf("%T", value)),
					zap.Any("value", value),
				)

				if at, ok := value.(string); ok && at != "" {
					agentType = at
				} else {
					logger.Warn("Resolved value is not a string",
						zap.String("type", fmt.Sprintf("%T", value)),
						zap.Any("value", value),
					)
				}
			} else {
				logger.Error("Field path resolution failed",
					zap.String("field_path", fieldPath),
					zap.Strings("available_keys", datahelpers.GetMapKeys(params.CollectedData)),
				)
			}
		}
	}

	if agentType == "" {
		return nil, fmt.Errorf("could not resolve agent_type - provide 'agent_type', 'agent_type_field', 'group_type', or 'group_type_field' in config. Config keys: %v. Tried to extract from: %v",
			datahelpers.GetMapKeys(config),
			datahelpers.GetMapKeys(params.CollectedData),
		)
	}

	logger.Info("Resolved agent type", zap.String("agent_type", agentType))

	// Query agent_definitions for briefing_questionnaire
	questionnaireJSON, err := queryBriefingQuestionnaire(ctx, params.DB, agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch questionnaire for agent type %s: %w", agentType, err)
	}

	// Parse questionnaire
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

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
