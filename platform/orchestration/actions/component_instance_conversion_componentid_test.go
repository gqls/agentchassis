// FILE: platform/orchestration/actions/component_instance_conversion_componentid_test.go
//
// Pass 0 of the deterministic converter: an id written as the OLD per-instance
// placeholder, {{.ComponentID}}.
//
// RFC_032 §8 (owner ruling 2026-08-22) converges the estate on {{.InstanceID}}
// and retires {{.ComponentID}}. The five templates still spelling the old name
// could not be filed through the proven conversion pipeline at all, because the
// declared-id harvest excludes braces and so reported "declares no literal
// element ids" on a template whose only id is exactly the one being fixed.
//
// Every test here was run against the UNPATCHED converter first and observed to
// fail for the stated reason — a test that has never failed is not evidence
// that the code does anything.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// The live generic-text-block shape as of 2026-08-22: one id, templated, and
// nothing binds it in script. 179 placements across 152 pages, 21 domains.
const genericTextBlockShape = `<section id="{{.ComponentID}}" class="section section--generic">
  <div class="container">
    <h2 class="section__title">{{.Title}}</h2>
    <div class="section__content">{{.Body}}</div>
  </div>
</section>`

func TestConvertTemplateToInstanceScope_TemplatedIDOnly(t *testing.T) {
	// Against the unpatched converter this failed with
	// ok=false, RefusedReason="template declares no literal element ids".
	out, rep, ok := ConvertTemplateToInstanceScope(genericTextBlockShape)
	if !ok {
		t.Fatalf("wrapper-only template refused: %s", rep.RefusedReason)
	}
	if rep.TemplatedIDSwaps != 1 {
		t.Errorf("TemplatedIDSwaps = %d, want 1", rep.TemplatedIDSwaps)
	}
	if !strings.Contains(out, `id="{{.InstanceID}}"`) {
		t.Errorf("wrapper id not swapped to the instance token:\n%s", out)
	}
	if strings.Contains(out, "ComponentID") {
		t.Errorf("the retired placeholder survived conversion:\n%s", out)
	}
	// The BARE token, never the {{.InstanceID}}- prefix: this id IS the
	// instance identity. The prefix form renders `c-generic-text-block-`
	// with nothing after the hyphen, which is a different defect wearing a
	// converted template's clothes.
	if strings.Contains(out, `id="{{.InstanceID}}-"`) {
		t.Errorf("wrapper took the PREFIX form, leaving a trailing hyphen:\n%s", out)
	}
}

// The acceptance gate is the half of the machinery that survives unchanged from
// the 124-row programme, so it must be shown to actually engage on this shape —
// rendering TWO instances through the real renderer and asking the real
// detector, not merely reporting that a placeholder is present.
func TestGateConvertedTemplate_TemplatedIDOnly(t *testing.T) {
	logger := zap.NewNop()

	// CONTROL first: the gate must REFUSE the unconverted template. Without
	// this, the pass below is not evidence — a gate that accepts everything
	// returns the same "clean" for a conversion that did nothing.
	if _, err := GateConvertedTemplate("generic-text-block", genericTextBlockShape, logger); err == nil {
		t.Fatal("gate accepted the UNCONVERTED template — it cannot then vouch for a converted one")
	}

	converted, rep, ok := ConvertTemplateToInstanceScope(genericTextBlockShape)
	if !ok {
		t.Fatalf("conversion refused: %s", rep.RefusedReason)
	}
	needsJudged, err := GateConvertedTemplate("generic-text-block", converted, logger)
	if err != nil {
		t.Fatalf("gate rejected the converted template: %v", err)
	}
	if needsJudged {
		t.Error("wrapper-only template routed to the judged pool; it declares no script at all")
	}
}

// The defect this whole seam exists to remove, asserted at the bytes: two
// instances on one page must not share an id. Asserting only that the output
// "contains a substitution" passes in the broken world too.
func TestConvertedTemplate_TwoInstancesDoNotCollide(t *testing.T) {
	logger := zap.NewNop()
	converted, rep, ok := ConvertTemplateToInstanceScope(genericTextBlockShape)
	if !ok {
		t.Fatalf("conversion refused: %s", rep.RefusedReason)
	}

	toks := InstanceTokensForPage([]string{"generic-text-block", "generic-text-block"})
	var page strings.Builder
	for _, tok := range toks {
		rc := &RenderContext{}
		BindInstanceToken(rc, tok)
		rendered, _, _, err := RenderTemplate(converted, rc, logger)
		if err != nil {
			t.Fatalf("render failed for token %q: %v", tok, err)
		}
		page.WriteString(rendered)
	}

	// Positive expectation, not merely an absence: both tokens must actually
	// appear. A render that produced two EMPTY ids would also report "no
	// duplicate non-empty ids" from the detector, because reElementID cannot
	// match an empty id — so the absence alone is not evidence.
	for _, tok := range toks {
		if !strings.Contains(page.String(), `id="`+tok+`"`) {
			t.Errorf("token %q absent from the assembled page — ids may have rendered empty:\n%s", tok, page.String())
		}
	}
	if toks[0] == toks[1] {
		t.Fatalf("the two instance tokens are identical (%q) — the counter did not advance", toks[0])
	}
	if collisions := DetectInstanceCollisions(page.String()); !collisions.Clean() {
		t.Errorf("two instances collide after conversion: %s", collisions.Summary())
	}

	// And the mutation control: the SAME assertion against the UNCONVERTED
	// template must fail, or it is not testing anything.
	var bad strings.Builder
	for range toks {
		rc := &RenderContext{}
		rc.ContentData = map[string]interface{}{"ComponentID": "8d81e665-3ee0-443d-a873-690268c15fbb"}
		rendered, _, _, err := RenderTemplate(genericTextBlockShape, rc, logger)
		if err != nil {
			t.Fatalf("control render failed: %v", err)
		}
		bad.WriteString(rendered)
	}
	if DetectInstanceCollisions(bad.String()).Clean() {
		t.Error("the detector reports the UNCONVERTED two-instance page as clean — it cannot then certify the converted one")
	}
}

