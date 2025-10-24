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
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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
	DefaultRequestTimeout = 180 * time.Second
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

	// Retry tracking
	retryCounters map[string]int
	retryMutex    sync.RWMutex
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
	// fmt.Fprintf(os.Stderr, "DEBUG uuid: ExecuteWorkflow START printf - headers: %+v\n", headers)
	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG uuid: ExecuteWorkflow failed to parse headers printf: %v\n", err)
		// Create minimal context if parsing fails
		execCtx = &types.ExecutionContext{
			CorrelationID:   headers["correlation_id"],
			ClientID:        headers["client_id"],
			MessageID:       headers["message_id"],
			OrchestrationID: headers["orchestration_id"],
			Timestamp:       time.Now(),
		}
	}
	// fmt.Fprintf(os.Stderr, "DEBUG uuid: ExecuteWorkflow parsed context printf: orch=%s, parent=%s\n", execCtx.OrchestrationID, execCtx.ParentOrchestrationID)

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

func (s *SagaCoordinator) ProcessResponse(ctx context.Context, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Info("ProcessResponse in coordinator.go")

	// Prepare values for tracing
	var inResponseToParentOrchestrationID string
	var inResponseToRequest string

	if execCtx.InResponseTo != nil {
		inResponseToParentOrchestrationID = execCtx.InResponseTo.ParentOrchestrationID
		inResponseToRequest = execCtx.InResponseTo.RequestID
	}

	s.tracer.TraceMessage(execCtx, "RECEIVE_RESPONSE", execCtx.ResponsesTopic,
		map[string]interface{}{
			"consuming_agent_type":       os.Getenv("AGENT_TYPE"),
			"consuming_agent_id":         os.Getenv("AGENT_ID"),
			"response_orch_id":           execCtx.OrchestrationID,
			"in_response_to_parent_orch": inResponseToParentOrchestrationID,
			"in_response_to_request":     inResponseToRequest,
		})

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

	// Create response preview safely
	responsePreview := createResponsePreview(response.Body.Body)

	s.logger.Info("RESPONSE_RECEIVED: Processing response in coordinator")
	contextLogger := s.logger.With(
		zap.String("request_id", requestID),
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("execCtx.InResponseTo.ParentOrchestrationID", execCtx.InResponseTo.ParentOrchestrationID),
		zap.String("from_agent", execCtx.FromAgentID),
		zap.String("status", execCtx.Status),
		zap.Int("retry_version", execCtx.RetryVersion),
		zap.String("response_preview", responsePreview),
	)

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go ProcessResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	repo := NewStateRepository(s.db, s.logger)

	// ALWAYS use FindByAwaitedRequestID first - this finds the orchestration that's waiting
	state, err := repo.FindByAwaitedRequestID(ctx, requestID)
	if err != nil {
		// No orchestration is waiting for this response - it's not for us
		contextLogger.Debug("No orchestration waiting for this response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return nil
	}

	if state == nil {
		return fmt.Errorf("no state found for request_id=%s", requestID)
	}

	// Verify this state is actually waiting for this request
	if _, exists := state.AwaitedRequests[requestID]; !exists {
		contextLogger.Debug("Orchestration not waiting for this request",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("request_id", requestID))
		return nil
	}

	// Additional check: verify this orchestrator owns this orchestration
	if state.ProcessingNode != "" && state.ProcessingNode != s.podName {
		contextLogger.Debug("Response for orchestration owned by different pod, ignoring",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("owner_pod", state.ProcessingNode),
			zap.String("my_pod", s.podName))
		return nil
	}

	contextLogger.Info("RESPONSE_MATCHED: Found orchestration for response",
		zap.String("state_orch_id", state.OrchestrationID),
		zap.String("state_status", string(state.Status)),
		zap.Int("awaited_requests", len(state.AwaitedRequests)))

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
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response types.ResponseMessage) error {
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

// createResponsePreview safely creates a preview string from response body
func createResponsePreview(body interface{}) string {
	if body == nil {
		return "<nil>"
	}

	switch v := body.(type) {
	case string:
		if len(v) > 500 {
			return v[:500] + "..."
		}
		return v

	case map[string]interface{}:
		// Convert map to JSON string for preview
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("<map with %d keys>", len(v))
		}
		preview := string(jsonBytes)
		if len(preview) > 500 {
			return preview[:500] + "..."
		}
		return preview

	case []byte:
		preview := string(v)
		if len(preview) > 500 {
			return preview[:500] + "..."
		}
		return preview

	default:
		// For any other type, use fmt.Sprintf
		preview := fmt.Sprintf("%v", v)
		if len(preview) > 500 {
			return preview[:500] + "..."
		}
		return preview
	}
}

// getOrCreateState gets or creates orchestration state
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, execCtx *types.ExecutionContext, plan models.WorkflowPlan, initialData []byte) (*OrchestrationState, string, bool, error) {
	repo := NewStateRepository(s.db, s.logger)

	// fmt.Fprintf(os.Stderr, "DEBUG uuid: getOrCreateState START printf - orch=%s, parent=%s\n",
	// execCtx.OrchestrationID, execCtx.ParentOrchestrationID)

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
			ownerAgentType := s.determineOwnerAgentType(execCtx)
			ownerAgentRole := execCtx.Sender.Role

			// When creating initial state, pass both ID and Type
			err := repo.CreateInitialState(ctx, execCtx.OrchestrationID, execCtx.OrchestrationName, execCtx.CorrelationID,
				ownerAgentID, ownerAgentType, ownerAgentRole, execCtx.ParentOrchestrationID, execCtx.ClientID, plan, initialData, execCtx)

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
	ownerAgentRole := execCtx.Sender.Role

	s.logger.Info("Creating new orchestration",
		zap.String("orchestration_id", newOrchestrationID),
		zap.String("orchestration_name", newOrchestrationName),
		zap.String("correlation_id", execCtx.CorrelationID))

	err := repo.CreateInitialState(ctx, newOrchestrationID, newOrchestrationName, execCtx.CorrelationID,
		ownerAgentID, ownerAgentType, ownerAgentRole, execCtx.ParentOrchestrationID, execCtx.ClientID, plan, initialData, execCtx)

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
	callstack := s.tracer.GetCallStack(12)

	s.logger.Info("Handling orchestration status handleOrchestrationStatus coordinator.go 512",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.Any("state.Status", state.Status),
		zap.Any("state.CurrentlyExecuting", state.CurrentlyExecuting),
		zap.Any("state.CurrentStep", state.CurrentStep),
		zap.Any("call stack handleOrchestrationStatus", callstack),
	)

	repo := NewStateRepository(s.db, s.logger)

	switch state.Status {
	case StatusInitialized:
		// New orchestration, start execution
		s.logger.Info("Status Initialized Starting new orchestration execution in handleOrchestrationStatus",
			zap.String("about to execute step", state.CurrentStep))

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
		s.logger.Info("Status Executing Step in handleOrchestrationStatus")
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
		s.logger.Info("Orchestration is awaiting responses in handleOrchestrationStatus",
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
	s.logger.Info("in continueExecution",
		zap.Any("DEBUGaa: orchestration state", state),
		zap.Any("Exec context", execCtx),
	)

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
			zap.Any("state loaded from repo - in continueExecution", state),
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
		//zap.Any("state", state),
	)

	// Route to appropriate handler
	if isLocalAction(step.Action) {
		return s.executeLocalAction(ctx, state, step, execCtx)
	} else if step.Topic != "" {
		return s.executeRemoteAction(ctx, state, step, execCtx)
	}

	return fmt.Errorf("unknown action: %s", step.Action)
}

// executeLocalAction - Main entry point, orchestrates local action execution
func (s *SagaCoordinator) executeLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	s.logger.Info("just into executeLocalAction look for execCtx before it gets changed",
		zap.String("step", state.CurrentStep),
		zap.String("action", step.Action),
		zap.Any("DEBUGaa: executeLocalAction execCtx before", execCtx),
		zap.Any("before state", state),
	)

	// 1. Prepare execution context for this step
	prepareExecutionContext(execCtx, state, step, s.podName)

	// 2. Create contextual logger
	contextLogger := createActionLogger(s.logger, execCtx, state.CurrentStep, step.Action)

	contextLogger.Info("Executing local action",
		zap.Any("config", step.Config),
		zap.Any("DEBUGaa: executeLocalAction step", step),
		zap.Any("DEBUGaa: executeLocalAction state after", state),
		zap.Any("DEBUGaa: executeLocalAction execCtx after", execCtx),
	)

	// 3. Handle retry logic for spawn actions
	if shouldRetry := s.handleSpawnRetry(state, step, contextLogger); !shouldRetry {
		return fmt.Errorf("spawn action failed after max retries I think this is a maximum number of spawned agents")
	}

	// 4. Get and validate action handler
	handler, err := getActionHandler(step.Action)
	if err != nil {
		return err
	}

	// 5. Build action parameters - get input data
	params := buildActionParams(ctx, execCtx, state, step, s, contextLogger)

	s.logger.Info("Executing local action",
		zap.Any("DEBUGaa: params sent to action handler", params),
	)

	// 6. Execute the action
	result, err := executeAction(ctx, handler, params, contextLogger)
	if err != nil {
		return handleActionError(err, step, contextLogger)
	}

	// 7. Process action result
	if err := processActionResult(state, result, step, execCtx, s, contextLogger); err != nil {
		return err
	}

	// 8. Record processing history
	recordActionExecution(state, execCtx, step, s.podName)

	// 9. Save state if needed - pass the logger
	return saveStateIfNeeded(ctx, state, s.db, s.logger)
}

