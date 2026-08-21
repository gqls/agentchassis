// FILE: platform/kafka/produce_retry_test.go
//
// bugs_open/040-kafka-dial: the opt-in bounded produce retry.
//
// The most important test in this file is the one that asserts NOTHING CHANGED:
// TestDeliverReplyDoesNotRetryByDefault. This is new behaviour on a seam every
// reply in the fleet crosses, and it ships opt-in with the unsafe side off (owner
// ruling 2026-08-02 §2) — so the default's byte-for-byte sameness is the promise,
// and a promise nobody tests is a comment.
package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go/protocol"
	"go.uber.org/zap"
)

// fastPolicy keeps the retry semantics but not the wall-clock.
var fastPolicy = RetryPolicy{Attempts: 4, BaseDelay: time.Millisecond, MaxDelay: 2 * time.Millisecond}

// THE DEFAULT-OFF PIN. Five callers do not pass WithRetry and must behave exactly
// as they did before this change: one produce attempt, FailedTransient.
//
// ⚠ WHAT THIS PROVES, AND WHAT IT DOES NOT — stated because the obvious claim is
// wrong and I checked rather than assuming. The default-off holds by TWO
// mechanisms IN SERIES: the cfg.retry flag, and the zero-value RetryPolicy whose
// Attempts of 0 is clamped to 1 by retrySend. Removing the flag check ALONE
// therefore changes nothing observable — applied as a mutation, it survives this
// test — because the zero policy still yields a single attempt.
//
// That is a real belt-and-braces property and it is why the default is hard to
// break by accident. It also means this test cannot discriminate the two, so
// what it actually pins is the OBSERVABLE promise ("no option, one attempt"),
// which is the promise the five existing callers depend on. The clamp's own half
// is pinned separately by TestRetryAlwaysSendsAtLeastOnce.
//
// The mutation that DOES fail this test: default cfg.retry to true AND give
// deliverConfig a non-zero default policy — i.e. actually turning the feature on
// by default, which is the thing the owner ruling forbids.
func TestDeliverReplyDoesNotRetryByDefault(t *testing.T) {
	p := &scriptedProducer{errs: []error{protocol.ErrNoLeader, nil}}

	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(),
		"system.generic.responses", map[string]string{}, []byte("k"), []byte("v"), nil)

	if outcome != FailedTransient {
		t.Errorf("outcome = %v, want FailedTransient", outcome)
	}
	if err == nil {
		t.Error("a failed produce must return its error")
	}
	if p.calls != 1 {
		t.Fatalf("produce called %d times without WithRetry, want 1 - the opt-in default is not off, and five existing callers just changed behaviour", p.calls)
	}
}

// Opted in, the same fail-then-succeed sequence is delivered on attempt 2.
//
// Mutation that must fail this: delete the retrySend loop — the outcome reverts
// to FailedTransient with one call.
func TestDeliverReplyWithRetryRecoversOnTheSecondAttempt(t *testing.T) {
	p := &scriptedProducer{errs: []error{protocol.ErrNoLeader, nil}}

	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(),
		"system.generic.responses", map[string]string{}, []byte("k"), []byte("v"), nil,
		WithRetry(fastPolicy))

	if err != nil {
		t.Fatalf("delivery failed despite a succeeding second attempt: %v", err)
	}
	if outcome != Delivered {
		t.Errorf("outcome = %v, want Delivered", outcome)
	}
	if p.calls != 2 {
		t.Errorf("produce called %d times, want 2", p.calls)
	}
}

// A deterministic error must stop the loop immediately: resending the same bytes
// is provably pointless, and burning the budget proving it delays the remedy that
// would work (an error response, or a degrade).
func TestRetryStopsImmediatelyOnDeterministicErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want DeliveryOutcome
	}{
		// bugs_open/274: the headers are wrong by construction, so no retry can
		// ever pass. Classified permanent, not transient.
		{"validation refusal", ErrMessageValidationFailed, FailedUndeliverable},
		// The broker's size limit does not move between attempts; DeliverReply's
		// answer is to DEGRADE, which is a different remedy and stays its own.
		{"too large with no degrade", errors.New(tooLargeMsg), FailedUndeliverable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &scriptedProducer{errs: []error{tc.err, nil}}
			outcome, _ := DeliverReply(context.Background(), p, zap.NewNop(),
				"system.generic.responses", map[string]string{}, []byte("k"), []byte("v"), nil,
				WithRetry(fastPolicy))

			if outcome != tc.want {
				t.Errorf("outcome = %v, want %v", outcome, tc.want)
			}
			if p.calls != 1 {
				t.Errorf("produce called %d times on a deterministic error, want 1 - the retry budget was spent proving what the error already said", p.calls)
			}
		})
	}
}

