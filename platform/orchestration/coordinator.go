// FILE: platform/orchestration/coordinator.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
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
	// Timeout for stuck orchestrations
	StuckOrchestrationTimeout = 5 * time.Minute
	// Default request timeout
	DefaultRequestTimeout = 120 * time.Second
)

var (
	ErrWaitingForResponse = errors.New("orchestration is waiting for responses")
	ErrVersionMismatch    = errors.New("optimistic lock failure: version mismatch")
)

// SagaCoordinator manages the execution of complex workflows
type SagaCoordinator struct {
	db          *sql.DB
	producer    kafka.Producer
	logger      *zap.Logger
	fuelManager *governance.FuelManager
	tracer      *types.TraceLogger

	// For stateless operation
	isStateless bool
	podName     string
}

// NewSagaCoordinator creates a new coordinator instance
func NewSagaCoordinator(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *SagaCoordinator {
	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		podName = fmt.Sprintf("coordinator-local-%d", os.Getpid())
	}

	return &SagaCoordinator{
		db:          db,
		producer:    producer,
		logger:      logger,
		fuelManager: governance.NewFuelManager(),
		tracer:      types.NewTraceLogger(logger),
		isStateless: os.Getenv("ENABLE_STATELESS_MODE") == "true",
		podName:     podName,
	}
}

// ExecuteWorkflow executes a workflow with stateless support
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Create minimal context if parsing fails
		execCtx = &types.ExecutionContext{
			CorrelationID:   headers["correlation_id"],
			ClientID:        headers["client_id"],
			MessageID:       headers["message_id"],
			OrchestrationID: headers["orchestration_id"],
			Timestamp:       time.Now(),
		}
	}

	l := s.logger.With(
		zap.String("correlation_id", execCtx.CorrelationID),
		zap.String("message_id", execCtx.MessageID),
		zap.Bool("stateless", s.isStateless),
		zap.String("pod_name", s.podName))

	l.Info("ExecuteWorkflow called coordinator.go 92 PLAN",
		zap.Any("plan", plan),
		zap.String("start_step_", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	if execCtx.ClientID == "" {
		return fmt.Errorf("client_id is required to execute a workflow")
	}

	// Get or create orchestration state
	state, orchestrationID, isNew, err := s.getOrCreateState(ctx, execCtx, plan, initialData)
	if err != nil {
		l.Error("Failed to get or create state", zap.Error(err))
		return err
	}

	l.Info("Orchestration state retrieved",
		zap.String("orchestration_id", orchestrationID),
		zap.Bool("is_new", isNew),
		zap.String("status", string(state.Status)))

	// Update context with orchestration ID
	execCtx.OrchestrationID = orchestrationID
	headers["orchestration_id"] = orchestrationID

	// Handle based on current status
	return s.handleOrchestrationStatus(ctx, state, execCtx, isNew)
}

// ProcessResponse handles responses using ExecutionContext
func (s *SagaCoordinator) ProcessResponse(ctx context.Context, execCtx *types.ExecutionContext, response []byte) error {

	// Extract request ID from InResponseTo
	var requestID string
	if execCtx.InResponseTo != nil {
		requestID = execCtx.InResponseTo.RequestID
	}
	if requestID == "" {
		requestID = execCtx.RequestID
	}
	if requestID == "" {
		return fmt.Errorf("no request ID in response")
	}

	s.logger.Info("RESPONSE_RECEIVED: Processing response in coordinator")
	contextLogger := s.logger.With(
		zap.String("request_id", requestID),
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("from_agent", execCtx.FromAgentID),
		zap.String("status", execCtx.Status),
		zap.Int("retry_version", execCtx.RetryVersion),
		zap.String("response_preview", string(response)[:min(500, len(response))]),
	)

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go ProcessResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	repo := NewStateRepository(s.db, s.logger)

	// Find state by request ID (not step ID!)
	state, err := repo.FindByAwaitedRequestID(ctx, requestID)
	if err != nil {
		// Try fallback to orchestration ID
		var orchID string
		if execCtx.InResponseTo != nil && execCtx.InResponseTo.ParentOrchestrationID != "" {
			orchID = execCtx.InResponseTo.ParentOrchestrationID
		} else {
			orchID = execCtx.ParentOrchestrationID
		}

		if orchID != "" {
			state, err = repo.GetState(ctx, orchID)
			if err != nil {
				return fmt.Errorf("unable to get state when looking for orchestration by request id: %s", requestID)
			}

			// Verify this state is actually waiting for this request
			if _, exists := state.AwaitedRequests[requestID]; !exists {
				return fmt.Errorf("orchestration %s not waiting for request %s", orchID, requestID)
			}
		} else {
			return fmt.Errorf("no orchestration found for request_id: %s", requestID)
		}
	}

	contextLogger.Info("RESPONSE_MATCHED: Found orchestration for response",
		zap.String("state_orch_id", state.OrchestrationID),
		zap.String("state_status", string(state.Status)),
		zap.Int("awaited_requests", len(state.AwaitedRequests)),
	)

	// Handle based on response status
	switch execCtx.Status {
	case "awaiting", "processing":
		return s.handleProgressUpdate(ctx, state, execCtx)
	case "complete":
		return s.handleCompleteResponse(ctx, state, requestID, execCtx, response)
	case "error_recoverable":
		return s.handleRecoverableError(ctx, state, requestID, execCtx, response)
	case "error_unrecoverable":
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	default:
		contextLogger.Warn("Unknown response status", zap.String("status", execCtx.Status))
		return nil
	}
}

// HandleResponse processes responses using headers (compatibility method)
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		s.logger.Error("Failed to create ExecutionContext from headers", zap.Error(err))

		// Try manual extraction for critical fields
		requestID := headers["in_response_to"]
		if requestID == "" {
			requestID = headers["causation_id"] // Legacy fallback
		}
		if requestID == "" {
			return fmt.Errorf("no request ID in headers")
		}

		// Create minimal ExecutionContext
		execCtx = &types.ExecutionContext{
			OrchestrationID: headers["orchestration_id"],
			CorrelationID:   headers["correlation_id"],
			ClientID:        headers["client_id"],
			Status:          headers["status"],
			InResponseTo: &types.ResponseContext{
				RequestID: requestID,
			},
		}
	}

	return s.ProcessResponse(ctx, execCtx, response)
}

