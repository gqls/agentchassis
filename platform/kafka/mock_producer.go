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

// ProduceWithValidation mocks the ProduceWithValidation method. Recorded as its
// own call rather than delegating to Produce, because this mock is testify-based
// and a test asserting On("Produce") must not be satisfied by a call the code
// made to ProduceWithValidation — the two are different assertions here, unlike
// in the plain recording mocks elsewhere.
func (m *MockProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	args := m.Called(ctx, topic, headers, key, value)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Compile-time proof that this mock still satisfies the interface it claims to
// mock. Its absence is exactly how it drifted: kafka.Producer gained a method,
// nothing here asserted the contract, and the mock silently stopped implementing
// it while this package still compiled (2026-07-31).
var _ Producer = (*MockProducer)(nil)
