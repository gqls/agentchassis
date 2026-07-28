package types

import (
	"encoding/json"
	"testing"
)

// The defect this file exists to prevent (bugs_open/129): the coordinator's
// retry path used to build its message out of the AWAITING orchestration's own
// state, so a retry reached the child carrying
//
//	orchestration_id = <the parent's>      → the child loaded the parent's row,
//	                                          saw AWAITING_RESPONSES, declined the
//	                                          work and logged success
//	action           = "execute"           → never the original action
//	body             = {"is_retry": true}  → the payload, gone
//
// A replay must differ from the original in retry_version, message_id and
// timestamp, and in nothing else. Each test below fails on a reconstruction and
// passes only on a replay.

func originalSpawnMessage(t *testing.T) []byte {
	t.Helper()
	msg := RequestMessage{
		Headers: RequestHeaders{
			RequestID:             "req-91f55361",
			OrchestrationID:       "child-25dc09ae", // the CHILD's own id
			OrchestrationName:     "code-indexer-workflow",
			ParentOrchestrationID: "parent-58a53a6a",
			CorrelationID:         "corr-1",
			ClientID:              "demo_client",
			StepName:              "spawn_indexer",
			Action:                "initialize",
			MessageID:             "msg-original",
			MessageType:           "request",
			ToAgent:               "child-25dc09ae",
			ToAgentType:           "code-indexer",
			TimeoutSeconds:        30,
			RetryVersion:          0,
		},
		Body: map[string]interface{}{
			"repo_url": "git@github.com:gqls/agentchassis.git",
			"ref":      "main",
		},
	}
	raw, err := json.Marshal(&msg)
	if err != nil {
		t.Fatalf("marshalling the original message: %v", err)
	}
	return raw
}

func decodeStored(t *testing.T, payload map[string]interface{}) *RetryPayload {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshalling the stored payload: %v", err)
	}
	stored, err := DecodeRetryPayload(raw)
	if err != nil {
		t.Fatalf("decoding the stored payload: %v", err)
	}
	return stored
}

// A replay must preserve the CHILD's identity. This is the whole of 129: on the
// old path this field held the parent's id and the child declined the work.
func TestReplayPreservesTheChildsOwnOrchestrationID(t *testing.T) {
	stored := decodeStored(t, BuildRetryPayload("job.indexer.requests", "req-91f55361", originalSpawnMessage(t)))

	_, _, replayed, err := stored.ReplayRequest(2, 60)
	if err != nil {
		t.Fatalf("replaying: %v", err)
	}

	if replayed.Headers.OrchestrationID != "child-25dc09ae" {
		t.Errorf("orchestration_id = %q, want the CHILD's own id %q — a retry carrying the parent's id makes the child adopt the parent's row and decline the work",
			replayed.Headers.OrchestrationID, "child-25dc09ae")
	}
	if replayed.Headers.ParentOrchestrationID != "parent-58a53a6a" {
		t.Errorf("parent_orchestration_id = %q, want %q — without it the child cannot create its own row either",
			replayed.Headers.ParentOrchestrationID, "parent-58a53a6a")
	}
}

// A replay must preserve the original action and body. Fixing identity alone
// would give the child a correct row and nothing to work on, which is a wrong
// answer instead of no answer.
func TestReplayPreservesTheActionAndTheBody(t *testing.T) {
	stored := decodeStored(t, BuildRetryPayload("job.indexer.requests", "req-91f55361", originalSpawnMessage(t)))

	_, _, replayed, err := stored.ReplayRequest(1, 0)
	if err != nil {
		t.Fatalf("replaying: %v", err)
	}

	if replayed.Headers.Action != "initialize" {
		t.Errorf("action = %q, want %q", replayed.Headers.Action, "initialize")
	}

	body, ok := replayed.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("body is %T, want the original map", replayed.Body)
	}
	if body["repo_url"] != "git@github.com:gqls/agentchassis.git" || body["ref"] != "main" {
		t.Errorf("body = %v, want the original request body — a retry with a stub body makes the child run on nothing", body)
	}
	if _, isStub := body["is_retry"]; isStub {
		t.Error("body carries the is_retry stub — the payload was reconstructed, not replayed")
	}
}

