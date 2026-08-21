// FILE: platform/kafka/dialer_test.go
package kafka

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	_ "github.com/gqls/agentchassis/platform/observability"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

// dialCount reads ai_persona_kafka_dial_total{broker,outcome} back out of
// prometheus.DefaultGatherer.
//
// Deliberately gathered from the DEFAULT registry rather than read off the
// collector directly: that is the registry promhttp.Handler() serves, so this
// also proves the metric is actually registered and would appear on /metrics.
// Given that the live fleet turned out to expose no ai_persona_* series at all,
// "the counter object incremented" is not the assertion worth making.
func dialCount(t *testing.T, broker, outcome string) float64 {
	t.Helper()
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != "ai_persona_kafka_dial_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			var gotBroker, gotOutcome string
			for _, lp := range m.GetLabel() {
				switch lp.GetName() {
				case "broker":
					gotBroker = lp.GetValue()
				case "outcome":
					gotOutcome = lp.GetValue()
				}
			}
			if gotBroker == broker && gotOutcome == outcome {
				return m.GetCounter().GetValue()
			}
		}
	}
	return 0
}

func TestClassifyDialErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil is ok", nil, dialOutcomeOK},
		{"connection refused", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, dialOutcomeRefused},
		{"context deadline", context.DeadlineExceeded, dialOutcomeTimeout},
		{"unclassified", errors.New("something else"), dialOutcomeError},
		{
			// The signature 040 is actually about: dial tcp <ip>:9092: i/o timeout.
			"tcp i/o timeout",
			&net.OpError{Op: "dial", Net: "tcp", Err: os.ErrDeadlineExceeded},
			dialOutcomeTimeout,
		},
		{
			// Must NOT be reported as a plain timeout — telling a resolution
			// stall apart from a connect stall is the whole point of the label.
			"dns timeout",
			&net.DNSError{Err: "i/o timeout", Name: "broker", IsTimeout: true},
			dialOutcomeDNSTimeout,
		},
		{
			"dns failure",
			&net.DNSError{Err: "no such host", Name: "broker", IsNotFound: true},
			dialOutcomeDNS,
		},
		{
			// Wrapped, because kafka-go wraps every dial error before we see it
			// ("failed to open connection to %s: %w").
			"wrapped tcp timeout",
			errors.New("failed to open connection to b:9092: " +
				(&net.OpError{Op: "dial", Err: os.ErrDeadlineExceeded}).Error()),
			dialOutcomeError, // a string-formatted error is NOT unwrappable
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyDialErr(tc.err); got != tc.want {
				t.Errorf("classifyDialErr(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

// A genuinely wrapped error must still classify, since that is how every real
// error arrives from kafka-go.
func TestClassifyDialErrUnwraps(t *testing.T) {
	inner := &net.OpError{Op: "dial", Net: "tcp", Err: os.ErrDeadlineExceeded}
	wrapped := &net.OpError{Op: "read", Err: inner}
	if got := classifyDialErr(wrapped); got != dialOutcomeTimeout {
		t.Errorf("wrapped timeout classified as %q, want %q", got, dialOutcomeTimeout)
	}
}

func TestBrokerLabelStripsPort(t *testing.T) {
	cases := map[string]string{
		"personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc:9092": "personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc",
		"10.20.161.217:9092": "10.20.161.217",
		"no-port-here":       "no-port-here",
	}
	for in, want := range cases {
		if got := brokerLabel(in); got != want {
			t.Errorf("brokerLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

// The positive control. A metric that reads zero because it was never wired is
// indistinguishable from a metric that reads zero because the bug is fixed —
// and that is not hypothetical here: verified 2026-07-26, the live Prometheus
// held ZERO ai_persona_* series because nothing ever served /metrics. This test
// asserts the counter actually moves when a dial fails.
func TestInstrumentedDialRecordsTimeout(t *testing.T) {
	const outcome = dialOutcomeTimeout

	// 203.0.113.0/24 is TEST-NET-3 (RFC 5737) — reserved for documentation and
	// guaranteed not routable, so this cannot accidentally reach anything.
	const addr = "203.0.113.1:9092"

	before := dialCount(t, "203.0.113.1", outcome)

	dial := instrumentedDialFunc(&net.Dialer{Timeout: 50 * time.Millisecond, DualStack: true})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := dial(ctx, "tcp", addr)
	if err == nil {
		conn.Close()
		t.Skip("unroutable address unexpectedly connected; cannot assert on timeout here")
	}
	if got := classifyDialErr(err); got != outcome {
		t.Skipf("dial failed as %q rather than %q (%v); environment-dependent", got, outcome, err)
	}

	after := dialCount(t, "203.0.113.1", outcome)
	if after <= before {
		t.Errorf("KafkaDialTotal{outcome=%q} did not increment: %v -> %v", outcome, before, after)
	}
}

// A successful dial must be counted too, otherwise the timeout count has no
// denominator and "5 timeouts" cannot be read as good or bad.
func TestInstrumentedDialRecordsSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	host, _, _ := net.SplitHostPort(ln.Addr().String())
	before := dialCount(t, host, dialOutcomeOK)

	dial := instrumentedDialFunc(&net.Dialer{Timeout: 2 * time.Second, DualStack: true})
	conn, err := dial(context.Background(), "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial to local listener failed: %v", err)
	}
	conn.Close()

	after := dialCount(t, host, dialOutcomeOK)
	if after != before+1 {
		t.Errorf("KafkaDialTotal{outcome=ok} = %v, want %v", after, before+1)
	}
}

// The council vetoed round 1 for changing shared-messaging defaults inside what
// was framed as an instrumentation change. This pins the promise that replaced
// it: ProducerTransport must reproduce kafka-go's DefaultTransport EXACTLY, and
// differ only by having an instrumented Dial. If someone retunes these, they are
// writing part 2 and this test should stop them doing it by accident.
func TestProducerTransportMatchesKafkaGoDefaults(t *testing.T) {
	tr := ProducerTransport()
	// Values from kafka-go v0.4.47 transport.go:191-210.
	if tr.DialTimeout != 5*time.Second {
		t.Errorf("DialTimeout = %v, want kafka-go default 5s", tr.DialTimeout)
	}
	if tr.IdleTimeout != 30*time.Second {
		t.Errorf("IdleTimeout = %v, want kafka-go default 30s", tr.IdleTimeout)
	}
	if tr.MetadataTTL != 6*time.Second {
		t.Errorf("MetadataTTL = %v, want kafka-go default 6s", tr.MetadataTTL)
	}
	if tr.Dial == nil {
		t.Error("Dial is nil — producer dials would not be counted")
	}
	// Shared, as a nil Transport was: all Writers in a process used kafka-go's
	// package-level DefaultTransport, so they shared one connection pool.
	if ProducerTransport() != tr {
		t.Error("ProducerTransport returned distinct instances; the pool would be split per Writer")
	}
}

// Each call site must keep the budget it had before this change.
func TestInstrumentedDialerPreservesCallerTimeout(t *testing.T) {
	for _, want := range []time.Duration{3 * time.Second, defaultDialerTimeout} {
		d := InstrumentedDialer(want)
		if d.Timeout != want {
			t.Errorf("InstrumentedDialer(%v).Timeout = %v", want, d.Timeout)
		}
		if d.DialFunc == nil {
			t.Errorf("InstrumentedDialer(%v).DialFunc is nil — dials would not be counted", want)
		}
	}
	if defaultDialerTimeout != 10*time.Second {
		t.Errorf("defaultDialerTimeout = %v, want kafka-go DefaultDialer's 10s", defaultDialerTimeout)
	}
}

// bugs_open/040-kafka-dial: an address with no host is refused before it is
// dialled, and counted under its own label.
//
// The mechanism this makes visible: net.SplitHostPort(":9092") SUCCEEDS with
// host="" — it is a valid address naming the local machine — so such a dial goes
// to the pod's own loopback and returns ECONNREFUSED instantly, hundreds per pod
// per burst. [MEASURED 2026-08-21] 94,419 `refused` in 7 days, 85,887 of them
// carrying an EMPTY broker label, which is brokerLabel's output for host="".
//
// The assertion is on the ERROR TEXT deliberately, not merely on "an error came
// back": on a machine where something happens to listen on localhost:9092 the
// dial would SUCCEED without the guard, and only the guard produces this string.
// That makes the test discriminate rather than pass by accident.
func TestInstrumentedDialRefusesAnEmptyHostBeforeDialling(t *testing.T) {
	const addr = ":9092"

	before := dialCount(t, "", dialOutcomeEmptyHost)
	refusedBefore := dialCount(t, "", dialOutcomeRefused)

	dial := instrumentedDialFunc(&net.Dialer{Timeout: 2 * time.Second, DualStack: true})
	conn, err := dial(context.Background(), "tcp", addr)
	if conn != nil {
		conn.Close()
		t.Fatal("a dial to an empty host returned a connection - the guard did not fire")
	}
	if err == nil {
		t.Fatal("a dial to an empty host returned no error")
	}
	if !strings.Contains(err.Error(), "empty host") {
		t.Fatalf("error %q does not name the guard; something else refused this dial and the test proves nothing", err)
	}

	if after := dialCount(t, "", dialOutcomeEmptyHost); after != before+1 {
		t.Errorf("empty_host count = %v, want %v", after, before+1)
	}
	// The label must be its OWN, not folded into refused: separating them is the
	// entire diagnostic value here (see the comment in instrumentedDialFunc).
	if after := dialCount(t, "", dialOutcomeRefused); after != refusedBefore {
		t.Errorf("refused count moved from %v to %v - an empty-host dial must not be counted as a refusal", refusedBefore, after)
	}
}

// CONTROL 1: an address SplitHostPort cannot parse must take the ordinary dial
// path, not the guard. Without this, a guard written as "anything SplitHostPort
// dislikes" would pass the test above and silently swallow real addresses.
func TestInstrumentedDialGuardIgnoresUnparseableAddresses(t *testing.T) {
	before := dialCount(t, "", dialOutcomeEmptyHost)

	dial := instrumentedDialFunc(&net.Dialer{Timeout: 50 * time.Millisecond, DualStack: true})
	conn, err := dial(context.Background(), "tcp", "9092") // no colon: SplitHostPort errors
	if conn != nil {
		conn.Close()
	}
	if err != nil && strings.Contains(err.Error(), "empty host") {
		t.Fatal("the guard fired on an address with no host:port separator - it must key on an EMPTY host, not on a parse failure")
	}
	if after := dialCount(t, "", dialOutcomeEmptyHost); after != before {
		t.Errorf("empty_host count moved from %v to %v on an unparseable address", before, after)
	}
}

// CONTROL 2: a populated host is untouched. Reuses the local listener pattern of
// TestInstrumentedDialRecordsSuccess so a real connection is made.
func TestInstrumentedDialGuardIgnoresAPopulatedHost(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot listen locally: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	before := dialCount(t, "", dialOutcomeEmptyHost)

	dial := instrumentedDialFunc(&net.Dialer{Timeout: 2 * time.Second, DualStack: true})
	conn, err := dial(context.Background(), "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("a populated host was refused: %v", err)
	}
	conn.Close()

	if after := dialCount(t, "", dialOutcomeEmptyHost); after != before {
		t.Errorf("empty_host count moved from %v to %v on a populated host", before, after)
	}
}
