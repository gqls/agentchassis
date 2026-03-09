package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/input_contracts"
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

	// 3. Extract the data to send to the agent (with contract validation)
	dataToSend, err := extractDataForAgent(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to extract data for agent: %w", err)
	}
	params.Logger.Info("",
		zap.Any("dataToSend", dataToSend),
	)

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

	// 10. Check if caller configured fire-and-forget (no waiting for child response).
	//     buildCallResult sets await_response: true by default.
	//     Only override to false when config explicitly requests it.
	//     The coordinator's processAwaitResponse reads this from the action result
	//     to decide whether to enter AWAITING_RESPONSES state.
	config := params.StepConfig.Config
	if awaitCfg, ok := config["await_response"].(bool); ok {
		if !awaitCfg {
			callResult["await_response"] = false
			params.Logger.Info("CallAgentAction: fire-and-forget mode — not awaiting child response",
				zap.String("target_agent", targetAgent.AgentType),
				zap.String("target_agent_id", targetAgent.AgentID),
				zap.String("request_id", requestMessage.Headers.RequestID),
				zap.String("child_orchestration", childOrchID),
				zap.String("child_requests_topic", targetAgent.RequestsTopic),
			)
		} else {
			params.Logger.Info("CallAgentAction: await_response explicitly true (default behaviour)",
				zap.String("target_agent", targetAgent.AgentType),
				zap.String("request_id", requestMessage.Headers.RequestID),
			)
		}
	} else {
		// No await_response in config — keep buildCallResult default (true)
		params.Logger.Info("CallAgentAction: awaiting child response (default)",
			zap.String("target_agent", targetAgent.AgentType),
			zap.String("request_id", requestMessage.Headers.RequestID),
		)
	}

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
// It also handles the .response wrapper pattern automatically
func resolveFieldPathCallAgent(path string, data map[string]interface{}) string {
	// First, try the direct path
	if value := traverseFieldPathGeneric(path, data); value != "" {
		return value
	}

	// If direct path failed and path has at least 2 parts, try with .response wrapper
	parts := strings.Split(path, ".")
	if len(parts) >= 2 {
		responsePath := parts[0] + ".response." + strings.Join(parts[1:], ".")
		if value := traverseFieldPathGeneric(responsePath, data); value != "" {
			return value
		}
	}

	return ""
}

