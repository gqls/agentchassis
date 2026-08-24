// FILE: platform/orchestration/error_route_completion_test.go
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
)

// bugs_open/354. These pin the discriminator that decides whether a completing
// run gets its failure recorded on the row.
//
// Every case below is a REAL population measured in the live table on
// 2026-08-22, not an invented shape — the counts in each test name are what that
// case was worth in a ~2-day retention window. That matters because the naive
// fix (stamp whenever __step_error is present) passes a hand-written
// happy-path test and mislabels 13 real runs, which the bug file warns is
// "worse than the bug".

// stateAt builds a completing state: the plan's terminal step, its config, and
// whatever __step_error the run is carrying.
func stateAt(terminal string, terminalConfig map[string]interface{}, stepError interface{}) *OrchestrationState {
	collected := map[string]interface{}{}
	if stepError != nil {
		collected["__step_error"] = stepError
	}
	return &OrchestrationState{
		CurrentStep:   terminal,
		CollectedData: collected,
		WorkflowPlan: models.WorkflowPlan{
			StartStep: "start",
			Steps: map[string]models.Step{
				terminal: {Action: "complete_workflow", Config: terminalConfig},
			},
		},
	}
}

func routedError(failedStep, message string) map[string]interface{} {
	return map[string]interface{}{"failed_step": failedStep, "message": message}
}

// The defect itself: tool-generator's save_tool refusal landing on complete_error.
// 12 rows. Also pins the message shaping shared with UpdateWorkItemStatusAction.
func TestErrorRouteTermination_DeclaredTerminalWithMarker_IsRecorded(t *testing.T) {
	st := stateAt("complete_error",
		map[string]interface{}{"outcome": "error"},
		routedError("save_tool", "tool birth refused (instance scope)"))

	msg, ok := errorRouteTermination(st)
	if !ok {
		t.Fatal("a declared error terminal carrying a routed failure must be recorded")
	}
	want := "step save_tool failed: tool birth refused (instance scope)"
	if msg != want {
		t.Errorf("message = %q, want %q", msg, want)
	}
}

// THE ONE THAT KILLS THE NAIVE FIX. build-dispatch-loop routes an item failure to
// mark_failed, marks the item, continues the loop and finishes its work at the
// ORDINARY terminal. __step_error is still set — it is never cleared — but the
// run genuinely succeeded. 13 rows, 27% of everything carrying the marker.
func TestErrorRouteTermination_RecoveredRunAtOrdinaryTerminal_IsUntouched(t *testing.T) {
	st := stateAt("complete",
		map[string]interface{}{}, // the ordinary terminal declares nothing
		routedError("process_item_iter_0_call_handler", "handler failed"))

	if msg, ok := errorRouteTermination(st); ok {
		t.Errorf("a recovered run must not be recorded as failed; got %q", msg)
	}
}

// The 2 rows in the bug's own §2(b) control: page-build-handler's
// check_page_found routes a not-found page to complete_error via a CONDITIONAL
// else_step. No step failed, so there is no __step_error. That is a SKIP, not a
// failure (bugs_closed/299), and requiring the marker excludes it for free.
func TestErrorRouteTermination_DeclaredTerminalReachedBySkip_IsUntouched(t *testing.T) {
	st := stateAt("complete_error",
		map[string]interface{}{"outcome": "error"},
		nil) // reached without any routed failure

	if msg, ok := errorRouteTermination(st); ok {
		t.Errorf("a skip into an error terminal must not be recorded as a failure; got %q", msg)
	}
}

// Rollout safety. The config migration is live immediately; the binary is not.
// A run spawned before the terminal was declared carries no outcome key and must
// behave exactly as it does today.
func TestErrorRouteTermination_UndeclaredTerminal_IsUntouched(t *testing.T) {
	st := stateAt("complete_error",
		map[string]interface{}{}, // not yet seeded
		routedError("save_tool", "tool birth refused"))

	if _, ok := errorRouteTermination(st); ok {
		t.Error("an undeclared terminal must be inert until its declaration is seeded")
	}
}

