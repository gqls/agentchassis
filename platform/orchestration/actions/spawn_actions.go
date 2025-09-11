// FILE: platform/orchestration/actions/spawn_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// SpawnAgentAction spawns a single agent with hierarchy tracking
func SpawnAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SpawnAgentAction starting",
		zap.Any("config", params.StepConfig))

	// Extract agent configuration
	agentType, ok := params.StepConfig.Config["agent_type"].(string)
	if !ok || agentType == "" {
		return nil, fmt.Errorf("agent_type is required")
	}

	role, _ := params.StepConfig.Config["role"].(string)
	if role == "" {
		role = agentType // Use agent type as role if not specified
	}

	// Generate unique agent ID and name
	agentID := uuid.New().String()
	agentName := GenerateAgentName(agentType, role)

	// Pre-generate request ID (critical for response matching)
	requestID := params.Headers["request_id"]
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	params.Logger.Info("Spawning agent",
		zap.String("agent_id", agentID),
		zap.String("agent_name", agentName),
		zap.String("agent_type", agentType),
		zap.String("role", role),
		zap.String("request_id", requestID))

	// Create subtree info for hierarchy tracking
	subtreeInfo := &types.SubtreeInfo{
		AgentID:       agentID,
		AgentType:     agentType,
		AgentName:     agentName,
		ParentAgentID: params.ExecutionContext.FromAgentID,
		Children:      make(map[string]*types.SubtreeInfo),
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted:   0,
			TasksFailed:      0,
			AverageLatencyMs: 0,
			FuelConsumed:     0,
			LastUpdated:      time.Now(),
		},
	}

	/*	// Store in orchestration state if we have DB access
		if params.DB != nil && params.ExecutionContext.OrchestrationID != "" {
			repo := orchestration.NewStateRepository(params.DB, params.Logger)
			if err := repo.AddSubtreeAgent(params.Context, params.ExecutionContext.OrchestrationID, subtreeInfo); err != nil {
				params.Logger.Error("Failed to add agent to subtree",
					zap.Error(err),
					zap.String("agent_id", agentID))
				// Don't fail the spawn, just log the error
			}
		}*/

	// Note: We can't directly add to state repository here due to circular import
	// Instead, return the subtree info and let the coordinator handle it

	// Create spawn message
	spawnMessage := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: params.ExecutionContext.FromAgentType,
				AgentID:   params.ExecutionContext.FromAgentID,
				PodName:   params.ExecutionContext.ProcessingNode,
			},
			CorrelationID:         params.ExecutionContext.CorrelationID,
			CorrelationName:       params.ExecutionContext.CorrelationName,
			OrchestrationID:       params.ExecutionContext.OrchestrationID,
			OrchestrationName:     params.ExecutionContext.OrchestrationName,
			ParentOrchestrationID: params.ExecutionContext.ParentOrchestrationID,
			StepID:                params.ExecutionContext.StepID,
			StepName:              "spawn_agent",
			RequestID:             requestID,
			RetryVersion:          0,
			MessageID:             uuid.New().String(),
			MessageType:           "request",
			ToAgent:               agentID,
			ToAgentType:           agentType,
			Action:                "initialize",
			Timestamp:             time.Now(),
			FuelBudget:            params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds:        30,
			ResponsesTopic:        fmt.Sprintf("system.agent.%s.responses", params.ExecutionContext.FromAgentType),
		},
		Body: map[string]interface{}{
			"agent_id":   agentID,
			"agent_name": agentName,
			"role":       role,
			"config":     params.StepConfig.Config["config"],
			"parent_info": map[string]interface{}{
				"parent_agent_id":         params.ExecutionContext.FromAgentID,
				"parent_agent_type":       params.ExecutionContext.FromAgentType,
				"parent_orchestration_id": params.ExecutionContext.OrchestrationID,
			},
		},
	}

	// Send spawn message to agent's request topic
	targetTopic := fmt.Sprintf("system.agent.%s.requests", agentType)

	messageBytes, err := json.Marshal(spawnMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spawn message: %w", err)
	}

	if err := params.Producer.Produce(ctx, targetTopic, spawnMessage.Headers.ToMap(), []byte(requestID), messageBytes); err != nil {
		return nil, fmt.Errorf("failed to send spawn message: %w", err)
	}

	params.Logger.Info("Agent spawn message sent",
		zap.String("agent_id", agentID),
		zap.String("topic", targetTopic),
		zap.String("request_id", requestID))

	// Return result with await_response flag
	return map[string]interface{}{
		"agent_id":          agentID,
		"agent_name":        agentName,
		"agent_type":        agentType,
		"role":              role,
		"request_id":        requestID,
		"await_response":    true, // Critical: tells orchestrator to wait
		"target_agent_type": agentType,
		"subtree_info":      subtreeInfo, // Return for coordinator to handle
		"topic_sent_to":     targetTopic,
	}, nil
}