// traverseFieldPathGeneric traverses a dot-notation path and returns string value
func traverseFieldPathGeneric(path string, data map[string]interface{}) string {
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

// REASON: In loop iterations, multiple spawned agents share the same role (e.g. "handler").
//
//	Go map iteration is non-deterministic, so findAgentByRole randomly returns the wrong
//	iteration's agent. Fix: when called from a loop step, prefer the key matching the
//	current iteration index.
func findAgentByRole(params ActionParams, targetRole string, agent *TargetAgentInfo) bool {
	// Check if we're inside a loop iteration — if so, prefer the current iteration's spawn result
	if loopIter := getLoopIteration(params); loopIter >= 0 {
		iterSuffix := fmt.Sprintf("_%d", loopIter)
		for stepName, stepData := range params.CollectedData {
			if !strings.HasSuffix(stepName, iterSuffix) {
				continue
			}
			if stepResult, ok := stepData.(map[string]interface{}); ok {
				spawnResult := unwrapSpawnResult(stepResult)
				if role, ok := spawnResult["role"].(string); ok && role == targetRole {
					populateAgentFromSpawnResult(agent, spawnResult)
					params.Logger.Info("Found agent with matching role (iteration-aware)",
						zap.String("role", targetRole),
						zap.String("agent_id", agent.AgentID),
						zap.String("agent_type", agent.AgentType),
						zap.String("from_step", stepName),
						zap.Int("loop_iteration", loopIter))
					return true
				}
			}
		}
		// If not found via iteration-specific key, fall through to general scan.
		// This handles cases where the spawn output_field doesn't follow the _{N} convention.
		params.Logger.Debug("No iteration-specific agent found, falling back to general scan",
			zap.String("role", targetRole),
			zap.Int("loop_iteration", loopIter))
	}

	// General scan — original behaviour (used outside loops, or as fallback)
	for stepName, stepData := range params.CollectedData {
		if stepResult, ok := stepData.(map[string]interface{}); ok {
			spawnResult := unwrapSpawnResult(stepResult)
			if role, ok := spawnResult["role"].(string); ok && role == targetRole {
				populateAgentFromSpawnResult(agent, spawnResult)
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

// getLoopIteration extracts the loop iteration index from the step config.
// Returns -1 if not in a loop context.
func getLoopIteration(params ActionParams) int {
	if params.StepConfig.Config == nil {
		return -1
	}
	iterVal, ok := params.StepConfig.Config["loop_iteration"]
	if !ok {
		return -1
	}
	switch v := iterVal.(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return -1
}

// populateAgentFromSpawnResult fills a TargetAgentInfo from spawn result data.
// Extracted from the old inline code in findAgentByRole to avoid duplication
// between the iteration-aware path and the general scan path.
func populateAgentFromSpawnResult(agent *TargetAgentInfo, spawnResult map[string]interface{}) {
	agent.AgentID, _ = spawnResult["agent_id"].(string)

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
}

// findAgentDataWithRole checks for agent data with matching role at top level
// or inside .response wrapper. Returns the map containing agent fields, or nil.
func findAgentDataWithRole(data map[string]interface{}, targetRole string) map[string]interface{} {
	// First, check top level
	if role, ok := data["role"].(string); ok && role == targetRole {
		return data
	}

	// Then check inside .response wrapper
	if response, ok := data["response"].(map[string]interface{}); ok {
		if role, ok := response["role"].(string); ok && role == targetRole {
			return response
		}
	}

	return nil
}

func findAgentByType(params ActionParams, targetAgentType string, agent *TargetAgentInfo) bool {
	// Look for spawn_<type> key
	spawnKey := fmt.Sprintf("spawn_%s", targetAgentType)
	if stepResult, ok := params.CollectedData[spawnKey].(map[string]interface{}); ok {
		// Unwrap .response if present
		spawnResult := unwrapSpawnResult(stepResult)

		agent.AgentID, _ = spawnResult["agent_id"].(string)

		// Extract from nested topics structure
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
			if stepResult, ok := stepData.(map[string]interface{}); ok {
				// Unwrap .response if present
				spawnResult := unwrapSpawnResult(stepResult)

				if agentType, ok := spawnResult["agent_type"].(string); ok && agentType == targetAgentType {
					agent.AgentID, _ = spawnResult["agent_id"].(string)

					// Extract from nested topics structure
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

// unwrapSpawnResult extracts the actual spawn data from a step result.
// Spawn results may be stored directly OR wrapped in a .response field.
// Returns the map containing agent fields (role, agent_id, topics, etc.)
func unwrapSpawnResult(stepData map[string]interface{}) map[string]interface{} {
	// Check if there's a .response wrapper with agent data inside
	if response, ok := stepData["response"].(map[string]interface{}); ok {
		// Verify this looks like spawn data (has agent_id or role)
		if _, hasAgentID := response["agent_id"]; hasAgentID {
			return response
		}
		if _, hasRole := response["role"]; hasRole {
			return response
		}
	}

	// No wrapper or wrapper doesn't contain agent data - use top level
	return stepData
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
		"requests_topic":        targetAgent.RequestsTopic,  // for retry routing
		"responses_topic":       targetAgent.ResponsesTopic, // for consistency
		"child_responses_topic": targetAgent.ResponsesTopic, // For debugging
		"await_response":        true,
		"target_agent_type":     targetAgent.AgentType,
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
			if stepResult, ok := stepData.(map[string]interface{}); ok {
				// Unwrap .response if present
				spawnResult := unwrapSpawnResult(stepResult)

				if role, ok := spawnResult["role"].(string); ok && role == targetRole {
					// Try nested topics first
					if topics, ok := spawnResult["topics"].(map[string]interface{}); ok {
						requestsTopic, _ := topics["requests"].(string)
						responsesTopic, _ := topics["responses"].(string)
						if requestsTopic != "" {
							params.Logger.Info("Found topics for role",
								zap.String("role", targetRole),
								zap.String("requests_topic", requestsTopic),
								zap.String("responses_topic", responsesTopic),
								zap.String("from_step", stepName))
							return requestsTopic, responsesTopic
						}
					}

					// Fallback to flat structure
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
	if stepResult, ok := params.CollectedData["spawn_agent"].(map[string]interface{}); ok {
		spawnResult := unwrapSpawnResult(stepResult)
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
			if stepResult, ok := stepData.(map[string]interface{}); ok {
				// Unwrap .response if present
				spawnResult := unwrapSpawnResult(stepResult)

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

// Data extraction - with explicit input_mapping support and contract validation
func extractDataForAgent(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	stepName := params.StepConfig.Name

	// Extract targetAgentType early so it's available in all branches
	targetAgentType, _ := config["agent_type"].(string)

	params.Logger.Info("extractDataForAgent Extracting data for agent",
		zap.String("step", stepName),
		zap.String("agent_type", targetAgentType),
		zap.Any("step_config", config))

	// PRIORITY 1: Check for new explicit input_mapping (PREFERRED)
	if inputMapping, ok := input_contracts.ParseInputMapping(config); ok {
		params.Logger.Info("Using explicit input_mapping",
			zap.String("step", stepName),
			zap.Int("mapping_count", len(inputMapping)))

		// Resolve the mapping to actual data
		inputData, err := input_contracts.ResolveInputMapping(
			params.CollectedData,
			inputMapping,
			params.Logger,
		)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", stepName, err)
		}

		// Validate against child agent's input contract
		if targetAgentType != "" {
			contract, err := input_contracts.GetAgentInputContract(ctx, params.DB, targetAgentType, params.Logger)
			if err != nil {
				params.Logger.Warn("Failed to load input contract, skipping validation",
					zap.String("agent_type", targetAgentType),
					zap.Error(err))
			} else if contract != nil {
				if err := input_contracts.ValidateInputContract(targetAgentType, inputData, contract, params.Logger); err != nil {
					return nil, fmt.Errorf("step %s: %w", stepName, err)
				}
			}
		}

		return inputData, nil
	}

	// PRIORITY 1a: Check for plural "input_fields" (Array of keys) - DEPRECATED
	if fields, ok := config["input_fields"].([]interface{}); ok {
		result := make(map[string]interface{})
		params.Logger.Info("Using input_fields list", zap.Any("fields", fields))

		params.Logger.Warn("DEPRECATED: Using input_fields instead of input_mapping",
			zap.String("step", stepName),
			zap.String("agent_type", targetAgentType),
			zap.String("hint", "Update workflow config to use input_mapping"))

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
		return result, nil
	}

	// PRIORITY 1b: Check for explicit input_data specification in config (map)
	if inputDataSpec, ok := config["input_data"].(map[string]interface{}); ok {
		cleanInputData := datahelpers.GetInputData(params.CollectedData, params.Logger)

		params.Logger.Warn("DEPRECATED: Using input_data map instead of input_mapping",
			zap.String("step", stepName),
			zap.String("agent_type", targetAgentType),
			zap.String("hint", "Update workflow config to use input_mapping"))

		return renderTemplatesInData(inputDataSpec, map[string]interface{}{"input_data": cleanInputData}, params.Logger), nil
	}

	// PRIORITY 2: Check for input_field reference (single string)
	if inputField, ok := config["input_field"].(string); ok {
		cleanInputData := datahelpers.GetInputData(params.CollectedData, params.Logger)

		params.Logger.Warn("DEPRECATED: Using input_field instead of input_mapping",
			zap.String("step", stepName),
			zap.String("agent_type", targetAgentType),
			zap.String("hint", "Update workflow config to use input_mapping"))

		if fieldData, err := datahelpers.GetFieldFromPath(cleanInputData, inputField, params.Logger); err == nil {
			return fieldData, nil
		}
	}

	// PRIORITY 3: Return the clean extracted data (default)
	params.Logger.Info("Using cleaned input_data default")
	return datahelpers.GetInputData(params.CollectedData, params.Logger), nil
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

// executeGoTemplate executes a template using Go's text/template package
// This properly handles {{if}}, {{range}}, {{with}}, {{end}} directives
func executeGoTemplate(templateStr string, data map[string]interface{}, logger *zap.Logger) (string, error) {
	tmpl, err := template.New("component").
		Option("missingkey=zero"). // Still useful for some types
		Funcs(template.FuncMap{
			"default": func(defaultVal, val interface{}) interface{} {
				if val == nil || val == "" {
					return defaultVal
				}
				return val
			},
			"eq": func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
			"ne": func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
			},
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
			"isset": func(val interface{}) bool {
				if val == nil {
					return false
				}
				if s, ok := val.(string); ok {
					return s != ""
				}
				return true
			},
			// safe returns "" for nil values instead of "<no value>"
			// Usage in templates: {{safe .name}} or {{.name | safe}}
			"safe": func(val interface{}) string {
				if val == nil {
					return ""
				}
				s := fmt.Sprintf("%v", val)
				if s == "<nil>" || s == "<no value>" {
					return ""
				}
				return s
			},
		}).Parse(templateStr)

	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}

	return buf.String(), nil
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
