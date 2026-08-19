// FILE: platform/orchestration/actions/conditional_fail_on_non_numeric_test.go
//
// bugs_open/313. A conditional whose numeric comparison cannot evaluate —
// `candidate_pages.count > 0` against a bare array, which has no `count` key —
// returned `false, nil` and silently routed to else_step on every run for four
// months, so the internal linker's only LLM step never executed and the runs
// still read `complete`.
//
// These cases pin both halves: the lenient default is frozen deliberately (it
// is the historical contract of 145 live conditional steps and must not change
// under anyone), and the opt-in `fail_on_non_numeric` is proven to fail the
// step, to name the field, and to leave the null-probe operators (==, !=,
// truthy) untouched — for those, nil is a legitimate operand.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func conditionalParams(config map[string]interface{}, collected map[string]interface{}) ActionParams {
	return ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{StepName: "check_candidates"},
		StepConfig:       models.Step{Config: config},
		CollectedData:    collected,
	}
}

// The 313 shape verbatim: the producer declared output_format array, so the
// field is a bare slice and `.count` resolves to nothing.
func arrayShapedCandidates() map[string]interface{} {
	return map[string]interface{}{
		"candidate_pages": []interface{}{
			map[string]interface{}{"name": "services", "url": "/services"},
			map[string]interface{}{"name": "about-us", "url": "/about-us"},
		},
	}
}

// The post-488 shape: output_format object, count present.
func objectShapedCandidates(n float64) map[string]interface{} {
	return map[string]interface{}{
		"candidate_pages": map[string]interface{}{
			"rows":  []interface{}{},
			"count": n,
		},
	}
}

// Lenient default, frozen: absent flag preserves the historical silent-false
// byte for byte. Kills the mutation that makes strict the default.
func TestConditionalLenientDefaultStillRoutesToElse(t *testing.T) {
	result, err := ConditionalBranchAction(context.Background(), conditionalParams(
		map[string]interface{}{
			"condition": "candidate_pages.count > 0",
			"then_step": "load_specs",
			"else_step": "complete_no_candidates",
		},
		arrayShapedCandidates(),
	))
	if err != nil {
		t.Fatalf("lenient mode must not error on an unresolvable numeric comparison, got: %v", err)
	}
	m := result.(map[string]interface{})
	if m["next_step_override"] != "complete_no_candidates" {
		t.Fatalf("lenient mode must take else_step, got %v", m["next_step_override"])
	}
}

// The opt-in: the same unevaluable comparison FAILS THE STEP, naming the field.
func TestConditionalFailOnNonNumericFailsTheStep(t *testing.T) {
	_, err := ConditionalBranchAction(context.Background(), conditionalParams(
		map[string]interface{}{
			"condition":           "candidate_pages.count > 0",
			"then_step":           "load_specs",
			"else_step":           "complete_no_candidates",
			"fail_on_non_numeric": true,
		},
		arrayShapedCandidates(),
	))
	if err == nil {
		t.Fatal("fail_on_non_numeric must fail the step on a non-numeric left side")
	}
	if !strings.Contains(err.Error(), "candidate_pages.count") {
		t.Fatalf("the error must name the field that failed to resolve, got: %v", err)
	}
}

// The flag must not disturb a comparison that evaluates — both outcomes.
func TestConditionalFailOnNonNumericInertWhenResolvable(t *testing.T) {
	for _, tc := range []struct {
		count    float64
		wantStep string
	}{
		{15, "load_specs"},
		{0, "complete_no_candidates"},
	} {
		result, err := ConditionalBranchAction(context.Background(), conditionalParams(
			map[string]interface{}{
				"condition":           "candidate_pages.count > 0",
				"then_step":           "load_specs",
				"else_step":           "complete_no_candidates",
				"fail_on_non_numeric": true,
			},
			objectShapedCandidates(tc.count),
		))
		if err != nil {
			t.Fatalf("count=%v: strict mode must not error on a resolvable comparison, got: %v", tc.count, err)
		}
		m := result.(map[string]interface{})
		if m["next_step_override"] != tc.wantStep {
			t.Fatalf("count=%v: want %s, got %v", tc.count, tc.wantStep, m["next_step_override"])
		}
	}
}

// Scope boundary: != null on an unresolvable path is a legitimate null probe
// and must stay silent even with the flag set. Kills the mutation that widens
// strictness past the numeric arms.
func TestConditionalFailOnNonNumericLeavesNullProbesAlone(t *testing.T) {
	result, err := ConditionalBranchAction(context.Background(), conditionalParams(
		map[string]interface{}{
			"condition":           "target_page.page_id != null",
			"then_step":           "load_candidate_pages",
			"else_step":           "complete_not_found",
			"fail_on_non_numeric": true,
		},
		map[string]interface{}{"target_page": map[string]interface{}{}},
	))
	if err != nil {
		t.Fatalf("a null probe must not error under fail_on_non_numeric, got: %v", err)
	}
	m := result.(map[string]interface{})
	if m["next_step_override"] != "complete_not_found" {
		t.Fatalf("unresolved != null must evaluate false, got %v", m["next_step_override"])
	}
}

// A non-numeric RIGHT side is a config typo and must also fail under the flag.
func TestConditionalFailOnNonNumericRightSide(t *testing.T) {
	_, err := ConditionalBranchAction(context.Background(), conditionalParams(
		map[string]interface{}{
			"condition":           "candidate_pages.count > banana",
			"then_step":           "a",
			"else_step":           "b",
			"fail_on_non_numeric": true,
		},
		objectShapedCandidates(3),
	))
	if err == nil {
		t.Fatal("fail_on_non_numeric must fail the step on a non-numeric right side")
	}
	if !strings.Contains(err.Error(), "banana") {
		t.Fatalf("the error must name the unparseable right side, got: %v", err)
	}
}

// The flag must propagate through AND/OR recursion, not just bare comparisons.
func TestConditionalFailOnNonNumericPropagatesThroughAnd(t *testing.T) {
	_, err := ConditionalBranchAction(context.Background(), conditionalParams(
		map[string]interface{}{
			"condition":           "target_page.page_id != null AND candidate_pages.count > 0",
			"then_step":           "a",
			"else_step":           "b",
			"fail_on_non_numeric": true,
		},
		map[string]interface{}{
			"target_page":     map[string]interface{}{"page_id": "p1"},
			"candidate_pages": []interface{}{map[string]interface{}{"name": "x"}},
		},
	))
	if err == nil {
		t.Fatal("fail_on_non_numeric must reach a numeric clause inside AND")
	}
	if !strings.Contains(err.Error(), "candidate_pages.count") {
		t.Fatalf("the error must name the failing clause's field, got: %v", err)
	}
}

// The exported lenient wrapper keeps its historical contract — external
// callers (two test files) depend on the three-argument form staying lenient.
func TestEvaluateStringConditionWrapperStaysLenient(t *testing.T) {
	met, err := evaluateStringCondition("candidate_pages.count > 0", arrayShapedCandidates(), zap.NewNop())
	if err != nil {
		t.Fatalf("the lenient wrapper must not error, got: %v", err)
	}
	if met {
		t.Fatal("an unresolvable numeric comparison must evaluate false in lenient mode")
	}
}
