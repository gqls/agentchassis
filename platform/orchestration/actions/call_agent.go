package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// CallAgentAction - Main entry point, orchestrates calling an already-spawned agent
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Log entry
	logCallStart(params)

	// 1. Extract and validate configuration
	targetAgentType, targetRole, err := extractCallConfiguration(params)
	if err != nil {
		return nil, fmt.Errorf("configuration extraction failed in call agent: %w", err)
	}

	// 2. Find the target agent from spawn results
	targetAgent, err := findTargetAgent(params, targetAgentType, targetRole)
	if err != nil {
		return nil, err
	}

	// 3. Extract the data to send to the agent
	dataToSend := extractDataForAgent(params)

	// 4. Determine the action the agent should perform
	targetAction := determineTargetAction(params.StepConfig)

	// 5. Build the complete request body
	requestBody := buildRequestBody(params, targetAction, dataToSend)

	// 6. Create child orchestration context
	childOrchID, childOrchName := createChildOrchestration(targetAgentType)

	// 7. Build the request message
	requestMessage := buildCallRequestMessage(
		params,
		targetAgent,
		childOrchID,
		childOrchName,
		targetAction,
		requestBody,
	)

	// 8. Send the message to the agent
	if err := sendAgentRequest(ctx, params, targetAgent.RequestsTopic, requestMessage); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 9. Build and return the result
	return buildCallResult(targetAgent, childOrchID, targetAction, requestMessage.Headers.RequestID), nil
}

// Configuration extraction
func extractCallConfiguration(params ActionParams) (targetAgentType string, targetRole string, err error) {
	config := params.StepConfig.Config

	targetAgentType, ok := config["agent_type"].(string)
	if !ok || targetAgentType == "" {
		return "", "", fmt.Errorf("agent_type not specified in config")
	}

	// Check for specific role (optional)
	targetRole, _ = config["target_role"].(string)

	return targetAgentType, targetRole, nil
}

// TargetAgentInfo holds information about the agent we're calling
type TargetAgentInfo struct {
	AgentID        string
	AgentType      string
	Role           string
	RequestsTopic  string
	ResponsesTopic string
}

// Find the target agent from spawn results
func findTargetAgent(params ActionParams, targetAgentType, targetRole string) (*TargetAgentInfo, error) {
	agent := &TargetAgentInfo{
		AgentType: targetAgentType,
		Role:      targetRole,
	}

	// If looking for a specific role
	if targetRole != "" {
		params.Logger.Info("Looking for agent by role",
			zap.String("target_role", targetRole))

		if found := findAgentByRole(params, targetRole, agent); found {
			return agent, nil
		}
		return nil, fmt.Errorf("no agent with role '%s' found", targetRole)
	}

	// Look by agent type
	if found := findAgentByType(params, targetAgentType, agent); found {
		return agent, nil
	}

	// Check for standard agents with legacy topics
	if isStandardAgent(targetAgentType) {
		agent.RequestsTopic = fmt.Sprintf("system.agent.%s.requests", targetAgentType)
		agent.ResponsesTopic = fmt.Sprintf("system.agent.%s.responses", targetAgentType)
		params.Logger.Info("Using legacy topics for standard agent",
			zap.String("agent_type", targetAgentType))
		return agent, nil
	}

	return nil, fmt.Errorf("no spawned %s agent found", targetAgentType)
}

func findAgentByRole(params ActionParams, targetRole string, agent *TargetAgentInfo) bool {
	// Search through spawn results
	for stepName, stepData := range params.CollectedData {
		if spawnResult, ok := stepData.(map[string]interface{}); ok {
			if role, ok := spawnResult["role"].(string); ok && role == targetRole {
				agent.AgentID, _ = spawnResult["agent_id"].(string)
				agent.RequestsTopic, _ = spawnResult["requests_topic"].(string)
				agent.ResponsesTopic, _ = spawnResult["responses_topic"].(string)

				params.Logger.Info("Found agent with matching role",
					zap.String("role", targetRole),
					zap.String("agent_id", agent.AgentID),
					zap.String("from_step", stepName))
				return true
			}
		}
	}
	return false
}

