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
	return &MessageProcessor{
		agentType:    agentType,
		db:           db,
		sqlDB:        sqlDB,
		producer:     producer,
		orchestrator: orchestrator,
		validator:    validator,
		configLoader: config.NewAgentConfigLoader(logger),
		logger:       logger,
	}
}

// ProcessMessage handles a single message
func (p *MessageProcessor) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	startTime := time.Now()
	headers := kafka.HeadersToMap(msg.Headers)

	p.logger.Info("ProcessMessage started - agent type check",
		zap.String("processor_agent_type", p.agentType),
		zap.String("env_agent_type", os.Getenv("AGENT_TYPE")))

	// Create a message context for this specific message
	msgCtx := &MessageContext{
		Message:   msg,
		Headers:   headers,
		StartTime: startTime,
		Logger: p.logger.With(
			zap.String("correlation_id", headers["correlation_id"]),
			zap.String("request_id", headers["request_id"]),
			zap.String("client_id", headers["client_id"]),
			zap.String("agent_instance_id", headers["agent_instance_id"]),
		),
	}

	msgCtx.Logger.Info("ProcessMessage started",
		zap.String("agent_type", p.agentType),
		zap.Int("payload_size", len(msg.Value)))

	// Extract action
	if err := msgCtx.ExtractAction(); err != nil {
		msgCtx.Logger.Error("Failed to extract action", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "invalid_payload")
	}

	msgCtx.Logger.Info("Action extracted", zap.String("action", msgCtx.Action))

	// Record metrics
	observability.AgentTasksReceived.WithLabelValues(p.agentType, msgCtx.Action).Inc()
	defer func() {
		observability.AgentProcessingDuration.WithLabelValues(p.agentType, msgCtx.Action).
			Observe(time.Since(startTime).Seconds())
	}()

	// Process the message
	if err := p.process(ctx, msgCtx); err != nil {
		msgCtx.Logger.Error("Processing failed in process()", zap.Error(err))
		return p.handleError(ctx, msgCtx, err, "processing_failed")
	}

	// Success
	observability.AgentTasksProcessed.WithLabelValues(p.agentType, msgCtx.Action, "success").Inc()
	msgCtx.Logger.Info("ProcessMessage completed successfully")
	return nil
}

