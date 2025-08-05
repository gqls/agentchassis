// FILE: platform/agentbase/client.go
package agentbase

import (
	"context"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"go.uber.org/zap"
	"time"
)

// AgentClient handles responses from other agents
type AgentClient struct {
	ctx              context.Context
	logger           *zap.Logger
	responseConsumer *kafka.Consumer
	processor        *messaging.MessageProcessor
	consumerGroup    string
	agentType        string
}

// NewAgentClient creates a new agent client
func NewAgentClient(
	ctx context.Context,
	logger *zap.Logger,
	responseConsumer *kafka.Consumer,
	processor *messaging.MessageProcessor,
	consumerGroup string,
	agentType string,
) *AgentClient {
	return &AgentClient{
		ctx:              ctx,
		logger:           logger,
		responseConsumer: responseConsumer,
		processor:        processor,
		consumerGroup:    consumerGroup,
		agentType:        agentType,
	}
}

// Run starts processing responses
func (c *AgentClient) Run() error {
	c.logger.Info("Starting agent client (response handler)",
		zap.String("agent_type", c.agentType),
		zap.String("response_group", c.consumerGroup+"-responses"))

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Agent client shutting down")
			return nil
		default:
			msg, err := c.responseConsumer.FetchMessage(c.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				c.logger.Error("Failed to fetch response message", zap.Error(err))
				observability.SystemErrors.WithLabelValues(c.agentType, "fetch_response").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			// Record metric
			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, c.consumerGroup+"-responses").Inc()

			// Process response asynchronously
			go c.processResponse(msg)
		}
	}
}

func (c *AgentClient) processResponse(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)

	c.logger.Info("Processing response",
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("causation_id", headers["causation_id"]))

	// Route to processor which will handle orchestration responses
	if err := c.processor.ProcessResponse(c.ctx, msg); err != nil {
		c.logger.Error("Failed to process response", zap.Error(err))
		observability.SystemErrors.WithLabelValues(c.agentType, "process_response").Inc()
	}

	// Always commit
	if err := c.responseConsumer.CommitMessages(context.Background(), msg); err != nil {
		c.logger.Error("Failed to commit response message", zap.Error(err))
		observability.SystemErrors.WithLabelValues(c.agentType, "commit_response").Inc()
	}
}

// Shutdown gracefully shuts down the client
func (c *AgentClient) Shutdown() error {
	return c.responseConsumer.Close()
}
