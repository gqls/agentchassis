// Tests for ConvertLiteralCodeSpansInHTML (bugs_open/277 §5, owner ruling
// 2026-08-20). The contract has four legs, each held by a property here:
//
//  1. REPAIR CONTRACT: after the transform, the discovery scan
//     (ExtractAssertionText + LiteralMarkdownPatterns) finds no code_span the
//     transform could reach — on the real production fixture, not only on
//     composed cases (a fixture you compose exercises its own rule).
//  2. CONVERSION ⊆ DETECTION: everything codeSpanConvertRe matches,
//     MDCodeSpanRe matches. Never the reverse requirement.
//  3. SPLICE FIDELITY: converted==0 ⇒ output is byte-identical, and a convert
//     changes ONLY the matched spans (asserted via reconstruction).
//  4. SKIP-SET PARITY: text inside every nonAssertionElements subtree is
//     untouched — including the tokenizer-buffer-mutation trap (TagName()
//     lower-cases Raw() in place, so mixed-case tags are the discriminating
//     input; see the ORDER IS LOAD-BEARING comment at the call site).

package datahelpers

import (
	"os"
	"strings"
	"testing"
)

func mustConvert(t *testing.T, in string) (string, int) {
	t.Helper()
	out, n, err := ConvertLiteralCodeSpansInHTML(in)
	if err != nil {
		t.Fatalf("ConvertLiteralCodeSpansInHTML error: %v", err)
	}
	return out, n
}

// scanCodeSpans runs the detector's own rendered_html scan (the verifier's
// predicate) and returns the code_span matches.
func scanCodeSpans(html string) []string {
	var matches []string
	for _, block := range ExtractAssertionText(html) {
		for _, pm := range LiteralMarkdownPatterns(block, true) {
			if pm[0] == "code_span" {
				matches = append(matches, pm[1])
			}
		}
	}
	return matches
}

func TestConvertCodeSpans_RealFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/ported_page_tool_cubic_bezier.html")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	in := string(raw)

	// Premise checks — if the fixture changes shape, fail HERE, not in the
	// assertions below: the prose span the detector flagged live
	// (`ease-in-out`, bugs_open/277 §5.1) and the JS template literal inside
	// <script> that an unguarded regex would corrupt.
	if !strings.Contains(in, "(`ease-in-out`)") {
		t.Fatal("PREMISE CHANGED: fixture lost the live prose code span")
	}
	if !strings.Contains(in, "`left 1s ${css}`") {
		t.Fatal("PREMISE CHANGED: fixture lost the in-script template literal")
	}
	if got := scanCodeSpans(in); len(got) != 1 || got[0] != "`ease-in-out`" {
		t.Fatalf("PREMISE CHANGED: detector sees %v, expected exactly [`ease-in-out`]", got)
	}

	out, n := mustConvert(t, in)

	if n != 1 {
		t.Fatalf("converted %d spans, expected exactly 1 (prose only, never the script literal)", n)
	}
	if !strings.Contains(out, "(<code>ease-in-out</code>)") {
		t.Error("prose span not converted to <code>")
	}
	if !strings.Contains(out, "`left 1s ${css}`") {
		t.Error("JS template literal inside <script> was modified — the skip set failed")
	}
	// Repair contract: the detector's own scan is now clean.
	if got := scanCodeSpans(out); len(got) != 0 {
		t.Errorf("detector still finds code spans after transform: %v", got)
	}
	// Splice fidelity: the only change is `x` → <code>x</code>, so the output
	// reconstructs from the input by one string substitution.
	want := strings.Replace(in, "(`ease-in-out`)", "(<code>ease-in-out</code>)", 1)
	if out != want {
		t.Error("output differs from input beyond the single converted span")
	}
	// Idempotency on real bytes.
	out2, n2 := mustConvert(t, out)
	if n2 != 0 || out2 != out {
		t.Errorf("not idempotent: second pass converted %d", n2)
	}
}

