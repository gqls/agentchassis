// FILE: platform/messaging/error_step_plan_test.go
package messaging

import (
	"testing"

	"go.uber.org/zap"
)

// TestConvertToWorkflowPlanCarriesErrorStep pins the field that was missing
// (bugs_open/085). convertToWorkflowPlan builds models.Step field by field, so a
// step-level key it does not name is silently dropped — and the coordinator's
// routeToErrorStepOrFail reads step.ErrorStep from the PERSISTED plan, not from
// agent_definitions. The measured effect of the omission: 0 of 14,209 persisted
// plan steps carried a step-level error_step, while 1,828 carried the
// config-level twin (which survives only because the whole config map is copied).
//
// The negative control matters as much as the positive case: a step with no
// error_step must still come out empty, so this test can fail.
func TestConvertToWorkflowPlanCarriesErrorStep(t *testing.T) {
	p := &MessageProcessor{logger: zap.NewNop()}

	plan := p.convertToWorkflowPlan(map[string]interface{}{
		"start_step": "resolve_links",
		"steps": map[string]interface{}{
			// The real shape from page-content-writer: error_step at step level
			// only, next_step pointing at the same handler.
			"resolve_links": map[string]interface{}{
				"action":     "call_agent",
				"next_step":  "select_sections",
				"error_step": "select_sections",
				"config": map[string]interface{}{
					"agent_type": "internal-link-resolver",
				},
			},
			// The shape that already worked: error_step inside config.
			"deploy_page": map[string]interface{}{
				"action":    "deploy_page",
				"next_step": "complete",
				"config": map[string]interface{}{
					"error_step": "mark_item_failed",
				},
			},
			// Negative control — no error handler declared anywhere.
			"select_sections": map[string]interface{}{
				"action":    "extract_fields",
				"next_step": "process_sections_loop",
			},
			"complete":         map[string]interface{}{"action": "complete"},
			"mark_item_failed": map[string]interface{}{"action": "update_work_item_status"},
		},
	})

	if got := plan.StartStep; got != "resolve_links" {
		t.Errorf("StartStep = %q, want %q", got, "resolve_links")
	}

	if got := plan.Steps["resolve_links"].ErrorStep; got != "select_sections" {
		t.Errorf("step-level error_step: ErrorStep = %q, want %q", got, "select_sections")
	}
	// Unchanged fields must survive alongside it.
	if got := plan.Steps["resolve_links"].NextStep; got != "select_sections" {
		t.Errorf("NextStep = %q, want %q", got, "select_sections")
	}
	if got, _ := plan.Steps["resolve_links"].Config["agent_type"].(string); got != "internal-link-resolver" {
		t.Errorf("config.agent_type = %q, want %q", got, "internal-link-resolver")
	}

	// The config-level route is the fallback in routeToErrorStepOrFail; it must
	// keep working exactly as before, and must NOT be promoted to the field.
	if got := plan.Steps["deploy_page"].ErrorStep; got != "" {
		t.Errorf("config-level error_step leaked into ErrorStep: %q, want empty", got)
	}
	if got, _ := plan.Steps["deploy_page"].Config["error_step"].(string); got != "mark_item_failed" {
		t.Errorf("config.error_step = %q, want %q", got, "mark_item_failed")
	}

	if got := plan.Steps["select_sections"].ErrorStep; got != "" {
		t.Errorf("negative control: ErrorStep = %q, want empty", got)
	}
}

// TestConvertToWorkflowPlanDanglingErrorStep covers the advisory warning path: a
// typo'd target must not change the plan or panic — the coordinator would route
// to it and fail with "step not found", i.e. the pre-fix outcome. The warning
// exists so the typo is visible at plan-build time rather than only when the
// step happens to fail.
func TestConvertToWorkflowPlanDanglingErrorStep(t *testing.T) {
	p := &MessageProcessor{logger: zap.NewNop()}

	plan := p.convertToWorkflowPlan(map[string]interface{}{
		"start_step": "a",
		"steps": map[string]interface{}{
			"a": map[string]interface{}{
				"action":     "noop",
				"error_step": "step_that_does_not_exist",
			},
		},
	})

	if got := plan.Steps["a"].ErrorStep; got != "step_that_does_not_exist" {
		t.Errorf("dangling error_step must still be carried verbatim, got %q", got)
	}
	if len(plan.Steps) != 1 {
		t.Errorf("plan gained or lost steps: %d, want 1", len(plan.Steps))
	}
}
