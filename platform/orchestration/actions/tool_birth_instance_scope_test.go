package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// Fixtures are the three shapes the generator actually produces, minimal.

// class A: ids + lookups, script already IIFE-scoped — mechanically convertible.
const birthClassA = `<div id="calc-box"><input id="amount" type="number"><button id="btn-go">Go</button><div id="result"></div></div>
<script>(function () { 'use strict';
  var btn = document.getElementById('btn-go');
  btn.addEventListener('click', function () {
    var v = document.getElementById('amount').value;
    document.getElementById('result').textContent = v;
  });
})();</script>`

// judged class: the script declares into global scope — not mechanically provable.
const birthGlobalScope = `<div><input id="amount"><button id="btn-go">Go</button><div id="result"></div></div>
<script>function go() {
  document.getElementById('result').textContent = document.getElementById('amount').value;
}
document.getElementById('btn-go').addEventListener('click', go);</script>`

// self-converted: the generator followed the prompt and emitted the convention itself.
const birthSelfConverted = `<div><input id="{{.InstanceID}}-amount"><button id="{{.InstanceID}}-btn-go">Go</button><div id="{{.InstanceID}}-result"></div></div>
<script>(function () { 'use strict';
  document.getElementById('{{.InstanceID}}-btn-go').addEventListener('click', function () {
    document.getElementById('{{.InstanceID}}-result').textContent = document.getElementById('{{.InstanceID}}-amount').value;
  });
})();</script>`

const birthNoIDs = `<div class="prose"><p>No interactive elements here.</p></div>`

func TestScopeToolBirth_armedConvertsClassA(t *testing.T) {
	tpl, rendered, info, refuse := ScopeToolBirthTemplate(birthClassA, "my-tool", true, zap.NewNop())
	if refuse != nil {
		t.Fatalf("class-A template must convert, got refusal: %v", refuse)
	}
	if !strings.Contains(tpl, `id="{{.InstanceID}}-amount"`) {
		t.Fatalf("armed mode must persist the CONVERTED template; got: %s", tpl[:120])
	}
	if strings.Contains(rendered, "{{.InstanceID}}") {
		t.Fatalf("rendered bytes must have the token BOUND — literal placeholder would ship to a live page")
	}
	if !strings.Contains(rendered, `id="c-my-tool-amount"`) {
		t.Fatalf("rendered bytes must carry the occurrence-0 token; got: %s", rendered[:160])
	}
	if got := info["instance_scope"]; got != "mechanically converted at birth" {
		t.Fatalf("info = %v", got)
	}
	// The converted output, doubled on one page, must be collision-free — and
	// the check must be able to fail: the UNCONVERTED template doubled is dirty.
	if c := DetectInstanceCollisions(rendered + rendered); len(c.DuplicateElementIDs) != 0 {
		// rendered+rendered fakes the same occurrence twice; real pages get
		// distinct tokens, so build the honest two-token page:
		t.Logf("same-token doubling collides as expected; checking distinct tokens")
	}
	toks := InstanceTokensForPage([]string{"my-tool", "my-tool"})
	var page strings.Builder
	for _, tok := range toks {
		rc := &RenderContext{}
		BindInstanceToken(rc, tok)
		out, _, _, err := RenderTemplate(tpl, rc, zap.NewNop())
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		page.WriteString(out)
	}
	if c := DetectInstanceCollisions(page.String()); len(c.DuplicateElementIDs) != 0 {
		t.Fatalf("two instances of the converted template collide: %v", c.DuplicateElementIDs)
	}
	// CONTROL (the check can fail): the unconverted original doubled IS dirty.
	if c := DetectInstanceCollisions(birthClassA + birthClassA); len(c.DuplicateElementIDs) == 0 {
		t.Fatalf("control failed: doubling the unconverted template should collide — the detector is inert")
	}
}

func TestScopeToolBirth_unarmedIsRecordOnly(t *testing.T) {
	tpl, rendered, info, refuse := ScopeToolBirthTemplate(birthClassA, "my-tool", false, zap.NewNop())
	if refuse != nil {
		t.Fatalf("unarmed mode must never refuse: %v", refuse)
	}
	if tpl != birthClassA || rendered != birthClassA {
		t.Fatalf("unarmed mode must return the caller's bytes VERBATIM — record-only means no new authority")
	}
	if s, _ := info["instance_scope"].(string); !strings.Contains(s, "record-only") {
		t.Fatalf("unarmed info must say record-only, got: %v", info["instance_scope"])
	}
}