// SpawnGroupAction spawns a group of agents with hierarchy tracking
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SpawnGroupAction starting",
		zap.Any("config", params.StepConfig))

	// Extract group configuration
	groupType, ok := params.StepConfig.Config["group_type"].(string)
	if !ok || groupType == "" {
		return nil, fmt.Errorf("group_type is required")
	}

	agents, ok := params.StepConfig.Config["agents"].(map[string]interface{})
	if !ok || len(agents) == 0 {
		return nil, fmt.Errorf("agents configuration is required")
	}

	// Generate group ID and name
	groupID := uuid.New().String()
	groupName := GenerateGroupName(groupType)

	// Pre-generate request ID for the group
	requestID := params.Headers["request_id"]
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	params.Logger.Info("Spawning agent group",
		zap.String("group_id", groupID),
		zap.String("group_name", groupName),
		zap.String("group_type", groupType),
		zap.Int("agent_count", len(agents)),
		zap.String("request_id", requestID))

	// Track spawned agents
	spawnedAgents := make(map[string]string)               // role -> agentID
	spawnedSubtrees := make(map[string]*types.SubtreeInfo) // Use types.SubtreeInfo

	// Spawn each agent in the group
	for role, agentConfig := range agents {
		agentConfigMap, ok := agentConfig.(map[string]interface{})
		if !ok {
			params.Logger.Error("Invalid agent config",
				zap.String("role", role))
			continue
		}

		agentType, ok := agentConfigMap["type"].(string)
		if !ok {
			params.Logger.Error("Agent type not specified",
				zap.String("role", role))
			continue
		}

		// Create spawn config for this agent
		spawnConfig := models.Step{
			Action: "spawn_agent",
			Config: map[string]interface{}{
				"agent_type": agentType,
				"role":       role,
				"config":     agentConfigMap["config"],
				"group_id":   groupID,
				"group_name": groupName,
			},
		}

		// Use SpawnAgentAction to spawn individual agent
		spawnParams := params // Copy params
		spawnParams.StepConfig = spawnConfig

		result, err := SpawnAgentAction(ctx, spawnParams)
		if err != nil {
			params.Logger.Error("Failed to spawn agent",
				zap.String("role", role),
				zap.Error(err))
			continue
		}

		agentResult, ok := result.(map[string]interface{})
		if !ok {
			params.Logger.Error("Unexpected result type from SpawnAgentAction",
				zap.String("role", role))
			continue
		}

		agentID, ok := agentResult["agent_id"].(string)
		if !ok {
			params.Logger.Error("No agent_id in result",
				zap.String("role", role))
			continue
		}

		spawnedAgents[role] = agentID

		// Track subtree info
		if subtreeInfo, ok := agentResult["subtree_info"].(*types.SubtreeInfo); ok {
			spawnedSubtrees[agentID] = subtreeInfo
		}

		params.Logger.Info("Successfully spawned agent in group",
			zap.String("role", role),
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType),
			zap.String("group_id", groupID))
	}

	params.Logger.Info("All agents in group spawned",
		zap.String("group_id", groupID),
		zap.Int("total_spawned", len(spawnedAgents)),
		zap.Any("spawned_agents", spawnedAgents))

	// Get or create workflow for the group
	workflow := extractGroupWorkflow(params.StepConfig.Config)
	if workflow == nil {
		// Create default workflow
		workflow = createDefaultGroupWorkflow()
	}

	// Validate and marshal workflow
	workflowBytes, err := json.Marshal(workflow)
	if err != nil {
		params.Logger.Error("Failed to marshal workflow",
			zap.Error(err))
		workflowBytes = []byte(`{"start_step": "execute", "steps": {"execute": {"action": "execute_llm_prompt"}}}`)
	}

	// Create group subtree info using types.SubtreeInfo
	groupSubtree := &types.SubtreeInfo{
		AgentID:       groupID,
		AgentType:     "group",
		AgentName:     groupName,
		ParentAgentID: params.ExecutionContext.FromAgentID,
		Children:      spawnedSubtrees,
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted: 0,
			TasksFailed:    0,
			LastUpdated:    time.Now(),
		},
	}

	params.Logger.Info("SpawnGroupAction completed",
		zap.String("group_id", groupID),
		zap.String("request_id", requestID),
		zap.Int("agents_spawned", len(spawnedAgents)))

	// Return result
	return map[string]interface{}{
		"group_id":          groupID,
		"group_name":        groupName,
		"agents":            spawnedAgents,
		"workflow":          workflowBytes,
		"request_id":        requestID,
		"await_response":    true, // Wait for group initialization
		"target_agent_type": groupType,
		"subtree_info":      groupSubtree, // Return for coordinator to handle
		"spawn_count":       len(spawnedAgents),
	}, nil
}

