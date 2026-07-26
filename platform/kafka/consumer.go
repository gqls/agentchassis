// FILE: platform/kafka/consumer.go (updated version)
package kafka

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
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

	sessionTimeout := envDurationOrDefault("KAFKA_SESSION_TIMEOUT", 60*time.Second)
	rebalanceTimeout := envDurationOrDefault("KAFKA_REBALANCE_TIMEOUT", 60*time.Second)
	heartbeatInterval := sessionTimeout / 3 // Kafka requires heartbeat < session timeout

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,                 // 1 byte
		MaxBytes:       10e6,              // 10MB
		CommitInterval: 0,                 // Manual commit
		StartOffset:    kafka.FirstOffset, // Start from beginning if no offset stored
		// bugs_open/040-kafka-dial: was an inline 10s dialer. Now the shared
		// instrumented one, so consumer dials are counted and share the
		// producer's timeout instead of silently disagreeing with it.
		Dialer: SharedDialer(),
		SessionTimeout:    sessionTimeout,
		RebalanceTimeout:  rebalanceTimeout,
		HeartbeatInterval: heartbeatInterval,
		MaxWait:           1 * time.Second,
	})

	logger.Info("Kafka consumer created",
		zap.Strings("brokers", brokers),
		zap.String("topic", topic),
		zap.String("groupID", groupID),
		zap.Duration("session_timeout", sessionTimeout),
		zap.Duration("rebalance_timeout", rebalanceTimeout),
		zap.Duration("heartbeat_interval", heartbeatInterval),
	)

	return &Consumer{
		reader: reader,
		logger: logger,
	}, nil
}

// NOTE (bugs_open/003 F3): there used to be a Consume() here that fetched a
// message and committed its offset BEFORE returning it to the caller —
// at-most-once delivery wearing an at-least-once comment. Every pod death
// annihilated whatever was in flight. It is deliberately DELETED, not
// deprecated: use FetchMessage → process → CommitMessages, like every
// current caller does.

// FetchMessage fetches the next message from the topic
// Returns the native kafka.Message type
func (c *Consumer) FetchMessage(ctx context.Context) (Message, error) {
	c.logger.Debug("FetchMessage called") // ADD THIS

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		// Adapters poll with a short timeout context: DeadlineExceeded here is
		// the NORMAL empty-poll case, not a fault. It used to log at ERROR every
		// poll interval on every idle adapter, drowning the real log (observed on
		// the browser-runner, 2026-07-14). errors.Is, not ==: kafka-go may wrap it.
		if errors.Is(err, context.Canceled) {
			return Message{}, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			c.logger.Debug("No messages available within poll timeout",
				zap.String("topic", c.reader.Config().Topic))
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

func envDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	seconds, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return time.Duration(seconds) * time.Second
}
