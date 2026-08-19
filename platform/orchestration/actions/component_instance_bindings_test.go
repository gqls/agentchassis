// FILE: platform/orchestration/actions/component_instance_bindings_test.go
//
// Fixtures are LIVE ROW BYTES exported 2026-08-19 — the CONVERTED state the
// mechanical batch actually wrote (content_components.html_template) and, for
// one component, the PRE-conversion snapshot from component_versions. A
// composed fixture would exercise the author's belief about the corpus; the
// author's belief is exactly what was wrong (bugs_open/283 §14).
//
// Each class is asserted in both directions: the shipped bytes must DETECT
// dirty, the repaired bytes must DETECT clean, and re-introducing one bare
// literal into the repaired bytes must detect dirty again — so a detector that
// reports nothing is distinguishable from one that is inert.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// Class A, array form, the SHIPPED bytes: `var ids = ['amount', …]` then
// `getElementById(id)` — tool-loan-repayment. The detector must see it and
// the repair must clear it without touching anything else.
func TestBindings_classA_array_loanRepayment(t *testing.T) {
	shipped := readFixture(t, "instance_bindings_converted_loan_repayment_283.html")

	ub := UnprefixedBindings(shipped)
	if len(ub) != 3 {
		t.Fatalf("CONTROL FAILED: the shipped loan-repayment template has 3 dangling literals (amount, interest, years); detector reports %d: %v", len(ub), ub)
	}

	repaired, rep, rejects, ok := RepairConvertedTemplateBindings(shipped)
	if !ok {
		t.Fatalf("repair refused: %v", rejects)
	}
	if rep.LiteralIDsRenamed != 3 || rep.ConcatSitesRenamed != 0 {
		t.Fatalf("repair counts: literals=%d concat=%d (want 3/0)", rep.LiteralIDsRenamed, rep.ConcatSitesRenamed)
	}
	if !strings.Contains(repaired, `var ids = ['{{.InstanceID}}-amount', '{{.InstanceID}}-interest', '{{.InstanceID}}-years'];`) {
		t.Fatal("the ids array must carry the prefix on every element")
	}
	if ub := UnprefixedBindings(repaired); len(ub) != 0 {
		t.Fatalf("repaired template must detect clean, got: %v", ub)
	}
	// Nothing outside the script bodies may move.
	if normaliseMarkup(markupOutsideScripts(shipped)) != normaliseMarkup(markupOutsideScripts(repaired)) {
		t.Fatal("repair touched markup outside <script> — it must only rewrite script bodies")
	}
	// Idempotent.
	again, rep2, _, _ := RepairConvertedTemplateBindings(repaired)
	if again != repaired || rep2.LiteralIDsRenamed != 0 {
		t.Fatal("repair must be idempotent over its own output")
	}
	// The gate must now pass this component (it is IIFE-scoped; only the
	// bindings were wrong).
	if needsJudged, err := GateConvertedTemplate("tool-loan-repayment", repaired, zap.NewNop()); err != nil || needsJudged {
		t.Fatalf("repaired loan-repayment must gate clean: needsJudged=%v err=%v", needsJudged, err)
	}
	// And the gate must REFUSE the shipped bytes — this is the check that was
	// missing on 2026-08-18.
	if needsJudged, err := GateConvertedTemplate("tool-loan-repayment", shipped, zap.NewNop()); err != nil || !needsJudged {
		t.Fatalf("CONTROL FAILED: the gate passed the shipped (dangling) bytes: needsJudged=%v err=%v", needsJudged, err)
	}

	// Mutation: put ONE bare literal back into the repaired template.
	mutated := strings.Replace(repaired, `'{{.InstanceID}}-interest'`, `'interest'`, 1)
	if mutated == repaired {
		t.Fatal("mutation did not apply")
	}
	if ub := UnprefixedBindings(mutated); len(ub) != 1 || !strings.Contains(ub[0], `"interest"`) {
		t.Fatalf("detector must see the single re-introduced literal, got: %v", ub)
	}
}

