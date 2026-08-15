package orchestration

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/validation"
	"go.uber.org/zap"
)

// bugs_open/274 — a completed child workflow's success reply omitted
// sender_agent_type and in_response_to_step_name, two of the five headers
// ValidateOutgoingMessage requires, so producer-side validation refused it
// deterministically and every completed child workflow was reported to its
// parent as FAILED (~16,869 agent_error_log rows, 2026-08-03..15).
//
// The existing 158 tests in reply_delivery_adoption_test.go use a double whose
// ProduceWithValidation deliberately does not validate — which is exactly how
// this bug stayed invisible to them. These tests therefore run the REAL
// validator over the headers the coordinator REALLY produced: the double only
// records, and validation.NewValidator judges the recording. No mirrored
// validation logic that could drift from production.

// headerRecordingProducer records every produce's headers and never refuses.
type headerRecordingProducer struct {
	produced []map[string]string
	topics   []string
}

func (p *headerRecordingProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	p.produced = append(p.produced, headers)
	p.topics = append(p.topics, topic)
	return nil
}

func (p *headerRecordingProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return p.Produce(ctx, topic, headers, key, value)
}

func (p *headerRecordingProducer) Close() error { return nil }

func completedChildState() *OrchestrationState {
	return &OrchestrationState{
		OrchestrationID: "orch-274-child",
		CorrelationID:   "corr-274",
		ClientID:        "client-274",
		OwnerAgentType:  "asset-deployer",
		OwnerAgentID:    "agent-274",
		OwnerAgentRole:  "worker",
		CollectedData: map[string]interface{}{
			"__parent_responses_topic__": "system.agent.parent.responses",
			"__reply_to_request_id__":    "req-274",
			// The shape __execution_context__ has after a DB round-trip: a
			// plain map keyed by the struct's JSON tags.
			"__execution_context__": map[string]interface{}{
				"step_name": "spawn_asset_deployer",
			},
			"result": "done",
		},
	}
}

// TestNotifyParentOfSuccessHeadersPassTheRealValidator is the fix: the reply a
// completed child sends its parent must satisfy the same validator the real
// producer consults.
func TestNotifyParentOfSuccessHeadersPassTheRealValidator(t *testing.T) {
	prod := &headerRecordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), podName: "pod-274"}

	s.notifyParentOfSuccess(context.Background(), completedChildState())

	if len(prod.produced) != 1 {
		t.Fatalf("a deliverable success must produce exactly ONE message, got %d", len(prod.produced))
	}
	headers := prod.produced[0]

	if !validation.NewValidator(zap.NewNop()).ValidateOutgoingMessage(headers) {
		t.Fatalf("the real validator refused the success reply's headers — the exact defect of bugs_open/274; headers: %v", headers)
	}
	if got := headers["sender_agent_type"]; got != "asset-deployer" {
		t.Errorf("sender_agent_type must be the child's own resolved type, got %q", got)
	}
	if got := headers["in_response_to_step_name"]; got != "spawn_asset_deployer" {
		t.Errorf("in_response_to_step_name must be the parent's spawning step, got %q", got)
	}
	if got := headers["is_error"]; got != "false" {
		t.Errorf("a success reply must not lean on the is_error exemption, got is_error=%q", got)
	}
	if got := headers["status"]; got != "complete" {
		t.Errorf("status must be complete, got %q", got)
	}
}

// TestNotifyParentOfSuccessHeadersMutationControl proves the first test can
// fail: with the identity blanked and both step-name sources removed, the same
// validator must REFUSE what the coordinator emits. If this ever passes
// validation, the validator's contract has changed shape and the test above is
// vacuous.
func TestNotifyParentOfSuccessHeadersMutationControl(t *testing.T) {
	prod := &headerRecordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), podName: "pod-274"}

	state := completedChildState()
	state.OwnerAgentType = ""
	delete(state.CollectedData, "__execution_context__")
	delete(state.CollectedData, "__work_request__")

	s.notifyParentOfSuccess(context.Background(), state)

	if len(prod.produced) == 0 {
		t.Fatalf("the recording double refuses nothing, so a message must still have been recorded")
	}
	if validation.NewValidator(zap.NewNop()).ValidateOutgoingMessage(prod.produced[0]) {
		t.Fatalf("headers with no sender and no step name PASSED the real validator — the guard these tests rely on is gone")
	}
}

func TestParentReplyStepNameRecovery(t *testing.T) {
	cases := []struct {
		name string
		data map[string]interface{}
		want string
	}{
		{
			name: "map-shaped execution context (the post-DB shape)",
			data: map[string]interface{}{
				"__execution_context__": map[string]interface{}{"step_name": "call_writer"},
			},
			want: "call_writer",
		},
		{
			name: "typed execution context (freshly built, pre-persist)",
			data: map[string]interface{}{
				"__execution_context__": &types.ExecutionContext{StepName: "spawn_hero_writer"},
			},
			want: "spawn_hero_writer",
		},
		{
			name: "falls back to __work_request__ when the context has no step name",
			data: map[string]interface{}{
				"__execution_context__": map[string]interface{}{"step_name": ""},
				"__work_request__":      map[string]interface{}{"step_name": "process_item"},
			},
			want: "process_item",
		},
		{
			name: "absent everywhere",
			data: map[string]interface{}{},
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parentReplyStepName(&OrchestrationState{CollectedData: tc.data})
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestNotifyParentOfFailureCarriesIdentityAndClientID — the failure arm always
// delivered (plain produce + is_error), but its envelope omitted the sender,
// the step name AND client_id, so it passed the parent's incoming validation
// only via the is_error bypass. It must now be truthful on its own merits.
func TestNotifyParentOfFailureCarriesIdentityAndClientID(t *testing.T) {
	prod := &headerRecordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), podName: "pod-274"}

	s.notifyParentOfFailure(context.Background(), completedChildState(), "the child broke")

	if len(prod.produced) != 1 {
		t.Fatalf("expected exactly one failure notice, got %d", len(prod.produced))
	}
	headers := prod.produced[0]
	if got := headers["client_id"]; got != "client-274" {
		t.Errorf("client_id missing from the failure envelope, got %q", got)
	}
	if got := headers["sender_agent_type"]; got != "asset-deployer" {
		t.Errorf("sender_agent_type missing from the failure envelope, got %q", got)
	}
	if got := headers["in_response_to_step_name"]; got != "spawn_asset_deployer" {
		t.Errorf("in_response_to_step_name missing from the failure envelope, got %q", got)
	}
	if got := headers["is_error"]; got != "true" {
		t.Errorf("a failure notice must carry is_error=true, got %q", got)
	}
}
