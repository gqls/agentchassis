// FILE: platform/orchestration/datahelpers/action_inputs_strategy6_test.go
//
// bugs_open/231 candidate 2 (Strategy 6) — the properties the council's REVISE
// round 1 asked to have ENFORCED rather than proved in prose.
//
// The submission's whole blast-radius argument is one invariant: for a field
// carrying a spec Default, no strategy other than Strategy 0, the Strategy 3
// bridge, and Strategy 6 can ever set it, because every other arm opens with
// `if _, hasValue := result.Values[field]; hasValue { continue }` and nothing
// ever deletes from Values. The `guardian` seat's gating objection was that this
// is "a static snapshot, not an invariant enforced going forward — nothing stops
// a future spec from having both a Default and a resolution path this proof
// didn't anticipate". Correct, and that is what TestDefaultBeatsTheRecursiveSearch
// below is for: it fails the moment a future strategy lets anything other than an
// explicit config entry overwrite a Default.
package datahelpers

import (
	"testing"

	"go.uber.org/zap"
)

// strategy6Spec is a defaulted-field spec with a deprecated alias onto the same
// field, so bridge-vs-canonical precedence is testable in one place.
func strategy6Spec() ActionInputSpec {
	return ActionInputSpec{
		Required:   []string{},
		Optional:   []string{"purpose", "max_items", "label"},
		Defaults:   map[string]interface{}{"purpose": "hero", "max_items": 10},
		Deprecated: map[string]string{"purpose_field": "purpose"},
	}
}

// THE INVARIANT. Strategies 1/2 use ExtractFields' aggressive recursive search,
// which WOULD find `spec.purpose` here — the same search that, per Strategy 0's
// own comment, "can find stale values from previous loop iterations". A Default
// must beat it, because the caller said nothing: candidate 2 deliberately did NOT
// widen the resolver to let collected_data outrank a Default, only explicit config.
//
// If this test starts failing, a strategy has been added or reordered such that
// something OTHER than an explicit config entry can overwrite a Default. That
// breaks the blast-radius proof recorded in bugs_open/231 and CTS-059, and it
// silently re-opens bugs_open/248 finding (b): deploy_image_asset reads
// WasDefaulted before letting an asset row's purpose win, so a Default overwritten
// by a search would read as a caller's stated choice.
func TestDefaultBeatsTheRecursiveSearch(t *testing.T) {
	logger := zap.NewNop()
	// No config at all — nothing explicit is being said about purpose.
	collected := map[string]interface{}{
		"spec":          map[string]interface{}{"purpose": "logo", "max_items": 99},
		"current_page":  map[string]interface{}{"purpose": "banner"},
		"claim_result":  map[string]interface{}{"purpose": "stale-from-iteration-0"},
		"anything_else": map[string]interface{}{"nested": map[string]interface{}{"purpose": "deep"}},
	}

	inputs, err := ExtractActionInputs(collected, map[string]interface{}{}, strategy6Spec(), logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}

	if got := inputs.Get("purpose"); got != "hero" {
		t.Fatalf("purpose = %q, want \"hero\": something other than an explicit config entry "+
			"overwrote a spec Default. That invalidates bugs_open/231's unreachability proof — "+
			"re-measure the census before assuming the blast radius is still the dead set.", got)
	}
	if got := inputs.GetInt("max_items", -1); got != 10 {
		t.Fatalf("max_items = %d, want 10 (the Default) — the recursive search beat a Default", got)
	}
	if !inputs.WasDefaulted("purpose") {
		t.Fatal("purpose came from the Default but WasDefaulted is false — provenance is lying, " +
			"and bugs_open/248's row-purpose fallback depends on it")
	}

	// THE CONTROL, and it is not optional. Everything above would pass just as
	// happily if ExtractFields simply never found these values — a test asserting
	// "the search did not win" is vacuous unless the search COULD have won. So run
	// data of the same nesting shape through a spec with no Defaults and confirm
	// the search does reach it.
	//
	// REPAIRED 2026-08-15 (RFC_029 §9 implementation notes; filed failing by
	// 260cb2393). The original control ran the FULL fixture above and asserted one
	// winner ("logo") out of FOUR same-named candidates — but which candidate the
	// whole-tree search met first was Go map-iteration order, a per-run coin flip,
	// so the control failed on most runs while the assertions it guards were fine.
	// The control's only job is to prove the search CAN reach a Defaults-blocked
	// value, so it gets a fixture with exactly ONE candidate per field: correct
	// under the old random walk, under RFC_029 Phase 1 (stable winner), and under
	// Phase 2 (conflicts refuse to resolve), because one candidate is never a
	// conflict. The `spec.*` nesting is identical to the main fixture's, so
	// reachability proven here transfers to the assertions above.
	singleCandidate := map[string]interface{}{
		"spec": map[string]interface{}{"purpose": "logo", "max_items": 99},
	}
	noDefaults := ActionInputSpec{Optional: []string{"purpose", "max_items"}}
	control, err := ExtractActionInputs(singleCandidate, map[string]interface{}{}, noDefaults, logger)
	if err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	if control.Get("purpose") != "logo" || control.GetInt("max_items", -1) != 99 {
		t.Fatalf("CONTROL FAILED: with no Defaults the recursive search resolved purpose=%q "+
			"max_items=%v, wanted \"logo\"/99. The search can no longer reach this fixture, so the "+
			"assertions above prove nothing about Defaults beating it — rebuild the fixture so the "+
			"search CAN win before trusting this test again.",
			control.Get("purpose"), control.GetRaw("max_items"))
	}
}

