// FILE: platform/kafka/produce_retry.go
//
// bugs_open/040-kafka-dial: a bounded, OPT-IN retry for produces that are worth
// retrying, adopted at the reply/terminal lane and nowhere else.
//
// WHY THIS EXISTS AND WHY IT IS NOT INSIDE Produce. [MEASURED 2026-08-21] over
// the retained agent_error_log window: 63 "Kafka write errors" plus 40 "topic
// partition has no leader" rows across 93 distinct orchestrations, recurring most
// days since 08-10. The single worst shape is a workflow whose every substantive
// step has already committed failing at its terminal complete_workflow send — the
// work is done and the head row says FAILED.
//
// kafka-go will not cover it. Its writer loops up to MaxAttempts (default 10)
// with backoff, but breaks after ONE attempt when
// `!isTemporary(err) && !isTransientNetworkError(err)`, and protocol.ErrNoLeader
// — the client-side "topic partition has no leader" — is a bare string type with
// no Temporary() method. So the canonical seconds-scale transient, a leadership
// election, is never retried internally. "Kafka write errors (1/1)" is the
// fingerprint: one message, one attempt.
//
// So this is not duplicating the library's machinery; it covers precisely what
// that machinery declines to cover. It is nonetheless BEHAVIOUR on shared
// messaging plumbing, and this case's round 1 was vetoed for shipping exactly
// that unmeasured and unannounced. Hence: opt-in, adopted only at named sites,
// with everything else byte-identical to today.
package kafka

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/gqls/agentchassis/platform/observability"
	"go.uber.org/zap"
)

// RetryPolicy bounds an application-level produce retry.
//
// Attempts counts TOTAL attempts, not extra ones: 1 means "no retry" and is the
// same behaviour as calling Produce directly.
type RetryPolicy struct {
	Attempts  int
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// DefaultReplyRetryPolicy is the policy the reply lane adopts.
//
// Worst case is bounded and worth stating plainly: 4 attempts × the writer's 10s
// WriteTimeout, plus ~3.5s of backoff, is about 44 seconds before a reply is
// finally reported undeliverable. That is a long time to hold a step — and it is
// the right trade against losing a completed workflow, which is what happens
// today. It is not appropriate for a hot dispatch path, which is one reason this
// is opt-in rather than a default inside Produce.
var DefaultReplyRetryPolicy = RetryPolicy{
	Attempts:  4,
	BaseDelay: 500 * time.Millisecond,
	MaxDelay:  4 * time.Second,
}

// isDeterministicProduceErr reports whether resending the same bytes is provably
// pointless, so the retry loop must stop immediately rather than burn its budget.
//
// Validation refusal: the headers are wrong by construction (bugs_open/274).
// Too large: the broker's size limit does not move between attempts — DeliverReply
// answers that one by DEGRADING, which is a different remedy and stays its own.
// Context cancelled: there is no one left to answer.
func isDeterministicProduceErr(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrMessageValidationFailed) ||
		IsMessageTooLarge(err) ||
		errors.Is(err, context.Canceled)
}

// retryDelay is exponential with full jitter, capped at MaxDelay.
//
// Jittered because these failures are CORRELATED: a leadership election fails
// every producer in the fleet at once, and an unjittered backoff would march them
// all back onto the broker in the same instant.
func (p RetryPolicy) retryDelay(attempt int) time.Duration {
	delay := p.BaseDelay << (attempt - 1)
	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	if delay <= 0 {
		return 0
	}
	// Jitter WITHIN [delay/2, delay], never above it.
	//
	// The obvious form — rand.Int63n(delay) + delay/2 — spans [delay/2, 1.5*delay)
	// and therefore BREAKS THE CAP by up to 50%, which a test caught here rather
	// than a reader. MaxDelay is a bound or it is decoration.
	half := int64(delay) / 2
	// #nosec G404 -- jitter, not a security decision.
	return time.Duration(half + rand.Int63n(int64(delay)-half+1))
}

// retrySend runs send under the policy and returns the LAST attempt's error
// VERBATIM.
//
// Returning the last error unwrapped is load-bearing, not tidiness: every caller
// wraps it in its own message ("failed to send response: %w",
// "failed to write message to kafka: %w"), every log search matches on those, and
// platform/errors' transient classifier matches on needles inside them. Adding a
// layer here would silently change what all of that sees.
func retrySend(ctx context.Context, logger *zap.Logger, policy RetryPolicy, topicLabel string, send func() error) error {
	if logger == nil {
		logger = zap.NewNop()
	}
	attempts := policy.Attempts
	if attempts < 1 {
		attempts = 1
	}

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		err = send()
		if err == nil {
			if attempt > 1 {
				logger.Warn("kafka produce succeeded after retry",
					zap.Int("attempt", attempt),
					zap.String("topic_class", topicLabel))
				observability.KafkaProduceRetryRecoveries.WithLabelValues(topicLabel).Inc()
			}
			return nil
		}
		if isDeterministicProduceErr(err) {
			// Resending the same bytes cannot succeed. Stop now and let the
			// caller apply the remedy that fits (an error response, or a
			// degrade), rather than spending the budget proving it again.
			return err
		}
		if attempt == attempts {
			break
		}
		delay := policy.retryDelay(attempt)
		logger.Warn("kafka produce failed, retrying",
			zap.Int("attempt", attempt),
			zap.Int("of", attempts),
			zap.Duration("backoff", delay),
			zap.String("topic_class", topicLabel),
			zap.Error(err))

		select {
		case <-ctx.Done():
			// The caller is gone; returning the produce error rather than the
			// context error keeps the caller's classification unchanged.
			return err
		case <-time.After(delay):
		}
	}
	return err
}

// ProduceWithRetry is Produce plus the bounded retry: identical semantics
// (no validation), identical error text, one policy.
//
// Callers that want validation as well go through DeliverReply with WithRetry.
func ProduceWithRetry(
	ctx context.Context,
	producer Producer,
	logger *zap.Logger,
	policy RetryPolicy,
	topic string,
	headers map[string]string,
	key, value []byte,
) error {
	if producer == nil {
		return errors.New("produce with retry: nil producer")
	}
	return retrySend(ctx, logger, policy, topicClass(topic), func() error {
		return producer.Produce(ctx, topic, headers, key, value)
	})
}

// DeliverOption configures DeliverReply. The zero set of options is today's
// behaviour exactly.
type DeliverOption func(*deliverConfig)

type deliverConfig struct {
	retry  bool
	policy RetryPolicy
}

// WithRetry opts one DeliverReply call into the bounded produce retry.
//
// OFF BY DEFAULT, deliberately and per the owner ruling of 2026-08-02 §2: this is
// new behaviour on a seam every reply in the fleet crosses, so the unsafe side is
// the default and each adopter says so at its own call site, where a reviewer of
// the CALLER can see it. The five callers that do not pass it behave exactly as
// before — pinned by a test.
//
// ⚠ What opting in accepts: a retry after a LOST ACK duplicates the reply.
// kafka-go v0.4.47 has no idempotent producer, so that cannot be prevented here.
// It is absorbed downstream by the parent's two-phase ClaimAwaitedRequest and by
// processed_messages — so only adopt this where the consumer is a parent
// orchestration with that dedupe, which is what every current adopter is.
func WithRetry(policy RetryPolicy) DeliverOption {
	return func(c *deliverConfig) {
		c.retry = true
		c.policy = policy
	}
}
