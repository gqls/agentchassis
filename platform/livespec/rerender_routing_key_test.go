// FILE: platform/livespec/rerender_routing_key_test.go
//
// The verification trap this seam inherits from bugs_open/410, discharged here:
// prove the refusal CAN fire, with a real observed unknown, not only that the
// happy path routes. A suite asserting only over today's five values passes on
// the day it is written and can never do anything else.
//
// MUTATION CHECKS for whoever edits ResolveRoutingReason:
//   - make the unknown branch return RoutingAssemble (the pre-split quiet
//     default) → TestResolveRoutingReason_UnknownRefuses must fail.
//   - make the empty branch return RoutingRefuse (over-strict — would refuse
//     every annotation-only item, the migration-696 shape) →
//     TestResolveRoutingReason_AbsentAssembles must fail.
// If either still passes, the resolver is not what produces the decision.

package livespec

import (
	"strings"
	"testing"
)

func TestResolveRoutingReason_UnknownRefuses(t *testing.T) {
	// tool_retirement is a REAL observed unknown: 16 live page_rerender items
	// carried it on 2026-08-31 and silently assembled (bugs_open/440's census).
	// Post-split, that exact value in the ROUTING key must refuse.
	for _, unknown := range []string{"tool_retirement", "light_palette_chrome_replaced", "no_such_reason_xq440"} {
		if _, d := ResolveRoutingReason(unknown); d != RoutingRefuse {
			t.Errorf("%q: got decision %d, want RoutingRefuse — an unknown routing key that "+
				"does not refuse re-creates bugs_open/440 behind the split that exists to end it", unknown, d)
		}
	}
}

func TestResolveRoutingReason_AbsentAssembles(t *testing.T) {
	// The annotation-only item is legal forever: migration 696 minted 11 of
	// them the day this seam was written, all correctly assemble-only.
	if _, d := ResolveRoutingReason(""); d != RoutingAssemble {
		t.Fatalf("got decision %d, want RoutingAssemble — refusing an absent key fails every "+
			"annotation-only item in the fleet", d)
	}
}

func TestResolveRoutingReason_EveryVocabularyValueRoutes(t *testing.T) {
	for _, want := range RerenderSectionReasons {
		got, d := ResolveRoutingReason(want.Name)
		if d != RoutingSections {
			t.Errorf("%q: got decision %d, want RoutingSections", want.Name, d)
		}
		if got != want {
			t.Errorf("%q: resolver returned %+v, want the vocabulary's own entry %+v — a "+
				"resolver that loses the scoping/stamping judgement re-creates the drift the "+
				"struct exists to prevent", want.Name, got, want)
		}
	}
}

func TestClauseRenderers_DeriveFromTheVocabulary(t *testing.T) {
	// Derived-by-loop so the clauses cannot drift from the list: every value
	// must appear under the right key(s), and the literal spec key must be the
	// constant, not a hand-typed spelling.
	trans, known := TransitionRerenderModeConditionClause(), CheckRoutingKnownConditionClause()
	for _, r := range RerenderSectionReasons {
		if !strings.Contains(trans, "input_data.spec."+RoutingReasonSpecKey+" == '"+r.Name+"'") ||
			!strings.Contains(trans, "input_data.spec.reason == '"+r.Name+"'") {
			t.Errorf("transition clause is missing %q under one of its two keys:\n%s", r.Name, trans)
		}
		if !strings.Contains(known, "input_data.spec."+RoutingReasonSpecKey+" == '"+r.Name+"'") {
			t.Errorf("known-clause is missing %q:\n%s", r.Name, known)
		}
	}
	if !strings.Contains(known, "input_data.spec."+RoutingReasonSpecKey+" == ''") {
		t.Errorf("known-clause is missing the absent-key disjunct — without it every legacy "+
			"item refuses at the read door:\n%s", known)
	}
	if strings.Contains(known, "input_data.spec.reason ==") {
		t.Errorf("known-clause must never test the ANNOTATION field:\n%s", known)
	}
}