// process determines how to handle the message based on agent configuration
func (p *MessageProcessor) process(ctx context.Context, msgCtx *MessageContext) error {
	// Initialize CollectedData if nil
	if msgCtx.CollectedData == nil {
		msgCtx.CollectedData = make(map[string]interface{})
	}

	msgCtx.Logger.Info("Starting message processing", zap.String("action", msgCtx.Action))

	// Validate headers
	if err := msgCtx.ValidateHeaders(); err != nil {
		msgCtx.Logger.Error("Header validation failed", zap.Error(err))
		return errors.ValidationError("headers", err.Error())
	}

	msgCtx.Logger.Info("Headers validated successfully")

	// Load the complete agent definition from the database
	agentDef, err := p.loadAgentDefinition(ctx, p.agentType)
	if err != nil {
		msgCtx.Logger.Warn("Failed to load agent definition, using defaults",
			zap.String("agent_type", p.agentType),
			zap.Error(err))
		// Continue with default behavior
		return p.processWithDefaults(ctx, msgCtx)
	}

	msgCtx.Logger.Info("Agent definition loaded",
		zap.String("display_name", agentDef.DisplayName),
		zap.String("category", agentDef.Category))

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

	// Extract workflow from agent definition
	var workflow models.WorkflowPlan
	if workflowConfig, ok := agentDef.DefaultConfig["workflow"].(map[string]interface{}); ok {
		// Convert workflow config to WorkflowPlan
		workflow = p.convertToWorkflowPlan(workflowConfig)
		msgCtx.Logger.Info("Workflow extracted from agent definition",
			zap.String("start_step", workflow.StartStep),
			zap.Int("step_count", len(workflow.Steps)))
	} else {
		// No workflow defined, use a simple default
		workflow = p.getDefaultWorkflow()
		msgCtx.Logger.Info("Using default workflow",
			zap.String("start_step", workflow.StartStep),
			zap.Int("step_count", len(workflow.Steps)))
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

	// Ensure agent_type is in headers for actions
	if msgCtx.Headers["agent_type"] == "" {
		msgCtx.Headers["agent_type"] = p.agentType
		msgCtx.Logger.Info("Set agent_type in headers in processor.go executeWorkflow",
			zap.String("agent_type", p.agentType))
	}

	// Start workflow timer
	workflowTimer := observability.StartWorkflowTimer(p.agentType, config.Workflow.StartStep)
	defer workflowTimer.Complete("success")

	// Update metrics
	observability.WorkflowsStarted.WithLabelValues(p.agentType, config.Workflow.StartStep, msgCtx.Headers["client_id"]).Inc()
	observability.ActiveWorkflows.WithLabelValues(p.agentType).Inc()
	defer observability.ActiveWorkflows.WithLabelValues(p.agentType).Dec()

	// Execute through orchestrator
	err := p.orchestrator.ExecuteWorkflow(ctx, config.Workflow, msgCtx.Headers, msgCtx.Message.Value)
	if err != nil {
		msgCtx.Logger.Error("Workflow execution failed", zap.Error(err))
		// Send failure response
		p.sendWorkflowResponse(ctx, msgCtx, map[string]interface{}{
			"error":  err.Error(),
			"status": "failed",
		})
	} else {
		msgCtx.Logger.Info("Workflow execution completed successfully")

		// Use p.sqlDB for the repository
		if p.sqlDB != nil {
			// Get the final result from orchestrator state
			repo := orchestration.NewStateRepository(p.sqlDB, msgCtx.Logger)
			state, _ := repo.GetState(ctx, msgCtx.Headers["correlation_id"])

			if err != nil {
				msgCtx.Logger.Warn("Failed to get final state", zap.Error(err))
				// Send basic success response
				p.sendWorkflowResponse(ctx, msgCtx, map[string]interface{}{
					"status":  "completed",
					"message": "Workflow completed but state retrieval failed",
				})
			} else if state != nil && state.CollectedData != nil {
				// Send success response with collected data
				p.sendWorkflowResponse(ctx, msgCtx, state.CollectedData)
			} else {
				// Send basic success response
				p.sendWorkflowResponse(ctx, msgCtx, map[string]interface{}{
					"status": "completed",
				})
			}
		} else {
			// No SQL DB available, send basic response
			p.sendWorkflowResponse(ctx, msgCtx, map[string]interface{}{
				"status":  "completed",
				"message": "Workflow completed",
			})
		}
	}

	return err
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
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("causation_id", headers["causation_id"]))

	// Route to orchestrator
	return p.orchestrator.HandleResponse(ctx, headers, msg.Value)
}

