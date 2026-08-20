// FILE: platform/orchestration/actions/head_assembly_test.go
//
// Tests for the per-page head identity splice and the document-language read
// (bugs_open/252, og/lang slug).
//
// Measured 2026-08-19, and what the fixtures below reproduce: 24 stored head
// rows fleet-wide, 22 carrying `og:url` pointing at the site HOMEPAGE, and 4
// (the `head-seo-standard` sites: finetuning.uk, leopardessconsulting.co.uk,
// ai-agent-orchestration.com, gaswholesalers.com) carrying BOTH a blank
// template placeholder AND a filled duplicate appended by injectBrandHeadTags,
// whose idempotency guard tests `rel="icon"` or `og:image` and so cannot see
// the blank. 700 assembled pages across 26 sites served the result.
//
// Every test here was proven load-bearing by running the named mutation and
// watching it fail; the mutations are recorded in
// docs024_key_docs_latest/bugfix_252_og_lang_assembly/NOTES_og_lang_assembly.md.
package actions

import (
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// testStoredHeadBrandDup is the REAL live shape of a head-seo-standard site's
// stored head: the template's own blank og pair, then injectBrandHeadTags'
// filled block with the homepage og:url, then the blank description the
// description splice targets. Copied structurally from
// https://ai-agent-orchestration.com/about.html as served 2026-08-19.
const testStoredHeadBrandDup = `<head>
    <meta charset="UTF-8">
    <title>Loan rules, plainly</title>
    <meta property="og:title" content="">
    <meta property="og:description" content="">
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="Example Co">
  <link rel="icon" href="/assets/images/favicon.png">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="Example Co">
  <meta property="og:title" content="Example Co">
  <meta property="og:description" content="A site-level tagline.">
  <meta property="og:image" content="https://lendzy.co.uk/assets/images/og-card.png">
  <meta property="og:url" content="https://lendzy.co.uk/">
  <meta name="twitter:card" content="summary_large_image">
    <meta name="description" content="">
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`

func countTag(s, needle string) int {
	return strings.Count(s, needle)
}

// ogTestPage is canonicalTestPage() plus a meta description. The shared fixture
// deliberately carries none — most of the fleet has none (bugs_open/320) and the
// canonical tests need that shape — so tests asserting og:description supply
// their own rather than mutating a fixture four other tests read.
func ogTestPage() *PageInfo {
	p := canonicalTestPage()
	p.MetaDesc = "Plain answers on UK loan rules."
	return p
}

func TestSpliceOpenGraphInjectsPerPageSet(t *testing.T) {
	// The 18-site "Document Head" case: a head with no og tags at all.
	page := ogTestPage()
	out := spliceOpenGraph(testStoredHead, page, zap.NewNop())

	for _, want := range []string{
		`<meta property="og:title" content="` + page.Title + `">`,
		`<meta property="og:description" content="` + page.MetaDesc + `">`,
		`<meta property="og:url" content="https://` + page.Domain + page.URL + `">`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in:\n%s", want, out)
		}
	}
	// Injected INSIDE the head, not after it.
	if strings.Index(out, "og:url") > strings.Index(out, "</head>") {
		t.Fatalf("og tags landed outside </head>:\n%s", out)
	}
	// Neighbours untouched.
	for _, keep := range []string{`charset="UTF-8"`, "styles.css", "<title>Loan rules, plainly</title>"} {
		if !strings.Contains(out, keep) {
			t.Fatalf("damaged a neighbour (%s):\n%s", keep, out)
		}
	}
}

