// FILE: cmd/config-key-audit/optionalexplicitwires_test.go
//
// RFC_029 §10.15 / council REVISE round 1. The guard that makes `?` adoption
// countable and unacknowledged adoption loud.
//
// MUTATION PROOFS (no-op form, so the package still compiles):
//   - in findOptionalExplicitWires, replace `Acknowledged: acked[...]` with
//     `Acknowledged: true` → TestOptionalExplicitWireIsUnacknowledgedByDefault
//     fails.
//   - replace the `if strict || !optional || base == ""` guard with
//     `if base == ""` → TestOptionalExplicitWiresIgnoreTheStrictMarker fails.
//   - make the walker top-level only → TestOptionalExplicitWireFoundInNested
//     Substeps fails, which is the census hole the guardian seat named.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// oneAgent decodes a workflow from JSON so the fixtures exercise the same
// decode path the live export does, rather than a hand-built struct that could
// agree with the walker by construction.
func oneAgent(t *testing.T, jsonWorkflow string) []liveAgent {
	t.Helper()
	var a liveAgent
	if err := json.Unmarshal([]byte(`{"type":"fixture-agent","workflow":`+jsonWorkflow+`}`), &a); err != nil {
		t.Fatalf("fixture does not decode: %v", err)
	}
	return []liveAgent{a}
}

// A `?` wire with no ack is a finding. This is the guard's whole purpose.
func TestOptionalExplicitWireIsUnacknowledgedByDefault(t *testing.T) {
	agents := oneAgent(t, `{"steps":{"save_tool":{"action":"create_tool_component",
	  "config":{"related_pages?":"input_data.spec.related_pages"}}}}`)

	got := findOptionalExplicitWires(agents, map[string]bool{})
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1: a live `?` wire must be reported", len(got))
	}
	if got[0].Acknowledged {
		t.Error("wire reads as acknowledged against an EMPTY acks map — the default must be " +
			"unacknowledged, or the guard passes before anyone has looked")
	}
	if got[0].Field != "related_pages" || got[0].Reference != "input_data.spec.related_pages" {
		t.Errorf("field/reference = %q/%q, want related_pages/input_data.spec.related_pages: "+
			"the ack is judged against the path the author wrote", got[0].Field, got[0].Reference)
	}

	// And the acknowledged case, so the test can distinguish the two states.
	acked := findOptionalExplicitWires(agents,
		map[string]bool{"create_tool_component.related_pages": true})
	if len(acked) != 1 || !acked[0].Acknowledged {
		t.Error("an acked (action, field) must read as acknowledged — otherwise acknowledging " +
			"one never clears it and the guard cannot be satisfied")
	}
}

// The OTHER surface is not this audit's business: a `?` inside input_mapping is
// ResolveInputMapping's long-standing marker (77 live keys), not an adoption of
// the new one. Reporting those would bury the finding in known-good noise.
func TestOptionalExplicitWiresIgnoreTheInputMappingSurface(t *testing.T) {
	agents := oneAgent(t, `{"steps":{"call_writer":{"action":"call_agent",
	  "config":{"input_mapping":{"rewrite_guidance?":"input_data.spec.suggestion"}}}}}`)

	if got := findOptionalExplicitWires(agents, map[string]bool{}); len(got) != 0 {
		t.Fatalf("findings = %#v, want none: a `?` under input_mapping belongs to the sibling "+
			"surface and is not an ExtractActionInputs adoption", got)
	}

	// CONTROL: the same key moved onto the step config IS reported. Without
	// this, the assertion above would also pass if the walker were broken.
	control := oneAgent(t, `{"steps":{"call_writer":{"action":"call_agent",
	  "config":{"rewrite_guidance?":"input_data.spec.suggestion"}}}}`)
	if got := findOptionalExplicitWires(control, map[string]bool{}); len(got) != 1 {
		t.Fatalf("CONTROL FAILED: the same key on the step config produced %d findings, want 1 "+
			"— the walker is not reaching step config at all, so the negative above is vacuous",
			len(got))
	}
}

