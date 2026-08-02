package datahelpers

import (
	"regexp"
	"strings"
	"testing"
)

// THE VACUOUS-NEGATIVE GUARD, and it is why every case below is a PAIR.
//
// "the script was left alone" is asserted as byte-identical output — and
// RepairPageLinks returns its input byte-identically down four other paths too
// (empty html, empty index, a runtime-fill marker, no anchor matched at all). A
// test that only pinned the script would still pass if the function had stopped
// working entirely, which is the shape a council seat caught in this lane's
// previous round. So each input also carries a REAL phantom anchor OUTSIDE the
// non-markup region, and each test asserts that one WAS repaired. The pair fails
// in both directions: over-repair breaks the first assertion, and a function
// that has quietly become a no-op breaks the second.

func nonMarkupIndex() PageURLIndex {
	return NewPageURLIndex([]string{"/index.html", "/tools/checker.html", "/about.html"})
}

const livePhantom = `<p><a href="/gone">Pricing</a></p>`
const livePhantomRepaired = `<p>Pricing</p>`

func assertRepairedOnlyOutside(t *testing.T, name, protected string) {
	t.Helper()
	in := protected + livePhantom
	got, repairs := RepairPageLinks(in, nonMarkupIndex())

	if want := protected + livePhantomRepaired; got != want {
		t.Errorf("%s: protected region was rewritten\n in : %s\n got: %s\nwant: %s", name, in, got, want)
	}
	// The control: exactly one repair, and it is the anchor in the prose.
	if len(repairs) != 1 || repairs[0].Action != LinkRepairUnlink || repairs[0].Href != "/gone" {
		t.Errorf("%s: the live phantom outside the region was not repaired — the negative above proves nothing; repairs=%+v", name, repairs)
	}
}

func TestRepairPageLinksLeavesAnchorsThatAreNotAnchors(t *testing.T) {
	cases := []struct{ name, protected string }{
		{
			// bugs_open/180 as filed: a tool builds its anchor at runtime. The
			// href capture cannot cross the ' that follows href=", so the shipped
			// regex read this as href="" and the unlink arm deleted a WORKING link.
			"js string concatenation",
			`<script>var h = '<p>' + t + ' <a href="' + q.link + '" target="_blank">See guide</a>.</p>';</script>`,
		},
		{
			// The other half of the class, and it takes a DIFFERENT arm: ${q.link}
			// is a non-empty href, so it is judged a phantom rather than an empty
			// control. One cause, two symptoms — which is why the fix is not a
			// denylist of `' +`.
			"js template literal",
			"<script>el.innerHTML = `<a href=\"${q.link}\">See guide</a>`;</script>",
		},
		{"style comment", `<style>/* was <a href="/gone">x</a> */ .a{color:red}</style>`},
		{"textarea shows markup as text", `<textarea><a href="/gone">paste me</a></textarea>`},
		{"title", `<title>How to write <a href="/gone">a link</a></title>`},
		{"html comment", `<!-- <a href="/gone">removed for now</a> -->`},
		{"uppercase script tag", `<SCRIPT>var s = '<a href="/gone">x</a>';</SCRIPT>`},
		{"script with attributes", `<script type="module" defer>var s = '<a href="/gone">x</a>';</script>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { assertRepairedOnlyOutside(t, c.name, c.protected) })
	}
}

// THE CASE THAT DECIDES MASK-VERSUS-FILTER. The regex's non-greedy </a> closes
// at the REAL anchor, so before the fix ONE match ran from inside the script to
// the end of a genuine phantom. Merely dropping matches that start in a span
// would drop the real anchor along with the decoy — and FindAll never revisits
// those bytes, so the phantom would ship. Masking removes the decoy and leaves
// the real anchor to be matched on its own.
func TestDecoyAnchorInsideScriptDoesNotSwallowTheRealOneAfterIt(t *testing.T) {
	in := `<script>var t = '<a href="/gone">';</script><a href="/gone">Pricing</a>`
	want := `<script>var t = '<a href="/gone">';</script>Pricing`
	got, repairs := RepairPageLinks(in, nonMarkupIndex())
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if len(repairs) != 1 {
		t.Fatalf("want the real phantom repaired exactly once, got %+v", repairs)
	}
}

// A comment INSIDE an anchor's content is not a reason to leave the anchor
// alone, and the content must survive the unlink verbatim — the offsets index
// the original bytes, never the mask.
func TestUnlinkKeepsMaskedBytesInInnerContentVerbatim(t *testing.T) {
	in := `<p><a href="/gone">Read <!-- keep me --> more</a></p>`
	want := `<p>Read <!-- keep me --> more</p>`
	if got, _ := RepairPageLinks(in, nonMarkupIndex()); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestRewriteArmStillFiresOutsideNonMarkup(t *testing.T) {
	in := `<script>var u = "/about";</script><a href="/about">About</a>`
	want := `<script>var u = "/about";</script><a href="/about.html">About</a>`
	got, repairs := RepairPageLinks(in, nonMarkupIndex())
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if len(repairs) != 1 || repairs[0].Action != LinkRepairRewrite {
		t.Errorf("want one rewrite, got %+v", repairs)
	}
}

func TestNonMarkupSpansCoverWholeElementsAndComments(t *testing.T) {
	html := `<p>a</p><script>x</script><b>c</b><!-- d --><style>e</style>`
	spans := NonMarkupSpans(html)
	inside := []string{"<script>x</script>", "<!-- d -->", "<style>e</style>"}
	for _, frag := range inside {
		at := strings.Index(html, frag)
		if !spans.Contains(at) {
			t.Errorf("%q at %d should be non-markup", frag, at)
		}
		if !spans.Contains(at + len(frag) - 1) {
			t.Errorf("%q end at %d should be non-markup", frag, at+len(frag)-1)
		}
	}
	for _, frag := range []string{"<p>a</p>", "<b>c</b>"} {
		at := strings.Index(html, frag)
		if spans.Contains(at) {
			t.Errorf("%q at %d is real markup and must not be masked", frag, at)
		}
	}
}

// The degrade direction for a WRITER is wide: markup the scanner cannot finish
// reading must not be rewritten. Each of these must protect the phantom that
// follows the damage, which is the opposite of what RuntimeFillSpans does with
// its own uncertainty and is argued at scanSpans.
func TestUnparseableMarkupIsTreatedAsNonMarkupToTheEnd(t *testing.T) {
	cases := []struct{ name, html string }{
		{"unclosed script", `<script>var t = "` + livePhantom},
		{"unterminated comment", `<!-- oops ` + livePhantom},
		{"unterminated quote in a start tag", `<div class="oops ` + livePhantom},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, repairs := RepairPageLinks(c.html, nonMarkupIndex())
			if got != c.html || len(repairs) != 0 {
				t.Errorf("a writer rewrote markup it could not parse\n in : %s\n got: %s\nrepairs=%+v", c.html, got, repairs)
			}
		})
	}
}

