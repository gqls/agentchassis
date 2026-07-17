// FILE: platform/orchestration/actions/feature_stage_route_action.go
//
// Feature builder delta 2 (DESIGN_stage_loop_delta2.md, E1-E5 approved
// 2026-07-17): the stage loop's ONLY new control machinery. Deterministic —
// no model judgement; it walks a council-approved staged plan
// (SCHEMA_staged_plan_v1.md) one stage per invocation.
//
// The load-bearing trick: each invocation emits the CURRENT stage as a
// SINGLE-PLAN shape (stage_plan — a fix-plan JSON string holding just that
// stage's edits), so the proven read/prepare actions loop without knowing
// they are looping. Alongside it: the ref to read at (base ref for stage 1,
// the feat/* branch thereafter, so later stages see earlier commits), the
// per-stage commit message, gate/edit flags, and — once the stages are
// exhausted — the PR payload and the go-test packages DERIVED from the plan's
// edited .go files (D6: the model never declares its own test surface).
//
// Like diagnose_route, this action reads its own PRIOR output back at
// state_field (default "stage.feature_state" — it MUST track the step's
// output_field, else the loop re-seeds every iteration). Seeding also
// enforces E4: a pre-existing feat/* branch is a HARD error, never silently
// reused — the fix loop's delete-stale-branches gotcha turned loud. The
// existence check uses GITHUB_READ_TOKEN (this agent rides the
// isRepoCloningAgent spawn gate, like fix-implementer).
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FeatureStageRouteInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"fix_correlation_id"},
	Optional: []string{
		"plan_field", "state_field", "branch_prefix",
		"base_ref", "base_ref_field", "council_field",
		"repo_owner", "repo_name", "require_fresh_branch",
	},
	Defaults: map[string]interface{}{
		"plan_field": "plan_row.body",
		// Reads the PRIOR route output back — must track this step's
		// output_field ("stage"), same trap diagnose_route documents.
		"state_field":          "stage.feature_state",
		"branch_prefix":        "feat/", // D2
		"base_ref":             "main",
		"base_ref_field":       "input_data.base_ref",
		"council_field":        "council_row.body",
		"repo_owner":           "gqls",
		"repo_name":            "agentchassis",
		"require_fresh_branch": true, // E4
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("feature_stage_route", FeatureStageRouteInputSpec)
}

// featureLoopState is the accumulated loop state, persisted in this action's
// own output and rehydrated on the next invocation. Current is the 1-based
// index of the stage emitted by the LAST invocation (0 = not seeded); a
// re-entry therefore marks stage Current done and advances.
type featureLoopState struct {
	Current int             `json:"current"`
	Branch  string          `json:"branch"`
	BaseRef string          `json:"base_ref"`
	Done    []doneStageNote `json:"done,omitempty"`
}

// doneStageNote is what later stages (and the PR body) need to know about an
// earlier one — derived from the plan, whose allowlist the prepare step has
// already enforced against the actual commit.
type doneStageNote struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Files []string `json:"files,omitempty"`
}

// FeatureStageRouteAction is the loop controller for the stage-loop
// implementer. One invocation = seed (stage 1) or advance (mark the previous
// stage done, emit the next / the terminal PR payload).
func FeatureStageRouteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "feature_stage_route"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, FeatureStageRouteInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	corr := strings.TrimSpace(inputs.Get("fix_correlation_id"))
	if corr == "" {
		return nil, fmt.Errorf("fix_correlation_id is empty")
	}
	short := corr
	if len(short) > 8 {
		short = short[:8]
	}

	// ── the staged plan (map-or-string, like every plan consumer) ───────────
	planRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "plan_field", "plan_row.body"))
	if planRaw == nil {
		return nil, fmt.Errorf("staged plan missing at plan_field")
	}
	pb, err := planBytes(planRaw)
	if err != nil {
		return nil, fmt.Errorf("plan not serialisable: %w", err)
	}
	var plan stagedPlan
	if err := json.Unmarshal(pb, &plan); err != nil {
		return nil, fmt.Errorf("plan does not match the staged-plan schema: %w", err)
	}
	if plan.PlanFormat != "staged-v1" || len(plan.Stages) == 0 {
		return nil, fmt.Errorf("feature_stage_route requires a staged-v1 plan with stages (got plan_format %q, %d stages) — single plans belong to fix-implementer", plan.PlanFormat, len(plan.Stages))
	}

	// ── rehydrate prior state, or seed ──────────────────────────────────────
	stateField := datahelpers.GetStringField(config, "state_field", "stage.feature_state")
	var prior *featureLoopState
	if raw := datahelpers.ExtractNestedField(params.CollectedData, stateField); raw != nil {
		b, err := json.Marshal(raw)
		if err == nil {
			var st featureLoopState
			if err := json.Unmarshal(b, &st); err == nil && st.Current > 0 {
				prior = &st
			}
		}
		if prior == nil {
			return nil, fmt.Errorf("feature_state at %q is unreadable — refusing to re-seed mid-run (it would restart the loop from stage 1 on a branch with commits)", stateField)
		}
	}

	if prior == nil {
		branch := datahelpers.GetStringField(config, "branch_prefix", "feat/") + short
		baseRef := datahelpers.GetStringField(config, "base_ref", "main")
		if brf := datahelpers.GetStringField(config, "base_ref_field", "input_data.base_ref"); brf != "" {
			if v := strings.TrimSpace(datahelpers.ExtractNestedFieldString(params.CollectedData, brf)); v != "" {
				baseRef = v
			}
		}
		// E4: a leftover feat/* branch means a prior run's commits would ride
		// into this one from a stale base — fail loudly, never reuse.
		if datahelpers.GetBoolField(config, "require_fresh_branch", true) {
			owner := datahelpers.GetStringField(config, "repo_owner", "gqls")
			repo := datahelpers.GetStringField(config, "repo_name", "agentchassis")
			exists, err := githubBranchExists(ctx, owner, repo, branch)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, fmt.Errorf("branch %s already exists — delete the stale branch before re-firing (E4: silent reuse of an old base is the fix loop's proven trap)", branch)
			}
		}
		prior = &featureLoopState{Current: 0, Branch: branch, BaseRef: baseRef}
	}

	councilRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "council_field", "council_row.body"))

	emit, next := nextStageEmit(plan, *prior, corr, short, councilRaw)
	logger.Info("feature_stage_route",
		zap.String("correlation_id", corr),
		zap.String("branch", next.Branch),
		zap.Int("stage", next.Current),
		zap.Int("stages_total", len(plan.Stages)),
		zap.Bool("stage_ready", emit["stage_ready"] == true))
	return emit, nil
}

