// FILE: platform/validation/subworkflow_test.go
//
// bugs_open/144 — steps nested inside a loop's sub-workflow reached production
// validated by nothing. These tests pin what the recursion does and, just as
// importantly, what it must NOT do: a reference out of a sub-workflow is legitimate,
// and a naive "validate each sub-workflow as a workflow in its own right" would
// reject workflows that run today.
//
// Every hard-error case below is paired with a positive control that validates clean,
// so a test that stops discriminating fails rather than passing vacuously.
package validation

import (
	"os"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestMain registers a local-action checker so "is this action local?" is decidable
// in this package's tests. Without one, actioncheck.IsLocalAction returns false for
// everything and every fixture would need a topic — which no live nested step carries
// (85 of 85, measured 2026-07-29), so the fixtures would stop resembling the thing
// under test. Prefix rule rather than a list: explicit at every call site.
func TestMain(m *testing.M) {
	actioncheck.RegisterLocalActionChecker(func(action string) bool {
		return strings.HasPrefix(action, "local_")
	})
	os.Exit(m.Run())
}

// planWithLoop wraps a sub-workflow config in the shape live definitions use: a
// top-level loop step whose config carries sub_workflow.steps.
func planWithLoop(subWorkflow map[string]interface{}) models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "the_loop",
		Steps: map[string]models.Step{
			"the_loop": {
				Action: "loop",
				Topic:  "system.test.requests",
				Config: map[string]interface{}{
					"items_field":   "input_data.pages",
					"item_variable": "page",
					"sub_workflow":  subWorkflow,
				},
			},
			"after_the_loop": {Action: "local_complete_workflow"},
		},
	}
}

func steps(m map[string]interface{}) map[string]interface{} { return m }

func validateWithLogs(t *testing.T, plan models.WorkflowPlan) (error, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zapcore.WarnLevel)
	return NewWorkflowValidator(zap.New(core)).ValidateWorkflow(plan), logs
}

// TestNestedStepWithNoActionIsRejected: the headline case. A nested step with no
// action ran, and failed at execution, having been checked by nothing.
func TestNestedStepWithNoActionIsRejected(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{
		"start_step": "write",
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{"description": "no action here"},
		}),
	})

	err, _ := validateWithLogs(t, plan)
	if err == nil {
		t.Fatal("a nested step with no action must be rejected")
	}
	// The path, not just the name: a definition can hold three steps called 'write'
	// at different depths, and "step 'write'" is unactionable in that case.
	if !strings.Contains(err.Error(), "steps.the_loop.sub_workflow.write") {
		t.Fatalf("error must locate the step by path, got: %v", err)
	}

	// Positive control: the same shape with an action validates clean.
	plan = planWithLoop(map[string]interface{}{
		"start_step": "write",
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{"action": "local_write_page_content"},
		}),
	})
	if err, _ := validateWithLogs(t, plan); err != nil {
		t.Fatalf("a well-formed sub-workflow must validate: %v", err)
	}
}

// TestNestedRemoteActionNeedsTopic mirrors the top-level rule. Live nested steps use
// local actions exclusively (20 distinct actions, all IsLocal, measured 2026-07-29),
// so this fires on the first remote one anybody nests.
func TestNestedRemoteActionNeedsTopic(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"call_out": map[string]interface{}{"action": "remote_thing"},
		}),
	})
	err, _ := validateWithLogs(t, plan)
	if err == nil || !strings.Contains(err.Error(), "requires a topic") {
		t.Fatalf("a nested remote action with no topic must be rejected, got: %v", err)
	}

	plan = planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"call_out": map[string]interface{}{"action": "remote_thing", "topic": "system.x.requests"},
		}),
	})
	if err, _ := validateWithLogs(t, plan); err != nil {
		t.Fatalf("a nested remote action WITH a topic must validate: %v", err)
	}
}

// TestReferenceOutOfSubWorkflowIsNotAnError is the anti-regression test for the
// mistake this fix nearly shipped. Loop expansion prefixes a next_step only when it
// names a sibling substep; anything else passes through untouched and resolves
// against the enclosing plan, or against a step the expander injects at runtime.
// Making that a hard error would reject workflows that run today.
func TestReferenceOutOfSubWorkflowIsNotAnError(t *testing.T) {
	for _, target := range []string{
		"after_the_loop",           // a top-level step of the enclosing plan
		"the_loop_complete",        // injected by the expander; in no definition
		"some_other_loop_complete", // ditto, in a sibling branch
	} {
		plan := planWithLoop(map[string]interface{}{
			"steps": steps(map[string]interface{}{
				"write": map[string]interface{}{"action": "local_write", "next_step": target},
			}),
		})
		err, logs := validateWithLogs(t, plan)
		if err != nil {
			t.Fatalf("next_step %q out of the sub-workflow must not fail validation: %v", target, err)
		}
		if n := logs.FilterMessageSnippet("names no step in scope").Len(); n != 0 {
			t.Fatalf("next_step %q must not be reported as dangling, got %d warnings", target, n)
		}
	}

	// Discriminating half: a target that resolves nowhere and is not an injected
	// name still gets reported — as a warning, because validation cannot see the
	// expanded plan and a false rejection is the worse failure.
	plan := planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{"action": "local_write", "next_step": "no_such_step"},
		}),
	})
	err, logs := validateWithLogs(t, plan)
	if err != nil {
		t.Fatalf("an unresolved reference must warn, not fail: %v", err)
	}
	if n := logs.FilterMessageSnippet("names no step in scope").Len(); n != 1 {
		t.Fatalf("expected 1 dangling-reference warning, got %d", n)
	}
}

