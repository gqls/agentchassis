// FILE: platform/orchestration/actions/discovery_checks/static_attribute_checks_test.go
//
// The markup in these tests is not invented. Every fixture is a verbatim
// element taken from a page fetched live on 2026-07-28, named in the test that
// uses it, because this workstream has already been burned once by a shape
// invented alongside its own fixtures: the fixtures agreed with the code and
// every real caller disagreed with both.
package discovery_checks

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func parseForTest(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("fixture will not parse: %v", err)
	}
	return doc
}

func runAttr(t *testing.T, html string, ch criteriaCheck) attributeOutcome {
	t.Helper()
	return evaluateAttributeCheck(parseForTest(t, html), ch)
}

// verbatim from https://vonc.com/provocations/index.html, fetched 2026-07-28.
// The hidden template row the loader clones — it ships href="#".
const voncTemplateRow = `<div class="provocations-archive__list">` +
	`<a class="provocations-archive__item" data-archive-template hidden href="#"></a>` +
	`</div>`

// verbatim from https://fundamentallyai.com/capabilities.html, fetched
// 2026-07-28. Four carousel cards, each with a real destination.
const fundamentallyCarouselCards = `<div data-hcc-track>` +
	`<a class="hero-card-carousel__card" href="/capabilities.html#review-council"></a>` +
	`<a class="hero-card-carousel__card" href="/capabilities.html#verification"></a>` +
	`<a class="hero-card-carousel__card" href="/capabilities.html#rapid-delivery"></a>` +
	`<a class="hero-card-carousel__card" href="/capabilities.html#embeddings"></a>` +
	`</div>`

// TestAttributeCheck_ZeroMatchesSkips is the rule the whole capability rests on.
//
// The inert rows of a feed-driven list are cloned client-side, so at Tier 2 the
// selector for them matches nothing. "No inert row carries a dead href" is then
// vacuously true — and a vacuous pass is the exact defect the register exists to
// end. It is equally not a failure: failing there would call a correct page
// broken, which is what happened on 2026-07-28 when a check already recorded as
// impossible was executed anyway.
func TestAttributeCheck_ZeroMatchesSkips(t *testing.T) {
	out := runAttr(t, voncTemplateRow, criteriaCheck{
		ID: "no_dead_row_hrefs", Type: "attribute_absent",
		Selector: ".provocations-archive__item--inert", Attributes: []string{"href", "tabindex"},
	})
	if out.verdict != attrSkip {
		t.Fatalf("zero matched elements must SKIP, got verdict %v (%s)", out.verdict, out.detail)
	}
	if !strings.Contains(out.detail, "nothing was asserted") {
		t.Errorf("a skip must say that nothing was asserted; got %q", out.detail)
	}
}

// TestAttributeCheck_AbsentFailsOnRealPlaceholderHref uses the real vonc
// template row, which ships href="#". The check must refute it: this is the
// element the entry's own clause calls "a dead control that no visitor can see
// but every sweep can find".
func TestAttributeCheck_AbsentFailsOnRealPlaceholderHref(t *testing.T) {
	out := runAttr(t, voncTemplateRow, criteriaCheck{
		ID: "template_row_not_a_control", Type: "attribute_absent",
		Selector: ".provocations-archive__item", Attributes: []string{"href"},
	})
	if out.verdict != attrFail {
		t.Fatalf("a served href=\"#\" must FAIL attribute_absent, got verdict %v (%s)", out.verdict, out.detail)
	}
	for _, want := range []string{"1 of 1", `href="#"`} {
		if !strings.Contains(out.detail, want) {
			t.Errorf("detail must name the count and the offending value; %q missing from %q", want, out.detail)
		}
	}
}

func TestAttributeCheck_AbsentPassesWhenTrulyAbsent(t *testing.T) {
	out := runAttr(t, `<div class="provocations-archive__list">`+
		`<a class="provocations-archive__item" data-archive-template hidden></a></div>`,
		criteriaCheck{
			ID: "template_row_not_a_control", Type: "attribute_absent",
			Selector: ".provocations-archive__item", Attributes: []string{"href", "tabindex"},
		})
	if out.verdict != attrPass {
		t.Fatalf("want pass, got verdict %v (%s)", out.verdict, out.detail)
	}
	if !strings.Contains(out.detail, "none of 1 element") {
		t.Errorf("a pass must carry the element count — it is the only thing separating a real assertion from a vacuous one; got %q", out.detail)
	}
}

// TestAttributeCheck_MatchesOnLiveCards is the payoff case: six of the eight
// attribute checks authored across the register are of this shape, on
// server-rendered brochure components.
func TestAttributeCheck_MatchesOnLiveCards(t *testing.T) {
	out := runAttr(t, fundamentallyCarouselCards, criteriaCheck{
		ID: "cards_have_real_destinations", Type: "attribute_matches",
		Selector: ".hero-card-carousel__card", Attribute: "href", NotMatches: `^(#|\s*)$`,
	})
	if out.verdict != attrPass {
		t.Fatalf("four real destinations must PASS, got verdict %v (%s)", out.verdict, out.detail)
	}
	if !strings.Contains(out.detail, "all 4 element(s)") {
		t.Errorf("detail must name how many elements were asserted over; got %q", out.detail)
	}
}

