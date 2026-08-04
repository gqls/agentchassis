// FILE: platform/messaging/validation_drop.go
//
// The "validation error" drop path, and its durable record (bugs_open/034).
//
// Two layers independently decide that a failed message is a validation error
// and must NOT be retried: MessageProcessor.handleError here, and
// agentbase.Agent.processMessage above it. Both decided it by substring match,
// against two lists that had silently drifted apart — this file makes that ONE
// list, so the classification cannot differ by layer again.
//
// The drop itself is deliberate and stays: retrying a genuinely malformed
// message is an infinite loop. What was wrong is that the drop left no trace a
// database query could find. The only residue was a zap.Warn on a pod that
// rotates within minutes and a Prometheus counter labelled (agent_type, reason)
// — enough to say something was dropped, never which thing. Every investigation
// here starts from the database, so from the investigator's point of view the
// message never existed. recordDroppedValidationError closes that.

package messaging

import (
	"context"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/orchestration"
	"go.uber.org/zap"
)

// ValidationErrorNeedles are the substrings that mark a processing failure as a
// non-retryable validation error. This is the single source of truth for both
// the messaging and agentbase layers.
//
// It is a substring match against the WHOLE error string, unanchored to any
// validation origin, and that is a known defect (bugs_open/034 fix candidate 2,
// not yet done): "invalid" is also a substring of "invalid character 'w' after
// object key:value pair" (a truncated-LLM parse failure), "invalid connection"
// (a database driver fault), "invalid memory address" (a recovered nil deref)
// and "x509: … invalid" (a TLS failure). Each of those is a real runtime error
// that this list silently reclassifies as "the caller sent us rubbish", and then
// drops without a retry.
//
// Replacing it with a typed sentinel (errors.Is(err, ErrValidation)) is the
// structural fix and is deliberately NOT attempted here — it needs every error
// construction site audited, and doing it blind would change retry behaviour
// fleet-wide. Until then the mitigation is visibility, not correctness: every
// drop records WHICH needle fired, so a misclassification is a queryable row
// rather than an invisible one.
//
// > **bugs_open/195 — that deferred structural fix is now DONE, and the reason
// > it could not wait is that this list missed its own commonest input.** The
// > fleet's most frequent permanent config error renders as
// > "WORKFLOW_INVALID: Invalid workflow configuration (caused by: …)", which
// > matches NONE of the four needles: "is required" loses to the wording
// > ("requires a topic"), and "invalid" loses to the CAPITAL I, because the
// > match is case-sensitive. So the error was classified transient, and
// > recordDroppedValidationError — the durable record 034 built — was never
// > reached, because the branch that calls it was never entered. See
// > MatchedPermanentFailure below: typed first, this list only as a fallback
// > for untyped errors. The needles are deliberately NOT case-folded, which
// > would widen every hazard listed above to its capitalised variants too.
//
// Ordering matters only for which needle gets reported when several match.
var ValidationErrorNeedles = []string{"is required", "validation", "invalid", "missing"}

// NonRetryablePermanentCodes are the typed DomainError codes that mark a
// processing failure PERMANENT: drop it, record it durably, never retry it.
//
// Closed and explicit, deliberately. The tempting shortcut — "classify on
// DomainError.Retryable" — is far too wide: errors.InternalError() sets
// Retryable false, so every genuinely transient internal fault would start
// being dropped fleet-wide. A code earns its place here only when retrying it
// can never succeed because the input is statically wrong.
var NonRetryablePermanentCodes = []errors.ErrorCode{
	errors.ErrWorkflowInvalid,
	errors.ErrValidation,
}