func TestScopeToolBirth_armedRefusesGlobalScope(t *testing.T) {
	tpl, rendered, info, refuse := ScopeToolBirthTemplate(birthGlobalScope, "my-tool", true, zap.NewNop())
	if refuse == nil {
		t.Fatalf("a global-scope script must be REFUSED at birth when armed; info=%v", info)
	}
	if tpl != "" || rendered != "" {
		t.Fatalf("a refusal must return no bytes to persist")
	}
	if !strings.Contains(refuse.Error(), "regenerate") {
		t.Fatalf("the refusal must tell the retry loop what to do, got: %v", refuse)
	}
	// Unarmed: same template passes through untouched, defect recorded.
	tpl2, _, info2, refuse2 := ScopeToolBirthTemplate(birthGlobalScope, "my-tool", false, zap.NewNop())
	if refuse2 != nil || tpl2 != birthGlobalScope {
		t.Fatalf("unarmed mode must pass through: refuse=%v", refuse2)
	}
	if info2["instance_scope_defect"] == nil {
		t.Fatalf("unarmed mode must still RECORD the defect")
	}
}

func TestScopeToolBirth_selfConvertedIsVerifiedNotTrusted(t *testing.T) {
	tpl, rendered, _, refuse := ScopeToolBirthTemplate(birthSelfConverted, "my-tool", true, zap.NewNop())
	if refuse != nil {
		t.Fatalf("clean self-converted template must pass: %v", refuse)
	}
	if tpl != birthSelfConverted {
		t.Fatalf("a clean self-converted template is kept as-is")
	}
	if strings.Contains(rendered, "{{.InstanceID}}") || !strings.Contains(rendered, "c-my-tool-") {
		t.Fatalf("rendered bytes must be bound: %s", rendered[:160])
	}

	// The broken variant: placeholder ids but a global-scope script. Trusting
	// the placeholder's presence would ship it; the gate must refuse.
	broken := strings.Replace(birthSelfConverted,
		"(function () { 'use strict';", "", 1)
	broken = strings.Replace(broken, "})();</script>", ";</script>", 1)
	if !strings.Contains(broken, "{{.InstanceID}}") {
		t.Fatalf("fixture defect: broken variant lost its placeholders")
	}
	_, _, _, refuseBroken := ScopeToolBirthTemplate(broken, "my-tool", true, zap.NewNop())
	if refuseBroken == nil {
		t.Fatalf("a self-converted template with a global-scope script must be refused — presence of the placeholder is a claim, not a verification")
	}
}

func TestScopeToolBirth_noIDsPassesThrough(t *testing.T) {
	for _, armed := range []bool{true, false} {
		tpl, rendered, _, refuse := ScopeToolBirthTemplate(birthNoIDs, "my-tool", armed, zap.NewNop())
		if refuse != nil || tpl != birthNoIDs || rendered != birthNoIDs {
			t.Fatalf("armed=%v: an id-less template is inert-safe and must pass through verbatim (refuse=%v)", armed, refuse)
		}
	}
}

// ---------------------------------------------------------------------------
// Pass 0 AT THE CALL SITE (RFC_032; council round 1 on that change, bug_historian
// edit 4 and guardian edit 1, both MEDIUM, both making the same point against my
// own submission's prose).
//
// Pass 0 lives in ConvertTemplateToInstanceScope, which this file calls, so both
// ARMED birth guards — create_tool_component and deploy_tool_to_site, live since
// v1.0.1322 — inherited a behaviour change that only the converter's own tests
// exercised. The seats' conditional ("if any caller treated the old refusal as
// an expected outcome, that path is now silently bypassed") is not hypothetical:
// the `declares no literal element ids` arm above IS such a caller, and it
// persists the caller's bytes verbatim in both modes. These tests pin what that
// arm now does and no longer does.
// ---------------------------------------------------------------------------

// A newborn spelling the RETIRED placeholder. Before pass 0 the converter
// refused it ("declares no literal element ids") and the arm above passed it
// through UNCHANGED, armed or not — minting, at the one seam that exists to stop
// that, a template whose id is identical on every instance.
const birthTemplatedID = `<section id="{{.ComponentID}}" class="tool">
  <p>Nothing interactive; the wrapper is the whole component.</p>
</section>`

