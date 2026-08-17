// FILE: platform/orchestration/actions/component_instance_scope_test.go
//
// These tests are built around one rule: a detector that reports nothing is
// indistinguishable from a detector that is inert, so every "clean" assertion
// here is paired with a MUTATION of the same input that must come out dirty.
// Asserting only the clean case would pass against a function that returns an
// empty report unconditionally.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// oneCalculator is the real shape of an LMC calculator component after the
// bugs_open/283 fix: ids namespaced with the instance token, lookups scoped,
// body wrapped in an IIFE, no window.onload assignment.
func oneCalculator(token string) string {
	return `
<section class="card">
  <input id="` + token + `-loanAmount" value="200000">
  <input id="` + token + `-interestRate" value="5">
  <button id="` + token + `-btn-calculate">Calculate</button>
  <div id="` + token + `-displayMonthly"></div>
</section>
<script src="/assets/js/calculators.js"></script>
<script>
(function () {
  const root = document.getElementById('` + token + `-loanAmount');
  root.addEventListener('input', function () {});
})();
</script>`
}

// oneCalculatorOldShape is the shape as it actually ships TODAY: literal ids,
// a top-level function declaration, and window.onload assigned.
func oneCalculatorOldShape() string {
	return `
<section class="card">
  <input id="loanAmount" value="200000">
  <input id="interestRate" value="5">
  <button id="btn-calculate" onclick="runCalc()">Calculate</button>
  <div id="displayMonthly"></div>
</section>
<script src="/assets/js/calculators.js"></script>
<script>
    function runCalc() {
        const P = parseFloat(document.getElementById('loanAmount').value);
        document.getElementById('displayMonthly').textContent = P;
    }
    window.onload = runCalc;
</script>`
}

func TestInstanceToken_isSelectorSafeAndStableAcrossPages(t *testing.T) {
	// An id may begin with a digit and getElementById will find it, but
	// querySelector("#3-foo") throws — so the token must not start with one, and
	// must carry nothing that would break a selector.
	for _, fn := range []string{"mortgages-repayment", "Stamp Duty/SDLT", "", "   ", "!!!", "9lives"} {
		for _, occ := range []int{0, 1, 7} {
			tok := InstanceToken(fn, occ)
			if tok == "" {
				t.Fatalf("InstanceToken(%q,%d) must never be empty — an empty token "+
					"is exactly the missingkey=zero failure this seam removes", fn, occ)
			}
			if tok[0] >= '0' && tok[0] <= '9' {
				t.Fatalf("token %q must start with a letter to be CSS-selector safe", tok)
			}
			if strings.ContainsAny(tok, " /!.#") {
				t.Fatalf("token %q must be selector-safe", tok)
			}
		}
	}

	// THE PROPERTY THE ORACLE NEEDS, and the one `position` and `data_uuid` both
	// fail: the same component alone on two different pages gets the SAME token,
	// whatever else sits beside it. Here it is second on one page and first on
	// the other.
	pageA := InstanceTokensForPage([]string{"hero", "mortgages-repayment", "faq"})
	pageB := InstanceTokensForPage([]string{"mortgages-repayment", "faq"})
	if pageA[1] != pageB[0] {
		t.Fatalf("a single-instance component must take the same token on every page, "+
			"got %q and %q — every hand-written selector would need per-page knowledge",
			pageA[1], pageB[0])
	}

	// And it must actually discriminate, or the assertion above would pass
	// against a function returning one constant.
	if pageA[0] == pageA[1] {
		t.Fatalf("CONTROL FAILED: different components share a token (%q)", pageA[0])
	}
}

func TestInstanceTokensForPage_repeatedComponentGetsDistinctTokens(t *testing.T) {
	got := InstanceTokensForPage([]string{"faq", "mortgages-repayment", "faq", "faq"})

	seen := map[string]int{}
	for _, tok := range got {
		seen[tok]++
	}
	for tok, n := range seen {
		if n > 1 {
			t.Fatalf("token %q assigned to %d instances on one page — %v", tok, n, got)
		}
	}
	// The FIRST instance takes the bare token, so a component that appears once
	// (every interactive component on every live page today) is unaffected by
	// this rule existing at all.
	if got[0] != InstanceToken("faq", 0) {
		t.Fatalf("first instance must take the bare token, got %q", got[0])
	}
	// Case is not a distinction an HTML id can rely on being preserved, so
	// "FAQ" and "faq" must count as the same component rather than silently
	// producing two "first" instances.
	if mixed := InstanceTokensForPage([]string{"FAQ", "faq"}); mixed[0] == mixed[1] {
		t.Fatalf("case-differing functions must still be counted as one component, got %v", mixed)
	}
}

