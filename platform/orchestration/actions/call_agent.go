package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("In CallAgentAction",
		zap.Any("params in callagentaction", params))

	// Keep stack trace for debugging
	current, caller := getFuncInfo(1)
	params.Logger.Info("CallAgentAction starting",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.Any("params.Headers", params.Headers),
	)

	config := params.StepConfig.Config
	targetAgentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified in config")
	}

	// Check if we're looking for a specific role
	targetRole, hasRole := config["target_role"].(string)

	var targetAgentID string
	var targetRequestsTopic string  // Changed from jobTopic
	var targetResponsesTopic string // track child's response topic

	if hasRole && targetRole != "" {
		params.Logger.Info("Looking for agent by role",
			zap.String("target_role", targetRole))

		// Search through spawn results for matching role
		for stepName, stepData := range params.CollectedData {
			if spawnResult, ok := stepData.(map[string]interface{}); ok {

				params.Logger.Info("In CallAgentAction where is parent response topic?",
					zap.Any("spawn result", spawnResult))

				// Check both the role field and the step name
				if role, ok := spawnResult["role"].(string); ok && role == targetRole {
					targetAgentID, _ = spawnResult["agent_id"].(string)
					targetRequestsTopic, _ = spawnResult["requests_topic"].(string)   // Changed from job_topic
					targetResponsesTopic, _ = spawnResult["responses_topic"].(string) // NEW
					params.Logger.Info("Found agent with matching role",
						zap.String("role", targetRole),
						zap.String("agent_id", targetAgentID),
						zap.String("requests_topic", targetRequestsTopic),
						zap.String("responses_topic", targetResponsesTopic),
						zap.String("from_step", stepName))
					break
				}
			}
		}
	} else {
		// Look by agent type
		spawnKey := fmt.Sprintf("spawn_%s", targetAgentType)
		if spawnResult, ok := params.CollectedData[spawnKey].(map[string]interface{}); ok {
			targetAgentID, _ = spawnResult["agent_id"].(string)
			targetRequestsTopic, _ = spawnResult["requests_topic"].(string)
			targetResponsesTopic, _ = spawnResult["responses_topic"].(string)
		}
	}

	if targetAgentID == "" {
		if hasRole {
			return nil, fmt.Errorf("no agent with role '%s' found", targetRole)
		}
		return nil, fmt.Errorf("no spawned %s agent found", targetAgentType)
	}

	// If we didn't find topics in the first search, do a second pass
	if targetRequestsTopic == "" {
		if hasRole && targetRole != "" {
			targetRequestsTopic, targetResponsesTopic = findTopicsForRole(params, targetRole)
		} else {
			// Find by agent type/ID
			for stepName, stepData := range params.CollectedData {
				if strings.HasPrefix(stepName, "spawn_") {
					if spawnResult, ok := stepData.(map[string]interface{}); ok {
						if agentID, ok := spawnResult["agent_id"].(string); ok && agentID == targetAgentID {
							targetRequestsTopic, _ = spawnResult["requests_topic"].(string)
							targetResponsesTopic, _ = spawnResult["responses_topic"].(string)
							break
						}
					}
				}
			}
		}
	}

	// Handle special standard agents (search, image) that use legacy topics
	if targetRequestsTopic == "" {
		if targetAgentType == "search" || targetAgentType == "image" {
			targetRequestsTopic = fmt.Sprintf("system.agent.%s.requests", targetAgentType)
			targetResponsesTopic = fmt.Sprintf("system.agent.%s.responses", targetAgentType)
			params.Logger.Info("Using legacy topics for standard agent",
				zap.String("agent_type", targetAgentType),
				zap.String("requests_topic", targetRequestsTopic))
		} else {
			return nil, fmt.Errorf("no requests topic found for agent %s", targetAgentID)
		}
	}

	params.Logger.Info("Using agent topics",
		zap.String("requests_topic", targetRequestsTopic),
		zap.String("responses_topic", targetResponsesTopic),
		zap.String("target_agent", targetAgentID))

	requestID := uuid.New().String()

	// Create a NEW orchestration ID for the child's workflow
	childOrchestrationID := uuid.New().String()
	childOrchestrationName := fmt.Sprintf("%s-workflow-%s", targetAgentType, time.Now().Format("1504"))

	params.Logger.Info("CALL_AGENT: Starting agent call",
		zap.String("target_agent_type", targetAgentType),
		zap.String("target_agent_id", targetAgentID),
		zap.String("child_orch_id", childOrchestrationID),
		zap.String("parent_orch_id", params.ExecutionContext.OrchestrationID))

	// Determine the action to send - KEEP this logic as is
	var targetAction string
	step := params.StepConfig

	if action, ok := step.Config["target_action"].(string); ok {
		targetAction = action
	} else if action, ok := step.Config["action"].(string); ok {
		targetAction = action
	} else {
		targetAction = "process" // Default for backward compatibility
	}

	// Get which input field to use - KEEP this logic
	inputField := "input_data"
	if field, ok := step.Config["input_field"].(string); ok && field != "" {
		inputField = field
	}

	// Extract the specific data for this field - KEEP all this data extraction logic
	var dataToSend interface{}

	params.Logger.Info("DEBUG: CallAgentAction CollectedData contents",
		zap.Any("all_keys", getMapKeys(params.CollectedData)),
		zap.Any("has_input_data", params.CollectedData["input_data"] != nil),
		zap.Any("has_initial_request", params.CollectedData["InitialRequestData"] != nil))

	// First, check if there's input_data in CollectedData
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if inputField != "input_data" {
			// Looking for a specific field like "first_calc" or "second_calc"
			if fieldData, exists := inputData[inputField]; exists {
				dataToSend = fieldData
				params.Logger.Info("Found field in collected input_data",
					zap.String("field", inputField),
					zap.Any("data", dataToSend))
			}
		} else {
			dataToSend = inputData
		}
	}

	// If not found, check the initial request data
	if dataToSend == nil {
		if initialData, ok := params.CollectedData["InitialRequestData"].(map[string]interface{}); ok {
			if inputField != "input_data" {
				if fieldData, exists := initialData[inputField]; exists {
					dataToSend = fieldData
				}
			}
		}
	}

	params.Logger.Info("CallAgentAction data extraction",
		zap.String("requested_field", inputField),
		zap.Any("data_found", dataToSend),
		zap.Any("all_collected_keys", getMapKeys(params.CollectedData)))

	// Default to all input_data if we haven't found anything
	if dataToSend == nil {
		dataToSend = params.CollectedData["input_data"]
		params.Logger.Warn("Using full input_data as fallback (didnot find correct input data field)",
			zap.String("requested_field", inputField))
	}

	// Build the request message body - KEEP this structure
	requestBody := map[string]interface{}{
		"action":     targetAction,
		"input_data": dataToSend,
	}

	// Add any additional config if needed
	if agentConfig, ok := params.CollectedData["agent_config"]; ok {
		requestBody["config"] = agentConfig
	}

	// Include relevant context from previous steps if needed - KEEP this logic
	if includeContext, ok := params.StepConfig.Config["include_context"].(bool); ok && includeContext {
		if contextSteps, ok := params.StepConfig.Config["context_steps"].([]interface{}); ok {
			localContext := make(map[string]interface{})
			for _, step := range contextSteps {
				if stepName, ok := step.(string); ok {
					if stepData, exists := params.CollectedData[stepName]; exists {
						localContext[stepName] = stepData
					}
				}
			}
			requestBody["context"] = localContext
		}
	}

	// Build the request - check ResponsesTopic
	actionRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: params.ExecutionContext.Sender,

			CorrelationID:   params.ExecutionContext.CorrelationID,
			CorrelationName: params.ExecutionContext.CorrelationName,
			ClientID:        params.ExecutionContext.ClientID,

			// Child orchestration context
			OrchestrationID:   childOrchestrationID,
			OrchestrationName: childOrchestrationName,
			StepID:            uuid.New().String(),
			StepName:          "process",
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
			Action:      "process",
			Timestamp:   time.Now(),

			FuelBudget:     params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds: 30,

			// Use parent's ResponsesTopic so child knows where to respond
			ResponsesTopic: params.ExecutionContext.ResponsesTopic,
			// Child will create its own topics when it processes this
			RequestsTopic: "",
		},
		Body: requestBody,
	}

	params.Logger.Info("CallAgentAction RequestMessage prepared",
		zap.String("Action", "process"),
		zap.String("child_orch_id", childOrchestrationID),
		zap.String("parent_responses_topic", params.ExecutionContext.ResponsesTopic),
		zap.String("sending_to_topic", targetRequestsTopic),
		zap.String("orchestration_id", actionRequest.Headers.OrchestrationID),
		zap.Any("DEBUGaa: requestBody", requestBody),
	)

	msgBytes, err := json.Marshal(actionRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := actionRequest.Headers.ToMap()
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("Sending request to child agent",
		zap.String("target_topic", targetRequestsTopic),
		zap.String("request_id", requestID),
		zap.String("child_orch_id", childOrchestrationID),
		zap.String("action", "process"))

	// Send to child's requests topic
	err = params.Producer.Produce(ctx, targetRequestsTopic, headers, key, msgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send CallAgentAction request: %w", err)
	}

	if params.Tracer != nil {
		params.Tracer.TraceMessage(params.ExecutionContext, "SEND_CHILD_REQUEST CallAgentAction", targetRequestsTopic,
			map[string]interface{}{
				"responses_topic_set": actionRequest.Headers.ResponsesTopic,
				"parent_orch_id":      params.ExecutionContext.OrchestrationID,
				"child_orch_id":       childOrchestrationID,
				"request_id":          actionRequest.Headers.RequestID,
				"Sender":              params.ExecutionContext.Sender,
			})
	}

	params.Logger.Info("CALL_AGENT: Request sent to agent",
		zap.String("topic", targetRequestsTopic),
		zap.String("request_id", requestID),
		zap.String("child_orch_id", childOrchestrationID),
		zap.Any("headers_sent", headers))

	// Return result indicating we're waiting for the response
	result := map[string]interface{}{
		"agent_called":        targetAgentID,
		"agent_type":          targetAgentType,
		"request_id":          requestID,
		"child_orchestration": childOrchestrationID,
		"action_sent":         "process",
		"await_response":      true,
		"target_agent_type":   targetAgentType,
		// Track child's response topic for debugging
		"child_responses_topic": targetResponsesTopic,
	}

	params.Logger.Info("Call agent action completed, awaiting response",
		zap.String("request_id", requestID),
		zap.Any("returning result to say we are waiting for response", result))

	return result, nil
}

