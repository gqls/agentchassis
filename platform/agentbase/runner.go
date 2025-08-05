// FILE: platform/agentbase/runner.go (updated)
package agentbase

import (
	"context"
	"fmt"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"go.uber.org/zap"
	"time"
)

// MessageRunner handles the message processing loop
type MessageRunner struct {
	ctx    context.Context
	logger *zap.Logger
	// consumer listens for e.g. new work requests from clients or other agents
	// listens on system.agent.generic.process topic
	consumer *kafka.Consumer
	// responseConsumer listens for agents responses from workflow execution
	// listens on system.responses.generic
	responseConsumer *kafka.Consumer
	processor        *messaging.MessageProcessor
	consumerGroup    string
	agentType        string
}

// NewMessageRunner creates a new message runner
func NewMessageRunner(
	ctx context.Context,
	logger *zap.Logger,
	consumer *kafka.Consumer,
	responseConsumer *kafka.Consumer, // ADD: Response consumer parameter
	processor *messaging.MessageProcessor,
	consumerGroup string,
	agentType string,
) *MessageRunner {
	return &MessageRunner{
		ctx:              ctx,
		logger:           logger,
		consumer:         consumer,
		responseConsumer: responseConsumer,
		processor:        processor,
		consumerGroup:    consumerGroup,
		agentType:        agentType,
	}
}

// Run starts the message processing loop
func (r *MessageRunner) Run() error {
	r.logger.Info("Starting message runner", zap.String("agent_type", r.agentType))

	// Start response handler in a separate goroutine
	go r.handleResponses()

	// Main message processing loop
	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("Message runner shutting down")
			return nil
		default:
			msg, err := r.consumer.FetchMessage(r.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				r.logger.Error("Failed to fetch message", zap.Error(err))
				observability.SystemErrors.WithLabelValues(r.agentType, "fetch_message").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			// Record metric
			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, r.consumerGroup).Inc()

			// Process asynchronously
			go r.processMessage(msg)
		}
	}
}

// handleResponses processes response messages
func (r *MessageRunner) handleResponses() {
	r.logger.Info("Starting response handler",
		zap.String("agent_type", r.agentType),
		zap.String("topic", fmt.Sprintf("system.responses.%s", r.agentType)))

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("Response handler shutting down")
			return
		default:
			msg, err := r.responseConsumer.FetchMessage(r.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				r.logger.Error("Failed to fetch response message", zap.Error(err))
				observability.SystemErrors.WithLabelValues(r.agentType, "fetch_response").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			// Record metric
			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, r.consumerGroup+"-responses").Inc()

			// Process response asynchronously
			go r.processResponse(msg)
		}
	}
}

// processResponse handles a response message
func (r *MessageRunner) processResponse(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)

	r.logger.Info("Processing response",
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("causation_id", headers["causation_id"]))

	// Route to processor which will handle orchestration responses
	if err := r.processor.ProcessResponse(r.ctx, msg); err != nil {
		r.logger.Error("Failed to process response", zap.Error(err))
		observability.SystemErrors.WithLabelValues(r.agentType, "process_response").Inc()
	}

	// Always commit
	if err := r.responseConsumer.CommitMessages(context.Background(), msg); err != nil {
		r.logger.Error("Failed to commit response message", zap.Error(err))
		observability.SystemErrors.WithLabelValues(r.agentType, "commit_response").Inc()
	}
}

func (r *MessageRunner) processMessage(msg kafka.Message) {
	if err := r.processor.ProcessMessage(r.ctx, msg); err != nil {
		r.logger.Error("Failed to process message", zap.Error(err))
	}

	// Always commit
	if err := r.consumer.CommitMessages(context.Background(), msg); err != nil {
		r.logger.Error("Failed to commit message", zap.Error(err))
		observability.SystemErrors.WithLabelValues(r.agentType, "commit_message").Inc()
	}
}
