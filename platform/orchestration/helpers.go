// FILE: platform/orchestration/helpers.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TimeoutMonitor handles timeout monitoring for child orchestrations and requests
type TimeoutMonitor struct {
	db       *sql.DB
	producer kafka.Producer
	logger   *zap.Logger
	repo     *StateRepository
}

// NewTimeoutMonitor creates a new timeout monitor
func NewTimeoutMonitor(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *TimeoutMonitor {
	return &TimeoutMonitor{
		db:       db,
		producer: producer,
		logger:   logger,
		repo:     NewStateRepository(db, logger),
	}
}

// MonitorRequest monitors a specific request for timeout
func (tm *TimeoutMonitor) MonitorRequest(orchestrationID string, requestID string, timeout time.Duration) {
	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
		defer ticker.Stop()

		for {
			select {
			case <-timer.C:
				// Timeout reached
				tm.handleRequestTimeout(orchestrationID, requestID)
				return

			case <-ticker.C:
				// Check if still waiting
				if !tm.isStillWaitingForRequest(orchestrationID, requestID) {
					tm.logger.Debug("Request completed, stopping timeout monitor",
						zap.String("orchestration_id", orchestrationID),
						zap.String("request_id", requestID))
					return
				}
			}
		}
	}()
}

// MonitorChildOrchestration monitors a child orchestration for timeout
func (tm *TimeoutMonitor) MonitorChildOrchestration(
	parentOrchID string,
	childOrchID string,
	requestID string,
	timeout time.Duration,
) {
	go func() {
		tm.logger.Info("Starting child orchestration timeout monitor",
			zap.String("parent_orchestration_id", parentOrchID),
			zap.String("child_orchestration_id", childOrchID),
			zap.String("request_id", requestID),
			zap.Duration("timeout", timeout))

		timer := time.NewTimer(timeout)
		defer timer.Stop()

		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-timer.C:
				// Timeout reached
				tm.handleChildTimeout(parentOrchID, childOrchID, requestID)
				return

			case <-ticker.C:
				// Check both parent and child states
				if !tm.isParentStillWaiting(parentOrchID, requestID) {
					tm.logger.Info("Parent no longer waiting, stopping child monitor",
						zap.String("child_orchestration_id", childOrchID))
					return
				}

				if tm.isChildCompleted(childOrchID) {
					tm.logger.Info("Child completed, stopping monitor",
						zap.String("child_orchestration_id", childOrchID))
					return
				}
			}
		}
	}()
}

// isStillWaitingForRequest checks if orchestration is still waiting for a request
func (tm *TimeoutMonitor) isStillWaitingForRequest(orchestrationID, requestID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := tm.repo.GetState(ctx, orchestrationID)
	if err != nil {
		tm.logger.Error("Failed to get orchestration state",
			zap.Error(err),
			zap.String("orchestration_id", orchestrationID))
		return false
	}

	// Check new request-based tracking
	if _, exists := state.AwaitedRequests[requestID]; exists {
		return true
	}

	// Fallback to legacy step-based tracking
	for _, step := range state.AwaitedSteps {
		if step == requestID {
			return true
		}
	}

	return false
}

// isParentStillWaiting checks if parent is still waiting for a request
func (tm *TimeoutMonitor) isParentStillWaiting(parentOrchID, requestID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := tm.repo.GetState(ctx, parentOrchID)
	if err != nil {
		return false
	}

	if state.Status != StatusAwaitingResponses {
		return false
	}

	// Check request-based tracking
	if _, exists := state.AwaitedRequests[requestID]; exists {
		return true
	}

	// Check legacy step-based tracking
	for _, step := range state.AwaitedSteps {
		if step == requestID {
			return true
		}
	}

	return false
}

// isChildCompleted checks if a child orchestration has completed
func (tm *TimeoutMonitor) isChildCompleted(childOrchID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := tm.repo.GetState(ctx, childOrchID)
	if err != nil {
		return true // Assume completed if we can't find it
	}

	return state.Status == StatusCompleted || state.Status == StatusFailed
}

