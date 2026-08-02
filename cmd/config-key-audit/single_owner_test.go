// FILE: cmd/config-key-audit/single_owner_test.go
//
// RFC 006 / bugs_closed/150. findSingleOwnerViolations is the durable half of
// "one promoter, one owner": migration 286 removed the two duplicate triage
// steps, but a removal is a one-off that the next agent to gain the step
// silently undoes. This check is what makes the second copy loud.
//
// The fixtures are the REAL shapes, not invented ones — the three-agent fan-out
// is the exact configuration that was live until 2026-08-02 (improvement-loop
// `triage_findings`, design-audit-agent `triage`, site-review-agent `triage`),
// and the single-owner fixture is what the fleet looks like after the migration.
// A detector whose only test is the state it was written to produce cannot tell
// you it still works, so both arms are here.
package main

import (
	"encoding/json"
	"testing"

	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// The pre-migration fleet: one site-wide promoter, three carriers. This is the
// input the check exists to reject.
const threeCarrierFleet = `[
	{"type": "improvement-loop", "workflow": {"start_step": "ensure_site_record", "steps": {
		"call_site_review":  {"action": "call_agent", "next_step": "triage_findings"},
		"triage_findings":   {"action": "triage_detected_items", "next_step": "increment_audit_pass"},
		"check_has_findings":{"action": "conditional"}
	}}},
	{"type": "design-audit-agent", "workflow": {"start_step": "ensure_site_record", "steps": {
		"triage": {"action": "triage_detected_items", "next_step": "complete"}
	}}},
	{"type": "site-review-agent", "workflow": {"start_step": "ensure_site_record", "steps": {
		"triage": {"action": "triage_detected_items", "next_step": "complete"}
	}}}
]`

func TestSingleOwnerViolationIsReportedWithEveryCarrier(t *testing.T) {
	agents, failed, err := decodeLiveAgents([]byte(threeCarrierFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}
	if failed != 0 {
		t.Fatalf("expected 0 undecodable rows, got %d", failed)
	}

	findings := findSingleOwnerViolations(agents, []string{"triage_detected_items"})
	if len(findings) != 1 {
		t.Fatalf("expected exactly 1 violation, got %d: %+v", len(findings), findings)
	}

	f := findings[0]
	if f.Action != "triage_detected_items" {
		t.Errorf("action = %q, want triage_detected_items", f.Action)
	}

	// EVERY carrier, not "the extras beyond the first". When this fires the
	// question is which agent should own the effect, and that cannot be
	// answered from a list that has already dropped one arbitrary candidate.
	wantOwners := []string{"design-audit-agent", "improvement-loop", "site-review-agent"}
	if len(f.Owners) != len(wantOwners) {
		t.Fatalf("owners = %v, want %v", f.Owners, wantOwners)
	}
	for i, w := range wantOwners {
		if f.Owners[i] != w {
			t.Errorf("owners[%d] = %q, want %q (sorted, so the report is stable across runs)", i, f.Owners[i], w)
		}
	}

	// Paths must name the STEP, since two agents calling it from differently
	// named steps is the normal case and "design-audit-agent" alone does not
	// tell you where to look.
	wantPath := "design-audit-agent.steps.triage"
	found := false
	for _, p := range f.Paths {
		if p == wantPath {
			found = true
		}
	}
	if !found {
		t.Errorf("paths = %v, want one entry %q", f.Paths, wantPath)
	}
}

// The post-migration fleet. The negative control: a check that fires on
// everything is worth nothing, and this is the state the fleet is actually in
// after 286, so a regression here means the detector has started crying wolf.
func TestOneOwnerIsNotAViolation(t *testing.T) {
	const oneCarrierFleet = `[
		{"type": "improvement-loop", "workflow": {"start_step": "ensure_site_record", "steps": {
			"triage_findings": {"action": "triage_detected_items", "next_step": "increment_audit_pass"}
		}}},
		{"type": "design-audit-agent", "workflow": {"start_step": "ensure_site_record", "steps": {
			"call_content_auditor": {"action": "call_agent", "next_step": "complete"},
			"complete":             {"action": "complete_workflow"}
		}}},
		{"type": "site-review-agent", "workflow": {"start_step": "ensure_site_record", "steps": {
			"write_strategic_findings": {"action": "write_audit_findings", "next_step": "complete"},
			"complete":                 {"action": "complete_workflow"}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(oneCarrierFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findSingleOwnerViolations(agents, []string{"triage_detected_items"})
	if len(findings) != 0 {
		t.Fatalf("expected no violations once the promoter has one owner, got %+v", findings)
	}

	out, err := json.Marshal(findings)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("zero findings must encode as '[]', not null, so a consumer cannot mistake it for a missing key; got %q", out)
	}
}

// An UNdeclared action carried by ten agents is not a finding. Most shared
// actions are legitimately shared; this check speaks only where an action has
// asserted that a second caller would be wrong.
func TestSharedActionsWithoutTheDeclarationAreNotFlagged(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(threeCarrierFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	// Same fleet, but nothing declared. call_agent and conditional appear in
	// several of these definitions and must stay silent.
	findings := findSingleOwnerViolations(agents, []string{"some_other_action"})
	if len(findings) != 0 {
		t.Fatalf("expected no violations when the carried action is not declared single-owner, got %+v", findings)
	}
}

// The same agent carrying the step twice is a DIFFERENT defect (a duplicated
// step inside one workflow) and must not be reported here — filing it under
// this heading would put two classes behind one message and make the owner
// question unanswerable.
func TestOneAgentCarryingItTwiceIsNotAnOwnershipViolation(t *testing.T) {
	const doubledWithinOneAgent = `[
		{"type": "improvement-loop", "workflow": {"start_step": "a", "steps": {
			"triage_findings": {"action": "triage_detected_items", "next_step": "triage_again"},
			"triage_again":    {"action": "triage_detected_items", "next_step": "complete"}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(doubledWithinOneAgent), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findSingleOwnerViolations(agents, []string{"triage_detected_items"})
	if len(findings) != 0 {
		t.Fatalf("a repeat within ONE agent is not an ownership violation, got %+v", findings)
	}
}

// A nested substep is a real carrier. bugs_open/144's lesson is that a
// hand-written traversal and the runtime's traversal go blind in different
// directions and then agree with each other, which is why every mode in this
// tool walks with validation.WalkSteps. If that ever regresses to a top-level
// loop, a promoter hidden inside a loop body would be invisible.
func TestANestedCarrierCounts(t *testing.T) {
	const nestedCarrier = `[
		{"type": "improvement-loop", "workflow": {"start_step": "a", "steps": {
			"triage_findings": {"action": "triage_detected_items", "next_step": "complete"}
		}}},
		{"type": "loop-agent", "workflow": {"start_step": "per_site", "steps": {
			"per_site": {"action": "loop", "config": {"substeps": {
				"triage": {"action": "triage_detected_items", "next_step": ""}
			}}}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(nestedCarrier), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findSingleOwnerViolations(agents, []string{"triage_detected_items"})
	if len(findings) != 1 {
		t.Fatalf("expected the nested carrier to count, got %d findings: %+v", len(findings), findings)
	}
	if len(findings[0].Owners) != 2 {
		t.Errorf("owners = %v, want both improvement-loop and loop-agent", findings[0].Owners)
	}
}

// The declaration must actually be REACHABLE from the registry, not merely
// present in the source. This is the difference between the field being set and
// the detector being armed: ListSingleOwnerActions is what the CLI calls, and if
// the spec registration or the listing ever stops including it, every export
// produces zero findings and reads as a clean fleet.
//
// It also pins the deliberate decision NOT to gate the listing on
// checksConfig() — triage_detected_items happens to declare ConfigKeys today,
// so a coupled implementation would pass this test by accident; the datahelpers
// test asserts the uncoupled behaviour directly.
func TestTheDeclarationIsReachableFromTheRegistry(t *testing.T) {
	declared := datahelpers.ListSingleOwnerActions()
	if len(declared) == 0 {
		t.Fatal("no action declares SingleOwner — the detector is inert and every export would read as clean")
	}
	for _, a := range declared {
		if a == "triage_detected_items" {
			return
		}
	}
	t.Errorf("triage_detected_items is not in ListSingleOwnerActions() = %v; RFC 006's ruling is unenforced", declared)
}
