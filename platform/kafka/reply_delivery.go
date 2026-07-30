// FILE: platform/kafka/reply_delivery.go
package kafka

import (
	"context"
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Reply delivery — the one place that knows what to do when the broker refuses
// a reply.
//
// THE RULE (016b §9, derived from bugs_closed/062): a response that cannot be
// delivered must become a deliverable error, never silence. The caller is
// listening on the reply topic, not reading this pod's logs, so a produce
// failure that is only logged starves it — in 062's case through 4 × 180s of
// retries on a failure that could never have succeeded.
//
// WHY THIS IS SHARED AND NOT COPIED (bugs_open/133): before this file, the rule
// held at exactly ONE of the nine reply-producing sites in the tree, because it
// existed only as an unexported block inside the webscrape batch handler. The
// obvious way to fix the eight others is to copy that block, which is how one
// rule becomes eight slightly different rules — the defect bugs_closed/144 was
// about. So the POLICY lives here with one implementation, and the two things
// that genuinely differ per caller stay with the caller: how to shrink an
// oversized payload, and what its error envelope looks like.
//
// This file is additive and inert: nothing changes for any producer until a
// call site opts in by calling DeliverReply.

// IsMessageTooLarge reports whether a produce error is the broker's
// message-size refusal — the one produce failure that is deterministic
// (resending the same bytes can never succeed), so it earns a degrade-and-
// retry rather than being surfaced as-is.
//
// Typed checks first (council fe468218, editquality seat): the producer wraps
// with %w so errors.Is/As unwrap to kafka-go's MessageSizeTooLarge (broker
// error code 10) or MessageTooLargeError (the writer's client-side pre-send
// detection). The substring fallback stays because WriteMessages can also
// surface the failure inside composite shapes (kafka.WriteErrors) that the
// unwrap chain does not always reach.
func IsMessageTooLarge(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, kafka.MessageSizeTooLarge) {
		return true
	}
	var tooLarge kafka.MessageTooLargeError
	if errors.As(err, &tooLarge) {
		return true
	}
	return strings.Contains(err.Error(), "Message Size Too Large")
}

// DeliveryOutcome says what happened to a reply, so the caller can tell the
// three cases apart without re-inspecting the error. They are deliberately
// distinct: Delivered and DeliveredDegraded both mean the caller was answered,
// but only the second means it was answered with less than was scraped.
type DeliveryOutcome int

const (
	// Delivered — the reply went out as built.
	Delivered DeliveryOutcome = iota
	// DeliveredDegraded — the broker refused the full reply; the degraded one
	// was accepted. The caller has been answered, with less in it.
	DeliveredDegraded
	// FailedTransient — produce failed for a reason that resending may fix
	// (broker unreachable, validation). The caller has NOT been answered;
	// the existing retry path is the right answer and DeliverReply has
	// deliberately not invented a new one.
	FailedTransient
	// FailedUndeliverable — the broker refuses even the degraded reply, or
	// there was nothing left to degrade. The caller has NOT been answered and
	// no retry can change that, so the caller MUST now send an error response.
	// This is the outcome that exists to stop the silence.
	FailedUndeliverable
)

func (o DeliveryOutcome) String() string {
	switch o {
	case Delivered:
		return "delivered"
	case DeliveredDegraded:
		return "delivered_degraded"
	case FailedTransient:
		return "failed_transient"
	case FailedUndeliverable:
		return "failed_undeliverable"
	}
	return "unknown"
}

// Answered reports whether the caller waiting on the reply topic got a reply.
// Use it rather than comparing against two constants, so a fourth outcome
// added later cannot silently read as "answered".
func (o DeliveryOutcome) Answered() bool {
	return o == Delivered || o == DeliveredDegraded
}

// ReplyDegrader produces a smaller version of a reply whose full form the
// broker refused. It returns the bytes to resend, or an error / nil bytes if
// there is nothing smaller worth sending — either of which means the caller
// gets FailedUndeliverable and must send an error response.
//
// A degrader must actually drop payload. Returning the same bytes wastes a
// second guaranteed-to-fail produce.
type ReplyDegrader func() ([]byte, error)

// DeliverReply produces a reply and, if the broker refuses it as too large,
// degrades it once and resends. It never invents a retry for transient
// failures and it never returns "delivered" for a message that was not.
//
// The caller MUST answer FailedUndeliverable with an error response — that is
// the whole point of the type. Callers that ignore it reintroduce the silent
// starve, which is why the outcome is returned rather than logged and dropped.
//
// degrade may be nil, meaning this reply has nothing to shrink; an oversize
// refusal then goes straight to FailedUndeliverable.
func DeliverReply(
	ctx context.Context,
	producer Producer,
	logger *zap.Logger,
	topic string,
	headers map[string]string,
	key, value []byte,
	degrade ReplyDegrader,
) (DeliveryOutcome, error) {
	if producer == nil {
		return FailedTransient, errors.New("reply delivery: nil producer")
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	err := producer.ProduceWithValidation(ctx, topic, headers, key, value)
	if err == nil {
		return Delivered, nil
	}

	if !IsMessageTooLarge(err) {
		// Transient: broker unreachable, validation refusal, context cancelled.
		// Resending the same bytes may well work, so the caller's existing
		// retry path stays in charge and we do not degrade a reply that was
		// never too big.
		logger.Error("Failed to produce reply",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("correlation_id", headers["correlation_id"]))
		return FailedTransient, err
	}

	if degrade == nil {
		logger.Error("Reply exceeds broker max message size and has no degraded form — caller must send an error response",
			zap.Error(err),
			zap.String("topic", topic),
			zap.Int("bytes", len(value)))
		return FailedUndeliverable, err
	}

	degradedValue, derr := degrade()
	if derr != nil || len(degradedValue) == 0 {
		logger.Error("Reply exceeds broker max message size and could not be degraded — caller must send an error response",
			zap.Error(err),
			zap.NamedError("degrade_error", derr),
			zap.String("topic", topic),
			zap.Int("bytes", len(value)))
		return FailedUndeliverable, err
	}

	logger.Warn("Reply exceeded broker max message size — resending degraded",
		zap.String("topic", topic),
		zap.String("correlation_id", headers["correlation_id"]),
		zap.Int("original_bytes", len(value)),
		zap.Int("degraded_bytes", len(degradedValue)))

	if perr := producer.ProduceWithValidation(ctx, topic, headers, key, degradedValue); perr != nil {
		logger.Error("Degraded reply also refused — caller must send an error response",
			zap.Error(perr),
			zap.String("topic", topic),
			zap.Int("degraded_bytes", len(degradedValue)))
		return FailedUndeliverable, perr
	}

	return DeliveredDegraded, nil
}