// handleRequestTimeout handles a request timeout
func (tm *TimeoutMonitor) handleRequestTimeout(orchestrationID, requestID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tm.logger.Warn("Request timed out",
		zap.String("orchestration_id", orchestrationID),
		zap.String("request_id", requestID))

	state, err := tm.repo.GetState(ctx, orchestrationID)
	if err != nil {
		tm.logger.Error("Failed to get state for timeout handling", zap.Error(err))
		return
	}

	// Check if still waiting
	awaited, exists := state.AwaitedRequests[requestID]
	if !exists {
		tm.logger.Info("Request no longer awaited, ignoring timeout")
		return
	}

	// Check retry count
	if awaited.RetryVersion < 3 {
		// Retry the request
		tm.retryTimedOutRequest(ctx, state, awaited)
	} else {
		// Max retries exceeded, fail the orchestration
		tm.failOrchestrationDueToTimeout(ctx, state, requestID, "Request timed out after max retries")
	}
}

// handleChildTimeout handles a child orchestration timeout
func (tm *TimeoutMonitor) handleChildTimeout(parentOrchID, childOrchID, requestID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tm.logger.Error("Child orchestration timed out",
		zap.String("parent_orchestration_id", parentOrchID),
		zap.String("child_orchestration_id", childOrchID),
		zap.String("request_id", requestID))

	// Send timeout response to parent
	tm.sendTimeoutResponse(ctx, parentOrchID, childOrchID, requestID)

	// Fail the child orchestration
	childState, err := tm.repo.GetState(ctx, childOrchID)
	if err == nil && childState.Status != StatusCompleted && childState.Status != StatusFailed {
		childState.Status = StatusFailed
		childState.Error = "Orchestration timed out"
		tm.repo.UpdateState(ctx, childState)
	}
}

// retryTimedOutRequest retries a timed-out request
func (tm *TimeoutMonitor) retryTimedOutRequest(ctx context.Context, state *OrchestrationState, awaited *AwaitedRequest) {
	awaited.RetryVersion++
	awaited.SentAt = time.Now()
	awaited.TimeoutAt = time.Now().Add(30 * time.Second)

	tm.logger.Info("Retrying timed-out request",
		zap.String("request_id", awaited.RequestID),
		zap.Int("retry_version", awaited.RetryVersion))

	// Create retry message
	retryRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			RequestID:       awaited.RequestID, // Same request ID
			RetryVersion:    awaited.RetryVersion,
			StepID:          awaited.StepID,
			StepName:        awaited.StepName,
			OrchestrationID: state.OrchestrationID,
			CorrelationID:   state.CorrelationID,
			ToAgentType:     awaited.TargetAgentType,
			MessageID:       uuid.New().String(),
			MessageType:     "request",
			Timestamp:       time.Now(),
			Action:          "retry",
		},
	}

	// Send retry
	topic := fmt.Sprintf("system.agent.%s.requests", awaited.TargetAgentType)
	retryBytes, _ := json.Marshal(retryRequest)

	if err := tm.producer.Produce(ctx, topic, retryRequest.Headers.ToMap(), []byte(awaited.RequestID), retryBytes); err != nil {
		tm.logger.Error("Failed to send retry request",
			zap.Error(err),
			zap.String("request_id", awaited.RequestID))

		// If we can't retry, fail the orchestration
		tm.failOrchestrationDueToTimeout(ctx, state, awaited.RequestID, "Failed to send retry request")
	} else {
		// Update the awaited request in state
		state.AwaitedRequests[awaited.RequestID] = awaited
		tm.repo.UpdateState(ctx, state)

		// Start new timeout monitor
		tm.MonitorRequest(state.OrchestrationID, awaited.RequestID, 30*time.Second)
	}
}

// failOrchestrationDueToTimeout fails an orchestration due to timeout
func (tm *TimeoutMonitor) failOrchestrationDueToTimeout(ctx context.Context, state *OrchestrationState, requestID, reason string) {
	tm.logger.Error("Failing orchestration due to timeout",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("request_id", requestID),
		zap.String("reason", reason))

	state.Status = StatusFailed
	state.Error = fmt.Sprintf("Timeout: %s (request_id: %s)", reason, requestID)

	// Add failure to processing history
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   tm.getPodName(),
		Action:    "timeout_failure",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Request %s: %s", requestID, reason),
	})

	// Remove from awaited requests
	delete(state.AwaitedRequests, requestID)

	// Update state
	if err := tm.repo.UpdateState(ctx, state); err != nil {
		tm.logger.Error("Failed to update state after timeout", zap.Error(err))
	}

	// Send failure notification if there's a parent
	if state.ParentOrchestrationID != "" {
		tm.sendTimeoutResponse(ctx, state.ParentOrchestrationID, state.OrchestrationID, requestID)
	}
}

