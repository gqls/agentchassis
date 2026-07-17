package actions

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// The structural gate between the LLM's plan and the artifact F1.1b will act
// on. Shape only — correctness belongs to the council (F2) and the human PR.
func TestValidateFixPlan(t *testing.T) {
	good := fixPlan{
		Summary: "add the missing mark_no_sections step and ground nav in build_status",
		Edits: []fixPlanEdit{
			{File: "platform/orchestration/actions/populate_nav_tables_action.go",
				Symbol: "loadPagesForNav", Operation: "modify",
				Rationale: "nav selects on status, never build_status (diagnosis citation 4)",
				Sketch:    "add AND build_status = 'deployed' to the WHERE clause"},
			{File: "agent_definitions:page-build-handler", Symbol: "check_has_ready_sections",
				Operation: "config_change",
				Rationale: "sectionless pages route to a success terminal",
				Sketch:    "else_step -> mark_no_sections (fail_work_item, needs_human_review)"},
		},
		GroundedIn: []string{"Content writer skipped — page has no sections defined"},
	}
	if p := validateFixPlan(good, 8); len(p) != 0 {
		t.Fatalf("good plan must validate, got: %v", p)
	}

	cases := []struct {
		name   string
		mutate func(fixPlan) fixPlan
		expect string
	}{
		{"no edits", func(p fixPlan) fixPlan { p.Edits = nil; return p }, "no edits"},
		{"ungrounded", func(p fixPlan) fixPlan { p.GroundedIn = nil; return p }, "grounded_in is empty"},
		{"path traversal", func(p fixPlan) fixPlan { p.Edits[0].File = "../../etc/passwd"; return p }, "traversal"},
		{"absolute path", func(p fixPlan) fixPlan { p.Edits[0].File = "/etc/passwd"; return p }, "traversal"},
		{"rogue operation", func(p fixPlan) fixPlan { p.Edits[0].Operation = "execute"; return p }, "allowlist"},
		{"edit cap", func(p fixPlan) fixPlan {
			for i := 0; i < 9; i++ {
				p.Edits = append(p.Edits, p.Edits[0])
			}
			return p
		}, "architecture change"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// deep-ish copy: reslice edits so mutations don't leak between cases
			cp := good
			cp.Edits = append([]fixPlanEdit{}, good.Edits...)
			cp.GroundedIn = append([]string{}, good.GroundedIn...)
			problems := validateFixPlan(tc.mutate(cp), 8)
			if len(problems) == 0 {
				t.Fatal("expected validation failure")
			}
			if !strings.Contains(strings.Join(problems, "; "), tc.expect) {
				t.Fatalf("want %q named, got %v", tc.expect, problems)
			}
		})
	}
}

// The first live proposer run (ed164fed): proposal.result arrived as a raw
// STRING (execute_llm_prompt stores the string when the model JSON is invalid)
// and, separately, was truncated at max_tokens. Both shapes must be handled.
func TestPlanBytes_StringAndMapShapes(t *testing.T) {
	// map shape (happy path) round-trips
	b, err := planBytes(map[string]interface{}{"summary": "s"})
	if err != nil || !strings.Contains(string(b), `"summary"`) {
		t.Fatalf("map shape: %v %s", err, b)
	}
	// string shape with fences is unwrapped
	b, err = planBytes("```json\n{\"summary\":\"s\"}\n```")
	if err != nil || string(b) != `{"summary":"s"}` {
		t.Fatalf("fenced string shape: %v %q", err, b)
	}
	// bare string passes through
	if b, _ = planBytes(`{"summary":"s"}`); string(b) != `{"summary":"s"}` {
		t.Fatalf("bare string shape: %q", b)
	}
}

// ── staged plans (feature builder delta 1 — SCHEMA_staged_plan_v1.md) ────────

func boolPtr(b bool) *bool { return &b }

