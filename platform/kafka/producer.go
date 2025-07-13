// FILE: platform/kafka/producer.go
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer defines the interface for Kafka message production
type Producer interface {
	Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	Close() error
}

// KafkaProducer wraps the kafka-go writer for standardized message production
type KafkaProducer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// NewProducer creates a new standardized Kafka producer
func NewProducer(brokers []string, logger *zap.Logger) (Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list cannot be empty")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Kafka producer created", zap.Strings("brokers", brokers))

	return &KafkaProducer{
		writer: writer,
		logger: logger,
	}, nil
}

// Produce sends a message to a specific topic with standard headers
func (p *KafkaProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
		Time:    time.Now().UTC(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("Failed to produce Kafka message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	p.logger.Debug("Successfully produced message", zap.String("topic", topic), zap.String("key", string(key)))
	return nil
}

// Close gracefully closes the producer's writer
func (p *KafkaProducer) Close() error {
	p.logger.Info("Closing Kafka producer...")
	return p.writer.Close()
}
