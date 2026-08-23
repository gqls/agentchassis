// FILE: cmd/config-key-audit/recurrencedeclared_test.go
//
// bugs_open/326. The positive fixture is the REAL pre-migration-572 shape of
// domain-submitter's create_research_item — the exact configuration that made a
// re-submitted customer build report COMPLETED and queue nothing — not an
// invented one, so the check is proven by FIRING against the state it was
// written to catch (the WFA-007 convention: proven by firing against the pre-fix
// config, not by passing).
//
// The negative fixtures are the shapes the rule must ACQUIT, and the second of
// them is the one that matters most: an explicit `false` is CLEAN. This check
// reports a missing declaration, never a wrong answer — a step that has decided
// it IS a detector has satisfied it. claims-auditor is the live example, and
// setting it `true` would break its revalidator-close loop.
package main

import (
	"testing"
)

// domain-submitter as it stood before migration 572 (step name, prefix and
// item_type verbatim from agent_definitions, 2026-08-23). One finding.
const preFix572DomainSubmitter = `[
	{"type": "domain-submitter", "workflow": {"start_step": "ensure_site_record", "steps": {
		"ensure_site_record": {"action": "upsert_site", "next_step": "create_research_item"},
		"create_research_item": {"action": "create_work_item", "next_step": "complete", "config": {
			"source": "domain-submitter",
			"site_id": "site_record.site_id",
			"summary": "Research and classify domain",
			"priority": 5,
			"severity": "high",
			"item_type": "needs_domain_research",
			"handler_agent": "domain-research-classifier",
			"item_pipeline": "build",
			"item_key_prefix": "research"}},
		"complete": {"action": "complete_workflow"}
	}}}
]`

