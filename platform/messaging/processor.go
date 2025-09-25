// FILE: platform/messaging/processor.go
package messaging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/validation"
	"go.uber.org/zap"
)

// MessageProcessor handles all message processing
type MessageProcessor struct {
	agentType    string
	agentID      string
	db           *sql.DB
	sqlDB        *sql.DB
	producer     kafka.Producer
	orchestrator *orchestration.SagaCoordinator
	validator    *validation.Validator
	configLoader *config.AgentConfigLoader
	logger       *zap.Logger
	tracer       *types.TraceLogger
	initializer  Initializer

	// For stateless operation
	isStateless bool
	podName     string
	stateRepo   *orchestration.StateRepository
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(
	agentType string,
	agentID string,
	db *sql.DB,
	producer kafka.Producer,
	orchestrator *orchestration.SagaCoordinator,
	validator *validation.Validator,
	logger *zap.Logger,
	initializer Initializer,
) *MessageProcessor {
	// Also get SQL DB connection if available
	var sqlDB *sql.DB
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		var err error
		sqlDB, err = sql.Open("pgx", connStr)
		if err != nil {
			logger.Error("Failed to create SQL DB connection", zap.Error(err))
		}
	} else if host := os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST"); host != "" {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host,
			os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT"),
			os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER"),
			os.Getenv("CLIENTS_DB_PASSWORD"),
			os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME"))
		var err error
		sqlDB, err = sql.Open("pgx", connStr)
		if err != nil {
			logger.Error("Failed to create SQL DB connection from env vars", zap.Error(err))
		}
	}

	// Create tracer if enabled
	var tracer *types.TraceLogger
	if os.Getenv("ENABLE_MESSAGE_TRACING") == "true" {
		tracer = types.NewTraceLogger(logger)
	}

	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		podName = fmt.Sprintf("%s-local-%d", agentType, os.Getpid())
	}

	return &MessageProcessor{
		agentType:    agentType,
		agentID:      agentID,
		db:           db,
		sqlDB:        sqlDB,
		producer:     producer,
		orchestrator: orchestrator,
		validator:    validator,
		configLoader: config.NewAgentConfigLoader(logger),
		logger:       logger,
		tracer:       tracer,
		initializer:  initializer,
		isStateless:  os.Getenv("ENABLE_STATELESS_MODE") == "true",
		podName:      podName,
		stateRepo:    orchestration.NewStateRepository(sqlDB, logger),
	}
}

// process determines how to handle the message based on agent configuration
func (p *MessageProcessor) process(ctx context.Context, msgCtx *MessageContext) error {
	current, caller := getFuncInfo(1)
	fmt.Fprint(os.Stderr, "DEBUG uuid: process START printf")
	p.logger.With(msgCtx.ExecutionContext.LogContext()...).Info("In file processor.go process 110",
		zap.String("function", current),
		zap.String("called_by (in process)", caller),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Initialize CollectedData if nil
	if msgCtx.CollectedData == nil {
		msgCtx.CollectedData = make(map[string]interface{})
	}

	msgCtx.Logger.Info("Starting message processing",
		zap.String("action", msgCtx.ExecutionContext.Action),
		zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID),
		zap.String("orchestration_name", msgCtx.ExecutionContext.OrchestrationName),
		zap.String("RESPONSES topic in process", msgCtx.ExecutionContext.ResponsesTopic),
		zap.String("request id in process", msgCtx.ExecutionContext.RequestID),
	)

	// Ensure ExecutionContext has required fields
	if msgCtx.ExecutionContext.OrchestrationID == "" {
		msgCtx.ExecutionContext.OrchestrationID = uuid.New().String()
		msgCtx.Logger.Info("Generated new orchestration_id",
			zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID))
	}

	// Sync headers from context for backward compatibility
	msgCtx.SyncHeadersFromContext()

	// Store parent context if this is a child
	if msgCtx.IsChildOrchestration() {
		msgCtx.CollectedData["__execution_context__"] = msgCtx.ExecutionContext
	}

	// Load agent definition
	agentDef, err := p.loadAgentDefinition(ctx, p.agentType)
	if err != nil {
		msgCtx.Logger.Warn("Failed to load agent definition, creating dynamic workflow",
			zap.String("agent_type", p.agentType),
			zap.Error(err))
		// Instead of processWithDefaults, create a workflow for the action
		workflow := p.createWorkflowForAction(msgCtx.ExecutionContext.Action)
		agentDef = &actions.AgentDefinition{
			Type:          p.agentType,
			DefaultConfig: make(map[string]interface{}),
		}
		agentDef.DefaultConfig["workflow"] = workflow
	}

	msgCtx.Logger.Info("Agent definition loaded",
		zap.String("display_name", agentDef.DisplayName),
		zap.String("category", agentDef.Category))

	var workflow models.WorkflowPlan
	// Select the appropriate workflow based on context
	workflow, err = p.selectWorkflow(ctx, agentDef, msgCtx)
	if err != nil {
		msgCtx.Logger.Error("Failed to select workflow", zap.Error(err))
		return err
	}

	// Store the agent configuration for actions to use
	msgCtx.CollectedData["agent_config"] = agentDef.DefaultConfig

	msgCtx.Logger.Info("Message value in msgCtx",
		zap.ByteString("msgCtx.Message.Value", msgCtx.Message.Value),
		zap.Any("agent config", agentDef.DefaultConfig))

	// Store the input data
	var inputPayload map[string]interface{}
	if err := json.Unmarshal(msgCtx.Message.Value, &inputPayload); err == nil {
		// Store the entire payload as input_data, not just a "data" field
		msgCtx.CollectedData["input_data"] = inputPayload
		msgCtx.CollectedData["input_action"] = inputPayload["action"]

		msgCtx.Logger.Info("Input data stored",
			zap.Any("input_data", inputPayload))
	}

	// Create agent config with the workflow
	agentConfig := &models.AgentConfig{
		CoreLogic: agentDef.DefaultConfig,
		Workflow:  workflow,
	}

	// Validate workflow
	if err := p.validator.ValidateWorkflow(agentConfig.Workflow); err != nil {
		msgCtx.Logger.Error("Invalid workflow configuration",
			zap.Error(err),
			zap.String("agent_type", p.agentType))
		return errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").
			WithCause(err).
			Build()
	}

	msgCtx.Logger.Info("Workflow validated successfully, executing workflow")

	// Add debug logging
	p.logger.Info("DEBUG: Executing workflow",
		zap.Any("agentConfig", agentConfig))

	fmt.Fprint(os.Stderr, "DEBUG uuid: process about to ExecuteWorkflow printf")

	// Execute the workflow
	err = p.executeWorkflow(ctx, msgCtx, agentConfig)

	// After handing off to the orchestrator, the message processor's job is done.
	// The orchestrator is now responsible for the workflow's lifecycle.
	if err != nil {
		// Only handle errors related to *starting* the workflow
		if err == orchestration.ErrWaitingForResponse {
			msgCtx.Logger.Info("Workflow started and is now waiting for a response.")
			return nil
		}
		msgCtx.Logger.Error("Failed to start workflow execution", zap.Error(err))
		fmt.Fprint(os.Stderr, "DEBUG uuid: process Failed to start workflow execution printf")
		return p.sendWorkflowFailureResponse(ctx, msgCtx, err)
	}

	// Check if this is a child workflow that completed
	if msgCtx.IsChildOrchestration() {
		// Check if workflow is complete
		if p.sqlDB != nil {
			repo := orchestration.NewStateRepository(p.sqlDB, msgCtx.Logger)
			state, _ := repo.GetState(ctx, msgCtx.ExecutionContext.OrchestrationID)
			if state != nil && state.Status == orchestration.StatusCompleted {
				msgCtx.Logger.Info("Child workflow completed, sending response to parent")
				return p.sendWorkflowSuccessResponse(ctx, msgCtx)
			}
		}
	}

	msgCtx.Logger.Info("Workflow successfully handed off to the orchestrator.")
	return nil // Return nil to acknowledge the message without sending a premature response.
}