func findAgentByType(params ActionParams, targetAgentType string, agent *TargetAgentInfo) bool {
	// Look for spawn_<type> key
	spawnKey := fmt.Sprintf("spawn_%s", targetAgentType)
	if spawnResult, ok := params.CollectedData[spawnKey].(map[string]interface{}); ok {
		agent.AgentID, _ = spawnResult["agent_id"].(string)
		agent.RequestsTopic, _ = spawnResult["requests_topic"].(string)
		agent.ResponsesTopic, _ = spawnResult["responses_topic"].(string)
		return agent.AgentID != ""
	}

	// Search all spawn results
	for stepName, stepData := range params.CollectedData {
		if strings.HasPrefix(stepName, "spawn_") {
			if spawnResult, ok := stepData.(map[string]interface{}); ok {
				if agentType, ok := spawnResult["agent_type"].(string); ok && agentType == targetAgentType {
					agent.AgentID, _ = spawnResult["agent_id"].(string)
					agent.RequestsTopic, _ = spawnResult["requests_topic"].(string)
					agent.ResponsesTopic, _ = spawnResult["responses_topic"].(string)
					return agent.AgentID != ""
				}
			}
		}
	}

	return false
}

func isStandardAgent(agentType string) bool {
	standardAgents := []string{"search", "image"}
	for _, standard := range standardAgents {
		if agentType == standard {
			return true
		}
	}
	return false
}

// Data extraction - clearer logic for finding the right data to send
func extractDataForAgent(params ActionParams) interface{} {
	// Get which input field to use
	inputField := "input_data"
	if field, ok := params.StepConfig.Config["input_field"].(string); ok && field != "" {
		inputField = field
	}

	params.Logger.Info("Extracting data for agent",
		zap.String("requested_field", inputField),
		zap.Any("available_keys", getMapKeys(params.CollectedData)),
		zap.Any("DEBUGaa: params.CollectedData for passing to the new agent in CallAgentAction 192", params.CollectedData),
	)

	// Define search paths from most to least specific
	searchPaths := [][]string{
		{"input_data", "input_data", inputField}, // Deeply nested
		{"input_data", inputField},               // Medium nested
		{inputField},                             // Top level
	}

	// Try each path
	for _, path := range searchPaths {
		if data, ok := getNestedInputValue(params.CollectedData, path...); ok {
			params.Logger.Info("Found data via path",
				zap.Strings("path", path))
			return data
		}
	}

	// Special case: if looking for input_data, try nested version
	if inputField == "input_data" {
		if data, ok := getNestedInputValue(params.CollectedData, "input_data", "input_data"); ok {
			params.Logger.Info("Using nested input_data")
			return data
		}
	}

	// Final fallback
	data := params.CollectedData["input_data"]
	if data == nil {
		params.Logger.Warn("No data found, using empty map")
		return make(map[string]interface{})
	}

	params.Logger.Info("Using top-level input_data as fallback")
	return data
}

// Determine what action the agent should perform
func determineTargetAction(stepConfig models.Step) string {
	// Check for explicit target_action
	if action, ok := stepConfig.Config["target_action"].(string); ok && action != "" {
		return action
	}

	// Check for generic action
	if action, ok := stepConfig.Config["action"].(string); ok && action != "" {
		return action
	}

	// Default
	return "process"
}

// Build the complete request body
func buildRequestBody(params ActionParams, targetAction string, dataToSend interface{}) map[string]interface{} {
	requestBody := map[string]interface{}{
		"action":     targetAction,
		"input_data": dataToSend,
	}

	// Add any agent config
	if agentConfig, ok := params.CollectedData["agent_config"]; ok {
		requestBody["config"] = agentConfig
	}

	// Include context from previous steps if requested
	if includeContext, ok := params.StepConfig.Config["include_context"].(bool); ok && includeContext {
		requestBody["context"] = extractContext(params)
	}

	params.Logger.Info("Built request body",
		zap.String("action", targetAction),
		zap.Any("has_data", dataToSend != nil),
		zap.Any("has_config", requestBody["config"] != nil),
		zap.Any("has_context", requestBody["context"] != nil))

	return requestBody
}

func extractContext(params ActionParams) map[string]interface{} {
	localContext := make(map[string]interface{})

	if contextSteps, ok := params.StepConfig.Config["context_steps"].([]interface{}); ok {
		for _, step := range contextSteps {
			if stepName, ok := step.(string); ok {
				if stepData, exists := params.CollectedData[stepName]; exists {
					localContext[stepName] = stepData
				}
			}
		}
	}

	return localContext
}

