// FILE: cmd/config-key-audit/loopitemkeys_test.go
//
// bugs_open/321. The positive fixture is the REAL pre-migration-493 shape of
// tool-suggester's create_items_loop — the exact configuration that lost 72% of
// its suggestions in production — not an invented one, so the check is proven by
// FIRING against the state it was written to catch (the WFA-007 convention:
// proven by firing against the pre-fix config, not by passing). The negative
// fixtures are the shapes the rule must acquit: tool-auditor's suffixed steps,
// top-level site-wide steps, loops over sites, and sub_workflows under non-loop
// parents that never execute.
package main

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"

	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
)

// The live tool-suggester loop as it stood before migration 493 (prefixes and
// step names verbatim from agent_definitions, 2026-08-19). Two findings.
const preFix493ToolSuggester = `[
	{"type": "tool-suggester", "workflow": {"start_step": "ensure_site_record", "steps": {
		"suggest_tools": {"action": "execute_llm_prompt", "next_step": "create_items_loop"},
		"create_items_loop": {"action": "loop", "next_step": "complete", "config": {
			"items_field": "evaluation.result.suggestions",
			"item_variable": "current_suggestion",
			"max_iterations": 10,
			"sub_workflow": {"start_step": "check_is_library", "steps": {
				"check_is_library": {"action": "conditional", "config": {
					"condition": "current_suggestion.tool_component_id != null",
					"then_step": "create_library_item", "else_step": "create_novel_item"}},
				"create_novel_item": {"action": "create_work_item", "next_step": "done", "config": {
					"item_type": "add_tool", "item_key_prefix": "add_tool_novel",
					"site_id": "input_data.site_id", "handler_agent": "tool-generator",
					"spec_data": "current_suggestion"}},
				"create_library_item": {"action": "create_work_item", "next_step": "done", "config": {
					"item_type": "add_tool", "item_key_prefix": "add_tool",
					"site_id": "input_data.site_id", "handler_agent": "tool-deployer",
					"spec_data": "current_suggestion"}},
				"done": {"action": "loop_complete"}
			}}
		}}
	}}}
]`

func TestLoopItemKeys_FiresOnThePreFixToolSuggester(t *testing.T) {
	agents, failed, err := decodeLiveAgents([]byte(preFix493ToolSuggester), "test")
	if err != nil || failed != 0 {
		t.Fatalf("decode: err=%v failed=%d", err, failed)
	}
	findings := findLoopSitewideItemKeys(agents)
	if len(findings) != 2 {
		t.Fatalf("expected exactly 2 findings on the pre-493 config, got %d: %+v", len(findings), findings)
	}
	// Sorted by path: create_library_item before create_novel_item.
	if findings[0].ItemKeyPrefix != "add_tool" || findings[1].ItemKeyPrefix != "add_tool_novel" {
		t.Errorf("prefixes = %q, %q", findings[0].ItemKeyPrefix, findings[1].ItemKeyPrefix)
	}
	for _, f := range findings {
		if f.LoopVariable != "current_suggestion" {
			t.Errorf("%s: loop_variable = %q, want current_suggestion (the fix hint)", f.Path, f.LoopVariable)
		}
		if f.LoopPath != "steps.create_items_loop" {
			t.Errorf("%s: loop_path = %q", f.Path, f.LoopPath)
		}
		if f.SuffixDeclaredButUnhonoured {
			t.Errorf("%s: suffix was never declared, flag must be false", f.Path)
		}
	}
}

// tool-auditor's shape: suffix present and honoured — the proven idiom, zero
// findings. This is also the post-493 shape of all four fixed steps.
const suffixedLoop = `[
	{"type": "tool-auditor", "workflow": {"start_step": "audit", "steps": {
		"findings_loop": {"action": "loop", "config": {
			"items_field": "audit_result.result.findings",
			"item_variable": "current_finding",
			"sub_workflow": {"start_step": "create_improve_item", "steps": {
				"create_improve_item": {"action": "create_work_item", "config": {
					"item_type": "improve_tool", "item_key_prefix": "audit_fix",
					"item_key_suffix_field": "tool_data.page_id",
					"site_id": "input_data.site_id"}}
			}}
		}}
	}}}
]`

