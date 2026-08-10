// FILE: platform/orchestration/actions/complete_work_item_verification_test.go
//
// Covers the completion-lie guard from bugs_open/017: CompleteWorkItemAction
// used to stamp 'complete' on items whose stored result was itself a record of
// failure, because it read only the delivery envelope (response_status) and
// never the saga's own verdict (response.status).

package actions

import (
	"errors"
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestHandlerReportedFailure(t *testing.T) {
	tests := []struct {
		name       string
		result     map[string]interface{}
		wantFailed bool
		wantDetail string
		// wantUnknown is the status the guard could not classify: the item
		// completes, but recordUnknownVerdict must surface it to agent_error_log.
		wantUnknown string
	}{
		{
			// The exact shape stored for work item e4fd567e (robot-hands,
			// 2026-07-17): delivery succeeded, the workflow never ran.
			name: "WORKFLOW_INVALID saga is a failure despite response_status complete",
			result: map[string]interface{}{
				"response": map[string]interface{}{
					"status": "failed",
					"error":  "WORKFLOW_INVALID: Invalid workflow configuration (caused by: step 'fix_text_colors' with action 'fix_forced_text_colors' requires a topic)",
				},
				"response_status":      "complete",
				"response_received_at": "2026-07-17T13:32:24Z",
			},
			wantFailed: true,
			wantDetail: "WORKFLOW_INVALID: Invalid workflow configuration (caused by: step 'fix_text_colors' with action 'fix_forced_text_colors' requires a topic)",
		},
		{
			// The overwhelming majority shape: 2905 of 2959 completed items on
			// the 2026-07-18 sweep carried no response.status at all.
			name: "healthy saga with no response.status completes",
			result: map[string]interface{}{
				"response":        map[string]interface{}{"components_fixed": 3},
				"response_status": "complete",
			},
			wantFailed: false,
		},
		{
			name:       "no response key at all completes",
			result:     map[string]interface{}{"commit_sha": "f32b208e5"},
			wantFailed: false,
		},
		{
			name: "explicit success completes",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": "success"},
			},
			wantFailed: false,
		},
		{
			// An error string alone must NOT block: handlers may carry a
			// non-fatal error field beside a successful outcome. Only an
			// explicit failure verdict blocks completion.
			name: "error string without a failure verdict completes",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": "success", "error": "1 of 4 pages skipped"},
			},
			wantFailed: false,
		},
		{
			name: "failure verdict with no detail still blocks",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": "failed"},
			},
			wantFailed: true,
			wantDetail: "handler returned status 'failed' with no error detail",
		},
		{
			name: "case and whitespace tolerated",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": " FAILED ", "error": "boom"},
			},
			wantFailed: true,
			wantDetail: "boom",
		},
		{
			name: "error verdict blocks",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": "error", "error": "adapter timeout"},
			},
			wantFailed: true,
			wantDetail: "adapter timeout",
		},
		{
			name:       "non-object response is ignored",
			result:     map[string]interface{}{"response": "just a string"},
			wantFailed: false,
		},
		{
			// Council objection (bug_historian, 2026-07-18, rounds 1+2): the
			// allowlist cannot know a future handler's dialect. An unrecognised
			// verdict must COMPLETE (a novel status is not evidence of failure)
			// but must not pass silently — it is returned as unknownVerdict so
			// the caller records it to agent_error_log, a queryable surface,
			// rather than only to an ephemeral pod log. Both halves pinned here.
			name: "unrecognised verdict completes rather than guessing",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": "timeout", "error": "upstream slow"},
			},
			wantFailed:  false,
			wantUnknown: "timeout",
		},
		{
			name: "explicit success vocabulary completes",
			result: map[string]interface{}{
				"response": map[string]interface{}{"status": "completed"},
			},
			wantFailed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			detail, failed, unknown := handlerReportedFailure(tc.result)
			if tc.wantUnknown != "" && unknown != tc.wantUnknown {
				t.Errorf("unknownVerdict = %q, want %q", unknown, tc.wantUnknown)
			}
			if tc.wantUnknown == "" && unknown != "" {
				t.Errorf("unknownVerdict = %q, want empty", unknown)
			}
			if failed != tc.wantFailed {
				t.Fatalf("handlerReportedFailure() failed = %v, want %v (detail %q)", failed, tc.wantFailed, detail)
			}
			if tc.wantFailed && detail != tc.wantDetail {
				t.Errorf("detail = %q, want %q", detail, tc.wantDetail)
			}
		})
	}
}