// A template carrying literal ids AND the retired placeholder converted
// "successfully" before pass 0: the literals were prefixed and the templated id
// was silently ignored, producing id="" on every instance — which the gate
// could not see, because the collision detector cannot match an empty id.
// Measured 2026-08-22: no live template is in this state. The test is for the
// arrival that half-learns the convention.
func TestConvertTemplateToInstanceScope_MixedLiteralAndTemplatedID(t *testing.T) {
	mixed := `<section id="{{.ComponentID}}" class="tool">
  <input id="amount" type="number">
  <button id="calc">Go</button>
  <script>
    (function () {
      var root = document;
      root.getElementById('calc').addEventListener('click', function () {
        root.getElementById('amount').value = 1;
      });
    })();
  </script>
</section>`
	out, rep, ok := ConvertTemplateToInstanceScope(mixed)
	if !ok {
		t.Fatalf("mixed template refused: %s", rep.RefusedReason)
	}
	if rep.TemplatedIDSwaps != 1 {
		t.Errorf("TemplatedIDSwaps = %d, want 1", rep.TemplatedIDSwaps)
	}
	if strings.Contains(out, "ComponentID") {
		t.Errorf("the retired placeholder survived on a MIXED template — this is the half-state the gate cannot catch:\n%s", out)
	}
	for _, want := range []string{`id="{{.InstanceID}}"`, `id="{{.InstanceID}}-amount"`, `id="{{.InstanceID}}-calc"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in:\n%s", want, out)
		}
	}
}

// The placeholder somewhere pass 0 cannot safely rewrite must REFUSE the whole
// component rather than convert around it. Half-converted is the one outcome
// this file's failure direction forbids.
func TestConvertTemplateToInstanceScope_ComponentIDOutsideIDAttrRefuses(t *testing.T) {
	for name, tpl := range map[string]string{
		"data attribute": `<section id="wrap" data-target="{{.ComponentID}}"><b id="x">hi</b></section>`,
		"css selector":   `<section id="wrap"><style>#{{.ComponentID}} { color: red; }</style></section>`,
		"script literal": `<section id="wrap"><script>var root = "{{.ComponentID}}";</script></section>`,
	} {
		t.Run(name, func(t *testing.T) {
			out, rep, ok := ConvertTemplateToInstanceScope(tpl)
			if ok {
				t.Fatalf("converted around a %s reference instead of refusing:\n%s", name, out)
			}
			if !strings.Contains(rep.RefusedReason, "ComponentID") {
				t.Errorf("refusal does not name the surviving placeholder: %q", rep.RefusedReason)
			}
			if out != tpl {
				t.Error("a refusal must return the template unmodified")
			}
		})
	}
}

// Idempotence: the retirement must not be re-runnable over its own output, the
// same contract the literal-id passes have.
func TestConvertTemplateToInstanceScope_SwapIsNotReRunnable(t *testing.T) {
	once, _, ok := ConvertTemplateToInstanceScope(genericTextBlockShape)
	if !ok {
		t.Fatal("first conversion refused")
	}
	twice, rep, ok := ConvertTemplateToInstanceScope(once)
	if ok {
		t.Error("second conversion reported success; it must be a declared no-op")
	}
	if !rep.AlreadyConverted {
		t.Errorf("AlreadyConverted not set on re-run; reason was %q", rep.RefusedReason)
	}
	if twice != once {
		t.Error("re-run mutated an already-converted template")
	}
}

