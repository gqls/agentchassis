// FILE: platform/orchestration/actions/subworkflow_decoder_lockstep_test.go
//
// bugs_open/144 — the workflow validator now descends into loop sub-workflows, and
// to do that honestly it must decode a nested step EXACTLY as the loop action does.
// parseSubsteps reads seven fields and silently drops the rest, so a validator that
// decoded by JSON round-trip would populate fields the executor never sees and then
// vouch for behaviour that does not happen.
//
// This test is the lockstep: validation.DecodeSubWorkflowStep must produce the same
// models.Step as parseSubsteps for the same input. If either side starts reading a
// new field, this fails — which is the only mechanism that stops the two drifting.
// Two hand-maintained copies of one contract is precisely the class of defect
// bugs_open/144 was: the runtime and the audit were blind in the same direction and
// agreed with each other, and consistent blindness reads exactly like correctness.
//
// It lives in this package rather than in validation because parseSubsteps is
// unexported here, and package validation is importable from here (no cycle: actions
// does not depend on validation).
package actions

import (
	"reflect"
	"sort"
	"testing"

	"github.com/gqls/agentchassis/platform/validation"
	"go.uber.org/zap"
)

// everyStepField carries every field models.Step declares a json tag for, so the
// test sees a difference whichever side changes.
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
		// decoder (processor.convertToWorkflowPlan) reads dependencies; the nested
		// one does not, and that difference is the thing under test.
		"dependencies":      []interface{}{"some_other_step"},
		"sub_tasks":         []interface{}{map[string]interface{}{"step_name": "a", "topic": "t"}},
		"name":              "write",
		"target_agent_type": "content-writer",
		"store_memory":      true,
		"timeout":           30,
	}
}

func TestSubWorkflowDecoderMatchesRuntime(t *testing.T) {
	raw := everyStepField()

	runtime, _, err := parseSubsteps(map[string]interface{}{"write": raw}, "", zap.NewNop())
	if err != nil {
		t.Fatalf("parseSubsteps failed on a well-formed step: %v", err)
	}

	mirror, dropped := validation.DecodeSubWorkflowStep(raw)

	if !reflect.DeepEqual(runtime["write"], mirror) {
		t.Fatalf("the validator's nested decoder and the loop action's have diverged.\n"+
			"  runtime (parseSubsteps):              %#v\n"+
			"  validator (DecodeSubWorkflowStep):    %#v\n"+
			"Whichever side gained a field, the other must gain it too — or the validator "+
			"is checking something the executor does not run, or missing something it does.",
			runtime["write"], mirror)
	}

	// The reported drop list must be exactly the fields the runtime ignored. Stated
	// explicitly rather than derived, so adding a field to the honoured set is a
	// deliberate edit here and not a silent widening.
	want := []string{"dependencies", "name", "store_memory", "sub_tasks", "target_agent_type", "timeout"}
	sort.Strings(dropped)
	if !reflect.DeepEqual(dropped, want) {
		t.Fatalf("dropped-field report is wrong:\n  got  %v\n  want %v", dropped, want)
	}
}

// TestSubWorkflowDecoderReportsNothingForAnHonouredStep is the positive control: a
// step using only honoured fields must report no drops. Without it the test above
// would pass just as well against a decoder that called everything dropped.
func TestSubWorkflowDecoderReportsNothingForAnHonouredStep(t *testing.T) {
	raw := map[string]interface{}{
		"action":       "save_page_sections",
		"description":  "save",
		"next_step":    "done",
		"error_step":   "fail",
		"output_field": "saved",
		"topic":        "",
		"config":       map[string]interface{}{},
	}

	runtime, _, err := parseSubsteps(map[string]interface{}{"save": raw}, "", zap.NewNop())
	if err != nil {
		t.Fatalf("parseSubsteps failed: %v", err)
	}
	mirror, dropped := validation.DecodeSubWorkflowStep(raw)

	if len(dropped) != 0 {
		t.Fatalf("no field here is dropped by the runtime, got %v", dropped)
	}
	if !reflect.DeepEqual(runtime["save"], mirror) {
		t.Fatalf("decoders disagree on an ordinary step:\n  runtime %#v\n  mirror  %#v", runtime["save"], mirror)
	}
}
