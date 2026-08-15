// FILE: platform/orchestration/actions/revalidate_unverified_claims_archivedskip_test.go
//
// The revalidator half of the 2026-08-15 scan exclusion (archived AND never
// deployed). The emit-side contract is pinned in
// discovery_checks/check_unverified_claims_archivedskip_test.go; this pins what
// an OPERATOR is told when that exclusion is the reason a page came back empty.
//
// Why it needs a test at all: ScanDeployedClaims returns no row for an excluded
// page, which is indistinguishable — at this layer — from a page that was
// deleted. The verdict is the same and that is correct (both refuse). The REASON
// is what a person reads in the review queue, and "page is absent" about a page
// they can see in the database sends them looking for a deletion that never
// happened. That is a real cost: the arm cannot close on its own, so a human has
// to act on this text.
//
// This asserts on shipped OUTPUT, not on a comment. The distinction matters —
// the sibling file's harness deliberately captures SQL rather than scanning
// source for exactly the opposite reason.
//
// PROVEN BY MUTATION (2026-08-15), and the limit is stated because it is real:
//
//	mutation applied to revalidate_unverified_claims.go   result
//	---------------------------------------------------   --------------------
//	full revert of Reason to its pre-2026-08-15 text      CAUGHT
//	drop the "not evidence the claims were removed" tail  CAUGHT
//	revert only the first clause, leaving                 NOT CAUGHT
//	  "never-published" standing in the second
//
// ⚠ The third row is this test's honest ceiling. It looks for the CONCEPT by
// one word each, so any rewording that happens to retain those words passes.
// It catches a cause being dropped; it cannot catch a cause being described
// badly. Tightening it to exact prose would trade that for a test that fails on
// every copy-edit, which is worse — but do not read a green run here as "the
// reason text is good", only as "no cause was silently deleted".

package actions

import (
	"strings"
	"testing"
)

// TestPageAbsentReasonNamesTheExclusion — the never-deployed cause must be
// legible in the reason, alongside the two that were always there.
func TestPageAbsentReasonNamesTheExclusion(t *testing.T) {
	got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, nil, publishedAfterEdit())

	if got.Arm != armPageAbsent {
		t.Fatalf("premise wrong: a nil scan no longer produces %s, it produced %s — "+
			"this test is asserting on the wrong arm", armPageAbsent, got.Arm)
	}
	if got.Verdict != revalidationUnknown {
		t.Errorf("an excluded or absent page must REFUSE, not resolve; got verdict %q", got.Verdict)
	}

	reason := strings.ToLower(got.Reason)

	// The three causes a reader has to be able to tell apart. Checked by concept,
	// not by exact phrasing, so a rewording does not fail this test but a dropped
	// cause does.
	for _, want := range []struct{ cause, needle string }{
		{"deleted / missing page", "absent"},
		{"archived and never deployed", "never"},
		{"nothing built on the page", "no component"},
	} {
		if !strings.Contains(reason, want.needle) {
			t.Errorf("the %s cause is not legible in the operator-facing reason (looked for %q). "+
				"An operator who cannot tell these apart cannot dispose of the item.\n%s",
				want.cause, want.needle, got.Reason)
		}
	}

	// The refusal must not read as evidence of a fix. This is the standing rule
	// across every arm in this ladder, and it is the one a rewrite is most likely
	// to lose while making the text friendlier.
	if !strings.Contains(reason, "not evidence") {
		t.Errorf("the reason no longer says the absence is NOT evidence the claims were removed; "+
			"withdrawing a page is not the same as substantiating what it said.\n%s", got.Reason)
	}
}
