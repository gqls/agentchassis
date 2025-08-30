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
	"github.com/gqls/agentchassis/platform/orchestration/types"
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
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	correlationID := headers["correlation_id"]
	l := s.logger.With(zap.String("correlation_id", correlationID))

	l.Info("ExecuteWorkflow called",
		zap.String("DEBUG_COOR_24: start_step", plan.StartStep),
		zap.Int("DEBUG_COOR_24: total_steps", len(plan.Steps)))

	clientID := headers["client_id"]
	if clientID == "" {
		l.Error("DEBUG_COOR_24: client_id header is required to execute a workflow")
		return fmt.Errorf("client_id header is required to execute a workflow")
	}

	// Get or create state with the plan
	state, orchestrationID, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData, headers)
	if err != nil {
		l.Error("DEBUG_COOR_24: Failed to get or create state", zap.Error(err))
		return err
	}

	// Set orchestration_id in headers for all subsequent operations
	headers["orchestration_id"] = orchestrationID
	headers["owner_agent_id"] = state.OwnerAgentID

	l.Info("Workflow state retrieved",
		zap.String("DEBUG_COOR_24: orchestration_id", orchestrationID),
		zap.String("DEBUG_COOR_24: status", string(state.Status)),
		zap.String("DEBUG_COOR_24: current_step", state.CurrentStep))

	// Check if workflow is already complete
	if state.Status == StatusCompleted || state.Status == StatusFailed {
		l.Info("Workflow already finished",
			zap.String("DEBUG_COOR_24: orchestration_id", orchestrationID),
			zap.String("DEBUG_COOR_24: status", string(state.Status)))
		return nil
	}

	// Continue execution from current step
	return s.continueExecution(ctx, state, headers)
}

// getOrCreateState returns orchestrationID explicitly
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte, headers map[string]string) (*OrchestrationState, string, error) {
	repo := NewStateRepository(s.db, s.logger)

	// Priority 1: If we have an explicit orchestration_id, use it
	if orchestrationID := headers["orchestration_id"]; orchestrationID != "" {
		state, err := repo.GetState(ctx, orchestrationID)
		if err == nil {
			s.logger.Info("Found existing state by orchestration_id",
				zap.String("orchestration_id", orchestrationID))
			return state, orchestrationID, nil
		}

		// If we have an orchestration_id but no state, this is likely a new child orchestration
		// Don't fall back to correlation_id lookup - just create new state with this ID
		if headers["parent_orchestration_id"] != "" {
			s.logger.Info("Creating child orchestration with explicit ID",
				zap.String("orchestration_id", orchestrationID),
				zap.String("parent_orchestration_id", headers["parent_orchestration_id"]))

			// Use the provided orchestration_id instead of generating a new one
			ownerAgentID := s.determineOwnerAgentID(headers)

			if err := repo.CreateInitialState(ctx, orchestrationID, correlationID, ownerAgentID,
				headers["parent_orchestration_id"], clientID, plan, initialData); err != nil {
				return nil, "", err
			}

			state, err = repo.GetState(ctx, orchestrationID)
			if err != nil {
				return nil, "", err
			}
			return state, orchestrationID, nil
		}
	}

	// Priority 2: For root orchestrations only, try correlation_id lookup
	// Never do this for child orchestrations
	if headers["parent_orchestration_id"] == "" {
		state, err := repo.GetStateByCorrelation(ctx, correlationID)
		if err == nil {
			s.logger.Info("Found existing root orchestration by correlation_id",
				zap.String("correlation_id", correlationID),
				zap.String("orchestration_id", state.OrchestrationID))
			return state, state.OrchestrationID, nil
		}
	}

	// Priority 3: Create new orchestration
	newOrchestrationID := uuid.New().String()
	if providedID := headers["orchestration_id"]; providedID != "" {
		// Respect the provided ID if it exists
		newOrchestrationID = providedID
	}

	ownerAgentID := s.determineOwnerAgentID(headers)
	parentOrchestrationID := headers["parent_orchestration_id"]

	s.logger.Info("Creating new orchestration",
		zap.String("orchestration_id", newOrchestrationID),
		zap.String("correlation_id", correlationID),
		zap.String("parent_orchestration_id", parentOrchestrationID),
		zap.Bool("is_child", parentOrchestrationID != ""))

	if err := repo.CreateInitialState(ctx, newOrchestrationID, correlationID, ownerAgentID,
		parentOrchestrationID, clientID, plan, initialData); err != nil {
		return nil, "", err
	}

	state, err := repo.GetState(ctx, newOrchestrationID)
	if err != nil {
		return nil, "", err
	}

	return state, newOrchestrationID, nil
}