// goodStagedPlan mirrors the schema doc's worked example (the F1.2 pilot):
// two stages, a to-be-created seed file, and an image-before-seed checklist —
// every new mechanism exercised once.
func goodStagedPlan() stagedPlan {
	return stagedPlan{
		PlanFormat: "staged-v1",
		Summary:    "make the implementer's read ref and base branch per-run inputs",
		GroundedIn: []string{"HANDOFF: ref/base are live-set to a stale branch"},
		Stages: []stagedPlanStage{
			{ID: "s1", Title: "ref/base as inputs",
				Goal: "both actions resolve ref/base from input_data at run time",
				Edits: []fixPlanEdit{
					{File: "platform/orchestration/actions/diagnose_read_repo_files_action.go",
						Symbol: "DiagnoseReadRepoFilesAction", Operation: "modify",
						Rationale: "ref is a config literal; a stale live value reads the wrong tree",
						Sketch:    "resolve ref via the action-input spec (input_data.ref -> config -> 'main')"},
				},
				ExpectedSymbols: []string{"input_data.ref"}},
			{ID: "s2", Title: "re-seed the implementer",
				Goal:      "the live workflow passes input_data.ref/base through",
				DependsOn: []string{"s1"},
				Edits: []fixPlanEdit{
					{File: "docs/fixloop/0NN_fix_implementer_v2_ref_input.sql",
						Operation: "add", ArtifactRole: "seed",
						Rationale: "the workflow config must name the new input fields",
						Sketch:    "v2 seed: read_current_files/prepare gain ref/base field refs"},
				},
				Gate: stagedGate{Build: boolPtr(false)}},
		},
		PostMergeChecklist: []stagedChecklistEntry{
			{Order: 1, Act: "image_deploy", Detail: "make build-agent-chassis-ref; bump IMAGE_TAG; verify pod"},
			{Order: 2, Act: "seed_apply", File: "docs/fixloop/0NN_fix_implementer_v2_ref_input.sql",
				Detail: "apply to clients_db"},
			{Order: 3, Act: "verify", Detail: "fire the implementer with an explicit ref"},
		},
	}
}

var testStagedCaps = stagedPlanCaps{MaxStages: 6, MaxStageEdits: 8, MaxTotalEdits: 24}

