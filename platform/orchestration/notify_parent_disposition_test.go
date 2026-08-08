// FILE: platform/orchestration/notify_parent_disposition_test.go
//
// bugs_open/217. notifyParentOfFailure used to hardcode error_unrecoverable /
// Recoverable:false — the third unconverged failure sender after 207's
// processor pair — and on the orchestrated path its response claims the
// awaited request FIRST, so the hardcoded verdict pre-empted every
// better-classified one. These tests pin the converged contract: the wire
// status is decided by the sequenced shared classifier
// (perrors.RetryDisposition) run over the failure prose, and the envelope's
// code and message stay verbatim whatever the disposition.
//
// The sequencing case is the load-bearing pin: prose carrying BOTH a
// permanent and a transient needle ("pq: invalid connection") must stay
// error_unrecoverable. It fails if anyone swaps RetryDisposition for
// MatchedTransientFailure, which is exactly the drift the RSH-007 landmine
// warns against.

package orchestration

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

func failedChildStateFixture(parentTopic, replyToRequestID string) *OrchestrationState {
	return &OrchestrationState{
		OrchestrationID: "child-orch-217",
		CorrelationID:   "corr-217",
		CurrentStep:     "call_child",
		CollectedData: map[string]interface{}{
			"__parent_responses_topic__": parentTopic,
			"__reply_to_request_id__":    replyToRequestID,
		},
	}
}

// notifyAndCapture runs notifyParentOfFailure over an unreachable DB (the
// error-log write is best-effort by design) and returns the single produced
// response envelope.
func notifyAndCapture(t *testing.T, errorMsg string) types.ResponseMessage {
	t.Helper()
	prod := &recordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), db: unreachableDB(t)}

	s.notifyParentOfFailure(context.Background(), failedChildStateFixture("system.agent.parent.responses", "req-217"), errorMsg)

	if len(prod.produced) != 1 {
		t.Fatalf("produced %d messages, want exactly 1", len(prod.produced))
	}
	var resp types.ResponseMessage
	if err := json.Unmarshal(prod.produced[0].Value, &resp); err != nil {
		t.Fatalf("unmarshal produced response: %v", err)
	}
	return resp
}

// TestNotifyParentTransientFailureIsRecoverable: the flip that is the fix.
// A deadline-exceeded child-orchestration failure (the live population's
// second-largest shape, 1,989 rows/14d) must reach the parent as
// error_recoverable so the 216 replay seam can buy a real retry.
func TestNotifyParentTransientFailureIsRecoverable(t *testing.T) {
	msg := "workflow failed: Request cef0a691 failed: failed to execute request: context deadline exceeded"
	resp := notifyAndCapture(t, msg)

	if resp.Headers.Status != "error_recoverable" {
		t.Errorf("Status = %q, want error_recoverable", resp.Headers.Status)
	}
	if resp.Body.Error == nil || !resp.Body.Error.Recoverable {
		t.Error("Body.Error.Recoverable = false, want true — status and body must move in lockstep")
	}
	if resp.Body.Error != nil && resp.Body.Error.Code != "CHILD_ORCHESTRATION_FAILED" {
		t.Errorf("Code = %q, want CHILD_ORCHESTRATION_FAILED preserved whatever the disposition", resp.Body.Error.Code)
	}
	if resp.Body.Error != nil && resp.Body.Error.Message != msg {
		t.Errorf("Message = %q, want the failure prose verbatim", resp.Body.Error.Message)
	}
}

// TestNotifyParentSequencingPermanentBeatsTransient: the sequencing pin.
// "pq: invalid connection" carries the permanent needle "invalid" AND the
// transient needle "connection"; permanent must be asked first, so the
// parent must NOT retry what the processor layer drops as
// validation-never-retry. This test fails if the sender ever consults the
// transient classifier alone.
func TestNotifyParentSequencingPermanentBeatsTransient(t *testing.T) {
	resp := notifyAndCapture(t, "workflow failed: pq: invalid connection")

	if resp.Headers.Status != "error_unrecoverable" {
		t.Errorf("Status = %q, want error_unrecoverable — permanent-first is the contract", resp.Headers.Status)
	}
	if resp.Body.Error == nil || resp.Body.Error.Recoverable {
		t.Error("Body.Error.Recoverable = true, want false for a permanent-needle failure")
	}
}

// TestNotifyParentUnclassifiableStaysTerminal: no needle, no retry — the
// pre-fix behaviour is preserved for everything the classifier cannot vouch
// for (975 of 11,970 rows in the 14d population).
func TestNotifyParentUnclassifiableStaysTerminal(t *testing.T) {
	resp := notifyAndCapture(t, "probe: unclassifiable failure xyzzy")

	if resp.Headers.Status != "error_unrecoverable" {
		t.Errorf("Status = %q, want error_unrecoverable", resp.Headers.Status)
	}
	if resp.Body.Error == nil || resp.Body.Error.Recoverable {
		t.Error("Body.Error.Recoverable = true, want false for unclassifiable prose")
	}
}

// TestNotifyParentNoParentProducesNothing: the early return is intact — a
// root orchestration (no parent topic / reply id) sends no envelope at all,
// classified or otherwise.
func TestNotifyParentNoParentProducesNothing(t *testing.T) {
	prod := &recordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), db: unreachableDB(t)}

	state := &OrchestrationState{
		OrchestrationID: "root-orch-217",
		CollectedData:   map[string]interface{}{},
	}
	s.notifyParentOfFailure(context.Background(), state, "context deadline exceeded")

	if len(prod.produced) != 0 {
		t.Fatalf("produced %d messages, want 0 — no parent means no notification", len(prod.produced))
	}
}
