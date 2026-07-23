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
)

// The exact regression that WAS bugs_open/052: the generic (listedOnly=false)
// listing path — tool-list, game-list, guide-list, archetype-grid — carried NO
// build-state filter at all (eligibility was ""), so a planned, never-built,
// 404 page passed the status filter and was listed. Guard: both branches carry
// a deployed_at floor, and neither is empty.
func TestPageListEligibilityAlwaysHasBuildStateFloor(t *testing.T) {
	for _, listedOnly := range []bool{false, true} {
		frag := pageListEligibilitySQL(listedOnly)
		if strings.TrimSpace(frag) == "" {
			t.Fatalf("pageListEligibilitySQL(%v) is empty — the never-built-page floor is gone (bugs_open/052)", listedOnly)
		}
		if !strings.Contains(frag, "deployed_at IS NOT NULL") {
			t.Errorf("pageListEligibilitySQL(%v) lacks a deployed_at floor: %q", listedOnly, frag)
		}
	}
}

// The generic floor must ALSO keep a page that is `deployed` yet never stamped
// (idea.uk/tool-audience-check, a bugs_open/040 shape): it serves 200 and is a
// real tool page, so a plain `deployed_at IS NOT NULL` floor would delist a
// working tool — "worse than the bug". The keep comes from the build_status
// disjunct.
func TestGenericListingKeepsDeployedButUnstampedPages(t *testing.T) {
	frag := pageListEligibilitySQL(false)
	if !strings.Contains(frag, "build_status = 'deployed'") {
		t.Errorf("generic listing floor must keep deployed-but-unstamped pages via a build_status disjunct: %q", frag)
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
