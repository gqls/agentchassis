// FILE: platform/orchestration/actions/subworkflow_decoder_lockstep_test.go
//
// bugs_open/144 — the workflow validator descends into loop sub-workflows, and to do
// that honestly it must see EXACTLY what the executor executes.
//
// The first version of this file compared two hand-written decoders. The council's
// reuse seat objected that a test proving two copies agree is a backstop rather than
// single-sourcing, and that the founding incident is what happens when the backstop is
// all there is. So there is now ONE decoder — models.DecodeSubWorkflowStep — called by
// parseSubsteps here and by the validator there.
//
// WHAT IS LEFT TO TEST, HONESTLY: while parseSubsteps delegates, the first assertion
// below cannot fail. It is a regression guard, not a proof — it exists to fail the day
// someone re-inlines the decode into this file, which is how the duplication would come
// back. The assertions that can fail today are in pkg/models/substep_decode_test.go
// (a new field on models.Step that neither side decides about) and in
// platform/validation/subworkflow_test.go (what the validator does with the result).
package actions

import (
	"reflect"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// everyStepField carries every field models.Step declares a json tag for, so the test
// sees a difference whichever side changes.
func everyStepField() map[string]interface{} {
	return map[string]interface{}{
		// Read by the nested decoder.
		"action":       "write_page_content",
		"description":  "write one page",
		"next_step":    "save_page",
		"error_step":   "mark_failed",
		"output_field": "page_html",
		"topic":        "system.writer.requests",
		"config":       map[string]interface{}{"page_field": "loop_item"},
		// NOT read by the nested decoder — dropped before execution. The top-level
		// decoder (processor.convertToWorkflowPlan) reads dependencies; this one does
		// not, and that difference is deliberate rather than an oversight.
		"dependencies":      []interface{}{"some_other_step"},
		"sub_tasks":         []interface{}{map[string]interface{}{"step_name": "a", "topic": "t"}},
		"name":              "write",
		"target_agent_type": "content-writer",
		"store_memory":      true,
		"timeout":           30,
	}
}

// TestParseSubstepsUsesTheSharedDecoder: fails if the decode is ever re-inlined here.
func TestParseSubstepsUsesTheSharedDecoder(t *testing.T) {
	raw := everyStepField()

	runtime, _, err := parseSubsteps(map[string]interface{}{"write": raw}, "", zap.NewNop())
	if err != nil {
		t.Fatalf("parseSubsteps failed on a well-formed step: %v", err)
	}

	shared, _ := models.DecodeSubWorkflowStep(raw)
	if !reflect.DeepEqual(runtime["write"], shared) {
		t.Fatalf("parseSubsteps no longer produces what models.DecodeSubWorkflowStep produces.\n"+
			"  parseSubsteps: %#v\n  shared:        %#v\n"+
			"The nested decode is single-sourced on purpose (bugs_open/144): the validator "+
			"decodes with the shared function, so a second decode here means the validator is "+
			"checking something the executor does not run.", runtime["write"], shared)
	}
}

// TestParseSubstepsDropsWhatTheValidatorReports pins the consequence rather than the
// mechanism: whatever the decoder does, a substep reaching the executor must not carry
// the fields the validator tells operators are being dropped.
func TestParseSubstepsDropsWhatTheValidatorReports(t *testing.T) {
	substeps, _, err := parseSubsteps(map[string]interface{}{"write": everyStepField()}, "", zap.NewNop())
	if err != nil {
		t.Fatalf("parseSubsteps failed: %v", err)
	}
	step := substeps["write"]

	if len(step.Dependencies) != 0 {
		t.Errorf("a substep must reach the executor with no Dependencies (the expander does not honour them), got %v", step.Dependencies)
	}
	if len(step.SubTasks) != 0 {
		t.Errorf("a substep must reach the executor with no SubTasks — this is why a nested fan_out is rejected at validation, got %v", step.SubTasks)
	}
	if step.Timeout != 0 || step.StoreMemory || step.TargetAgentType != "" || step.Name != "" {
		t.Errorf("a substep carried a field the loop decoder does not read: %#v", step)
	}

	// And the fields that ARE honoured must survive, or the drop test above would pass
	// against a decoder that dropped everything.
	if step.Action != "write_page_content" || step.ErrorStep != "mark_failed" || step.OutputField != "page_html" {
		t.Errorf("honoured fields did not survive: %#v", step)
	}
}
