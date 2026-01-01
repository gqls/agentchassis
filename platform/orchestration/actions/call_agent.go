package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// CallAgentAction - Main entry point, orchestrates calling an already-spawned agent from the parent (this is on parent)
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Log entry
	logCallStart(params)

	params.Logger.Info("DEBUGaa: CallAgentAction starting - a) prompt b) data for prompt",
		zap.Any("params", params),
	)

	// 1. Extract and validate configuration
	targetAgentType, targetRole, err := extractCallConfiguration(params)
	if err != nil {
		return nil, fmt.Errorf("configuration extraction failed in call agent: %w", err)
	}

	// 2. Find the target agent from spawn results
	targetAgent, err := findTargetAgent(params, targetAgentType, targetRole)
	if err != nil {
		params.Logger.Info("in CallAgentAction failed to find target agent",
			zap.Error(err),
		)
		return nil, err
	}
	params.Logger.Info("in CallAgentAction found target agent",
		zap.Any("targetAgent", targetAgent),
	)

	// 3. Extract the data to send to the agent
	dataToSend := extractDataForAgent(params)
	params.Logger.Info("",
		zap.Any("dataToSend", dataToSend),
	)
	/**
	dataToSend": {
	    "business_name": "Golden Crust Bakery",
	    "business_type": "artisanal bakery"
	  }
	*/

	// 4. Determine the action the agent should perform
	targetAction := determineTargetAction(params.StepConfig)
	params.Logger.Info("in CallAgentAction determined target action",
		zap.String("targetAction", targetAction),
	)

	// 5. Build the complete request body
	requestBody := buildRequestBody(params, targetAction, dataToSend)
	params.Logger.Info("in CallAgentAction built request body",
		zap.Any("requestBody", requestBody),
		zap.Any("DEBUGaa: original params", params),
	)

	// 6. Create child orchestration context
	//childOrchID, childOrchName := createChildOrchestration(targetAgentType)
	agentTypeForOrch := targetAgentType
	if agentTypeForOrch == "" {
		agentTypeForOrch = targetAgent.AgentType
	}
	childOrchID, childOrchName := createChildOrchestration(agentTypeForOrch)

	// 7. Build the request message
	requestMessage := buildCallRequestMessage(
		params,
		targetAgent,
		childOrchID,
		childOrchName,
		targetAction,
		requestBody,
	)
	params.Logger.Info("in CallAgentAction built request message for the child spawn",
		zap.Any("requestMessage", requestMessage),
	)

	// 8. Send the message to the agent
	if err := sendAgentRequest(ctx, params, targetAgent.RequestsTopic, requestMessage); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 9. Build and return the result
	callResult := buildCallResult(targetAgent, childOrchID, targetAction, requestMessage.Headers.RequestID)
	params.Logger.Info("in CallAgentAction built call result on agent",
		zap.Any("callResult", callResult),
		zap.Any("agent type", os.Getenv("AGENT_TYPE")),
	)

	return callResult, nil
}

// Configuration extraction
func extractCallConfiguration(params ActionParams) (targetAgentType string, targetRole string, err error) {
	config := params.StepConfig.Config

	// Check for target_role first - if specified, agent_type is optional
	targetRole, _ = config["target_role"].(string)

	// Try static agent_type
	targetAgentType, _ = config["agent_type"].(string)

	// Try dynamic agent_type_field if static not provided
	if targetAgentType == "" {
		if agentTypeField, ok := config["agent_type_field"].(string); ok && agentTypeField != "" {
			if resolved := resolveFieldPathCallAgent(agentTypeField, params.CollectedData); resolved != "" {
				targetAgentType = resolved
				params.Logger.Info("Resolved agent_type from field",
					zap.String("field", agentTypeField),
					zap.String("resolved", targetAgentType))
			}
		}
	}

	// If we have a target_role, agent_type is optional (we can find the agent by role)
	if targetRole != "" {
		return targetAgentType, targetRole, nil
	}

	// Without target_role, we need agent_type
	if targetAgentType == "" {
		return "", "", fmt.Errorf("agent_type not specified in config (required when target_role is not set)")
	}

	return targetAgentType, targetRole, nil
}