func TestValidateStagedPlan(t *testing.T) {
	if p := validateStagedPlan(goodStagedPlan(), false, testStagedCaps); len(p) != 0 {
		t.Fatalf("good staged plan must validate, got: %v", p)
	}

	cases := []struct {
		name   string
		mutate func(stagedPlan) stagedPlan
		expect string
	}{
		{"wrong plan_format", func(p stagedPlan) stagedPlan { p.PlanFormat = "staged-v2"; return p },
			"plan_format must be"},
		{"no stages", func(p stagedPlan) stagedPlan { p.Stages = nil; return p }, "no stages"},
		{"ungrounded", func(p stagedPlan) stagedPlan { p.GroundedIn = nil; return p }, "grounded_in is empty"},
		{"stage cap", func(p stagedPlan) stagedPlan {
			for i := 0; i < 7; i++ {
				st := p.Stages[0]
				st.ID = fmt.Sprintf("x%d", i)
				st.DependsOn = nil
				p.Stages = append(p.Stages, st)
			}
			return p
		}, "stages exceeds cap"},
		{"per-stage edit cap", func(p stagedPlan) stagedPlan {
			for i := 0; i < 9; i++ {
				e := p.Stages[0].Edits[0]
				e.File = fmt.Sprintf("a/b%d.go", i)
				p.Stages[0].Edits = append(p.Stages[0].Edits, e)
			}
			return p
		}, "per-stage cap"},
		{"bad stage id", func(p stagedPlan) stagedPlan { p.Stages[0].ID = "Stage One!"; return p }, "must match"},
		{"duplicate stage id", func(p stagedPlan) stagedPlan { p.Stages[1].ID = "s1"; p.Stages[1].DependsOn = nil; return p },
			"already used"},
		{"empty goal", func(p stagedPlan) stagedPlan { p.Stages[0].Goal = " "; return p }, "goal is empty"},
		{"unknown dep", func(p stagedPlan) stagedPlan { p.Stages[1].DependsOn = []string{"s9"}; return p },
			"unknown stage"},
		{"forward dep", func(p stagedPlan) stagedPlan { p.Stages[0].DependsOn = []string{"s2"}; return p },
			"strictly earlier"},
		{"self dep", func(p stagedPlan) stagedPlan { p.Stages[0].DependsOn = []string{"s1"}; return p },
			"strictly earlier"},
		{"added twice", func(p stagedPlan) stagedPlan {
			e := p.Stages[1].Edits[0]
			p.Stages[0].Edits = append(p.Stages[0].Edits, e)
			return p
		}, "added more than once"},
		{"modify before add", func(p stagedPlan) stagedPlan {
			p.Stages[0].Edits = append(p.Stages[0].Edits, fixPlanEdit{
				File: p.Stages[1].Edits[0].File, Operation: "modify",
				Rationale: "r", Sketch: "wire the new seed in"})
			return p
		}, "before the stage that adds it"},
		{"create then delete", func(p stagedPlan) stagedPlan {
			p.Stages[1].Edits = append(p.Stages[1].Edits, fixPlanEdit{
				File: p.Stages[1].Edits[0].File, Operation: "remove",
				Rationale: "r", Sketch: "drop it again", ArtifactRole: "seed"})
			return p
		}, "create-then-delete"},
		{"rogue artifact_role", func(p stagedPlan) stagedPlan { p.Stages[0].Edits[0].ArtifactRole = "binary"; return p },
			"artifact_role"},
		{"config_change as seed", func(p stagedPlan) stagedPlan {
			p.Stages[0].Edits[0].Operation = "config_change"
			p.Stages[0].Edits[0].ArtifactRole = "seed"
			return p
		}, "contradictory"},
		{"build off for code stage", func(p stagedPlan) stagedPlan { p.Stages[0].Gate.Build = boolPtr(false); return p },
			"only all-seed/doc stages may skip the build gate"},
		{"empty expected symbol", func(p stagedPlan) stagedPlan {
			p.Stages[0].ExpectedSymbols = append(p.Stages[0].ExpectedSymbols, " ")
			return p
		}, "expected_symbols contains an empty string"},
		{"seed without checklist", func(p stagedPlan) stagedPlan { p.PostMergeChecklist = nil; return p },
			"post_merge_checklist is required"},
		{"seed uncovered", func(p stagedPlan) stagedPlan {
			p.PostMergeChecklist = []stagedChecklistEntry{
				{Order: 1, Act: "image_deploy", Detail: "d"}}
			return p
		}, "no seed_apply checklist entry"},
		{"seed_apply names unshipped file", func(p stagedPlan) stagedPlan {
			p.PostMergeChecklist[1].File = "docs/fixloop/other.sql"
			return p
		}, "no seed edit ships"},
		{"seed before image", func(p stagedPlan) stagedPlan {
			p.PostMergeChecklist[0].Order = 5 // image now after the seed at 2
			return p
		}, "image_deploy must come strictly before any seed_apply"},
		{"code+seed without image_deploy", func(p stagedPlan) stagedPlan {
			p.PostMergeChecklist = []stagedChecklistEntry{
				{Order: 1, Act: "seed_apply", File: "docs/fixloop/0NN_fix_implementer_v2_ref_input.sql", Detail: "apply"},
				{Order: 2, Act: "verify", Detail: "check"},
			}
			return p
		}, "plan ships code and a seed"},
		{"duplicate order", func(p stagedPlan) stagedPlan { p.PostMergeChecklist[2].Order = 2; return p },
			"already used"},
		{"rogue act", func(p stagedPlan) stagedPlan {
			p.PostMergeChecklist = append(p.PostMergeChecklist,
				stagedChecklistEntry{Order: 4, Act: "self_merge", Detail: "d"})
			return p
		}, "act \"self_merge\" not in the allowlist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problems := validateStagedPlan(tc.mutate(goodStagedPlan()), false, testStagedCaps)
			if len(problems) == 0 {
				t.Fatal("expected validation failure")
			}
			if !strings.Contains(strings.Join(problems, "; "), tc.expect) {
				t.Fatalf("want %q named, got %v", tc.expect, problems)
			}
		})
	}

	// top-level edits are the legacy shape leaking into a staged plan
	if problems := validateStagedPlan(goodStagedPlan(), true, testStagedCaps); len(problems) == 0 ||
		!strings.Contains(strings.Join(problems, "; "), "top-level edits") {
		t.Fatalf("top-level edits must be rejected, got %v", problems)
	}
}

