// FILE: platform/kafka/reply_delivery_test.go
package kafka

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// scriptedProducer returns a pre-set error per call and records what it was
// asked to send, so a test can assert both the outcome and the number of
// produce attempts. The attempt COUNT is half the contract: "degrade and
// resend ONCE" is a claim about how many times the broker is asked.
type scriptedProducer struct {
	errs  []error
	sent  [][]byte
	calls int
}

func (s *scriptedProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return s.ProduceWithValidation(ctx, topic, headers, key, value)
}

func (s *scriptedProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	s.sent = append(s.sent, value)
	i := s.calls
	s.calls++
	if i < len(s.errs) {
		return s.errs[i]
	}
	return nil
}

func (s *scriptedProducer) Close() error { return nil }

// scriptedProducer must satisfy the real interface, or these tests prove
// nothing about the real call path.
var _ Producer = (*scriptedProducer)(nil)

const tooLargeMsg = "Message Size Too Large: the server has a configurable maximum message size"

func TestIsMessageTooLargeClassification(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unrelated", errors.New("connection refused"), false},
		{"sentinel", kafka.MessageSizeTooLarge, true},
		{"sentinel wrapped", fmt.Errorf("failed to write message to kafka: %w", kafka.MessageSizeTooLarge), true},
		{"typed client-side", kafka.MessageTooLargeError{}, true},
		{"typed wrapped", fmt.Errorf("produce: %w", kafka.MessageTooLargeError{}), true},
		{"substring only", errors.New(tooLargeMsg), true},
		// The producer's real wrapping, which is how this arrives in practice.
		{"substring wrapped", fmt.Errorf("failed to write message to kafka: %w", errors.New(tooLargeMsg)), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsMessageTooLarge(c.err); got != c.want {
				t.Errorf("IsMessageTooLarge(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestDeliverReplyCleanSendDoesNotDegrade(t *testing.T) {
	p := &scriptedProducer{}
	degradeCalled := false
	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(), "t", nil, nil,
		[]byte("full"), func() ([]byte, error) { degradeCalled = true; return []byte("small"), nil })

	if outcome != Delivered || err != nil {
		t.Fatalf("outcome=%v err=%v, want delivered/nil", outcome, err)
	}
	if degradeCalled {
		t.Error("degrader ran on a successful produce")
	}
	if p.calls != 1 {
		t.Errorf("produce attempts = %d, want 1", p.calls)
	}
	if !outcome.Answered() {
		t.Error("Delivered must count as answered")
	}
}

// A transient failure must NOT be degraded. Shrinking a reply that was never
// too big loses payload for no reason, and the caller's existing retry can
// still deliver the full one.
func TestDeliverReplyTransientFailureIsNotDegraded(t *testing.T) {
	p := &scriptedProducer{errs: []error{errors.New("broker unreachable")}}
	degradeCalled := false
	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(), "t", nil, nil,
		[]byte("full"), func() ([]byte, error) { degradeCalled = true; return []byte("small"), nil })

	if outcome != FailedTransient {
		t.Fatalf("outcome = %v, want FailedTransient", outcome)
	}
	if err == nil {
		t.Error("transient failure must return the error for the caller to log/retry on")
	}
	if degradeCalled {
		t.Error("degrader ran on a transient failure — it must only run on a size refusal")
	}
	if p.calls != 1 {
		t.Errorf("produce attempts = %d, want 1 (no resend)", p.calls)
	}
	if outcome.Answered() {
		t.Error("FailedTransient must not count as answered")
	}
}

func TestDeliverReplyOversizeDegradesAndResendsExactlyOnce(t *testing.T) {
	p := &scriptedProducer{errs: []error{kafka.MessageSizeTooLarge}}
	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(), "t", nil, nil,
		[]byte("the full reply"), func() ([]byte, error) { return []byte("stub"), nil })

	if outcome != DeliveredDegraded || err != nil {
		t.Fatalf("outcome=%v err=%v, want DeliveredDegraded/nil", outcome, err)
	}
	if p.calls != 2 {
		t.Fatalf("produce attempts = %d, want exactly 2", p.calls)
	}
	if string(p.sent[1]) != "stub" {
		t.Errorf("resent %q, want the degraded bytes", p.sent[1])
	}
	if !outcome.Answered() {
		t.Error("DeliveredDegraded must count as answered — the caller did get a reply")
	}
}