func (p *MessageProcessor) createWorkflowForAction(action string) map[string]interface{} {
	return map[string]interface{}{
		"start_step": action,
		"steps": map[string]interface{}{
			action: map[string]interface{}{
				"action":      action,
				"description": fmt.Sprintf("Execute %s", action),
				"next_step":   "complete",
			},
			"complete": map[string]interface{}{
				"action":      "complete_workflow",
				"description": "Complete workflow",
			},
		},
	}
}

func (p *MessageProcessor) validateNoSelfRecursion(workflow models.WorkflowPlan, agentType string) error {
	for stepName, step := range workflow.Steps {
		if step.Action == "call_agent" {
			if targetType, ok := step.Config["agent_type"].(string); ok && targetType == agentType {
				return fmt.Errorf("workflow step '%s' would cause self-recursion: agent type '%s' cannot call itself", stepName, agentType)
			}
		}
	}
	return nil
}

// loadAgentDefinition loads the agent definition from the database
func (p *MessageProcessor) loadAgentDefinition(ctx context.Context, agentType string) (*actions.AgentDefinition, error) {
	p.logger.Debug("Loading agent definition", zap.String("agent_type", agentType))

	query := `
        SELECT type, display_name, description, category, default_config, capabilities
        FROM agent_definitions
        WHERE type = $1
    `

	var def actions.AgentDefinition
	var configJSON json.RawMessage // Read as RawMessage first
	var capabilitiesJSON json.RawMessage

	err := p.db.QueryRowContext(ctx, query, agentType).Scan(
		&def.Type,
		&def.DisplayName,
		&def.Description,
		&def.Category,
		&configJSON,
		&capabilitiesJSON,
	)

	if err != nil {
		p.logger.Error("Failed to query agent definition",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to load agent definition: %w", err)
	}

	p.logger.Info("Agent definition loaded from DB",
		zap.String("type", def.Type),
		zap.String("display_name", def.DisplayName),
		zap.String("category", def.Category),
		zap.String("raw_config", string(configJSON)))

	// Parse the JSON config into map
	if err := json.Unmarshal(configJSON, &def.DefaultConfig); err != nil {
		p.logger.Error("Failed to parse agent config JSON", zap.Error(err))
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	// Parse capabilities if present
	if capabilitiesJSON != nil {
		json.Unmarshal(capabilitiesJSON, &def.Capabilities)
	}

	p.logger.Info("Agent definition loaded successfully",
		zap.String("type", def.Type),
		zap.String("display_name", def.DisplayName),
		zap.String("category", def.Category))

	return &def, nil
}

// convertToWorkflowPlan converts the workflow config from DB to WorkflowPlan
func (p *MessageProcessor) convertToWorkflowPlan(workflowConfig map[string]interface{}) models.WorkflowPlan {
	plan := models.WorkflowPlan{
		Steps: make(map[string]models.Step),
	}

	// Get start step
	if startStep, ok := workflowConfig["start_step"].(string); ok {
		plan.StartStep = startStep
	}

	// Convert steps
	if steps, ok := workflowConfig["steps"].(map[string]interface{}); ok {
		for stepName, stepData := range steps {
			if stepMap, ok := stepData.(map[string]interface{}); ok {
				step := models.Step{
					Action:      p.getStringValue(stepMap, "action"),
					Description: p.getStringValue(stepMap, "description"),
					NextStep:    p.getStringValue(stepMap, "next_step"),
					Topic:       p.getStringValue(stepMap, "topic"),
				}

				// Get config if present
				if config, ok := stepMap["config"].(map[string]interface{}); ok {
					step.Config = config
				}

				// Get dependencies if present
				if deps, ok := stepMap["dependencies"].([]interface{}); ok {
					for _, dep := range deps {
						if depStr, ok := dep.(string); ok {
							step.Dependencies = append(step.Dependencies, depStr)
						}
					}
				}

				plan.Steps[stepName] = step
				p.logger.Debug("Converted workflow step",
					zap.String("step_name", stepName),
					zap.String("action", step.Action),
					zap.String("next_step", step.NextStep))
			}
		}
	}

	return plan
}

// getStringValue safely extracts a string value from a map
func (p *MessageProcessor) getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getDefaultWorkflow returns a simple default workflow
func (p *MessageProcessor) getDefaultWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "process",
		Steps: map[string]models.Step{
			"process": {
				Action:      "execute_llm_prompt",
				Description: "Process the task",
				NextStep:    "respond",
			},
			"respond": {
				Action:      "send_notification",
				Description: "Send response",
				NextStep:    "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Complete the workflow",
			},
		},
	}
}

// New response methods using ExecutionContext
func (p *MessageProcessor) sendWorkflowSuccessResponse(ctx context.Context, msgCtx *MessageContext) error {

	current, caller := getFuncInfo(1)
	caller, caller_called_by := getFuncInfo(2)

	msgCtx.Logger.With(msgCtx.ExecutionContext.LogContext()...).Info("In file processor.go sendWorkflowSuccessResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("caller_called_by", caller_called_by),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp: ", time.Now().UTC().Format(time.RFC3339)),
	)

	// Get final state if available
	var finalResult interface{}
	if p.sqlDB != nil {
		repo := orchestration.NewStateRepository(p.sqlDB, msgCtx.Logger)
		state, err := repo.GetState(ctx, msgCtx.ExecutionContext.OrchestrationID)
		if err == nil && state != nil {
			finalResult = state.CollectedData
		}
	}

	if finalResult == nil {
		finalResult = map[string]interface{}{"status": "completed"}
	}

	return p.sendWorkflowResponse(ctx, msgCtx, finalResult)
}

func (p *MessageProcessor) sendWorkflowFailureResponse(ctx context.Context, msgCtx *MessageContext, err error) error {
	current, caller := getFuncInfo(1)
	caller, caller_called_by := getFuncInfo(2)

	msgCtx.Logger.With(msgCtx.ExecutionContext.LogContext()...).Info("In file processor.go ",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("caller_called_by", caller_called_by),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp: ", time.Now().UTC().Format(time.RFC3339)),
	)

	return p.sendWorkflowResponse(ctx, msgCtx, map[string]interface{}{
		"error":  err.Error(),
		"status": "failed",
	})
}

