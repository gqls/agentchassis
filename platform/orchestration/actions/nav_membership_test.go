package actions

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// THE INVARIANT THESE TESTS PIN (bugs_open/149 A2, fixed 2026-07-31):
//
//	pages.in_header / in_footer DECLARES nav membership. A page's URL shape may
//	decide WHERE it appears. It may never decide WHETHER it appears.
//
// Before the fix, classifyPagesForNav dropped any page under /tools/, /blog/,
// /guides/ … with a bare `continue` placed BEFORE either flag was read. The
// consequence was not a cosmetic one: check_orphan_pages raises `nav_drift` for a
// flagged page missing from nav, routes it to nav-updater, and nav-updater's first
// step is populate_nav_tables — so the item completed having placed nothing, and
// the next sweep raised it again. Measured on the live fleet 2026-07-31: a
// nav_drift item for gamesdesign.co.uk naming four /tools/ pages was `complete` on
// 07-29 and all four were still absent from site_nav_items two days later, while
// the same check and handler had repaired robot-hands.com's /learning-center.html
// and /news.html — whose only difference is the URL prefix.
//
// The tests are deliberately about the CLASSIFIER and not about SQL: the defect
// was one ordering decision between two in-memory rules, so that is the unit worth
// pinning. TestChildPageURLCannotVetoADeclaredFlag is the one that fails if the
// URL-shaped rule ever regains the power to drop a flagged page.

func navTestPage(name, url, pageType string, inHeader, inFooter bool) pageNavInfo {
	return pageNavInfo{
		ID:       uuid.New(),
		Name:     name,
		Title:    name,
		URL:      url,
		PageType: pageType,
		InHeader: inHeader,
		InFooter: inFooter,
	}
}

func navNames(pages []pageNavInfo) []string {
	out := make([]string, 0, len(pages))
	for _, p := range pages {
		out = append(out, p.Name)
	}
	return out
}

func navContains(pages []pageNavInfo, name string) bool {
	for _, p := range pages {
		if p.Name == name {
			return true
		}
	}
	return false
}

// TestChildPageURLCannotVetoADeclaredFlag is the regression test for the defect.
// Every case is a real fleet row as at 2026-07-31.
func TestChildPageURLCannotVetoADeclaredFlag(t *testing.T) {
	logger := zap.NewNop()

	cases := []struct {
		name        string
		page        pageNavInfo
		wantUtility bool
		why         string
	}{
		{
			name:        "tool page, in_header only (gamesdesign bayesian-ranking)",
			page:        navTestPage("bayesian-ranking", "/tools/bayesian-ranking.html", "tool", true, false),
			wantUtility: true,
			why:         "named by a nav_drift item that completed on 07-29 and was still absent on 07-31",
		},
		{
			name:        "tool page, both flags (gamesdesign tool-drop-rate-tuner)",
			page:        navTestPage("tool-drop-rate-tuner", "/tools/tool-drop-rate-tuner.html", "tool", true, true),
			wantUtility: true,
			why:         "same item, same outcome",
		},
		{
			name:        "tool page, in_footer only (oufe tool-recovery-waterfall)",
			page:        navTestPage("tool-recovery-waterfall", "/tools/tool-recovery-waterfall/index.html", "tool", false, true),
			wantUtility: true,
			why:         "the /tools/<name>/index.html URL convention must behave like the flat one",
		},
		{
			name:        "guide page under /guides/ (vetcomparison guide-independent-strategy)",
			page:        navTestPage("guide-independent-strategy", "/guides/guide-independent-strategy.html", "guide", false, true),
			wantUtility: true,
			why:         "the defect was never tool-specific — every child prefix was affected",
		},
		{
			name:        "tool page with NO flag declared (leopardess password-entropy)",
			page:        navTestPage("password-entropy", "/tools/password-entropy.html", "tool", false, false),
			wantUtility: false,
			why:         "declaring nothing must still mean no nav item — the fix widens what the flags CAN say, not what silence says",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			primary, legal, utility := classifyPagesForNav([]pageNavInfo{tc.page}, logger)

			if navContains(primary, tc.page.Name) {
				t.Fatalf("child page %q reached PRIMARY nav; a child page is never primary (the parent listing represents it in the main menu). primary=%v",
					tc.page.Name, navNames(primary))
			}
			if len(legal) != 0 {
				t.Fatalf("child page %q was classified LEGAL: %v", tc.page.Name, navNames(legal))
			}

			got := navContains(utility, tc.page.Name)
			if got != tc.wantUtility {
				t.Fatalf("child page %q in utility = %v, want %v (%s). utility=%v",
					tc.page.Name, got, tc.wantUtility, tc.why, navNames(utility))
			}
		})
	}
}

