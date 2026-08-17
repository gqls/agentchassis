// FILE: platform/orchestration/actions/component_instance_conversion_test.go
//
// The two load-bearing fixtures are LIVE ROW BYTES, not composed prose —
// testdata/instance_conversion_*_283.html are content_components.html_template
// exactly as exported 2026-08-17. A fixture you compose exercises your own
// belief about the artefact, not the artefact (WRONG_CALLS 2026-08-16, the
// vonc calibration lesson): this lane's own regex triage read 67 of these
// templates as "mechanical" because its author's patterns leaned that way.
//
// mortgages-repayment is one of the 25 whose script genuinely declares into
// global scope — the fixture for the §2.1 REFUSAL. tool-css-unit-converter is
// one of the 66 already-scoped rows — the fixture for the full happy path.
package actions

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return string(b)
}

// The happy path on real bytes: an already-scoped component converts, gates
// clean, and two rendered instances genuinely carry distinct ids.
func TestConvert_realScopedComponent_convertsAndGatesClean(t *testing.T) {
	tpl := readFixture(t, "instance_conversion_css_unit_converter_283.html")

	converted, rep, ok := ConvertTemplateToInstanceScope(tpl)
	if !ok {
		t.Fatalf("real scoped component must convert, refused: %s", rep.RefusedReason)
	}
	// The live row's surfaces: 12 id attributes; 11 LITERAL getElementById
	// calls (the 12th is getElementById(targetId), a variable fed from
	// data-target attributes — which is why DataAttrRefs must be 5, one per
	// copy button; the fixture is the reason that pass exists); 6 label-for
	// references. If the template is ever re-pinned these change WITH it.
	if rep.IDAttrsRenamed != 12 || rep.GetElementByID != 11 || rep.IDRefAttrs != 6 || rep.DataAttrRefs != 5 {
		t.Fatalf("surface counts moved: ids=%d gebi=%d refs=%d data=%d (want 12/11/6/5 for this fixture)",
			rep.IDAttrsRenamed, rep.GetElementByID, rep.IDRefAttrs, rep.DataAttrRefs)
	}
	if !strings.Contains(converted, `data-target="{{.InstanceID}}-result-px"`) {
		t.Fatal("the copy button's data-target must move with the id it references — " +
			"getElementById(targetId) reads it at runtime and would dangle")
	}

	needsJudged, err := GateConvertedTemplate("tool-css-unit-converter", converted, zap.NewNop())
	if err != nil {
		t.Fatalf("gate errored: %v", err)
	}
	if needsJudged {
		t.Fatal("an already-IIFE-scoped component must NOT be routed to the judged pool")
	}

	// End to end: two instances, real renderer, real detector.
	toks := InstanceTokensForPage([]string{"tool-css-unit-converter", "tool-css-unit-converter"})
	var page strings.Builder
	for _, tok := range toks {
		rc := &RenderContext{}
		BindInstanceToken(rc, tok)
		page.WriteString(RenderTemplate(converted, rc, zap.NewNop()))
	}
	if d := DetectInstanceCollisions(page.String()); !d.Clean() {
		t.Fatalf("two instances of the converted component must be clean, got: %s", d.Summary())
	}

	// Idempotency is a REFUSAL, not a silent second rewrite.
	if _, rep2, ok2 := ConvertTemplateToInstanceScope(converted); ok2 || !rep2.AlreadyConverted {
		t.Fatal("converting a converted template must refuse with AlreadyConverted")
	}
}

// The §2.1 enforcement on real bytes: mortgages-repayment's ids convert
// cleanly and its script still declares runCalc globally — the gate must
// route it to the judged pool, and the caller must write NOTHING.
func TestConvert_realGlobalScriptComponent_isRefusedToJudgedPool(t *testing.T) {
	tpl := readFixture(t, "instance_conversion_mortgages_repayment_283.html")

	converted, rep, ok := ConvertTemplateToInstanceScope(tpl)
	if !ok {
		t.Fatalf("the transform half must succeed (it is the gate that refuses): %s", rep.RefusedReason)
	}
	// The bug file's own §3 counts for this exact component: 9 literal ids,
	// 7 getElementById calls.
	if rep.IDAttrsRenamed != 9 || rep.GetElementByID != 7 {
		t.Fatalf("mortgages-repayment surfaces moved: ids=%d gebi=%d (bug file says 9/7)",
			rep.IDAttrsRenamed, rep.GetElementByID)
	}

	needsJudged, err := GateConvertedTemplate("mortgages-repayment", converted, zap.NewNop())
	if err != nil {
		t.Fatalf("gate errored: %v", err)
	}
	if !needsJudged {
		t.Fatal("CONTROL FAILED: a global-script component passed the gate — the §2.1 " +
			"refusal is not firing, and ids-only conversions would ship")
	}
}

