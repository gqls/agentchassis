// FILE: platform/agentbase/agent.go
package agentbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
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
	"github.com/gqls/agentchassis/platform/storage"
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
// Agent should I guess belong only on the agent itself and not be passed. we shouldn't be able to access another agent's Agent
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
	replyToRequestID      string // The request ID parent is waiting for (for error propagation)

	// Configuration
	config        *AgentConfig
	DynamicConfig map[string]interface{}
	isStateless   bool
	spawned       bool      // Whether agent was dynamically spawned
	spawnTime     time.Time // When agent was spawned
	initialized   bool      // Whether agent is initialized

	// Core components
	db            *sql.DB
	storageClient storage.Client
	producer      kafka.Producer
	orchestrator  *orchestration.SagaCoordinator
	processor     *messaging.MessageProcessor
	validator     *validation.Validator
	logger        *zap.Logger

	// Kafka consumers - ALL agents have both
	requestConsumer  *kafka.Consumer
	responseConsumer *kafka.Consumer

	// Topics
	requestsTopic  string // as an agent this is my requests topic
	responsesTopic string // as an agent this is my responses topic
	replyToTopic   string // as an agent/workflow this is where workflow wants me to reply to - basically per agent for now

	// State management
	stateRepo *orchestration.StateRepository

	// Lifecycle
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	shutdownChan chan struct{}
	shutdownOnce sync.Once

	// Metrics
	messagesProcessed uint64
	lastActivity      time.Time
}

// New creates a new agent with the standard constructor signature
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
		requestsTopic:  os.Getenv("REQUESTS_TOPIC"),
		responsesTopic: os.Getenv("RESPONSES_TOPIC"),
		replyToTopic:   os.Getenv("PARENT_RESPONSES_TOPIC"),

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

