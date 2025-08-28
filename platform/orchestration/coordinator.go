// FILE: platform/orchestration/coordinator.go (enhanced with logging)
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
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"go.uber.org/zap"
)

const (
	// Topic for notifications to the UI
	NotificationTopic = "system.notifications.ui"
	// Topic for receiving resume commands
	ResumeWorkflowTopic = "system.commands.workflow.resume"
)

// SagaCoordinator manages the execution of complex workflows
type SagaCoordinator struct {
	db          *sql.DB
	producer    kafka.Producer
	logger      *zap.Logger
	fuelManager *governance.FuelManager
}

var actionRegistry = map[string]actions.ActionHandler{
	"validate_input":      actions.ValidateInputAction,
	"transform_data":      actions.TransformDataAction,
	"send_notification":   actions.SendNotificationAction,
	"spawn_agent":         actions.SpawnAgentAction,
	"spawn_group":         actions.SpawnGroupAction,
	"call_agent":          actions.CallAgentAction,
	"discover_agents":     actions.DiscoverAgentsAction,
	"execute_llm_prompt":  actions.ExecuteLLMPromptAction,
	"start_orchestration": actions.StartOrchestrationAction,
	// generic actions
	"validate_schema":    actions.ValidateSchemaAction,
	"retrieve_memory":    actions.RetrieveMemoryAction,
	"store_memory":       actions.StoreMemoryAction,
	"validate_assets":    actions.ValidateAssetsAction,
	"deploy_to_hosting":  actions.DeployToHostingAction,
	"http_request":       actions.HTTPRequestAction,
	"conditional_branch": actions.ConditionalBranchAction,
	"aggregate_data":     actions.AggregateDataAction,
	"cache_lookup":       actions.CacheLookupAction,

	"ai_text_generate_anthropic": actions.ExecuteLLMPromptAction, // Map to existing action

	"plan_agent_team":       actions.PlanAgentTeamAction,
	"review_performance":    actions.ReviewPerformanceAction,
	"approve_agent_changes": actions.ApproveAgentChangesAction,
	"conditional_route":     actions.ConditionalRouteAction,

	// HTML-specific actions
	"generate_html": actions.GenerateHTMLAction,
	"process_html":  actions.ProcessHTMLAction,
	"validate_html": actions.ValidateHTMLAction,

	// Storage actions
	"route_storage": actions.RouteStorageAction,
	"upload_to_s3":  actions.UploadToS3Action,
	"s3_upload":     actions.UploadToS3Action,
	"store_result":  actions.StoreResultAction,

	"complete_workflow": actions.CompleteWorkflowAction,

	"evaluate_task": actions.EvaluateTaskAction,
}

// NewSagaCoordinator creates a new coordinator instance
func NewSagaCoordinator(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *SagaCoordinator {
	return &SagaCoordinator{
		db:          db,
		producer:    producer,
		logger:      logger,
		fuelManager: governance.NewFuelManager(),
	}
}

// ExecuteWorkflow now stores the plan and continues execution
// ExecuteWorkflow now properly manages orchestration_id
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	correlationID := headers["correlation_id"]
	l := s.logger.With(zap.String("correlation_id", correlationID))

	l.Info("ExecuteWorkflow called",
		zap.String("start_step", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	clientID := headers["client_id"]
	if clientID == "" {
		l.Error("client_id header is required to execute a workflow")
		return fmt.Errorf("client_id header is required to execute a workflow")
	}

	// Get or create state with the plan
	state, orchestrationID, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData, headers)
	if err != nil {
		l.Error("Failed to get or create state", zap.Error(err))
		return err
	}

	// CRITICAL: Set orchestration_id in headers for all subsequent operations
	headers["orchestration_id"] = orchestrationID
	headers["owner_agent_id"] = state.OwnerAgentID

	l.Info("Workflow state retrieved",
		zap.String("orchestration_id", orchestrationID),
		zap.String("status", string(state.Status)),
		zap.String("current_step", state.CurrentStep))

	// Check if workflow is already complete
	if state.Status == StatusCompleted || state.Status == StatusFailed {
		l.Info("Workflow already finished",
			zap.String("orchestration_id", orchestrationID),
			zap.String("status", string(state.Status)))
		return nil
	}

	// Continue execution from current step
	return s.continueExecution(ctx, state, headers)
}

// Updated getOrCreateState to return orchestrationID explicitly
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte, headers map[string]string) (*OrchestrationState, string, error) {
	repo := NewStateRepository(s.db, s.logger)

	// Check if we already have an orchestration_id in headers
	if orchestrationID := headers["orchestration_id"]; orchestrationID != "" {
		// Try to get by orchestration ID first
		state, err := repo.GetState(ctx, orchestrationID)
		if err == nil {
			return state, orchestrationID, nil
		}
		s.logger.Warn("orchestration_id in headers but state not found, creating new",
			zap.String("orchestration_id", orchestrationID),
			zap.Error(err))
	}

	// Try to get by correlation ID
	state, err := repo.GetStateByCorrelation(ctx, correlationID)
	if err == nil {
		// Found existing state
		return state, state.OrchestrationID, nil
	}

	// Create new orchestration
	orchestrationID := uuid.New().String()
	ownerAgentID := headers["agent_id"]
	if ownerAgentID == "" {
		ownerAgentID = os.Getenv("AGENT_ID")
		if ownerAgentID == "" {
			ownerAgentID = "00000000-0000-0000-0000-000000000001"
		}
	}

	ParentOrchestrationID := headers["parent_orchestration_id"]
	if ParentOrchestrationID == "" {
		ParentOrchestrationID = headers["parent_orchestration_id"]
	}

	if err := repo.CreateInitialState(ctx, orchestrationID, correlationID, ownerAgentID, ParentOrchestrationID, clientID, plan, initialData); err != nil {
		return nil, "", err
	}

	// Get the created state
	state, err = repo.GetState(ctx, orchestrationID)
	if err != nil {
		return nil, "", err
	}

	return state, orchestrationID, nil
}

