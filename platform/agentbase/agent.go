// FILE: platform/agentbase/agent.go
package agentbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/validation"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// AgentConfig contains configuration for creating an agent
type AgentConfig struct {
	// Basic configuration
	AgentType    string
	AgentName    string // Optional, will be generated if empty
	AgentVersion string
	Role         string // Optional functional role

	// Parent info (for spawned agents)
	ParentAgentID         string
	ParentAgentType       string
	ParentOrchestrationID string

	// System configuration
	EnableStateless bool
	KafkaBrokers    []string
	DatabaseURL     string
	LogLevel        string

	// Behavior configuration
	MaxRetries   int
	RetryBackoff time.Duration

	// Dynamic configuration (for runtime customization)
	DynamicConfig map[string]interface{}
}

// Agent represents both statically configured and dynamically spawned agents
type Agent struct {
	// Identity
	AgentID      string // Unique identifier
	AgentName    string // Human-readable name
	AgentType    string // Type of agent
	AgentVersion string
	Role         string // Functional role in a group/team

	// Pod identity (for stateless operation)
	PodName   string // Kubernetes pod name
	NodeName  string // Kubernetes node
	Namespace string // Kubernetes namespace

	// Parent relationship (for spawned agents)
	ParentAgentID         string
	ParentAgentType       string
	ParentOrchestrationID string

	// Configuration
	config        *AgentConfig
	DynamicConfig map[string]interface{}
	isStateless   bool
	spawned       bool      // Whether agent was dynamically spawned
	spawnTime     time.Time // When agent was spawned
	initialized   bool      // Whether agent is initialized

	// Core components
	db           *sql.DB
	producer     kafka.Producer
	orchestrator *orchestration.SagaCoordinator
	processor    *messaging.MessageProcessor
	validator    *validation.Validator
	logger       *zap.Logger

	// Kafka consumers - ALL agents have both
	requestConsumer  *kafka.Consumer
	responseConsumer *kafka.Consumer

	// Topics
	requestsTopic  string
	responsesTopic string

	// State management
	stateRepo *orchestration.StateRepository

	// Lifecycle
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	shutdownChan chan struct{}

	// Metrics
	messagesProcessed uint64
	lastActivity      time.Time
}