// initializeComponents initializes all agent components
func (a *Agent) initializeComponents() error {

	// Debug logging for consumer configuration
	a.logger.Info("Agent InitializeComponents",
		zap.String("requests_topic", a.requestsTopic),
		zap.Strings("kafka_brokers", a.config.KafkaBrokers))

	// Initialize database if URL provided
	if a.config.DatabaseURL != "" {
		db, err := sql.Open("pgx", a.config.DatabaseURL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		// Set a low, fixed number of max connections per pod
		db.SetMaxOpenConns(4)

		// Only keep one idle connection open per pod
		db.SetMaxIdleConns(1)

		// Close connections after a while to force recycling
		db.SetConnMaxLifetime(time.Minute * 10)

		a.db = db
		a.stateRepo = orchestration.NewStateRepository(db, a.logger)
	}

	// Create validator
	a.validator = validation.NewValidator(a.logger)

	// Create Kafka producer with validator injected
	producer, err := kafka.NewProducerWithValidator(a.config.KafkaBrokers, a.logger, a.validator)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	a.producer = producer

	// Initialize storage client for image operations (optional - may not be configured)
	// This follows the same pattern as the image adapter
	var storageClient storage.Client
	storageConfig := config.ObjectStorageConfig{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Bucket:          os.Getenv("IMAGE_BUCKET"),
		AccessKeyEnvVar: "B2_APPLICATION_KEY_ID",
		SecretKeyEnvVar: "B2_APPLICATION_KEY",
	}

	// Only create if bucket is configured
	if storageConfig.Bucket != "" {
		var err error
		storageClient, err = storage.NewS3Client(a.ctx, storageConfig, *a.logger)
		if err != nil {
			a.logger.Warn("Failed to initialize storage client - image deployment will be unavailable",
				zap.Error(err))
			// Don't fail startup - storage is optional for most workflows
		} else {
			a.logger.Info("Storage client initialized",
				zap.String("bucket", storageConfig.Bucket),
				zap.String("endpoint", storageConfig.Endpoint))
		}
	} else {
		a.logger.Info("Storage client not configured (IMAGE_BUCKET not set)")
	}
	a.storageClient = storageClient

	// EVERY agent gets an orchestrator - they all orchestrate something
	if a.db != nil {
		a.orchestrator = orchestration.NewSagaCoordinator(a.db, producer, a.storageClient, a.logger)
	} else {
		a.logger.Warn("No database configured, orchestration capabilities will be limited")
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

func (a *Agent) setupConsumers() error {
	requestsTopic := os.Getenv("REQUESTS_TOPIC")
	responsesTopic := os.Getenv("RESPONSES_TOPIC")

	if requestsTopic == "" {
		// Only the main orchestrator listens on the generic topic
		// waiting for client requests
		requestsTopic = "system.agent.generic.requests"
		a.logger.Info("Listening on generic topic for client requests")
	} else {
		a.logger.Info("Listening on job-specific topic",
			zap.String("topic", requestsTopic))
	}

	// Get consumer group from environment
	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = fmt.Sprintf("%s-consumers-%s", a.AgentType, a.AgentID[0:8])
	}

	// Use the consumer group from environment
	requestConsumer, err := kafka.NewConsumer(
		a.config.KafkaBrokers,
		requestsTopic,
		consumerGroup,
		a.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create request consumer: %w", err)
	}

	a.requestConsumer = requestConsumer

	// this is listening only, not replying to anything
	if responsesTopic == "" {
		// Only the main orchestrator listens on the generic topic
		// waiting for client requests
		responsesTopic = "system.agent.generic.responses"
		a.logger.Info("Listening on generic topic for client responses")
	} else {
		a.logger.Info("Listening on job-specific topic for responses",
			zap.String("responses topic", responsesTopic))
	}

	a.logger.Info("Setting up responseConsumer with:",
		zap.String("responses topic", responsesTopic),
		zap.String("groupID", a.AgentID[0:8]),
	)

	// Simple consumer group
	responseConsumer, err := kafka.NewConsumer(
		a.config.KafkaBrokers,
		responsesTopic,
		a.AgentID,
		a.logger,
	)

	a.responseConsumer = responseConsumer

	return err
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

	// yet another sleep
	time.Sleep(2 * time.Second)

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

	// Start idle monitor if configured (for spawned Job agents)
	if idleStr := os.Getenv("IDLE_TIMEOUT_SECONDS"); idleStr != "" {
		if idleSecs, err := strconv.Atoi(idleStr); err == nil && idleSecs > 0 {
			a.wg.Add(1)
			go a.idleMonitor(time.Duration(idleSecs) * time.Second)
		}
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
		zap.String("listening on requests topic", a.requestsTopic))

	// if it's a valid request we should check the reply to is what we'd expect
	// as an agent my reply to topic should already be in env, we can decide later if we want messages to override this. for now no.
	// does it exist? it should.
	envParentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
	envMyAgentType := os.Getenv("AGENT_TYPE")
	envRequestsTopic := os.Getenv("REQUESTS_TOPIC")
	envResponsesTopic := os.Getenv("RESPONSES_TOPIC")

	// when replying I either reply to their requests or my parent responses topic
	a.logger.Info("Starting response processor agent.go processRequests",
		zap.String("requests topic from agent struct - my requests topic?", a.requestsTopic),
		zap.String("parent responses topic from environment", envParentResponsesTopic),
		zap.String("my responses topic from environment", envResponsesTopic),
		zap.String("my requests topic from environment", envRequestsTopic),
		zap.String("which agent am I from environment", envMyAgentType),
	)

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Request processor stopping due to context cancellation")
			return
		default:
			msg, err := a.requestConsumer.Consume(a.ctx)

			if err != nil {
				// Handle timeout specifically
				if err == context.DeadlineExceeded {
					// Don't log, just continue - timeouts are normal
					continue
				}
				if err != context.Canceled {
					a.logger.Error("Failed to consume request message", zap.Error(err))
				}
				continue
			}

			// Skip empty messages
			if msg.Value == nil || len(msg.Value) == 0 {
				// This is a timeout or empty message, skip it
				a.logger.Info("Skipping empty requests message")
				continue
			}

			a.logger.Info("Request consumer received message",
				zap.Int("value_length", len(msg.Value)),
			)

			// Update activity timestamp for idle timeout
			a.lastActivity = time.Now()

			// Process the message
			a.processMessage(msg, "request")
		}
	}
}