// getOrCreateState gets or creates orchestration state
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, execCtx *types.ExecutionContext, plan models.WorkflowPlan, initialData []byte) (*OrchestrationState, string, bool, error) {
	repo := NewStateRepository(s.db, s.logger)

	// Generate orchestration name if not provided
	orchestrationName := execCtx.OrchestrationName
	if orchestrationName == "" {
		// Generate readable name based on agent type and timestamp
		orchestrationName = fmt.Sprintf("%s-%s-%s",
			s.determineOwnerAgentType(execCtx),
			execCtx.Action,
			time.Now().Format("0102-1504"))
	}

	// Priority 1: Use explicit orchestration_id
	if execCtx.OrchestrationID != "" {
		state, err := repo.GetState(ctx, execCtx.OrchestrationID)
		if err == nil {
			s.logger.Info("Found existing state by orchestration_id",
				zap.String("orchestration_id", execCtx.OrchestrationID))
			return state, execCtx.OrchestrationID, false, nil
		}

		// If we have a parent, this is likely a new child orchestration
		if execCtx.ParentOrchestrationID != "" {
			// Generate name for child orchestration if not provided
			childOrchestrationName := execCtx.OrchestrationName
			if childOrchestrationName == "" {
				childOrchestrationName = fmt.Sprintf("%s-%s-%s",
					s.determineOwnerAgentType(execCtx),
					execCtx.Action,
					time.Now().Format("0102-1504"))
			}

			s.logger.Info("Creating child orchestration",
				zap.String("orchestration_id", execCtx.OrchestrationID),
				zap.String("orchestration_name", childOrchestrationName),
				zap.String("parent_orchestration_id", execCtx.ParentOrchestrationID))

			ownerAgentID := s.determineOwnerAgentID(execCtx)
			ownerAgentType := s.determineOwnerAgentType(execCtx) // ADD THIS

			// When creating initial state, pass both ID and Type
			err := repo.CreateInitialState(ctx, execCtx.OrchestrationID, execCtx.OrchestrationName, execCtx.CorrelationID,
				ownerAgentID, ownerAgentType, execCtx.ParentOrchestrationID, execCtx.ClientID, plan, initialData)

			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					state, err = repo.GetState(ctx, execCtx.OrchestrationID)
					if err == nil {
						return state, execCtx.OrchestrationID, false, nil
					}
				}
				return nil, "", false, fmt.Errorf("failed to create child orchestration: %w", err)
			}

			state, err = repo.GetState(ctx, execCtx.OrchestrationID)
			if err != nil {
				return nil, "", false, fmt.Errorf("failed to get newly created state: %w", err)
			}
			return state, execCtx.OrchestrationID, true, nil
		}
	}

	// Priority 2: For root orchestrations, try correlation_id lookup
	if execCtx.ParentOrchestrationID == "" {
		state, err := repo.GetStateByCorrelation(ctx, execCtx.CorrelationID)
		if err == nil {
			s.logger.Info("Found existing root orchestration by correlation_id",
				zap.String("correlation_id", execCtx.CorrelationID))
			return state, state.OrchestrationID, false, nil
		}
	}

	// Priority 3: Create new orchestration
	newOrchestrationID := execCtx.OrchestrationID
	if newOrchestrationID == "" {
		newOrchestrationID = uuid.New().String()
	}

	// Generate name for new orchestration if not provided
	newOrchestrationName := execCtx.OrchestrationName
	if newOrchestrationName == "" {
		newOrchestrationName = fmt.Sprintf("%s-%s-%s",
			s.determineOwnerAgentType(execCtx),
			execCtx.Action,
			time.Now().Format("0102-1504"))
	}

	ownerAgentID := s.determineOwnerAgentID(execCtx)
	ownerAgentType := s.determineOwnerAgentType(execCtx)

	s.logger.Info("Creating new orchestration",
		zap.String("orchestration_id", newOrchestrationID),
		zap.String("orchestration_name", newOrchestrationName),
		zap.String("correlation_id", execCtx.CorrelationID))

	err := repo.CreateInitialState(ctx, newOrchestrationID, newOrchestrationName, execCtx.CorrelationID,
		ownerAgentID, ownerAgentType, execCtx.ParentOrchestrationID, execCtx.ClientID, plan, initialData)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			// Race condition - try to find existing
			if execCtx.ParentOrchestrationID == "" {
				state, err := repo.GetStateByCorrelation(ctx, execCtx.CorrelationID)
				if err == nil {
					return state, state.OrchestrationID, false, nil
				}
			} else {
				state, err := repo.GetState(ctx, newOrchestrationID)
				if err == nil {
					return state, newOrchestrationID, false, nil
				}
			}
		}
		return nil, "", false, fmt.Errorf("failed to create orchestration: %w", err)
	}

	state, err := repo.GetState(ctx, newOrchestrationID)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to get newly created state: %w", err)
	}

	return state, newOrchestrationID, true, nil
}

