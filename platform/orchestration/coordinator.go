// FILE: platform/orchestration/coordinator.go (enhanced with logging)
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// Add the action registry
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
	"store_result":       actions.StoreResultAction,

	"ai_text_generate_anthropic": actions.ExecuteLLMPromptAction, // Map to existing action
	// "s3_upload":                  actions.DeployToHostingAction,  // Map to existing similar action
	"upload_to_s3": actions.UploadToS3Action, // Add the real S3 upload action
	"s3_upload":    actions.UploadToS3Action, // Also map s3_upload to it

	"plan_agent_team":       actions.PlanAgentTeamAction,
	"review_performance":    actions.ReviewPerformanceAction,
	"approve_agent_changes": actions.ApproveAgentChangesAction,
	"conditional_route":     actions.ConditionalRouteAction,
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
		zap.String("start_step", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	clientID := headers["client_id"]
	if clientID == "" {
		l.Error("client_id header is required to execute a workflow")
		return fmt.Errorf("client_id header is required to execute a workflow")
	}

	// Get or create state with the plan
	state, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData)
	if err != nil {
		l.Error("Failed to get or create state", zap.Error(err))
		return err
	}

	l.Info("Workflow state retrieved",
		zap.String("status", string(state.Status)),
		zap.String("current_step", state.CurrentStep))

	// Check if workflow is already complete
	if state.Status == StatusCompleted || state.Status == StatusFailed {
		l.Info("Workflow already finished", zap.String("status", string(state.Status)))
		return nil
	}

	// Continue execution from current step
	return s.continueExecution(ctx, state, headers)
}