// Oversize + degrade still degrades and resends exactly once, with the retry on:
// the degraded send goes through the same policy, and the ORIGINAL send must not
// have been retried, because too-large is deterministic.
func TestRetryDoesNotDisturbTheDegradePath(t *testing.T) {
	p := &scriptedProducer{errs: []error{errors.New(tooLargeMsg), nil}}

	outcome, err := DeliverReply(context.Background(), p, zap.NewNop(),
		"system.generic.responses", map[string]string{}, []byte("k"), []byte("original"), func() ([]byte, error) {
			return []byte("small"), nil
		},
		WithRetry(fastPolicy))

	if err != nil {
		t.Fatalf("degrade path failed: %v", err)
	}
	if outcome != DeliveredDegraded {
		t.Errorf("outcome = %v, want DeliveredDegraded", outcome)
	}
	if p.calls != 2 {
		t.Fatalf("produce called %d times, want 2 (original + degraded) - the oversize send was retried, which it must never be", p.calls)
	}
	if string(p.sent[1]) != "small" {
		t.Errorf("second send carried %q, want the degraded payload", p.sent[1])
	}
}

// A cancelled context stops the loop, and the PRODUCE error is returned rather
// than the context error, so the caller's classification is unchanged.
func TestRetryStopsOnCancelledContextAndReturnsTheProduceError(t *testing.T) {
	sentinel := errors.New("broker unreachable")
	p := &scriptedProducer{errs: []error{sentinel, sentinel, sentinel, sentinel}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ProduceWithRetry(ctx, p, zap.NewNop(), fastPolicy,
		"system.generic.responses", map[string]string{}, []byte("k"), []byte("v"))

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want the produce error verbatim so the caller's classification is unchanged", err)
	}
	if p.calls != 1 {
		t.Errorf("produce called %d times under a cancelled context, want 1", p.calls)
	}
}

// THE ERROR-TEXT CONTRACT. Every caller wraps this error in its own message and
// platform/errors' transient classifier matches needles inside them, so the retry
// must return the LAST attempt's error verbatim — adding a layer here silently
// changes what all of that sees.
func TestRetryReturnsTheLastErrorVerbatim(t *testing.T) {
	last := errors.New("the final failure, exactly as the writer phrased it")
	p := &scriptedProducer{errs: []error{
		errors.New("first"), errors.New("second"), errors.New("third"), last,
	}}

	err := ProduceWithRetry(context.Background(), p, zap.NewNop(), fastPolicy,
		"system.generic.responses", map[string]string{}, []byte("k"), []byte("v"))

	if err == nil {
		t.Fatal("an exhausted retry must return an error")
	}
	if err.Error() != last.Error() {
		t.Errorf("err = %q, want the last attempt's error verbatim %q", err, last)
	}
	if p.calls != 4 {
		t.Errorf("produce called %d times, want 4 (the policy's Attempts)", p.calls)
	}
}

// Attempts < 1 must mean one attempt, not zero: a policy that silently sends
// nothing is the worst possible failure mode for this helper.
func TestRetryAlwaysSendsAtLeastOnce(t *testing.T) {
	for _, attempts := range []int{0, -1, 1} {
		p := &scriptedProducer{}
		if err := ProduceWithRetry(context.Background(), p, zap.NewNop(),
			RetryPolicy{Attempts: attempts}, "t", map[string]string{}, nil, []byte("v")); err != nil {
			t.Fatalf("Attempts=%d: %v", attempts, err)
		}
		if p.calls != 1 {
			t.Errorf("Attempts=%d produced %d times, want exactly 1", attempts, p.calls)
		}
	}
}

// The backoff is jittered because these failures are CORRELATED — a leadership
// election fails every producer in the fleet at once, and an unjittered backoff
// marches them all back onto the broker in the same instant.
func TestRetryDelayIsJitteredAndCapped(t *testing.T) {
	p := RetryPolicy{Attempts: 6, BaseDelay: 100 * time.Millisecond, MaxDelay: 400 * time.Millisecond}

	seen := map[time.Duration]struct{}{}
	for i := 0; i < 50; i++ {
		seen[p.retryDelay(2)] = struct{}{}
	}
	if len(seen) < 2 {
		t.Error("retryDelay returned the same value 50 times - the jitter is not jittering, and a correlated failure will produce a thundering herd")
	}
	for i := 1; i <= 6; i++ {
		if d := p.retryDelay(i); d > p.MaxDelay {
			t.Errorf("retryDelay(%d) = %v, exceeds MaxDelay %v", i, d, p.MaxDelay)
		}
	}
}