func (s *SagaCoordinator) determineOwnerAgentID(headers map[string]string) string {
	if ownerAgentID := headers["agent_id"]; ownerAgentID != "" {
		return ownerAgentID
	}
	if ownerAgentID := os.Getenv("AGENT_ID"); ownerAgentID != "" {
		return ownerAgentID
	}
	return "00000000-0000-0000-0000-000000000001"
}

// executeLocalAction handles actions that run within the orchestrator
func (s *SagaCoordinator) executeLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) (interface{}, error) {
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Create minimal context
		execCtx = &types.ExecutionContext{
			OrchestrationID: state.OrchestrationID,
			CorrelationID:   state.CorrelationID,
			ClientID:        state.ClientID,
			OwnerAgentID:    state.OwnerAgentID,
			FuelBudget:      1000,
		}
	}

	contextLogger := s.logger.With(execCtx.LogContext()...)

	contextLogger.Info("TRACE: executeLocalAction entry",
		zap.String("DEBUG_COOR_26: action", step.Action),
		zap.String("DEBUG_COOR_26: step", state.CurrentStep),
		zap.Int("DEBUG_COOR_26: fuel_available", execCtx.FuelBudget))

	// For start_orchestration, log the child creation
	if step.Action == "start_orchestration" {
		headers["parent_orchestration_id"] = execCtx.OrchestrationID
		contextLogger.Info("TRACE: Creating child orchestration",
			zap.String("DEBUG_COOR_26: parent_orch", execCtx.OrchestrationID),
			zap.Int("DEBUG_COOR_26: parent_fuel", execCtx.FuelBudget))
	}

	l := s.logger.With(
		zap.String("DEBUG_COOR_26b: orchestration_id", execCtx.OrchestrationID),
		zap.String("DEBUG_COOR_26b: correlation_id", execCtx.CorrelationID),
		zap.String("DEBUG_COOR_26b: action", step.Action),
	)

	handler, ok := actionRegistry[step.Action]
	if !ok {
		return nil, fmt.Errorf("local action '%s' not found in registry", step.Action)
	}

	// For call_agent actions, prepare child context
	if step.Action == "call_agent" {
		// Use the actual agent type, not "orchestrator"
		execCtx.ReplyToTopic = fmt.Sprintf("system.agent.%s.responses", execCtx.OwnerAgentType)
		l.Info("Prepared context for call_agent",
			zap.String("DEBUG_COOR_26c: reply_to_topic", execCtx.ReplyToTopic),
			zap.String("DEBUG_COOR_26c: owner_agent_type", execCtx.OwnerAgentType))
	}

	// For start_orchestration, mark parent relationship
	if step.Action == "start_orchestration" {
		// Headers will be used by the action to create child context
		headers["parent_orchestration_id"] = execCtx.OrchestrationID
		l.Info("DEBUG_COOR_26d: Prepared context for start_orchestration",
			zap.String("DEBUG_COOR_26d: parent_orchestration_id", execCtx.OrchestrationID))
	}

	// Convert back to headers for action params (for backward compatibility)
	actionHeaders := execCtx.ToHeaders()

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
		AgentType:       execCtx.FromAgentType,
		CurrentStep:     state.CurrentStep,
	}

	// Execute the action
	result, err := handler(ctx, params)
	if err != nil {
		l.Error("DEBUG_COOR_26: Local action failed", zap.Error(err))
		return nil, fmt.Errorf("local action failed: %w", err)
	}

	// Handle await_response results
	if resultMap, ok := result.(map[string]interface{}); ok {
		if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
			var requestID string

			if step.Action == "start_orchestration" {
				if id, ok := resultMap["new_correlation_id"].(string); ok {
					requestID = id
				}
			} else if id, ok := resultMap["request_id"].(string); ok {
				requestID = id
			}

			if requestID != "" {
				l.Info("Local action requires waiting for response",
					zap.String("DEBUG_COOR_26: request_id", requestID),
					zap.String("DEBUG_COOR_26: action", step.Action))

				state.CollectedData[state.CurrentStep] = result
				state.Status = StatusAwaitingResponses
				state.AwaitedSteps = []string{requestID}
				state.ExecutionMetadata.CompletedSteps++
				state.ExecutionMetadata.Checkpoints[step.Action] = time.Now().UTC()

				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					return result, fmt.Errorf("failed to update state for waiting: %w", err)
				}

				if step.Action == "start_orchestration" {
					timeout := 5 * time.Minute
					if step.Timeout > 0 {
						timeout = step.Timeout
					}
					go s.handleChildOrchestrationTimeout(ctx, state.OrchestrationID, requestID, timeout)
				}

				return result, nil
			}
		}
	}

	// Store result and update metadata
	state.CollectedData[state.CurrentStep] = result
	state.ExecutionMetadata.CompletedSteps++
	state.ExecutionMetadata.Checkpoints[step.Action] = time.Now().UTC()

	// Move to next step if defined
	if step.NextStep != "" {
		l.Info("Moving to next step",
			zap.String("DEBUG_COOR_28: next_step", step.NextStep),
			zap.String("DEBUG_COOR_28: orchestration_id", state.OrchestrationID))

		state.CurrentStep = step.NextStep

		repo := NewStateRepository(s.db, s.logger)
		if err := repo.UpdateState(ctx, state); err != nil {
			return result, fmt.Errorf("failed to update state: %w", err)
		}

		// Continue execution with updated headers
		return result, s.continueExecution(ctx, state, actionHeaders)
	}

	return result, nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	// Create ExecutionContext
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		return fmt.Errorf("invalid execution context: %w", err)
	}

	l := s.logger.With(
		zap.String("DEBUG_COOR_29: correlation_id", execCtx.CorrelationID),
		zap.String("DEBUG_COOR_29: orchestration_id", execCtx.OrchestrationID),
		zap.String("DEBUG_COOR_29: action", step.Action),
		zap.String("DEBUG_COOR_29: topic", step.Topic),
	)

	// Create new context for the remote call
	remoteCtx := &types.ExecutionContext{
		CorrelationID:   execCtx.CorrelationID,
		ClientID:        execCtx.ClientID,
		OrchestrationID: execCtx.OrchestrationID,
		MessageID:       uuid.New().String(),
		RequestID:       uuid.New().String(),
		MessageType:     "request",
		OwnerAgentID:    execCtx.OwnerAgentID,
		FromAgentID:     execCtx.OwnerAgentID,
		FromAgentType:   execCtx.FromAgentType,
		ReplyToTopic:    fmt.Sprintf("system.agent.%s.responses", execCtx.FromAgentType),
		FuelBudget:      execCtx.FuelBudget,
		Timestamp:       time.Now().UTC(),
		Version:         "2.0",
	}

	// Prepare the message
	payload := models.TaskRequest{
		Action: step.Action,
		Data:   state.CollectedData,
	}
	payloadBytes, _ := json.Marshal(payload)

	l.Info("Sending remote action",
		zap.String("DEBUG_COOR_29: request_id", remoteCtx.RequestID),
		zap.String("DEBUG_COOR_29: topic", step.Topic))

	// Send the message
	if err := s.producer.Produce(ctx, step.Topic, remoteCtx.ToHeaders(),
		[]byte(remoteCtx.CorrelationID), payloadBytes); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Update state to await response
	state.Status = StatusAwaitingResponses
	state.CurrentStep = step.NextStep
	state.AwaitedSteps = []string{remoteCtx.RequestID}

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	l.Info("Remote action initiated",
		zap.String("DEBUG_COOR_29: request_id", remoteCtx.RequestID))

	return nil
}

