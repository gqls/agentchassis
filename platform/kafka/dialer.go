// FILE: platform/kafka/dialer.go
package kafka

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/platform/observability"
	"github.com/segmentio/kafka-go"
)

// bugs_open/040-kafka-dial. Every Kafka dial in the fleet is counted here.
//
// SCOPE — read before adding anything to this file.
//
// This is instrumentation ONLY. It deliberately changes no timeout, no pool
// setting, and no dial behaviour of any kind. Each call site keeps the exact
// budget it had before, which is why the constructors below take a timeout
// rather than choosing one.
//
// The first version of this change did more: it unified the fleet's four
// divergent dial configurations behind one shared dialer, dropped the default
// timeout from 10s to 5s, and raised the producer's IdleTimeout 30s→5m and
// MetadataTTL 6s→30s. The council gate vetoed that (guardian, HIGH, round 1,
// corr 7abe1a57) and was right to: those are behaviour changes to shared
// messaging plumbing and to failover reactivity across every pipeline, bundled
// into what was presented as a metrics change. Nothing measured them; they were
// argued from a bug-file remark and from the Java client's defaults.
//
// The consolidation is not abandoned, it is SEQUENCED. Once these counters have
// shipped, `ai_persona_kafka_dial_duration_seconds` says what the real dial
// latency distribution is, and a timeout can be chosen from that rather than
// from an opinion. That is a separate change and needs an architecture review.
//
// So: if you are about to add a default here, you are writing part 2. Don't.

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

	// empty_host is not a failure MODE, it is a failure SHAPE: the address had no
	// host at all, so no dial was ever worth attempting. It is kept out of
	// "refused" deliberately — see instrumentedDialFunc.
	dialOutcomeEmptyHost = "empty_host"
)

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
		// Refuse an address with no host before dialling it, and count it under
		// its own label.
		//
		// WHY THIS IS HERE AND NOT AT A CALL SITE. `net.SplitHostPort(":9092")`
		// SUCCEEDS with host="" — it is not a malformed address, it is a valid one
		// naming the local machine. So a dial to ":9092" goes to the pod's own
		// loopback, where nothing listens, and comes back ECONNREFUSED instantly.
		// A single pod can rack up hundreds in a burst, which is exactly the shape
		// bugs_open/040 measured: [MEASURED 2026-08-21] 94,419 `refused` over 7
		// days, of which 85,887 carry an EMPTY broker label — that label is
		// brokerLabel's output for host="", so it is not a missing label, it is a
		// dial to ":9092".
		//
		// topic_manager.go's getController was one producer of that address and is
		// fixed (controllerAddress rejects an empty Host, v1.0.1291). The bursts
		// RECURRED after that fix, so at least one more producer exists and it is
		// not in this repository: kafka-go's own consumer-group coordinator lookup
		// builds `net.JoinHostPort(out.Coordinator.Host, ...)` straight from a
		// FindCoordinator response with no validation (consumergroup.go), and every
		// Reader in the fleet dials through this dialer. Vendored code cannot be
		// guarded upstream, and this function is the one choke point BELOW all of
		// it.
		//
		// This changes no outcome — a dial to ":9092" fails either way. What it
		// buys is the discriminator: post-roll, a non-zero empty_host says the
		// remaining producer is library-internal, and `refused` with an empty
		// broker label must be STRUCTURALLY ZERO. If it is not, a third producer
		// exists outside the instrumented dial path, and that is the disconfirming
		// result to look for.
		if host, _, splitErr := net.SplitHostPort(address); splitErr == nil && host == "" {
			// broker label stays "" so the series is continuous with the 85,887
			// events already recorded under it.
			observability.KafkaDialTotal.WithLabelValues(brokerLabel(address), dialOutcomeEmptyHost).Inc()
			// No duration observation: nothing was dialled, and a ~0s sample would
			// drag the latency distribution this bug is read from.
			return nil, fmt.Errorf("kafka: refusing dial to structurally invalid address %q: empty host (bugs_open/040)", address)
		}

		start := time.Now()
		conn, err := base.DialContext(ctx, network, address)
		outcome := classifyDialErr(err)

		observability.KafkaDialTotal.WithLabelValues(brokerLabel(address), outcome).Inc()
		observability.KafkaDialDuration.WithLabelValues(outcome).Observe(time.Since(start).Seconds())

		return conn, err
	}
}

// InstrumentedDialer returns a kafka.Dialer that counts its dials and is
// otherwise identical to what the caller had before.
//
// The caller passes its OWN existing timeout — there is no default here on
// purpose (see the scope note at the top of this file). DualStack matches both
// kafka-go's DefaultDialer and the previous inline dialers.
func InstrumentedDialer(timeout time.Duration) *kafka.Dialer {
	return &kafka.Dialer{
		Timeout: timeout,
		// kafka-go ignores DualStack, KeepAlive, LocalAddr and FallbackDelay
		// once DialFunc is set (see the DialFunc doc comment in its dialer.go),
		// so the inner net.Dialer below is where they actually take effect.
		DialFunc: instrumentedDialFunc(&net.Dialer{
			Timeout:   timeout,
			DualStack: true,
		}),
	}
}

// defaultDialerTimeout is kafka-go's package-level DefaultDialer timeout
// (dialer.go: `Timeout: 10 * time.Second, DualStack: true`). Sites that used the
// bare kafka.Dial/kafka.DialContext helpers were getting this, so they pass it
// explicitly now and their behaviour is unchanged.
const defaultDialerTimeout = 10 * time.Second

// producerTransport instruments the producer's dials while reproducing
// kafka-go's DefaultTransport EXACTLY in every other respect.
//
// Package-level, and shared by every Writer, because that is what the previous
// behaviour was: a nil Writer.Transport falls through to kafka-go's
// package-level DefaultTransport, so all Writers in a process already shared one
// connection pool. Giving each Writer its own Transport would split that pool —
// a behaviour change, and not the one this file is for.
//
// Every value below is copied from kafka-go v0.4.47 transport.go and must stay
// copied. Retuning them is part 2.
//
//	Dial:        (&net.Dialer{Timeout: 3s, DualStack: true}).DialContext  (:126-131)
//	DialTimeout: 5s   (:191-196)
//	IdleTimeout: 30s  (:198-203)
//	MetadataTTL: 6s   (:205-210)
var producerTransport = &kafka.Transport{
	Dial: instrumentedDialFunc(&net.Dialer{
		Timeout:   3 * time.Second,
		DualStack: true,
	}),
	DialTimeout: 5 * time.Second,
	IdleTimeout: 30 * time.Second,
	MetadataTTL: 6 * time.Second,
}

// ProducerTransport is the instrumented stand-in for kafka-go's DefaultTransport.
func ProducerTransport() *kafka.Transport { return producerTransport }
