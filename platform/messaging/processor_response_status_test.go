// FILE: platform/messaging/processor_response_status_test.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// These tests pin the response envelope this processor puts on the wire
// (bugs_open/196). They drive the real senders through a recording producer, so
// they assert what an awaiting parent would actually receive — the status header
// it switches on, the typed ErrorInfo it reads the message from, and the legacy
// body blob its Go readers parse.

type recordedResponse struct {
	topic   string
	headers map[string]string
	key     string
	message types.ResponseMessage
}

type recordingResponseProducer struct {
	sent []recordedResponse
	err  error
}

func (r *recordingResponseProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	var msg types.ResponseMessage
	if err := json.Unmarshal(value, &msg); err != nil {
		return fmt.Errorf("test producer could not unmarshal response: %w", err)
	}
	copied := make(map[string]string, len(headers))
	for k, v := range headers {
		copied[k] = v
	}
	r.sent = append(r.sent, recordedResponse{topic: topic, headers: copied, key: string(key), message: msg})
	return r.err
}

func (r *recordingResponseProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return r.Produce(ctx, topic, headers, key, value)
}

func (r *recordingResponseProducer) Close() error { return nil }

var _ kafka.Producer = (*recordingResponseProducer)(nil)

func responseTestProcessor(p kafka.Producer) *MessageProcessor {
	return &MessageProcessor{
		agentType: "test-agent",
		agentRole: "worker",
		producer:  p,
		logger:    zap.NewNop(),
	}
}

// responseTestContext is a child answering a parent: both topics set, so the
// senders reach Produce rather than bailing on a missing responses topic.
func responseTestContext() *MessageContext {
	return &MessageContext{
		Headers: map[string]string{},
		ExecutionContext: &types.ExecutionContext{
			CorrelationID:         "corr-196",
			OrchestrationID:       "orch-child",
			OrchestrationName:     "child-workflow",
			ParentOrchestrationID: "orch-parent",
			ClientID:              "client-196",
			MessageID:             "msg-196",
			RequestID:             "req-196",
			ReplyToRequestID:      "req-196",
			StepID:                "step-1",
			StepName:              "call_the_child",
			Action:                "process",
			ResponsesTopic:        "parent.responses",
			ReplyToTopic:          "parent.responses",
			FuelBudget:            1000,
			TimeoutSeconds:        300,
			Timestamp:             time.Now(),
			Version:               "2.0",
		},
		Logger:        zap.NewNop(),
		CollectedData: map[string]interface{}{},
		StartTime:     time.Now(),
	}
}

// onlyResponse asserts exactly one message was produced and returns it.
func onlyResponse(t *testing.T, p *recordingResponseProducer) recordedResponse {
	t.Helper()
	if len(p.sent) != 1 {
		t.Fatalf("produced %d messages, want 1", len(p.sent))
	}
	return p.sent[0]
}

// coordinatorArm mirrors the parent's routing decision on a response status:
// `switch execCtx.Status` in SagaCoordinator.HandleResponse
// (platform/orchestration/coordinator.go). Copied rather than called — the real
// switch sits behind a live DB and a claimed state row — because the mapping,
// not the plumbing, is the contract a sender has to satisfy. Keep in step with
// it: the arms are what decides whether a child's failure is treated as output.
func coordinatorArm(status string) string {
	switch status {
	case "awaiting", "processing":
		return "progress"
	case "complete", "success":
		return "complete"
	case "error_recoverable":
		return "recoverable"
	case "error_unrecoverable", "failed", "error":
		return "unrecoverable"
	default:
		// The switch warns on an unknown status and treats it as unrecoverable.
		return "unrecoverable"
	}
}

