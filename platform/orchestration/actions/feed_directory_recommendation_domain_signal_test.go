// FILE: platform/orchestration/actions/feed_directory_recommendation_domain_signal_test.go
//
// Regression test for the DOMAIN-signal half of matchVerticalDirectory.
//
// matchVerticalDirectory collects domain-derived signals by ranging over
// verticalDirectoryMap — and Go randomises map iteration order. The outer
// loop then returns on the FIRST exact match. So a domain containing two
// keywords with OPPOSITE recommendations resolved differently run to run.
//
// This is precisely the defect the PARTIAL-match arm already documents and
// guards against ("map iteration order is random, and this map mixes
// recommending and non-recommending entries, so a signal containing two keys
// would flip between opposite outcomes run to run") — the same bug lived one
// level up, in the loop that BUILDS the signal list, and the fix there is the
// same rule: most specific (longest) key wins, lexicographic tie-break.
//
// Reachability is not hypothetical: `mortgage-refinance.co.uk` is a live
// domain in the portfolio register's M4 entry — the same entry as the Phase C
// pilot — and "refinance" contains "finance". It matched both "mortgage"
// (recommended) and "finance" (deliberately NOT recommended).
package actions

import (
	"testing"

	"go.uber.org/zap"
)

// TestMatchVerticalDirectory_DomainSignalIsDeterministic pins the outcome for a
// domain that contains two keywords with opposite recommendations. Before the
// fix this flipped; a single run could pass by luck, so it iterates.
//
// It cannot be a "run it once and see" test: map order is randomised PER RANGE
// STATEMENT, so the failure is probabilistic. 200 iterations makes a 50/50 flip
// astronomically unlikely to hide (P(miss) = 2^-199).
func TestMatchVerticalDirectory_DomainSignalIsDeterministic(t *testing.T) {
	logger := zap.NewNop()

	cases := []struct {
		name         string
		domain       string
		wantKind     string
		wantRecommed bool
	}{
		{
			// Both "mortgage" (recommended) and "finance" (NOT recommended)
			// are substrings. Longest-key wins => mortgage, which is also the
			// correct answer: a remortgage/refinance site wants the lender
			// directory.
			name:         "domain contains two opposite keywords",
			domain:       "mortgage-refinance.co.uk",
			wantKind:     "mortgage-lender",
			wantRecommed: true,
		},
		{
			// The Phase C pilot domain: one keyword only, must stay stable.
			name:         "pilot domain, single keyword",
			domain:       "remortgagecalculator.uk",
			wantKind:     "mortgage-lender",
			wantRecommed: true,
		},
		{
			// "finance" alone stays deliberately not-recommended: too generic
			// to pick a provider class.
			name:         "generic finance domain stays not-recommended",
			domain:       "personalfinance.co.uk",
			wantKind:     "",
			wantRecommed: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 200; i++ {
				// Empty industry/site_type/category is the REAL shape: all four
				// comparable finance sites on the estate have industry NULL and
				// site_type "interactive-platform", so the domain signal is the
				// only one that can fire. Measured 2026-08-17.
				got := matchVerticalDirectory("", "", "", tc.domain, logger)
				if got == nil {
					if tc.wantKind == "" && !tc.wantRecommed {
						continue // no match is an acceptable "not recommended"
					}
					t.Fatalf("iteration %d: %s: got no match, want kind=%q recommended=%v",
						i, tc.domain, tc.wantKind, tc.wantRecommed)
				}
				if got.Recommended != tc.wantRecommed {
					t.Fatalf("iteration %d: %s: Recommended=%v (kind %q), want %v — "+
						"a flip here is the map-iteration-order defect, not a flake",
						i, tc.domain, got.Recommended, got.Kind, tc.wantRecommed)
				}
				if tc.wantKind != "" && got.Kind != tc.wantKind {
					t.Fatalf("iteration %d: %s: Kind=%q, want %q",
						i, tc.domain, got.Kind, tc.wantKind)
				}
			}
		})
	}
}

// TestMatchVerticalDirectory_ExplicitSignalStillBeatsDomain guards the ordering
// the action relies on: an explicit classification signal is consulted before
// any domain-derived one, so a site the classifier has actually understood is
// never overridden by a coincidence in its domain string.
func TestMatchVerticalDirectory_ExplicitSignalStillBeatsDomain(t *testing.T) {
	logger := zap.NewNop()

	// industry "savings" on a domain containing "mortgage": the explicit
	// signal must win.
	got := matchVerticalDirectory("savings", "", "", "mortgagecalculator.co.uk", logger)
	if got == nil {
		t.Fatal("got no match for industry=savings")
	}
	if got.Kind != "savings-provider" {
		t.Fatalf("Kind=%q, want savings-provider — the domain signal overrode an explicit one", got.Kind)
	}
}
