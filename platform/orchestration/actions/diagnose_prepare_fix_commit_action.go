// FILE: platform/orchestration/actions/diagnose_prepare_fix_commit_action.go
//
// F1.1b(c) part 2: the fix-implementer's SAFETY CORE. Sits between the
// sketch_to_files LLM step and the git-adapter calls, and enforces the
// owner-decided constraint (Q-C, 2026-07-07): the approved plan's file list is
// a HARD ALLOWLIST — edit operations apply ONLY to named files.
//
// Deterministic Go, no model judgement:
//  1. every LLM-produced file path must exactly match a modify/add edit in the
//     APPROVED plan (config_change edits target agent_definitions rows, not
//     repo files — a fabricated file for one is rejected as out-of-allowlist);
//  2. every modify/add edit in the plan must be implemented — a missing file
//     is an INCOMPLETE implementation, not a smaller one;
//  3. empty bodies are rejected; when the original file bodies are provided,
//     an unchanged body is rejected too (the no-op discipline the plan
//     validator applies to edits, applied again to implementations).
//
// On success it assembles everything the git steps need: the repo-relative
// files map (GitCommitData shape, no domain — platform commits are
// repo-relative), a deterministic branch name (fix/<short-correlation>), the
// commit message, and the PR title/body carrying the Q-H hand-off package
// (diagnosis conclusion + plan + council decision). A validation failure FAILS
// the step: nothing reaches the write surface.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnosePrepareFixCommitInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"fix_correlation_id"},
	Optional: []string{
		"plan_field", "files_field", "originals_field",
		"diagnosis_field", "council_field", "repo_name", "base_branch", "base_branch_field",
		"branch_field", "commit_message_field", "expected_symbols_field",
	},
	Defaults: map[string]interface{}{
		"plan_field":      "plan_row.body",
		"files_field":     "implementation.result",
		"diagnosis_field": "diagnosis_row.conclusion",
		"council_field":   "council_row.body",
		"repo_name":       "agentchassis",
		"base_branch":     "main",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_prepare_fix_commit", DiagnosePrepareFixCommitInputSpec)
}

// implFile is one whole-file output from the sketch_to_files step.
type implFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// implWire is the sketch_to_files output contract.
type implWire struct {
	Files []implFile `json:"files"`
	Notes string     `json:"notes,omitempty"`
}

// planEditLite is the slice of the fix-plan schema this action needs.
type planEditLite struct {
	File      string `json:"file"`
	Symbol    string `json:"symbol,omitempty"`
	Operation string `json:"operation"`
	Rationale string `json:"rationale,omitempty"`
}

type planLite struct {
	Summary    string         `json:"summary"`
	Edits      []planEditLite `json:"edits"`
	GroundedIn []string       `json:"grounded_in,omitempty"`
	Risks      string         `json:"risks,omitempty"`
}

