// FILE: platform/orchestration/actions/nav_declaration_test.go
//
// bugs_open/407 — pins the per-site header declaration.
//
// TWO TESTS ARE LOAD-BEARING AND WERE WRITTEN FIRST:
//
//	TestUndeclaredSitePrimaryOrderUnchanged  the 51-site no-op, and it pins ORDER
//	                                         rather than membership — membership is
//	                                         already pinned in nav_membership_test.go
//	                                         and would not catch a mis-ordered fallback
//	TestDeclarationGivesATotalOrderWithinATier  the SAME-TIER tie, which is half the
//	                                         measured damage and the half a
//	                                         tier-bump design cannot fix
//
// Every precedence test carries its own UNDECLARED CONTROL over the same corpus,
// so a mutation that deletes the declared branch fails the declared half while the
// control half still passes. That is deliberate: a refusal or precedence test that
// only asserts "the outcome differs" cannot tell its own mechanism from the next
// one along, which is a mistake this session already made once today and wrote up.
//
// ⚠ TWO OF THESE ROWS WERE FALSE WHEN FIRST WRITTEN — the second time in one day,
// and the cause was different from the first, which is the part worth recording.
// This morning's pair failed because the tests were passing on a SECOND failure
// downstream. This pair failed because the FIXTURES DID NOT EXERCISE THE PATH:
//
//   - the total-order test declared ["pricing-transparency","service-areas"],
//     which is already alphabetical, so a mutation sorting the declared names
//     changed nothing and the ordering guarantee could be deleted unnoticed;
//   - the de-duplication test used a corpus in which every page was in_header, so
//     UTILITY WAS EMPTY and there was nothing to duplicate — the guard could be
//     removed and the test still passed.
//
// Neither is a "guard in series". A test can be perfectly targeted, assert exactly
// the right property, and still be vacuous because its INPUT cannot produce the
// failure. Both fixtures now carry an explicit setup assertion that fails loudly
// if the corpus stops exercising the path — a fixture that has quietly stopped
// discriminating is indistinguishable from a guard that works.
//
//	mutation                                            test that catches it
//	--------------------------------------------------  -----------------------------------------
//	drop the declared branch (always fallback)          DeclaredSlotBeatsTier1
//	honour only the tier, not the declared ORDER        DeclarationGivesATotalOrderWithinATier
//	run the declared branch on an EMPTY declaration     UndeclaredSitePrimaryOrderUnchanged
//	let a declared page stay in utility as well         DeclaredPageIsNotDuplicatedAcrossGroups
//	require in_header before placing a declared page    DeclaredMembershipDoesNotRequireInHeader
//	swallow a declared name that resolves to nothing    DeclaredNameWithNoPageIsReportedNotSilent
//	promote a declared system or legal page             DeclaredSystemAndLegalPagesAreIneligible
//	refuse to override neverPrimaryTypes                DeclarationOverridesNeverPrimaryType
//	refuse to override the child-URL bar                DeclarationOverridesChildURLBar
//	ignore the per-site cap                             PerSiteMaxHeaderItemsOverridesStepConfig
//	treat a malformed declaration as "no declaration"   SiteConfigChromeParseShapes
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// --- corpora --------------------------------------------------------------

// mixedTierCorpus spans all three tiers so an ordering assertion over it is
// discriminating: tier 1 (index, about, contact), tier 2 (pricing, blog) and
// tier 3 (approach) — finetuning.uk's real shape, which is the case the owner hit.
func mixedTierCorpus() []pageNavInfo {
	mk := func(name, url string, order int) pageNavInfo {
		p := navTestPage(name, url, "content", true, false)
		p.NavOrder = order
		return p
	}
	return []pageNavInfo{
		mk("approach", "/approach.html", 4), // tier 3, and the site says it matters
		mk("index", "/index.html", 1),       // tier 1
		mk("about", "/about.html", 10),      // tier 1
		mk("contact", "/contact.html", 20),  // tier 1
		mk("pricing", "/pricing.html", 30),  // tier 2
		mk("blog", "/blog.html", 40),        // tier 2
	}
}

// fourTierThreeTies is gaswholesalers.com's real shape: FOUR tier-3 pages, ALL at
// nav_order 100, today resolved by load order. This is the corpus a tier-bump
// design cannot help.
func fourTierThreeTies() []pageNavInfo {
	mk := func(name string) pageNavInfo {
		p := navTestPage(name, "/"+name+".html", "content", true, false)
		p.NavOrder = 100
		return p
	}
	return []pageNavInfo{
		mk("why-gas-wholesalers"), mk("how-pricing-works"),
		mk("service-areas"), mk("pricing-transparency"),
	}
}

