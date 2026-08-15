// FILE: platform/orchestration/actions/queryresolve/page_list_eligibility_test.go
//
// Tests for the build-state floor every page-listing derivation must carry
// (bugs_open/052). The load-bearing property: a listing regenerated from the
// page set on each render must NOT advertise a page that would 404, and it must
// NOT delist a page that still serves. These are DB-free tests of the WHERE
// fragment the resolvers splice in — the SQL semantics themselves are verified
// against live data in the bug's RUNBOOK, but the fragment's SHAPE is pinned
// here so the fix cannot silently regress.

package queryresolve

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// The exact regression that WAS bugs_open/052: the generic (listedOnly=false)
// listing path — tool-list, game-list, guide-list, archetype-grid — carried NO
// build-state filter at all (eligibility was ""), so a planned, never-built,
// 404 page passed the status filter and was listed. Guard: neither branch is
// empty, and each carries its floor — the article branch its own deployed_at
// literal, the generic branch the estate's canonical "has shipped" predicate.
func TestPageListEligibilityAlwaysHasBuildStateFloor(t *testing.T) {
	for _, listedOnly := range []bool{false, true} {
		frag := pageListEligibilitySQL(listedOnly)
		if strings.TrimSpace(frag) == "" {
			t.Fatalf("pageListEligibilitySQL(%v) is empty — the never-built-page floor is gone (bugs_open/052)", listedOnly)
		}
	}
	if listed := pageListEligibilitySQL(true); !strings.Contains(listed, "deployed_at IS NOT NULL") {
		t.Errorf("article listing floor lacks a deployed_at floor: %q", listed)
	}
	if generic := pageListEligibilitySQL(false); !strings.Contains(generic, "deployed_at IS NULL") {
		t.Errorf("generic listing floor lacks a deployed_at axis: %q", generic)
	}
}

// The generic floor must ALSO keep a page that is `deployed` yet never stamped
// (idea.uk/tool-audience-check, a bugs_open/040 shape): it serves 200 and is a
// real tool page, so a plain `deployed_at IS NOT NULL` floor would delist a
// working tool — "worse than the bug".
//
// Since bugs_open/185 fix candidate 2 the floor is not spelled here at all: it
// is datahelpers.PageHasShippedPredicateFor("p"), verbatim, and THAT builder is
// where the keep lives (its `COALESCE(build_status,'') <> 'deployed'` conjunct
// is false for the unstamped-deployed row, so the negation keeps it —
// links_shipped_predicate_test.go pins that only 'deployed' is ever named). This
// test therefore guards the DERIVATION: a hand-respelled floor, however
// equivalent today, is exactly the drift the derivation exists to end.
func TestGenericListingIsTheCanonicalShippedPredicateVerbatim(t *testing.T) {
	frag := pageListEligibilitySQL(false)
	want := datahelpers.PageHasShippedPredicateFor("p")
	if !strings.Contains(frag, want) {
		t.Errorf("generic listing floor must BE the canonical shipped predicate, not a respelling:\n  floor: %q\n  want:  %q", frag, want)
	}
	if strings.Contains(frag, "OR p.build_status") {
		t.Errorf("generic listing floor carries a hand-written disjunct beside the canonical predicate: %q", frag)
	}
	// The property the old test asserted, now stated where it is actually
	// decided: the canonical predicate names 'deployed' and no other status.
	if !strings.Contains(want, "'deployed'") {
		t.Errorf("canonical shipped predicate no longer names 'deployed' — the unstamped-deployed keep is gone: %q", want)
	}
}

// The stricter article contract (listedOnly=true) additionally demands real
// section content, so plan-era scaffold rows and never-built duplicates are
// excluded from article listings. It must stay stricter than the generic floor.
func TestListedOnlyIsStricterThanGeneric(t *testing.T) {
	listed := pageListEligibilitySQL(true)
	if !strings.Contains(listed, "jsonb_array_length") {
		t.Errorf("listedOnly floor must still require non-empty sections: %q", listed)
	}
	if listed == pageListEligibilitySQL(false) {
		t.Error("listedOnly and generic floors must differ — tool pages legitimately have empty sections")
	}
}
