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
		out, err := RenderTemplate(tpl, rc, zap.NewNop())
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
