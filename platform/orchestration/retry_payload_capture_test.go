package orchestration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// bugs_open/129. The retry path replays the message the action actually
// produced. Two things have to hold for that to work at all, and both are the
// kind that break silently under a later edit — hence these tests rather than a
// comment.

//  1. The payload must survive the hop from the action's result map onto the
//     awaited request. An action that records nothing must yield nothing, not an
//     empty document: "absent" is what makes the coordinator refuse to retry,
//     and an empty jsonb string would be indistinguishable from a real payload
//     at the column while being unusable at replay time.
func TestExtractRetryPayloadLiftsWhatTheActionRecorded(t *testing.T) {
	produced := types.BuildRetryPayload("job.indexer.requests", "req-1", []byte(`{"headers":{"orchestration_id":"child-1"}}`))

	got := extractRetryPayload(map[string]interface{}{
		"await_response":       true,
		"requests_topic":       "job.indexer.requests",
		types.RetryPayloadKey:  produced,
		"target_agent_type":    "code-indexer",
		"unrelated_result_key": 7,
	})
	if len(got) == 0 {
		t.Fatal("payload not lifted out of the action result — every retry would then be refused")
	}

	stored, err := types.DecodeRetryPayload(got)
	if err != nil {
		t.Fatalf("lifted payload does not decode: %v", err)
	}
	if stored.Topic != "job.indexer.requests" {
		t.Errorf("topic = %q, want the topic the action sent to", stored.Topic)
	}
}

func TestExtractRetryPayloadYieldsNothingWhenTheActionRecordedNothing(t *testing.T) {
	for _, result := range []map[string]interface{}{
		{"await_response": true},
		{"await_response": true, types.RetryPayloadKey: nil},
		nil,
	} {
		if got := extractRetryPayload(result); len(got) != 0 {
			t.Errorf("extractRetryPayload(%v) = %q, want nothing — an absent payload must stay absent so the coordinator refuses to synthesise a retry", result, got)
		}
	}
}

//  2. RequestPayload must NEVER be serialised into
//     orchestration_states.awaited_requests. That column is a JSONB blob
//     rewritten on every state update; a request body in it would be
//     re-serialised many times per orchestration. The payload belongs on the
//     per-request row, read back only at retry time. The `json:"-"` tag is the
//     whole of that decision and nothing else enforces it.
func TestRequestPayloadStaysOutOfTheStateJSONB(t *testing.T) {
	req := &AwaitedRequest{
		RequestID:       "req-1",
		OrchestrationID: "parent-1",
		StepName:        "spawn_indexer",
		RequestPayload:  json.RawMessage(`{"topic":"t","key":"k","message":{"body":{"huge":"payload"}}}`),
	}

	encoded, err := json.Marshal(map[string]*AwaitedRequest{"req-1": req})
	if err != nil {
		t.Fatalf("marshalling the awaited-request map as the state column does: %v", err)
	}

	if strings.Contains(string(encoded), "huge") || strings.Contains(string(encoded), "request_payload") {
		t.Errorf("the request payload leaked into the state JSONB: %s", encoded)
	}
	// …and the fields that DO belong there are still present, so this test
	// cannot pass by the struct having stopped serialising altogether.
	if !strings.Contains(string(encoded), "spawn_indexer") {
		t.Errorf("the awaited request no longer serialises its own fields: %s", encoded)
	}
}

// nullableJSON is what keeps an unrecorded payload out of the jsonb column as
// SQL NULL rather than as an empty string, which the column rejects.
func TestNullableJSONBindsAbsentAsNull(t *testing.T) {
	if got := nullableJSON(nil); got != nil {
		t.Errorf("nullableJSON(nil) = %v, want nil so the driver binds SQL NULL", got)
	}
	if got := nullableJSON(json.RawMessage(``)); got != nil {
		t.Errorf("nullableJSON(empty) = %v, want nil", got)
	}
	got := nullableJSON(json.RawMessage(`{"topic":"t"}`))
	if b, ok := got.([]byte); !ok || string(b) != `{"topic":"t"}` {
		t.Errorf("nullableJSON(doc) = %v, want the document as bytes", got)
	}
}