// Prepare the execution context for this step
func prepareExecutionContext(execCtx *types.ExecutionContext, state *OrchestrationState, step models.Step, podName string) {
	// Generate step ID if not present
	if execCtx.StepID == "" {
		execCtx.StepID = uuid.New().String()
	}

	// Set step information
	execCtx.StepName = state.CurrentStep
	execCtx.Action = step.Action
	//execCtx.RequestID = uuid.New().String()
	//execCtx.MessageID = uuid.New().String()
	execCtx.Timestamp = time.Now().UTC()

	// Set reply-to topic from environment
	if parentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC"); parentResponsesTopic != "" {
		execCtx.ReplyToTopic = parentResponsesTopic
	}

	// Ensure sender is populated
	if execCtx.Sender.AgentType == "" {
		execCtx.Sender = types.AgentIdentity{
			AgentType:    state.OwnerAgentType,
			AgentID:      state.OwnerAgentID,
			PodName:      podName,
			AgentVersion: os.Getenv("AGENT_VERSION"),
			Role:         state.OwnerAgentRole,
		}
	}

	// Ensure this is marked as a request
	execCtx.MessageType = "request"

	// Clear any response-specific fields
	execCtx.InResponseTo = nil
	execCtx.Status = ""
	execCtx.IsComplete = false
}