// Class A, config-object form, the SHIPPED bytes that were SERVING on
// robot-hands.com: `{ id: 'gsfc-accel', … }` then `getElementById(field.id)`.
func TestBindings_classA_configObjects_gripperSafetyFactor(t *testing.T) {
	shipped := readFixture(t, "instance_bindings_converted_gripper_sf_283.html")
	if ub := UnprefixedBindings(shipped); len(ub) != 5 {
		t.Fatalf("CONTROL FAILED: 5 field ids dangle on the served gripper page; detector reports %d: %v", len(ub), ub)
	}
	repaired, rep, _, ok := RepairConvertedTemplateBindings(shipped)
	if !ok || rep.LiteralIDsRenamed != 5 {
		t.Fatalf("repair: ok=%v literals=%d (want 5)", ok, rep.LiteralIDsRenamed)
	}
	if !strings.Contains(repaired, `{ id: '{{.InstanceID}}-gsfc-accel'`) {
		t.Fatal("object VALUE position must be renamed (it is the binding)")
	}
	if ub := UnprefixedBindings(repaired); len(ub) != 0 {
		t.Fatalf("repaired gripper must detect clean, got: %v", ub)
	}
	if needsJudged, err := GateConvertedTemplate("tool-gripper-safety-factor-calculator", repaired, zap.NewNop()); err != nil || needsJudged {
		t.Fatalf("repaired gripper must gate clean: %v %v", needsJudged, err)
	}
}

// Class C, the SHIPPED bytes: `getElementById('block-' + name)` for declared
// block-<name> ids — tool-affordability-complaint-checker, serving on three
// domains.
func TestBindings_classC_concatLookup_affordability(t *testing.T) {
	shipped := readFixture(t, "instance_bindings_converted_affordability_283.html")
	ub := UnprefixedBindings(shipped)
	if len(ub) == 0 || !strings.Contains(strings.Join(ub, " "), `"block-"`) {
		t.Fatalf("CONTROL FAILED: the 'block-' concatenated lookups must be detected, got: %v", ub)
	}
	repaired, rep, _, ok := RepairConvertedTemplateBindings(shipped)
	if !ok || rep.ConcatSitesRenamed == 0 {
		t.Fatalf("repair: ok=%v concat=%d", ok, rep.ConcatSitesRenamed)
	}
	if !strings.Contains(repaired, `getElementById('{{.InstanceID}}-block-' + `) {
		t.Fatal("the concatenated lookup must carry the prefix on its static half")
	}
	if ub := UnprefixedBindings(repaired); len(ub) != 0 {
		t.Fatalf("repaired affordability must detect clean, got: %v", ub)
	}
	if needsJudged, err := GateConvertedTemplate("tool-affordability-complaint-checker", repaired, zap.NewNop()); err != nil || needsJudged {
		t.Fatalf("repaired affordability must gate clean: %v %v", needsJudged, err)
	}
}

// Class B + C, the SHIPPED bytes: markup built in JS with `id="name-' + index
// + '"`, looked up by `getElementById('name-' + index)`, labelled by
// `for="name-' + index + '"` — tool-supplier-comparison-calculator, serving on
// gaswholesalers.com. Pass 1 had prefixed the declaration (its regex captured
// the fragment as an "id"); the lookup and the label were left behind.
func TestBindings_classB_dynamicDeclaration_supplierComparison(t *testing.T) {
	shipped := readFixture(t, "instance_bindings_converted_supplier_comparison_283.html")
	ids, fragments, rejects := BindingIDSets(shipped)
	if len(rejects) != 0 {
		t.Fatalf("no dynamic declaration here lacks a static prefix, got rejects %v", rejects)
	}
	if len(fragments) == 0 || !contains(fragments, "name-") {
		t.Fatalf("fragments must include 'name-' (dynamic declaration), got %v (ids %d)", fragments, len(ids))
	}
	if ub := UnprefixedBindings(shipped); len(ub) == 0 {
		t.Fatal("CONTROL FAILED: the shipped supplier-comparison template has unprefixed dynamic lookups")
	}
	repaired, rep, _, ok := RepairConvertedTemplateBindings(shipped)
	if !ok || rep.ConcatSitesRenamed == 0 {
		t.Fatalf("repair: ok=%v concat=%d", ok, rep.ConcatSitesRenamed)
	}
	for _, want := range []string{
		`getElementById('{{.InstanceID}}-name-' + index)`,
		`for="{{.InstanceID}}-name-' + index + '"`,
	} {
		if !strings.Contains(repaired, want) {
			t.Fatalf("repaired template must contain %q", want)
		}
	}
	if ub := UnprefixedBindings(repaired); len(ub) != 0 {
		t.Fatalf("repaired supplier-comparison must detect clean, got: %v", ub)
	}
	if needsJudged, err := GateConvertedTemplate("tool-supplier-comparison-calculator", repaired, zap.NewNop()); err != nil || needsJudged {
		t.Fatalf("repaired supplier-comparison must gate clean: %v %v", needsJudged, err)
	}
}