func navOrderOf(pages []pageNavInfo) []string { return navNames(pages) }

func requireOrderPrefix(t *testing.T, got []pageNavInfo, want ...string) {
	t.Helper()
	names := navOrderOf(got)
	if len(names) < len(want) {
		t.Fatalf("primary has %d pages, want at least %d: %v", len(names), len(want), names)
	}
	for i, w := range want {
		if names[i] != w {
			t.Fatalf("primary[%d] = %q, want %q (full order: %v)", i, names[i], w, names)
		}
	}
}

// --- the 51-site no-op ----------------------------------------------------

// TestUndeclaredSitePrimaryOrderUnchanged is the proof that bugs_open/407's change
// is inert for every site which has not declared a header. [MEASURED 2026-08-26]
// that is all 51 of them.
//
// It pins the ORDER, not the membership. A change that ran the declared branch on
// an empty declaration and reordered the fallback would pass every membership
// assertion in this package and still be a fleet-wide regression.
func TestUndeclaredSitePrimaryOrderUnchanged(t *testing.T) {
	primary, _, _, rep := classifyPagesForNavDeclared(mixedTierCorpus(), siteNavDeclaration{}, zap.NewNop())

	// tier 1 by nav_order, then tier 2 by nav_order, then tier 3.
	requireOrderPrefix(t, primary, "index", "about", "contact", "pricing", "blog", "approach")

	if len(rep.Placed)+len(rep.Missing)+len(rep.Ineligible)+len(rep.FlagDisagreed) != 0 {
		t.Fatalf("an undeclared site must produce an EMPTY declaration report, got %+v", rep)
	}
}

// TestUndeclaredResultMatchesTheLegacyEntryPoint pins the wrapper: the exported
// signature nav_membership_test.go uses must agree with the declared path given an
// empty declaration, or that file's zero diff proves nothing.
func TestUndeclaredResultMatchesTheLegacyEntryPoint(t *testing.T) {
	corpus := mixedTierCorpus()
	legacyPrimary, legacyLegal, legacyUtility := classifyPagesForNav(corpus, zap.NewNop())
	newPrimary, newLegal, newUtility, _ := classifyPagesForNavDeclared(corpus, siteNavDeclaration{}, zap.NewNop())

	for _, pair := range []struct {
		name string
		a, b []pageNavInfo
	}{{"primary", legacyPrimary, newPrimary}, {"legal", legacyLegal, newLegal}, {"utility", legacyUtility, newUtility}} {
		if strings.Join(navNames(pair.a), ",") != strings.Join(navNames(pair.b), ",") {
			t.Fatalf("%s differs between the legacy wrapper and the declared path with an empty declaration:\n  legacy %v\n  new    %v",
				pair.name, navNames(pair.a), navNames(pair.b))
		}
	}
}

// --- precedence -----------------------------------------------------------

// TestDeclaredSlotBeatsTier1 is finetuning.uk's case: a tier-3 page the SITE says
// matters, against a header full of tier 1 and 2. The control in the same corpus
// is what makes it discriminating — undeclared, that page ranks LAST.
func TestDeclaredSlotBeatsTier1(t *testing.T) {
	decl := siteNavDeclaration{HeaderSlots: []string{"approach", "index"}, Source: navDeclSourceSiteConfig}
	primary, _, _, rep := classifyPagesForNavDeclared(mixedTierCorpus(), decl, zap.NewNop())
	requireOrderPrefix(t, primary, "approach", "index")
	if strings.Join(rep.Placed, ",") != "approach,index" {
		t.Fatalf("report must name what it placed, in order; got %v", rep.Placed)
	}

	control, _, _, _ := classifyPagesForNavDeclared(mixedTierCorpus(), siteNavDeclaration{}, zap.NewNop())
	names := navNames(control)
	if names[len(names)-1] != "approach" {
		t.Fatalf("CONTROL: undeclared, the tier-3 page must rank LAST — if it does not, this test is "+
			"not measuring the declaration. order=%v", names)
	}
}

