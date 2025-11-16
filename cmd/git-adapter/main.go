package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gqls/agentchassis/platform/config"
	"go.uber.org/zap"
)

// Adapter is the main service struct
type Adapter struct {
	cfg        *config.Config
	log        *zap.Logger
	consumer   sarama.ConsumerGroup
	producer   sarama.SyncProducer
	gitClient  *GitHubClient
	kafkaTopic string
	kafkaGroup string
}

// NewAdapter creates the git-adapter service
func NewAdapter(cfg *config.Config, log *zap.Logger) (*Adapter, error) {
	// Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0 // Or your Kafka version
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Producer.Return.Successes = true

	consumer, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync producer: %w", err)
	}

	gitClient, err := NewGitHubClient(cfg.Github.Token, cfg.Github.Org, cfg.Github.APIBase, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create github client: %w", err)
	}

	return &Adapter{
		cfg:        cfg,
		log:        log,
		consumer:   consumer,
		producer:   producer,
		gitClient:  gitClient,
		kafkaTopic: cfg.Kafka.Topics[0], // Assuming first topic is the one we listen to
		kafkaGroup: cfg.Kafka.ConsumerGroup,
	}, nil
}

// Run starts the Kafka consumer loop
func (a *Adapter) Run(ctx context.Context) error {
	a.log.Info("Adapter is running, listening for messages...")
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop
			if err := a.consumer.Consume(ctx, []string{a.kafkaTopic}, a); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					a.log.Info("Consumer group closed")
					return
				}
				a.log.Error("Error from consumer", zap.Error(err))
			}
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-ctx.Done()
	a.log.Info("Context cancelled, shutting down adapter run loop")
	wg.Wait()
	if err := a.consumer.Close(); err != nil {
		a.log.Error("Error closing consumer", zap.Error(err))
	}
	if err := a.producer.Close(); err != nil {
		a.log.Error("Error closing producer", zap.Error(err))
	}
	return nil
}

// --- Sarama ConsumerGroupHandler Implementation ---

func (a *Adapter) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (a *Adapter) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim is the core message processing loop
func (a *Adapter) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				a.log.Info("Message channel closed")
				return nil
			}

			a.log.Info("Received message",
				zap.String("topic", message.Topic),
				zap.Int64("offset", message.Offset),
			)

			// 1. Deserialize the *exact* request your agent sends
			var req AdapterRequest
			if err := json.Unmarshal(message.Value, &req); err != nil {
				a.log.Error("Failed to unmarshal request", zap.Error(err), zap.ByteString("value", message.Value))
				session.MarkMessage(message, "") // Mark as processed
				continue
			}

			// 2. Get response topic from headers
			responsesTopic := req.Headers.ResponsesTopic
			if responsesTopic == "" {
				a.log.Error("No responses_topic in headers, discarding message", zap.String("request_id", req.Headers.RequestID))
				session.MarkMessage(message, "")
				continue
			}

			// 3. Handle the request
			var responsePayload interface{}
			switch req.Body.Action {
			case "commit":
				responsePayload = a.handleCommit(req)
			default:
				responsePayload = map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("unknown action: %s", req.Body.Action),
				}
			}

			// 4. Send the response
			if err := a.sendResponse(responsesTopic, req.Headers, responsePayload); err != nil {
				a.log.Error("Failed to send response", zap.Error(err))
			}

			// 5. Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleCommit routes the message to the git client
func (a *Adapter) handleCommit(req AdapterRequest) interface{} {
	var commitData GitCommitData
	if err := json.Unmarshal(req.Body.Data, &commitData); err != nil {
		a.log.Error("Failed to unmarshal commit data", zap.Error(err))
		return map[string]interface{}{"success": false, "error": "failed to parse commit data"}
	}

	repoURL, err := a.gitClient.CommitToRepo(context.Background(), commitData)
	if err != nil {
		a.log.Error("Failed to process git commit", zap.Error(err), zap.String("request_id", req.Headers.RequestID))
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
func (a *Adapter) sendResponse(topic string, headers AdapterHeaders, payload interface{}) error {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal response payload: %w", err)
	}

	// Create a minimal response envelope
	responseMsg := map[string]interface{}{
		"headers": map[string]string{
			"correlation_id":   headers.CorrelationID,
			"orchestration_id": headers.OrchestrationID,
			"request_id":       headers.RequestID,
			"message_type":     "response",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		},
		"body": json.RawMessage(responseBody), // Embed the payload as raw JSON
	}

	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal full response: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(headers.RequestID), // Use RequestID as key
		Value: sarama.ByteString(responseBytes),
	}

	_, _, err = a.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send response message: %w", err)
	}

	a.log.Info("Sent response", zap.String("topic", topic), zap.String("request_id", headers.RequestID))
	return nil
}
