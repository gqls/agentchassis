// FILE: pkg/models/substep_decode_test.go
//
// The guard that replaced a lockstep test between two hand-written decoders
// (bugs_open/144, council reuse-seat objection on corr 9194bc97). With one decoder
// there is nothing to hold in lockstep — so the guard moves to the thing that can
// still drift silently: a field added to Step that neither side decides about.
package models

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// knowinglyDropped is every json-tagged field of Step that a nested step may carry and
// the loop executor does NOT read. Stated explicitly, not derived — deriving it from
// the honoured set would make this test agree with whatever the code does, which is
// the failure mode it exists to prevent.
var knowinglyDropped = map[string]string{
	"dependencies":      "the TOP-LEVEL decoder reads this; the loop decoder never has. A nested step's dependencies describe ordering that does not happen.",
	"sub_tasks":         "so a nested fan_out would execute with none — which is why the validator rejects a nested fan_out outright.",
	"store_memory":      "never read for a substep.",
	"timeout":           "never read for a substep; the loop's own max_iterations and the step timeout config are unrelated.",
	"name":              "the substep's map key IS its name; a `name` field is decoration.",
	"target_agent_type": "never read for a substep.",
}

// TestEveryStepFieldIsHonouredOrKnowinglyDropped: add a field to Step and this test
// makes you decide what a nested step does with it. Without it, a new field is
// silently ignored by the executor AND by the validator at the same time — which is
// bugs_open/144's exact shape, one level down.
func TestEveryStepFieldIsHonouredOrKnowinglyDropped(t *testing.T) {
	stepType := reflect.TypeOf(Step{})

	var undecided []string
	for i := 0; i < stepType.NumField(); i++ {
		tag := stepType.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			continue
		}
		if SubWorkflowStepFields[name] {
			continue
		}
		if _, known := knowinglyDropped[name]; known {
			continue
		}
		undecided = append(undecided, name)
	}

	if len(undecided) > 0 {
		sort.Strings(undecided)
		t.Fatalf("models.Step gained field(s) %v that a nested sub-workflow step neither honours nor knowingly drops.\n"+
			"Decide, and say which in code:\n"+
			"  - the loop executor should read it   → add it to SubWorkflowStepFields AND to DecodeSubWorkflowStep\n"+
			"  - it is meaningless inside a loop     → add it to knowinglyDropped here, with the reason\n"+
			"Doing neither means the executor and the validator both ignore it silently, which is bugs_open/144 again.",
			undecided)
	}
}

// TestSubWorkflowStepFieldsAreAllRealStepFields is the other direction: an honoured
// key that does not correspond to a Step field would be decoded into nothing, and the
// report of dropped keys would then be quietly wrong.
func TestSubWorkflowStepFieldsAreAllRealStepFields(t *testing.T) {
	stepType := reflect.TypeOf(Step{})
	real := map[string]bool{}
	for i := 0; i < stepType.NumField(); i++ {
		if tag := stepType.Field(i).Tag.Get("json"); tag != "" && tag != "-" {
			real[strings.Split(tag, ",")[0]] = true
		}
	}
	for honoured := range SubWorkflowStepFields {
		if !real[honoured] {
			t.Errorf("SubWorkflowStepFields honours %q, which is not a json field of models.Step — it decodes into nothing", honoured)
		}
	}
}

func TestDecodeSubWorkflowStepReadsSevenFieldsAndReportsTheRest(t *testing.T) {
	step, dropped := DecodeSubWorkflowStep(map[string]interface{}{
		"action":       "write_page_content",
		"description":  "write one page",
		"next_step":    "save_page",
		"error_step":   "mark_failed",
		"output_field": "page_html",
		"topic":        "system.writer.requests",
		"config":       map[string]interface{}{"page_field": "loop_item"},
		"dependencies": []interface{}{"some_other_step"},
		"timeout":      30,
	})

	if step.Action != "write_page_content" || step.ErrorStep != "mark_failed" || step.Topic != "system.writer.requests" {
		t.Fatalf("honoured fields did not survive the decode: %#v", step)
	}
	if step.OutputField != "page_html" || step.NextStep != "save_page" || step.Description != "write one page" {
		t.Fatalf("honoured fields did not survive the decode: %#v", step)
	}
	if step.Config["page_field"] != "loop_item" {
		t.Fatalf("config did not survive the decode: %#v", step.Config)
	}
	// The fields the executor does not read must be absent from the decoded step, not
	// merely reported: a validator built on a Step carrying Dependencies would enforce
	// ordering that never happens.
	if len(step.Dependencies) != 0 {
		t.Fatalf("dependencies must NOT be decoded for a nested step, got %v", step.Dependencies)
	}
	if step.Timeout != 0 {
		t.Fatalf("timeout must NOT be decoded for a nested step, got %v", step.Timeout)
	}
	if !reflect.DeepEqual(dropped, []string{"dependencies", "timeout"}) {
		t.Fatalf("dropped list wrong: %v", dropped)
	}

	// Positive control: a step using only honoured fields reports nothing dropped.
	// Without it, a decoder that called everything dropped would pass the above.
	_, none := DecodeSubWorkflowStep(map[string]interface{}{"action": "save_page_sections"})
	if len(none) != 0 {
		t.Fatalf("an ordinary step must report nothing dropped, got %v", none)
	}
}

// TestDecodeSubWorkflowStepAlwaysGivesAUsableConfig: the executor writes iteration
// variables into step.Config, so a nil map would panic on the first write.
func TestDecodeSubWorkflowStepAlwaysGivesAUsableConfig(t *testing.T) {
	for _, raw := range []map[string]interface{}{
		{"action": "x"},
		{"action": "x", "config": "a string, not an object"},
		{"action": "x", "config": nil},
	} {
		step, _ := DecodeSubWorkflowStep(raw)
		if step.Config == nil {
			t.Fatalf("config must never decode to nil (input %v)", raw)
		}
		step.Config["loop_item"] = 1 // must not panic
	}
}
