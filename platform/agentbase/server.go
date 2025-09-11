// FILE: platform/agentbase/server.go
package agentbase

import (
	"context"
	"time"

	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// AgentServer handles incoming work requests
type AgentServer struct {
	ctx           context.Context
	logger        *zap.Logger
	consumer      *kafka.Consumer
	processor     *messaging.MessageProcessor
	consumerGroup string
	agentType     string
}

// NewAgentServer creates a new agent server
func NewAgentServer(
	ctx context.Context,
	logger *zap.Logger,
	consumer *kafka.Consumer,
	processor *messaging.MessageProcessor,
	consumerGroup string,
	agentType string,
) *AgentServer {
	return &AgentServer{
		ctx:           ctx,
		logger:        logger,
		consumer:      consumer,
		processor:     processor,
		consumerGroup: consumerGroup,
		agentType:     agentType,
	}
}

// Run starts processing incoming requests
func (s *AgentServer) Run() error {
	s.logger.Info("Starting agent server",
		zap.String("agent_type", s.agentType),
		zap.String("consumer_group", s.consumerGroup))

	s.logger.Info("Entering message consumption loop")

	messageCount := 0
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Agent server shutting down",
				zap.Int("messages_processed", messageCount))
			return nil
		default:
			s.logger.Debug("About to fetch message", zap.Int("message_count", messageCount))

			msg, err := s.consumer.FetchMessage(s.ctx)
			if err != nil {
				if err == context.Canceled {
					s.logger.Debug("Context canceled during fetch")
					continue
				}
				s.logger.Error("FetchMessage returned error",
					zap.Error(err),
					zap.Int("message_count", messageCount))
				observability.SystemErrors.WithLabelValues(s.agentType, "fetch_message").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			messageCount++
			s.logger.Info("Message received",
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
				zap.Int("message_count", messageCount),
				zap.Int("message_size", len(msg.Value)))

			// Log headers for debugging
			headers := kafka.HeadersToMap(msg.Headers)
			s.logger.Info("Message headers",
				zap.String("correlation_id", headers["correlation_id"]),
				zap.String("request_id", headers["request_id"]),
				zap.String("client_id", headers["client_id"]),
				zap.String("agent_instance_id", headers["agent_instance_id"]),
				zap.String("fuel_budget", headers["fuel_budget"]))

			// Record metric
			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, s.consumerGroup).Inc()

			// Process asynchronously
			s.processMessage(msg, messageCount)
		}
	}
}

func (s *AgentServer) processMessage(msg kafka.Message, msgNum int) {
	startTime := time.Now()
	headers := kafka.HeadersToMap(msg.Headers)

	s.logger.Info("Processing message started",
		zap.Int("message_num", msgNum),
		zap.String("correlation_id", headers["correlation_id"]))

	if err := s.processor.ProcessMessage(s.ctx, msg); err != nil {
		s.logger.Error("Failed to process message",
			zap.Error(err),
			zap.Int("message_num", msgNum),
			zap.String("correlation_id", headers["correlation_id"]),
			zap.Duration("processing_time", time.Since(startTime)))
	} else {
		s.logger.Info("Message processed successfully",
			zap.Int("message_num", msgNum),
			zap.String("correlation_id", headers["correlation_id"]),
			zap.Duration("processing_time", time.Since(startTime)))
	}

	// Always commit
	if err := s.consumer.CommitMessages(context.Background(), msg); err != nil {
		s.logger.Error("Failed to commit message",
			zap.Error(err),
			zap.Int("message_num", msgNum))
		observability.SystemErrors.WithLabelValues(s.agentType, "commit_message").Inc()
	} else {
		s.logger.Debug("Message committed successfully",
			zap.Int("message_num", msgNum),
			zap.String("correlation_id", headers["correlation_id"]))
	}
}

// Shutdown gracefully shuts down the server
func (s *AgentServer) Shutdown() error {
	return s.consumer.Close()
}
