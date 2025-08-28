// platform/orchestration/actions/orchestration_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// StartOrchestrationAction spawns a child orchestration and waits for its completion
func StartOrchestrationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Starting orchestration action",
		zap.String("current_step", params.CurrentStep))

	// Step 1: Validate we have our own orchestration context
	parentContext, err := extractParentContext(params)
	if err != nil {
		return nil, err
	}

	// Step 2: Check if already started (idempotency)
	if existing := checkExistingChild(params); existing != nil {
		params.Logger.Info("Returning existing child orchestration")
		return existing, nil
	}

	// Step 3: Find the spawn data from previous step
	spawnData, err := findSpawnData(params)
	if err != nil {
		return nil, err
	}

	// Step 4: Extract and prepare workflow
	workflowJSON, err := extractWorkflow(spawnData, params.Logger)
	if err != nil {
		return nil, err
	}

	// Step 5: Create child orchestration
	childIDs, err := createChildOrchestration(ctx, params, parentContext, spawnData, workflowJSON)
	if err != nil {
		return nil, err
	}

	// Step 6: Start timeout monitor
	startTimeoutMonitor(params, parentContext, childIDs)

	// Step 7: Return result for parent to wait on
	return buildStartOrchestrationResult(childIDs, spawnData), nil
}

// ParentContext holds the parent orchestration's context
type ParentContext struct {
	OrchestrationID string
	CorrelationID   string
	AgentType       string
}

// extractParentContext gets the current (parent) orchestration's context
func extractParentContext(params ActionParams) (*ParentContext, error) {
	orchestrationID := params.Headers["orchestration_id"]
	if orchestrationID == "" {
		return nil, fmt.Errorf("current orchestration_id not found - cannot start child without parent context")
	}

	return &ParentContext{
		OrchestrationID: orchestrationID,
		CorrelationID:   params.Headers["correlation_id"],
		AgentType:       params.Headers["agent_type"],
	}, nil
}

// checkExistingChild checks if this step already created a child
func checkExistingChild(params ActionParams) interface{} {
	stepKey := fmt.Sprintf("%s_started", params.CurrentStep)

	if _, alreadyStarted := params.CollectedData[stepKey]; !alreadyStarted {
		return nil
	}

	// Try to return the existing result
	if existingChild, ok := params.CollectedData[params.CurrentStep]; ok {
		if childMap, ok := existingChild.(map[string]interface{}); ok {
			if childID, ok := childMap["child_orchestration_id"].(string); ok && childID != "" {
				params.Logger.Info("Found existing child",
					zap.String("child_orchestration_id", childID))
				return existingChild
			}
		}
	}

	params.Logger.Warn("Child marked as started but can't find result")
	return nil
}

// findSpawnData locates the spawn result from previous steps
func findSpawnData(params ActionParams) (map[string]interface{}, error) {
	var spawnResult map[string]interface{}
	found := false

	// Look for spawn results in collected data
	for stepName, data := range params.CollectedData {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Check for spawn result indicators
			hasWorkflow := dataMap["workflow"] != nil
			hasAgents := dataMap["agents"] != nil
			hasGroupID := dataMap["group_id"] != nil

			if (hasWorkflow && hasAgents) || (hasGroupID && hasAgents) {
				spawnResult = dataMap
				found = true
				params.Logger.Debug("Found spawn result",
					zap.String("from_step", stepName))
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no spawn result found in collected data")
	}

	return spawnResult, nil
}

// extractWorkflow gets the workflow JSON from spawn data
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
	case map[string]interface{}:
		bytes, err := json.Marshal(wf)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal workflow: %w", err)
		}
		return json.RawMessage(bytes), nil
	default:
		return nil, fmt.Errorf("workflow has unexpected type: %T", workflow)
	}
}

// ChildOrchestrationIDs holds all IDs for the child orchestration
type ChildOrchestrationIDs struct {
	OrchestrationID string // Child's orchestration ID
	MessageID       string // Message ID for the creation request
	RequestID       string // Request ID that parent will wait for
}

// createChildOrchestration creates the child orchestration
func createChildOrchestration(
	ctx context.Context,
	params ActionParams,
	parent *ParentContext,
	spawnData map[string]interface{},
	workflowJSON json.RawMessage,
) (*ChildOrchestrationIDs, error) {

	// Generate IDs for the child
	childIDs := &ChildOrchestrationIDs{
		OrchestrationID: uuid.New().String(),
		MessageID:       uuid.New().String(),
		RequestID:       uuid.New().String(), // This is what parent will wait for
	}

	// MY response topic
	myResponseTopic := fmt.Sprintf("system.agent.%s.responses", parent.AgentType)

	// Store the request ID for the child to know what to respond to
	// params.CollectedData["parent_request_id"] = childIDs.RequestID

	// Build headers for child
	childHeaders := make(map[string]string)
	for k, v := range params.Headers {
		childHeaders[k] = v
	}

	// Set orchestration hierarchy and parent context
	childHeaders["orchestration_id"] = childIDs.OrchestrationID
	childHeaders["parent_orchestration_id"] = parent.OrchestrationID
	childHeaders["parent_reply_to_topic"] = myResponseTopic // Where child responds
	childHeaders["parent_agent_type"] = parent.AgentType
	childHeaders["correlation_id"] = parent.CorrelationID
	childHeaders["message_id"] = childIDs.MessageID
	childHeaders["request_id"] = childIDs.RequestID

	// Add agent mappings from spawn data
	if agents, ok := spawnData["agents"].(map[string]interface{}); ok {
		for role, agentID := range agents {
			if id, ok := agentID.(string); ok {
				childHeaders[fmt.Sprintf("agent_%s", role)] = id
			}
		}
	}

	// Mark this step as started
	stepKey := fmt.Sprintf("%s_started", params.CurrentStep)
	params.CollectedData[stepKey] = true

	// Get orchestrator interface
	orchestrator, ok := params.SagaCoordinator.(interface {
		CreateNewOrchestration(context.Context, string, map[string]string, json.RawMessage) error
	})
	if !ok || orchestrator == nil {
		return nil, fmt.Errorf("SagaCoordinator not available")
	}

	// Create the orchestration
	err := orchestrator.CreateNewOrchestration(ctx, childIDs.OrchestrationID, childHeaders, workflowJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestration: %w", err)
	}

	params.Logger.Info("Child orchestration created",
		zap.String("child_orchestration_id", childIDs.OrchestrationID),
		zap.String("parent_will_wait_for", childIDs.RequestID))

	return childIDs, nil
}