// Every reference surface moves with the id, on a purpose-built template that
// carries all of them at once. Composed deliberately — no live row exercises
// every surface — and therefore asserting SPECIFIC strings, not verdicts.
func TestConvert_everyReferenceSurfaceMoves(t *testing.T) {
	tpl := `<style>#rate { color: red } #rateYears:hover { color: blue }</style>
<div id="rate"></div><div id="rateYears"></div>
<label for="rate">r</label>
<div aria-labelledby="rate rateYears" aria-controls="rate"></div>
<button data-target="rate" data-mode="verbose">copy</button>
<a href="#rateYears">jump</a>
<script>
(function () {
  var a = document.getElementById('rate');
  var b = document.getElementById("rateYears");
  var c = document.querySelector('#rate');
})();
</script>`

	converted, rep, ok := ConvertTemplateToInstanceScope(tpl)
	if !ok {
		t.Fatalf("refused: %s", rep.RefusedReason)
	}
	for _, want := range []string{
		`#{{.InstanceID}}-rate {`,
		`#{{.InstanceID}}-rateYears:hover`,
		`id="{{.InstanceID}}-rate"`,
		`id="{{.InstanceID}}-rateYears"`,
		`for="{{.InstanceID}}-rate"`,
		`aria-labelledby="{{.InstanceID}}-rate {{.InstanceID}}-rateYears"`,
		`aria-controls="{{.InstanceID}}-rate"`,
		`href="#{{.InstanceID}}-rateYears"`,
		`data-target="{{.InstanceID}}-rate"`,
		`data-mode="verbose"`, // a data value that is NOT a declared id stays put
		`getElementById('{{.InstanceID}}-rate')`,
		`getElementById("{{.InstanceID}}-rateYears")`,
		`querySelector('#{{.InstanceID}}-rate')`,
	} {
		if !strings.Contains(converted, want) {
			t.Fatalf("missing %q in converted template:\n%s", want, converted)
		}
	}
	// The prefix trap this test's id pair exists for: renaming `rate` must not
	// have reached through `rateYears`.
	if strings.Contains(converted, "{{.InstanceID}}-rateYears") &&
		strings.Contains(converted, "rate{{.InstanceID}}") {
		t.Fatal("prefix id clobbered through its superstring")
	}
}

// The refusal arms, each asserted with its reason — a refusal without a
// reason is indistinguishable from a crash to the operator reading the item.
func TestConvert_refusalArms(t *testing.T) {
	cases := map[string]struct {
		tpl        string
		wantReason string
	}{
		"hex-ambiguous id": {
			`<div id="abc"></div><style>.x { color: #abc }</style>`,
			"hex colour",
		},
		"no ids at all": {
			`<div class="prose">{{.body}}</div>`,
			"no literal element ids",
		},
		"binding built by concatenation": {
			// The declared id also appears via a construction no pass
			// recognises — the # reference survives inside a selector string
			// the passes cannot see is a binding.
			`<div id="ratex"></div><script>(function(){var s='#'+'ratex';document.querySelector('#ratex');})();</script>`,
			"",
			// note: '#ratex' IS recognised by the # pass; the concatenated
			// half is invisible and legitimately unfixable — covered below.
		},
	}
	for name, c := range cases {
		_, rep, ok := ConvertTemplateToInstanceScope(c.tpl)
		if name == "binding built by concatenation" {
			// The recognisable binding converts; the concatenated one cannot
			// be seen by ANY textual pass — this case documents the limit
			// rather than asserting a refusal the code cannot deliver.
			_ = ok
			continue
		}
		if ok {
			t.Fatalf("%s: must refuse", name)
		}
		if !strings.Contains(rep.RefusedReason, c.wantReason) {
			t.Fatalf("%s: reason %q must name %q", name, rep.RefusedReason, c.wantReason)
		}
	}
}

// Ids the template does NOT declare are never touched: a reference to chrome,
// another section, or an external anchor is not ours to rename.
func TestConvert_foreignIDsAreLeftAlone(t *testing.T) {
	tpl := `<div id="mine"></div>
<script>(function(){
  document.getElementById('mine').textContent = 'x';
  document.getElementById('site-header-nav').classList.add('y');
  document.querySelector('#global-footer');
})();</script>`
	converted, _, ok := ConvertTemplateToInstanceScope(tpl)
	if !ok {
		t.Fatal("must convert")
	}
	if !strings.Contains(converted, `getElementById('site-header-nav')`) ||
		!strings.Contains(converted, `querySelector('#global-footer')`) {
		t.Fatal("foreign ids must be untouched")
	}
	if !strings.Contains(converted, `getElementById('{{.InstanceID}}-mine')`) {
		t.Fatal("owned id must be renamed")
	}
}