// executeLocalAction handles actions that run within the orchestrator
func (s *SagaCoordinator) executeLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) (interface{}, error) {
	l := s.logger.With(
		zap.String("DEBUG_COOR_11: orchestration_id", state.OrchestrationID), // Add this
		zap.String("DEBUG_COOR_11: correlation_id", state.CorrelationID),
		zap.String("DEBUG_COOR_11: action", step.Action),
	)

	handler, ok := actionRegistry[step.Action]
	if !ok {
		return nil, fmt.Errorf("local action '%s' not found in registry", step.Action)
	}

	// Prepare headers for this orchestration's context
	actionHeaders := make(map[string]string)
	for k, v := range headers {
		actionHeaders[k] = v
	}

	// CRITICAL: Ensure orchestration_id is always present
	actionHeaders["orchestration_id"] = state.OrchestrationID
	actionHeaders["correlation_id"] = state.CorrelationID
	actionHeaders["client_id"] = state.ClientID

	// Log with virtual topic for consistency
	virtualTopic := fmt.Sprintf("local.action.%s", step.Action)
	l.Info("Executing local action",
		zap.String("DEBUG_COOR_12: virtual_topic", virtualTopic),
		zap.String("DEBUG_COOR_12: orchestration_id", state.OrchestrationID))

	// Verify producer is available
	if s.producer == nil {
		l.Error("Producer is nil in SagaCoordinator")
		return nil, fmt.Errorf("producer not available for action execution")
	}

	// For call_agent actions, ensure proper context
	if step.Action == "call_agent" {
		// The agent being called needs the current orchestration context
		actionHeaders["orchestration_id"] = state.OrchestrationID
		actionHeaders["correlation_id"] = state.CorrelationID

		// Set reply information so responses come back correctly
		actionHeaders["reply_to_topic"] = fmt.Sprintf("system.agent.%s.responses", headers["agent_type"])
		actionHeaders["from_agent_id"] = state.OwnerAgentID
		actionHeaders["from_agent_type"] = headers["agent_type"]

		l.Info("Prepared headers for call_agent",
			zap.String("DEBUG_COOR_14: orchestration_id", actionHeaders["orchestration_id"]),
			zap.String("DEBUG_COOR_14: correlation_id", actionHeaders["correlation_id"]))
	}

	// For start_orchestration, ensure it knows its parent
	if step.Action == "start_orchestration" {
		// The child orchestration needs to know THIS orchestration is its parent
		actionHeaders["parent_orchestration_id"] = state.OrchestrationID // THIS is critical
		actionHeaders["parent_correlation_id"] = state.CorrelationID
		actionHeaders["parent_agent_type"] = headers["agent_type"]

		l.Info("Prepared headers for start_orchestration",
			zap.String("DEBUG_COOR_15: parent_orchestration_id", state.OrchestrationID),
			zap.String("DEBUG_COOR_15: parent_correlation_id", state.CorrelationID))
	}

	// Prepare action parameters
	params := actions.ActionParams{
		Context:         ctx,
		Headers:         actionHeaders,
		StepConfig:      step,
		InputData:       state.InitialRequestData,
		CollectedData:   state.CollectedData,
		SagaCoordinator: s,
		Producer:        s.producer,
		DB:              s.db,
		Logger:          s.logger,
		AgentType:       actionHeaders["agent_type"],
		CurrentStep:     state.CurrentStep,
	}

	// If agent_type is not in headers, try to get it from other sources
	if params.AgentType == "" {
		if agentConfig, ok := state.CollectedData["agent_config"].(map[string]interface{}); ok {
			if at, ok := agentConfig["agent_type"].(string); ok {
				params.AgentType = at
				actionHeaders["agent_type"] = at
			}
		}
	}

	l.Info("Executing local action with params",
		zap.String("DEBUG_COOR_16: agent_type", params.AgentType),
		zap.String("DEBUG_COOR_16: action", step.Action),
		zap.String("DEBUG_COOR_16: orchestration_id", state.OrchestrationID))

	// Execute the action
	result, err := handler(ctx, params)
	if err != nil {
		l.Error("Local action failed", zap.Error(err))
		return nil, fmt.Errorf("local action failed: %w", err)
	}

	// CHECK FOR AWAIT_RESPONSE FLAG
	if resultMap, ok := result.(map[string]interface{}); ok {
		if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
			var requestID string

			// Check for different ID fields based on action type
			if step.Action == "start_orchestration" {
				// For start_orchestration, we wait for the child's correlation_id
				if id, ok := resultMap["new_correlation_id"].(string); ok {
					requestID = id
				}
			} else if id, ok := resultMap["request_id"].(string); ok {
				// For other actions like call_agent
				requestID = id
			}

			if requestID != "" {
				l.Info("Local action requires waiting for response",
					zap.String("DEBUG_COOR_17: request_id", requestID),
					zap.String("DEBUG_COOR_17: action", step.Action),
					zap.String("DEBUG_COOR_17: orchestration_id", state.OrchestrationID))

				// Store result first
				if state.CurrentStep != "" {
					state.CollectedData[state.CurrentStep] = result
				} else {
					state.CollectedData[step.Action] = result
				}

				// Update state to wait
				state.Status = StatusAwaitingResponses
				state.AwaitedSteps = []string{requestID}

				// Update metadata
				state.ExecutionMetadata.CompletedSteps++
				state.ExecutionMetadata.Checkpoints[step.Action] = time.Now().UTC()

				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					return result, fmt.Errorf("failed to update state for waiting: %w", err)
				}

				// Set up timeout if needed
				if step.Action == "start_orchestration" {
					timeout := 5 * time.Minute
					if step.Timeout > 0 {
						timeout = step.Timeout
					}
					// Pass the current orchestration_id, not correlation_id
					go s.handleChildOrchestrationTimeout(ctx, state.OrchestrationID, requestID, timeout)
				}

				return result, nil
			}
		}
	}

	// Store result
	if state.CurrentStep != "" {
		state.CollectedData[state.CurrentStep] = result
	} else {
		state.CollectedData[step.Action] = result
	}

	// Update metadata
	state.ExecutionMetadata.CompletedSteps++
	state.ExecutionMetadata.Checkpoints[step.Action] = time.Now().UTC()

	// Move to next step
	if step.NextStep != "" {
		l.Info("Moving to next step",
			zap.String("next_step", step.NextStep),
			zap.String("orchestration_id", state.OrchestrationID))

		state.CurrentStep = step.NextStep

		repo := NewStateRepository(s.db, s.logger)
		if err := repo.UpdateState(ctx, state); err != nil {
			return result, fmt.Errorf("failed to update state: %w", err)
		}

		// Continue execution with properly set headers
		return result, s.continueExecution(ctx, state, actionHeaders)
	}

	return result, nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	l := s.logger.With(
		zap.String("DEBUG_COOR_19: correlation_id", state.CorrelationID),
		zap.String("DEBUG_COOR_19: orchestration_id", state.OrchestrationID),
		zap.String("DEBUG_COOR_19: action", step.Action),
		zap.String("DEBUG_COOR_19: topic", step.Topic),
	)

	// Prepare the message
	payload := models.TaskRequest{
		Action: step.Action,
		Data:   state.CollectedData,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Create new request ID
	newRequestID := uuid.NewString()
	outHeaders := make(map[string]string)
	for k, v := range headers {
		outHeaders[k] = v
	}
	outHeaders["causation_id"] = headers["request_id"]
	outHeaders["in_response_to"] = headers["request_id"]
	outHeaders["correlation_id"] = state.CorrelationID
	outHeaders["orchestration_id"] = state.OrchestrationID
	outHeaders["request_id"] = newRequestID
	outHeaders["client_id"] = state.ClientID

	// Add response routing information
	outHeaders["reply_to_topic"] = fmt.Sprintf("system.agent.%s.responses", headers["agent_type"])
	outHeaders["from_agent_id"] = state.OwnerAgentID
	outHeaders["from_agent_type"] = headers["agent_type"]

	l.Info("Sending remote action",
		zap.String("DEBUG_COOR_20: request_id", newRequestID),
		zap.String("DEBUG_COOR_20: topic", step.Topic),
		zap.String("DEBUG_COOR_20: orchestration_id", state.OrchestrationID),
		zap.String("DEBUG_COOR_20: correlation_id", state.CorrelationID))

	// Send the message
	if err := s.producer.Produce(ctx, step.Topic, outHeaders,
		[]byte(state.CorrelationID), payloadBytes); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Update state to await response
	state.Status = StatusAwaitingResponses
	state.CurrentStep = step.NextStep
	state.AwaitedSteps = []string{newRequestID}

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	l.Info("Remote action initiated",
		zap.String("DEBUG_COOR_21: request_id", newRequestID),
		zap.String("DEBUG_COOR_21: orchestration_id", state.OrchestrationID))

	return nil
}