// completeWorkflow - uses ExecutionContext for parent notification
func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState, headers map[string]string) error {

	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Fallback: create minimal context from state
		execCtx = &types.ExecutionContext{
			OrchestrationID:       state.OrchestrationID,
			CorrelationID:         state.CorrelationID,
			ClientID:              state.ClientID,
			OwnerAgentID:          state.OwnerAgentID,
			ParentOrchestrationID: state.ParentOrchestrationID,
			FuelBudget:            0, // Will extract from collected_data
		}
	}

	// Extract fuel from collected_data
	if fuel, ok := state.CollectedData["__fuel_budget__"].(float64); ok {
		execCtx.FuelBudget = int(fuel)
	}

	contextLogger := s.logger.With(execCtx.LogContext()...)

	l := contextLogger.With(
		zap.String("DEBUG_COOR_30: correlation_id", state.CorrelationID),
		zap.String("DEBUG_COOR_30: orchestration_id", state.OrchestrationID),
		zap.String("DEBUG_COOR_30: parent_orchestration_id", state.ParentOrchestrationID),
	)

	l.Info("Completing workflow",
		zap.Int("DEBUG_COOR_30: completed_steps", state.ExecutionMetadata.CompletedSteps),
		zap.Int("DEBUG_COOR_30: total_steps", len(state.WorkflowPlan.Steps)))

	state.Status = StatusCompleted
	finalResult, _ := json.Marshal(state.CollectedData)
	state.FinalResult = finalResult

	now := time.Now().UTC()
	state.ExecutionMetadata.EndTime = &now
	state.ExecutionMetadata.CompletedSteps++

	repo := NewStateRepository(s.db, s.logger)
	err = repo.UpdateState(ctx, state)
	if err != nil {
		l.Error("DEBUG_COOR_30: Failed to update workflow completion state", zap.Error(err))
		return err
	}

	// Check if this is a child orchestration needing to notify parent
	if state.ParentOrchestrationID != "" && state.CollectedData != nil {
		// Look for parent context that was stored during creation
		if parentCtx, ok := state.CollectedData["__parent_context__"].(map[string]interface{}); ok {
			// Create response targeting the PARENT's orchestration
			responseCtx := &types.ExecutionContext{
				CorrelationID:   state.CorrelationID,
				ClientID:        state.ClientID,
				OrchestrationID: state.ParentOrchestrationID, // PARENT's orchestration
				MessageID:       uuid.New().String(),
				RequestID:       fmt.Sprintf("%v", parentCtx["request_id"]),
				MessageType:     "response",
				InResponseTo:    fmt.Sprintf("%v", parentCtx["request_id"]),
				FromAgentID:     state.OwnerAgentID,
				OwnerAgentID:    state.ParentOrchestrationID, // Parent owns this orchestration
				ReplyToTopic:    fmt.Sprintf("%v", parentCtx["reply_to_topic"]),
				FuelBudget:      extractFuelBudget(state),
				Timestamp:       time.Now().UTC(),
				Version:         "2.0",
			}

			contextLogger.Info("TRACE: Child notifying parent of completion",
				zap.String("child_orch", state.OrchestrationID),
				zap.String("parent_orch", state.ParentOrchestrationID),
				zap.String("response_orch_id", responseCtx.OrchestrationID),
				zap.Int("fuel_returning", extractFuelBudget(state)),
				zap.String("reply_topic", responseCtx.ReplyToTopic))

			// Send completion to parent
			parentResponse := models.TaskResponse{
				Success: true,
				Data: map[string]interface{}{
					"status":         "completed",
					"correlation_id": state.CorrelationID,
					"final_result":   state.CollectedData,
				},
			}

			responseBytes, _ := json.Marshal(parentResponse)
			err := s.producer.Produce(ctx, responseCtx.ReplyToTopic, responseCtx.ToHeaders(),
				[]byte(state.CorrelationID), responseBytes)

			if err != nil {
				s.logger.Error("Failed to notify parent", zap.Error(err))
			} else {
				s.logger.Info("Parent notified of completion",
					zap.String("parent_orchestration_id", state.ParentOrchestrationID),
					zap.String("response_orchestration_id", responseCtx.OrchestrationID))
			}
		}
	}

	return nil
}