// The single-section paths (RenderComponentAction, the section editor) cannot
// see the page, so they supply occurrence 0. Assert that this AGREES with the
// canonical token in the common case — the whole reason it is not a second
// rule — and that where it does not, the guard is what catches it.
func TestSingleSectionToken_agreesWithCanonical_andCollidesDetectably(t *testing.T) {
	rc := &RenderContext{}
	BindSingleSectionInstanceToken(rc, "mortgages-repayment")
	single, _ := rc.ContentData[InstanceContentKey].(string)

	canonical := InstanceTokensForPage([]string{"hero", "mortgages-repayment"})[1]
	if single != canonical {
		t.Fatalf("the single-section path must supply the SAME token as the canonical "+
			"derivation for a once-per-page component, got %q vs %q — differing tokens "+
			"would mean an instance's ids depend on which action last rendered it",
			single, canonical)
	}

	// Where the assumption is wrong — the component really does appear twice and
	// both were rendered one section at a time — the tokens collide. That is the
	// known cost, and it must be DETECTED rather than silent.
	if DetectInstanceCollisions(oneCalculator(single) + oneCalculator(single)).Clean() {
		t.Fatal("CONTROL FAILED: two same-token instances must be reported by the guard")
	}
}

// The default is the safety-critical property here, not the armed behaviour:
// 13 active pages already emit duplicate ids, so a guard that defaults to
// refusing would fail their next re-render and turn a latent defect into an
// outage. Assert the default explicitly rather than relying on Go's zero value
// staying convenient.
func TestEnforceInstanceScope_defaultsOff(t *testing.T) {
	if enforceInstanceScope(nil) {
		t.Fatal("a nil config must not arm the refusal")
	}
	if enforceInstanceScope(map[string]interface{}{}) {
		t.Fatal("an empty config must not arm the refusal")
	}
	// A non-bool value must not arm it either — a config carrying the string
	// "true" is a misconfiguration, and reading it as armed would refuse renders
	// on the strength of a typo.
	if enforceInstanceScope(map[string]interface{}{"enforce_instance_scope": "true"}) {
		t.Fatal("a string value must not arm the refusal")
	}
	// And it must actually be reachable — otherwise the two assertions above
	// would pass against a function that returns false unconditionally.
	if !enforceInstanceScope(map[string]interface{}{"enforce_instance_scope": true}) {
		t.Fatal("CONTROL FAILED: the guard cannot be armed at all")
	}
}

func TestDetect_fixedShapeIsClean_andMutationProvesTheDetectorCanFail(t *testing.T) {
	toks := InstanceTokensForPage([]string{"mortgages-repayment", "mortgages-repayment"})
	// Two instances, distinct tokens — the whole point of the fix.
	page := oneCalculator(toks[0]) + oneCalculator(toks[1])

	got := DetectInstanceCollisions(page)
	if !got.Clean() {
		t.Fatalf("two correctly-scoped instances must be clean, got: %s", got.Summary())
	}

	// MUTATION 1 — give both instances the SAME token, which is exactly what
	// {{.ComponentID}} does on the rerender path. If this still reads clean the
	// detector is inert and the clean assertion above proved nothing.
	same := oneCalculator(toks[0]) + oneCalculator(toks[0])
	if m := DetectInstanceCollisions(same); m.Clean() {
		t.Fatal("CONTROL FAILED: two instances sharing a token must collide; " +
			"the detector cannot see duplicate ids")
	} else if len(m.DuplicateElementIDs) != 4 {
		t.Fatalf("expected all 4 ids duplicated, got %d: %v",
			len(m.DuplicateElementIDs), m.DuplicateElementIDs)
	}
}

