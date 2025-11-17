package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// Config holds the adapter configuration
type Config struct {
	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroup    string
	GitHubToken   string
	GitHubOrg     string // Optional: to create repos in an organization
	GitHubAPIBase string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal("KAFKA_BROKERS env var not set")
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "system.adapter.git.requests"
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN env var not set")
	}

	apiBase := os.Getenv("GITHUB_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}

	return Config{
		KafkaBrokers:  strings.Split(brokers, ","),
		KafkaTopic:    topic,
		KafkaGroup:    "git-adapter-group",
		GitHubToken:   token,
		GitHubOrg:     os.Getenv("GITHUB_ORG"), // e.g., "my-company"
		GitHubAPIBase: apiBase,
	}
}

// GitAdapterHandler handles Git adapter messages from Kafka
type GitAdapterHandler struct {
	cfg          Config
	logger       *zap.Logger
	githubClient *GitHubClient
	producer     sarama.SyncProducer
}

// NewGitAdapterHandler creates a new handler instance
func NewGitAdapterHandler(cfg Config, logger *zap.Logger, producer sarama.SyncProducer) (*GitAdapterHandler, error) {
	// Initialize GitHub client
	githubClient, err := NewGitHubClient(cfg.GitHubToken, cfg.GitHubOrg, cfg.GitHubAPIBase, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return &GitAdapterHandler{
		cfg:          cfg,
		logger:       logger,
		githubClient: githubClient,
		producer:     producer,
	}, nil
}

// Setup is called at the beginning of a new session, before ConsumeClaim
func (h *GitAdapterHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called at the end of a session, once all ConsumeClaim goroutines have exited
func (h *GitAdapterHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim processes messages from Kafka
func (h *GitAdapterHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.logger.Info("Received message",
			zap.String("topic", message.Topic),
			zap.Int32("partition", message.Partition),
			zap.Int64("offset", message.Offset),
		)

		// Parse the request
		var req AdapterRequest
		if err := json.Unmarshal(message.Value, &req); err != nil {
			h.logger.Error("Failed to unmarshal request",
				zap.Error(err),
				zap.ByteString("value", message.Value))
			session.MarkMessage(message, "")
			continue
		}

		// Get response topic from headers
		responsesTopic := req.Headers.ResponsesTopic
		if responsesTopic == "" {
			h.logger.Error("No responses_topic in headers, discarding message",
				zap.String("request_id", req.Headers.RequestID))
			session.MarkMessage(message, "")
			continue
		}

		// Handle the request
		responsePayload := h.handleRequest(req)

		// Send the response
		if err := h.sendResponse(responsesTopic, req.Headers, responsePayload); err != nil {
			h.logger.Error("Failed to send response", zap.Error(err))
		}

		// Mark message as processed
		session.MarkMessage(message, "")
	}
	return nil
}

// handleRequest routes the message to the appropriate handler
func (h *GitAdapterHandler) handleRequest(req AdapterRequest) interface{} {
	h.logger.Info("Handling request",
		zap.String("action", req.Body.Action),
		zap.String("request_id", req.Headers.RequestID),
	)

	switch req.Body.Action {
	case "commit":
		return h.handleCommitAction(req.Body.Data)
	case "create_repo":
		return h.handleCreateRepoAction(req.Body.Data)
	default:
		h.logger.Error("Unknown action",
			zap.String("action", req.Body.Action),
			zap.String("request_id", req.Headers.RequestID))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("unknown action: %s", req.Body.Action),
		}
	}
}

// handleCommitAction handles the git commit action
func (h *GitAdapterHandler) handleCommitAction(data json.RawMessage) interface{} {
	var commitData GitCommitData
	if err := json.Unmarshal(data, &commitData); err != nil {
		h.logger.Error("Failed to parse commit data", zap.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to parse commit data: %v", err),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repoURL, err := h.githubClient.CommitToRepo(ctx, commitData)
	if err != nil {
		h.logger.Error("Failed to commit to repo",
			zap.Error(err),
			zap.String("repo", commitData.RepoName))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	h.logger.Info("Successfully committed to repo",
		zap.String("repo", commitData.RepoName),
		zap.String("url", repoURL),
		zap.Int("files", len(commitData.Files)))

	return map[string]interface{}{
		"success":  true,
		"repo_url": repoURL,
	}
}

// handleCreateRepoAction handles repository creation
func (h *GitAdapterHandler) handleCreateRepoAction(data json.RawMessage) interface{} {
	var repoData struct {
		RepoName    string `json:"repo_name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
	}

	if err := json.Unmarshal(data, &repoData); err != nil {
		h.logger.Error("Failed to parse repo creation data", zap.Error(err))
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to parse repo data: %v", err),
		}
	}

	// For now, we'll use the commit function which creates the repo if needed
	// You could expand this to have a dedicated CreateRepo method in GitHubClient
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create with an initial README
	commitData := GitCommitData{
		RepoName: repoData.RepoName,
		Files: map[string]string{
			"README.md": fmt.Sprintf("# %s\n\n%s", repoData.RepoName, repoData.Description),
		},
		CommitMessage: "Initial commit",
	}

	repoURL, err := h.githubClient.CommitToRepo(ctx, commitData)
	if err != nil {
		h.logger.Error("Failed to create repo",
			zap.Error(err),
			zap.String("repo", repoData.RepoName))
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":  true,
		"repo_url": repoURL,
	}
}

// sendResponse sends the response back through Kafka
func (h *GitAdapterHandler) sendResponse(topic string, headers AdapterHeaders, payload interface{}) error {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal response payload: %w", err)
	}

	// Create response envelope matching your agent's expected format
	responseMsg := map[string]interface{}{
		"headers": map[string]string{
			"correlation_id":   headers.CorrelationID,
			"orchestration_id": headers.OrchestrationID,
			"request_id":       headers.RequestID,
			"message_type":     "response",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
			"from_agent":       "git-adapter",
		},
		"body": json.RawMessage(responseBody),
	}

	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal full response: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(headers.RequestID),
		Value: sarama.ByteEncoder(responseBytes), // Fixed: ByteEncoder instead of ByteString
	}

	partition, offset, err := h.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send response message: %w", err)
	}

	h.logger.Info("Sent response",
		zap.String("topic", topic),
		zap.String("request_id", headers.RequestID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// Run starts the git adapter service
func Run() error {
	// Load configuration
	cfg := LoadConfig()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting git-adapter service",
		zap.Strings("brokers", cfg.KafkaBrokers),
		zap.String("topic", cfg.KafkaTopic),
		zap.String("group", cfg.KafkaGroup),
	)

	// Configure Sarama
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	// Create producer for responses
	producer, err := sarama.NewSyncProducer(cfg.KafkaBrokers, saramaConfig)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	defer producer.Close()

	// Create handler
	handler, err := NewGitAdapterHandler(cfg, logger, producer)
	if err != nil {
		logger.Fatal("Failed to create handler", zap.Error(err))
	}

	// Create consumer group
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroup, saramaConfig)
	if err != nil {
		logger.Fatal("Failed to create consumer group", zap.Error(err))
	}

	// Start consuming
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Consume messages
			if err := client.Consume(ctx, []string{cfg.KafkaTopic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					logger.Info("Consumer group closed")
					return
				}
				logger.Error("Error from consumer", zap.Error(err))
			}
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Handle graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	case sig := <-sigterm:
		logger.Info("Termination signal received", zap.String("signal", sig.String()))
	}

	// Cleanup
	cancel()
	wg.Wait()

	if err := client.Close(); err != nil {
		logger.Error("Failed to close consumer client", zap.Error(err))
	}

	logger.Info("Git adapter service stopped")
	return nil
}
