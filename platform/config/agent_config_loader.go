// FILE: platform/config/agent_config_loader.go
package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AgentConfigLoader handles loading agent configurations from the database
type AgentConfigLoader struct {
	logger *zap.Logger
}

// NewAgentConfigLoader creates a new agent config loader
func NewAgentConfigLoader(logger *zap.Logger) *AgentConfigLoader {
	return &AgentConfigLoader{logger: logger}
}

// LoadFromDatabase fetches agent configuration from a client-specific schema
func (l *AgentConfigLoader) LoadFromDatabase(ctx context.Context, db *pgxpool.Pool, clientID, agentInstanceID, agentType string) (*models.AgentConfig, error) {
	query := fmt.Sprintf(`
		SELECT name, config, template_id 
		FROM client_%s.agent_instances 
		WHERE id = $1 AND is_active = true
	`, clientID)

	var name string
	var configJSON []byte
	var templateID string

	err := db.QueryRow(ctx, query, agentInstanceID).Scan(&name, &configJSON, &templateID)
	if err != nil {
		if err == pgx.ErrNoRows {
			l.logger.Warn("Agent instance not found, using default configuration",
				zap.String("agent_instance_id", agentInstanceID))
			return l.GetDefaultConfig(agentInstanceID, agentType), nil
		}
		return nil, fmt.Errorf("failed to query agent instance: %w", err)
	}

	// Parse and build config
	return l.parseConfig(configJSON, agentInstanceID, agentType)
}

// LoadFromJSON loads agent configuration from JSON data
func (l *AgentConfigLoader) LoadFromJSON(data []byte, agentInstanceID, agentType string) (*models.AgentConfig, error) {
	return l.parseConfig(data, agentInstanceID, agentType)
}

// GetDefaultConfig returns a default configuration for an agent type
func (l *AgentConfigLoader) GetDefaultConfig(agentInstanceID, agentType string) *models.AgentConfig {
	return &models.AgentConfig{
		AgentID:   agentInstanceID,
		AgentType: agentType,
		Version:   1,
		CoreLogic: l.getDefaultCoreLogic(agentType),
		Workflow:  l.getDefaultWorkflow(agentType),
	}
}

func (l *AgentConfigLoader) parseConfig(configJSON []byte, agentInstanceID, agentType string) (*models.AgentConfig, error) {
	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	// Extract workflow
	var workflow models.WorkflowPlan
	if workflowData, ok := config["workflow"]; ok {
		workflowBytes, _ := json.Marshal(workflowData)
		if err := json.Unmarshal(workflowBytes, &workflow); err != nil {
			l.logger.Warn("Failed to parse workflow, using default", zap.Error(err))
			workflow = l.getDefaultWorkflow(agentType)
		}
	} else {
		workflow = l.getDefaultWorkflow(agentType)
	}

	// Extract memory configuration
	var memoryConfig models.MemoryConfiguration
	if memData, ok := config["memory_config"]; ok {
		memBytes, _ := json.Marshal(memData)
		json.Unmarshal(memBytes, &memoryConfig)
	}

	return &models.AgentConfig{
		AgentID:      agentInstanceID,
		AgentType:    agentType,
		Version:      1,
		CoreLogic:    config,
		Workflow:     workflow,
		MemoryConfig: memoryConfig,
	}, nil
}

func (l *AgentConfigLoader) getDefaultCoreLogic(agentType string) map[string]interface{} {
	// Different defaults for different agent types
	switch agentType {
	case "copywriter":
		return map[string]interface{}{
			"model":       "claude-3-sonnet",
			"temperature": 0.7,
			"max_tokens":  2000,
		}
	case "researcher":
		return map[string]interface{}{
			"model":       "claude-3-opus",
			"temperature": 0.3,
			"max_tokens":  4000,
		}
	default:
		return map[string]interface{}{
			"model":       "claude-3-haiku",
			"temperature": 0.5,
			"max_tokens":  1000,
		}
	}
}

func (l *AgentConfigLoader) getDefaultWorkflow(agentType string) models.WorkflowPlan {
	// Could have type-specific default workflows
	return models.WorkflowPlan{
		StartStep: "process",
		Steps: map[string]models.Step{
			"process": {
				Action:      "ai_text_generate",
				Description: "Process the request",
				NextStep:    "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Mark workflow as complete",
			},
		},
	}
}