// processResponses handles incoming response messages
func (a *Agent) processResponses() {
	defer a.wg.Done()

	// my parent responses topic should not be altered by the incoming message
	// does it exist? it should.
	envParentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
	envMyAgentType := os.Getenv("AGENT_TYPE")
	envRequestsTopic := os.Getenv("REQUESTS_TOPIC")
	envResponsesTopic := os.Getenv("RESPONSES_TOPIC")

	// when replying I either reply to their requests or my parent responses topic
	a.logger.Info("Starting response processor agent.go processResponses",
		zap.String("topic from agent struct - is this not my reply to but my own responses topic", a.responsesTopic), // it is responses topic of calculator not parent
		zap.String("parent responses topic from environment", envParentResponsesTopic),
		zap.String("my responses topic from environment", envResponsesTopic),
		zap.String("my requests topic from environment", envRequestsTopic),
		zap.String("which agent am I from environment", envMyAgentType),
	)

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("warn: Response processor stopping due to context cancellation")
			return
		default:
			msg, err := a.responseConsumer.Consume(a.ctx)

			if err != nil {
				if err != context.Canceled && err != context.DeadlineExceeded {
					a.logger.Error("Failed to consume response message", zap.Error(err))
				}
				continue
			}

			// Skip empty messages
			if msg.Value == nil || len(msg.Value) == 0 {
				// This is a timeout or empty message, skip it
				a.logger.Info("Fail: Skipping empty response message")
				continue
			}

			// Only log after we've confirmed it's a real message
			a.logger.Info("Response consumer received message",
				//zap.Any("message", msg),
				zap.Int("value_length", len(msg.Value)),
				zap.String("topic", msg.Topic))

			// Update activity timestamp for idle timeout
			a.lastActivity = time.Now()

			// Process the message
			a.processMessage(msg, "response")
		}
	}
}