func TestLoopItemKeys_AcquitsAnHonouredSuffix(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(suffixedLoop), "test")
	if err != nil {
		t.Fatal(err)
	}
	if findings := findLoopSitewideItemKeys(agents); len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %+v", findings)
	}
}

// A suffix DECLARED but not honoured — empty string, or a non-string — is
// silently ignored by the action's `f, ok := ...(string); ok && f != ""` read,
// which reverts to the site-wide key with no error. Both must be findings, and
// both must carry the flag that tells the reader the author believed the door
// was closed.
const unhonouredSuffixes = `[
	{"type": "empty-suffix", "workflow": {"start_step": "l", "steps": {
		"l": {"action": "loop", "config": {
			"items_field": "x.items", "item_variable": "it",
			"sub_workflow": {"start_step": "c", "steps": {
				"c": {"action": "create_work_item", "config": {
					"item_key_prefix": "p", "item_key_suffix_field": "",
					"site_id": "input_data.site_id"}}
			}}
		}}
	}}},
	{"type": "boolean-suffix", "workflow": {"start_step": "l", "steps": {
		"l": {"action": "loop", "config": {
			"items_field": "x.items", "item_variable": "it",
			"sub_workflow": {"start_step": "c", "steps": {
				"c": {"action": "create_work_item", "config": {
					"item_key_prefix": "p", "item_key_suffix_field": true,
					"site_id": "input_data.site_id"}}
			}}
		}}
	}}}
]`

func TestLoopItemKeys_ConvictsADeclaredButUnhonouredSuffix(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(unhonouredSuffixes), "test")
	if err != nil {
		t.Fatal(err)
	}
	findings := findLoopSitewideItemKeys(agents)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(findings), findings)
	}
	for _, f := range findings {
		if !f.SuffixDeclaredButUnhonoured {
			t.Errorf("%s/%s: suffix_declared_but_unhonoured must be true", f.Agent, f.Path)
		}
	}
}

// The acquittals that keep the once-per-site fleet out of the report:
//   - a TOP-LEVEL step with a prefix (site-wide key is the intended dedupe);
//   - a nested step with NO prefix (item_key NULL, outside idx_swi_dedup);
//   - a loop over SITES (site_id rooted at the loop variable — distinct
//     site_id per iteration, no collision possible);
//   - a sub_workflow under a NON-LOOP parent (never executes).
const acquittals = `[
	{"type": "top-level-site-wide", "workflow": {"start_step": "c", "steps": {
		"c": {"action": "create_work_item", "config": {
			"item_key_prefix": "site_plan", "site_id": "input_data.site_id"}}
	}}},
	{"type": "no-prefix", "workflow": {"start_step": "l", "steps": {
		"l": {"action": "loop", "config": {
			"items_field": "x.items", "item_variable": "it",
			"sub_workflow": {"start_step": "c", "steps": {
				"c": {"action": "create_work_item", "config": {
					"site_id": "input_data.site_id"}}
			}}
		}}
	}}},
	{"type": "loop-over-sites", "workflow": {"start_step": "l", "steps": {
		"l": {"action": "loop", "config": {
			"items_field": "sites.rows", "item_variable": "current_site",
			"sub_workflow": {"start_step": "c", "steps": {
				"c": {"action": "create_work_item", "config": {
					"item_key_prefix": "sweep", "site_id": "current_site.id"}}
			}}
		}}
	}}},
	{"type": "non-loop-parent", "workflow": {"start_step": "outer", "steps": {
		"outer": {"action": "conditional", "config": {
			"sub_workflow": {"start_step": "c", "steps": {
				"c": {"action": "create_work_item", "config": {
					"item_key_prefix": "dead", "site_id": "input_data.site_id"}}
			}}
		}}
	}}}
]`

func TestLoopItemKeys_AcquittalsStayQuiet(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(acquittals), "test")
	if err != nil {
		t.Fatal(err)
	}
	if findings := findLoopSitewideItemKeys(agents); len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %+v", findings)
	}
}