// Read the declaration STRICTLY: only "error" activates, so a typo or a future
// vocabulary fails toward today's behaviour rather than toward stamping.
func TestErrorRouteTermination_OutcomeIsReadStrictly(t *testing.T) {
	for _, outcome := range []interface{}{"failure", "Error", "ERROR", "", true, 1, nil} {
		cfg := map[string]interface{}{"outcome": outcome}
		st := stateAt("complete_error", cfg, routedError("save_tool", "boom"))
		if _, ok := errorRouteTermination(st); ok {
			t.Errorf("outcome %#v must not activate the record", outcome)
		}
	}
}

// __step_error is written by one producer but has been observed malformed; the
// shared accessor is hardened and this pins that we reuse it rather than
// re-implementing the parse. Companion to
// update_work_item_status_owned_refusal_test.go.
func TestErrorRouteTermination_MalformedMarker_IsInertNotFatal(t *testing.T) {
	for name, marker := range map[string]interface{}{
		"bare string": "something wrote a bare string here",
		"number":      500,
		"array":       []interface{}{"a", "b"},
		"empty map":   map[string]interface{}{},
		"message is a number": map[string]interface{}{
			"failed_step": "save_tool", "message": 500},
	} {
		t.Run(name, func(t *testing.T) {
			st := stateAt("complete_error", map[string]interface{}{"outcome": "error"}, marker)
			if _, ok := errorRouteTermination(st); ok {
				t.Errorf("a malformed __step_error must be inert, not recorded")
			}
		})
	}
}

// The awaited-request timeout shape ("Request <id> timed out after N retries")
// carries no "step " prefix and so gains one; an action error already has it and
// must not get a second.
func TestErrorRouteTermination_PrefixIsNotDoubled(t *testing.T) {
	already := "step save_tool failed: tool birth refused"
	st := stateAt("complete_error", map[string]interface{}{"outcome": "error"},
		routedError("save_tool", already))

	msg, ok := errorRouteTermination(st)
	if !ok {
		t.Fatal("expected the record")
	}
	if msg != already {
		t.Errorf("prefix was doubled: %q", msg)
	}
}

// A marker with no failed_step (possible for the timeout shape) still records —
// the message is the point; the prefix is a convenience.
func TestErrorRouteTermination_MarkerWithoutFailedStep_StillRecords(t *testing.T) {
	st := stateAt("complete_error", map[string]interface{}{"outcome": "error"},
		map[string]interface{}{"message": "Request abc timed out after 3 retries"})

	msg, ok := errorRouteTermination(st)
	if !ok {
		t.Fatal("expected the record")
	}
	if msg != "Request abc timed out after 3 retries" {
		t.Errorf("message = %q", msg)
	}
}

// A sparse plan must not panic or record. Loop substeps are injected under
// iteration-prefixed names, but nothing guarantees every CurrentStep is present.
func TestErrorRouteTermination_StepAbsentFromPlan_IsInert(t *testing.T) {
	st := stateAt("complete_error", map[string]interface{}{"outcome": "error"},
		routedError("save_tool", "boom"))
	st.CurrentStep = "a_step_not_in_the_plan"

	if _, ok := errorRouteTermination(st); ok {
		t.Error("a step absent from the plan declares nothing and must be inert")
	}
}

// A clean run — the control that makes the whole thing meaningful. ~3,020 rows.
func TestErrorRouteTermination_CleanRun_IsUntouched(t *testing.T) {
	st := stateAt("complete", map[string]interface{}{}, nil)
	if _, ok := errorRouteTermination(st); ok {
		t.Error("a clean run must never be recorded as failed")
	}
}

func TestErrorRouteTermination_NilState_IsInert(t *testing.T) {
	if _, ok := errorRouteTermination(nil); ok {
		t.Error("nil state must be inert")
	}
}
