// FILE: platform/kafka/producer_test.go
package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewProducer(t *testing.T) {
	// Test with empty brokers
	_, err := NewProducer([]string{}, zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka brokers list cannot be empty")

	// Test with valid brokers (won't actually connect in unit test)
	producer, err := NewProducer([]string{"localhost:9092"}, zap.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, producer)

	// Clean up
	producer.Close()
}

func TestProducerInterface(t *testing.T) {
	// Ensure KafkaProducer implements Producer interface
	var _ Producer = (*KafkaProducer)(nil)
}
