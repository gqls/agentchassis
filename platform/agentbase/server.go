// FILE: platform/agentbase/server.go
package agentbase

import (
	"context"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"go.uber.org/zap"
	"time"
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

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Agent server shutting down")
			return nil
		default:
			msg, err := s.consumer.FetchMessage(s.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				s.logger.Error("Failed to fetch message", zap.Error(err))
				observability.SystemErrors.WithLabelValues(s.agentType, "fetch_message").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			// Record metric
			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, s.consumerGroup).Inc()

			// Process asynchronously
			go s.processMessage(msg)
		}
	}
}

func (s *AgentServer) processMessage(msg kafka.Message) {
	if err := s.processor.ProcessMessage(s.ctx, msg); err != nil {
		s.logger.Error("Failed to process message", zap.Error(err))
	}

	// Always commit
	if err := s.consumer.CommitMessages(context.Background(), msg); err != nil {
		s.logger.Error("Failed to commit message", zap.Error(err))
		observability.SystemErrors.WithLabelValues(s.agentType, "commit_message").Inc()
	}
}

// Shutdown gracefully shuts down the server
func (s *AgentServer) Shutdown() error {
	return s.consumer.Close()
}
