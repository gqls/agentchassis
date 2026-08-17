package actions

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// bugs_open/289. A sub-workflow's terminal substep is declared
// `action: loop_complete`, which is also the action of the loop's OWN end-step.
// Injected once per iteration, it used to run the whole-loop aggregator from
// inside every lap and nest each earlier iteration's aggregate into its own, so
// collected_data doubled per iteration (2^N). tool-auditor reached 22 MB and
// completed 1 run in 63.
//
// These tests are written so that they can FAIL: each guard test has a control
// that removes only the guard's own signal from otherwise identical input and
// asserts the OLD behaviour returns. A test that only checks the fixed path
// would pass just as happily against a function that never aggregates at all.

const (
	priorAggregateMarker = "MARKER_FROM_ITERATION_0_AGGREGATE"
	freshOutputMarker    = "MARKER_FROM_ITERATION_1_OWN_OUTPUT"
)

// collectedDataWithPriorAggregate is the shape that caused the blow-up: the
// previous iteration's terminal substep result is sitting in CollectedData under
// the runtime-generated key, and it is large and self-similar.
func collectedDataWithPriorAggregate() map[string]interface{} {
	return map[string]interface{}{
		"loop_metadata": map[string]interface{}{
			"loop_name":        "create_items_loop",
			"total_iterations": float64(10),
			"first_substep":    "check_target_class",
		},
		// Iteration 0's terminal aggregate — the thing that must not be copied.
		"create_items_loop_iter_0_done": map[string]interface{}{
			"iterations": 10,
			"count":      10,
			"results": []interface{}{
				map[string]interface{}{
					"iteration":    0,
					"name":         "item_0",
					"payload":      priorAggregateMarker,
					"done":         map[string]interface{}{"nested": priorAggregateMarker},
					"item_created": map[string]interface{}{"id": "abc"},
				},
			},
		},
		"create_items_loop_iter_1_check_target_class": map[string]interface{}{
			"routed": freshOutputMarker,
		},
		"item_created_0": map[string]interface{}{"id": "abc"},
		"item_created_1": map[string]interface{}{"id": "def", "note": freshOutputMarker},
	}
}

func runLoopComplete(t *testing.T, stepConfig map[string]interface{}, collected map[string]interface{}) map[string]interface{} {
	t.Helper()
	out, err := LoopCompleteAction(context.Background(), ActionParams{
		Context:          context.Background(),
		ExecutionContext: &types.ExecutionContext{StepName: "create_items_loop_iter_1_done"},
		StepConfig:       models.Step{Action: "loop_complete", Config: stepConfig},
		CollectedData:    collected,
		Logger:           zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("LoopCompleteAction returned an error: %v", err)
	}
	result, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", out)
	}
	return result
}