// THE END-TO-END TEST THE COUNCIL'S GUARDIAN SEAT ASKED FOR (bugs_open/283 §3.6).
// Everything above tests the detector against hand-written HTML, which cannot
// tell you whether the RENDER LAYER actually puts two different tokens on two
// instances. This drives the real seam — a template with {{.InstanceID}}, the
// real RenderContext binding, the real RenderTemplate — and asserts on the
// assembled page.
func TestRenderLayer_twoInstancesOnOnePageGetDifferentIDs(t *testing.T) {
	logger := zap.NewNop()
	tmpl := `<section><input id="{{.InstanceID}}-loanAmount"><button id="{{.InstanceID}}-btn-calculate"></button></section>`

	// Render the SAME component twice, as a page-assembling path does: one
	// counter walked in render order.
	instances := NewInstanceCounter()
	var page strings.Builder
	for i := 0; i < 2; i++ {
		rc := &RenderContext{}
		BindInstanceToken(rc, instances.Next("mortgages-repayment"))
		out := RenderTemplate(tmpl, rc, logger)
		if strings.Contains(out, "{{") {
			t.Fatalf("instance %d did not render: %q", i, out)
		}
		if strings.Contains(out, `id="-`) {
			t.Fatalf("instance %d rendered an EMPTY token: %q — this is the "+
				"missingkey=zero failure the seam exists to remove", i, out)
		}
		page.WriteString(out)
	}

	if got := DetectInstanceCollisions(page.String()); !got.Clean() {
		t.Fatalf("two rendered instances must carry distinct ids, got: %s", got.Summary())
	}

	// MUTATION — the regression test for §3.3, the objection that the value was
	// bound at three call sites while the mechanism stayed generic. Render the
	// same template through a path that binds NOTHING (any of the five call
	// sites that did not, before this change). Both instances must then collide,
	// proving the clean result above is produced by the binding and not by the
	// template happening to be harmless.
	var unbound strings.Builder
	for i := 0; i < 2; i++ {
		unbound.WriteString(RenderTemplate(tmpl, &RenderContext{}, logger))
	}
	if DetectInstanceCollisions(unbound.String()).Clean() {
		t.Fatal("CONTROL FAILED: an unbound InstanceID must produce colliding ids — " +
			"either the detector or this test is inert")
	}
}

// The shared layer cannot invent a token, so its job is to make the absence
// loud. Assert the predicate that drives that, since it is what decides whether
// any of the eleven call sites is reported at all.
func TestTemplateNeedsInstanceID_matchesTheSpellingsGoAccepts(t *testing.T) {
	for _, tmpl := range []string{
		`<input id="{{.InstanceID}}-loanAmount">`,
		`<input id="{{ .InstanceID }}-loanAmount">`,
		`<input id="{{- .InstanceID -}}-loanAmount">`,
	} {
		if !TemplateNeedsInstanceID(tmpl) {
			t.Fatalf("must detect the token in %q — an undetected reference renders "+
				"empty and silently collides", tmpl)
		}
	}
	// Must not fire on a template that does not use it, or every render of the
	// other 238 components logs an error nobody can act on.
	for _, tmpl := range []string{
		`<input id="loanAmount">`,
		`<div id="{{.ComponentID}}">`,
		`<div>{{.InstanceIDs}}</div>`,
	} {
		if TemplateNeedsInstanceID(tmpl) {
			t.Fatalf("must not fire on %q", tmpl)
		}
	}
}

func TestDetect_todaysShapeCollidesOnAllThreeClasses(t *testing.T) {
	// One instance of today's shape is already unscoped, but does not yet
	// duplicate anything — the defect is latent, not live. Assert that
	// precisely, because "one instance is broken" would be the wrong claim.
	single := DetectInstanceCollisions(oneCalculatorOldShape())
	if len(single.DuplicateElementIDs) != 0 {
		t.Fatalf("a single instance must not duplicate ids, got %v", single.DuplicateElementIDs)
	}
	if single.UnscopedInlineScripts != 1 {
		t.Fatalf("today's shape declares into global scope; want 1 unscoped, got %d",
			single.UnscopedInlineScripts)
	}
	if single.Clean() {
		t.Fatal("a component that cannot be safely repeated must not read clean")
	}

	// Two instances: all three classes fire at once.
	page := oneCalculatorOldShape() + oneCalculatorOldShape()
	got := DetectInstanceCollisions(page)
	if len(got.DuplicateElementIDs) != 4 {
		t.Fatalf("want 4 duplicate ids, got %d: %v",
			len(got.DuplicateElementIDs), got.DuplicateElementIDs)
	}
	if got.WindowOnloadAssignments != 2 {
		t.Fatalf("want 2 window.onload assignments, got %d", got.WindowOnloadAssignments)
	}
	if got.UnscopedInlineScripts != 2 {
		t.Fatalf("want 2 unscoped scripts, got %d", got.UnscopedInlineScripts)
	}
	// The summary is what a refusal message shows an operator; it must name the
	// cause rather than merely say "invalid".
	if s := got.Summary(); !strings.Contains(s, "window.onload") ||
		!strings.Contains(s, "duplicate element id") {
		t.Fatalf("summary must name each cause, got %q", s)
	}
}