func TestConvertCodeSpans_SkipContexts(t *testing.T) {
	// Every case must come back byte-identical with converted==0. Mixed-case
	// tags are DELIBERATE: TagName() lower-cases the buffer Raw() aliases, so
	// a Write-after-TagName reorder passes every all-lowercase case and fails
	// exactly these.
	cases := map[string]string{
		"script":          `<div><script>let a = ` + "`x${y}`" + `;</script></div>`,
		"script mixed":    `<div><SCRIPT>let a = ` + "`x${y}`" + `;</SCRIPT></div>`,
		"style":           `<style>.a::before{content:'` + "`code`" + `'}</style>`,
		"existing code":   `<p>use <code>` + "`raw`" + `</code> here</p>`,
		"pre":             `<pre>run ` + "`make build`" + ` first</pre>`,
		"textarea":        `<textarea>type ` + "`cmd`" + ` here</textarea>`,
		"svg":             `<svg><text>` + "`t`" + `</text></svg>`,
		"attribute value": `<p title="use ` + "`x`" + ` here">plain text</p>`,
		"non-alnum open":  `<p>` + "`${x}`" + ` and ` + "`/api`" + ` never fire</p>`,
		"unterminated":    `<p>a stray ` + "`" + ` backtick</p>`,
		"crosses tags":    `<p>` + "`fetch" + `<em>()</em>` + "`" + `</p>`,
		"contains angle":  `<p>` + "`a < b`" + `</p>`,
		"over length":     `<p>` + "`a" + strings.Repeat("x", 90) + "`" + `</p>`,
		"newline inside":  "<p>`a\nb`</p>",
	}
	for name, in := range cases {
		out, n, err := ConvertLiteralCodeSpansInHTML(in)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", name, err)
			continue
		}
		if n != 0 {
			t.Errorf("%s: converted %d spans, expected 0", name, n)
		}
		if out != in {
			t.Errorf("%s: output not byte-identical to input:\n in: %q\nout: %q", name, in, out)
		}
	}
}

func TestConvertCodeSpans_Converts(t *testing.T) {
	cases := []struct {
		name, in, want string
		n              int
	}{
		{"simple", "<p>use `fetch()` to load</p>", "<p>use <code>fetch()</code> to load</p>", 1},
		{"percent", "<p>set `33%` width</p>", "<p>set <code>33%</code> width</p>", 1},
		{"two spans one node", "<p>`a1` and `b2`</p>", "<p><code>a1</code> and <code>b2</code></p>", 2},
		{"span after skipped subtree", "<div><code>`kept`</code><p>`x1`</p></div>",
			"<div><code>`kept`</code><p><code>x1</code></p></div>", 1},
		{"bare text no tags", "plain `tok` text", "plain <code>tok</code> text", 1},
		{"entity in span", "<p>`a&amp;b`</p>", "<p><code>a&amp;b</code></p>", 1},
	}
	for _, c := range cases {
		out, n := mustConvert(t, c.in)
		if out != c.want || n != c.n {
			t.Errorf("%s:\n got (%d): %q\nwant (%d): %q", c.name, n, out, c.n, c.want)
		}
		// Idempotency for every converting case.
		out2, n2 := mustConvert(t, out)
		if n2 != 0 || out2 != out {
			t.Errorf("%s: not idempotent (second pass converted %d)", c.name, n2)
		}
		// Repair contract for every converting case.
		if got := scanCodeSpans(out); len(got) != 0 {
			t.Errorf("%s: detector still finds %v after transform", c.name, got)
		}
	}
}

// TestConvertCodeSpans_ConversionSubsetOfDetection holds leg 2: any text the
// conversion pattern matches, the detection pattern matches. The generator
// walks a corpus of interior bytes chosen to probe the boundary (the chars the
// two classes disagree on, the length bound, the opening discipline).
func TestConvertCodeSpans_ConversionSubsetOfDetection(t *testing.T) {
	interiors := []string{
		"x", "fetch()", "ease-in-out", "33%", "feTurbulence",
		"a&amp;b", "a b c", "a.b/c_d-e", strings.Repeat("y", 81),
		// Boundary probes: these must NOT convert; if the conversion class
		// widens past detection, one of them starts matching only one side.
		"a<b", "a>b", "a\nb", strings.Repeat("y", 82), "", "$x", "/api",
	}
	for _, interior := range interiors {
		s := "`" + interior + "`"
		conv := codeSpanConvertRe.MatchString(s)
		det := MDCodeSpanRe.MatchString(s)
		if conv && !det {
			t.Errorf("conversion matches %q but detection does not — conversion has grown past detection", s)
		}
	}
}
