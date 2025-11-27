// internal/adapters/git/adapter.go
package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// GitAdapter is the main git adapter service
type GitAdapter struct {
	ctx           context.Context
	cancel        context.CancelFunc
	cfg           *config.ServiceConfig
	logger        *zap.Logger
	consumer      *kafka.Consumer
	producer      kafka.Producer
	githubClient  *GitHubClient
	healthServer  *http.Server
	shutdownOnce  sync.Once
	shutdownWg    sync.WaitGroup
	requestsTopic string
	adapterID     uuid.UUID
}

// NewAdapter creates a new git adapter instance
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*GitAdapter, error) {
	// Create adapter context
	adapterCtx, cancel := context.WithCancel(ctx)

	// debug
	envVars := os.Environ()
	logger.Info("New Git Adapter - environment variables",
		zap.Strings("DEBUGaa: env vars", envVars),
	)

	// Determine request topic
	requestsTopic := os.Getenv("REQUESTS_TOPIC")
	if requestsTopic == "" {
		requestsTopic = "system.adapter.git.requests"
	}

	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = "git.adapter.group"
	}

	adapterID, _ := uuid.NewUUID()

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestsTopic, consumerGroup, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create Kafka producer
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Initialize GitHub client
	githubToken := os.Getenv("GITHUB_TOKEN")
	githubOrg := os.Getenv("GITHUB_ORG")
	githubAPIBase := os.Getenv("GITHUB_API_BASE")
	if githubAPIBase == "" {
		githubAPIBase = "https://api.github.com"
	}

	githubClient, err := NewGitHubClient(githubToken, githubOrg, githubAPIBase, logger)
	if err != nil {
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	adapter := &GitAdapter{
		ctx:           adapterCtx,
		cancel:        cancel,
		cfg:           cfg,
		logger:        logger.With(zap.String("component", "git-adapter")),
		consumer:      consumer,
		producer:      producer,
		githubClient:  githubClient,
		requestsTopic: requestsTopic,
		adapterID:     adapterID,
	}

	logger.Info("Git adapter initialized",
		zap.Strings("kafka_brokers", cfg.Infrastructure.KafkaBrokers),
		zap.String("request_topic", requestsTopic),
		zap.String("consumer_group", consumerGroup),
		zap.String("github_org", githubOrg),
		zap.String("adapter_id", adapterID.String()),
	)

	return adapter, nil
}

