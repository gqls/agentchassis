// platform/orchestration/actions/spawn_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/discovery"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
)

// SpawnAgentAction creates or reuses an agent instance
func SpawnAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	clientID := params.Headers["client_id"]
	if clientID == "" {
		return nil, fmt.Errorf("client_id not specified in headers")
	}

	// Check if agent already exists
	var existingID string
	query := fmt.Sprintf(`
        SELECT id FROM client_%s.agent_instances 
        WHERE config->>'agent_type' = $1 
        AND is_active = true
        LIMIT 1
    `, clientID)

	err := params.DB.QueryRowContext(ctx, query, agentType).Scan(&existingID)

	if err == nil {
		return map[string]interface{}{
			"agent_id": existingID,
			"topic":    fmt.Sprintf("system.agent.%s.process", agentType),
			"status":   "reused",
		}, nil
	}

	// Create new agent instance
	agentID := uuid.New().String()
	workflow := buildWorkflowForType(agentType)

	agentConfig := map[string]interface{}{
		"agent_type":   agentType,
		"workflow":     workflow,
		"topic":        fmt.Sprintf("system.agent.%s.process", agentType),
		"capabilities": getCapabilitiesForType(agentType),
	}

	configJSON, err := json.Marshal(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	insertQuery := fmt.Sprintf(`
        INSERT INTO client_%s.agent_instances 
        (id, template_id, owner_user_id, name, config, is_active)
        VALUES ($1, $2, $3, $4, $5, true)
    `, clientID)

	userID := params.Headers["user_id"]
	if userID == "" {
		userID = "system"
	}

	_, err = params.DB.ExecContext(ctx, insertQuery,
		agentID,
		"2a540b98-85d5-4762-a692-538bcf1be395", // generic template
		userID,
		fmt.Sprintf("%s-%s", agentType, time.Now().Format("20060102-150405")),
		configJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return map[string]interface{}{
		"agent_id": agentID,
		"topic":    fmt.Sprintf("system.agent.%s.process", agentType),
		"status":   "created",
	}, nil
}

// SpawnGroupAction spawns a complete agent group
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	groupType, ok := config["group_type"].(string)
	if !ok {
		return nil, fmt.Errorf("group_type not specified")
	}

	// Find the best matching group
	var groupID, groupName string
	var agentConfigs json.RawMessage
	var workflow json.RawMessage

	// Fix: Use QueryRowContext with context as first parameter
	err := params.DB.QueryRowContext(ctx, `
        SELECT id, name, agent_configs, orchestration_workflow
        FROM agent_groups
        WHERE group_type = $1
        ORDER BY usage_count DESC, version DESC
        LIMIT 1
    `, groupType).Scan(&groupID, &groupName, &agentConfigs, &workflow)

	if err != nil {
		return nil, fmt.Errorf("no group found for type %s: %w", groupType, err)
	}

	// Parse agent configs
	var agents []map[string]interface{}
	if err := json.Unmarshal(agentConfigs, &agents); err != nil {
		return nil, fmt.Errorf("failed to parse agent configs: %w", err)
	}

	// Spawn each agent in the group
	spawnedAgents := make(map[string]string)

	for _, agentConfig := range agents {
		role, ok := agentConfig["role"].(string)
		if !ok {
			continue
		}

		agentType, ok := agentConfig["agent_type"].(string)
		if !ok {
			continue
		}

		result, err := SpawnAgentAction(ctx, ActionParams{
			StepConfig: models.Step{
				Config: map[string]interface{}{
					"agent_type": agentType,
				},
			},
			Headers: params.Headers,
			DB:      params.DB,
			Logger:  params.Logger,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to spawn %s: %w", role, err)
		}

		agentResult, ok := result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected result type from SpawnAgentAction")
		}

		agentID, ok := agentResult["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id not found in spawn result")
		}

		spawnedAgents[role] = agentID
	}

	// Update group usage - Fix: Use ExecContext
	_, err = params.DB.ExecContext(ctx, `
        UPDATE agent_groups 
        SET usage_count = usage_count + 1, 
            last_used_at = NOW() 
        WHERE id = $1
    `, groupID)

	if err != nil && params.Logger != nil {
		params.Logger.Error("Failed to update group usage", zap.Error(err))
	}

	return map[string]interface{}{
		"group_id":   groupID,
		"group_name": groupName,
		"agents":     spawnedAgents,
		"workflow":   workflow,
	}, nil
}