func (p *MessageProcessor) sendWorkflowResponse(ctx context.Context, msgCtx *MessageContext, result interface{}) error {
	// Determine response topic from headers
	responseTopic := ""

	// Check for reply_to_topic in headers (should be type-based)
	if replyTopic := msgCtx.Headers["reply_to_topic"]; replyTopic != "" {
		responseTopic = replyTopic
		msgCtx.Logger.Info("Using reply_to_topic from headers",
			zap.String("reply_to_topic", replyTopic))
	} else {
		// Fallback: construct from parent agent TYPE if available
		if parentAgentType := msgCtx.Headers["parent_agent_type"]; parentAgentType != "" {
			responseTopic = fmt.Sprintf("system.agent.%s.responses", parentAgentType)
			msgCtx.Logger.Info("Constructed response topic from parent_agent_type",
				zap.String("parent_agent_type", parentAgentType),
				zap.String("response_topic", responseTopic))
		} else if parentAgentID := msgCtx.Headers["parent_agent_id"]; parentAgentID != "" {
			// Try to get type from DB using agent ID
			parentType := p.getAgentTypeFromID(ctx, parentAgentID)
			if parentType != "" {
				responseTopic = fmt.Sprintf("system.agent.%s.responses", parentType)
			}
		} else {
			// Last resort fallback
			responseTopic = "system.orchestrator.responses"
			msgCtx.Logger.Warn("No reply_to_topic or parent agent info, using legacy topic")
		}
	}

	// Check if this was a new format message
	if msgCtx.Headers["message_version"] == "2.0" && responseTopic != "" {
		// New format - use AgentMessage
		response := models.AgentMessage{
			MessageID:     uuid.New().String(),
			CorrelationID: msgCtx.Headers["correlation_id"],
			FromAgentID:   os.Getenv("AGENT_ID"),
			ToAgentID:     msgCtx.Headers["from_agent_id"], // Reply to sender
			MessageType:   "response",
			Action:        "response",
			Data: map[string]interface{}{
				"result":     result,
				"request_id": msgCtx.Headers["message_id"],
			},
			Timestamp: time.Now(),
			Version:   "2.0",
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		return p.producer.Produce(ctx, responseTopic,
			response.ToHeaders(),
			[]byte(response.CorrelationID),
			responseBytes)
	}

	// Legacy format - use TaskResponse
	// Convert result to map[string]interface{} if needed
	var responseData map[string]interface{}

	switch v := result.(type) {
	case map[string]interface{}:
		responseData = v
	case string:
		// If it's a string (like an error), wrap it
		responseData = map[string]interface{}{
			"message": v,
		}
	case error:
		// If it's an error, include the error message
		responseData = map[string]interface{}{
			"error": v.Error(),
		}
	default:
		// For any other type, wrap it in a result field
		responseData = map[string]interface{}{
			"result": result,
		}
	}

	// Prepare response
	response := models.TaskResponse{
		Success: true,
		Data:    responseData,
	}

	// If the result contains an error, mark as not successful
	if _, hasError := responseData["error"]; hasError {
		response.Success = false
		if errStr, ok := responseData["error"].(string); ok {
			response.Error = errStr
		}
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Prepare headers for response
	responseHeaders := make(map[string]string)
	responseHeaders["correlation_id"] = msgCtx.Headers["correlation_id"]
	responseHeaders["causation_id"] = msgCtx.Headers["request_id"]
	responseHeaders["agent_type"] = p.agentType

	// Use agent ID from environment if not in headers
	agentInstanceID := msgCtx.Headers["agent_instance_id"]
	if agentInstanceID == "" {
		agentInstanceID = os.Getenv("AGENT_ID")
	}
	responseHeaders["agent_instance_id"] = agentInstanceID

	// Include reply context
	responseHeaders["from_agent_id"] = os.Getenv("AGENT_ID")
	if toAgent := msgCtx.Headers["from_agent_id"]; toAgent != "" {
		responseHeaders["to_agent_id"] = toAgent
	}

	msgCtx.Logger.Info("Sending workflow response",
		zap.String("response_topic", responseTopic),
		zap.String("correlation_id", responseHeaders["correlation_id"]),
		zap.String("causation_id", responseHeaders["causation_id"]),
		zap.Bool("success", response.Success))

	// Send response
	err = p.producer.Produce(ctx, responseTopic, responseHeaders,
		[]byte(responseHeaders["correlation_id"]), responseBytes)

	if err != nil {
		msgCtx.Logger.Error("Failed to send response", zap.Error(err))
		return fmt.Errorf("failed to send response: %w", err)
	}

	msgCtx.Logger.Info("Response sent successfully",
		zap.String("topic", responseTopic))
	return nil
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
		zap.String("from_agent", headers["from_agent_id"]),
		zap.String("to_agent", headers["to_agent_id"]),
		zap.String("message_type", headers["message_type"]))

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
		clientID = "demo_client" // Default client
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

	// If not found in instances, try to extract from the agent ID itself
	// Sometimes agent IDs are structured like "agent-{type}-{uuid}"
	if strings.Contains(agentID, "-") {
		parts := strings.Split(agentID, "-")
		if len(parts) >= 2 {
			// Check if this looks like a known agent type
			possibleType := parts[1]
			if p.isKnownAgentType(ctx, possibleType) {
				p.logger.Debug("Extracted agent type from ID pattern",
					zap.String("agent_id", agentID),
					zap.String("agent_type", possibleType))
				return possibleType
			}
		}
	}

	// Try to get from orchestrator state if this is a child agent
	stateQuery := `
		SELECT collected_data->>'agent_type'
		FROM orchestrator_state 
		WHERE correlation_id = $1
		LIMIT 1
	`

	err = p.db.QueryRow(ctx, stateQuery, agentID).Scan(&agentType)
	if err == nil && agentType != "" {
		p.logger.Debug("Found agent type from orchestrator state",
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType))
		return agentType
	}

	// Check if it's a well-known agent ID
	wellKnownAgents := map[string]string{
		"00000000-0000-0000-0000-000000000001": "orchestrator",
		"00000000-0000-0000-0000-000000000002": "generic",
	}

	if knownType, ok := wellKnownAgents[agentID]; ok {
		p.logger.Debug("Found well-known agent type",
			zap.String("agent_id", agentID),
			zap.String("agent_type", knownType))
		return knownType
	}

	p.logger.Warn("Could not determine agent type from ID",
		zap.String("agent_id", agentID))

	// Default fallback
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