func TestSpliceOpenGraphSelfHealsDuplicatedBrandTags(t *testing.T) {
	// The load-bearing case: the live 4-site shape must come out with ONE of
	// each page-scoped tag, carrying the PAGE's values, and every site-scoped
	// tag preserved.
	page := ogTestPage()
	out := spliceOpenGraph(testStoredHeadBrandDup, page, zap.NewNop())

	if n := countTag(out, "og:title"); n != 1 {
		t.Fatalf("expected exactly 1 og:title, got %d:\n%s", n, out)
	}
	if n := countTag(out, "og:url"); n != 1 {
		t.Fatalf("expected exactly 1 og:url, got %d:\n%s", n, out)
	}
	if n := countTag(out, "og:description"); n != 1 {
		t.Fatalf("expected exactly 1 og:description, got %d:\n%s", n, out)
	}
	// The surviving values are the PAGE's, not the site's.
	if !strings.Contains(out, `<meta property="og:title" content="`+page.Title+`">`) {
		t.Fatalf("og:title is not the page title:\n%s", out)
	}
	if strings.Contains(out, `content="https://lendzy.co.uk/">`) {
		t.Fatalf("the baked homepage og:url survived:\n%s", out)
	}
	// No blank page-scoped tag anywhere.
	for _, blank := range []string{`property="og:title" content=""`, `property="og:description" content=""`} {
		if strings.Contains(out, blank) {
			t.Fatalf("a blank placeholder survived (%s):\n%s", blank, out)
		}
	}
	// Site-scoped tags are none of our business and must be byte-preserved.
	for _, keep := range []string{
		`<meta property="og:type" content="website">`,
		`<meta property="og:site_name" content="Example Co">`,
		`<meta property="og:image" content="https://lendzy.co.uk/assets/images/og-card.png">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<link rel="icon" href="/assets/images/favicon.png">`,
	} {
		if !strings.Contains(out, keep) {
			t.Fatalf("stripped a site-scoped tag we do not own (%s):\n%s", keep, out)
		}
	}
	// og:type appears twice in the fixture; we must not have deduped it.
	if n := countTag(out, "og:type"); n != 2 {
		t.Fatalf("og:type count changed (%d) — the strip is too wide:\n%s", n, out)
	}
	// Removal must not leave indented blank lines behind.
	if strings.Contains(out, "\n\n") {
		t.Fatalf("removal left an empty line behind:\n%s", out)
	}
}

func TestSpliceOpenGraphAgreesWithCanonicalAndJSONLD(t *testing.T) {
	// The three page-identity assertions must name ONE url. This is asserted,
	// not commented, because the previous version of this contract was a
	// comment claiming two literals were "byte-identical".
	reOgURL := regexp.MustCompile(`property="og:url" content="([^"]+)"`)
	reCanon := regexp.MustCompile(`rel="canonical" href="([^"]+)"`)
	reID := regexp.MustCompile(`"@id"\s*:\s*"([^"]+)"`)

	for _, url := range []string{"/about.html", "/index.html"} {
		page := canonicalTestPage()
		page.URL = url

		og := reOgURL.FindStringSubmatch(spliceOpenGraph(testStoredHead, page, zap.NewNop()))
		canon := reCanon.FindStringSubmatch(injectCanonicalLink(testStoredHead, page, zap.NewNop()))
		ld := reID.FindStringSubmatch(injectPageJSONLD(testStoredHead, page, zap.NewNop()))
		if og == nil || canon == nil || ld == nil {
			t.Fatalf("%s: one of the three emitted nothing (og=%v canon=%v ld=%v)", url, og, canon, ld)
		}
		if og[1] != canon[1] || og[1] != ld[1] {
			t.Fatalf("%s: identities disagree — og:url=%q canonical=%q @id=%q", url, og[1], canon[1], ld[1])
		}
	}

	// And the root case must be the normalised bare form, not /index.html —
	// otherwise the three could agree on the wrong url.
	page := canonicalTestPage()
	page.URL = "/index.html"
	out := spliceOpenGraph(testStoredHead, page, zap.NewNop())
	if !strings.Contains(out, `content="https://`+page.Domain+`/">`) {
		t.Fatalf("root og:url was not normalised to the bare /:\n%s", out)
	}
}

