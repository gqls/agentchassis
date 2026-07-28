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
//
// staged-v1 (feature builder delta 1, owner-approved 2026-07-17): the same
// action also validates STAGED plans — an ordered sequence of constrained edit
// plans with cross-stage file discipline and an encoded image-then-seed
// checklist. Discriminated by plan_format/stages in the body; the legacy
// single-plan path is unchanged. Schema and rules:
// docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/SCHEMA_staged_plan_v1.md
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnosePersistFixPlanInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"fix_correlation_id"},
	Optional:    []string{"plan_field", "max_edits", "max_plan_bytes", "max_stages", "max_total_edits"},
	Defaults: map[string]interface{}{
		// execute_llm_prompt with output_format=json leaves the parsed object
		// under <output_field>.result; the workflow sets output_field "proposal".
		"plan_field": "proposal.result",
		// max_edits is the whole-plan cap for a single plan and the PER-STAGE
		// cap for a staged one (D3). max_plan_bytes here is the single-plan
		// default; an unset config gets 131072 for staged plans in the action.
		"max_edits":       8,
		"max_plan_bytes":  32768,
		"max_stages":      6,
		"max_total_edits": 24,
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
	// staged plans only: code (default) | seed | doc. A seed is a DB-side file
	// shipped IN the PR, never executed by the loop — it must be covered by the
	// plan's post_merge_checklist (validated below). Ignored on single plans.
	ArtifactRole string `json:"artifact_role,omitempty"`
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

	// staged-v1 discriminator: plan_format and/or a stages key selects the
	// staged path; everything else is the legacy single plan, validated exactly
	// as before. The probe tolerates invalid JSON — the unmarshal below owns
	// that error (and its truncation hint).
	probe := probePlan(planJSON)
	staged := probe.PlanFormat != "" || rawPresent(probe.Stages)

	// A staged plan is a sequence of edit plans, so its default byte cap is
	// larger (D3). An explicit config cap wins for both shapes.
	defaultMaxBytes := 32768
	if staged {
		defaultMaxBytes = 131072
	}
	maxBytes := datahelpers.GetIntField(params.StepConfig.Config, "max_plan_bytes", defaultMaxBytes)
	if len(planJSON) > maxBytes {
		return nil, fmt.Errorf("plan too large: %d bytes (cap %d)", len(planJSON), maxBytes)
	}

	var editCount, stageCount int
	var files []string
	var summary string
	if staged {
		var plan stagedPlan
		if err := json.Unmarshal(planJSON, &plan); err != nil {
			if !json.Valid(planJSON) {
				return nil, fmt.Errorf("plan JSON is invalid — likely truncated at the propose step's max_tokens; raise it or shrink the plan: %w", err)
			}
			return nil, fmt.Errorf("plan does not match the staged-plan schema: %w", err)
		}
		caps := stagedPlanCaps{
			MaxStages:     datahelpers.GetIntField(params.StepConfig.Config, "max_stages", 6),
			MaxStageEdits: datahelpers.GetIntField(params.StepConfig.Config, "max_edits", 8),
			MaxTotalEdits: datahelpers.GetIntField(params.StepConfig.Config, "max_total_edits", 24),
		}
		if problems := validateStagedPlan(plan, rawPresent(probe.Edits), caps); len(problems) > 0 {
			return nil, fmt.Errorf("staged plan failed validation: %s", strings.Join(problems, "; "))
		}
		stageCount = len(plan.Stages)
		summary = plan.Summary
		for _, st := range plan.Stages {
			for _, e := range st.Edits {
				files = append(files, e.File)
			}
		}
		editCount = len(files)
	} else {
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
		summary = plan.Summary
		files = make([]string, 0, len(plan.Edits))
		for _, e := range plan.Edits {
			files = append(files, e.File)
		}
		editCount = len(plan.Edits)
	}

	meta := map[string]interface{}{
		"edit_count": editCount,
		"files":      files,
		"summary":    summary,
	}
	if staged {
		meta["stage_count"] = stageCount
		meta["plan_format"] = "staged-v1"
	}
	metadata, _ := json.Marshal(meta)

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
		zap.Int("edits", editCount),
		zap.Int("stages", stageCount),
		zap.Int("bytes", len(planJSON)))

	result := map[string]interface{}{
		"persisted":  true,
		"edit_count": editCount,
		"files":      files,
		"summary":    summary,
		// The validated plan verbatim, for the council reviewers' prompts —
		// a string, so template rendering is exact rather than a Go map dump.
		"plan_json": string(planJSON),
	}
	if staged {
		// Downstream conditionals (the implementer's stage loop, delta 2)
		// branch on these rather than re-probing the body.
		result["stage_count"] = stageCount
		result["plan_format"] = "staged-v1"
	}
	return result, nil
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
		problems = append(problems, editProblems(fmt.Sprintf("edit %d", i+1), e)...)
	}
	return problems
}