// sendTimeoutResponse sends a timeout response to parent
func (tm *TimeoutMonitor) sendTimeoutResponse(ctx context.Context, parentOrchID, childOrchID, requestID string) {
	// Get parent state to find the awaited request details
	parentState, err := tm.repo.GetState(ctx, parentOrchID)
	if err != nil {
		tm.logger.Error("Failed to get parent state for timeout response", zap.Error(err))
		return
	}

	// Find the awaited request to get the response topic
	awaitedReq, exists := parentState.AwaitedRequests[requestID]
	if !exists {
		tm.logger.Error("Request not found in parent's awaited list",
			zap.String("request_id", requestID),
			zap.String("parent_orch_id", parentOrchID))
		return
	}

	// Use the response topic from the awaited request
	responseTopic := awaitedReq.ResponseTopic
	if responseTopic == "" {
		// Fallback if not set (for backward compatibility)
		if parentState.OwnerAgentType != "" {
			responseTopic = fmt.Sprintf("system.agent.%s.responses", parentState.OwnerAgentType)
		} else {
			responseTopic = "system.orchestrator.responses"
		}
		tm.logger.Warn("ResponseTopic not set in awaited request, using fallback",
			zap.String("fallback_topic", responseTopic))
	}

	response := &types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender: types.AgentIdentity{
				AgentType:    "timeout-monitor",
				AgentID:      tm.getPodName(),
				PodName:      tm.getPodName(),
				AgentVersion: "1.0.0",
			},

			// Response tracking - matching what the parent expects
			InResponseToRequestID:      requestID,
			InResponseToStepID:         awaitedReq.StepID,
			InResponseToStepName:       awaitedReq.StepName,
			InResponseToParentOrchID:   parentOrchID,
			InResponseToParentOrchName: "", // Could store this too if needed
			InResponseToMessageID:      "", // Could store this too
			InResponseToAction:         "", // Could store this too
			RetryCount:                 awaitedReq.RetryVersion,

			// My context (the timeout monitor acting on behalf of child)
			MyOrchestrationID:   childOrchID,
			MyOrchestrationName: fmt.Sprintf("timeout-monitor-for-%s", childOrchID),

			// Identity
			CorrelationID: parentState.CorrelationID,
			ClientID:      parentState.ClientID,
			MessageType:   "response",
			FromAgent:     tm.getPodName(),
			ToAgent:       parentState.OwnerAgentID,
			ToAgentType:   parentState.OwnerAgentType,

			// Status flags
			IsComplete:          true,
			IsError:             true,
			IsMultipartResponse: false,
			PartCount:           1,
			Status:              "error_unrecoverable",

			// Timing & Resources
			TimeSent:                   time.Now(),
			TimeSpent:                  time.Since(awaitedReq.SentAt),
			OverallTimeBudgetRemaining: 0,
			TopicSentTo:                responseTopic,
			FuelUsed:                   0,
			RemainingFuelBudget:        0,
		},
		Body: types.ResponseBody{
			Success: false,
			Headers: nil,
			Body: struct {
				Result      interface{}      `json:"result"`
				Calculation interface{}      `json:"calculation,omitempty"`
				Error       *types.ErrorInfo `json:"error,omitempty"`
			}{
				Result:      nil,
				Calculation: nil,
				Error: &types.ErrorInfo{
					Code:        "TIMEOUT",
					Message:     fmt.Sprintf("Child orchestration %s timed out after %v", childOrchID, time.Since(awaitedReq.SentAt)),
					Recoverable: false,
					RetryAfter:  0,
				},
			},
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		tm.logger.Error("Failed to marshal timeout response", zap.Error(err))
		return
	}

	// Send to the correct response topic
	if err := tm.producer.Produce(ctx, responseTopic, response.Headers.ToMap(),
		[]byte(requestID), responseBytes); err != nil {
		tm.logger.Error("Failed to send timeout response",
			zap.Error(err),
			zap.String("topic", responseTopic),
			zap.String("request_id", requestID))
	} else {
		tm.logger.Info("Sent timeout response",
			zap.String("topic", responseTopic),
			zap.String("request_id", requestID),
			zap.String("parent_orch_id", parentOrchID),
			zap.String("child_orch_id", childOrchID))
	}
}

func (tm *TimeoutMonitor) getPodName() string {
	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		podName = "timeout-monitor"
	}
	return podName
}

// ParentContext holds the parent orchestration's context
type ParentContext struct {
	OrchestrationID   string
	OrchestrationName string
	CorrelationID     string
	CorrelationName   string
	AgentType         string
	AgentID           string
	ClientID          string
}

