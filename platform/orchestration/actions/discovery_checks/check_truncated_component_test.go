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
			"multiple unterminated, fixed order",
			`<style>.a{}` + `<section><div><script>x` + strings.Repeat("q", 100),
			[]string{"<script", "<style", "<section", "<div"},
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

// truncationTagPairsMirrorGuard fails if truncationTagPairs drifts from the
// canonical list in actions.balancedPairs (component_write_guard.go). The
// package boundary forbids importing that list (actions imports discovery_checks
// — the reverse is an import cycle), so this hardcoded expectation is the guard:
// change one list without the other and this test breaks, forcing reconciliation.
func TestTruncationTagPairsMirrorGuard(t *testing.T) {
	// MUST equal actions.balancedPairs in
	// platform/orchestration/actions/component_write_guard.go.
	want := [][2]string{
		{"<script", "</script>"},
		{"<style", "</style>"},
		{"<section", "</section>"},
		{"<div", "</div>"},
		{"<fieldset", "</fieldset>"},
	}
	if len(truncationTagPairs) != len(want) {
		t.Fatalf("truncationTagPairs has %d pairs, want %d — reconcile with actions.balancedPairs",
			len(truncationTagPairs), len(want))
	}
	for i, w := range want {
		if truncationTagPairs[i].open != w[0] || truncationTagPairs[i].close != w[1] {
			t.Errorf("truncationTagPairs[%d] = {%q,%q}, want {%q,%q} — reconcile with actions.balancedPairs",
				i, truncationTagPairs[i].open, truncationTagPairs[i].close, w[0], w[1])
		}
	}
}