// substeps WINS over sub_workflow at execution (loop_actions.go:91 reads it
// first; validation.subWorkflowsOf mirrors that precedence — sharedoutputs.go's
// header records this package going blind on exactly this). A loop carrying a
// CLEAN sub_workflow and a DEFECTIVE substeps body must be convicted on the
// substeps half — the one that runs.
const bothShapes = `[
	{"type": "both-shapes", "workflow": {"start_step": "l", "steps": {
		"l": {"action": "loop", "config": {
			"items_field": "x.items", "item_variable": "it",
			"substeps": {
				"c": {"action": "create_work_item", "config": {
					"item_key_prefix": "executed", "site_id": "input_data.site_id"}}
			},
			"sub_workflow": {"start_step": "c", "steps": {
				"c": {"action": "create_work_item", "config": {
					"item_key_prefix": "inert", "item_key_suffix_field": "it.id",
					"site_id": "input_data.site_id"}}
			}}
		}}
	}}}
]`

func TestLoopItemKeys_JudgesTheExecutedShape(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(bothShapes), "test")
	if err != nil {
		t.Fatal(err)
	}
	findings := findLoopSitewideItemKeys(agents)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (the executed substeps half), got %d: %+v", len(findings), findings)
	}
	if findings[0].ItemKeyPrefix != "executed" {
		t.Errorf("convicted prefix = %q — the check judged the inert sub_workflow half", findings[0].ItemKeyPrefix)
	}
}

// A loop nested in a loop: the inner create_work_item's parent is the INNER
// loop, and the finding must name it.
const loopInLoop = `[
	{"type": "nested-loops", "workflow": {"start_step": "outer", "steps": {
		"outer": {"action": "loop", "config": {
			"items_field": "a.items", "item_variable": "oa",
			"sub_workflow": {"start_step": "inner", "steps": {
				"inner": {"action": "loop", "config": {
					"items_field": "oa.parts", "item_variable": "part",
					"sub_workflow": {"start_step": "c", "steps": {
						"c": {"action": "create_work_item", "config": {
							"item_key_prefix": "deep", "site_id": "input_data.site_id"}}
					}}
				}}
			}}
		}}
	}}}
]`

func TestLoopItemKeys_LoopInLoopNamesTheInnerLoop(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(loopInLoop), "test")
	if err != nil {
		t.Fatal(err)
	}
	findings := findLoopSitewideItemKeys(agents)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
	if findings[0].LoopPath != "steps.outer.sub_workflow.inner" {
		t.Errorf("loop_path = %q, want steps.outer.sub_workflow.inner", findings[0].LoopPath)
	}
	if findings[0].LoopVariable != "part" {
		t.Errorf("loop_variable = %q, want part (the inner loop's)", findings[0].LoopVariable)
	}
}

// Registry pinning (the parity property, without a second implementation): the
// rule above hard-codes the strings "create_work_item", "item_key_prefix" and
// "item_key_suffix_field". If the action is renamed, retired, or stops
// declaring either key, this test fails loudly — instead of the detector
// counting zero for ever (the two-actions-counted-as-ZERO failure RFC_022's
// cron walked into).
func TestLoopItemKeys_RegistryStillMatchesTheHardcodedRule(t *testing.T) {
	if !actioncheck.IsLocalAction("create_work_item") {
		t.Fatal("create_work_item is no longer a registered local action — the detector's rule is auditing a ghost")
	}
	spec, ok := datahelpers.GetActionInputSpec("create_work_item")
	if !ok {
		t.Fatal("create_work_item no longer declares an ActionInputSpec")
	}
	found := map[string]bool{}
	for _, k := range spec.ConfigKeys {
		found[k] = true
	}
	for _, want := range []string{"item_key_prefix", "item_key_suffix_field"} {
		if !found[want] {
			t.Errorf("create_work_item's ConfigKeys no longer declare %q — the detector reads a key the action has disowned", want)
		}
	}
}
