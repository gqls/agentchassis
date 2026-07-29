// FILE: pkg/models/substep_decode.go
//
// ONE decoder for a step nested inside a loop's sub-workflow, shared by the executor
// and the validator.
//
// bugs_open/144 was two hand-written traversals disagreeing, and the first fix for it
// shipped a hand-written second DECODER pinned to the first by a lockstep test. The
// council's reuse seat objected — correctly: a test that two copies agree is a
// backstop, not single-sourcing, and the founding incident is exactly what happens
// when the backstop is the only thing holding them together. So the decode lives here,
// in the package that owns Step, and both sides call it:
//
//   - platform/orchestration/actions/loop_actions.go  parseSubsteps  (the executor)
//   - platform/validation/subworkflow.go              (the validator)
//
// WHY IT IS NOT models.Step's json tags. A nested step is NOT decoded like a
// top-level one. The top-level decoder (messaging.convertToWorkflowPlan) also reads
// `dependencies`; this one does not, and never has. Unmarshalling a nested step into
// Step by JSON round-trip would populate fields the executor never sees, and any
// validator built on that would vouch for behaviour that does not happen. The
// difference is not a bug to be tidied away — it is the contract, and this file is
// where it is stated once.
package models

import "sort"

// SubWorkflowStepFields is exactly the set of keys a nested step definition is read
// for. Anything else present on a nested step is dropped before execution.
//
// Adding a field to Step does NOT add it here. That is deliberate: the honoured set is
// a statement about what the loop executor does, not about what the struct can hold.
// pkg/models/substep_decode_test.go enumerates Step's json tags and fails if a new one
// is neither honoured here nor listed as knowingly dropped — so a new field forces a
// decision instead of being silently ignored by both sides at once.
var SubWorkflowStepFields = map[string]bool{
	"action":       true,
	"config":       true,
	"description":  true,
	"error_step":   true,
	"next_step":    true,
	"output_field": true,
	"topic":        true,
}

// DecodeSubWorkflowStep decodes one nested step exactly as the loop action executes
// it, and returns the keys it had to drop, sorted.
//
// The dropped list is the point of the second return value: a nested `dependencies`
// or `sub_tasks` is not an error, it is config describing behaviour that does not
// happen — the same defect class as an unknown config key, one level up. Callers
// report it; nobody enforces it, because there is nothing to enforce.
func DecodeSubWorkflowStep(raw map[string]interface{}) (Step, []string) {
	step := Step{
		Action:      rawStepString(raw, "action"),
		Description: rawStepString(raw, "description"),
		NextStep:    rawStepString(raw, "next_step"),
		// error_step is the step-level twin of config.error_step and
		// routeToErrorStepOrFail prefers it; omitting it made every step-level
		// declaration inert fleet-wide (bugs_open/086). Expansion prefixes it per
		// iteration when it names a sibling substep.
		ErrorStep:   rawStepString(raw, "error_step"),
		OutputField: rawStepString(raw, "output_field"),
		Topic:       rawStepString(raw, "topic"),
	}

	if config, ok := raw["config"].(map[string]interface{}); ok {
		step.Config = config
	} else {
		// An absent config becomes an empty map, never nil: the executor writes
		// iteration variables into it, and a nil map would panic on the first write.
		step.Config = make(map[string]interface{})
	}

	var dropped []string
	for key := range raw {
		if !SubWorkflowStepFields[key] {
			dropped = append(dropped, key)
		}
	}
	sort.Strings(dropped)

	return step, dropped
}

func rawStepString(m map[string]interface{}, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}