// Create child orchestration identifiers
func createChildOrchestration(agentType string) (orchestrationID, orchestrationName string) {
	orchestrationID = uuid.New().String()
	orchestrationName = fmt.Sprintf("%s-workflow-%s", agentType, time.Now().Format("1504"))
	return orchestrationID, orchestrationName
}

// Build the request message
func buildCallRequestMessage(
	params ActionParams,
	targetAgent *TargetAgentInfo,
	childOrchID string,
	childOrchName string,
	targetAction string,
	requestBody map[string]interface{},
) *types.RequestMessage {
	requestID := uuid.New().String()

	return &types.RequestMessage{
		Headers: types.RequestHeaders{
			// Sender identity
			Sender: params.ExecutionContext.Sender,

			// Core identity
			CorrelationID:   params.ExecutionContext.CorrelationID,
			CorrelationName: params.ExecutionContext.CorrelationName,
			ClientID:        params.ExecutionContext.ClientID,

			// Child orchestration context
			OrchestrationID:   childOrchID,
			OrchestrationName: childOrchName,
			StepID:            uuid.New().String(),
			StepName:          "process",
			RequestID:         requestID,
			RetryVersion:      0,

			// Parent tracking
			ParentOrchestrationID:   params.ExecutionContext.OrchestrationID,
			ParentOrchestrationName: params.ExecutionContext.OrchestrationName,
			ParentRequestID:         params.ExecutionContext.RequestID,

			// Message metadata
			MessageID:   uuid.New().String(),
			MessageType: "request",
			FromAgent:   params.ExecutionContext.Sender.AgentID,
			ToAgent:     targetAgent.AgentID,
			ToAgentType: targetAgent.AgentType,
			Action:      targetAction,
			Timestamp:   time.Now(),

			// Resource management
			FuelBudget:     params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds: 30,

			// Routing - child needs to know where to respond
			ResponsesTopic: params.ExecutionContext.ResponsesTopic,
			RequestsTopic:  "", // Child sets its own
		},
		Body: requestBody,
	}
}

// Send the request to the agent
func sendAgentRequest(ctx context.Context, params ActionParams, targetTopic string, message *types.RequestMessage) error {
	// Marshal the message
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Convert headers
	headers := message.Headers.ToMap()
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("Sending request to agent",
		zap.String("target_topic", targetTopic),
		zap.String("request_id", message.Headers.RequestID),
		zap.String("orchestration_id", message.Headers.OrchestrationID),
		zap.String("action", message.Headers.Action))

	// Send the message
	if err := params.Producer.Produce(ctx, targetTopic, headers, key, msgBytes); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Trace if available
	if params.Tracer != nil {
		params.Tracer.TraceMessage(params.ExecutionContext, "SEND_CHILD_REQUEST", targetTopic,
			map[string]interface{}{
				"request_id":      message.Headers.RequestID,
				"child_orch_id":   message.Headers.OrchestrationID,
				"parent_orch_id":  params.ExecutionContext.OrchestrationID,
				"responses_topic": message.Headers.ResponsesTopic,
			})
	}

	return nil
}

// Build the result to return
func buildCallResult(targetAgent *TargetAgentInfo, childOrchID, targetAction, requestID string) map[string]interface{} {
	return map[string]interface{}{
		"agent_called":          targetAgent.AgentID,
		"agent_type":            targetAgent.AgentType,
		"request_id":            requestID,
		"child_orchestration":   childOrchID,
		"action_sent":           targetAction,
		"await_response":        true,
		"target_agent_type":     targetAgent.AgentType,
		"child_responses_topic": targetAgent.ResponsesTopic, // For debugging
	}
}

// Helper functions
func logCallStart(params ActionParams) {
	current, caller := getFuncInfo(1)
	params.Logger.Info("CallAgentAction starting",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.Any("config", params.StepConfig),
		zap.Any("headers", params.Headers))
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

// Add this helper function to platform/orchestration/actions/call_agent.go

// getNestedValue safely traverses a map[string]interface{} to find a value.
func getNestedInputValue(data map[string]interface{}, path ...string) (interface{}, bool) {
	var current interface{} = data
	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
