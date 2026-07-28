package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RetryPayloadKey is the reserved key under which an action that sends an
// awaited request records the message it actually produced, so the coordinator
// can REPLAY it on timeout instead of reconstructing one.
//
// bugs_open/129. The coordinator used to build the retry out of the AWAITING
// orchestration's own state — which handed the child the PARENT's
// orchestration_id (the child then loaded the parent's row, saw
// AWAITING_RESPONSES, declined the work and logged success), dropped the body
// entirely, and forced Action:"execute" regardless of what was originally sent.
// Measured 2026-07-28: 430 of 430 retried requests in 14 days took that path and
// 294 exhausted their budget.
//
// It travels in the action's result map alongside the other retry-routing keys
// the coordinator already reads from there (`requests_topic`, `responses_topic`,
// `target_agent_type`), and createAwaitedRequest lifts it onto the awaited
// request. An action that sends an awaited request and does NOT set it gets no
// retries at all — the coordinator refuses rather than emitting a poisoned
// message — so a missed site fails loudly and countably instead of corrupting a
// child.
const RetryPayloadKey = "retry_payload"

// RetryPayload is the exact producer.Produce call that was made for an awaited
// request. Persisted as awaited_requests.request_payload (jsonb, nullable).
type RetryPayload struct {
	Topic   string          `json:"topic"`
	Key     string          `json:"key"`
	Message json.RawMessage `json:"message"`
}

// BuildRetryPayload records a produced request message for later replay.
// messageBytes must be the exact bytes handed to the producer, so that a replay
// differs from the original in nothing but retry_version, message_id and
// timestamp.
func BuildRetryPayload(topic, key string, messageBytes []byte) map[string]interface{} {
	return map[string]interface{}{
		"topic":   topic,
		"key":     key,
		"message": json.RawMessage(messageBytes),
	}
}

// DecodeRetryPayload parses a stored payload back into the arguments of the
// original send.
func DecodeRetryPayload(raw json.RawMessage) (*RetryPayload, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("no stored request payload")
	}
	var p RetryPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("stored request payload is not decodable: %w", err)
	}
	if p.Topic == "" || len(p.Message) == 0 {
		return nil, fmt.Errorf("stored request payload is incomplete (topic=%q, message_bytes=%d)", p.Topic, len(p.Message))
	}
	return &p, nil
}

// ReplayRequest re-stamps a stored request message as retry attempt
// retryVersion and returns the headers and bytes to produce. It is the ONLY
// sanctioned way to build a retry: the identity, action and body all come from
// the original message, and the headers handed to Kafka are regenerated from
// that same message so the two cannot disagree.
//
// timeoutSeconds <= 0 leaves the original timeout in place.
func (p *RetryPayload) ReplayRequest(retryVersion, timeoutSeconds int) (map[string]string, []byte, *RequestMessage, error) {
	var msg RequestMessage
	if err := json.Unmarshal(p.Message, &msg); err != nil {
		return nil, nil, nil, fmt.Errorf("stored request message is not a RequestMessage: %w", err)
	}

	msg.Headers.RetryVersion = retryVersion
	msg.Headers.MessageID = uuid.New().String()
	msg.Headers.Timestamp = time.Now()
	if timeoutSeconds > 0 {
		msg.Headers.TimeoutSeconds = timeoutSeconds
	}

	bytes, err := json.Marshal(&msg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal replayed request: %w", err)
	}
	return msg.Headers.ToMap(), bytes, &msg, nil
}
