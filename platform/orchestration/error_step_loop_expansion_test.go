// FILE: platform/orchestration/error_step_loop_expansion_test.go
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// TestLoopExpansionCarriesErrorStep pins the loop half of bugs_open/086.
//
// Two rules, and they are different:
//   - an error_step naming a substep of THIS loop must gain the iteration
//     prefix, or it would name a step that does not exist in the expanded plan;
//   - an error_step naming a step OUTSIDE the loop (mark_item_failed,
//     complete_error, ...) must pass through untouched.
//
// It must NOT go through resolveIterationNextStep: an error handler is not a
// chain link, so the last iteration must not roll it over to the loop's
// _complete step.
func TestLoopExpansionCarriesErrorStep(t *testing.T) {
	s := &SagaCoordinator{logger: zap.NewNop()}

	state := &OrchestrationState{
		CurrentStep:   "build_pages_loop",
		CollectedData: map[string]interface{}{},
		WorkflowPlan: models.WorkflowPlan{
			StartStep: "build_pages_loop",
			Steps: map[string]models.Step{
				"build_pages_loop": {Action: "loop"},
				"complete":         {Action: "complete"},
				"mark_item_failed": {Action: "update_work_item_status"},
			},
		},
	}

	loopResult := map[string]interface{}{
		"loop_name":    "build_pages_loop",
		"loop_var":     "page",
		"next_step":    "complete",
		"output_field": "built_pages",
		"items":        []interface{}{map[string]interface{}{"name": "about"}, map[string]interface{}{"name": "pricing"}},
		"substeps": map[string]models.Step{
			// in-loop target: must be prefixed per iteration
			"write_page_content": {Action: "call_agent", NextStep: "record_result", ErrorStep: "record_result"},
			// external target: must survive verbatim
			"record_result": {Action: "update_work_item_status", ErrorStep: "mark_item_failed"},
		},
		"substep_order": []string{"write_page_content", "record_result"},
	}

	if err := s.handleLoopExpansion(state, loopResult, zap.NewNop()); err != nil {
		t.Fatalf("handleLoopExpansion: %v", err)
	}

	for _, iter := range []int{0, 1} {
		inLoop := state.WorkflowPlan.Steps[iterName("build_pages_loop", iter, "write_page_content")]
		want := iterName("build_pages_loop", iter, "record_result")
		if inLoop.ErrorStep != want {
			t.Errorf("iter %d: in-loop error_step = %q, want %q", iter, inLoop.ErrorStep, want)
		}

		external := state.WorkflowPlan.Steps[iterName("build_pages_loop", iter, "record_result")]
		if external.ErrorStep != "mark_item_failed" {
			t.Errorf("iter %d: external error_step = %q, want %q (unprefixed)", iter, external.ErrorStep, "mark_item_failed")
		}
	}
}

func iterName(loopName string, iterIdx int, substep string) string {
	return loopName + "_iter_" + string(rune('0'+iterIdx)) + "_" + substep
}
