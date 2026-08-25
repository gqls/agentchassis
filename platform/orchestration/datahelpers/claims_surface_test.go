// FILE: platform/orchestration/datahelpers/claims_surface_test.go
//
// The page-type gate on the prose number scan (bugs_open/102).
//
// Every fixture in the editorial test below is a VERBATIM live false positive,
// measured 2026-07-28 with cmd/claimscan against each opted-in site's own
// evidence register over its live rendered_html: 124 unregistered-number
// findings fleet-wide, of which 59 sit on the editorial page types kept below and
// every one is a worked example, widget help text, or a quoted third-party figure. The
// business-surface fixtures are live TRUE positives from the same run, so the
// two directions are graded against the same corpus.
//
// The four properties this file pins, in the order they matter:
//  1. an editorial page raises no prose number findings;
//  2. the SAME text on a business page (and on an unknown surface) still does —
//     without this, "quiet" and "broken" are indistinguishable;
//  3. a BANNED claim is still raised on an editorial page — the case that
//     motivated the whole check was "70+ agents across eight functional
//     departments" found on a guide;
//  4. the types that LOOK editorial and are deliberately not (blog-index,
//     section-index, report, and since bugs_open/364 entity-directory and
//     entity-page) stay scanned, each with its reason attached, so adding one
//     back is a deliberate act rather than a tidy-up.
//
// Extended 2026-08-24 (bugs_open/364) with the tracker/directory three, measured
// the same way: 20 findings at ZERO precision, every one a third party's figure
// in an aggregated listing. That addition knowingly fails the second half of the
// membership bar — see editorialPageTypes' comment and
// TestTrackerPagesGiveUpTheirFirstPersonClaims, which pins what it costs.

package datahelpers

import "testing"

const surfaceTestEB = `{
  "audit_doc": "docs/x/AUDIT.md",
  "facts": [
    {"id": "F1", "claim": "orchestration state records", "value": 90790, "kind": "metric",
     "source": {"sql": "SELECT count(*) FROM orchestration_states"},
     "verified_at": "2026-07-28", "tolerance": "gte", "context_terms": ["orchestration", "state record"]}
  ],
  "banned_claims": [
    {"pattern": "eight (functional |business )?(departments|functions)", "reason": "L0 audit U1: invented taxonomy"}
  ]
}`

func mustParseSurfaceEB(t *testing.T) *EvidenceBase {
	t.Helper()
	eb, err := ParseEvidenceBase([]byte(surfaceTestEB))
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if eb == nil {
		t.Fatal("ParseEvidenceBase returned nil for a populated register")
	}
	return eb
}

// Live false positives, one per editorial page type, with the site they were
// measured on. Each one is business-shaped to the lexical gate and is not a
// claim about the business.
var editorialFalsePositives = []struct {
	pageType string
	site     string
	block    string
}{
	{"blog-post", "gamesdesign.co.uk",
		"In a game with 10,000 active players farming that item, roughly 180 of them will hit that wall."},
	{"blog-post", "ai-agent-orchestration.com",
		"For HTTP-based agents, expose a endpoint that returns 200 if the agent's main loop is running."},
	{"guide", "gamesdesign.co.uk",
		"Genshin Impact uses a hard Pity Timer at 90 pulls for five-star items."},
	{"tool", "gamesdesign.co.uk",
		"Cumulative % of players who have received the item by each hour of play. " +
			"The shaded band shows where the middle 50% of players get it (25th–75th percentile)."},
	{"game", "gamesdesign.co.uk",
		"Project the net accumulation across 10, 20, and 40 hours. If the curve climbs, " +
			"you have an inflation problem."},
	{"news-index", "robot-hands.com",
		"[Insights] Market report projects cobot tending cells to hold 38% share, driven by demand from small manufacturing customers."},
}

