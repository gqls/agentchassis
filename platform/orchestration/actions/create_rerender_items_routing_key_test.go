// FILE: platform/orchestration/actions/create_rerender_items_routing_key_test.go
//
// bugs_open/440 phase 1b: this action is the FIRST sanctioned producer of
// `spec.routing_reason` (REB-008 forbids a second before RFC_062 lands).
//
// WHAT THESE TESTS EXIST TO PIN, and it is not "the field gets written":
//
//  1. LOCKSTEP. RoutingKey is set exactly when KeyReason is. Stamping on
//     "merely known" would silently re-route image_landed-without-a-component,
//     whose assemble is REB-001's DESIGNED degrade — a behaviour change wearing
//     a foundation's clothes, and invisible until the phase-3 flip made it
//     production behaviour weeks later.
//  2. NO BAD KEY CAN EVER BE PRODUCED. An unknown reason must leave the routing
//     key EMPTY, not carry the unknown value — otherwise phase 3's refusal
//     would fire on items this very action minted, and REB-008's constraint
//     would be violated by its own first producer.
//  3. THE DEDUP KEY IS UNTOUCHED. pageRerenderItemKey takes keyReason alone; if
//     the routing key ever enters it, items stop deduping against their own
//     history (idx_swi_dedup) — the one way this additive change could have
//     altered live behaviour.
//
// MUTATION CHECKS for whoever edits rerenderModeFor:
//   - set m.RoutingKey unconditionally in the `known` branch (i.e. outside the
//     StampReason guard) → TestRoutingKeyIsStampedInLockstepWithReason fails on
//     image_landed-without-component.
//   - set m.RoutingKey = reason in the `!known` branch →
//     TestUnknownReasonProducesNoRoutingKey fails.
// If either still passes, the lockstep is not what is producing the behaviour.

package actions

import (
	"testing"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/livespec"
)

func TestRoutingKeyIsStampedInLockstepWithReason(t *testing.T) {
	// The matrix that matters: every vocabulary value, with and without a
	// component_id, plus the empty and unknown cases. The invariant is one
	// line — RoutingKey is non-empty exactly when KeyReason is — and it holds
	// across the whole matrix rather than at a hand-picked example.
	cases := []string{"", "no_such_reason_xq440", "tool_retirement"}
	for _, r := range livespec.RerenderSectionReasons {
		cases = append(cases, r.Name)
	}
	for _, reason := range cases {
		for _, componentID := range []string{"", "11111111-1111-1111-1111-111111111111"} {
			m := rerenderModeFor(reason, componentID)
			if (m.RoutingKey != "") != (m.KeyReason != "") {
				t.Errorf("reason=%q componentID=%q: KeyReason=%q but RoutingKey=%q — the two "+
					"must be set together. Stamping a routing key where the item stamps no "+
					"reason re-routes at RFC_062's flip a page that assembles today (REB-001's "+
					"designed degrade), which is a behaviour change disguised as a foundation",
					reason, componentID, m.KeyReason, m.RoutingKey)
			}
			if m.RoutingKey != "" && m.RoutingKey != m.KeyReason {
				t.Errorf("reason=%q componentID=%q: RoutingKey=%q must equal KeyReason=%q — the "+
					"gate flip assumes the two fields carry the same value",
					reason, componentID, m.RoutingKey, m.KeyReason)
			}
		}
	}
}

func TestUnknownReasonProducesNoRoutingKey(t *testing.T) {
	// tool_retirement and light_palette_chrome_replaced are REAL observed
	// unknowns (bugs_open/440's census: 16 and 13 live items, silently
	// assembled). This action must never mint a routing key phase 3 would then
	// refuse — the refusal is for values that arrive from OUTSIDE the
	// vocabulary, never for items the estate produced itself.
	for _, unknown := range []string{"tool_retirement", "light_palette_chrome_replaced", "not_a_reason"} {
		// PRECONDITION, asserted rather than assumed (editquality advisory, round
		// 934327db): these poison values must still be OUTSIDE the vocabulary. If
		// a later lane declares one of them legitimately, this fails LOUDLY with
		// an instruction — the alternative is a test that quietly stops testing
		// anything, which is the 410 poison-row trap in another costume.
		if _, declared := livespec.RerenderSectionReasonByName(unknown); declared {
			t.Fatalf("%q is now IN the vocabulary, so it no longer poisons this test — pick "+
				"another out-of-vocabulary value here; do NOT delete the case", unknown)
		}
		m := rerenderModeFor(unknown, "11111111-1111-1111-1111-111111111111")
		// The early-return path itself, asserted in the diff rather than only in
		// the mutation narrative (editquality advisory): an unknown reason yields
		// NOTHING — not scoped, not stamped, no key.
		if m.Scoped || m.StampReason || m.KeyReason != "" {
			t.Errorf("%q: expected the unknown branch to return empty-handed, got "+
				"Scoped=%v StampReason=%v KeyReason=%q — if this branch ever falls through to "+
				"the assignment block, REB-008's no-bad-producer guarantee is gone",
				unknown, m.Scoped, m.StampReason, m.KeyReason)
		}
		if m.RoutingKey != "" {
			t.Errorf("%q: produced RoutingKey=%q — an unknown value in the routing key would make "+
				"phase 3 refuse an item this action minted, violating REB-008's constraint at its "+
				"own first producer", unknown, m.RoutingKey)
		}
		if m.UnknownReason != unknown {
			t.Errorf("%q: UnknownReason=%q — the loud-but-assemble report (bugs_open/404) must "+
				"survive this change untouched", unknown, m.UnknownReason)
		}
	}
}

func TestRoutingKeyResolvesAsSectionsForEveryStampedValue(t *testing.T) {
	// The round trip the flip depends on: whatever this producer stamps must
	// classify as RoutingSections at the consumer (livespec.ResolveRoutingReason,
	// REB-008). A value that stamped but did not resolve would refuse at phase 3.
	for _, r := range livespec.RerenderSectionReasons {
		m := rerenderModeFor(r.Name, "11111111-1111-1111-1111-111111111111")
		if m.RoutingKey == "" {
			continue // legitimately unstamped in this combination; covered above
		}
		if _, d := livespec.ResolveRoutingReason(m.RoutingKey); d != livespec.RoutingSections {
			t.Errorf("%q: producer stamped %q but the consumer classifies it %d, not RoutingSections "+
				"— producer and consumer disagree, which is the drift this vocabulary already "+
				"suffered once (bugs_open/404)", r.Name, m.RoutingKey, d)
		}
	}
}

func TestRoutingKeyIsNotPartOfTheDedupKey(t *testing.T) {
	// pageRerenderItemKey must still discriminate on keyReason alone. If the
	// routing key ever enters it, every item stops matching its own history and
	// idx_swi_dedup silently stops deduping.
	siteID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	withReason := pageRerenderItemKey("about", siteID, livespec.ReasonCTALinksStale)
	assembleOnly := pageRerenderItemKey("about", siteID, "")
	if withReason == assembleOnly {
		t.Fatal("the dedup key no longer discriminates render modes (bugs_open/024 defect 6)")
	}
	// The key for a stamped item must be derivable from keyReason alone — i.e.
	// unchanged by anything this phase added.
	m := rerenderModeFor(livespec.ReasonCTALinksStale, "")
	if got := pageRerenderItemKey("about", siteID, m.KeyReason); got != withReason {
		t.Fatalf("dedup key changed for a stamped item: %q vs %q — phase 1b must be inert for "+
			"dedup, and this is the assertion that proves it", got, withReason)
	}
}