// CallAgentAction calls another agent with discovery
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	// If specific agent_id provided, use it directly
	if agentID, ok := config["agent_id"].(string); ok {
		return callSpecificAgent(ctx, params, agentID)
	}

	// Otherwise, discover best agent
	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	// Try to find existing agent
	discover := NewAgentDiscovery(params.DB)
	if discover == nil {
		return nil, fmt.Errorf("failed to create discovery service")
	}

	matches, err := discover.DiscoverAgents(ctx, discovery.Requirements{
		AgentType: agentType,
		ClientID:  params.Headers["client_id"],
	})

	if err != nil || len(matches) == 0 {
		// No agent found, spawn one
		spawnResult, err := SpawnAgentAction(ctx, ActionParams{
			StepConfig: models.Step{
				Config: map[string]interface{}{
					"agent_type": agentType,
				},
			},
			Headers: params.Headers,
			DB:      params.DB,
			Logger:  params.Logger,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to spawn agent: %w", err)
		}

		sr, ok := spawnResult.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected spawn result type")
		}

		agentID, ok := sr["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id not found in spawn result")
		}

		return callSpecificAgent(ctx, params, agentID)
	}

	// Use best match
	return callSpecificAgent(ctx, params, matches[0].AgentID)
}

// callSpecificAgent calls a specific agent by ID
func callSpecificAgent(ctx context.Context, params ActionParams, agentID string) (interface{}, error) {
	clientID := params.Headers["client_id"]
	if clientID == "" {
		return nil, fmt.Errorf("client_id not specified")
	}

	// Get agent configuration to find its topic
	var agentConfig json.RawMessage

	query := fmt.Sprintf(`
        SELECT config FROM client_%s.agent_instances 
        WHERE id = $1
    `, clientID)

	err := params.DB.QueryRowContext(ctx, query, agentID).Scan(&agentConfig)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(agentConfig, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent config: %w", err)
	}

	topic, ok := config["topic"].(string)
	if !ok {
		return nil, fmt.Errorf("topic not found in agent config")
	}

	// Prepare payload
	payload := models.TaskRequest{
		Action: params.StepConfig.Action,
		Data:   params.CollectedData,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create headers
	newRequestID := uuid.New().String()
	outHeaders := make(map[string]string)
	for k, v := range params.Headers {
		outHeaders[k] = v
	}
	outHeaders["causation_id"] = params.Headers["request_id"]
	outHeaders["request_id"] = newRequestID
	outHeaders["target_agent_id"] = agentID

	// Send message
	err = params.Producer.Produce(ctx, topic, outHeaders,
		[]byte(params.Headers["correlation_id"]), payloadBytes)

	if err != nil {
		return nil, fmt.Errorf("failed to call agent: %w", err)
	}

	return map[string]interface{}{
		"agent_called": agentID,
		"request_id":   newRequestID,
		"topic":        topic,
	}, nil
}

// DiscoverAgentsAction for explicit discovery
func DiscoverAgentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	discover := NewAgentDiscovery(params.DB)
	if discover == nil {
		return nil, fmt.Errorf("failed to create discovery service")
	}

	requirements := discovery.Requirements{
		AgentType: agentType,
		ClientID:  params.Headers["client_id"],
	}

	if caps, ok := config["capabilities"].([]string); ok {
		requirements.Capabilities = caps
	}

	matches, err := discover.DiscoverAgents(ctx, requirements)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"found_agents": len(matches),
		"agents":       matches,
	}, nil
}

// NewAgentDiscovery creates a discovery service from the database connection
func NewAgentDiscovery(db interface{}) *discovery.AgentDiscovery {
	switch d := db.(type) {
	case *sql.DB:
		// For sql.DB, we need connection string to create pgxpool
		// This is a limitation - in production, pass pgxpool directly
		connStr := "postgres://user:pass@localhost/db" // This needs proper config
		config, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			return nil
		}
		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			return nil
		}
		return discovery.NewAgentDiscovery(pool)
	case *pgxpool.Pool:
		return discovery.NewAgentDiscovery(d)
	default:
		return nil
	}
}

// Helper functions

func buildWorkflowForType(agentType string) map[string]interface{} {
	// This would typically load from templates or configuration
	// For now, return a basic workflow
	return map[string]interface{}{
		"start_step": "process",
		"steps": map[string]interface{}{
			"process": map[string]interface{}{
				"action":    "process_task",
				"next_step": "complete",
			},
			"complete": map[string]interface{}{
				"action": "complete_workflow",
			},
		},
	}
}

func getCapabilitiesForType(agentType string) []string {
	capabilities := map[string][]string{
		"copywriter":      {"writing", "marketing", "creative"},
		"researcher":      {"research", "analysis", "data"},
		"developer":       {"coding", "html", "css", "javascript"},
		"designer":        {"design", "ui", "ux", "graphics"},
		"domain-analyst":  {"analysis", "research", "categorization"},
		"site-architect":  {"planning", "structure", "navigation"},
		"html-developer":  {"html", "css", "javascript", "frontend"},
		"visual-designer": {"design", "graphics", "branding"},
		"site-publisher":  {"deployment", "hosting", "publishing"},
	}

	if caps, ok := capabilities[agentType]; ok {
		return caps
	}
	return []string{agentType}
}
