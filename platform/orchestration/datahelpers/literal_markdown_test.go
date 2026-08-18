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
		"- first point\n- second point\n# A heading\n**bold** and `code`",
		// composites that need more than one pass
		"**[bold link](https://example.com/x)**",
		"# **Bold heading**",
		"## [Title](https://example.com) with `code` after",
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
