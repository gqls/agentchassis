package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// CallAgentAction creates child orchestrations with proper context
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get parent execution context
	parentCtx, err := types.FromHeaders(params.Headers)
	if err != nil {
		return nil, fmt.Errorf("invalid parent context: %w", err)
	}

	// Validate parent context
	if parentCtx.OrchestrationID == "" {
		return nil, fmt.Errorf("parent orchestration_id required")
	}

	// Check for idempotency
	stepKey := fmt.Sprintf("%s_request", params.CurrentStep)
	if existingRequest, exists := params.CollectedData[stepKey]; exists {
		if reqMap, ok := existingRequest.(map[string]interface{}); ok {
			if reqID, ok := reqMap["request_id"].(string); ok && reqID != "" {
				return existingRequest, nil
			}
		}
	}

	// Get target agent details
	config := params.StepConfig.Config
	targetAgentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified in config")
	}

	targetAgentID, err := findOrSpawnAgent(ctx, params, targetAgentType)
	if err != nil {
		return nil, fmt.Errorf("failed to find or spawn agent: %w", err)
	}

	// Create child execution context
	childCtx := parentCtx.CreateChildContext(targetAgentID, targetAgentType)

	// Build message for child
	msg := models.AgentMessage{
		MessageID:       childCtx.MessageID,
		RequestID:       childCtx.RequestID,
		CorrelationID:   childCtx.CorrelationID,
		OrchestrationID: childCtx.OrchestrationID,
		FromAgentID:     childCtx.FromAgentID,
		ToAgentID:       childCtx.ToAgentID,
		MessageType:     childCtx.MessageType,
		Action:          "process",
		Data:            params.CollectedData,
		Timestamp:       childCtx.Timestamp,
		Version:         childCtx.Version,
	}

	// Add parent context to message data for child to use when responding
	msg.Data["__execution_context__"] = childCtx

	// Track request in parent's orchestration
	trackRequest(ctx, params.DB, childCtx.RequestID, parentCtx.OrchestrationID, targetAgentID)

	// Send to child's topic
	targetTopic := fmt.Sprintf("system.agent.%s.requests", targetAgentType)
	msgBytes, _ := json.Marshal(msg)

	err = params.Producer.Produce(ctx, targetTopic, childCtx.ToHeaders(),
		[]byte(childCtx.OrchestrationID), msgBytes)
	if err != nil {
		failRequest(ctx, params.DB, childCtx.RequestID)
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Store result
	result := map[string]interface{}{
		"agent_called":        targetAgentID,
		"agent_type":          targetAgentType,
		"request_id":          childCtx.RequestID,
		"child_orchestration": childCtx.OrchestrationID,
		"await_response":      true,
	}

	params.CollectedData[stepKey] = result
	return result, nil
}
