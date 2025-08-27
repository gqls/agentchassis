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
func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
	correlationID := headers["correlation_id"]
	l := s.logger.With(zap.String("correlation_id", correlationID))

	l.Info("ExecuteWorkflow called",
		zap.String("DEBUG_COOR_1: start_step", plan.StartStep),
		zap.Int("DEBUG_COOR_1: total_steps", len(plan.Steps)))

	clientID := headers["client_id"]
	if clientID == "" {
		l.Error("DEBUG_COOR_2: client_id header is required to execute a workflow")
		return fmt.Errorf("client_id header is required to execute a workflow")
	}

	// Get or create state with the plan
	state, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData)
	if err != nil {
		l.Error("Failed to get or create state", zap.Error(err))
		return err
	}

	l.Info("Workflow state retrieved",
		zap.String("DEBUG_COOR_3: status", string(state.Status)),
		zap.String("DEBUG_COOR_3: current_step", state.CurrentStep))

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
		zap.String("DEBUG_COOR_4: correlation_id", state.CorrelationID),
		zap.String("DEBUG_COOR_4: current_step", state.CurrentStep),
	)

	l.Info("Continuing workflow execution",
		zap.String("DEBUG_COOR_5: current_step", state.CurrentStep),
		zap.Int("DEBUG_COOR_5: total_steps", len(state.WorkflowPlan.Steps)))

	// Get current step from the stored plan
	currentStepConfig, exists := state.WorkflowPlan.Steps[state.CurrentStep]
	if !exists {
		errorMsg := fmt.Sprintf("step '%s' not found in plan", state.CurrentStep)
		l.Error("Step not found", zap.String("missing_step", state.CurrentStep))
		return s.failWorkflow(ctx, state, errorMsg)
	}

	l.Info("Executing step",
		zap.String("DEBUG_COOR_6: step", state.CurrentStep),
		zap.Any("DEBUG_COOR_6: currentStepConfig", currentStepConfig),
		zap.String("DEBUG_COOR_6: action", currentStepConfig.Action),
		zap.String("DEBUG_COOR_6: description", currentStepConfig.Description))

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

	l.Info("Determining action type",
		zap.String("DEBUG_COOR_7: action", currentStepConfig.Action),
		zap.Bool("DEBUG_COOR_7: is_local", isLocalAction(currentStepConfig.Action)),
		zap.String("DEBUG_COOR_7: topic", currentStepConfig.Topic))

	// Execute based on action type
	var execErr error
	switch {
	case currentStepConfig.Action == "complete_workflow":
		l.Info("Completing workflow")
		// Execute the action first (which handles parent notification)
		_, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else {
			// Then complete the workflow
			execErr = s.completeWorkflow(ctx, state)
		}

	case currentStepConfig.Action == "call_agent":
		// SPECIAL HANDLING: call_agent sends message and waits for response
		l.Info("Executing call_agent with wait-for-response")

		l.Info("currentStepConfig.Action is call_agent",
			zap.Any("DEBUG_COOR_8: headers", headers),
			zap.Bool("DEBUG_COOR_8: is_local", isLocalAction(currentStepConfig.Action)),
			zap.Any("DEBUG_COOR_8: currentStepConfig", currentStepConfig),
			zap.Any("DEBUG_COOR_8: state", state),
			zap.String("DEBUG_COOR_8: topic", currentStepConfig.Topic))

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
						zap.String("DEBUG_COOR_9: request_id", requestID),
						zap.String("DEBUG_COOR_9: agent_called", fmt.Sprintf("%v", resultMap["agent_called"])),
						zap.String("DEBUG_COOR_9: topic", fmt.Sprintf("%v", resultMap["topic"])),
						zap.String("DEBUG_COOR_9: next_step", currentStepConfig.NextStep))

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

		// In continueExecution method, add this case:
	case currentStepConfig.Action == "start_orchestration":
		l.Info("Executing start_orchestration with wait-for-response")

		// Execute the action to create the child orchestration
		result, err := s.executeLocalAction(ctx, state, currentStepConfig, headers)
		if err != nil {
			execErr = err
		} else if resultMap, ok := result.(map[string]interface{}); ok {
			// Check if we need to wait for the child orchestration
			if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
				if requestID, ok := resultMap["new_correlation_id"].(string); ok {
					// Store that we've already executed this step
					state.CollectedData[state.CurrentStep] = result

					// Update state to wait for child orchestration
					state.Status = StatusAwaitingResponses
					state.AwaitedSteps = []string{requestID}

					// Mark this step as completed in metadata
					state.ExecutionMetadata.Checkpoints[state.CurrentStep] = time.Now().UTC()

					repo := NewStateRepository(s.db, s.logger)
					if err := repo.UpdateState(ctx, state); err != nil {
						execErr = fmt.Errorf("failed to update state for waiting: %w", err)
					} else {
						l.Info("Waiting for child orchestration",
							zap.String("DEBUG_COOR_10: child_correlation_id", requestID),
							zap.String("DEBUG_COOR_10: next_step", currentStepConfig.NextStep))

						// Set up timeout for child orchestration
						timeout := 5 * time.Minute
						if currentStepConfig.Timeout > 0 {
							timeout = currentStepConfig.Timeout
						}
						go s.handleChildOrchestrationTimeout(ctx, state.CorrelationID, requestID, timeout)

						return nil // Exit and wait for child to complete
					}
				}
			}
		}
		// If we get here, there was an error or await_response wasn't true
		if execErr != nil {
			execRecord.Result = "failed"
			execRecord.Error = execErr.Error()
		}

	case currentStepConfig.Action == "fan_out":
		l.Info("Handling fan-out")
		execErr = s.handleFanOut(ctx, headers, currentStepConfig, state)
	case currentStepConfig.Action == "pause_for_human_input":
		l.Info("Pausing for human input")
		execErr = s.handlePauseForHumanInput(ctx, headers, currentStepConfig, state)
	case isLocalAction(currentStepConfig.Action):
		l.Info("Executing local action.", zap.String("action", currentStepConfig.Action))
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
			zap.String("DEBUG_COOR_10: step", state.CurrentStep),
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

	// Log with virtual topic for consistency
	virtualTopic := fmt.Sprintf("local.action.%s", step.Action)
	l.Info("Executing local action",
		zap.String("DEBUG_COOR_12: virtual_topic", virtualTopic))

	// Verify producer is available
	if s.producer == nil {
		l.Error("Producer is nil in SagaCoordinator")
		return nil, fmt.Errorf("producer not available for action execution")
	}

	// For call_agent actions, ensure proper correlation context
	if step.Action == "call_agent" {
		// The agent being called should use THIS orchestration's correlation_id
		// NOT the grandparent's
		actionHeaders["correlation_id"] = state.CorrelationID

		// Remove parent_correlation_id from headers going to sub-agents
		// They don't need to know about the grandparent
		delete(actionHeaders, "parent_correlation_id")

		l.Info("Adjusted headers for call_agent",
			zap.String("DEBUG_COOR_12: correlation_id", actionHeaders["correlation_id"]),
			zap.String("DEBUG_COOR_12: original_parent_correlation_id", headers["parent_correlation_id"]))
	}

	// For start_orchestration, ensure it knows its parent
	if step.Action == "start_orchestration" {
		// The child orchestration needs to know THIS orchestration is its parent
		actionHeaders["parent_correlation_id"] = state.CorrelationID
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

	params.Logger.Info("in execute local action",
		zap.Any("headers", actionHeaders),
		zap.Any("agent config", state.CollectedData["agent_config"]))

	// If agent_type is not in headers, try to get it from other sources
	if params.AgentType == "" {
		// Try to get from collected data or state
		if agentConfig, ok := state.CollectedData["agent_config"].(map[string]interface{}); ok {
			if at, ok := agentConfig["agent_type"].(string); ok {
				params.AgentType = at
				actionHeaders["agent_type"] = at
			}
		}
	}

	// Log what we're passing
	l.Info("Executing local action with params",
		zap.String("DEBUG_COOR_14: agent_type", params.AgentType),
		zap.String("DEBUG_COOR_14: aaction", step.Action))

	// Execute the action
	result, err := handler(ctx, params)
	if err != nil {
		l.Error("Local action failed", zap.Error(err))
		return nil, fmt.Errorf("local action failed: %w", err)
	}

	l.Info("Local action completed successfully", zap.Any("result", result))

	// CHECK FOR AWAIT_RESPONSE FLAG
	if resultMap, ok := result.(map[string]interface{}); ok {
		if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {

			var requestID string

			// Check for different ID fields based on action type
			if step.Action == "start_orchestration" {
				// For start_orchestration, use new_correlation_id
				if id, ok := resultMap["new_correlation_id"].(string); ok {
					requestID = id
				}
			} else if id, ok := resultMap["request_id"].(string); ok {
				// For other actions like call_agent
				requestID = id
			}

			if requestID != "" {
				l.Info("Local action requires waiting for response",
					zap.String("DEBUG_COOR_15: arequest_id", requestID),
					zap.String("DEBUG_COOR_15: action", step.Action),
					zap.Bool("DEBUG_COOR_15: await_response", true))

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

				l.Info("Action requires waiting for response",
					zap.String("DEBUG_COOR_16: request_id", requestID),
					zap.String("DEBUG_COOR_16: action", step.Action))

				// Set up timeout for child orchestration if needed
				if step.Action == "start_orchestration" {
					timeout := 5 * time.Minute
					if step.Timeout > 0 {
						timeout = step.Timeout
					}
					go s.handleChildOrchestrationTimeout(ctx, state.CorrelationID, requestID, timeout)
				}

				// Don't continue execution - wait for a response
				return result, nil
			}
		}
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
			zap.String("DEBUG_COOR_17: next_step", step.NextStep))

		// Continue execution with the ADJUSTED headers
		return result, s.continueExecution(ctx, state, actionHeaders)
	} else {
		l.Info("DEBUG_COOR_18: No next step specified, workflow may be complete")
	}

	return result, nil
}