// TestReproducesTheBug_completeStampedFailureRoutesToCompleteArm documents the
// defect, and is why the status seam in sendWorkflowResponseWithStatus exists.
//
// The first half is the envelope the pre-fix senders built: every failure went
// through sendWorkflowResponse, which hard-coded status "complete" and
// Success:true. Nothing in it marks a failure, so the parent's switch takes the
// complete arm and applyResponseToState records the error blob as that step's
// OUTPUT — the workflow then continues on junk. The second half drives the real
// sender and fails on the pre-fix code, which is what makes this a reproduction
// rather than a description.
func TestReproducesTheBug_completeStampedFailureRoutesToCompleteArm(t *testing.T) {
	// The pre-fix envelope, built by hand: a failure wearing a success.
	preFix := types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Status:     "complete",
			IsComplete: true,
			IsError:    false,
		},
		Body: types.ResponseBody{
			Success: true,
			Body: map[string]interface{}{
				"error":  "WORKFLOW_INVALID: Invalid workflow configuration",
				"status": "failed",
			},
			Error: nil,
		},
	}

	if arm := coordinatorArm(preFix.Headers.Status); arm != "complete" {
		t.Fatalf("pre-fix envelope routes to %q; the bug is that it routes to the complete arm", arm)
	}
	// And the parent had nothing else to notice with: the two fields a
	// complete-arm reader would consult both say "fine".
	if !preFix.Body.Success || preFix.Body.Error != nil {
		t.Fatal("pre-fix envelope must present as a success — that was the defect")
	}

	// Now the real sender on the same failure.
	p := &recordingResponseProducer{}
	proc := responseTestProcessor(p)
	failure := errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").Build()

	if err := proc.sendErrorResponse(context.Background(), responseTestContext(), failure); err != nil {
		t.Fatalf("sendErrorResponse: %v", err)
	}

	got := onlyResponse(t, p)
	if arm := coordinatorArm(got.message.Headers.Status); arm != "unrecoverable" {
		t.Errorf("failure response routes to %q arm (status %q), want unrecoverable",
			arm, got.message.Headers.Status)
	}
	// The wire header is what the coordinator's legacy header path parses, so it
	// has to carry the status too, not just the marshalled body.
	if got.headers["status"] != "error_unrecoverable" {
		t.Errorf("kafka status header = %q, want error_unrecoverable", got.headers["status"])
	}
}

// TestErrorResponseStatusIsSequencedFromTheError pins the status/ErrorInfo
// mapping. The decision goes through RetryDisposition — permanent first, then
// transient, else terminal (bugs_open/207). Until that convergence this test
// pinned the opposite for untyped transient prose (typed-only, everything
// terminal); the flips below are the fix, taken deliberately, not a drift.
func TestErrorResponseStatusIsSequencedFromTheError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		wantStatus      string
		wantCode        string
		wantRecoverable bool
	}{
		{
			// A bare error carries no code. In production handleError wraps one
			// in errors.InternalError before it reaches here, so an empty code is
			// the direct-call edge case, not the live path. FLIPPED by 207: this
			// is the census's ~30% headline shape, now recoverable.
			name:            "untyped transient prose is recoverable",
			err:             fmt.Errorf("llm call failed: context deadline exceeded"),
			wantStatus:      "error_recoverable",
			wantCode:        "",
			wantRecoverable: true,
		},
		{
			// FLIPPED by 207, through the deliberate Retryable=false fall-through:
			// InternalError defaults Retryable=false while wrapping a transient
			// cause, and the cause text renders into Error().
			name:            "untyped transient as handleError wraps it",
			err:             errors.InternalError("Processing failed", fmt.Errorf("context deadline exceeded")),
			wantStatus:      "error_recoverable",
			wantCode:        "INTERNAL_ERROR",
			wantRecoverable: true,
		},
		{
			// The sequencing pin, and why the sender uses RetryDisposition rather
			// than MatchedTransientFailure alone: this prose carries a permanent
			// needle ("invalid") AND a transient one ("connection"). Permanent
			// must win, or handleError's validation branch drops the message
			// while this status tells the coordinator to retry it. Fails if the
			// two questions are ever reordered.
			name:            "permanent needle outranks transient needle",
			err:             fmt.Errorf("pq: invalid connection"),
			wantStatus:      "error_unrecoverable",
			wantCode:        "",
			wantRecoverable: false,
		},
		{
			// Unclassifiable stays terminal: no permanent needle, no transient
			// needle, no type. The convergence must not fail open.
			name:            "unclassifiable untyped error stays terminal",
			err:             fmt.Errorf("runtime error: index out of range [3] with length 2"),
			wantStatus:      "error_unrecoverable",
			wantCode:        "",
			wantRecoverable: false,
		},
		{
			// The motivating failure: the fleet's commonest permanent config
			// error (bugs_closed/195), Retryable false by default.
			name:            "DomainError not retryable",
			err:             errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").Build(),
			wantStatus:      "error_unrecoverable",
			wantCode:        "WORKFLOW_INVALID",
			wantRecoverable: false,
		},
		{
			// Dormant today — nothing in production constructs a retryable —
			// but the arm the coordinator retries on, so it is pinned.
			name:            "DomainError marked retryable",
			err:             errors.New(errors.ErrAgentOverloaded, "downstream busy").AsRetryable(nil).Build(),
			wantStatus:      "error_recoverable",
			wantCode:        "AGENT_OVERLOADED",
			wantRecoverable: true,
		},
		{
			// A %w wrap must not lose the classification (bugs_closed/195):
			// CodeOf and IsRetryable read the whole chain.
			name:            "wrapped retryable DomainError",
			err:             fmt.Errorf("sending request: %w", errors.New(errors.ErrAgentOverloaded, "downstream busy").AsRetryable(nil).Build()),
			wantStatus:      "error_recoverable",
			wantCode:        "AGENT_OVERLOADED",
			wantRecoverable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &recordingResponseProducer{}
			proc := responseTestProcessor(p)

			if err := proc.sendErrorResponse(context.Background(), responseTestContext(), tt.err); err != nil {
				t.Fatalf("sendErrorResponse: %v", err)
			}

			got := onlyResponse(t, p)
			h := got.message.Headers

			if h.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", h.Status, tt.wantStatus)
			}
			if h.IsComplete {
				t.Error("IsComplete = true on a failure response")
			}
			if !h.IsError {
				t.Error("IsError = false on a failure response")
			}
			if got.message.Body.Success {
				t.Error("Body.Success = true on a failure response")
			}

			errInfo := got.message.Body.Error
			if errInfo == nil {
				t.Fatal("Body.Error is nil; the coordinator reads the real message from it")
			}
			if errInfo.Code != tt.wantCode {
				t.Errorf("Body.Error.Code = %q, want %q", errInfo.Code, tt.wantCode)
			}
			if errInfo.Message != tt.err.Error() {
				t.Errorf("Body.Error.Message = %q, want %q", errInfo.Message, tt.err.Error())
			}
			if errInfo.Recoverable != tt.wantRecoverable {
				t.Errorf("Body.Error.Recoverable = %v, want %v", errInfo.Recoverable, tt.wantRecoverable)
			}

			// Wire header and marshalled body must agree — the coordinator may
			// read either depending on how it built its ExecutionContext.
			if got.headers["status"] != tt.wantStatus {
				t.Errorf("kafka status header = %q, want %q", got.headers["status"], tt.wantStatus)
			}
			if got.headers["is_error"] != "true" {
				t.Errorf("kafka is_error header = %q, want true", got.headers["is_error"])
			}
		})
	}
}

