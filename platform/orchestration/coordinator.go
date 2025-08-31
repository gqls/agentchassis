// FILE: platform/orchestration/coordinator.go (enhanced with logging)
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
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
	// whether to log hefty messages and headers
	LogMessageDetails = true
)

// SagaCoordinator manages the execution of complex workflows
type SagaCoordinator struct {
	db          *sql.DB
	producer    kafka.Producer
	logger      *zap.Logger
	fuelManager *governance.FuelManager
	tracer      *TraceLogger
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
	"complete_workflow":   actions.CompleteWorkflowAction,

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

	"evaluate_task": actions.EvaluateTaskAction,
}

// NewSagaCoordinator creates a new coordinator instance
func NewSagaCoordinator(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *SagaCoordinator {
	return &SagaCoordinator{
		db:          db,
		producer:    producer,
		logger:      logger,
		fuelManager: governance.NewFuelManager(),
		tracer:      NewTraceLogger(logger),
	}
}

// ExecuteWorkflow now stores the plan and continues execution
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	correlationID := headers["correlation_id"]
	messageID := headers["message_id"]
	clientID := headers["client_id"]

	l := s.logger.With(
		zap.String("correlation_id", correlationID),
		zap.String("message_id", messageID))

	l.Info("ExecuteWorkflow called",
		zap.String("start_step", plan.StartStep),
		zap.Int("total_steps", len(plan.Steps)))

	if clientID == "" {
		l.Error("client_id header is required to execute a workflow")
		return fmt.Errorf("client_id header is required to execute a workflow")
	}

	// Step 1: Check message deduplication
	repo := NewStateRepository(s.db, s.logger)

	// If no message_id, generate one
	if messageID == "" {
		messageID = uuid.New().String()
		headers["message_id"] = messageID
	}

	// Check if we've already processed this message
	alreadyProcessed, err := repo.HasProcessedMessage(ctx, messageID)
	if err != nil {
		l.Error("Failed to check message deduplication", zap.Error(err))
		// Continue anyway - better to risk duplicate than fail
	} else if alreadyProcessed {
		l.Info("Message already processed, skipping",
			zap.String("message_id", messageID))
		return nil // Already handled
	}

	// Step 2: Get or create orchestration state
	state, orchestrationID, isNewlyCreated, err := s.getOrCreateStateWithFlag(ctx, correlationID, clientID, plan, initialData, headers)
	if err != nil {
		l.Error("Failed to get or create state", zap.Error(err))
		return err
	}

	// Record that we're processing this message
	if err := repo.RecordMessageProcessing(ctx, messageID, correlationID, orchestrationID); err != nil {
		l.Error("Failed to record message processing", zap.Error(err))
		// Continue anyway
	}

	// Set orchestration_id in headers for all subsequent operations
	headers["orchestration_id"] = orchestrationID
	headers["owner_agent_id"] = state.OwnerAgentID

	l.Info("Workflow state retrieved",
		zap.String("orchestration_id", orchestrationID),
		zap.String("status", string(state.Status)),
		zap.String("current_step", state.CurrentStep),
		zap.Bool("newly_created", isNewlyCreated))

	// Step 3: Handle based on current state
	switch state.Status {
	case StatusInitialized:
		// New orchestration, start execution
		l.Info("Starting new orchestration execution")

		// Mark as executing
		if err := repo.SetExecutingStep(ctx, orchestrationID, state.CurrentStep); err != nil {
			l.Error("Failed to set executing step", zap.Error(err))
			return err
		}

		// Reload state to get updated version
		state, err = repo.GetState(ctx, orchestrationID)
		if err != nil {
			return fmt.Errorf("failed to reload state after setting executing: %w", err)
		}

		return s.continueExecution(ctx, state, headers)

	case StatusExecutingStep:
		// Check if it's stuck
		if state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout {
			l.Warn("Found stuck orchestration, taking over",
				zap.String("stuck_step", *state.CurrentlyExecuting),
				zap.Duration("stuck_for", time.Since(state.LastActivity)))

			// Clear the stuck execution
			if err := repo.ClearExecutingStep(ctx, orchestrationID); err != nil {
				l.Error("Failed to clear stuck execution", zap.Error(err))
				return err
			}

			// Reload and continue
			state, err = repo.GetState(ctx, orchestrationID)
			if err != nil {
				return fmt.Errorf("failed to reload state: %w", err)
			}

			return s.continueExecution(ctx, state, headers)
		}

		// Still actively executing
		l.Info("Orchestration is actively executing, skipping duplicate",
			zap.String("executing_step", *state.CurrentlyExecuting),
			zap.String("processing_node", state.ProcessingNode))
		return nil

	case StatusAwaitingResponses:
		l.Info("Orchestration is awaiting responses, skipping duplicate",
			zap.Int("awaited_count", len(state.AwaitedSteps)))
		return nil

	case StatusCompleted:
		l.Info("Workflow already completed")
		return nil

	case StatusFailed:
		l.Info("Workflow previously failed",
			zap.String("error", state.Error))
		return nil

	default:
		l.Error("Unknown orchestration status",
			zap.String("status", string(state.Status)))
		return fmt.Errorf("unknown orchestration status: %s", state.Status)
	}
}