// processMessage handles a single message (request or response)
func (a *Agent) processMessage(msg kafka.Message, messageType string) {

	envRequestsTopic := os.Getenv("REQUESTS_TOPIC")
	envResponsesTopic := os.Getenv("RESPONSES_TOPIC")
	envParentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")

	a.logger.Info("process single message agent.go processMessage",
		zap.String("messageType", messageType),
		//zap.Any("full message in Agent processMessage 610", msg),
		zap.String("what is this agent (from env)", os.Getenv("AGENT_TYPE")),
		zap.String("what is this agents REQUESTS_TOPIC (from env)", envRequestsTopic),
		zap.String("what is this agents RESPONSES_TOPIC (from env)", envResponsesTopic),
		zap.String("what is this agents PARENT_RESPONSES_TOPIC (from env)", envParentResponsesTopic),
	)

	// FIRST THING: Check for empty message
	if msg.Value == nil || len(msg.Value) == 0 {
		a.logger.Debug("Ignoring empty message in processMessage",
			zap.String("messageType", messageType),
			zap.String("topic", msg.Topic))
		return
	}

	startTime := time.Now()
	a.lastActivity = startTime
	a.messagesProcessed++

	// Extract headers
	headers := kafka.HeadersToMap(msg.Headers)

	a.logger.Info("process single message agent.go processMessage whats in headers",
		zap.Any("headers", headers),
	)

	// Check if this is an error message
	if headers["is_error"] == "true" {
		a.logger.Info("Received error message",
			zap.String("correlation_id", headers["correlation_id"]),
			zap.String("error_from", headers["from_agent_type"]),
			zap.String("message_type", messageType))

		if messageType == "response" {
			// Error RESPONSES must reach the coordinator so it can:
			//   - Route to error_step (e.g. mark_failed)
			//   - Handle continue_on_error in loop actions
			//   - Call handleUnrecoverableError / handleRecoverableError
			// Do NOT short-circuit here — fall through to normal processing.
			a.logger.Info("Error response — routing to coordinator for error_step handling",
				zap.String("in_response_to_request_id", headers["in_response_to_request_id"]))
		} else {
			// Error REQUESTS: pass up to parent (original behavior)
			if a.spawned && a.ParentAgentID != "" {
				headers["error_chain"] = headers["error_chain"] + "," + a.AgentType
				headers["from_agent_type"] = a.AgentType

				parentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
				a.producer.Produce(context.Background(), parentResponsesTopic, headers, msg.Key, msg.Value)
			}
			return
		}
	}

	// Validate incoming message
	validator := validation.NewValidator(a.logger)
	if !validator.ValidateIncomingMessage(headers) {
		// Send error response
		errorHeaders := make(map[string]string)
		for k, v := range headers {
			errorHeaders[k] = v
		}
		errorHeaders["is_error"] = "true"
		errorHeaders["error_reason"] = "validation_failed"
		errorHeaders["from_agent_type"] = a.AgentType

		errorBody := map[string]string{
			"error": "Message missing required fields",
			"agent": a.AgentType,
		}
		bodyBytes, _ := json.Marshal(errorBody)

		parentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
		// this is an error
		a.producer.Produce(context.Background(),
			parentResponsesTopic, errorHeaders, msg.Key, bodyBytes)

		return
	}

	// Check for empty headers as additional validation
	if len(headers) == 0 && len(msg.Value) > 0 {
		a.logger.Warn("Message has no headers but has value - possible malformed message",
			zap.String("messageType", messageType),
			zap.Int("value_length", len(msg.Value)))
		// You might want to handle this case specially
	}

	// Create or parse ExecutionContext
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		a.logger.Error("error: legacy do I get here? agent.go processMessage",
			zap.Any("error is", err),
		)
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

	// Check if orchestration_id is missing for request messages
	if execCtx.OrchestrationID == "" && messageType == "request" {
		a.logger.Error("Request message missing orchestration_id",
			zap.String("correlation_id", execCtx.CorrelationID),
			zap.String("request_id", execCtx.RequestID),
			zap.Any("headers", headers))

		// Send error response to client
		// Create proper error response
		responseMessage := types.ResponseMessage{
			Headers: types.ResponseHeaders{
				// Response tracking - what we're responding to
				InResponseToRequestID: execCtx.RequestID,
				InResponseToMessageID: execCtx.MessageID,
				InResponseToAction:    headers["action"],
				RetryCount:            0,

				// The orchestration that should process this response
				// Since there's no orchestration_id, we can't route properly
				OrchestrationID:   "", // Empty because request didn't have one
				OrchestrationName: "",

				// My Context - this agent's info
				MyOrchestrationID:   a.AgentID, // Use agent ID as fallback
				MyOrchestrationName: fmt.Sprintf("%s-validation", a.AgentType),
				MyRequestsTopic:     "", // Not applicable for validation error
				MyResponsesTopic:    "", // Not applicable for validation error

				// Identity
				CorrelationID:   execCtx.CorrelationID,
				CorrelationName: headers["correlation_name"],
				ClientID:        headers["client_id"],
				MessageType:     "response",
				FromAgent:       a.AgentID,
				ToAgent:         headers["from_agent_id"], // Send back to sender if available
				ToAgentType:     headers["from_agent_type"],

				// Status Flags
				IsComplete:          true, // This completes the request
				IsError:             true, // This is an error
				IsMultipartResponse: false,
				PartCount:           0,
				Status:              "error_validation",

				// Sender identity
				Sender: types.AgentIdentity{
					AgentID:      a.AgentID,
					AgentType:    a.AgentType,
					PodName:      a.PodName,
					AgentVersion: os.Getenv("AGENT_VERSION"),
					Role:         a.Role,
				},

				// Timing & Resources
				TimeSent:                   time.Now(),
				TimeSpent:                  time.Since(startTime),
				OverallTimeBudgetRemaining: 0, // No budget for failed validation
				TopicSentTo:                headers["responses_topic"],
				FuelUsed:                   0, // No fuel used for validation error
				RemainingFuelBudget:        0,
			},
			Body: types.ResponseBody{
				Success: false,
				Headers: map[string]interface{}{
					// Include the problematic headers for debugging
					"received_headers": headers,
					"missing_field":    "orchestration_id",
				},
				Body: map[string]interface{}{
					"error_type":    "validation_error",
					"error_message": "orchestration_id is required in request headers",
					"details": map[string]interface{}{
						"correlation_id": execCtx.CorrelationID,
						"request_id":     execCtx.RequestID,
						"message_id":     execCtx.MessageID,
						"timestamp":      time.Now().Unix(),
					},
				},
				Error: &types.ErrorInfo{
					Code:        "MISSING_ORCHESTRATION_ID",
					Message:     "orchestration_id is required in request headers",
					Recoverable: false, // Client must fix their request
					RetryAfter:  0,     // Don't retry
				},
			},
		}
		a.sendErrorResponse(execCtx, &responseMessage)

		// Mark message as processed despite the error to prevent infinite retry
		observability.MessagesDropped.WithLabelValues(a.AgentType, "missing_orchestration_id").Inc()
		return
	}

	// Add agent context
	execCtx.ProcessingNode = a.PodName
	execCtx.ToAgentID = a.AgentID
	execCtx.ToAgentType = a.AgentType

	// For static agents, capture the response topic from request headers
	if a.IsStaticAgent() && headers["responses_topic"] != "" {
		execCtx.ResponsesTopic = headers["responses_topic"]
		a.logger.Info("Static agent storing response topic from request",
			zap.String("responses_topic", execCtx.ResponsesTopic))
	}

	contextLogger := a.logger.With(execCtx.LogContext()...)

	contextLogger.Info("Processing message STORE_EXEC_CONTEXT ish",
		zap.String("message_type", messageType),
		zap.String("message_id", execCtx.MessageID),
		zap.String("stored_responses_topic should be the parents one", execCtx.ResponsesTopic),
		zap.String("I am in this agent type:", os.Getenv("AGENT_TYPE")),
		zap.String("stored_request_id", execCtx.RequestID),
		zap.String("orchestration_id", execCtx.OrchestrationID),
		zap.String("parent_orch_id", execCtx.ParentOrchestrationID),
		zap.Int("payload_size", len(msg.Value)))

	// Check for duplicates if stateless
	if a.isStateless && a.stateRepo != nil {
		isDuplicate, err := a.stateRepo.HasProcessedMessage(a.ctx, execCtx.CorrelationID, execCtx.RequestID, a.AgentID, execCtx.RetryVersion)
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

		// Check if this is a validation error
		errMsg := err.Error()
		if strings.Contains(errMsg, "is required") ||
			strings.Contains(errMsg, "validation") ||
			strings.Contains(errMsg, "invalid") {

			contextLogger.Warn("Validation error in message processing - not calling handleProcessingError",
				zap.Error(err),
				zap.String("message_type", messageType),
				zap.String("correlation_id", execCtx.CorrelationID))

			// Record metric but don't retry
			observability.MessagesDropped.WithLabelValues(a.AgentType, "validation_error").Inc()

			// Just return - don't call handleProcessingError
			// This prevents the error from being propagated and causing a retry
			return
		}

		// For non-validation errors, handle normally
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

	// IMPORTANT: If this agent has a parent, also notify the parent orchestrator
	// This ensures the parent knows this child workflow has failed
	// We check PARENT_RESPONSES_TOPIC env var to determine if we have a parent
	parentResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
	if parentResponsesTopic != "" && a.replyToRequestID != "" {
		a.logger.Info("Propagating error to parent orchestrator",
			zap.String("parent_topic", parentResponsesTopic),
			zap.String("parent_orchestration_id", a.ParentOrchestrationID),
			zap.String("reply_to_request_id", a.replyToRequestID),
			zap.Error(err))

		// Build error headers to notify parent
		errorHeaders := map[string]string{
			"correlation_id":                         execCtx.CorrelationID,
			"orchestration_id":                       a.ParentOrchestrationID,
			"message_type":                           "response",
			"is_complete":                            "false",
			"is_error":                               "true",
			"status":                                 "failed",
			"sender_agent_type":                      a.AgentType,
			"sender_agent_id":                        a.AgentID,
			"sender_pod_name":                        a.PodName,
			"from_agent_type":                        a.AgentType,
			"error_chain":                            a.AgentType,
			"in_response_to_request_id":              a.replyToRequestID, // The original request from parent
			"in_response_to_parent_orchestration_id": a.ParentOrchestrationID,
		}

		errorBody := map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":        "CHILD_PROCESSING_ERROR",
				"message":     fmt.Sprintf("Child orchestration %s failed: %s", a.AgentType, err.Error()),
				"recoverable": recoverable,
				"agent_type":  a.AgentType,
				"agent_id":    a.AgentID,
			},
		}
		bodyBytes, _ := json.Marshal(errorBody)

		if err := a.producer.Produce(context.Background(), parentResponsesTopic, errorHeaders, []byte(execCtx.CorrelationID), bodyBytes); err != nil {
			a.logger.Error("Failed to propagate error to parent",
				zap.Error(err),
				zap.String("parent_topic", parentResponsesTopic))
		} else {
			a.logger.Info("Successfully propagated error to parent orchestrator",
				zap.String("parent_topic", parentResponsesTopic),
				zap.String("reply_to_request_id", a.replyToRequestID))
		}
	} else if parentResponsesTopic != "" {
		a.logger.Warn("Cannot propagate error to parent - missing replyToRequestID",
			zap.String("parent_topic", parentResponsesTopic),
			zap.String("reply_to_request_id", a.replyToRequestID))
	}
}

