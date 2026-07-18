package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

// The stage loop's pure core: seed → stage 1 (reads base), advance → stage 2
// (reads the branch, digest carries stage 1), advance past the end → PR
// payload + derived test packages. The k8s/HTTP edges (branch freshness) are
// exercised live, like the build gate's Job machinery.

func routeTestPlan() stagedPlan {
	p := goodStagedPlan() // the F1.2 pilot shape from the validation tests
	if probs := validateStagedPlan(p, false, testStagedCaps); len(probs) != 0 {
		panic("route tests need a valid plan: " + strings.Join(probs, "; "))
	}
	return p
}

func TestNextStageEmit_SeedReadsBase(t *testing.T) {
	plan := routeTestPlan()
	seed := featureLoopState{Current: 0, Branch: "feat/ab12cd34", BaseRef: "main"}
	emit, next := nextStageEmit(plan, seed, "ab12cd34-full-corr", "ab12cd34", nil)

	if emit["stage_ready"] != true || emit["stage_id"] != "s1" {
		t.Fatalf("seed must emit stage 1 ready: %v", emit)
	}
	if emit["read_ref"] != "main" {
		t.Fatalf("stage 1 must read the BASE ref, got %v", emit["read_ref"])
	}
	if emit["gate_build"] != true || emit["has_repo_edits"] != true {
		t.Fatalf("stage 1 flags wrong: %v", emit)
	}
	if next.Current != 1 || len(next.Done) != 0 {
		t.Fatalf("seed state wrong: %+v", next)
	}
	// The emitted stage_plan is a STRING carrying a valid single-plan shape
	// with ONLY this stage's edits — the contract read/prepare rely on.
	var sp fixPlan
	if err := json.Unmarshal([]byte(emit["stage_plan"].(string)), &sp); err != nil {
		t.Fatalf("stage_plan not a single-plan JSON string: %v", err)
	}
	if len(sp.Edits) != 1 || sp.Edits[0].File != plan.Stages[0].Edits[0].File {
		t.Fatalf("stage_plan must carry exactly stage 1's edits: %+v", sp.Edits)
	}
	if msg := emit["commit_message"].(string); !strings.Contains(msg, "feat(ab12cd34) stage s1:") ||
		!strings.Contains(msg, "Stage 1/2") {
		t.Fatalf("commit message wrong: %q", msg)
	}
}

func TestNextStageEmit_AdvanceReadsBranchAndCarriesDigest(t *testing.T) {
	plan := routeTestPlan()
	afterStage1 := featureLoopState{Current: 1, Branch: "feat/ab12cd34", BaseRef: "main"}
	emit, next := nextStageEmit(plan, afterStage1, "ab12cd34-full-corr", "ab12cd34", nil)

	if emit["stage_id"] != "s2" || emit["stage_ready"] != true {
		t.Fatalf("advance must emit stage 2: %v", emit)
	}
	if emit["read_ref"] != "feat/ab12cd34" {
		t.Fatalf("stage 2 must read the BRANCH (to see stage 1's commit), got %v", emit["read_ref"])
	}
	if emit["gate_build"] != false {
		t.Fatalf("stage 2 (all-seed) must skip the build gate: %v", emit)
	}
	if len(next.Done) != 1 || next.Done[0].ID != "s1" {
		t.Fatalf("stage 1 must be marked done: %+v", next.Done)
	}
	if digest := emit["prior_stages"].(string); !strings.Contains(digest, "stage s1") ||
		!strings.Contains(digest, plan.Stages[0].Edits[0].File) {
		t.Fatalf("prior-stages digest must name stage 1 and its files: %q", digest)
	}
}

func TestNextStageEmit_TerminalPRPayload(t *testing.T) {
	plan := routeTestPlan()
	afterStage2 := featureLoopState{
		Current: 2, Branch: "feat/ab12cd34", BaseRef: "main",
		Done: []doneStageNote{{ID: "s1", Title: "ref/base as inputs",
			Files: []string{plan.Stages[0].Edits[0].File}}},
	}
	council := map[string]interface{}{"decision": "approved"}
	emit, next := nextStageEmit(plan, afterStage2, "ab12cd34-full-corr", "ab12cd34", council)

	if emit["stage_ready"] != false {
		t.Fatalf("past the last stage must emit terminal: %v", emit)
	}
	if next.Current != 3 || len(next.Done) != 2 {
		t.Fatalf("terminal state wrong: %+v", next)
	}
	title := emit["pr_title"].(string)
	if !strings.HasPrefix(title, "feat(ab12cd34): ") {
		t.Fatalf("pr title wrong: %q", title)
	}
	body := emit["pr_body"].(string)
	for _, want := range []string{
		"### Stages",
		"**s1**", "**s2**",
		"### Post-merge checklist",
		"- [ ] 1. **image_deploy**",
		"- [ ] 2. **seed_apply** `docs/fixloop/0NN_fix_implementer_v2_ref_input.sql`",
		"### Council decision",
		"nothing merges itself",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("pr body missing %q:\n%s", want, body)
		}
	}
	// D6: test packages derived from the plan's .go edits — the .sql add
	// contributes nothing.
	pkgs, _ := emit["test_packages"].([]string)
	if len(pkgs) != 1 || pkgs[0] != "./platform/orchestration/actions" {
		t.Fatalf("derived test packages wrong: %v", pkgs)
	}
}

