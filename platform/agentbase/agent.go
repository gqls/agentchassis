// FILE: platform/agentbase/agent.go (refactored)
package agentbase

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/health"
	"github.com/gqls/agentchassis/platform/infrastructure"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/validation"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Agent represents a generic agent chassis
type Agent struct {
	ctx           context.Context
	cfg           *config.ServiceConfig
	logger        *zap.Logger
	agentID       uuid.UUID
	agentType     string
	consumerGroup string

	// Components
	server       *AgentServer
	client       *AgentClient
	healthServer *health.Server

	// Managers
	infraManager *infrastructure.Manager
}

// New creates a new agent with defaults from config
func New(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	agentType := "generic"
	topic := "system.agent.generic.process"

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
	var err error

	// Get or generate agent ID
	// Get agent ID from environment (MUST be set for each agent)
	agentIDStr := os.Getenv("AGENT_ID")
	if agentIDStr == "" {
		logger.Warn("AGENT_ID not set, generated one", zap.String("agent_id is blank", agentIDStr))
		// return nil, fmt.Errorf("AGENT_ID is not set")
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		logger.Warn("AGENT_ID not set, generated one", zap.Any("agent_id is weird", agentID))
		// return nil, fmt.Errorf("invalid AGENT_ID: %w", err)
	}

	// Consumer group
	consumerGroup := fmt.Sprintf("%s-group", agentType)
	if cfg.Custom != nil {
		if cg, ok := cfg.Custom["kafka_consumer_group"].(string); ok {
			consumerGroup = cg
		}
	}

	// Check if we should use agent-specific topics
	useNewTopics := kafka.ShouldUseNewTopics()

	// Determine actual topics to use
	var requestTopic, responseTopic string
	var requestConsumerGroup, responseConsumerGroup string

	if useNewTopics {
		// NEW MODEL: Agent-specific topics
		requestTopic = fmt.Sprintf("system.agent.%s.requests", agentID)
		responseTopic = fmt.Sprintf("system.agent.%s.responses", agentID)
		requestConsumerGroup = fmt.Sprintf("%s-requests", agentID)
		responseConsumerGroup = fmt.Sprintf("%s-responses", agentID)

		logger.Info("Using agent-specific topics",
			zap.Any("agent_id", agentID),
			zap.String("request_topic", requestTopic),
			zap.String("response_topic", responseTopic))

		// Create topics if they don't exist
		topicManager := kafka.NewTopicManager(cfg.Infrastructure.KafkaBrokers, logger)
		if err := topicManager.CreateAgentInstanceTopics(ctx, agentID); err != nil {
			logger.Warn("Failed to create agent topics (may already exist)", zap.Error(err))
		}
	} else {
		// OLD MODEL: Type-based topics (current behavior)
		requestTopic = topic // Use provided topic
		responseTopic = fmt.Sprintf("system.responses.%s", agentType)

		// Consumer groups from config or defaults
		requestConsumerGroup = fmt.Sprintf("%s-group", agentType)
		if cfg.Custom != nil {
			if cg, ok := cfg.Custom["kafka_consumer_group"].(string); ok {
				requestConsumerGroup = cg
			}
		}
		responseConsumerGroup = fmt.Sprintf("%s-responses", requestConsumerGroup)

		logger.Info("Using type-based topics (legacy mode)",
			zap.String("agent_type", agentType),
			zap.String("request_topic", requestTopic),
			zap.String("response_topic", responseTopic))
	}

	logger.Info("Agent topics configured",
		zap.Any("agent_id", agentID),
		zap.String("request_topic", topic),
		zap.String("response_topic", responseTopic),
		zap.Bool("using_new_topics", kafka.ShouldUseNewTopics()))

	// Initialize infrastructure with the determined request topic
	infraManager := infrastructure.NewManager(logger)
	if err := infraManager.Initialize(ctx, cfg, requestTopic, requestConsumerGroup); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	connections := infraManager.GetConnections()

	// Create response consumer
	responseConsumer, err := kafka.NewConsumer(
		cfg.Infrastructure.KafkaBrokers,
		responseTopic,
		responseConsumerGroup,
		logger,
	)

	if err != nil {
		infraManager.Close()
		return nil, fmt.Errorf("failed to create response consumer: %w", err)
	}

	// Create components
	components, err := createComponents(connections, agentType, logger)
	if err != nil {
		infraManager.Close()
		return nil, fmt.Errorf("failed to create components: %w", err)
	}

	// Create server (handles incoming requests)
	server := NewAgentServer(
		ctx,
		logger,
		connections.KafkaConsumer,
		components.messageProcessor,
		consumerGroup,
		agentType,
	)

	// Create client (handles responses)
	client := NewAgentClient(
		ctx,
		logger,
		responseConsumer,
		components.messageProcessor,
		consumerGroup,
		agentType,
	)

	// Create health server
	healthServer := createHealthServer(cfg, connections, agentType, logger)

	// Record metrics
	observability.AgentPoolSize.WithLabelValues(agentType).Inc()

	return &Agent{
		ctx:           ctx,
		cfg:           cfg,
		logger:        logger,
		agentID:       agentID,
		agentType:     agentType,
		consumerGroup: consumerGroup,
		server:        server,
		client:        client,
		healthServer:  healthServer,
		infraManager:  infraManager,
	}, nil
}

