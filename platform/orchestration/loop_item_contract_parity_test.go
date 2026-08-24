// FILE: platform/orchestration/loop_item_contract_parity_test.go
//
// THE CROSS-PACKAGE PIN FOR bugs_closed/283 / RFC_032 step 3.
//
// The occurrence fix in platform/orchestration/actions reads three things that
// loop expansion PUBLISHES and that nothing else in that package produces: the
// injected `loop_item_index` and `loop_name` on a substep's own config, and the
// per-iteration items parked in CollectedData under datahelpers.LoopItemKey.
//
// The council's standing objection to a change like that is coupling: an action
// reading another component's output can rot silently when that component
// changes, because BOTH sides keep compiling and the reader simply finds
// nothing — and "found nothing" is a legitimate state for this reader (it means
// "no loop context", which falls back to occurrence 0). So the drift would show
// up as quietly wrong element ids on multi-instance pages, not as a failure.
//
// This test is the answer to that objection, and it is why the coupling claim in
// component_instance_occurrence.go is enforceable rather than asserted: it runs
// the REAL expander over a real loop and asserts the REAL reader agrees with it.
// It lives in this package (not in actions) because handleLoopExpansion is
// unexported; it imports the actions package to get the reader, which is the
// direction the dependency already runs.
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"go.uber.org/zap"
)

// TestLoopExpansion_placementReaderAgreesWithInjectedState drives the expander
// over a five-section page whose 1st, 3rd and 5th sections share a function, then
// asks the actions-package reader what iteration 4 (the third generic-text-block)
// sees. It must see exactly the four preceding functions, in order.
//
// MUTATIONS THAT KILL IT — all three RUN, not asserted (2026-08-24):
//   - the expander re-introduces a hardcoded key literal instead of calling
//     datahelpers.LoopItemKey (writer and reader genuinely drift; this test is
//     the ONLY one in the repo that fails);
//   - the expander stops injecting `loop_item_index` (LoopKnown goes false and
//     every instance silently reverts to occurrence 0 — again, only this test);
//   - actions.functionOfLoopItem reads the component's display name instead of
//     its function (the decoy above is deliberately not a function name).
//
// ⚠ AND ONE MUTATION THAT DOES **NOT** KILL IT, which is worth more than the
// three that do. Changing datahelpers.LoopItemKey's FORMAT leaves every test
// green. That looks like a hole and is not: since this change single-sourced the
// spelling, the writer and the reader move together, so the format is no longer
// a thing that CAN drift — the helper is a guard sitting in series with this
// test. The drift that remains possible is a caller going around the helper,
// which is mutation 1. Recorded because a mutation that passes usually means a
// guard in series rather than a blind test, and the two are indistinguishable
// until you look.
func TestLoopExpansion_placementReaderAgreesWithInjectedState(t *testing.T) {
	s := &SagaCoordinator{logger: zap.NewNop()}

	// The live shape: sectionPlanItem carries the resolved function at the top
	// level AND inside `component`, whose `name` is a DISPLAY name.
	functions := []string{"generic-text-block", "hero", "generic-text-block", "faq", "generic-text-block"}
	items := make([]interface{}, 0, len(functions))
	for _, fn := range functions {
		items = append(items, map[string]interface{}{
			"name":     fn,
			"function": fn,
			"component": map[string]interface{}{
				"function": fn,
				"name":     "DISPLAY NAME THAT MUST NOT BE READ",
			},
		})
	}

	state := &OrchestrationState{
		CurrentStep:   "process_sections_loop",
		CollectedData: map[string]interface{}{},
		WorkflowPlan: models.WorkflowPlan{
			StartStep: "process_sections_loop",
			Steps: map[string]models.Step{
				"process_sections_loop": {Action: "loop"},
				"compile_page":          {Action: "compile_page_sections"},
			},
		},
	}

	loopResult := map[string]interface{}{
		"loop_name":     "process_sections_loop",
		"loop_var":      "current_section",
		"next_step":     "compile_page",
		"output_field":  "rendered_sections",
		"items":         items,
		"substeps":      map[string]models.Step{"render_section": {Action: "render_component"}},
		"substep_order": []string{"render_section"},
	}

	if err := s.handleLoopExpansion(state, loopResult, zap.NewNop()); err != nil {
		t.Fatalf("handleLoopExpansion: %v", err)
	}

	const iter = 4
	step, ok := state.WorkflowPlan.Steps[iterName("process_sections_loop", iter, "render_section")]
	if !ok {
		t.Fatalf("the expander did not inject an iteration-%d render step", iter)
	}

	// THE PARITY ASSERTION: the real reader, over the real injected config and
	// the real CollectedData the expander just wrote.
	p := actions.PlacementFromLoopStep(step.Config, state.CollectedData)

	if !p.LoopKnown {
		t.Fatal("the reader could not see the loop context the expander just injected — " +
			"loop_item_index/loop_name are the published contract this fix depends on, and " +
			"a miss here reverts every instance to occurrence 0 without failing anything")
	}
	if len(p.PriorFunctions) != iter {
		t.Fatalf("iteration %d must see %d preceding items, saw %d (%v) — the item-key spelling has drifted",
			iter, iter, len(p.PriorFunctions), p.PriorFunctions)
	}
	for i, want := range functions[:iter] {
		if p.PriorFunctions[i] != want {
			t.Fatalf("prior %d: reader saw %q, expander wrote %q", i, p.PriorFunctions[i], want)
		}
	}

	// And the consequence that actually matters: this placement yields the same
	// token the canonical whole-page walk assigns to the same section. If these
	// two ever disagree, a page's ids depend on which path last rendered it —
	// the entire defect class RFC_032 exists to close.
	rc := &actions.RenderContext{}
	actions.DeriveAndBindInstanceToken(t.Context(), nil, rc, functions[iter], p, zap.NewNop())
	got, _ := rc.ContentData[actions.InstanceContentKey].(string)
	want := actions.InstanceTokensForPage(functions)[iter]
	if got != want {
		t.Fatalf("loop-derived token %q != canonical token %q for the same section", got, want)
	}
}