// serialise is how the real damage was measured — the defect was a SIZE defect,
// so the assertion is made against the serialised bytes, not the map shape.
func serialise(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// The guard: an injected per-iteration terminal must NOT re-aggregate the loop.
func TestLoopCompleteIterationTerminalDoesNotAggregate(t *testing.T) {
	result := runLoopComplete(t, map[string]interface{}{
		"loop_name":               "create_items_loop",
		"loop_iteration":          1,
		"loop_iteration_terminal": true,
	}, collectedDataWithPriorAggregate())

	if got := result["status"]; got != "iteration_complete" {
		t.Errorf("status = %v, want iteration_complete", got)
	}
	if got := result["iteration"]; got != 1 {
		t.Errorf("iteration = %v, want 1", got)
	}
	if _, present := result["results"]; present {
		t.Error("a per-iteration terminal must not carry a `results` aggregate")
	}

	// The whole point: the previous iteration's aggregate must not be inside it.
	if body := serialise(t, result); strings.Contains(body, priorAggregateMarker) {
		t.Errorf("iteration terminal copied the PREVIOUS iteration's aggregate — this is the 2^N blow-up of bugs_open/289; result was: %s", body)
	}
}

// The control that makes the test above meaningful: strip ONLY the two signals
// that mark this step as a per-iteration terminal, leave the input otherwise
// identical, and the old swallowing behaviour must come back. Without this, the
// guard test would also pass against a LoopCompleteAction that never aggregates.
func TestLoopCompleteEndStepStillAggregatesAndWouldHaveSwallowedIt(t *testing.T) {
	result := runLoopComplete(t, map[string]interface{}{
		"loop_name":        "create_items_loop",
		"total_iterations": 2,
		// No loop_iteration, no loop_iteration_terminal: this is the loop's own
		// end-step, which is SUPPOSED to aggregate.
	}, collectedDataWithPriorAggregate())

	if _, present := result["results"]; !present {
		t.Fatal("the loop's own end-step must still aggregate — it lost its `results`")
	}
	if got := result["iterations"]; got != 2 {
		t.Errorf("iterations = %v, want 2", got)
	}

	// Proves the input really does contain something swallowable, so the guard
	// test above is testing the guard and not an inert fixture.
	if body := serialise(t, result); !strings.Contains(body, priorAggregateMarker) {
		t.Error("control failed: the end-step did NOT pick up the prior aggregate, so the guard test proves nothing about the guard")
	}
}

// The fallback arm, which is what rescues loops already expanded and persisted
// before this fix shipped: those plans carry `loop_iteration` but no explicit
// flag, and must still be treated as per-iteration terminals.
func TestLoopCompleteIterationTerminalDetectedFromIterationAlone(t *testing.T) {
	result := runLoopComplete(t, map[string]interface{}{
		"loop_name":      "create_items_loop",
		"loop_iteration": float64(3), // float64: a JSON round-trip through the persisted plan
	}, collectedDataWithPriorAggregate())

	if got := result["status"]; got != "iteration_complete" {
		t.Errorf("status = %v, want iteration_complete", got)
	}
	if got := result["iteration"]; got != 3 {
		t.Errorf("iteration = %v, want 3 (float64 from a persisted plan must be read as an int)", got)
	}
	if body := serialise(t, result); strings.Contains(body, priorAggregateMarker) {
		t.Error("an in-flight plan's iteration terminal still aggregated")
	}
}

// A nested loop's own end-step is built by handleLoopExpansion rather than
// injected as a substep, so it carries total_iterations and no iteration index.
// It must keep aggregating even while the orchestration sits inside an OUTER
// loop's iteration — which is exactly why the flag is set explicitly at
// injection instead of being inferred from the step's name or position.
func TestLoopCompleteNestedInnerEndStepStillAggregates(t *testing.T) {
	result := runLoopComplete(t, map[string]interface{}{
		"loop_name":             "inner_loop",
		"total_iterations":      1,
		"substep_output_fields": []interface{}{"item_created"},
	}, map[string]interface{}{
		"item_created_0": map[string]interface{}{"id": "inner"},
	})

	if _, present := result["results"]; !present {
		t.Fatal("a nested loop's own end-step must still aggregate")
	}
	if got := result["count"]; got != 1 {
		t.Errorf("count = %v, want 1", got)
	}
}

// isLoopIterationTerminal is the whole discriminator, so it gets its own table —
// including the two shapes that must NOT trip it.
func TestIsLoopIterationTerminal(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]interface{}
		want   bool
	}{
		{"nil config", nil, false},
		{"explicit flag", map[string]interface{}{"loop_iteration_terminal": true}, true},
		{"explicit flag false", map[string]interface{}{"loop_iteration_terminal": false}, false},
		{"iteration index only (in-flight plan)", map[string]interface{}{"loop_iteration": 0}, true},
		{"iteration index zero is still an iteration", map[string]interface{}{"loop_iteration": float64(0)}, true},
		{"loop end-step", map[string]interface{}{"loop_name": "l", "total_iterations": 10}, false},
		{"loop end-step with output fields", map[string]interface{}{"substep_output_fields": []interface{}{"x"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isLoopIterationTerminal(tc.config); got != tc.want {
				t.Errorf("isLoopIterationTerminal(%v) = %v, want %v", tc.config, got, tc.want)
			}
		})
	}
}