// The converter itself, on the PRE-conversion snapshot: with pass 5 in place
// a fresh conversion of tool-loan-repayment must come out with the array
// prefixed and the detector clean — the batch would not have shipped it.
func TestConvert_freshConversionCarriesBindings_loanRepayment(t *testing.T) {
	bare := readFixture(t, "instance_bindings_bare_loan_repayment_283.html")
	if strings.Contains(bare, "InstanceID") {
		t.Fatal("fixture must be the PRE-conversion snapshot")
	}
	converted, rep, ok := ConvertTemplateToInstanceScope(bare)
	if !ok {
		t.Fatalf("refused: %s", rep.RefusedReason)
	}
	if rep.Bindings.LiteralIDsRenamed != 3 {
		t.Fatalf("fresh conversion must rename the 3 array literals, got %d", rep.Bindings.LiteralIDsRenamed)
	}
	if len(rep.UnprefixedBindings) != 0 {
		t.Fatalf("fresh conversion must report no unprefixed bindings, got %v", rep.UnprefixedBindings)
	}
	if needsJudged, err := GateConvertedTemplate("tool-loan-repayment", converted, zap.NewNop()); err != nil || needsJudged {
		t.Fatalf("fresh conversion must gate clean: %v %v", needsJudged, err)
	}
	// Mutation on the CONVERTER: disable pass 5 by feeding the output of the
	// old passes — simulated by stripping the prefix from the array literals —
	// and the gate must refuse.
	old := strings.Replace(converted, `['{{.InstanceID}}-amount', '{{.InstanceID}}-interest', '{{.InstanceID}}-years']`, `['amount', 'interest', 'years']`, 1)
	if old == converted {
		t.Fatal("mutation did not apply")
	}
	if needsJudged, err := GateConvertedTemplate("tool-loan-repayment", old, zap.NewNop()); err != nil || !needsJudged {
		t.Fatalf("CONTROL FAILED: the gate must refuse the pre-pass-5 shape: %v %v", needsJudged, err)
	}
}

// Already-sound templates must stay untouched: the css-unit-converter fixture
// binds getElementById(targetId) off data-target attributes (renamed by pass
// 3b) and has no bare literals — the detector must be silent and the repair a
// no-op.
func TestBindings_soundTemplateIsUntouched(t *testing.T) {
	tpl := readFixture(t, "instance_conversion_css_unit_converter_283.html")
	converted, _, ok := ConvertTemplateToInstanceScope(tpl)
	if !ok {
		t.Fatal("fixture must convert")
	}
	if ub := UnprefixedBindings(converted); len(ub) != 0 {
		t.Fatalf("a sound template must detect clean, got %v", ub)
	}
	repaired, rep, _, _ := RepairConvertedTemplateBindings(converted)
	if repaired != converted || rep.LiteralIDsRenamed != 0 || rep.ConcatSitesRenamed != 0 {
		t.Fatal("repair must be a no-op on a sound template")
	}
}

