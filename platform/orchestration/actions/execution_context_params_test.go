package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

func TestExecutionContextParam_IgnoresNonNamespacedPaths(t *testing.T) {
	execCtx := &types.ExecutionContext{CorrelationID: "corr-1"}

	// The whole safety argument for this feature is that it is additive: every
	// path shape that exists in a live workflow today must fall straight
	// through. Includes the shapes that LOOK adjacent.
	for _, path := range []string{
		"claimed.work_item_id",
		"correlation_id",
		"input_data.correlation_id",
		"ctx.correlation_id",
		"$ctxcorrelation_id",
		"",
	} {
		_, isCtx, err := executionContextParam(execCtx, path)
		if isCtx {
			t.Errorf("path %q was claimed by the $ctx. namespace; it must fall through to collected_data", path)
		}
		if err != nil {
			t.Errorf("path %q: fall-through must not error, got %v", path, err)
		}
	}
}

func TestExecutionContextParam_ResolvesEveryKnownField(t *testing.T) {
	execCtx := &types.ExecutionContext{
		CorrelationID:         "corr-1",
		OrchestrationID:       "orch-1",
		ParentOrchestrationID: "parent-1",
		OrchestrationName:     "name-1",
		ClientID:              "client-1",
		RequestID:             "req-1",
		StepName:              "claim_item",
		GroupID:               "group-1",
	}

	cases := map[string]string{
		"$ctx.correlation_id":          "corr-1",
		"$ctx.orchestration_id":        "orch-1",
		"$ctx.parent_orchestration_id": "parent-1",
		"$ctx.orchestration_name":      "name-1",
		"$ctx.client_id":               "client-1",
		"$ctx.request_id":              "req-1",
		"$ctx.step_name":               "claim_item",
		"$ctx.group_id":                "group-1",
	}
	for path, want := range cases {
		got, isCtx, err := executionContextParam(execCtx, path)
		if !isCtx {
			t.Errorf("path %q: not recognised as a $ctx. path", path)
			continue
		}
		if err != nil {
			t.Errorf("path %q: unexpected error %v", path, err)
			continue
		}
		if got != want {
			t.Errorf("path %q: got %q, want %q", path, got, want)
		}
	}
}

// The failing branches are the point of this helper: a bind that silently
// produced "" would fill a stamp column with empty strings, which looks
// populated in every subsequent query. Each of these must fail the step.
func TestExecutionContextParam_FailsLoudRatherThanBindingAPlaceholder(t *testing.T) {
	full := &types.ExecutionContext{CorrelationID: "corr-1"}

	t.Run("unknown field", func(t *testing.T) {
		_, isCtx, err := executionContextParam(full, "$ctx.correlationid")
		if !isCtx {
			t.Fatal("a typo inside the namespace must stay inside the namespace, not fall through to collected_data")
		}
		if err == nil {
			t.Fatal("unknown execution-context field bound without error")
		}
		// The message has to name the field set, or the next person guesses again.
		if !strings.Contains(err.Error(), "correlation_id") {
			t.Errorf("error should list the known fields, got: %v", err)
		}
	})

	t.Run("empty value", func(t *testing.T) {
		_, isCtx, err := executionContextParam(full, "$ctx.parent_orchestration_id")
		if !isCtx || err == nil {
			t.Fatalf("an empty context field must error; isCtx=%v err=%v", isCtx, err)
		}
	})

	t.Run("nil execution context", func(t *testing.T) {
		_, isCtx, err := executionContextParam(nil, "$ctx.correlation_id")
		if !isCtx || err == nil {
			t.Fatalf("a nil execution context must error; isCtx=%v err=%v", isCtx, err)
		}
	})
}
