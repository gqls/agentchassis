// FILE: platform/orchestration/coordinator.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/gqls/agentchassis/platform/kafka"
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

// NewSagaCoordinator creates a new coordinator instance
func NewSagaCoordinator(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *SagaCoordinator {
	return &SagaCoordinator{
		db:          db,
		producer:    producer,
		logger:      logger,
		fuelManager: governance.NewFuelManager(),
	}
}

// ExecuteWorkflow manages the execution of a workflow plan
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	correlationID := headers["correlation_id"]
	l := s.logger.With(zap.String("correlation_id", correlationID))

	// Get clientID from headers to pass to state creation
	clientID := headers["client_id"]
	if clientID == "" {
		return fmt.Errorf("client_id header is required to execute a workflow")
	}

	// Get or create state
	state, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData)
	if err != nil {
		return err
	}

	// Check if workflow is already complete
	if state.Status == StatusCompleted || state.Status == StatusFailed {
		l.Info("Workflow already finished", zap.String("status", string(state.Status)))
		return nil
	}

	// Get current step configuration
	currentStepConfig, ok := plan.Steps[state.CurrentStep]
	if !ok {
		return s.failWorkflow(ctx, state, fmt.Sprintf("step '%s' not found in plan", state.CurrentStep))
	}

	// Check dependencies
	if !s.dependenciesMet(currentStepConfig.Dependencies, state) {
		l.Info("Dependencies not met, waiting", zap.Strings("dependencies", currentStepConfig.Dependencies))
		return nil
	}

	// Check fuel budget
	fuel, err := governance.GetFuelFromHeader(headers)
	if err != nil {
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to get fuel from headers: %v", err))
	}

	if !s.fuelManager.HasEnoughFuel(fuel, currentStepConfig.Action) {
		return s.failWorkflow(ctx, state, fmt.Sprintf("insufficient fuel for action '%s': have %d, need %d",
			currentStepConfig.Action, fuel, s.fuelManager.GetCost(currentStepConfig.Action)))
	}

	// Deduct fuel and update headers
	remainingFuel := s.fuelManager.DeductFuel(fuel, currentStepConfig.Action)
	governance.SetFuelHeader(headers, remainingFuel)

	// Execute the action
	switch currentStepConfig.Action {
	case "fan_out":
		return s.handleFanOut(ctx, headers, currentStepConfig, state)
	case "pause_for_human_input":
		return s.handlePauseForHumanInput(ctx, headers, currentStepConfig, state)
	case "complete_workflow":
		return s.completeWorkflow(ctx, state)
	default:
		return s.handleStandardAction(ctx, headers, currentStepConfig, state)
	}
}

// getOrCreateState retrieves existing state or creates new one
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte) (*OrchestrationState, error) {
	repo := NewStateRepository(s.db, s.logger)

	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		// State doesn't exist, create it
		if err := repo.CreateInitialState(ctx, correlationID, clientID, plan.StartStep, initialData); err != nil {
			return nil, fmt.Errorf("failed to create initial state: %w", err)
		}
		return repo.GetState(ctx, correlationID)
	}

	return state, nil
}

// dependenciesMet checks if all required dependencies have been completed
func (s *SagaCoordinator) dependenciesMet(dependencies []string, state *OrchestrationState) bool {
	for _, dep := range dependencies {
		if _, ok := state.CollectedData[dep]; !ok {
			return false
		}
	}
	return true
}

// handleStandardAction sends a message to the specified topic
func (s *SagaCoordinator) handleStandardAction(ctx context.Context, headers map[string]string, step models.Step, state *OrchestrationState) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID))

	// Prepare the message payload
	payload := models.TaskRequest{
		Action: step.Action,
		Data:   state.CollectedData,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Create new request ID for this sub-task
	newRequestID := uuid.NewString()
	outHeaders := make(map[string]string)
	for k, v := range headers {
		outHeaders[k] = v
	}
	outHeaders["causation_id"] = headers["request_id"]
	outHeaders["request_id"] = newRequestID

	// Send the message
	if err := s.producer.Produce(ctx, step.Topic, outHeaders, []byte(state.CorrelationID), payloadBytes); err != nil {
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

	l.Info("Standard action executed", zap.String("action", step.Action), zap.String("topic", step.Topic))
	return nil
}

// handleFanOut sends multiple parallel requests
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

// handlePauseForHumanInput pauses the workflow and notifies the UI
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

// HandleResponse processes a response from a sub-task
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
	correlationID := headers["correlation_id"]
	causationID := headers["causation_id"]

	l := s.logger.With(
		zap.String("correlation_id", correlationID),
		zap.String("causation_id", causationID),
	)

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Parse response
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(response, &taskResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Store response data
	state.CollectedData[causationID] = taskResponse.Data

	// Remove from awaited steps
	newAwaitedSteps := make([]string, 0)
	for _, step := range state.AwaitedSteps {
		if step != causationID {
			newAwaitedSteps = append(newAwaitedSteps, step)
		}
	}
	state.AwaitedSteps = newAwaitedSteps

	// If all responses received, set status back to running
	if len(state.AwaitedSteps) == 0 {
		state.Status = StatusRunning
	}

	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	l.Info("Response processed", zap.Int("remaining_awaited", len(state.AwaitedSteps)))

	// If all responses received, continue workflow
	if len(state.AwaitedSteps) == 0 {
		// Need to reload the workflow plan - this would come from the agent config
		// For now, we'll need to pass it through somehow
		// This is a limitation we'll address in the actual implementation
	}

	return nil
}

// ResumeWorkflow resumes a paused workflow after human input
func (s *SagaCoordinator) ResumeWorkflow(ctx context.Context, headers map[string]string, resumeData []byte) error {
	correlationID := headers["correlation_id"]
	l := s.logger.With(zap.String("correlation_id", correlationID))

	var resumePayload struct {
		Approved bool                   `json:"approved"`
		Feedback map[string]interface{} `json:"feedback,omitempty"`
	}
	if err := json.Unmarshal(resumeData, &resumePayload); err != nil {
		return fmt.Errorf("failed to unmarshal resume payload: %w", err)
	}

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	if state.Status != StatusPausedForHuman {
		return fmt.Errorf("workflow not in paused state: %s", state.Status)
	}

	if !resumePayload.Approved {
		state.Status = StatusFailed
		state.Error = "Workflow rejected by user"
		return repo.UpdateState(ctx, state)
	}

	// Add feedback to collected data
	if resumePayload.Feedback != nil {
		state.CollectedData["human_feedback"] = resumePayload.Feedback
	}

	state.Status = StatusRunning
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	l.Info("Workflow resumed after human approval")

	// Continue workflow execution
	// This would trigger re-execution with the current state

	return nil
}

// completeWorkflow marks the workflow as completed
func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState) error {
	state.Status = StatusCompleted
	finalResult, _ := json.Marshal(state.CollectedData)
	state.FinalResult = finalResult

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

// failWorkflow marks the workflow as failed
func (s *SagaCoordinator) failWorkflow(ctx context.Context, state *OrchestrationState, errorMsg string) error {
	state.Status = StatusFailed
	state.Error = errorMsg

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state to failed: %w", err)
	}

	// IMPORTANT: Return the error message as an error
	return fmt.Errorf(errorMsg)
}
