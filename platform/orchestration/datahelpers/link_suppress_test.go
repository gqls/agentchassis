package datahelpers

import (
	"strings"
	"testing"
)

// The suppression pass (bugs_open/328) removes anchors whose target EXISTS as a
// pages row and has never been served. Every test here pairs the thing that must
// go with something that must stay, in the SAME document, because a pass that
// strips everything satisfies any test that only asserts the removal — and
// "stopped emitting internal links" is a state this platform has reached and not
// noticed (bugs_open/313, closed 2026-08-19).

const servableHref = "/calculators.html"
const refusedHref = "/your-rights.html"

func refusedSet() PageURLSet { return NewPageURLSet([]string{refusedHref}) }

// TestProseAnchorIsUnlinkedAndTheServableSiblingSurvives is the mixed batch: the
// classless prose anchor loses its <a> and keeps its words; the anchor beside it
// whose target serves is byte-identical.
func TestProseAnchorIsUnlinkedAndTheServableSiblingSurvives(t *testing.T) {
	html := `<p>Read the <a href="` + refusedHref + `">guide to your rights</a> before you use our ` +
		`<a href="` + servableHref + `">calculators</a>.</p>`

	got, repairs := SuppressRefusedPageLinks(html, refusedSet())

	if strings.Contains(got, `href="`+refusedHref+`"`) {
		t.Errorf("the refused anchor survived:\n%s", got)
	}
	if !strings.Contains(got, "guide to your rights") {
		t.Errorf("the prose arm dropped the anchor TEXT — it must keep it:\n%s", got)
	}
	// The positive control. Without this a pass that stripped every link would
	// still satisfy the assertion above.
	if !strings.Contains(got, `<a href="`+servableHref+`">calculators</a>`) {
		t.Errorf("the SERVABLE sibling was altered — the pass is stripping links it was not asked to:\n%s", got)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairSuppress {
		t.Errorf("expected exactly one %q repair, got %+v", LinkRepairSuppress, repairs)
	}
}

// TestLabelledControlIsDroppedWhole pins the owner's 2026-08-23 decision. The
// measured shape is a classed card link whose inner content is a short label
// plus an arrow glyph; unlinking it would leave "Read your rights →" as bare
// text in the middle of the card, which is a standing landmine on the sibling
// unlink arm.
func TestLabelledControlIsDroppedWhole(t *testing.T) {
	html := `<div class="card"><h3>Your rights</h3>` +
		`<a class="info-card-grid__card-link" href="` + refusedHref + `">Read your rights` +
		`<em class="info-card-grid__card-link-arrow" aria-hidden="true">&rarr;</em></a></div>`

	got, repairs := SuppressRefusedPageLinks(html, refusedSet())

	if strings.Contains(got, "Read your rights") || strings.Contains(got, "card-link-arrow") {
		t.Errorf("the control arm left orphaned label/arrow text behind — the whole element must go:\n%s", got)
	}
	if !strings.Contains(got, "<h3>Your rights</h3>") {
		t.Errorf("the control arm removed more than the anchor — the card's own content must survive:\n%s", got)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairDropControl {
		t.Errorf("expected exactly one %q repair, got %+v", LinkRepairDropControl, repairs)
	}
}

// TestClassedAnchorWithSentenceInnerTextFallsBackToUnlink is the guard on the
// control arm. A class attribute alone must not license deleting words: a
// template that classes a prose link would otherwise lose the sentence. Both
// conditions are required, and the fallback direction is under-repair.
func TestClassedAnchorWithSentenceInnerTextFallsBackToUnlink(t *testing.T) {
	sentence := "a long editorial sentence about your statutory rights that plainly is not a button label"
	if len([]rune(sentence)) <= suppressControlLabelMax {
		t.Fatalf("test fixture is shorter than the control threshold (%d) and cannot exercise the fallback", suppressControlLabelMax)
	}
	html := `<p><a class="body-prose__link" href="` + refusedHref + `">` + sentence + `</a></p>`

	got, repairs := SuppressRefusedPageLinks(html, refusedSet())

	if !strings.Contains(got, sentence) {
		t.Errorf("a classed anchor with sentence-length text was dropped WHOLE — prose was deleted:\n%s", got)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairSuppress {
		t.Errorf("expected the unlink fallback (%q), got %+v", LinkRepairSuppress, repairs)
	}
}

// TestNonPageAndUnrefusedHrefsAreUntouched — this set can judge page links only.
func TestNonPageAndUnrefusedHrefsAreUntouched(t *testing.T) {
	html := `<a href="https://example.com/your-rights.html">external</a>` +
		`<a href="mailto:hi@example.com">mail</a>` +
		`<a href="#section">fragment</a>` +
		`<a href="/assets/css/styles.css">asset</a>` +
		`<a href="">empty</a>` +
		`<a href="/glossary.html">another page</a>`

	got, repairs := SuppressRefusedPageLinks(html, refusedSet())

	if got != html {
		t.Errorf("output was not byte-identical:\n got: %s\nwant: %s", got, html)
	}
	if len(repairs) != 0 {
		t.Errorf("expected no repairs, got %+v", repairs)
	}
}

// TestEmptySetIsAByteIdenticalNoOp. An empty set means "nothing is refused",
// which is the normal healthy state; the caller is required to turn a FAILED
// lookup into a nil set plus a logged skip rather than an empty one.
func TestEmptySetIsAByteIdenticalNoOp(t *testing.T) {
	html := `<p><a href="` + refusedHref + `">rights</a></p>`
	for name, set := range map[string]PageURLSet{"nil": nil, "empty": NewPageURLSet(nil)} {
		got, repairs := SuppressRefusedPageLinks(html, set)
		if got != html || len(repairs) != 0 {
			t.Errorf("%s set was not a no-op: got %q, %+v", name, got, repairs)
		}
	}
}

// TestRuntimeFillShellIsExempt — a runtime-fill document's hrefs are hydrated
// client-side, so this set cannot judge them. Whole-input, and the fail-safe
// direction for a WRITER is to touch nothing (link_repair.go's own argument).
func TestRuntimeFillShellIsExempt(t *testing.T) {
	html := `<div data-runtime-fill="tool"><a href="` + refusedHref + `">rights</a></div>`
	got, repairs := SuppressRefusedPageLinks(html, refusedSet())
	if got != html || len(repairs) != 0 {
		t.Errorf("a runtime-fill shell was rewritten: got %q, %+v", got, repairs)
	}
}

// TestAnchorQuotedInsideScriptIsNotDeleted is bugs_open/180's hazard, which
// bites this file harder than it bites the unlink arm: BOTH arms here delete
// markup, and the control arm deletes a whole element. MarkupMatches is what
// makes the difference; ReplaceAllString would corrupt valid JavaScript.
func TestAnchorQuotedInsideScriptIsNotDeleted(t *testing.T) {
	html := `<script>var tpl = '<a href="` + refusedHref + `">rights</a>';</script>` +
		`<p><a href="` + refusedHref + `">rights</a></p>`

	got, _ := SuppressRefusedPageLinks(html, refusedSet())

	if !strings.Contains(got, `var tpl = '<a href="`+refusedHref+`">rights</a>';`) {
		t.Errorf("an anchor quoted inside <script> was treated as markup and rewritten:\n%s", got)
	}
	if strings.Contains(got, `<p><a href="`+refusedHref+`">`) {
		t.Errorf("the real anchor in the body was NOT suppressed:\n%s", got)
	}
}

// TestNormalisationMatchesTheIndexForm — the set is built from stored pages.url
// values, and a writer may emit the same target with a trailing slash or an
// index.html tail. NormalizePagePath is the one normal form; this pins that the
// suppression set uses it rather than a raw string compare.
func TestNormalisationMatchesTheIndexForm(t *testing.T) {
	set := NewPageURLSet([]string{"/guides/index.html"})
	for _, href := range []string{"/guides/index.html", "/guides/", "/guides"} {
		html := `<p><a href="` + href + `">guides</a></p>`
		got, repairs := SuppressRefusedPageLinks(html, set)
		if len(repairs) != 1 {
			t.Errorf("href %q did not resolve to the refused target: got %q", href, got)
		}
	}
}
