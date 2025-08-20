// FILE: platform/messaging/processor.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	return &MessageProcessor{
		agentType:    agentType,
		db:           db,
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

	// Determine processing mode
	processingMode := p.determineProcessingMode(agentDef, inputPayload)

	msgCtx.Logger.Info("Processing mode determined",
		zap.String("agent_type", p.agentType),
		zap.String("processing_mode", processingMode),
		zap.String("action", msgCtx.Action))

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
		zap.Int("total_steps", len(config.Workflow.Steps)))

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
	} else {
		msgCtx.Logger.Info("Workflow execution completed successfully")
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
	errorTopic := fmt.Sprintf("system.errors.%s", p.agentType)

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
