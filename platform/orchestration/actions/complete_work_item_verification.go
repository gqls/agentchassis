// FILE: platform/orchestration/actions/complete_work_item_verification.go
//
// Completion gate: before CompleteWorkItemAction stamps 'complete', ask
// the item-type's verifier (if any) whether the defect is actually gone.
//
// Why: dispatch loops call complete_work_item on EVERY successful handler
// saga. A saga that no-ops "successfully" — e.g. page-build-handler's
// complete_error path for a page with no spec sections — therefore used to
// stamp its item complete with the defect untouched. robot-hands'
// gripper-detail product sections were marked complete on 2026-07-10 and
// still served empty markup on 2026-07-14; the two-strike rule then parked
// the re-detections as non-dispatchable 'unresolved' zombies.
//
// Policy:
//   - no verifier registered for the item_type → complete as before
//   - verifier errors → fail OPEN (complete, recording the error under
//     result._verification) — discovery re-detection + two-strike is the
//     backstop, and failing closed would wedge items on transient errors
//   - verifier says the defect persists → do NOT complete; route into the
//     same attempt machinery as fail_work_item (attempt_count+1 →
//     'triaged' for retry, 'failed' when attempts are exhausted)

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// verifyBeforeComplete runs the registered verifier for the item, if any.
// Returns a map to embed at result._verification (nil when no verifier is
// registered) and whether completion may proceed.
func verifyBeforeComplete(ctx context.Context, db *sql.DB, itemID uuid.UUID, logger *zap.Logger) (map[string]interface{}, bool) {
	var itemType string
	var specJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT item_type, COALESCE(spec, '{}'::jsonb)
		FROM site_work_items WHERE id = $1
	`, itemID).Scan(&itemType, &specJSON)
	if err != nil {
		// Row missing or unreadable — the completion UPDATE will no-op or
		// fail on its own; nothing to verify here.
		return nil, true
	}

	verifier := checks.GetVerifier(itemType)
	if verifier == nil {
		return nil, true
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		logger.Warn("verifyBeforeComplete: unparseable spec — failing open",
			zap.String("item_id", itemID.String()),
			zap.String("item_type", itemType),
			zap.Error(err))
		return map[string]interface{}{
			"status":    "error",
			"item_type": itemType,
			"error":     "unparseable spec: " + err.Error(),
		}, true
	}

	result, err := verifier(ctx, db, spec, logger)
	if err != nil {
		logger.Warn("verifyBeforeComplete: verifier error — failing open",
			zap.String("item_id", itemID.String()),
			zap.String("item_type", itemType),
			zap.Error(err))
		return map[string]interface{}{
			"status":    "error",
			"item_type": itemType,
			"error":     err.Error(),
		}, true
	}

	if result.Resolved {
		return map[string]interface{}{
			"status":    "verified",
			"item_type": itemType,
			"detail":    result.Detail,
		}, true
	}

	logger.Warn("verifyBeforeComplete: defect persists — blocking completion",
		zap.String("item_id", itemID.String()),
		zap.String("item_type", itemType),
		zap.String("detail", result.Detail))
	return map[string]interface{}{
		"status":    "defect_persists",
		"item_type": itemType,
		"detail":    result.Detail,
	}, false
}

// handlerReportedFailure reports whether the result we are about to store is
// itself a record of failure.
//
// The envelope the coordinator writes always carries response_status='complete'
// once a reply arrives (coordinator.go ~2398) — that records DELIVERY, not
// success. The saga's own outcome lives at response.status, with response.error
// carrying the reason: a workflow that never ran at all (an unregistered or
// mistyped action fails validation) comes back as
// response.status='failed' + 'WORKFLOW_INVALID: …'.
//
// Deliberately narrow, keyed on an explicit failure verdict rather than on the
// presence of an error string: a handler may legitimately carry a non-fatal
// error string beside a successful outcome. Measured against live data before
// choosing the predicate — on the 2026-07-18 sweep, 'failed' was the ONLY value
// response.status had ever held anywhere in site_work_items (2905 completed
// items carried no response.status at all; 54 carried 'failed'; nothing else
// existed). Over the 30 days to that date this guard would have blocked 6 of
// 1662 completions, and all 6 were genuine failures.
//
// The allowlist is therefore already a superset of observed reality; the
// default branch logs any future dialect rather than swallowing it.
func handlerReportedFailure(result map[string]interface{}, logger *zap.Logger) (string, bool) {
	resp, ok := result["response"].(map[string]interface{})
	if !ok {
		return "", false
	}

	status, _ := resp["status"].(string)
	switch normalised := strings.ToLower(strings.TrimSpace(status)); normalised {
	case "failed", "failure", "error":
	case "", "success", "complete", "completed", "ok":
		return "", false
	default:
		// An unrecognised verdict. We complete (the conservative choice — a
		// novel status is not evidence of failure), but say so loudly: the
		// allowlist above is a superset of every value response.status has
		// EVER held in this database, so a new one appearing means a handler
		// has started speaking a dialect this guard does not read. That is the
		// precise moment to widen the allowlist — and without this line it
		// would pass silently, which is the very failure mode bugs_open/017
		// was about.
		logger.Warn("handlerReportedFailure: unrecognised response.status — completing, but this verdict is unknown to the guard",
			zap.String("response_status", status),
			zap.String("action", "widen the allowlist in handlerReportedFailure if this is a failure verdict"))
		return "", false
	}

	detail, _ := resp["error"].(string)
	if detail = strings.TrimSpace(detail); detail == "" {
		detail = "handler returned status '" + status + "' with no error detail"
	}
	return detail, true
}

// failUnverifiedCompletion routes a blocked completion into the same
// attempt machinery as FailWorkItemAction's default branch. The handler
// saga's result (including _verification) is preserved for forensics, and
// the claim is released so the item returns to the dispatchable pool.
//
// errorMsg is the recorded reason and reason is the caller-facing code; the
// two callers are the post-fix verifier (defect still present) and the
// handler-failure guard (the saga reported its own failure).
func failUnverifiedCompletion(ctx context.Context, db *sql.DB, itemID uuid.UUID, agentType, resultJSON, errorMsg, reason string, logger *zap.Logger) (interface{}, error) {
	var newStatus string
	err := db.QueryRowContext(ctx, `
		UPDATE site_work_items
		SET attempt_count = attempt_count + 1,
		    error = $2,
		    result = $3::jsonb,
		    status = CASE
		        WHEN attempt_count + 1 >= max_attempts THEN 'failed'
		        ELSE 'triaged'
		    END,
		    claimed_by = NULL,
		    claimed_at = NULL,
		    handled_by = $4,
		    updated_at = NOW()
		WHERE id = $1
		  AND status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')
		RETURNING status
	`, itemID, errorMsg, resultJSON, agentType).Scan(&newStatus)
	if err == sql.ErrNoRows {
		return map[string]interface{}{
			"completed": false,
			"verified":  false,
			"item_id":   itemID.String(),
			"reason":    "already_flagged_or_terminal",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fail unverified completion: %w", err)
	}

	logger.Warn("CompleteWorkItemAction: completion blocked",
		zap.String("item_id", itemID.String()),
		zap.String("new_status", newStatus),
		zap.String("reason", reason),
		zap.String("detail", errorMsg))

	return map[string]interface{}{
		"completed":  false,
		"verified":   false,
		"item_id":    itemID.String(),
		"new_status": newStatus,
		"will_retry": newStatus == "triaged",
		"reason":     reason,
		"detail":     errorMsg,
	}, nil
}