// MatchedPermanentFailure classifies err as a permanent failure and returns an
// audit token naming WHAT matched, or "" for "not permanent".
//
// The token is the point, and it is inherited from MatchedValidationNeedle
// below: returning WHICH thing matched rather than a bool is what makes an
// unanchored classification auditable after the fact. Typed matches report
// "code:WORKFLOW_INVALID"; the legacy substring fallback reports the bare
// needle, so the two are distinguishable in a query.
//
// Order matters: the typed code is exact and cannot be defeated by rewording,
// capitalisation, or %w wrapping, so it is tried first. The substring fallback
// remains only for errors that carry no DomainError at all.
//
// An error explicitly built AsRetryable is never permanent, whatever its code
// and whatever its prose says — an author who wrote "retry this" outranks both
// the list and the substring fallback. That early return is load-bearing, not
// tidiness: without it an AsRetryable(ErrValidation) error skips the typed
// branch and is then classified permanent by the fallback matching the word
// "validation" in its own message. Structure must not be overridden by prose;
// being overridden by prose is the entire bug this function exists to fix.
//
// Note what is deliberately NOT done: a non-retryable DomainError whose code is
// not on the list still falls through to the substring fallback. Suppressing
// that would be more principled — the error has already declared its type — but
// it would move existing drops back to retry fleet-wide, which is exactly the
// blast radius 034's lane deferred this work to avoid. This change only ever
// ADDS permanent classifications; it removes none.
func MatchedPermanentFailure(err error) string {
	if err == nil {
		return ""
	}
	if de, ok := errors.AsDomainError(err); ok {
		if de.Retryable {
			return ""
		}
		for _, code := range NonRetryablePermanentCodes {
			if de.Code == code {
				return "code:" + string(code)
			}
		}
	}
	return MatchedValidationNeedle(err.Error())
}

// MatchedValidationNeedle returns the first needle contained in errMsg, or ""
// if none match. Returning the needle rather than a bool is the point: it is
// what makes an unanchored match auditable after the fact.
func MatchedValidationNeedle(errMsg string) string {
	for _, needle := range ValidationErrorNeedles {
		if strings.Contains(errMsg, needle) {
			return needle
		}
	}
	return ""
}

// recordDroppedValidationError persists a dropped validation error to
// agent_error_log so the drop is investigable from the database.
//
// Best-effort, mirroring the 017 precedent (c80fffc83): a failure to record must
// never change the drop decision the caller has already made. Severity is
// 'warning' — a genuine validation error is correctly not retried — but the
// matched needle travels in the context because the match is unanchored and may
// have caught a real runtime failure.
//
// This and agentbase's identically-named method are deliberately two thin
// wrappers over ONE writer (orchestration.LogAgentError), not two
// implementations: they exist separately only because the two layers reach their
// db/logger/context through different structs. Change the row shape in
// LogAgentError; do not fork the wrappers (council 180d7c68, reuse seat).
func (p *MessageProcessor) recordDroppedValidationError(msgCtx *MessageContext, matchedNeedle string, procErr error) {
	db := p.sqlDB
	if db == nil {
		db = p.db
	}
	if db == nil || msgCtx == nil || procErr == nil {
		return
	}

	execCtx := msgCtx.ExecutionContext

	action := execCtx.Action
	if action == "" {
		action = "process_message"
	}

	// Detached from the request context on purpose: this runs on a path that has
	// already failed, and a cancelled request context would silently take the
	// record with it — which is the exact failure this fix exists to end.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orchestration.LogAgentError(ctx, db, p.logger, orchestration.AgentErrorEntry{
		OrchestrationID: execCtx.OrchestrationID,
		AgentType:       p.agentType,
		AgentID:         p.agentID,
		PodName:         p.podName,
		StepName:        execCtx.StepName,
		Action:          action,
		ErrorMessage:    procErr.Error(),
		ErrorCode:       "VALIDATION_ERROR_DROPPED",
		Severity:        "warning",
		Context: map[string]interface{}{
			"correlation_id":  execCtx.CorrelationID,
			"message_id":      execCtx.MessageID,
			"request_id":      execCtx.RequestID,
			"client_id":       execCtx.ClientID,
			"message_type":    execCtx.MessageType,
			"from_agent_type": execCtx.FromAgentType,
			"matched_needle":  matchedNeedle,
			"dropped_at":      "messaging.handleError",
			"retried":         false,
			"note":            "message dropped without retry (see bugs_open/034). matched_needle is an unanchored substring — if this is not a genuine validation error, the classification caught a real runtime failure.",
		},
	})

	p.logger.Info("VALIDATION_ERROR_DROPPED recorded to agent_error_log",
		zap.String("correlation_id", execCtx.CorrelationID),
		zap.String("matched_needle", matchedNeedle))
}