// executeRemoteAction sends work to another agent
func (s *SagaCoordinator) executeRemoteAction(ctx context.Context, state *OrchestrationState, step models.Step, headers map[string]string) error {
	l := s.logger.With(
		zap.String("DEBUG_COOR_19: correlation_id", state.CorrelationID),
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
	outHeaders["request_id"] = newRequestID

	l.Info("Sending remote action",
		zap.String("DEBUG_COOR_20: request_id", newRequestID),
		zap.String("DEBUG_COOR_20: topic", step.Topic))

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
		zap.String("DEBUG_COOR_21: request_id", newRequestID))

	return nil
}

// Rest of the methods remain the same but with enhanced logging...
// (I'll include key ones with logging enhancements)

// completeWorkflow marks the workflow as completed with enhanced tracking
func (s *SagaCoordinator) completeWorkflow(ctx context.Context, state *OrchestrationState) error {
	l := s.logger.With(zap.String("correlation_id", state.CorrelationID))

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

	// Check if this is a child orchestration and notify parent
	if parentCorrelationID := state.CollectedData["parent_correlation_id"]; parentCorrelationID != nil {
		if parentID, ok := parentCorrelationID.(string); ok && parentID != "" {
			l.Info("This is a child orchestration, checking for parent notification",
				zap.String("DEBUG_COOR_23: parent_correlation_id", parentID))

			// Determine parent's response topic
			parentResponseTopic := ""

			// First check for parent_agent_type in collected data
			if parentAgentType, ok := state.CollectedData["parent_agent_type"].(string); ok && parentAgentType != "" {
				parentResponseTopic = fmt.Sprintf("system.agent.%s.responses", parentAgentType)
				l.Info("Using parent agent's response topic",
					zap.String("DEBUG_COOR_24: parent_agent_type", parentAgentType),
					zap.String("DEBUG_COOR_24: parent_response_topic", parentResponseTopic))
			} else if replyTopic, ok := state.CollectedData["reply_to_topic"].(string); ok && replyTopic != "" {
				// Check if we have a reply_to_topic stored
				parentResponseTopic = replyTopic
				l.Info("Using stored reply_to_topic",
					zap.String("DEBUG_COOR_25: parent_response_topic2", parentResponseTopic))
			} else {
				// Fallback for backward compatibility
				parentResponseTopic = "system.agent.generic.responses"
				l.Warn("DEBUG_COOR_26: No parent agent type or reply topic found, using legacy topic")
			}

			// Prepare the completion notification
			parentResponse := models.TaskResponse{
				Success: true,
				Data: map[string]interface{}{
					"status":         "completed",
					"correlation_id": state.CorrelationID,
					"final_result":   state.CollectedData,
					"execution_stats": map[string]interface{}{
						"completed_steps": state.ExecutionMetadata.CompletedSteps,
						"total_steps":     state.ExecutionMetadata.TotalSteps,
						"duration_ms":     now.Sub(state.ExecutionMetadata.StartTime).Milliseconds(),
					},
				},
			}

			responseBytes, _ := json.Marshal(parentResponse)

			// Create headers for the response
			responseHeaders := map[string]string{
				"correlation_id": parentID,            // Parent's correlation
				"causation_id":   state.CorrelationID, // Child's correlation as causation
				"message_type":   "orchestration_complete",
				"from_agent_id":  os.Getenv("AGENT_ID"), // This agent
			}

			// Include parent correlation for tracing
			if parentID != "" {
				responseHeaders["parent_correlation_id"] = parentID
			}

			l.Info("Notifying parent orchestration of completion",
				zap.String("DEBUG_COOR_27: parent_correlation_id", parentID),
				zap.String("DEBUG_COOR_27: parent_response_topic3", parentResponseTopic),
				zap.String("DEBUG_COOR_27: from_agent", responseHeaders["from_agent_id"]))

			// Send to parent's response topic
			err := s.producer.Produce(ctx,
				parentResponseTopic,
				responseHeaders,
				[]byte(parentID),
				responseBytes)

			if err != nil {
				l.Error("Failed to notify parent orchestration",
					zap.Error(err),
					zap.String("DEBUG_COOR_28: topic", parentResponseTopic))
				// Don't fail the child completion, just log the error
			} else {
				l.Info("Parent orchestration notified of completion",
					zap.String("DEBUG_COOR_29: topic", parentResponseTopic))
			}
		}
	}

	l.Info("Workflow completed and all notifications sent")
	return nil
}

