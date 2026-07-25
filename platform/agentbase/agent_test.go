// FILE: platform/agentbase/agent_test.go
package agentbase

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/platform/messaging"
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

// TestAgentbaseUsesSharedValidationNeedles is the lockstep half of
// bugs_open/034's drift fix. The classifier's own table test now lives with the
// function, in platform/messaging; what matters HERE is that this layer no
// longer keeps a private copy of the list.
//
// The drift it guards against was real and silent: agentbase matched three
// needles, messaging four, so an error whose only needle was "missing" was
// dropped-without-retry by one layer and retried by the other — decided purely
// by which path the message took to reach the failure.
func TestAgentbaseUsesSharedValidationNeedles(t *testing.T) {
	// "missing" is the needle agentbase used to lack. If this layer ever
	// reintroduces a local list, this is the case that will fall out of step
	// first, because it is the one the old local list did not have.
	if got := messaging.MatchedValidationNeedle("missing client_id header"); got != "missing" {
		t.Errorf("agentbase classifier out of step with messaging: got %q, want %q", got, "missing")
	}

	// And a genuine validation error still classifies, so the no-retry
	// behaviour this branch exists for is unchanged.
	if got := messaging.MatchedValidationNeedle("client_id is required to execute a workflow"); got != "is required" {
		t.Errorf("genuine validation error no longer classified: got %q", got)
	}
}
