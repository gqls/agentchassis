// FILE: platform/orchestration/actions/create_rerender_items_reason_parity_test.go
//
// bugs_open/404 — the parity half, at the reader that drifted.
//
// THE BUG FILE NAMED THIS TEST'S OWN TRAP and it is worth restating, because
// avoiding it is why this file looks the way it does:
//
//	"the assertion must read the LIVE condition, not a copy of it pasted into the
//	 test. A parity test whose two sides are both maintained by the same author,
//	 in the same file, cannot come out the other way."
//
// Defused by construction. This file asserts Go against livespec's ONE
// definition; the definition is asserted against the LIVE database every morning
// by config-key-audit --live-declaration-drift; and a corpus lint
// (livespec/rerender_reasons_test.go) catches a migration widening the live gate
// at COMMIT time. No side of any comparison is a copy: one side is always either
// the single definition or the production database.
//
// It exercises rerenderModeFor — the REAL function the action calls — rather than
// re-implementing the decision, so a test that agreed with a broken action is not
// available as a failure mode.
//
// ⚠ WHAT THESE TESTS CANNOT DO, AND WHERE THAT LIVES INSTEAD. Every test here
// reads its expectation FROM livespec.RerenderSectionReasons — that is what makes
// them conformance tests (does the Go behave as the table says?) and it is also
// what stops them detecting a change to the TABLE ITSELF. Delete StampAlways from
// template_changed and these all still pass, because the assertion moves with the
// data.
//
// That half is livespec's TestRerenderSectionReasonTableSemantics, which pins the
// table against hardcoded expectations WITH the reasoning in its failure text.
// Verified by mutation 2026-08-26: removing template_changed's StampAlways — the
// exact shape of this bug — is caught there and by nothing here.
//
// So the two files are not redundant and neither is sufficient: one asks whether
// the table says the right thing, the other whether the code does what the table
// says. Do not "simplify" either away on the grounds that the other exists.
//
//	mutation                                          test that catches it
//	------------------------------------------------  --------------------------------------
//	template_changed loses StampAlways (today's bug)  EveryDeclaredReasonStampsItsReason
//	an unknown reason silently assembles unreported   UnknownReasonAssemblesAndIsReported
//	scoping without a component_id                    ScopedReasonNeedsAComponentToScopeBy
//	the empty reason starts warning or stamping       EmptyReasonIsUnchangedAndSilent
//	the item key stops discriminating the two modes   DeclaredReasonKeyIsNotTheAssembleKey
package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// TestEveryDeclaredReasonStampsItsReason loops the DEFINITION, never a copy. A
// sixth reason added to livespec and not handled here fails immediately — which
// is the property the two 2026-08-18 migrations needed and did not have.
func TestEveryDeclaredReasonStampsItsReason(t *testing.T) {
	const comp = "1b1e0d6a-0000-4000-8000-000000000001"
	for _, r := range livespec.RerenderSectionReasons {
		t.Run(r.Name, func(t *testing.T) {
			m := rerenderModeFor(r.Name, comp)
			// THE DISCRIMINATING ASSERTION. An assemble-only item and a
			// re-resolving item are indistinguishable by STATUS — both complete —
			// so the spec's reason is the only thing that tells them apart.
			if m.KeyReason != r.Name {
				t.Fatalf("%s: KeyReason = %q, want %q. An empty KeyReason means the item carries no "+
					"reason, so page-rerender's gate routes it to ASSEMBLE and re-ships the stored "+
					"HTML unchanged while completing green — bugs_open/404 exactly.", r.Name, m.KeyReason, r.Name)
			}
			if !m.StampReason {
				t.Fatalf("%s: StampReason is false with a component_id supplied", r.Name)
			}
			if m.UnknownReason != "" {
				t.Fatalf("%s: reported as unknown, but it is in the declared vocabulary", r.Name)
			}
			if got := m.Scoped; got != r.ComponentScoped {
				t.Fatalf("%s: Scoped = %v with a component_id supplied, want %v (the table's own "+
					"ComponentScoped)", r.Name, got, r.ComponentScoped)
			}
		})
	}
}

