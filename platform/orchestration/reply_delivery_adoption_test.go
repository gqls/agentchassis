package orchestration

import (
	"context"
	"strings"
	"testing"

	segkafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// bugs_open/158 item 1 — owner ruling 2026-08-03, option (b): adopt
// platform/kafka.DeliverReply at the plumbing sites first, because every agent
// inherits them.
//
// THESE TESTS LIVE HERE, NOT IN platform/kafka, ON PURPOSE. 158's landmine:
//
//	"FailedUndeliverable is a return value a caller can ignore. An adapter that
//	 calls DeliverReply and drops the outcome compiles, passes every
//	 platform/kafka test, and reintroduces the silent starve unchanged."
//
// So the thing under test is not that DeliverReply reports the refusal — it is
// that THIS caller answers it. Each test asserts a real message goes out.

// sizeLimitedProducer refuses any payload over maxBytes with the broker's
// size error and records everything it accepts.
//
// Failing on SIZE rather than unconditionally is what makes the test faithful:
// the whole scenario is a large success payload being refused while the small
// failure notice that replaces it gets through. A producer that failed every
// call could not tell "answered with an error" from "not answered at all",
// which is the only distinction these tests exist to make.
type sizeLimitedProducer struct {
	maxBytes int
	accepted []struct {
		Topic string
		Value []byte
	}
	refused int
}

func (p *sizeLimitedProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	if len(value) > p.maxBytes {
		p.refused++
		return segkafka.MessageSizeTooLarge
	}
	p.accepted = append(p.accepted, struct {
		Topic string
		Value []byte
	}{topic, value})
	return nil
}

func (p *sizeLimitedProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return p.Produce(ctx, topic, headers, key, value)
}

func (p *sizeLimitedProducer) Close() error { return nil }

// TestNotifyParentOfSuccessAnswersAnUndeliverableReply is the behaviour the fix
// exists for: the child finished, its result is too large for the bus, and the
// parent must be told SOMETHING. Before the fix this logged and returned, and
// the parent waited out its timeout with the cause visible only in this pod's
// logs.
func TestNotifyParentOfSuccessAnswersAnUndeliverableReply(t *testing.T) {
	prod := &sizeLimitedProducer{maxBytes: 2048}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop()}

	state := &OrchestrationState{
		OrchestrationID: "orch-158-b",
		CollectedData: map[string]interface{}{
			"__parent_responses_topic__": "system.agent.parent.responses",
			"__reply_to_request_id__":    "req-158",
			// Large enough that the success response cannot fit, while the
			// failure notice that replaces it comfortably can.
			"big_output": strings.Repeat("x", 8192),
		},
	}

	s.notifyParentOfSuccess(context.Background(), state)

	if prod.refused == 0 {
		t.Fatalf("test is vacuous — the success reply was not refused, so the answering branch never ran")
	}
	if len(prod.accepted) == 0 {
		t.Fatalf("the parent was told NOTHING: %d produce(s) refused and no replacement sent — "+
			"this is exactly the silence bugs_closed/062 forbids", prod.refused)
	}
	got := string(prod.accepted[len(prod.accepted)-1].Value)
	if !strings.Contains(got, "could not be delivered") {
		t.Errorf("the replacement must say WHY the parent is hearing failure after a completed workflow; got: %.400s", got)
	}
}

// TestNotifyParentOfSuccessStaysQuietWhenDeliverySucceeds is the negative
// control: nothing about the happy path may change, or every existing
// completion grows a spurious failure notice.
func TestNotifyParentOfSuccessStaysQuietWhenDeliverySucceeds(t *testing.T) {
	prod := &sizeLimitedProducer{maxBytes: 1 << 20}
	s := &SagaCoordinator{producer: prod, logger: zap.NewNop()}

	state := &OrchestrationState{
		OrchestrationID: "orch-158-b-ok",
		CollectedData: map[string]interface{}{
			"__parent_responses_topic__": "system.agent.parent.responses",
			"__reply_to_request_id__":    "req-158-ok",
			"small_output":               "fine",
		},
	}

	s.notifyParentOfSuccess(context.Background(), state)

	if prod.refused != 0 {
		t.Fatalf("nothing should have been refused at a 1MB limit; refused=%d", prod.refused)
	}
	if len(prod.accepted) != 1 {
		t.Fatalf("a successful completion must send exactly ONE message, got %d", len(prod.accepted))
	}
	if strings.Contains(string(prod.accepted[0].Value), "could not be delivered") {
		t.Errorf("a delivered success must not carry the failure wording")
	}
}

// NOT TESTED HERE, and stated rather than quietly omitted: the third adopted
// site, TimeoutMonitor.sendTimeoutResponse (helpers.go), reads orchestration
// state from the database (StateRepository.GetState) BEFORE it produces, so a
// unit test of its delivery branch needs a live DB or a repository fake that
// does not exist in this package yet. Its change is the smallest of the three —
// same call, honest outcome logging, no fallback channel to choose — but it is
// covered by review and by scripts/pattern-check.py, NOT by a test. Anyone
// adding a repository fake here should finish the job.