// TestDeclarationGivesATotalOrderWithinATier is gaswholesalers.com's case and the
// reason the declaration is an ordered array rather than a priority hint: four
// tier-3 pages at identical nav_order, where a tier bump changes nothing at all.
func TestDeclarationGivesATotalOrderWithinATier(t *testing.T) {
	// ⚠ THE DECLARED ORDER IS DELIBERATELY NOT ALPHABETICAL, AND NOT THE CORPUS
	// ORDER. A fixture whose declared order coincides with either is satisfied by a
	// mutation that re-sorts the list, and the first version of this test was
	// exactly that: sorting the declared names alphabetically changed nothing and
	// the test passed with the ordering guarantee gone. Declared "service-areas"
	// before "pricing-transparency" (reverse-alphabetical) and both after
	// "why-gas-wholesalers" in the corpus, so only the SITE's order produces this.
	decl := siteNavDeclaration{
		HeaderSlots: []string{"service-areas", "pricing-transparency"},
		Source:      navDeclSourceSiteConfig,
	}
	primary, _, _, _ := classifyPagesForNavDeclared(fourTierThreeTies(), decl, zap.NewNop())
	requireOrderPrefix(t, primary, "service-areas", "pricing-transparency")

	// CONTROL: every page is tier 3 at nav_order 100, so without a declaration the
	// site has no way whatsoever to express which of the four it cares about.
	control, _, _, _ := classifyPagesForNavDeclared(fourTierThreeTies(), siteNavDeclaration{}, zap.NewNop())
	if len(control) != 4 {
		t.Fatalf("CONTROL: expected all four tied pages as candidates, got %v", navNames(control))
	}
}

// TestDeclarationOverridesNeverPrimaryType is idea.uk's case, and the one the
// mechanism replay identified as a THIRD cause: /report.html is page_type "tool"
// on a FLAT url, so neverPrimaryTypes bars it and no rebuild would ever place it.
func TestDeclarationOverridesNeverPrimaryType(t *testing.T) {
	page := navTestPage("report", "/report.html", "tool", true, false)
	page.NavOrder = 3
	corpus := append(mixedTierCorpus(), page)

	decl := siteNavDeclaration{HeaderSlots: []string{"report"}, Source: navDeclSourceSiteConfig}
	primary, _, utility, _ := classifyPagesForNavDeclared(corpus, decl, zap.NewNop())
	requireOrderPrefix(t, primary, "report")
	if navContains(utility, "report") {
		t.Fatal("a declared page must appear in exactly one group; it is in primary AND utility")
	}

	control, _, controlUtility, _ := classifyPagesForNavDeclared(corpus, siteNavDeclaration{}, zap.NewNop())
	if navContains(control, "report") || !navContains(controlUtility, "report") {
		t.Fatalf("CONTROL: undeclared, a page_type=tool page must be barred from primary and land in "+
			"utility — if it does not, this test is not measuring the override. primary=%v utility=%v",
			navNames(control), navNames(controlUtility))
	}
}

// TestDeclarationOverridesChildURLBar — the same authority, for the URL-shaped bar.
func TestDeclarationOverridesChildURLBar(t *testing.T) {
	page := navTestPage("gripper-payload-calculator", "/tools/gripper-payload-calculator.html", "content", true, false)
	corpus := append(mixedTierCorpus(), page)

	decl := siteNavDeclaration{HeaderSlots: []string{"gripper-payload-calculator"}, Source: navDeclSourceSiteConfig}
	primary, _, _, _ := classifyPagesForNavDeclared(corpus, decl, zap.NewNop())
	requireOrderPrefix(t, primary, "gripper-payload-calculator")

	control, _, controlUtility, _ := classifyPagesForNavDeclared(corpus, siteNavDeclaration{}, zap.NewNop())
	if navContains(control, "gripper-payload-calculator") || !navContains(controlUtility, "gripper-payload-calculator") {
		t.Fatalf("CONTROL: undeclared, a /tools/ page must stay out of primary and land in utility. primary=%v utility=%v",
			navNames(control), navNames(controlUtility))
	}
}

