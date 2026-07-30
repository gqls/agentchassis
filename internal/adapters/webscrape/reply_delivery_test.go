// FILE: internal/adapters/webscrape/reply_delivery_test.go
package webscrape

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	platformkafka "github.com/gqls/agentchassis/platform/kafka"
	segkafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// bugs_open/133 defect B, as a test: a reply the broker refuses must end as a
// deliverable error, never as a log line and silence.
//
// These drive the real sendSuccessResponse / sendBatchSuccessResponse through a
// scripted producer, so they test the WIRING — that the adapter answers
// FailedUndeliverable with an error response — not just the shared policy,
// which platform/kafka tests separately. An adapter that called DeliverReply
// and ignored its outcome would pass those and fail these.

type recordingProducer struct {
	errs []error
	sent []recordedMessage
}

type recordedMessage struct {
	topic   string
	headers map[string]string
	value   []byte
}

func (r *recordingProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return r.ProduceWithValidation(ctx, topic, headers, key, value)
}

func (r *recordingProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	i := len(r.sent)
	copied := make(map[string]string, len(headers))
	for k, v := range headers {
		copied[k] = v
	}
	r.sent = append(r.sent, recordedMessage{topic: topic, headers: copied, value: value})
	if i < len(r.errs) {
		return r.errs[i]
	}
	return nil
}

func (r *recordingProducer) Close() error { return nil }

var _ platformkafka.Producer = (*recordingProducer)(nil)

func testAdapter(p platformkafka.Producer) *Adapter {
	return &Adapter{ctx: context.Background(), logger: zap.NewNop(), producer: p}
}

// statusesOf pulls the header every response carries, so a test asserts on what
// the caller would actually switch on.
func statusesOf(msgs []recordedMessage) []string {
	out := make([]string, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, m.headers["status"])
	}
	return out
}

func TestSingleScrapeReplyDeliveredCleanlySendsOneMessage(t *testing.T) {
	p := &recordingProducer{}
	a := testAdapter(p)

	a.sendSuccessResponse("req-1", "corr-1", "orch-1", "reply.topic", "client-1", "step-1",
		map[string]interface{}{"markdown_content": "hello"})

	if len(p.sent) != 1 {
		t.Fatalf("sent %d messages, want 1: %v", len(p.sent), statusesOf(p.sent))
	}
	if got := p.sent[0].headers["status"]; got != "complete" {
		t.Errorf("status = %q, want complete", got)
	}
}

// The whole point of the bug: the caller must be answered.
func TestSingleScrapeOversizeReplyDegradesThenSendsErrorIfStillRefused(t *testing.T) {
	p := &recordingProducer{errs: []error{segkafka.MessageSizeTooLarge, segkafka.MessageSizeTooLarge}}
	a := testAdapter(p)

	result := map[string]interface{}{
		"raw_html":         bigString(200000),
		"markdown_content": bigString(200000),
	}
	a.sendSuccessResponse("req-2", "corr-2", "orch-2", "reply.topic", "client-2", "step-2", result)

	if len(p.sent) != 3 {
		t.Fatalf("sent %d messages, want 3 (full, degraded, error): %v", len(p.sent), statusesOf(p.sent))
	}

	// The degraded attempt must actually be smaller, or the resend is a wasted
	// produce that was always going to fail.
	if len(p.sent[1].value) >= len(p.sent[0].value) {
		t.Errorf("degraded reply is %d bytes vs original %d — it did not shrink",
			len(p.sent[1].value), len(p.sent[0].value))
	}

	// And the third message is a real error response the caller can act on.
	final := p.sent[2]
	if final.headers["status"] != "error_recoverable" {
		t.Errorf("final status = %q, want error_recoverable", final.headers["status"])
	}
	if !strings.Contains(string(final.value), "could not be delivered") {
		t.Error("the error response does not say why the caller is getting an error")
	}
	if final.topic != "reply.topic" {
		t.Errorf("error response went to %q, not the reply topic", final.topic)
	}
}

func TestSingleScrapeOversizeReplySucceedsOnTheDegradedResend(t *testing.T) {
	p := &recordingProducer{errs: []error{segkafka.MessageSizeTooLarge}}
	a := testAdapter(p)

	a.sendSuccessResponse("req-3", "corr-3", "orch-3", "reply.topic", "client-3", "step-3",
		map[string]interface{}{"raw_html": bigString(200000)})

	if len(p.sent) != 2 {
		t.Fatalf("sent %d messages, want 2 (full, degraded): %v", len(p.sent), statusesOf(p.sent))
	}
	if p.sent[1].headers["status"] != "complete" {
		t.Errorf("the degraded reply must still be a success response, got %q", p.sent[1].headers["status"])
	}
	// No error response — the caller was answered.
	for _, m := range p.sent {
		if m.headers["status"] == "error_recoverable" {
			t.Error("an error response went out even though the degraded reply was accepted")
		}
	}

	// The degraded reply must SAY it is degraded, or the caller cannot tell a
	// stub from a page that really is that short.
	var envelope map[string]interface{}
	if err := json.Unmarshal(p.sent[1].value, &envelope); err != nil {
		t.Fatalf("degraded reply is not valid JSON: %v", err)
	}
	body := envelope["body"].(map[string]interface{})["body"].(map[string]interface{})
	data := body["data"].(map[string]interface{})
	if data["degraded_for_transport"] != true {
		t.Error("degraded reply does not carry degraded_for_transport")
	}
	if _, present := data["raw_html"]; present {
		t.Error("raw_html survived into the degraded reply")
	}
}