// A mismatched quote pair around the placeholder. Go's regexp is RE2 and has no
// backreferences, so the two quotes are separate capture groups and the pattern
// alone cannot require them to match — the pairing is enforced in the replace
// function (council round 1: editquality edit 3 and bug_historian edit 1, both
// asking for a `\2` that cannot be written in this engine).
//
// The output must not be a tidied `id="{{.InstanceID}}"`: that would launder a
// malformed attribute the browser never parsed as an id into a well-formed one,
// which is a silent repair of evidence, not a conversion.
func TestConvertTemplateToInstanceScope_MismatchedQuotesAreNotSwapped(t *testing.T) {
	tpl := `<section id="{{.ComponentID}}' class="section"><p>body</p></section>`
	out, rep, ok := ConvertTemplateToInstanceScope(tpl)
	if ok {
		t.Fatalf("a mismatched-quote attribute must not convert:\n%s", out)
	}
	if rep.TemplatedIDSwaps != 0 {
		t.Errorf("TemplatedIDSwaps = %d, want 0 — the malformed attribute was counted as a swap", rep.TemplatedIDSwaps)
	}
	if strings.Contains(out, "InstanceID") {
		t.Errorf("the malformed attribute was rewritten:\n%s", out)
	}
	if out != tpl {
		t.Error("a refusal must return the template unmodified")
	}
	if !strings.Contains(rep.RefusedReason, "ComponentID") {
		t.Errorf("the refusal must name the placeholder, got %q", rep.RefusedReason)
	}
	// CONTROL: the same template with MATCHING quotes converts. Without it this
	// test passes against a converter that refuses everything.
	good := `<section id="{{.ComponentID}}" class="section"><p>body</p></section>`
	if _, repGood, okGood := ConvertTemplateToInstanceScope(good); !okGood || repGood.TemplatedIDSwaps != 1 {
		t.Fatalf("control failed: the well-formed twin must convert (ok=%v swaps=%d reason=%q)",
			okGood, repGood.TemplatedIDSwaps, repGood.RefusedReason)
	}
}

// The refusal REASON is part of the converter's contract, because
// tool_birth_instance_scope.go routes on its text: a template with no literal
// ids takes an arm that persists the caller's bytes VERBATIM in both modes.
//
// A template with no literal ids that carries {{.ComponentID}} somewhere pass 0
// cannot rewrite reaches that same early return — and it is NOT inert, because
// the placeholder renders the same value on every instance. So the two cases
// must leave under reasons that can be told apart. (The call-site consequence is
// pinned in tool_birth_instance_scope_test.go; this pins the reason itself.)
func TestConvertTemplateToInstanceScope_NoLiteralIDsButComponentIDElsewhere(t *testing.T) {
	for name, tpl := range map[string]string{
		"data attribute": `<section data-target="{{.ComponentID}}"><p>body</p></section>`,
		"css selector":   `<section><style>#{{.ComponentID}} { color: red; }</style></section>`,
		"script literal": `<section><script>var root = "{{.ComponentID}}";</script></section>`,
	} {
		t.Run(name, func(t *testing.T) {
			_, rep, ok := ConvertTemplateToInstanceScope(tpl)
			if ok {
				t.Fatalf("%s: converted a template pass 0 cannot fix", name)
			}
			// The TYPED signals are the contract the birth guard routes on
			// (council r2, bug_historian: a bool cannot be reworded). Both must
			// be set here — no literal ids AND a placeholder pass 0 cannot swap
			// — and it is that COMBINATION that must not read as inert.
			if !rep.NoLiteralElementIDs {
				t.Errorf("NoLiteralElementIDs not set; the guard cannot classify this refusal")
			}
			if !rep.ComponentIDUnswappable {
				t.Errorf("ComponentIDUnswappable not set — the birth guard's inert-safe arm would swallow this template")
			}
			if !strings.Contains(rep.RefusedReason, "ComponentID") {
				t.Errorf("the reason should still NAME the placeholder for a human reading it: %q", rep.RefusedReason)
			}
		})
	}
	// CONTROL: a template with genuinely nothing to scope keeps the ORIGINAL
	// reason, unqualified. The narrowing must distinguish the two cases, not
	// rename both — the inert-safe pass-through is correct for this one.
	_, rep, ok := ConvertTemplateToInstanceScope(`<div class="prose"><p>No ids here.</p></div>`)
	if ok {
		t.Fatal("control: an id-less template cannot convert")
	}
	if !rep.NoLiteralElementIDs || rep.ComponentIDUnswappable {
		t.Fatalf("control failed: a genuinely id-less template must be NoLiteralElementIDs without ComponentIDUnswappable (got %v/%v) — otherwise the split classified both cases the same way",
			rep.NoLiteralElementIDs, rep.ComponentIDUnswappable)
	}
	if !strings.Contains(rep.RefusedReason, "declares no literal element ids") || strings.Contains(rep.RefusedReason, "ComponentID") {
		t.Fatalf("control failed: an id-less template must keep the plain reason, got %q", rep.RefusedReason)
	}
}

// The pre-existing contract this change must NOT break: TemplateNeedsInstanceID
// keys on {{.InstanceID}} only, so a template still spelling the old name is
// correctly reported as NOT namespaced. component_instance_scope_test.go pins
// this too; repeated here because pass 0 is the change most likely to blur it.
func TestTemplateNeedsInstanceID_UnmovedByPass0(t *testing.T) {
	if TemplateNeedsInstanceID(`<div id="{{.ComponentID}}">`) {
		t.Error("TemplateNeedsInstanceID fired on the retired placeholder")
	}
	converted, _, ok := ConvertTemplateToInstanceScope(genericTextBlockShape)
	if !ok {
		t.Fatal("conversion refused")
	}
	if !TemplateNeedsInstanceID(converted) {
		t.Error("a converted template is not reported as needing the token — the render seam would not bind it")
	}
}