// getOrCreateState retrieves existing state or creates new one - now includes the plan
func (s *SagaCoordinator) getOrCreateState(ctx context.Context, correlationID string, clientID string, plan models.WorkflowPlan, initialData []byte) (*OrchestrationState, error) {
	l := s.logger.With(zap.String("correlation_id", correlationID))

	repo := NewStateRepository(s.db, s.logger)

	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		l.Info("DEBUG_COOR_30: State doesn't exist, creating new one")
		// State doesn't exist, create it with the plan
		if err := repo.CreateInitialState(ctx, correlationID, clientID, plan, initialData); err != nil {
			l.Error("DEBUG_COOR_30: Failed to create initial state", zap.Error(err))
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

	l.Error("DEBUG_COOR_31: Failing workflow", zap.String("error", errorMsg))

	state.Status = StatusFailed
	state.Error = errorMsg

	repo := NewStateRepository(s.db, s.logger)
	if err := repo.UpdateState(ctx, state); err != nil {
		l.Error("DEBUG_COOR_31: Failed to update state to failed", zap.Error(err))
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
			zap.String("DEBUG_COOR_32: step", record.Step))
	}
}

// HandleResponse processes responses and continues workflow
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {

	correlationID := headers["correlation_id"]
	causationID := headers["causation_id"]

	// Check if this is a response FROM a child orchestration
	if parentCorrelationID := headers["parent_correlation_id"]; parentCorrelationID != "" {
		// Check if the response is from a completed child orchestration
		var taskResponse models.TaskResponse
		if err := json.Unmarshal(response, &taskResponse); err == nil {
			// Check if this is a final result from child orchestration
			if status, ok := taskResponse.Data["status"].(string); ok && status == "completed" {
				s.logger.Info("Received completion from child orchestration",
					zap.String("DEBUG_COOR_33: child_correlation_id", correlationID),
					zap.String("DEBUG_COOR_33: parent_correlation_id", parentCorrelationID))

				// Update the PARENT orchestration
				repo := NewStateRepository(s.db, s.logger)
				parentState, err := repo.GetState(ctx, parentCorrelationID)
				if err != nil {
					s.logger.Error("Failed to get parent state",
						zap.String("DEBUG_COOR_34: parent_correlation_id", parentCorrelationID),
						zap.Error(err))
					// Continue with normal processing
				} else {
					// Store child result in parent's collected data
					if parentState.CollectedData == nil {
						parentState.CollectedData = make(map[string]interface{})
					}

					// Store under the step that started this child orchestration
					parentState.CollectedData["child_orchestration_result"] = taskResponse.Data

					// Continue parent workflow if needed
					if len(parentState.AwaitedSteps) == 0 {
						parentState.Status = StatusRunning
						if currentStep := parentState.WorkflowPlan.Steps[parentState.CurrentStep]; currentStep.NextStep != "" {
							parentState.CurrentStep = currentStep.NextStep
						}

						if err := repo.UpdateState(ctx, parentState); err == nil {
							// Continue parent execution
							return s.continueExecution(ctx, parentState, headers)
						}
					}
				}
			}
		}
	}

	l := s.logger.With(
		zap.String("DEBUG_COOR_35: correlation_id", correlationID),
		zap.String("DEBUG_COOR_35: causation_id", causationID),
	)

	l.Info("Handling workflow response")

	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		l.Error("Failed to get state for response", zap.Error(err))
		return fmt.Errorf("DEBUG_COOR_36: failed to get state: %w", err)
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
				zap.String("DEBUG_COOR_37: next_step", currentStep.NextStep))
		}

		if err := repo.UpdateState(ctx, state); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		// Ensure required headers for continuation
		if headers["fuel_budget"] == "" {
			if state != nil && state.CollectedData != nil {
				if fuel, ok := state.CollectedData["initial_fuel_budget"].(string); ok {
					headers["fuel_budget"] = fuel
				}
			}
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
		zap.Int("DEBUG_COOR_38: remaining_awaited", len(state.AwaitedSteps)))

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
			"headers":               headers,
			"timestamp":             time.Now().UTC(),
			"message":               "Starting website build orchestration",
			"parent_correlation_id": headers["parent_correlation_id"],
		},
	}
	initialDataBytes, _ := json.Marshal(initialData)

	// Create the initial state
	repo := NewStateRepository(s.db, s.logger)
	if err := repo.CreateInitialState(ctx, correlationID, clientID, plan, initialDataBytes); err != nil {
		return fmt.Errorf("failed to create orchestration state: %w", err)
	}

	l.Info("New orchestration created",
		zap.String("DEBUG_COOR_39: client_id", clientID),
		zap.String("DEBUG_COOR_39: start_step", plan.StartStep),
		zap.Int("DEBUG_COOR_39: total_steps", len(plan.Steps)))

	// Start execution immediately
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		return fmt.Errorf("failed to get created state: %w", err)
	}

	// Store parent correlation ID in collected data for child to access later
	if parentID := headers["parent_correlation_id"]; parentID != "" {
		state.CollectedData["parent_correlation_id"] = parentID
		if parentType := headers["parent_agent_type"]; parentType != "" {
			state.CollectedData["parent_agent_type"] = parentType
		}
		// Save this update
		if err := repo.UpdateState(ctx, state); err != nil {
			l.Error("DEBUG_COOR_40: Failed to store parent correlation ID", zap.Error(err))
		}
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
		headers["agent_instance_id"] = "orchestrator-" + correlationID
	}

	// Start the workflow execution
	return s.continueExecution(ctx, state, headers)
}

