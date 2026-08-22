// FILE: cmd/config-key-audit/commitshaexposure_test.go
//
// The standing form of migration 537's guard (bugs_closed/334). Every test
// pins the DECISION functions, not the I/O — this lane's own recorded trap
// (staged_component_build handoff 2026-08-21 §6.3) is a guard test that
// asserted inputs while the branch sat untested behind os.Exit.
//
// MUTATION PROOFS (no-op form, so the package still compiles):
//   - in commitShaExposureFindings, replace `Exposed: exposed[p]` with
//     `Exposed: true` → TestCommitShaExposureUnexposedProducerIsAFinding fails.
//   - in exposesCommitSha, drop the `step.Action != "complete_workflow"` guard
//     → TestCommitShaExposureResultMappingAloneIsNotExposure fails.
//   - make the walker top-level only →
//     TestCommitShaExposureNestedCompleteStepCountsAsExposed fails, which pins
//     the one DELIBERATE difference from 537's top-level-only SQL guard.
//   - in commitShaExposureRefusal, drop any refusal branch →
//     TestCommitShaExposureRefusals fails on that branch.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

// An unexposed member of (producers ∩ handlers) is THE finding; an exposed one
// is reported but clean; an acked one is a recorded exception.
func TestCommitShaExposureUnexposedProducerIsAFinding(t *testing.T) {
	exposedAgent := oneAgent(t, `{"steps":{"complete":{"action":"complete_workflow",
	  "config":{"result_mapping":{"commit_sha":"deploy_result.response.data.commit_sha"}}}}}`)[0]
	exposedAgent.Type = "good-handler"
	bareAgent := oneAgent(t, `{"steps":{"complete":{"action":"complete_workflow",
	  "config":{"output_fields":["status"]}}}}`)[0]
	bareAgent.Type = "new-handler"
	agents := []liveAgent{exposedAgent, bareAgent}

	producers := []string{"good-handler", "new-handler"}
	handlers := []string{"good-handler", "new-handler"}

	got := commitShaExposureFindings(agents, producers, handlers, map[string]bool{})
	if len(got) != 2 {
		t.Fatalf("findings = %d, want 2: every member of the intersection is reported", len(got))
	}
	// Sorted by agent: good-handler, new-handler.
	if !got[0].Exposed || got[0].Agent != "good-handler" {
		t.Errorf("good-handler must read exposed, got %+v", got[0])
	}
	if got[1].Exposed || got[1].Acknowledged {
		t.Errorf("new-handler must read UNEXPOSED and unacknowledged against an empty acks "+
			"map — the default must be a finding, or the guard passes before anyone looked: %+v", got[1])
	}
	if unexposedUnackedCount(got) != 1 {
		t.Errorf("unexposedUnackedCount = %d, want 1 — this is the exit-code decision", unexposedUnackedCount(got))
	}

	// The acknowledged case must be distinguishable, and must clear the exit.
	acked := commitShaExposureFindings(agents, producers, handlers,
		map[string]bool{"new-handler": true})
	if !acked[1].Acknowledged || unexposedUnackedCount(acked) != 0 {
		t.Error("an acked handler must read acknowledged and not count toward the exit code " +
			"— otherwise a recorded exception never clears the alarm")
	}
}

// Only the intersection is the population: a producer that handles no
// dispatched work (a dispatcher like build-dispatch-loop, whose collected_data
// carries every handler's shas) and a handler that produces no commits are
// both out of scope — 537's guard drew exactly this boundary, and widening it
// is how tool-generator became a false positive in the first census.
func TestCommitShaExposureOnlyTheIntersectionIsReported(t *testing.T) {
	agents := oneAgent(t, `{"steps":{"complete":{"action":"complete_workflow",
	  "config":{"result_mapping":{"commit_sha":"x"}}}}}`)
	agents[0].Type = "both-sets"

	got := commitShaExposureFindings(agents,
		[]string{"both-sets", "dispatcher-not-a-handler"},
		[]string{"both-sets", "handler-that-never-commits"},
		map[string]bool{})
	if len(got) != 1 || got[0].Agent != "both-sets" {
		t.Fatalf("findings = %+v, want exactly [both-sets]: producers-only and handlers-only "+
			"agents are out of the population", got)
	}
}