// completeWorkflow marks the workflow as completed with enhanced tracking
func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState) error {
	l := s.logger.With(
		zap.String("DEBUG_COOR_22: correlation_id", state.CorrelationID),
		zap.String("DEBUG_COOR_22: orchestration_id", state.OrchestrationID),
	)

	l.Info("Completing workflow",
		zap.Int("DEBUG_COOR_22: completed_steps", state.ExecutionMetadata.CompletedSteps),
		zap.Int("DEBUG_COOR_22: total_steps", len(state.WorkflowPlan.Steps)))

	state.Status = StatusCompleted
	finalResult, _ := json.Marshal(state.CollectedData)
	state.FinalResult = finalResult

	// Update metadata
	now := time.Now().UTC()
	state.ExecutionMetadata.EndTime = &now
	state.ExecutionMetadata.CompletedSteps++

	repo := NewStateRepository(s.db, s.logger)
	err := repo.UpdateState(ctx, state)

	if err != nil {
		l.Error("Failed to update workflow completion state", zap.Error(err))
	} else {
		l.Info("Workflow completed successfully")
	}

	// Check if this is a child orchestration by looking at ParentOrchestrationID
	if parentContext, ok := state.CollectedData["__parent_context__"].(map[string]interface{}); ok {
		parentOrchestrationID, _ := parentContext["orchestration_id"].(string)
		replyToTopic, _ := parentContext["reply_to_topic"].(string)
		requestID, _ := parentContext["request_id"].(string)

		if parentOrchestrationID != "" && replyToTopic != "" {
			// Send completion to parent using stored context
			parentResponse := models.TaskResponse{
				Success: true,
				Data: map[string]interface{}{
					"status":         "completed",
					"correlation_id": state.CorrelationID,
					"final_result":   state.CollectedData,
				},
			}

			responseHeaders := map[string]string{
				"orchestration_id": parentOrchestrationID,
				"in_response_to":   requestID,
				"correlation_id":   state.CorrelationID,
				"message_type":     "response",
			}

			responseBytes, _ := json.Marshal(parentResponse)
			err := s.producer.Produce(ctx, replyToTopic, responseHeaders,
				[]byte(state.CorrelationID), responseBytes)

			if err != nil {
				s.logger.Error("Failed to notify parent", zap.Error(err))
			}
		}
	}

	return nil
}