// buildChildHeaders creates headers for the child orchestration
func buildChildHeaders(
	parentHeaders map[string]string,
	parent *ParentContext,
	childIDs *ChildOrchestrationIDs,
	spawnData map[string]interface{},
) map[string]string {
	headers := make(map[string]string)

	// Copy parent headers
	for k, v := range parentHeaders {
		headers[k] = v
	}

	// Set orchestration hierarchy
	headers["orchestration_id"] = childIDs.OrchestrationID
	headers["parent_orchestration_id"] = parent.OrchestrationID
	headers["correlation_id"] = parent.CorrelationID
	headers["message_id"] = childIDs.MessageID
	headers["request_id"] = childIDs.RequestID
	headers["parent_agent_type"] = parent.AgentType

	// Add agent mappings from spawn data
	if agents, ok := spawnData["agents"].(map[string]interface{}); ok {
		for role, agentID := range agents {
			if id, ok := agentID.(string); ok {
				headers[fmt.Sprintf("agent_%s", role)] = id
			}
		}
	}

	return headers
}

// startTimeoutMonitor starts a goroutine to monitor child timeout
func startTimeoutMonitor(params ActionParams, parent *ParentContext, childIDs *ChildOrchestrationIDs) {
	timeout := 5 * time.Minute
	if configTimeout, ok := params.StepConfig.Config["child_timeout_minutes"].(float64); ok {
		timeout = time.Duration(configTimeout) * time.Minute
	}

	go monitorChildTimeout(
		params.DB,
		params.Producer,
		params.Logger,
		parent,
		childIDs,
		timeout,
	)
}

// monitorChildTimeout monitors for child orchestration timeout
func monitorChildTimeout(
	db *sql.DB,
	producer interface{},
	logger *zap.Logger,
	parent *ParentContext,
	childIDs *ChildOrchestrationIDs,
	timeout time.Duration,
) {
	time.Sleep(timeout)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if parent is still waiting for this request
	if !isParentStillWaiting(ctx, db, parent.OrchestrationID, childIDs.RequestID, logger) {
		logger.Info("Timeout check passed - parent no longer waiting",
			zap.String("child_request_id", childIDs.RequestID))
		return
	}

	// Send timeout notification
	// sendTimeoutNotification(ctx, producer, parent, childIDs, timeout, logger)
}

// isParentStillWaiting checks if parent is still waiting for a request
func isParentStillWaiting(ctx context.Context, db *sql.DB, parentOrchID, requestID string, logger *zap.Logger) bool {
	if db == nil {
		return false
	}

	var status string
	var awaitedStepsJSON []byte

	query := `SELECT status, awaited_steps FROM orchestration_states WHERE orchestration_id = $1`
	err := db.QueryRowContext(ctx, query, parentOrchID).Scan(&status, &awaitedStepsJSON)
	if err != nil {
		logger.Error("Failed to check parent state", zap.Error(err))
		return false
	}

	if status != "AWAITING_RESPONSES" {
		return false
	}

	var awaitedSteps []string
	if err := json.Unmarshal(awaitedStepsJSON, &awaitedSteps); err != nil {
		return false
	}

	for _, step := range awaitedSteps {
		if step == requestID {
			return true
		}
	}

	return false
}

// buildStartOrchestrationResult builds the return value
// the data that gets stored in the parent's CollectedData - stored locally in parent's state
/*
This return value tells the parent orchestration:
"I started a child orchestration"
"The request_id is REQ-123"
"You should wait for a response to REQ-123"
Then the parent does: state.AwaitedSteps = []string{"REQ-123"}  // From the request_id field above
Later, when the child COMPLETES and sends a response MESSAGE (not the return value), it includes:
responseHeaders := map[string]string{
    "in_response_to": "REQ-123",  // This matches what parent is waiting for
    // ...
}
The in_response_to field belongs in the response MESSAGE headers, not in the action's return value. The action return value just needs to tell the parent what request_id to wait for.
*/
func buildStartOrchestrationResult(childIDs *ChildOrchestrationIDs, spawnData map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":                 "orchestration_started",
		"child_orchestration_id": childIDs.OrchestrationID,
		"request_id":             childIDs.RequestID,
		"await_response":         true,
		"group_id":               spawnData["group_id"],
	}
}

func getMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
