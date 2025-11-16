package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gqls/agentchassis/platform/config" // Assuming this path
	"go.uber.org/zap"
)

// GitAdapterHandler implements sarama.ConsumerGroupHandler
// It is the core Kafka consumer for the git-adapter service.
type GitAdapterHandler struct {
	log       *zap.Logger
	producer  sarama.SyncProducer
	gitClient *GitHubClient // The client that does the actual work
}

// NewAdapterHandler creates a new Kafka consumer handler
func NewAdapterHandler(cfg *config.Config, log *zap.Logger) (*GitAdapterHandler, error) {
	// Sarama Producer config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0 // Or your Kafka version
	saramaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync producer: %w", err)
	}

	// Create the GitHub Client, passing in config
	gitClient, err := NewGitHubClient(
		cfg.Github.Token,
		cfg.Github.Org,
		cfg.Github.APIBase,
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github client: %w", err)
	}

	return &GitAdapterHandler{
		log:       log,
		producer:  producer,
		gitClient: gitClient,
	}, nil
}

// Run is a helper to start the consumer group (called by main.go)
func (h *GitAdapterHandler) Run(ctx context.Context, cfg *config.Config) error {
	h.log.Info("Starting consumer group", zap.String("group", cfg.Kafka.ConsumerGroup), zap.String("topic", cfg.Kafka.Topics[0]))

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	consumer, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, saramaConfig)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	defer consumer.Close()
	defer h.producer.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := consumer.Consume(ctx, cfg.Kafka.Topics, h); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					h.log.Info("Consumer group closed")
					return
				}
				h.log.Error("Error from consumer", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-ctx.Done()
	h.log.Info("Context cancelled, shutting down adapter")
	wg.Wait()
	return nil
}

// --- Sarama ConsumerGroupHandler Implementation ---

func (h *GitAdapterHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *GitAdapterHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim is the core message processing loop
func (h *GitAdapterHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				h.log.Info("Message channel closed")
				return nil
			}

			h.log.Info("Received message",
				zap.String("topic", message.Topic),
				zap.Int64("offset", message.Offset),
			)

			var req AdapterRequest
			if err := json.Unmarshal(message.Value, &req); err != nil {
				h.log.Error("Failed to unmarshal request", zap.Error(err), zap.ByteString("value", message.Value))
				session.MarkMessage(message, "")
				continue
			}

			responsesTopic := req.Headers.ResponsesTopic
			if responsesTopic == "" {
				h.log.Error("No responses_topic in headers, discarding message", zap.String("request_id", req.Headers.RequestID))
				session.MarkMessage(message, "")
				continue
			}

			// Handle the request
			var responsePayload interface{}
			switch req.Body.Action {
			case "commit":
				responsePayload = h.handleCommit(session.Context(), req)
			default:
				responsePayload = map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("unknown action: %s", req.Body.Action),
				}
			}

			// Send the response
			if err := h.sendResponse(responsesTopic, req.Headers, responsePayload); err != nil {
				h.log.Error("Failed to send response", zap.Error(err))
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleCommit routes the message to the git client
func (h *GitAdapterHandler) handleCommit(ctx context.Context, req AdapterRequest) interface{} {
	var commitData GitCommitData
	if err := json.Unmarshal(req.Body.Data, &commitData); err != nil {
		h.log.Error("Failed to unmarshal commit data", zap.Error(err))
		return map[string]interface{}{"success": false, "error": "failed to parse commit data"}
	}

	repoURL, err := h.gitClient.CommitToRepo(ctx, commitData)
	if err != nil {
		h.log.Error("Failed to process git commit", zap.Error(err), zap.String("request_id", req.Headers.RequestID))
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

// sendResponse sends a JSON payload back to the specified topic
func (h *GitAdapterHandler) sendResponse(topic string, headers AdapterHeaders, payload interface{}) error {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal response payload: %w", err)
	}

	responseMsg := map[string]interface{}{
		"headers": map[string]string{
			"correlation_id":   headers.CorrelationID,
			"orchestration_id": headers.OrchestrationID,
			"request_id":       headers.RequestID,
			"message_type":     "response",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		},
		"body": json.RawMessage(responseBody),
	}

	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal full response: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(headers.RequestID),
		// **FIX:** Use sarama.ByteEncoder, which is the correct type for the Value field.
		Value: sarama.ByteEncoder(responseBytes),
	}

	_, _, err = h.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send response message: %w", err)
	}

	h.log.Info("Sent response", zap.String("topic", topic), zap.String("request_id", headers.RequestID))
	return nil
}
