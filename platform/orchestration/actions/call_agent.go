package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("In CallAgentAction")

	config := params.StepConfig.Config
	targetAgentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified in config")
	}

	// Get the previously spawned agent
	var targetAgentID string
	var spawnedAgentInfo map[string]interface{}

	// Check spawn_calculator step result
	spawnKey := fmt.Sprintf("spawn_%s", targetAgentType)
	if spawnResult, ok := params.CollectedData[spawnKey].(map[string]interface{}); ok {
		targetAgentID, _ = spawnResult["agent_id"].(string)
		spawnedAgentInfo = spawnResult
		params.Logger.Info("Found spawned agent",
			zap.String("agent_type", targetAgentType),
			zap.String("agent_id", targetAgentID),
			zap.String("DEBUGaa: spawn_key", spawnKey)) // ?
	}

	if targetAgentID == "" {
		return nil, fmt.Errorf("no spawned %s agent found in step %s", targetAgentType, spawnKey)
	}

	// Check if we're still waiting for the spawn to complete
	if awaitResponse, ok := spawnedAgentInfo["await_response"].(bool); ok && awaitResponse {
		// The spawn step should have completed before we get here
		// This is a workflow ordering issue
		return nil, fmt.Errorf("spawn step %s has not completed yet", spawnKey)
	}

	requestID := uuid.New().String()

	// Create child orchestration ID for this call
	childOrchID := uuid.New().String() // todo: who's orchestration is this?
	childOrchName := fmt.Sprintf("%s-calc-%s", targetAgentType, time.Now().Format("1504"))

	// Build the calculation request
	// The calculator already has the data from initialization, but we send it again
	// to ensure context is preserved
	actionRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: params.ExecutionContext.Sender,

			CorrelationID:   params.ExecutionContext.CorrelationID,
			CorrelationName: params.ExecutionContext.CorrelationName,
			ClientID:        params.ExecutionContext.ClientID,

			// Child orchestration context
			OrchestrationID:   childOrchID,
			OrchestrationName: childOrchName,
			StepID:            uuid.New().String(),
			StepName:          "process", // Match workflow start step
			RequestID:         requestID,
			RetryVersion:      0,

			// Parent tracking
			ParentOrchestrationID:   params.ExecutionContext.OrchestrationID,
			ParentOrchestrationName: params.ExecutionContext.OrchestrationName,
			ParentRequestID:         params.ExecutionContext.RequestID,

			MessageID:   uuid.New().String(),
			MessageType: "request",
			FromAgent:   params.ExecutionContext.Sender.AgentID,
			ToAgent:     targetAgentID,
			ToAgentType: targetAgentType,
			Action:      "process", // Trigger the workflow
			Timestamp:   time.Now(),

			FuelBudget:     params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds: 30,

			ResponsesTopic: fmt.Sprintf("system.agent.%s.responses",
				params.ExecutionContext.Sender.AgentType),
		},
		//Body: params.CollectedData["input_data"], // Send the calculation data
		Body: map[string]interface{}{
			"input_data": params.CollectedData["input_data"], // Wrap it in the expected key
		},
	}

	params.Logger.Info("DEBUGaa: 2 CallAgentAction RequestMessage for calculator actionSending calculation request",
		zap.Any("request message", actionRequest),
		zap.String("Action", "process"),
		zap.String("child_orch_id", childOrchID),
		zap.String("orchestration id is child orchestration id", actionRequest.Headers.OrchestrationID),
	)

	// Send to calculator's requests topic
	targetTopic := fmt.Sprintf("system.agent.%s.requests", targetAgentType)
	msgBytes, err := json.Marshal(actionRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := actionRequest.Headers.ToMap()
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("Sending calculation request",
		zap.String("target_topic", targetTopic),
		zap.String("request_id", requestID),
		zap.String("child_orch_id", childOrchID),
		zap.String("action", "process"))

	err = params.Producer.Produce(ctx, targetTopic, headers, key, msgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send CallAgentAction request: %w", err)
	}

	// Return result indicating we're waiting for the calculation
	result := map[string]interface{}{
		"agent_called":        targetAgentID,
		"agent_type":          targetAgentType,
		"request_id":          requestID,
		"child_orchestration": childOrchID,
		"action_sent":         "process",
		"await_response":      true,
		"target_agent_type":   targetAgentType,
	}

	params.Logger.Info("Call agent action completed, awaiting response",
		zap.String("request_id", requestID))

	return result, nil
}