// NewAgent creates a new agent with the standard constructor signature
func New(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	// Extract agent type from config or environment
	agentType := ""
	if cfg.Custom != nil {
		if at, ok := cfg.Custom["agent_type"].(string); ok {
			agentType = at
		}
	}
	if agentType == "" {
		agentType = os.Getenv("AGENT_TYPE")
	}
	if agentType == "" {
		agentType = "generic" // fallback
	}

	// Generate agent ID
	agentID := os.Getenv("AGENT_ID")
	if agentID == "" {
		// Generate a stable UUID based on hostname if AGENT_ID not set
		agentID = "00000000-0000-0000-0000-000000000002"
	}

	// Generate agent name
	agentName := fmt.Sprintf("%s-%s", agentType, time.Now().Format("0102-1504"))

	// Build database URL
	var databaseURL string
	if cfg.Infrastructure.ClientsDatabase.Host != "" {
		dbPassword := os.Getenv(cfg.Infrastructure.ClientsDatabase.PasswordEnvVar)
		databaseURL = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Infrastructure.ClientsDatabase.Host,
			cfg.Infrastructure.ClientsDatabase.Port,
			cfg.Infrastructure.ClientsDatabase.User,
			dbPassword,
			cfg.Infrastructure.ClientsDatabase.DBName,
			cfg.Infrastructure.ClientsDatabase.SSLMode)
	}

	// Extract version
	agentVersion := os.Getenv("AGENT_VERSION")
	if agentVersion == "" {
		agentVersion = "1.0.0"
	}

	// Create child context with cancel
	agentCtx, cancel := context.WithCancel(ctx)

	// Enhance logger with agent context
	agentLogger := logger.With(
		zap.String("agent_type", agentType),
		zap.String("agent_id", agentID),
		zap.String("pod_name", os.Getenv("HOSTNAME")),
		zap.Bool("stateless", true))

	agent := &Agent{
		// Identity
		AgentID:      agentID,
		AgentName:    agentName,
		AgentType:    agentType,
		AgentVersion: agentVersion,

		// Pod identity
		PodName:   os.Getenv("HOSTNAME"),
		NodeName:  os.Getenv("NODE_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),

		// Configuration
		config: &AgentConfig{
			AgentType:       agentType,
			AgentName:       agentName,
			AgentVersion:    agentVersion,
			EnableStateless: true,
			KafkaBrokers:    cfg.Infrastructure.KafkaBrokers,
			DatabaseURL:     databaseURL,
			MaxRetries:      3,
			RetryBackoff:    5 * time.Second,
			DynamicConfig:   cfg.Custom,
		},
		DynamicConfig: cfg.Custom,
		isStateless:   true,

		// Core components
		logger:       agentLogger,
		ctx:          agentCtx,
		cancel:       cancel,
		shutdownChan: make(chan struct{}),

		// Topics
		requestsTopic:  fmt.Sprintf("system.agent.%s.requests", agentType),
		responsesTopic: fmt.Sprintf("system.agent.%s.responses", agentType),

		// Metrics
		lastActivity: time.Now(),
	}

	// Initialize components
	if err := agent.initializeComponents(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	agent.initialized = true

	agentLogger.Info("Agent created",
		zap.String("agent_id", agent.AgentID),
		zap.String("requests_topic", agent.requestsTopic),
		zap.String("responses_topic", agent.responsesTopic))

	return agent, nil
}

// NewAgent creates a new agent (static or dynamic)
func NewAgent(config AgentConfig) (*Agent, error) {
	// Generate agent ID if not provided
	agentID := os.Getenv("AGENT_ID")
	if agentID == "" {
		// Generate a stable UUID based on hostname if AGENT_ID not set
		agentID = "00000000-0000-0000-0000-000000000003"
	}

	// Generate agent name if not provided
	agentName := config.AgentName
	if agentName == "" {
		agentName = fmt.Sprintf("%s-%s", config.AgentType, time.Now().Format("0102-1504"))
	}

	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	logger = logger.With(
		zap.String("agent_type", config.AgentType),
		zap.String("agent_id", agentID),
		zap.String("pod_name", os.Getenv("HOSTNAME")),
		zap.Bool("stateless", config.EnableStateless))

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		// Identity
		AgentID:      agentID,
		AgentName:    agentName,
		AgentType:    config.AgentType,
		AgentVersion: config.AgentVersion,
		Role:         config.Role,

		// Pod identity
		PodName:   os.Getenv("HOSTNAME"),
		NodeName:  os.Getenv("NODE_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),

		// Parent relationship
		ParentAgentID:         config.ParentAgentID,
		ParentAgentType:       config.ParentAgentType,
		ParentOrchestrationID: config.ParentOrchestrationID,

		// Configuration
		config:        &config,
		DynamicConfig: config.DynamicConfig,
		isStateless:   config.EnableStateless,
		spawned:       config.ParentAgentID != "", // If has parent, it was spawned

		// Core components
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		shutdownChan: make(chan struct{}),

		// Topics
		requestsTopic:  fmt.Sprintf("system.agent.%s.requests", config.AgentType),
		responsesTopic: fmt.Sprintf("system.agent.%s.responses", config.AgentType),

		// Metrics
		lastActivity: time.Now(),
	}

	// Initialize components
	if err := agent.initializeComponents(); err != nil {
		cancel()
		return nil, err
	}

	agent.initialized = true

	logger.Info("Agent created",
		zap.String("agent_id", agent.AgentID),
		zap.String("requests_topic", agent.requestsTopic),
		zap.String("responses_topic", agent.responsesTopic),
		zap.Bool("spawned", agent.spawned))

	return agent, nil
}

// initializeComponents initializes all agent components
func (a *Agent) initializeComponents() error {
	// Initialize database if URL provided
	if a.config.DatabaseURL != "" {
		db, err := sql.Open("pgx", a.config.DatabaseURL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		a.db = db
		a.stateRepo = orchestration.NewStateRepository(db, a.logger)
	}

	// Create Kafka producer
	producer, err := kafka.NewProducer(a.config.KafkaBrokers, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	a.producer = producer

	// Create validator
	a.validator = validation.NewValidator(a.logger)

	// EVERY agent gets an orchestrator - they all orchestrate something
	if a.db != nil {
		a.orchestrator = orchestration.NewSagaCoordinator(a.db, producer, a.logger)
	} else {
		a.logger.Warn("No database configured, orchestration capabilities will be limited")
		// Could create an in-memory orchestrator here if needed
	}

	// Create message processor
	a.processor = messaging.NewMessageProcessor(
		a.AgentType,
		a.AgentID,
		a.Role,
		a.db,
		producer,
		a.orchestrator,
		a.validator,
		a.logger,
		a,
	)

	// Set up consumers
	if err := a.setupConsumers(); err != nil {
		return fmt.Errorf("failed to setup consumers: %w", err)
	}

	return nil
}

// setupConsumers sets up Kafka consumers for the agent
func (a *Agent) setupConsumers() error {

	a.logger.Info("setupConsumers in agent.go")

	// Consumer group naming for stateless operation
	requestConsumerGroup := fmt.Sprintf("%s-request-consumers", a.AgentType)
	responseConsumerGroup := fmt.Sprintf("%s-response-consumers", a.AgentType)

	// Create request consumer (ALL agents need this)
	requestConsumer, err := kafka.NewConsumer(
		a.config.KafkaBrokers,
		a.requestsTopic,
		requestConsumerGroup,
		a.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create request consumer: %w", err)
	}
	a.requestConsumer = requestConsumer

	// Create response consumer (ALL agents need this since they all orchestrate)
	responseConsumer, err := kafka.NewConsumer(
		a.config.KafkaBrokers,
		a.responsesTopic,
		responseConsumerGroup,
		a.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create response consumer: %w", err)
	}
	a.responseConsumer = responseConsumer

	return nil
}

// Run starts the agent's message processing loops
func (a *Agent) Run() error {
	if !a.initialized {
		return fmt.Errorf("agent not initialized")
	}

	a.logger.Info("Agent starting agent.go Run",
		zap.String("agent_id", a.AgentID),
		zap.String("agent_type", a.AgentType),
		zap.String("role", a.Role))

	// Start request processing
	a.wg.Add(1)
	go a.processRequests()

	// Start response processing (ALL agents process responses)
	a.wg.Add(1)
	go a.processResponses()

	// Start health check
	a.wg.Add(1)
	go a.healthCheck()

	// If spawned, send heartbeats to parent
	if a.spawned && a.ParentAgentID != "" {
		a.wg.Add(1)
		go a.sendHeartbeats()
	}

	// Wait for shutdown signal
	<-a.shutdownChan

	a.logger.Info("Agent shutting down")
	a.cancel()
	a.wg.Wait()

	return nil
}

// processRequests handles incoming request messages
func (a *Agent) processRequests() {
	defer a.wg.Done()

	a.logger.Info("Starting request processor agent.go processRequests",
		zap.String("topic", a.requestsTopic))

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			msg, err := a.requestConsumer.Consume(a.ctx)
			a.logger.Info("DEBUG uuidparse: agent.go processRequests the message was:",
				zap.Any("message", msg))
			if err != nil {
				if err != context.Canceled {
					a.logger.Error("Failed to consume request message", zap.Error(err))
				}
				continue
			}
			// Process the message
			a.processMessage(msg, "request")
		}
	}
}

// processResponses handles incoming response messages
func (a *Agent) processResponses() {
	defer a.wg.Done()

	a.logger.Info("Starting response processor agent.go processResponses",
		zap.String("topic", a.responsesTopic))

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			msg, err := a.responseConsumer.Consume(a.ctx)
			if err != nil {
				if err != context.Canceled {
					a.logger.Error("Failed to consume response message", zap.Error(err))
				}
				continue
			} else {
				a.logger.Info("Response consumer received message",
					zap.Int("size", len(msg.Value)))
			}
			a.processMessage(msg, "response")
		}
	}
}

