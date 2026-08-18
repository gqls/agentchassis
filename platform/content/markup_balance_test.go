package content

import (
	"reflect"
	"strings"
	"testing"
)

// A structurally whole tool of the class bugs_open/303 is about: its own
// JavaScript mentions structural tags in a comment, a regex and a string —
// unpaired mentions that a substring count reads as unterminated opens. This
// is the bug file's verification recipe (c0cfb873's shape with the injected
// comment): it MUST pass markup-context counting and MUST fail the old
// substring counting — asserted by TestSubstringCountWouldRefuseMentionTool
// below, so this fixture cannot silently stop exercising the defect.
const mentionTool = `<section class="tool"><div class="io">` +
	`<fieldset><label for="in">Input</label><textarea id="in"></textarea></fieldset>` +
	`<div id="out"></div></div>` +
	`<style>.tool textarea{width:100%}</style>` +
	`<script>
	// protect <style> and <script> blocks before minifying
	var protect = /<script[^>]*>/g;
	var alt = /<(pre|textarea|script|style)\b[^>]*>/g;
	function explain(){ return "wraps every <div> it finds"; }
	(function(){ document.getElementById('in'); })();
	</script></section>`

// oldSubstringUnbalanced is the PRE-FIX predicate, reproduced verbatim so the
// tests can prove fixtures actually discriminate old from new behaviour.
func oldSubstringUnbalanced(html string) []string {
	folded := strings.ToLower(html)
	var bad []string
	for _, p := range StructuralTagPairs {
		if strings.Count(folded, p.Open) > strings.Count(folded, p.Close) {
			bad = append(bad, p.Open)
		}
	}
	return bad
}

func TestStructuralTagPairsPinned(t *testing.T) {
	// The canonical list, pinned. It used to be pinned twice — actions.balancedPairs
	// and a hand-maintained mirror in discovery_checks (import cycle) — both now
	// delegate here. Changing it means re-running the calibration in
	// component_write_guard.go's header.
	want := []StructuralTagPair{
		{"<script", "</script>"},
		{"<style", "</style>"},
		{"<section", "</section>"},
		{"<div", "</div>"},
		{"<fieldset", "</fieldset>"},
	}
	if !reflect.DeepEqual(StructuralTagPairs, want) {
		t.Fatalf("StructuralTagPairs = %v, want %v", StructuralTagPairs, want)
	}
}