// TestVerificationDecision covers the fail-closed flip (owner ruling on RFC_017,
// 2026-08-08). Every case below EXCEPT the explicit opt-in row would have passed
// under the old policy with mayComplete=true — so the table is written to fail
// loudly if the default is ever flipped back by accident rather than by ruling.
func TestVerificationDecision(t *testing.T) {
	closed := checks.VerifierPolicy{}                    // the default
	open := checks.VerifierPolicy{FailOpenOnError: true} // explicit opt-in

	tests := []struct {
		name            string
		result          checks.VerifyResult
		err             error
		policy          checks.VerifierPolicy
		wantMayComplete bool
		wantStatus      string
	}{
		{
			// THE RULING, in one row. Pre-2026-08-08 this returned true.
			name:            "verifier error fails CLOSED by default",
			err:             errors.New("cannot verify: component 123 no longer exists"),
			policy:          closed,
			wantMayComplete: false,
			wantStatus:      "error",
		},
		{
			// The escape hatch must still work, or the flip is a wall rather than a
			// default and the next author routes around the registry entirely.
			name:            "verifier error fails OPEN only with the explicit opt-in",
			err:             errors.New("transient: dial tcp: connection refused"),
			policy:          open,
			wantMayComplete: true,
			wantStatus:      "error",
		},
		{
			name:            "resolved completes",
			result:          checks.VerifyResult{Resolved: true, Detail: "no findings"},
			policy:          closed,
			wantMayComplete: true,
			wantStatus:      "verified",
		},
		{
			// An opt-in to fail-open must NOT leak into the verdict path: it licenses
			// completing when the check could not RUN, never when it ran and said no.
			name:            "fail-open policy does not rescue a persisting defect",
			result:          checks.VerifyResult{Resolved: false, Detail: "18 findings still present"},
			policy:          open,
			wantMayComplete: false,
			wantStatus:      "defect_persists",
		},
		{
			// err wins over a Resolved:true carried alongside it — a verifier that
			// returns both has not verified anything.
			name:            "error beats a resolved result returned with it",
			result:          checks.VerifyResult{Resolved: true, Detail: "ignore me"},
			err:             errors.New("boom"),
			policy:          closed,
			wantMayComplete: false,
			wantStatus:      "error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, mayComplete := verificationDecision("empty_section", tc.result, tc.err, tc.policy)
			if mayComplete != tc.wantMayComplete {
				t.Fatalf("mayComplete = %v, want %v (payload %v)", mayComplete, tc.wantMayComplete, payload)
			}
			if got, _ := payload["status"].(string); got != tc.wantStatus {
				t.Errorf("status = %q, want %q", got, tc.wantStatus)
			}
			if payload["item_type"] != "empty_section" {
				t.Errorf("item_type = %v, want empty_section", payload["item_type"])
			}
			if tc.err != nil {
				// The stored row must say which policy produced the outcome; the
				// RFC_017 census read exactly this payload and could not tell a
				// completed error from a blocked one.
				if got, ok := payload["fail_open"].(bool); !ok || got != tc.policy.FailOpenOnError {
					t.Errorf("fail_open = %v (ok=%v), want %v", got, ok, tc.policy.FailOpenOnError)
				}
			}
		})
	}
}

// TestBlockedCompletionReason: the two blocking causes must not share a sentence.
// Before the flip only a persisting defect could block, so the caller hard-coded
// "found the defect still present" and read the text from "detail" — on an error
// payload that produces a false claim with an empty body.
func TestBlockedCompletionReason(t *testing.T) {
	msg, reason := blockedCompletionReason(map[string]interface{}{
		"status": "error",
		"error":  "cannot verify: component 123 no longer exists",
	})
	if reason != "verification_unavailable" {
		t.Errorf("reason = %q, want verification_unavailable", reason)
	}
	if strings.Contains(msg, "still present") {
		t.Errorf("an unrunnable verification must not claim the defect was found present: %q", msg)
	}
	if !strings.Contains(msg, "component 123 no longer exists") {
		t.Errorf("message must carry the verifier's own text, got %q", msg)
	}

	msg, reason = blockedCompletionReason(map[string]interface{}{
		"status": "defect_persists",
		"detail": "18 findings still present",
	})
	if reason != "verification_failed" {
		t.Errorf("reason = %q, want verification_failed", reason)
	}
	if !strings.Contains(msg, "18 findings still present") {
		t.Errorf("message must carry the detail, got %q", msg)
	}
}

// TestBlockedCompletionReasonOutOfScope covers the third blocking cause
// (bugs_open/213). It is a distinct claim from the other two and must not borrow
// either one's sentence: the verifier did not fail, and it did not find a defect —
// it declined to grade this item at all, because its predicate is not the one the
// item describes.
//
// The wording matters more here than for the other two. An operator who reads
// "found the defect still present" on one of these goes looking for a defect that
// nobody looked for; an operator who reads "verification could not run" goes
// looking for an outage. What is actually owed is a verifier for this item's own
// shape.
func TestBlockedCompletionReasonOutOfScope(t *testing.T) {
	msg, reason := blockedCompletionReason(map[string]interface{}{
		"status": "out_of_scope",
		"detail": "this verifier re-runs the discovery check's site-wide aggregate predicate",
	})
	if reason != "verifier_scope_mismatch" {
		t.Errorf("reason = %q, want verifier_scope_mismatch", reason)
	}
	if strings.Contains(msg, "still present") {
		t.Errorf("an ungraded item must not claim the defect was found present: %q", msg)
	}
	if strings.Contains(msg, "could not run") {
		t.Errorf("an out-of-scope disclaimer is not an error — it must not read as one: %q", msg)
	}
	if !strings.Contains(msg, "site-wide aggregate predicate") {
		t.Errorf("message must carry the verifier's own reason, got %q", msg)
	}
}
