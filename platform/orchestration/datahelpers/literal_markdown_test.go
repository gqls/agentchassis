package datahelpers

import (
	"strings"
	"testing"
)

// The repair contract (bugs_open/184): after StripLiteralMarkdown, the scan
// finds nothing. If this property breaks, either the verifier strands items in
// 'failed' for ever (scan stricter than strip) or a still-served defect gets
// stamped repaired (strip stricter than scan). Run over the live corpus that
// motivated the bug plus generated composites.
func TestStripThenScanFindsNothing(t *testing.T) {
	corpus := []string{
		// the three founding rows (bugs_open/184 §Scope)
		"Banks evaluate your application using a **Decision Engine** (an automated algorithm).",
		"**Recommended next steps:**",
		"**the `animation`**",
		// the widened live symptom (2026-08-17 contribution + this session's findings)
		"# Getting a mortgage\nSome prose.",
		"## Understanding rates\n### Sub heading",
		"## [OpenAI safety update](https://openai.com/news/safety-alignment/)",
		"Research by [James Kettle](https://portswigger.net/research/james-kettle) shows...",
		"[read the guide](/guides/first-time-buyer/index.html)",
		"Set `true` to enable, or use `ease-in-out` timing with `feTurbulence`.",
		// Tier 1 leaves the `- ` markers here; only the feed tier removes them, and
		// the fixpoint property holds either way because a strip-only rule can
		// never leave something for the scan to find.
		"- first point\n- second point\n# A heading\n**bold** and `code`",
		// composites that need more than one pass
		"**[bold link](https://example.com/x)**",
		"# **Bold heading**",
		"## [Title](https://example.com) with `code` after",
		// THE LIVE CORPUS (bugs_open/332) — verbatim from boxingonline.ugg2.com's
		// served news page and /data/news-archive.json, 2026-09-03. Every one is a
		// half-pattern manufactured by our own 197-byte snippet cut.
		"Itauma (14-1, 12 KOs) ultimately punched himself out and [lost in the ninth round](https://sports.yahoo.com/boxing/live/moses-itauma-vs-filip-hrgovic-live-results-round-by-round-updates-...",
		"- Tennis (W)\n- [NLL (Lacrosse)](https://www.espn.com/boxin...",
		"A total of seven fights were announced, including [Alexander Volkanovski](https://bloodyelbow.com/tag/alexander-v...",
		"[![Results: Thorslund, Thibeault win titles in...",
		"# The housing market's fragile place\n![](https://www.axios.com/_next/image?url=https%3A%2F%2Fimages.axios.com%2F...",
	}
	for _, in := range corpus {
		cleaned, _ := StripLiteralMarkdown(in, true)
		if got := LiteralMarkdownPatterns(cleaned, true); len(got) != 0 {
			t.Errorf("scan found %v after strip\n input: %q\n output: %q", got, in, cleaned)
		}
	}
}

// The letter-guards must hold for the stripper exactly as they hold for the
// detector: text that the check would never flag must pass through unchanged.
func TestStripLeavesGuardedTextAlone(t *testing.T) {
	unchanged := []string{
		"3 * 4 = 12 and 2**10 is 1024",
		"a ** b in maths notation",
		"#fff and #1 rated and issue #12",
		"prices from £99**", // no opening letter-led pair
		"array[0](call) is indexing, not a link",
		"[1](https://example.com) numeric citations do not fire",
		"see [the docs](ftp://example.com/x) — non-http scheme does not fire",
		"plain prose with nothing to do",
	}
	for _, in := range unchanged {
		out, did := StripLiteralMarkdown(in, true)
		if did || out != in {
			t.Errorf("guarded text was changed:\n in:  %q\n out: %q", in, out)
		}
	}
}

// Strip-only: the output must never contain a character sequence that opens
// markup the input did not have, and must be no longer than the input.
func TestStripNeverInserts(t *testing.T) {
	cases := []string{
		"**x** and `y` and # h and [t](https://u.example/)",
		"## [Title](https://example.com)",
	}
	for _, in := range cases {
		out, _ := StripLiteralMarkdown(in, true)
		if len(out) > len(in) {
			t.Errorf("output longer than input: %q -> %q", in, out)
		}
		if strings.Contains(out, "<") && !strings.Contains(in, "<") {
			t.Errorf("strip introduced markup: %q -> %q", in, out)
		}
	}
}

