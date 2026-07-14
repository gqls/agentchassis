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

// failUnverifiedCompletion routes a blocked completion into the same
// attempt machinery as FailWorkItemAction's default branch. The handler
// saga's result (including _verification) is preserved for forensics, and
// the claim is released so the item returns to the dispatchable pool.
func failUnverifiedCompletion(ctx context.Context, db *sql.DB, itemID uuid.UUID, agentType, resultJSON, detail string, logger *zap.Logger) (interface{}, error) {
	errorMsg := "completion blocked: post-fix verification found the defect still present: " + detail

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

	logger.Warn("CompleteWorkItemAction: completion blocked by verification",
		zap.String("item_id", itemID.String()),
		zap.String("new_status", newStatus),
		zap.String("detail", detail))

	return map[string]interface{}{
		"completed":  false,
		"verified":   false,
		"item_id":    itemID.String(),
		"new_status": newStatus,
		"will_retry": newStatus == "triaged",
		"reason":     "verification_failed",
		"detail":     detail,
	}, nil
}