func TestUnbalancedStructuralTags(t *testing.T) {
	cases := []struct {
		name string
		html string
		want []string
	}{
		// ---- the bugs_open/303 false-positive class: mentions, not markup ----
		{"tool that mentions tags in JS comment/regex/string balances", mentionTool, nil},
		{
			"unpaired mention in an HTML comment is not markup",
			`<div class="a">text</div><!-- an unpaired <style mention, and <script too -->`,
			nil,
		},
		{
			"tag literal inside a CSS string is style text",
			`<style>.a::before{content:"<div>"}</style><div>x</div></div>` /* stray close is fine */, nil,
		},
		{
			"substring is not a tag: <divider does not open a div",
			`<divider><span>x</span></divider>`,
			nil,
		},
		{
			"'>' inside a quoted attribute does not end the tag",
			`<div data-note="a>b" class="c"><span>x</span></div>`,
			nil,
		},

		// ---- the true-positive class the guard exists for (012/046) ----
		{
			"cut mid-JavaScript: script opened, never closed",
			`<section><div>x</div></section><script>var tiers=['Common','Rare','Epic`,
			[]string{"<script"},
		},
		{
			"cut mid-markup: div left open",
			`<section><div class="calc"><p>x</p>`,
			[]string{"<section", "<div"},
		},
		{
			"cut inside the open tag itself",
			`<section><p>x</p></section><script src="/a.js`,
			[]string{"<script"},
		},
		{
			"cut right at the open token",
			`<section><p>x</p></section><script`,
			[]string{"<script"},
		},
		{
			"close cut before its '>' does not count as a close",
			`<script>var x=1;</script`,
			[]string{"<script"},
		},
		{
			"unterminated style: everything after the cut is style text",
			// The old counter also reported the <section>/<div> that follow the
			// cut; a browser reads them as CSS, so only <style is unbalanced.
			// Same verdict (truncated), shorter and more honest token list.
			`<style>.a{color:red` + `<section><div>`,
			[]string{"<style"},
		},

		// ---- browser-faithful raw-text handling ----
		{
			// bugs_open/303 ADDENDUM — the class that cannot be reworded: a
			// tool whose OUTPUT is a script tag. The open literal must appear
			// in the template string and the close MUST be escaped (<\/script>
			// — an unescaped close would terminate the outer script; a hard JS
			// rule, not style). The substring count is imbalanced BY
			// CONSTRUCTION for every correct implementation; in markup context
			// both mentions sit in the outer script's raw-text body, and the
			// element balances exactly as a browser reads it.
			"tool that OUTPUTS a script tag: open literal + escaped close in a template string",
			"<section><div id=\"out\"></div></section>" +
				"<script>const s=`<script type=\"application/ld+json\">${json}<\\/script>`;emit(s);</script>",
			nil,
		},
		{
			"literal </script> inside a JS string ends the element (as in a browser)",
			`<script>var s="</script>";`,
			nil, // open 1, close 1 — the trailing '";' is page text, not a tag imbalance
		},
		{
			"</scriptFoo inside JS does not end the element",
			`<script>var s="</scriptic>";</script>`,
			nil,
		},
		{"case-insensitive", `<SCRIPT>var x=1;`, []string{"<script"}},
		{"empty input", ``, nil},
		{"unterminated comment consumes the rest", `<div>x</div><!-- cut here <div`, nil},
	}

	for _, tc := range cases {
		if got := UnbalancedStructuralTags(tc.html); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: UnbalancedStructuralTags = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestSubstringCountWouldRefuseMentionTool proves the central fixture still
// discriminates: the OLD substring predicate refuses mentionTool (that is
// bugs_open/303) while the markup-context scanner passes it. If a future edit
// to the fixture makes the old predicate pass too, this test fails — the
// fixture would no longer be exercising the defect it was written for.
func TestSubstringCountWouldRefuseMentionTool(t *testing.T) {
	if got := oldSubstringUnbalanced(mentionTool); len(got) == 0 {
		t.Fatalf("fixture no longer trips the substring counter — it stopped exercising bugs_open/303")
	}
	if got := UnbalancedStructuralTags(mentionTool); len(got) != 0 {
		t.Fatalf("markup-context scanner refuses the mention tool: %v — the bugs_open/303 false positive is back", got)
	}

	// The addendum class: an output-tool template whose substring imbalance is
	// forced by JS escaping rules, so no rewording can dodge the old counter.
	outputTool := "<section><div id=\"out\"></div></section>" +
		"<script>const s=`<script type=\"application/ld+json\">${json}<\\/script>`;emit(s);</script>"
	if got := oldSubstringUnbalanced(outputTool); len(got) == 0 {
		t.Fatalf("output-tool fixture no longer trips the substring counter — it stopped exercising the 303 addendum")
	}
	if got := UnbalancedStructuralTags(outputTool); len(got) != 0 {
		t.Fatalf("markup-context scanner refuses the output tool: %v", got)
	}
}

// TestTruncationVerdictAgreesWithOldCounterOnCuts: for genuinely cut inputs the
// two predicates must agree there IS an imbalance (the token lists may differ —
// see the raw-text reporting note in the file header).
func TestTruncationVerdictAgreesWithOldCounterOnCuts(t *testing.T) {
	cuts := []string{
		`<section><div>x</div></section><script>var tiers=['Common','Rare','Epic`,
		`<style>.a{color:red` + `<section><div>`,
		`<section><div class="calc"><p>x</p>`,
		`<fieldset><label>a</label>` + strings.Repeat("w", 100),
	}
	for _, cut := range cuts {
		if len(oldSubstringUnbalanced(cut)) == 0 {
			t.Errorf("old counter passes a cut input — bad fixture: %.60q", cut)
		}
		if len(UnbalancedStructuralTags(cut)) == 0 {
			t.Errorf("scanner passes a cut input the old counter caught: %.60q", cut)
		}
	}
}

func TestStructuralTagCountsPerPair(t *testing.T) {
	counts := StructuralTagCounts(`<div><div><script>var a="<div>";</script></div>`)
	byOpen := map[string]TagBalance{}
	for _, tb := range counts {
		byOpen[tb.Open] = tb
	}
	if d := byOpen["<div"]; d.Opens != 2 || d.Closes != 1 {
		t.Errorf("<div counts = %d/%d, want 2/1 (the mention inside the script must not count)", d.Opens, d.Closes)
	}
	if s := byOpen["<script"]; s.Opens != 1 || s.Closes != 1 {
		t.Errorf("<script counts = %d/%d, want 1/1", s.Opens, s.Closes)
	}
}
