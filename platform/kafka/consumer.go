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
	return newConsumerWithStart(brokers, topic, groupID, logger, kafka.FirstOffset)
}

// NewConsumerFromLatest is NewConsumer with StartOffset LastOffset, for a
// consumer whose group id is EPHEMERAL (per-pod) on a topic with long history.
// With FirstOffset such a consumer replays the topic's entire past on every
// pod start — measured 2026-07-28 on the chassis's response lane: 12,280
// messages, ~530/min, ~23 minutes of processing history while deaf to fresh
// traffic, on every restart, growing with the topic (chassis_replica_scaling
// NOTES). The blind window this trades for is the pod-restart seconds, and
// only for messages nothing else would redeliver — the F2 retry driver
// re-sends any await whose response lands there (bugs_open/003).
// Stable-group consumers must keep NewConsumer: their stored offsets make
// StartOffset irrelevant after the first ever start.
func NewConsumerFromLatest(brokers []string, topic, groupID string, logger *zap.Logger) (*Consumer, error) {
	return newConsumerWithStart(brokers, topic, groupID, logger, kafka.LastOffset)
}

func newConsumerWithStart(brokers []string, topic, groupID string, logger *zap.Logger, startOffset int64) (*Consumer, error) {
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
		MinBytes:       1,           // 1 byte
		MaxBytes:       10e6,        // 10MB
		CommitInterval: 0,           // Manual commit
		StartOffset:    startOffset, // Only consulted when the group has no stored offset
		// bugs_open/040-kafka-dial: same 10s budget as before, now counted.
		// The timeout is passed explicitly and deliberately unchanged — see the
		// scope note in dialer.go.
		Dialer:            InstrumentedDialer(10 * time.Second),
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