func TestSpliceOpenGraphOmitsDescriptionItCannotFill(t *testing.T) {
	// Correct-or-absent: no meta description means NO og:description, and any
	// blank one already in the head must go. 55.7% of active pages were in
	// this state on 2026-08-19 (bugs_open/320).
	page := ogTestPage()
	page.MetaDesc = ""
	out := spliceOpenGraph(testStoredHeadBrandDup, page, zap.NewNop())

	if countTag(out, "og:description") != 0 {
		t.Fatalf("expected no og:description at all:\n%s", out)
	}
	// The page still says who it is and where it lives.
	if !strings.Contains(out, "og:title") || !strings.Contains(out, "og:url") {
		t.Fatalf("an empty description suppressed the other two tags:\n%s", out)
	}
}

func TestSpliceOpenGraphIsIdempotent(t *testing.T) {
	page := ogTestPage()
	once := spliceOpenGraph(testStoredHeadBrandDup, page, zap.NewNop())
	twice := spliceOpenGraph(once, page, zap.NewNop())
	if once != twice {
		t.Fatalf("second pass changed the head:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}

func TestSpliceOpenGraphSkipsAnOgURLItCannotAssertTruthfully(t *testing.T) {
	// Same eligibility as injectCanonicalLink, and for the same reason: a
	// wrong url is acted on, an absent one is not. og:title must survive each
	// skip — the url being unusable says nothing about the title.
	cases := map[string]func(*PageInfo){
		"no domain":          func(p *PageInfo) { p.Domain = "" },
		"not root-relative":  func(p *PageInfo) { p.URL = "about.html" },
		"carries a fragment": func(p *PageInfo) { p.URL = "/tools/calc.html#results" },
		"carries a query":    func(p *PageInfo) { p.URL = "/search.html?q=loan" },
	}
	for name, mutate := range cases {
		page := ogTestPage()
		mutate(page)
		out := spliceOpenGraph(testStoredHeadBrandDup, page, zap.NewNop())
		if strings.Contains(out, "og:url") {
			t.Fatalf("%s: emitted an og:url it cannot assert:\n%s", name, out)
		}
		// The stale baked value must STILL be gone — a skip is not a bail-out.
		if strings.Contains(out, "https://lendzy.co.uk/\"") {
			t.Fatalf("%s: skipped the injection AND left the stale homepage url:\n%s", name, out)
		}
		if !strings.Contains(out, "og:title") {
			t.Fatalf("%s: an ineligible url suppressed og:title:\n%s", name, out)
		}
	}
}

func TestSpliceOpenGraphEscapesAttributeBreakers(t *testing.T) {
	page := ogTestPage()
	page.Title = `Loans: "cap" rules & fees`
	page.MetaDesc = `Fees & "charges" explained`
	out := spliceOpenGraph(testStoredHead, page, zap.NewNop())

	if strings.Contains(out, `content="Loans: "cap"`) {
		t.Fatalf("raw quote broke out of the attribute:\n%s", out)
	}
	for _, want := range []string{`Loans: &quot;cap&quot; rules &amp; fees`, `Fees &amp; &quot;charges&quot; explained`} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected escaped %q in:\n%s", want, out)
		}
	}
}

// testStoredHeadNoBlankDesc is the DISCRIMINATING shape for the interaction
// between the two splices: blank og placeholders, and NO exact
// `<meta name="description" content="">`, which forces spliceMetaDescription
// onto its legacy "fill the first content="" anywhere" fallback. On the 24 real
// stored heads today every one carries the exact blank description tag, so the
// targeted path is taken and the order cannot matter — which is precisely why
// the order needs a test built on the shape where it CAN.
const testStoredHeadNoBlankDesc = `<head>
    <title>Loan rules, plainly</title>
    <meta property="og:title" content="">
    <meta property="og:description" content="">
    <meta property="og:image" content="">
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`

func TestSpliceOpenGraphRunsAfterTheDescriptionSpliceForAReason(t *testing.T) {
	// Pins assemblePage's call ORDER, and fails if the two lines are swapped.
	//
	// The hazard is displacement, not collision. spliceMetaDescription's legacy
	// fallback consumes the FIRST blank content attribute in the head. If the
	// og splice runs first it clears the blanks it owns, promoting og:image to
	// first — and og:image is deliberately outside its strip set, so the page
	// description lands there and STAYS. Running the og splice last means the
	// fallback can only take a blank og tag the splice is about to rewrite.
	page := ogTestPage()

	// The shipped order.
	head := spliceMetaDescription(testStoredHeadNoBlankDesc, page.MetaDesc)
	head = spliceOpenGraph(head, page, zap.NewNop())

	// The description must NOT have leaked into og:image.
	if strings.Contains(head, `property="og:image" content="`+page.MetaDesc) {
		t.Fatalf("the page description leaked into og:image — the splices are in the wrong order:\n%s", head)
	}
	// og:image is not ours to touch: it must be exactly as we found it.
	if !strings.Contains(head, `<meta property="og:image" content="">`) {
		t.Fatalf("og:image was modified; it is outside this fix's scope:\n%s", head)
	}
	// And the page's own identity is still stated, once each.
	if !strings.Contains(head, `<meta property="og:title" content="`+page.Title+`">`) {
		t.Fatalf("og:title missing or wrong:\n%s", head)
	}
	if !strings.Contains(head, `<meta property="og:description" content="`+page.MetaDesc+`">`) {
		t.Fatalf("og:description missing or wrong:\n%s", head)
	}
	for _, tag := range []string{"og:title", "og:description", "og:url"} {
		if n := countTag(head, tag); n != 1 {
			t.Fatalf("expected exactly 1 %s, got %d:\n%s", tag, n, head)
		}
	}
}

func TestSpliceOpenGraphAndDescriptionSpliceCoexistOnTheLiveHeadShape(t *testing.T) {
	// The 24-real-heads case: the exact blank description tag IS present, so
	// spliceMetaDescription takes its targeted path and the two never compete.
	// Asserted so a future change to either function cannot quietly start them
	// competing on the shape the whole fleet actually has.
	page := ogTestPage()
	head := spliceMetaDescription(testStoredHeadBrandDup, page.MetaDesc)
	head = spliceOpenGraph(head, page, zap.NewNop())

	if !strings.Contains(head, `<meta name="description" content="`+page.MetaDesc+`">`) {
		t.Fatalf("the description did not land in the description tag:\n%s", head)
	}
	if !strings.Contains(head, `<meta property="og:description" content="`+page.MetaDesc+`">`) {
		t.Fatalf("og:description lost its value:\n%s", head)
	}
	if countTag(head, "og:description") != 1 {
		t.Fatalf("ended with more than one og:description:\n%s", head)
	}
	if strings.Contains(head, `content=""`) {
		t.Fatalf("a blank content attribute survived both splices:\n%s", head)
	}
}

// testStoredHeadFragment is webdesign.co.uk's real stored head: a bare
// fragment with NO <head> open tag and NO </head> close tag (confirmed
// 2026-08-19 in site_components and on the served page, which begins
// `<!DOCTYPE html><html lang="en"><meta charset="utf-8">`). It matters out of
// proportion to its one row: that site has 117 assembled pages, the most in
// the fleet. It also carries rel="icon", which makes injectBrandHeadTags skip
// its whole block (bugs_open/322 item 4), so it has no og tags at all today.
const testStoredHeadFragment = `<meta charset="utf-8">
<title>Web design</title>
<link rel="icon" href="/assets/images/favicon.png">
<link rel="stylesheet" href="/assets/css/styles.css">`

func TestSpliceOpenGraphHandlesAHeadWithNoCloseTag(t *testing.T) {
	page := ogTestPage()
	out := spliceOpenGraph(testStoredHeadFragment, page, zap.NewNop())

	// The tags must still be emitted — this site gains its first per-page
	// Open Graph identity from this change, on 117 pages.
	for _, tag := range []string{"og:title", "og:description", "og:url"} {
		if countTag(out, tag) != 1 {
			t.Fatalf("expected exactly 1 %s on a fragment head:\n%s", tag, out)
		}
	}
	// Appended after the existing content, so they land inside the head the
	// browser implies rather than after the body starts.
	if !strings.HasPrefix(out, testStoredHeadFragment) {
		t.Fatalf("fragment was modified rather than appended to:\n%s", out)
	}
	// Idempotent on this shape too.
	if twice := spliceOpenGraph(out, page, zap.NewNop()); twice != out {
		t.Fatalf("not idempotent on a fragment head:\n%s", twice)
	}
	// And it declares no language, so the document keeps today's default.
	if got := headLangAttr(testStoredHeadFragment); got != "" {
		t.Fatalf("fragment head should declare no lang, got %q", got)
	}
}

func TestInjectBrandHeadTagsEmitsNoOgURL(t *testing.T) {
	// The emitter half of bugs_open/252. injectBrandHeadTags writes into the
	// PER-SITE stored head, so it must not assert a page URL — it has no page
	// to ask, and its origin-rooted og:url is what made 700 assembled pages
	// claim the homepage. The site-scoped tags it DOES own must survive.
	ctx := &RenderContext{
		Domain:      "lendzy.co.uk",
		CompanyName: "Lendzy",
		Tagline:     "Know the rules before you borrow",
	}
	out := injectBrandHeadTags("<head><title>t</title></head>", ctx, false, zap.NewNop())

	if strings.Contains(out, "og:url") {
		t.Fatalf("the per-site brand block still asserts a page URL:\n%s", out)
	}
	for _, keep := range []string{"og:type", "og:site_name", "og:title", "og:description", "og:image", "twitter:card"} {
		if !strings.Contains(out, keep) {
			t.Fatalf("removing og:url took %s with it:\n%s", keep, out)
		}
	}
}

func TestHeadLangAttrReadsTheHeadAndNotTheHeader(t *testing.T) {
	cases := []struct {
		name string
		head string
		want string
	}{
		{"declared", `<head lang="en-GB"><title>x</title></head>`, "en-GB"},
		{"single quotes", `<head lang='cy'><title>x</title></head>`, "cy"},
		{"attribute not first", `<head data-x="1" lang="en-GB">`, "en-GB"},
		{"none declared", testStoredHead, ""},
		// <header lang> is a body element and says nothing about the document.
		{"header is not head", `<head><title>x</title></head><header lang="fr">nav</header>`, ""},
		// A lang inside some other element must not be harvested.
		{"inner element", `<head><meta property="og:locale" content="en_GB"><p lang="de">x</p></head>`, ""},
	}
	for _, c := range cases {
		if got := headLangAttr(c.head); got != c.want {
			t.Fatalf("%s: headLangAttr = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestHTMLDocumentOpenDefaultsToTodaysBytes(t *testing.T) {
	// The behaviour-unchanged pin: a site declaring no language must produce
	// EXACTLY the line the hardcoded write produced before this change, or
	// every non-opted page in the fleet churns on the next rerender.
	const before = "<!DOCTYPE html>\n<html lang=\"en\">\n"
	if got := htmlDocumentOpen(""); got != before {
		t.Fatalf("default output changed:\ngot  %q\nwant %q", got, before)
	}
	if got := htmlDocumentOpen("en-GB"); got != "<!DOCTYPE html>\n<html lang=\"en-GB\">\n" {
		t.Fatalf("declared language not emitted: %q", got)
	}
	// End to end through the real read, on the shape the migration creates.
	if got := htmlDocumentOpen(headLangAttr(`<head lang="en-GB"><title>x</title></head>`)); !strings.Contains(got, `lang="en-GB"`) {
		t.Fatalf("head-declared language did not reach the html tag: %q", got)
	}
}