// ExtractParentContext gets the current (parent) orchestration's context
func ExtractParentContext(headers map[string]string) (*ParentContext, error) {
	orchestrationID := headers["orchestration_id"]
	if orchestrationID == "" {
		return nil, fmt.Errorf("current orchestration_id not found - cannot start child without parent context")
	}

	return &ParentContext{
		OrchestrationID:   orchestrationID,
		OrchestrationName: headers["orchestration_name"],
		CorrelationID:     headers["correlation_id"],
		CorrelationName:   headers["correlation_name"],
		AgentType:         headers["agent_type"],
		AgentID:           headers["agent_id"],
		ClientID:          headers["client_id"],
	}, nil
}

// ChildOrchestrationIDs holds all IDs for the child orchestration
type ChildOrchestrationIDs struct {
	OrchestrationID   string // Child's orchestration ID
	OrchestrationName string // Child's orchestration name
	MessageID         string // Message ID for the creation request
	RequestID         string // Request ID that parent will wait for
}

// CreateChildOrchestrationIDs generates IDs for a child orchestration
func CreateChildOrchestrationIDs(childName string) *ChildOrchestrationIDs {
	if childName == "" {
		childName = fmt.Sprintf("child-orch-%s", time.Now().Format("1504"))
	}

	return &ChildOrchestrationIDs{
		OrchestrationID:   uuid.New().String(),
		OrchestrationName: childName,
		MessageID:         uuid.New().String(),
		RequestID:         uuid.New().String(),
	}
}

// BuildChildHeaders builds headers for a child orchestration
func BuildChildHeaders(
	parentHeaders map[string]string,
	parent *ParentContext,
	childIDs *ChildOrchestrationIDs,
	spawnData map[string]interface{},
) map[string]string {
	headers := make(map[string]string)

	// Copy relevant parent headers
	for k, v := range parentHeaders {
		if k != "orchestration_id" && k != "message_id" && k != "request_id" {
			headers[k] = v
		}
	}

	// Set orchestration hierarchy
	headers["orchestration_id"] = childIDs.OrchestrationID
	headers["orchestration_name"] = childIDs.OrchestrationName
	headers["parent_orchestration_id"] = parent.OrchestrationID
	headers["parent_orchestration_name"] = parent.OrchestrationName
	headers["correlation_id"] = parent.CorrelationID
	headers["correlation_name"] = parent.CorrelationName
	headers["message_id"] = childIDs.MessageID
	headers["request_id"] = childIDs.RequestID
	headers["parent_agent_type"] = parent.AgentType
	headers["parent_agent_id"] = parent.AgentID
	headers["client_id"] = parent.ClientID

	// Set response topic for child to respond to parent
	headers["parent_reply_to_topic"] = fmt.Sprintf("system.agent.%s.responses", parent.AgentType)

	// Add agent mappings from spawn data
	if agents, ok := spawnData["agents"].(map[string]interface{}); ok {
		for role, agentID := range agents {
			if id, ok := agentID.(string); ok {
				headers[fmt.Sprintf("agent_%s", role)] = id
			}
		}

		// Set agent_instance_id to orchestrator if present
		if orchestratorID, ok := agents["orchestrator"].(string); ok {
			headers["agent_instance_id"] = orchestratorID
		}
	}

	// If no agent_instance_id set, use parent's
	if headers["agent_instance_id"] == "" {
		headers["agent_instance_id"] = parent.AgentID
	}

	return headers
}

// OrchestratorHelper provides helper functions for orchestration
type OrchestratorHelper struct {
	db             *sql.DB
	producer       kafka.Producer
	logger         *zap.Logger
	repo           *StateRepository
	timeoutMonitor *TimeoutMonitor
}

