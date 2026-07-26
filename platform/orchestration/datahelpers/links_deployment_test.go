package datahelpers

import (
	"strings"
	"testing"
)

// NeverDeployedPagePredicate decides, in SQL, whether a link to a page would
// 404. Two consumers depend on it agreeing with live reality: the chrome
// renderer, which must not WRITE such a link, and the post-deploy audit, which
// must FLAG one. Getting it wrong in either direction is expensive — the
// renderer would either ship 404s or delete working navigation.
//
// The measurement behind it (fleet-wide, 2026-07-20, against live HTTP):
//
//	needs_rebuild AND deployed_at IS NOT NULL   34 pages — 34/34 return 200
//	deployed_at IS NULL                         22 pages — every one tested 404s
//
// So the tempting simplification, build_status <> 'deployed', would have been
// wrong about 34 of the 56 rows it selects.
func TestNeverDeployedPagePredicateKeysOnDeployedAt(t *testing.T) {
	if !strings.Contains(NeverDeployedPagePredicate, "deployed_at IS NULL") {
		t.Fatalf("predicate must key on deployed_at IS NULL, got %q", NeverDeployedPagePredicate)
	}

	// A page deployed once and later flagged needs_rebuild still serves its old
	// artefact. Singling that status out would false-flag 34 live pages.
	if strings.Contains(NeverDeployedPagePredicate, "'needs_rebuild'") {
		t.Errorf("predicate must not single out needs_rebuild, got %q", NeverDeployedPagePredicate)
	}

	// build_status appears only as the second conjunct that excludes the one
	// fleet row marked 'deployed' yet never stamped. If it ever becomes the
	// primary test, the 34-page false-positive class is back.
	if !strings.Contains(NeverDeployedPagePredicate, "COALESCE(build_status, '') <> 'deployed'") {
		t.Errorf("predicate lost the build_status escape conjunct, got %q", NeverDeployedPagePredicate)
	}
}
