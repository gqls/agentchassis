// FILE: internal/adapters/webscrape/batch_handler_test.go
//
// The bugs_open/062 properties: a cut must be VISIBLE (truncated: true),
// the oversize-retry strip must actually shrink what did not fit, and the
// message-too-large classification must never catch a transient produce
// failure (a transient keeps the coordinator-retry path; only the
// deterministic refusal gets degrade-and-resend).

package webscrape

import (
	"errors"
	"strings"
	"testing"
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

func TestIsKafkaMessageTooLargeClassification(t *testing.T) {
	// The broker's refusal, as the kafka client actually wraps it (observed
	// live, bugs_open/062).
	refusal := errors.New("failed to write message to kafka: [10] Message Size Too Large: the server has a configurable maximum message size to avoid unbounded memory allocation and the client attempted to produce a message larger than this maximum")
	if !isKafkaMessageTooLarge(refusal) {
		t.Error("the broker's size refusal must classify as message-too-large")
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