// editProblems is the per-edit shape gate, shared verbatim between the single
// fix plan and each staged-plan stage.
func editProblems(tag string, e fixPlanEdit) []string {
	var problems []string
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
	if reason := noOpEditReason(e.Sketch); reason != "" {
		problems = append(problems, fmt.Sprintf("%s: %s — a fix plan proposes changes, not observations; drop the edit or make it real", tag, reason))
	}
	return problems
}

// noOpEditReason flags edits that change nothing. The first live plan
// (correlation e08c5b01, 2026-07-10) passed validation with an edit whose
// sketch said "No code change required" and another proposing only a
// clarifying comment — structurally valid, semantically empty. Explicit
// phrases only: over-blocking a real edit is worse than letting the council
// catch a subtle no-op.
func noOpEditReason(sketch string) string {
	s := strings.ToLower(sketch)
	switch {
	case strings.Contains(s, "no code change"),
		strings.Contains(s, "no change required"),
		strings.Contains(s, "no change is required"),
		strings.Contains(s, "no change needed"),
		strings.Contains(s, "no change is needed"):
		return "sketch declares no code change"
	case strings.Contains(s, "clarifying note"),
		strings.Contains(s, "clarifying comment"),
		strings.Contains(s, "add a comment"),
		strings.Contains(s, "comment-only"):
		return "sketch is comment-only"
	}
	return ""
}

// ── staged plans (feature builder delta 1 — SCHEMA_staged_plan_v1.md) ────────

// stagedPlanStage is one constrained edit plan inside a staged (feature) plan.
// Array order is execution order; depends_on is validated documentation for
// the council, not a scheduler input (v1 runs nothing in parallel).
type stagedPlanStage struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Goal      string        `json:"goal"`
	DependsOn []string      `json:"depends_on,omitempty"`
	Edits     []fixPlanEdit `json:"edits"`
	// Each symbol must appear verbatim in one of the stage's produced file
	// bodies — checked deterministically at implement time (delta 2), not here.
	ExpectedSymbols []string   `json:"expected_symbols,omitempty"`
	Gate            stagedGate `json:"gate"`
}

// stagedGate: build defaults to true; false is legal only for a stage whose
// edits are all seed/doc (nothing for gofmt+go build to gate).
type stagedGate struct {
	Build *bool `json:"build,omitempty"`
}

// stagedChecklistEntry is one owner act after the PR merges. The loop never
// performs these — encoding them is delta 3's whole point: a feature is done
// only after the owner merges AND applies, two deliberate human acts.
type stagedChecklistEntry struct {
	Order  int    `json:"order"`
	Act    string `json:"act"` // image_deploy | seed_apply | verify
	File   string `json:"file,omitempty"`
	Detail string `json:"detail"`
}

type stagedPlan struct {
	PlanFormat         string                 `json:"plan_format"` // literal "staged-v1"
	Summary            string                 `json:"summary"`
	Stages             []stagedPlanStage      `json:"stages"`
	PostMergeChecklist []stagedChecklistEntry `json:"post_merge_checklist,omitempty"`
	GroundedIn         []string               `json:"grounded_in"`
	Risks              string                 `json:"risks,omitempty"`
}