func (s *SagaCoordinator) handleTimeout(ctx context.Context, correlationID string, requestID string, timeout time.Duration) {
	time.Sleep(timeout)

	// Check if still waiting for this request
	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, correlationID)
	if err != nil {
		s.logger.Error("Failed to get state for timeout check",
			zap.String("DEBUG_COOR_41: correlation_id", correlationID),
			zap.Error(err))
		return
	}

	// Check if still waiting for this specific request
	for _, awaitedStep := range state.AwaitedSteps {
		if awaitedStep == requestID {
			s.logger.Error("Timeout waiting for agent response",
				zap.String("DEBUG_COOR_42: correlation_id", correlationID),
				zap.String("DEBUG_COOR_42: request_id", requestID),
				zap.Duration("DEBUG_COOR_42: timeout", timeout))

			// Fail the workflow
			s.failWorkflow(ctx, state, fmt.Sprintf("timeout after %v waiting for agent response (request_id: %s)", timeout, requestID))
			return
		}
	}

	// If we get here, the response was already received
	s.logger.Info("Timeout check passed - response already received",
		zap.String("DEBUG_COOR_43: correlation_id", correlationID),
		zap.String("DEBUG_COOR_43: request_id", requestID))
}

func (s *SagaCoordinator) handleChildOrchestrationTimeout(ctx context.Context, parentCorrelationID string, childCorrelationID string, timeout time.Duration) {
	time.Sleep(timeout)

	// Check if parent is still waiting
	repo := NewStateRepository(s.db, s.logger)
	state, err := repo.GetState(ctx, parentCorrelationID)
	if err != nil {
		s.logger.Error("Failed to get parent state for timeout check",
			zap.String("DEBUG_COOR_44: parent_correlation_id", parentCorrelationID),
			zap.Error(err))
		return
	}

	// Check if still waiting for this child
	for _, awaitedStep := range state.AwaitedSteps {
		if awaitedStep == childCorrelationID {
			s.logger.Error("Child orchestration timeout",
				zap.String("DEBUG_COOR_45: parent_correlation_id", parentCorrelationID),
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

			// Simulate response from child
			headers := map[string]string{
				"correlation_id": parentCorrelationID,
				"causation_id":   childCorrelationID,
			}

			// Process the timeout as a response
			s.HandleResponse(ctx, headers, responseBytes)
			return
		}
	}

	s.logger.Info("Timeout check passed - child completed in time",
		zap.String("DEBUG_COOR_46: child_correlation_id", childCorrelationID))
}
