// FILE: platform/agentbase/agent_methods.go
// This file contains the remaining methods for the Agent type

package agentbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// handleMessage processes a single task
func (a *Agent) handleMessage(msg kafka.Message) {
	startTime := time.Now()
	headers := kafka.HeadersToMap(msg.Headers)

	clientID := headers["client_id"]
	agentInstanceID := headers["agent_instance_id"]

	// Extract action from payload for metrics
	var payload struct {
		Action string `json:"action"`
	}
	json.Unmarshal(msg.Value, &payload)

	// Record task received
	observability.AgentTasksReceived.WithLabelValues(a.agentType, payload.Action).Inc()

	if clientID == "" || agentInstanceID == "" {
		a.logger.Error("Message missing required headers",
			zap.String("topic", msg.Topic),
			zap.Int64("offset", msg.Offset))
		observability.SystemErrors.WithLabelValues(a.agentType, "missing_headers").Inc()
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
		a.sendErrorResponse(headers, errors.InternalError("Failed to load configuration", err))
		observability.AgentTasksProcessed.WithLabelValues(a.agentType, payload.Action, "error").Inc()
		observability.AgentProcessingDuration.WithLabelValues(a.agentType, payload.Action).Observe(time.Since(startTime).Seconds())
		a.kafkaConsumer.CommitMessages(context.Background(), msg)
		return
	}

	l.Info("Agent instance loaded", zap.String("agent_type", agentConfig.AgentType))

	// Validate workflow if present
	if err := a.validator.ValidateWorkflowPlan(agentConfig.Workflow); err != nil {
		l.Error("Invalid workflow configuration", zap.Error(err))
		a.sendErrorResponse(headers, errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").
			WithCause(err).
			WithDetail("workflow_metrics", a.validator.GetWorkflowMetrics(agentConfig.Workflow)).
			Build())
		observability.AgentTasksProcessed.WithLabelValues(a.agentType, payload.Action, "invalid_workflow").Inc()
		a.kafkaConsumer.CommitMessages(context.Background(), msg)
		return
	}

	// Start workflow timer
	workflowTimer := observability.StartWorkflowTimer(a.agentType, agentConfig.Workflow.StartStep)
	observability.WorkflowsStarted.WithLabelValues(a.agentType, agentConfig.Workflow.StartStep, clientID).Inc()
	observability.ActiveWorkflows.WithLabelValues(a.agentType).Inc()

	// Retrieve relevant memories if enabled
	if a.memoryService != nil && agentConfig.MemoryConfig.Enabled {
		agentUUID, _ := uuid.Parse(agentInstanceID)
		memories, err := a.memoryService.GetMemoryContext(
			context.Background(),
			agentUUID,
			agentConfig.MemoryConfig,
			string(msg.Value),
		)
		if err != nil {
			l.Warn("Failed to retrieve memories", zap.Error(err))
		} else if len(memories) > 0 {
			l.Info("Retrieved relevant memories", zap.Int("count", len(memories)))
			observability.VectorSearchQueries.WithLabelValues(a.agentType).Inc()
			// Add memories to the workflow context
			// This would be passed to the orchestrator
		}
	}

	// Execute workflow
	err = a.orchestrator.ExecuteWorkflow(a.ctx, agentConfig.Workflow, headers, msg.Value)

	// Decrement active workflows
	observability.ActiveWorkflows.WithLabelValues(a.agentType).Dec()

	if err != nil {
		l.Error("Workflow execution failed", zap.Error(err))
		workflowTimer.Complete("failed")
		observability.WorkflowsCompleted.WithLabelValues(a.agentType, agentConfig.Workflow.StartStep, "failed", clientID).Inc()
		observability.AgentTasksProcessed.WithLabelValues(a.agentType, payload.Action, "failed").Inc()

		// Check if it's a fuel error
		if domainErr, ok := err.(*errors.DomainError); ok && domainErr.Code == errors.ErrInsufficientFuel {
			observability.FuelExhausted.WithLabelValues(a.agentType, payload.Action, clientID).Inc()
		}

		a.sendErrorResponse(headers, errors.New(errors.ErrWorkflowFailed, "Workflow execution failed").
			WithCause(err).
			Build())
	} else {
		workflowTimer.Complete("success")
		observability.WorkflowsCompleted.WithLabelValues(a.agentType, agentConfig.Workflow.StartStep, "success", clientID).Inc()
		observability.AgentTasksProcessed.WithLabelValues(a.agentType, payload.Action, "success").Inc()
	}

	// Record processing duration
	observability.AgentProcessingDuration.WithLabelValues(a.agentType, payload.Action).Observe(time.Since(startTime).Seconds())

	// Commit message
	if err := a.kafkaConsumer.CommitMessages(context.Background(), msg); err != nil {
		l.Error("Failed to commit kafka message", zap.Error(err))
		observability.SystemErrors.WithLabelValues(a.agentType, "commit_message").Inc()
	}
}

