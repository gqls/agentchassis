package discovery_checks

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// TestAccumulateLinkIssuesRuntimeFill covers the data-runtime-fill guard: an
// empty href inside a client-hydrated shell is by design (not emitted), while a
// phantom non-empty href baked into that same shell is still a real defect.
func TestAccumulateLinkIssuesRuntimeFill(t *testing.T) {
	targets := sitePageTargets{
		valid:   datahelpers.NewPageURLSet([]string{"/real.html"}),
		unbuilt: map[string]string{},
	}

	cases := []struct {
		name        string
		html        string
		wantEmpty   int // empty_internal_href occurrences recorded
		wantPhantom int // phantom_internal_link occurrences recorded
	}{
		{
			name:      "normal component empty href is flagged",
			html:      `<a href="">Browse</a>`,
			wantEmpty: 1, wantPhantom: 0,
		},
		{
			name:      "runtime-fill shell empty href is exempt",
			html:      `<div data-runtime-fill="lobby"><a href="">Card</a></div>`,
			wantEmpty: 0, wantPhantom: 0,
		},
		{
			name:      "runtime-fill shell phantom target still flagged",
			html:      `<div data-runtime-fill="lobby"><a href="/ghost.html">Card</a></div>`,
			wantEmpty: 0, wantPhantom: 1,
		},
		{
			name:      "normal component phantom target flagged, real target not",
			html:      `<a href="/ghost.html">x</a><a href="/real.html">y</a>`,
			wantEmpty: 0, wantPhantom: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			counts := make(map[plKey]int)
			accumulateLinkIssues(counts, make(map[plKey]string), "page_component", "p", "id", "slot", tc.html, targets)

			var gotEmpty, gotPhantom int
			for k, n := range counts {
				switch k.issue {
				case "empty_internal_href":
					gotEmpty += n
				case "phantom_internal_link":
					gotPhantom += n
				}
			}
			if gotEmpty != tc.wantEmpty {
				t.Errorf("empty_internal_href = %d, want %d", gotEmpty, tc.wantEmpty)
			}
			if gotPhantom != tc.wantPhantom {
				t.Errorf("phantom_internal_link = %d, want %d", gotPhantom, tc.wantPhantom)
			}
		})
	}
}

// TestAccumulateLinkIssuesUnbuiltTarget covers bugs_open/049 mechanism 2: a href
// that resolves to a real pages row whose page has never been deployed. It is
// NOT a phantom (the row exists) and it is NOT fine (the URL 404s), so it gets
// its own issue type and carries the TARGET page id for remediation.
func TestAccumulateLinkIssuesUnbuiltTarget(t *testing.T) {
	targets := sitePageTargets{
		valid: datahelpers.NewPageURLSet([]string{"/real.html", "/never-built.html"}),
		unbuilt: map[string]string{
			"/never-built.html": "target-page-id",
		},
	}

	counts := make(map[plKey]int)
	targetIDs := make(map[plKey]string)
	accumulateLinkIssues(counts, targetIDs, "site_component", "", "", "footer",
		`<a href="/real.html">ok</a><a href="/never-built.html">dead</a><a href="/ghost.html">phantom</a>`,
		targets)

	var unbuilt, phantom int
	var gotTargetID string
	for k, n := range counts {
		switch k.issue {
		case "unbuilt_internal_link":
			unbuilt += n
			gotTargetID = targetIDs[k]
		case "phantom_internal_link":
			phantom += n
		}
	}
	if unbuilt != 1 {
		t.Errorf("unbuilt_internal_link = %d, want 1", unbuilt)
	}
	if phantom != 1 {
		t.Errorf("phantom_internal_link = %d, want 1 (/ghost.html has no row at all)", phantom)
	}
	// Without the target id the work item would be filed against the page
	// CONTAINING the link, whose rebuild re-emits the same correct href and
	// resolves nothing.
	if gotTargetID != "target-page-id" {
		t.Errorf("target page id = %q, want %q", gotTargetID, "target-page-id")
	}
}

// A deployed page must never be reported, and the unbuilt lookup must agree with
// PageURLSet on normalisation — /never-built.html and /never-built are the same
// target, so a link written either way has to resolve to the same verdict.
func TestUnbuiltLookupUsesSharedNormalisation(t *testing.T) {
	targets := sitePageTargets{
		valid:   datahelpers.NewPageURLSet([]string{"/news/index.html"}),
		unbuilt: map[string]string{datahelpers.NormalizePagePath("/news/index.html"): "news-id"},
	}

	// Both spellings normalise to the same page; both must be flagged.
	for _, href := range []string{"/news/index.html", "/news/", "/news"} {
		counts := make(map[plKey]int)
		accumulateLinkIssues(counts, make(map[plKey]string), "page_component", "p", "id", "slot",
			`<a href="`+href+`">x</a>`, targets)

		var unbuilt int
		for k, n := range counts {
			if k.issue == "unbuilt_internal_link" {
				unbuilt += n
			}
		}
		if unbuilt != 1 {
			t.Errorf("href %q: unbuilt_internal_link = %d, want 1", href, unbuilt)
		}
	}
}

// The predicate is deployed_at IS NULL, not build_status <> 'deployed'. Measured
// 2026-07-20: 34 of 34 needs_rebuild-but-once-deployed pages return 200, so
// keying on build_status alone would have been wrong about 34 of the 56 rows it
// selects. This asserts the SQL still says so — the constant is the contract.
func TestNeverDeployedPredicateKeysOnDeployedAt(t *testing.T) {
	if !strings.Contains(neverDeployedPredicate, "deployed_at IS NULL") {
		t.Fatalf("predicate must key on deployed_at IS NULL, got %q", neverDeployedPredicate)
	}
	// A page once deployed and later flagged needs_rebuild still serves its old
	// artefact; selecting it would produce false positives on live pages.
	if strings.Contains(neverDeployedPredicate, "'needs_rebuild'") {
		t.Errorf("predicate must not single out needs_rebuild, got %q", neverDeployedPredicate)
	}
}