// TestWorkflowFailureResponseStatusIsSequenced covers the second error sender:
// workflow-start failures, which 196's bug file did not name and which used to
// be stamped complete by the same shared path. A typed permanent failure stays
// terminal through the sequenced decision (bugs_open/207).
func TestWorkflowFailureResponseStatusIsSequenced(t *testing.T) {
	p := &recordingResponseProducer{}
	proc := responseTestProcessor(p)
	failure := errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").
		WithCause(fmt.Errorf("step 'done' with action 'complete' requires a topic")).
		Build()

	if err := proc.sendWorkflowFailureResponse(context.Background(), responseTestContext(), failure); err != nil {
		t.Fatalf("sendWorkflowFailureResponse: %v", err)
	}

	got := onlyResponse(t, p)
	if got.message.Headers.Status != "error_unrecoverable" {
		t.Errorf("Status = %q, want error_unrecoverable", got.message.Headers.Status)
	}
	if got.message.Body.Success {
		t.Error("Body.Success = true on a workflow failure")
	}
	if got.message.Body.Error == nil {
		t.Fatal("Body.Error is nil on a workflow failure")
	}
	if got.message.Body.Error.Code != "WORKFLOW_INVALID" {
		t.Errorf("Body.Error.Code = %q, want WORKFLOW_INVALID", got.message.Body.Error.Code)
	}
	if got.message.Body.Error.Message != failure.Error() {
		t.Errorf("Body.Error.Message = %q, want %q", got.message.Body.Error.Message, failure.Error())
	}
}

// TestWorkflowFailureTransientProseIsRecoverable is 207's flip on the
// workflow-start sender: an untyped dial failure while starting a workflow is
// exactly the shape a retry can cure, and it used to be terminal here.
func TestWorkflowFailureTransientProseIsRecoverable(t *testing.T) {
	p := &recordingResponseProducer{}
	proc := responseTestProcessor(p)
	failure := fmt.Errorf("failed to start workflow: dial tcp 10.0.0.1:5432: connect: connection refused")

	if err := proc.sendWorkflowFailureResponse(context.Background(), responseTestContext(), failure); err != nil {
		t.Fatalf("sendWorkflowFailureResponse: %v", err)
	}

	got := onlyResponse(t, p)
	if got.message.Headers.Status != "error_recoverable" {
		t.Errorf("Status = %q, want error_recoverable", got.message.Headers.Status)
	}
	if arm := coordinatorArm(got.message.Headers.Status); arm != "recoverable" {
		t.Errorf("routes to %q arm, want recoverable", arm)
	}
	if got.message.Body.Error == nil {
		t.Fatal("Body.Error is nil on a workflow failure")
	}
	if !got.message.Body.Error.Recoverable {
		t.Error("Body.Error.Recoverable = false; determineStatus re-derives the status from it, so it must move in lockstep")
	}
}

