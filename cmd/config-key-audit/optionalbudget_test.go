// FILE: cmd/config-key-audit/optionalbudget_test.go
//
// RFC 022. censusOptionalKeys is the counter that closes the ruling's stated
// blind spot: the interim exempted every individually-inert opt-in field from
// architecture review, so only the accumulated COUNT can notice the tenth.
//
// The fixture is the REAL motivating shape, not an invented one: append_doc_note
// is the shared action bugs_open/223's note_body_suffix_field landed on (8 live
// consumers at the time of the ruling), and landmine-verifier/council-gate are
// two of its real carriers. The single-consumer arm exists because the budget
// must NOT fire there — a lone consumer's surface is its own business, and a
// counter that taxes it recreates exactly the per-change tax the ruling retired.
package main

import (
	"testing"

	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// Two agents share append_doc_note; only one carries diagnose_code_lookup.
// A loop's body hides one carrier inside substeps — the bugs_open/144 class —
// so a hand-written descent would undercount consumers and the budget would
// silently never fire. WalkSteps sees it; this fixture is what asserts that.
const sharedActionFleet = `[
	{"type": "landmine-verifier", "workflow": {"start_step": "load_entry", "steps": {
		"run_checks":      {"action": "diagnose_code_lookup", "next_step": "verify"},
		"persist_verdict": {"action": "append_doc_note"}
	}}},
	{"type": "council-gate", "workflow": {"start_step": "loop_seats", "steps": {
		"loop_seats": {"action": "loop_over_items", "config": {"sub_workflow": {"steps": {
			"persist": {"action": "append_doc_note"}
		}}}}
	}}}
]`

func TestOptionalKeyCensusCountsDeclarationsAndDistinctConsumers(t *testing.T) {
	agents, failed, err := decodeLiveAgents([]byte(sharedActionFleet), "test")
	if err != nil || failed != 0 {
		t.Fatalf("decodeLiveAgents: err=%v failed=%d", err, failed)
	}

	rows := censusOptionalKeys(agents, -1)
	if len(rows) == 0 {
		t.Fatal("census over a real registry returned no rows — the actions import has come unlinked")
	}

	byAction := make(map[string]optionalKeyCensusRow, len(rows))
	for _, r := range rows {
		byAction[r.Action] = r
		if r.OverBudget {
			t.Errorf("no budget was given, so nothing can be over budget: %+v", r)
		}
	}

	spec, ok := datahelpers.GetActionInputSpec("append_doc_note")
	if !ok || len(spec.Optional) == 0 {
		t.Fatal("fixture premise broken: append_doc_note no longer registers optional keys")
	}
	note, ok := byAction["append_doc_note"]
	if !ok {
		t.Fatal("append_doc_note missing from the census")
	}
	// The count is the DECLARATION's, read from the registry — asserting the
	// live spec's own length rather than a literal, so this test measures the
	// counter and not the fleet's current shape.
	if note.OptionalKeys != len(spec.Optional) {
		t.Errorf("optional_keys = %d, want the declaration's %d", note.OptionalKeys, len(spec.Optional))
	}
	// Two distinct consumers, one of them reachable ONLY through a loop's
	// sub-workflow — undercounting here means a hand-written descent got back in.
	if note.Consumers != 2 {
		t.Errorf("append_doc_note consumers = %d (%v), want 2 — the sub-workflow carrier must be seen",
			note.Consumers, note.Agents)
	}

	if lookup, ok := byAction["diagnose_code_lookup"]; ok && lookup.Consumers != 1 {
		t.Errorf("diagnose_code_lookup consumers = %d (%v), want 1", lookup.Consumers, lookup.Agents)
	}
}

func TestOptionalKeyBudgetFiresOnlyOnSharedActions(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(sharedActionFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	// Budget 0: every shared action with >= 1 optional key is over budget, and
	// no single-consumer action may be — whatever its surface. This pins the
	// shared-only rule without depending on any action's current key count.
	rows := censusOptionalKeys(agents, 0)
	firedShared, firedLone := false, false
	for _, r := range rows {
		if r.OverBudget {
			if r.Consumers >= 2 {
				firedShared = true
			} else {
				firedLone = true
			}
		}
	}
	if !firedShared {
		t.Error("budget 0 must flag a shared action carrying optional keys (append_doc_note qualifies)")
	}
	if firedLone {
		t.Error("a single-consumer action must never be over budget — the ruling's tax falls on SHARED seams only")
	}

	// Findings sort first, so the report's top is the answer.
	if len(rows) > 0 && !rows[0].OverBudget {
		for _, r := range rows {
			if r.OverBudget {
				t.Error("over-budget rows must sort before the rest")
				break
			}
		}
	}

	// A budget no shared action exceeds produces zero findings — the healthy
	// state must stay representable or the check decays into noise.
	for _, r := range censusOptionalKeys(agents, 10000) {
		if r.OverBudget {
			t.Errorf("budget 10000 must produce no findings, got %+v", r)
		}
	}
}
