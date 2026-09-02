// FILE: platform/orchestration/actions/skip_link_assembly_test.go
//
// Guards the skip link assemblePage emits (improvement_loop lane, 2026-09-02).
//
// WHY THESE TESTS ARE SHAPED THIS WAY. The finding this fixes —
// head_essentials_missing — was 867 rows across 26 sites, and it is graded by a
// detector in another package with its own predicate. A test that asserted my
// own literals would go green while the detector still failed the page, so the
// central assertion here runs THE DETECTOR'S OWN SELECTORS, copied from
// discovery_checks/check_site_structural_validity.go:1065:
//
//	hasSkipLink = doc.Find(`a[href="#content"]`).Length() > 0 || doc.Find(`.skip-link`).Length() > 0
//
// If that predicate is ever narrowed, these tests must be re-read — the
// duplication is deliberate and is the point, because the two packages cannot
// share the helper without the check depending on the renderer.
//
// The ORDER is pinned as hard as the presence. A skip link that is not the
// first focusable element makes a user tab through the whole nav to reach it,
// which is the problem it exists to solve; and a target emitted before the
// header would skip to the wrong place. Both are silent failures — the page
// still contains both elements and the detector still passes.
package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// assembleWithChrome runs the real assemblePage over a page with a header, a
// footer and one rendered section.
func assembleWithChrome(t *testing.T) string {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM site_components").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
			AddRow("head", `<head><title></title></head>`).
			AddRow("header", `<header><nav><a href="/">Home</a></nav></header>`).
			AddRow("footer", `<footer><p>&copy; Example</p></footer>`))
	expectAssemblyQueries(mock, `["hero"]`,
		componentRows().AddRow(
			`<section id="hero"><h1>Example</h1><p>Body copy that is long enough to count as visible content.</p></section>`,
			"hero"))

	page := &PageInfo{
		ID:     uuid.New(),
		SiteID: uuid.New(),
		Name:   "index",
		Title:  "Example",
		URL:    "/index.html",
	}

	html, _, err := assemblePage(context.Background(), db, page, zap.NewNop())
	if err != nil {
		t.Fatalf("assemblePage: %v", err)
	}
	if html == "" {
		t.Fatal("assemblePage returned empty HTML — the fixture is wrong, not the code")
	}
	return html
}

func TestSkipLink_SatisfiesTheDetectorsOwnPredicate(t *testing.T) {
	html := assembleWithChrome(t)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	byHref := doc.Find(`a[href="#content"]`).Length()
	byClass := doc.Find(`.skip-link`).Length()
	if byHref == 0 && byClass == 0 {
		t.Fatalf("head_essentials_missing would still fire: no a[href=\"#content\"] and no .skip-link\n%s", html)
	}
	// Both signals, not just one: the check accepts either, and a site whose
	// stylesheet renames the class must still pass on the href, and vice versa.
	if byHref == 0 {
		t.Error(`the href signal is missing — a[href="#content"] found 0`)
	}
	if byClass == 0 {
		t.Error(`the class signal is missing — .skip-link found 0`)
	}

	// The link must point at something that exists. Without this the page grows
	// a dangling fragment on every URL, which check_phantom_internal_links_
	// fragments files as its own finding — trading one backlog for another.
	if doc.Find(`#content`).Length() == 0 {
		t.Error(`the skip link's target #content does not exist on the page`)
	}
	if _, ok := doc.Find(`#content`).Attr("tabindex"); !ok {
		t.Error(`#content has no tabindex — browsers will move the viewport but not the focus ring`)
	}

	// THE RULES MUST REACH THE PAGE, and this assertion exists because its
	// absence was caught by mutation: deleting the injectSkipLinkCSS call left
	// every other test in this file green while every page on the estate would
	// have rendered a visible "Skip to content" above its header. The link is
	// only invisible because these rules travel with it.
	if !strings.Contains(html, skipLinkMarker) {
		t.Fatal("the skip-link CSS never reached the assembled page — the link would render VISIBLE on every page")
	}
	if !strings.Contains(html, ".skip-link:focus") {
		t.Error("the :focus rule is missing — the link would stay off-screen even when focused, so a sighted keyboard user cannot see where they are")
	}
	if headEnd := strings.Index(html, "</head>"); headEnd < 0 || strings.Index(html, skipLinkMarker) > headEnd {
		t.Error("the CSS block is not inside <head>")
	}
}