// The outcome that exists to stop the silence: both produces refused.
func TestDeliverReplyUndeliverableWhenEvenDegradedIsRefused(t *testing.T) {
	p := &scriptedProducer{errs: []error{kafka.MessageSizeTooLarge, kafka.MessageSizeTooLarge}}
	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(), "t", nil, nil,
		[]byte("the full reply"), func() ([]byte, error) { return []byte("stub"), nil })

	if outcome != FailedUndeliverable {
		t.Fatalf("outcome = %v, want FailedUndeliverable", outcome)
	}
	if err == nil {
		t.Error("undeliverable must return an error to put in the error response")
	}
	if p.calls != 2 {
		t.Errorf("produce attempts = %d, want 2 — it must not keep retrying a deterministic refusal", p.calls)
	}
	if outcome.Answered() {
		t.Error("FailedUndeliverable must not count as answered — this is the silent-starve guard")
	}
}

func TestDeliverReplyNoDegraderGoesStraightToUndeliverable(t *testing.T) {
	p := &scriptedProducer{errs: []error{kafka.MessageSizeTooLarge}}
	outcome, _ := DeliverReply(context.Background(), p, zap.NewNop(), "t", nil, nil,
		[]byte("full"), nil)

	if outcome != FailedUndeliverable {
		t.Fatalf("outcome = %v, want FailedUndeliverable", outcome)
	}
	if p.calls != 1 {
		t.Errorf("produce attempts = %d, want 1 — nothing to resend", p.calls)
	}
}

// A degrader that fails, or that has nothing left to send, must not be treated
// as a delivery.
func TestDeliverReplyDegraderFailureIsUndeliverable(t *testing.T) {
	for _, c := range []struct {
		name    string
		degrade ReplyDegrader
	}{
		{"degrader errors", func() ([]byte, error) { return nil, errors.New("marshal failed") }},
		{"degrader returns nothing", func() ([]byte, error) { return nil, nil }},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := &scriptedProducer{errs: []error{kafka.MessageSizeTooLarge}}
			outcome, err := DeliverReply(context.Background(), p, zap.NewNop(), "t", nil, nil,
				[]byte("full"), c.degrade)
			if outcome != FailedUndeliverable {
				t.Fatalf("outcome = %v, want FailedUndeliverable", outcome)
			}
			if err == nil {
				t.Error("want the original size error preserved for the error response")
			}
			if p.calls != 1 {
				t.Errorf("produce attempts = %d, want 1 — nothing sendable was produced", p.calls)
			}
		})
	}
}

// Guard rails: a nil producer is a programming error, not a silent success.
func TestDeliverReplyNilProducerIsNotASuccess(t *testing.T) {
	outcome, err := DeliverReply(context.Background(), nil, nil, "t", nil, nil, []byte("x"), nil)
	if outcome.Answered() || err == nil {
		t.Fatalf("nil producer gave outcome=%v err=%v; must not read as answered", outcome, err)
	}
}

// Every outcome must have a distinct name, so a log line naming one is
// unambiguous evidence in an incident.
func TestDeliveryOutcomeNamesAreDistinct(t *testing.T) {
	seen := map[string]bool{}
	for _, o := range []DeliveryOutcome{Delivered, DeliveredDegraded, FailedTransient, FailedUndeliverable} {
		s := o.String()
		if s == "unknown" {
			t.Errorf("outcome %d has no name", int(o))
		}
		if seen[s] {
			t.Errorf("duplicate outcome name %q", s)
		}
		seen[s] = true
	}
}