// TestSectionIndexChildStaysPrimaryEligible protects the exception that was
// already correct and must not be collateral damage: a section-index page sits
// under the section prefix but is the section's PARENT, so it belongs in the main
// menu. Keyed on page_type, never on URL.
func TestSectionIndexChildStaysPrimaryEligible(t *testing.T) {
	logger := zap.NewNop()

	for _, pageType := range []string{"section-index", "blog-index", "news-index", "entity-directory"} {
		t.Run(pageType, func(t *testing.T) {
			page := navTestPage("tools-index", "/tools/index.html", pageType, true, true)
			primary, _, utility := classifyPagesForNav([]pageNavInfo{page}, logger)

			if !navContains(primary, page.Name) {
				t.Fatalf("%s page under a child prefix did not reach primary nav: primary=%v utility=%v",
					pageType, navNames(primary), navNames(utility))
			}
		})
	}
}

// TestNeverPrimaryTypeOutsideAChildPathStillGoesToUtility pins the OTHER half of
// the unified rule — the branch that already worked. A tool page at a flat URL
// (three fleet sites use /tool-name.html) must behave identically to one under
// /tools/, because the whole point of the fix is that URL shape stopped being a
// separate decision.
func TestNeverPrimaryTypeOutsideAChildPathStillGoesToUtility(t *testing.T) {
	logger := zap.NewNop()

	page := navTestPage("tool-llm-cost-calculator", "/tool-llm-cost-calculator.html", "tool", true, false)
	primary, _, utility := classifyPagesForNav([]pageNavInfo{page}, logger)

	if navContains(primary, page.Name) {
		t.Fatalf("a tool page reached primary nav: %v", navNames(primary))
	}
	if !navContains(utility, page.Name) {
		t.Fatalf("a flagged tool page at a flat URL did not reach utility: utility=%v", navNames(utility))
	}
}

// TestOrdinaryPagesAreUnaffected is the control. The fix touched one ordering
// decision, and the cheapest way to be wrong about it is to change the answer for
// pages that were never in scope — which is how a nav fix empties a header.
func TestOrdinaryPagesAreUnaffected(t *testing.T) {
	logger := zap.NewNop()

	pages := []pageNavInfo{
		navTestPage("index", "/index.html", "content", true, true),
		navTestPage("about", "/about.html", "content", true, true),
		navTestPage("contact", "/contact.html", "content", true, true),
		navTestPage("privacy", "/privacy.html", "content", false, true),
		navTestPage("careers", "/careers.html", "content", false, true),
		navTestPage("404", "/404.html", "content", true, true),
	}

	primary, legal, utility := classifyPagesForNav(pages, logger)

	for _, want := range []string{"index", "about", "contact"} {
		if !navContains(primary, want) {
			t.Fatalf("%q missing from primary nav: %v", want, navNames(primary))
		}
	}
	if !navContains(legal, "privacy") {
		t.Fatalf("privacy missing from legal: %v", navNames(legal))
	}
	if !navContains(utility, "careers") {
		t.Fatalf("careers (in_footer only) missing from utility: %v", navNames(utility))
	}
	if navContains(primary, "404") || navContains(utility, "404") || navContains(legal, "404") {
		t.Fatalf("system page 404 reached nav: primary=%v utility=%v legal=%v",
			navNames(primary), navNames(utility), navNames(legal))
	}
}
