// FILE: platform/orchestration/actions/complete_work_item_verification_test.go
//
// Covers the completion-lie guard from bugs_open/017: CompleteWorkItemAction
// used to stamp 'complete' on items whose stored result was itself a record of
// failure, because it read only the delivery envelope (response_status) and
// never the saga's own verdict (response.status).

package actions

import "testing"

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