// NewOrchestratorHelper creates a new orchestrator helper
func NewOrchestratorHelper(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *OrchestratorHelper {
	return &OrchestratorHelper{
		db:             db,
		producer:       producer,
		logger:         logger,
		repo:           NewStateRepository(db, logger),
		timeoutMonitor: NewTimeoutMonitor(db, producer, logger),
	}
}

// StartChildOrchestration starts a child orchestration with timeout monitoring
func (oh *OrchestratorHelper) StartChildOrchestration(
	ctx context.Context,
	parent *ParentContext,
	childName string,
	workflow json.RawMessage,
	spawnData map[string]interface{},
	timeout time.Duration,
) (*ChildOrchestrationIDs, error) {

	// Create child IDs
	childIDs := CreateChildOrchestrationIDs(childName)

	// Build child headers
	childHeaders := BuildChildHeaders(nil, parent, childIDs, spawnData)

	// Create child orchestration state
	childState := &OrchestrationState{
		OrchestrationID:       childIDs.OrchestrationID,
		CorrelationID:         parent.CorrelationID,
		OwnerAgentID:          childHeaders["agent_instance_id"],
		ParentOrchestrationID: parent.OrchestrationID,
		ClientID:              parent.ClientID,
		Status:                StatusInitialized,
		CurrentStep:           "",
		AwaitedRequests:       make(map[string]*AwaitedRequest),
		SubtreeAgents:         make(map[string]*types.SubtreeInfo),
		ProcessingHistory:     []ProcessingRecord{},
		CollectedData:         spawnData,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Version:               1,
	}

	// Parse and set workflow
	if workflow != nil {
		var workflowPlan models.WorkflowPlan
		if err := json.Unmarshal(workflow, &workflowPlan); err != nil {
			oh.logger.Error("Failed to parse workflow", zap.Error(err))
			return nil, fmt.Errorf("failed to parse workflow: %w", err)
		}
		childState.WorkflowPlan = workflowPlan
	}

	// Create the orchestration in database
	if err := oh.repo.CreateState(ctx, childState); err != nil {
		return nil, fmt.Errorf("failed to create child orchestration state: %w", err)
	}

	// Add to parent's awaited requests
	parentAwaitedReq := &AwaitedRequest{
		RequestID:       childIDs.RequestID,
		StepID:          uuid.New().String(),
		StepName:        "child_orchestration",
		RetryVersion:    0,
		TargetAgentType: childHeaders["agent_instance_id"],
		SentAt:          time.Now(),
		TimeoutAt:       time.Now().Add(timeout),
	}

	if err := oh.repo.AddAwaitedRequest(ctx, parent.OrchestrationID, parentAwaitedReq); err != nil {
		oh.logger.Error("Failed to add child to parent's awaited requests", zap.Error(err))
	}

	// Start timeout monitoring
	oh.timeoutMonitor.MonitorChildOrchestration(
		parent.OrchestrationID,
		childIDs.OrchestrationID,
		childIDs.RequestID,
		timeout,
	)

	oh.logger.Info("Child orchestration started with monitoring",
		zap.String("parent_orchestration_id", parent.OrchestrationID),
		zap.String("child_orchestration_id", childIDs.OrchestrationID),
		zap.String("request_id", childIDs.RequestID),
		zap.Duration("timeout", timeout))

	return childIDs, nil
}

// GetOrchestratorAgentType finds the agent_type for the orchestrator role in a group
func (oh *OrchestratorHelper) GetOrchestratorAgentType(ctx context.Context, groupID string) (string, error) {
	var agentConfigsJSON []byte
	query := "SELECT agent_configs FROM agent_group_definitions WHERE group_id = $1"
	err := oh.db.QueryRowContext(ctx, query, groupID).Scan(&agentConfigsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			oh.logger.Warn("No agent group definition found", zap.String("group_id", groupID))
			return "", fmt.Errorf("agent group definition not found for group_id: %s", groupID)
		}
		return "", fmt.Errorf("failed to query agent group definition: %w", err)
	}

	var agentConfigs []map[string]interface{}
	if err := json.Unmarshal(agentConfigsJSON, &agentConfigs); err != nil {
		return "", fmt.Errorf("failed to unmarshal agent_configs: %w", err)
	}

	for _, config := range agentConfigs {
		if role, ok := config["role"].(string); ok && role == "orchestrator" {
			if agentType, ok := config["agent_type"].(string); ok {
				oh.logger.Info("Found orchestrator agent type",
					zap.String("group_id", groupID),
					zap.String("agent_type", agentType))
				return agentType, nil
			}
		}
	}

	return "", fmt.Errorf("orchestrator agent type not found for group_id: %s", groupID)
}

func (tm *TimeoutMonitor) getAgentType(ctx context.Context, agentID string) string {
	// Try to get from agent instances table
	var agentType string
	query := `
        SELECT agent_type 
        FROM agent_instances 
        WHERE agent_id = $1
        LIMIT 1
    `

	err := tm.db.QueryRowContext(ctx, query, agentID).Scan(&agentType)
	if err == nil && agentType != "" {
		return agentType
	}

	// Fallback to parsing the ID
	return ""
}
