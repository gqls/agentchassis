// FILE: platform/orchestration/actions/fail_workflow_action_test.go

package actions

import "testing"

// The reason is what a human reads in agent_error_log when a report never
// arrived, so a runtime message must beat the static label — otherwise every
// failure reads the same and the actual error is lost.
func TestFailWorkflowReasonPrecedence(t *testing.T) {
	collected := map[string]interface{}{
		"verify": map[string]interface{}{
			"error": "prose asserts number \"999\" not present in the fact block",
		},
		"blank": map[string]interface{}{"error": "   "},
	}

	cases := []struct {
		name   string
		config map[string]interface{}
		want   string
	}{
		{
			name:   "runtime message wins",
			config: map[string]interface{}{"reason": "static", "reason_field": "verify.error"},
			want:   "prose asserts number \"999\" not present in the fact block",
		},
		{
			name:   "falls back to static when the path is absent",
			config: map[string]interface{}{"reason": "verification failed", "reason_field": "nope.missing"},
			want:   "verification failed",
		},
		{
			name:   "a blank runtime message is not a message",
			config: map[string]interface{}{"reason": "verification failed", "reason_field": "blank.error"},
			want:   "verification failed",
		},
		{
			name:   "never silently empty",
			config: map[string]interface{}{},
			want:   "no reason given",
		},
	}
	for _, c := range cases {
		if got := failWorkflowReason(c.config, collected); got != c.want {
			t.Errorf("%s: failWorkflowReason = %q, want %q", c.name, got, c.want)
		}
	}
}