func TestUndeclaredRecurrence_FiresOnThePreFixDomainSubmitter(t *testing.T) {
	agents, failed, err := decodeLiveAgents([]byte(preFix572DomainSubmitter), "test")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if failed != 0 {
		t.Fatalf("%d agent(s) failed to decode", failed)
	}

	findings := findUndeclaredRecurrence(agents)
	if len(findings) != 1 {
		t.Fatalf("want exactly 1 finding on the pre-572 front door, got %d: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Agent != "domain-submitter" || f.Path != "steps.create_research_item" {
		t.Errorf("finding names the wrong step: %+v", f)
	}
	if f.ItemKeyPrefix != "research" || f.ItemType != "needs_domain_research" {
		t.Errorf("finding must carry the routing hints a reviewer needs: %+v", f)
	}
	if f.DeclaredUnhonoured {
		t.Error("nothing was declared at all, so declared_unhonoured must be false — " +
			"conflating 'never said' with 'said it wrong' sends the reader to the wrong fix")
	}
}

// The three shapes the rule must acquit, each for a different reason.
func TestUndeclaredRecurrence_AcquitsWhatItMust(t *testing.T) {
	for _, tc := range []struct {
		name string
		json string
		why  string
	}{
		{
			name: "explicit true",
			why:  "declared as an action request — migration 572's post-state",
			json: `[{"type": "domain-submitter", "workflow": {"start_step": "s", "steps": {
				"s": {"action": "create_work_item", "config": {
					"item_type": "needs_domain_research", "item_key_prefix": "research",
					"recurrence_expected": true}}}}}]`,
		},
		{
			name: "explicit false",
			why: "declared as a DETECTED DEFECT. claims-auditor is the live case: its " +
				"revalidator-close loop writes 'complete' into the two-strike window by " +
				"design, so it NEEDS the counter. A check that reported this would be " +
				"pushing every reader toward breaking it.",
			json: `[{"type": "claims-auditor", "workflow": {"start_step": "s", "steps": {
				"s": {"action": "create_work_item", "config": {
					"item_type": "claims_unverified", "item_key_prefix": "claims_llm",
					"recurrence_expected": false}}}}}]`,
		},
		{
			name: "no item_key_prefix",
			why: "no prefix -> item_key is NULL -> the row sits outside idx_swi_dedup AND " +
				"writeWorkItem's brake is gated on itemKey != \"\", so the brake cannot " +
				"fire and there is nothing to declare. grounded-explainer is the live case.",
			json: `[{"type": "grounded-explainer", "workflow": {"start_step": "s", "steps": {
				"s": {"action": "create_work_item", "config": {
					"item_type": "grounded_draft_review"}}}}}]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			agents, _, err := decodeLiveAgents([]byte(tc.json), "test")
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := findUndeclaredRecurrence(agents); len(got) != 0 {
				t.Errorf("must acquit (%s), got %+v", tc.why, got)
			}
		})
	}
}

// A declaration the ACTION CANNOT READ is worse than no declaration, because the
// author believes the door is shut. create_work_item reads the key with a bare
// `config["recurrence_expected"].(bool)`, so a JSON string "true" is silently
// ignored and the brake runs anyway.
//
// This is the same shape --loop-sitewide-item-keys reports for its suffix field
// (suffix_declared_but_unhonoured), and it is reported here as a finding with
// declared_unhonoured=true so the two cases are distinguishable in the output.
func TestUndeclaredRecurrence_ConvictsADeclarationTheActionCannotRead(t *testing.T) {
	const stringNotBool = `[{"type": "some-agent", "workflow": {"start_step": "s", "steps": {
		"s": {"action": "create_work_item", "config": {
			"item_type": "needs_rerender", "item_key_prefix": "rerender",
			"recurrence_expected": "true"}}}}}]`

	agents, _, err := decodeLiveAgents([]byte(stringNotBool), "test")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	findings := findUndeclaredRecurrence(agents)
	if len(findings) != 1 {
		t.Fatalf(`want 1 finding for a string "true" the action ignores, got %d`, len(findings))
	}
	if !findings[0].DeclaredUnhonoured {
		t.Error("declared_unhonoured must be true: the author opted in and the runtime disagreed, " +
			"which needs a different fix from never having declared")
	}
}

// Nested steps are IN scope, unlike --loop-sitewide-item-keys, whose defect is
// specifically about loop nesting. The brake fires wherever a keyed item is
// written, so a sub_workflow step is exactly as exposed as a top-level one.
//
// This is not hypothetical: 8 of the 19 live findings at the commit that shipped
// this check are nested (tool-auditor ×2, tool-suggester ×2, internal-linker,
// component-quality-auditor, and others), so a top-level-only walk would have
// undercounted by nearly half — bugs_open/144's cost, in this check's own terms.
func TestUndeclaredRecurrence_SeesNestedSteps(t *testing.T) {
	const nested = `[{"type": "tool-auditor", "workflow": {"start_step": "create_items_loop", "steps": {
		"create_items_loop": {"action": "loop", "config": {
			"items_field": "audit.findings",
			"item_variable": "current_finding",
			"sub_workflow": {"start_step": "create_improve_item", "steps": {
				"create_improve_item": {"action": "create_work_item", "config": {
					"item_type": "improve_tool", "item_key_prefix": "audit_fix",
					"item_key_suffix_field": "current_finding.id"}},
				"done": {"action": "loop_complete"}}}}}}}}]`

	agents, _, err := decodeLiveAgents([]byte(nested), "test")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	findings := findUndeclaredRecurrence(agents)
	if len(findings) != 1 {
		t.Fatalf("a nested keyed step must be seen; got %d findings", len(findings))
	}
	if findings[0].Path != "steps.create_items_loop.sub_workflow.create_improve_item" {
		t.Errorf("path must be the qualified nested path, got %q", findings[0].Path)
	}
	// And note what does NOT acquit it: a per-item item_key_suffix_field is set
	// here, which satisfies --loop-sitewide-item-keys entirely. The two checks
	// ask different questions of the same step, and a step can pass one while
	// failing the other.
}