// Create a contextual logger for the action
func createActionLogger(baseLogger *zap.Logger, execCtx *types.ExecutionContext, stepName, action string) *zap.Logger {
	return baseLogger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", stepName),
		zap.String("action", action),
		zap.String("step_id", execCtx.StepID))
}

// Handle retry logic for spawn actions
func (s *SagaCoordinator) handleSpawnRetry(state *OrchestrationState, step models.Step, logger *zap.Logger) bool {
	if step.Action != "spawn_agent" {
		return true // Not a spawn action, continue
	}

	retryKey := fmt.Sprintf("spawn_retry_%s_%s", state.OrchestrationID, step.Name)
	retryCount := s.getRetryCount(retryKey)

	if retryCount >= 30 {
		logger.Error("Spawn action exceeded max retries",
			zap.String("step_name", state.CurrentStep),
			zap.Int("retry_count", retryCount))
		return false
	}

	s.incrementRetryCount(retryKey)
	return true
}

// Get and validate action handler
func getActionHandler(action string) (actions.ActionFunc, error) {
	handler, exists := actions.GetAction(action)
	if !exists {
		return nil, fmt.Errorf("unknown action: %s, not found in registry", action)
	}
	return handler, nil
}

// Build action parameters
func buildActionParams(ctx context.Context, execCtx *types.ExecutionContext, state *OrchestrationState,
	step models.Step, coordinator *SagaCoordinator, logger *zap.Logger) actions.ActionParams {

	logger.Info("in buildActionParams",
		zap.Any("DEBUGaa: in buildActionParams state.CollectedData", state.CollectedData), // good here in generic but not in hero
		zap.Any("DEBUGaa: in buildActionParams headers are execCtx.ToHeaders", execCtx.ToHeaders()),
		zap.Any("DEBUGaa: in buildActionParams current step", state.CurrentStep),
	)
	return actions.ActionParams{
		Context:          ctx,
		ExecutionContext: execCtx,
		StepConfig:       step,
		Headers:          execCtx.ToHeaders(),
		CollectedData:    state.CollectedData,
		Logger:           logger,
		Producer:         coordinator.producer,
		DB:               coordinator.db,
		Tracer:           coordinator.tracer,
		CurrentStep:      state.CurrentStep,
	}
}

// Execute the action handler
func executeAction(ctx context.Context, handler actions.ActionFunc, params actions.ActionParams, logger *zap.Logger) (interface{}, error) {
	logger.Info("Calling action handler",
		zap.String("action", params.StepConfig.Action))

	result, err := handler(ctx, params)
	if err != nil {
		return nil, err
	}

	logger.Info("Action handler completed - in executeAction",
		zap.String("action", params.StepConfig.Action),
		zap.Any("DEBUGaa: result", result),
	)

	return result, nil
}

// Handle action execution errors
func handleActionError(err error, step models.Step, logger *zap.Logger) error {
	logger.Error("Local action failed in handleActionError",
		zap.String("step_name", step.Name),
		zap.String("action", step.Action),
		zap.Error(err))

	// Special handling for spawn failures with topic issues
	if step.Action == "spawn_agent" && strings.Contains(err.Error(), "Unknown Topic") {
		logger.Warn("Spawn failed due to missing topic, adding delay for retry")
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("failed to execute action %s: %w", step.Action, err)
}

// Process the result from the action
func processActionResult(state *OrchestrationState, result interface{}, step models.Step,
	execCtx *types.ExecutionContext, coordinator *SagaCoordinator, logger *zap.Logger) error {

	// Store result in collected data
	if err := storeActionResult(state, result, logger); err != nil {
		return err
	}

	// Process result based on type
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Handle subtree information (from spawn actions)
		processSubtreeInfo(state, resultMap, logger)

		// Check if action requires waiting for response
		if needsWaiting := processAwaitResponse(state, resultMap, execCtx, step, coordinator, logger); needsWaiting {
			// State needs to wait for response
			state.Status = StatusAwaitingResponses
		}
	}

	return nil
}

// Store action result in collected data
func storeActionResult(state *OrchestrationState, result interface{}, logger *zap.Logger) error {
	if state.CurrentStep == "" {
		logger.Error("Cannot store result - CurrentStep is empty")
		return fmt.Errorf("current step is empty")
	}

	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}

	state.CollectedData[state.CurrentStep] = result

	logger.Info("Stored action result",
		zap.String("step", state.CurrentStep),
		zap.Any("result_keys", getMapKeys(result)),
		zap.Any("DEBUGaa: result_value", result))

	return nil
}