// Refuse-contexts: a bare literal equal to an id in a comparison, a case
// label, an object KEY or a computed property access is NOT renamed — it is
// reported, so the gate routes the component to the judged pool where a
// reader can tell. Composed on purpose: no live row exercises every context.
func TestBindings_refuseContextsAreReportedNotRenamed(t *testing.T) {
	tpl := `<div id="rate"></div><div id="term"></div>
<script>
(function(){
  var ids = ['rate', 'term'];                 // array: rename
  if (mode === 'rate') { }                    // comparison: skip
  switch (k) { case 'term': break; }          // case label: skip
  var cfg = { 'rate': 1 };                    // object key: skip
  var v = data['term'];                       // computed access: skip
  ids.forEach(function(id){ document.getElementById(id); });
})();
</script>`
	converted, rep, ok := ConvertTemplateToInstanceScope(tpl)
	if !ok {
		t.Fatalf("refused: %s", rep.RefusedReason)
	}
	if rep.Bindings.LiteralIDsRenamed != 2 {
		t.Fatalf("only the two array elements rename, got %d", rep.Bindings.LiteralIDsRenamed)
	}
	if !strings.Contains(converted, `['{{.InstanceID}}-rate', '{{.InstanceID}}-term']`) {
		t.Fatal("array elements must be renamed")
	}
	for _, keep := range []string{`mode === 'rate'`, `case 'term':`, `{ 'rate': 1 }`, `data['term']`} {
		if !strings.Contains(converted, keep) {
			t.Fatalf("%q must be left alone", keep)
		}
	}
	if len(rep.Bindings.SkippedContexts) != 4 {
		t.Fatalf("4 skipped contexts expected, got %v", rep.Bindings.SkippedContexts)
	}
	// And the skipped ones are REPORTED, so the gate refuses to judged.
	if len(rep.UnprefixedBindings) == 0 {
		t.Fatal("skipped literals must surface in UnprefixedBindings")
	}
	if needsJudged, err := GateConvertedTemplate("x", converted, zap.NewNop()); err != nil || !needsJudged {
		t.Fatalf("gate must route to the judged pool on reported bindings: %v %v", needsJudged, err)
	}
}

// A dynamic declaration with NO static prefix cannot be carried: refuse.
func TestBindings_dynamicDeclarationWithoutPrefixRefuses(t *testing.T) {
	tpl := `<div id="host"></div>
<script>(function(){ var h = document.getElementById('host'); h.innerHTML = '<input id="' + someId + '">'; })();</script>`
	_, rep, ok := ConvertTemplateToInstanceScope(tpl)
	if ok {
		t.Fatal("must refuse: the dynamic declaration has no static `-`/`_` prefix")
	}
	if !strings.Contains(rep.RefusedReason, "no static") {
		t.Fatalf("reason must name the cause, got %q", rep.RefusedReason)
	}
}

// The judged gate refuses a rewrite that left a binding unprefixed, even when
// ids and script scope are clean.
func TestJudged_refusesUnprefixedBindingInRewrite(t *testing.T) {
	bare := readFixture(t, "instance_judged_loans_standard_calc_283.html")
	baseline, rep, ok := ConvertTemplateToInstanceScope(bare)
	if !ok {
		t.Fatalf("refused: %s", rep.RefusedReason)
	}
	// The canary's `const inputs = ['amount','interest','years']` is class A —
	// pass 5 now prefixes it in the baseline.
	if !strings.Contains(baseline, `['{{.InstanceID}}-amount', '{{.InstanceID}}-interest', '{{.InstanceID}}-years']`) {
		t.Fatal("baseline must carry the prefixed inputs array")
	}
	// A correct rewrite: IIFE-wrapped, nothing else changed.
	good := strings.Replace(baseline, "<script>\n    const inputs", "<script>\n(function(){\n    const inputs", 1)
	good = strings.Replace(good, "    calculateLoan();\n</script>", "    calculateLoan();\n})();\n</script>", 1)
	if good == baseline {
		t.Fatal("rewrite fixture did not apply")
	}
	if issues := JudgedConversionIssues("loans-standard-calc", baseline, good, zap.NewNop()); len(issues) != 0 {
		t.Fatalf("a correct rewrite must pass, got %v", issues)
	}
	// Mutation: the LLM "tidies" the array back to bare ids.
	bad := strings.Replace(good, `['{{.InstanceID}}-amount', '{{.InstanceID}}-interest', '{{.InstanceID}}-years']`, `['amount', 'interest', 'years']`, 1)
	issues := JudgedConversionIssues("loans-standard-calc", baseline, bad, zap.NewNop())
	if len(issues) == 0 || !strings.Contains(strings.Join(issues, " "), "unprefixed bindings") {
		t.Fatalf("CONTROL FAILED: the judged gate must refuse the unprefixed array, got %v", issues)
	}
}