// handleOrchestrationStatus handles orchestration based on its current status
func (s *SagaCoordinator) handleOrchestrationStatus(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext, isNew bool) error {
	s.logger.Info("Handling orchestration status handleOrchestrationStatus coordinator.go 360",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.Any("DEBUGaa: state.Status", state.Status),
		zap.Any("state.CurrentlyExecuting", state.CurrentlyExecuting),
	)

	repo := NewStateRepository(s.db, s.logger)

	switch state.Status {
	case StatusInitialized:
		// New orchestration, start execution
		s.logger.Info("Starting new orchestration execution")

		if err := repo.SetExecutingStep(ctx, state.OrchestrationID, state.CurrentStep); err != nil {
			return err
		}

		// Reload state to get updated version
		state, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to reload state: %w", err)
		}

		return s.continueExecution(ctx, state, execCtx)

	case StatusExecutingStep:
		// Check if stuck
		if state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout {
			s.logger.Warn("Found stuck orchestration, taking over",
				zap.String("stuck_step", *state.CurrentlyExecuting))

			if err := repo.ClearExecutingStep(ctx, state.OrchestrationID); err != nil {
				return err
			}

			state, err := repo.GetState(ctx, state.OrchestrationID)
			if err != nil {
				return fmt.Errorf("failed to reload state: %w", err)
			}

			return s.continueExecution(ctx, state, execCtx)
		}

		s.logger.Info("Orchestration is actively executing")
		return nil

	case StatusAwaitingResponses:
		s.logger.Info("Orchestration is awaiting responses",
			zap.Int("awaited_count", len(state.AwaitedRequests)))
		return ErrWaitingForResponse

	case StatusCompleted:
		s.logger.Info("Workflow already completed")
		return nil

	case StatusFailed:
		s.logger.Info("Workflow previously failed",
			zap.String("error", state.Error))
		return nil

	default:
		return fmt.Errorf("unknown orchestration status: %s", state.Status)
	}
}

