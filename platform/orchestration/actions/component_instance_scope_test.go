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

func TestInstanceToken_isUniqueAndSelectorSafe(t *testing.T) {
	a, b := InstanceToken(1), InstanceToken(2)
	if a == b {
		t.Fatalf("token must differ by position, got %q for both", a)
	}
	// An id may begin with a digit and getElementById will find it, but
	// querySelector("#3-foo") throws — so the token must not start with one.
	for _, tok := range []string{InstanceToken(0), InstanceToken(3), InstanceToken(42)} {
		if tok == "" || (tok[0] >= '0' && tok[0] <= '9') {
			t.Fatalf("token %q must start with a letter to be CSS-selector safe", tok)
		}
	}
}

func TestInstanceTokenFromSlot_isSafeAndNeverEmpty(t *testing.T) {
	// Never empty: an absent InstanceID renders as "" under missingkey=zero and
	// silently puts every instance back on identical ids.
	for _, in := range []string{"", "   ", "!!!", "---"} {
		if got := InstanceTokenFromSlot(in); got == "" {
			t.Fatalf("InstanceTokenFromSlot(%q) must never be empty", in)
		}
	}
	// Selector-safe: no spaces or punctuation that would break querySelector.
	if got := InstanceTokenFromSlot("tool 1/x"); strings.ContainsAny(got, " /") {
		t.Fatalf("token %q must be selector-safe", got)
	}
	// Positional slots — the common case — do discriminate.
	if InstanceTokenFromSlot("tool-1") == InstanceTokenFromSlot("tool-2") {
		t.Fatal("positional slot names must produce different tokens")
	}
}

// The documented WEAKNESS, asserted deliberately. If someone later "fixes" this
// to look unique, they will have created a token that appears to guarantee
// something it cannot, which is worse than a known-weak one. 13 active pages
// repeat a slot name; on those this returns the same token and the GUARD is what
// catches it, not this function.
func TestInstanceTokenFromSlot_repeatedSlotDoesNotDiscriminate(t *testing.T) {
	if InstanceTokenFromSlot("generic-text-block") != InstanceTokenFromSlot("generic-text-block") {
		t.Fatal("same slot must give the same token - this is a pure function")
	}
	page := oneCalculator(InstanceTokenFromSlot("generic-text-block")) +
		oneCalculator(InstanceTokenFromSlot("generic-text-block"))
	if DetectInstanceCollisions(page).Clean() {
		t.Fatal("the guard must catch what the slot-derived token cannot prevent")
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
	// Two instances, distinct tokens — the whole point of the fix.
	page := oneCalculator(InstanceToken(1)) + oneCalculator(InstanceToken(2))

	got := DetectInstanceCollisions(page)
	if !got.Clean() {
		t.Fatalf("two correctly-scoped instances must be clean, got: %s", got.Summary())
	}

	// MUTATION 1 — give both instances the SAME token, which is exactly what
	// {{.ComponentID}} does on the rerender path. If this still reads clean the
	// detector is inert and the clean assertion above proved nothing.
	same := oneCalculator(InstanceToken(1)) + oneCalculator(InstanceToken(1))
	if m := DetectInstanceCollisions(same); m.Clean() {
		t.Fatal("CONTROL FAILED: two instances sharing a token must collide; " +
			"the detector cannot see duplicate ids")
	} else if len(m.DuplicateElementIDs) != 4 {
		t.Fatalf("expected all 4 ids duplicated, got %d: %v",
			len(m.DuplicateElementIDs), m.DuplicateElementIDs)
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
