// FILE: platform/kafka/interfaces.go
package kafka

import "context"

// Producer interface for Kafka message production
type Producer interface {
	Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	Close() error
}

// Consumer interface for Kafka message consumption
type Consumer interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	Close() error
}

// Message represents a Kafka message
type Message interface {
	GetHeaders() []Header
	GetKey() []byte
	GetValue() []byte
	GetTopic() string
	GetPartition() int
	GetOffset() int64
}

// Header represents a Kafka message header
type Header struct {
	Key   string
	Value []byte
}