// THE FILLER IS LOAD-BEARING, and the input below is the one that shows it.
//
// The first input I wrote for this — `<style>.a{}</style>src=""` — passed with a
// SPACE filler too, and not because a space is safe: the manufactured match
// began ON the mask, so dropMatchesInSpans discarded it one level down. Two
// guards in SERIES, and the mutation that should have exposed the weaker one was
// absorbed by the stronger. The discriminating case is a match that begins
// OUTSIDE the span and only exists because the span turned to whitespace: here
// `\s*` cannot cross `<style>…`, but it crosses sixteen spaces happily, and the
// match starts at the space before `src`, where the offset filter has no view.
// With a space filler this deletes the <style> element and the src attribute
// together. Mutate maskFiller to ' ' and this test fails.
func TestMaskFillerCannotManufactureAMatchStartingOutsideASpan(t *testing.T) {
	srcAttrRe := regexp.MustCompile(`(?i)\ssrc\s*=\s*(?:""|'')`)
	html := `<p>a</p> src=<style>.a{}</style>""`
	if got := srcAttrRe.FindAllStringIndex(html, -1); len(got) != 0 {
		t.Fatalf("premise broken: the ORIGINAL document already matches at %v", got)
	}
	if spans := NonMarkupSpans(html); len(spans) != 1 {
		t.Fatalf("premise broken: the <style> element was not recognised as a span (%+v), so nothing is masked and this test cannot discriminate", spans)
	}
	if out := ReplaceAllInMarkup(srcAttrRe, html, ""); out != html {
		t.Errorf("the mask invented a match the document did not contain\n in : %q\n out: %q", html, out)
	}
}

func TestReplaceAllInMarkupIsByteIdenticalWhenNothingMatchesOutside(t *testing.T) {
	deadAnchor := regexp.MustCompile(`(?is)<a\b[^>]*\shref\s*=\s*(?:""|'')[^>]*>.*?</a>`)
	protected := `<script>var s = '<a href="">dead</a>';</script>`
	live := `<nav><a href="">Home</a></nav>`

	if out := ReplaceAllInMarkup(deadAnchor, protected, ""); out != protected {
		t.Errorf("dropped a control that only exists inside a script string\n in : %s\n out: %s", protected, out)
	}
	// Control: the same pattern still deletes a real dead control.
	if out := ReplaceAllInMarkup(deadAnchor, protected+live, ""); out != protected+`<nav></nav>` {
		t.Errorf("the real dead control was not dropped — the negative above proves nothing; out=%s", out)
	}
}

// RuntimeFillSpans and NonMarkupSpans now come off ONE walk. This pins that
// sharing the walk did not make the marker visible inside a script string —
// the property runtime_fill.go's rawTextElements comment claims.
func TestSharedWalkKeepsTheMarkerInvisibleInsideScripts(t *testing.T) {
	html := `<script>var a = '<div data-runtime-fill>x</div>';</script><section>y</section>`
	if spans := RuntimeFillSpans(html); len(spans) != 0 {
		t.Errorf("a marker quoted inside a script must not open a shell: %+v", spans)
	}
	if spans := NonMarkupSpans(html); len(spans) != 1 {
		t.Errorf("want exactly the script span, got %+v", spans)
	}
}