// resolveFieldPathCallAgent extracts a nested string value from a map using dot notation
// e.g., "spawn_builder.agent_type" or "confirmed_type.recommended_builder"
func resolveFieldPathCallAgent(path string, data map[string]interface{}) string {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return ""
			}
			current = val
		default:
			return ""
		}
	}

	if result, ok := current.(string); ok {
		return result
	}
	return ""
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

				// NEW: Also extract agent_type from spawn result
				if agentType, ok := spawnResult["agent_type"].(string); ok {
					agent.AgentType = agentType
				}

				// Extract topics from nested structure
				if topics, ok := spawnResult["topics"].(map[string]interface{}); ok {
					agent.RequestsTopic, _ = topics["requests"].(string)
					agent.ResponsesTopic, _ = topics["responses"].(string)
				}

				// FALLBACK: Try flat structure for backward compatibility
				if agent.RequestsTopic == "" {
					agent.RequestsTopic, _ = spawnResult["requests_topic"].(string)
					agent.ResponsesTopic, _ = spawnResult["responses_topic"].(string)
				}

				params.Logger.Info("Found agent with matching role",
					zap.String("role", targetRole),
					zap.String("agent_id", agent.AgentID),
					zap.String("agent_type", agent.AgentType),
					zap.String("requests_topic", agent.RequestsTopic),
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

		// NEW: Extract from nested topics structure
		if topics, ok := spawnResult["topics"].(map[string]interface{}); ok {
			agent.RequestsTopic, _ = topics["requests"].(string)
			agent.ResponsesTopic, _ = topics["responses"].(string)
		}

		// FALLBACK: Try flat structure
		if agent.RequestsTopic == "" {
			agent.RequestsTopic, _ = spawnResult["requests_topic"].(string)
			agent.ResponsesTopic, _ = spawnResult["responses_topic"].(string)
		}

		return agent.AgentID != ""
	}

	// Search all spawn results
	for stepName, stepData := range params.CollectedData {
		if strings.HasPrefix(stepName, "spawn_") {
			if spawnResult, ok := stepData.(map[string]interface{}); ok {
				if agentType, ok := spawnResult["agent_type"].(string); ok && agentType == targetAgentType {
					agent.AgentID, _ = spawnResult["agent_id"].(string)

					// NEW: Extract from nested topics structure
					if topics, ok := spawnResult["topics"].(map[string]interface{}); ok {
						agent.RequestsTopic, _ = topics["requests"].(string)
						agent.ResponsesTopic, _ = topics["responses"].(string)
					}

					// FALLBACK: Try flat structure
					if agent.RequestsTopic == "" {
						agent.RequestsTopic, _ = spawnResult["requests_topic"].(string)
						agent.ResponsesTopic, _ = spawnResult["responses_topic"].(string)
					}

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
	params.Logger.Info("buildRequestBody data into function",
		zap.String("action", targetAction),
		zap.Any("data_to send", dataToSend),
	)

	requestBody := map[string]interface{}{
		"action":     targetAction,
		"input_data": dataToSend,
	}

	// Add prompt from step config if present
	if prompt, ok := params.StepConfig.Config["prompt"].(string); ok && prompt != "" {
		requestBody["prompt"] = prompt
	}

	// Add any agent config - maybe reinstate this if we want to pass workflows down in messages
	//if agentConfig, ok := params.CollectedData["agent_config"]; ok {
	//	requestBody["config"] = agentConfig
	//}
	// Filter agent_config to remove workflow before passing to child
	// Child agents should use their own workflows, not inherit parent's workflow
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		// Create a clean copy without workflow
		cleanConfig := make(map[string]interface{})
		for k, v := range agentConfig {
			// Skip workflow-related fields that should not be passed to children
			if k != "workflow" && k != "task_workflow" && k != "orchestration_workflow" && k != "orchestrator_workflow" {
				cleanConfig[k] = v
			}
		}

		// Only add config if it has meaningful content after filtering
		if len(cleanConfig) > 0 {
			requestBody["config"] = cleanConfig
			params.Logger.Debug("Added filtered config to request body",
				zap.Int("config_fields", len(cleanConfig)),
				zap.Strings("excluded_fields", []string{"workflow", "task_workflow", "orchestration_workflow"}),
			)
		}
	}

	// Include context from previous steps if requested
	if includeContext, ok := params.StepConfig.Config["include_context"].(bool); ok && includeContext {
		requestBody["context"] = extractContext(params)
	}

	params.Logger.Info("Built request body with clean data",
		zap.String("action", targetAction),
		zap.Any("context", requestBody["context"]),
		zap.Any("data_fields input data", requestBody["input_data"]),
	)

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

	// Use the BuildRequestMessage helper
	message := datahelpers.BuildRequestMessage(
		params.ExecutionContext,
		targetAgent.AgentType,
		targetAction,
		requestBody["input_data"].(map[string]interface{}),
		nil, // config will be added to body separately if needed
		params.Logger,
	)

	// Override specific fields for child orchestration
	message.Headers.OrchestrationID = childOrchID
	message.Headers.OrchestrationName = childOrchName

	// Set parent orchestration info so child knows where to respond
	message.Headers.ParentOrchestrationID = params.ExecutionContext.OrchestrationID
	message.Headers.ParentOrchestrationName = params.ExecutionContext.OrchestrationName

	message.Headers.ToAgent = targetAgent.AgentID
	message.Headers.RequestID = uuid.New().String()
	message.Headers.ReplyToRequestID = message.Headers.RequestID

	parentResponsesTopic := params.ExecutionContext.ResponsesTopic
	if parentResponsesTopic == "" {
		// Fallback to environment variable if not set in context
		parentResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	message.Headers.ParentResponsesTopic = parentResponsesTopic
	message.Headers.ReplyToTopic = parentResponsesTopic
	requestBody["parent_responses_topic"] = parentResponsesTopic

	// Update the body with the complete request body (includes prompt, config, etc.)
	message.Body = requestBody

	return message
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
			ORDER BY version DESC
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

// Data extraction - with explicit template-based specification
func extractDataForAgent(params ActionParams) interface{} {
	params.Logger.Info("extractDataForAgent Extracting data for agent",
		zap.Any("step_config", params.StepConfig.Config))

	// PRIORITY 1: Check for plural "input_fields" (Array of keys)
	if fields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		result := make(map[string]interface{})
		params.Logger.Info("Using input_fields list", zap.Any("fields", fields))

		for _, f := range fields {
			fieldName, ok := f.(string)
			if !ok {
				continue
			}

			// Determine the key to store in result (use base name for nested paths)
			resultKey := fieldName
			if strings.Contains(fieldName, ".") {
				// For "input_data.reviewed_brief", store as "reviewed_brief"
				parts := strings.Split(fieldName, ".")
				resultKey = parts[len(parts)-1]
			}

			// Try 1: Direct lookup in CollectedData (for top-level keys)
			if val, exists := params.CollectedData[fieldName]; exists {
				result[resultKey] = val
				continue
			}

			// Try 2: ExtractNestedField (for dotted paths like "input_data.reviewed_brief")
			if val := datahelpers.ExtractNestedField(params.CollectedData, fieldName); val != nil {
				result[resultKey] = val
				continue
			}

			// Try 3: Search common nested locations
			searchPaths := []string{
				"input_data." + fieldName,
				"site_record.content_data." + fieldName,
				"__raw_message__." + fieldName,
			}

			found := false
			for _, searchPath := range searchPaths {
				if val := datahelpers.ExtractNestedField(params.CollectedData, searchPath); val != nil {
					result[resultKey] = val
					params.Logger.Info("extractDataForAgent: Found field via fallback search",
						zap.String("field", fieldName),
						zap.String("found_at", searchPath),
					)
					found = true
					break
				}
			}

			if !found {
				// Check raw message as last resort
				if raw, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
					if val, exists := raw[fieldName]; exists {
						result[resultKey] = val
						continue
					}
				}

				params.Logger.Warn("extractDataForAgent: Field not found",
					zap.String("field", fieldName),
					zap.Strings("tried_paths", append([]string{fieldName}, searchPaths...)),
				)
			}
		}
		return result
	}

	// PRIORITY 1b: Check for explicit input_data specification in config (map)
	if inputDataSpec, ok := params.StepConfig.Config["input_data"].(map[string]interface{}); ok {
		// ... (Keep existing template rendering logic)
		cleanInputData := datahelpers.GetInputData(params.CollectedData, params.Logger)
		return renderTemplatesInData(inputDataSpec, map[string]interface{}{"input_data": cleanInputData}, params.Logger)
	}

	// PRIORITY 2: Check for input_field reference (single string)
	if inputField, ok := params.StepConfig.Config["input_field"].(string); ok {
		// ... (Keep existing logic)
		cleanInputData := datahelpers.GetInputData(params.CollectedData, params.Logger)
		if fieldData, err := datahelpers.GetFieldFromPath(cleanInputData, inputField, params.Logger); err == nil {
			return fieldData
		}
	}

	// PRIORITY 3: Return the clean extracted data (default)
	params.Logger.Info("Using cleaned input_data default")
	return datahelpers.GetInputData(params.CollectedData, params.Logger)
}

// Render templates in data structure
func renderTemplatesInData(data map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	logger.Info("in renderTemplatesInData callAgentAction",
		zap.Any("data", data),
		zap.Any("collectedData", collectedData),
	)

	result := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Render template strings like "{{.input_data.business_type}}"
			rendered := renderTemplate(v, collectedData, logger)
			result[key] = rendered

		case map[string]interface{}:
			// Recursively render nested maps
			result[key] = renderTemplatesInData(v, collectedData, logger)

		case []interface{}:
			// Handle arrays
			renderedArray := make([]interface{}, len(v))
			for i, item := range v {
				if itemStr, ok := item.(string); ok {
					renderedArray[i] = renderTemplate(itemStr, collectedData, logger)
				} else if itemMap, ok := item.(map[string]interface{}); ok {
					renderedArray[i] = renderTemplatesInData(itemMap, collectedData, logger)
				} else {
					renderedArray[i] = item
				}
			}
			result[key] = renderedArray

		default:
			// Pass through non-template values as-is
			result[key] = value
		}
	}

	return result
}

