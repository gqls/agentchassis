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
// Policy (CHANGED 2026-08-08 by owner ruling on architecture_review/RFC_017 —
// the first three lines below used to read "verifier errors → fail OPEN"):
//   - no verifier registered for the item_type → complete as before
//   - verifier errors → fail CLOSED. Do NOT complete; route into the attempt
//     machinery, exactly as for a persisting defect. "I could not check" is no
//     longer treated as "I checked and it is fixed".
//   - unless that item type registered VerifierPolicy{FailOpenOnError: true},
//     which is the explicit, per-type opt-in to the old behaviour
//   - verifier says the defect persists → do NOT complete; route into the
//     same attempt machinery as fail_work_item (attempt_count+1 →
//     'triaged' for retry, 'failed' when attempts are exhausted)
//
// Why the flip, in one line of evidence: the old policy's justification was that
// "discovery re-detection + two-strike is the backstop". Measured 2026-08-08 over
// the registry's whole life, the error path had fired twice, both times the page
// still declared the slot (so the honest verdict was Resolved:false), both items
// were stamped 'complete' at attempt_count 0, and five days later no re-detection
// had followed although the detector's own predicate still matched. The backstop
// did not catch the only two cases it was ever asked to catch. Zero of the errors
// were infrastructural — the transient-DB-blip case fail-open was protecting
// against has never once been observed.

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
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
)

// verifyBeforeComplete runs the registered verifier for the item, if any.
// Returns a map to embed at result._verification (nil when no verifier is
// registered) and whether completion may proceed.
func verifyBeforeComplete(ctx context.Context, db *sql.DB, itemID uuid.UUID, logger *zap.Logger) (map[string]interface{}, bool) {
	var itemType string
	var specJSON []byte
	var siteID uuid.UUID
	var pageID uuid.NullUUID
	err := db.QueryRowContext(ctx, `
		SELECT item_type, COALESCE(spec, '{}'::jsonb), site_id, page_id
		FROM site_work_items WHERE id = $1
	`, itemID).Scan(&itemType, &specJSON, &siteID, &pageID)
	if err != nil {
		// Row missing or unreadable — the completion UPDATE will no-op or
		// fail on its own; nothing to verify here.
		return nil, true
	}

	verifier, policy := checks.GetVerifier(itemType)
	if verifier == nil {
		return nil, true
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		// Same class as a verifier error — verification could not run — so it takes
		// the same policy. Deliberately not exempted: an unparseable spec is if
		// anything MORE suspicious than a verifier's own error, and exempting it
		// would leave a second silent completion path behind the one RFC_017 closed.
		return verificationOutcome(itemType,
			checks.VerifyResult{}, fmt.Errorf("unparseable spec: %w", err), policy, itemID, logger)
	}

	target := checks.VerifyTarget{
		ItemID:   itemID,
		SiteID:   siteID,
		ItemType: itemType,
		Spec:     spec,
	}
	if pageID.Valid {
		target.PageID = &pageID.UUID
	}

	// Scope gate (bugs_open/213): ask the verifier whether its predicate speaks for
	// THIS item before letting it grade one. A verifier registered for an item_type
	// grades every row carrying that name, including rows filed by a producer who
	// meant something else by it entirely — and because the verifier answers its own
	// question correctly, the mismatch reads as a clean pass. Opt-in: Grades is nil
	// for every type that has not asked for this, which is today's behaviour exactly.
	//
	// Placed BEFORE the verifier call, not after, because running a predicate over
	// the wrong item is not merely uninformative — its Resolved:true is the thing
	// that closed 11 defects untouched.
	if policy.Grades != nil {
		if speaks, why := policy.Grades(target); !speaks {
			logger.Warn("verifyBeforeComplete: verifier disclaims this item — blocking completion (out of scope)",
				zap.String("item_id", itemID.String()),
				zap.String("item_type", itemType),
				zap.String("why", why))
			return map[string]interface{}{
				"status":    "out_of_scope",
				"item_type": itemType,
				"detail":    why,
			}, false
		}
	}

	result, err := verifier(ctx, db, target, logger)
	return verificationOutcome(itemType, result, err, policy, itemID, logger)
}

// verificationOutcome turns a verifier's return into the _verification payload
// and the may-complete decision. Split out from verifyBeforeComplete so the
// POLICY is testable without a database — the fail-open/fail-closed branch is the
// whole subject of RFC_017 and was previously reachable only through a live DB
// call, which is why no test asserted it and the behaviour was documented rather
// than proven. verificationDecision below is the pure core; this wrapper only logs.
func verificationOutcome(itemType string, result checks.VerifyResult, err error,
	policy checks.VerifierPolicy, itemID uuid.UUID, logger *zap.Logger) (map[string]interface{}, bool) {

	payload, mayComplete := verificationDecision(itemType, result, err, policy)

	switch payload["status"] {
	case "error":
		if mayComplete {
			logger.Warn("verifyBeforeComplete: verifier error — failing OPEN by explicit policy",
				zap.String("item_id", itemID.String()),
				zap.String("item_type", itemType),
				zap.Error(err))
		} else {
			logger.Warn("verifyBeforeComplete: verification could not run — blocking completion (fail-closed)",
				zap.String("item_id", itemID.String()),
				zap.String("item_type", itemType),
				zap.Error(err))
		}
	case "defect_persists":
		logger.Warn("verifyBeforeComplete: defect persists — blocking completion",
			zap.String("item_id", itemID.String()),
			zap.String("item_type", itemType),
			zap.String("detail", result.Detail))
	}

	return payload, mayComplete
}