// continueExecution executes from the current step
func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext) error {
	repo := NewStateRepository(s.db, s.logger)

	// This initial check remains outside the loop. If the function is called
	// on a state that's already waiting, we should exit immediately.
	if state.Status == StatusAwaitingResponses {
		s.logger.Info("Already in waiting state", zap.String("orchestration_id", state.OrchestrationID))
		return nil
	}

	// This loop replaces the recursion. It will run for each step in the workflow
	// until a 'return' statement is hit.
	for {
		l := s.logger.With(
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("current_step", state.CurrentStep),
			zap.String("exec_ctx_sender_type", execCtx.Sender.AgentType),
			zap.String("exec_ctx_message_type", execCtx.MessageType),
			zap.Any("exec_ctx_in_response_to", execCtx.InResponseTo),
			zap.Any("state.Status", state.Status),
		)

		l.Info("Continuing workflow execution continueExecution coordinator.go",
			zap.Int("total_steps", len(state.WorkflowPlan.Steps)))

		// Mark the current step as executing for this iteration
		if err := repo.SetExecutingStep(ctx, state.OrchestrationID, state.CurrentStep); err != nil {
			return err
		}

		// Reload the state to ensure we have the latest version for this step
		reloadedState, err := repo.GetState(ctx, state.OrchestrationID)
		if err != nil {
			return fmt.Errorf("failed to reload state: %w", err)
		}
		state = reloadedState

		l = s.logger.With(
			zap.Any("DEBUGAA: state loaded from repo - in continueExecution", state),
			zap.String("current_step", state.CurrentStep),
			zap.Any("state.Status", state.Status),
			zap.Any("currentStepConfig", state.WorkflowPlan.Steps[state.CurrentStep]),
		)

		currentStepConfig, exists := state.WorkflowPlan.Steps[state.CurrentStep]
		if !exists {
			l.Error("Current step doesnt exist in the plan", zap.Any("state.WorkflowPlan", state.WorkflowPlan))
			return s.failWorkflow(ctx, state, fmt.Sprintf("step '%s' not found", state.CurrentStep))
		}

		if !s.dependenciesMet(currentStepConfig.Dependencies, state) {
			return s.skipStep(ctx, state, "dependencies not met")
		}

		// Execute the single step for this iteration
		err = s.executeStep(ctx, state, currentStepConfig, execCtx)
		if err != nil {
			return s.failWorkflow(ctx, state, fmt.Sprintf("step %s failed: %v", state.CurrentStep, err))
		}

		// If the step requires a pause, save the state and exit the loop.
		if state.Status == StatusAwaitingResponses {
			l.Info("Execution paused - waiting for responses")
			return repo.UpdateState(ctx, state)
		}

		// --- This is the core step-transition logic ---
		if currentStepConfig.NextStep != "" {
			l.Info("currentStepConfig.NextStep was not blank", zap.String("next_step", currentStepConfig.NextStep))
			state.CurrentStep = currentStepConfig.NextStep

			if err := repo.UpdateState(ctx, state); err != nil {
				return err
			}

			// Go to the top of the for loop to run the next step
			continue
		} else {
			l.Info("currentStepConfig.NextStep WAS blank, completing workflow.")
			// No next step, so complete the workflow and exit the loop
			return s.completeWorkflow(ctx, state)
		}
	}
}

// executeStep executes a single workflow step
func (s *SagaCoordinator) executeStep(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	s.logger.Info("Executing step in executeStep before executeLocalAction",
		zap.String("step", state.CurrentStep),
		zap.String("action", step.Action),
		zap.Any("state", state),
	)

	// Route to appropriate handler
	if isLocalAction(step.Action) {
		return s.executeLocalAction(ctx, state, step, execCtx)
	} else if step.Topic != "" {
		return s.executeRemoteAction(ctx, state, step, execCtx)
	}

	return fmt.Errorf("unknown action: %s", step.Action)
}