// Components holds the processing components
type Components struct {
	messageProcessor *messaging.MessageProcessor
	orchestrator     *orchestration.SagaCoordinator
	validator        *validation.WorkflowValidator
}

func createComponents(connections *infrastructure.Connections, agentType string, logger *zap.Logger) (*Components, error) {
	// Create orchestrator
	connConfig := connections.ClientsDB.Config().ConnConfig.Copy()
	stdDB := stdlib.OpenDB(*connConfig)
	orchestrator := orchestration.NewSagaCoordinator(stdDB, connections.KafkaProducer, logger)

	// Create validator
	validator := validation.NewWorkflowValidator()

	// Create message processor
	messageProcessor := messaging.NewMessageProcessor(
		agentType,
		connections.ClientsDB,
		connections.KafkaProducer,
		orchestrator,
		validator,
		logger,
	)

	return &Components{
		messageProcessor: messageProcessor,
		orchestrator:     orchestrator,
		validator:        validator,
	}, nil
}

func createHealthServer(cfg *config.ServiceConfig, connections *infrastructure.Connections, agentType string, logger *zap.Logger) *health.Server {
	healthServer := health.NewServer(
		agentType,
		health.Config{
			HealthPort:  "8080",
			MetricsPort: "9090",
		},
		health.Checkers{
			"database": func(ctx context.Context) error {
				return connections.ClientsDB.Ping(ctx)
			},
			"kafka": func(ctx context.Context) error {
				// Simplified check - could be enhanced
				return nil
			},
		},
		logger,
	)

	// Add monitoring endpoints
	connConfig := connections.ClientsDB.Config().ConnConfig.Copy()
	stdDB := stdlib.OpenDB(*connConfig)
	monitor := orchestration.NewWorkflowMonitor(stdDB)
	healthServer.AddMonitoringEndpoints(monitor)

	return healthServer
}

// Update Run method:
func (a *Agent) Run() error {
	a.logger.Info("Agent starting", zap.String("type", a.agentType))

	// Start health server
	a.healthServer.Start()

	// Bootstrap with core manager
	if err := a.Bootstrap(); err != nil {
		a.logger.Error("Bootstrap failed", zap.Error(err))
		// Decide if bootstrap failure should be fatal
		// For now, log and continue
	}

	// Create error channel for goroutines
	errCh := make(chan error, 2)

	// Start server in goroutine
	go func() {
		if err := a.server.Run(); err != nil {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Start client in goroutine
	go func() {
		if err := a.client.Run(); err != nil {
			errCh <- fmt.Errorf("client error: %w", err)
		}
	}()

	// Wait for error or context cancellation
	select {
	case err := <-errCh:
		return err
	case <-a.ctx.Done():
		return nil
	}
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown() error {
	a.logger.Info("Agent shutting down")

	// Shutdown both server and client
	if err := a.server.Shutdown(); err != nil {
		a.logger.Error("Error shutting down server", zap.Error(err))
	}

	if err := a.client.Shutdown(); err != nil {
		a.logger.Error("Error shutting down client", zap.Error(err))
	}

	observability.AgentPoolSize.WithLabelValues(a.agentType).Dec()
	return a.infraManager.Close()
}