func extractFuelBudget(state *OrchestrationState) int {
	if fuel, ok := state.CollectedData["__fuel_budget__"].(float64); ok {
		return int(fuel)
	}
	return 0
}

// failWorkflow marks the workflow as failed
func (s *SagaCoordinator) failWorkflow(ctx context.Context, state *OrchestrationState, errorMsg string) error {
	l := s.logger.With(
		zap.String("DEBUG_COOR_32: orchestration_id", state.OrchestrationID),
		zap.String("DEBUG_COOR_32: correlation_id", state.CorrelationID))

	l.Error("Failing workflow",
		zap.String("error", errorMsg))

	state.Status = StatusFailed
	state.Error = errorMsg

	now := time.Now().UTC()
	state.ExecutionMetadata.EndTime = &now

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		l.Error("Failed to update state to failed", zap.Error(err))
		return fmt.Errorf("failed to update state to failed: %w", err)
	}

	// Notify parent if this is a child orchestration
	if state.ParentOrchestrationID != "" {
		parentState, err := repo.GetState(ctx, state.ParentOrchestrationID)
		if err == nil && parentState != nil {
			// Create failure response context
			responseCtx := &types.ExecutionContext{
				CorrelationID:   parentState.CorrelationID,
				OrchestrationID: state.ParentOrchestrationID,
				MessageID:       uuid.New().String(),
				MessageType:     "response",
				InResponseTo:    state.CorrelationID,
				FromAgentID:     state.OwnerAgentID,
				ClientID:        state.ClientID,
				Timestamp:       time.Now().UTC(),
				Version:         "2.0",
			}

			failureResponse := models.TaskResponse{
				Success: false,
				Error:   errorMsg,
				Data: map[string]interface{}{
					"status":         "failed",
					"correlation_id": state.CorrelationID,
					"error":          errorMsg,
				},
			}

			responseBytes, _ := json.Marshal(failureResponse)
			parentResponseTopic := fmt.Sprintf("system.agent.%s.responses",
				parentState.CollectedData["agent_type"])

			if err := s.producer.Produce(ctx,
				parentResponseTopic,
				responseCtx.ToHeaders(),
				[]byte(parentState.CorrelationID),
				responseBytes); err != nil {
				l.Error("Failed to notify parent of failure", zap.Error(err))
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

// HandleResponse processes responses and continues workflow with ExecutionContext
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		s.logger.Error("Failed to create ExecutionContext from headers", zap.Error(err))
		return err
	}

	contextLogger := s.logger.With(execCtx.LogContext()...)

	contextLogger.Info("DEBUG_COOR_33: HandleResponse received",
		zap.String("in_response_to", execCtx.InResponseTo))

	// Create a repository instance to interact with the database
	repo := NewStateRepository(s.db, s.logger)

	// Find the parent orchestration state that is waiting for this response
	state, err := repo.FindByAwaitedRequestID(ctx, execCtx.InResponseTo)
	if err != nil {
		contextLogger.Error("DEBUG_COOR_33: Failed to find orchestration state for response",
			zap.String("awaited_request_id", execCtx.InResponseTo),
			zap.Error(err))
		return err
	}

	// Parse the incoming response
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(response, &taskResponse); err != nil {
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to unmarshal response: %v", err))
	}

	// Process the response data and update the state
	if err := s.processResponseData(ctx, state, execCtx, taskResponse); err != nil {
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to process response data: %v", err))
	}

	// If no more steps are awaited, continue the workflow
	if len(state.AwaitedSteps) == 0 {
		return s.continueWorkflowAfterResponse(ctx, state, execCtx)
	}

	// If still waiting for other responses, just save the updated state
	contextLogger.Info("Still awaiting other responses",
		zap.Int("remaining_count", len(state.AwaitedSteps)))
	return repo.UpdateState(ctx, state)
}

// ResponseIDs holds all the different IDs from a response
type ResponseIDs struct {
	OrchestrationID       string // The orchestration that sent this response
	CorrelationID         string // Business transaction ID
	InResponseTo          string // What request this is responding to
	ParentOrchestrationID string // If this is from a child, who's the parent
}

// extractResponseIDs clearly extracts all IDs from headers
func (s *SagaCoordinator) extractResponseIDs(headers map[string]string) (*ResponseIDs, error) {
	ids := &ResponseIDs{
		OrchestrationID:       headers["orchestration_id"],
		CorrelationID:         headers["correlation_id"],
		InResponseTo:          headers["in_response_to"],
		ParentOrchestrationID: headers["parent_orchestration_id"],
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
		zap.String("DEBUG_COOR_34: orchestration_id", ids.OrchestrationID),
		zap.String("DEBUG_COOR_34: correlation_id", ids.CorrelationID),
		zap.String("DEBUG_COOR_34: in_response_to", ids.InResponseTo),
		zap.String("DEBUG_COOR_34: parent_orch_id", ids.ParentOrchestrationID))

	return ids, nil
}

// findTargetOrchestration determines which orchestration should handle this response
func (s *SagaCoordinator) findTargetOrchestration(ctx context.Context, ids *ResponseIDs, headers map[string]string) (*OrchestrationState, error) {
	repo := NewStateRepository(s.db, s.logger)

	// The target is usually in the "orchestration_id" header
	// This should be the parent's orchestration_id for child responses
	targetOrchestrationID := headers["orchestration_id"]

	// Try to get by orchestration_id
	if targetOrchestrationID != "" {
		state, err := repo.GetState(ctx, targetOrchestrationID)
		if err == nil {
			s.logger.Info("Found state by orchestration_id",
				zap.String("orchestration_id", targetOrchestrationID))
			return state, nil
		}
	}

	return nil, fmt.Errorf("no state found for orchestration_id=%s", targetOrchestrationID)
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

// processResponseData stores the response and updates orchestration state updated to use ExecutionContext
func (s *SagaCoordinator) processResponseData(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext, response models.TaskResponse) error {
	// Store response under the request ID we're responding to
	responseKey := execCtx.InResponseTo

	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}

	state.CollectedData[responseKey] = response.Data

	// Remove from awaited steps
	updatedAwaitedSteps := []string{}
	for _, awaitedID := range state.AwaitedSteps {
		if awaitedID != execCtx.InResponseTo {
			updatedAwaitedSteps = append(updatedAwaitedSteps, awaitedID)
		}
	}
	state.AwaitedSteps = updatedAwaitedSteps

	return nil
}

// continueWorkflowAfterResponse continues execution after all responses received
func (s *SagaCoordinator) continueWorkflowAfterResponse(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext) error {
	s.logger.Info("All responses received, continuing workflow",
		zap.String("DEBUG_COOR_35: orchestration_id", state.OrchestrationID),
		zap.String("DEBUG_COOR_35: current_step", state.CurrentStep))

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

	// Convert ExecutionContext back to headers for continuation
	continueHeaders := execCtx.ToHeaders()

	// Ensure critical fields are from state
	continueHeaders["DEBUG_COOR_36: orchestration_id"] = state.OrchestrationID
	continueHeaders["DEBUG_COOR_36: correlation_id"] = state.CorrelationID
	continueHeaders["DEBUG_COOR_36: client_id"] = state.ClientID
	continueHeaders["DEBUG_COOR_36: owner_agent_id"] = state.OwnerAgentID

	return s.continueExecution(ctx, state, continueHeaders)
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
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		return fmt.Errorf("invalid execution context: %w", err)
	}

	l := s.logger.With(
		zap.String("DEBUG_COOR_37: correlation_id", execCtx.CorrelationID),
		zap.String("DEBUG_COOR_37: orchestration_id", execCtx.OrchestrationID))

	awaitedSteps := make([]string, 0, len(step.SubTasks))

	for _, subTask := range step.SubTasks {
		// Create new context for each subtask
		subCtx := &types.ExecutionContext{
			CorrelationID:   execCtx.CorrelationID,
			ClientID:        execCtx.ClientID,
			OrchestrationID: execCtx.OrchestrationID,
			MessageID:       uuid.New().String(),
			RequestID:       uuid.New().String(),
			MessageType:     "request",
			OwnerAgentID:    execCtx.OwnerAgentID,
			FromAgentID:     execCtx.OwnerAgentID,
			FromAgentType:   execCtx.FromAgentType,
			ReplyToTopic:    fmt.Sprintf("system.agent.%s.responses", execCtx.FromAgentType),
			FuelBudget:      execCtx.FuelBudget,
			Timestamp:       time.Now().UTC(),
			Version:         "2.0",
		}

		payload := models.TaskRequest{
			Action: subTask.StepName,
			Data:   state.CollectedData,
		}
		payloadBytes, _ := json.Marshal(payload)

		if err := s.producer.Produce(ctx, subTask.Topic, subCtx.ToHeaders(),
			[]byte(state.OrchestrationID), payloadBytes); err != nil {
			return fmt.Errorf("failed to produce fan-out message: %w", err)
		}

		awaitedSteps = append(awaitedSteps, subCtx.RequestID)
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
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		return fmt.Errorf("invalid execution context: %w", err)
	}

	l := s.logger.With(
		zap.String("correlation_id", execCtx.CorrelationID),
		zap.String("orchestration_id", execCtx.OrchestrationID))

	state.Status = StatusPausedForHuman
	state.CurrentStep = step.NextStep

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Send notification using ExecutionContext
	notification := map[string]interface{}{
		"event_type":       "WORKFLOW_PAUSED_FOR_APPROVAL",
		"correlation_id":   execCtx.CorrelationID,
		"orchestration_id": execCtx.OrchestrationID,
		"client_id":        execCtx.ClientID,
		"message":          fmt.Sprintf("Step '%s' requires your approval", step.Description),
		"data_for_review":  state.CollectedData,
	}
	notificationBytes, _ := json.Marshal(notification)

	notificationCtx := &types.ExecutionContext{
		CorrelationID:   execCtx.CorrelationID,
		OrchestrationID: execCtx.OrchestrationID,
		ClientID:        execCtx.ClientID,
		MessageID:       uuid.New().String(),
		MessageType:     "notification",
		FromAgentID:     execCtx.OwnerAgentID,
		FromAgentType:   execCtx.FromAgentType,
		Timestamp:       time.Now().UTC(),
		Version:         "2.0",
	}

	if err := s.producer.Produce(ctx, NotificationTopic, notificationCtx.ToHeaders(),
		[]byte(state.OrchestrationID), notificationBytes); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	l.Info("Workflow paused for human input")
	return nil
}

