// FILE: platform/orchestration/actions/discovery_checks/cta_classify_anchor_test.go
//
// Pins ctaClassifyAnchor — the single definition of "this CTA is misdirected",
// shared by MisdirectedCTACheck.Run (detection) and VerifyPageRerenderResolved
// (completion verification).
//
// Why this test exists: the predicate was EXTRACTED from Run's inline loop on
// 2026-07-20 so verification could reuse it. The package's pre-existing tests
// covered only ctaTokens and ctaAreaExcluded, so they would have passed
// unchanged even if the extraction had altered the logic — "the tests pass" was
// not evidence. These cases pin the behaviour the inline version had.
//
// The three-outcome contract is load-bearing: Run distinguishes "names no page"
// (falls through to its unknown-destination rules) from "names a page, href
// agrees" (healthy, skip). Collapsing those two would make every generic link a
// phantom-destination finding.

package discovery_checks

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func ctaTestPages() []ctaMatchPage {
	return []ctaMatchPage{
		{
			ID: "1", Name: "gauntlet", Title: "The Gauntlet", URL: "/gauntlet.html",
			Interactive: true,
			tokens:      map[string]bool{"gauntlet": true},
		},
		{
			ID: "2", Name: "pricing", Title: "Pricing", URL: "/pricing.html",
			tokens: map[string]bool{"pricing": true},
		},
	}
}

func TestCTAClassifyAnchor(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		href       string
		wantNamed  bool
		wantMisdir bool
		wantTarget string
	}{
		{
			// The case the check was built for: copy names a real page, href
			// points at a different real page (so the phantom check is blind).
			name: "copy names one page, href points elsewhere",
			text: "Enter the Gauntlet", href: "/contact.html",
			wantNamed: true, wantMisdir: true, wantTarget: "/gauntlet.html",
		},
		{
			name: "copy names a page and href agrees",
			text: "Enter the Gauntlet", href: "/gauntlet.html",
			wantNamed: true, wantMisdir: false,
		},
		{
			// Must be named=false, NOT a healthy match — Run relies on this to
			// apply its unknown-destination rules.
			name: "copy names no real page",
			text: "Widget Emporium", href: "/nowhere.html",
			wantNamed: false, wantMisdir: false,
		},
		{
			// "Learn More" reduces to zero distinctive tokens. If this ever
			// returned named=true it would match arbitrary pages and flood the
			// queue with false misdirects.
			name: "generic text names nothing",
			text: "Learn More", href: "/contact.html",
			wantNamed: false, wantMisdir: false,
		},
		{
			name: "generic text with only stopwords",
			text: "Get Started Today", href: "/pricing.html",
			wantNamed: false, wantMisdir: false,
		},
		{
			// Normalisation is applied to BOTH sides, so a query string must not
			// make an otherwise-correct link read as misdirected.
			name: "query string does not make a correct href misdirected",
			text: "Pricing", href: "/pricing.html?utm_source=nav",
			wantNamed: true, wantMisdir: false,
		},
		{
			name: "fragment does not make a correct href misdirected",
			text: "Enter the Gauntlet", href: "/gauntlet.html#top",
			wantNamed: true, wantMisdir: false,
		},
		{
			// NOTE (2026-07-20): NormalizePagePath does not insert a leading
			// slash, so a RELATIVE href ("pricing.html") would not equal a page
			// url ("/pricing.html") and would read as a misdirect. Not asserted
			// as either correct or buggy here because it is unreachable with
			// current data — a live scan found 0 components with relative
			// internal hrefs against 169 with absolute ones — and the behaviour
			// predates the extraction this file pins. If relative hrefs ever
			// start being emitted, this is where the false positives will come
			// from.
			name: "trailing slash is normalised away",
			text: "Pricing", href: "/pricing.html/",
			wantNamed: true, wantMisdir: false,
		},
		{
			// Interactive pages outrank content pages as match candidates.
			name: "interactive page preferred as suggested target",
			text: "gauntlet pricing", href: "/contact.html",
			wantNamed: true, wantMisdir: true, wantTarget: "/gauntlet.html",
		},
	}

	pages := ctaTestPages()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, named := ctaClassifyAnchor(
				datahelpers.Anchor{Text: tc.text, Href: tc.href}, "hero", pages)

			if named != tc.wantNamed {
				t.Fatalf("named = %v, want %v", named, tc.wantNamed)
			}
			if (got != nil) != tc.wantMisdir {
				t.Fatalf("misdirect = %v, want %v", got != nil, tc.wantMisdir)
			}
			if tc.wantMisdir {
				if got.SuggestedTarget != tc.wantTarget {
					t.Errorf("suggested target = %q, want %q", got.SuggestedTarget, tc.wantTarget)
				}
				if got.Href != tc.href || got.Text != tc.text {
					t.Errorf("misdirect did not carry the anchor through: %+v", got)
				}
				if got.SlotName != "hero" {
					t.Errorf("slot name = %q, want %q", got.SlotName, "hero")
				}
			}
		})
	}
}