func (p *MessageProcessor) handleError(ctx context.Context, msgCtx *MessageContext, err error, errorType string) error {
	msgCtx.Logger.Error("Processing failed", zap.Error(err))
	observability.AgentTasksProcessed.WithLabelValues(p.agentType, msgCtx.ExecutionContext.Action, errorType).Inc()

	// Check for specific error types
	if domainErr, ok := err.(*errors.DomainError); ok {
		if domainErr.Code == errors.ErrInsufficientFuel {
			observability.FuelExhausted.WithLabelValues(p.agentType, msgCtx.ExecutionContext.Action, msgCtx.Headers["client_id"]).Inc()
		}
		p.sendErrorResponse(ctx, msgCtx, domainErr)
	} else {
		p.sendErrorResponse(ctx, msgCtx, errors.InternalError("Processing failed", err))
	}

	return err
}

// ProcessResponse handles response messages for orchestrated workflows
func (p *MessageProcessor) ProcessResponse(ctx context.Context, msg kafka.Message) error {
	current, caller := getFuncInfo(1)

	p.logger.Info("In file processor.go ",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp: ", time.Now().UTC().Format(time.RFC3339)),
	)

	headers := kafka.HeadersToMap(msg.Headers)

	p.logger.Info("Processing orchestration response",
		zap.String("DEBUG_PROCESSOR_3: correlation_id", headers["correlation_id"]),
		zap.String("DEBUG_PROCESSOR_3: orchestration_id", headers["orchestration_id"]),
		zap.String("DEBUG_PROCESSOR_3: causation_id", headers["causation_id"]))

	// Route to orchestrator
	return p.orchestrator.HandleResponse(ctx, headers, msg.Value)
}

func (p *MessageProcessor) sendWorkflowResponse(ctx context.Context, msgCtx *MessageContext, result interface{}) error {
	current, caller := getFuncInfo(1)
	_, caller_called_by := getFuncInfo(2)

	p.logger.With(msgCtx.ExecutionContext.LogContext()...).Info("RESPONSE_CREATION: Starting to create response processor.go sendWorkflowResponse",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("caller_called_by in sendWorkflowResponse", caller_called_by),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),

		zap.Any("result_data", result),
	)

	// Create response context
	responseCtx := msgCtx.CreateResponseContext()
	contextLogger := p.logger.With(msgCtx.ExecutionContext.LogContext()...)

	contextLogger.Info("CRITICAL_FLOW: sendWorkflowResponse called",
		zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID),
		zap.String("orchestration_name", msgCtx.ExecutionContext.OrchestrationName),
		zap.String("parent_orchestration_id", msgCtx.ExecutionContext.ParentOrchestrationID),
		zap.String("original_request_id", msgCtx.ExecutionContext.RequestID),
		zap.Any("in_response_to", msgCtx.ExecutionContext.InResponseTo),
		zap.String("responses_topic", responseCtx.ResponsesTopic),
		zap.Any("result_type", fmt.Sprintf("%T", result)))

	// The 'result' is the full 'CollectedData' map from the orchestration state.
	// We need to find the result of the last action within this map to see
	// if we should suppress the response or use a specific request_id.
	if allData, ok := result.(map[string]interface{}); ok {
		var lastActionResult map[string]interface{}

		// Find the map that contains the 'await_response' key, as this identifies the action result.
		for _, value := range allData {
			if resultMap, isMap := value.(map[string]interface{}); isMap {
				if _, found := resultMap["await_response"]; found {
					lastActionResult = resultMap
					break
				}
			}
		}

		if lastActionResult != nil {
			// **FIX**: If the action is waiting for a response, DO NOT send a response now.
			// This prevents the premature response and the subsequent error.
			if await, ok := lastActionResult["await_response"].(bool); ok && await {
				contextLogger.Info("CRITICAL_FLOW: Suppressing response because workflow is waiting",
					zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID))
				return nil
			}

			// If the action was NOT waiting but generated a request_id, use that ID for the response.
			if requestID, ok := lastActionResult["request_id"].(string); ok && requestID != "" {
				contextLogger.Info("Using action's specific request_id for response",
					zap.String("action_request_id", requestID),
					zap.String("original_request_id", msgCtx.ExecutionContext.RequestID))

				// FIX: Update InResponseTo properly - it's a *ResponseContext
				if responseCtx.InResponseTo != nil {
					responseCtx.InResponseTo.RequestID = requestID
				} else {
					responseCtx.InResponseTo = &types.ResponseContext{
						RequestID:               requestID,
						StepID:                  msgCtx.ExecutionContext.StepID,
						StepName:                msgCtx.ExecutionContext.StepName,
						MessageID:               msgCtx.ExecutionContext.MessageID,
						Action:                  msgCtx.ExecutionContext.Action,
						ParentOrchestrationID:   msgCtx.ExecutionContext.ParentOrchestrationID,
						ParentOrchestrationName: msgCtx.ExecutionContext.ParentOrchestrationName,
					}
				}
				// Update the request ID in the response context
				responseCtx.RequestID = requestID
			}
		}
	}

	// Trace outgoing response
	if p.tracer != nil {
		p.tracer.TraceMessage(responseCtx, "sending_response in sendWorkflowResponse", responseCtx.ResponsesTopic, msgCtx.Message.Value)
	}

	// Determine target orchestration
	targetOrchestrationID := msgCtx.ExecutionContext.OrchestrationID
	if msgCtx.ExecutionContext.ParentOrchestrationID != "" {
		targetOrchestrationID = msgCtx.ExecutionContext.ParentOrchestrationID
		contextLogger.Info("TRACE: Child sending response to parent",
			zap.String("child_orch", msgCtx.ExecutionContext.OrchestrationID),
			zap.String("parent_orch", targetOrchestrationID))
	}

	// Build response headers
	responseHeaders := responseCtx.ToHeaders()

	contextLogger.Info("TRACE: Sending workflow response",
		zap.String("response_orch_id", targetOrchestrationID),
		zap.String("RESPONSES_topic", responseCtx.ResponsesTopic),
		zap.Int("fuel_returning", msgCtx.ExecutionContext.FuelBudget),
		zap.Any("response_headers", responseHeaders))

	// Validate we have a reply topic
	if responseCtx.ResponsesTopic == "" {
		// Try to determine from context
		if msgCtx.ExecutionContext.ResponsesTopic != "" {
			responseCtx.ResponsesTopic = msgCtx.ExecutionContext.ResponsesTopic
		} else if msgCtx.ExecutionContext.FromAgentType != "" {
			responseCtx.ResponsesTopic = fmt.Sprintf("system.agent.%s.responses",
				msgCtx.ExecutionContext.FromAgentType)
		} else {
			// Fallback to generic
			responseCtx.ResponsesTopic = "system.agent.generic.responses"
		}

		p.logger.Warn("Had to construct ResponsesTopic",
			zap.String("responses_topic", responseCtx.ResponsesTopic))
	}

	// Build response message
	response := models.AgentMessage{
		MessageID:         responseCtx.MessageID,
		CorrelationID:     responseCtx.CorrelationID,
		OrchestrationID:   responseCtx.OrchestrationID,
		OrchestrationName: responseCtx.OrchestrationName,
		FromAgentID:       responseCtx.FromAgentID,
		ToAgentID:         responseCtx.ToAgentID,
		MessageType:       "response",
		Action:            "response",
		Data: map[string]interface{}{
			"result":         result,
			"in_response_to": responseCtx.RequestID, // Use the request ID here
		},
		Timestamp: responseCtx.Timestamp,
		Version:   responseCtx.Version,
	}

	// After building response headers
	p.logger.Info("RESPONSE_HEADERS: Built response headers",
		zap.Any("headers", responseHeaders),
		zap.String("target_topic", responseCtx.ResponsesTopic),
	)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Get correlation ID for message key
	key := []byte(responseCtx.CorrelationID)
	if responseCtx.CorrelationID == "" {
		key = []byte(responseCtx.MessageID)
	}

	// Before sending
	p.logger.Info("RESPONSE_SEND: Sending response message",
		zap.String("topic", responseCtx.ResponsesTopic),
		zap.String("key", string(key)),
		zap.Int("payload_size", len(responseBytes)))

	// Send using response context headers
	err = p.producer.Produce(ctx,
		responseCtx.ResponsesTopic,
		responseHeaders,
		key,
		responseBytes)

	if err != nil {
		p.logger.Error("KAFKA_SEND_ERROR: Failed to send message",
			zap.String("topic", responseCtx.ResponsesTopic),
			zap.Error(err))
	} else {
		p.logger.Info("KAFKA_SENT: Message sent successfully",
			zap.String("topic", responseCtx.ResponsesTopic),
			zap.String("key", string(key)),
			zap.Any("headers", responseHeaders))
	}

	return err
}