// CreateNewOrchestration creates a new orchestration instance
func (s *SagaCoordinator) CreateNewOrchestration(ctx context.Context, orchestrationID string, headers map[string]string, initialData json.RawMessage) error {
	// Parse the incoming data
	var data map[string]interface{}
	if err := json.Unmarshal(initialData, &data); err != nil {
		s.logger.Error("Failed to unmarshal initial data",
			zap.Error(err),
			zap.String("raw_data", string(initialData)))
		return fmt.Errorf("failed to unmarshal initial data: %w", err)
	}

	// Extract workflow
	var plan models.WorkflowPlan
	if workflowData, ok := data["workflow"].(map[string]interface{}); ok {
		// It's already a map, convert to WorkflowPlan
		planBytes, _ := json.Marshal(workflowData)
		if err := json.Unmarshal(planBytes, &plan); err != nil {
			s.logger.Error("Failed to parse workflow",
				zap.Error(err),
				zap.Any("DEBUG_COOR_38: workflow_data", workflowData))
			return fmt.Errorf("failed to parse workflow: %w", err)
		}
	} else {
		s.logger.Error("No workflow found in initial data",
			zap.Any("data_keys", getMapKeys(data)))
		return fmt.Errorf("no workflow found in initial data")
	}

	// Validate workflow has required fields
	if plan.StartStep == "" {
		return fmt.Errorf("workflow plan missing start_step")
	}
	if len(plan.Steps) == 0 {
		return fmt.Errorf("workflow plan has no steps")
	}

	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		return fmt.Errorf("invalid execution context: %w", err)
	}

	// Validate required fields
	if execCtx.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	l := s.logger.With(
		zap.String("DEBUG_COOR_39: orchestration_id", execCtx.OrchestrationID),
		zap.String("DEBUG_COOR_39: parent_orchestration_id", execCtx.ParentOrchestrationID),
		zap.String("DEBUG_COOR_39: start_step", plan.StartStep),
		zap.Int("DEBUG_COOR_39: step_count", len(plan.Steps)))

	l.Info("Creating orchestration with workflow")

	// Merge initial_data from input with execution context
	collectedData := make(map[string]interface{})

	// Add any initial_data that was passed in
	if inputData, ok := data["initial_data"].(map[string]interface{}); ok {
		for k, v := range inputData {
			collectedData[k] = v
		}
	}

	// Add execution context
	collectedData["__execution_context__"] = execCtx
	collectedData["timestamp"] = time.Now().UTC()
	collectedData["orchestration_id"] = execCtx.OrchestrationID

	// If this is a child, preserve parent context
	if execCtx.IsChildOrchestration() {
		if parentCtx, ok := collectedData["__parent_context__"].(map[string]interface{}); ok {
			// Already has parent context from StartOrchestrationAction
			l.Info("Parent context preserved",
				zap.Any("DEBUG_COOR_40: parent_context", parentCtx))
		}
	}

	collectedDataBytes, _ := json.Marshal(collectedData)

	// Create the initial state
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.CreateInitialState(ctx,
		execCtx.OrchestrationID,
		execCtx.CorrelationID,
		execCtx.OwnerAgentID,
		execCtx.ParentOrchestrationID,
		execCtx.ClientID,
		plan,
		collectedDataBytes); err != nil {
		return fmt.Errorf("failed to create orchestration state: %w", err)
	}

	l.Info("New orchestration created",
		zap.String("DEBUG_COOR_41: owner_agent_id", execCtx.OwnerAgentID),
		zap.String("DEBUG_COOR_41: client_id", execCtx.ClientID))

	// Get the created state
	state, err := repo.GetState(ctx, execCtx.OrchestrationID)
	if err != nil {
		return fmt.Errorf("failed to get created state: %w", err)
	}

	// Start execution with ExecutionContext headers
	return s.continueExecution(ctx, state, execCtx.ToHeaders())
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
// continueExecution executes from the current step using the stored plan
func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, headers map[string]string) error {
	l := s.logger.With(
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("correlation_id", state.CorrelationID),
		zap.String("current_step", state.CurrentStep),
	)

	l.Info("Continuing workflow execution",
		zap.String("current_step", state.CurrentStep),
		zap.Int("total_steps", len(state.WorkflowPlan.Steps)))

	// Handle fuel management
	fuelBudget, err := s.manageFuel(ctx, state, headers)
	if err != nil {
		return s.failWorkflow(ctx, state, err.Error())
	}

	// Get and validate current step
	currentStepConfig, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		errorMsg := fmt.Sprintf("step '%s' not found in plan", state.CurrentStep)
		l.Error("Step not found", zap.String("missing_step", state.CurrentStep))
		return s.failWorkflow(ctx, state, errorMsg)
	}

	// Check dependencies
	if !s.dependenciesMet(currentStepConfig.Dependencies, state) {
		return s.skipStep(ctx, state, "dependencies not met")
	}

	// Execute the step
	return s.executeStep(ctx, state, currentStepConfig, headers, fuelBudget)
}