// continueExecution executes from the current step using the stored plan
func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, headers map[string]string) error {
	l := s.logger.With(
		zap.String("correlation_id", state.CorrelationID),
		zap.String("current_step", state.CurrentStep),
	)

	l.Info("Continuing workflow execution",
		zap.String("current_step", state.CurrentStep),
		zap.Int("total_steps", len(state.WorkflowPlan.Steps)))

	// Get current step from the stored plan
	currentStepConfig, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		errorMsg := fmt.Sprintf("step '%s' not found in plan", state.CurrentStep)
		l.Error("Step not found", zap.String("missing_step", state.CurrentStep))
		return s.failWorkflow(ctx, state, errorMsg)
	}

	l.Info("Executing step",
		zap.String("step", state.CurrentStep),
		zap.String("action", currentStepConfig.Action),
		zap.String("description", currentStepConfig.Description))

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
		execErr = s.completeWorkflow(ctx, state)

	case currentStepConfig.Action == "call_agent":
		// SPECIAL HANDLING: call_agent sends message and waits for response
		l.Info("Executing call_agent with wait-for-response")

		// Execute the action to send the message
		result, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else if resultMap, ok := result.(map[string]interface{}); ok {
			// Check if message was sent successfully
			if messageSent, ok := resultMap["message_sent"].(bool); ok && messageSent {
				requestID := resultMap["request_id"].(string)

				// Update state to wait for response
				state.Status = StatusAwaitingResponses
				state.AwaitedSteps = []string{requestID}
				// Don't update CurrentStep yet - we'll do that when response arrives

				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					execErr = fmt.Errorf("failed to update state for waiting: %w", err)
				} else {
					l.Info("Waiting for agent response",
						zap.String("request_id", requestID),
						zap.String("agent_called", fmt.Sprintf("%v", resultMap["agent_called"])),
						zap.String("topic", fmt.Sprintf("%v", resultMap["topic"])),
						zap.String("next_step", currentStepConfig.NextStep))

					// Set up timeout goroutine
					timeout := 60 * time.Second // Default timeout
					if currentStepConfig.Timeout > 0 {
						timeout = currentStepConfig.Timeout
					}

					go s.handleTimeout(ctx, state.CorrelationID, requestID, timeout)

					return nil // Exit and wait for response
				}
			} else {
				execErr = fmt.Errorf("call_agent did not send message successfully")
			}
		} else {
			execErr = fmt.Errorf("unexpected result type from call_agent")
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
			// For non-call_agent local actions, continue to next step
			if currentStepConfig.NextStep != "" {
				l.Info("Moving to next step", zap.String("next_step", currentStepConfig.NextStep))
				state.CurrentStep = currentStepConfig.NextStep

				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					execErr = fmt.Errorf("failed to update state: %w", err)
				} else {
					// Continue execution immediately for local actions
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

// executeLocalAction handles actions that run within the orchestrator
func (s *SagaCoordinator) executeLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) (interface{}, error) {
	l := s.logger.With(
		zap.String("correlation_id", state.CorrelationID),
		zap.String("action", step.Action),
	)

	handler, ok := actionRegistry[step.Action]
	if !ok {
		return nil, fmt.Errorf("local action '%s' not found in registry", step.Action)
	}

	// Log with virtual topic for consistency
	virtualTopic := fmt.Sprintf("local.action.%s", step.Action)
	l.Info("Executing local action",
		zap.String("virtual_topic", virtualTopic))

	// Verify producer is available
	if s.producer == nil {
		l.Error("Producer is nil in SagaCoordinator")
		return nil, fmt.Errorf("producer not available for action execution")
	}

	// Prepare action parameters
	params := actions.ActionParams{
		Context:         ctx,
		Headers:         headers,
		StepConfig:      step,
		InputData:       state.InitialRequestData,
		CollectedData:   state.CollectedData,
		SagaCoordinator: s,
		Producer:        s.producer,
		DB:              s.db,
		Logger:          s.logger,
		AgentType:       headers["agent_type"],
		CurrentStep:     state.CurrentStep,
	}

	params.Logger.Info("in execute local action",
		zap.Any("headers", headers),
		zap.Any("agent config", state.CollectedData["agent_config"]))

	// If agent_type is not in headers, try to get it from other sources
	if params.AgentType == "" {
		// Try to get from collected data or state
		if agentConfig, ok := state.CollectedData["agent_config"].(map[string]interface{}); ok {
			if at, ok := agentConfig["agent_type"].(string); ok {
				params.AgentType = at
			}
		}
	}

	// Log what we're passing
	l.Info("Executing local action with params",
		zap.String("agent_type", params.AgentType),
		zap.String("action", step.Action))

	// Execute the action
	result, err := handler(ctx, params)
	if err != nil {
		l.Error("Local action failed", zap.Error(err))
		return nil, fmt.Errorf("local action failed: %w", err)
	}

	l.Info("Local action completed successfully", zap.Any("result", result))

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
		l.Info("Moving to next step", zap.String("next_step", step.NextStep))
		state.CurrentStep = step.NextStep

		// Update state
		repo := NewStateRepository(s.db, s.logger)
		if err := repo.UpdateState(ctx, state); err != nil {
			return result, fmt.Errorf("failed to update state: %w", err)
		}

		l.Info("Local action completed, continuing workflow",
			zap.String("next_step", step.NextStep))

		// Continue execution immediately for local actions
		return result, s.continueExecution(ctx, state, headers)
	} else {
		l.Info("No next step specified, workflow may be complete")
	}

	return result, nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	l := s.logger.With(
		zap.String("correlation_id", state.CorrelationID),
		zap.String("action", step.Action),
		zap.String("topic", step.Topic),
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
	outHeaders["request_id"] = newRequestID

	l.Info("Sending remote action",
		zap.String("request_id", newRequestID),
		zap.String("topic", step.Topic))

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
		zap.String("request_id", newRequestID))

	return nil
}

// Rest of the methods remain the same but with enhanced logging...
// (I'll include key ones with logging enhancements)

