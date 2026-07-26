// FILE: platform/kafka/producer.go
package kafka

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer defines the interface for Kafka message production
type Producer interface {
	Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	Close() error
}

// MessageValidator interface for outgoing message validation.
// Implemented by validation.Validator - injected to avoid cyclic import.
type MessageValidator interface {
	ValidateOutgoingMessage(headers map[string]string) bool
}

// KafkaProducer wraps the kafka-go writer for standardized message production
type KafkaProducer struct {
	writer    *kafka.Writer
	logger    *zap.Logger
	validator MessageValidator // injected, can be nil
}

// NewProducer creates a new standardized Kafka producer without validation
func NewProducer(brokers []string, logger *zap.Logger) (Producer, error) {
	return NewProducerWithValidator(brokers, logger, nil)
}

// NewProducerWithValidator creates a new Kafka producer with an injected validator.
func NewProducerWithValidator(brokers []string, logger *zap.Logger, validator MessageValidator) (Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list cannot be empty")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		WriteTimeout: 10 * time.Second,
		// bugs_open/040-kafka-dial: leaving Transport nil fell through to
		// kafka-go's DefaultTransport, whose dial timeout is 3s — so the
		// producer and the consumer had been dialling the same brokers on
		// different budgets, and neither was counted.
		Transport: SharedTransport(),
	}

	logger.Info("Kafka producer created", zap.Strings("brokers", brokers))

	return &KafkaProducer{
		writer:    writer,
		logger:    logger,
		validator: validator,
	}, nil
}

// SetValidator allows setting the validator after construction.
// Useful when the validator can't be provided at construction time.
func (p *KafkaProducer) SetValidator(v MessageValidator) {
	p.validator = v
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
	current, caller := getFuncInfo(1)
	p.logger.Info("agent producer.go Successfully produced message",
		zap.String("sent to topic", send_to_topic),
		zap.String("key", string(key)),
		zap.String("value", string(value)),
		zap.Any("headers", headers),
		zap.String("current function", current),
		zap.String("caller", caller),
		zap.Time("time", time.Now().UTC()),
	)

	return nil
}

// ProduceWithValidation validates using the injected validator before sending
func (p *KafkaProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	// If no validator injected, just produce without validation
	if p.validator == nil {
		return p.Produce(ctx, topic, headers, key, value)
	}

	if !p.validator.ValidateOutgoingMessage(headers) {
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

// Helper to get current and caller function names
func getFuncInfo(skip int) (current, caller string) {
	// skip=0 => this func, skip=1 => its caller, skip=2 => caller's caller, etc.
	if pc, _, _, ok := runtime.Caller(skip); ok {
		current = runtime.FuncForPC(pc).Name()
	}
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		caller = runtime.FuncForPC(pc).Name()
	}
	return
}