// Run starts the main message processing loop
// This method should ONLY return when actually shutting down
func (a *GitAdapter) Run() error {
	a.logger.Info("Starting git adapter message processing",
		zap.String("topic", a.requestsTopic),
	)

	a.shutdownWg.Add(1)
	defer a.shutdownWg.Done()

	// Keep track of consecutive errors for backoff
	consecutiveErrors := 0

	// Message processing loop - runs forever until shutdown
	for {
		// Check if we're shutting down
		select {
		case <-a.ctx.Done():
			a.logger.Info("Shutdown signal received, stopping message processing")
			return nil // Return nil, not an error - this is normal shutdown
		default:
			// Continue processing
		}

		// Try to consume a message
		// Use a child context with timeout to avoid blocking forever
		consumeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg, err := a.consumer.Consume(consumeCtx)
		cancel() // Always clean up the context

		// Handle the consume result
		if err != nil {
			// First check if WE are shutting down (not the consume context)
			select {
			case <-a.ctx.Done():
				a.logger.Debug("Adapter shutting down during consume")
				return nil // Normal shutdown, not an error
			default:
				// Not shutting down, handle the error
			}

			// Determine if this is a real error or just a timeout
			isRealError := false

			switch {
			case errors.Is(err, context.DeadlineExceeded):
				// Timeout - this is normal, not an error
				// Reset error counter since this isn't a real error
				consecutiveErrors = 0
				// Just continue to next iteration silently
				continue

			case errors.Is(err, context.Canceled):
				// Context was cancelled, check if we're shutting down
				if a.ctx.Err() != nil {
					return nil
				}
				// Otherwise just continue
				continue

			case strings.Contains(err.Error(), "EOF"):
				// Kafka EOF - can happen during rebalancing or shutdown
				// Log at debug level only
				a.logger.Debug("Kafka EOF received, continuing", zap.Error(err))
				continue

			case strings.Contains(strings.ToLower(err.Error()), "timeout"):
				// Another form of timeout
				consecutiveErrors = 0
				continue

			default:
				// This is a real error
				isRealError = true
				consecutiveErrors++
				a.logger.Warn("Error consuming message",
					zap.Error(err),
					zap.Int("consecutive_errors", consecutiveErrors),
				)
			}

			// Handle real errors with backoff
			if isRealError {
				// Implement exponential backoff for real errors
				backoffDuration := time.Duration(min(consecutiveErrors, 10)) * time.Second

				a.logger.Debug("Backing off after error",
					zap.Duration("backoff", backoffDuration),
					zap.Int("consecutive_errors", consecutiveErrors),
				)

				// Sleep with cancellation check
				select {
				case <-time.After(backoffDuration):
					// Continue after backoff
				case <-a.ctx.Done():
					// Shutdown during backoff
					return nil
				}

				// If we've had too many consecutive errors, log a warning
				if consecutiveErrors > 10 {
					a.logger.Error("Too many consecutive consume errors",
						zap.Int("count", consecutiveErrors),
						zap.Error(err),
					)
					// Reset counter to avoid overflow
					consecutiveErrors = 10
				}
			}

			// Continue to next iteration
			continue
		}

		// Successfully received a message
		consecutiveErrors = 0 // Reset error counter

		// Process the message in a safe way
		func() {
			defer func() {
				if r := recover(); r != nil {
					a.logger.Error("Panic while processing message",
						zap.Any("panic", r),
						zap.Stack("stack"),
					)
				}
			}()

			a.processMessage(&msg)
		}()
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Alternative simpler version if you prefer
func (a *GitAdapter) RunSimpleNotUsed() error {
	a.logger.Info("Starting git adapter (simple mode)")

	a.shutdownWg.Add(1)
	defer a.shutdownWg.Done()

	for {
		// Check shutdown
		if a.ctx.Err() != nil {
			return nil // Normal shutdown
		}

		// Try to get a message (with timeout)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg, err := a.consumer.Consume(ctx)
		cancel()

		if err != nil {
			// Only process if not shutting down
			if a.ctx.Err() == nil {
				// Only log if it's not a timeout
				if !errors.Is(err, context.DeadlineExceeded) &&
					!errors.Is(err, context.Canceled) &&
					!strings.Contains(err.Error(), "EOF") {
					a.logger.Warn("Consume error (continuing)", zap.Error(err))
					time.Sleep(time.Second)
				}
			} else {
				return nil // Shutting down
			}
			continue
		}

		// Process message
		a.processMessage(&msg)
	}
}

// processMessage handles an individual Kafka message
func (a *GitAdapter) processMessage(msg *kafka.Message) {
	startTime := time.Now()

	a.logger.Debug("Processing message",
		zap.String("topic", msg.Topic),
		zap.Int64("offset", msg.Offset),
		zap.Any("headers", msg.Headers),
		zap.Any("DEBUGaa: msg - looking for replytotopic", msg),
	)

	// Parse the request
	var req AdapterRequest
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		a.logger.Error("Failed to unmarshal request",
			zap.Error(err),
			zap.Int("message_size", len(msg.Value)),
		)
		return
	}

	// Validate request
	responsesTopic := req.Headers.ResponsesTopic
	if responsesTopic == "" {
		a.logger.Error("No responses_topic in headers",
			zap.Any("req.Headers", req.Headers),
		)
		return
	}

	a.logger.Info("In adapter processsMessage Handling request",
		zap.String("action", req.Body.Action),
		zap.String("request_id", req.Headers.RequestID),
		zap.String("correlation_id", req.Headers.CorrelationID),
		zap.String("responses_topic is this right?", responsesTopic),
		//zap.Any("DEBUGaa: request headers", req.Headers),
	)

	// Handle the request based on action and get response payload
	var responsePayload interface{}
	switch req.Body.Action {
	case "commit":
		responsePayload = a.handleCommitAction(req.Body.Data)
	case "create_repo":
		responsePayload = a.handleCreateRepoAction(req.Body.Data)
	case "delete_repo":
		responsePayload = a.handleDeleteRepoAction(req.Body.Data)
	default:
		a.logger.Error("Unknown action",
			zap.String("action", req.Body.Action),
		)
		responsePayload = map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("unknown action: %s", req.Body.Action),
		}
	}

	a.logger.Info("processMessage about to sendError or Success Response",
		zap.String("topic", responsesTopic),
	)

	// Send the response based on success/failure
	if respMap, ok := responsePayload.(map[string]interface{}); ok {
		if success, ok := respMap["success"].(bool); ok && !success {
			// It's an error response
			if errMsg, ok := respMap["error"].(string); ok {
				a.sendErrorResponse(responsesTopic, req.Headers, fmt.Errorf(errMsg))
			} else {
				a.sendErrorResponse(responsesTopic, req.Headers, fmt.Errorf("operation failed"))
			}
		} else {
			// It's a success response
			a.sendSuccessResponse(responsesTopic, req.Headers, responsePayload)
		}
	} else {
		// Default to sending as success
		a.sendSuccessResponse(responsesTopic, req.Headers, responsePayload)
	}

	// Log processing time
	a.logger.Info("Message processed",
		zap.String("request_id", req.Headers.RequestID),
		zap.Duration("processing_time", time.Since(startTime)),
	)
}

