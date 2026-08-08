// TestCTATokens and TestBestPageMatch moved to
// platform/orchestration/datahelpers/label_match_test.go — ctaTokens and
// bestPageMatch were extracted there as LabelTokens/BestLabelMatch so the
// audit-time detector and the write-time resolver share one definition
// (bugs_open/203 follow-on). ctaAreaExcluded stays local: it duplicates
// resolve_internal_links' areasExcludedFromCTA deliberately, to avoid an
// import cycle (actions imports this package).

package discovery_checks

import (
	"testing"
)

func TestCTAAreaExcluded(t *testing.T) {
	cases := map[string]bool{
		"/contact.html":              true, // top-level page — the firstPathSegment blind spot
		"/contact/index.html":        true,
		"/legal/privacy.html":        true,
		"/about.html":                true,
		"/tools/gauntlet/index.html": false,
		"/archetypes.html":           false,
		"/":                          false,
	}
	for url, want := range cases {
		if got := ctaAreaExcluded(url); got != want {
			t.Errorf("ctaAreaExcluded(%q) = %v, want %v", url, got, want)
		}
	}
}