func TestSkipLink_IsTheFirstFocusableElementAndTargetsPastTheHeader(t *testing.T) {
	html := assembleWithChrome(t)

	body := strings.Index(html, "<body>")
	anchor := strings.Index(html, skipLinkAnchor)
	header := strings.Index(html, "<header>")
	target := strings.Index(html, skipLinkTarget)
	main := strings.Index(html, "<main>")

	for name, idx := range map[string]int{
		"<body>": body, "skip anchor": anchor, "<header>": header,
		"skip target": target, "<main>": main,
	} {
		if idx < 0 {
			t.Fatalf("%s is absent from the assembled page:\n%s", name, html)
		}
	}

	if !(body < anchor && anchor < header) {
		t.Errorf("the skip link must be the FIRST focusable element: body=%d anchor=%d header=%d", body, anchor, header)
	}
	if !(header < target && target < main) {
		t.Errorf("the target must sit between the header and <main>: header=%d target=%d main=%d", header, target, main)
	}
}

// The demand control. These assertions are only worth anything if they CAN
// fail, and the two above are made over a page this package builds — so the
// obvious way to fool them is a predicate that passes on anything. Feed the
// same assertions a page assembled the way it was BEFORE this change and they
// must fail.
func TestSkipLink_TheAssertionsFailOnAPreChangePage(t *testing.T) {
	preChange := `<!DOCTYPE html><html lang="en"><head><title>Example</title></head>
<body>
<header><nav><a href="/">Home</a></nav></header>
<main><section><h1>Example</h1></section></main>
<footer><p>Example</p></footer>
</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(preChange))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Find(`a[href="#content"]`).Length() != 0 || doc.Find(`.skip-link`).Length() != 0 {
		t.Fatal("the control page already satisfies the detector — it cannot prove anything")
	}
	if strings.Contains(preChange, skipLinkTarget) {
		t.Fatal("the control page already carries the target — it cannot prove anything")
	}
}

func TestInjectSkipLinkCSS_LandsBeforeHeadCloseAndIsIdempotent(t *testing.T) {
	head := `<head><title>x</title></head>`

	once := injectSkipLinkCSS(head)
	if !strings.Contains(once, skipLinkMarker) {
		t.Fatal("the CSS block was not injected")
	}
	if strings.Index(once, skipLinkMarker) > strings.Index(once, "</head>") {
		t.Error("the CSS block landed after </head>")
	}

	twice := injectSkipLinkCSS(once)
	if strings.Count(twice, skipLinkMarker) != 1 {
		t.Errorf("not idempotent: %d copies of the block after a second pass", strings.Count(twice, skipLinkMarker))
	}
}

func TestInjectSkipLinkCSS_HeadWithNoCloseTagStillShipsTheRules(t *testing.T) {
	// A truncated or hand-written head component is real on this estate; the
	// rules must not be silently dropped, or the link renders visible.
	out := injectSkipLinkCSS(`<title>x</title>`)
	if !strings.Contains(out, skipLinkMarker) {
		t.Fatal("the CSS block was dropped when the head had no </head>")
	}
}

// TestSkipLinkCSS_ReferencesOnlyCustomPropertiesTheEstateDefines exists because
// the council's render_guardian seat caught the first version of this block
// referencing `--brand-accent` / `--brand-primary`, which are defined NOWHERE on
// this estate (corr 3c71ec77, medium). The failure mode is the nasty kind: the
// CSS is valid, the fallback fires, the link renders — it simply renders
// hard-coded black-on-white on every site while LOOKING like it inherits the
// brand. Nothing visual is broken enough to notice.
//
// The allow-list is measured, not guessed: fetched from four live stylesheets on
// 2026-09-02 (cookly.uk, finetuning.uk, agritec.uk, webdesign.co.uk), which
// define 51 custom properties between them, `--brand-*` zero times and
// `--color-primary` 12-19 times each. Extend the list only with a name you have
// confirmed the same way — the point of the test is that an invented name fails.
func TestSkipLinkCSS_ReferencesOnlyCustomPropertiesTheEstateDefines(t *testing.T) {
	defined := map[string]bool{
		"--color-primary":      true,
		"--color-primary-text": true,
	}

	refs := regexp.MustCompile(`var\((--[a-z0-9-]+)`).FindAllStringSubmatch(skipLinkCSSBlock, -1)
	if len(refs) == 0 {
		t.Fatal("the block references no custom properties at all — it should inherit the site's brand, not hardcode a colour")
	}
	for _, m := range refs {
		if !defined[m[1]] {
			t.Errorf("skipLinkCSSBlock references %q, which this estate does not define — "+
				"the fallback would fire on every site and the link would be hardcoded, "+
				"looking brand-aware while being nothing of the sort. Confirm the name against "+
				"a live styles.css before adding it to the allow-list.", m[1])
		}
	}

	// Every reference must carry a fallback: a site that has not been through a
	// design run yet defines none of these, and `var(--x)` with no fallback
	// resolves to nothing — an unpainted link on a white ground.
	for _, m := range refs {
		if !regexp.MustCompile(regexp.QuoteMeta("var("+m[1]) + `\s*,`).MatchString(skipLinkCSSBlock) {
			t.Errorf("var(%s) has no fallback — a site with no design run yet would render the link unpainted", m[1])
		}
	}
}
