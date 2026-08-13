// FILE: platform/orchestration/actions/feed_directory_recommendation_action_test.go
//
// Phase B (2026-08-13): the vertical→directory-kind matcher. Pure-function
// table tests — the DB write half reuses the news action's proven
// supersede-then-insert shape verbatim and is not re-tested here.

package actions

import (
	"testing"

	"go.uber.org/zap"
)

func TestMatchVerticalDirectory(t *testing.T) {
	cases := []struct {
		name     string
		industry string
		siteType string
		category string
		domain   string
		wantKind string // "" = no recommendation (nil or Recommended:false)
	}{
		{"exact mortgage industry", "mortgage", "", "", "example.co.uk", "mortgage-lender"},
		{"exact savings", "savings", "", "", "example.co.uk", "savings-provider"},
		{"banking aliases to savings-provider", "banking", "", "", "example.co.uk", "savings-provider"},
		{"exact health insurance", "health insurance", "", "", "example.co.uk", "health-insurer"},
		{"partial: mortgage broker industry", "mortgage brokerage", "", "", "example.co.uk", "mortgage-lender"},

		// The longest-key determinism rule this matcher adds over
		// matchVerticalNews: a signal containing BOTH "health insurance" and
		// "insurance" must always take the longer, more specific key. (Both
		// resolve to health-insurer today, but the rule is what's pinned —
		// when more insurer kinds land, the specific one must keep winning.)
		{"longest partial key wins", "private health insurance advice", "", "", "example.co.uk", "health-insurer"},

		// "finance" alone is an explicit NOT-recommended entry, not a miss:
		// too generic to choose a provider class.
		{"bare finance is refused", "finance", "", "", "example.co.uk", ""},

		// A signal containing both "insurance" (recommends) and "finance"
		// (refuses): "insurance" is longer, so the recommendation wins —
		// deterministically, not by map iteration luck.
		{"insurance beats finance by length", "insurance finance", "", "", "example.co.uk", "health-insurer"},

		{"domain-derived signal", "", "", "", "bestmortgagedeals.co.uk", "mortgage-lender"},
		{"no signal at all", "", "", "", "", ""},
		{"unrelated vertical", "veterinary", "clinic", "", "vetexample.co.uk", ""},
	}

	logger := zap.NewNop()
	for _, c := range cases {
		got := matchVerticalDirectory(c.industry, c.siteType, c.category, c.domain, logger)
		gotKind := ""
		if got != nil && got.Recommended {
			gotKind = got.Kind
		}
		if gotKind != c.wantKind {
			t.Errorf("%s: matchVerticalDirectory(%q,%q,%q,%q) kind = %q, want %q",
				c.name, c.industry, c.siteType, c.category, c.domain, gotKind, c.wantKind)
		}
	}
}

// Every recommending entry must name a kind that actually exists in
// directoryPublishProfiles — a recommendation for an unrenderable kind would
// plan a page nothing can ever populate. This is the lockstep guard between
// the two tables, in the same spirit as the planner-vocabulary caveat in
// SEED_directory_components.sql.
func TestVerticalDirectoryMapKindsAreRenderable(t *testing.T) {
	for signal, cfg := range verticalDirectoryMap {
		if !cfg.Recommended {
			continue
		}
		if cfg.Kind == "" || cfg.SpecKey == "" {
			t.Errorf("entry %q recommends but is missing Kind/SpecKey", signal)
			continue
		}
		if _, ok := directoryPublishProfiles[cfg.Kind]; !ok {
			t.Errorf("entry %q recommends kind %q, which has no directoryPublishProfiles entry — unrenderable", signal, cfg.Kind)
		}
	}
}
