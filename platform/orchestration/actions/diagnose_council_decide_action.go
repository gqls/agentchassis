// FILE: platform/orchestration/actions/diagnose_council_decide_action.go
//
// F2.1 of the diagnosis→fix loop: the council's DECISION layer. Reviewers are
// LLM steps (parallel opinions, per Q-D); the aggregation is deliberately
// DETERMINISTIC Go — a decision you can read off the rules beats a third
// model opinion about two model opinions. Q-D (owner, 2026-07-07): all
// opinions are advisory by default; a hard_veto reviewer's negative verdict
// is a BLOCK. v1 places the hard-veto flag in step config
// (hard_veto_from: [reviewer names]); the definition-column vs council-config
// placement question stays open for F2.2.
//
// Decision rules, in order:
//  1. any hard-veto reviewer says veto            → "rejected"
//  2. any reviewer says veto (advisory veto)      → "rejected"
//  3. any reviewer says object                    → "revise"
//  4. all approve                                 → "approved"
//
// The report is persisted to diagnosis_artifacts (kind='council_report') on
// the SAME correlation_id as the diagnosis and the plan, and the decision is
// returned so the workflow can route on it. Like the plan-persist action and
// unlike the bundle write-through, a malformed reviewer verdict FAILS the
// step: a council that cannot read its reviewers must not wave a plan through.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseCouncilDecideInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"fix_correlation_id"},
	Optional:   []string{"review_fields", "hard_veto_from", "max_rounds"},
	Defaults:   map[string]interface{}{"max_rounds": 2},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_council_decide", DiagnoseCouncilDecideInputSpec)
}

// councilReview is one reviewer's structured opinion (the verdict-wire-style
// contract the F2 design asked for: verdict + objections + suggestions).
type councilReview struct {
	Reviewer   string `json:"reviewer"`
	Verdict    string `json:"verdict"` // approve | object | veto
	Objections []struct {
		Edit     int    `json:"edit"` // 1-based index into the plan's edits; 0 = plan-wide
		Problem  string `json:"problem"`
		Severity string `json:"severity,omitempty"` // low | medium | high
	} `json:"objections,omitempty"`
	Missing []string `json:"missing,omitempty"` // mechanisms the plan should cover but doesn't
	Notes   string   `json:"notes,omitempty"`
}

var councilVerdicts = map[string]bool{"approve": true, "object": true, "veto": true}

