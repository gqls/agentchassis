// FILE: platform/orchestration/datahelpers/cta_positional_test.go
//
// Pins bugs_open/436's lever at the RANKING — the binding point the 391 review
// ruled on (constraint 1): the loaders SELECT eligibility and decide nothing;
// RankCTAPositionalCandidates is where "this page must never be a CTA
// destination" becomes true, for all three of the ranking's callers at once —
// the build-time resolver, the rerender recompute, and the site header
// fallback, whose output is never persisted and therefore has no diff for any
// content_data check to see.

package datahelpers

import "testing"

// fossilSite is 391's shape reduced to a fixture: an off-topic tool whose
// nav_order (1, set at page creation and never looked at again) beats every
// sibling, plus the on-topic pack at the default 100.
func fossilSite() []CTAPositionalCandidate {
	return []CTAPositionalCandidate{
		{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html", Area: "tools", NavOrder: 1},
		{Name: "tool-agent-cost-model", Title: "Agent Cost Modeller", URL: "/tools/agent-cost-model.html", Area: "tools", NavOrder: 100},
		{Name: "tool-risk-checker", Title: "Deployment Risk Checker", URL: "/tools/risk-checker.html", Area: "tools", NavOrder: 100},
	}
}

// TestRankRefusesAnOptedOutPageBothDirections is bugs_open/436's own
// verification bar: "Seed a site with two tools where the nav_order-first one
// is off-topic … assert the primary CTA is NOT it — then flip the opt-out and
// assert it IS eligible again. Both directions in one run."
func TestRankRefusesAnOptedOutPageBothDirections(t *testing.T) {
	pages := fossilSite()

	// Direction 1: no flag set (the zero value) — today's behaviour, the
	// fossil wins. This is also the safety property of the INVERTED field: a
	// candidate built with no knowledge of the flag must be eligible, or the
	// flag's arrival would have silently emptied every ranking fleet-wide.
	ranked := RankCTAPositionalCandidates("", pages)
	if len(ranked) != 3 || ranked[0].Name != "tool-password-entropy" {
		t.Fatalf("baseline: expected the fossil to win as today, got %+v", ranked)
	}

	// Direction 2: the owner's lever. Opting the fossil out hands rank-1 to
	// the next candidate; the page is dropped, not demoted.
	pages[0].IneligibleAsCTATarget = true
	ranked = RankCTAPositionalCandidates("", pages)
	if len(ranked) != 2 {
		t.Fatalf("opted-out page must leave the ranking entirely, got %+v", ranked)
	}
	if ranked[0].Name != "tool-agent-cost-model" {
		t.Errorf("rank-1 after opt-out = %q, want tool-agent-cost-model", ranked[0].Name)
	}

	// And back: flipping the flag off re-admits the page with no other state.
	pages[0].IneligibleAsCTATarget = false
	ranked = RankCTAPositionalCandidates("", pages)
	if len(ranked) != 3 || ranked[0].Name != "tool-password-entropy" {
		t.Errorf("flag off must restore today's ranking exactly, got %+v", ranked)
	}
}

// TestRankHeaderFormRefusesAnOptedOutPage asserts the header fallback's exact
// call shape (render_site_components_action.go:190 — pageName "") separately,
// as bugs_open/436 requires: that caller's output is never persisted, so this
// unit is the only cheap place its behaviour is visible at all. pageName ""
// must not read as "no page to exclude, so exclude nothing".
func TestRankHeaderFormRefusesAnOptedOutPage(t *testing.T) {
	pages := fossilSite()
	pages[0].IneligibleAsCTATarget = true
	ranked := RankCTAPositionalCandidates("", pages)
	for _, c := range ranked {
		if c.Name == "tool-password-entropy" {
			t.Fatalf("header-form ranking (pageName \"\") still offers the opted-out page: %+v", ranked)
		}
	}
	if len(ranked) == 0 || ranked[0].Name != "tool-agent-cost-model" {
		t.Errorf("header-form rank-1 = %+v, want tool-agent-cost-model", ranked)
	}
}

// TestRankPreservesThePreLeverBehaviour pins the three filters and the sort
// that predate the lever, byte for byte: excluded areas, self-exclusion, and
// (NavOrder, Name) ordering. If this fails after an edit to the ranking, the
// edit changed more than eligibility.
func TestRankPreservesThePreLeverBehaviour(t *testing.T) {
	pages := []CTAPositionalCandidate{
		{Name: "contact-index", Title: "Contact us", URL: "/contact/index.html", Area: "contact", NavOrder: 1},
		{Name: "services", Title: "Services", URL: "/services.html", NavOrder: 5},
		{Name: "guides", Title: "Guides", URL: "/guides/index.html", Area: "guides", NavOrder: 5},
		{Name: "blog", Title: "Blog", URL: "/blog/index.html", Area: "blog", NavOrder: 20},
	}
	ranked := RankCTAPositionalCandidates("services", pages)
	if len(ranked) != 2 {
		t.Fatalf("want contact excluded (area) and services excluded (self), got %+v", ranked)
	}
	// NavOrder tie at 5 is gone (services excluded); guides at 5 precedes blog at 20.
	if ranked[0].Name != "guides" || ranked[1].Name != "blog" {
		t.Errorf("ordering changed: got [%s %s], want [guides blog]", ranked[0].Name, ranked[1].Name)
	}
	if CTAExcludedDestination("/legal/privacy.html") != true || CTAExcludedDestination("/tools/x.html") != false {
		t.Error("CTAExcludedDestination changed meaning")
	}
}