// A transient failure must NOT produce an error response — the coordinator's
// retry is the right answer and an error response would end the step early.
// This is the behaviour the old code had for ALL failures, and it is correct
// for exactly this one.
func TestSingleScrapeTransientFailureDoesNotSendAnErrorResponse(t *testing.T) {
	p := &recordingProducer{errs: []error{context.DeadlineExceeded}}
	a := testAdapter(p)

	a.sendSuccessResponse("req-4", "corr-4", "orch-4", "reply.topic", "client-4", "step-4",
		map[string]interface{}{"markdown_content": "hello"})

	if len(p.sent) != 1 {
		t.Fatalf("sent %d messages, want 1 — a transient failure must not degrade or error: %v",
			len(p.sent), statusesOf(p.sent))
	}
}

// No reply topic is the one case where silence is right: there is nowhere to
// send anything. It must not panic or attempt a produce.
func TestSingleScrapeNoReplyTopicProducesNothing(t *testing.T) {
	p := &recordingProducer{}
	a := testAdapter(p)

	a.sendSuccessResponse("req-5", "corr-5", "orch-5", "", "client-5", "step-5",
		map[string]interface{}{"markdown_content": "hello"})

	if len(p.sent) != 0 {
		t.Errorf("sent %d messages with no reply topic", len(p.sent))
	}
}

// A non-map result cannot be stripped. It must still reach the caller as an
// error rather than vanishing.
func TestSingleScrapeUndegradableResultStillAnswersTheCaller(t *testing.T) {
	p := &recordingProducer{errs: []error{segkafka.MessageSizeTooLarge}}
	a := testAdapter(p)

	a.sendSuccessResponse("req-6", "corr-6", "orch-6", "reply.topic", "client-6", "step-6",
		"a string, not a map")

	if len(p.sent) != 2 {
		t.Fatalf("sent %d messages, want 2 (full, error): %v", len(p.sent), statusesOf(p.sent))
	}
	if p.sent[1].headers["status"] != "error_recoverable" {
		t.Errorf("final status = %q, want error_recoverable", p.sent[1].headers["status"])
	}
}

// ---------------------------------------------------------------------------
// The batch path, repointed at the shared policy — bugs_closed/062 must not
// regress. These assert the behaviour its own fix established.
// ---------------------------------------------------------------------------

func batchResult() map[string]interface{} {
	return map[string]interface{}{
		"results": []map[string]interface{}{
			{"url": "https://a.example", "raw_html": bigString(200000), "content": bigString(200000)},
		},
		"success_count": 1,
		"error_count":   0,
		"total_count":   1,
	}
}

func TestBatchReplyStillDegradesThenErrorsAfterTheRepoint(t *testing.T) {
	p := &recordingProducer{errs: []error{segkafka.MessageSizeTooLarge, segkafka.MessageSizeTooLarge}}
	a := testAdapter(p)

	a.sendBatchSuccessResponse("req-b1", "corr-b1", "orch-b1", "reply.topic",
		"client-b1", "step-b1", "stepid-b1", batchResult())

	if len(p.sent) != 3 {
		t.Fatalf("sent %d messages, want 3 (full, stripped, error): %v", len(p.sent), statusesOf(p.sent))
	}
	if len(p.sent[1].value) >= len(p.sent[0].value) {
		t.Errorf("stripped batch reply did not shrink: %d vs %d", len(p.sent[1].value), len(p.sent[0].value))
	}
	if p.sent[2].headers["status"] != "error_recoverable" {
		t.Errorf("final batch status = %q, want error_recoverable", p.sent[2].headers["status"])
	}
	if p.sent[2].headers["is_error"] != "true" {
		t.Error("batch error response missing is_error header")
	}
}

func TestBatchTransientFailureStillDoesNotErrorOrStrip(t *testing.T) {
	p := &recordingProducer{errs: []error{context.DeadlineExceeded}}
	a := testAdapter(p)

	a.sendBatchSuccessResponse("req-b2", "corr-b2", "orch-b2", "reply.topic",
		"client-b2", "step-b2", "stepid-b2", batchResult())

	if len(p.sent) != 1 {
		t.Fatalf("sent %d messages, want 1: %v", len(p.sent), statusesOf(p.sent))
	}
}

func TestBatchOversizeSucceedsOnStrippedResend(t *testing.T) {
	p := &recordingProducer{errs: []error{segkafka.MessageSizeTooLarge}}
	a := testAdapter(p)

	a.sendBatchSuccessResponse("req-b3", "corr-b3", "orch-b3", "reply.topic",
		"client-b3", "step-b3", "stepid-b3", batchResult())

	if len(p.sent) != 2 {
		t.Fatalf("sent %d messages, want 2: %v", len(p.sent), statusesOf(p.sent))
	}
	if p.sent[1].headers["status"] != "complete" {
		t.Errorf("stripped batch reply status = %q, want complete", p.sent[1].headers["status"])
	}
	if strings.Contains(string(p.sent[1].value), `"raw_html"`) {
		t.Error("raw_html survived the batch strip")
	}
}