// Process subtree information from spawn actions
func processSubtreeInfo(state *OrchestrationState, result map[string]interface{}, logger *zap.Logger) {
	subtreeInfo, ok := result["subtree_info"].(*types.SubtreeInfo)
	if !ok {
		return
	}

	if state.SubtreeAgents == nil {
		state.SubtreeAgents = make(map[string]*types.SubtreeInfo)
	}

	state.SubtreeAgents[subtreeInfo.AgentID] = subtreeInfo

	logger.Info("Added agent to subtree",
		zap.String("agent_id", subtreeInfo.AgentID),
		zap.String("agent_type", subtreeInfo.AgentType),
		zap.String("agent_name", subtreeInfo.AgentName))
}

// Process await_response flag and setup waiting if needed
func processAwaitResponse(state *OrchestrationState, result map[string]interface{},
	execCtx *types.ExecutionContext, step models.Step, coordinator *SagaCoordinator, logger *zap.Logger) bool {

	// Check if action requires waiting
	awaitResponse, ok := result["await_response"].(bool)
	if !ok || !awaitResponse {
		return false
	}

	// Extract request ID
	requestID := extractRequestID(result, execCtx)
	if requestID == "" {
		logger.Error("No request ID for awaited response")
		return false
	}

	// Determine response topic
	responsesTopic := determineResponsesTopic(result, execCtx, logger)
	if responsesTopic == "" {
		logger.Error("No responses topic available for awaited request",
			zap.String("request_id", requestID))
		return false
	}

	// Create awaited request entry
	awaitedReq := createAwaitedRequest(requestID, execCtx, state, step, result, responsesTopic)

	// Add to state
	if state.AwaitedRequests == nil {
		state.AwaitedRequests = make(map[string]*AwaitedRequest)
	}
	state.AwaitedRequests[requestID] = awaitedReq

	logger.Info("Action requires waiting for response",
		zap.String("request_id", requestID),
		zap.String("target_agent_type", awaitedReq.TargetAgentType),
		zap.String("target_agent_id", awaitedReq.TargetAgentID),
		zap.String("responses_topic", responsesTopic),
		zap.Int("total_awaited", len(state.AwaitedRequests)))

	// Setup timeout handler
	go coordinator.handleRequestTimeout(context.Background(), state.OrchestrationID, requestID, awaitedReq.TimeoutAt)

	return true
}

// Extract request ID from result or context
func extractRequestID(result map[string]interface{}, execCtx *types.ExecutionContext) string {
	// Try to get from result first
	if reqID, ok := result["request_id"].(string); ok && reqID != "" {
		return reqID
	}
	// Fall back to execution context
	return execCtx.RequestID
}

// Determine the responses topic for waiting
func determineResponsesTopic(result map[string]interface{}, execCtx *types.ExecutionContext, logger *zap.Logger) string {
	// Priority order:
	// 1. Environment variable (most reliable)
	if topic := os.Getenv("RESPONSES_TOPIC"); topic != "" {
		logger.Info("Using RESPONSES_TOPIC from environment",
			zap.String("topic", topic))
		return topic
	}

	// 2. Result from action
	if topic, ok := result["responses_topic"].(string); ok && topic != "" {
		logger.Info("Using responses_topic from action result",
			zap.String("topic", topic))
		return topic
	}

	// 3. Execution context
	if execCtx.ResponsesTopic != "" {
		logger.Info("Using ResponsesTopic from execution context",
			zap.String("topic", execCtx.ResponsesTopic))
		return execCtx.ResponsesTopic
	}

	return ""
}

// Create an awaited request entry
func createAwaitedRequest(requestID string, execCtx *types.ExecutionContext, state *OrchestrationState,
	step models.Step, result map[string]interface{}, responsesTopic string) *AwaitedRequest {

	return &AwaitedRequest{
		RequestID:       requestID,
		StepID:          execCtx.StepID,
		StepName:        state.CurrentStep,
		RetryVersion:    0,
		TargetAgentType: extractTargetAgentType(step, result),
		TargetAgentID:   extractTargetAgentID(result),
		ResponsesTopic:  responsesTopic,
		SentAt:          time.Now(),
		TimeoutAt:       time.Now().Add(getTimeout(step)),
	}
}

// Extract target agent type from step or result
func extractTargetAgentType(step models.Step, result map[string]interface{}) string {
	// Try result first
	if agentType, ok := result["target_agent_type"].(string); ok && agentType != "" {
		return agentType
	}
	// Try step config
	if agentType, ok := step.Config["agent_type"].(string); ok && agentType != "" {
		return agentType
	}
	return "unknown"
}

// Extract target agent ID from result
func extractTargetAgentID(result map[string]interface{}) string {
	// Try multiple possible keys
	keys := []string{"agent_id", "target_agent_id", "agent_called"}
	for _, key := range keys {
		if id, ok := result[key].(string); ok && id != "" {
			return id
		}
	}
	return ""
}