// The one deliberate difference from 537's SQL guard: a complete step nested in
// a loop sub-workflow still counts as exposed, in both container spellings.
// The guard's jsonb_each walk was top-level only; a second traversal blind to
// nesting is bugs_open/144.
func TestCommitShaExposureNestedCompleteStepCountsAsExposed(t *testing.T) {
	for _, container := range []string{"sub_workflow", "substeps"} {
		var wf string
		if container == "sub_workflow" {
			wf = `{"steps":{"loop":{"action":"loop_items","config":{"sub_workflow":{"steps":
			  {"complete":{"action":"complete_workflow","config":{"result_mapping":{"commit_sha":"x"}}}}}}}}}`
		} else {
			wf = `{"steps":{"loop":{"action":"loop_items","config":{"substeps":
			  {"complete":{"action":"complete_workflow","config":{"result_mapping":{"commit_sha":"x"}}}}}}}}`
		}
		agents := oneAgent(t, wf)
		if !exposesCommitSha(agents[0]) {
			t.Errorf("%s: a nested complete_workflow with result_mapping.commit_sha must "+
				"count as exposed — a handler whose complete step moves into a sub-workflow "+
				"must not start paging", container)
		}
	}
}

// Exposure means precisely 537's predicate: the COMPLETE step's result_mapping
// carrying commit_sha. A result_mapping on some other action is not the reply
// wire, and a complete step without the key exposes nothing.
func TestCommitShaExposureResultMappingAloneIsNotExposure(t *testing.T) {
	wrongAction := oneAgent(t, `{"steps":{"save":{"action":"save_to_database",
	  "config":{"result_mapping":{"commit_sha":"x"}}}}}`)
	if exposesCommitSha(wrongAction[0]) {
		t.Error("result_mapping.commit_sha on a non-complete_workflow step must not count — " +
			"the wire reads the workflow REPLY, which only the complete step shapes")
	}

	noKey := oneAgent(t, `{"steps":{"complete":{"action":"complete_workflow",
	  "config":{"result_mapping":{"status":"x"}}}}}`)
	if exposesCommitSha(noKey[0]) {
		t.Error("a complete_workflow whose result_mapping lacks commit_sha must not count")
	}

	// CONTROL: the conjunction of both properties IS exposure — without this,
	// the two negatives above would also pass if the function always said no.
	control := oneAgent(t, `{"steps":{"complete":{"action":"complete_workflow",
	  "config":{"result_mapping":{"commit_sha":"x"}}}}}`)
	if !exposesCommitSha(control[0]) {
		t.Fatal("CONTROL FAILED: the canonical exposed shape reads unexposed — both " +
			"negatives above are vacuous")
	}
}

// Every vacuity branch refuses, and the healthy state does not. Each of the
// three blind states produces a report byte-identical to a clean one, which is
// why they must exit 2 rather than print.
func TestCommitShaExposureRefusals(t *testing.T) {
	if msg := commitShaExposureRefusal([]string{}, []string{"h"}, 1); msg == "" {
		t.Error("zero producers in 30 days must refuse — items record shas daily, so an " +
			"empty set means the query went blind")
	}
	if msg := commitShaExposureRefusal([]string{"p"}, []string{}, 1); msg == "" {
		t.Error("zero handlers in 7 days must refuse — dispatch runs daily")
	}
	if msg := commitShaExposureRefusal([]string{"p"}, []string{"h"}, 0); msg == "" {
		t.Error("zero exposing agents fleet-wide must refuse — 537's own second guard: the " +
			"519–540 standardisation absent means rollback or a blind walk, never a clean fleet")
	}
	if msg := commitShaExposureRefusal([]string{"p"}, []string{"h"}, 1); msg != "" {
		t.Errorf("the healthy state must not refuse, got: %s", msg)
	}
}

// The acks loader: a documentation key is skipped, an entry with an empty
// reason does not count, a real entry does.
func TestCommitShaExposureAcksLoader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "acks.json")
	if err := os.WriteFile(path, []byte(`{
	  "_comment": "doc key, skipped",
	  "hollow-handler": {"reason": "  ", "date": "2026-08-22"},
	  "real-exception": {"reason": "renders only; its commit_sha hits are another item's envelope",
	                     "date": "2026-08-22", "review": "bugs_closed/334 §9"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	acked, err := loadCommitShaExposureAcks(path)
	if err != nil {
		t.Fatalf("loadCommitShaExposureAcks: %v", err)
	}
	if acked["hollow-handler"] {
		t.Error("an entry with a blank reason must not count as acknowledged — a hollow ack " +
			"is the exact shape the sibling gate's seat objected to")
	}
	if !acked["real-exception"] {
		t.Error("a complete entry must count as acknowledged")
	}
	if acked["_comment"] {
		t.Error("documentation keys must be skipped")
	}
}