// getOrCreateStateWithFlag returns orchestration with a flag indicating if it was just created
func (s *SagaCoordinator) getOrCreateStateWithFlag(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte, headers map[string]string) (*OrchestrationState, string, bool, error) {
	repo := NewStateRepository(s.db, s.logger)

	// Priority 1: If we have an explicit orchestration_id, use it
	orchestrationID := headers["orchestration_id"]
	if orchestrationID != "" {
		state, err := repo.GetState(ctx, orchestrationID)
		if err == nil {
			s.logger.Info("Found existing state by orchestration_id",
				zap.String("orchestration_id", orchestrationID),
				zap.String("status", string(state.Status)))
			return state, orchestrationID, false, nil
		}

		// If we have an orchestration_id but no state, this is likely a new child orchestration
		if headers["parent_orchestration_id"] != "" {
			s.logger.Info("Creating child orchestration with explicit ID",
				zap.String("orchestration_id", orchestrationID),
				zap.String("parent_orchestration_id", headers["parent_orchestration_id"]))

			ownerAgentID := s.determineOwnerAgentID(headers)

			err := repo.CreateInitialState(ctx, orchestrationID, correlationID, ownerAgentID,
				headers["parent_orchestration_id"], clientID, plan, initialData)

			if err != nil {
				// Check if someone else created it
				if strings.Contains(err.Error(), "duplicate key") {
					state, err = repo.GetState(ctx, orchestrationID)
					if err == nil {
						return state, orchestrationID, false, nil
					}
				}
				return nil, "", false, fmt.Errorf("failed to create child orchestration: %w", err)
			}

			// Get the newly created state
			state, err = repo.GetState(ctx, orchestrationID)
			if err != nil {
				return nil, "", false, fmt.Errorf("failed to get newly created state: %w", err)
			}
			return state, orchestrationID, true, nil
		}
	}

	// Priority 2: For root orchestrations only, try correlation_id lookup
	if headers["parent_orchestration_id"] == "" {
		state, err := repo.GetStateByCorrelation(ctx, correlationID)
		if err == nil {
			s.logger.Info("Found existing root orchestration by correlation_id",
				zap.String("correlation_id", correlationID),
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("status", string(state.Status)))
			return state, state.OrchestrationID, false, nil
		}
	}

	// Priority 3: Create new orchestration
	newOrchestrationID := orchestrationID
	if newOrchestrationID == "" {
		newOrchestrationID = uuid.New().String()
	}

	ownerAgentID := s.determineOwnerAgentID(headers)
	parentOrchestrationID := headers["parent_orchestration_id"]

	s.logger.Info("Creating new orchestration",
		zap.String("orchestration_id", newOrchestrationID),
		zap.String("correlation_id", correlationID),
		zap.String("parent_orchestration_id", parentOrchestrationID),
		zap.Bool("is_child", parentOrchestrationID != ""))

	err := repo.CreateInitialState(ctx, newOrchestrationID, correlationID, ownerAgentID,
		parentOrchestrationID, clientID, plan, initialData)

	if err != nil {
		// Handle race condition
		if strings.Contains(err.Error(), "duplicate key") {
			// Try to find the existing one
			if parentOrchestrationID == "" {
				state, err := repo.GetStateByCorrelation(ctx, correlationID)
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

	// Get the newly created state
	state, err := repo.GetState(ctx, newOrchestrationID)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to get newly created state: %w", err)
	}

	return state, newOrchestrationID, true, nil
}

// handleStartOrchestration handles child orchestration starts
func (s *SagaCoordinator) handleStartOrchestration(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Executing start_orchestration with wait-for-response")

	_, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	// If executeLocalAction already handled the waiting state, just return
	if state.Status == StatusAwaitingResponses {
		// Timeout already set up by executeLocalAction
		return nil
	}

	// Only handle non-waiting results
	return nil
}

// StartStuckOrchestrationRecovery runs periodically to recover stuck orchestrations
func (s *SagaCoordinator) StartStuckOrchestrationRecovery(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.recoverStuckOrchestrations(ctx)
		}
	}
}