// completeWorkflow marks the workflow as completed with enhanced tracking
func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID))

	l.Info("Completing workflow",
		zap.Int("completed_steps", state.ExecutionMetadata.CompletedSteps),
		zap.Int("total_steps", len(state.WorkflowPlan.Steps)))

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

	return err
}

// getOrCreateState retrieves existing state or creates new one - now includes the plan
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte) (*OrchestrationState, error) {
	l := s.logger.With(zap.String("correlation_id", correlationID))

	repo := NewStateRepository(s.db, s.logger)

	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		l.Info("State doesn't exist, creating new one")
		// State doesn't exist, create it with the plan
		if err := repo.CreateInitialState(ctx, correlationID, clientID, plan, initialData); err != nil {
			l.Error("Failed to create initial state", zap.Error(err))
			return nil, fmt.Errorf("failed to create initial state: %w", err)
		}
		return repo.GetState(ctx, correlationID)
	}

	l.Info("Retrieved existing state", zap.String("status", string(state.Status)))
	return state, nil
}

// failWorkflow marks the workflow as failed
func (s *SagaCoordinator) failWorkflow(ctx context.Context, state *OrchestrationState, errorMsg string) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID))

	l.Error("Failing workflow", zap.String("error", errorMsg))

	state.Status = StatusFailed
	state.Error = errorMsg

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		l.Error("Failed to update state to failed", zap.Error(err))
		return fmt.Errorf("failed to update state to failed: %w", err)
	}

	// IMPORTANT: Return the error message as an error
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
			zap.String("step", record.Step))
	}
}

// HandleResponse and other methods would also get enhanced logging...
// (keeping response brief, but the pattern is the same)

// HandleResponse processes responses and continues workflow
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
	correlationID := headers["correlation_id"]
	causationID := headers["causation_id"]

	l := s.logger.With(
		zap.String("correlation_id", correlationID),
		zap.String("causation_id", causationID),
	)

	l.Info("Handling workflow response")

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		l.Error("Failed to get state for response", zap.Error(err))
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Parse response
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(response, &taskResponse); err != nil {
		l.Error("Failed to unmarshal response", zap.Error(err))
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	l.Info("Response parsed successfully", zap.Any("response_data", taskResponse.Data))

	// Store response data under the causation_id (request_id)
	state.CollectedData[causationID] = taskResponse.Data

	// Also store under the current step name if available
	if state.CurrentStep != "" {
		// For call_agent responses, we want to store the actual result
		state.CollectedData[state.CurrentStep] = taskResponse.Data
	}

	// Remove from awaited steps
	newAwaitedSteps := make([]string, 0)
	for _, step := range state.AwaitedSteps {
		if step != causationID {
			newAwaitedSteps = append(newAwaitedSteps, step)
		}
	}
	state.AwaitedSteps = newAwaitedSteps

	// Update metadata
	state.ExecutionMetadata.CompletedSteps++

	// If all responses received, continue workflow
	if len(state.AwaitedSteps) == 0 {
		l.Info("All responses received, continuing workflow")
		state.Status = StatusRunning

		// Move to next step
		currentStep := state.WorkflowPlan.Steps[state.CurrentStep]
		if currentStep.NextStep != "" {
			state.CurrentStep = currentStep.NextStep
			l.Info("Moving to next step after response",
				zap.String("next_step", currentStep.NextStep))
		}

		if err := repo.UpdateState(ctx, state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		// Ensure required headers for continuation
		if headers["fuel_budget"] == "" {
			headers["fuel_budget"] = "1000"
		}
		if headers["client_id"] == "" {
			headers["client_id"] = state.ClientID
		}
		if headers["agent_instance_id"] == "" {
			headers["agent_instance_id"] = "00000000-0000-0000-0000-000000000001"
		}

		// Continue execution with the stored plan
		return s.continueExecution(ctx, state, headers)
	}

	// Still waiting for more responses
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	l.Info("Response processed",
		zap.Int("remaining_awaited", len(state.AwaitedSteps)))

	return nil
}

// Add other missing methods like handleFanOut, handlePauseForHumanInput etc. with similar logging enhancements
func (s *SagaCoordinator) handleFanOut(ctx context.Context, headers map[string]string, step models.Step, state *OrchestrationState) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID))

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

		if err := s.producer.Produce(ctx, subTask.Topic, outHeaders, []byte(state.CorrelationID), payloadBytes); err != nil {
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
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID))

	state.Status = StatusPausedForHuman
	state.CurrentStep = step.NextStep

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Send notification
	notification := map[string]interface{}{
		"event_type":      "WORKFLOW_PAUSED_FOR_APPROVAL",
		"correlation_id":  state.CorrelationID,
		"project_id":      headers["project_id"],
		"client_id":       headers["client_id"],
		"message":         fmt.Sprintf("Step '%s' requires your approval", step.Description),
		"data_for_review": state.CollectedData,
	}
	notificationBytes, _ := json.Marshal(notification)

	if err := s.producer.Produce(ctx, NotificationTopic, headers, []byte(state.CorrelationID), notificationBytes); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	l.Info("Workflow paused for human input")
	return nil
}

