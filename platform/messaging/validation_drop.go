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
// Ordering matters only for which needle gets reported when several match.
var ValidationErrorNeedles = []string{"is required", "validation", "invalid", "missing"}

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