func (p *MessageProcessor) determineResponsesTopic(ctx context.Context, msgCtx *MessageContext, isChildResponse bool, parentContext map[string]interface{}) string {
	// Priority 1: Child responding to parent's specified topic
	if isChildResponse {
		if responsesTopic, ok := parentContext["responses_topic"].(string); ok && responsesTopic != "" {
			p.logger.Info("Using parent's responses_topic",
				zap.String("responses_topic", responsesTopic))
			return responsesTopic
		}
	}

	// Priority 2: Explicit responses_topic in headers
	if responsesTopic := msgCtx.Headers["responses_topic"]; responsesTopic != "" {
		p.logger.Info("Using responses_topic from headers",
			zap.String("responses_topic", responsesTopic))
		return responsesTopic
	}

	// Priority 3: Construct from parent agent type
	if parentAgentType := msgCtx.Headers["parent_agent_type"]; parentAgentType != "" {
		topic := fmt.Sprintf("system.agent.%s.responses", parentAgentType)
		p.logger.Info("Constructed topic from parent_agent_type",
			zap.String("parent_agent_type", parentAgentType),
			zap.String("responses_topic", topic))
		return topic
	}

	// Priority 4: Look up parent agent type from ID
	if parentAgentID := msgCtx.Headers["parent_agent_id"]; parentAgentID != "" {
		if parentType := p.getAgentTypeFromID(ctx, parentAgentID); parentType != "" {
			topic := fmt.Sprintf("system.agent.%s.responses", parentType)
			p.logger.Info("Constructed topic from parent_agent_id lookup",
				zap.String("parent_agent_id", parentAgentID),
				zap.String("parent_type", parentType),
				zap.String("responses_topic", topic))
			return topic
		}
	}

	// Fallback
	p.logger.Warn("Using fallback response topic")
	return "system.agent.generic.responses"
}

