package datahelpers

import (
	"strings"
	"testing"
)

// The page set these tests repair against. Deliberately mixed: a .html fleet
// plus a directory-style page, so the extension-less rewrite arm and the
// leave-alone arm are both exercised against realistic stored values.
func testIndex() PageURLIndex {
	return NewPageURLIndex([]string{
		"/index.html",
		"/contact.html",
		"/services.html",
		"/tools/llm-cost-calculator.html",
		"/guides/index.html", // stored directory form; normalises to /guides
	})
}

func TestRepairPageLinks_PhantomIsUnlinkedAndTextSurvives(t *testing.T) {
	in := `<p>Read our <a href="/pricing.html">pricing guide</a> for details.</p>`
	got, repairs := RepairPageLinks(in, testIndex())

	want := `<p>Read our pricing guide for details.</p>`
	if got != want {
		t.Errorf("markup:\n got %q\nwant %q", got, want)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairUnlink || repairs[0].Href != "/pricing.html" {
		t.Errorf("repairs = %+v", repairs)
	}
	if strings.Contains(got, "href") {
		t.Error("a dead href survived the repair")
	}
}

func TestRepairPageLinks_UnlinkKeepsInnerMarkup(t *testing.T) {
	in := `<a class="btn" href="/nope"><strong>Get</strong> started</a>`
	got, _ := RepairPageLinks(in, testIndex())

	want := `<strong>Get</strong> started`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRepairPageLinks_ExtensionlessTargetIsRewrittenNotRemoved(t *testing.T) {
	// bugs_open/049 measured 8 of these fleet-wide: the target EXISTS and
	// returns 200 at its .html form. Unlinking them would be a content loss.
	in := `<p><a href="/contact">Talk to us</a></p>`
	got, repairs := RepairPageLinks(in, testIndex())

	want := `<p><a href="/contact.html">Talk to us</a></p>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairRewrite || repairs[0].NewHref != "/contact.html" {
		t.Errorf("repairs = %+v", repairs)
	}
}

func TestRepairPageLinks_RewriteEmitsTheStoredURLNotAConstructedOne(t *testing.T) {
	// The index holds "/Services.HTML"; the rewrite must hand back that exact
	// string, not a lowercased or reassembled one. bugs_closed/029's rule.
	ix := NewPageURLIndex([]string{"/Services.HTML"})
	got, repairs := RepairPageLinks(`<a href="/services">x</a>`, ix)

	if !strings.Contains(got, `href="/Services.HTML"`) {
		t.Errorf("rewrite did not emit the stored URL: %q", got)
	}
	// Fatalf, not an unguarded index: repairs[0] on an empty slice panics the
	// whole test BINARY, and every test declared after this one then silently
	// never runs. Found by inducing a fault to check these tests could fail at
	// all — the run reported "4 passed" that had in fact never executed.
	if len(repairs) != 1 {
		t.Fatalf("want 1 repair, got %+v", repairs)
	}
	if repairs[0].NewHref != "/Services.HTML" {
		t.Errorf("NewHref = %q", repairs[0].NewHref)
	}
}

func TestRepairPageLinks_RewritePreservesFragmentAndQuery(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`<a href="/contact#form">a</a>`, `<a href="/contact.html#form">a</a>`},
		{`<a href="/contact?ref=hero">a</a>`, `<a href="/contact.html?ref=hero">a</a>`},
	} {
		got, _ := RepairPageLinks(tc.in, testIndex())
		if got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestRepairPageLinks_ValidLinksAreByteIdentical(t *testing.T) {
	in := `<a href="/contact.html">Contact</a> and ` +
		`<a  href='/services.html'  class="x" >Services</a> and ` +
		`<a href="/guides">Guides</a>` // directory form, stored as /guides/index.html

	got, repairs := RepairPageLinks(in, testIndex())
	if got != in {
		t.Errorf("clean input was perturbed:\n got %q\nwant %q", got, in)
	}
	if len(repairs) != 0 {
		t.Errorf("unexpected repairs: %+v", repairs)
	}
}

func TestRepairPageLinks_NonPageScopesUntouched(t *testing.T) {
	in := `<a href="https://example.com/x">ext</a>` +
		`<a href="//cdn.example.com/y">proto-rel</a>` +
		`<a href="mailto:a@b.com">mail</a>` +
		`<a href="tel:+441234">tel</a>` +
		`<a href="#section">frag</a>` +
		`<a href="/assets/brochure.pdf">asset</a>`

	got, repairs := RepairPageLinks(in, testIndex())
	if got != in {
		t.Errorf("a non-page scope was rewritten:\n got %q\nwant %q", got, in)
	}
	if len(repairs) != 0 {
		t.Errorf("unexpected repairs: %+v", repairs)
	}
}

func TestRepairPageLinks_EmptyHrefIsUnlinked(t *testing.T) {
	in := `<p>See <a href="" class="btn">Browse All</a> now</p>`
	got, repairs := RepairPageLinks(in, testIndex())

	want := `<p>See Browse All now</p>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairUnlink || repairs[0].Href != "" {
		t.Errorf("repairs = %+v", repairs)
	}
}

func TestRepairPageLinks_RuntimeFillShellIsExempt(t *testing.T) {
	// Its hrefs hydrate client-side; stripping them breaks the shell. Same
	// exemption every other link check applies.
	in := `<div data-runtime-fill="lobby"><a href="/not-a-page">x</a><a href="">y</a></div>`
	got, repairs := RepairPageLinks(in, testIndex())

	if got != in {
		t.Errorf("runtime-fill shell was modified: %q", got)
	}
	if len(repairs) != 0 {
		t.Errorf("unexpected repairs: %+v", repairs)
	}
}

func TestRepairPageLinks_EmptyIndexIsANoOp(t *testing.T) {
	// A failed page query must never cause every link on the page to be
	// stripped. The caller skips repair on a load failure; this is the belt.
	in := `<a href="/anything">text</a>`
	got, repairs := RepairPageLinks(in, PageURLIndex{})

	if got != in || len(repairs) != 0 {
		t.Errorf("empty index repaired something: %q %+v", got, repairs)
	}
}

func TestRepairPageLinks_MixedDocumentTouchesOnlyTheBadAnchors(t *testing.T) {
	in := `<nav><a href="/contact.html">Contact</a></nav>` +
		`<p>Our <a href="/invented">approach</a> and our ` +
		`<a href="/tools/llm-cost-calculator">calculator</a>.</p>` +
		`<a href="https://x.test">ext</a>`

	want := `<nav><a href="/contact.html">Contact</a></nav>` +
		`<p>Our approach and our ` +
		`<a href="/tools/llm-cost-calculator.html">calculator</a>.</p>` +
		`<a href="https://x.test">ext</a>`

	got, repairs := RepairPageLinks(in, testIndex())
	if got != want {
		t.Errorf("\n got %q\nwant %q", got, want)
	}
	if len(repairs) != 2 {
		t.Fatalf("want 2 repairs, got %+v", repairs)
	}
	if repairs[0].Action != LinkRepairUnlink || repairs[1].Action != LinkRepairRewrite {
		t.Errorf("repairs = %+v", repairs)
	}
}

// The induced fault the bug file asks for, at unit scale: the real
// webdesign.co.uk homepage shipped 9 phantom /tools/* links on 2026-07-25 and
// oufe.com's shipped 6. Every link on the page is dead; all the text must
// survive and no href may remain.
func TestRepairPageLinks_PageWhereEveryLinkIsPhantom(t *testing.T) {
	in := `<section>` +
		`<a href="/tools/typography">Typography scale</a>` +
		`<a href="/tools/colour">Colour contrast</a>` +
		`<a href="/tools/css">CSS layout</a>` +
		`<a href="/guides-hub">Guides</a>` +
		`</section>`

	got, repairs := RepairPageLinks(in, testIndex())

	if strings.Contains(got, "href") || strings.Contains(got, "<a") {
		t.Errorf("a link survived: %q", got)
	}
	for _, text := range []string{"Typography scale", "Colour contrast", "CSS layout", "Guides"} {
		if !strings.Contains(got, text) {
			t.Errorf("lost anchor text %q from %q", text, got)
		}
	}
	if len(repairs) != 4 {
		t.Errorf("want 4 repairs, got %d: %+v", len(repairs), repairs)
	}
}

func TestPageURLIndex_LookupAndContains(t *testing.T) {
	ix := testIndex()

	for _, tc := range []struct {
		href      string
		wantOK    bool
		wantStore string
	}{
		{"/contact.html", true, "/contact.html"},
		{"/contact.html#form", true, "/contact.html"}, // fragment dropped for matching
		{"/guides", true, "/guides/index.html"},       // directory form resolves to its stored url
		{"/guides/", true, "/guides/index.html"},
		{"/contact", false, ""}, // extension-less MISS is the point: /contact really 404s
		{"/nope", false, ""},
	} {
		stored, ok := ix.Lookup(tc.href)
		if ok != tc.wantOK || (ok && stored != tc.wantStore) {
			t.Errorf("Lookup(%q) = %q,%v; want %q,%v", tc.href, stored, ok, tc.wantStore, tc.wantOK)
		}
		if ix.Contains(tc.href) != tc.wantOK {
			t.Errorf("Contains(%q) = %v, want %v", tc.href, ix.Contains(tc.href), tc.wantOK)
		}
	}
}