// TestDeclaredMembershipDoesNotRequireInHeader pins the re-plan-survival semantics.
// sync_pages_to_db upserts `in_header = EXCLUDED.in_header`, so a declaration that
// required the flag would be silently defeated by the next re-plan — which is this
// bug, returning.
func TestDeclaredMembershipDoesNotRequireInHeader(t *testing.T) {
	page := navTestPage("your-own-model", "/your-own-model.html", "content", false, false)
	corpus := append(mixedTierCorpus(), page)

	decl := siteNavDeclaration{HeaderSlots: []string{"your-own-model"}, Source: navDeclSourceSiteConfig}
	primary, _, _, rep := classifyPagesForNavDeclared(corpus, decl, zap.NewNop())
	requireOrderPrefix(t, primary, "your-own-model")
	if strings.Join(rep.FlagDisagreed, ",") != "your-own-model" {
		t.Fatalf("the flag disagreement must be REPORTED so the plan can be corrected at its source; got %v",
			rep.FlagDisagreed)
	}

	control, _, _, _ := classifyPagesForNavDeclared(corpus, siteNavDeclaration{}, zap.NewNop())
	if navContains(control, "your-own-model") {
		t.Fatalf("CONTROL: with in_header=false and no declaration the page must NOT be primary; order=%v",
			navNames(control))
	}
}

// TestDeclaredPageIsNotDuplicatedAcrossGroups — a page appears in exactly one group.
func TestDeclaredPageIsNotDuplicatedAcrossGroups(t *testing.T) {
	// ⚠ THE CORPUS MUST CONTAIN A PAGE THAT CLASSIFICATION SENDS TO UTILITY, or
	// there is nothing to duplicate and the de-duplication guard can be deleted
	// with this test still passing. The first version declared only in_header
	// pages and was exactly that. `report` (page_type tool) and `footer-only`
	// both land in utility undeclared; `report` is then declared, so the guard is
	// the only thing keeping it out of two groups.
	corpus := append(mixedTierCorpus(),
		navTestPage("report", "/report.html", "tool", true, true),
		navTestPage("footer-only", "/footer-only.html", "content", false, true),
	)
	decl := siteNavDeclaration{HeaderSlots: []string{"report", "blog", "index"}, Source: navDeclSourceSiteConfig}
	primary, legal, utility, _ := classifyPagesForNavDeclared(corpus, decl, zap.NewNop())

	if !navContains(primary, "report") {
		t.Fatalf("setup is wrong: the declared page must reach primary, or this test measures nothing. primary=%v",
			navNames(primary))
	}
	if !navContains(utility, "footer-only") {
		t.Fatalf("setup is wrong: utility must be non-empty, or the de-duplication guard is unexercised. utility=%v",
			navNames(utility))
	}
	seen := map[string]int{}
	for _, group := range [][]pageNavInfo{primary, legal, utility} {
		for _, p := range group {
			seen[p.Name]++
		}
	}
	for name, n := range seen {
		if n != 1 {
			t.Fatalf("page %q appears in %d groups; it must appear in exactly one", name, n)
		}
	}
	// and once WITHIN primary
	if len(navNames(primary)) != len(mixedTierCorpus())+1 {
		t.Fatalf("primary has %d entries, want %d (the six in_header pages plus the declared tool page) "+
			"— a declared candidate was duplicated: %v",
			len(primary), len(mixedTierCorpus())+1, navNames(primary))
	}
}

// --- refusals, each reported rather than silent ---------------------------

func TestDeclaredNameWithNoPageIsReportedNotSilent(t *testing.T) {
	decl := siteNavDeclaration{HeaderSlots: []string{"a-page-that-does-not-exist", "index"}, Source: navDeclSourceSiteConfig}
	primary, _, _, rep := classifyPagesForNavDeclared(mixedTierCorpus(), decl, zap.NewNop())

	if strings.Join(rep.Missing, ",") != "a-page-that-does-not-exist" {
		t.Fatalf("a declared name that resolves to nothing must be REPORTED — asserted on the return "+
			"value, because a log line is not a record. got %v", rep.Missing)
	}
	requireOrderPrefix(t, primary, "index") // the remaining slots are unaffected
}

func TestDeclaredSystemAndLegalPagesAreIneligible(t *testing.T) {
	corpus := append(mixedTierCorpus(),
		navTestPage("privacy", "/privacy.html", "content", true, true),
		navTestPage("404", "/404.html", "content", true, false),
	)
	decl := siteNavDeclaration{HeaderSlots: []string{"privacy", "404", "index"}, Source: navDeclSourceSiteConfig}
	primary, legal, _, rep := classifyPagesForNavDeclared(corpus, decl, zap.NewNop())

	if navContains(primary, "privacy") || navContains(primary, "404") {
		t.Fatalf("system and legal pages stay ineligible even when declared — those are CORRECTNESS, "+
			"not preference. primary=%v", navNames(primary))
	}
	if !navContains(legal, "privacy") {
		t.Fatalf("a declared legal page must stay in the legal group, not vanish; legal=%v", navNames(legal))
	}
	if len(rep.Ineligible) != 2 {
		t.Fatalf("both refusals must be reported with a reason; got %v", rep.Ineligible)
	}
	requireOrderPrefix(t, primary, "index")
}