// failWorkflow marks the workflow as failed
func (s *SagaCoordinator) failWorkflow(ctx context.Context, state *OrchestrationState, errorMsg string) error {
	l := s.logger.With(
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("correlation_id", state.CorrelationID))

	l.Error("Failing workflow",
		zap.String("error", errorMsg),
		zap.String("orchestration_id", state.OrchestrationID))

	state.Status = StatusFailed
	state.Error = errorMsg

	// Set end time for metadata
	now := time.Now().UTC()
	state.ExecutionMetadata.EndTime = &now

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		l.Error("Failed to update state to failed",
			zap.Error(err),
			zap.String("orchestration_id", state.OrchestrationID))
		return fmt.Errorf("failed to update state to failed: %w", err)
	}

	// If this is a child orchestration, notify parent of failure
	if state.ParentOrchestrationID != "" {
		l.Info("Notifying parent of child failure",
			zap.String("parent_orchestration_id", state.ParentOrchestrationID))

		// Get parent's state to find its correlation ID
		parentState, err := repo.GetState(ctx, state.ParentOrchestrationID)
		if err == nil && parentState != nil {
			// Prepare failure notification
			failureResponse := models.TaskResponse{
				Success: false,
				Error:   errorMsg,
				Data: map[string]interface{}{
					"status":         "failed",
					"correlation_id": state.CorrelationID,
					"error":          errorMsg,
					"execution_stats": map[string]interface{}{
						"completed_steps": state.ExecutionMetadata.CompletedSteps,
						"total_steps":     state.ExecutionMetadata.TotalSteps,
						"duration_ms":     now.Sub(state.ExecutionMetadata.StartTime).Milliseconds(),
					},
				},
			}

			responseBytes, _ := json.Marshal(failureResponse)

			// Create headers with proper orchestration context
			responseHeaders := map[string]string{
				"orchestration_id": state.ParentOrchestrationID, // Parent's orchestration ID
				"correlation_id":   parentState.CorrelationID,   // Parent's correlation ID
				"in_response_to":   state.CorrelationID,         // Child's correlation
				"causation_id":     state.CorrelationID,         // For backward compatibility
				"message_type":     "response",
				"from_agent_id":    state.OwnerAgentID,
				"client_id":        state.ClientID,
			}

			// Determine parent's response topic
			parentResponseTopic := "system.agent.generic.responses"
			if parentAgentType, ok := parentState.CollectedData["agent_type"].(string); ok {
				parentResponseTopic = fmt.Sprintf("system.agent.%s.responses", parentAgentType)
			}

			// Send failure notification to parent
			if err := s.producer.Produce(ctx,
				parentResponseTopic,
				responseHeaders,
				[]byte(parentState.CorrelationID),
				responseBytes); err != nil {
				l.Error("Failed to notify parent of child failure", zap.Error(err))
			} else {
				l.Info("Parent notified of child failure")
			}
		}
	}

	return fmt.Errorf(errorMsg)
}

// Helper functions remain the same
func isLocalAction(action string) bool {
	_, exists := actionRegistry[action]
	return exists
}

func (s *SagaCoordinator) dependenciesMet(dependencies []string, state *OrchestrationState) bool {
	for _, dep := range dependencies {
		if _, ok := state.CollectedData[dep]; !ok {
			return false
		}
	}
	return true
}

func (s *SagaCoordinator) recordExecution(ctx context.Context, state *OrchestrationState, record ExecutionRecord) {
	state.ExecutionPath = append(state.ExecutionPath, record)

	// Update metadata based on result
	switch record.Result {
	case "failed":
		state.ExecutionMetadata.FailedSteps++
	case "skipped":
		state.ExecutionMetadata.SkippedSteps++
	}

	// Don't fail the workflow if we can't update tracking
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		s.logger.Error("Failed to record execution",
			zap.Error(err),
			zap.String("DEBUG_COOR_32: step", record.Step))
	}
}

// HandleResponse processes responses and continues workflow
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
	s.logger.Info("HandleResponse called", zap.Any("headers", headers))

	// Step 1: Extract and validate IDs
	ids, err := s.extractResponseIDs(headers)
	if err != nil {
		return fmt.Errorf("failed to extract IDs: %w", err)
	}

	// Step 2: Find the correct orchestration state
	state, err := s.findTargetOrchestration(ctx, ids, headers)
	if err != nil {
		return fmt.Errorf("failed to find orchestration: %w", err)
	}

	// Step 3: Parse and validate the response
	taskResponse, err := s.parseResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Step 4: Process the response
	if err := s.processResponseData(ctx, state, ids, taskResponse); err != nil {
		return fmt.Errorf("failed to process response: %w", err)
	}

	// Step 5: Check if workflow can continue
	if len(state.AwaitedSteps) == 0 {
		return s.continueWorkflowAfterResponse(ctx, state, headers)
	}

	// Still waiting for more responses
	s.logger.Info("Still waiting for responses",
		zap.Int("remaining", len(state.AwaitedSteps)),
		zap.Strings("awaited", state.AwaitedSteps))

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

// ResponseIDs holds all the different IDs from a response
type ResponseIDs struct {
	OrchestrationID string // The orchestration that sent this response
	CorrelationID   string // Business transaction ID
	InResponseTo    string // What request this is responding to
	ParentOrchID    string // If this is from a child, who's the parent
}

// extractResponseIDs clearly extracts all IDs from headers
func (s *SagaCoordinator) extractResponseIDs(headers map[string]string) (*ResponseIDs, error) {
	ids := &ResponseIDs{
		OrchestrationID: headers["orchestration_id"],
		CorrelationID:   headers["correlation_id"],
		InResponseTo:    headers["in_response_to"],
		ParentOrchID:    headers["parent_orchestration_id"],
	}

	// Fallback for legacy messages
	if ids.InResponseTo == "" {
		ids.InResponseTo = headers["causation_id"]
	}

	if ids.InResponseTo == "" {
		return nil, fmt.Errorf("no in_response_to or causation_id in headers")
	}

	if ids.CorrelationID == "" {
		return nil, fmt.Errorf("missing correlation_id in headers")
	}

	s.logger.Info("Extracted response IDs",
		zap.String("orchestration_id", ids.OrchestrationID),
		zap.String("correlation_id", ids.CorrelationID),
		zap.String("in_response_to", ids.InResponseTo),
		zap.String("parent_orch_id", ids.ParentOrchID))

	return ids, nil
}

