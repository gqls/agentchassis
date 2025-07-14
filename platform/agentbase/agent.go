// FILE: platform/agentbase/agent.go (with fixed imports and types)
package agentbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/database"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/memory"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/validation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Agent is the core struct for the "Agent Chassis"
type Agent struct {
	ctx           context.Context
	cfg           *config.ServiceConfig
	logger        *zap.Logger
	clientsDB     *pgxpool.Pool
	kafkaConsumer *kafka.Consumer
	kafkaProducer kafka.Producer
	orchestrator  *orchestration.SagaCoordinator
	memoryService *memory.Service
	validator     *validation.WorkflowValidator
	metricsServer *observability.MetricsServer
	agentType     string
	consumerGroup string
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

	// Consumer group name
	consumerGroup := fmt.Sprintf("%s-group", agentType)
	if cfg.Custom != nil {
		if cg, ok := cfg.Custom["kafka_consumer_group"].(string); ok {
			consumerGroup = cg
		}
	}

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, topic, consumerGroup, logger)
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

	// Create standard DB handle for orchestrator using pgx v5 stdlib
	connConfig := clientsPool.Config().ConnConfig.Copy()
	stdDB := stdlib.OpenDB(*connConfig)

	// Initialize orchestrator
	sagaCoordinator := orchestration.NewSagaCoordinator(stdDB, producer, logger)

	// Initialize memory service (placeholder for now - needs AI client)
	// In real implementation, you'd initialize the AI client based on config
	var memoryService *memory.Service
	// memoryService = memory.NewService(clientsPool, aiClient, logger)

	// Initialize workflow validator
	validator := validation.NewWorkflowValidator()

	// Initialize metrics server
	metricsServer := observability.NewMetricsServer("9090")

	logger.Info("Agent chassis initialized",
		zap.String("agent_type", agentType),
		zap.String("topic", topic),
		zap.String("consumer_group", consumerGroup),
	)

	// Record agent pool size
	observability.AgentPoolSize.WithLabelValues(agentType).Inc()

	return &Agent{
		ctx:           ctx,
		cfg:           cfg,
		logger:        logger,
		clientsDB:     clientsPool,
		kafkaConsumer: consumer,
		kafkaProducer: producer,
		orchestrator:  sagaCoordinator,
		memoryService: memoryService,
		validator:     validator,
		metricsServer: metricsServer,
		agentType:     agentType,
		consumerGroup: consumerGroup,
	}, nil
}

// Run starts the main Kafka consumer loop
func (a *Agent) Run() error {
	a.logger.Info("Agent running", zap.String("type", a.agentType))

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Agent shutting down")
			observability.AgentPoolSize.WithLabelValues(a.agentType).Dec()
			return a.cleanup()
		default:
			msg, err := a.kafkaConsumer.FetchMessage(a.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				observability.SystemErrors.WithLabelValues(a.agentType, "fetch_message").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			// Record message consumed
			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, a.consumerGroup).Inc()

			// Process each message in a goroutine
			go a.handleMessage(msg)
		}
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