// THE CONVERSION TRAP, and the reason bugs_open/283 §11 sequences ids and
// scripts as ONE step rather than two phases.
//
// Namespacing the element ids is the mechanical, obvious, regex-able half of
// converting a component — and doing ONLY that produces a page which reads clean
// on the id check, looks converted in the template, and still cross-talks. The
// global function name is the reason: both instances' scripts declare
// `runCalc` at top level, the second declaration replaces the first, and every
// button on the page then runs the last instance's logic against the last
// instance's (correctly namespaced, correctly unique) fields.
//
// The fixture's shape is the real one, not one composed to make the point:
// content_components `mortgages-repayment`, read live 2026-08-17, carries 9
// literal ids, 7 getElementById calls, one `onclick="runCalc()"` and
// `window.onload = runCalc`.
func TestIDOnlyConversion_readsCleanOnIDsAndIsStillBroken(t *testing.T) {
	// Today's shape with ONLY the ids namespaced — scripts untouched.
	idOnly := func(token string) string {
		s := oneCalculatorOldShape()
		for _, id := range []string{"loanAmount", "interestRate", "displayMonthly", "btn-calculate"} {
			s = strings.ReplaceAll(s, `id="`+id+`"`, `id="`+token+`-`+id+`"`)
			s = strings.ReplaceAll(s, `getElementById('`+id+`')`, `getElementById('`+token+`-`+id+`')`)
		}
		return s
	}

	toks := InstanceTokensForPage([]string{"mortgages-repayment", "mortgages-repayment"})
	page := idOnly(toks[0]) + idOnly(toks[1])
	got := DetectInstanceCollisions(page)

	// The half that DID work — and this is exactly what makes the trap dangerous.
	if len(got.DuplicateElementIDs) != 0 {
		t.Fatalf("id namespacing should have removed every duplicate id, got %v",
			got.DuplicateElementIDs)
	}

	// The half that did not. Both must still fire, or this test is asserting
	// nothing and the "ids alone are not enough" claim in 283 §11 is unevidenced.
	if got.WindowOnloadAssignments != 2 {
		t.Fatalf("want 2 surviving window.onload assignments, got %d — only the LAST "+
			"one runs, so every earlier instance never initialises",
			got.WindowOnloadAssignments)
	}
	if got.UnscopedInlineScripts != 2 {
		t.Fatalf("want 2 surviving global-scope scripts, got %d — the second "+
			"`function runCalc` replaces the first, so both buttons run the last "+
			"instance's logic", got.UnscopedInlineScripts)
	}
	if got.Clean() {
		t.Fatal("CONTROL FAILED: an id-only conversion must NOT read clean — that is " +
			"the whole finding")
	}
}

// The token is not a valid JavaScript identifier, which removes one of the two
// obvious ways to de-collide the global function name and forces the IIFE route.
// Asserted rather than described, because a converter author will reach for
// `function runCalc_{{.InstanceID}}()` and get a syntax error at render time on
// a page that has already shipped.
func TestInstanceToken_isNotAValidJSIdentifier(t *testing.T) {
	tok := InstanceToken("mortgages-repayment", 0)
	if !strings.Contains(tok, "-") {
		t.Fatalf("token %q no longer contains a hyphen — if the token shape changed to "+
			"something JS-identifier-safe, 283 §11's 'the IIFE route is forced' reasoning "+
			"needs revisiting, not just this test", tok)
	}
}

// The owner's actual use case: different calculators listed on one page. These
// do not repeat a component at all, so an id-uniqueness check scoped to "the
// same component twice" would miss it entirely.
func TestDetect_differentComponentsSharingAnIdNameStillCollide(t *testing.T) {
	stampDuty := `<section><input id="price"><button id="btn-calculate"></button></section>`
	carFinance := `<section><input id="price"><button id="btn-calculate"></button></section>`

	got := DetectInstanceCollisions(stampDuty + carFinance)
	if got.Clean() {
		t.Fatal("two DIFFERENT components sharing id names must collide")
	}
	want := map[string]bool{"price": true, "btn-calculate": true}
	if len(got.DuplicateElementIDs) != len(want) {
		t.Fatalf("want %v, got %v", want, got.DuplicateElementIDs)
	}
	for _, id := range got.DuplicateElementIDs {
		if !want[id] {
			t.Fatalf("unexpected duplicate %q, want only %v", id, want)
		}
	}
}

func TestDetect_ignoresSharedScriptSrcAndUnrenderedTemplates(t *testing.T) {
	// calculators.js is one shared file legitimately referenced by every
	// calculator; counting it as a per-instance body would make every
	// multi-calculator page permanently dirty for no reason.
	twoRefs := `<script src="/assets/js/calculators.js"></script>
	            <script src="/assets/js/calculators.js"></script>`
	if got := DetectInstanceCollisions(twoRefs); !got.Clean() {
		t.Fatalf("shared <script src> must not count as a collision: %s", got.Summary())
	}

	// An UNRENDERED template is not evidence of anything — the substitution has
	// not happened yet, so two occurrences of the same placeholder are expected.
	tmpl := `<input id="{{.InstanceID}}-loanAmount"><input id="{{.InstanceID}}-rate">`
	if got := DetectInstanceCollisions(tmpl); !got.Clean() {
		t.Fatalf("unrendered template must not report collisions: %s", got.Summary())
	}
}
