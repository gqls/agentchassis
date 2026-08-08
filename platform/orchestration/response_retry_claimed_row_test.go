// FILE: platform/orchestration/response_retry_claimed_row_test.go
//
// bugs_open/216. The recoverable arm decides on the row the CALLER claimed —
// never on a re-read, never on the in-memory copy. The defect these tests pin:
// handleRecoverableError used to re-read the row with a predicate
// ('waiting','retrying') that its own claim (status='processing') had just made
// unsatisfiable, so every cross-pod response-driven retry fell to the
// in-memory fallback, whose request_payload is never populated (json:"-",
// bugs_closed/129), and died at RETRY_PAYLOAD_UNAVAILABLE — 4ms after
// persisting the retry_version bump that makes the row look retried.
//
// The scenario is deliberately hostile in the way production is: the DB is
// unreachable (a re-read CANNOT succeed), and the in-memory awaited set holds
// the row WITHOUT its payload (what any state loaded from the DB carries).
// Only the claimed row passed in has the payload. If a re-read or a fallback
// ever creeps back into this path, the replay below stops happening.

package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// recordingProducer records every accepted produce, keyed by topic.
type recordingProducer struct {
	produced []struct {
		Topic string
		Value []byte
	}
}

func (p *recordingProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	p.produced = append(p.produced, struct {
		Topic string
		Value []byte
	}{topic, value})
	return nil
}

func (p *recordingProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return p.Produce(ctx, topic, headers, key, value)
}

func (p *recordingProducer) Close() error { return nil }

// unreachableDB returns a *sql.DB whose every query fails fast (nothing
// listens on port 1). The one DB write on the replay path
// (UpdateAwaitedRequestRetry) is log-and-continue by design, so an
// unreachable DB exercises exactly the world the defect lived in: the row
// cannot be re-read, and only what the caller passed in can decide.
func unreachableDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", "postgres://u:p@127.0.0.1:1/none?connect_timeout=1")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func claimedRowFixture(t *testing.T, requestID string, withPayload bool) *AwaitedRequest {
	t.Helper()
	row := &AwaitedRequest{
		RequestID:       requestID,
		OrchestrationID: "parent-orch-216",
		StepName:        "call_child",
		RetryVersion:    0,
		TargetAgentID:   "agent-216",
		TargetAgentType: "reasoning-agent",
		RequestsTopic:   "system.agent.reasoning-agent.requests",
		ResponsesTopic:  "system.agent.parent.responses",
		SentAt:          time.Now().Add(-time.Minute),
		TimeoutAt:       time.Now().Add(2 * time.Minute),
		Status:          "processing",
	}
	if withPayload {
		stored, err := json.Marshal(types.BuildRetryPayload(
			"system.agent.reasoning-agent.requests",
			requestID,
			[]byte(`{"headers":{"orchestration_id":"child-orch-216","action":"process","message_type":"request"}}`),
		))
		if err != nil {
			t.Fatalf("marshal retry payload: %v", err)
		}
		row.RequestPayload = stored
	}
	return row
}

func recoverableResponseFixture() types.ResponseMessage {
	return types.ResponseMessage{
		Headers: types.ResponseHeaders{Status: "error_recoverable", IsError: true},
		Body: types.ResponseBody{
			Success: false,
			Error:   &types.ErrorInfo{Code: "TIMEOUT", Message: "context deadline exceeded", Recoverable: true},
		},
	}
}

// TestRecoverableRetryReplaysFromTheClaimedRow is the 216 mechanism, inverted:
// the DB is unreachable and the in-memory copy has no payload, so the ONLY way
// the replay can go out is off the claimed row the caller passed through.
func TestRecoverableRetryReplaysFromTheClaimedRow(t *testing.T) {
	const requestID = "req-216"
	prod := &recordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), db: unreachableDB(t)}

	claimed := claimedRowFixture(t, requestID, true)

	// The in-memory copy, as any DB-loaded state carries it: same row, NO
	// payload (json:"-"). Pre-216 code decided on this copy and refused.
	inMemory := *claimed
	inMemory.RequestPayload = nil
	state := &OrchestrationState{
		OrchestrationID: "parent-orch-216",
		Status:          StatusAwaitingResponses,
		AwaitedRequests: map[string]*AwaitedRequest{requestID: &inMemory},
		CollectedData:   map[string]interface{}{},
	}

	execCtx := &types.ExecutionContext{RequestID: requestID, Status: "error_recoverable"}
	if err := s.handleRecoverableError(context.Background(), state, requestID, execCtx, recoverableResponseFixture(), claimed); err != nil {
		t.Fatalf("handleRecoverableError: %v", err)
	}

	var replays []json.RawMessage
	for _, p := range prod.produced {
		if p.Topic == "system.agent.reasoning-agent.requests" {
			replays = append(replays, p.Value)
		}
	}
	if len(replays) != 1 {
		t.Fatalf("want exactly 1 replay on the original requests topic, got %d (all produces: %d) — "+
			"the retry died before the wire, which is the bugs_open/216 shape", len(replays), len(prod.produced))
	}

	var msg types.RequestMessage
	if err := json.Unmarshal(replays[0], &msg); err != nil {
		t.Fatalf("replayed message does not decode as a RequestMessage: %v", err)
	}
	if msg.Headers.OrchestrationID != "child-orch-216" {
		t.Errorf("replay carries orchestration_id %q, want the CHILD's (child-orch-216) — a parent id here is bugs_open/129 reborn", msg.Headers.OrchestrationID)
	}
	if msg.Headers.RetryVersion != 1 {
		t.Errorf("replay retry_version = %d, want 1", msg.Headers.RetryVersion)
	}
}

// TestRecoverableRetryStillRefusesWithoutAStoredPayload is the negative
// control: passing the claimed row through must not weaken the bugs_open/129
// refusal. A claimed row with NO payload refuses to synthesise — nothing goes
// out on the requests topic.
func TestRecoverableRetryStillRefusesWithoutAStoredPayload(t *testing.T) {
	const requestID = "req-216-bare"
	prod := &recordingProducer{}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop(), db: unreachableDB(t)}

	claimed := claimedRowFixture(t, requestID, false)
	state := &OrchestrationState{
		OrchestrationID: "parent-orch-216",
		Status:          StatusAwaitingResponses,
		AwaitedRequests: map[string]*AwaitedRequest{requestID: claimed},
		CollectedData:   map[string]interface{}{},
		WorkflowPlan:    models.WorkflowPlan{Steps: map[string]models.Step{}},
	}

	execCtx := &types.ExecutionContext{RequestID: requestID, Status: "error_recoverable"}
	// The refusal routes into handleUnrecoverableError → failWorkflow, whose DB
	// writes all fail here; the assertion is only that no request is emitted.
	_ = s.handleRecoverableError(context.Background(), state, requestID, execCtx, recoverableResponseFixture(), claimed)

	for _, p := range prod.produced {
		if p.Topic == "system.agent.reasoning-agent.requests" {
			t.Fatalf("a payload-less request was replayed anyway — the coordinator synthesised a retry it was built to refuse")
		}
	}
}
