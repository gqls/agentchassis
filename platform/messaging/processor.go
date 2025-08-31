// FILE: platform/messaging/processor.go
package messaging

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
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// MessageProcessor handles processing of Kafka messages for agents
type MessageProcessor struct {
	agentType    string
	db           *pgxpool.Pool
	sqlDB        *sql.DB
	producer     kafka.Producer
	orchestrator *orchestration.SagaCoordinator
	validator    *validation.WorkflowValidator
	configLoader *config.AgentConfigLoader
	logger       *zap.Logger
	tracer       *orchestration.TraceLogger
}

// AgentDefinition represents an agent's configuration from the database
type AgentDefinition struct {
	Type          string
	DisplayName   string
	Description   string
	Category      string
	DefaultConfig map[string]interface{}
	Capabilities  []string
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(
	agentType string,
	db *pgxpool.Pool,
	producer kafka.Producer,
	orchestrator *orchestration.SagaCoordinator,
	validator *validation.WorkflowValidator,
	logger *zap.Logger,
) *MessageProcessor {

	// Create sql.DB connection
	sqlDB, err := createSQLDB()
	if err != nil {
		logger.Error("Failed to create SQL DB connection", zap.Error(err))
		// Continue without it - will fail later if needed
	}

	// Create tracer if enabled
	var tracer *orchestration.TraceLogger
	if os.Getenv("ENABLE_MESSAGE_TRACING") == "true" {
		tracer = orchestration.NewTraceLogger(logger)
	}

	return &MessageProcessor{
		agentType:    agentType,
		db:           db,
		sqlDB:        sqlDB,
		producer:     producer,
		orchestrator: orchestrator,
		validator:    validator,
		configLoader: config.NewAgentConfigLoader(logger),
		logger:       logger,
		tracer:       tracer,
	}
}

// process determines how to handle the message based on agent configuration
func (p *MessageProcessor) process(ctx context.Context, msgCtx *MessageContext) error {
	// Initialize CollectedData if nil
	if msgCtx.CollectedData == nil {
		msgCtx.CollectedData = make(map[string]interface{})
	}

	msgCtx.Logger.Info("Starting message processing",
		zap.String("action", msgCtx.Action),
		zap.String("orchestration_id", msgCtx.ExecutionContext.OrchestrationID))

	/*	// The processor KNOWS its own agent type
		msgCtx.ExecutionContext.FromAgentType = p.agentType
		msgCtx.ExecutionContext.OwnerAgentID = os.Getenv("AGENT_ID")

		// Set reply topic based on THIS agent's type
		msgCtx.ExecutionContext.ReplyToTopic = fmt.Sprintf("system.agent.%s.responses", p.agentType)
	*/

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

	// Load agent definition and continue processing...
	agentDef, err := p.loadAgentDefinition(ctx, p.agentType)
	if err != nil {
		msgCtx.Logger.Warn("Failed to load agent definition, using defaults",
			zap.String("agent_type", p.agentType),
			zap.Error(err))
		return p.processWithDefaults(ctx, msgCtx)
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

	// Validate for self-recursion only if not intentionally orchestrating
	if !p.isIntentionalOrchestration(msgCtx) {
		if err := p.validateNoSelfRecursion(workflow, p.agentType); err != nil {
			msgCtx.Logger.Error("Workflow validation failed - self-recursion detected",
				zap.Error(err),
				zap.String("agent_type", p.agentType))
			return errors.New(errors.ErrWorkflowInvalid, "Invalid workflow: self-recursion detected").
				WithCause(err).
				Build()
		}
	}

	// Store the agent configuration for actions to use
	msgCtx.CollectedData["agent_config"] = agentDef.DefaultConfig

	// Store the input data
	var inputPayload map[string]interface{}
	if err := json.Unmarshal(msgCtx.Message.Value, &inputPayload); err == nil {
		if data, ok := inputPayload["data"]; ok {
			msgCtx.CollectedData["input_data"] = data
			msgCtx.Logger.Info("Input data stored", zap.Any("input_data", data))
		}
		// Store the action for the workflow to reference
		msgCtx.CollectedData["input_action"] = inputPayload["action"]
	}

	// Add debug logging
	p.logger.Info("DEBUG: Determining processing mode",
		zap.String("agent_type", p.agentType),
		zap.String("category", agentDef.Category),
		zap.Any("default_config_mode", agentDef.DefaultConfig["processing_mode"]),
		zap.String("msgCtx.action", msgCtx.Action),
		zap.Bool("has_workflow", workflow.StartStep != ""),
		zap.String("workflow_start", workflow.StartStep),
		zap.Any("payload_action", inputPayload["action"]))

	// Determine processing mode
	processingMode := p.determineProcessingMode(agentDef, inputPayload)

	// If this is a task agent but has a workflow, log a warning
	if processingMode == "task" && len(workflow.Steps) > 0 {
		msgCtx.Logger.Warn("Task agent has workflow - will execute as orchestrator",
			zap.String("agent_type", p.agentType),
			zap.Int("step_count", len(workflow.Steps)))
	}

	msgCtx.Logger.Info("Processing mode determined",
		zap.String("agent_type", p.agentType),
		zap.String("processing_mode", processingMode),
		zap.String("action", msgCtx.Action),
		zap.String("category", agentDef.Category),
		zap.Any("workflow_start_step", workflow.StartStep))

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

	// Execute the workflow
	return p.executeWorkflow(ctx, msgCtx, agentConfig)
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
func (p *MessageProcessor) loadAgentDefinition(ctx context.Context, agentType string) (*AgentDefinition, error) {
	p.logger.Debug("Loading agent definition", zap.String("agent_type", agentType))

	query := `
		SELECT type, display_name, description, category, default_config, capabilities
		FROM agent_definitions
		WHERE type = $1
	`

	var def AgentDefinition
	var configJSON []byte
	var capabilitiesJSON []byte

	err := p.db.QueryRow(ctx, query, agentType).Scan(
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

	// Parse the JSON config
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

// determineProcessingMode figures out how this agent should process messages
func (p *MessageProcessor) determineProcessingMode(agentDef *AgentDefinition, payload map[string]interface{}) string {
	// Check if explicitly configured
	if mode, ok := agentDef.DefaultConfig["processing_mode"].(string); ok {
		return mode
	}

	// Check agent category
	if agentDef.Category == "orchestrator" {
		return "orchestrator"
	}

	// Check based on action in the message
	action, _ := payload["action"].(string)
	switch action {
	case "spawn_group", "start_orchestration", "spawn_agent":
		return "orchestrator"
	case "call_agent", "process_task":
		return "task"
	default:
		// Default based on agent type
		orchestratorTypes := map[string]bool{
			"generic":         true,
			"orchestrator":    true,
			"website-builder": true,
		}
		if orchestratorTypes[p.agentType] {
			return "orchestrator"
		}
		return "task"
	}
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

// processWithDefaults handles processing when no agent definition is found
func (p *MessageProcessor) processWithDefaults(ctx context.Context, msgCtx *MessageContext) error {
	msgCtx.Logger.Info("Processing with defaults")

	// Use a simple default workflow
	workflow := p.getDefaultWorkflow()

	agentConfig := &models.AgentConfig{
		CoreLogic: map[string]interface{}{
			"agent_type": p.agentType,
		},
		Workflow: workflow,
	}

	return p.executeWorkflow(ctx, msgCtx, agentConfig)
}

func (p *MessageProcessor) executeWorkflow(ctx context.Context, msgCtx *MessageContext, config *models.AgentConfig) error {
	msgCtx.Logger.Info("Executing workflow",
		zap.String("start_step", config.Workflow.StartStep),
		zap.Int("total_steps", len(config.Workflow.Steps)),
		zap.String("agent_type", p.agentType))

	// Ensure ExecutionContext has agent type
	if msgCtx.ExecutionContext.FromAgentType == "" {
		msgCtx.ExecutionContext.FromAgentType = p.agentType
	}

	// Sync to headers for orchestrator compatibility
	msgCtx.SyncHeadersFromContext()

	// Update metrics
	observability.WorkflowsStarted.WithLabelValues(p.agentType, config.Workflow.StartStep, msgCtx.Headers["client_id"]).Inc()
	observability.ActiveWorkflows.WithLabelValues(p.agentType).Inc()
	defer observability.ActiveWorkflows.WithLabelValues(p.agentType).Dec()

	// Execute through orchestrator
	err := p.orchestrator.ExecuteWorkflow(ctx, config.Workflow, msgCtx.Headers, msgCtx.Message.Value)

	if err != nil {
		msgCtx.Logger.Error("Workflow execution failed", zap.Error(err))
		return p.sendWorkflowFailureResponse(ctx, msgCtx, err)
	}

	// Workflow started successfully (might be running, waiting, or completed)
	// The coordinator handles all state management internally
	return p.sendWorkflowSuccessResponse(ctx, msgCtx)
}

// New response methods using ExecutionContext
func (p *MessageProcessor) sendWorkflowSuccessResponse(ctx context.Context, msgCtx *MessageContext) error {
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
	return p.sendWorkflowResponse(ctx, msgCtx, map[string]interface{}{
		"error":  err.Error(),
		"status": "failed",
	})
}

func (p *MessageProcessor) handleError(ctx context.Context, msgCtx *MessageContext, err error, errorType string) error {
	msgCtx.Logger.Error("Processing failed", zap.Error(err))
	observability.AgentTasksProcessed.WithLabelValues(p.agentType, msgCtx.Action, errorType).Inc()

	// Check for specific error types
	if domainErr, ok := err.(*errors.DomainError); ok {
		if domainErr.Code == errors.ErrInsufficientFuel {
			observability.FuelExhausted.WithLabelValues(p.agentType, msgCtx.Action, msgCtx.Headers["client_id"]).Inc()
		}
		p.sendErrorResponse(ctx, msgCtx, domainErr)
	} else {
		p.sendErrorResponse(ctx, msgCtx, errors.InternalError("Processing failed", err))
	}

	return err
}

func (p *MessageProcessor) sendErrorResponse(ctx context.Context, msgCtx *MessageContext, domainErr *errors.DomainError) {
	responseHeaders := msgCtx.CreateResponseHeaders(p.agentType)
	domainErr.TraceID = msgCtx.Headers["correlation_id"]

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   domainErr,
		"agent":   p.agentType,
	}

	responseBytes, _ := json.Marshal(errorResponse)
	errorTopic := fmt.Sprintf("system.agent.%s.errors", p.agentType)

	if err := p.producer.Produce(ctx, errorTopic, responseHeaders,
		[]byte(msgCtx.Headers["correlation_id"]), responseBytes); err != nil {
		msgCtx.Logger.Error("Failed to send error response", zap.Error(err))
		observability.SystemErrors.WithLabelValues(p.agentType, "produce_error").Inc()
	} else {
		observability.KafkaMessagesProduced.WithLabelValues(errorTopic).Inc()
	}
}

// ProcessResponse handles response messages for orchestrated workflows
func (p *MessageProcessor) ProcessResponse(ctx context.Context, msg kafka.Message) error {
	headers := kafka.HeadersToMap(msg.Headers)

	p.logger.Info("Processing orchestration response",
		zap.String("DEBUG_PROCESSOR_3: correlation_id", headers["correlation_id"]),
		zap.String("DEBUG_PROCESSOR_3: orchestration_id", headers["orchestration_id"]),
		zap.String("DEBUG_PROCESSOR_3: causation_id", headers["causation_id"]))

	// Route to orchestrator
	return p.orchestrator.HandleResponse(ctx, headers, msg.Value)
}

func (p *MessageProcessor) sendWorkflowResponse(ctx context.Context, msgCtx *MessageContext, result interface{}) error {
	// Create response context
	responseCtx := msgCtx.CreateResponseContext()

	contextLogger := p.logger.With(msgCtx.ExecutionContext.LogContext()...)

	// Check if we're completing an action that was waiting
	// The result from spawn_group contains the request_id we should respond with
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Check for await_response flag and request_id
		if awaitResponse, ok := resultMap["await_response"].(bool); ok && awaitResponse {
			if requestID, ok := resultMap["request_id"].(string); ok && requestID != "" {
				// This is the request_id the orchestrator is waiting for
				responseCtx.InResponseTo = requestID
				responseCtx.RequestID = requestID

				contextLogger.Info("Using action's request_id for response",
					zap.String("action_request_id", requestID),
					zap.String("original_request_id", msgCtx.ExecutionContext.RequestID))
			}
		}
	}

	// Trace outgoing response
	if p.tracer != nil {
		p.tracer.TraceMessage(responseCtx, "sending_response", responseCtx.ReplyToTopic, msgCtx.Message.Value)
	}

	// Determine target orchestration
	targetOrchestrationID := msgCtx.ExecutionContext.OrchestrationID
	if msgCtx.ExecutionContext.ParentOrchestrationID != "" {
		targetOrchestrationID = msgCtx.ExecutionContext.ParentOrchestrationID
		contextLogger.Info("TRACE: Child sending response to parent",
			zap.String("child_orch", msgCtx.ExecutionContext.OrchestrationID),
			zap.String("parent_orch", targetOrchestrationID))
	}

	responseHeaders := map[string]string{
		"correlation_id":   msgCtx.ExecutionContext.CorrelationID,
		"orchestration_id": targetOrchestrationID,
		"message_type":     "response",
		"in_response_to":   responseCtx.InResponseTo,
		"fuel_budget":      fmt.Sprintf("%d", msgCtx.ExecutionContext.FuelBudget),
		// Add other required headers
		"message_id":      responseCtx.MessageID,
		"request_id":      responseCtx.RequestID,
		"from_agent_id":   responseCtx.FromAgentID,
		"from_agent_type": responseCtx.FromAgentType,
		"to_agent_id":     responseCtx.ToAgentID,
		"to_agent_type":   responseCtx.ToAgentType,
		"client_id":       msgCtx.ExecutionContext.ClientID,
		"timestamp":       responseCtx.Timestamp.Format(time.RFC3339),
		"version":         responseCtx.Version,
	}

	contextLogger.Info("TRACE: Sending workflow response",
		zap.String("response_orch_id", targetOrchestrationID),
		zap.String("reply_topic", msgCtx.ExecutionContext.ReplyToTopic),
		zap.Int("fuel_returning", msgCtx.ExecutionContext.FuelBudget),
		zap.Any("response_headers", responseHeaders))

	// Validate we have a reply topic
	if responseCtx.ReplyToTopic == "" {
		// Try to determine from context
		if msgCtx.ExecutionContext.ReplyToTopic != "" {
			responseCtx.ReplyToTopic = msgCtx.ExecutionContext.ReplyToTopic
		} else if msgCtx.ExecutionContext.FromAgentType != "" {
			responseCtx.ReplyToTopic = fmt.Sprintf("system.agent.%s.responses",
				msgCtx.ExecutionContext.FromAgentType)
		} else {
			// Fallback to generic
			responseCtx.ReplyToTopic = "system.agent.generic.responses"
		}

		p.logger.Warn("Had to construct ReplyToTopic",
			zap.String("reply_to_topic", responseCtx.ReplyToTopic))
	}

	// Build response message
	response := models.AgentMessage{
		MessageID:       responseCtx.MessageID,
		CorrelationID:   responseCtx.CorrelationID,
		OrchestrationID: responseCtx.OrchestrationID,
		FromAgentID:     responseCtx.FromAgentID,
		ToAgentID:       responseCtx.ToAgentID,
		MessageType:     "response",
		Action:          "response",
		Data: map[string]interface{}{
			"result":         result,
			"in_response_to": responseCtx.InResponseTo,
		},
		Timestamp: responseCtx.Timestamp,
		Version:   responseCtx.Version,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send using response context headers
	return p.producer.Produce(ctx,
		responseCtx.ReplyToTopic,
		responseHeaders,
		[]byte(responseCtx.CorrelationID),
		responseBytes)
}

func (p *MessageProcessor) determineResponseTopic(ctx context.Context, msgCtx *MessageContext, isChildResponse bool, parentContext map[string]interface{}) string {
	// Priority 1: Child responding to parent's specified topic
	if isChildResponse {
		if replyTopic, ok := parentContext["reply_to_topic"].(string); ok && replyTopic != "" {
			p.logger.Info("Using parent's reply_to_topic",
				zap.String("reply_to_topic", replyTopic))
			return replyTopic
		}
	}

	// Priority 2: Explicit reply_to_topic in headers
	if replyTopic := msgCtx.Headers["reply_to_topic"]; replyTopic != "" {
		p.logger.Info("Using reply_to_topic from headers",
			zap.String("reply_to_topic", replyTopic))
		return replyTopic
	}

	// Priority 3: Construct from parent agent type
	if parentAgentType := msgCtx.Headers["parent_agent_type"]; parentAgentType != "" {
		topic := fmt.Sprintf("system.agent.%s.responses", parentAgentType)
		p.logger.Info("Constructed topic from parent_agent_type",
			zap.String("parent_agent_type", parentAgentType),
			zap.String("response_topic", topic))
		return topic
	}

	// Priority 4: Look up parent agent type from ID
	if parentAgentID := msgCtx.Headers["parent_agent_id"]; parentAgentID != "" {
		if parentType := p.getAgentTypeFromID(ctx, parentAgentID); parentType != "" {
			topic := fmt.Sprintf("system.agent.%s.responses", parentType)
			p.logger.Info("Constructed topic from parent_agent_id lookup",
				zap.String("parent_agent_id", parentAgentID),
				zap.String("parent_type", parentType),
				zap.String("response_topic", topic))
			return topic
		}
	}

	// Fallback
	p.logger.Warn("Using fallback response topic")
	return "system.agent.generic.responses"
}

func (p *MessageProcessor) sendResponse(ctx context.Context, msgCtx *MessageContext, headers map[string]string, topic string, result interface{}) error {
	response := models.AgentMessage{
		MessageID:       uuid.New().String(),
		CorrelationID:   headers["correlation_id"],
		OrchestrationID: headers["orchestration_id"],
		FromAgentID:     headers["from_agent_id"],
		ToAgentID:       headers["to_agent_id"],
		MessageType:     "response",
		Action:          "response",
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

	return p.producer.Produce(ctx, topic, response.ToHeaders(),
		[]byte(response.CorrelationID), responseBytes)
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
	host := os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST")
	port := os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT")
	user := os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER")
	password := os.Getenv("CLIENTS_DB_PASSWORD")
	dbname := os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return sql.Open("pgx", connStr)
}

func (p *MessageProcessor) processNewAgentMessage(ctx context.Context, msg kafka.Message, headers map[string]string, startTime time.Time) error {
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
		Action:        agentMsg.Action,
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
	err := p.db.QueryRow(ctx, query, agentID).Scan(&agentType)
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
	err := p.db.QueryRow(ctx, query, agentType).Scan(&exists)
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

func (p *MessageProcessor) selectWorkflow(ctx context.Context, agentDef *AgentDefinition, msgCtx *MessageContext) (models.WorkflowPlan, error) {
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

func (p *MessageProcessor) determineWorkflowMode(msgCtx *MessageContext, agentDef *AgentDefinition) string {
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
	action := msgCtx.Action
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

// Update ProcessMessage to use the tracer properly
func (p *MessageProcessor) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	startTime := time.Now()
	headers := kafka.HeadersToMap(msg.Headers)

	// Create ExecutionContext for consistent logging
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		p.logger.Error("Failed to create ExecutionContext",
			zap.Error(err),
			zap.Any("headers", headers))
		// Continue with basic processing
		return p.processWithoutContext(ctx, msg, headers)
	}

	contextLogger := p.logger.With(execCtx.LogContext()...)

	contextLogger.Info("TRACE: ProcessMessage entry",
		zap.String("agent_type", p.agentType),
		zap.String("processing_orch", execCtx.OrchestrationID),
		zap.Int("fuel_received", execCtx.FuelBudget),
		zap.String("action", extractAction(msg.Value)))

	// Create logger with full context
	//contextLogger := p.logger.With(execCtx.LogContext()...)

	// Or use compact logging for less verbosity
	// p.Logger.Debug("Quick check", execCtx.LogCompact()...)

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

	// Create message context with ExecutionContext
	msgCtx, err := NewMessageContext(msg, headers)
	if err != nil {
		contextLogger.Error("Failed to create message context", zap.Error(err))
		return err
	}

	msgCtx.Logger = contextLogger

	// Extract action
	if err := msgCtx.ExtractAction(); err != nil {
		contextLogger.Error("Failed to extract action", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "invalid_payload")
	}

	// Validate context
	if err := msgCtx.ValidateContext(); err != nil {
		contextLogger.Error("Context validation failed", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "invalid_context")
	}

	// Record metrics
	observability.AgentTasksReceived.WithLabelValues(p.agentType, msgCtx.Action).Inc()
	defer func() {
		observability.AgentProcessingDuration.WithLabelValues(p.agentType, msgCtx.Action).
			Observe(time.Since(startTime).Seconds())
	}()

	// Process the message
	if err := p.process(ctx, msgCtx); err != nil {
		contextLogger.Error("Processing failed", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "processing_failed")
	}

	// Success
	observability.AgentTasksProcessed.WithLabelValues(p.agentType, msgCtx.Action, "success").Inc()
	contextLogger.Info("ProcessMessage completed successfully")

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

	action := extractAction(msg.Value)
	p.logger.Warn("Processing without ExecutionContext",
		zap.String("action", action),
		zap.String("correlation_id", headers["correlation_id"]))

	// Create a minimal message context
	msgCtx := &MessageContext{
		Message:       msg,
		Headers:       headers,
		Action:        action,
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
