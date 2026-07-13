// FILE: platform/orchestration/actions/diagnose_run_checks_action.go
//
// F2.3 of the diagnosis→fix loop: the council's VERIFY step. Benchmark
// evidence (2026-07-10, run aadd532a): the guardian's round-3 blocking
// objections were all "containable by pre-deploy audit queries" — read-only
// facts the platform can answer (a column's type; whether a component function
// name exists in a table; a fleet-wide count). The proposer had no step that
// could run them, so it burned its last revise round blind and exhausted one
// verification short of approval.
//
// This action closes that gap. Reviewers may attach structured `checks`
// ([{sql, why}], SELECT/WITH only) to their review JSON; this step collects
// them from the configured fields and runs each under the SAME containment as
// the diagnosis loop's data_requests — IsReadOnlySQL lint → READ ONLY tx →
// statement_timeout → EXPLAIN cost gate → capped rendered rows — then hands
// the results to the next repropose so an objection that hinged on an
// unverified fact can be settled with evidence instead of another guess.
//
// The wire format deliberately matches the diagnosis verdict's data_requests
// ({sql, why}): one convention, one extractor (dataRequestsFromCollected), one
// executor (runDataRequests). Reuse, not a second mechanism.
package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseRunChecksInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{
		"check_fields", "max_checks",
		"explain_max_rows", "explain_max_cost", "row_cap", "cell_chars",
	},
	Defaults: map[string]interface{}{
		// Same size-guard defaults as diagnose_load_runtime's data_requests.
		"max_checks":       8,
		"explain_max_rows": 50000,
		"explain_max_cost": 0,
		"row_cap":          200,
		"cell_chars":       600,
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_run_checks", DiagnoseRunChecksInputSpec)
}

// DiagnoseRunChecksAction runs the reviewers' read-only verification queries.
func DiagnoseRunChecksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_run_checks"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_run_checks: no DB handle")
	}

	checkFields := configStringSlice(config, "check_fields", nil)
	if len(checkFields) == 0 {
		return nil, fmt.Errorf("no check_fields configured — a verify step with nowhere to look is a wiring mistake")
	}

	// Collect {sql, why} entries from every configured reviewer field. Absent
	// or empty fields are fine — reviewers only attach checks when a verdict
	// hinges on an unverified fact.
	var reqs []dataReq
	for _, field := range checkFields {
		reqs = append(reqs, dataRequestsFromCollected(params.CollectedData, field)...)
	}

	// Bound the spend: reviewers occasionally ask for everything at once. A
	// dropped check is not silent — it is named in the output so the model
	// (and the human) can see coverage was capped.
	maxChecks := datahelpers.GetIntField(config, "max_checks", 8)
	dropped := 0
	if maxChecks > 0 && len(reqs) > maxChecks {
		dropped = len(reqs) - maxChecks
		reqs = reqs[:maxChecks]
	}

	if len(reqs) == 0 {
		logger.Info("diagnose_run_checks: reviewers requested no checks")
		return map[string]interface{}{
			"results_text":   "(reviewers requested no verification checks this round)",
			"checks_run":     0,
			"checks_dropped": 0,
		}, nil
	}

	maxRows := datahelpers.GetIntField(config, "explain_max_rows", 50000)
	maxCost := datahelpers.GetIntField(config, "explain_max_cost", 0)
	rowCap := datahelpers.GetIntField(config, "row_cap", 200)
	cellChars := datahelpers.GetIntField(config, "cell_chars", 600)

	var b strings.Builder
	b.WriteString("Read-only verification queries your reviewers asked for, now answered:\n")
	runDataRequests(ctx, params.DB, reqs, &b, maxRows, maxCost, rowCap, cellChars)
	if dropped > 0 {
		fmt.Fprintf(&b, "\n> %d further check(s) dropped (max_checks=%d) — coverage was capped, not complete.\n", dropped, maxChecks)
	}

	logger.Info("diagnose_run_checks: ran reviewer checks",
		zap.Int("checks_run", len(reqs)),
		zap.Int("checks_dropped", dropped),
		zap.String("orchestration_id", orchIDForLog(params)),
	)

	return map[string]interface{}{
		"results_text":   b.String(),
		"checks_run":     len(reqs),
		"checks_dropped": dropped,
	}, nil
}

// orchIDForLog is a nil-safe accessor for the run id in log lines (constitution:
// put the run id in log lines).
func orchIDForLog(params ActionParams) string {
	if params.ExecutionContext == nil {
		return ""
	}
	return params.ExecutionContext.OrchestrationID
}