func (s *SagaCoordinator) recoverStuckOrchestrations(ctx context.Context) {
	repo := NewStateRepository(s.db, s.logger)

	stuckIDs, err := repo.CheckStuckOrchestrations(ctx, StuckOrchestrationTimeout)
	if err != nil {
		s.logger.Error("Failed to check for stuck orchestrations", zap.Error(err))
		return
	}

	for _, orchestrationID := range stuckIDs {
		s.logger.Info("Recovering stuck orchestration",
			zap.String("orchestration_id", orchestrationID))

		state, err := repo.GetState(ctx, orchestrationID)
		if err != nil {
			s.logger.Error("Failed to get stuck orchestration state",
				zap.String("orchestration_id", orchestrationID),
				zap.Error(err))
			continue
		}

		// Clear the stuck execution
		if err := repo.ClearExecutingStep(ctx, orchestrationID); err != nil {
			s.logger.Error("Failed to clear stuck execution",
				zap.String("orchestration_id", orchestrationID),
				zap.Error(err))
			continue
		}

		// Mark as failed with timeout error
		s.failWorkflow(ctx, state, fmt.Sprintf("Step '%s' timed out after %v",
			*state.CurrentlyExecuting, StuckOrchestrationTimeout))
	}
}

// continueExecution executes from the current step
func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, headers map[string]string) error {
	repo := NewStateRepository(s.db, s.logger)

	l := s.logger.With(
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("correlation_id", state.CorrelationID),
		zap.String("current_step", state.CurrentStep))

	l.Info("Continuing workflow execution",
		zap.String("current_step", state.CurrentStep),
		zap.Int("total_steps", len(state.WorkflowPlan.Steps)))

	// Mark the step as executing
	if err := repo.SetExecutingStep(ctx, state.OrchestrationID, state.CurrentStep); err != nil {
		l.Error("Failed to mark step as executing", zap.Error(err))
		return err
	}

	// Reload state to get updated version
	state, err := repo.GetState(ctx, state.OrchestrationID)
	if err != nil {
		return fmt.Errorf("failed to reload state: %w", err)
	}

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

// getOrCreateState returns orchestrationID explicitly
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte, headers map[string]string) (*OrchestrationState, string, error) {
	repo := NewStateRepository(s.db, s.logger)

	// Priority 1: If we have an explicit orchestration_id, use it
	orchestrationID := headers["orchestration_id"]
	if orchestrationID != "" {
		state, err := repo.GetState(ctx, orchestrationID)
		if err == nil {
			s.logger.Info("Found existing state by orchestration_id",
				zap.String("orchestration_id", orchestrationID),
				zap.String("status", string(state.Status)))

			// Check if this is a duplicate request that we should ignore
			if state.Status == StatusRunning || state.Status == StatusAwaitingResponses {
				// Already processing, return existing state
				return state, orchestrationID, nil
			}

			// If completed or failed, also return existing state
			if state.Status == StatusCompleted || state.Status == StatusFailed {
				return state, orchestrationID, nil
			}

			return state, orchestrationID, nil
		}

		// If we have an orchestration_id but no state, this is likely a new child orchestration
		if headers["parent_orchestration_id"] != "" {
			s.logger.Info("Creating child orchestration with explicit ID",
				zap.String("orchestration_id", orchestrationID),
				zap.String("parent_orchestration_id", headers["parent_orchestration_id"]))

			ownerAgentID := s.determineOwnerAgentID(headers)

			// Try to create with retry on conflict
			for attempts := 0; attempts < 3; attempts++ {
				err := repo.CreateInitialState(ctx, orchestrationID, correlationID, ownerAgentID,
					headers["parent_orchestration_id"], clientID, plan, initialData)

				if err == nil {
					// Successfully created
					state, err = repo.GetState(ctx, orchestrationID)
					if err != nil {
						return nil, "", fmt.Errorf("failed to get newly created state: %w", err)
					}
					return state, orchestrationID, nil
				}

				// Check if it's a duplicate key error
				if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
					// Someone else created it, try to get it
					state, err = repo.GetState(ctx, orchestrationID)
					if err == nil {
						s.logger.Info("Another process created the state, using existing",
							zap.String("orchestration_id", orchestrationID))
						return state, orchestrationID, nil
					}
					// If we can't get it, retry
					time.Sleep(time.Duration(attempts*100) * time.Millisecond)
					continue
				}

				// Some other error
				return nil, "", fmt.Errorf("failed to create child orchestration: %w", err)
			}

			return nil, "", fmt.Errorf("failed to create state after retries")
		}
	}

	// Priority 2: For root orchestrations only, try correlation_id lookup
	if headers["parent_orchestration_id"] == "" {
		state, err := repo.GetStateByCorrelation(ctx, correlationID)
		if err == nil {
			s.logger.Info("Found existing root orchestration by correlation_id",
				zap.String("correlation_id", correlationID),
				zap.String("orchestration_id", state.OrchestrationID),
				zap.String("status", string(state.Status)))

			// Check status to avoid duplicate processing
			if state.Status == StatusRunning || state.Status == StatusAwaitingResponses {
				return state, state.OrchestrationID, nil
			}

			return state, state.OrchestrationID, nil
		}
	}

	// Priority 3: Create new orchestration with retry logic
	newOrchestrationID := orchestrationID
	if newOrchestrationID == "" {
		newOrchestrationID = uuid.New().String()
	}

	ownerAgentID := s.determineOwnerAgentID(headers)
	parentOrchestrationID := headers["parent_orchestration_id"]

	s.logger.Info("Creating new orchestration",
		zap.String("orchestration_id", newOrchestrationID),
		zap.String("correlation_id", correlationID),
		zap.String("parent_orchestration_id", parentOrchestrationID),
		zap.Bool("is_child", parentOrchestrationID != ""))

	// Try to create with retry on conflict
	for attempts := 0; attempts < 3; attempts++ {
		err := repo.CreateInitialState(ctx, newOrchestrationID, correlationID, ownerAgentID,
			parentOrchestrationID, clientID, plan, initialData)

		if err == nil {
			// Successfully created
			state, err := repo.GetState(ctx, newOrchestrationID)
			if err != nil {
				return nil, "", fmt.Errorf("failed to get newly created state: %w", err)
			}
			return state, newOrchestrationID, nil
		}

		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			// For root orchestrations, try to find by correlation
			if parentOrchestrationID == "" {
				state, err := repo.GetStateByCorrelation(ctx, correlationID)
				if err == nil {
					s.logger.Info("Another process created the state, using existing",
						zap.String("correlation_id", correlationID),
						zap.String("orchestration_id", state.OrchestrationID))
					return state, state.OrchestrationID, nil
				}
			} else {
				// For child orchestrations, try by orchestration ID
				state, err := repo.GetState(ctx, newOrchestrationID)
				if err == nil {
					s.logger.Info("Another process created the child state, using existing",
						zap.String("orchestration_id", newOrchestrationID))
					return state, newOrchestrationID, nil
				}
			}

			// If we still can't find it, generate a new ID and retry
			newOrchestrationID = uuid.New().String()
			s.logger.Info("Retrying with new orchestration ID",
				zap.String("new_orchestration_id", newOrchestrationID),
				zap.Int("attempt", attempts+1))
			time.Sleep(time.Duration(attempts*100) * time.Millisecond)
			continue
		}

		// Some other error
		return nil, "", fmt.Errorf("failed to create orchestration: %w", err)
	}

	return nil, "", fmt.Errorf("failed to create state after retries")
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
		zap.String("action", step.Action),
		zap.String("step", state.CurrentStep),
		zap.Int("fuel_available", execCtx.FuelBudget))

	// Pre-generate request ID for actions that will need responses
	var preGeneratedRequestID string
	needsResponse := step.Action == "call_agent" || step.Action == "start_orchestration" ||
		step.Action == "spawn_group" || step.Action == "spawn_agent"

	if needsResponse {
		preGeneratedRequestID = uuid.New().String()
		headers["request_id"] = preGeneratedRequestID

		contextLogger.Info("Pre-generated request ID for action",
			zap.String("action", step.Action),
			zap.String("request_id", preGeneratedRequestID))
	}

	// Special handling for spawn_group
	if step.Action == "spawn_group" {
		// Ensure we have the group type
		if groupType, ok := step.Config["group_type"].(string); ok {
			headers["group_type"] = groupType
		}
		headers["parent_orchestration_id"] = state.OrchestrationID
		headers["reply_to_topic"] = fmt.Sprintf("system.agent.%s.responses", execCtx.OwnerAgentType)
	}

	// For child orchestration actions
	if step.Action == "start_orchestration" {
		childOrchestrationID := uuid.New().String()
		headers["child_orchestration_id"] = childOrchestrationID
		headers["parent_orchestration_id"] = state.OrchestrationID
		headers["reply_to_topic"] = fmt.Sprintf("system.agent.%s.responses", execCtx.OwnerAgentType)

		contextLogger.Info("Prepared child orchestration",
			zap.String("parent_orch", state.OrchestrationID),
			zap.String("child_orch", childOrchestrationID))
	}

	handler, ok := actionRegistry[step.Action]
	if !ok {
		return nil, fmt.Errorf("local action '%s' not found in registry", step.Action)
	}

	// Convert to headers for action params
	actionHeaders := execCtx.ToHeaders()
	for k, v := range headers {
		actionHeaders[k] = v
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
		AgentType:       execCtx.FromAgentType,
		CurrentStep:     state.CurrentStep,
	}

	// Execute the action
	result, err := handler(ctx, params)
	if err != nil {
		contextLogger.Error("Local action failed", zap.Error(err))
		return nil, fmt.Errorf("local action failed: %w", err)
	}

	// Handle await_response results
	if resultMap, ok := result.(map[string]interface{}); ok {
		if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
			// Use the pre-generated request ID
			requestID := preGeneratedRequestID
			if resultRequestID, ok := resultMap["request_id"].(string); ok && resultRequestID != "" {
				requestID = resultRequestID
			}

			if requestID != "" {
				contextLogger.Info("Action requires waiting for response",
					zap.String("action", step.Action),
					zap.String("request_id", requestID),
					zap.String("orchestration_id", state.OrchestrationID))

				// Store result
				state.CollectedData[state.CurrentStep] = result

				// CRITICAL: Use the repository to add the awaited request
				repo := NewStateRepository(s.db, s.logger)

				// First, update the state to AWAITING_RESPONSES
				state.Status = StatusAwaitingResponses
				state.AwaitedSteps = []string{requestID}

				// Update metadata
				state.ExecutionMetadata.CompletedSteps++
				state.ExecutionMetadata.Checkpoints[step.Action] = time.Now().UTC()

				// Save the state with the awaited request
				if err := repo.UpdateState(ctx, state); err != nil {
					contextLogger.Error("Failed to update state with awaited request",
						zap.Error(err),
						zap.String("request_id", requestID))
					return result, fmt.Errorf("failed to update state for waiting: %w", err)
				}

				// Trace the awaited steps update
				if s.tracer != nil {
					s.tracer.TraceAwaitedSteps(execCtx, state.AwaitedSteps, step.Action)
				}

				contextLogger.Info("Successfully added request to awaited steps",
					zap.String("request_id", requestID),
					zap.String("orchestration_id", state.OrchestrationID),
					zap.String("pregenerated request id", preGeneratedRequestID),
					zap.Strings("all_awaited_steps", state.AwaitedSteps),
					zap.String("status", string(state.Status)))

				// Set up timeout for child orchestrations
				if step.Action == "start_orchestration" {
					timeout := 5 * time.Minute
					if step.Timeout > 0 {
						timeout = step.Timeout
					}
					go s.handleChildOrchestrationTimeout(ctx, state.OrchestrationID, requestID, timeout)
				}

				return result, nil
			} else {
				contextLogger.Error("No request ID available for waiting action",
					zap.String("action", step.Action))
			}
		}
	}

	// Store result and update metadata for non-waiting actions
	state.CollectedData[state.CurrentStep] = result
	state.ExecutionMetadata.CompletedSteps++
	state.ExecutionMetadata.Checkpoints[step.Action] = time.Now().UTC()

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
	s.logger.Info("CRITICAL: HandleResponse Entry",
		zap.String("in_response_to", headers["in_response_to"]),
		zap.String("causation_id", headers["causation_id"]),
		zap.String("request_id", headers["request_id"]),
		zap.String("orchestration_id", headers["orchestration_id"]))

	// Create ExecutionContext from headers
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		s.logger.Error("Failed to create ExecutionContext from headers", zap.Error(err))
		return err
	}

	if s.tracer != nil {
		s.tracer.TraceMessage(execCtx, "processing_response", "", response)
	}

	contextLogger := s.logger.With(execCtx.LogContext()...)

	// The key is in_response_to - this tells us which request this is responding to
	requestID := execCtx.InResponseTo
	if requestID == "" {
		// Fallback to causation_id for backward compatibility
		requestID = headers["causation_id"]
	}

	if requestID == "" {
		contextLogger.Error("No request ID in response",
			zap.Any("headers", headers))
		return fmt.Errorf("no in_response_to or causation_id in response headers")
	}

	contextLogger.Info("HandleResponse received",
		zap.String("request_id", requestID),
		zap.String("orchestration_id_in_header", execCtx.OrchestrationID))

	// Create a repository instance
	repo := NewStateRepository(s.db, s.logger)

	// Find the orchestration that is waiting for this request ID
	state, err := repo.FindByAwaitedRequestID(ctx, requestID)
	if err != nil {
		// Log more details to help debug
		contextLogger.Error("Failed to find orchestration waiting for request",
			zap.String("request_id", requestID),
			zap.Error(err))

		// Try to find by orchestration_id as fallback
		if execCtx.OrchestrationID != "" {
			state, err = repo.GetState(ctx, execCtx.OrchestrationID)
			if err == nil {
				contextLogger.Info("Found state by orchestration_id fallback",
					zap.String("orchestration_id", execCtx.OrchestrationID),
					zap.Strings("awaited_steps", state.AwaitedSteps))

				// Check if this state is actually waiting for this request
				isWaiting := false
				for _, awaited := range state.AwaitedSteps {
					if awaited == requestID {
						isWaiting = true
						break
					}
				}

				if !isWaiting {
					contextLogger.Error("State found but not waiting for this request",
						zap.String("request_id", requestID),
						zap.Strings("awaited_steps", state.AwaitedSteps))
					return fmt.Errorf("orchestration %s not waiting for request %s", execCtx.OrchestrationID, requestID)
				}
			} else {
				return fmt.Errorf("no orchestration found waiting for request_id: %s", requestID)
			}
		} else {
			return err
		}
	}

	contextLogger.Info("Found orchestration waiting for response",
		zap.String("orchestration_id", state.OrchestrationID),
		zap.String("request_id", requestID),
		zap.Int("remaining_awaited", len(state.AwaitedSteps)-1))

	// Parse the incoming response
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(response, &taskResponse); err != nil {
		contextLogger.Error("Failed to unmarshal response", zap.Error(err))
		return s.failWorkflow(ctx, state, fmt.Sprintf("failed to unmarshal response: %v", err))
	}

	// Store the response data under the request ID
	if state.CollectedData == nil {
		state.CollectedData = make(map[string]interface{})
	}
	state.CollectedData[fmt.Sprintf("response_%s", requestID)] = taskResponse.Data

	// If this is from a child orchestration, also store under a more intuitive key
	if childOrchID, ok := taskResponse.Data["orchestration_id"].(string); ok {
		state.CollectedData[fmt.Sprintf("child_%s", childOrchID)] = taskResponse.Data
	}

	// Remove this request from awaited steps
	err = repo.RemoveAwaitedRequest(ctx, state.OrchestrationID, requestID)
	if err != nil {
		contextLogger.Error("Failed to remove awaited request", zap.Error(err))
		return err
	}

	// Reload state to get updated awaited steps
	state, err = repo.GetState(ctx, state.OrchestrationID)
	if err != nil {
		return fmt.Errorf("failed to reload state: %w", err)
	}

	// Handle fuel return if present
	if fuelReturned, ok := taskResponse.Data["fuel_budget"].(float64); ok {
		currentFuel := 0
		if fuel, ok := state.CollectedData["__fuel_budget__"].(float64); ok {
			currentFuel = int(fuel)
		}
		state.CollectedData["__fuel_budget__"] = currentFuel + int(fuelReturned)

		contextLogger.Info("Fuel returned from child",
			zap.Int("returned", int(fuelReturned)),
			zap.Int("new_total", currentFuel+int(fuelReturned)))
	}

	// If no more steps are awaited, continue the workflow
	if len(state.AwaitedSteps) == 0 {
		contextLogger.Info("All responses received, continuing workflow",
			zap.String("orchestration_id", state.OrchestrationID),
			zap.String("current_step", state.CurrentStep))

		// Mark as running again
		state.Status = StatusRunning

		// Move to next step if defined
		if currentStep, exists := state.WorkflowPlan.Steps[state.CurrentStep]; exists && currentStep.NextStep != "" {
			state.CurrentStep = currentStep.NextStep
			contextLogger.Info("Moving to next step",
				zap.String("next_step", currentStep.NextStep))
		}

		// Update state
		if err := repo.UpdateState(ctx, state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		// Continue execution with restored context
		continueHeaders := make(map[string]string)
		for k, v := range headers {
			continueHeaders[k] = v
		}
		continueHeaders["orchestration_id"] = state.OrchestrationID
		continueHeaders["correlation_id"] = state.CorrelationID
		continueHeaders["client_id"] = state.ClientID
		continueHeaders["owner_agent_id"] = state.OwnerAgentID

		return s.continueExecution(ctx, state, continueHeaders)
	}

	// Still waiting for other responses
	contextLogger.Info("Still awaiting other responses",
		zap.Int("remaining_count", len(state.AwaitedSteps)),
		zap.Strings("awaited_requests", state.AwaitedSteps))

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
	repo := NewStateRepository(s.db, s.logger)

	s.logger.Info("Executing step",
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

	case "spawn_group":
		execErr = s.handleSpawnGroup(ctx, state, step, headers)

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

	// Clear executing step if we're not waiting for responses
	if state.Status != StatusAwaitingResponses {
		if err := repo.ClearExecutingStep(ctx, state.OrchestrationID); err != nil {
			s.logger.Error("Failed to clear executing step", zap.Error(err))
		}
	}

	return execErr
}

// handleSpawnGroup handles the spawn_group action
func (s *SagaCoordinator) handleSpawnGroup(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Executing spawn_group with wait-for-response")

	result, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	// The executeLocalAction for spawn_group returns:
	// - await_response: true (handled by executeLocalAction)
	// - group_id: the group ID
	// - group_name: the group name
	// - agents: map of role->agentID
	// - workflow: the orchestration workflow

	// Since SpawnGroupAction returns await_response=true,
	// executeLocalAction should have already:
	// 1. Set state.Status = StatusAwaitingResponses
	// 2. Added group_id to state.AwaitedSteps

	// We don't actually need to do anything else here because:
	// - The group doesn't send back a response as a group
	// - The spawned agents will start processing immediately
	// - The workflow execution continues separately

	// The only thing we might want to do is verify the state was updated
	if state.Status != StatusAwaitingResponses {
		s.logger.Warn("spawn_group didn't set awaiting state as expected",
			zap.Any("result", result))
	}

	return nil
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

// handleLocalAction handles local actions that can continue immediately
func (s *SagaCoordinator) handleLocalAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	s.logger.Info("Executing local action", zap.String("action", step.Action))

	_, err := s.executeLocalAction(ctx, state, step, headers)
	if err != nil {
		return err
	}

	// CHECK: Don't continue if action set state to waiting
	if state.Status == StatusAwaitingResponses {
		s.logger.Info("Action requires waiting, not continuing",
			zap.String("action", step.Action))
		return nil
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

func (s *SagaCoordinator) handleGroupTimeout(ctx context.Context, orchestrationID string, groupID string, timeout time.Duration) {
	time.Sleep(timeout)

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, orchestrationID)
	if err != nil {
		s.logger.Error("Failed to get state for group timeout check",
			zap.String("orchestration_id", orchestrationID),
			zap.Error(err))
		return
	}

	// Check if still waiting for this group
	for _, awaitedStep := range state.AwaitedSteps {
		if awaitedStep == groupID {
			s.logger.Error("Agent group timeout",
				zap.String("orchestration_id", orchestrationID),
				zap.String("group_id", groupID),
				zap.Duration("timeout", timeout))

			// Fail the workflow
			s.failWorkflow(ctx, state, fmt.Sprintf("agent group timeout after %v (group_id: %s)", timeout, groupID))
			return
		}
	}

	s.logger.Info("Group timeout check passed - group completed in time",
		zap.String("group_id", groupID))
}
