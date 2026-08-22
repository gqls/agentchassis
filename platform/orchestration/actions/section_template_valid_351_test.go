// FILE: platform/orchestration/actions/section_template_valid_351_test.go
//
// bugs_open/351 — pins the two predicates that stopped the platform reusing a
// calculator it already owned.
//
// THE DEFECT, IN ONE LINE: sectionTemplateValid used the substring "</section>"
// as a proxy for "not truncated", and a calculator is a <div>-wrapped widget
// that never contains one. Measured 2026-08-21: 0 of 22 active section-level
// calculators passed; all 22 were structurally complete.
//
// Every guard below is proven load-bearing by breaking it and watching a NAMED
// test fail. A guard no test can be made to fail against is decoration:
//
//	mutation                                              test that catches it
//	----------------------------------------------------  -------------------------------
//	restore the "</section>" substring test               CalculatorWidgetIsValid
//	drop the UnbalancedStructuralTags arm                 UnbalancedMarkupIsStillInvalid
//	drop the endsCleanly arm                              CutMidStreamIsStillInvalid
//	accept a bare "}}" instead of stripping {{end}}       CutAfterAnActionIsStillInvalid
//	strip only once instead of looping                    NestedTrailingEndsAreStripped
//	strip {{if}}/{{range}}/placeholders too               NonEndTrailingActionIsNotStripped
//	drop the <100 short-template arm                      ShortTemplateStillAllowed
//	make the {{end}} match case-insensitive               CaseVariantIsNotStripped

package actions

import "testing"

// A real calculator: <div>-wrapped, self-contained, ends on </script>.
// Shape taken from b89f91e1 "Mortgages Repayment", the incumbent whose
// invisibility was bugs_open/351's originating case.
const calc351 = `<div class="calc-grid">
  <div class="card">
    <div class="form-group">
      <label>{{.label_1}}</label>
      <input type="number" id="{{.InstanceID}}-loanAmount" value="250000">
    </div>
  </div>
  <script>
    (function () {
      function runCalc() { /* … */ }
      document.getElementById('{{.InstanceID}}-btn').addEventListener('click', runCalc);
    })();
  </script>
</div>`

func TestSectionTemplateValid351_CalculatorWidgetIsValid(t *testing.T) {
	if len(calc351) < 100 {
		t.Fatalf("fixture too short — it would pass on the stub arm and prove nothing")
	}
	if !sectionTemplateValid(calc351) {
		t.Fatalf("a complete <div>-wrapped calculator must be valid; the </section> proxy rejected 22 of these")
	}
}

func TestSectionTemplateValid351_CutMidStreamIsStillInvalid(t *testing.T) {
	cut := `<section class="hero"><div class="inner"><h1>Headline</h1><p>Body copy that runs on for a while so the template clears the hundred byte floor and reaches the real test, then stops mid-sen`
	if sectionTemplateValid(cut) {
		t.Fatalf("a template cut mid-sentence must be invalid — this is the whole point of the guard")
	}
}

func TestSectionTemplateValid351_UnbalancedMarkupIsStillInvalid(t *testing.T) {
	// Ends on '>' so endsCleanly alone would pass it; only the structural arm catches it.
	bad := `<div class="wrap"><div class="inner"><h1>Headline</h1><p>Body copy long enough to clear the hundred byte floor so the structural arm is what decides this case.</p></div>`
	if sectionTemplateValid(bad) {
		t.Fatalf("unbalanced structural tags must be invalid even when the text ends on '>'")
	}
}

func TestSectionTemplateValid351_ShortTemplateStillAllowed(t *testing.T) {
	if !sectionTemplateValid(`<div>stub`) {
		t.Fatalf("short templates are intentional stubs and that arm is unchanged")
	}
	if !sectionTemplateValid("") {
		t.Fatalf("empty template must stay valid")
	}
}

// ── endsCleanly: the trailing {{end}} tolerance ────────────────────────────

func TestEndsCleanly351_ConditionalWrapperIsClean(t *testing.T) {
	// about-commercial-block's real shape, and case-studies-grid's.
	for _, s := range []string{
		`<section class="about"><div class="inner"><p>Copy</p></div></section>{{end}}`,
		`<div class="grid"><script>(function(){})();</script></div>{{end}}`,
		`<section><p>x</p></section>{{- end -}}`,
		`<section><p>x</p></section>  {{ end }}  `,
	} {
		if !endsCleanly(s) {
			t.Errorf("a complete conditionally-wrapped template must end cleanly: %q", s)
		}
	}
}

func TestEndsCleanly351_NestedTrailingEndsAreStripped(t *testing.T) {
	if !endsCleanly(`<section><p>x</p></section>{{end}}{{end}}`) {
		t.Fatalf("nested wrappers close with repeated {{end}} — the strip must loop")
	}
}

// THE DISCRIMINATING CASE, and the reason a bare "}}" suffix rule was rejected:
// a template cut immediately after a complete mid-template action also ends "}}".
func TestEndsCleanly351_CutAfterAnActionIsStillInvalid(t *testing.T) {
	for _, s := range []string{
		`<section><p>Intro copy</p>{{if .flag}}`,
		`<section><p>{{.placeholder}}`,
		`<section><ul>{{range .items}}`,
	} {
		if endsCleanly(s) {
			t.Errorf("a cut ending on a non-{{end}} action must NOT read as clean: %q", s)
		}
	}
}

func TestEndsCleanly351_NonEndTrailingActionIsNotStripped(t *testing.T) {
	// Only {{end}} is strippable. Stripping any action would pass the cuts above.
	if endsCleanly(`<section><p>x</p></section>{{template "foo"}}`) {
		t.Fatalf("only {{end}} may be stripped — a trailing non-end action is suspicious")
	}
}

func TestEndsCleanly351_CaseVariantIsNotStripped(t *testing.T) {
	// `end` is a Go template keyword with no case variants; matching {{END}}
	// would widen the tolerance for nothing and is deliberately not done.
	if endsCleanly(`<section><p>x</p></section>{{END}}`) {
		t.Fatalf("{{END}} is not a Go template keyword and must not be stripped")
	}
}

func TestEndsCleanly351_PlainMarkupUnaffected(t *testing.T) {
	if !endsCleanly(`<section><p>x</p></section>`) {
		t.Fatalf("ordinary markup must be unaffected by the change")
	}
	if endsCleanly(`<section><p>truncated here`) {
		t.Fatalf("ordinary truncation must be unaffected by the change")
	}
}