// Render a single template string
func renderTemplate(template string, collectedData map[string]interface{}, logger *zap.Logger) interface{} {
	logger.Info("in renderTemplate callAgentAction",
		zap.String("template", template),
		zap.Any("collectedData", collectedData),
	)

	// Check if it's a template (contains {{...}})
	if !strings.Contains(template, "{{") {
		return template
	}

	// Parse template syntax: {{.path.to.field}}
	tmplRegex := regexp.MustCompile(`\{\{\.([^}]+)\}\}`)
	matches := tmplRegex.FindStringSubmatch(template)

	if len(matches) < 2 {
		logger.Warn("Invalid template syntax",
			zap.String("template", template))
		return template
	}

	// Extract path (e.g., "input_data.business_type")
	path := matches[1]
	pathParts := strings.Split(path, ".")

	// Navigate to the value
	value, found := getNestedInputValue(collectedData, pathParts...)
	if !found {
		logger.Warn("Template path not found in collected data",
			zap.String("path", path),
			zap.Strings("path_parts", pathParts))
		return ""
	}

	logger.Info("Resolved template",
		zap.String("template", template),
		zap.String("path", path),
		zap.Any("value", value))

	// If the entire string is just the template, return the value directly
	if template == matches[0] {
		return value
	}

	// Otherwise, replace the template in the string
	result := strings.ReplaceAll(template, matches[0], fmt.Sprintf("%v", value))
	return result
}

