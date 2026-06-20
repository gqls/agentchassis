// FILE: platform/orchestration/actions/diagnose_emit_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env.
//
// diagnose_emit shapes the human-facing diagnosis report from the loop's terminal
// state (which diagnose_route wrote into its result when it stopped). It is the
// single place that turns internal loop state into the report a person reads, so
// the responsibility is clear and the output is one well-formed object regardless
// of how the loop ended.
//
// READ-ONLY by design (DESIGN §4): it reads fields already in collected_data and
// returns a formatted map. It does NOT write to the DB. (There is deliberately no
// "diagnoses" table here — persisting a diagnosis is a separate decision that
// would need its own schema; the report goes back to the CALLER via
// complete_workflow's result_from, on the responses topic. If you later want it
// persisted, add a thin writer + a table in its own migration.) It NEVER emits a
// fix.
//
// Reads (defaults; diagnose_route is a router so its result lands under the step
// name "route" — see the migration):
//   status_field     "route.status"          CONFIRMED | UNVERIFIABLE
//   conclusion_field "route.conclusion"       the engine's conclusion text
//   stopped_by_field "route.stopped_by"       confirmed | iteration-cap | scope-not-narrowing | evidence-not-growing | hypothesis-thrash
//   trail_field      "route.evidence_trail"   the per-iteration trail (array)

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseEmitInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{"status_field", "conclusion_field", "stopped_by_field", "trail_field"},
	Defaults: map[string]interface{}{
		"status_field":     "route.status",
		"conclusion_field": "route.conclusion",
		"stopped_by_field": "route.stopped_by",
		"trail_field":      "route.evidence_trail",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_emit", DiagnoseEmitInputSpec)
}

func DiagnoseEmitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	status := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "status_field", "route.status"))
	conclusion := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "conclusion_field", "route.conclusion"))
	stoppedBy := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "stopped_by_field", "route.stopped_by"))
	trail := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "trail_field", "route.evidence_trail"))

	if status == "" && conclusion == "" {
		// The loop produced nothing to report — surface that honestly rather than
		// an empty success.
		logger.Warn("diagnose_emit: no terminal diagnosis found; emitting an explicit not-determined report")
		status = "UNVERIFIABLE"
		conclusion = "The diagnosis loop produced no terminal result (no verdict reached a stop). Hand to a human with the trail."
	}

	// One concise human line + the full structured detail. CONFIRMED is the only
	// status that asserts a cause; everything else is explicitly "not confirmed".
	var summary string
	if status == "CONFIRMED" {
		summary = "Diagnosis CONFIRMED — see conclusion for the cited cause."
	} else {
		summary = fmt.Sprintf("Diagnosis NOT confirmed (stopped: %s). Best-effort trail attached for a human; no fix proposed.", stoppedBy)
	}

	logger.Info("diagnose_emit: report shaped",
		zap.String("status", status),
		zap.String("stopped_by", stoppedBy))

	// Returned under this step's output_field (the migration sets it to
	// "diagnosis"); complete_workflow's result_from forwards it to the caller.
	return map[string]interface{}{
		"status":         status,
		"summary":        summary,
		"conclusion":     conclusion,
		"stopped_by":     stoppedBy,
		"evidence_trail": trail,
		"is_fix":         false, // explicit: this is a diagnosis, never a fix
	}, nil
}
