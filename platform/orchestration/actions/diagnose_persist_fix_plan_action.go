// FILE: platform/orchestration/actions/diagnose_persist_fix_plan_action.go
//
// F1.1a of the diagnosis→fix loop: validate and persist a CONSTRAINED EDIT
// PLAN produced by the fix-proposer agent from a CONFIRMED diagnosis.
//
// This slice deliberately writes NO code anywhere: the plan is a reviewable
// artifact in diagnosis_artifacts (kind='fix_plan'), fetchable by
// correlation_id like the bundles. The git branch + PR (F1.1b) is a separate
// slice behind the isolated write token (Q-C, decided 2026-07-07) — an agent
// whose only write surface is its own artifacts table cannot need one yet.
//
// Unlike the bundle write-through (observability, degrades on failure), a plan
// that fails validation MUST fail the step: persisting a malformed plan would
// hand F1.1b garbage to turn into a branch. The workflow routes the error to
// its complete_error step (config-level error_step — 001 §16).
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnosePersistFixPlanInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"fix_correlation_id"},
	Optional: []string{"plan_field", "max_edits", "max_plan_bytes"},
	Defaults: map[string]interface{}{
		// execute_llm_prompt with output_format=json leaves the parsed object
		// under <output_field>.result; the workflow sets output_field "proposal".
		"plan_field":     "proposal.result",
		"max_edits":      8,
		"max_plan_bytes": 32768,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_persist_fix_plan", DiagnosePersistFixPlanInputSpec)
}

// fixPlanEdit is one constrained edit. "Constrained" is the load-bearing word:
// an allowlisted operation on a named file/symbol with a rationale grounded in
// the diagnosis — never a free-form patch.
type fixPlanEdit struct {
	File      string `json:"file"`
	Symbol    string `json:"symbol,omitempty"`
	Operation string `json:"operation"` // modify | add | remove | config_change
	Rationale string `json:"rationale"`
	Sketch    string `json:"sketch"` // the intended change, described or diff-sketched
}

type fixPlan struct {
	Summary    string        `json:"summary"`
	Edits      []fixPlanEdit `json:"edits"`
	GroundedIn []string      `json:"grounded_in"` // citation quotes from the diagnosis
	Risks      string        `json:"risks,omitempty"`
}

var allowedFixOperations = map[string]bool{
	"modify": true, "add": true, "remove": true, "config_change": true,
}

// DiagnosePersistFixPlanAction validates the proposer's plan structurally and
// persists it to diagnosis_artifacts under the DIAGNOSIS run's correlation_id,
// so diagnosis → bundles → plan all join on one key.
func DiagnosePersistFixPlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "diagnose_persist_fix_plan"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		DiagnosePersistFixPlanInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	corr := strings.TrimSpace(inputs.Get("fix_correlation_id"))
	if corr == "" {
		return nil, fmt.Errorf("fix_correlation_id is empty")
	}

	planField := datahelpers.GetStringField(params.StepConfig.Config, "plan_field", "proposal.result")
	raw := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if raw == nil {
		return nil, fmt.Errorf("no plan found at %q", planField)
	}
	planJSON, err := planBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("plan not serialisable: %w", err)
	}

	maxBytes := datahelpers.GetIntField(params.StepConfig.Config, "max_plan_bytes", 32768)
	if len(planJSON) > maxBytes {
		return nil, fmt.Errorf("plan too large: %d bytes (cap %d)", len(planJSON), maxBytes)
	}

	var plan fixPlan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		// The first live run (ed164fed, 2026-07-10) failed here twice over: the
		// proposal hit max_tokens and arrived TRUNCATED, so execute_llm_prompt
		// stored the raw string instead of a parsed map. Say which failure this is.
		if !json.Valid(planJSON) {
			return nil, fmt.Errorf("plan JSON is invalid — likely truncated at the propose step's max_tokens; raise it or shrink the plan: %w", err)
		}
		return nil, fmt.Errorf("plan does not match the fix-plan schema: %w", err)
	}
	if problems := validateFixPlan(plan, datahelpers.GetIntField(params.StepConfig.Config, "max_edits", 8)); len(problems) > 0 {
		return nil, fmt.Errorf("plan failed validation: %s", strings.Join(problems, "; "))
	}

	files := make([]string, 0, len(plan.Edits))
	for _, e := range plan.Edits {
		files = append(files, e.File)
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"edit_count": len(plan.Edits),
		"files":      files,
		"summary":    plan.Summary,
	})

	// iteration 0 = a run-level artifact (not tied to one loop iteration).
	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO diagnosis_artifacts (
			correlation_id, orchestration_id, iteration, kind, body,
			metadata, source_agent, created_by
		) VALUES ($1, $2, 0, 'fix_plan', $3, $4::jsonb, $5, 'diagnose_persist_fix_plan')
	`, corr,
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		string(planJSON),
		string(metadata),
		nullIfEmpty(params.AgentType),
	); err != nil {
		return nil, fmt.Errorf("persist fix plan: %w", err)
	}

	logger.Info("diagnose_persist_fix_plan: plan persisted",
		zap.String("correlation_id", corr),
		zap.Int("edits", len(plan.Edits)),
		zap.Int("bytes", len(planJSON)))

	return map[string]interface{}{
		"persisted":  true,
		"edit_count": len(plan.Edits),
		"files":      files,
		"summary":    plan.Summary,
	}, nil
}

// planBytes coerces the proposal to JSON bytes whichever shape it arrived in:
// a parsed map (execute_llm_prompt's output_format=json happy path) or a raw
// string (what it stores when the model's JSON did not parse — code fences or
// truncation). Same map-or-string defence as decodeAnalysisOutput.
func planBytes(raw interface{}) ([]byte, error) {
	switch v := raw.(type) {
	case string:
		s := strings.TrimSpace(v)
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		return []byte(strings.TrimSpace(s)), nil
	case []byte:
		return v, nil
	default:
		return json.Marshal(raw)
	}
}

// validateFixPlan is the structural gate between the LLM's output and the
// artifact F1.1b will act on. It checks shape, not correctness — the council
// (F2) judges correctness; a human reviews the eventual PR.
func validateFixPlan(p fixPlan, maxEdits int) []string {
	var problems []string
	if strings.TrimSpace(p.Summary) == "" {
		problems = append(problems, "summary is empty")
	}
	if len(p.Edits) == 0 {
		problems = append(problems, "no edits")
	}
	if len(p.Edits) > maxEdits {
		problems = append(problems, fmt.Sprintf("%d edits exceeds cap %d — a fix this broad is architecture change, not a constrained fix", len(p.Edits), maxEdits))
	}
	if len(p.GroundedIn) == 0 {
		problems = append(problems, "grounded_in is empty — a fix plan must quote the diagnosis evidence it rests on")
	}
	for i, e := range p.Edits {
		tag := fmt.Sprintf("edit %d", i+1)
		f := strings.TrimSpace(e.File)
		switch {
		case f == "":
			problems = append(problems, tag+": file is empty")
		case strings.Contains(f, ".."), strings.HasPrefix(f, "/"), strings.ContainsAny(f, " \t\n"):
			problems = append(problems, tag+": file path must be repo-relative with no traversal or whitespace")
		}
		if !allowedFixOperations[e.Operation] {
			problems = append(problems, fmt.Sprintf("%s: operation %q not in the allowlist", tag, e.Operation))
		}
		if strings.TrimSpace(e.Rationale) == "" {
			problems = append(problems, tag+": rationale is empty")
		}
		if strings.TrimSpace(e.Sketch) == "" {
			problems = append(problems, tag+": sketch is empty")
		}
	}
	return problems
}