// Record the action execution in processing history
func recordActionExecution(state *OrchestrationState, execCtx *types.ExecutionContext, step models.Step, podName string) {
	state.ProcessingHistory = append(state.ProcessingHistory, ProcessingRecord{
		PodName:   podName,
		StepID:    execCtx.StepID,
		StepName:  state.CurrentStep,
		Action:    "executed",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Action %s executed by %s", step.Action, podName),
	})
}

// Save state if it needs to be persisted (e.g., waiting for responses)
func saveStateIfNeeded(ctx context.Context, state *OrchestrationState, db *sql.DB, logger *zap.Logger) error {
	if state.Status != StatusAwaitingResponses {
		// State will be saved by continueExecution
		return nil
	}

	// We're waiting for responses, need to save now
	repo := NewStateRepository(db, logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to save state before waiting: %w", err)
	}

	return nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, execCtx *types.ExecutionContext) error {
	contextLogger := s.logger.With(
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", state.CurrentStep),
		zap.String("action", step.Action))

	contextLogger.Info("Executing remote action",
		zap.String("topic", step.Topic),
		zap.String("target_agent_type", step.TargetAgentType))

	// Update execution context
	execCtx.RequestID = uuid.New().String()
	execCtx.StepID = uuid.New().String()
	execCtx.StepName = state.CurrentStep
	execCtx.Action = step.Action
	execCtx.ToAgentType = step.TargetAgentType

	// ResponsesTopic should already be set in execCtx from the process() function
	if execCtx.ResponsesTopic == "" {
		contextLogger.Error("No responses topic configured",
			zap.String("step", state.CurrentStep))
		return fmt.Errorf("no responses topic configured for remote action")
	}

	// Prepare the message
	requestMsg := &types.RequestMessage{
		Headers: execCtx.ToRequestHeaders(),
		Body: models.TaskRequest{
			Action: step.Action,
			Data:   state.CollectedData,
		},
	}

	msgBytes, err := json.Marshal(requestMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send the message
	if err := s.producer.Produce(ctx, step.Topic, requestMsg.Headers.ToMap(),
		[]byte(execCtx.CorrelationID), msgBytes); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	contextLogger.Info("Remote request sent",
		zap.String("request_id", execCtx.RequestID),
		zap.String("topic", step.Topic),
		zap.String("responses_topic", execCtx.ResponsesTopic))

	// Create awaited request
	awaitedReq := &AwaitedRequest{
		RequestID:       execCtx.RequestID,
		StepID:          execCtx.StepID,
		StepName:        state.CurrentStep,
		RetryVersion:    0,
		TargetAgentType: step.TargetAgentType,
		ResponsesTopic:  execCtx.ResponsesTopic,
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

	contextLogger.Info("Remote action initiated, waiting for response",
		zap.String("request_id", execCtx.RequestID),
		zap.String("awaiting_on_topic", execCtx.ResponsesTopic))

	// Set up timeout
	go s.handleRequestTimeout(ctx, state.OrchestrationID, execCtx.RequestID, awaitedReq.TimeoutAt)

	return nil
}

func (s *SagaCoordinator) getResponsesTopicFromResult(result map[string]interface{}, execCtx *types.ExecutionContext) string {
	// First check if the action result includes a responses_topic
	// This would be set by spawn_agent or call_agent actions
	if topic, ok := result["responses_topic"].(string); ok && topic != "" {
		return topic
	}

	// For child agents, they might include the child's response topic
	if topic, ok := result["child_responses_topic"].(string); ok && topic != "" {
		return topic
	}

	// Otherwise use the execution context's topic
	// This is where THIS orchestration expects to receive responses
	if execCtx.ResponsesTopic != "" {
		return execCtx.ResponsesTopic
	}

	// This shouldn't happen with proper setup
	s.logger.Error("No responses topic available",
		zap.Any("result_keys", getMapKeys(result)),
		zap.String("exec_responses_topic", execCtx.ResponsesTopic))
	return ""
}

func (s *SagaCoordinator) getTargetAgentID(result map[string]interface{}) string {
	if agentID, ok := result["agent_id"].(string); ok {
		return agentID
	}
	if agentID, ok := result["target_agent_id"].(string); ok {
		return agentID
	}
	if agentID, ok := result["agent_called"].(string); ok {
		return agentID
	}
	return ""
}

func getMapKeys(m interface{}) []string {
	if m == nil {
		return []string{}
	}

	switch v := m.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys
	default:
		return []string{}
	}
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
func (s *SagaCoordinator) handleCompleteResponse(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Info("in handleCompleteResponse orchestrator ",
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("step_name", execCtx.StepName),
		zap.String("step_id", execCtx.StepID),
		zap.String("functional role", execCtx.FunctionalRole),
		zap.String("requestId from arguments", requestID),
		zap.Any("state.CollectedData in handleCompleteResponse is:", state.CollectedData),
		zap.Any("execCts.InResponseTo in handleCompleteResponse is:", execCtx.InResponseTo),
	)
	contextLogger := s.logger

	current, caller := getFuncInfo(1)

	contextLogger.Info("In file coordinator.go handleCompleteResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Find the awaited request
	var awaitedReq *AwaitedRequest

	// Find the awaited request
	awaitedReq, exists := state.AwaitedRequests[requestID]
	if !exists {
		s.logger.Error("No awaited request found for response",
			zap.String("request_id", requestID),
			zap.Any("state.AwaitedRequests (awaited_request_ids)", state.AwaitedRequests),
		)
		return fmt.Errorf("no awaited request found for request_id: %s", requestID)
	}

	s.logger.Info("Found awaited request",
		zap.String("step_name", awaitedReq.StepName),
		zap.String("step_id", awaitedReq.StepID))

	for rid, req := range state.AwaitedRequests {
		if req.RequestID == response.Headers.InResponseToRequestID {
			awaitedReq = req
			requestID = rid
			break
		}
	}

	// Parse response body - handle flexible structure
	var responseBodyData map[string]interface{}

	// Response.Body might be already a map or might be raw JSON
	switch bodyData := response.Body.Body.(type) {
	case map[string]interface{}:
		responseBodyData = bodyData
		s.logger.Info("Response body is already a map")

	case []byte:
		if err := json.Unmarshal(bodyData, &responseBodyData); err != nil {
			s.logger.Error("Failed to unmarshal response body bytes", zap.Error(err))
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		s.logger.Info("Unmarshaled response body from bytes")

	case string:
		if err := json.Unmarshal([]byte(bodyData), &responseBodyData); err != nil {
			s.logger.Error("Failed to unmarshal response body string", zap.Error(err))
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		s.logger.Info("Unmarshaled response body from string")

	default:
		// Try to marshal and unmarshal to get it into map form
		jsonBytes, err := json.Marshal(bodyData)
		if err != nil {
			s.logger.Error("Failed to marshal response body", zap.Error(err))
			return fmt.Errorf("failed to marshal response body: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &responseBodyData); err != nil {
			s.logger.Error("Failed to unmarshal marshaled response body", zap.Error(err))
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		s.logger.Info("Converted response body to map via marshal/unmarshal")
	}

	// Normalize the response data before storing
	normalisedData := datahelpers.NormalizeResponseData(response.Body, s.logger)

	// Store under the step name in CollectedData
	stepName := awaitedReq.StepName
	state.CollectedData[stepName] = normalisedData

	s.logger.Info("handleCompleteResponse: stored normalized response",
		zap.String("step_name", stepName),
		zap.Int("original_fields", len(responseBodyData)),
		zap.Int("normalized_fields", len(normalisedData)),
		zap.Strings("normalised data_keys", getMapKeys(normalisedData)),
		zap.Any("DEBUGaa: normalised data", normalisedData),
	)

	// Remove from awaited requests
	delete(state.AwaitedRequests, requestID)

	// ALSO find the step in CollectedData and update its metadata
	// This preserves any existing step data while adding response info
	for existingStepName, stepData := range state.CollectedData {
		if existingStepName == "" {
			s.logger.Warn("Skipping empty step name in CollectedData")
			continue
		}

		if stepMap, ok := stepData.(map[string]interface{}); ok {
			// Check if this step data contains our request_id
			if storedReqID, exists := stepMap["request_id"]; exists && storedReqID == requestID {
				// Update the existing step data with response info
				stepMap["response"] = normalisedData
				stepMap["response_received_at"] = time.Now().Format(time.RFC3339)
				stepMap["response_status"] = "complete"

				s.logger.Info("Updated step metadata with response",
					zap.String("step_name", existingStepName),
					zap.String("request_id", requestID))
				break
			}
		}
	}

	s.logger.Info("handleCompleteResponse: removed from awaited requests",
		zap.String("request_id", requestID),
		zap.Int("remaining_awaited", len(state.AwaitedRequests)))

	// Update state in database
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		return fmt.Errorf("failed to save response data: %w", err)
	}

	// Check if all responses received
	if len(state.AwaitedRequests) == 0 {
		// If no more awaited requests, continue workflow
		s.logger.Info("All responses received, continuing workflow")

		// ADVANCE TO NEXT STEP
		currentStep := state.WorkflowPlan.Steps[state.CurrentStep]
		if currentStep.NextStep != "" {
			state.CurrentStep = currentStep.NextStep
			repo.UpdateState(ctx, state) // Save the step advancement
		}

		s.logger.Info("All responses received - steps",
			zap.Any("current step_object", currentStep),
			zap.String("next, (now current) step_name", state.CurrentStep),
		)

		var stateExecCtx types.ExecutionContext
		if execCtxData, ok := state.CollectedData["__execution_context__"].(map[string]interface{}); ok {
			// Marshal the map to JSON, then unmarshal to ExecutionContext
			execCtxBytes, err := json.Marshal(execCtxData)
			if err != nil {
				s.logger.Error("Failed to marshal execution context map", zap.Error(err))
			} else {
				if err := json.Unmarshal(execCtxBytes, &stateExecCtx); err != nil {
					s.logger.Error("Failed to unmarshal execution context", zap.Error(err))
				} else {
					s.logger.Info("Successfully extracted execution context from CollectedData",
						zap.String("request_id", stateExecCtx.RequestID),
						zap.String("reply_to_request_id", stateExecCtx.ReplyToRequestID))
				}
			}
		} else {
			s.logger.Warn("__execution_context__ not found or wrong type in CollectedData",
				zap.String("type", fmt.Sprintf("%T", state.CollectedData["__execution_context__"])))
		}
		s.logger.Info("Execution Context from Collected Data",
			zap.Any("DEBUGaa: Execution context from collected data", stateExecCtx))

		// Create fresh execution context for continuing
		freshExecCtx := &types.ExecutionContext{
			CorrelationID:   state.CorrelationID,
			OrchestrationID: state.OrchestrationID,
			ClientID:        state.ClientID,

			// Reset to request mode
			MessageType:      "request",
			MessageID:        uuid.New().String(),
			ReplyToRequestID: stateExecCtx.RequestID,

			// Set sender to the current orchestrator
			Sender: types.AgentIdentity{
				AgentType:    state.OwnerAgentType,
				AgentID:      state.OwnerAgentID,
				PodName:      s.podName,
				AgentVersion: os.Getenv("AGENT_VERSION"),
			},

			// Clear any response fields
			InResponseTo: nil, //?
			Status:       "",
			IsComplete:   false,

			// Resources
			FuelBudget:     state.FuelBudget,
			TimeoutSeconds: 30,
			Timestamp:      time.Now(),
			Version:        "2.0",
		}

		state.LastActivity = time.Now()

		state.Status = StatusExecutingStep
		state.AwaitedRequests = make(map[string]*AwaitedRequest)

		if err := repo.UpdateState(ctx, state); err != nil {
			return fmt.Errorf("failed to update state after clearing awaited requests after responses: %w", err)
		}

		s.logger.Info("handleCompleteResponse: all responses received continuing workflow",
			zap.Any("DEBUGaa: fresh exec ctx:", freshExecCtx),
		)

		return s.continueExecution(ctx, state, freshExecCtx)
	}

	s.logger.Info("handleCompleteResponse: still awaiting responses",
		zap.Int("remaining", len(state.AwaitedRequests)))

	return nil

	/*// Parse response with flexible structure handling
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to unmarshal response: %v", err))
	}

	contextLogger.Debug("Raw response structure", zap.Any("raw", rawResponse))

	// Extract the actual data - simpler logic!
	var responseData interface{}

	// Check if there's a "body" field
	if body, ok := rawResponse["body"].(map[string]interface{}); ok {
		// Check if there's a nested body (from CompleteWorkflowAction responses)
		if innerBody, ok := body["body"].(map[string]interface{}); ok {
			responseData = innerBody
		} else if process, ok := body["process"].(map[string]interface{}); ok {
			// If body contains "process" with actual result, use that
			responseData = process
		} else {
			// Otherwise use the whole body
			responseData = body
		}
	} else {
		// No "body" field - use entire response as data
		responseData = rawResponse
	}

	// Store the response data
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}

	contextLogger.Info("DEBUG: CollectedData keys before searching for request",
		zap.Any("keys", getMapKeys(state.CollectedData)))

	// Find the step that owns this request and store the response there
	for stepName, stepData := range state.CollectedData {
		if stepName == "" {
			contextLogger.Warn("Skipping empty step name in CollectedData")
			continue
		}
		if stepMap, ok := stepData.(map[string]interface{}); ok {
			if storedReqID, exists := stepMap["request_id"]; exists && storedReqID == requestID {
				stepMap["response"] = responseData
				stepMap["response_received_at"] = time.Now()
				stepMap["response_status"] = "complete"

				contextLogger.Info("Stored response in parent step",
					zap.String("step_name", stepName),
					zap.String("request_id", requestID),
					zap.Any("stepData", stepMap),
					zap.Any("response_data :", responseData))
				break
			}
		}
	}*/

	// Handle special case for spawn_agent responses
	/*if awaitedReq, exists := state.AwaitedRequests[requestID]; exists {

		stepName := awaitedReq.StepName

		// Store the response under the step that initiated it
		if stepData, ok := state.CollectedData[stepName].(map[string]interface{}); ok {
			stepData["response"] = responseData
			stepData["response_received_at"] = time.Now()
			stepData["response_status"] = "complete"

			if _, hasReqID := stepData["request_id"]; !hasReqID {
				stepData["request_id"] = requestID
			}

			contextLogger.Info("Stored response under step",
				zap.String("step_name", stepName),
				zap.String("request_id", requestID),
				zap.Any("response_data", responseData),
			)
		} else {
			// If the step data doesn't exist or isn't a map, create it
			state.CollectedData[stepName] = map[string]interface{}{
				"response":             responseData,
				"response_received_at": time.Now(),
				"response_status":      "complete",
				"request_id":           requestID,
			}

			contextLogger.Info("Created step data and stored response",
				zap.String("step_name", stepName),
				zap.Any("response_data", responseData))
		}

		if awaitedReq.StepName == "spawn_agent" {
			s.logger.Info("In handleCompleteResponse, stepname in awaited requests is spawn_agent")
			// Extract agent info from response
			if agentID, ok := responseData.(map[string]interface{})["agent_id"]; ok {
				state.CollectedData["spawn_calculator"] = map[string]interface{}{
					"agent_id":   agentID,
					"agent_name": responseData.(map[string]interface{})["agent_name"],
					"agent_type": responseData.(map[string]interface{})["agent_type"],
					"status":     "initialized",
					"spawned_at": time.Now(),
				}
			}
		}
	}*/

	/*	repo := NewStateRepository(s.db, s.logger)
		if err := repo.UpdateStateWithVersion(ctx, state); err != nil {
			return fmt.Errorf("failed to save response data: %w", err)
		}*/

	/*// Remove from awaited requests
	if err := repo.RemoveAwaitedRequest(ctx, state.OrchestrationID, requestID); err != nil {
		return err
	}*/

	/*	// Reload state
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

		return nil*/
}

// handleRecoverableError handles errors that can be retried
func (s *SagaCoordinator) handleRecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Warn("Recoverable error received",
		zap.String("request_id", requestID),
		zap.Int("retry_version", execCtx.RetryVersion),
	)

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
	state.AwaitedRequests[requestID] = awaited

	topic := os.Getenv("PARENT_RESPONSES_TOPIC")
	if topic == "" {
		topic = execCtx.ReplyToTopic
		if execCtx.ReplyToTopic == "" {
			topic = execCtx.ResponsesTopic
		}
	}

	s.logger.Info("Retrying request",
		zap.String("request_id", requestID),
		zap.Int("retry_version", awaited.RetryVersion),
		zap.String("DEBUGaa: where to send this? execCtx.ReplyToTopic:", execCtx.ReplyToTopic),
		zap.String("DEBUGaa: where to send this? execCtx.ResponseTopic:", execCtx.ResponsesTopic),
		zap.String("DEBUGaa: os.Getenv(PARENT_RESPONSES_TOPIC)", os.Getenv("PARENT_RESPONSES_TOPIC")),
	)

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
	retryBytes, _ := json.Marshal(retryRequest)

	return s.producer.Produce(ctx, topic, retryRequest.Headers.ToMap(), []byte(requestID), retryBytes)
}

// handleUnrecoverableError handles fatal errors
func (s *SagaCoordinator) handleUnrecoverableError(ctx context.Context, state *OrchestrationState, requestID string, execCtx *types.ExecutionContext, response types.ResponseMessage) error {
	s.logger.Error("Unrecoverable error received",
		zap.String("request_id", requestID))

	// Extract error message from the ResponseMessage structure
	errorMsg := "Unknown error"

	// Check if there's an Error field in the ResponseBody
	if response.Body.Error != nil {
		errorMsg = response.Body.Error.Message
		if response.Body.Error.Code != "" {
			errorMsg = fmt.Sprintf("%s (code: %s)", errorMsg, response.Body.Error.Code)
		}
	} else if response.Body.Body != nil {
		// If no Error field, check the Body field for error information
		switch body := response.Body.Body.(type) {
		case map[string]interface{}:
			if errStr, ok := body["error"].(string); ok {
				errorMsg = errStr
			}
		case string:
			errorMsg = body
		}
	}

	return s.failWorkflow(ctx, state, fmt.Sprintf("Request %s failed: %s", requestID, errorMsg))
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
		s.logger.Error("Request timed out",
			zap.String("request_id", requestID),
			zap.Int("retry_version", awaited.RetryVersion),
		)

		// Retry or fail
		if awaited.RetryVersion < 3 {
			// Create minimal ExecutionContext for retry
			execCtx := &types.ExecutionContext{
				RequestID: requestID,
				Status:    "error_recoverable",
			}

			// Create a timeout response message instead of raw bytes
			timeoutResponse := types.ResponseMessage{
				Headers: types.ResponseHeaders{
					InResponseToRequestID: requestID,
					InResponseToStepID:    awaited.StepID,
					InResponseToStepName:  awaited.StepName,
					InResponseToAction:    "unknown",
					Status:                "error_recoverable",
					IsError:               true,
					MessageType:           "response",
					TimeSent:              time.Now(),
					OrchestrationID:       orchestrationID,
					CorrelationID:         state.CorrelationID,
				},
				Body: types.ResponseBody{
					Success: false,
					Error: &types.ErrorInfo{
						Code:        "TIMEOUT",
						Message:     "Request timed out",
						Recoverable: true,
						Details: map[string]interface{}{
							"timeout_after": timeoutAt.Sub(awaited.TimeoutAt).String(),
							"retry_count":   awaited.RetryVersion,
						},
					},
				},
			}

			s.handleRecoverableError(ctx, state, requestID, execCtx, timeoutResponse)
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

// Helper methods for retry tracking (add these to SagaCoordinator struct)
func (s *SagaCoordinator) getRetryCount(key string) int {
	// Simple in-memory tracking - you might want to use Redis or database
	s.retryMutex.RLock()
	defer s.retryMutex.RUnlock()

	if s.retryCounters == nil {
		return 0
	}
	return s.retryCounters[key]
}

func (s *SagaCoordinator) incrementRetryCount(key string) {
	s.retryMutex.Lock()
	defer s.retryMutex.Unlock()

	if s.retryCounters == nil {
		s.retryCounters = make(map[string]int)
	}
	s.retryCounters[key]++
}
