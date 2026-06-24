// FILE: platform/orchestration/actions/diagnose_route_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Grounded in REAL contracts:
//   - ExecuteLLMPromptAction (ai_actions.go) returns {"result": <parsed JSON>},
//     so the verdict step's output is read at <verdict_field>.result.* — which is
//     exactly the VerdictWire shape (outcome/citations/next_scope/...).
//   - The coordinator (coordinator.go getNextStepFromResult) honours a "next_step"
//     key in an action's result map to OVERRIDE the workflow's next step — the
//     same mechanism conditional_route uses. That is how this action drives the
//     loop: it returns next_step = the gather step (iterate) or the emit step (stop).
//   - The loop guards + call-graph re-scope are the ALREADY-TESTED pkg/diagnose
//     engine (15 tests). This action is a thin chassis adapter over them.
//
// DESIGN DECISION (this turn): the VERDICT is its OWN execute_llm_prompt workflow
// step (observable, single responsibility), NOT called from inside a monolithic
// diagnose_run. This action — diagnose_route — is the per-iteration CONTROLLER
// that runs AFTER the verdict step: it applies the guards, decides continue/stop,
// computes the next scope by following the call graph, and routes the workflow.
// So the loop is expressed in the WORKFLOW (gather → verdict → route → back to
// gather | emit), with the guard/re-scope COMPLEXITY kept in Go (the guideline).
//
// READ-ONLY: pure decision logic over data already in collected_data. No DB, no
// LLM, no spawn, no writes.

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/pkg/diagnose"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseRouteInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"verdict_field", "state_field", "analysis_field",
		"gather_step", "emit_step", "max_iterations",
	},
	Defaults: map[string]interface{}{
		"verdict_field":  "verdict.result",  // ExecuteLLMPromptAction wraps JSON under .result
		"state_field":    "diagnose_state",  // where this action persists loop state across iterations
		"analysis_field": "repo_analysis",   // analyser Output, for call-graph re-scope
		"gather_step":    "assemble_bundle", // the step to loop BACK to for the next iteration
		"emit_step":      "emit",            // the step to proceed to when the loop stops
		"max_iterations": 5,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_route", DiagnoseRouteInputSpec)
}

