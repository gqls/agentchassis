// FILE: platform/orchestration/actions/fail_workflow_action.go
//
// fail_workflow — end a workflow with a FAILURE verdict, deliberately.
//
// The missing half of complete_workflow. Every workflow could already end in
// success (complete_workflow) or die by accident (a step erroring with no
// error_step), but there was no way to say "this run handled its failure
// cleanly and the outcome is still failure".
//
// Why that gap matters, concretely. A handler saga that catches its own
// failure, tidies up, and then calls complete_workflow reports SUCCESS to its
// caller — and complete_work_item's guard (complete_work_item_verification.go:
// handlerReportedFailure) reads response.status to decide whether the work item
// may be stamped 'complete'. A tidy-up path built on complete_workflow
// therefore produces exactly the state that guard exists to prevent: an item
// marked complete beside the evidence that the work failed. That was 54 items
// across 6 sites in the 2026-07-18 sweep, and the guard can only catch it if
// the handler actually says 'failed'.
//
// So: do the cleanup steps, then end HERE. The step returns an error, which is
// the orchestrator's existing failure signal — the caller's call_agent
// error_step fires, and the response carries a failure status the guard reads.
// Nothing new in the failure machinery; this only makes the failure
// intentional and greppable instead of relying on a step that happens to
// break.
//
// Config:
//   - reason:       static failure message (optional)
//   - reason_field: dotted path to a failure message in collected_data,
//                   preferred when present and non-empty
//
// First user: the gripper dossier's report-builder, whose failure path must
// publish the island's status sidecar (so the visitor is told the report
// failed) BEFORE the run is allowed to end in failure.

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FailWorkflowInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fail_workflow", FailWorkflowInputSpec)
}

func FailWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "fail_workflow"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	// initialize must NOT fail — it is a registration probe, not a run.
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	reason := failWorkflowReason(params.StepConfig.Config, params.CollectedData)
	logger.Warn("fail_workflow: ending this run with a FAILURE verdict", zap.String("reason", reason))
	return nil, fmt.Errorf("workflow failed: %s", reason)
}

// failWorkflowReason prefers a runtime message (reason_field) over the static
// one, so the failure carries the actual error rather than a generic label.
// Pure, so the precedence is unit-testable without an orchestrator.
func failWorkflowReason(config, collected map[string]interface{}) string {
	if rf, _ := config["reason_field"].(string); rf != "" {
		if v := datahelpers.ExtractNestedField(collected, rf); v != nil {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	if r, _ := config["reason"].(string); strings.TrimSpace(r) != "" {
		return strings.TrimSpace(r)
	}
	return "no reason given"
}