// handleCommitAction handles git commit requests - returns payload
func (a *GitAdapter) handleCommitAction(data json.RawMessage) interface{} {
	var commitData GitCommitData
	if err := json.Unmarshal(data, &commitData); err != nil {
		a.logger.Error("Failed to parse commit data", zap.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to parse commit data: %v", err),
		}
	}

	a.logger.Info("in HandleCommitAction",
		zap.Any("commitData", commitData),
	)

	// Create timeout context for GitHub operations
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	// Perform the commit
	repoURL, err := a.githubClient.CommitToRepo(ctx, commitData)
	if err != nil {
		a.logger.Error("Failed to commit to repo",
			zap.Error(err),
			zap.String("repo", commitData.RepoName),
		)
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	a.logger.Info("Successfully committed to repo",
		zap.String("repo", commitData.RepoName),
		zap.String("url", repoURL),
		zap.Int("files", len(commitData.Files)),
	)

	// Return success payload
	return map[string]interface{}{
		"success":        true,
		"repo_url":       repoURL,
		"repo_name":      commitData.RepoName,
		"files_count":    len(commitData.Files),
		"commit_message": commitData.CommitMessage,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

// handleCreateRepoAction handles repository creation requests - returns payload
func (a *GitAdapter) handleCreateRepoAction(data json.RawMessage) interface{} {
	var repoData struct {
		RepoName    string `json:"repo_name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		AutoInit    bool   `json:"auto_init"`
	}

	if err := json.Unmarshal(data, &repoData); err != nil {
		a.logger.Error("Failed to parse repo creation data", zap.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to parse repo data: %v", err),
		}
	}

	// Set default for auto_init
	if !repoData.AutoInit {
		repoData.AutoInit = true
	}

	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	// Create initial README content
	readmeContent := fmt.Sprintf("# %s\n\n%s\n", repoData.RepoName, repoData.Description)
	if repoData.Description == "" {
		readmeContent = fmt.Sprintf("# %s\n\nRepository created by Git Adapter\n", repoData.RepoName)
	}

	// Use commit function to create repo with initial content
	commitData := GitCommitData{
		RepoName: repoData.RepoName,
		Files: map[string]string{
			"README.md": readmeContent,
		},
		CommitMessage: "Initial commit",
	}

	repoURL, err := a.githubClient.CommitToRepo(ctx, commitData)
	if err != nil {
		a.logger.Error("Failed to create repo",
			zap.Error(err),
			zap.String("repo", repoData.RepoName),
		)
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	a.logger.Info("Successfully created repo",
		zap.String("repo", repoData.RepoName),
		zap.String("url", repoURL),
	)

	return map[string]interface{}{
		"success":     true,
		"repo_url":    repoURL,
		"repo_name":   repoData.RepoName,
		"description": repoData.Description,
		"private":     repoData.Private,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
}

// handleDeleteRepoAction handles repository deletion requests - returns payload
func (a *GitAdapter) handleDeleteRepoAction(data json.RawMessage) interface{} {
	// This would need to be implemented in GitHubClient
	// For now, return not implemented
	return map[string]interface{}{
		"success": false,
		"error":   "delete_repo action not yet implemented",
	}
}

// sendSuccessResponse sends a successful response back via Kafka
func (a *GitAdapter) sendSuccessResponse(topic string, requestHeaders AdapterHeaders, data interface{}) {
	// Build response headers as map[string]string for ProduceWithValidation
	responseHeaders := map[string]string{
		// Core orchestration context
		"correlation_id":            requestHeaders.CorrelationID,
		"orchestration_id":          requestHeaders.OrchestrationID,
		"client_id":                 requestHeaders.ClientID,
		"request_id":                requestHeaders.RequestID,
		"parent_request_id":         requestHeaders.ParentRequestID,
		"parent_orchestration_id":   requestHeaders.ParentOrchestrationID,
		"in_response_to_request_id": requestHeaders.RequestID,
		"in_response_to_step_name":  requestHeaders.StepName,

		// Response metadata
		"message_type": "response",
		"message_id":   uuid.New().String(),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"status":       "success",

		// Agent identification
		"sender_agent_id":   a.adapterID.String(),
		"sender_agent_type": "git-adapter",
		"sender_pod_name":   os.Getenv("HOSTNAME"),

		// Resource tracking
		"fuel_used": "10",
	}

	// Add step tracking if present
	if requestHeaders.StepID != "" {
		responseHeaders["step_id"] = requestHeaders.StepID
		responseHeaders["step_name"] = requestHeaders.StepName
		responseHeaders["in_response_to"] = requestHeaders.StepID
	}

	responseBody := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	responseMsg := map[string]interface{}{
		"headers": responseHeaders,
		"body":    responseBody,
	}

	// Marshal the message
	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		a.logger.Error("Failed to marshal success response", zap.Error(err))
		return
	}

	// Use correlation ID as key
	key := []byte(requestHeaders.CorrelationID)

	// Call ProduceWithValidation with headers as map[string]string
	err = a.producer.ProduceWithValidation(a.ctx, topic, responseHeaders, key, responseBytes)
	if err != nil {
		a.logger.Error("Failed to send success response", zap.Error(err))
		return
	}

	a.logger.Info("Success response sent",
		zap.String("topic", topic),
		zap.String("request_id", requestHeaders.RequestID),
	)
}

// sendErrorResponse sends an error response back via Kafka
func (a *GitAdapter) sendErrorResponse(topic string, requestHeaders AdapterHeaders, err error) {
	// Build response headers as map[string]string for ProduceWithValidation
	responseHeaders := map[string]string{
		// Core orchestration context
		"correlation_id":            requestHeaders.CorrelationID,
		"orchestration_id":          requestHeaders.OrchestrationID,
		"client_id":                 requestHeaders.ClientID,
		"request_id":                requestHeaders.RequestID,
		"parent_request_id":         requestHeaders.ParentRequestID,
		"parent_orchestration_id":   requestHeaders.ParentOrchestrationID,
		"in_response_to_request_id": requestHeaders.RequestID,
		"in_response_to_step_name":  requestHeaders.StepName,

		// Error-specific metadata
		"message_type": "response",
		"status":       "error",
		"error":        err.Error(),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),

		// Agent identification
		"sender_agent_id":   a.adapterID.String(),
		"sender_agent_type": "git-adapter",
		"sender_pod_name":   os.Getenv("HOSTNAME"),

		// Resource tracking
		"fuel_used": "5",
	}

	// Add step tracking if present
	if requestHeaders.StepID != "" {
		responseHeaders["step_id"] = requestHeaders.StepID
		responseHeaders["step_name"] = requestHeaders.StepName
		responseHeaders["in_response_to"] = requestHeaders.StepID
	}

	responseBody := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"message":     err.Error(),
			"type":        "GitAdapterError",
			"recoverable": true,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		},
	}

	responseMsg := map[string]interface{}{
		"headers": responseHeaders,
		"body":    responseBody,
	}

	responseBytes, marshalErr := json.Marshal(responseMsg)
	if marshalErr != nil {
		a.logger.Error("Failed to marshal error response", zap.Error(marshalErr))
		return
	}

	a.logger.Info("sendErrorResponse",
		zap.String("topic", topic),
		zap.String("request_id", requestHeaders.RequestID),
	)
	key := []byte(requestHeaders.CorrelationID)

	// Call ProduceWithValidation with headers as map[string]string
	produceErr := a.producer.ProduceWithValidation(a.ctx, topic, responseHeaders, key, responseBytes)
	if produceErr != nil {
		a.logger.Error("Failed to send error response",
			zap.Error(produceErr),
			zap.String("original_error", err.Error()),
		)
	}

	a.logger.Info("Error response sent",
		zap.String("topic", topic),
		zap.String("request_id", requestHeaders.RequestID),
		zap.String("error", err.Error()),
	)
}

// Shutdown gracefully shuts down the adapter
func (a *GitAdapter) Shutdown() {
	a.shutdownOnce.Do(func() {
		a.logger.Info("Shutting down git adapter")

		// Cancel context
		a.cancel()

		// Close Kafka connections
		if a.consumer != nil {
			if err := a.consumer.Close(); err != nil {
				a.logger.Error("Error closing consumer", zap.Error(err))
			}
		}

		if a.producer != nil {
			if err := a.producer.Close(); err != nil {
				a.logger.Error("Error closing producer", zap.Error(err))
			}
		}

		// Shutdown health server
		if a.healthServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.healthServer.Shutdown(ctx); err != nil {
				a.logger.Error("Failed to shutdown health server", zap.Error(err))
			}
		}

		// Wait for goroutines to finish
		done := make(chan struct{})
		go func() {
			a.shutdownWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			a.logger.Info("All goroutines finished")
		case <-time.After(5 * time.Second):
			a.logger.Warn("Shutdown timeout waiting for goroutines")
		}

		a.logger.Info("Git adapter shutdown complete")
	})
}

// StartHealthServer starts the health check HTTP server
func (a *GitAdapter) StartHealthServer(port string) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Ready check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check if consumer is connected
		if a.consumer != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready"}`))
			a.logger.Debug("Readiness probe hit", zap.String("path", "/ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
			a.logger.Debug("Readiness probe failed", zap.String("path", "/ready"))

		}
	})

	// Live check endpoint
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		// Check if consumer is connected
		if a.consumer != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"live"}`))
			a.logger.Debug("Liveness probe hit", zap.String("path", "/live"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not live"}`))
			a.logger.Debug("Liveness probe failed", zap.String("path", "/live"))
		}
	})

	// Metrics endpoint (placeholder)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add Prometheus metrics
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# HELP git_adapter_info git adapter information\n"))
		w.Write([]byte("# TYPE git_adapter_info gauge\n"))
		w.Write([]byte("git_adapter_info{version=\"1.0.0\"} 1\n"))
	})

	a.healthServer = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	a.logger.Info("Health server starting",
		zap.String("port", port),
		zap.Strings("endpoints", []string{"/live", "/ready", "/health"}),
	)

	go func() {
		if err := a.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Health server error", zap.Error(err))
		}
	}()
}