// DiagnoseRouteAction is the loop controller. It runs once per iteration, after
// the verdict step. It returns a result map carrying:
//   - "next_step": the gather step (iterate) or the emit step (stop) — the
//     coordinator obeys this to route the workflow;
//   - "scope": the next iteration's scope (symbols/tables/runtime), which the
//     gather step reads to assemble the next bundle;
//   - "diagnose_state": the accumulated loop state (iteration, evidence trail,
//     hypothesis history) threaded into the next iteration;
//   - "data_requests": the verdict's read-only SELECTs (when iterating), which the
//     next load_runtime executes under a read-only transaction;
//   - "status"/"conclusion": when stopping, the engine's terminal result.
func DiagnoseRouteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	verdictField := datahelpers.GetStringField(config, "verdict_field", "verdict.result")
	stateField := datahelpers.GetStringField(config, "state_field", "diagnose_state")
	analysisField := datahelpers.GetStringField(config, "analysis_field", "repo_analysis")
	gatherStep := datahelpers.GetStringField(config, "gather_step", "assemble_bundle")
	emitStep := datahelpers.GetStringField(config, "emit_step", "emit")
	maxIter := datahelpers.GetIntField(config, "max_iterations", 5)

	// 1) Parse the verdict the LLM step produced (its JSON, under .result).
	verdictRaw := datahelpers.ExtractNestedField(params.CollectedData, verdictField)
	if verdictRaw == nil {
		return nil, fmt.Errorf("diagnose_route: no verdict at %q (expected ExecuteLLMPromptAction output under .result)", verdictField)
	}
	verdict, err := diagnose.ParseVerdictValue(verdictRaw) // wire-shape map -> domain Verdict
	if err != nil {
		return nil, fmt.Errorf("diagnose_route: parse verdict: %w", err)
	}

	// 2) Rehydrate loop state from the prior iteration (empty on the first).
	var st diagnose.LoopState
	if prev := datahelpers.ExtractNestedField(params.CollectedData, stateField); prev != nil {
		if err := diagnose.DecodeLoopState(prev, &st); err != nil {
			return nil, fmt.Errorf("diagnose_route: decode loop state: %w", err)
		}
	}
	if st.MaxIterations == 0 {
		st.MaxIterations = maxIter
	}

	// 3) Call-graph for re-scope (follow the evidence, not the symptom — §1a).
	var cg diagnose.CallGraph
	if analysisRaw := datahelpers.ExtractNestedField(params.CollectedData, analysisField); analysisRaw != nil {
		if g, err := diagnose.NewCallGraphFromValue(analysisRaw); err == nil {
			cg = g
		} else {
			logger.Warn("diagnose_route: could not build call graph; re-scope will use verdict next_scope verbatim", zap.Error(err))
		}
	}

	// 4) ONE engine step: apply guards, decide continue/stop, compute next scope.
	//    This is the tested pkg/diagnose logic, exposed as a single advance.
	decision := diagnose.Advance(&st, verdict, cg)

	// 5) Translate the engine decision into a workflow route + carry state/scope.
	result := map[string]interface{}{
		"diagnose_state": diagnose.EncodeLoopState(&st),
		"iteration":      st.Iteration,
	}

	if decision.Stop {
		// Loop ends — go to emit; carry the terminal result for emit/complete.
		result["next_step"] = emitStep
		result["status"] = decision.Status.String()
		result["conclusion"] = decision.Conclusion
		result["stopped_by"] = decision.StoppedBy
		result["evidence_trail"] = diagnose.EncodeTrail(st.Trail)
		logger.Info("diagnose_route: loop stopping",
			zap.String("status", decision.Status.String()),
			zap.String("stopped_by", decision.StoppedBy),
			zap.Int("iterations", st.Iteration))
		return result, nil
	}

	// Iterate — loop back to the gather step with the narrowed scope.
	result["next_step"] = gatherStep
	result["scope"] = diagnose.EncodeScope(decision.NextScope)
	result["hypothesis"] = decision.Hypothesis // the (possibly revised) hypothesis for the next bundle

	// Forward the verdict's read-only data_requests so the next gather (load_runtime)
	// runs them under a read-only transaction and folds the rows into the next bundle.
	// Read them from the RAW verdict wire value (verdict.result) rather than a typed
	// Verdict field — the data_requests are model-supplied wire data, and reading the
	// map directly keeps this independent of the engine's Verdict shape (the chassis
	// pkg/diagnose Verdict does not carry them). Keep ONLY the read-only ones
	// (diagnose.IsReadOnlySQL) — the route-layer filter; load_runtime re-lints (defence
	// in depth) and the read-only transaction is the real backstop. Code re-scope (the
	// call graph, carried in Scope) and data re-gather (these) are separate channels.
	dataReqs, droppedDR := readOnlyDataRequestsFromWire(verdictRaw)
	if len(dataReqs) > 0 {
		result["data_requests"] = dataReqs
	}
	if droppedDR > 0 {
		logger.Warn("diagnose_route: dropped non-read-only data_requests at route", zap.Int("dropped", droppedDR))
	}

	logger.Info("diagnose_route: iterating",
		zap.Int("next_iteration", st.Iteration+1),
		zap.Int("scope_size", len(decision.NextScope.Symbols)),
		zap.Int("data_requests", len(dataReqs)),
		zap.String("next_step", gatherStep))
	return result, nil
}

// readOnlyDataRequestsFromWire pulls the verdict's data_requests out of the RAW
// verdict wire value (the parsed LLM JSON at verdict.result) as a []{sql,why} list,
// keeping ONLY the read-only ones (diagnose.IsReadOnlySQL) and counting the rest as
// dropped. Reading the wire map directly (not a typed Verdict field) keeps this
// independent of the engine's Verdict shape; load_runtime re-lints and runs each
// survivor under a read-only transaction.
func readOnlyDataRequestsFromWire(verdictRaw interface{}) (kept []interface{}, dropped int) {
	m, ok := verdictRaw.(map[string]interface{})
	if !ok {
		return nil, 0
	}
	arr, ok := m["data_requests"].([]interface{})
	if !ok {
		return nil, 0
	}
	for _, it := range arr {
		dm, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		sqlText, _ := dm["sql"].(string)
		why, _ := dm["why"].(string)
		if strings.TrimSpace(sqlText) == "" {
			continue
		}
		if diagnose.IsReadOnlySQL(sqlText) != nil {
			dropped++
			continue
		}
		kept = append(kept, map[string]interface{}{"sql": sqlText, "why": why})
	}
	return kept, dropped
}
