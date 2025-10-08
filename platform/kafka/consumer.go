// FILE: platform/kafka/consumer.go (updated version)
package kafka

import (
	"context"
	"fmt"
	"time"

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

	logger.Info("consumer.go setting up NewConsumer")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,                 // 10KB
		MaxBytes:       10e6,              // 10MB
		CommitInterval: time.Second,       // Manual commit
		StartOffset:    kafka.FirstOffset, // Start from beginning if no offset stored
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		},
		// Add these for better consumer group behavior
		SessionTimeout:   10 * time.Second,
		RebalanceTimeout: 10 * time.Second,
		Partition:        0,
		MaxWait:          1 * time.Second, // Don't wait too long for messages
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

// Consume fetches the next message from the topic
func (c *Consumer) Consume(ctx context.Context) (Message, error) {
	c.logger.Debug("Consume() called, attempting to fetch message")

	// Use a timeout context to prevent infinite blocking
	fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg, err := c.reader.FetchMessage(fetchCtx)
	if err != nil {
		if err == context.DeadlineExceeded {
			// Timeout is normal when no messages available
			c.logger.Debug("No messages available within timeout")
			return Message{}, context.DeadlineExceeded
		}
		if err != context.Canceled {
			c.logger.Error("Failed to fetch message",
				zap.Error(err),
				zap.String("topic", c.reader.Config().Topic))
		}
		return Message{}, err
	}

	c.logger.Info("Message fetched successfully",
		zap.String("topic", msg.Topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.Int("size", len(msg.Value)))

	// After successful processing, commit the offset
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		c.logger.Error("Failed to commit message", zap.Error(err))
	}

	return msg, nil
}

// FetchMessage fetches the next message from the topic
// Returns the native kafka.Message type
func (c *Consumer) FetchMessage(ctx context.Context) (Message, error) {
	c.logger.Debug("FetchMessage called") // ADD THIS

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		if err == context.Canceled {
			return Message{}, err
		}
		c.logger.Error("Failed to fetch message from Kafka",
			zap.Error(err),
			zap.String("topic", c.reader.Config().Topic)) // ADD topic info
		return Message{}, err
	}

	c.logger.Debug("Message fetched", // ADD THIS
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset))

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