// executeLocalAction executes a local action, handles subtree info from both spawn actions
func (s *SagaCoordinator) executeLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	// Update execution context for this step
	execCtx.StepID = uuid.New().String()
	execCtx.StepName = step.Action
	execCtx.Action = step.Action

	contextLogger := s.logger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", execCtx.StepName),
		zap.String("action", step.Action),
		zap.Any("state.CollectedData", state.CollectedData),
	)

	contextLogger.Info("In executeLocalAction",
		zap.Any("DEBUGaa: executionContext", execCtx))

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go executeLocalAction",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Pre-generate request ID for actions that need responses
	needsResponse := step.Action == "call_agent" || step.Action == "start_orchestration" ||
		step.Action == "spawn_group" || step.Action == "spawn_agent"

	if needsResponse {
		// Only generate new request ID if we don't already have one
		if execCtx.RequestID == "" {
			execCtx.RequestID = uuid.New().String()
			contextLogger.Info("Generated new request ID",
				zap.String("request_id", execCtx.RequestID))
		} else {
			contextLogger.Info("Using existing request ID for retry",
				zap.String("request_id", execCtx.RequestID))
		}
	}

	// Get action handler from global registry
	handler, exists := actions.GetAction(step.Action)
	if !exists {
		return fmt.Errorf("unknown action: %s, not found in registry", step.Action)
	}

	// Ensure ExecutionContext has Sender before creating params
	if execCtx.Sender.AgentType == "" {
		execCtx.Sender = types.AgentIdentity{
			AgentType:    state.OwnerAgentType, // From the state
			AgentID:      state.OwnerAgentID,
			PodName:      s.podName,
			AgentVersion: os.Getenv("AGENT_VERSION"),
		}
	}

	// Ensure this is a request, not a response
	execCtx.MessageType = "request"

	// Clear any response-specific fields
	execCtx.InResponseTo = nil
	execCtx.Status = ""
	execCtx.IsComplete = false

	// Prepare action params
	params := actions.ActionParams{
		Context:          ctx,
		ExecutionContext: execCtx,
		StepConfig:       step,
		Headers:          execCtx.ToHeaders(),
		CollectedData:    state.CollectedData,
		Logger:           contextLogger,
		Producer:         s.producer,
		DB:               s.db,
		Tracer:           s.tracer,
	}

	// Execute action
	result, err := handler(ctx, params)
	if err != nil {
		contextLogger.Error("Local action failed", zap.Error(err), zap.Any("params", params))
		return fmt.Errorf("local action failed: %w", err)
	}

	// Store result
	state.CollectedData[state.CurrentStep] = result

	// Handle subtree info if returned (for spawn actions), handle all state modifications in memory
	needsWaiting := false
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Handle subtree info
		if subtreeInfo, ok := resultMap["subtree_info"].(*types.SubtreeInfo); ok {
			if state.SubtreeAgents == nil {
				state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
			}
			state.SubtreeAgents[subtreeInfo.AgentID] = subtreeInfo

			contextLogger.Info("Added agent to subtree",
				zap.String("agent_id", subtreeInfo.AgentID),
				zap.String("agent_type", subtreeInfo.AgentType))
		}

		// Check if action requires waiting
		if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
			requestID := execCtx.RequestID
			if resultRequestID, ok := resultMap["request_id"].(string); ok && resultRequestID != "" {
				requestID = resultRequestID
			}

			if requestID != "" {
				// Create awaited request entry IN MEMORY
				awaitedReq := &AwaitedRequest{
					RequestID:       requestID,
					StepID:          execCtx.StepID,
					StepName:        step.Action,
					RetryVersion:    0,
					TargetAgentType: getTargetAgentType(step, resultMap),
					ResponseTopic:   execCtx.ResponsesTopic,
					SentAt:          time.Now(),
					TimeoutAt:       time.Now().Add(getTimeout(step)),
				}

				// Add to state IN MEMORY
				if state.AwaitedRequests == nil {
					state.AwaitedRequests = make(map[string]*AwaitedRequest)
				}
				state.AwaitedRequests[requestID] = awaitedReq
				state.Status = StatusAwaitingResponses
				needsWaiting = true

				contextLogger.Info("Action requires waiting for response",
					zap.String("request_id", requestID),
					zap.String("target_agent_type", awaitedReq.TargetAgentType))

				// SAVE STATE BEFORE EXECUTING ACTION
				repo := NewStateRepository(s.db, s.logger)
				if err := repo.UpdateState(ctx, state); err != nil {
					return fmt.Errorf("failed to save awaited request: %w", err)
				}

				// Set up timeout handler
				go s.handleRequestTimeout(ctx, state.OrchestrationID, requestID, awaitedReq.TimeoutAt)
			}
		}
	}

	// Add processing record IN MEMORY
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		StepID:    execCtx.StepID,
		StepName:  step.Action,
		Action:    "executed",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Executed by %s", s.podName),
	})

	// If we need to wait, save state NOW before returning
	if needsWaiting {
		repo := NewStateRepository(s.db, s.logger)
		if err := repo.UpdateState(ctx, state); err != nil {
			return fmt.Errorf("failed to save state before waiting: %w", err)
		}
		contextLogger.Info("State saved before waiting for response")
	}

	// NO DATABASE SAVE HERE - let continueExecution handle it
	return nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	contextLogger := s.logger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", execCtx.StepName),
		zap.String("action", step.Action),
		zap.Any("state.CollectedData", state.CollectedData),
	)
	contextLogger.Info("In executeRemoteAction",
		zap.Any("ExecutionContext", execCtx))

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go executeRemoteAction",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Update execution context
	execCtx.RequestID = uuid.New().String()
	execCtx.StepID = uuid.New().String()
	execCtx.StepName = step.Action
	execCtx.Action = step.Action
	execCtx.ToAgentType = step.TargetAgentType
	execCtx.ResponsesTopic = fmt.Sprintf("system.agent.%s.responses", execCtx.Sender.AgentType)

	l := s.logger.With(
		zap.String("request_id", execCtx.RequestID),
		zap.String("topic", step.Topic))

	// Prepare the message
	requestMsg := &types.RequestMessage{
		Headers: execCtx.ToRequestHeaders(),
		Body: models.TaskRequest{
			Action: step.Action,
			Data:   state.CollectedData,
		},
	}

	msgBytes, _ := json.Marshal(requestMsg)

	// Send the message
	if err := s.producer.Produce(ctx, step.Topic, requestMsg.Headers.ToMap(), []byte(execCtx.CorrelationID), msgBytes); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Create awaited request
	awaitedReq := &AwaitedRequest{
		RequestID:       execCtx.RequestID,
		StepID:          execCtx.StepID,
		StepName:        step.Action,
		RetryVersion:    0,
		TargetAgentType: step.TargetAgentType,
		ResponseTopic:   execCtx.ResponsesTopic, // Store where we expect responses
		SentAt:          time.Now(),
		TimeoutAt:       time.Now().Add(getTimeout(step)),
	}

	// Update state
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.AddAwaitedRequest(ctx, state.OrchestrationID, awaitedReq); err != nil {
		return err
	}

	state.Status = StatusAwaitingResponses
	state.CurrentStep = step.NextStep

	if err := repo.UpdateState(ctx, state); err != nil {
		return err
	}

	l.Info("Remote action initiated")

	// Set up timeout
	go s.handleRequestTimeout(ctx, state.OrchestrationID, execCtx.RequestID, awaitedReq.TimeoutAt)

	return nil
}