// processMessage handles a single message (request or response)
func (a *Agent) processMessage(msg kafka.Message, messageType string) {

	a.logger.Info("process single message agent.go processMessage",
		zap.String("messageType", messageType))

	startTime := time.Now()
	a.lastActivity = startTime
	a.messagesProcessed++

	// Extract headers
	headers := kafka.HeadersToMap(msg.Headers)

	// Create or parse ExecutionContext
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Create minimal context for legacy messages
		execCtx = &types.ExecutionContext{
			MessageID:      uuid.New().String(),
			MessageType:    messageType,
			Timestamp:      time.Now(),
			ProcessingNode: a.PodName,
		}

		// Extract basic fields from headers
		if corrID := headers["correlation_id"]; corrID != "" {
			execCtx.CorrelationID = corrID
		}
		if orchID := headers["orchestration_id"]; orchID != "" {
			execCtx.OrchestrationID = orchID
		}
		if reqID := headers["request_id"]; reqID != "" {
			execCtx.RequestID = reqID
		}
	}

	// Add agent context
	execCtx.ProcessingNode = a.PodName
	execCtx.ToAgentID = a.AgentID
	execCtx.ToAgentType = a.AgentType

	contextLogger := a.logger.With(execCtx.LogContext()...)

	contextLogger.Info("Processing message STORE_EXEC_CONTEXT ish",
		zap.String("message_type", messageType),
		zap.String("message_id", execCtx.MessageID),
		zap.String("stored_responses_topic", execCtx.ResponsesTopic),
		zap.String("stored_request_id", execCtx.RequestID),
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("parent_orch_id", execCtx.ParentOrchestrationID),
		zap.Int("payload_size", len(msg.Value)))

	// Check for duplicates if stateless
	if a.isStateless && a.stateRepo != nil {
		isDuplicate, err := a.stateRepo.HasProcessedMessage(a.ctx, execCtx.CorrelationID, execCtx.RequestID, a.AgentID)
		if err != nil {
			contextLogger.Error("Failed to check for duplicate", zap.Error(err))
		} else if isDuplicate {
			contextLogger.Info("Duplicate message ignored")
			observability.MessagesDropped.WithLabelValues(a.AgentType, "duplicate").Inc()
			return
		}

		// Record processing
		if err := a.stateRepo.RecordMessageProcessing(a.ctx, execCtx, a.AgentID); err != nil {
			contextLogger.Error("Failed to record message processing", zap.Error(err))
		}
	}

	// Update performance metrics if spawned
	if a.spawned {
		defer a.updatePerformanceMetrics(err == nil)
	}

	// Process through message processor
	if err := a.processor.ProcessMessage(a.ctx, msg); err != nil {
		contextLogger.Error("Failed to process message (agent.go)",
			zap.Error(err),
			zap.Duration("duration", time.Since(startTime)))

		observability.MessagesFailed.WithLabelValues(a.AgentType, messageType).Inc()
		a.handleProcessingError(execCtx, err)
	} else {
		contextLogger.Info("Message processed successfully (agent.go)",
			zap.Duration("duration", time.Since(startTime)))

		observability.MessagesProcessed.WithLabelValues(a.AgentType, messageType).Inc()
		observability.MessageProcessingDuration.WithLabelValues(a.AgentType, messageType).Observe(time.Since(startTime).Seconds())
	}
}

