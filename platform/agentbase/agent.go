// FILE: platform/agentbase/agent.go (refactored)
package agentbase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/health"
	"github.com/gqls/agentchassis/platform/infrastructure"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/types"
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

// SpawnChildAgent creates and spawns a child agent
"fmt"
"os"
"time"

"github.com/google/uuid"
"github.com/gqls/agentchassis/platform/types"
"go.uber.org/zap"
)

// SpawnChildAgent creates and spawns a child agent using existing database schema
func (a *Agent) SpawnChildAgent(ctx *types.ExecutionContext, childType string, body interface{}) error {
	// Create child context
	childCtx := ctx.CreateChildContext(childType)

	// Set sender identity
	childCtx.Sender = types.AgentIdentity{
		AgentID:      a.agentID.String(),
		AgentType:    a.agentType,
		PodName:      a.getPodName(),
		AgentVersion: a.getAgentVersion(),
	}

	childCtx.ToAgentType = childType
	childCtx.Action = "initialize"
	childCtx.RequestsTopic = fmt.Sprintf("system.agent.%s.requests", childType)
	childCtx.ResponsesTopic = fmt.Sprintf("system.agent.%s.responses", a.agentType)

	// Convert to request message
	request := types.RequestMessage{
		Headers: childCtx.ToRequestHeaders(),
		Body:    body,
	}

	// Get database connection
	db := a.infraManager.GetConnections().ClientsDB

	// Create child orchestration state in database
	childAgentID := uuid.New() // The child agent's ID (will be assigned when it initializes)

	_, err := db.Exec(a.ctx, `
		INSERT INTO orchestration_states (
			orchestration_id,
			correlation_id,
			owner_agent_id,
			client_id,
			status,
			current_step,
			workflow_plan,
			parent_orchestration_id,
			fuel_budget,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		childCtx.OrchestrationID,           // Child's orchestration ID
		childCtx.CorrelationID,             // Same correlation across all
		childAgentID,                        // Placeholder for child agent
		childCtx.ClientID,                   // Client ID
		"pending",                           // Initial status
		"initialization",                    // Current step
		map[string]interface{}{              // Workflow plan
			"type": "child_agent_workflow",
			"agent_type": childType,
			"parent_orchestration": ctx.OrchestrationID,
		},
		ctx.OrchestrationID,                 // Parent orchestration ID
		childCtx.FuelBudget,                 // Fuel budget
		time.Now(),
		time.Now(),
	)

	if err != nil {
		a.logger.Error("Failed to create child orchestration state",
			zap.String("child_orchestration_id", childCtx.OrchestrationID),
			zap.Error(err))
		return fmt.Errorf("failed to create child orchestration state: %w", err)
	}

	// Track the request in orchestration_requests table
	_, err = db.Exec(a.ctx, `
		INSERT INTO orchestration_requests (
			request_id,
			orchestration_id,
			from_agent_id,
			to_agent_id,
			message_type,
			status,
			timeout_at,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		childCtx.RequestID,                              // Request ID
		ctx.OrchestrationID,                            // Parent's orchestration
		a.agentID,                                      // From parent agent
		childAgentID,                                    // To child agent (placeholder)
		"spawn_request",                                 // Message type
		"pending",                                       // Status
		time.Now().Add(time.Duration(childCtx.TimeoutSeconds) * time.Second), // Timeout
		time.Now(),
	)

	if err != nil {
		a.logger.Error("Failed to track orchestration request",
			zap.String("request_id", childCtx.RequestID),
			zap.Error(err))
		return fmt.Errorf("failed to track orchestration request: %w", err)
	}

	// Also track in pending_requests for timeout handling
	_, err = db.Exec(a.ctx, `
		INSERT INTO pending_requests (
			request_id,
			orchestration_id,
			to_agent_id,
			status,
			timeout_at,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		childCtx.RequestID,
		ctx.OrchestrationID,
		childAgentID,
		"pending",
		time.Now().Add(time.Duration(childCtx.TimeoutSeconds) * time.Second),
		time.Now(),
	)

	if err != nil {
		a.logger.Warn("Failed to add to pending_requests",
			zap.String("request_id", childCtx.RequestID),
			zap.Error(err))
		// Non-fatal, continue
	}

	// Update parent orchestration's awaited_steps
	_, err = db.Exec(a.ctx, `
		UPDATE orchestration_states 
		SET 
			awaited_steps = awaited_steps || $1::jsonb,
			updated_at = NOW()
		WHERE orchestration_id = $2
	`,
		fmt.Sprintf(`["%s"]`, childCtx.RequestID), // Add request ID to awaited steps
		ctx.OrchestrationID,
	)

	if err != nil {
		a.logger.Error("Failed to update parent awaited_steps",
			zap.String("orchestration_id", ctx.OrchestrationID),
			zap.Error(err))
		// Continue anyway - the request is tracked
	}

	// Get Kafka producer
	producer := a.infraManager.GetConnections().KafkaProducer

	// Convert request to JSON bytes
	messageBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create Kafka headers
	headers := make(map[string]string)
	headers["correlation_id"] = childCtx.CorrelationID
	headers["orchestration_id"] = childCtx.OrchestrationID
	headers["parent_orchestration_id"] = ctx.OrchestrationID
	headers["request_id"] = childCtx.RequestID
	headers["message_type"] = "request"
	headers["from_agent_type"] = a.agentType
	headers["to_agent_type"] = childType
	headers["timestamp"] = time.Now().Format(time.RFC3339)

	// Use parent orchestration ID as partition key
	key := []byte(ctx.OrchestrationID)

	// Publish the message
	err = producer.Produce(
		a.ctx,
		childCtx.RequestsTopic,
		headers,
		key,
		messageBytes,
	)

	if err != nil {
		// Update request status to failed
		db.Exec(a.ctx, `
			UPDATE orchestration_requests 
			SET status = 'failed', completed_at = NOW() 
			WHERE request_id = $1
		`, childCtx.RequestID)

		return fmt.Errorf("failed to spawn child agent: %w", err)
	}

	// Log system event
	_, _ = db.Exec(a.ctx, `
		INSERT INTO system_events (
			event_type, entity_type, entity_id, client_id, 
			metadata, severity, source
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		"agent_spawned",
		"agent",
		childAgentID.String(),
		childCtx.ClientID,
		map[string]interface{}{
			"parent_agent": a.agentID.String(),
			"parent_type": a.agentType,
			"child_type": childType,
			"orchestration_id": childCtx.OrchestrationID,
		},
		"info",
		a.agentType,
	)

	a.logger.Info("Successfully spawned child agent",
		zap.String("parent_agent", a.agentType),
		zap.String("child_type", childType),
		zap.String("child_request_id", childCtx.RequestID),
		zap.String("child_orchestration_id", childCtx.OrchestrationID))

	return nil
}

// ProcessChildResponse handles responses from child agents using existing schema
func (a *Agent) ProcessChildResponse(ctx context.Context, response types.ResponseMessage) error {
	db := a.infraManager.GetConnections().ClientsDB

	// Check if this response is for a request we're tracking
	var orchestrationID uuid.UUID
	var fromAgentID uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT orchestration_id, from_agent_id 
		FROM orchestration_requests 
		WHERE request_id = $1 AND status = 'pending'
	`, response.Headers.InResponseToRequestID).Scan(&orchestrationID, &fromAgentID)

	if err != nil {
		a.logger.Debug("Response not for a pending request",
			zap.String("request_id", response.Headers.InResponseToRequestID))
		return nil
	}

	// Handle based on response status
	switch response.Headers.Status {
	case "awaiting", "processing":
		// Progress update - just log
		a.logger.Info("Child agent progress",
			zap.String("request_id", response.Headers.InResponseToRequestID),
			zap.String("status", response.Headers.Status))

		// Update orchestration state's execution metadata
		_, _ = db.Exec(ctx, `
			UPDATE orchestration_states 
			SET 
				execution_metadata = execution_metadata || $1::jsonb,
				last_activity = NOW()
			WHERE orchestration_id = $2
		`,
			map[string]interface{}{
				"last_child_update": response.Headers.Status,
				"last_child_update_time": time.Now(),
			},
			orchestrationID,
		)
		return nil

	case "complete":
		// Update request status
		_, err = db.Exec(ctx, `
			UPDATE orchestration_requests 
			SET status = 'completed', completed_at = NOW() 
			WHERE request_id = $1
		`, response.Headers.InResponseToRequestID)

		if err != nil {
			a.logger.Error("Failed to update request status",
				zap.Error(err))
		}

		// Update pending_requests
		_, _ = db.Exec(ctx, `
			UPDATE pending_requests 
			SET status = 'completed', completed_at = NOW() 
			WHERE request_id = $1
		`, response.Headers.InResponseToRequestID)

		// Remove from parent's awaited_steps
		_, err = db.Exec(ctx, `
			UPDATE orchestration_states 
			SET 
				awaited_steps = (
					SELECT jsonb_agg(elem)
					FROM jsonb_array_elements(awaited_steps) elem
					WHERE elem::text != $1
				),
				collected_data = collected_data || $2::jsonb,
				updated_at = NOW()
			WHERE orchestration_id = $3
		`,
			fmt.Sprintf(`"%s"`, response.Headers.InResponseToRequestID),
			map[string]interface{}{
				"child_result": response.Body,
				"child_orchestration": response.Headers.MyOrchestrationID,
			},
			orchestrationID,
		)

		if err != nil {
			a.logger.Error("Failed to update parent orchestration",
				zap.Error(err))
		}

		// Log completion event
		_, _ = db.Exec(ctx, `
			INSERT INTO system_events (
				event_type, entity_type, entity_id, 
				metadata, severity, source
			) VALUES ($1, $2, $3, $4, $5, $6)
		`,
			"child_agent_completed",
			"orchestration",
			orchestrationID.String(),
			map[string]interface{}{
				"request_id": response.Headers.InResponseToRequestID,
				"child_orchestration": response.Headers.MyOrchestrationID,
			},
			"info",
			a.agentType,
		)

		// Check if all children are complete
		return a.checkAndCompleteOrchestration(ctx, orchestrationID)

	case "error_recoverable":
		a.logger.Warn("Child agent error (recoverable)",
			zap.String("request_id", response.Headers.InResponseToRequestID))
		// TODO: Implement retry logic
		return nil

	case "error_unrecoverable":
		// Mark request as failed
		_, _ = db.Exec(ctx, `
			UPDATE orchestration_requests 
			SET status = 'failed', completed_at = NOW() 
			WHERE request_id = $1
		`, response.Headers.InResponseToRequestID)

		// Update orchestration state
		_, _ = db.Exec(ctx, `
			UPDATE orchestration_states 
			SET 
				status = 'failed',
				error = $1,
				updated_at = NOW()
			WHERE orchestration_id = $2
		`,
			fmt.Sprintf("Child agent failed: %v", response.Body),
			orchestrationID,
		)

		// Log error event
		_, _ = db.Exec(ctx, `
			INSERT INTO system_events (
				event_type, entity_type, entity_id, 
				metadata, severity, source
			) VALUES ($1, $2, $3, $4, $5, $6)
		`,
			"child_agent_failed",
			"orchestration",
			orchestrationID.String(),
			map[string]interface{}{
				"request_id": response.Headers.InResponseToRequestID,
				"error": response.Body,
			},
			"error",
			a.agentType,
		)
	}

	return nil
}

// checkAndCompleteOrchestration checks if all children are done
func (a *Agent) checkAndCompleteOrchestration(ctx context.Context, orchestrationID uuid.UUID) error {
	db := a.infraManager.GetConnections().ClientsDB

	// Check if there are any pending requests
	var pendingCount int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM orchestration_requests 
		WHERE orchestration_id = $1 AND status = 'pending'
	`, orchestrationID).Scan(&pendingCount)

	if err != nil || pendingCount > 0 {
		return nil // Still waiting for responses
	}

	// All children complete - update orchestration
	_, err = db.Exec(ctx, `
		UPDATE orchestration_states 
		SET 
			status = 'completed',
			updated_at = NOW()
		WHERE orchestration_id = $1 AND status != 'failed'
	`, orchestrationID)

	return err
}

// Helper methods
func (a *Agent) getPodName() string {
	if podName := os.Getenv("HOSTNAME"); podName != "" {
		return podName
	}
	return a.agentID.String()
}

func (a *Agent) getAgentVersion() string {
	if version := os.Getenv("AGENT_VERSION"); version != "" {
		return version
	}
	return "v1.0.0"
}

// sendResponseToParent sends a response back to the parent orchestration
func (a *Agent) sendResponseToParent(ctx context.Context, orchestrationID string, childResponse types.ResponseMessage) error {
	// Get parent orchestration details from database
	db := a.infraManager.GetConnections().ClientsDB

	var parentOrchID, parentRequestID, parentResponseTopic string
	err := db.QueryRow(ctx, `
		SELECT parent_orchestration_id, parent_request_id, parent_response_topic
		FROM orchestration_states
		WHERE orchestration_id = $1
	`, orchestrationID).Scan(&parentOrchID, &parentRequestID, &parentResponseTopic)

	if err != nil {
		// No parent to respond to (might be top-level orchestration)
		return nil
	}

	// Create response to parent
	parentResponse := types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender: types.AgentIdentity{
				AgentID:      a.agentID.String(),
				AgentType:    a.agentType,
				PodName:      a.getPodName(),
				AgentVersion: a.getAgentVersion(),
			},
			InResponseToRequestID: parentRequestID,
			MyOrchestrationID:     orchestrationID,
			Status:                "complete",
			IsComplete:            true,
			CorrelationID:         childResponse.Headers.CorrelationID,
			TimeSent:              time.Now(),
		},
		Body: childResponse.Body,
	}

	// Marshal and send
	messageBytes, err := json.Marshal(parentResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal parent response: %w", err)
	}

	headers := make(map[string]string)
	headers["correlation_id"] = parentResponse.Headers.CorrelationID
	headers["message_type"] = "response"
	headers["status"] = "complete"

	producer := a.infraManager.GetConnections().KafkaProducer
	return producer.Produce(ctx, parentResponseTopic, headers, []byte(parentOrchID), messageBytes)
}