func TestDeriveTestPackages_SortedDeduped(t *testing.T) {
	plan := stagedPlan{Stages: []stagedPlanStage{
		{Edits: []fixPlanEdit{
			{File: "platform/b/x.go", Operation: "modify"},
			{File: "platform/a/y.go", Operation: "add"},
			{File: "docs/seed.sql", Operation: "add"},
		}},
		{Edits: []fixPlanEdit{
			{File: "platform/b/z.go", Operation: "modify"},
			{File: "agent_definitions:some-agent", Operation: "config_change"},
		}},
	}}
	pkgs := deriveTestPackages(plan)
	if len(pkgs) != 2 || pkgs[0] != "./platform/a" || pkgs[1] != "./platform/b" {
		t.Fatalf("want [./platform/a ./platform/b], got %v", pkgs)
	}
}

func TestStageRepoFiles_ConfigChangeExcluded(t *testing.T) {
	st := stagedPlanStage{Edits: []fixPlanEdit{
		{File: "a/b.go", Operation: "modify"},
		{File: "agent_definitions:x", Operation: "config_change"},
	}}
	files := stageRepoFiles(st)
	if len(files) != 1 || files[0] != "a/b.go" {
		t.Fatalf("config_change must not surface as a repo file: %v", files)
	}
}

// Council-gate review 5a65ec4c (bug-historian, high): a CONFIGURED routed
// field that resolves empty must FAIL, never silently fall back — falling back
// reads the wrong tree / commits to the wrong branch / gates on less than
// intended. Unset stays the single-plan path. These lock the distinction that
// the three action-level guards enforce.
func TestRoutedFieldEmptyIsAnError_Contract(t *testing.T) {
	// The guards live in the actions (which need DB/k8s), so this test locks the
	// SHAPE the loop depends on: the router always emits non-empty values for
	// every field the stage workflow configures. If this ever emits "", the
	// action-level guards are what turn it into a loud failure rather than a
	// wrong-tree read.
	plan := routeTestPlan()
	seed := featureLoopState{Current: 0, Branch: "feat/ab12cd34", BaseRef: "main"}
	emit, next := nextStageEmit(plan, seed, "ab12cd34-full-corr", "ab12cd34", nil)
	for _, k := range []string{"read_ref", "branch", "commit_message"} {
		if v, _ := emit[k].(string); strings.TrimSpace(v) == "" {
			t.Fatalf("router emitted empty %q — the action guards would (correctly) fail the run", k)
		}
	}
	// ...and on the advance path too, where the branch matters most.
	emit2, _ := nextStageEmit(plan, next, "ab12cd34-full-corr", "ab12cd34", nil)
	if v, _ := emit2["branch"].(string); strings.TrimSpace(v) == "" {
		t.Fatal("router emitted empty branch on advance — stage 2 would commit to the wrong branch")
	}
	// Terminal emission must carry a non-empty derived package list whenever the
	// plan has .go edits, else the end gate is configured-but-empty (an error).
	term, _ := nextStageEmit(plan, featureLoopState{Current: 2, Branch: "feat/ab12cd34", BaseRef: "main"},
		"ab12cd34-full-corr", "ab12cd34", nil)
	if pkgs, _ := term["test_packages"].([]string); len(pkgs) == 0 {
		t.Fatal("terminal emitted no test packages for a plan containing .go edits")
	}
}

// Delta-2 prepare additions: expected symbols must appear in a produced body.
func TestMissingExpectedSymbols(t *testing.T) {
	files := map[string]interface{}{
		"a/b.go": map[string]interface{}{"content": "func ResolveRef() {}", "encoding": "utf-8"},
	}
	if m := missingExpectedSymbols([]string{"ResolveRef"}, files); len(m) != 0 {
		t.Fatalf("present symbol reported missing: %v", m)
	}
	if m := missingExpectedSymbols([]string{"ResolveRef", "NotThere"}, files); len(m) != 1 || m[0] != "NotThere" {
		t.Fatalf("want [NotThere], got %v", m)
	}
	if m := missingExpectedSymbols([]string{" ", ""}, files); len(m) != 0 {
		t.Fatalf("blank symbols must be ignored: %v", m)
	}
}