// handleProcessingError handles errors during message processing
func (a *Agent) handleProcessingError(execCtx *types.ExecutionContext, err error) {
	// Determine if error is recoverable
	recoverable := isRecoverableError(err)

	// Send error response if we have enough context
	if execCtx.OrchestrationID != "" && execCtx.RequestID != "" {
		errorResponse := &types.ResponseMessage{
			Headers: types.ResponseHeaders{
				// Sender identity
				Sender: types.AgentIdentity{
					AgentType:    a.AgentType,
					AgentID:      a.AgentID,
					PodName:      a.PodName,
					AgentVersion: a.AgentVersion,
					Role:         a.Role,
				},

				// Response tracking - what we're responding to
				InResponseToRequestID:      execCtx.RequestID,
				InResponseToStepID:         execCtx.StepID,
				InResponseToStepName:       execCtx.StepName,
				InResponseToParentOrchID:   execCtx.ParentOrchestrationID,
				InResponseToParentOrchName: execCtx.ParentOrchestrationName,
				InResponseToMessageID:      execCtx.MessageID,
				InResponseToAction:         execCtx.Action,
				RetryCount:                 execCtx.RetryVersion,

				// My context (the error handler)
				MyOrchestrationID:   execCtx.OrchestrationID,
				MyOrchestrationName: execCtx.OrchestrationName,
				MyRequestsTopic:     a.requestsTopic,
				MyResponsesTopic:    a.responsesTopic,

				// Identity
				CorrelationID:   execCtx.CorrelationID,
				CorrelationName: execCtx.CorrelationName,
				ClientID:        execCtx.ClientID,
				MessageType:     "response",
				FromAgent:       a.AgentID,
				ToAgent:         execCtx.FromAgentID,
				ToAgentType:     execCtx.FromAgentType,

				// Status flags
				IsComplete:          false, // Error is not a successful completion
				IsError:             true,
				IsMultipartResponse: false,
				PartCount:           1,
				Status:              getErrorStatus(recoverable),

				// Timing & Resources
				TimeSent:                   time.Now(),
				TimeSpent:                  time.Since(execCtx.Timestamp),
				OverallTimeBudgetRemaining: execCtx.TimeoutSeconds - int(time.Since(execCtx.Timestamp).Seconds()),
				TopicSentTo:                "", // Will be set by sendErrorResponse
				FuelUsed:                   10, // Some nominal fuel for error handling
				RemainingFuelBudget:        execCtx.FuelBudget - 10,
			},
			Body: types.ResponseBody{
				Success: false,
				Headers: nil,
				Body:    nil, // No result for error
				Error: &types.ErrorInfo{
					Code:        "PROCESSING_ERROR",
					Message:     err.Error(),
					Recoverable: recoverable,
					RetryAfter:  30, // seconds
				},
			},
		}

		// Send error response
		a.sendErrorResponse(execCtx, errorResponse)
	}
}