// DiagnoseCouncilDecideAction aggregates the reviewer steps' verdicts.
func DiagnoseCouncilDecideAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "diagnose_council_decide"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		DiagnoseCouncilDecideInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	corr := strings.TrimSpace(inputs.Get("fix_correlation_id"))
	if corr == "" {
		return nil, fmt.Errorf("fix_correlation_id is empty")
	}

	reviewFields := configStringSlice(params.StepConfig.Config, "review_fields", nil)
	if len(reviewFields) == 0 {
		return nil, fmt.Errorf("no review_fields configured — a council with no reviewers decides nothing")
	}
	hardVeto := map[string]bool{}
	for _, r := range configStringSlice(params.StepConfig.Config, "hard_veto_from", nil) {
		hardVeto[strings.ToLower(strings.TrimSpace(r))] = true
	}

	var reviews []councilReview
	for _, field := range reviewFields {
		raw := datahelpers.ExtractNestedField(params.CollectedData, field)
		if raw == nil {
			return nil, fmt.Errorf("reviewer output missing at %q", field)
		}
		rb, err := planBytes(raw) // same map-or-string defence as the plan itself
		if err != nil {
			return nil, fmt.Errorf("reviewer output at %q not serialisable: %w", field, err)
		}
		var rv councilReview
		if err := json.Unmarshal(rb, &rv); err != nil {
			if !json.Valid(rb) {
				return nil, fmt.Errorf("reviewer output at %q is invalid JSON — likely truncated at max_tokens: %w", field, err)
			}
			return nil, fmt.Errorf("reviewer output at %q does not match the review schema: %w", field, err)
		}
		rv.Verdict = strings.ToLower(strings.TrimSpace(rv.Verdict))
		if !councilVerdicts[rv.Verdict] {
			return nil, fmt.Errorf("reviewer %q returned unknown verdict %q (want approve|object|veto)", rv.Reviewer, rv.Verdict)
		}
		if strings.TrimSpace(rv.Reviewer) == "" {
			rv.Reviewer = field
		}
		reviews = append(reviews, rv)
	}

	decision, decidedBy := decideCouncil(reviews, hardVeto)

	report, _ := json.Marshal(map[string]interface{}{
		"decision":   decision,
		"decided_by": decidedBy,
		"reviews":    reviews,
	})
	metadata, _ := json.Marshal(map[string]interface{}{
		"decision":  decision,
		"reviewers": len(reviews),
	})
	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO diagnosis_artifacts (
			correlation_id, orchestration_id, iteration, kind, body,
			metadata, source_agent, created_by
		) VALUES ($1, $2, 0, 'council_report', $3, $4::jsonb, $5, 'diagnose_council_decide')
	`, corr,
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		string(report),
		string(metadata),
		nullIfEmpty(params.AgentType),
	); err != nil {
		return nil, fmt.Errorf("persist council report: %w", err)
	}

	// The revise loop's counter IS the council_report count for THIS PROPOSER
	// RUN — one per round, just written, so this reflects the current round.
	// Sourcing it from the durable artifact avoids threading loop state through
	// the workflow. Scoped by orchestration_id, NOT correlation alone: the
	// correlation belongs to the DIAGNOSIS and accumulates reports across
	// proposer re-runs (e08c5b01 had a report from a prior run, which would have
	// started this run at round 2 and exhausted it before any repropose). A
	// count failure must not strand a decided council: fall back to round 1,
	// which caps the loop at once more.
	round := 1
	if err := params.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM diagnosis_artifacts
		 WHERE correlation_id = $1 AND kind = 'council_report' AND orchestration_id = $2`,
		corr, nullIfEmpty(params.ExecutionContext.OrchestrationID)).Scan(&round); err != nil {
		logger.Warn("diagnose_council_decide: round count failed; treating as round 1", zap.Error(err))
		round = 1
	}
	maxRounds := datahelpers.GetIntField(params.StepConfig.Config, "max_rounds", 2)
	// Only a 'revise' decision with rounds left loops back; 'exhausted' is a
	// revise that ran out of rounds — terminal, and named so the human sees the
	// loop gave up rather than approved.
	shouldRevise := decision == "revise" && round < maxRounds
	if decision == "revise" && !shouldRevise {
		decision = "exhausted"
		decidedBy = fmt.Sprintf("%s — revise cap reached (%d rounds)", decidedBy, maxRounds)
	}

	logger.Info("diagnose_council_decide: decided",
		zap.String("correlation_id", corr),
		zap.String("decision", decision),
		zap.String("decided_by", decidedBy),
		zap.Int("reviewers", len(reviews)),
		zap.Int("round", round),
		zap.Bool("should_revise", shouldRevise))

	return map[string]interface{}{
		"decision":      decision,
		"decided_by":    decidedBy,
		"reviewers":     len(reviews),
		"round":         round,
		"should_revise": shouldRevise,
	}, nil
}

// decideCouncil applies the ordered rules. decidedBy names the rule that fired
// so the report is auditable without re-deriving.
func decideCouncil(reviews []councilReview, hardVeto map[string]bool) (decision, decidedBy string) {
	for _, r := range reviews {
		if r.Verdict == "veto" && hardVeto[strings.ToLower(r.Reviewer)] {
			return "rejected", "hard veto from " + r.Reviewer
		}
	}
	for _, r := range reviews {
		if r.Verdict == "veto" {
			return "rejected", "veto from " + r.Reviewer
		}
	}
	for _, r := range reviews {
		if r.Verdict == "object" {
			return "revise", "objection from " + r.Reviewer
		}
	}
	return "approved", "all reviewers approve"
}