// D4 refinement (found live by the F1.2 pilot, run bcc96877): a SEED-ONLY
// plan ships no code, so its truthful checklist is seed_apply→verify with no
// image_deploy — validation must accept it. Ordering is still enforced if an
// image_deploy entry is present anyway.
func TestValidateStagedPlan_SeedOnlyNeedsNoImageDeploy(t *testing.T) {
	p := goodStagedPlan()
	p.Stages = p.Stages[1:] // drop the code stage; the seed stage remains
	p.Stages[0].DependsOn = nil
	p.PostMergeChecklist = []stagedChecklistEntry{
		{Order: 1, Act: "seed_apply", File: "docs/fixloop/0NN_fix_implementer_v2_ref_input.sql", Detail: "apply to clients_db"},
		{Order: 2, Act: "verify", Detail: "fire with an explicit ref and confirm"},
	}
	if problems := validateStagedPlan(p, false, testStagedCaps); len(problems) != 0 {
		t.Fatalf("seed-only plan without image_deploy must validate, got: %v", problems)
	}
	// an image_deploy AFTER the seed is still a lie about ordering
	p.PostMergeChecklist = append(p.PostMergeChecklist,
		stagedChecklistEntry{Order: 3, Act: "image_deploy", Detail: "late"})
	if problems := validateStagedPlan(p, false, testStagedCaps); len(problems) == 0 ||
		!strings.Contains(strings.Join(problems, "; "), "strictly before") {
		t.Fatalf("misordered image_deploy must still fail, got: %v", problems)
	}
}

// The discriminator must never mistake one shape for the other: a legacy plan
// probes legacy, a staged plan (with or without plan_format) probes staged,
// and invalid JSON probes legacy so the existing truncation hint still fires.
func TestProbePlan_Discriminator(t *testing.T) {
	legacy := []byte(`{"summary":"s","edits":[{"file":"a.go"}],"grounded_in":["q"]}`)
	if p := probePlan(legacy); p.PlanFormat != "" || rawPresent(p.Stages) {
		t.Fatal("legacy plan must not probe as staged")
	}
	stagedByFormat := []byte(`{"plan_format":"staged-v1","summary":"s"}`)
	if p := probePlan(stagedByFormat); p.PlanFormat == "" {
		t.Fatal("plan_format must probe as staged")
	}
	stagedByStages := []byte(`{"summary":"s","stages":[{"id":"s1"}]}`)
	if p := probePlan(stagedByStages); !rawPresent(p.Stages) {
		t.Fatal("a stages key must probe as staged")
	}
	truncated := []byte(`{"plan_format":"staged-v1","stages":[{"id":`)
	p := probePlan(truncated) // must not panic; field values are best-effort
	_ = p
	if rawPresent(json.RawMessage(`null`)) || rawPresent(json.RawMessage(``)) {
		t.Fatal("null/absent must not read as present")
	}
}

// F1.1b(a): the first live plan's two semantic no-ops must now be rejected.
func TestValidateFixPlan_RejectsNoOpEdits(t *testing.T) {
	base := fixPlan{
		Summary:    "s",
		GroundedIn: []string{"quote"},
		Edits: []fixPlanEdit{{
			File: "a/b.go", Operation: "modify", Rationale: "r",
			Sketch: "add AND build_status = 'deployed' to the WHERE clause",
		}},
	}
	if p := validateFixPlan(base, 8); len(p) != 0 {
		t.Fatalf("real edit must pass: %v", p)
	}
	for _, sketch := range []string{
		"No code change required in applyAddToPage itself — the function is correct",
		"In the block comment above childPrefixes, add a clarifying note: 'section-index pages are exempt'",
		"Audit the branch; no change is needed here for the nav",
	} {
		p := base
		p.Edits = []fixPlanEdit{{File: "a/b.go", Operation: "modify", Rationale: "r", Sketch: sketch}}
		if problems := validateFixPlan(p, 8); len(problems) == 0 {
			t.Fatalf("no-op sketch must be rejected: %q", sketch)
		}
	}
}