// stagedPlanCaps: MaxStageEdits reuses the max_edits config knob — the
// single-plan cap becomes the per-stage cap (D3).
type stagedPlanCaps struct {
	MaxStages, MaxStageEdits, MaxTotalEdits int
}

var stagedArtifactRoles = map[string]bool{"": true, "code": true, "seed": true, "doc": true}
var stagedChecklistActs = map[string]bool{"image_deploy": true, "seed_apply": true, "verify": true}
var stageIDPattern = regexp.MustCompile(`^[a-z0-9_-]{1,16}$`)

// stagedProbe reads only the discriminator keys before committing to a schema.
// Invalid JSON is deliberately ignored at probe time — the full unmarshal owns
// that error (and its truncation hint).
type stagedProbe struct {
	PlanFormat string          `json:"plan_format"`
	Stages     json.RawMessage `json:"stages"`
	Edits      json.RawMessage `json:"edits"`
}

func probePlan(b []byte) stagedProbe {
	var p stagedProbe
	_ = json.Unmarshal(b, &p)
	return p
}

func rawPresent(m json.RawMessage) bool {
	s := strings.TrimSpace(string(m))
	return s != "" && s != "null"
}

// stageLabel names a stage by id when it has one, else by position — problems
// must stay addressable even when the id itself is what's wrong.
func stageLabel(st stagedPlanStage, i int) string {
	if id := strings.TrimSpace(st.ID); id != "" {
		return id
	}
	return fmt.Sprintf("%d", i+1)
}