// nextStageEmit is the pure loop step: mark the prior stage done (when there
// is one), then emit the next stage — or the terminal PR payload when the
// stages are exhausted. Tested directly.
func nextStageEmit(plan stagedPlan, prior featureLoopState, corr, short string, councilRaw interface{}) (map[string]interface{}, featureLoopState) {
	next := prior
	if next.Current > 0 && next.Current <= len(plan.Stages) {
		done := plan.Stages[next.Current-1]
		next.Done = append(next.Done, doneStageNote{
			ID: done.ID, Title: done.Title, Files: stageRepoFiles(done),
		})
	}
	next.Current++

	stateMap := map[string]interface{}{}
	if b, err := json.Marshal(next); err == nil {
		_ = json.Unmarshal(b, &stateMap)
	}

	if next.Current > len(plan.Stages) {
		title, body := buildStagedPRPayload(corr, short, plan, next.Done, councilRaw)
		return map[string]interface{}{
			"stage_ready":   false,
			"branch":        next.Branch,
			"base_ref":      next.BaseRef,
			"stages_total":  len(plan.Stages),
			"pr_title":      title,
			"pr_body":       body,
			"test_packages": deriveTestPackages(plan),
			"feature_state": stateMap,
		}, next
	}

	st := plan.Stages[next.Current-1]
	stagePlanJSON, _ := json.Marshal(fixPlan{
		Summary:    fmt.Sprintf("%s — stage %s: %s", firstSentence(plan.Summary), st.ID, st.Title),
		Edits:      st.Edits,
		GroundedIn: plan.GroundedIn,
		Risks:      plan.Risks,
	})
	readRef := next.Branch
	if next.Current == 1 {
		readRef = next.BaseRef
	}
	return map[string]interface{}{
		"stage_ready":  true,
		"stage_id":     st.ID,
		"stage_title":  st.Title,
		"stage_goal":   st.Goal,
		"stage_index":  next.Current,
		"stages_total": len(plan.Stages),
		// The single-plan shape the proven read/prepare actions consume — a
		// STRING so planBytes and template rendering see exact JSON.
		"stage_plan":     string(stagePlanJSON),
		"has_repo_edits": len(stageRepoFiles(st)) > 0,
		"read_ref":       readRef,
		"gate_build":     st.Gate.Build == nil || *st.Gate.Build,
		"expected_symbols": append([]string{}, st.ExpectedSymbols...),
		"branch":           next.Branch,
		"base_ref":         next.BaseRef,
		"commit_message": fmt.Sprintf("feat(%s) stage %s: %s\n\nStage %d/%d of a council-approved staged plan (correlation %s).\nHuman review terminal — do not merge without review.",
			short, st.ID, firstSentence(st.Title), next.Current, len(plan.Stages), corr),
		"prior_stages": priorStagesDigest(next.Done),
		"feature_state": stateMap,
	}, next
}