// findTargetOrchestration determines which orchestration should handle this response
func (s *SagaCoordinator) findTargetOrchestration(ctx context.Context, ids *ResponseIDs, headers map[string]string) (*OrchestrationState, error) {
	repo := NewStateRepository(s.db, s.logger)

	// The target is usually in the "orchestration_id" header
	// This should be the parent's orchestration_id for child responses
	targetOrchID := headers["orchestration_id"]

	// Try to get by orchestration_id
	if targetOrchID != "" {
		state, err := repo.GetState(ctx, targetOrchID)
		if err == nil {
			s.logger.Info("Found state by orchestration_id",
				zap.String("orchestration_id", targetOrchID))
			return state, nil
		}
	}

	// Fallback: try correlation_id
	if ids.CorrelationID != "" {
		state, err := repo.GetStateByCorrelation(ctx, ids.CorrelationID)
		if err == nil {
			s.logger.Info("Found state by correlation_id",
				zap.String("correlation_id", ids.CorrelationID))
			return state, nil
		}
	}

	return nil, fmt.Errorf("no state found for orchestration_id=%s, correlation_id=%s",
		targetOrchID, ids.CorrelationID)
}

// parseResponse unmarshals the response payload
func (s *SagaCoordinator) parseResponse(response []byte) (models.TaskResponse, error) {
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(response, &taskResponse); err != nil {
		s.logger.Error("Failed to unmarshal response", zap.Error(err))
		return taskResponse, err
	}

	s.logger.Info("Response parsed",
		zap.Bool("success", taskResponse.Success),
		zap.Any("data_keys", getMapKeys(taskResponse.Data)))

	return taskResponse, nil
}

// processResponseData stores the response and updates orchestration state
func (s *SagaCoordinator) processResponseData(ctx context.Context, state *OrchestrationState, ids *ResponseIDs, response models.TaskResponse) error {
	// Initialize collected data if needed
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}

	// Store response data under the request we're responding to
	responseKey := ids.InResponseTo

	// Store the response
	if finalResult, ok := response.Data["final_result"]; ok {
		// Child orchestration completion
		state.CollectedData[responseKey] = map[string]interface{}{
			"status":          "completed",
			"correlation_id":  response.Data["correlation_id"],
			"final_result":    finalResult,
			"execution_stats": response.Data["execution_stats"],
		}
	} else {
		// Regular agent response
		state.CollectedData[responseKey] = response.Data
	}

	s.logger.Info("Stored response data",
		zap.String("key", responseKey),
		zap.Int("data_size", len(response.Data)))

	// Remove from awaited steps
	updatedAwaitedSteps := []string{}
	found := false

	for _, awaitedID := range state.AwaitedSteps {
		if awaitedID == ids.InResponseTo {
			found = true
			s.logger.Info("Removed from awaited steps",
				zap.String("request_id", awaitedID))
		} else {
			updatedAwaitedSteps = append(updatedAwaitedSteps, awaitedID)
		}
	}

	if !found {
		s.logger.Warn("Response not in awaited steps",
			zap.String("in_response_to", ids.InResponseTo),
			zap.Strings("awaited_steps", state.AwaitedSteps))
	}

	state.AwaitedSteps = updatedAwaitedSteps
	state.ExecutionMetadata.CompletedSteps++

	// Complete request tracking if applicable
	if ids.InResponseTo != "" {
		if err := s.completeRequest(ctx, ids.InResponseTo); err != nil {
			s.logger.Warn("Failed to complete request tracking",
				zap.String("request_id", ids.InResponseTo),
				zap.Error(err))
		}
	}

	return nil
}

// continueWorkflowAfterResponse continues execution after all responses received
func (s *SagaCoordinator) continueWorkflowAfterResponse(ctx context.Context, state *OrchestrationState, originalHeaders map[string]string) error {
	s.logger.Info("All responses received, continuing workflow",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("current_step", state.CurrentStep))

	state.Status = StatusRunning

	// Move to next step if defined
	if currentStep, exists := state.WorkflowPlan.Steps[state.CurrentStep]; exists && currentStep.NextStep != "" {
		state.CurrentStep = currentStep.NextStep
		s.logger.Info("Moving to next step",
			zap.String("next_step", currentStep.NextStep))
	}

	// Update state first
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Build headers for continuation
	continueHeaders := s.buildContinuationHeaders(state, originalHeaders)

	return s.continueExecution(ctx, state, continueHeaders)
}

// buildContinuationHeaders creates headers for continuing workflow execution
func (s *SagaCoordinator) buildContinuationHeaders(state *OrchestrationState, originalHeaders map[string]string) map[string]string {
	headers := make(map[string]string)

	// Copy original headers
	for k, v := range originalHeaders {
		headers[k] = v
	}

	// Override with correct orchestration context
	headers["orchestration_id"] = state.OrchestrationID
	headers["correlation_id"] = state.CorrelationID
	headers["client_id"] = state.ClientID
	headers["owner_agent_id"] = state.OwnerAgentID

	// Preserve fuel budget
	if headers["fuel_budget"] == "" {
		if fuel, ok := state.CollectedData["initial_fuel_budget"].(string); ok {
			headers["fuel_budget"] = fuel
		} else {
			headers["fuel_budget"] = "1000"
		}
	}

	return headers
}

