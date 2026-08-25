package main

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/livespec"
)

// The two subtleties that make a naive version of this check confidently wrong,
// pinned so a later edit has to decide rather than drift (both measured
// 2026-08-24, both named in the file header).

// TRAP 1: `status` DEFAULTS to `complete`. UpdateWorkItemStatusAction sets
// newStatus := "complete" and only THEN reads config["status"], so a step that
// omits the key is a complete arm. A filter spelled `status == "complete"` is
// blind to it, and blind in the direction that hides an unguarded completer.
func TestCompletesUnarmed_AbsentStatusIsAnUnarmedCompleteArm(t *testing.T) {
	cases := []struct {
		name string
		cfg  map[string]interface{}
		want bool
	}{
		{"status absent — DEFAULTS to complete", map[string]interface{}{}, true},
		{"status empty string — also defaults", map[string]interface{}{"status": ""}, true},
		{"status complete, unarmed", map[string]interface{}{"status": "complete"}, true},
		{"status complete, ARMED", map[string]interface{}{"status": "complete", "verify_before_complete": true}, false},
		{"armed:false is not armed", map[string]interface{}{"status": "complete", "verify_before_complete": false}, true},
		{"armed as a STRING is not a bool — Go reads it as unarmed, so the audit must too",
			map[string]interface{}{"status": "complete", "verify_before_complete": "true"}, true},
		{"failed", map[string]interface{}{"status": "failed"}, false},
		{"needs_human_review", map[string]interface{}{"status": "needs_human_review"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			step := models.Step{Action: "update_work_item_status", Config: c.cfg}
			if got := completesUnarmed(step); got != c.want {
				t.Errorf("completesUnarmed(%v) = %v, want %v", c.cfg, got, c.want)
			}
		})
	}
	// A different action with the same config shape must be ignored entirely.
	if completesUnarmed(models.Step{Action: "complete_work_item", Config: map[string]interface{}{}}) {
		t.Error("complete_work_item is the GUARDED writer — it must never appear in this census")
	}
}

// TRAP 2: a top-level walk misses NESTED steps. Measured on the sibling writer: a
// $.workflow.steps scan finds only 2 of the 4 live complete_work_item callers,
// because the dispatch loops carry theirs inside a sub_workflow. This asserts the
// traversal reaches one.
func TestFindUnarmedCompleterDrift_SeesNestedSteps(t *testing.T) {
	agents := []liveAgent{{
		Type: "nested-agent",
		Workflow: models.WorkflowPlan{Steps: map[string]models.Step{
			"process_item": {
				Action: "loop_over_items",
				Config: map[string]interface{}{
					"sub_workflow": map[string]interface{}{
						"steps": map[string]interface{}{
							"close_inner": map[string]interface{}{
								"action": "update_work_item_status",
								"config": map[string]interface{}{"status": "complete"},
							},
						},
					},
				},
			},
		}},
	}}

	got := findUnarmedCompleterDrift(agents, nil)
	var found bool
	for _, f := range got {
		if f.Kind == "undeclared" && f.Step == "close_inner" {
			found = true
		}
	}
	if !found {
		t.Fatalf("a NESTED unarmed complete arm must be found — a top-level-only walk is the bugs_open/144 "+
			"blindness this mode exists to avoid. got %+v", got)
	}
}

// Both directions, and a clean case that must stay clean. Without the clean case
// the two above would pass just as happily for a check that reports everything.
func TestFindUnarmedCompleterDrift_BothDirectionsAndAClean(t *testing.T) {
	agents := []liveAgent{{
		Type: "a1",
		Workflow: models.WorkflowPlan{Steps: map[string]models.Step{
			"close_it":  {Action: "update_work_item_status", Config: map[string]interface{}{"status": "complete"}},
			"armed_one": {Action: "update_work_item_status", Config: map[string]interface{}{"status": "complete", "verify_before_complete": true}},
			"fail_it":   {Action: "update_work_item_status", Config: map[string]interface{}{"status": "failed"}},
		}},
	}}

	// CLEAN: declared exactly. An ARMED arm must NOT need declaring — arming is the
	// whole point, and demanding an entry for it would punish the fix.
	clean := findUnarmedCompleterDrift(agents, []livespec.UnarmedCompleter{
		{Agent: "a1", Step: "close_it", ItemType: "t", Why: "w"},
	})
	if len(clean) != 0 {
		t.Fatalf("a matching declaration must report nothing, got %+v", clean)
	}

	// UNDECLARED — the dangerous direction.
	if got := findUnarmedCompleterDrift(agents, nil); len(got) != 1 ||
		got[0].Kind != "undeclared" || got[0].Step != "close_it" {
		t.Fatalf("an undeclared live arm must be reported once, got %+v", got)
	}

	// STALE — the tidy-up direction.
	got := findUnarmedCompleterDrift(agents, []livespec.UnarmedCompleter{
		{Agent: "a1", Step: "close_it", ItemType: "t", Why: "w"},
		{Agent: "a1", Step: "gone", ItemType: "t", Why: "w"},
	})
	if len(got) != 1 || got[0].Kind != "stale" || got[0].Step != "gone" {
		t.Fatalf("a declared arm with no live counterpart must be reported stale, got %+v", got)
	}
}

// The declaration must actually match live config. This is a UNIT test of the
// shape, not of production — the production comparison is the mode itself, run
// against live-workflows.json. Asserted because an entry whose Agent or Step is
// misspelled reports as BOTH a stale declaration and an undeclared live arm, and
// a reader meeting two findings looks for two problems.
func TestUnarmedVerifiedCompletersDeclarationIsWellFormed(t *testing.T) {
	if len(livespec.UnarmedVerifiedCompleters) == 0 {
		t.Fatal("empty declaration — this mode would report every live arm as undeclared")
	}
	for _, d := range livespec.UnarmedVerifiedCompleters {
		if d.Agent == "" || d.Step == "" || d.ItemType == "" || d.Why == "" {
			t.Errorf("incomplete entry %+v", d)
		}
	}
}