// Helper functions

func GenerateAgentName(agentType, role string) string {
	timestamp := time.Now().Format("0102-1504")
	if role != "" && role != agentType {
		return fmt.Sprintf("%s-%s-%s", agentType, role, timestamp)
	}
	return fmt.Sprintf("%s-%s", agentType, timestamp)
}

func GenerateGroupName(groupType string) string {
	timestamp := time.Now().Format("0102-1504")
	return fmt.Sprintf("%s-group-%s", groupType, timestamp)
}

func extractGroupWorkflow(config map[string]interface{}) map[string]interface{} {
	if workflow, ok := config["workflow"].(map[string]interface{}); ok {
		return workflow
	}

	// Try to extract from JSON string
	if workflowStr, ok := config["workflow"].(string); ok {
		var workflow map[string]interface{}
		if err := json.Unmarshal([]byte(workflowStr), &workflow); err == nil {
			return workflow
		}
	}

	return nil
}

func createDefaultGroupWorkflow() map[string]interface{} {
	return map[string]interface{}{
		"start_step": "execute",
		"steps": map[string]interface{}{
			"execute": map[string]interface{}{
				"action":      "execute_llm_prompt",
				"description": "Execute group task",
				"next_step":   "complete",
			},
			"complete": map[string]interface{}{
				"action":      "complete_workflow",
				"description": "Complete group workflow",
			},
		},
	}
}

