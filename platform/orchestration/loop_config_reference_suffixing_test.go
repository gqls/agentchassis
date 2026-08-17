// FILE: platform/orchestration/loop_config_reference_suffixing_test.go
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// TestLoopConfigReferenceSuffixing pins the generic suffixing pass in
// prefixConfigStepReferences (bugs_open/287 §9a, spawn_record slug).
//
// The old behaviour rewrote sibling-output references only for an allow-list of
// config keys; complete_work_item's "result" was not on it, so the live
// build-dispatch-loop's `"result": "handler_result"` was never rewritten to the
// iteration-suffixed key and could only resolve through the resolver's
// whole-tree search — which is how ~75% of dispatch-loop completions stored the
// SPAWN RECORD instead of the handler's reply. The fix stops enumerating: ANY
// top-level string config value that is reference-shaped and whose first dotted
// segment names a sibling output_field is rewritten.
//
// Mutation checks (run by hand; recorded in the lane's NOTES):
//   - delete the generic pass in prefixConfigStepReferences -> the "result",
//     "result!" and "commit_sha" assertions fail;
//   - delete the referenceShapedConfigValue gate -> the prose and condition
//     assertions fail (whole-string rewrite of an expression).
func TestLoopConfigReferenceSuffixing(t *testing.T) {
	s := &SagaCoordinator{logger: zap.NewNop()}

	state := &OrchestrationState{
		CurrentStep:   "process_item",
		CollectedData: map[string]interface{}{},
		WorkflowPlan: models.WorkflowPlan{
			StartStep: "process_item",
			Steps: map[string]models.Step{
				"process_item": {Action: "loop"},
				"complete":     {Action: "complete"},
			},
		},
	}

	loopResult := map[string]interface{}{
		"loop_name":    "process_item",
		"loop_var":     "current_item",
		"next_step":    "complete",
		"output_field": "processed_items",
		"items": []interface{}{
			map[string]interface{}{"id": "a"},
			map[string]interface{}{"id": "b"},
		},
		"substeps": map[string]models.Step{
			// claim_result deliberately doubles as a SUBSTEP NAME and an
			// OUTPUT FIELD (see check_claim's then_step below): a step-ref key
			// must be step-prefixed, never data-suffixed.
			"claim_result": {Action: "claim_work_item", OutputField: "claim_result", NextStep: "check_claim"},
			"check_claim": {
				Action:   "conditional",
				NextStep: "spawn_handler",
				Config: map[string]interface{}{
					// expression: excluded by the shape gate (spaces/operators)
					"condition": "claim_result.claimed == true",
					// step-ref key colliding with an output field: step-prefixed only
					"then_step": "claim_result",
					"else_step": "done",
				},
			},
			"spawn_handler": {Action: "spawn_agent", OutputField: "handler_spawned", NextStep: "call_handler"},
			"call_handler": {
				Action:      "call_agent",
				OutputField: "handler_result",
				NextStep:    "deploy",
				Config: map[string]interface{}{
					"input_mapping": map[string]interface{}{
						// sibling-output reference inside input_mapping: suffixed (existing behaviour)
						"previous": "claim_result.work_item_id",
						// loop-variable reference: untouched
						"spec": "current_item.spec",
					},
				},
			},
			"deploy": {Action: "git_commit", OutputField: "page_deployed", NextStep: "mark_complete"},
			"mark_complete": {
				Action:      "complete_work_item",
				OutputField: "item_completed",
				NextStep:    "done",
				Config: map[string]interface{}{
					// THE bug: a previously-unlisted key referencing a sibling output
					"result": "handler_result",
					// the strict spelling (RFC_029 §9 D3) rides the same pass
					"result!": "handler_result",
					// dotted sibling-output reference on an unlisted key
					"commit_sha": "page_deployed.commit_sha",
					// loop-variable reference: not a sibling output, untouched
					"work_item_id": "current_item.id",
					// prose that NAMES an output field: excluded by the shape gate
					"note": "handler_result is the reply",
					// legacy allow-listed key: still works
					"content_from": "handler_result.response",
					// Strategy-1 field NAMES in an array: never rewritten
					"input_fields": []interface{}{"handler_result", "claim_result"},
				},
			},
			"done": {Action: "loop_complete"},
		},
		"substep_order": []string{"claim_result", "check_claim", "spawn_handler", "call_handler", "deploy", "mark_complete", "done"},
	}

	if err := s.handleLoopExpansion(state, loopResult, zap.NewNop()); err != nil {
		t.Fatalf("handleLoopExpansion: %v", err)
	}

	for _, iter := range []int{0, 1} {
		suffix := map[int]string{0: "_0", 1: "_1"}[iter]

		mc := state.WorkflowPlan.Steps[iterName("process_item", iter, "mark_complete")]
		if mc.Action == "" {
			t.Fatalf("iter %d: expanded mark_complete step missing", iter)
		}
		cfg := mc.Config

		if got, want := cfg["result"], "handler_result"+suffix; got != want {
			t.Errorf("iter %d: result = %v, want %q", iter, got, want)
		}
		if got, want := cfg["result!"], "handler_result"+suffix; got != want {
			t.Errorf("iter %d: result! = %v, want %q", iter, got, want)
		}
		if got, want := cfg["commit_sha"], "page_deployed"+suffix+".commit_sha"; got != want {
			t.Errorf("iter %d: commit_sha = %v, want %q", iter, got, want)
		}
		if got, want := cfg["content_from"], "handler_result"+suffix+".response"; got != want {
			t.Errorf("iter %d: content_from = %v, want %q", iter, got, want)
		}
		if got, want := cfg["work_item_id"], "current_item.id"; got != want {
			t.Errorf("iter %d: work_item_id = %v, want %q (loop var, untouched)", iter, got, want)
		}
		if got, want := cfg["note"], "handler_result is the reply"; got != want {
			t.Errorf("iter %d: note = %v, want %q (prose, untouched)", iter, got, want)
		}
		if arr, ok := cfg["input_fields"].([]interface{}); !ok || len(arr) != 2 ||
			arr[0] != "handler_result" || arr[1] != "claim_result" {
			t.Errorf("iter %d: input_fields = %v, want field NAMES untouched", iter, cfg["input_fields"])
		}

		cc := state.WorkflowPlan.Steps[iterName("process_item", iter, "check_claim")]
		if got, want := cc.Config["condition"], "claim_result.claimed == true"; got != want {
			t.Errorf("iter %d: condition = %v, want %q (expression, untouched)", iter, got, want)
		}
		if got, want := cc.Config["then_step"], iterName("process_item", iter, "claim_result"); got != want {
			t.Errorf("iter %d: then_step = %v, want %q (step-prefixed, never data-suffixed)", iter, got, want)
		}

		ch := state.WorkflowPlan.Steps[iterName("process_item", iter, "call_handler")]
		im, _ := ch.Config["input_mapping"].(map[string]interface{})
		if im == nil {
			t.Fatalf("iter %d: call_handler input_mapping missing", iter)
		}
		if got, want := im["previous"], "claim_result"+suffix+".work_item_id"; got != want {
			t.Errorf("iter %d: input_mapping.previous = %v, want %q", iter, got, want)
		}
		if got, want := im["spec"], "current_item.spec"; got != want {
			t.Errorf("iter %d: input_mapping.spec = %v, want %q (loop var, untouched)", iter, got, want)
		}
	}
}

// TestReferenceShapedConfigValue pins the shape gate itself.
func TestReferenceShapedConfigValue(t *testing.T) {
	shaped := []string{"handler_result", "page_deployed.commit_sha", "a.b.c", "_x", "x9._y2"}
	for _, s := range shaped {
		if !referenceShapedConfigValue.MatchString(s) {
			t.Errorf("%q should be reference-shaped", s)
		}
	}
	unshaped := []string{
		"", "claim_result.claimed == true", "handler_result is the reply",
		"a b", "items[0].x", "9lives", ".leading", "trailing.", "a..b",
		"reviewed_content.approved == true OR reviewed_content.ok == true",
	}
	for _, s := range unshaped {
		if referenceShapedConfigValue.MatchString(s) {
			t.Errorf("%q should NOT be reference-shaped", s)
		}
	}
}
