// FILE: platform/orchestration/substep_continue_on_error_test.go
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// expandLoopWithSubsteps runs one loop expansion over two items and returns the
// state holding the injected plan. The loop is always named "work_loop" and
// always runs two iterations, so a test can assert that a resolution is made per
// substep per iteration rather than once for the whole loop.
func expandLoopWithSubsteps(t *testing.T, loopContinueOnError interface{}, substeps map[string]models.Step, order []string) *OrchestrationState {
	t.Helper()

	s := &SagaCoordinator{logger: zap.NewNop()}

	state := &OrchestrationState{
		CurrentStep:   "work_loop",
		CollectedData: map[string]interface{}{},
		WorkflowPlan: models.WorkflowPlan{
			StartStep: "work_loop",
			Steps: map[string]models.Step{
				"work_loop": {Action: "loop"},
				"complete":  {Action: "complete"},
			},
		},
	}

	loopResult := map[string]interface{}{
		"loop_name":    "work_loop",
		"loop_var":     "item",
		"next_step":    "complete",
		"output_field": "results",
		"items": []interface{}{
			map[string]interface{}{"name": "first"},
			map[string]interface{}{"name": "second"},
		},
		"substeps":      substeps,
		"substep_order": order,
	}
	// Mirrors LoopAction, which only sets the key when the loop declares it.
	if loopContinueOnError != nil {
		loopResult["continue_on_error"] = loopContinueOnError
	}

	if err := s.handleLoopExpansion(state, loopResult, zap.NewNop()); err != nil {
		t.Fatalf("handleLoopExpansion: %v", err)
	}
	return state
}

// injectedContinueOnError reads the resolved flag off one injected iteration step.
func injectedContinueOnError(t *testing.T, state *OrchestrationState, iterIdx int, substep string) bool {
	t.Helper()

	stepName := iterName("work_loop", iterIdx, substep)
	step, ok := state.WorkflowPlan.Steps[stepName]
	if !ok {
		t.Fatalf("injected step %q not found in expanded plan", stepName)
	}
	value, ok := step.Config["continue_on_error"].(bool)
	if !ok {
		t.Fatalf("step %q: continue_on_error = %#v, want a bool", stepName, step.Config["continue_on_error"])
	}
	return value
}

// wouldSkipIteration asks the read side — the single decision point all three
// coordinator call sites go through — what it makes of the injected step.
func wouldSkipIteration(state *OrchestrationState, iterIdx int, substep string) bool {
	state.CurrentStep = iterName("work_loop", iterIdx, substep)
	return shouldContinueLoopOnError(state, zap.NewNop())
}

// TestSubstepContinueOnErrorTolerantSubstepInStrictLoop covers the first direction:
// the loop itself is strict, and one substep opts into tolerance.
func TestSubstepContinueOnErrorTolerantSubstepInStrictLoop(t *testing.T) {
	state := expandLoopWithSubsteps(t, nil,
		map[string]models.Step{
			"best_effort": {
				Action:   "call_agent",
				NextStep: "record",
				Config:   map[string]interface{}{"continue_on_error": true},
			},
			"record": {Action: "update_work_item_status"},
		},
		[]string{"best_effort", "record"},
	)

	for _, iter := range []int{0, 1} {
		if got := injectedContinueOnError(t, state, iter, "best_effort"); !got {
			t.Errorf("iter %d: best_effort continue_on_error = %v, want true (substep opts in against a strict loop)", iter, got)
		}
		if !wouldSkipIteration(state, iter, "best_effort") {
			t.Errorf("iter %d: shouldContinueLoopOnError = false, want true for a tolerant substep", iter)
		}
	}
}

// TestSubstepContinueOnErrorStrictSubstepInTolerantLoop covers the direction that
// prevents the silent-drop class, and the one a "read the substep only if truthy"
// implementation gets wrong by treating a declared false as no declaration.
func TestSubstepContinueOnErrorStrictSubstepInTolerantLoop(t *testing.T) {
	state := expandLoopWithSubsteps(t, true,
		map[string]models.Step{
			"must_succeed": {
				Action:   "call_agent",
				NextStep: "record",
				Config:   map[string]interface{}{"continue_on_error": false},
			},
			"record": {Action: "update_work_item_status"},
		},
		[]string{"must_succeed", "record"},
	)

	for _, iter := range []int{0, 1} {
		if got := injectedContinueOnError(t, state, iter, "must_succeed"); got {
			t.Errorf("iter %d: must_succeed continue_on_error = %v, want false (substep opts out of a tolerant loop)", iter, got)
		}
		if wouldSkipIteration(state, iter, "must_succeed") {
			t.Errorf("iter %d: shouldContinueLoopOnError = true, want false for a strict substep", iter)
		}
	}
}