// sendErrorResponse sends an error response
func (a *Agent) sendErrorResponse(execCtx *types.ExecutionContext, response *types.ResponseMessage) {
	// Determine response topic
	var responsesTopic string

	// For static agents, use the topic specified in the request
	if a.IsStaticAgent() && execCtx.ResponsesTopic != "" {
		responsesTopic = execCtx.ResponsesTopic
		a.logger.Debug("Static agent using request-specified response topic",
			zap.String("responses_topic", responsesTopic))
	} else if execCtx.ResponsesTopic != "" {
		responsesTopic = execCtx.ResponsesTopic
	} else if os.Getenv("PARENT_RESPONSES_TOPIC") != "" {
		// Fallback to parent responses topic for spawned agents
		responsesTopic = os.Getenv("PARENT_RESPONSES_TOPIC")
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
	a.logger.Info("Sending response with headers",
		zap.String("sender_role", headers["sender_role"]),
		zap.Any("all_headers", headers))

	// Use correlation ID as key
	key := []byte(execCtx.CorrelationID)
	if execCtx.CorrelationID == "" {
		key = []byte(execCtx.MessageID)
	}

	// Use producer's signature // error so no message validation
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

	// Signal shutdown — protected by sync.Once (idle monitor may have closed it already)
	a.shutdownOnce.Do(func() {
		close(a.shutdownChan)
	})

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

	// Clean up ephemeral Kafka topics (only for spawned agents with EPHEMERAL_TOPICS=true)
	a.cleanupEphemeralTopics()

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
	a.logger.Info("DEBUGaa: SendInitializationResponse agent.go 947 looking for correct reply-to topic, check I am on sending agent",
		zap.Any("spawnRequest.Headers", spawnRequest.Headers),
		zap.Any("spawnRequest.Body", spawnRequest.Body),
		zap.String("(TopicSentTo)/ParentResponsesTopic", spawnRequest.Headers.ParentResponsesTopic),
		zap.Any("whats in the agent", a),
	)

	current, caller := getFuncInfo(1)
	a.logger.Info("In agent.go SendInitializationResponse, callstack",
		zap.String("current function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Extract role if not already set
	if a.Role == "" {
		if body, ok := spawnRequest.Body.(map[string]interface{}); ok {
			if role, ok := body["role"].(string); ok {
				a.Role = role // Store it!
			}
		}
	}

	// Store the request ID that parent is waiting for - needed for error propagation
	if spawnRequest.Headers.RequestID != "" {
		a.replyToRequestID = spawnRequest.Headers.RequestID
	}
	// Store parent orchestration ID if not already set
	if a.ParentOrchestrationID == "" && spawnRequest.Headers.ParentOrchestrationID != "" {
		a.ParentOrchestrationID = spawnRequest.Headers.ParentOrchestrationID
	}
	a.logger.Info("Stored spawn context for erro r propagation",
		zap.String("reply_to_request_id", a.replyToRequestID),
		zap.String("parent_orchestration_id", a.ParentOrchestrationID))

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
			InResponseToRequestID:      spawnRequest.Headers.RequestID, //?
			ReplyToRequestID:           spawnRequest.Headers.ReplyToRequestID,
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
			TopicSentTo:         spawnRequest.Headers.ParentResponsesTopic,

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

	// Send to parent's response topic or request-specified topic
	var responsesTopic string

	// For static agents, prefer the topic from request headers
	if a.IsStaticAgent() && spawnRequest.Headers.ResponsesTopic != "" {
		responsesTopic = spawnRequest.Headers.ResponsesTopic
		a.logger.Info("Static agent using request-specified response topic - SendInitializationResponse",
			zap.String("response_stopic", responsesTopic))
	} else if spawnRequest.Headers.ResponsesTopic != "" {
		// Use topic from request if specified
		responsesTopic = spawnRequest.Headers.ResponsesTopic
		a.logger.Info("request was specified in the request headers",
			zap.String("responses_topic", responsesTopic))
	} else {
		// Fallback to environment variable
		responsesTopic = os.Getenv("PARENT_RESPONSES_TOPIC")
		if responsesTopic == "" {
			responsesTopic = spawnRequest.Headers.ParentResponsesTopic
			a.logger.Warn("error: environment var PARENT_RESPONSES_TOPIC was blank so probably sending to wrong topic now",
				zap.String("responsesTopic", responsesTopic),
			)
		}
		if responsesTopic == "" {
			responsesTopic = "system.agent.generic.responses"
			a.logger.Warn("Using default responses topic - no other responses topics found - SendInitializationResponse",
				zap.String("responsesTopic", responsesTopic))
		}
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

	// Wait for consumers to be ready
	time.Sleep(2 * time.Second)

	return a.producer.ProduceWithValidation(a.ctx, responsesTopic, headers, key, responseBytes)
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
	// heartbeat so no message validation
	a.producer.Produce(a.ctx, topic, headers, key, heartbeatBytes)
}

// idleMonitor checks lastActivity and triggers shutdown if the agent
// has been idle for longer than IDLE_TIMEOUT_SECONDS. This ensures
// spawned agent Jobs exit cleanly so K8s can reclaim resources.
//
// Only runs when IDLE_TIMEOUT_SECONDS > 0 (set by job spawner from
// agent_definitions.idle_timeout_seconds).
func (a *Agent) idleMonitor(timeout time.Duration) {
	defer a.wg.Done()

	checkInterval := 10 * time.Second
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	a.logger.Info("Idle monitor started",
		zap.Duration("timeout", timeout),
		zap.Duration("check_interval", checkInterval))

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			idle := time.Since(a.lastActivity)
			if idle >= timeout {
				a.logger.Info("Idle timeout reached — shutting down",
					zap.Duration("idle_duration", idle),
					zap.Duration("timeout", timeout),
					zap.Time("last_activity", a.lastActivity))

				// Signal shutdown — protected by sync.Once to prevent
				// panic if Shutdown() is also called (e.g. from signal handler)
				a.shutdownOnce.Do(func() {
					close(a.shutdownChan)
				})
				return
			}
		}
	}
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

// IsStaticAgent checks if this agent is statically deployed
func (a *Agent) IsStaticAgent() bool {
	// Static agents have well-known request topics FIXME
	return strings.HasPrefix(a.requestsTopic, "system.agent.")
}

// cleanupEphemeralTopics deletes the agent's request and response topics.
// Only runs when EPHEMERAL_TOPICS=true (set by job spawner for per-spawn topics).
// When we move to shared topics, this env var won't be set and cleanup is skipped.
func (a *Agent) cleanupEphemeralTopics() {
	if os.Getenv("EPHEMERAL_TOPICS") != "true" {
		return
	}

	brokersEnv := os.Getenv("SERVICE_INFRASTRUCTURE_KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = os.Getenv("KAFKA_BROKERS")
	}
	if brokersEnv == "" {
		a.logger.Warn("Cannot clean up topics — no KAFKA_BROKERS configured")
		return
	}
	brokers := strings.Split(brokersEnv, ",")

	topicManager := kafka.NewTopicManager(brokers, a.logger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	topics := []string{a.requestsTopic, a.responsesTopic}
	for _, topic := range topics {
		if topic == "" || !strings.HasPrefix(topic, "job.") {
			continue // Only clean up job.* topics, never system topics
		}
		if err := topicManager.DeleteTopic(ctx, topic); err != nil {
			a.logger.Warn("Failed to clean up topic",
				zap.String("topic", topic),
				zap.Error(err))
		} else {
			a.logger.Info("Cleaned up ephemeral topic",
				zap.String("topic", topic))
		}
	}
}