// The SAME six live false positives, now keyed by the COMPONENT they sit in
// rather than the page (RFC_053 / bugs_open/364 Phase 2). Measured 2026-08-24 on
// ai-agent-orchestration.com; every one is a THIRD PARTY's figure in an
// aggregated listing, tripping businessClaimContextRe on `agents?` because the
// site is about agents. Note two are not statistics at all — a version string,
// and a digit inside an acronym.
//
// They moved out of editorialPageTypes' table deliberately: gating these by PAGE
// took the same pages' first-person hero and call-to-action with them, which is
// the cost TestTrackerPageHeroAndCTAAreScannedAgain now pins as RECOVERED.
var thirdPartyComponentFalsePositives = []struct {
	component string
	pageType  string
	site      string
	block     string
}{
	{"adoption-tracker-listing", "adoption-tracker", "ai-agent-orchestration.com",
		"rollout_scope Over 80% of Fortune 500 deploying active agents built with Copilot Studio or Agent Framework source"},
	{"adoption-tracker-listing", "adoption-tracker", "ai-agent-orchestration.com",
		"roi_claimed only 95 of 1,837 respondents reported AI agents live in production respondents source"},
	{"protocol-tracker-listing", "protocol-tracker", "ai-agent-orchestration.com",
		"agent_framework JSON-RPC 2.0 client-server with Tools, Resources, Prompts, and Sampling"},
	{"protocol-tracker-listing", "protocol-tracker", "ai-agent-orchestration.com",
		"Agent-to-Agent Protocol (A2A) Linux Foundation"},
	{"model-directory-listing", "model-directory", "ai-agent-orchestration.com",
		"protocol_adopted Salesforce Headless 360 platform routes customer and agent interactions via MCP source"},
}

// The types that look editorial and are NOT, each for its own reason — pinned so
// that adding one back is a deliberate act with evidence attached, not a tidy-up.
// See editorialPageTypes' comment; all three were settled by measurement after
// council round 1 objected to 'blog-index' as an unmeasured extrapolation.
var notEditorialPageTypes = []struct {
	pageType string
	because  string
}{
	{"blog-index", "never measured: 3 pages fleet-wide, zero findings even against an empty register"},
	{"section-index", "2 of its 20 pages are about-index / contact-index — marketing bodies under an index name"},
	{"report", "its false positives are model numbers in product names, a different mechanism"},
	{"entity-directory", "bugs_open/364: 4 pages fleet-wide, ZERO measured findings — analogy to the trackers is not a measurement"},
	{"entity-page", "bugs_open/364: 21 pages fleet-wide, ZERO measured findings — same bar that keeps blog-index out"},
}

func TestTypesDeliberatelyNotEditorialStayScanned(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	block := "We hold 45,000 client records across the estate."
	for _, tc := range notEditorialPageTypes {
		if f := eb.ScanUnregisteredNumbers([]string{block}, ClaimSurface{PageType: tc.pageType}); len(f) != 1 {
			t.Errorf("page_type %q must still be scanned (%s), got %+v", tc.pageType, tc.because, f)
		}
	}
}

// (1) The editorial types raise nothing from prose numbers.
func TestEditorialPageTypesRaiseNoProseNumbers(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	for _, fp := range editorialFalsePositives {
		findings := eb.ScanUnregisteredNumbers([]string{fp.block}, ClaimSurface{PageType: fp.pageType})
		if len(findings) != 0 {
			t.Errorf("%s (%s): teaching content flagged as a business claim: %+v",
				fp.pageType, fp.site, findings)
		}
	}
}

// (2) THE NEGATIVE CONTROL. Identical text on a business surface, and on an
// unknown one, is still scanned. A fix graded only on findings going down is
// indistinguishable from a scanner that was switched off.
func TestSameTextOnBusinessSurfaceIsStillScanned(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	for _, fp := range editorialFalsePositives {
		for _, surface := range []ClaimSurface{
			{PageType: "content"},       // the type carrying the live TRUE positives
			{PageType: "landing"},       //
			{PageType: "report"},        // deliberately NOT editorial — see claims.go
			{PageType: "section-index"}, // ditto: about-index / contact-index live here
			{PageType: "blog-index"},    // ditto: never measured
			{},                          // unknown: site chrome, or no page in hand
			{PageType: "some-type-nobody-has-invented-yet"},
		} {
			if f := eb.ScanUnregisteredNumbers([]string{fp.block}, surface); len(f) == 0 {
				t.Errorf("page_type %q: the scan went quiet on business-surface text %q",
					surface.PageType, fp.block)
			}
		}
	}
}

// A live TRUE positive keeps flagging, and a registered figure keeps not
// flagging, on a business surface — the fix must not touch either.
func TestBusinessSurfaceUnchangedByTheFix(t *testing.T) {
	eb := mustParseSurfaceEB(t)

	unregistered := "We hold 45,000 client records across the estate."
	if f := eb.ScanUnregisteredNumbers([]string{unregistered}, ClaimSurface{PageType: "content"}); len(f) != 1 {
		t.Errorf("an unregistered business figure must still flag on a content page, got %+v", f)
	}

	registered := "The platform has processed 90,790 orchestration state records."
	if f := eb.ScanUnregisteredNumbers([]string{registered}, ClaimSurface{PageType: "content"}); len(f) != 0 {
		t.Errorf("a REGISTERED figure must not flag: %+v", f)
	}
}

