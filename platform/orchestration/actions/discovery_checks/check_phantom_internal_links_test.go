package discovery_checks

import (
	"os"
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

// This check must judge "would a link to this page 404" by the SAME rule the
// chrome renderer uses when it WRITES nav links, or the platform flags links it
// authored itself — which is exactly what bugs_open/049 mechanism 2 was. The
// predicate moved to datahelpers on 2026-07-26 so both share one definition;
// what is worth pinning here is that this check still uses the shared one rather
// than reintroducing a local copy that can drift.
//
// The predicate's own content is pinned in datahelpers/links_deployment_test.go.
func TestPhantomCheckUsesTheSharedNeverDeployedPredicate(t *testing.T) {
	if !strings.Contains(datahelpers.NeverDeployedPagePredicate, "deployed_at IS NULL") {
		t.Fatalf("shared predicate must key on deployed_at IS NULL, got %q",
			datahelpers.NeverDeployedPagePredicate)
	}

	src, err := os.ReadFile("check_phantom_internal_links.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if !strings.Contains(string(src), "datahelpers.NeverDeployedPagePredicate") {
		t.Error("check must build its query from datahelpers.NeverDeployedPagePredicate")
	}
	if strings.Contains(string(src), "const neverDeployedPredicate") {
		t.Error("a local copy of the predicate has come back; it will drift from the renderer's")
	}
}