// DiagnosePrepareFixCommitAction validates the implementation against the
// approved plan's allowlist and assembles the git payloads.
func DiagnosePrepareFixCommitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_prepare_fix_commit"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, DiagnosePrepareFixCommitInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	corr := strings.TrimSpace(inputs.Get("fix_correlation_id"))
	if corr == "" {
		return nil, fmt.Errorf("fix_correlation_id is empty")
	}

	// ── the approved plan (map-or-string, like the persist action) ──────────
	planRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "plan_field", "plan_row.body"))
	if planRaw == nil {
		return nil, fmt.Errorf("approved plan missing at plan_field")
	}
	pb, err := planBytes(planRaw)
	if err != nil {
		return nil, fmt.Errorf("plan not serialisable: %w", err)
	}
	var plan planLite
	if err := json.Unmarshal(pb, &plan); err != nil {
		return nil, fmt.Errorf("plan does not match the fix-plan schema: %w", err)
	}

	// ── the implementation files ─────────────────────────────────────────────
	filesRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "files_field", "implementation.result"))
	if filesRaw == nil {
		return nil, fmt.Errorf("implementation missing at files_field")
	}
	fb, err := planBytes(filesRaw)
	if err != nil {
		return nil, fmt.Errorf("implementation not serialisable: %w", err)
	}
	var impl implWire
	if err := json.Unmarshal(fb, &impl); err != nil {
		if !json.Valid(fb) {
			return nil, fmt.Errorf("implementation is invalid JSON — likely truncated at max_tokens: %w", err)
		}
		return nil, fmt.Errorf("implementation does not match the {files:[{path,content}]} schema: %w", err)
	}

	// originals (optional): path → current body, for no-op detection.
	originals := map[string]string{}
	if of := datahelpers.GetStringField(config, "originals_field", ""); of != "" {
		if raw, ok := datahelpers.ExtractNestedField(params.CollectedData, of).(map[string]interface{}); ok {
			for p, v := range raw {
				if s, ok := v.(string); ok {
					originals[p] = s
				}
			}
		}
	}

	files, violations := validateImplementation(plan, impl, originals)
	if len(violations) > 0 {
		return nil, fmt.Errorf("implementation rejected (%d violations): %s",
			len(violations), strings.Join(violations, "; "))
	}

	// Stage-loop seam (delta 2): the stage router declares symbols that must
	// literally appear in the stage's produced files — a deterministic check in
	// the same spirit as the allowlist. Unset keeps single-plan behaviour.
	if sf := datahelpers.GetStringField(config, "expected_symbols_field", ""); sf != "" {
		if missing := missingExpectedSymbols(collectedStringSlice(params.CollectedData, sf), files); len(missing) > 0 {
			return nil, fmt.Errorf("implementation rejected: expected symbols not present in any produced file: %s",
				strings.Join(missing, ", "))
		}
	}

	// ── payload assembly ─────────────────────────────────────────────────────
	short := corr
	if len(short) > 8 {
		short = short[:8]
	}
	branch := "fix/" + short
	repoName := datahelpers.GetStringField(config, "repo_name", "agentchassis")
	baseBranch := datahelpers.GetStringField(config, "base_branch", "main")
	// F1.2: a per-run base branch wins over the literal when configured, mirroring
	// read_current_files' ref_field — so the implementer bases the fix branch and
	// PR on the branch the run names (input_data.base_branch), not a literal that
	// goes stale the moment the active branch moves. Unset/unresolvable keeps the
	// literal default (main).
	if bf := datahelpers.GetStringField(config, "base_branch_field", ""); bf != "" {
		if v := strings.TrimSpace(datahelpers.ExtractNestedFieldString(params.CollectedData, bf)); v != "" {
			baseBranch = v
		}
	}

	commitMessage := fmt.Sprintf("fix(%s): %s\n\nAutomated fix-implementer commit for diagnosis %s.\nPlan approved by the review council; human review terminal — do not merge without review.",
		short, firstSentence(plan.Summary), corr)

	// Stage-loop seam (delta 2): the router supplies the feat/* branch and the
	// per-stage commit message; unset fields keep the derivations above.
	// CONFIGURED-BUT-EMPTY IS AN ERROR (council-gate review 5a65ec4c,
	// bug-historian, high): falling back here would derive fix/<corr> and commit
	// a stage's files to a DIFFERENT branch than the one the loop is building —
	// silently, with no failed work item. Unset keeps the single-plan derivation.
	if bf := datahelpers.GetStringField(config, "branch_field", ""); bf != "" {
		v := strings.TrimSpace(datahelpers.ExtractNestedFieldString(params.CollectedData, bf))
		if v == "" {
			return nil, fmt.Errorf("branch_field %q is configured but resolved empty — refusing to fall back to %q, which would commit to the wrong branch", bf, branch)
		}
		branch = v
	}
	if mf := datahelpers.GetStringField(config, "commit_message_field", ""); mf != "" {
		v := strings.TrimSpace(datahelpers.ExtractNestedFieldString(params.CollectedData, mf))
		if v == "" {
			return nil, fmt.Errorf("commit_message_field %q is configured but resolved empty", mf)
		}
		commitMessage = v
	}

	diagnosis := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "diagnosis_field", "diagnosis_row.conclusion"))
	councilRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "council_field", "council_row.body"))

	prTitle := fmt.Sprintf("fix(%s): %s", short, firstSentence(plan.Summary))
	prBody := buildFixPRBody(corr, plan, diagnosis, councilRaw)

	logger.Info("diagnose_prepare_fix_commit: implementation validated",
		zap.String("correlation_id", corr),
		zap.String("orchestration_id", orchIDForLog(params)),
		zap.String("branch", branch),
		zap.Int("files", len(files)))

	return map[string]interface{}{
		"files":          files,
		"branch":         branch,
		"base_branch":    baseBranch,
		"repo_name":      repoName,
		"commit_message": commitMessage,
		"pr_title":       prTitle,
		"pr_body":        prBody,
		"files_count":    len(files),
	}, nil
}

