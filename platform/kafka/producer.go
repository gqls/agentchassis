// FILE: platform/kafka/producer.go
package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/platform/observability"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
	"go.uber.org/zap"
)

// ErrMessageValidationFailed is returned by ProduceWithValidation when the
// injected validator rejects a non-error message's headers. It is DETERMINISTIC
// for the message: the same bytes can never pass on a retry, so callers (and
// DeliverReply's classification) must treat it as permanent, not transient
// (bugs_open/274). The text is unchanged from the previous bare error so every
// existing log search still matches.
var ErrMessageValidationFailed = errors.New("message validation failed")

// Producer defines the interface for Kafka message production
type Producer interface {
	Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
	Close() error
}

// MessageValidator interface for outgoing message validation.
// Implemented by validation.Validator - injected to avoid cyclic import.
type MessageValidator interface {
	ValidateOutgoingMessage(headers map[string]string) bool
}

// KafkaProducer wraps the kafka-go writer for standardized message production
type KafkaProducer struct {
	writer    *kafka.Writer
	logger    *zap.Logger
	validator MessageValidator // injected, can be nil
}

// NewProducer creates a new standardized Kafka producer without validation
func NewProducer(brokers []string, logger *zap.Logger) (Producer, error) {
	return NewProducerWithValidator(brokers, logger, nil)
}

// NewProducerWithValidator creates a new Kafka producer with an injected validator.
func NewProducerWithValidator(brokers []string, logger *zap.Logger, validator MessageValidator) (Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list cannot be empty")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		WriteTimeout: 10 * time.Second,
		// bugs_open/040-kafka-dial: nil Transport fell through to kafka-go's
		// DefaultTransport, so producer dials were never counted. This one is
		// byte-for-byte DefaultTransport's settings plus an instrumented Dial —
		// same 3s dial, same 30s IdleTimeout, same 6s MetadataTTL, same shared
		// pool. Retuning any of it is a separate, reviewed change.
		Transport: ProducerTransport(),
	}

	logger.Info("Kafka producer created", zap.Strings("brokers", brokers))

	return &KafkaProducer{
		writer:    writer,
		logger:    logger,
		validator: validator,
	}, nil
}

// SetValidator allows setting the validator after construction.
// Useful when the validator can't be provided at construction time.
func (p *KafkaProducer) SetValidator(v MessageValidator) {
	p.validator = v
}

// Produce outcomes. Mirrors classifyDialErr's bounded-label discipline on the
// other side of the connection: the dial counters see a CONNECTION fail, these
// see a WRITE fail, and until bugs_open/040's 2026-08-21 round nothing did.
const (
	produceOutcomeOK = "ok"
	// The two no-leader errors are kept APART because they behave OPPOSITELY
	// inside kafka-go, and that difference is the whole diagnostic value:
	//   client_no_leader — protocol.ErrNoLeader, a bare string type with no
	//     Temporary() method, so the writer's retry loop BREAKS AFTER ONE
	//     ATTEMPT. This is the "Kafka write errors (1/1)" fingerprint.
	//   broker_no_leader — kafka.LeaderNotAvailable (error code 5), which IS in
	//     kafka-go's temporary set and IS retried internally.
	// Collapsing them would leave a future reader unable to tell
	// "exhausted immediately" from "retried and still failed" — the council's
	// editquality seat raised exactly that (corr a414d81b, round 1).
	produceOutcomeClientNoLeader = "client_no_leader"
	produceOutcomeBrokerNoLeader = "broker_no_leader"
	produceOutcomeTooLarge       = "too_large"
	produceOutcomeTimeout        = "timeout"
	produceOutcomeCanceled       = "canceled"
	produceOutcomeNetwork        = "network"
	produceOutcomeOther          = "other"
)

// classifyProduceErr maps a WriteMessages error onto a bounded metric label.
//
// no_leader is checked first and is the one worth understanding, because it is
// the fleet's most common produce failure and kafka-go will NEVER retry it
// internally: the writer's retry loop breaks after one attempt when
// `!isTemporary(err) && !isTransientNetworkError(err)`, and protocol.ErrNoLeader
// is a bare string type with no Temporary() method. That is the observed
// "Kafka write errors (1/1)" fingerprint — one message, one attempt, one error —
// on a condition that is usually over in seconds (a leadership election).
//
// The substring arm is belt-and-braces alongside the typed check, for the same
// reason IsMessageTooLarge carries one: WriteMessages surfaces failures inside
// composite kafka.WriteErrors shapes whose Error() text embeds the member texts
// but whose unwrap chain does not always reach them.
func classifyProduceErr(err error) string {
	if err == nil {
		return produceOutcomeOK
	}

	// Broker-side first: it is the TYPED check, and its message text does not
	// contain "has no leader", so the two arms cannot shadow each other.
	if errors.Is(err, kafka.LeaderNotAvailable) {
		return produceOutcomeBrokerNoLeader
	}
	if errors.Is(err, protocol.ErrNoLeader) || strings.Contains(err.Error(), "has no leader") {
		// The substring arm is belt-and-braces beside the typed one: WriteMessages
		// surfaces failures inside composite kafka.WriteErrors shapes whose
		// Error() text embeds the member texts but whose unwrap chain does not
		// always reach them.
		return produceOutcomeClientNoLeader
	}
	if IsMessageTooLarge(err) {
		return produceOutcomeTooLarge
	}
	if errors.Is(err, context.Canceled) {
		return produceOutcomeCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return produceOutcomeTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return produceOutcomeTimeout
	}
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrUnexpectedEOF) {
		return produceOutcomeNetwork
	}

	return produceOutcomeOther
}

