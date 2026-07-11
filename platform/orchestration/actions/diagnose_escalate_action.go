// FILE: platform/orchestration/actions/diagnose_escalate_action.go
//
// F2.3 of the diagnosis→fix loop: the ESCALATION terminal. Before this,
// 'rejected' (guardian veto) and 'exhausted' (revise budget spent) were silent
// completes — the veto, with the reviewer's own recommended alternative inside
// it, was persisted and nothing consumed it. Benchmark evidence (2026-07-10):
// run 8c770fd5's guardian veto correctly identified an architecture-level fix
// dressed as a point-edit and NAMED the safe alternative; run aadd532a
// exhausted one verification short of approval with a concrete pre-deploy
// checklist in its final report. Both are decisions a human must take, and
// both died in the artifact table.
//
// Escalation makes "this needs a human / architecture review" a FIRST-CLASS
// SUCCESS OUTCOME, not a failure: it persists kind='escalation' carrying the
// whole hand-off package — the decision and why, the diagnosis conclusion, the
// final plan, and both reviews (whose notes/missing hold the recommended
// alternative and any unrun checklist). One fetch by correlation_id shows a
// human everything needed to decide. F1.1b(c)'s human-review surface will
// carry this package into the PR body.
//
// A failed escalation FAILS the step — loud, unlike the bundle write-through's
// degrade-to-warning: here the artifact IS the outcome.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseEscalateInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"fix_correlation_id"},
	Optional: []string{"council_field", "plan_field", "diagnosis_field", "review_fields"},
	Defaults: map[string]interface{}{
		"council_field":   "council",
		"plan_field":      "plan_persisted.plan_json",
		"diagnosis_field": "diagnosis_row.conclusion",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_escalate", DiagnoseEscalateInputSpec)
}

// DiagnoseEscalateAction persists the human hand-off package.
func DiagnoseEscalateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_escalate"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_escalate: no DB handle")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, DiagnoseEscalateInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	corr := strings.TrimSpace(inputs.Get("fix_correlation_id"))
	if corr == "" {
		return nil, fmt.Errorf("fix_correlation_id is empty")
	}

	// The council decision is the REASON for the escalation — without it the
	// package is meaningless, so its absence is a hard error. Everything else
	// is best-effort: escalation must be reachable from odd states, so what is
	// present is included and what is absent is named.
	councilField := datahelpers.GetStringField(config, "council_field", "council")
	councilRaw := datahelpers.ExtractNestedField(params.CollectedData, councilField)
	council, ok := councilRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("council decision missing at %q — an escalation needs its reason", councilField)
	}
	reason, _ := council["decision"].(string)
	decidedBy, _ := council["decided_by"].(string)
	round := datahelpers.GetIntField(council, "round", 0)

	diagnosis := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "diagnosis_field", "diagnosis_row.conclusion"))
	if diagnosis == "" {
		diagnosis = "(diagnosis conclusion not present in collected data)"
	}

	// The final plan, verbatim if present.
	var plan interface{}
	if raw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "plan_field", "plan_persisted.plan_json")); raw != nil {
		plan = raw
	} else {
		plan = "(no plan in collected data — all plans remain fetchable from diagnosis_artifacts kind=fix_plan by correlation_id)"
	}

	// Both reviews: their notes/missing carry the reviewer-recommended
	// alternative and the unrun pre-deploy checklist.
	var reviewsOut []interface{}
	for _, field := range configStringSlice(config, "review_fields", nil) {
		if raw := datahelpers.ExtractNestedField(params.CollectedData, field); raw != nil {
			reviewsOut = append(reviewsOut, raw)
		}
	}

	body, err := json.Marshal(map[string]interface{}{
		"reason":               reason,
		"decided_by":           decidedBy,
		"round":                round,
		"diagnosis_conclusion": diagnosis,
		"final_plan":           plan,
		"reviews":              reviewsOut,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal escalation package: %w", err)
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"reason": reason,
		"round":  round,
	})

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO diagnosis_artifacts (
			correlation_id, orchestration_id, iteration, kind, body,
			metadata, source_agent, created_by
		) VALUES ($1, $2, 0, 'escalation', $3, $4::jsonb, $5, 'diagnose_escalate')
	`, corr,
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		string(body),
		string(metadata),
		nullIfEmpty(params.AgentType),
	); err != nil {
		return nil, fmt.Errorf("persist escalation: %w", err)
	}

	logger.Info("diagnose_escalate: hand-off package persisted",
		zap.String("correlation_id", corr),
		zap.String("orchestration_id", orchIDForLog(params)),
		zap.String("reason", reason),
		zap.Int("round", round))

	return map[string]interface{}{
		"escalated": true,
		"reason":    reason,
		"round":     round,
		"kind":      "escalation",
	}, nil
}