// sendErrorResponse sends an error response
func (a *Agent) sendErrorResponse(execCtx *types.ExecutionContext, response *types.ResponseMessage) {
	// Determine response topic
	var responsesTopic string
	if execCtx.ResponsesTopic != "" {
		responsesTopic = execCtx.ResponsesTopic
	} else if execCtx.FromAgentType != "" {
		responsesTopic = fmt.Sprintf("system.agent.%s.responses", execCtx.FromAgentType)
	} else {
		responsesTopic = "system.agent.generic.errors"
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		a.logger.Error("Failed to marshal error response", zap.Error(err))
		return
	}

	// Convert headers to map
	headers := response.Headers.ToMap()

	// Use correlation ID as key
	key := []byte(execCtx.CorrelationID)
	if execCtx.CorrelationID == "" {
		key = []byte(execCtx.MessageID)
	}

	// Use YOUR producer's signature
	if err := a.producer.Produce(a.ctx, responsesTopic, headers, key, responseBytes); err != nil {
		a.logger.Error("Failed to send error response",
			zap.Error(err),
			zap.String("topic", responsesTopic))
	}
}

// healthCheck performs periodic health checks
func (a *Agent) healthCheck() {
	defer a.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return

		case <-ticker.C:
			// Check agent health
			isHealthy := a.checkHealth()

			// Update metrics
			if isHealthy {
				observability.AgentHealth.WithLabelValues(a.AgentType, a.PodName).Set(1)
			} else {
				observability.AgentHealth.WithLabelValues(a.AgentType, a.PodName).Set(0)
			}

			// Log health status
			a.logger.Debug("Health check",
				zap.Bool("healthy", isHealthy),
				zap.Uint64("messages_processed", a.messagesProcessed),
				zap.Duration("since_last_activity", time.Since(a.lastActivity)))
		}
	}
}

