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

	// set up topics when agent is created
	requestTopic  string
	responseTopic string
}

// New creates a new agent with defaults from config
func New(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	agentType := "generic"
	topic := "system.agent.generic.requests"

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
	// Get agent ID from environment
	agentIDStr := os.Getenv("AGENT_ID")
	if agentIDStr == "" {
		logger.Warn("AGENT_ID not set, generating one")
		// Generate a new ID if not provided
		agentIDStr = uuid.New().String()
	}

	// Parse the agent ID
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		logger.Error("Invalid AGENT_ID format",
			zap.String("agent_id_str", agentIDStr),
			zap.Error(err))
		return nil, fmt.Errorf("invalid AGENT_ID: %w", err)
	}

	var requestTopic, responseTopic string
	var requestConsumerGroup, responseConsumerGroup string

	// Topics are ALWAYS deterministic based on agent type
	requestTopic = fmt.Sprintf("system.agent.%s.requests", agentType)
	responseTopic = fmt.Sprintf("system.agent.%s.responses", agentType)

	// Create topics if they don't exist
	topicManager := kafka.NewTopicManager(cfg.Infrastructure.KafkaBrokers, logger)

	// Use the existing CreateAgentTypeTopics method which creates both request and response topics
	if err := topicManager.CreateAgentTypeTopics(ctx, agentType); err != nil {
		logger.Warn("Failed to ensure topics exist",
			zap.String("agent_type", agentType),
			zap.Error(err))
		// Don't fail - topics might already exist
	}

	logger.Info("Agent topics configured",
		zap.String("agent_type", agentType),
		zap.String("request_topic", requestTopic),
		zap.String("response_topic", responseTopic))

	// Always set topics based on agent type, including for "generic"
	if agentType != "" {
		// Use type-based topics with new naming convention
		requestTopic = fmt.Sprintf("system.agent.%s.requests", agentType)
		responseTopic = fmt.Sprintf("system.agent.%s.responses", agentType)
		requestConsumerGroup = fmt.Sprintf("%s-requests-group", agentType)
		responseConsumerGroup = fmt.Sprintf("%s-responses-group", agentType)

		logger.Info("Using type-based topics",
			zap.String("agent_id", agentID.String()),
			zap.String("agent_type", agentType),
			zap.String("request_topic", requestTopic),
			zap.String("response_topic", responseTopic))

		// Create topics if they don't exist
		topicManager := kafka.NewTopicManager(cfg.Infrastructure.KafkaBrokers, logger)
		if err := topicManager.CreateAgentTypeTopics(ctx, agentType); err != nil {
			logger.Warn("Failed to create agent topics (may already exist)",
				zap.Error(err),
				zap.String("agent_type", agentType))
		}
	} else {
		// This should never happen, but have a fallback
		return nil, fmt.Errorf("agent type cannot be empty")
	}

	// Special handling for orchestrator-type agents that need to listen to multiple topics
	if agentID.String() == "00000000-0000-0000-0000-000000000001" || agentType == "orchestrator" || agentType == "website-builder" {
		// These agents also need to listen to orchestrator responses
		// But the primary topics are still based on their type
		logger.Info("Orchestrator-type agent detected, will listen to additional response topics",
			zap.String("agent_type", agentType))
	}

	logger.Info("Final agent configuration",
		zap.String("agent_id", agentID.String()),
		zap.String("agent_type", agentType),
		zap.String("request_topic", requestTopic),
		zap.String("response_topic", responseTopic),
		zap.String("request_consumer_group", requestConsumerGroup),
		zap.String("response_consumer_group", responseConsumerGroup))

	// Initialize infrastructure with the request topic
	infraManager := infrastructure.NewManager(logger)
	if err := infraManager.Initialize(ctx, cfg, requestTopic, requestConsumerGroup); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	connections := infraManager.GetConnections()

	// Create response consumer for listening to responses from child agents
	responseConsumer, err := kafka.NewConsumer(
		cfg.Infrastructure.KafkaBrokers,
		responseTopic,
		responseConsumerGroup,
		logger,
	)
	if err != nil {
		infraManager.Close()
		return nil, fmt.Errorf("failed to create response consumer for topic %s: %w", responseTopic, err)
	}

	// Create components
	components, err := createComponents(connections, agentType, logger)
	if err != nil {
		infraManager.Close()
		responseConsumer.Close()
		return nil, fmt.Errorf("failed to create components: %w", err)
	}

	// Create server (handles incoming requests on request topic)
	server := NewAgentServer(
		ctx,
		logger,
		connections.KafkaConsumer, // This consumer is already set up for requestTopic
		components.messageProcessor,
		requestConsumerGroup,
		agentType,
	)

	// Create client (handles responses from children on response topic)
	client := NewAgentClient(
		ctx,
		logger,
		responseConsumer, // This consumer listens to responseTopic
		components.messageProcessor,
		responseConsumerGroup,
		agentType,
	)

	// Create health server
	healthServer := createHealthServer(cfg, connections, agentType, logger)

	// Record metrics
	observability.AgentPoolSize.WithLabelValues(agentType).Inc()

	logger.Info("Agent successfully initialized",
		zap.String("agent_id", agentID.String()),
		zap.String("agent_type", agentType),
		zap.String("listening_for_requests_on", requestTopic),
		zap.String("listening_for_responses_on", responseTopic))

	return &Agent{
		ctx:           ctx,
		cfg:           cfg,
		logger:        logger,
		agentID:       agentID,
		agentType:     agentType,
		consumerGroup: requestConsumerGroup,
		server:        server,
		client:        client,
		healthServer:  healthServer,
		infraManager:  infraManager,
		requestTopic:  requestTopic,
		responseTopic: responseTopic,
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
