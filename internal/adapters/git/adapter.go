// internal/adapters/git/adapter.go
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
func (a *GitAdapter) Run() error {
	a.logger.Info("Starting git adapter message processing",
		zap.String("topic", a.requestsTopic),
	)

	a.shutdownWg.Add(1)
	defer a.shutdownWg.Done()

	// Message processing loop
	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Context cancelled, stopping consumer")
			return a.ctx.Err()

		default:
			// Consume a message
			msg, err := a.consumer.Consume(a.ctx)
			if err != nil {
				if a.ctx.Err() != nil {
					return a.ctx.Err()
				}
				a.logger.Error("Error consuming message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			// Process the message
			a.processMessage(&msg)
		}
	}
}

// processMessage handles an individual Kafka message
func (a *GitAdapter) processMessage(msg *kafka.Message) {
	startTime := time.Now()

	a.logger.Debug("Processing message",
		zap.String("topic", msg.Topic),
		zap.Int64("offset", msg.Offset),
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
			zap.String("request_id", req.Headers.RequestID),
		)
		return
	}

	a.logger.Info("Handling request",
		zap.String("action", req.Body.Action),
		zap.String("request_id", req.Headers.RequestID),
		zap.String("correlation_id", req.Headers.CorrelationID),
		zap.String("responses_topic", responsesTopic),
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
		"correlation_id":          requestHeaders.CorrelationID,
		"orchestration_id":        requestHeaders.OrchestrationID,
		"request_id":              requestHeaders.RequestID,
		"parent_request_id":       requestHeaders.ParentRequestID,
		"parent_orchestration_id": requestHeaders.ParentOrchestrationID,

		// Response metadata
		"message_type": "response",
		"message_id":   fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"status":       "success",

		// Agent identification
		"from_agent":      a.adapterID.String(),
		"from_agent_type": "git-adapter",

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
		"correlation_id":          requestHeaders.CorrelationID,
		"orchestration_id":        requestHeaders.OrchestrationID,
		"request_id":              requestHeaders.RequestID,
		"parent_request_id":       requestHeaders.ParentRequestID,
		"parent_orchestration_id": requestHeaders.ParentOrchestrationID,

		// Error-specific metadata
		"message_type": "response",
		"status":       "error",
		"error":        err.Error(),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),

		// Agent identification
		"from_agent":      a.adapterID.String(),
		"from_agent_type": "git-adapter",

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
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
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

	go func() {
		if err := a.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Health server error", zap.Error(err))
		}
	}()
}