// Helper to complete a request in the tracking table
func (s *SagaCoordinator) completeRequest(ctx context.Context, requestID string) error {
	query := `
		UPDATE pending_requests 
		SET status = 'completed', 
		    completed_at = NOW()
		WHERE request_id = $1 AND status = 'pending'
	`

	_, err := s.db.ExecContext(ctx, query, requestID)
	if err != nil {
		s.logger.Error("Failed to complete request",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	return nil
}

// Helper to ensure headers have required values
func (s *SagaCoordinator) ensureHeaders(headers map[string]string, state *OrchestrationState) {
	// Ensure orchestration_id
	if headers["orchestration_id"] == "" {
		headers["orchestration_id"] = state.OrchestrationID
	}

	// Ensure client_id
	if headers["client_id"] == "" {
		headers["client_id"] = state.ClientID
	}

	// Ensure owner_agent_id
	if headers["owner_agent_id"] == "" {
		headers["owner_agent_id"] = state.OwnerAgentID
	}

	// Try to restore fuel_budget from collected data
	if headers["fuel_budget"] == "" {
		if fuel, ok := state.CollectedData["initial_fuel_budget"].(string); ok {
			headers["fuel_budget"] = fuel
		} else {
			headers["fuel_budget"] = "1000" // Default fallback
		}
	}

	// Ensure agent_instance_id
	if headers["agent_instance_id"] == "" {
		headers["agent_instance_id"] = state.OwnerAgentID
	}
}

// Add other missing methods like handleFanOut, handlePauseForHumanInput etc. with similar logging enhancements
func (s *SagaCoordinator) handleFanOut(ctx context.Context, headers map[string]string, step models.Step, state *OrchestrationState) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID),
		zap.String("orchestration_id", state.OrchestrationID))

	awaitedSteps := make([]string, 0, len(step.SubTasks))

	for _, subTask := range step.SubTasks {
		payload := models.TaskRequest{
			Action: subTask.StepName,
			Data:   state.CollectedData,
		}
		payloadBytes, _ := json.Marshal(payload)

		newRequestID := uuid.NewString()
		outHeaders := make(map[string]string)
		for k, v := range headers {
			outHeaders[k] = v
		}
		outHeaders["causation_id"] = headers["request_id"]
		outHeaders["request_id"] = newRequestID

		if err := s.producer.Produce(ctx, subTask.Topic, outHeaders, []byte(state.OrchestrationID), payloadBytes); err != nil {
			return fmt.Errorf("failed to produce fan-out message: %w", err)
		}

		awaitedSteps = append(awaitedSteps, newRequestID)
	}

	// Update state
	state.Status = StatusAwaitingResponses
	state.CurrentStep = step.NextStep
	state.AwaitedSteps = awaitedSteps

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	l.Info("Fan-out executed", zap.Int("subtasks", len(step.SubTasks)))
	return nil
}

func (s *SagaCoordinator) handlePauseForHumanInput(ctx context.Context, headers map[string]string, step models.Step, state *OrchestrationState) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID),
		zap.String("orchestration_id", state.OrchestrationID))

	state.Status = StatusPausedForHuman
	state.CurrentStep = step.NextStep

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Send notification
	notification := map[string]interface{}{
		"event_type":       "WORKFLOW_PAUSED_FOR_APPROVAL",
		"correlation_id":   state.CorrelationID,
		"orchestration_id": state.OrchestrationID,
		"project_id":       headers["project_id"],
		"client_id":        headers["client_id"],
		"message":          fmt.Sprintf("Step '%s' requires your approval", step.Description),
		"data_for_review":  state.CollectedData,
	}
	notificationBytes, _ := json.Marshal(notification)

	if err := s.producer.Produce(ctx, NotificationTopic, headers, []byte(state.OrchestrationID), notificationBytes); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	l.Info("Workflow paused for human input")
	return nil
}

// CreateNewOrchestration
func (s *SagaCoordinator) CreateNewOrchestration(ctx context.Context, orchestrationID string, headers map[string]string, workflowJSON json.RawMessage) error {
	l := s.logger.With(zap.String("orchestration_id", orchestrationID))

	// Validate we have required headers
	clientID := headers["client_id"]
	if clientID == "" {
		return fmt.Errorf("client_id header is required")
	}

	// Parse the workflow
	var plan models.WorkflowPlan
	if err := json.Unmarshal(workflowJSON, &plan); err != nil {
		return fmt.Errorf("failed to unmarshal workflow: %w", err)
	}

	// Generate orchestration ID for this new orchestration
	newOrchestrationID := uuid.New().String()
	s.logger.With(zap.String("newOrchestration_id", newOrchestrationID))
	correlationID := headers["correlation_id"]

	// Prepare initial data
	initialData := map[string]interface{}{
		"action": "start_orchestration",
		"data": map[string]interface{}{
			"headers":          headers,
			"timestamp":        time.Now().UTC(),
			"message":          "Starting orchestration",
			"orchestration_id": newOrchestrationID,
		},
	}

	// Determine owner agent ID (the orchestrator creating this)
	ownerAgentID := headers["agent_id"]
	if ownerAgentID == "" {
		ownerAgentID = os.Getenv("AGENT_ID")
		if ownerAgentID == "" {
			ownerAgentID = "00000000-0000-0000-0000-000000000001" // Default orchestrator ID
		}
	}

	// Get parent orchestration ID if this is a child orchestration
	ParentOrchestrationID := orchestrationID // The current orchestration becomes the parent
	if ParentOrchestrationID == "" {
		// Try alternative header names
		ParentOrchestrationID = headers["parent_orchestration_id"]
		if ParentOrchestrationID == "" {
			ParentOrchestrationID = headers["parent_orchestration_id"]
		}
	}

	// Store parent context if this is a child orchestration
	if ParentOrchestrationID != "" {
		initialData["__parent_context__"] = map[string]interface{}{
			"orchestration_id":  ParentOrchestrationID,
			"reply_to_topic":    headers["parent_reply_to_topic"],
			"request_id":        headers["request_id"],
			"parent_agent_type": headers["parent_agent_type"],
		}
	}

	initialDataBytes, _ := json.Marshal(initialData)

	l.Info("Creating child orchestration with parent tracking",
		zap.String("DEBUG_COOR_57: child_orchestration_id", newOrchestrationID),
		zap.String("DEBUG_COOR_57: parent_orchestration_id", ParentOrchestrationID),
		zap.String("DEBUG_COOR_57: parent_correlation_id", headers["parent_correlation_id"]))

	// Create the initial state
	repo := NewStateRepository(s.db, s.logger)
	// Create the initial state WITH parent orchestration ID
	if err := repo.CreateInitialState(ctx,
		newOrchestrationID, // New child orchestration ID
		correlationID,
		ownerAgentID,
		ParentOrchestrationID, // Parent's orchestration ID - CRITICAL
		clientID,
		plan,
		initialDataBytes); err != nil {
		return fmt.Errorf("failed to create orchestration state: %w", err)
	}

	l.Info("New orchestration created",
		zap.String("DEBUG_COOR_58: new child orchestration_id", newOrchestrationID),
		zap.String("DEBUG_COOR_58: owner_agent_id", ownerAgentID),
		zap.String("DEBUG_COOR_58: parent_orchestration_id", ParentOrchestrationID),
		zap.String("DEBUG_COOR_58: client_id", clientID),
		zap.String("DEBUG_COOR_58: start_step", plan.StartStep),
		zap.Int("DEBUG_COOR_58: total_steps", len(plan.Steps)))

	// Get the created state using orchestration ID
	state, err := repo.GetState(ctx, newOrchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get created state: %w", err)
	}

	// Ensure headers have required fields for execution
	if headers["fuel_budget"] == "" {
		if state != nil && state.CollectedData != nil {
			if fuel, ok := state.CollectedData["initial_fuel_budget"].(string); ok {
				headers["fuel_budget"] = fuel
			}
		}
	}
	if headers["agent_instance_id"] == "" {
		headers["agent_instance_id"] = "orchestrator-" + newOrchestrationID
	}

	// Add orchestration_id to headers for continueExecution
	headers["orchestration_id"] = newOrchestrationID

	// Store parent correlation ID in collected data for child to access later
	if parentCtx, ok := initialData["__parent_context__"]; ok {
		state.CollectedData["__parent_context__"] = parentCtx
		// Save this update
		if err := repo.UpdateState(ctx, state); err != nil {
			l.Error("Failed to store parent correlation ID", zap.Error(err))
		}
	}

	// Start the workflow execution
	return s.continueExecution(ctx, state, headers)
}

