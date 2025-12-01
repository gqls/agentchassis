// FILE: platform/orchestration/actions/agent/agent.go
// Package agent provides agent management actions: spawn, call, discover
package agent

import (
	"context"
	"fmt"

	"github.com/aqls/personae/platform/orchestration/actions/registry"
)

func init() {
	// Register all agent management actions
	registry.Register("spawn_agent", registry.ActionDefinition{
		Func:        SpawnAgentAction,
		Category:    registry.CategoryAgent,
		Description: "Spawns a new agent instance (K8s pod or in-process)",
		Status:      registry.StatusActive,
	})

	registry.Register("spawn_agent_k8s", registry.ActionDefinition{
		Func:        SpawnAgentAction, // Alias to spawn_agent
		Category:    registry.CategoryAgent,
		Description: "Spawns agent in Kubernetes (alias for spawn_agent)",
		Status:      registry.StatusDeprecated,
	})

	registry.Register("spawn_group", registry.ActionDefinition{
		Func:        SpawnGroupAction,
		Category:    registry.CategoryAgent,
		Description: "Spawns multiple agents as a group",
		Status:      registry.StatusActive,
	})

	registry.Register("call_agent", registry.ActionDefinition{
		Func:        CallAgentAction,
		Category:    registry.CategoryAgent,
		Description: "Sends a message to an agent and awaits response",
		Status:      registry.StatusActive,
	})

	registry.Register("discover_agents", registry.ActionDefinition{
		Func:        DiscoverAgentsAction,
		Category:    registry.CategoryAgent,
		Description: "Discovers available agents matching criteria",
		Status:      registry.StatusActive,
	})

	registry.Register("start_orchestration", registry.ActionDefinition{
		Func:        StartOrchestrationAction,
		Category:    registry.CategoryAgent,
		Description: "Starts a new orchestration workflow",
		Status:      registry.StatusActive,
	})
}

// SpawnAgentAction spawns a new agent instance
// This is a stub - the real implementation would be migrated from spawn_actions.go
func SpawnAgentAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	agentType, ok := params.Config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("spawn_agent requires 'agent_type' in config")
	}

	role, _ := params.Config["role"].(string)

	params.Logger.Info("Spawning agent") // zap.String("agent_type", agentType),
	// zap.String("role", role),

	// TODO: Migrate actual spawn logic from spawn_actions.go
	// This includes:
	// - Getting agent definition from database
	// - Creating K8s deployment or in-process agent
	// - Setting up topic subscriptions
	// - Registering in spawned_agents map

	return map[string]interface{}{
		"status":     "spawned",
		"agent_type": agentType,
		"role":       role,
		// "agent_id": generatedAgentID,
		// "topics": topicsMap,
	}, nil
}

// SpawnGroupAction spawns multiple agents
func SpawnGroupAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	agents, ok := params.Config["agents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("spawn_group requires 'agents' array in config")
	}

	params.Logger.Info("Spawning agent group") // zap.Int("count", len(agents)),

	// TODO: Migrate from spawn_actions.go
	// Spawn each agent in the group

	return map[string]interface{}{
		"status": "group_spawned",
		"count":  len(agents),
	}, nil
}

// CallAgentAction sends a message to an agent and waits for response
// This is a stub - the real implementation would be migrated from call_agent.go
func CallAgentAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	agentType, ok := params.Config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("call_agent requires 'agent_type' in config")
	}

	targetRole, _ := params.Config["target_role"].(string)
	timeoutSeconds := 30
	if ts, ok := params.Config["timeout_seconds"].(float64); ok {
		timeoutSeconds = int(ts)
	}

	params.Logger.Info("Calling agent") // zap.String("agent_type", agentType),
	// zap.String("target_role", targetRole),
	// zap.Int("timeout_seconds", timeoutSeconds),

	// TODO: Migrate from call_agent.go
	// This includes:
	// - Building the message from input_fields
	// - Sending to agent's request topic
	// - Awaiting response on response topic
	// - Handling timeout

	return map[string]interface{}{
		"status":     "called",
		"agent_type": agentType,
		"role":       targetRole,
	}, nil
}

// DiscoverAgentsAction finds available agents
func DiscoverAgentsAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// TODO: Migrate from discovery_actions.go
	return map[string]interface{}{
		"status": "discovered",
		"agents": []string{},
	}, nil
}

// StartOrchestrationAction starts a new workflow
func StartOrchestrationAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	groupType, ok := params.Config["group_type"].(string)
	if !ok {
		return nil, fmt.Errorf("start_orchestration requires 'group_type' in config")
	}

	params.Logger.Info("Starting orchestration") // zap.String("group_type", groupType),

	// TODO: Migrate orchestration start logic

	return map[string]interface{}{
		"status":     "started",
		"group_type": groupType,
	}, nil
}