// `!` is a different marker with a different (loud) failure mode; it has its
// own enforcement in the resolver and must not be reported as an unacknowledged
// silent-absence wire.
func TestOptionalExplicitWiresIgnoreTheStrictMarker(t *testing.T) {
	agents := oneAgent(t, `{"steps":{"mark_complete":{"action":"complete_work_item",
	  "config":{"result!":"handler_result","work_item_id!":"current_item.id"}}}}`)

	if got := findOptionalExplicitWires(agents, map[string]bool{}); len(got) != 0 {
		t.Fatalf("findings = %#v, want none: `!` fails loudly and is not the silent-absence "+
			"class this guard audits", got)
	}
}

// The walker must see NESTED steps, under both container spellings. This is the
// exact hole the guardian seat named in the census: `substeps` is what execution
// prefers, and a descent that knows only `sub_workflow` reports a confident zero.
func TestOptionalExplicitWireFoundInNestedSubsteps(t *testing.T) {
	for _, container := range []string{"sub_workflow", "substeps"} {
		var wf string
		if container == "sub_workflow" {
			wf = `{"steps":{"loop":{"action":"loop_items","config":{"sub_workflow":{"steps":
			  {"inner":{"action":"complete_work_item","config":{"commit_sha?":"handler_result.sha"}}}}}}}}`
		} else {
			wf = `{"steps":{"loop":{"action":"loop_items","config":{"substeps":
			  {"inner":{"action":"complete_work_item","config":{"commit_sha?":"handler_result.sha"}}}}}}}`
		}
		agents := oneAgent(t, wf)
		got := findOptionalExplicitWires(agents, map[string]bool{})
		if len(got) != 1 {
			t.Errorf("%s: findings = %d, want 1 — a nested `?` wire is invisible to this guard, "+
				"which is how a confident zero gets reported over a real population", container, len(got))
			continue
		}
		if got[0].Action != "complete_work_item" {
			t.Errorf("%s: action = %q, want complete_work_item: the ack is keyed on the "+
				"CONSUMING action", container, got[0].Action)
		}
	}
}

// An acks entry with an empty `downstream` does NOT count. An ack that says
// nothing about what was checked is the hollow ack the objection named, and it
// would otherwise read as acknowledged to both this tool and a grep.
func TestHollowAckDoesNotCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acks.json")
	if err := os.WriteFile(path, []byte(`{
	  "_doc": "documentation keys are skipped",
	  "create_tool_component.related_pages": {"downstream": "   ", "date": "2026-08-21"},
	  "complete_work_item.commit_sha": {"downstream": "315 lane confirmed an absent sha leaves the stamp unwritten rather than writing an empty one", "date": "2026-08-21"}
	}`), 0o600); err != nil {
		t.Fatal(err)
	}

	acked, err := loadOptionalExplicitAcks(path)
	if err != nil {
		t.Fatalf("loadOptionalExplicitAcks: %v", err)
	}
	if acked["create_tool_component.related_pages"] {
		t.Error("a whitespace-only `downstream` counted as an ack — an ack must state what " +
			"was checked, or the guard is satisfied by typing the key")
	}
	if !acked["complete_work_item.commit_sha"] {
		t.Error("CONTROL FAILED: a real ack did not count, so the assertion above cannot " +
			"distinguish hollow from real")
	}
	if acked["_doc"] {
		t.Error("a leading-underscore documentation key was read as an ack")
	}
}

// A missing acks file must be an ERROR, never an empty ack set: silently
// treating "cannot read the file" as "nothing is acknowledged" turns a broken
// invocation into a wall of findings, and the reverse default would turn it
// into a pass. Both readings are wrong; refusing to run is the third option.
func TestMissingAcksFileIsAnError(t *testing.T) {
	if _, err := loadOptionalExplicitAcks(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("a missing acks file returned no error — the mode would then run with an " +
			"empty ack set and report every wire, which reads as a finding rather than as " +
			"a check that did not run")
	}
}