// --- the parse ------------------------------------------------------------

func TestSiteConfigChromeParseShapes(t *testing.T) {
	// The input is the site_config aspect's `data` jsonb. Note the cases that
	// carry header_cta_url alongside: the declaration must coexist with the header
	// config already living under `chrome` on three sites, and must not be
	// disturbed by keys it does not own.
	cases := []struct {
		name       string
		raw        string
		wantSlots  string
		wantCap    int
		wantSource string
	}{
		{"no spec row at all", ``, "", 0, navDeclSourceDefault},
		{"empty data", `{}`, "", 0, navDeclSourceDefault},
		{"other aspects' keys only", `{"analytics":{"gtm":"GTM-X"},"locale":"en-GB"}`, "", 0, navDeclSourceDefault},
		{"chrome present, no slots (oufe's real shape)", `{"chrome":{"header_cta_url":"/cases/index.html","header_cta_label":"Read the cases"}}`, "", 0, navDeclSourceDefault},
		{"a real declaration beside the CTA keys", `{"chrome":{"header_cta_url":"/x.html","header_slots":["Index","Your-Own-Model"],"max_header_items":9}}`, "index,your-own-model", 9, navDeclSourceSiteConfig},
		{"cap only", `{"chrome":{"max_header_items":6}}`, "", 6, navDeclSourceSiteConfig},
		{"cap as a numeric string", `{"chrome":{"header_slots":["index"],"max_header_items":"9"}}`, "index", 9, navDeclSourceSiteConfig},
		{"empty slot array", `{"chrome":{"header_slots":[]}}`, "", 0, navDeclSourceDefault},
		{"chrome is not an object", `{"chrome":"index,about"}`, "", 0, navDeclSourceInvalid},
		{"slots is not an array", `{"chrome":{"header_slots":"index"}}`, "", 0, navDeclSourceInvalid},
		{"a non-string member is skipped, not fatal", `{"chrome":{"header_slots":["index",7,"about"]}}`, "index,about", 0, navDeclSourceInvalid},
		{"duplicates collapse and are flagged", `{"chrome":{"header_slots":["index","Index"]}}`, "index", 0, navDeclSourceInvalid},
		{"cap zero is refused", `{"chrome":{"header_slots":["index"],"max_header_items":0}}`, "index", 0, navDeclSourceInvalid},
		{"cap negative is refused", `{"chrome":{"header_slots":["index"],"max_header_items":-1}}`, "index", 0, navDeclSourceInvalid},
		{"cap fractional is refused", `{"chrome":{"header_slots":["index"],"max_header_items":2.5}}`, "index", 0, navDeclSourceInvalid},
		{"cap non-numeric string is refused", `{"chrome":{"header_slots":["index"],"max_header_items":"lots"}}`, "index", 0, navDeclSourceInvalid},
		{"data is not json", `not json`, "", 0, navDeclSourceInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := siteNavDeclarationFromSiteConfig([]byte(tc.raw))
			if strings.Join(got.HeaderSlots, ",") != tc.wantSlots {
				t.Fatalf("slots = %v, want %q", got.HeaderSlots, tc.wantSlots)
			}
			if got.MaxHeaderItems != tc.wantCap {
				t.Fatalf("cap = %d, want %d", got.MaxHeaderItems, tc.wantCap)
			}
			if got.Source != tc.wantSource {
				t.Fatalf("source = %q, want %q — a declaration that is present and unreadable must NEVER "+
					"read as %q, because that is the silent-ignore this whole change removes",
					got.Source, tc.wantSource, navDeclSourceDefault)
			}
		})
	}
}

func TestPerSiteMaxHeaderItemsOverridesStepConfig(t *testing.T) {
	if got := (siteNavDeclaration{MaxHeaderItems: 9}).EffectiveCap(8); got != 9 {
		t.Fatalf("the site's own cap must win over the fleet step config; got %d", got)
	}
	if got := (siteNavDeclaration{}).EffectiveCap(8); got != 8 {
		t.Fatalf("with no per-site cap the step config must apply unchanged; got %d", got)
	}
	if got := (siteNavDeclaration{HeaderSlots: []string{"index"}}).EffectiveCap(8); got != 8 {
		t.Fatalf("declaring SLOTS must not silently change the cap; got %d", got)
	}
}