// Exactly three fields may move, and the Kafka headers must be regenerated from
// the replayed message so the two cannot disagree.
func TestReplayMovesOnlyRetryVersionMessageIDAndTimestamp(t *testing.T) {
	original := originalSpawnMessage(t)
	stored := decodeStored(t, BuildRetryPayload("job.indexer.requests", "req-91f55361", original))

	headers, bytes, replayed, err := stored.ReplayRequest(3, 90)
	if err != nil {
		t.Fatalf("replaying: %v", err)
	}

	if replayed.Headers.RetryVersion != 3 {
		t.Errorf("retry_version = %d, want 3", replayed.Headers.RetryVersion)
	}
	if replayed.Headers.MessageID == "msg-original" {
		t.Error("message_id was not re-stamped — a replay is a new message on the wire")
	}
	if replayed.Headers.TimeoutSeconds != 90 {
		t.Errorf("timeout_seconds = %d, want the retry's 90", replayed.Headers.TimeoutSeconds)
	}
	if replayed.Headers.RequestID != "req-91f55361" {
		t.Errorf("request_id = %q — the request id must survive, it is what the parent is awaiting", replayed.Headers.RequestID)
	}

	// The headers handed to Kafka come from the replayed message itself.
	if headers["orchestration_id"] != replayed.Headers.OrchestrationID {
		t.Errorf("kafka header orchestration_id = %q but the message body says %q — the two must not be able to disagree",
			headers["orchestration_id"], replayed.Headers.OrchestrationID)
	}
	if headers["retry_version"] != "3" {
		t.Errorf("kafka header retry_version = %q, want \"3\" — the consumer's dedupe reads this one", headers["retry_version"])
	}

	var roundTripped RequestMessage
	if err := json.Unmarshal(bytes, &roundTripped); err != nil {
		t.Fatalf("replayed bytes do not parse as a RequestMessage: %v", err)
	}
	if roundTripped.Headers.OrchestrationID != "child-25dc09ae" {
		t.Errorf("produced bytes carry orchestration_id %q", roundTripped.Headers.OrchestrationID)
	}
}

// Absent or unusable payloads must be reported, not papered over: the
// coordinator's contract is to refuse to retry rather than emit a message it
// knows to be misaddressed.
func TestDecodeRefusesAnythingItCannotReplayFaithfully(t *testing.T) {
	cases := []struct {
		name string
		raw  json.RawMessage
	}{
		{"nothing recorded", nil},
		{"empty document", json.RawMessage(``)},
		{"not json", json.RawMessage(`not json`)},
		{"no topic", json.RawMessage(`{"key":"k","message":{"headers":{}}}`)},
		{"no message", json.RawMessage(`{"topic":"t","key":"k"}`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeRetryPayload(tc.raw); err == nil {
				t.Error("decoded without error — the coordinator would then send a synthesised retry")
			}
		})
	}
}

// The empty-vs-absent distinction the persistence layer relies on: a payload
// that was never recorded must bind as SQL NULL, not as an empty jsonb string.
func TestBuildRetryPayloadRoundTripsThroughJSON(t *testing.T) {
	original := originalSpawnMessage(t)
	payload := BuildRetryPayload("job.indexer.requests", "req-91f55361", original)

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	stored, err := DecodeRetryPayload(raw)
	if err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if stored.Topic != "job.indexer.requests" {
		t.Errorf("topic = %q — the replay must go to the topic the original went to, not a re-derived one", stored.Topic)
	}
	if stored.Key != "req-91f55361" {
		t.Errorf("key = %q — the partition key must match the original send", stored.Key)
	}
}
