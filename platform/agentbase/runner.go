// FILE: platform/agentbase/runner.go
package agentbase

import (
	"context"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/messaging"
	"github.com/gqls/agentchassis/platform/observability"
	"go.uber.org/zap"
	"time"
)

// MessageRunner handles the message processing loop
type MessageRunner struct {
	ctx           context.Context
	logger        *zap.Logger
	consumer      *kafka.Consumer
	processor     *messaging.MessageProcessor
	consumerGroup string
	agentType     string
}

// NewMessageRunner creates a new message runner
func NewMessageRunner(
	ctx context.Context,
	logger *zap.Logger,
	consumer *kafka.Consumer,
	processor *messaging.MessageProcessor,
	consumerGroup string,
	agentType string,
) *MessageRunner {
	return &MessageRunner{
		ctx:           ctx,
		logger:        logger,
		consumer:      consumer,
		processor:     processor,
		consumerGroup: consumerGroup,
		agentType:     agentType,
	}
}

// Run starts the message processing loop
func (r *MessageRunner) Run() error {
	r.logger.Info("Starting message runner", zap.String("agent_type", r.agentType))

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