// (3) THE REGRESSION CONTROL. ScanBannedClaims is not surface-gated, and the
// reason is historical: check_unverified_claims' first live run (2026-07-16)
// found "70+ agents across eight functional departments" on a GUIDE. A fix that
// skipped editorial pages wholesale would have un-caught exactly that.
// ScanBannedClaims takes no ClaimSurface AT ALL — that absence is the property,
// and a loop over page types here would be a check that cannot fail. So the
// assertion is the discriminating pair instead: on a guide, one block yields the
// banned finding and NOT the heuristic number finding.
func TestBannedClaimsAreStillCaughtOnEditorialPages(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	block := "At Leopardess, we hold 45,000 client records across eight functional departments."
	guide := ClaimSurface{PageType: "guide"}

	if f := eb.ScanBannedClaims([]string{block}); len(f) != 1 {
		t.Fatalf("the banned claim must still be caught on a guide, got %+v", f)
	}
	if f := eb.ScanUnregisteredNumbers([]string{block}, guide); len(f) != 0 {
		t.Fatalf("the heuristic number scan must be silent on a guide, got %+v", f)
	}
	// And the same block on a business page yields both.
	content := ClaimSurface{PageType: "content"}
	if f := eb.ScanUnregisteredNumbers([]string{block}, content); len(f) != 1 {
		t.Fatalf("the number must still flag on a content page, got %+v", f)
	}
}

// Page types arrive from a varchar column and from workflow config; neither is
// normalised at the source.
func TestPageTypeMatchingIsCaseAndSpaceInsensitive(t *testing.T) {
	for _, pt := range []string{"guide", "Guide", "GUIDE", "  blog-post  ", "\tTool\n"} {
		if (ClaimSurface{PageType: pt}).ProseNumbersAreClaims() {
			t.Errorf("page_type %q should be editorial", pt)
		}
	}
	for _, pt := range []string{"", "content", "landing", "report", "entity-page", "guides", "blogpost",
		"blog-index", "section-index"} {
		if !(ClaimSurface{PageType: pt}).ProseNumbersAreClaims() {
			t.Errorf("page_type %q must stay scanned — unknown and business types are noisy by design", pt)
		}
	}
}

// The stat scan is structural, not lexical, so it is NOT surface-gated: a stat
// card on a guide is still a published figure in a claim-shaped field. Pinned
// here because "the guide fix" is exactly the change someone would later extend
// to the stat path by analogy.
func TestStatScanIsNotSurfaceGated(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	claims := ExtractStatClaims("stats-band", map[string]interface{}{
		"stat1_value": "4,200",
		"stat1_label": "Client Deployments",
	})
	if len(claims) == 0 {
		t.Fatal("precondition: the stat extractor should see this card")
	}
	findings := eb.ScanStatClaims(claims)
	if len(findings) == 0 {
		t.Error("a stat card's figure must be audited whatever page it sits on — " +
			"if this now depends on page type, bugs_open/102's boundary has been widened wrongly")
	}
}

// ============================================================================
// COMPONENT GRAIN (RFC_053 / bugs_open/364 Phase 2)
// ============================================================================

// (5) A declared third-party data component raises nothing from prose numbers,
// wherever it sits.
func TestThirdPartyComponentsRaiseNoProseNumbers(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	for _, fp := range thirdPartyComponentFalsePositives {
		findings := eb.ScanUnregisteredNumbers([]string{fp.block},
			ClaimSurface{PageType: fp.pageType, ComponentFunction: fp.component})
		if len(findings) != 0 {
			t.Errorf("%s on %s (%s): a third party's figure flagged as OUR claim: %+v",
				fp.component, fp.pageType, fp.site, findings)
		}
	}
}

// (6) THE NEGATIVE CONTROL, and the one that matters most here. The identical
// text in an UNDECLARED component on the same page is still scanned. Without
// this, "the listing went quiet" and "the scan went quiet" are the same
// observation — which is the failure mode this whole file exists to prevent.
func TestSameTextInAnUndeclaredComponentIsStillScanned(t *testing.T) {
	eb := mustParseSurfaceEB(t)
	for _, fp := range thirdPartyComponentFalsePositives {
		for _, component := range []string{
			"hero",              // the marketing surface on the very same page
			"call-to-action",    // ditto
			"case-studies-list", // a list of OUR OWN work — measured, three genuine claims
			"generic-text-block",
			"",                                    // unknown: no component in hand
			"a-component-nobody-has-invented-yet", // a new template must stay noisy
		} {
			f := eb.ScanUnregisteredNumbers([]string{fp.block},
				ClaimSurface{PageType: fp.pageType, ComponentFunction: component})
			if len(f) == 0 {
				t.Errorf("component %q on %s: the scan went quiet on text that is only exempt "+
					"inside %q — a component-grain exemption must not leak to its neighbours",
					component, fp.pageType, fp.component)
			}
		}
	}
}