// validateImplementation is the ALLOWLIST core (Q-C), extracted pure so the
// tests exercise the real logic. modify/add edits name repo files;
// config_change targets agent_definitions rows and MUST NOT surface as a file;
// remove is excluded by plan rule 4. Returns the GitCommitData-shaped files
// map and every violation found (empty violations == valid).
func validateImplementation(plan planLite, impl implWire, originals map[string]string) (map[string]interface{}, []string) {
	allow := map[string]bool{}
	for _, e := range plan.Edits {
		op := strings.ToLower(strings.TrimSpace(e.Operation))
		if op == "modify" || op == "add" {
			allow[strings.TrimSpace(e.File)] = true
		}
	}
	if len(allow) == 0 {
		return nil, []string{"approved plan has no modify/add edits — nothing to implement as a commit (config_change-only plans are applied to agent_definitions, not the repo)"}
	}

	seen := map[string]bool{}
	files := map[string]interface{}{}
	var violations []string
	for _, f := range impl.Files {
		path := strings.TrimSpace(f.Path)
		switch {
		case path == "":
			violations = append(violations, "a file entry has an empty path")
		case !allow[path]:
			violations = append(violations, fmt.Sprintf("%s is OUTSIDE the approved plan's allowlist", path))
		case strings.TrimSpace(f.Content) == "":
			violations = append(violations, fmt.Sprintf("%s has an empty body", path))
		case seen[path]:
			violations = append(violations, fmt.Sprintf("%s appears twice in the implementation", path))
		default:
			if orig, ok := originals[path]; ok && orig == f.Content {
				violations = append(violations, fmt.Sprintf("%s is byte-identical to the original — a no-op implementation", path))
				continue
			}
			seen[path] = true
			files[path] = map[string]interface{}{"content": f.Content, "encoding": "utf-8"}
		}
	}
	for path := range allow {
		if !seen[path] {
			violations = append(violations, fmt.Sprintf("plan edit for %s was NOT implemented", path))
		}
	}
	return files, violations
}

// buildFixPRBody renders the Q-H hand-off package as PR markdown: what was
// diagnosed, what the plan changes and why, and what the council decided.
// Tolerant of absences — it names what is missing rather than failing, because
// the PR is the human surface and must always carry whatever exists.
func buildFixPRBody(corr string, plan planLite, diagnosis string, councilRaw interface{}) string {
	var b strings.Builder
	b.WriteString("## Automated fix proposal\n\n")
	fmt.Fprintf(&b, "Diagnosis correlation: `%s` — full artifact trail in `diagnosis_artifacts` (bundles, plans, council reports).\n\n", corr)

	b.WriteString("### Diagnosis\n\n")
	if strings.TrimSpace(diagnosis) != "" {
		b.WriteString(diagnosis + "\n\n")
	} else {
		b.WriteString("_(diagnosis conclusion not present in run data — fetch by correlation id)_\n\n")
	}

	b.WriteString("### Plan (council-approved)\n\n")
	if strings.TrimSpace(plan.Summary) != "" {
		b.WriteString(plan.Summary + "\n\n")
	}
	for i, e := range plan.Edits {
		fmt.Fprintf(&b, "%d. `%s` — %s `%s`: %s\n", i+1, e.File, e.Operation, e.Symbol, e.Rationale)
	}
	b.WriteString("\n")
	if len(plan.GroundedIn) > 0 {
		b.WriteString("### Grounded in\n\n")
		for _, g := range plan.GroundedIn {
			fmt.Fprintf(&b, "> %s\n\n", g)
		}
	}
	if strings.TrimSpace(plan.Risks) != "" {
		b.WriteString("### Risks (from the plan)\n\n" + plan.Risks + "\n\n")
	}

	b.WriteString("### Council decision\n\n")
	if councilRaw != nil {
		if cb, err := planBytes(councilRaw); err == nil {
			b.WriteString("```json\n" + string(cb) + "\n```\n\n")
		}
	} else {
		b.WriteString("_(council report not present in run data — fetch by correlation id)_\n\n")
	}

	b.WriteString("---\n_Created by the diagnosis→fix loop. Human review terminal: nothing merges itself._\n")
	return b.String()
}

// firstSentence trims a plan summary to a title-sized first sentence.
func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, ".\n"); i > 0 {
		s = s[:i]
	}
	if len(s) > 100 {
		s = s[:100]
	}
	return s
}

// missingExpectedSymbols returns the declared symbols that appear verbatim in
// NONE of the produced file bodies (files is the GitCommitData-shaped map).
// Pure so the tests exercise the real logic.
func missingExpectedSymbols(symbols []string, files map[string]interface{}) []string {
	var bodies []string
	for _, v := range files {
		if m, ok := v.(map[string]interface{}); ok {
			if s, ok := m["content"].(string); ok {
				bodies = append(bodies, s)
			}
		}
	}
	var missing []string
	for _, sym := range symbols {
		sym = strings.TrimSpace(sym)
		if sym == "" {
			continue
		}
		found := false
		for _, b := range bodies {
			if strings.Contains(b, sym) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, sym)
		}
	}
	return missing
}

// collectedStringSlice reads a []string out of collected_data at path,
// tolerating the []interface{} shape JSON round-trips produce.
func collectedStringSlice(collected map[string]interface{}, path string) []string {
	raw := datahelpers.ExtractNestedField(collected, path)
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