// checkHealth checks if the agent is healthy
func (a *Agent) checkHealth() bool {
	// Check database connection
	if a.db != nil {
		if err := a.db.Ping(); err != nil {
			a.logger.Error("Database ping failed", zap.Error(err))
			return false
		}
	}

	// Check if we're processing messages
	if time.Since(a.lastActivity) > 5*time.Minute {
		a.logger.Warn("No activity for 5 minutes")
		// This might be normal if no messages are coming in
	}

	return true
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown() error {
	a.logger.Info("Agent shutdown initiated")

	// Signal shutdown
	close(a.shutdownChan)

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("Agent shutdown complete")
	case <-time.After(30 * time.Second):
		a.logger.Warn("Agent shutdown timeout")
	}

	// Close resources
	if a.requestConsumer != nil {
		a.requestConsumer.Close()
	}
	if a.responseConsumer != nil {
		a.responseConsumer.Close()
	}
	if a.producer != nil {
		a.producer.Close()
	}
	if a.db != nil {
		a.db.Close()
	}

	return nil
}

// GetIdentity returns the agent's identity for message headers
func (a *Agent) GetIdentity() types.AgentIdentity {
	return types.AgentIdentity{
		AgentType:    a.AgentType,
		AgentID:      a.AgentID,
		PodName:      a.PodName,
		AgentVersion: a.AgentVersion,
		Role:         a.Role,
	}
}

// SpawnChildAgent spawns a child agent and tracks it in the hierarchy
func (a *Agent) SpawnChildAgent(ctx context.Context, agentType, role string, config map[string]interface{}) (string, error) {
	// Generate child agent ID
	childAgentID := uuid.New().String()

	// Generate IDs for the spawn request
	correlationID := uuid.New().String()
	orchestrationID := uuid.New().String()
	requestID := uuid.New().String()
	messageID := uuid.New().String()

	// Extract client ID from context or config if available
	clientID := ""
	if configClientID, ok := config["client_id"].(string); ok {
		clientID = configClientID
	} else if ctxClientID := ctx.Value("client_id"); ctxClientID != nil {
		// Could also extract from context if you store it there
		clientID = ctxClientID.(string)
	}

	// Create properly populated request headers
	spawnRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			// Sender identity
			Sender: a.GetIdentity(),

			// Core identity
			CorrelationID:   correlationID,
			CorrelationName: fmt.Sprintf("spawn-%s-%s", agentType, childAgentID[:8]),
			ClientID:        clientID,
			FunctionalRole:  role,

			// Orchestration context
			OrchestrationID:   orchestrationID,
			OrchestrationName: fmt.Sprintf("%s-spawn-child", a.AgentType),
			StepID:            uuid.New().String(),
			StepName:          "spawn_child_agent",
			RequestID:         requestID,
			RetryVersion:      0,
			// Parent orchestration would be a's current orchestration if it has one
			ParentOrchestrationID:   "", // Set if a is in an orchestration
			ParentOrchestrationName: "",
			ParentRequestID:         "",

			// Message metadata
			MessageID:   messageID,
			MessageType: "request",
			FromAgent:   a.AgentID,
			ToAgent:     childAgentID, // The child we're spawning
			ToAgentType: agentType,
			Action:      "initialize",
			Timestamp:   time.Now(),

			// Resource management
			FuelBudget:     1000, // Or inherit from parent
			TimeoutSeconds: 30,

			// Routing
			RequestsTopic:  fmt.Sprintf("system.agent.%s.requests", agentType),
			ResponsesTopic: fmt.Sprintf("system.agent.%s.responses", a.AgentType),
		},
		Body: map[string]interface{}{
			"agent_id": childAgentID,
			"role":     role,
			"config":   config,
			"parent": map[string]interface{}{
				"agent_id":          a.AgentID,
				"agent_type":        a.AgentType,
				"orchestration_id":  orchestrationID,
				"correlation_id":    correlationID,
				"parent_request_id": requestID,
			},
		},
	}

	// Send spawn request
	topic := fmt.Sprintf("system.agent.%s.requests", agentType)
	spawnBytes, err := json.Marshal(spawnRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spawn request: %w", err)
	}

	// Convert headers to map
	headers := spawnRequest.Headers.ToMap()

	// Use correlation ID as the message key for Kafka partitioning
	key := []byte(correlationID)

	// Use the correct producer.Produce signature: (ctx, topic, headers, key, value)
	if err := a.producer.Produce(ctx, topic, headers, key, spawnBytes); err != nil {
		return "", fmt.Errorf("failed to spawn child agent: %w", err)
	}

	a.logger.Info("Spawned child agent",
		zap.String("child_id", childAgentID),
		zap.String("child_type", agentType),
		zap.String("role", role),
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID))

	return childAgentID, nil
}