// Helper function updated for new topic names
func findTopicsForRole(params ActionParams, targetRole string) (string, string) {
	for stepName, stepData := range params.CollectedData {
		if strings.HasPrefix(stepName, "spawn_") {
			if spawnResult, ok := stepData.(map[string]interface{}); ok {
				if role, ok := spawnResult["role"].(string); ok && role == targetRole {
					requestsTopic, _ := spawnResult["requests_topic"].(string)
					responsesTopic, _ := spawnResult["responses_topic"].(string)
					params.Logger.Info("Found topics for role",
						zap.String("role", targetRole),
						zap.String("requests_topic", requestsTopic),
						zap.String("responses_topic", responsesTopic),
						zap.String("from_step", stepName))
					return requestsTopic, responsesTopic
				}
			}
		}
	}
	return "", ""
}

func findOrSpawnAgent(ctx context.Context, params ActionParams, targetAgentType string) (string, error) {
	params.Logger.Info("call_agent.go findOrSpawnAgent",
		zap.String("agent_type", targetAgentType),
	)

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

func findJobTopicForRole(params ActionParams, targetRole string) string {
	// Search through all spawn results for matching role
	for stepName, stepData := range params.CollectedData {
		if strings.HasPrefix(stepName, "spawn_") {
			if spawnResult, ok := stepData.(map[string]interface{}); ok {
				// Check if this spawn result matches our target role
				if role, ok := spawnResult["role"].(string); ok && role == targetRole {
					if jobTopic, ok := spawnResult["job_topic"].(string); ok && jobTopic != "" {
						params.Logger.Info("Found job topic for role",
							zap.String("role", targetRole),
							zap.String("job_topic", jobTopic),
							zap.String("from_step", stepName))
						return jobTopic
					}
				}
			}
		}
	}
	return ""
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