// Strategy 6 must not fire for a field with NO Default. There the same dotless
// string is a single-segment REFERENCE (Strategy 4), and that asymmetry is
// CTS-059's registered landmine — pinned here so it cannot drift silently.
func TestStrategy6DoesNotTouchANonDefaultedField(t *testing.T) {
	logger := zap.NewNop()
	spec := ActionInputSpec{Optional: []string{"spec_data"}} // no Defaults at all
	collected := map[string]interface{}{
		"site_plan": map[string]interface{}{"pages": 3},
	}
	config := map[string]interface{}{"spec_data": "site_plan"}

	inputs, err := ExtractActionInputs(collected, config, spec, logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	// Strategy 4 resolved the reference. If Strategy 6 had claimed it, the value
	// would be the literal string "site_plan" instead of the resolved object.
	if got := inputs.GetMap("spec_data"); got == nil {
		t.Fatalf("spec_data = %#v, want the resolved site_plan object: a dotless string on a "+
			"NON-defaulted field must stay a reference (Strategy 4), not become a literal",
			inputs.GetRaw("spec_data"))
	}
}

// CANONICAL BEATS DEPRECATED. Raised by the editquality seat: Strategy 3 runs
// before Strategy 6, so on a step carrying both spellings the alias would have
// claimed the field and the canonical key would have lost — the opposite of what
// a migration is for, and inconsistent with the dotted case, where Strategy 0
// already beats the bridge.
func TestCanonicalKeyBeatsTheDeprecatedAlias(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"stored": map[string]interface{}{"purpose": "logo"},
	}
	config := map[string]interface{}{
		"purpose_field": "stored.purpose", // deprecated alias, resolves to "logo"
		"purpose":       "banner",         // canonical key, dotless literal
	}

	inputs, err := ExtractActionInputs(collected, config, strategy6Spec(), logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.Get("purpose"); got != "banner" {
		t.Fatalf("purpose = %q, want \"banner\" — the canonical config key must beat a "+
			"deprecated alias whatever order the strategies run in", got)
	}
	// The alias is still REPORTED even though it did not decide the value: it is
	// present and it is deprecated, and silence would remove the only signal that
	// tells the author to drop it.
	if len(inputs.DeprecatedUsed) != 1 || inputs.DeprecatedUsed[0] != "purpose_field" {
		t.Errorf("DeprecatedUsed = %v, want [purpose_field]: losing the precedence contest must "+
			"not also silence the deprecation warning", inputs.DeprecatedUsed)
	}
	if inputs.WasDefaulted("purpose") {
		t.Error("purpose was supplied but is still marked defaulted")
	}
}

// The bridge's behaviour for a NON-defaulted field must be exactly what it was
// before candidate 2 touched Strategy 3's skip. Raised by the guardian seat:
// the change alters bridge precedence fleet-wide for every `*_field` alias, and
// the two flipped tests only covered the defaulted case.
//
// REPAIRED 2026-08-15 (RFC_029 §9 implementation). The old fixture held BOTH
// site_id values in plain containers, so Strategy 2's whole-tree search reached
// two same-named candidates BEFORE the bridge ever ran — and "the bridge fired"
// was true only when the randomised map walk happened to meet site_record first,
// a coin flip this test lost intermittently on pristine HEAD. RFC_029's
// determinism froze that coin (shallowest-first, sorted keys: "explicit" beats
// "site_record"), turning the intermittent failure permanent. The bridge is the
// deciding arm ONLY when nothing earlier in the chain resolves the field, so the
// fixture now places both sources under keys on the search's documented
// infrastructure skip-list (isInfrastructureKey) — which the search never enters
// but ExtractNestedField's plain path walk (Strategy 0 and the bridge) does.
// The search was never this test's subject; now it cannot vote.
func TestBridgeIsUnchangedForANonDefaultedField(t *testing.T) {
	logger := zap.NewNop()
	spec := ActionInputSpec{
		Optional:   []string{"site_id"}, // deliberately NO Defaults
		Deprecated: map[string]string{"site_id_field": "site_id"},
	}
	collected := map[string]interface{}{
		"agent_config":          map[string]interface{}{"site_id": "from-the-bridge"},
		"__execution_context__": map[string]interface{}{"site_id": "from-strategy-0"},
	}

	t.Run("bridge supplies a field nothing else set", func(t *testing.T) {
		inputs, err := ExtractActionInputs(collected,
			map[string]interface{}{"site_id_field": "agent_config.site_id"}, spec, logger)
		if err != nil {
			t.Fatalf("ExtractActionInputs: %v", err)
		}
		if got := inputs.Get("site_id"); got != "from-the-bridge" {
			t.Errorf("site_id = %q, want \"from-the-bridge\": the bridge must still fire for a "+
				"non-defaulted field with no other source", got)
		}
	})

	t.Run("bridge yields to an explicit canonical path", func(t *testing.T) {
		inputs, err := ExtractActionInputs(collected, map[string]interface{}{
			"site_id":       "__execution_context__.site_id", // Strategy 0, resolves
			"site_id_field": "agent_config.site_id",
		}, spec, logger)
		if err != nil {
			t.Fatalf("ExtractActionInputs: %v", err)
		}
		if got := inputs.Get("site_id"); got != "from-strategy-0" {
			t.Errorf("site_id = %q, want \"from-strategy-0\": the has-value skip must still stop "+
				"the bridge overwriting a caller-supplied value", got)
		}
	})
}

// The kind guard, at the resolver rather than in the offline detector: a config
// scalar whose type differs from the Default's is refused and the Default stands.
func TestStrategy6KindGuardRefusesAWrongTypedLiteral(t *testing.T) {
	logger := zap.NewNop()
	config := map[string]interface{}{
		"max_items": "25", // string against an int Default
		"purpose":   3.0,  // number against a string Default
	}
	inputs, err := ExtractActionInputs(map[string]interface{}{}, config, strategy6Spec(), logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.GetInt("max_items", -1); got != 10 {
		t.Errorf("max_items = %d, want 10: a string literal must not override an int Default", got)
	}
	if got := inputs.Get("purpose"); got != "hero" {
		t.Errorf("purpose = %q, want \"hero\": a numeric literal must not override a string Default", got)
	}
	// Both must still read as defaulted — a refused override is not a supplied value.
	if !inputs.WasDefaulted("max_items") || !inputs.WasDefaulted("purpose") {
		t.Error("a refused override left the field marked as caller-supplied")
	}
}