// StartOrchestrationAction starts a child orchestration
func StartOrchestrationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StartOrchestrationAction starting")

	// Check if we're resuming from a previous attempt
	if existingChild := checkExistingChild(params); existingChild != nil {
		params.Logger.Info("Found existing child orchestration, not spawning new one")
		return existingChild, nil
	}

	// Find spawn data from previous steps
	spawnData, err := findSpawnData(params)
	if err != nil {
		return nil, fmt.Errorf("failed to find spawn data: %w", err)
	}

	// Extract workflow
	workflow, err := extractWorkflow(spawnData, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow: %w", err)
	}

	// Generate child orchestration ID
	childOrchID := uuid.New().String()
	childOrchName := fmt.Sprintf("child-%s-%s", params.ExecutionContext.OrchestrationName, time.Now().Format("1504"))

	// Pre-generate request ID
	requestID := params.Headers["request_id"]
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	// Create start orchestration message
	startMessage := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: params.ExecutionContext.FromAgentType,
				AgentID:   params.ExecutionContext.FromAgentID,
				PodName:   params.ExecutionContext.ProcessingNode,
			},
			CorrelationID:           params.ExecutionContext.CorrelationID,
			CorrelationName:         params.ExecutionContext.CorrelationName,
			OrchestrationID:         childOrchID,
			OrchestrationName:       childOrchName,
			ParentOrchestrationID:   params.ExecutionContext.OrchestrationID,
			ParentOrchestrationName: params.ExecutionContext.OrchestrationName,
			ParentRequestID:         requestID,
			StepID:                  uuid.New().String(),
			StepName:                "start_child_orchestration",
			RequestID:               requestID,
			MessageID:               uuid.New().String(),
			MessageType:             "request",
			Action:                  "start_orchestration",
			Timestamp:               time.Now(),
			FuelBudget:              params.ExecutionContext.FuelBudget - 200,
			TimeoutSeconds:          300,
		},
		Body: map[string]interface{}{
			"workflow":   workflow,
			"spawn_data": spawnData,
			"parent_info": map[string]interface{}{
				"parent_orchestration_id": params.ExecutionContext.OrchestrationID,
				"parent_request_id":       requestID,
			},
		},
	}

	// Determine target agent for orchestration
	targetAgentType := "orchestrator" // Default
	if agentType, ok := spawnData["agent_type"].(string); ok {
		targetAgentType = agentType
	}

	// Send to target agent's request topic
	targetTopic := fmt.Sprintf("system.agent.%s.requests", targetAgentType)

	messageBytes, err := json.Marshal(startMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal start message: %w", err)
	}

	if err := params.Producer.Produce(ctx, targetTopic, startMessage.Headers.ToMap(), []byte(requestID), messageBytes); err != nil {
		return nil, fmt.Errorf("failed to send start orchestration message: %w", err)
	}

	params.Logger.Info("Child orchestration started",
		zap.String("child_orchestration_id", childOrchID),
		zap.String("child_orchestration_name", childOrchName),
		zap.String("request_id", requestID),
		zap.String("target_topic", targetTopic))

	return map[string]interface{}{
		"child_orchestration_id":   childOrchID,
		"child_orchestration_name": childOrchName,
		"request_id":               requestID,
		"await_response":           true,
		"target_agent_type":        targetAgentType,
		"workflow":                 workflow,
	}, nil
}

// Helper functions for StartOrchestrationAction

func checkExistingChild(params ActionParams) interface{} {
	// Check if we already started a child orchestration
	if params.CollectedData["start_orchestration"] != nil {
		if childMap, ok := params.CollectedData["start_orchestration"].(map[string]interface{}); ok {
			if childID, ok := childMap["child_orchestration_id"].(string); ok && childID != "" {
				params.Logger.Info("Found existing child",
					zap.String("child_orchestration_id", childID))
				return childMap
			}
		}
	}
	return nil
}

func findSpawnData(params ActionParams) (map[string]interface{}, error) {
	// Look for spawn results in collected data
	for stepName, data := range params.CollectedData {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Check for spawn result indicators
			if dataMap["workflow"] != nil || dataMap["agents"] != nil || dataMap["group_id"] != nil {
				params.Logger.Debug("Found spawn result",
					zap.String("from_step", stepName))
				return dataMap, nil
			}
		}
	}
	return nil, fmt.Errorf("no spawn result found in collected data")
}

func extractWorkflow(spawnData map[string]interface{}, logger *zap.Logger) (json.RawMessage, error) {
	workflow, ok := spawnData["workflow"]
	if !ok {
		return nil, fmt.Errorf("workflow not found in spawn result")
	}

	// Convert to JSON based on type
	switch wf := workflow.(type) {
	case json.RawMessage:
		return wf, nil
	case []byte:
		return json.RawMessage(wf), nil
	case string:
		return json.RawMessage(wf), nil
	case map[string]interface{}:
		return json.Marshal(wf)
	default:
		return nil, fmt.Errorf("unexpected workflow type: %T", workflow)
	}
}