func TestAttributeCheck_MatchesRefutesPlaceholderDestination(t *testing.T) {
	// One card degraded to a placeholder href — the no-inert-control defect.
	html := strings.Replace(fundamentallyCarouselCards, `href="/capabilities.html#verification"`, `href="#"`, 1)
	out := runAttr(t, html, criteriaCheck{
		ID: "cards_have_real_destinations", Type: "attribute_matches",
		Selector: ".hero-card-carousel__card", Attribute: "href", NotMatches: `^(#|\s*)$`,
	})
	if out.verdict != attrFail {
		t.Fatalf("a placeholder href must FAIL, got verdict %v (%s)", out.verdict, out.detail)
	}
	if !strings.Contains(out.detail, "1 of 4") || !strings.Contains(out.detail, "element 2") {
		t.Errorf("detail must locate the offender among its peers; got %q", out.detail)
	}
}

// TestAttributeCheck_MissingAttributeFails: once an element has matched, the
// contract says it carries the attribute. Absence is the strongest form of not
// matching, and treating it as a pass would make "cards are links" green for a
// grid of divs.
func TestAttributeCheck_MissingAttributeFails(t *testing.T) {
	out := runAttr(t, `<div><span class="stat-band__value" data-countup></span></div>`,
		criteriaCheck{
			ID: "accessible_name_carries_the_real_figure", Type: "attribute_matches",
			Selector: ".stat-band__value", Attribute: "aria-label", Matches: `\S`,
		})
	if out.verdict != attrFail {
		t.Fatalf("a missing attribute must FAIL, got verdict %v (%s)", out.verdict, out.detail)
	}
	if !strings.Contains(out.detail, "has no aria-label attribute") {
		t.Errorf("detail must say the attribute is absent rather than that it mismatched; got %q", out.detail)
	}
}

// TestAttributeCheck_AssertsNothingSkips — a check that names no attribute, or
// no pattern, must never report a pass. The write-time validator refuses these
// outright (P10); this is the runtime backstop for criteria that never went
// through it, such as a hand-authored tool-doc fence.
func TestAttributeCheck_AssertsNothingSkips(t *testing.T) {
	cases := map[string]criteriaCheck{
		"absent with no attributes": {
			ID: "x", Type: "attribute_absent", Selector: ".hero-card-carousel__card"},
		"matches with no attribute": {
			ID: "x", Type: "attribute_matches", Selector: ".hero-card-carousel__card", Matches: `\S`},
		"matches with no pattern": {
			ID: "x", Type: "attribute_matches", Selector: ".hero-card-carousel__card", Attribute: "href"},
	}
	for name, ch := range cases {
		out := runAttr(t, fundamentallyCarouselCards, ch)
		if out.verdict != attrSkip {
			t.Errorf("%s: must SKIP, got verdict %v (%s)", name, out.verdict, out.detail)
		}
		if !strings.Contains(out.detail, "asserts nothing") {
			t.Errorf("%s: detail must say so plainly; got %q", name, out.detail)
		}
	}
}

// TestAttributeCheck_BadPatternIsACriteriaDefectNotAPageDefect — an
// uncompilable regex or selector says nothing about the page. Failing there
// would blame the site for a typo in the criteria document, which is the
// harness-lies-about-the-page class this workstream keeps finding.
func TestAttributeCheck_BadPatternIsACriteriaDefectNotAPageDefect(t *testing.T) {
	for name, ch := range map[string]criteriaCheck{
		"bad regex": {
			ID: "x", Type: "attribute_matches",
			Selector: ".hero-card-carousel__card", Attribute: "href", NotMatches: `^(unclosed`},
		"bad selector": {
			ID: "x", Type: "attribute_absent",
			Selector: `a[href=`, Attributes: []string{"href"}},
	} {
		out := runAttr(t, fundamentallyCarouselCards, ch)
		if out.verdict != attrSkip {
			t.Errorf("%s: must SKIP, got verdict %v (%s)", name, out.verdict, out.detail)
		}
		if !strings.Contains(out.detail, "criteria defect, not a page defect") {
			t.Errorf("%s: detail must say whose fault it is; got %q", name, out.detail)
		}
	}
}

// TestAttributeCheck_MalformedSelectorIsNotTheSameAsNoMatch guards the reason
// the selector is compiled explicitly instead of being handed to goquery.Find:
// Find returns an empty selection for a malformed selector, which would make a
// typo indistinguishable from the ordinary client-side-rendering case.
func TestAttributeCheck_MalformedSelectorIsNotTheSameAsNoMatch(t *testing.T) {
	malformed := runAttr(t, fundamentallyCarouselCards, criteriaCheck{
		ID: "x", Type: "attribute_absent", Selector: `a[href=`, Attributes: []string{"href"}})
	noMatch := runAttr(t, fundamentallyCarouselCards, criteriaCheck{
		ID: "x", Type: "attribute_absent", Selector: `.nothing-here`, Attributes: []string{"href"}})
	if malformed.detail == noMatch.detail {
		t.Fatal("a malformed selector and an honest zero-match report the same thing — the reason cascadia.Compile is called explicitly has been undone")
	}
}

