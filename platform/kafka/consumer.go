// FILE: platform/kafka/consumer.go (updated version)
package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Consumer wraps the kafka-go reader for standardized consumption
type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

// NewConsumer creates a new standardized Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, logger *zap.Logger) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list cannot be empty")
	}
	if topic == "" {
		return nil, fmt.Errorf("kafka topic cannot be empty")
	}
	if groupID == "" {
		return nil, fmt.Errorf("kafka groupID cannot be empty")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,    // Manual commit
	})

	logger.Info("Kafka consumer created",
		zap.Strings("brokers", brokers),
		zap.String("topic", topic),
		zap.String("groupID", groupID),
	)

	return &Consumer{
		reader: reader,
		logger: logger,
	}, nil
}

// FetchMessage fetches the next message from the topic
// Returns the native kafka.Message type
func (c *Consumer) FetchMessage(ctx context.Context) (Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		if err == context.Canceled {
			return Message{}, err
		}
		c.logger.Error("Failed to fetch message from Kafka", zap.Error(err))
		return Message{}, err
	}
	return msg, nil
}

// CommitMessages commits the offset for the given messages
func (c *Consumer) CommitMessages(ctx context.Context, msgs ...Message) error {
	err := c.reader.CommitMessages(ctx, msgs...)
	if err != nil {
		c.logger.Error("Failed to commit Kafka messages", zap.Error(err))
	}
	return err
}

// Close gracefully closes the consumer's reader
func (c *Consumer) Close() error {
	c.logger.Info("Closing Kafka consumer...")
	return c.reader.Close()
}