// No literal ids AND the placeholder somewhere pass 0 cannot rewrite. This one
// reaches the SAME early return as a genuinely id-less template, which is why
// the reason string is load-bearing: routed by text alone it takes the
// inert-safe arm and ships.
const birthComponentIDInDataAttr = `<section data-target="{{.ComponentID}}" class="tool">
  <p>The placeholder is a data attribute value, not an id.</p>
</section>`

func TestScopeToolBirth_armedConvertsTemplatedID(t *testing.T) {
	tpl, rendered, info, refuse := ScopeToolBirthTemplate(birthTemplatedID, "my-tool", true, zap.NewNop())
	if refuse != nil {
		t.Fatalf("a templated-id newborn must convert at birth, got refusal: %v", refuse)
	}
	if !strings.Contains(tpl, `id="{{.InstanceID}}"`) {
		t.Fatalf("armed mode must persist the CONVERTED template; got: %s", tpl)
	}
	if strings.Contains(tpl, "ComponentID") {
		t.Fatalf("the retired placeholder was persisted at birth: %s", tpl)
	}
	// The rendered bytes are what a live page serves (tools carry the template
	// verbatim as rendered_html), so an unbound or empty token here ships to a
	// reader. Assert the token POSITIVELY: `id=""` also contains no placeholder.
	// Occurrence 0's token is `c-my-tool` with no numeric suffix — InstanceToken
	// only appends one from the second instance on (`c-my-tool-2`), so asserting
	// a "-0" suffix would fail against correct output.
	if !strings.Contains(rendered, `id="c-my-tool"`) {
		t.Fatalf("rendered bytes must carry the bound occurrence-0 token; got: %s", rendered)
	}
	if n, _ := info["instance_scope_templated_id_swaps"].(int); n != 1 {
		t.Errorf("info must report the swap so the generator's spelling is visible; got %v", info["instance_scope_templated_id_swaps"])
	}
	// Unarmed on the same bytes: record-only means no new authority.
	tpl2, rendered2, _, refuse2 := ScopeToolBirthTemplate(birthTemplatedID, "my-tool", false, zap.NewNop())
	if refuse2 != nil || tpl2 != birthTemplatedID || rendered2 != birthTemplatedID {
		t.Fatalf("unarmed mode must return the caller's bytes verbatim (refuse=%v)", refuse2)
	}
}

func TestScopeToolBirth_armedRefusesComponentIDOutsideIDAttr(t *testing.T) {
	tpl, rendered, info, refuse := ScopeToolBirthTemplate(birthComponentIDInDataAttr, "my-tool", true, zap.NewNop())
	if refuse == nil {
		t.Fatalf("a newborn carrying the retired placeholder outside an id attribute must be REFUSED when armed — it renders the same value on every instance; info=%v", info)
	}
	if tpl != "" || rendered != "" {
		t.Fatalf("a refusal must return no bytes to persist")
	}
	if !strings.Contains(refuse.Error(), "ComponentID") {
		t.Errorf("the refusal must name what is wrong so the retry loop can act: %v", refuse)
	}
	// Unarmed: passes through, but the defect is RECORDED — the record-only
	// mode's whole purpose. Before the reason split it recorded nothing,
	// because the inert-safe arm reported "nothing to scope".
	tpl2, _, info2, refuse2 := ScopeToolBirthTemplate(birthComponentIDInDataAttr, "my-tool", false, zap.NewNop())
	if refuse2 != nil || tpl2 != birthComponentIDInDataAttr {
		t.Fatalf("unarmed mode must pass through: refuse=%v", refuse2)
	}
	if info2["instance_scope_defect"] == nil {
		t.Fatalf("unarmed mode must RECORD the defect, got info=%v", info2)
	}
	// CONTROL, and the reason this test is not just the converter's test again:
	// a template with no ids and no placeholder must STILL take the inert arm.
	// If this fails, the narrowing above swallowed the case it was meant to
	// leave alone.
	tpl3, _, info3, refuse3 := ScopeToolBirthTemplate(birthNoIDs, "my-tool", true, zap.NewNop())
	if refuse3 != nil || tpl3 != birthNoIDs {
		t.Fatalf("control failed: a genuinely id-less template must still pass through armed (refuse=%v, info=%v)", refuse3, info3)
	}
}
