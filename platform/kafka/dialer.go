// FILE: platform/kafka/dialer.go
package kafka

import (
	"context"
	"errors"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/platform/observability"
	"github.com/segmentio/kafka-go"
)

// bugs_open/040-kafka-dial. Before this file the fleet had four independent and
// mutually inconsistent dial configurations: the consumer dialled with a 10s
// timeout, the producer with 3s (kafka-go's DefaultTransport, because Transport
// was left nil), the topic manager with 10s (bare kafka.Dial -> DefaultDialer)
// and the health probe with 3s. None was configurable and none was counted, so
// the one bug that spans all of them could not be measured.
//
// Everything that opens a connection to a broker now goes through here.

// defaultDialTimeout replaces kafka-go's 10s. 040 §4.6: a dial that cannot
// complete in 10s on an in-cluster network is pathological regardless of cause,
// and a long timeout converts a lost SYN into a stalled orchestration. Failing
// sooner and retrying is strictly better than hanging.
const defaultDialTimeout = 5 * time.Second

// Dial outcomes. dns and dns_timeout are kept apart from timeout deliberately:
// 040's residual flake is unexplained, and whether the stall is in resolution or
// in the TCP connect is the question that decides where to look next.
const (
	dialOutcomeOK         = "ok"
	dialOutcomeTimeout    = "timeout"
	dialOutcomeRefused    = "refused"
	dialOutcomeDNS        = "dns"
	dialOutcomeDNSTimeout = "dns_timeout"
	dialOutcomeError      = "error"
)

// DialTimeout is the broker dial budget. KAFKA_DIAL_TIMEOUT is read as a whole
// number of SECONDS, not a Go duration string — envDurationOrDefault (consumer.go)
// is shared with KAFKA_SESSION_TIMEOUT and friends, and a value like "5s" parses
// as malformed and silently falls back to the default.
func DialTimeout() time.Duration {
	return envDurationOrDefault("KAFKA_DIAL_TIMEOUT", defaultDialTimeout)
}

// classifyDialErr maps a dial error onto a bounded metric label.
func classifyDialErr(err error) string {
	if err == nil {
		return dialOutcomeOK
	}

	// Checked before the generic timeout test on purpose: a resolution failure
	// that happens to be a timeout is a DNS problem, and labelling it "timeout"
	// would hide exactly the distinction this metric exists to draw.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsTimeout {
			return dialOutcomeDNSTimeout
		}
		return dialOutcomeDNS
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return dialOutcomeRefused
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return dialOutcomeTimeout
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return dialOutcomeTimeout
	}

	return dialOutcomeError
}

// brokerLabel strips the port so the metric's cardinality stays at the number of
// brokers (three) plus the bootstrap service, rather than growing per connection.
func brokerLabel(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}
	return host
}

// instrumentedDialFunc records the outcome and latency of every dial.
//
// Note this measures DNS resolution *and* the TCP connect together. kafka-go's
// own lookupHost is a no-op when Resolver is nil (dialer.go), so the name is
// still unresolved when it reaches us and Go's net.Dialer does the lookup inside
// this call. That is what we want: the 10s budget 040 observed being consumed
// was the two combined, and splitting them is what the outcome label is for.
func instrumentedDialFunc(base *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		start := time.Now()
		conn, err := base.DialContext(ctx, network, address)
		outcome := classifyDialErr(err)

		observability.KafkaDialTotal.WithLabelValues(brokerLabel(address), outcome).Inc()
		observability.KafkaDialDuration.WithLabelValues(outcome).Observe(time.Since(start).Seconds())

		return conn, err
	}
}

func newNetDialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout:   timeout,
		DualStack: true,
		KeepAlive: 30 * time.Second,
	}
}

var (
	sharedDialerOnce    sync.Once
	sharedDialer        *kafka.Dialer
	sharedTransportOnce sync.Once
	sharedTransport     *kafka.Transport
)

// SharedDialer is the fleet's single consumer/admin dialer.
func SharedDialer() *kafka.Dialer {
	sharedDialerOnce.Do(func() {
		sharedDialer = SharedDialerWithTimeout(DialTimeout())
	})
	return sharedDialer
}

// SharedDialerWithTimeout builds an instrumented dialer with a caller-chosen
// budget. Used by the health probe, which deliberately keeps a shorter one than
// the data path so a wedged broker is noticed rather than waited on.
func SharedDialerWithTimeout(timeout time.Duration) *kafka.Dialer {
	return &kafka.Dialer{
		Timeout: timeout,
		// kafka-go ignores DualStack, KeepAlive, LocalAddr and FallbackDelay once
		// DialFunc is set (see the DialFunc doc comment in dialer.go), so the
		// inner net.Dialer below is where they actually take effect.
		DialFunc: instrumentedDialFunc(newNetDialer(timeout)),
	}
}

// SharedTransport is the producer-side equivalent, and is a singleton because a
// kafka.Transport owns the per-broker connection pool — one per process means
// producers share connections instead of each keeping its own set.
func SharedTransport() *kafka.Transport {
	sharedTransportOnce.Do(func() {
		timeout := DialTimeout()
		sharedTransport = &kafka.Transport{
			Dial:        instrumentedDialFunc(newNetDialer(timeout)),
			DialTimeout: timeout,
			// kafka-go defaults these to 30s and 6s. Both are aggressive enough
			// that a low-traffic agent re-dials almost every time it produces,
			// which is pure dial volume on a path 040 shows is not always
			// reliable. The Java client's equivalent metadata age is 300s.
			IdleTimeout: envDurationOrDefault("KAFKA_IDLE_TIMEOUT", 5*time.Minute),
			MetadataTTL: envDurationOrDefault("KAFKA_METADATA_TTL", 30*time.Second),
		}
	})
	return sharedTransport
}