// TestScopedReasonNeedsAComponentToScopeBy pins the difference between the two
// gates — the subtlety that makes putting a reason in the wrong one a new defect.
func TestScopedReasonNeedsAComponentToScopeBy(t *testing.T) {
	for _, r := range livespec.RerenderSectionReasons {
		t.Run(r.Name, func(t *testing.T) {
			m := rerenderModeFor(r.Name, "") // no component_id
			if m.Scoped {
				t.Fatalf("%s: Scoped without a component_id — there is nothing to scope BY", r.Name)
			}
			switch {
			case r.StampAlways:
				if m.KeyReason != r.Name {
					t.Fatalf("%s: StampAlways, so it must stamp even unscoped; got %q", r.Name, m.KeyReason)
				}
				if r.ComponentScoped && len(m.Warnings) == 0 {
					t.Fatalf("%s: stamped but unscoped — every page in the caller's list gets a "+
						"sections re-render, and that must be said out loud", r.Name)
				}
			default:
				// REB-001's designed degrade for the legacy pair: a reason without a
				// component_id falls back to assemble. Preserving it is deliberate.
				if m.KeyReason != "" {
					t.Fatalf("%s: is not StampAlways, so without a component_id it must degrade to "+
						"assemble (REB-001's design); got KeyReason %q", r.Name, m.KeyReason)
				}
			}
		})
	}
}

// TestUnknownReasonAssemblesAndIsReported — verbatim_adoption_deploy is the live
// worked example: adopt_verbatim stamps it on items that are SUPPOSED to
// assemble, so the answer here is loud, not refused.
func TestUnknownReasonAssemblesAndIsReported(t *testing.T) {
	m := rerenderModeFor("verbatim_adoption_deploy", "")
	if m.KeyReason != "" || m.StampReason || m.Scoped {
		t.Fatalf("an undeclared reason must assemble: got KeyReason=%q StampReason=%v Scoped=%v",
			m.KeyReason, m.StampReason, m.Scoped)
	}
	if m.UnknownReason != "verbatim_adoption_deploy" {
		t.Fatalf("an undeclared reason must be REPORTED — in the result, not only in a log line; "+
			"got UnknownReason=%q", m.UnknownReason)
	}
	if len(m.Warnings) == 0 {
		t.Fatal("an undeclared reason must warn: it will complete green having changed nothing, " +
			"and that is the one outcome nobody can see from the item")
	}
	if !strings.Contains(strings.Join(m.Warnings, " "), "bugs_open/404") {
		t.Fatal("the warning must name the bug, so whoever reads it can find out what happened")
	}
}

// TestEmptyReasonIsUnchangedAndSilent guards the ordinary case. [MEASURED
// 2026-08-26] 17,844 live+archived items carry no reason and are correct: a
// site-wide refresh IS assemble-only. Warning on those would drown the signal.
func TestEmptyReasonIsUnchangedAndSilent(t *testing.T) {
	for _, comp := range []string{"", "1b1e0d6a-0000-4000-8000-000000000001"} {
		m := rerenderModeFor("", comp)
		if m.KeyReason != "" || m.StampReason || m.Scoped {
			t.Fatalf("no reason must mean assemble-only; got %+v", m)
		}
		if m.UnknownReason != "" || len(m.Warnings) != 0 {
			t.Fatalf("no reason is the ORDINARY case and must be silent — 17,844 live items are "+
				"this shape; got UnknownReason=%q warnings=%v", m.UnknownReason, m.Warnings)
		}
	}
}

// TestDeclaredReasonKeyIsNotTheAssembleKey closes the loop to the dedup key,
// which is the other half of bugs_open/024 defect 6: the two modes must not be
// able to suppress each other.
func TestDeclaredReasonKeyIsNotTheAssembleKey(t *testing.T) {
	siteID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	assembleKey := pageRerenderItemKey("index", siteID, "")
	for _, r := range livespec.RerenderSectionReasons {
		m := rerenderModeFor(r.Name, "1b1e0d6a-0000-4000-8000-000000000001")
		got := pageRerenderItemKey("index", siteID, m.KeyReason)
		if got == assembleKey {
			t.Fatalf("%s: item key %q is the ASSEMBLE key — the two modes would dedup against each "+
				"other and one would silently suppress the other (bugs_open/024 defect 6)", r.Name, got)
		}
		if !strings.HasSuffix(got, r.Name) {
			t.Fatalf("%s: item key %q does not end in its reason, so the mode is not readable "+
				"from the key", r.Name, got)
		}
	}
}

// TestFixerPageStatusFragmentMatchesCanonicalPredicate closes a seam livespec's
// leaf status forces open: it cannot import datahelpers, so the two spellings of
// "the platform still wants this page served" are tied HERE, in the package that
// imports both. Without this the declaration could drift to a hand-typed
// `p.status='active'` that no longer matches what the estate means by it.
func TestFixerPageStatusFragmentMatchesCanonicalPredicate(t *testing.T) {
	d := livespec.MustGet("workflow.component-template-fixer.create_rerender")
	want := datahelpers.PageWantedLivePredicateFor("p")
	for _, f := range d.Fragments {
		if f.Text == want {
			return
		}
	}
	t.Fatalf("the fixer declaration must pin the CANONICAL liveness predicate %q; fragments are %v",
		want, d.Fragments)
}
