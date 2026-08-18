package discovery_checks

import (
	"reflect"
	"strings"
	"testing"
)

// A structurally whole tool: balanced script/style/section/div, ends on a
// closed tag. Padded past truncationMinLen so the stub exemption does not apply.
const wholeTool = `<style>.t{color:#000}</style>` +
	`<section class="tool"><div class="calc"><fieldset><label>x</label>` +
	`<input id="x"></fieldset><div id="out"></div></div></section>` +
	`<script>(function(){var v=1;document.getElementById('x');})();</script>` +
	`<!-- pad pad pad pad pad pad pad pad pad pad pad pad pad pad pad pad -->`

func TestUnterminatedTagPairs(t *testing.T) {
	// The census signature: everything balanced up to a <script> that is opened
	// and never closed — the exact shape of the bugs_open/046 casualties. Padded
	// past truncationMinLen.
	censusCut := `<style>.t{color:#000}</style>` +
		`<section><div class="calc"><input></div></section>` +
		`<script>function go(){var total=0;total+=1;` +
		strings.Repeat("compute();", 8)

	cases := []struct {
		name string
		html string
		want []string
	}{
		{"whole tool balances", wholeTool, nil},
		{"census cut: script opened never closed", censusCut, []string{"<script"}},
		{"unterminated style", `<style>.a{color:red}` + strings.Repeat("x", 100), []string{"<style"}},
		{"unterminated div", `<div class="a"><span>hi</span>` + strings.Repeat("y", 100), []string{"<div"}},
		{"unterminated section", `<section class="a"><p>hi</p>` + strings.Repeat("z", 100), []string{"<section"}},
		{"unterminated fieldset", `<fieldset><label>a</label>` + strings.Repeat("w", 100), []string{"<fieldset"}},
		{
			// Cut inside the <style> body: to a browser (and to the markup-
			// context counter, bugs_open/303) everything after the cut is CSS
			// text, so only <style reports unbalanced. The old substring
			// counter also listed the section/div/script that follow the cut;
			// the VERDICT (truncated) is identical, the token list shorter.
			"cut inside a raw-text body reports the enclosing tag",
			`<style>.a{}` + `<section><div><script>x` + strings.Repeat("q", 100),
			[]string{"<style"},
		},
		{
			"multiple unterminated in markup context, fixed order",
			`<section><div><script>x();</script><script>y` + strings.Repeat("q", 100),
			[]string{"<script", "<section", "<div"},
		},
		{
			// bugs_open/303: a tool whose own JavaScript MENTIONS tags must not
			// sweep as a truncation casualty (and, via the verifier, an item on
			// a since-fixed tool must be resolvable).
			"mentions inside script are not unterminated tags",
			`<section class="t"><div id="o"></div></section>` +
				`<script>/* protect <style> and <script> blocks */var r=/<div[^>]*>/g;x();</script>` +
				strings.Repeat(" ", 40),
			nil,
		},
		{"case-insensitive open tag", `<SCRIPT>var x=1;` + strings.Repeat("p", 100), []string{"<script"}},
		{"short stub exempt", `<script>x`, nil},
		{"empty exempt", ``, nil},
		{"self-closing-ish balanced script src", `<script src="/a.js"></script>` + strings.Repeat(" ", 100), nil},
	}

	for _, tc := range cases {
		got := unterminatedTagPairs(tc.html)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: unterminatedTagPairs = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// The hand-maintained mirror of the pair list — and the drift guard that
// policed it — was retired by bugs_open/303: this check now imports the
// canonical list directly (content.StructuralTagPairs, a leaf package both
// sides can reach). The list itself is pinned by TestStructuralTagPairsPinned
// in platform/content/markup_balance_test.go.