// (7) THE COST THE INTERIM PAID, NOW RECOVERED — this is the inversion the
// interim's own test demanded rather than a deletion.
//
// From 2026-08-22 to 2026-08-25 the three tracker page types sat in
// editorialPageTypes, which silenced the prose number scan on those pages'
// first-person hero and call-to-action along with their third-party listing.
// TestTrackerPagesGiveUpTheirFirstPersonClaims asserted that loss deliberately,
// so it would stay visible, and its failure message said: "if Phase 2
// (component-grain surface) has landed, INVERT this test rather than deleting
// it." This is that inversion. If it ever fails, page-grain gating has come
// back and the site's own claims are unpoliced again on those pages.
func TestTrackerPageHeroAndCTAAreScannedAgain(t *testing.T) {
	const claim = "We hold 45,000 client records across the estate."

	empty := &EvidenceBase{}
	for _, pt := range []string{"adoption-tracker", "protocol-tracker", "model-directory"} {
		for _, component := range []string{"hero", "call-to-action"} {
			f := empty.ScanUnregisteredNumbers([]string{claim},
				ClaimSurface{PageType: pt, ComponentFunction: component})
			if len(f) == 0 {
				t.Errorf("page_type %q, component %q: a first-person quantified claim is NOT scanned — "+
					"the page-grain interim has returned and its blind spot with it (RFC_053)", pt, component)
			}
		}
		// And the listing on that same page stays exempt, so this is a real
		// separation and not the exclusion simply having been removed.
		if f := empty.ScanUnregisteredNumbers([]string{claim},
			ClaimSurface{PageType: pt, ComponentFunction: pt + "-listing"}); len(f) != 0 {
			t.Errorf("page_type %q: the listing component should still be exempt, got %+v", pt, f)
		}
	}
}

// (8) The three page types must NOT be in editorialPageTypes any more. Pinned
// separately from the behaviour above because re-adding one would look like a
// tidy-up and would silently undo Phase 2 while every other test still passed.
func TestTrackerPageTypesAreNoLongerPageGated(t *testing.T) {
	for _, pt := range []string{"adoption-tracker", "protocol-tracker", "model-directory"} {
		if !(ClaimSurface{PageType: pt}).ProseNumbersAreClaims() {
			t.Errorf("page_type %q is page-gated again — Phase 2 moved this decision to the "+
				"component (thirdPartyDataComponents); re-adding the page type takes the hero "+
				"and CTA with it, which is the breach the interim recorded", pt)
		}
	}
}

// TestThePluralOfOrchestrationReachesTheGate pins bugs_open/364 §5b.
//
// businessClaimContextRe is an allow-list of NOUNS, and it carried
// `orchestration` singular with no `s?` — so "We run over 1,600 orchestrations a
// day across 13 live production systems" (verbatim live copy, protocol-tracker
// call-to-action, ai-agent-orchestration.com) was never scanned at all. Nothing
// reported that: a false NEGATIVE in this layer is silent by construction.
//
// The council raised it as a fast-follow on the interim (correlation
// b8df25dc, bug_historian and compliance seats): leaving it unfixed meant the
// interim's "measured loss is zero" rested on a SECOND bug staying unfixed,
// which is luck, not a safety property.
//
// The empty register is deliberate, for the usual reason — a populated one would
// support the value and the assertion would hold whatever the gate did.
func TestThePluralOfOrchestrationReachesTheGate(t *testing.T) {
	const claim = "We run over 1,600 orchestrations a day across 13 live production systems."

	empty := &EvidenceBase{}
	if f := empty.ScanUnregisteredNumbers([]string{claim}, ClaimSurface{PageType: "content"}); len(f) == 0 {
		t.Error("the plural 'orchestrations' must reach businessClaimContextRe — " +
			"a first-person quantified claim went unscanned because the gate carried only the singular")
	}

	// The singular must keep working — this is the direction that was never broken,
	// and a fix graded only on the new case cannot see if it broke the old one.
	const singular = "Our orchestration count reached 4,200 last week."
	if f := empty.ScanUnregisteredNumbers([]string{singular}, ClaimSurface{PageType: "content"}); len(f) == 0 {
		t.Error("regression: the singular 'orchestration' no longer reaches the gate")
	}
}