// SpawnChildAgentWithContext spawns a child agent using the current execution context
func (a *Agent) SpawnChildAgentWithContext(ctx context.Context, execCtx *types.ExecutionContext, agentType, role string, config map[string]interface{}) (string, error) {
	// Generate child agent ID
	childAgentID := uuid.New().String()

	// Create child context from parent
	childCtx := execCtx.CreateChildContext(agentType)

	// Override with specific values
	childCtx.ToAgentType = agentType
	childCtx.Action = "initialize"
	childCtx.FunctionalRole = role

	// Create spawn request
	spawnRequest := &types.RequestMessage{
		Headers: childCtx.ToRequestHeaders(),
		Body: map[string]interface{}{
			"agent_id": childAgentID,
			"role":     role,
			"config":   config,
			"parent": map[string]interface{}{
				"agent_id":          a.AgentID,
				"agent_type":        a.AgentType,
				"orchestration_id":  execCtx.OrchestrationID,
				"correlation_id":    execCtx.CorrelationID,
				"parent_request_id": execCtx.RequestID,
			},
		},
	}

	// Send spawn request
	topic := fmt.Sprintf("system.agent.%s.requests", agentType)
	spawnBytes, err := json.Marshal(spawnRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spawn request: %w", err)
	}

	headers := spawnRequest.Headers.ToMap()
	key := []byte(childCtx.CorrelationID)

	if err := a.producer.Produce(ctx, topic, headers, key, spawnBytes); err != nil {
		return "", fmt.Errorf("failed to spawn child agent: %w", err)
	}

	return childAgentID, nil
}

// Helper functions

func isRecoverableError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "temporary")
}

func getErrorStatus(recoverable bool) string {
	if recoverable {
		return "error_recoverable"
	}
	return "error_unrecoverable"
}

// SendInitializationResponse sends a response confirming agent initialization
func (a *Agent) SendInitializationResponse(spawnRequest *types.RequestMessage) error {
	a.logger.Info("DEBUGaa: SendInitializationResponse agent.go 947",
		zap.Any("spawnRequest.Headers", spawnRequest.Headers),
	)

	response := &types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender: types.AgentIdentity{
				AgentType:    a.AgentType,
				AgentID:      a.AgentID,
				PodName:      a.PodName,
				AgentVersion: a.AgentVersion,
				Role:         a.Role,
			},

			// Response tracking
			InResponseToRequestID:      spawnRequest.Headers.RequestID,
			InResponseToStepID:         spawnRequest.Headers.StepID,
			InResponseToStepName:       spawnRequest.Headers.StepName,
			InResponseToParentOrchID:   spawnRequest.Headers.ParentOrchestrationID,
			InResponseToParentOrchName: spawnRequest.Headers.ParentOrchestrationName,
			InResponseToMessageID:      spawnRequest.Headers.MessageID,
			InResponseToAction:         spawnRequest.Headers.Action,

			// Set OrchestrationID to the parent's orchestration since that's who processes the response
			OrchestrationID:   spawnRequest.Headers.ParentOrchestrationID,
			OrchestrationName: spawnRequest.Headers.ParentOrchestrationName,

			// My context Set MyOrchestrationID to our own (the child's) orchestration
			MyOrchestrationID:   spawnRequest.Headers.OrchestrationID,
			MyOrchestrationName: spawnRequest.Headers.OrchestrationName,
			MyRequestsTopic:     a.requestsTopic,
			MyResponsesTopic:    a.responsesTopic,

			// Identity
			CorrelationID:   spawnRequest.Headers.CorrelationID,
			CorrelationName: spawnRequest.Headers.CorrelationName,
			ClientID:        spawnRequest.Headers.ClientID,
			MessageType:     "response",
			Status:          "complete",
			IsComplete:      true,
			TimeSent:        time.Now(),
			TimeSpent:       time.Since(a.spawnTime),
		},
		Body: types.ResponseBody{
			Success: true,
			Headers: nil,
			Body: map[string]interface{}{ // Direct assignment
				"agent_id":    a.AgentID,
				"agent_name":  a.AgentName,
				"agent_type":  a.AgentType,
				"role":        a.Role,
				"initialized": true,
				"topics": map[string]string{
					"requests":  a.requestsTopic,
					"responses": a.responsesTopic,
				},
			},
			Error: nil,
		},
	}

	// Send to parent's response topic
	responsesTopic := spawnRequest.Headers.ResponsesTopic
	if responsesTopic == "" {
		responsesTopic = fmt.Sprintf("system.agent.%s.responses", spawnRequest.Headers.Sender.AgentType)
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal initialization response: %w", err)
	}

	headers := response.Headers.ToMap()
	key := []byte(spawnRequest.Headers.CorrelationID)

	a.logger.Info("DEBUGaa: ResponseHeaders in SendInitializationResponse 1021",
		zap.Any("response", response),
	)

	return a.producer.Produce(a.ctx, responsesTopic, headers, key, responseBytes)
}

