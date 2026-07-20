// FILE: platform/health/kafka_reachability.go
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaReachability tracks whether ANY broker in a list has been reachable
// recently, via a cheap periodic TCP dial, and serves that state as liveness
// and readiness handlers. It exists because the fleet's probes were pointing
// at hardcoded-200 endpoints (bugs_open/003 §4.1): a pod that could not reach
// any broker reported healthy forever and was never rescheduled. One reachable
// broker is enough — a client can bootstrap the cluster from any live broker.
//
// This deliberately measures a DIFFERENT signal from Agent.checkHealth()
// (platform/agentbase/agent.go): that one pings the DB and logs message
// inactivity, which its own comment notes "might be normal if no messages are
// coming in" — an idle-but-healthy agent is indistinguishable there from a
// wedged one. Broker reachability is not ambiguous that way, which is why
// liveness keys on it and not on activity.
//
// Liveness policy (HandleHealth): fail only after UnhealthyAfter of
// CONTINUOUS unreachability — any successful dial resets the window, so
// intermittent flakes (bugs_open/040) cannot trip it; only the sustained
// wedge state does. A restart can help there: the replacement may land on a
// node with a working broker path. A database outage deliberately does not
// fail liveness — a restart cannot fix a down DB, and a fleet restart storm
// during a DB blip would compound it.
type KafkaReachability struct {
	brokers []string
	logger  *zap.Logger
	dialer  *kafka.Dialer

	started atomic.Bool
	lastOK  atomic.Int64 // unix seconds of the last successful broker dial

	// UnhealthyAfter is how long ALL brokers must be continuously
	// unreachable before HandleHealth reports failure.
	// Env override: KAFKA_UNHEALTHY_AFTER_SECONDS (default 300).
	UnhealthyAfter time.Duration
	// ProbeInterval is the dial cadence.
	// Env override: KAFKA_HEALTH_PROBE_INTERVAL_SECONDS (default 30).
	ProbeInterval time.Duration
	// DialTimeout bounds each TCP dial attempt.
	DialTimeout time.Duration
}

// NewKafkaReachability seeds lastOK to now, so a pod that never reaches a
// broker still gets the full UnhealthyAfter window before liveness fails —
// no crash-loop at boot.
//
// Pass the broker list from kafkaplatform.GetBrokers() where one is available:
// the callers' Kafka clients resolve brokers from the environment
// (SERVICE_INFRASTRUCTURE_KAFKA_BROKERS, else KAFKA_BROKERS), and a health
// check that probes a DIFFERENT list reports on a Kafka the process is not
// using. That divergence was real and was caught in test on 2026-07-20 — the
// config-file value and the env value disagreed, and the endpoint cheerfully
// reported "ok" for a pod that could not reach its brokers at all.
func NewKafkaReachability(brokers []string, logger *zap.Logger) *KafkaReachability {
	k := &KafkaReachability{
		brokers:        brokers,
		logger:         logger,
		dialer:         &kafka.Dialer{Timeout: 3 * time.Second, DualStack: true},
		UnhealthyAfter: envSecondsOrDefault("KAFKA_UNHEALTHY_AFTER_SECONDS", 300),
		ProbeInterval:  envSecondsOrDefault("KAFKA_HEALTH_PROBE_INTERVAL_SECONDS", 30),
		DialTimeout:    3 * time.Second,
	}
	k.lastOK.Store(time.Now().Unix())
	return k
}

// Run probes the broker list every ProbeInterval until ctx is cancelled.
func (k *KafkaReachability) Run(ctx context.Context) {
	ticker := time.NewTicker(k.ProbeInterval)
	defer ticker.Stop()
	for {
		k.dialAny()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// dialAny succeeds if any broker completes a Kafka METADATA exchange, not
// merely a TCP handshake. A bare TCP dial is not evidence: proven 2026-07-20
// on a deliberately wedged pod, where `wget 10.255.255.1:9092` timed out and
// every real client logged `i/o timeout`, yet the endpoint reported "ok" —
// something in the pod network completes a TCP connect to addresses that
// serve no Kafka. Brokers() forces a real protocol round-trip.
func (k *KafkaReachability) dialAny() {
	for _, broker := range k.brokers {
		ctx, cancel := context.WithTimeout(context.Background(), k.DialTimeout)
		conn, err := k.dialer.DialContext(ctx, "tcp", broker)
		cancel()
		if err != nil {
			continue
		}
		_, metaErr := conn.Brokers()
		conn.Close()
		if metaErr == nil {
			k.lastOK.Store(time.Now().Unix())
			return
		}
	}
	k.logger.Warn("health: no Kafka broker reachable",
		zap.Strings("brokers", k.brokers),
		zap.Duration("kafka_unreachable_for", k.Silence()))
}

// SetStarted marks the owning process's main work loop as launched; readiness
// reports not_ready until this is called.
func (k *KafkaReachability) SetStarted() {
	k.started.Store(true)
}

// Silence is how long ago the last successful broker dial was.
func (k *KafkaReachability) Silence() time.Duration {
	return time.Since(time.Unix(k.lastOK.Load(), 0))
}

// Checker adapts the reachability signal to the Checkers map used by
// health.Server, for binaries that already run one.
func (k *KafkaReachability) Checker() CheckFunc {
	return func(_ context.Context) error {
		if silence := k.Silence(); silence > k.UnhealthyAfter {
			return fmt.Errorf("no Kafka broker reachable for %s", silence.Round(time.Second))
		}
		return nil
	}
}

// HandleHealth is a liveness handler: 503 only after UnhealthyAfter of
// continuous all-broker unreachability.
func (k *KafkaReachability) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	silence := k.Silence()
	if silence > k.UnhealthyAfter {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":                    "unhealthy",
			"reason":                    "no Kafka broker reachable",
			"kafka_unreachable_seconds": int(silence.Seconds()),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":                    "ok",
		"kafka_last_ok_seconds_ago": int(silence.Seconds()),
	})
}

// HandleReady is a readiness handler: ready once SetStarted has been called
// and a broker was reachable within roughly two probe cycles.
func (k *KafkaReachability) HandleReady(w http.ResponseWriter, _ *http.Request) {
	silence := k.Silence()
	recentWindow := 2*k.ProbeInterval + k.DialTimeout
	if !k.started.Load() || silence > recentWindow {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":                    "not_ready",
			"process_started":           k.started.Load(),
			"kafka_last_ok_seconds_ago": int(silence.Seconds()),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}

func writeJSON(w http.ResponseWriter, code int, body map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}

func envSecondsOrDefault(key string, defaultSeconds int) time.Duration {
	if val := os.Getenv(key); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}