// stageRepoFiles lists a stage's repo-file edits (modify/add) — the files its
// commit will contain; config_change edits target agent rows, not the repo.
func stageRepoFiles(st stagedPlanStage) []string {
	var files []string
	for _, e := range st.Edits {
		op := strings.ToLower(strings.TrimSpace(e.Operation))
		if op == "modify" || op == "add" {
			files = append(files, strings.TrimSpace(e.File))
		}
	}
	return files
}

// priorStagesDigest renders what earlier stages committed, for the implement
// prompt — so stage N wires what stage N-1 created without re-reading blind.
func priorStagesDigest(done []doneStageNote) string {
	if len(done) == 0 {
		return "(none — this is the first stage)"
	}
	var b strings.Builder
	for i, d := range done {
		fmt.Fprintf(&b, "%d. stage %s (%s): committed %s\n", i+1, d.ID, d.Title, strings.Join(d.Files, ", "))
	}
	return b.String()
}

// deriveTestPackages computes the end gate's go-test scope from the plan's
// edited .go files (D6): the union of their packages, sorted, deduplicated.
func deriveTestPackages(plan stagedPlan) []string {
	seen := map[string]bool{}
	for _, st := range plan.Stages {
		for _, f := range stageRepoFiles(st) {
			if !strings.HasSuffix(f, ".go") {
				continue
			}
			if dir := path.Dir(f); dir != "." && dir != "" {
				seen["./"+dir] = true
			}
		}
	}
	pkgs := make([]string, 0, len(seen))
	for p := range seen {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)
	return pkgs
}

// buildStagedPRPayload renders the ONE PR for the whole feature: the stage
// list, the owner's post-merge checklist as an ordered task list (the two
// human acts live IN the PR, where the merge happens), config_change edits
// for the human, and the council decision. Tolerant of absences, like
// buildFixPRBody — the PR is the human surface.
func buildStagedPRPayload(corr, short string, plan stagedPlan, done []doneStageNote, councilRaw interface{}) (string, string) {
	title := fmt.Sprintf("feat(%s): %s", short, firstSentence(plan.Summary))

	var b strings.Builder
	b.WriteString("## Automated feature build (staged plan)\n\n")
	fmt.Fprintf(&b, "Correlation: `%s` — full artifact trail in `diagnosis_artifacts` (staged plan, council reports).\n\n", corr)
	if strings.TrimSpace(plan.Summary) != "" {
		b.WriteString(plan.Summary + "\n\n")
	}

	b.WriteString("### Stages (one commit each, council-approved)\n\n")
	for i, st := range plan.Stages {
		files := stageRepoFiles(st)
		fmt.Fprintf(&b, "%d. **%s** — %s\n   %s\n", i+1, st.ID, st.Title, st.Goal)
		if len(files) > 0 {
			fmt.Fprintf(&b, "   files: `%s`\n", strings.Join(files, "`, `"))
		}
	}
	b.WriteString("\n")

	var configChanges []string
	for _, st := range plan.Stages {
		for _, e := range st.Edits {
			if strings.ToLower(strings.TrimSpace(e.Operation)) == "config_change" {
				configChanges = append(configChanges, fmt.Sprintf("`%s` (stage %s): %s — %s", e.File, st.ID, e.Rationale, e.Sketch))
			}
		}
	}
	if len(configChanges) > 0 {
		b.WriteString("### config_change edits (NOT in this diff — applied by a human)\n\n")
		for _, c := range configChanges {
			b.WriteString("- " + c + "\n")
		}
		b.WriteString("\n")
	}

	if len(plan.PostMergeChecklist) > 0 {
		b.WriteString("### Post-merge checklist (owner acts, in order — the feature is done only after these)\n\n")
		sorted := append([]stagedChecklistEntry{}, plan.PostMergeChecklist...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Order < sorted[j].Order })
		for _, c := range sorted {
			if strings.TrimSpace(c.File) != "" {
				fmt.Fprintf(&b, "- [ ] %d. **%s** `%s` — %s\n", c.Order, c.Act, c.File, c.Detail)
			} else {
				fmt.Fprintf(&b, "- [ ] %d. **%s** — %s\n", c.Order, c.Act, c.Detail)
			}
		}
		b.WriteString("\n")
	}

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

	b.WriteString("---\n_Created by the feature builder (staged plan). Human review terminal: nothing merges itself; the checklist above is the owner's, not the loop's._\n")
	return title, b.String()
}

// githubBranchExists asks the GitHub API whether the branch already exists —
// the E4 freshness check, run once at seed time with the pod-scoped read token.
func githubBranchExists(ctx context.Context, owner, repo, branch string) (bool, error) {
	token := strings.TrimSpace(os.Getenv("GITHUB_READ_TOKEN"))
	if token == "" {
		return false, fmt.Errorf("GITHUB_READ_TOKEN not in env — is this agent on the isRepoCloningAgent spawn gate?")
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", owner, repo, branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("branch existence check: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("branch existence check: HTTP %d for %s", resp.StatusCode, branch)
	}
}