// CreateNewOrchestration and other methods would also get logging enhancements
func (s *SagaCoordinator) CreateNewOrchestration(ctx context.Context, correlationID string, headers map[string]string, workflowJSON json.RawMessage) error {
	l := s.logger.With(zap.String("correlation_id", correlationID))

	// Parse the workflow
	var plan models.WorkflowPlan
	if err := json.Unmarshal(workflowJSON, &plan); err != nil {
		return fmt.Errorf("failed to unmarshal workflow: %w", err)
	}

	// Validate we have required headers
	clientID := headers["client_id"]
	if clientID == "" {
		return fmt.Errorf("client_id header is required")
	}

	// Prepare initial data from headers
	initialData := map[string]interface{}{
		"action": "start_orchestration",
		"data": map[string]interface{}{
			"headers":   headers,
			"timestamp": time.Now().UTC(),
			"message":   "Starting website build orchestration",
		},
	}
	initialDataBytes, _ := json.Marshal(initialData)

	// Create the initial state
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.CreateInitialState(ctx, correlationID, clientID, plan, initialDataBytes); err != nil {
		return fmt.Errorf("failed to create orchestration state: %w", err)
	}

	l.Info("New orchestration created",
		zap.String("client_id", clientID),
		zap.String("start_step", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	// Start execution immediately
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		return fmt.Errorf("failed to get created state: %w", err)
	}

	// Ensure headers have required fields for execution
	if headers["fuel_budget"] == "" {
		headers["fuel_budget"] = "1000" // Default fuel budget
	}
	if headers["agent_instance_id"] == "" {
		headers["agent_instance_id"] = "orchestrator-" + correlationID
	}

	// Start the workflow execution
	return s.continueExecution(ctx, state, headers)
}

// FILE: platform/orchestration/coordinator.go
// Add this new method to handle timeouts

func (s *SagaCoordinator) handleTimeout(ctx context.Context, correlationID string, requestID string, timeout time.Duration) {
	time.Sleep(timeout)

	// Check if still waiting for this request
	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		s.logger.Error("Failed to get state for timeout check",
			zap.String("correlation_id", correlationID),
			zap.Error(err))
		return
	}

	// Check if still waiting for this specific request
	for _, awaitedStep := range state.AwaitedSteps {
		if awaitedStep == requestID {
			s.logger.Error("Timeout waiting for agent response",
				zap.String("correlation_id", correlationID),
				zap.String("request_id", requestID),
				zap.Duration("timeout", timeout))

			// Fail the workflow
			s.failWorkflow(ctx, state, fmt.Sprintf("timeout after %v waiting for agent response (request_id: %s)", timeout, requestID))
			return
		}
	}

	// If we get here, the response was already received
	s.logger.Info("Timeout check passed - response already received",
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID))
}
