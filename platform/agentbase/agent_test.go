// FILE: platform/agentbase/agent_test.go
package agentbase

import (
	"context"
	"testing"
	"time"

	"github.com/gqls/ai-persona-system/pkg/models"
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockKafkaConsumer for testing
type MockKafkaConsumer struct {
	mock.Mock
}

func (m *MockKafkaConsumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaConsumer) Close() error {
	return nil
}

func TestAgentHandleMessage(t *testing.T) {
	// Create test message
	testMessage := kafka.Message{
		Topic: "test.topic",
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte("test-correlation-id")},
			{Key: "request_id", Value: []byte("test-request-id")},
			{Key: "client_id", Value: []byte("test-client")},
			{Key: "agent_instance_id", Value: []byte("test-agent-id")},
		},
		Value: []byte(`{"action": "test", "data": {}}`),
	}

	// Create mock dependencies
	mockConsumer := new(MockKafkaConsumer)
	mockProducer := new(MockKafkaProducer)
	logger := zap.NewNop()

	// Create test agent
	agent := &Agent{
		ctx:           context.Background(),
		logger:        logger,
		kafkaConsumer: mockConsumer,
		kafkaProducer: mockProducer,
		agentType:     "test-agent",
	}

	// Set expectations
	mockConsumer.On("CommitMessages", mock.Anything, []kafka.Message{testMessage}).Return(nil)

	// Execute
	agent.handleMessage(testMessage)

	// Verify
	mockConsumer.AssertExpectations(t)
}