// CallAgentAction creates child orchestrations with proper context
func CallAgentActionOLD(ctx context.Context, params ActionParams) (interface{}, error) {
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
	childCtx := parentCtx.CreateChildContext(targetAgentType)
	childCtx.ToAgentID = targetAgentID

	// Build message for child
	msg := models.AgentMessage{
		MessageID:         childCtx.MessageID,
		RequestID:         childCtx.RequestID,
		CorrelationID:     childCtx.CorrelationID,
		OrchestrationID:   childCtx.OrchestrationID,
		OrchestrationName: childCtx.OrchestrationName,
		FromAgentID:       childCtx.FromAgentID,
		ToAgentID:         childCtx.ToAgentID,
		MessageType:       childCtx.MessageType,
		Action:            "process",
		Data:              params.CollectedData,
		Timestamp:         childCtx.Timestamp,
		Version:           childCtx.Version,
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

func findOrSpawnAgent(ctx context.Context, params ActionParams, targetAgentType string) (string, error) {
	// Check if agent already exists in collected data
	agentKey := fmt.Sprintf("%s_agent_id", targetAgentType)
	if agentID, ok := params.CollectedData[agentKey].(string); ok && agentID != "" {
		params.Logger.Info("Found existing agent",
			zap.String("agent_type", targetAgentType),
			zap.String("agent_id", agentID))
		return agentID, nil
	}

	// Check if we have a spawned agent from a previous step
	if spawnResult, ok := params.CollectedData["spawn_agent"].(map[string]interface{}); ok {
		if spawnedType, _ := spawnResult["agent_type"].(string); spawnedType == targetAgentType {
			if agentID, ok := spawnResult["agent_id"].(string); ok && agentID != "" {
				params.Logger.Info("Using previously spawned agent",
					zap.String("agent_type", targetAgentType),
					zap.String("agent_id", agentID))
				return agentID, nil
			}
		}
	}

	// Need to spawn a new agent
	params.Logger.Info("Spawning new agent for call_agent",
		zap.String("target_type", targetAgentType))

	spawnConfig := params.StepConfig
	spawnConfig.Action = "spawn_agent"
	spawnConfig.Config = map[string]interface{}{
		"agent_type": targetAgentType,
		"role":       targetAgentType,
	}

	spawnParams := params
	spawnParams.StepConfig = spawnConfig

	result, err := SpawnAgentAction(ctx, spawnParams)
	if err != nil {
		return "", fmt.Errorf("failed to spawn agent: %w", err)
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if agentID, ok := resultMap["agent_id"].(string); ok {
			params.CollectedData[agentKey] = agentID
			return agentID, nil
		}
	}

	return "", fmt.Errorf("failed to get agent_id from spawn result")
}

func trackRequest(ctx context.Context, db *sql.DB, requestID, orchestrationID, targetAgentID string) {
	if db == nil {
		return
	}

	// Parse UUIDs
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		reqUUID = uuid.New() // Generate new if invalid
	}

	orchUUID, err := uuid.Parse(orchestrationID)
	if err != nil {
		return // Can't proceed without valid orchestration ID
	}

	targetUUID, err := uuid.Parse(targetAgentID)
	if err != nil {
		targetUUID = uuid.New() // Generate new if invalid
	}

	// Use the existing pending_requests table with UUID columns
	query := `
        INSERT INTO pending_requests 
        (request_id, orchestration_id, to_agent_id, status, timeout_at, created_at)
        VALUES ($1, $2, $3, 'pending', $4, NOW())
        ON CONFLICT (request_id) DO NOTHING
    `

	timeout := time.Now().Add(30 * time.Second)
	if _, err := db.ExecContext(ctx, query, reqUUID, orchUUID, targetUUID, timeout); err != nil {
		// Continue on error
	}

	// Also log to system_events table
	eventMetadata := map[string]interface{}{
		"request_id":       requestID,
		"target_agent_id":  targetAgentID,
		"orchestration_id": orchestrationID,
		"action":           "call_agent",
	}

	metadataJSON, _ := json.Marshal(eventMetadata)

	eventQuery := `
        INSERT INTO system_events 
        (event_type, entity_type, entity_id, metadata, severity, source, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `

	db.ExecContext(ctx, eventQuery,
		"AGENT_CALL",        // event_type
		"orchestration",     // entity_type
		orchestrationID,     // entity_id
		metadataJSON,        // metadata
		"info",              // severity
		"call_agent_action") // source
}

func failRequest(ctx context.Context, db *sql.DB, requestID string) {
	if db == nil {
		return
	}

	// Parse UUID
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return
	}

	// Update in pending_requests table
	query := `
        UPDATE pending_requests 
        SET status = 'failed', 
            completed_at = NOW()
        WHERE request_id = $1
    `

	if _, err := db.ExecContext(ctx, query, reqUUID); err != nil {
		// Continue on error
	}

	// Also need to update orchestration_states to remove from awaited_requests JSONB
	// This is more complex with JSONB
	updateOrchQuery := `
        UPDATE orchestration_states
        SET awaited_requests = awaited_requests - $1
        WHERE awaited_requests ? $1
    `

	db.ExecContext(ctx, updateOrchQuery, requestID)

	// Log failure to system_events
	eventMetadata := map[string]interface{}{
		"request_id": requestID,
		"reason":     "send_failed",
		"timestamp":  time.Now(),
	}

	metadataJSON, _ := json.Marshal(eventMetadata)

	eventQuery := `
        INSERT INTO system_events 
        (event_type, entity_type, entity_id, metadata, severity, source, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `

	db.ExecContext(ctx, eventQuery,
		"REQUEST_FAILED",    // event_type
		"request",           // entity_type
		requestID,           // entity_id
		metadataJSON,        // metadata
		"error",             // severity
		"call_agent_action") // source
}

func generateID() string {
	return uuid.New().String()
}

// Check if agent type exists in agent_definitions
func isValidAgentType(ctx context.Context, db *sql.DB, agentType string) bool {
	if db == nil {
		return true // Assume valid if we can't check
	}

	query := `
        SELECT EXISTS(
            SELECT 1 FROM agent_definitions 
            WHERE type = $1 AND is_active = true
        )
    `

	var exists bool
	err := db.QueryRowContext(ctx, query, agentType).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

// Log agent activity to system_events
func logAgentActivity(ctx context.Context, db *sql.DB, agentID, eventType, details string) {
	if db == nil {
		return
	}

	metadata := map[string]interface{}{
		"agent_id":  agentID,
		"details":   details,
		"timestamp": time.Now(),
	}

	metadataJSON, _ := json.Marshal(metadata)

	query := `
        INSERT INTO system_events 
        (event_type, entity_type, entity_id, metadata, severity, source, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `

	db.ExecContext(ctx, query,
		eventType,        // event_type
		"agent",          // entity_type
		agentID,          // entity_id
		metadataJSON,     // metadata
		"info",           // severity
		"agent_activity") // source
}