// handleProgressUpdate handles progress updates from agents
func (s *SagaCoordinator) handleProgressUpdate(ctx context.Context, state *OrchestrationState, execCtx *types.ExecutionContext) error {
	s.logger.Info("In handleProgressUpdate. Progress update received",
		zap.String("status", execCtx.Status),
		zap.String("from_agent", execCtx.Sender.AgentID))

	// Add processing record
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		StepID:    execCtx.StepID,
		StepName:  execCtx.StepName,
		Action:    fmt.Sprintf("progress_%s", execCtx.Status),
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Progress from %s", execCtx.Sender.AgentID),
	})

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

// handleCompleteResponse processes a successful response
func (s *SagaCoordinator) handleCompleteResponse(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response []byte) error {
	contextLogger := s.logger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", execCtx.StepName),
		zap.String("requestId", requestID),
		zap.Any("state.CollectedData", state.CollectedData),
	)
	contextLogger.Info("In handleCompleteResponse",
		zap.Any("ExecutionContext", execCtx))

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go handleCompleteResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	repo := NewStateRepository(s.db, s.logger)

	// Parse the response
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(response, &taskResponse); err != nil {
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to unmarshal response: %v", err))
	}

	// Store the response data
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}
	state.CollectedData[fmt.Sprintf("response_%s", requestID)] = taskResponse.Data

	// Find which step this response belongs to
	/*if awaitedReq, exists := state.AwaitedRequests[requestID]; exists {

		contextLogger.Info("DEBUGaa: coordinator.go handleCompleteResponse parse response structure",
			zap.Any("taskResponse", taskResponse),
		)

		stepName := awaitedReq.StepName
		if stepName == "spawn_agent" {
			// Store the spawn result under the step name
			state.CollectedData["spawn_calculator"] = map[string]interface{}{
				"agent_id":   taskResponse.Data["agent_id"],
				"agent_name": taskResponse.Data["agent_name"],
				"agent_type": taskResponse.Data["agent_type"],
				"job_name":   taskResponse.Data["job_name"],
				"status":     "initialized",
				"spawned_at": time.Now(),
			}
		}
	}*/

	if awaitedReq, exists := state.AwaitedRequests[requestID]; exists {
		if awaitedReq.StepName == "spawn_agent" {
			// Parse the nested response structure
			if body, ok := taskResponse.Data["body"].(map[string]interface{}); ok {
				if result, ok := body["result"].(map[string]interface{}); ok {
					state.CollectedData["spawn_calculator"] = map[string]interface{}{
						"agent_id":   result["agent_id"],
						"agent_name": result["agent_name"],
						"agent_type": result["agent_type"],
						"status":     "initialized",
						"spawned_at": time.Now(),
					}
				}
			}
		}
	}

	// Remove from awaited requests
	if err := repo.RemoveAwaitedRequest(ctx, state.OrchestrationID, requestID); err != nil {
		return err
	}

	// Reload state
	state, err := repo.GetState(ctx, state.OrchestrationID)
	if err != nil {
		return fmt.Errorf("failed to reload state: %w", err)
	}

	// If no more awaited requests, continue workflow
	if len(state.AwaitedRequests) == 0 {
		s.logger.Info("All responses received, continuing workflow")

		// ADVANCE TO NEXT STEP
		currentStep := state.WorkflowPlan.Steps[state.CurrentStep]
		if currentStep.NextStep != "" {
			state.CurrentStep = currentStep.NextStep
			repo.UpdateState(ctx, state) // Save the step advancement
		}

		// Create fresh execution context for continuing
		freshExecCtx := &types.ExecutionContext{
			CorrelationID:   state.CorrelationID,
			OrchestrationID: state.OrchestrationID,
			ClientID:        state.ClientID,

			// Reset to request mode
			MessageType: "request",
			MessageID:   uuid.New().String(),

			// Set sender to the current orchestrator
			Sender: types.AgentIdentity{
				AgentType:    state.OwnerAgentType,
				AgentID:      state.OwnerAgentID,
				PodName:      s.podName,
				AgentVersion: os.Getenv("AGENT_VERSION"),
			},

			// Clear any response fields
			InResponseTo: nil,
			Status:       "",
			IsComplete:   false,

			// Resources
			FuelBudget:     state.FuelBudget,
			TimeoutSeconds: 30,
			Timestamp:      time.Now(),
			Version:        "2.0",
		}

		state.LastActivity = time.Now()

		if err := repo.UpdateState(ctx, state); err != nil {
			return fmt.Errorf("failed to save state transition: %w", err)
		}

		return s.continueExecution(ctx, state, freshExecCtx)
	}

	s.logger.Info("Still waiting for responses",
		zap.Int("remaining", len(state.AwaitedRequests)))

	return nil
}

