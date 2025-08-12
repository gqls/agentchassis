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

	// Extract action
	if err := msgCtx.ExtractAction(); err != nil {
		return p.handleError(ctx, msgCtx, err, "invalid_payload")
	}

	// Record metrics
	observability.AgentTasksReceived.WithLabelValues(p.agentType, msgCtx.Action).Inc()
	defer func() {
		observability.AgentProcessingDuration.WithLabelValues(p.agentType, msgCtx.Action).
			Observe(time.Since(startTime).Seconds())
	}()

	// Process the message
	if err := p.process(ctx, msgCtx); err != nil {
		return p.handleError(ctx, msgCtx, err, "processing_failed")
	}

	// Success
	observability.AgentTasksProcessed.WithLabelValues(p.agentType, msgCtx.Action, "success").Inc()
	return nil
}

// In platform/messaging/processor.go
func (p *MessageProcessor) process(ctx context.Context, msgCtx *MessageContext) error {
	// Initialize CollectedData if nil
	if msgCtx.CollectedData == nil {
		msgCtx.CollectedData = make(map[string]interface{})
	}

	// Validate headers
	if err := msgCtx.ValidateHeaders(); err != nil {
		return errors.ValidationError("headers", err.Error())
	}

	// Load agent configuration from the agent_instances table
	agentConfig, err := p.configLoader.LoadFromDatabase(
		ctx,
		p.db,
		msgCtx.Headers["client_id"],
		msgCtx.Headers["agent_instance_id"],
		p.agentType,
	)
	if err != nil {
		return errors.InternalError("Failed to load configuration", err)
	}

	// Also load the agent definition for the prompt template
	// This is needed for execute_llm_prompt action
	var agentDefinitionConfig []byte // Use []byte for JSON columns
	query := `SELECT default_config FROM agent_definitions WHERE type = $1`

	// Use the pgxpool's QueryRow which returns a Row that has Scan method
	row := p.db.QueryRow(ctx, query, p.agentType)
	err = row.Scan(&agentDefinitionConfig)

	if err == nil && len(agentDefinitionConfig) > 0 {
		var defConfig map[string]interface{}
		if err := json.Unmarshal(agentDefinitionConfig, &defConfig); err == nil {
			// Merge definition config with instance config
			// Instance config takes precedence
			if agentConfig.CoreLogic == nil {
				agentConfig.CoreLogic = make(map[string]interface{})
			}

			// Add prompt_template and ai_service from definition if not in instance
			if promptTemplate, ok := defConfig["prompt_template"]; ok {
				if _, exists := agentConfig.CoreLogic["prompt_template"]; !exists {
					agentConfig.CoreLogic["prompt_template"] = promptTemplate
				}
			}
			if aiService, ok := defConfig["ai_service"]; ok {
				if _, exists := agentConfig.CoreLogic["ai_service"]; !exists {
					agentConfig.CoreLogic["ai_service"] = aiService
				}
			}

			// Also add temperature and other config values
			if temperature, ok := defConfig["temperature"]; ok {
				if _, exists := agentConfig.CoreLogic["temperature"]; !exists {
					agentConfig.CoreLogic["temperature"] = temperature
				}
			}
			if maxTokens, ok := defConfig["max_tokens"]; ok {
				if _, exists := agentConfig.CoreLogic["max_tokens"]; !exists {
					agentConfig.CoreLogic["max_tokens"] = maxTokens
				}
			}
		}
	} else if err != nil {
		// Log but don't fail - not all agents need prompts
		p.logger.Debug("No agent definition found or error loading",
			zap.String("agent_type", p.agentType),
			zap.Error(err))
	}

	// Store the enriched config in collected data for actions to use
	msgCtx.CollectedData["agent_config"] = agentConfig.CoreLogic

	// Also store the raw input data
	var inputPayload map[string]interface{}
	if err := json.Unmarshal(msgCtx.Message.Value, &inputPayload); err == nil {
		if data, ok := inputPayload["data"]; ok {
			msgCtx.CollectedData["input_data"] = data
		}
	}

	// Validate workflow
	if err := p.validator.ValidateWorkflow(agentConfig.Workflow); err != nil {
		return errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").
			WithCause(err).
			WithDetail("workflow_metrics", p.validator.GetWorkflowMetrics(agentConfig.Workflow)).
			Build()
	}

	// Execute workflow
	return p.executeWorkflow(ctx, msgCtx, agentConfig)
}

func (p *MessageProcessor) executeWorkflow(ctx context.Context, msgCtx *MessageContext, config *models.AgentConfig) error {
	// Start workflow timer
	workflowTimer := observability.StartWorkflowTimer(p.agentType, config.Workflow.StartStep)
	defer workflowTimer.Complete("success")

	// Update metrics
	observability.WorkflowsStarted.WithLabelValues(p.agentType, config.Workflow.StartStep, msgCtx.Headers["client_id"]).Inc()
	observability.ActiveWorkflows.WithLabelValues(p.agentType).Inc()
	defer observability.ActiveWorkflows.WithLabelValues(p.agentType).Dec()

	// Execute through orchestrator
	return p.orchestrator.ExecuteWorkflow(ctx, config.Workflow, msgCtx.Headers, msgCtx.Message.Value)
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