func (s *SagaCoordinator) handleTimeout(ctx context.Context, orchestrationID string, requestID string, timeout time.Duration) {
	time.Sleep(timeout)

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, orchestrationID) // Use orchestrationID
	if err != nil {
		s.logger.Error("Failed to get state for timeout check",
			zap.String("orchestration_id", orchestrationID),
			zap.Error(err))
		return
	}

	// Check if still waiting for this specific request
	for _, awaitedStep := range state.AwaitedSteps {
		if awaitedStep == requestID {
			s.logger.Error("Timeout waiting for agent response",
				zap.String("DEBUG_COOR_42: orchestration_id", orchestrationID),
				zap.String("DEBUG_COOR_42: request_id", requestID),
				zap.Duration("DEBUG_COOR_42: timeout", timeout))

			// Fail the workflow
			s.failWorkflow(ctx, state, fmt.Sprintf("timeout after %v waiting for agent response (request_id: %s)", timeout, requestID))
			return
		}
	}

	// If we get here, the response was already received
	s.logger.Info("Timeout check passed - response already received",
		zap.String("DEBUG_COOR_43: orchestration_id", orchestrationID),
		zap.String("DEBUG_COOR_43: request_id", requestID))
}

func (s *SagaCoordinator) handleChildOrchestrationTimeout(ctx context.Context, parentOrchestrationID string, childCorrelationID string, timeout time.Duration) {
	time.Sleep(timeout)

	// Check if parent is still waiting
	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, parentOrchestrationID)
	if err != nil {
		s.logger.Error("Failed to get parent state for timeout check",
			zap.String("DEBUG_COOR_44: parent_orchestration_id", parentOrchestrationID),
			zap.Error(err))
		return
	}

	// Check if still waiting for this child
	for _, awaitedStep := range state.AwaitedSteps {
		if awaitedStep == childCorrelationID { // Use the actual parameter name
			s.logger.Error("Child orchestration timeout",
				zap.String("DEBUG_COOR_45: parent_orchestration_id", parentOrchestrationID),
				zap.String("DEBUG_COOR_45: child_correlation_id", childCorrelationID),
				zap.Duration("DEBUG_COOR_45: timeout", timeout))

			// Create timeout response
			timeoutResponse := models.TaskResponse{
				Success: false,
				Error:   fmt.Sprintf("Child orchestration timeout after %v", timeout),
				Data: map[string]interface{}{
					"status":         "timeout",
					"correlation_id": childCorrelationID,
				},
			}

			responseBytes, _ := json.Marshal(timeoutResponse)

			// Build proper headers for HandleResponse
			headers := map[string]string{
				"orchestration_id": parentOrchestrationID, // Parent's orchestration ID
				"correlation_id":   state.CorrelationID,   // Parent's correlation ID
				"in_response_to":   childCorrelationID,    // What we're timing out
				"causation_id":     childCorrelationID,    // For backward compatibility
			}

			// Process the timeout as a response
			s.HandleResponse(ctx, headers, responseBytes)
			return
		}
	}

	s.logger.Info("Timeout check passed - child completed in time",
		zap.String("DEBUG_COOR_46: child_correlation_id", childCorrelationID))
}