// handleRecoverableError handles errors that can be retried
func (s *SagaCoordinator) handleRecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response []byte) error {
	awaited := state.AwaitedRequests[requestID]
	if awaited == nil {
		return fmt.Errorf("no awaited request found for %s", requestID)
	}

	// Check retry count
	if awaited.RetryVersion >= 3 {
		s.logger.Error("Max retries exceeded",
			zap.String("request_id", requestID))
		return s.handleUnrecoverableError(ctx, state, requestID, execCtx, response)
	}

	// Retry the request
	awaited.RetryVersion++
	awaited.SentAt = time.Now()
	awaited.TimeoutAt = time.Now().Add(30 * time.Second)

	s.logger.Info("Retrying request",
		zap.String("request_id", requestID),
		zap.Int("retry_version", awaited.RetryVersion))

	// Create retry request with same request ID
	retryRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender:            execCtx.Sender,
			RequestID:         requestID,            // Same request ID
			RetryVersion:      awaited.RetryVersion, // Incremented retry version
			StepID:            awaited.StepID,
			StepName:          awaited.StepName,
			OrchestrationID:   state.OrchestrationID,
			OrchestrationName: state.OrchestrationName,
			CorrelationID:     state.CorrelationID,
			ToAgentType:       awaited.TargetAgentType,
			ClientID:          state.ClientID,
			MessageID:         uuid.New().String(),
			MessageType:       "request",
			Timestamp:         time.Now(),
			Action:            "retry",
		},
	}

	// Send retry
	topic := fmt.Sprintf("system.agent.%s.requests", awaited.TargetAgentType)
	retryBytes, _ := json.Marshal(retryRequest)

	return s.producer.Produce(ctx, topic, retryRequest.Headers.ToMap(), []byte(requestID), retryBytes)
}

// handleUnrecoverableError handles fatal errors
func (s *SagaCoordinator) handleUnrecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response []byte) error {
	s.logger.Error("Unrecoverable error received",
		zap.String("request_id", requestID))

	var errorResponse struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	json.Unmarshal(response, &errorResponse)

	return s.failWorkflow(ctx, state, fmt.Sprintf("Request %s failed: %s", requestID, errorResponse.Error))
}