func TestStripKeepsVisibleText(t *testing.T) {
	cases := map[string]string{
		"**Decision Engine**": "Decision Engine",
		"`animation`":         "animation",
		"# Guide to X":        "Guide to X",
		"[James Kettle](https://portswigger.net/research/james-kettle)": "James Kettle",
		"## [Title](https://example.com/a)":                             "Title",
		"**the `animation`**":                                           "the animation",
	}
	for in, want := range cases {
		got, _ := StripLiteralMarkdown(in, true)
		if got != want {
			t.Errorf("StripLiteralMarkdown(%q) = %q, want %q", in, got, want)
		}
	}
}

// Markup-bearing values keep their code spans and link-shaped text: the same
// suppression rule the detector applies (a value carrying markup is not a
// text-typed field).
func TestMarkupBearingValueSuppression(t *testing.T) {
	in := "<p>run `npm install` then see [docs](https://example.com)</p>"
	out, did := StripLiteralMarkdown(in, !HTMLMarkupRe.MatchString(in))
	if did || out != in {
		t.Errorf("markup-bearing value was changed:\n in:  %q\n out: %q", in, out)
	}
}

func TestStripFromContentDataWalks(t *testing.T) {
	cd := map[string]interface{}{
		"headline":  "**Decision Engine** explained",
		"_built_at": "2026-08-18T00:00:00Z **not touched**",
		"items": []interface{}{
			map[string]interface{}{"summary": "## [Title](https://example.com/x)"},
			map[string]interface{}{"summary": "clean already"},
		},
		"body_html": "<p>`kept` because markup-bearing</p>",
		"count":     3,
	}
	changed := StripLiteralMarkdownFromContentData(cd)
	if len(changed) != 2 {
		t.Fatalf("changed = %v, want exactly headline and items[0].summary", changed)
	}
	if cd["headline"] != "Decision Engine explained" {
		t.Errorf("headline = %q", cd["headline"])
	}
	if got := cd["items"].([]interface{})[0].(map[string]interface{})["summary"]; got != "Title" {
		t.Errorf("items[0].summary = %q", got)
	}
	if !strings.Contains(cd["_built_at"].(string), "**not touched**") {
		t.Errorf("_-prefixed key was touched: %q", cd["_built_at"])
	}
	if !strings.Contains(cd["body_html"].(string), "`kept`") {
		t.Errorf("markup-bearing value was touched: %q", cd["body_html"])
	}
}

// Strip-to-empty cannot manufacture a blank (council 060bcc0a r5,
// render_guardian): every strip pattern keeps at least one letter/digit of
// visible text — bold/code/link captures start with [A-Za-z0-9] — and the
// heading strip deletes only the `#… ` prefix. So if the strip yields "", the
// input had no letter or digit in it, i.e. it was already displayably empty.
// Bare image markdown, the case the seat named, keeps its alt text.
func TestStripToEmptyOnlyFromAlreadyEmptyInput(t *testing.T) {
	hasAlnum := func(s string) bool {
		return strings.IndexFunc(s, func(r rune) bool {
			return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		}) >= 0
	}
	inputs := []string{
		"# ", "## \n### ", "# \n", "**x**", "`1`", "[a](https://e.com)",
		"![alt](https://e.com/i.png)", "![](https://e.com/i.png)", "# **Bold**",
		"## `code`", "**`x`**", "# ![alt](https://e.com/i.png)", "---", "   ",
		"#  x", "# \n# y", "**a** ", " `b` ", "#   ",
	}
	for _, in := range inputs {
		got, _ := StripLiteralMarkdown(in, true)
		// Whitespace-only counts as empty (council 060bcc0a r6, bug_historian):
		// a heading strip leaves the rest of the line, so the only way to reach
		// "" or "   " is from an input with nothing but markers and spaces.
		if strings.TrimSpace(got) == "" && hasAlnum(in) {
			t.Errorf("StripLiteralMarkdown(%q) emptied a value that had visible text (got %q)", in, got)
		}
		if got != "" && !hasAlnum(got) && hasAlnum(in) {
			t.Errorf("StripLiteralMarkdown(%q) = %q lost every letter/digit", in, got)
		}
	}
	// CHANGED 2026-09-03 (bugs_open/332): was "!alt", now "alt".
	//
	// This assertion is council 060bcc0a r5/r6's, and the argument for moving it
	// belongs here rather than in a commit message nobody will read again.
	//
	// The seat's property is that a bare image token KEEPS ITS VISIBLE TEXT — it
	// is evidence for the blank-manufacture guard above, not a specification of
	// the output. "alt" serves that property better than "!alt": it still carries
	// letters, so both clauses of the guard pass unchanged, and the visitor stops
	// seeing a stray "!" that existed only because mdLinkStripRe ran before any
	// image rule and ate "[alt](url)" out of "![alt](url)". The "!" was a
	// leftover of strip ORDER; it was never argued for by anyone.
	//
	// So the guard is not weakened — the output is tightened while the guard
	// holds. What is retired is an incidental character, not a protection.
	if got, _ := StripLiteralMarkdown("![alt](https://e.com/i.png)", true); got != "alt" {
		t.Errorf("bare image markdown = %q, want the alt text kept (%q)", got, "alt")
	}
}

