// FILE: internal/adapters/webscrape/batch_handler_test.go
//
// The bugs_open/062 properties: a cut must be VISIBLE (truncated: true),
// the oversize-retry strip must actually shrink what did not fit, and the
// message-too-large classification must never catch a transient produce
// failure (a transient keeps the coordinator-retry path; only the
// deterministic refusal gets degrade-and-resend).

package webscrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/segmentio/kafka-go"
)

func TestTruncateBatchResultMarksTheCut(t *testing.T) {
	big := strings.Repeat("x", 200)
	result := map[string]interface{}{
		"content":  big,
		"markdown": big,
		"title":    "untouched",
	}
	truncateBatchResult(result, 100)

	if got := result["content"].(string); len(got) != 100 {
		t.Errorf("content length = %d, want 100", len(got))
	}
	if got := result["markdown"].(string); len(got) != 100 {
		t.Errorf("markdown length = %d, want 100", len(got))
	}
	if result["truncated"] != true {
		t.Error("a cut result must carry truncated: true — an invisible cut is the damage")
	}
	if result["title"] != "untouched" {
		t.Errorf("non-content field was modified: %v", result["title"])
	}
}

func TestTruncateBatchResultNoOpUnderCap(t *testing.T) {
	result := map[string]interface{}{"content": "small"}
	truncateBatchResult(result, 100)

	if result["content"] != "small" {
		t.Errorf("under-cap content was modified: %v", result["content"])
	}
	if _, marked := result["truncated"]; marked {
		t.Error("an uncut result must NOT be marked truncated")
	}
}

func TestStripBatchResultsForRetryDropsRawAndShrinks(t *testing.T) {
	big := strings.Repeat("y", oversizeStripContentCap*3)
	results := []map[string]interface{}{
		{"content": big, "raw_html": big, "html_content": big, "url": "https://example.com"},
	}
	stripBatchResultsForRetry(results)

	r := results[0]
	if _, has := r["raw_html"]; has {
		t.Error("raw_html must be dropped on the oversize retry — it is what did not fit")
	}
	if _, has := r["html_content"]; has {
		t.Error("html_content must be dropped on the oversize retry")
	}
	if got := r["content"].(string); len(got) != oversizeStripContentCap {
		t.Errorf("content length = %d, want the strip cap %d", len(got), oversizeStripContentCap)
	}
	if r["truncated"] != true {
		t.Error("the stripped result must carry truncated: true")
	}
	if r["url"] != "https://example.com" {
		t.Errorf("identity field was modified: %v", r["url"])
	}
}

// The envelope's contract with the chassis: processor.go unmarshals the
// message value into types.ResponseMessage, whose ResponseHeaders.IsComplete
// is a real bool. This handler shipped from birth with `"is_complete":
// "true"` (a string), which failed that unmarshal and dropped EVERY batch
// response ever sent (bugs_open/062 defect 4). This test round-trips the
// real envelope through the real type, so the contract can never silently
// regress to the string form again.
func TestBatchSuccessEnvelopeUnmarshalsIntoResponseMessage(t *testing.T) {
	envelope := buildBatchSuccessEnvelope(
		"req-1", "corr-1", "orch-1", "client-1", "scrape_pages", "step-1",
		map[string]interface{}{
			"results":       []map[string]interface{}{{"index": 0, "url": "https://example.com", "content": "x", "success": true}},
			"success_count": 1, "error_count": 0, "total_count": 1,
		})

	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	var parsed types.ResponseMessage
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("the envelope MUST unmarshal into types.ResponseMessage — the chassis drops it otherwise: %v", err)
	}
	if !parsed.Headers.IsComplete {
		t.Error("is_complete must arrive as a true bool")
	}
	if parsed.Headers.InResponseToRequestID != "req-1" {
		t.Errorf("in_response_to_request_id = %q, want req-1", parsed.Headers.InResponseToRequestID)
	}
	if parsed.Headers.Status != "complete" {
		t.Errorf("status = %q, want complete", parsed.Headers.Status)
	}

	// The regression this guards against, demonstrated: the string form
	// fails the typed unmarshal outright.
	bad := []byte(`{"headers":{"is_complete":"true"},"body":{}}`)
	if err := json.Unmarshal(bad, &parsed); err == nil {
		t.Error("expected the string form of is_complete to fail the typed unmarshal — if this passes, the type contract changed and this test needs rethinking")
	}
}

func TestIsKafkaMessageTooLargeClassification(t *testing.T) {
	// Typed: the broker's error code, wrapped the way the producer wraps it
	// (%w), must classify via errors.Is.
	if !isKafkaMessageTooLarge(fmt.Errorf("failed to write message to kafka: %w", kafka.MessageSizeTooLarge)) {
		t.Error("the wrapped typed broker refusal must classify as message-too-large")
	}
	// Typed: the writer's client-side pre-send detection, via errors.As.
	if !isKafkaMessageTooLarge(fmt.Errorf("write: %w", kafka.MessageTooLargeError{})) {
		t.Error("the wrapped client-side MessageTooLargeError must classify as message-too-large")
	}
	// Fallback: the failure as a bare string (composite shapes like
	// kafka.WriteErrors can break the unwrap chain — observed live form,
	// bugs_open/062).
	refusal := errors.New("failed to write message to kafka: [10] Message Size Too Large: the server has a configurable maximum message size to avoid unbounded memory allocation and the client attempted to produce a message larger than this maximum")
	if !isKafkaMessageTooLarge(refusal) {
		t.Error("the broker's size refusal as a string must classify as message-too-large")
	}

	// Transient failures must NOT classify — they keep the coordinator-retry
	// path, where a resend can genuinely succeed.
	for _, transient := range []error{
		errors.New("dial tcp: connection refused"),
		errors.New("context deadline exceeded"),
		nil,
	} {
		if isKafkaMessageTooLarge(transient) {
			t.Errorf("transient error %v must not classify as message-too-large", transient)
		}
	}
}