// topicClass collapses a topic name onto a BOUNDED metric label.
//
// ⚠ This is the cardinality guard, and it is not optional. Reply topics are
// minted per job — job.<correlation>-<orchestration>-<step>.responses — and the
// cluster has carried ~25,000 topics at once. A raw `topic` label on a counter
// that fires on every produce in the fleet would be an unbounded series
// explosion in Prometheus, which is a much worse outage than the one this metric
// exists to measure. Any new arm added here must be a FIXED string, never a
// substring of the input.
func topicClass(topic string) string {
	switch {
	case topic == "":
		return "unknown"
	case strings.HasPrefix(topic, "job."):
		return "job"
	case strings.HasPrefix(topic, "system.agent."):
		return "system.agent"
	case strings.HasPrefix(topic, "system."):
		// Collapse to the FAMILY, and only to a family this list already knows.
		//
		// ⚠ CORRECTED after the council's editquality seat objected (HIGH, corr
		// a414d81b): this arm used to `return topic`, which is precisely the
		// "returns a substring of its input" case the rule above forbids — the
		// author stated the rule and then broke it in the next arm. It is not
		// hypothetical: [MEASURED 2026-08-21 against the live cluster] of 937
		// `system.*` topics, 859 are caught by the system.agent arm above and
		// **78 would have reached here as distinct label values** — and the two
		// biggest residue families, `system.errors.<agent-type>` (18) and
		// `system.responses.<agent-type>` (17), grow with every new agent type.
		// Bounded in practice, unbounded in principle, which is not a guard.
		if family, ok := systemTopicFamily(topic); ok {
			return family
		}
		return "system.other"
	default:
		return "other"
	}
}

// knownSystemTopicFamilies is the CLOSED set of `system.<family>` prefixes
// topicClass may emit. Enumerated 2026-08-21 from the live cluster's 937
// `system.*` topics, not invented — a family not on this list collapses to
// "system.other" rather than minting a label.
//
// Adding a family here is the only way to grow this metric's cardinality, which
// is the point: it is a compile-time decision a reviewer can see, not a runtime
// consequence of somebody naming a topic.
var knownSystemTopicFamilies = map[string]struct{}{
	"errors":       {},
	"responses":    {},
	"adapter":      {},
	"events":       {},
	"orchestrator": {},
	"metrics":      {},
	"audit":        {},
	"human":        {},
	"dlq":          {},
	"dispatch":     {},
	"commands":     {},
	"thunder":      {},
	"compliance":   {},
}

// systemTopicFamily returns the fixed "system.<family>" label for a system topic,
// and false when the family is not one this build knows about.
func systemTopicFamily(topic string) (string, bool) {
	rest := strings.TrimPrefix(topic, "system.")
	family := rest
	if i := strings.IndexByte(rest, '.'); i >= 0 {
		family = rest[:i]
	}
	if _, known := knownSystemTopicFamilies[family]; !known {
		return "", false
	}
	return "system." + family, true
}

// Produce sends a message to a specific topic with standard headers
func (p *KafkaProducer) Produce(ctx context.Context, send_to_topic string, headers map[string]string, key, value []byte) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Topic:   send_to_topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
		Time:    time.Now().UTC(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	observability.KafkaProduceTotal.WithLabelValues(topicClass(send_to_topic), classifyProduceErr(err)).Inc()
	if err != nil {
		p.logger.Error("Failed to produce Kafka message",
			zap.String("send_to_topic", send_to_topic),
			zap.String("produce_outcome", classifyProduceErr(err)),
			zap.Error(err),
		)
		// Error text deliberately unchanged: every log search, and every needle in
		// platform/errors' transient classifier, matches on this wrapping.
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}
	current, caller := getFuncInfo(1)
	p.logger.Info("agent producer.go Successfully produced message",
		zap.String("sent to topic", send_to_topic),
		zap.String("key", string(key)),
		zap.String("value", string(value)),
		zap.Any("headers", headers),
		zap.String("current function", current),
		zap.String("caller", caller),
		zap.Time("time", time.Now().UTC()),
	)

	return nil
}

// ProduceWithValidation validates using the injected validator before sending
func (p *KafkaProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	// If no validator injected, just produce without validation
	if p.validator == nil {
		return p.Produce(ctx, topic, headers, key, value)
	}

	if !p.validator.ValidateOutgoingMessage(headers) {
		// Check if it's an error message - those we always send
		if headers["is_error"] == "true" {
			p.logger.Warn("Error message failed validation but sending anyway",
				zap.String("topic", topic),
				zap.String("correlation_id", headers["correlation_id"]))
			return p.Produce(ctx, topic, headers, key, value)
		}

		return ErrMessageValidationFailed
	}

	// Validation passed, send the message
	return p.Produce(ctx, topic, headers, key, value)
}

// Close gracefully closes the producer's writer
func (p *KafkaProducer) Close() error {
	p.logger.Info("Closing Kafka producer...")
	return p.writer.Close()
}

// Helper to get current and caller function names
func getFuncInfo(skip int) (current, caller string) {
	// skip=0 => this func, skip=1 => its caller, skip=2 => caller's caller, etc.
	if pc, _, _, ok := runtime.Caller(skip); ok {
		current = runtime.FuncForPC(pc).Name()
	}
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		caller = runtime.FuncForPC(pc).Name()
	}
	return
}
