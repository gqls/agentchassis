// FILE: platform/agentbase/agent.go
package agentbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/ai-persona-system/pkg/models"
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/gqls/ai-persona-system/platform/database"
	"github.com/gqls/ai-persona-system/platform/kafka"
	"github.com/gqls/ai-persona-system/platform/orchestration"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Agent is the core struct for the "Agent Chassis"
type Agent struct {
	ctx           context.Context
	cfg           *config.ServiceConfig
	logger        *zap.Logger
	clientsDB     *pgxpool.Pool
	kafkaConsumer *kafka.Consumer
	kafkaProducer *kafka.Producer
	orchestrator  *orchestration.SagaCoordinator
	agentType     string
}

// New creates and initializes the agent chassis with defaults from config
func New(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	// Default agent type and topic
	agentType := "generic"
	topic := "system.agent.generic.process"

	// Override from config if available
	if cfg.Custom != nil {
		if at, ok := cfg.Custom["agent_type"].(string); ok {
			agentType = at
		}
		if t, ok := cfg.Custom["topic"].(string); ok {
			topic = t
		}
	}

	return NewWithType(ctx, cfg, logger, agentType, topic)
}

// NewWithType creates an agent with specific type and topic
func NewWithType(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, agentType string, topic string) (*Agent, error) {
	// Initialize database connection
	clientsPool, err := database.NewPostgresConnection(ctx, cfg.Infrastructure.ClientsDatabase, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clients database: %w", err)
	}

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, topic, agentType+"-group", logger)
	if err != nil {
		clientsPool.Close()
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		clientsPool.Close()
		consumer.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Create standard DB handle for orchestrator
	stdDB, err := sql.Open("pgx", clientsPool.Config().ConnString())
	if err != nil {
		clientsPool.Close()
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("failed to get standard DB handle: %w", err)
	}

	// Initialize orchestrator
	sagaCoordinator := orchestration.NewSagaCoordinator(stdDB, producer, logger)

	logger.Info("Agent chassis initialized",
		zap.String("agent_type", agentType),
		zap.String("topic", topic),
	)

	return &Agent{
		ctx:           ctx,
		cfg:           cfg,
		logger:        logger,
		clientsDB:     clientsPool,
		kafkaConsumer: consumer,
		kafkaProducer: producer,
		orchestrator:  sagaCoordinator,
		agentType:     agentType,
	}, nil
}

// Run starts the main Kafka consumer loop
func (a *Agent) Run() error {
	a.logger.Info("Agent running", zap.String("type", a.agentType))

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Agent shutting down")
			return a.cleanup()
		default:
			msg, err := a.kafkaConsumer.FetchMessage(a.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			// Process each message in a goroutine
			go a.handleMessage(msg)
		}
	}
}

// handleMessage processes a single task
func (a *Agent) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)

	clientID := headers["client_id"]
	agentInstanceID := headers["agent_instance_id"]

	if clientID == "" || agentInstanceID == "" {
		a.logger.Error("Message missing required headers",
			zap.String("topic", msg.Topic),
			zap.Int64("offset", msg.Offset))
		a.kafkaConsumer.CommitMessages(context.Background(), msg)
		return
	}

	l := a.logger.With(
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("request_id", headers["request_id"]),
		zap.String("client_id", clientID),
		zap.String("agent_instance_id", agentInstanceID),
	)

	// Load agent configuration
	agentConfig, err := a.loadAgentConfig(clientID, agentInstanceID)
	if err != nil {
		l.Error("Failed to load agent configuration", zap.Error(err))
		a.sendErrorResponse(headers, fmt.Sprintf("Failed to load configuration: %v", err))
		a.kafkaConsumer.CommitMessages(context.Background(), msg)
		return
	}

	l.Info("Agent instance loaded", zap.String("agent_type", agentConfig.AgentType))

	// Execute workflow
	if err := a.orchestrator.ExecuteWorkflow(a.ctx, agentConfig.Workflow, headers, msg.Value); err != nil {
		l.Error("Workflow execution failed", zap.Error(err))
		a.sendErrorResponse(headers, fmt.Sprintf("Workflow execution failed: %v", err))
	}

	// Commit message
	if err := a.kafkaConsumer.CommitMessages(context.Background(), msg); err != nil {
		l.Error("Failed to commit kafka message", zap.Error(err))
	}
}