// handleRequestTimeout handles request timeouts
func (s *SagaCoordinator) handleRequestTimeout(ctx context.Context, orchestrationID, requestID string, timeoutAt time.Time) {
	time.Sleep(time.Until(timeoutAt))

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, orchestrationID)
	if err != nil {
		return
	}

	// Check if still waiting
	if awaited, exists := state.AwaitedRequests[requestID]; exists {
		s.logger.Warn("Request timed out",
			zap.String("request_id", requestID))

		// Retry or fail
		if awaited.RetryVersion < 3 {
			// Create minimal ExecutionContext for retry
			execCtx := &types.ExecutionContext{
				RequestID: requestID,
				Status:    "error_recoverable",
			}
			s.handleRecoverableError(ctx, state, requestID, execCtx, []byte(`{"error": "timeout"}`))
		} else {
			s.failWorkflow(ctx, state, fmt.Sprintf("Request %s timed out after %d retries", requestID, awaited.RetryVersion))
		}
	}
}

// Helper methods
func (s *SagaCoordinator) determineOwnerAgentType(execCtx *types.ExecutionContext) string {
	if execCtx.Sender.AgentType != "" {
		return execCtx.Sender.AgentType
	}
	if agentType := os.Getenv("AGENT_TYPE"); agentType != "" {
		return agentType
	}
	s.logger.Error("Did not find agent type in determineOwnerAgentType (coordinator 887)")
	return "generic"
}

func (s *SagaCoordinator) determineOwnerAgentID(execCtx *types.ExecutionContext) string {
	if execCtx.Sender.AgentID != "" {
		return execCtx.Sender.AgentID
	}
	if ownerID := os.Getenv("AGENT_ID"); ownerID != "" {
		return ownerID
	}
	return "00000000-0000-0000-0000-000000000001"
}

func (s *SagaCoordinator) dependenciesMet(dependencies []string, state *OrchestrationState) bool {
	for _, dep := range dependencies {
		if _, ok := state.CollectedData[dep]; !ok {
			return false
		}
	}
	return true
}

func (s *SagaCoordinator) skipStep(ctx context.Context, state *OrchestrationState, reason string) error {
	s.logger.Info("Skipping step",
		zap.String("step", state.CurrentStep),
		zap.String("reason", reason))

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		StepName:  state.CurrentStep,
		Action:    "skipped",
		Timestamp: time.Now(),
		Details:   reason,
	})

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

func (s *SagaCoordinator) failWorkflow(ctx context.Context, state *OrchestrationState, errorMsg string) error {
	state.Status = StatusFailed
	state.Error = errorMsg

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		Action:    "workflow_failed",
		Timestamp: time.Now(),
		Details:   errorMsg,
	})

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState) error {

	safeDataKeys := make([]string, 0, len(state.CollectedData))
	for k := range state.CollectedData {
		safeDataKeys = append(safeDataKeys, k)
	}

	s.logger.Info("WORKFLOW_COMPLETION: Completing workflow completeWorkflow coordinator.go",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("parent_orchestration_id", state.ParentOrchestrationID),
		zap.Strings("collected_data_keys", safeDataKeys), // Just log the keys
		zap.String("owner_agent_type", state.OwnerAgentType))

	state.Status = StatusCompleted

	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   s.podName,
		Action:    "workflow_completed",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Completed after %d steps", state.ExecutionMetadata.CompletedSteps),
	})

	// After status update
	s.logger.Info("WORKFLOW_COMPLETION: Status updated",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("new_status", string(state.Status)))

	repo := NewStateRepository(s.db, s.logger)
	return repo.UpdateState(ctx, state)
}

func isLocalAction(action string) bool {
	return actions.IsLocalAction(action)
}

func getTargetAgentType(step models.Step, result map[string]interface{}) string {
	if agentType, ok := result["target_agent_type"].(string); ok {
		return agentType
	}
	if agentType, ok := step.Config["agent_type"].(string); ok {
		return agentType
	}
	return "unknown"
}

func getTimeout(step models.Step) time.Duration {
	if step.Timeout > 0 {
		return step.Timeout
	}
	return DefaultRequestTimeout
}

func extractFuelBudget(state *OrchestrationState) int {
	if fuel, ok := state.CollectedData["__fuel_budget__"].(float64); ok {
		return int(fuel)
	}
	return 1000 // Default
}

// Helper to get current and caller function names
func getFuncInfo(skip int) (current, caller string) {
	// skip=0 => this func, skip=1 => its caller, skip=2 => caller's caller, etc.
	if pc, _, _, ok := runtime.Caller(skip); ok {
		current = runtime.FuncForPC(pc).Name()
	}
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		caller = runtime.FuncForPC(pc).Name()
	}
	return
}
