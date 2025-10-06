// FILE: platform/kafka/producer.go
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/validation"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer defines the interface for Kafka message production
type Producer interface {
	Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
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
func (p *KafkaProducer) Produce(ctx context.Context, send_to_topic string, headers map[string]string, key, value []byte) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Topic:   send_to_topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
		Time:    time.Now().UTC(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("Failed to produce Kafka message",
			zap.String("send_to_topic", send_to_topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	p.logger.Info("agent producer.go Successfully produced message",
		zap.String("sent to topic", send_to_topic),
		zap.String("key", string(key)),
		zap.String("value", string(value)),
		zap.Any("headers", headers),
	)
	return nil
}

// ProduceWithValidation validates using the validation package before sending
func (p *KafkaProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	// Use the existing ValidateOutgoingMessage from validation package
	validator := validation.NewValidator(p.logger)
	if !validator.ValidateOutgoingMessage(headers) {
		// Check if it's an error message - those we always send
		if headers["is_error"] == "true" {
			p.logger.Warn("Error message failed validation but sending anyway",
				zap.String("topic", topic),
				zap.String("correlation_id", headers["correlation_id"]))
			return p.Produce(ctx, topic, headers, key, value)
		}

		return fmt.Errorf("message validation failed")
	}

	// Validation passed, send the message
	return p.Produce(ctx, topic, headers, key, value)
}

// Close gracefully closes the producer's writer
func (p *KafkaProducer) Close() error {
	p.logger.Info("Closing Kafka producer...")
	return p.writer.Close()
}