// ---------------------------------------------------------------------------
// bugs_open/332 — the truncated shapes, and the tier relationship
// ---------------------------------------------------------------------------

// THE MOST IMPORTANT TEST IN THIS FILE, because it is the only guard on the one
// failure mode neither a served-artefact grep nor any existing property can see.
//
// A truncated link's URL carries the "..." that firecrawl.go appended at its cut.
// Deleting the URL deletes the marker, and the result is a grammatical sentence
// the source never wrote — prettier than today's output and less true, on a paid
// customer's page. TestStripNeverInserts asserts only length; the scan is clean
// by construction; nothing else would notice.
func TestTruncatedLinkStripKeepsTheTruncationMarker(t *testing.T) {
	cases := map[string]string{
		"and [lost in the ninth round](https://sports.yahoo.com/boxing/live/moses-itauma-...": "and lost in the ninth round...",
		"including [Alexander Volkanovski](https://bloodyelbow.com/tag/alexander-v...":        "including Alexander Volkanovski...",
		"[NLL (Lacrosse)](https://www.espn.com/boxin...":                                      "NLL (Lacrosse)...",
		// A tail with no marker must not gain one — strip-only cuts both ways.
		"see [the guide](https://example.com/a/very/long/path/that/just/ends": "see the guide",
		// The unicode ellipsis is the same case and must survive too.
		"read [the notice](https://example.com/notice…": "read the notice…",
	}
	for in, want := range cases {
		got, _ := StripLiteralMarkdown(in, true)
		if got != want {
			t.Errorf("StripLiteralMarkdown(%q)\n  = %q\n want %q", in, got, want)
		}
		if len(got) > len(in) {
			t.Errorf("re-emitting the marker made the output longer: %q -> %q", in, got)
		}
	}
}

// The left word boundary is what pays for the closing paren the pattern gave up.
// Without it these fire, and each one deletes a path a reader needed.
func TestTruncatedLinkNeedsALeftWordBoundary(t *testing.T) {
	unchanged := []string{
		"config[Debug](/api/v2/logs",         // letter before the bracket
		"array[0](/api/v2/logs",              // digit link text, guarded twice over
		"see [the docs](ftp://example.com/x", // non-http scheme
		"a bracket [alone at the end",        // no "](", so not this pattern's business
	}
	for _, in := range unchanged {
		out, did := StripLiteralMarkdown(in, true)
		if did || out != in {
			t.Errorf("guarded text was changed by the truncated-link rule:\n in:  %q\n out: %q", in, out)
		}
	}
}

// Tier 2 is a strict superset of tier 1 and still satisfies the repair contract.
// Both directions, because "superset" alone would be satisfied by a rule that
// also broke Scan(Strip(x)) == nothing.
func TestFeedDisplayStripIsASupersetThatKeepsTheContract(t *testing.T) {
	corpus := []string{
		"- Tennis (W)\n- [NLL (Lacrosse)](https://www.espn.com/boxin...",
		"- NFL\n- [MLB](https://www.espn.com/boxing/story/_/id...",
		"[![Results: Thorslund, Thibeault win titles in...",
		"**CountryBox: Where Music Meets Bo...",
		"# Getting a mortgage\nSome prose.",
		"plain prose with nothing to do",
		"**Decision Engine** explained",
	}
	for _, in := range corpus {
		tier1, _ := StripLiteralMarkdown(in, true)
		feed, _ := StripFeedDisplayMarkdown(in, true)
		if len(feed) > len(tier1) {
			t.Errorf("feed strip is not a subset of tier 1's output:\n in:    %q\n tier1: %q\n feed:  %q", in, tier1, feed)
		}
		if got := LiteralMarkdownPatterns(feed, true); len(got) != 0 {
			t.Errorf("scan found %v after the FEED strip\n in:   %q\n feed: %q", got, in, feed)
		}
	}
}