// verificationDecision is the pure policy core. Table-tested.
func verificationDecision(itemType string, result checks.VerifyResult, err error,
	policy checks.VerifierPolicy) (map[string]interface{}, bool) {

	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"item_type": itemType,
			"error":     err.Error(),
			// Recorded so a payload says which policy produced the outcome. Without
			// it, 'error' + complete and 'error' + blocked are indistinguishable in
			// the stored row, and the census that produced RFC_017 read exactly this
			// column.
			"fail_open": policy.FailOpenOnError,
		}, policy.FailOpenOnError
	}

	if result.Resolved {
		return map[string]interface{}{
			"status":    "verified",
			"item_type": itemType,
			"detail":    result.Detail,
		}, true
	}

	return map[string]interface{}{
		"status":    "defect_persists",
		"item_type": itemType,
		"detail":    result.Detail,
	}, false
}

// blockedCompletionReason renders the recorded reason for a blocked completion.
//
// Exists because the two blocking causes are not the same claim and must not
// share a sentence: before RFC_017 only 'defect_persists' could block, so the
// caller hard-coded "found the defect still present" — which, once an error can
// block too, would record a finding the verifier never made, on a payload whose
// text lives under "error" rather than "detail" (so the message would also have
// come out empty). Returns (message, reason code).
func blockedCompletionReason(v map[string]interface{}) (string, string) {
	if status, _ := v["status"].(string); status == "error" {
		text, _ := v["error"].(string)
		return "completion blocked: verification could not run, and this item type fails closed (RFC_017): " + text,
			"verification_unavailable"
	}
	// Third cause, and it is a distinct claim from the other two: the verifier did
	// not fail and did not find a defect — it declined to grade this item at all,
	// because its predicate is not the one the item describes (bugs_open/213). An
	// operator reading "found the defect still present" here would go looking for a
	// defect nobody looked for. What is actually owed is a verifier for this item's
	// own shape, or a route to a handler that has one.
	if status, _ := v["status"].(string); status == "out_of_scope" {
		detail, _ := v["detail"].(string)
		return "completion blocked: the verifier registered for this item_type does not grade this item, so nothing checked it (bugs_open/213): " + detail,
			"verifier_scope_mismatch"
	}
	detail, _ := v["detail"].(string)
	return "completion blocked: post-fix verification found the defect still present: " + detail,
		"verification_failed"
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
// default branch records any future dialect rather than swallowing it — see
// recordUnknownVerdict for why a log line alone was not enough.
// Returns (detail, failed, unknownVerdict). unknownVerdict is non-empty when the
// saga reported a status this guard does not recognise: the item still COMPLETES
// (a novel status is not evidence of failure — inverting that would retry real
// work), but the caller must record it. Kept as a return value rather than a log
// call inside here so the function stays pure and table-testable, and so the
// record is written where the DB handle lives.
func handlerReportedFailure(result map[string]interface{}) (string, bool, string) {
	resp, ok := result["response"].(map[string]interface{})
	if !ok {
		return "", false, ""
	}

	status, _ := resp["status"].(string)
	switch normalised := strings.ToLower(strings.TrimSpace(status)); normalised {
	case "failed", "failure", "error":
	case "", "success", "complete", "completed", "ok":
		return "", false, ""
	default:
		// The allowlist above is a superset of every value response.status has
		// EVER held in this database, so a new one appearing means a handler has
		// started speaking a dialect this guard cannot read. That is precisely
		// when to widen the allowlist — and it must not depend on someone
		// happening to read pod logs, which is why this is surfaced rather than
		// merely logged (council objection, bug_historian, 2026-07-18).
		return "", false, status
	}

	detail, _ := resp["error"].(string)
	if detail = strings.TrimSpace(detail); detail == "" {
		detail = "handler returned status '" + status + "' with no error detail"
	}
	return detail, true, ""
}

// recordUnknownVerdict persists an unrecognised handler verdict to
// agent_error_log — a queryable, alertable surface — as well as the pod log.
//
// Why not a log line alone (council objection, bug_historian, 2026-07-18): a
// zap.Warn lives only in an ephemeral pod log, and 016b's own deploy-verification
// pattern records that pod logs do not survive rollouts. A handler that starts
// emitting a new failure dialect would therefore leave NO queryable trace, which
// is the same silent-failure shape bugs_open/017 was filed about. Severity is
// 'warning', not 'error': the completion itself was legitimate under the
// conservative rule; what needs attention is the vocabulary drift.
//
// Best-effort by design — a failure to record must never block a completion that
// the guard has already judged legitimate.
func recordUnknownVerdict(ctx context.Context, params ActionParams, itemID uuid.UUID, status string, logger *zap.Logger) {
	logger.Warn("handlerReportedFailure: unrecognised response.status — completing, but this verdict is unknown to the guard",
		zap.String("response_status", status),
		zap.String("item_id", itemID.String()),
		zap.String("remedy", "widen the allowlist in handlerReportedFailure if this is a failure verdict"))

	if params.DB == nil {
		return
	}

	// work_item_id is the caller's explicit item, not the one params carries —
	// set it so the merge cannot substitute input_data.work_item_id. The
	// PROVENANCE is the running step's (this guard records its own verdict), so
	// inheritance is declared rather than left to a silent merge.
	LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
		WorkItemID:   itemID.String(),
		Action:       "complete_work_item",
		ErrorMessage: "unrecognised handler verdict '" + status + "' — item completed, but this guard cannot tell success from failure for this vocabulary",
		ErrorCode:    "UNKNOWN_HANDLER_VERDICT",
		Severity:     "warning",
		Context: map[string]interface{}{
			"response_status": status,
			"guard":           "handlerReportedFailure",
			"known_verdicts":  []string{"failed", "failure", "error"},
			"remedy":          "if this is a failure verdict, widen the allowlist in handlerReportedFailure (complete_work_item_verification.go); see bugs_open/017",
		},
	}, logger)
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