func (p *MessageProcessor) sendResponse(ctx context.Context, msgCtx *MessageContext, headers map[string]string, topic string, result interface{}) error {
	current, caller := getFuncInfo(1)

	p.logger.With(msgCtx.ExecutionContext.LogContext()...).Info("In file processor.go",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	response := models.AgentMessage{
		MessageID:         uuid.New().String(),
		CorrelationID:     headers["correlation_id"],
		OrchestrationID:   headers["orchestration_id"],
		OrchestrationName: headers["orchestration_name"],
		FromAgentID:       headers["from_agent_id"],
		ToAgentID:         headers["to_agent_id"],
		MessageType:       "response",
		Action:            "response",
		Data: map[string]interface{}{
			"result":         result,
			"in_response_to": headers["in_response_to"],
		},
		Timestamp: time.Now(),
		Version:   "2.0",
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal v2 response: %w", err)
	}

	msgCtx.Logger.Info("Sending V2 response",
		zap.String("topic", topic),
		zap.String("orchestration_id", headers["orchestration_id"]),
		zap.String("in_response_to", headers["in_response_to"]))

	err = p.producer.Produce(ctx,
		topic,
		response.ToHeaders(),
		[]byte(response.CorrelationID),
		responseBytes,
	)
	if err != nil {
		p.logger.Error("KAFKA_SEND_ERROR: Failed to send message",
			zap.String("topic", topic),
			zap.Error(err))
	} else {
		p.logger.Info("KAFKA_SENT: Message sent successfully",
			zap.String("topic", topic),
			zap.String("key", string(response.CorrelationID)),
			zap.Any("headers", headers))
	}

	return err
}

func (p *MessageProcessor) normalizeResponseData(result interface{}) map[string]interface{} {
	switch v := result.(type) {
	case map[string]interface{}:
		return v
	case string:
		return map[string]interface{}{"message": v}
	case error:
		return map[string]interface{}{"error": v.Error()}
	default:
		return map[string]interface{}{"result": result}
	}
}

func createSQLDB() (*sql.DB, error) {
	host := os.Getenv("CLIENTS_DB_HOST")
	port := os.Getenv("CLIENTS_DB_PORT")
	user := os.Getenv("CLIENTS_DB_USER")
	password := os.Getenv("CLIENTS_DB_PASSWORD")
	dbname := os.Getenv("DATABASE_DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return sql.Open("pgx", connStr)
}

func (p *MessageProcessor) processNewAgentMessage(ctx context.Context, msg kafka.Message, headers map[string]string, startTime time.Time) error {
	current, caller := getFuncInfo(1)

	p.logger.Info("In file processor.go",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	p.logger.Info("Processing new format agent message",
		zap.String("DEBUG_PROCESSOR_5: from_agent", headers["from_agent_id"]),
		zap.String("DEBUG_PROCESSOR_5: to_agent", headers["to_agent_id"]),
		zap.String("DEBUG_PROCESSOR_5: message_type", headers["message_type"]))

	// Parse the message
	var agentMsg models.AgentMessage
	if err := json.Unmarshal(msg.Value, &agentMsg); err != nil {
		return fmt.Errorf("failed to unmarshal agent message: %w", err)
	}

	// Create MessageContext for compatibility with existing workflow
	msgCtx := &MessageContext{
		Message:       msg,
		Headers:       headers,
		StartTime:     startTime,
		Logger:        p.logger.With(zap.String("message_id", agentMsg.MessageID)),
		CollectedData: agentMsg.Data,
	}

	// Execute using existing workflow engine
	if err := p.process(ctx, msgCtx); err != nil {
		return p.handleError(ctx, msgCtx, err, "processing_failed")
	}

	return nil
}

// getAgentTypeFromID retrieves the agent type from the database using the agent ID
func (p *MessageProcessor) getAgentTypeFromID(ctx context.Context, agentID string) string {
	current, caller := getFuncInfo(1)

	p.logger.Info("In file processor.go",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// First, try to get from the agent_instances table in the client schema
	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		clientID = "demo_client"
	}

	// Try to get agent type from agent instance config
	query := fmt.Sprintf(`
        SELECT config->>'agent_type' 
        FROM client_%s.agent_instances 
        WHERE id = $1 AND is_active = true
        LIMIT 1
    `, clientID)

	var agentType string
	err := p.db.QueryRowContext(ctx, query, agentID).Scan(&agentType)
	if err == nil && agentType != "" {
		p.logger.Debug("Found agent type from instance",
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType))
		return agentType
	}

	// If not found in instances, try to extract from the agent ID pattern
	if strings.Contains(agentID, "-") {
		parts := strings.Split(agentID, "-")
		if len(parts) >= 2 {
			possibleType := parts[1]
			if p.isKnownAgentType(ctx, possibleType) {
				p.logger.Debug("Extracted agent type from ID pattern",
					zap.String("agent_id", agentID),
					zap.String("agent_type", possibleType))
				return possibleType
			}
		}
	}

	p.logger.Warn("Could not determine agent type from ID",
		zap.String("agent_id", agentID))

	return "generic"
}

// isKnownAgentType checks if a given type exists in agent_definitions
func (p *MessageProcessor) isKnownAgentType(ctx context.Context, agentType string) bool {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM agent_definitions 
			WHERE type = $1 AND is_active = true
		)
	`

	var exists bool
	err := p.db.QueryRowContext(ctx, query, agentType).Scan(&exists)
	if err != nil {
		p.logger.Debug("Error checking agent type existence",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return false
	}

	return exists
}

// getAgentTypeAndIDFromHeaders extracts both agent type and ID from message headers
// This is a helper function to ensure we always have both pieces of information
func (p *MessageProcessor) getAgentTypeAndIDFromHeaders(headers map[string]string) (agentType string, agentID string) {
	// Try to get from explicit headers first
	agentType = headers["parent_agent_type"]
	agentID = headers["parent_agent_id"]

	// If we have ID but not type, look it up
	if agentID != "" && agentType == "" {
		agentType = p.getAgentTypeFromID(context.Background(), agentID)
	}

	// If we still don't have the type, try from_agent headers
	if agentType == "" {
		agentType = headers["from_agent_type"]
		if agentType == "" && headers["from_agent_id"] != "" {
			agentType = p.getAgentTypeFromID(context.Background(), headers["from_agent_id"])
		}
	}

	return agentType, agentID
}

func (p *MessageProcessor) selectWorkflow(ctx context.Context, agentDef *actions.AgentDefinition, msgCtx *MessageContext) (models.WorkflowPlan, error) {
	p.logger.Info("DEBUG: selectWorkflow entry",
		zap.Any("default_config_keys", agentDef.DefaultConfig),
		zap.Any("workflow_exists", agentDef.DefaultConfig["workflow"] != nil),
		zap.Any("orchestration_workflow_exists", agentDef.DefaultConfig["orchestration_workflow"] != nil),
		zap.Any("task_workflow_exists", agentDef.DefaultConfig["task_workflow"] != nil))

	// Log the actual workflow content
	if wf, ok := agentDef.DefaultConfig["workflow"]; ok {
		wfMap, _ := wf.(map[string]interface{})
		if wfMap != nil {
			p.logger.Info("DEBUG: Found workflow",
				zap.String("start_step", fmt.Sprintf("%v", wfMap["start_step"])))
		}
	}

	// Check if explicit workflow mode is requested
	workflowMode := msgCtx.Headers["workflow_mode"]
	if workflowMode == "" {
		// Determine based on context
		workflowMode = p.determineWorkflowMode(msgCtx, agentDef)
	}

	p.logger.Info("Selecting workflow",
		zap.String("agent_type", p.agentType),
		zap.String("workflow_mode", workflowMode),
		zap.String("from_agent", msgCtx.Headers["from_agent_id"]))

	var workflowConfig map[string]interface{}

	// Always check for the main workflow field first
	if wf, ok := agentDef.DefaultConfig["workflow"].(map[string]interface{}); ok {
		workflowConfig = wf
	}

	// Only override if specific mode workflows exist
	switch workflowMode {
	case "task":
		if taskWf, ok := agentDef.DefaultConfig["task_workflow"].(map[string]interface{}); ok {
			workflowConfig = taskWf
		}
	case "orchestration":
		if orchWf, ok := agentDef.DefaultConfig["orchestration_workflow"].(map[string]interface{}); ok {
			workflowConfig = orchWf
		}
	}

	// If no workflow found, create a default based on mode
	if workflowConfig == nil {
		if workflowMode == "task" {
			return p.getDefaultTaskWorkflow(), nil
		}
		return p.getDefaultOrchestrationWorkflow(), nil
	}

	return p.convertToWorkflowPlan(workflowConfig), nil
}

func (p *MessageProcessor) determineWorkflowMode(msgCtx *MessageContext, agentDef *actions.AgentDefinition) string {
	// Safe type checking
	if delegationPrefs, ok := agentDef.DefaultConfig["delegation_preferences"].(map[string]interface{}); ok {
		if preferDelegation, ok := delegationPrefs["prefer_delegation"].(bool); ok && preferDelegation {
			if p.isComplexRequest(msgCtx) {
				return "orchestration"
			}
		}
	}

	// Check if called by another agent (subordinate call)
	if msgCtx.Headers["from_agent_id"] != "" &&
		msgCtx.Headers["from_agent_id"] != "00000000-0000-0000-0000-000000000001" {
		return "task"
	}

	// Default based on action
	if action := msgCtx.Headers["action"]; action == "process" || action == "execute" {
		return "task"
	}

	return "orchestration"
}

func (p *MessageProcessor) isComplexRequest(msgCtx *MessageContext) bool {
	// Analyze the request to determine complexity
	// This could look at:
	// - Size of input data
	// - Multiple subtasks mentioned
	// - Keywords indicating complexity
	// - Historical performance data

	// For now, simple heuristic
	if inputData, ok := msgCtx.CollectedData["input_data"].(map[string]interface{}); ok {
		// Check for multiple requirements or complex structure
		if len(inputData) > 5 {
			return true
		}
	}

	return false
}

func (p *MessageProcessor) getDefaultTaskWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "execute",
		Steps: map[string]models.Step{
			"execute": {
				Action:      "execute_llm_prompt",
				Description: "Execute the task",
				NextStep:    "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Complete the task",
			},
		},
	}
}

func (p *MessageProcessor) getDefaultOrchestrationWorkflow() models.WorkflowPlan {
	p.logger.Info("In file processor.go ",
		zap.String("Function: ", "getDefaultOrchestrationWorkflow"),
		zap.String("timestamp: ", time.Now().UTC().Format(time.RFC3339)),
	)

	return models.WorkflowPlan{
		StartStep: "analyze",
		Steps: map[string]models.Step{
			"analyze": {
				Action:      "evaluate_task",
				Description: "Analyze task complexity",
				NextStep:    "decide",
			},
			"decide": {
				Action:      "conditional_route",
				Description: "Decide execution strategy",
				Config: map[string]interface{}{
					"condition_field": "complexity",
					"routes": map[string]interface{}{
						"simple":  "execute",
						"complex": "delegate",
					},
				},
			},
			"delegate": {
				Action:      "call_agent",
				Description: "Delegate to specialized agent",
				NextStep:    "complete",
			},
			"execute": {
				Action:      "execute_llm_prompt",
				Description: "Execute locally",
				NextStep:    "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Complete orchestration",
			},
		},
	}
}

func (p *MessageProcessor) isIntentionalOrchestration(msgCtx *MessageContext) bool {
	// Check if this is an explicit orchestration request
	if mode := msgCtx.Headers["workflow_mode"]; mode == "orchestration" {
		return true
	}

	// Check if the action indicates orchestration
	action := msgCtx.ExecutionContext.Action
	return action == "spawn_group" || action == "start_orchestration" || action == "orchestrate"
}

// Add the helper function to extract action from message
func extractAction(msgValue []byte) string {
	var payload struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(msgValue, &payload); err != nil {
		return "unknown"
	}
	return payload.Action
}

func (p *MessageProcessor) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var callstack []string
	if p.tracer != nil {
		callstack = p.tracer.GetCallStack(12)
	}

	current, caller := getFuncInfo(1)
	p.logger.Info("In processor.go ProcessMessage start",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
		zap.Strings("call stack for processor ProcessMessage", callstack),
	)

	startTime := time.Now()
	headers := kafka.HeadersToMap(msg.Headers)

	// Create ExecutionContext for consistent logging
	execCtx, err := types.FromHeaders(headers)
	if err == nil {
		p.logger.Info("DEBUG returntopic: ExecutionContext created",
			zap.String("responses_topic", execCtx.ResponsesTopic),
			zap.Any("all_headers", headers))
	}
	if err != nil {
		p.logger.Error("FAILED TO CREATE EXECUTION CONTEXT: Failed to create ExecutionContext",
			zap.Error(err),
			zap.Any("headers", headers))
		// Continue with basic processing
		return p.processWithoutContext(ctx, msg, headers)
	}

	contextLogger := p.logger.With(execCtx.LogContext()...)
	contextLogger.Info("In processor.go 1072 ProcessMessage",
		zap.Bool("DEBUGaa: is p.sqlDB exists or is it different driver:", p.sqlDB != nil),
	)

	if p.sqlDB != nil {
		contextLogger.Info("In processor.go 1072 ProcessMessage. p.sqlDB is not nil")

		// Use request_id for deduplication, not message_id
		if execCtx.RequestID != "" {
			repo := orchestration.NewStateRepository(p.sqlDB, p.logger)
			isDuplicate, checkErr := repo.HasProcessedMessage(ctx,
				execCtx.CorrelationID,
				execCtx.RequestID,
				execCtx.ToAgentID)

			if checkErr != nil {
				contextLogger.Error("Failed to check for duplicate request",
					zap.String("request_id", execCtx.RequestID),
					zap.Error(checkErr))
			} else if isDuplicate {
				contextLogger.Warn("Duplicate request detected and ignored",
					zap.String("request_id", execCtx.RequestID),
					zap.String("correlation_id", execCtx.CorrelationID),
					zap.String("orchestration_id", execCtx.OrchestrationID),
				)
				return nil
			}

			// Record this request as processed
			if err := repo.RecordMessageProcessing(ctx, execCtx, execCtx.ToAgentID); err != nil {
				contextLogger.Error("Failed to record request processing", zap.Error(err))
			}
		}
	}

	contextLogger.Info("TRACE: processor.go ProcessMessage entry",
		zap.String("agent_type", p.agentType),
		zap.String("processing_orch", execCtx.OrchestrationID),
		zap.Int("fuel_received", execCtx.FuelBudget),
		zap.String("action", extractAction(msg.Value)))

	// Trace incoming message if enabled in env vars
	if p.tracer != nil {
		p.tracer.TraceMessage(execCtx, "received", msg.Topic, len(msg.Value))
		defer func() {
			if execCtx.MessageType == "response" ||
				execCtx.MessageType == "error" {
				// Dump trace on completion
				p.tracer.DumpTrace(execCtx.CorrelationID)
			}
		}()
	}

	msgCtx, err := NewMessageContext(msg, headers, p.agentType, p.logger)
	if err != nil {
		contextLogger.Error("Failed to create message context", zap.Error(err))
		return err
	}

	msgCtx.Logger = contextLogger

	// Validate context
	if err := msgCtx.ValidateContext(); err != nil {
		contextLogger.Error("Context validation failed", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "invalid_context")
	}

	// Record metrics
	observability.AgentTasksReceived.WithLabelValues(p.agentType, msgCtx.ExecutionContext.Action).Inc()
	defer func() {
		observability.AgentProcessingDuration.WithLabelValues(p.agentType, msgCtx.ExecutionContext.Action).
			Observe(time.Since(startTime).Seconds())
	}()

	// initialise response
	var requestData map[string]interface{}
	json.Unmarshal(msg.Value, &requestData)

	p.logger.Info("processor.go 1146 ProcessMessage",
		zap.String("Action from ExecutionContext", msgCtx.ExecutionContext.Action),
	)

	// Always use the action from the ExecutionContext (derived from the header)
	// for protocol-level decisions like initialization.
	if msgCtx.ExecutionContext.Action == "initialize" {
		contextLogger.Info("Handling protocol action: initialize")

		var spawnRequest types.RequestMessage
		if err := json.Unmarshal(msg.Value, &spawnRequest); err != nil {
			contextLogger.Error("Failed to unmarshal spawn request during initialization", zap.Error(err))
			return err
		}

		// This will now be called correctly, sending the confirmation response.
		return p.initializer.SendInitializationResponse(&spawnRequest)
	}

	contextLogger.Info("processor.go 1161 ProcessMessage",
		zap.String("MessageType from ExecutionContext", msgCtx.ExecutionContext.MessageType),
	)

	// Route responses to orchestrator
	if msgCtx.ExecutionContext.MessageType == "response" {
		contextLogger.Info("Routing response to orchestrator",
			zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID),
			zap.String("status", msgCtx.ExecutionContext.Status),
			zap.Bool("has_orchestrator", p.orchestrator != nil))

		// The orchestrator needs to handle this response
		if p.orchestrator != nil {
			return p.orchestrator.ProcessResponse(ctx, msgCtx.ExecutionContext, msg.Value)
		}

		// If no orchestrator, log and return
		contextLogger.Warn("No orchestrator available to handle response")
		return nil
	}

	// For ALL request messages, extract the body properly
	// This is generic for any agent type and any action
	var requestPayload map[string]interface{}
	if err := json.Unmarshal(msg.Value, &requestPayload); err == nil {
		// Handle both RequestMessage and AgentMessage formats

		// Format 1: RequestMessage with headers and body
		if body, ok := requestPayload["body"].(map[string]interface{}); ok {
			// Merge body contents into CollectedData
			for key, value := range body {
				msgCtx.CollectedData[key] = value
			}

			// Special handling for "data" field if present
			if data, ok := body["data"].(map[string]interface{}); ok {
				msgCtx.CollectedData["input_data"] = data
			}
		}

		// Format 2: AgentMessage with data field
		if data, ok := requestPayload["data"].(map[string]interface{}); ok {
			// If no body was found, use data directly
			if _, hasBody := requestPayload["body"]; !hasBody {
				msgCtx.CollectedData["input_data"] = data
			}
		}

		// Store the action if present
		if action, ok := requestPayload["action"].(string); ok && action != "" {
			msgCtx.CollectedData["requested_action"] = action
		}
	}

	// Process the message
	if err := p.process(ctx, msgCtx); err != nil {
		contextLogger.Error("Processing failed", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "processing_failed")
	}

	// Success
	observability.AgentTasksProcessed.WithLabelValues(p.agentType, msgCtx.ExecutionContext.Action, "success").Inc()
	contextLogger.Info("ProcessMessage (processor.go) completed successfully")

	// Trace outgoing response if one was sent
	if p.tracer != nil && msgCtx.ExecutionContext.MessageType == "request" {
		// The response would have been sent in process()
		// We could track it there too
		p.tracer.TraceMessage(execCtx, "received", msg.Topic, msg.Value)
	}

	return nil
}

// Add fallback for when ExecutionContext creation fails
func (p *MessageProcessor) processWithoutContext(ctx context.Context, msg kafka.Message, headers map[string]string) error {
	// Basic processing without ExecutionContext
	// This ensures the system doesn't break if context is malformed

	current, caller := getFuncInfo(1)

	p.logger.Info("In file processor.go",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	action := extractAction(msg.Value)
	p.logger.Warn("Processing without ExecutionContext",
		zap.String("action", action),
		zap.String("correlation_id", headers["correlation_id"]))

	// Create a minimal message context
	msgCtx := &MessageContext{
		Message:       msg,
		Headers:       headers,
		StartTime:     time.Now(),
		Logger:        p.logger,
		CollectedData: make(map[string]interface{}),
	}

	// Try to process anyway
	if err := p.process(ctx, msgCtx); err != nil {
		p.logger.Error("Processing failed without context",
			zap.Error(err),
			zap.String("action", action))
		return err
	}

	return nil
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

func (p *MessageProcessor) processRequest(ctx context.Context, msgCtx *MessageContext) error {
	p.logger.Info("Processing request processRequest processor.go 1295",
		zap.String("action", msgCtx.ExecutionContext.Action),
		zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID),
		zap.String("request_id", msgCtx.ExecutionContext.RequestID),
		zap.Bool("stateless", msgCtx.IsStateless))

	// Set our agent information
	msgCtx.ExecutionContext.FromAgentType = p.agentType
	msgCtx.ExecutionContext.FromAgentID = p.podName
	msgCtx.ExecutionContext.ProcessingNode = p.podName

	// Load agent configuration
	agentConfig, err := p.configLoader.LoadAgentConfig(p.agentType)
	if err != nil {
		p.logger.Error("Failed to load agent config", zap.Error(err))
		return p.sendErrorResponse(ctx, msgCtx, fmt.Errorf("configuration error: %w", err))
	}

	// Everything goes through workflow execution now
	return p.executeWorkflow(ctx, msgCtx, agentConfig)
}

// processResponse handles incoming responses
func (p *MessageProcessor) processResponse(ctx context.Context, msgCtx *MessageContext) error {
	execCtx := msgCtx.ExecutionContext

	// Check if we have InResponseTo context
	if execCtx.InResponseTo == nil {
		p.logger.Error("Response missing InResponseTo context",
			zap.String("message_id", execCtx.MessageID))
		return fmt.Errorf("response missing InResponseTo context")
	}

	p.logger.Info("Processing response",
		zap.String("request_id", execCtx.InResponseTo.RequestID),
		zap.String("status", execCtx.Status),
		zap.String("from_agent", execCtx.FromAgentID),
		zap.Bool("stateless", msgCtx.IsStateless))

	// Handle response through orchestrator
	if p.orchestrator != nil {
		// Create orchestrator context from InResponseTo
		orchestratorCtx := &types.ExecutionContext{
			// Use the parent orchestration ID if available, otherwise use current
			OrchestrationID: execCtx.InResponseTo.ParentOrchestrationID,
			RequestID:       execCtx.InResponseTo.RequestID,
			StepID:          execCtx.InResponseTo.StepID,
			StepName:        execCtx.InResponseTo.StepName,
			Status:          execCtx.Status,
			RetryVersion:    execCtx.RetryVersion, // Use RetryVersion
		}

		// If no parent orchestration ID, use the current orchestration ID
		if orchestratorCtx.OrchestrationID == "" {
			orchestratorCtx.OrchestrationID = execCtx.OrchestrationID
		}

		return p.orchestrator.ProcessResponse(ctx, orchestratorCtx, msgCtx.Message.Value)
	}

	// Direct response handling for non-orchestrated agents
	return p.handleDirectResponse(ctx, msgCtx)
}

// executeWorkflow executes a workflow through the orchestrator
func (p *MessageProcessor) executeWorkflow(ctx context.Context, msgCtx *MessageContext, config *models.AgentConfig) error {

	p.logger.Info("Executing workflow executeWorkflow processor.go 1361",
		zap.String("start_step", config.Workflow.StartStep),
		zap.Int("total_steps", len(config.Workflow.Steps)),
		zap.String("agent_type", p.agentType),
		zap.Bool("stateless", msgCtx.IsStateless),
		zap.Any("processor 1364 executeWorkflow - does msgCtx have initial data?", msgCtx),
	)

	// Ensure ExecutionContext has agent type
	if msgCtx.ExecutionContext.FromAgentType == "" {
		msgCtx.ExecutionContext.FromAgentType = p.agentType
	}

	// Convert ExecutionContext to headers for orchestrator
	headers := msgCtx.ExecutionContext.ToHeaders()
	fmt.Fprintf(os.Stderr, "DEBUG uuid: processor.executeWorkflow small e printf - headers: %+v\n", headers)

	// Update metrics
	observability.WorkflowsStarted.WithLabelValues(p.agentType, config.Workflow.StartStep, msgCtx.ExecutionContext.ClientID).Inc()
	observability.ActiveWorkflows.WithLabelValues(p.agentType).Inc()
	defer observability.ActiveWorkflows.WithLabelValues(p.agentType).Dec()

	// Include raw message in CollectedData if not already there
	if _, exists := msgCtx.CollectedData["__raw_message__"]; !exists {
		var rawMsg interface{}
		if err := json.Unmarshal(msgCtx.Message.Value, &rawMsg); err == nil {
			msgCtx.CollectedData["__raw_message__"] = rawMsg
		}
	}

	// Marshal CollectedData (which now includes everything)
	initialDataBytes, err := json.Marshal(msgCtx.CollectedData)
	if err != nil {
		p.logger.Error("Failed to marshal collected data", zap.Error(err))
		// Fallback to raw message if marshaling fails
		initialDataBytes = msgCtx.Message.Value
	}

	// Execute through orchestrator (use existing ExecuteWorkflow method)
	return p.orchestrator.ExecuteWorkflow(ctx, config.Workflow, headers, initialDataBytes)
}

// sendSuccessResponse sends a successful response
func (p *MessageProcessor) sendSuccessResponse(ctx context.Context, msgCtx *MessageContext, result interface{}) error {
	response := &types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender: types.AgentIdentity{
				AgentType:    p.agentType,
				AgentID:      p.podName,
				PodName:      p.podName,
				AgentVersion: os.Getenv("AGENT_VERSION"),
			},

			// Response tracking - individual fields, not InResponseTo struct
			InResponseToRequestID:      msgCtx.ExecutionContext.RequestID,
			InResponseToStepID:         msgCtx.ExecutionContext.StepID,
			InResponseToStepName:       msgCtx.ExecutionContext.StepName,
			InResponseToParentOrchID:   msgCtx.ExecutionContext.ParentOrchestrationID,
			InResponseToParentOrchName: msgCtx.ExecutionContext.ParentOrchestrationName,
			InResponseToMessageID:      msgCtx.ExecutionContext.MessageID,
			InResponseToAction:         msgCtx.ExecutionContext.Action,
			RetryCount:                 msgCtx.ExecutionContext.RetryVersion,

			// My context
			MyOrchestrationID:   msgCtx.ExecutionContext.OrchestrationID,
			MyOrchestrationName: msgCtx.ExecutionContext.OrchestrationName,
			MyRequestsTopic:     fmt.Sprintf("system.agent.%s.requests", p.agentType),
			MyResponsesTopic:    fmt.Sprintf("system.agent.%s.responses", p.agentType),

			// Identity
			CorrelationID:   msgCtx.ExecutionContext.CorrelationID,
			CorrelationName: msgCtx.ExecutionContext.CorrelationName,
			ClientID:        msgCtx.ExecutionContext.ClientID,
			MessageType:     "response",
			FromAgent:       p.podName,
			ToAgent:         msgCtx.ExecutionContext.FromAgentID,
			ToAgentType:     msgCtx.ExecutionContext.FromAgentType,

			// Status flags
			Status:              "complete",
			IsComplete:          true,
			IsError:             false,
			IsMultipartResponse: false,
			PartCount:           1,

			// Timing & Resources
			TimeSent:                   time.Now(),
			TimeSpent:                  time.Since(msgCtx.StartTime),
			OverallTimeBudgetRemaining: msgCtx.ExecutionContext.TimeoutSeconds - int(time.Since(msgCtx.StartTime).Seconds()),
			TopicSentTo:                "",  // Will be set below
			FuelUsed:                   100, // Calculate actual fuel usage
			RemainingFuelBudget:        msgCtx.ExecutionContext.FuelBudget - 100,
		},
		Body: types.ResponseBody{
			Success: true,
			Headers: nil,
			Body:    result, // Direct assignment
			Error:   nil,
		},
	}

	// Determine response topic
	responsesTopic := msgCtx.ExecutionContext.ResponsesTopic
	if responsesTopic == "" {
		// Use sender's response topic
		if msgCtx.ExecutionContext.FromAgentType != "" {
			responsesTopic = fmt.Sprintf("system.agent.%s.responses", msgCtx.ExecutionContext.FromAgentType)
		} else {
			responsesTopic = "system.agent.generic.responses"
		}
	}

	// Update the topic in headers
	response.Headers.TopicSentTo = responsesTopic

	// Send response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Convert headers to map for Kafka
	headers := response.Headers.ToMap()

	// Use correlation ID as key
	key := []byte(msgCtx.ExecutionContext.CorrelationID)
	if msgCtx.ExecutionContext.CorrelationID == "" {
		key = []byte(msgCtx.ExecutionContext.MessageID)
	}

	// Producer.Produce signature: (ctx, topic, headers, key, value)
	return p.producer.Produce(ctx, responsesTopic, headers, key, responseBytes)
}

// sendErrorResponse sends an error response
func (p *MessageProcessor) sendErrorResponse(ctx context.Context, msgCtx *MessageContext, err error) error {
	p.logger.Error("Sending error response",
		zap.Error(err),
		zap.String("request_id", msgCtx.ExecutionContext.RequestID))

	response := &types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender: types.AgentIdentity{
				AgentType:    p.agentType,
				AgentID:      p.podName,
				PodName:      p.podName,
				AgentVersion: os.Getenv("AGENT_VERSION"),
			},

			// Response tracking
			InResponseToRequestID:    msgCtx.ExecutionContext.RequestID,
			InResponseToStepID:       msgCtx.ExecutionContext.StepID,
			InResponseToParentOrchID: msgCtx.ExecutionContext.ParentOrchestrationID,
			InResponseToMessageID:    msgCtx.ExecutionContext.MessageID,
			InResponseToAction:       msgCtx.ExecutionContext.Action,

			// My context
			MyOrchestrationID: msgCtx.ExecutionContext.OrchestrationID,

			// Identity
			CorrelationID: msgCtx.ExecutionContext.CorrelationID,
			ClientID:      msgCtx.ExecutionContext.ClientID,
			MessageType:   "response",

			// Status flags
			Status:     "error_unrecoverable",
			IsError:    true,
			IsComplete: true,

			// Timing
			TimeSent: time.Now(),
		},
		Body: types.ResponseBody{
			Success: false,
			Headers: nil,
			Body:    nil, // No result for error
			Error: &types.ErrorInfo{
				Code:        "PROCESSING_ERROR",
				Message:     err.Error(),
				Recoverable: false,
			},
		},
	}

	// Determine response topic
	responsesTopic := msgCtx.ExecutionContext.ResponsesTopic
	if responsesTopic == "" && msgCtx.ExecutionContext.FromAgentType != "" {
		responsesTopic = fmt.Sprintf("system.agent.%s.responses", msgCtx.ExecutionContext.FromAgentType)
	}
	if responsesTopic == "" {
		responsesTopic = "system.agent.generic.errors"
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal error response: %w", err)
	}

	headers := response.Headers.ToMap()
	key := []byte(msgCtx.ExecutionContext.CorrelationID)

	err = p.producer.Produce(ctx, responsesTopic, headers, key, responseBytes)
	if err != nil {
		p.logger.Error("KAFKA_SEND_ERROR: Failed to send message",
			zap.String("topic", responsesTopic),
			zap.Error(err))
	} else {
		p.logger.Info("KAFKA_SENT: Message sent successfully",
			zap.String("topic", responsesTopic),
			zap.String("key", string(key)),
			zap.Any("headers", headers))
	}

	return err
}

// handleDirectResponse handles responses for non-orchestrated flows
func (p *MessageProcessor) handleDirectResponse(ctx context.Context, msgCtx *MessageContext) error {
	// For simple agents that don't use orchestration
	p.logger.Info("Handling direct response",
		zap.String("request_id", msgCtx.ExecutionContext.InResponseTo.RequestID),
		zap.String("status", msgCtx.ExecutionContext.Status))

	// Update metrics
	if msgCtx.ExecutionContext.Status == "complete" {
		observability.ResponsesCompleted.WithLabelValues(p.agentType, "success").Inc()
	} else if strings.Contains(msgCtx.ExecutionContext.Status, "error") {
		observability.ResponsesCompleted.WithLabelValues(p.agentType, "error").Inc()
	}

	return nil
}
