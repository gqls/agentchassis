// FILE: platform/kafka/produce_outcome_test.go
//
// bugs_open/040-kafka-dial, the produce side. The dial counters see a CONNECTION
// fail; until this metric nothing had ever seen a WRITE fail — and writes were
// failing. [MEASURED 2026-08-21] over the retained agent_error_log window: 63
// "Kafka write errors" plus 40 "topic partition has no leader" rows across 93
// distinct orchestrations, recurring most days since 08-10.
//
// Two things are pinned here, and the second is the one that could take the
// cluster down if it regressed:
//  1. the outcome vocabulary, against the LIVE error shapes verbatim;
//  2. topicClass's cardinality guard — per-job topics MUST collapse.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"testing"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
)

func TestClassifyProduceErr(t *testing.T) {
	// The two shapes actually seen in production, quoted from agent_error_log.
	liveNoLeader := kafkago.WriteErrors{protocol.ErrNoLeader}
	wrappedNoLeader := fmt.Errorf("failed to write message to kafka: %w", protocol.ErrNoLeader)

	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil is ok", nil, produceOutcomeOK},

		// protocol.ErrNoLeader is a BARE STRING TYPE with no Temporary() method,
		// so kafka-go's writer breaks out of its retry loop after one attempt
		// (writer.go: `if !isTemporary(err) && !isTransientNetworkError(err) {
		// break }`). That is the "Kafka write errors (1/1)" fingerprint: one
		// message, one attempt, one error, on a condition usually over in seconds.
		{"live composite: WriteErrors{ErrNoLeader}", liveNoLeader, produceOutcomeNoLeader},
		{"wrapped by our own producer", wrappedNoLeader, produceOutcomeNoLeader},
		{"bare protocol.ErrNoLeader", protocol.ErrNoLeader, produceOutcomeNoLeader},
		// The broker-side code-5 error is a DIFFERENT type with a DIFFERENT text
		// ("Leader Not Available: ..."), which is why the typed check and the
		// substring check both earn their place rather than one being redundant.
		{"broker-side LeaderNotAvailable (code 5)", kafkago.LeaderNotAvailable, produceOutcomeNoLeader},

		{"too large", kafkago.MessageSizeTooLarge, produceOutcomeTooLarge},
		{"context cancelled", context.Canceled, produceOutcomeCanceled},
		{"deadline exceeded", context.DeadlineExceeded, produceOutcomeTimeout},
		{"connection refused", syscall.ECONNREFUSED, produceOutcomeNetwork},
		{"connection reset", syscall.ECONNRESET, produceOutcomeNetwork},
		{"broken pipe", syscall.EPIPE, produceOutcomeNetwork},
		{"unexpected EOF", io.ErrUnexpectedEOF, produceOutcomeNetwork},
		{"anything else", errors.New("something nobody has seen yet"), produceOutcomeOther},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyProduceErr(tc.err); got != tc.want {
				t.Errorf("classifyProduceErr(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

// A net.Error that reports Timeout() must classify as timeout, not network.
func TestClassifyProduceErrPrefersTimeoutOverNetwork(t *testing.T) {
	var timeoutErr net.Error = &net.OpError{Op: "write", Err: &timeoutStub{}}
	if got := classifyProduceErr(timeoutErr); got != produceOutcomeTimeout {
		t.Errorf("a timing-out net.Error classified as %q, want %q", got, produceOutcomeTimeout)
	}
}

type timeoutStub struct{}

func (timeoutStub) Error() string   { return "i/o timeout" }
func (timeoutStub) Timeout() bool   { return true }
func (timeoutStub) Temporary() bool { return true }

// THE CARDINALITY GUARD. Per-job reply topics are minted one per job and this
// cluster has carried ~25,000 topics at once; a raw topic label on a counter that
// fires on every produce in the fleet is an unbounded series explosion in
// Prometheus — a worse outage than the one the metric exists to measure.
//
// Mutation that must fail this: return the topic unchanged (or add any arm that
// returns a substring of its input rather than a fixed string).
func TestTopicClassCollapsesPerJobTopics(t *testing.T) {
	cases := []struct {
		topic string
		want  string
	}{
		// Real shapes, from live awaited_requests rows.
		{"job.12b85f92-012885b6-page-build-handler-process_item_iter_2_spawn_handler.responses", "job"},
		{"job.7abe1a57-e3db-4b71.requests", "job"},
		{"system.agent.build-dispatch-loop.requests", "system.agent"},
		{"system.agent.council-gate.requests", "system.agent"},
		// The non-per-agent system set is small and fixed, so it stays legible.
		{"system.generic.responses", "system.generic.responses"},
		{"some-other-topic", "other"},
		{"", "unknown"},
	}
	for _, tc := range cases {
		if got := topicClass(tc.topic); got != tc.want {
			t.Errorf("topicClass(%q) = %q, want %q", tc.topic, got, tc.want)
		}
	}
}

// The bound stated as an assertion rather than as a comment: many distinct job
// topics must yield exactly ONE series label.
func TestTopicClassIsBoundedAcrossManyJobTopics(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 5000; i++ {
		seen[topicClass(fmt.Sprintf("job.corr-%d-orch-%d-step_%d.responses", i, i, i))] = struct{}{}
	}
	if len(seen) != 1 {
		t.Fatalf("5,000 distinct per-job topics produced %d label values, want 1 - this is a Prometheus cardinality bomb, not a metric", len(seen))
	}
}