// TestEvaluateStaticCriteria_AttributeChecksFlowThrough proves the wiring, not
// just the helper: the switch in evaluateStaticCriteria must route these types
// to the DOM path and project the three verdicts onto the right lists.
func TestEvaluateStaticCriteria_AttributeChecksFlowThrough(t *testing.T) {
	doc := criteriaDoc{Checks: []criteriaCheck{
		{ID: "template_row_not_a_control", Type: "attribute_absent",
			Selector: ".provocations-archive__item", Attributes: []string{"href"}},
		{ID: "no_dead_row_hrefs", Type: "attribute_absent",
			Selector: ".provocations-archive__item--inert", Attributes: []string{"href"}},
	}}
	// data-runtime-fill suppresses the built-in dead-control sweep, so the only
	// outcomes here are the two attribute checks'.
	ev := evaluateStaticCriteria(doc, 200, `<div data-runtime-fill="true">`+voncTemplateRow+`</div>`)
	if len(ev.failed) != 1 || ev.failed[0].id != "template_row_not_a_control" {
		t.Errorf("want the served placeholder href failed, got failed=%v", ev.failed)
	}
	if len(ev.skipped) != 1 || ev.skipped[0].id != "no_dead_row_hrefs" {
		t.Errorf("want the client-side-rendered rows skipped, got skipped=%v", ev.skipped)
	}
	if len(ev.passed) != 0 {
		t.Errorf("nothing here should pass, got passed=%v", ev.passed)
	}
}

// TestEveryStaticCheckTypeIsClassified is the enforcement the council gate's
// architecture seat asked for when it objected that "a doc comment is not an
// enforcement mechanism — nothing stops a third future check type from being
// added confirm-style by someone who never reads that comment".
//
// It reads the `case` labels out of evaluateStaticCriteria's own switch — the
// same source-lockstep technique the capability table uses — and fails the build
// for any handled type that is not classified as confirming or refuting. The
// point is not the table; it is that ADDING A CHECK TYPE NOW FAILS UNTIL YOU
// DECIDE WHICH GUARANTEE IT GIVES, which is the moment you are made to read the
// rule. A classification that can be skipped is a comment with extra steps.
func TestEveryStaticCheckTypeIsClassified(t *testing.T) {
	src, err := os.ReadFile(filepath.Clean("check_tool_acceptance.go"))
	if err != nil {
		t.Fatalf("cannot read the evaluator (this test runs from the package dir): %v", err)
	}
	body := string(src)
	if i := strings.Index(body, "func evaluateStaticCriteria("); i >= 0 {
		body = body[i:]
		if j := strings.Index(body[1:], "\nfunc "); j >= 0 {
			body = body[:j+1]
		}
	} else {
		t.Fatal("evaluateStaticCriteria not found — the evaluator's shape changed; fix this test, not the table")
	}

	caseRE := regexp.MustCompile(`(?m)^\s*case\s+((?:"[a-z_]+"\s*,?\s*)+):`)
	valueRE := regexp.MustCompile(`"([a-z_]+)"`)
	handled := map[string]bool{}
	for _, m := range caseRE.FindAllStringSubmatch(body, -1) {
		for _, v := range valueRE.FindAllStringSubmatch(m[1], -1) {
			handled[v[1]] = true
		}
	}
	if len(handled) == 0 {
		t.Fatal("no check types found in the evaluator's switch — fix this test rather than the table")
	}

	for typ := range handled {
		c, r := experienceStaticConfirming[typ], experienceStaticRefuting[typ]
		switch {
		case c && r:
			t.Errorf("check type %q is classified BOTH confirming and refuting — it gives one guarantee or the other", typ)
		case !c && !r:
			t.Errorf("check type %q is handled by evaluateStaticCriteria but classified in neither table.\n"+
				"  Decide before shipping it: CONFIRMING means it must never fail a page for markup the browser builds;\n"+
				"  REFUTING means it may fail a page for what it actually SERVES, and must SKIP (never pass, never fail)\n"+
				"  when its selector matches nothing. Tier 2's guarantee is mixed, and this is where that is recorded.", typ)
		}
	}
	// And the other direction: a classified type nothing implements is a stale
	// entry, which is how a table starts lying.
	for typ := range experienceStaticRefuting {
		if !handled[typ] {
			t.Errorf("%q is classified refuting but the evaluator does not handle it", typ)
		}
	}
	for typ := range experienceStaticConfirming {
		if !handled[typ] {
			t.Errorf("%q is classified confirming but the evaluator does not handle it", typ)
		}
	}
}