// Extract by field reference (e.g., "spawn_researcher.result")
func extractByFieldReference(params ActionParams, fieldRef string) interface{} {
	parts := strings.Split(fieldRef, ".")

	value, found := getNestedInputValue(params.CollectedData, parts...)
	if !found {
		params.Logger.Warn("Field reference not found",
			zap.String("field_ref", fieldRef))
		return make(map[string]interface{})
	}

	params.Logger.Info("Extracted data via field reference",
		zap.String("field_ref", fieldRef),
		zap.Any("value", value))

	return value
}

// Extract default input_data (backward compatibility)
func extractDefaultInputData(params ActionParams) interface{} {
	// Try to find clean input_data
	if inputData, ok := params.CollectedData["input_data"]; ok {
		// Check if it's a raw message structure
		if dataMap, ok := inputData.(map[string]interface{}); ok {
			// If it has "body" and "headers", extract from body
			if body, hasBody := dataMap["body"]; hasBody {
				if bodyMap, ok := body.(map[string]interface{}); ok {
					if bodyInputData, ok := bodyMap["input_data"]; ok {
						params.Logger.Info("Extracted clean input_data from message body")
						return bodyInputData
					}
				}
			}
			// If already clean (has business fields), use directly
			if _, hasBizName := dataMap["business_name"]; hasBizName {
				params.Logger.Info("Using direct input_data (already clean)")
				return dataMap
			}
		}
	}

	// Fallback to empty
	params.Logger.Warn("No input_data found, using empty map")
	return make(map[string]interface{})
}

// helper function to platform/orchestration/actions/call_agent.go

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