// continueExecution executes from the current step using the stored plan
func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, headers map[string]string) error {
	l := s.logger.With(
		zap.String("DEBUG_COOR_47: orchestration_id", state.OrchestrationID),
		zap.String("DEBUG_COOR_47: correlation_id", state.CorrelationID),
		zap.String("DEBUG_COOR_47: current_step", state.CurrentStep),
	)

	l.Info("Continuing workflow execution",
		zap.String("DEBUG_COOR_48: current_step", state.CurrentStep),
		zap.Int("DEBUG_COOR_48: total_steps", len(state.WorkflowPlan.Steps)))

	// Get current step from the stored plan
	currentStepConfig, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		errorMsg := fmt.Sprintf("step '%s' not found in plan", state.CurrentStep)
		l.Error("Step not found", zap.String("missing_step", state.CurrentStep))
		return s.failWorkflow(ctx, state, errorMsg)
	}

	l.Info("Executing step",
		zap.String("DEBUG_COOR_49: step", state.CurrentStep),
		zap.String("DEBUG_COOR_49: action", currentStepConfig.Action),
		zap.String("DEBUG_COOR_49: description", currentStepConfig.Description))

	// Record execution start
	execRecord := ExecutionRecord{
		Step:      state.CurrentStep,
		Action:    currentStepConfig.Action,
		StartTime: time.Now().UTC(),
	}

	// Check dependencies
	if !s.dependenciesMet(currentStepConfig.Dependencies, state) {
		l.Info("Dependencies not met, waiting", zap.Strings("dependencies", currentStepConfig.Dependencies))
		execRecord.Result = "skipped"
		execRecord.Error = "dependencies not met"
		s.recordExecution(ctx, state, execRecord)
		return nil
	}

	// Check fuel budget
	fuel, err := governance.GetFuelFromHeader(headers)
	if err != nil {
		l.Error("Failed to get fuel from headers", zap.Error(err))
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to get fuel from headers: %v", err))
	}

	if !s.fuelManager.HasEnoughFuel(fuel, currentStepConfig.Action) {
		execRecord.Result = "failed"
		execRecord.Error = fmt.Sprintf("insufficient fuel: have %d, need %d",
			fuel, s.fuelManager.GetCost(currentStepConfig.Action))
		s.recordExecution(ctx, state, execRecord)
		return s.failWorkflow(ctx, state, execRecord.Error)
	}

	// Deduct fuel
	remainingFuel := s.fuelManager.DeductFuel(fuel, currentStepConfig.Action)
	governance.SetFuelHeader(headers, remainingFuel)
	l.Info("Fuel deducted", zap.Int("remaining_fuel", remainingFuel))

	// Execute based on action type
	var execErr error
	switch {
	case currentStepConfig.Action == "complete_workflow":
		l.Info("Completing workflow")
		l.Info("DEBUG_COOR_50: Completing workflow",
			zap.Any("DEBUG_COOR_49: step", state),
			zap.Any("DEBUG_COOR_49: step", currentStepConfig),
			zap.Any("DEBUG_COOR_49: step", headers),
		)

		_, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else {
			execErr = s.completeWorkflow(ctx, state)
		}

	case currentStepConfig.Action == "call_agent":
		l.Info("Executing call_agent with wait-for-response")
		result, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else if resultMap, ok := result.(map[string]interface{}); ok {
			if messageSent, ok := resultMap["message_sent"].(bool); ok && messageSent {
				requestID, _ := resultMap["request_id"].(string)

				// Update state to wait for response
				state.Status = StatusAwaitingResponses
				state.AwaitedSteps = []string{requestID}

				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					execErr = fmt.Errorf("failed to update state for waiting: %w", err)
				} else {
					l.Info("Waiting for agent response",
						zap.String("request_id", requestID))

					// Set up timeout
					timeout := 60 * time.Second
					if currentStepConfig.Timeout > 0 {
						timeout = currentStepConfig.Timeout
					}
					go s.handleTimeout(ctx, state.OrchestrationID, requestID, timeout)
					return nil // Exit and wait for response
				}
			}
		}

	case currentStepConfig.Action == "start_orchestration":
		l.Info("Executing start_orchestration with wait-for-response")
		result, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else if resultMap, ok := result.(map[string]interface{}); ok {
			if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
				if requestID, ok := resultMap["new_correlation_id"].(string); ok {
					state.CollectedData[state.CurrentStep] = result
					state.Status = StatusAwaitingResponses
					state.AwaitedSteps = []string{requestID}
					state.ExecutionMetadata.Checkpoints[state.CurrentStep] = time.Now().UTC()

					repo := NewStateRepository(s.db, s.logger)
					if err := repo.UpdateState(ctx, state); err != nil {
						execErr = fmt.Errorf("failed to update state for waiting: %w", err)
					} else {
						l.Info("Waiting for child orchestration",
							zap.String("child_correlation_id", requestID))

						timeout := 5 * time.Minute
						if currentStepConfig.Timeout > 0 {
							timeout = currentStepConfig.Timeout
						}
						go s.handleChildOrchestrationTimeout(ctx, state.OrchestrationID, requestID, timeout)
						return nil
					}
				}
			}
		}

	case currentStepConfig.Action == "fan_out":
		l.Info("Handling fan-out")
		execErr = s.handleFanOut(ctx, headers, currentStepConfig, state)

	case currentStepConfig.Action == "pause_for_human_input":
		l.Info("Pausing for human input")
		execErr = s.handlePauseForHumanInput(ctx, headers, currentStepConfig, state)

	case isLocalAction(currentStepConfig.Action):
		l.Info("Executing local action", zap.String("action", currentStepConfig.Action))
		_, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else {
			if currentStepConfig.NextStep != "" {
				l.Info("Moving to next step", zap.String("next_step", currentStepConfig.NextStep))
				state.CurrentStep = currentStepConfig.NextStep

				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					execErr = fmt.Errorf("failed to update state: %w", err)
				} else {
					return s.continueExecution(ctx, state, headers)
				}
			}
		}

	case currentStepConfig.Topic != "":
		l.Info("Executing remote action", zap.String("topic", currentStepConfig.Topic))
		execErr = s.executeRemoteAction(ctx, state, currentStepConfig, headers)

	default:
		errorMsg := fmt.Sprintf("unknown action: %s", currentStepConfig.Action)
		l.Error("Unknown action", zap.String("action", currentStepConfig.Action))
		execErr = fmt.Errorf(errorMsg)
	}

	// Record execution result
	endTime := time.Now().UTC()
	execRecord.EndTime = &endTime
	if execErr != nil {
		execRecord.Result = "failed"
		execRecord.Error = execErr.Error()
		l.Error("Step execution failed",
			zap.String("step", state.CurrentStep),
			zap.Error(execErr))
	} else {
		execRecord.Result = "success"
		l.Info("Step execution succeeded", zap.String("step", state.CurrentStep))
	}
	s.recordExecution(ctx, state, execRecord)

	return execErr
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
