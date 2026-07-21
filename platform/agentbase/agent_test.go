// FILE: platform/agentbase/agent_test.go
package agentbase

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
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
	// This test needs to be redesigned since handleMessage is private
	// and the Agent struct expects real Kafka connections
	// For now, we'll skip this test or make it integration-only
	t.Skip("Skipping unit test that requires real Kafka connections")
}

// TestMatchedValidationNeedle pins the substring classifier behind the
// dropped-validation-error branch (bugs_open/034). It documents BOTH halves:
// genuine validation errors are still classified (so retry behaviour is
// unchanged), and the unanchored match still catches non-validation runtime
// failures — which is precisely why recordDroppedValidationError now persists a
// durable row and records which needle fired.
func TestMatchedValidationNeedle(t *testing.T) {
	tests := []struct {
		name string
		err  string
		want string
	}{
		// Genuine validation errors — behaviour preserved.
		{"client_id required", "client_id is required to execute a workflow", "is required"},
		{"generic validation", "payload failed validation", "validation"},
		{"invalid field", "field foo is invalid", "invalid"},

		// Unanchored-match hazard (bugs_open/034): these are NOT validation
		// errors, yet they match. The classifier still drops them; the point of
		// the fix is that the drop is now durably recorded, not that these stop
		// matching.
		{"truncated-LLM parse failure", "unrecoverable after control-char repair: invalid character 'w' after object key:value pair", "invalid"},
		{"db driver error", "pq: invalid connection", "invalid"},
		{"nil-pointer recovery", "runtime error: invalid memory address or nil pointer dereference", "invalid"},

		// Ordinary failures that must NOT be classified as validation errors.
		{"plain timeout", "context deadline exceeded", ""},
		{"kafka unavailable", "broker not available", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchedValidationNeedle(tt.err); got != tt.want {
				t.Errorf("matchedValidationNeedle(%q) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}