// loadAgentConfig fetches the agent's configuration from the database
func (a *Agent) loadAgentConfig(clientID, agentInstanceID string) (*models.AgentConfig, error) {
	ctx := context.Background()

	// Query the agent instance configuration
	query := fmt.Sprintf(`
		SELECT name, config, template_id 
		FROM client_%s.agent_instances 
		WHERE id = $1 AND is_active = true
	`, clientID)

	var name string
	var configJSON []byte
	var templateID string

	err := a.clientsDB.QueryRow(ctx, query, agentInstanceID).Scan(&name, &configJSON, &templateID)
	if err != nil {
		// If not found in database, return a default configuration
		if err.Error() == "no rows in result set" {
			a.logger.Warn("Agent instance not found, using default configuration",
				zap.String("agent_instance_id", agentInstanceID))
			return a.getDefaultConfig(agentInstanceID), nil
		}
		return nil, fmt.Errorf("failed to query agent instance: %w", err)
	}

	// Parse the configuration
	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	// Extract workflow if present, otherwise use default
	var workflow models.WorkflowPlan
	if workflowData, ok := config["workflow"]; ok {
		workflowBytes, _ := json.Marshal(workflowData)
		if err := json.Unmarshal(workflowBytes, &workflow); err != nil {
			a.logger.Warn("Failed to parse workflow, using default", zap.Error(err))
			workflow = a.getDefaultWorkflow()
		}
	} else {
		workflow = a.getDefaultWorkflow()
	}

	return &models.AgentConfig{
		AgentID:   agentInstanceID,
		AgentType: a.agentType,
		Version:   1,
		CoreLogic: config,
		Workflow:  workflow,
	}, nil
}

// getDefaultConfig returns a default configuration for agents
func (a *Agent) getDefaultConfig(agentInstanceID string) *models.AgentConfig {
	return &models.AgentConfig{
		AgentID:   agentInstanceID,
		AgentType: a.agentType,
		Version:   1,
		CoreLogic: map[string]interface{}{
			"model":       "claude-3-opus",
			"temperature": 0.7,
		},
		Workflow: a.getDefaultWorkflow(),
	}
}

// getDefaultWorkflow returns a simple default workflow
func (a *Agent) getDefaultWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "generate",
		Steps: map[string]models.Step{
			"generate": {
				Action:      "ai_text_generate",
				Description: "Generate text using AI",
				NextStep:    "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Mark workflow as complete",
			},
		},
	}
}

// sendErrorResponse sends an error response back via Kafka
func (a *Agent) sendErrorResponse(headers map[string]string, errorMsg string) {
	responseHeaders := a.createResponseHeaders(headers)

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
		"agent":   a.agentType,
	}

	responseBytes, _ := json.Marshal(errorResponse)

	// Send to error topic
	errorTopic := fmt.Sprintf("system.errors.%s", a.agentType)
	if err := a.kafkaProducer.Produce(a.ctx, errorTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to send error response", zap.Error(err))
	}
}

// createResponseHeaders creates response headers with proper causality tracking
func (a *Agent) createResponseHeaders(originalHeaders map[string]string) map[string]string {
	return map[string]string{
		"correlation_id": originalHeaders["correlation_id"],
		"causation_id":   originalHeaders["request_id"],
		"request_id":     uuid.NewString(),
		"client_id":      originalHeaders["client_id"],
		"agent_type":     a.agentType,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

// cleanup gracefully shuts down all resources
func (a *Agent) cleanup() error {
	var errs []error

	if err := a.kafkaConsumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close kafka consumer: %w", err))
	}

	if err := a.kafkaProducer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close kafka producer: %w", err))
	}

	a.clientsDB.Close()

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// GetType returns the agent type
func (a *Agent) GetType() string {
	return a.agentType
}

// GetConfig returns the service configuration
func (a *Agent) GetConfig() *config.ServiceConfig {
	return a.cfg
}