// The feed tier removes more, so it needs the blank-manufacture guard too — the
// council's objection was about the strip's output reaching a visitor, and this
// output reaches three surfaces.
func TestFeedDisplayStripNeverManufacturesABlank(t *testing.T) {
	hasAlnum := func(s string) bool {
		return strings.IndexFunc(s, func(r rune) bool {
			return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		}) >= 0
	}
	inputs := []string{
		"- ", "* ", "- x", "[Beta", "![A", "[![Results: a title in...",
		"**CountryBox: Where Music Meets Bo...", "- Tennis (W)",
		"![](https://e.com/i.png)", "---", "   ", "- [a](https://e.com/x",
	}
	for _, in := range inputs {
		got, _ := StripFeedDisplayMarkdown(in, true)
		if strings.TrimSpace(got) == "" && hasAlnum(in) {
			t.Errorf("StripFeedDisplayMarkdown(%q) emptied a value that had visible text (got %q)", in, got)
		}
		if got != "" && !hasAlnum(got) && hasAlnum(in) {
			t.Errorf("StripFeedDisplayMarkdown(%q) = %q lost every letter/digit", in, got)
		}
	}
}

// The four guards on the bold tail, and the one residual, stated rather than
// discovered later. `**kwargs` / `**args` are live shapes on this estate's own
// AI-facing sites, which is why the rule is feed-only in the first place.
// CORRECTED 2026-09-03: my first fixture here was "in Python use **kwargs and
// **args here", which DOES get stripped — by the PRE-EXISTING complete-bold rule
// (the text between the two pairs is "kwargs and ", a valid **…** match), not by
// the tail rule at all. Isolating the two showed the tail rule never fired on it.
// A guard test that fails for a reason other than the rule it guards is worse
// than no test: it would have been "fixed" by weakening the wrong pattern.
//
// RESIDUAL, stated rather than left to be discovered: a single unclosed pair
// followed by a long phrase — "pass **kwargs to the function" — DOES fire. That
// is the honest cost of the rule. It does not occur in the live corpus: run over
// all 5,112 non-empty feed rows [MEASURED 2026-09-03], the tail rule changes
// exactly 7 rows and every one is a genuinely truncated bold opening
// ("**CountryBox: Where Music Meets Bo...", "**Eric and Wendy Schmidt AI in
// Science African Faculty Fellow..."). Confining the rule to feed snippets is
// what keeps that residual away from prose we author.
func TestFeedBoldTailGuards(t *testing.T) {
	unchanged := []string{
		"in Python use **args here",    // phrase guard: only 4 chars after the space
		"prices from £99**",            // nothing after the pair
		"O(n**k",                       // letter before the pair
		"Free delivery**Terms apply",   // letter before the pair
		"3 * 4 = 12 and 2**10 is 1024", // digit guard
	}
	for _, in := range unchanged {
		out, _ := StripFeedDisplayMarkdown(in, true)
		if out != in {
			t.Errorf("bold-tail guard failed:\n in:  %q\n out: %q", in, out)
		}
	}
	if got, _ := StripFeedDisplayMarkdown("Tuesday, 7:00 pm ET **CountryBox: Where Music Meets Bo...", true); got != "Tuesday, 7:00 pm ET CountryBox: Where Music Meets Bo..." {
		t.Errorf("the live case did not strip: %q", got)
	}
}

// The feed tier must leave OUR content_data alone if it is ever pointed at it by
// mistake — this is the blast-radius statement, executable.
func TestFeedDisplayStripLeavesAuthoredProseAlone(t *testing.T) {
	unchanged := []string{
		"We answer in one working day.",
		"Rates from 4.5% APRC. Your home may be repossessed.",
		"<p>run `npm install` then see [docs](https://example.com)</p>",
	}
	for _, in := range unchanged {
		out, _ := StripFeedDisplayMarkdown(in, !HTMLMarkupRe.MatchString(in))
		if out != in {
			t.Errorf("authored prose was changed:\n in:  %q\n out: %q", in, out)
		}
	}
}
