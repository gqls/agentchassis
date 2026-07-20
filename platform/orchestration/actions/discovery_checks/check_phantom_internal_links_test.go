package discovery_checks

import (
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
