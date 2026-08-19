// Tests for the truncated_component revalidator's pure decision core.
//
// The database glue (revalidateTruncatedComponent) is one delegation call to
// checks.VerifyTruncatedComponentResolved, whose own behaviour — markup-context
// counting, deactivated → resolved, missing row → error — is pinned by that
// package's tests. What is asserted here is the seam this file owns: the
// mapping from the completion verifier's vocabulary onto the sweep's, which is
// what decides whether a parked item closes, stays, or waits.
package actions

import (
	"errors"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestTruncatedComponentVerdictMapping(t *testing.T) {
	const componentID = "fc56f085-8e9a-4f6b-8e8d-600f9a1381e2"

	cases := []struct {
		name        string
		result      checks.VerifyResult
		err         error
		wantVerdict string
		wantArm     string
	}{
		{
			// The direction RFC_017 fixed at completion time, preserved here: an
			// answer that could not be computed must not close the item.
			name:        "verifier error stays parked as unknown",
			err:         errors.New("cannot verify: component no longer exists"),
			wantVerdict: revalidationUnknown,
			wantArm:     "verifier_error",
		},
		{
			name:        "balanced template resolves",
			result:      checks.VerifyResult{Resolved: true, Detail: "html_template balances every paired tag"},
			wantVerdict: revalidationResolved,
			wantArm:     "verifier_resolved",
		},
		{
			name:        "deactivated component resolves",
			result:      checks.VerifyResult{Resolved: true, Detail: "component is deactivated — no longer served"},
			wantVerdict: revalidationResolved,
			wantArm:     "verifier_resolved",
		},
		{
			name:        "still truncated still holds",
			result:      checks.VerifyResult{Resolved: false, Detail: "still truncated: unterminated <script"},
			wantVerdict: revalidationStillHolds,
			wantArm:     "verifier_still_truncated",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := truncatedComponentVerdict(componentID, tc.result, tc.err)
			if v.Verdict != tc.wantVerdict {
				t.Errorf("verdict = %q, want %q", v.Verdict, tc.wantVerdict)
			}
			if v.Arm != tc.wantArm {
				t.Errorf("arm = %q, want %q (the arm must name the rung, not be left for the unreported: fallback)", v.Arm, tc.wantArm)
			}
			if v.Reason == "" {
				t.Error("reason is empty — the recorded verdict must carry the verifier's evidence sentence")
			}
			if got, _ := v.Evidence["component_id"].(string); got != componentID {
				t.Errorf("evidence.component_id = %q, want %q", got, componentID)
			}
		})
	}
}

// The detail sentence is the evidence a human reads off result.revalidation, so
// on the two decided outcomes it must be the verifier's own words, not a
// paraphrase this file could let drift.
func TestTruncatedComponentVerdictCarriesVerifierDetailVerbatim(t *testing.T) {
	const detail = "still truncated: unterminated <style"
	v := truncatedComponentVerdict("x", checks.VerifyResult{Resolved: false, Detail: detail}, nil)
	if v.Reason != detail {
		t.Errorf("reason = %q, want the verifier's detail verbatim %q", v.Reason, detail)
	}
}