// TestSubstepContinueOnErrorNoDeclarationInheritsLoop is the inertness proof: a
// substep that says nothing behaves exactly as it did before the override existed,
// for every loop-level value.
func TestSubstepContinueOnErrorNoDeclarationInheritsLoop(t *testing.T) {
	cases := []struct {
		name string
		loop interface{}
		want bool
	}{
		{name: "loop tolerant", loop: true, want: true},
		{name: "loop strict", loop: false, want: false},
		{name: "loop unset", loop: nil, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := expandLoopWithSubsteps(t, tc.loop,
				map[string]models.Step{
					"silent": {Action: "call_agent", NextStep: "record"},
					"record": {Action: "update_work_item_status"},
				},
				[]string{"silent", "record"},
			)

			for _, iter := range []int{0, 1} {
				if got := injectedContinueOnError(t, state, iter, "silent"); got != tc.want {
					t.Errorf("iter %d: silent continue_on_error = %v, want %v (inherited)", iter, got, tc.want)
				}
				if got := wouldSkipIteration(state, iter, "silent"); got != tc.want {
					t.Errorf("iter %d: shouldContinueLoopOnError = %v, want %v", iter, got, tc.want)
				}
			}
		})
	}
}

// TestSubstepContinueOnErrorMalformedDeclarationFallsBack pins the loud-but-safe
// handling of a declaration that is present with the wrong type: the loop's value
// stands, and expansion does not panic.
func TestSubstepContinueOnErrorMalformedDeclarationFallsBack(t *testing.T) {
	cases := []struct {
		name string
		loop interface{}
		want bool
	}{
		{name: "loop tolerant", loop: true, want: true},
		{name: "loop unset", loop: nil, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := expandLoopWithSubsteps(t, tc.loop,
				map[string]models.Step{
					"mistyped": {
						Action:   "call_agent",
						NextStep: "record",
						Config:   map[string]interface{}{"continue_on_error": "true"},
					},
					"record": {Action: "update_work_item_status"},
				},
				[]string{"mistyped", "record"},
			)

			for _, iter := range []int{0, 1} {
				if got := injectedContinueOnError(t, state, iter, "mistyped"); got != tc.want {
					t.Errorf("iter %d: mistyped continue_on_error = %v, want %v (loop value)", iter, got, tc.want)
				}
			}
		})
	}
}

// TestSubstepContinueOnErrorResolvesPerSubstepNotPerLoop is the granularity claim
// itself: within one loop, a declaring substep and a silent sibling must end up
// with different values, on every iteration.
func TestSubstepContinueOnErrorResolvesPerSubstepNotPerLoop(t *testing.T) {
	t.Run("tolerant declarer beside a strict sibling", func(t *testing.T) {
		state := expandLoopWithSubsteps(t, false,
			map[string]models.Step{
				"best_effort": {
					Action:   "call_agent",
					NextStep: "must_succeed",
					Config:   map[string]interface{}{"continue_on_error": true},
				},
				"must_succeed": {Action: "update_work_item_status"},
			},
			[]string{"best_effort", "must_succeed"},
		)

		for _, iter := range []int{0, 1} {
			if got := injectedContinueOnError(t, state, iter, "best_effort"); !got {
				t.Errorf("iter %d: declaring substep continue_on_error = %v, want true", iter, got)
			}
			if got := injectedContinueOnError(t, state, iter, "must_succeed"); got {
				t.Errorf("iter %d: silent sibling continue_on_error = %v, want false (the loop's value)", iter, got)
			}
			if !wouldSkipIteration(state, iter, "best_effort") {
				t.Errorf("iter %d: shouldContinueLoopOnError = false for the declaring substep, want true", iter)
			}
			if wouldSkipIteration(state, iter, "must_succeed") {
				t.Errorf("iter %d: shouldContinueLoopOnError = true for the silent sibling, want false", iter)
			}
		}
	})

	t.Run("strict declarer beside a tolerant sibling", func(t *testing.T) {
		state := expandLoopWithSubsteps(t, true,
			map[string]models.Step{
				"must_succeed": {
					Action:   "call_agent",
					NextStep: "best_effort",
					Config:   map[string]interface{}{"continue_on_error": false},
				},
				"best_effort": {Action: "update_work_item_status"},
			},
			[]string{"must_succeed", "best_effort"},
		)

		for _, iter := range []int{0, 1} {
			if got := injectedContinueOnError(t, state, iter, "must_succeed"); got {
				t.Errorf("iter %d: declaring substep continue_on_error = %v, want false", iter, got)
			}
			if got := injectedContinueOnError(t, state, iter, "best_effort"); !got {
				t.Errorf("iter %d: silent sibling continue_on_error = %v, want true (the loop's value)", iter, got)
			}
			if wouldSkipIteration(state, iter, "must_succeed") {
				t.Errorf("iter %d: shouldContinueLoopOnError = true for the declaring substep, want false", iter)
			}
			if !wouldSkipIteration(state, iter, "best_effort") {
				t.Errorf("iter %d: shouldContinueLoopOnError = false for the silent sibling, want true", iter)
			}
		}
	})
}