// manageFuel handles all fuel-related logic in one place
func (s *SagaCoordinator) manageFuel(ctx context.Context, state *OrchestrationState, headers map[string]string) (int, error) {
	// Get fuel from stored state or headers (single source of truth)
	fuelBudget := 1000 // default

	// Check collected_data first (persisted state)
	if fuel, ok := state.CollectedData["__fuel_budget__"].(float64); ok {
		fuelBudget = int(fuel)
	} else if fuelStr := headers["fuel_budget"]; fuelStr != "" {
		// Fall back to headers if not in state
		fmt.Sscanf(fuelStr, "%d", &fuelBudget)
	}

	// Check if we have enough fuel
	if fuelBudget < 1 {
		return 0, fmt.Errorf("insufficient fuel: have %d, need 1", fuelBudget)
	}

	// Deduct fuel ONCE
	fuelBudget--

	// Store updated fuel in both state and headers
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}
	state.CollectedData["__fuel_budget__"] = fuelBudget
	headers["fuel_budget"] = fmt.Sprintf("%d", fuelBudget)

	s.logger.Info("Fuel managed",
		zap.Int("remaining_fuel", fuelBudget))

	return fuelBudget, nil
}

// executeStep handles the actual step execution based on action type
func (s *SagaCoordinator) executeStep(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string, fuelBudget int) error {
	s.logger.With(
		zap.String("step", state.CurrentStep),
		zap.String("action", step.Action))

	// Record execution start
	execRecord := s.startExecutionRecord(state.CurrentStep, step.Action)

	// Route to appropriate handler based on action type
	var execErr error
	switch step.Action {
	case "complete_workflow":
		execErr = s.handleCompleteWorkflow(ctx, state, step, headers)

	case "call_agent":
		execErr = s.handleCallAgent(ctx, state, step, headers)

	case "start_orchestration":
		execErr = s.handleStartOrchestration(ctx, state, step, headers)

	case "fan_out":
		execErr = s.handleFanOut(ctx, headers, step, state)

	case "pause_for_human_input":
		execErr = s.handlePauseForHumanInput(ctx, headers, step, state)

	default:
		if isLocalAction(step.Action) {
			execErr = s.handleLocalAction(ctx, state, step, headers)
		} else if step.Topic != "" {
			execErr = s.executeRemoteAction(ctx, state, step, headers)
		} else {
			execErr = fmt.Errorf("unknown action: %s", step.Action)
		}
	}

	// Record execution result
	s.finishExecutionRecord(ctx, state, execRecord, execErr)

	return execErr
}

