// FILE: platform/orchestration/actions/error_step_substep_test.go
package actions

import (
	"testing"

	"go.uber.org/zap"
)

// TestParseSubstepsCarriesErrorStep pins the third drop site of bugs_open/086.
// parseSubsteps builds models.Step field by field exactly as
// convertToWorkflowPlan does, so a substep's step-level error_step never reached
// loop expansion. No agent definition declares one today (censused live: 0 of
// them), which is why this half was latent rather than biting — the test exists
// so the first one that does is not silently inert.
func TestParseSubstepsCarriesErrorStep(t *testing.T) {
	substeps, order, err := parseSubsteps(map[string]interface{}{
		"write_page_content": map[string]interface{}{
			"action":     "call_agent",
			"next_step":  "record_result",
			"error_step": "record_result",
			"config":     map[string]interface{}{"agent_type": "page-content-writer"},
		},
		"record_result": map[string]interface{}{
			"action": "update_work_item_status",
		},
	}, "write_page_content", zap.NewNop())
	if err != nil {
		t.Fatalf("parseSubsteps: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("substep order = %v, want 2 entries", order)
	}

	if got := substeps["write_page_content"].ErrorStep; got != "record_result" {
		t.Errorf("ErrorStep = %q, want %q", got, "record_result")
	}
	// Negative control — a substep with no error_step must stay empty.
	if got := substeps["record_result"].ErrorStep; got != "" {
		t.Errorf("negative control: ErrorStep = %q, want empty", got)
	}
	// Neighbouring fields unaffected.
	if got := substeps["write_page_content"].NextStep; got != "record_result" {
		t.Errorf("NextStep = %q, want %q", got, "record_result")
	}
}