// TestSuccessResponseStatusStillComplete is the no-op half. Splitting the sender
// must leave the success envelope exactly as it was: the same status, the same
// flags, the result unwrapped in Body.Body and no Error key at all.
func TestSuccessResponseStatusStillComplete(t *testing.T) {
	p := &recordingResponseProducer{}
	proc := responseTestProcessor(p)
	result := map[string]interface{}{"generated_text": "hello"}

	if err := proc.sendWorkflowResponse(context.Background(), responseTestContext(), result); err != nil {
		t.Fatalf("sendWorkflowResponse: %v", err)
	}

	got := onlyResponse(t, p)
	if got.message.Headers.Status != "complete" {
		t.Errorf("Status = %q, want complete", got.message.Headers.Status)
	}
	if !got.message.Headers.IsComplete {
		t.Error("IsComplete = false on a success response")
	}
	if got.message.Headers.IsError {
		t.Error("IsError = true on a success response")
	}
	if !got.message.Body.Success {
		t.Error("Body.Success = false on a success response")
	}
	if got.message.Body.Error != nil {
		t.Errorf("Body.Error = %+v, want nil on a success response", got.message.Body.Error)
	}
	if arm := coordinatorArm(got.message.Headers.Status); arm != "complete" {
		t.Errorf("success response routes to %q arm, want complete", arm)
	}
	if got.topic != "parent.responses" {
		t.Errorf("topic = %q, want parent.responses", got.topic)
	}
	if got.key != "corr-196" {
		t.Errorf("key = %q, want the correlation id", got.key)
	}

	body, ok := got.message.Body.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Body.Body = %T, want the result map", got.message.Body.Body)
	}
	if body["generated_text"] != "hello" {
		t.Errorf("Body.Body = %v, want the result unwrapped", body)
	}
}

// TestErrorResponseKeepsLegacyBodyBlobDialect guards the compatibility half of
// the fix. Several Go readers parse `status: "failed"` out of the body — the 017
// lane's handlerReportedFailure guard, loop_actions, multipage_actions,
// git_deployer_actions, fixloop_digest_action — so the typed ErrorInfo is
// ADDITIVE. Changing the blob's shape here silently breaks all of them, and
// nothing else in the tree would fail.
func TestErrorResponseKeepsLegacyBodyBlobDialect(t *testing.T) {
	failure := errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").Build()

	senders := map[string]func(*MessageProcessor) error{
		"sendErrorResponse": func(proc *MessageProcessor) error {
			return proc.sendErrorResponse(context.Background(), responseTestContext(), failure)
		},
		"sendWorkflowFailureResponse": func(proc *MessageProcessor) error {
			return proc.sendWorkflowFailureResponse(context.Background(), responseTestContext(), failure)
		},
	}

	for name, send := range senders {
		t.Run(name, func(t *testing.T) {
			p := &recordingResponseProducer{}
			proc := responseTestProcessor(p)
			if err := send(proc); err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			body, ok := onlyResponse(t, p).message.Body.Body.(map[string]interface{})
			if !ok {
				t.Fatalf("Body.Body is not the error blob")
			}
			if body["status"] != "failed" {
				t.Errorf(`Body.Body["status"] = %v, want "failed"`, body["status"])
			}
			if body["error"] != failure.Error() {
				t.Errorf(`Body.Body["error"] = %v, want %q`, body["error"], failure.Error())
			}
			if len(body) != 2 {
				t.Errorf("Body.Body has %d keys (%v), want exactly error and status", len(body), body)
			}
		})
	}
}

// TestErrorResponseNeedsAResponsesTopic keeps the existing guard visible: with
// no responses topic there is nobody to answer, and the sender must say so
// rather than produce to an empty topic.
func TestErrorResponseNeedsAResponsesTopic(t *testing.T) {
	p := &recordingResponseProducer{}
	proc := responseTestProcessor(p)
	msgCtx := responseTestContext()
	msgCtx.ExecutionContext.ResponsesTopic = ""

	if err := proc.sendErrorResponse(context.Background(), msgCtx, fmt.Errorf("boom")); err == nil {
		t.Error("sendErrorResponse with no responses topic returned nil")
	}
	if len(p.sent) != 0 {
		t.Errorf("produced %d messages with no responses topic, want 0", len(p.sent))
	}
}