// TestDroppedNestedFieldsAreReported: the runtime decoder reads seven fields and
// drops the rest, so a nested `dependencies` describes ordering that never happens.
// Reported, not rejected — the step itself is valid, its extra fields are inert.
func TestDroppedNestedFieldsAreReported(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{
				"action":       "local_write",
				"dependencies": []interface{}{"nothing"},
				"timeout":      "30s",
				"sub_tasks":    []interface{}{},
			},
		}),
	})

	err, logs := validateWithLogs(t, plan)
	if err != nil {
		t.Fatalf("dropped fields must not fail validation: %v", err)
	}
	entries := logs.FilterMessageSnippet("does not read").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 dropped-field warning, got %d", len(entries))
	}
	got, _ := entries[0].ContextMap()["dropped_fields"].([]interface{})
	want := map[string]bool{"dependencies": true, "timeout": true, "sub_tasks": true}
	if len(got) != len(want) {
		t.Fatalf("expected the three dropped fields, got %v", got)
	}
	for _, f := range got {
		if !want[f.(string)] {
			t.Fatalf("unexpected field reported as dropped: %q (all of %v)", f, got)
		}
	}

	// Positive control: a step carrying only honoured fields reports nothing. Without
	// this the test above would pass just as well if the warning fired on everything.
	plan = planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{
				"action": "local_write", "description": "d", "next_step": "",
				"error_step": "", "output_field": "html", "topic": "",
				"config": map[string]interface{}{},
			},
		}),
	})
	_, logs = validateWithLogs(t, plan)
	if n := logs.FilterMessageSnippet("does not read").Len(); n != 0 {
		t.Fatalf("a step using only honoured fields must report nothing, got %d", n)
	}
}

// TestNestedStartStepMustExist: the loop action tolerates a bad start_step by
// auto-detecting one, so the loop runs in an order nobody chose. That tolerance is
// what made it invisible.
func TestNestedStartStepMustExist(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{
		"start_step": "typo",
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{"action": "local_write"},
		}),
	})
	err, _ := validateWithLogs(t, plan)
	if err == nil || !strings.Contains(err.Error(), "start_step") {
		t.Fatalf("a start_step naming no step must be rejected, got: %v", err)
	}

	// Absent start_step is NORMAL — the loop auto-detects from the next_step chain,
	// and 2 of the 20 live carriers have none. Requiring it would reject them.
	plan = planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"write": map[string]interface{}{"action": "local_write"},
		}),
	})
	if err, _ := validateWithLogs(t, plan); err != nil {
		t.Fatalf("a sub-workflow with no start_step must validate: %v", err)
	}
}

// TestNestedFanOutRejected: sub_tasks is not read by the nested decoder, so a nested
// fan_out reaches the executor with none. The top-level rule cannot catch it —
// it inspects the definition, which looks complete.
func TestNestedFanOutRejected(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"spread": map[string]interface{}{
				"action":    "fan_out",
				"sub_tasks": []interface{}{map[string]interface{}{"step_name": "a", "topic": "t"}},
			},
		}),
	})
	err, _ := validateWithLogs(t, plan)
	if err == nil || !strings.Contains(err.Error(), "fan_out") {
		t.Fatalf("a nested fan_out must be rejected even with sub_tasks present, got: %v", err)
	}
}

// TestNestedStrictConfigKeyIsRejected: the bugs_open/101 rule now applies at both
// depths, from one implementation. 0 of the 66 live nested (action, key) pairs trips
// it, measured before shipping.
func TestNestedStrictConfigKeyIsRejected(t *testing.T) {
	// StrictConfig, not CheckConfig: CheckConfig only opts the action into
	// detection, and StrictConfig is what turns an unrecognised key into a hard
	// error (datahelpers.IsStrictConfigAction).
	datahelpers.RegisterActionInputSpec("local_strict_nested", datahelpers.ActionInputSpec{
		ConfigKeys:   []string{"page_field"},
		StrictConfig: true,
	})

	plan := planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"commit": map[string]interface{}{
				"action": "local_strict_nested",
				"config": map[string]interface{}{"page_field": "x", "commit_sha": "y"},
			},
		}),
	})
	err, _ := validateWithLogs(t, plan)
	if err == nil || !strings.Contains(err.Error(), "commit_sha") {
		t.Fatalf("an unrecognised key on a strict action must be rejected inside a sub-workflow too, got: %v", err)
	}

	plan = planWithLoop(map[string]interface{}{
		"steps": steps(map[string]interface{}{
			"commit": map[string]interface{}{
				"action": "local_strict_nested",
				"config": map[string]interface{}{"page_field": "x"},
			},
		}),
	})
	if err, _ := validateWithLogs(t, plan); err != nil {
		t.Fatalf("a declared key must validate: %v", err)
	}
}

