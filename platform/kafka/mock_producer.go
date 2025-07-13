// FILE: platform/kafka/mock_producer.go
// Mock producer for testing
package kafka

import (
	"context"
	"github.com/stretchr/testify/mock"
)

// MockProducer is a mock implementation of the Producer interface
type MockProducer struct {
	mock.Mock
}

// Produce mocks the Produce method
func (m *MockProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	args := m.Called(ctx, topic, headers, key, value)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}