// validateStagedPlan is the staged counterpart of validateFixPlan: shape, not
// correctness. Beyond the shared per-edit rules it owns the cross-stage file
// discipline (add-once, no modify-before-create, no create-then-delete) and
// the seed/checklist contract — exactly the properties the implementer's
// per-stage allowlist (delta 2) and the owner's apply checklist will trust.
func validateStagedPlan(p stagedPlan, hasTopLevelEdits bool, caps stagedPlanCaps) []string {
	var problems []string
	if p.PlanFormat != "staged-v1" {
		problems = append(problems, fmt.Sprintf("plan_format must be %q (got %q)", "staged-v1", p.PlanFormat))
	}
	if hasTopLevelEdits {
		problems = append(problems, "staged plan must not carry top-level edits — every edit belongs to a stage")
	}
	if strings.TrimSpace(p.Summary) == "" {
		problems = append(problems, "summary is empty")
	}
	if len(p.GroundedIn) == 0 {
		problems = append(problems, "grounded_in is empty — a staged plan must quote the approved spec it rests on")
	}
	if len(p.Stages) == 0 {
		problems = append(problems, "no stages")
	}
	if caps.MaxStages > 0 && len(p.Stages) > caps.MaxStages {
		problems = append(problems, fmt.Sprintf("%d stages exceeds cap %d — a build this broad needs splitting into more than one feature", len(p.Stages), caps.MaxStages))
	}

	idIndex := map[string]int{}
	for i, st := range p.Stages {
		tag := fmt.Sprintf("stage %d", i+1)
		id := strings.TrimSpace(st.ID)
		switch {
		case id == "":
			problems = append(problems, tag+": id is empty")
		case !stageIDPattern.MatchString(id):
			problems = append(problems, fmt.Sprintf("%s: id %q must match %s", tag, id, stageIDPattern.String()))
		default:
			if _, dup := idIndex[id]; dup {
				problems = append(problems, fmt.Sprintf("%s: id %q is already used", tag, id))
			} else {
				idIndex[id] = i
			}
		}
	}

	// First pass: where each path is created, so the second pass can see
	// forward references (a modify in stage 2 of a file added in stage 4).
	addedAt := map[string]int{}
	for i, st := range p.Stages {
		for _, e := range st.Edits {
			if strings.ToLower(strings.TrimSpace(e.Operation)) != "add" {
				continue
			}
			f := strings.TrimSpace(e.File)
			if f == "" {
				continue // reported by editProblems in the second pass
			}
			if _, dup := addedAt[f]; dup {
				problems = append(problems, fmt.Sprintf("%s is added more than once across the plan", f))
			} else {
				addedAt[f] = i
			}
		}
	}

	totalEdits := 0
	for i, st := range p.Stages {
		sTag := fmt.Sprintf("stage %s", stageLabel(st, i))
		if strings.TrimSpace(st.Title) == "" {
			problems = append(problems, sTag+": title is empty")
		}
		if strings.TrimSpace(st.Goal) == "" {
			problems = append(problems, sTag+": goal is empty — the goal is the council's reviewable contract")
		}
		if len(st.Edits) == 0 {
			problems = append(problems, sTag+": no edits")
		}
		if caps.MaxStageEdits > 0 && len(st.Edits) > caps.MaxStageEdits {
			problems = append(problems, fmt.Sprintf("%s: %d edits exceeds the per-stage cap %d", sTag, len(st.Edits), caps.MaxStageEdits))
		}
		totalEdits += len(st.Edits)

		seenDep := map[string]bool{}
		for _, dep := range st.DependsOn {
			depIdx, known := idIndex[dep]
			switch {
			case seenDep[dep]:
				problems = append(problems, fmt.Sprintf("%s: depends_on lists %q twice", sTag, dep))
			case !known:
				problems = append(problems, fmt.Sprintf("%s: depends_on references unknown stage %q", sTag, dep))
			case depIdx >= i:
				problems = append(problems, fmt.Sprintf("%s: depends_on %q must reference a strictly earlier stage", sTag, dep))
			}
			seenDep[dep] = true
		}

		for _, sym := range st.ExpectedSymbols {
			if strings.TrimSpace(sym) == "" {
				problems = append(problems, sTag+": expected_symbols contains an empty string")
			}
		}

		buildGate := st.Gate.Build == nil || *st.Gate.Build
		seenPath := map[string]bool{}
		for j, e := range st.Edits {
			tag := fmt.Sprintf("%s edit %d", sTag, j+1)
			problems = append(problems, editProblems(tag, e)...)

			role := strings.ToLower(strings.TrimSpace(e.ArtifactRole))
			if !stagedArtifactRoles[role] {
				problems = append(problems, fmt.Sprintf("%s: artifact_role %q not in the allowlist (code|seed|doc)", tag, e.ArtifactRole))
			}
			op := strings.ToLower(strings.TrimSpace(e.Operation))
			if op == "config_change" && (role == "seed" || role == "doc") {
				problems = append(problems, fmt.Sprintf("%s: a config_change targets an agent_definitions row, not a repo file — artifact_role %q is contradictory", tag, role))
			}
			if !buildGate && (role == "" || role == "code") {
				problems = append(problems, fmt.Sprintf("%s: gate.build=false but this is a code edit — only all-seed/doc stages may skip the build gate", tag))
			}

			f := strings.TrimSpace(e.File)
			if f == "" {
				continue
			}
			if seenPath[f] {
				problems = append(problems, fmt.Sprintf("%s: %s appears in more than one edit of this stage", tag, f))
			}
			seenPath[f] = true
			switch op {
			case "modify":
				if ai, ok := addedAt[f]; ok && ai > i {
					problems = append(problems, fmt.Sprintf("%s: modifies %s before the stage that adds it (stage %s)", tag, f, stageLabel(p.Stages[ai], ai)))
				}
			case "remove":
				if _, ok := addedAt[f]; ok {
					problems = append(problems, fmt.Sprintf("%s: removes %s, which this plan itself adds — create-then-delete churn", tag, f))
				}
			}
		}
	}
	if caps.MaxTotalEdits > 0 && totalEdits > caps.MaxTotalEdits {
		problems = append(problems, fmt.Sprintf("%d edits in total exceeds cap %d — a build this broad needs splitting", totalEdits, caps.MaxTotalEdits))
	}

	problems = append(problems, checklistProblems(p)...)
	return problems
}