// TestSubstepsTakesPrecedenceOverSubWorkflow mirrors loop_actions.go:74-88. Both
// shapes are live (18 sub_workflow carriers, 2 substeps), and a step carrying both
// runs only one of them.
func TestSubstepsTakesPrecedenceOverSubWorkflow(t *testing.T) {
	plan := models.WorkflowPlan{
		StartStep: "the_loop",
		Steps: map[string]models.Step{
			"the_loop": {
				Action: "local_loop",
				Config: map[string]interface{}{
					"substeps": map[string]interface{}{
						"good": map[string]interface{}{"action": "local_write"},
					},
					// Broken, and ignored at execution — so it must not fail validation.
					"sub_workflow": map[string]interface{}{
						"steps": map[string]interface{}{"bad": map[string]interface{}{}},
					},
				},
			},
		},
	}

	err, logs := validateWithLogs(t, plan)
	if err != nil {
		t.Fatalf("the ignored sub_workflow half must not fail validation: %v", err)
	}
	if n := logs.FilterMessageSnippet("only 'substeps' is executed").Len(); n != 1 {
		t.Fatalf("carrying both shapes must be reported once, got %d", n)
	}
}

// TestEmptySubWorkflowRejected: the loop action fails at execution on this. Failing
// at validation names the definition instead of the run.
func TestEmptySubWorkflowRejected(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{"steps": steps(map[string]interface{}{})})
	err, _ := validateWithLogs(t, plan)
	if err == nil || !strings.Contains(err.Error(), "defines no steps") {
		t.Fatalf("a sub-workflow with no steps must be rejected, got: %v", err)
	}
}

// TestNestedCycleDetected: cycle detection now reaches inside.
func TestNestedCycleDetected(t *testing.T) {
	plan := planWithLoop(map[string]interface{}{
		"start_step": "a",
		"steps": steps(map[string]interface{}{
			"a": map[string]interface{}{"action": "local_write", "next_step": "b"},
			"b": map[string]interface{}{"action": "local_write", "next_step": "a"},
		}),
	})
	err, _ := validateWithLogs(t, plan)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("a cycle inside a sub-workflow must be detected, got: %v", err)
	}
}

// TestSubWorkflowDepthLimit: the backstop is reachable and reports where it stopped.
// Live maximum nesting is 1.
func TestSubWorkflowDepthLimit(t *testing.T) {
	// Build a sub-workflow nested maxSubWorkflowDepth+1 deep.
	inner := map[string]interface{}{
		"steps": map[string]interface{}{"leaf": map[string]interface{}{"action": "local_write"}},
	}
	for i := 0; i <= maxSubWorkflowDepth; i++ {
		inner = map[string]interface{}{
			"steps": map[string]interface{}{
				"nest": map[string]interface{}{
					"action": "local_loop",
					"config": map[string]interface{}{"sub_workflow": inner},
				},
			},
		}
	}
	err, _ := validateWithLogs(t, planWithLoop(inner))
	if err == nil || !strings.Contains(err.Error(), "maximum supported depth") {
		t.Fatalf("nesting past the backstop must be rejected, got: %v", err)
	}

	// And a legal depth still validates, so the constant is a limit rather than a ban.
	legal := map[string]interface{}{
		"steps": map[string]interface{}{
			"nest": map[string]interface{}{
				"action": "local_loop",
				"config": map[string]interface{}{"sub_workflow": map[string]interface{}{
					"steps": map[string]interface{}{"leaf": map[string]interface{}{"action": "local_write"}},
				}},
			},
		},
	}
	if err, _ := validateWithLogs(t, planWithLoop(legal)); err != nil {
		t.Fatalf("two levels of nesting must validate: %v", err)
	}
}

// TestTopLevelValidationUnchanged: the recursion must not have altered what the
// validator already did. A plan with no sub-workflow anywhere behaves as before.
func TestTopLevelValidationUnchanged(t *testing.T) {
	plan := models.WorkflowPlan{
		StartStep: "one",
		Steps: map[string]models.Step{
			"one": {Action: "local_a", NextStep: "two"},
			"two": {Action: "local_b"},
		},
	}
	if err, _ := validateWithLogs(t, plan); err != nil {
		t.Fatalf("an ordinary plan must still validate: %v", err)
	}

	plan.Steps["one"] = models.Step{Action: "local_a", NextStep: "missing"}
	if err, _ := validateWithLogs(t, plan); err == nil {
		t.Fatal("a dangling top-level next_step must still be rejected")
	}
}
