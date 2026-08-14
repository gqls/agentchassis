// FILE: platform/messaging/validation_drop.go
//
// The "validation error" drop path, and its durable record (bugs_open/034).
//
// Two layers independently decide that a failed message is a validation error
// and must NOT be retried: MessageProcessor.handleError here, and
// agentbase.Agent.processMessage above it. Both decided it by substring match,
// against two lists that had silently drifted apart — one shared list is what
// ended that, so the classification cannot differ by layer again.
//
// THE CLASSIFICATION CORE NOW LIVES IN platform/errors (permanent_failure.go),
// moved 2026-08-08 (bugs_open/217): the coordinator's child-orchestration
// failure sender needs RetryDisposition, and this package imports
// platform/orchestration (see the recorder below), so the classifier had to
// move DOWN to a leaf both layers can import. The names below are re-exports
// aliasing that one implementation — same backing slices, same functions — so
// every existing caller (processor.go, agentbase/agent.go) and every pinning
// test keeps compiling against the single source of truth. Do NOT add a second
// list here; that is the 034 → 195 drift this file exists to prevent.
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
	"time"

	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/orchestration"
	"go.uber.org/zap"
)

// ValidationErrorNeedles re-exports the shared permanent-needle list. The
// documentation of record — the unanchored-match hazards, the 195 case-fold
// argument — travels with the list: platform/errors/permanent_failure.go.
var ValidationErrorNeedles = errors.ValidationErrorNeedles

// NonRetryablePermanentCodes re-exports the typed permanent code list.
var NonRetryablePermanentCodes = errors.NonRetryablePermanentCodes

// MatchedPermanentFailure delegates to the shared classifier in
// platform/errors — typed code first, substring fallback second. See
// permanent_failure.go for the contract and its history (034 → 195).
func MatchedPermanentFailure(err error) string {
	return errors.MatchedPermanentFailure(err)
}

// MatchedValidationNeedle delegates to the shared substring fallback.
func MatchedValidationNeedle(errMsg string) string {
	return errors.MatchedValidationNeedle(errMsg)
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
//
// This function's orchestration.LogAgentError call is WHY the classification
// core could not stay in this package (messaging → orchestration, so
// orchestration could never import messaging).
func (p *MessageProcessor) recordDroppedValidationError(msgCtx *MessageContext, matchedNeedle string, procErr error) {
	// This preferred p.sqlDB and fell back to p.db — the operands the opposite way
	// round from the rest of the file, so it reached for the handle that was
	// always nil on a chassis pod and only ever recorded anything by falling
	// through. bugs_open/259 deleted p.sqlDB, so there is one handle to read.
	db := p.db
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