// checklistProblems enforces the seed discipline (delta 3): seeds ship as PR
// files the owner applies AFTER the image, so the checklist — not prose —
// carries that ordering, and validation makes the wrong order unexpressible.
func checklistProblems(p stagedPlan) []string {
	var problems []string
	seedFiles := map[string]bool{}
	needChecklist := false
	shipsCode := false
	for _, st := range p.Stages {
		for _, e := range st.Edits {
			role := strings.ToLower(strings.TrimSpace(e.ArtifactRole))
			op := strings.ToLower(strings.TrimSpace(e.Operation))
			if role == "seed" {
				needChecklist = true
				if f := strings.TrimSpace(e.File); f != "" {
					seedFiles[f] = true
				}
			}
			if op == "config_change" {
				needChecklist = true
			}
			// Code edits are what make an image deploy necessary; seed/doc
			// files ship in the PR but change no binary.
			if op != "config_change" && (role == "" || role == "code") {
				shipsCode = true
			}
		}
	}
	if len(p.PostMergeChecklist) == 0 {
		if needChecklist {
			problems = append(problems, "post_merge_checklist is required when the plan ships seed files or config_change edits")
		}
		return problems
	}

	seenOrder := map[int]bool{}
	minImage, minSeed := 0, 0 // 0 = none seen
	covered := map[string]int{}
	for i, c := range p.PostMergeChecklist {
		tag := fmt.Sprintf("checklist entry %d", i+1)
		if c.Order <= 0 {
			problems = append(problems, tag+": order must be a positive integer")
		} else if seenOrder[c.Order] {
			problems = append(problems, fmt.Sprintf("%s: order %d is already used", tag, c.Order))
		}
		seenOrder[c.Order] = true
		if !stagedChecklistActs[c.Act] {
			problems = append(problems, fmt.Sprintf("%s: act %q not in the allowlist (image_deploy|seed_apply|verify)", tag, c.Act))
		}
		if strings.TrimSpace(c.Detail) == "" {
			problems = append(problems, tag+": detail is empty")
		}
		switch c.Act {
		case "image_deploy":
			if minImage == 0 || c.Order < minImage {
				minImage = c.Order
			}
		case "seed_apply":
			f := strings.TrimSpace(c.File)
			switch {
			case f == "":
				problems = append(problems, tag+": seed_apply must name the seed file it applies")
			case !seedFiles[f]:
				problems = append(problems, fmt.Sprintf("%s: seed_apply names %s, which no seed edit ships", tag, f))
			default:
				covered[f]++
			}
			if minSeed == 0 || c.Order < minSeed {
				minSeed = c.Order
			}
		}
	}
	for f := range seedFiles {
		if covered[f] == 0 {
			problems = append(problems, fmt.Sprintf("seed file %s has no seed_apply checklist entry", f))
		} else if covered[f] > 1 {
			problems = append(problems, fmt.Sprintf("seed file %s has %d seed_apply entries — exactly one expected", f, covered[f]))
		}
	}
	// Image-before-seed is hard ONLY when the plan ships code: a seed-only
	// plan has no image to wait for, and demanding an image_deploy entry
	// would make the checklist lie (found live by the F1.2 pilot's run 2 —
	// the designer's truthful seed_apply→verify checklist was refused).
	if minSeed > 0 {
		switch {
		case minImage > 0 && minImage >= minSeed:
			problems = append(problems, "image_deploy must come strictly before any seed_apply — a seed naming an unregistered action fails at runtime (image first, then seed)")
		case minImage == 0 && shipsCode:
			problems = append(problems, "plan ships code and a seed: an image_deploy entry must come strictly before any seed_apply — a seed naming an unregistered action fails at runtime (image first, then seed)")
		}
	}
	return problems
}