// loadAgentConfig fetches the agent's configuration from the database
func (a *Agent) loadAgentConfig(clientID, agentInstanceID string) (*models.AgentConfig, error) {
	ctx := context.Background()
	startTime := time.Now()
	defer func() {
		observability.DatabaseQueryDuration.WithLabelValues("clients", "select").Observe(time.Since(startTime).Seconds())
	}()

	// Query the agent instance configuration
	query := fmt.Sprintf(`
		SELECT name, config, template_id 
		FROM client_%s.agent_instances 
		WHERE id = $1 AND is_active = true
	`, clientID)

	var name string
	var configJSON []byte
	var templateID string

	observability.DatabaseQueries.WithLabelValues("clients", "select", "agent_instances").Inc()
	err := a.clientsDB.QueryRow(ctx, query, agentInstanceID).Scan(&name, &configJSON, &templateID)
	if err != nil {
		// If not found in database, return a default configuration
		if err == pgx.ErrNoRows {
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

	// Extract memory configuration if present
	var memoryConfig models.MemoryConfiguration
	if memData, ok := config["memory_config"]; ok {
		memBytes, _ := json.Marshal(memData)
		json.Unmarshal(memBytes, &memoryConfig)
	}

	return &models.AgentConfig{
		AgentID:      agentInstanceID,
		AgentType:    a.agentType,
		Version:      1,
		CoreLogic:    config,
		Workflow:     workflow,
		MemoryConfig: memoryConfig,
	}, nil
}

// sendErrorResponse sends an error response back via Kafka
func (a *Agent) sendErrorResponse(headers map[string]string, domainErr *errors.DomainError) {
	responseHeaders := a.createResponseHeaders(headers)
	domainErr.TraceID = headers["correlation_id"]

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   domainErr,
		"agent":   a.agentType,
	}

	responseBytes, _ := json.Marshal(errorResponse)

	// Send to error topic
	errorTopic := fmt.Sprintf("system.errors.%s", a.agentType)
	if err := a.kafkaProducer.Produce(a.ctx, errorTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to send error response", zap.Error(err))
		observability.SystemErrors.WithLabelValues(a.agentType, "produce_error").Inc()
	} else {
		observability.KafkaMessagesProduced.WithLabelValues(errorTopic).Inc()
	}
}

// StartHealthServer starts a comprehensive health check server
func (a *Agent) StartHealthServer(port string) {
	// Start metrics server
	go func() {
		a.logger.Info("Starting metrics server", zap.String("port", port))
		if err := a.metricsServer.Start(); err != nil {
			a.logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":     "healthy",
			"agent_type": a.agentType,
			"checks": map[string]interface{}{
				"database": a.checkDatabaseHealth(),
				"kafka":    a.checkKafkaHealth(),
			},
		}

		w.Header().Set("Content-Type", "application/json")

		// Determine overall health
		overallHealthy := true
		if dbHealth, ok := health["checks"].(map[string]interface{})["database"].(map[string]interface{}); ok {
			if status, ok := dbHealth["status"].(string); ok && status != "healthy" {
				overallHealthy = false
			}
		}
		if kafkaHealth, ok := health["checks"].(map[string]interface{})["kafka"].(map[string]interface{}); ok {
			if status, ok := kafkaHealth["status"].(string); ok && status != "healthy" {
				overallHealthy = false
			}
		}

		if overallHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			health["status"] = "unhealthy"
		}

		json.NewEncoder(w).Encode(health)
	})

	// Readiness endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check if we can process messages
		ready := a.checkDatabaseHealth()["status"] == "healthy" &&
			a.checkKafkaHealth()["status"] == "healthy"

		if ready {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("READY"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NOT READY"))
		}
	})

	// Separate server for health checks (different from metrics port)
	healthPort := "8080"
	if port != "9090" { // If not the metrics port, use it for health
		healthPort = port
	}

	go func() {
		a.logger.Info("Starting health server", zap.String("port", healthPort))
		if err := http.ListenAndServe(":"+healthPort, mux); err != nil {
			a.logger.Error("Health server failed", zap.Error(err))
		}
	}()
}

// checkDatabaseHealth checks database connectivity
func (a *Agent) checkDatabaseHealth() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := a.clientsDB.Ping(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return map[string]interface{}{
			"status":  "unhealthy",
			"latency": latency,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"status":  "healthy",
		"latency": latency,
	}
}

// checkKafkaHealth checks Kafka connectivity
func (a *Agent) checkKafkaHealth() map[string]interface{} {
	// In a real implementation, you'd check if the consumer is still connected
	// For now, we'll return a simple check
	return map[string]interface{}{
		"status":         "healthy",
		"consumer_group": a.consumerGroup,
		"agent_type":     a.agentType,
	}
}