// handleCompleteWorkflow handles workflow completion
func (s *SagaCoordinator) handleCompleteWorkflow(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Completing workflow")

	_, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	return s.completeWorkflow(ctx, state, headers)
}

// handleCallAgent handles agent calls that need responses
func (s *SagaCoordinator) handleCallAgent(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Executing call_agent with wait-for-response")

	result, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil
	}

	messageSent, _ := resultMap["message_sent"].(bool)
	if !messageSent {
		return nil
	}

	requestID, _ := resultMap["request_id"].(string)

	// Update state to wait for response
	state.Status = StatusAwaitingResponses
	state.AwaitedSteps = []string{requestID}

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state for waiting: %w", err)
	}

	s.logger.Info("Waiting for agent response", zap.String("request_id", requestID))

	// Set up timeout
	timeout := s.getTimeout(step, 60*time.Second)
	go s.handleTimeout(ctx, state.OrchestrationID, requestID, timeout)

	return nil
}

// handleStartOrchestration handles child orchestration starts
func (s *SagaCoordinator) handleStartOrchestration(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Executing start_orchestration with wait-for-response")

	result, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok || !resultMap["await_response"].(bool) {
		return nil
	}

	requestID, ok := resultMap["request_id"].(string)
	if !ok {
		return fmt.Errorf("no request_id in start_orchestration result")
	}

	// Store result and update state
	state.CollectedData[state.CurrentStep] = result
	state.Status = StatusAwaitingResponses
	state.AwaitedSteps = []string{requestID}
	state.ExecutionMetadata.Checkpoints[state.CurrentStep] = time.Now().UTC()

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state for waiting: %w", err)
	}

	s.logger.Info("Waiting for child orchestration", zap.String("request_id", requestID))

	timeout := s.getTimeout(step, 5*time.Minute)
	go s.handleChildOrchestrationTimeout(ctx, state.OrchestrationID, requestID, timeout)

	return nil
}