// registerWithParent registers this agent with its parent orchestration
func (a *Agent) registerWithParent(spawnRequest *types.RequestMessage) {
	if a.stateRepo == nil || a.ParentOrchestrationID == "" {
		return
	}

	// Create subtree info
	subtreeInfo := &types.SubtreeInfo{
		AgentID:       a.AgentID,
		AgentType:     a.AgentType,
		AgentName:     a.AgentName,
		ParentAgentID: a.ParentAgentID,
		Children:      make(map[string]*types.SubtreeInfo),
		CreatedAt:     a.spawnTime,
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted: 0,
			TasksFailed:    0,
			LastUpdated:    time.Now(),
		},
	}

	// Add to parent orchestration's subtree
	if err := a.stateRepo.AddSubtreeAgent(context.Background(), a.ParentOrchestrationID, subtreeInfo); err != nil {
		a.logger.Error("Failed to register with parent orchestration",
			zap.Error(err),
			zap.String("parent_orchestration_id", a.ParentOrchestrationID))
	} else {
		a.logger.Info("Registered with parent orchestration",
			zap.String("parent_orchestration_id", a.ParentOrchestrationID),
			zap.String("parent_agent_id", a.ParentAgentID))
	}
}

// updatePerformanceMetrics updates this agent's performance in the parent orchestration
func (a *Agent) updatePerformanceMetrics(success bool) {
	if a.stateRepo == nil || a.ParentOrchestrationID == "" {
		return
	}

	metrics := &types.PerformanceMetrics{
		LastUpdated: time.Now(),
	}

	if success {
		metrics.TasksCompleted = 1
	} else {
		metrics.TasksFailed = 1
	}

	// Update in parent orchestration
	a.stateRepo.UpdateAgentPerformance(
		context.Background(),
		a.ParentOrchestrationID,
		a.AgentID,
		metrics,
	)
}

// sendHeartbeats sends periodic heartbeats to parent
func (a *Agent) sendHeartbeats() {
	defer a.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat to the parent
func (a *Agent) sendHeartbeat() {
	if a.ParentAgentID == "" {
		return // No parent to report to
	}

	heartbeat := map[string]interface{}{
		"agent_id":           a.AgentID,
		"agent_type":         a.AgentType,
		"status":             "healthy",
		"messages_processed": a.messagesProcessed,
		"uptime_seconds":     time.Since(a.spawnTime).Seconds(),
		"last_activity":      a.lastActivity,
	}

	// Send to parent's response topic
	topic := fmt.Sprintf("system.agent.%s.responses", a.ParentAgentType)
	heartbeatBytes, _ := json.Marshal(heartbeat)

	headers := map[string]string{
		"message_type": "heartbeat",
		"agent_id":     a.AgentID,
		"agent_type":   a.AgentType,
	}

	key := []byte(a.AgentID)
	a.producer.Produce(a.ctx, topic, headers, key, heartbeatBytes)
}