// sendErrorToParent sends an error response to the parent orchestration
func (a *Agent) sendErrorToParent(ctx context.Context, orchestrationID string, errorResponse types.ResponseMessage) error {
	// Similar to sendResponseToParent but with error status
	// Implementation would be similar with Status: "error_unrecoverable"
	return nil // TODO: Implement
}

// retryChildRequest retries a failed child request
func (a *Agent) retryChildRequest(ctx context.Context, requestID string, retryCount int) error {
	// TODO: Implement retry logic
	// This would:
	// 1. Load the original request from database
	// 2. Increment retry_version
	// 3. Resend the request
	// 4. Update retry_count in awaited_requests
	return nil
}

// HandleSpawnRequest handles incoming requests to spawn child agents
// This would be called by your message processor
func (a *Agent) HandleSpawnRequest(ctx *types.ExecutionContext, spawnRequest map[string]interface{}) error {
	// Extract child type from request
	childType, ok := spawnRequest["agent_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid agent_type in spawn request")
	}

	// Extract config for child
	childConfig := spawnRequest["config"]
	if childConfig == nil {
		childConfig = map[string]interface{}{}
	}

	// Spawn the child
	return a.SpawnChildAgent(ctx, childType, map[string]interface{}{
		"action": "initialize",
		"config": childConfig,
	})
}