// handleLocalAction handles local actions that can continue immediately
func (s *SagaCoordinator) handleLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Executing local action", zap.String("action", step.Action))

	_, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	if step.NextStep == "" {
		return nil
	}

	// Move to next step
	s.logger.Info("Moving to next step", zap.String("next_step", step.NextStep))
	state.CurrentStep = step.NextStep

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Continue execution
	return s.continueExecution(ctx, state, headers)
}

// Helper functions

func (s *SagaCoordinator) startExecutionRecord(step, action string) ExecutionRecord {
	return ExecutionRecord{
		Step:      step,
		Action:    action,
		StartTime: time.Now().UTC(),
	}
}

func (s *SagaCoordinator) finishExecutionRecord(ctx context.Context, state *OrchestrationState, record ExecutionRecord, err error) {
	endTime := time.Now().UTC()
	record.EndTime = &endTime

	if err != nil {
		record.Result = "failed"
		record.Error = err.Error()
		s.logger.Error("Step execution failed",
			zap.String("step", record.Step),
			zap.Error(err))
	} else {
		record.Result = "success"
		s.logger.Info("Step execution succeeded",
			zap.String("step", record.Step))
	}

	s.recordExecution(ctx, state, record)
}

func (s *SagaCoordinator) skipStep(ctx context.Context, state *OrchestrationState, reason string) error {
	s.logger.Info("Skipping step",
		zap.String("step", state.CurrentStep),
		zap.String("reason", reason))

	record := ExecutionRecord{
		Step:      state.CurrentStep,
		Action:    "skipped",
		StartTime: time.Now().UTC(),
		Result:    "skipped",
		Error:     reason,
	}

	s.recordExecution(ctx, state, record)
	return nil
}

func (s *SagaCoordinator) getTimeout(step models.Step, defaultTimeout time.Duration) time.Duration {
	if step.Timeout > 0 {
		return step.Timeout
	}
	return defaultTimeout
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
