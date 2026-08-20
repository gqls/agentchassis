// FILE: platform/orchestration/datahelpers/action_inputs_optional_explicit_test.go
//
// RFC_029 §9 D2's explicit-mapping vehicle — the opt-in `?` OPTIONAL-EXPLICIT
// marker on ExtractActionInputs' config surface: `"field?": "<reference>"` means
// the field resolves via that explicit reference OR IS ABSENT. Like `!` it never
// meets the whole-tree search, the nested-object fallback or the deprecated
// bridge; unlike `!`, an unresolved reference is not an error — a spec Default
// stands, an Optional field is simply absent, and a Required field fails the
// ordinary required-field validation. This is ResolveInputMapping's `?` (WDS-014:
// "a path that resolves is forwarded; a missing one is skipped") arriving on the
// step-config surface, completing the two-surface mirror CTS-060 describes.
//
// Why absence must be reachable: bugs_open/330 — a tool whose spec carried no
// related_pages had the whole-tree search substitute ANOTHER suggestion's pages,
// and nine tools cross-linked to one unrelated tool's targets. The correct
// behaviour there was "absent" (a handled, documented state); no marker could
// express it (`!` hard-fails, and related_pages is legitimately absent in most
// specs). `?` is that missing middle.
//
// MUTATION PROOFS (run by hand when editing the marker, per the lane practice —
// mutate to a NO-OP so the package still compiles, never to a build break):
//   - in withoutStrict, replace `if !explicitOnly(f)` with `if true` →
//     TestOptionalExplicitMarkerYieldsAbsenceInsteadOfSearching fails (the decoy
//     resolves), and the strict decoy test fails with it.
//   - in the Strategy 3 loop, replace `if explicitOnly(newField)` with
//     `if false` → TestOptionalExplicitFieldTakesNothingFromTheDeprecatedBridge
//     fails.
//   - in the nested-object block, replace `if explicitOnly(field)` with
//     `if false` → TestOptionalExplicitFieldSkipsNestedObjectFallback fails.
package datahelpers

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// The happy path: a `?` dotted reference resolves exactly like an unmarked one —
// the marker costs nothing where the mapping is sound.
func TestOptionalExplicitMarkerResolvesItsExplicitReference(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"related_pages"}}
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{
			"spec": map[string]interface{}{"related_pages": []interface{}{"page-a"}},
		},
		"decoy": map[string]interface{}{"related_pages": []interface{}{"page-WRONG"}},
	}
	config := map[string]interface{}{"related_pages?": "input_data.spec.related_pages"}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	got, ok := inputs.GetRaw("related_pages").([]interface{})
	if !ok || len(got) != 1 || got[0] != "page-a" {
		t.Fatalf("related_pages = %#v, want [page-a]: the `?` reference must resolve exactly "+
			"as the unmarked one does", inputs.GetRaw("related_pages"))
	}
}

// THE MARKER'S WHOLE MEANING: when the reference cannot resolve, the field is
// ABSENT — no error, and never the same-named decoy the search would find (the
// control proves it would). This is bugs_open/330's fix shape: for the field
// that opts in, "I tried the declared path and found nothing" resolves to
// nothing, instead of being handed to the search as though it were never asked.
func TestOptionalExplicitMarkerYieldsAbsenceInsteadOfSearching(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"related_pages"}}
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{
			"spec": map[string]interface{}{"name": "no related_pages here"},
		},
		"suggestions_echo": map[string]interface{}{
			"related_pages": []interface{}{"page-WRONG"},
		},
	}

	inputs, err := ExtractActionInputs(collected,
		map[string]interface{}{"related_pages?": "input_data.spec.related_pages"},
		spec, zap.NewNop())
	if err != nil {
		t.Fatalf("`?` field with an unresolvable reference must not error, got: %v", err)
	}
	if raw := inputs.GetRaw("related_pages"); raw != nil {
		t.Fatalf("related_pages = %#v, want ABSENT: `?` means this reference or nothing — "+
			"the search must never substitute the decoy", raw)
	}

	// THE CONTROL, and it is not optional: the identical fixture WITHOUT the
	// marker must resolve the decoy (unique candidate → the search takes it).
	// If this stops passing, the fixture no longer exercises the search and the
	// absence above proves nothing about the marker suppressing it.
	control, err := ExtractActionInputs(collected,
		map[string]interface{}{"related_pages": "input_data.spec.related_pages"},
		spec, zap.NewNop())
	if err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	got, ok := control.GetRaw("related_pages").([]interface{})
	if !ok || len(got) != 1 || got[0] != "page-WRONG" {
		t.Fatalf("CONTROL FAILED: unmarked related_pages = %#v, want the decoy (the search "+
			"should reach it) — rebuild the fixture so the search CAN win before trusting "+
			"the absence assertion above", control.GetRaw("related_pages"))
	}
}

// A spec Default STANDS for a `?` field whose reference missed — this is
// exactly where `?` and `!` part company (`!` treats a standing Default as
// unresolved and fails; TestStrictMarkerWithDefaultStillFailsWhenUnresolved
// pins that side). A Default is declared behaviour, not a search guess.
func TestOptionalExplicitMarkerLetsASpecDefaultStand(t *testing.T) {
	spec := ActionInputSpec{
		Optional: []string{"category"},
		Defaults: map[string]interface{}{"category": "interactive"},
	}
	collected := map[string]interface{}{
		"foreign": map[string]interface{}{"category": "WRONG"},
	}
	config := map[string]interface{}{"category?": "input_data.spec.category"}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.Get("category"); got != "interactive" {
		t.Fatalf("category = %q, want the spec default \"interactive\": a `?` miss falls back "+
			"to the DECLARED default, never to the search's foreign value", got)
	}
}

// A `?` SINGLE-SEGMENT reference resolves via the explicit collected_data key
// (Strategy 4's explicit arm), and when that key is absent the field is absent —
// never the foreign same-named key the search can see.
func TestOptionalExplicitMarkerResolvesSingleSegmentReference(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"spec_data"}}
	collected := map[string]interface{}{
		"site_plan": map[string]interface{}{"pages": 3},
		"foreign":   map[string]interface{}{"spec_data": "not-this"},
	}
	config := map[string]interface{}{"spec_data?": "site_plan"}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.GetMap("spec_data"); got == nil || got["pages"] != 3 {
		t.Fatalf("spec_data = %#v, want the resolved site_plan object: a `?` dotless string "+
			"is a single-segment REFERENCE, resolved explicitly, never a search result",
			inputs.GetRaw("spec_data"))
	}

	delete(collected, "site_plan")
	inputs, err = ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs after delete: %v", err)
	}
	if raw := inputs.GetRaw("spec_data"); raw != nil {
		t.Fatalf("spec_data = %#v, want ABSENT: the single-segment reference is gone and `?` "+
			"must not fall back to the foreign spec_data the search can see", raw)
	}
}

// A spec-REQUIRED field wired `?` whose reference missed fails the ORDINARY
// required-field validation — the author's contradiction ("required, but may be
// absent") surfaces as the standard missing-field error naming the field, not
// as a strict-marker error.
func TestOptionalExplicitRequiredFieldStillFailsValidation(t *testing.T) {
	spec := ActionInputSpec{Required: []string{"site_id"}}
	collected := map[string]interface{}{}
	config := map[string]interface{}{"site_id?": "input_data.site_id"}

	_, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err == nil {
		t.Fatal("required field absent after a `?` miss must still fail required validation")
	}
	if !strings.Contains(err.Error(), "missing required fields") ||
		!strings.Contains(err.Error(), "site_id") {
		t.Errorf("error %q must be the ordinary missing-required error naming the field", err.Error())
	}
	if strings.Contains(err.Error(), "'!'") {
		t.Errorf("error %q must not be the strict error — `?` is not `!`", err.Error())
	}
}

// When one field carries BOTH markers, `!` wins: the author who wrote the
// strict marker demanded resolution, so an unresolvable reference is the loud
// strict failure, not `?`'s quiet absence. (Same rule input_mapping.go applies
// to its degenerate combo.)
func TestStrictWinsWhenBothMarkersNameOneField(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"asset_id"}}
	collected := map[string]interface{}{}
	config := map[string]interface{}{
		"asset_id!": "stored.asset_id",
		"asset_id?": "stored.asset_id",
	}

	_, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err == nil {
		t.Fatal("both markers on one field with an unresolvable reference: `!` must win and " +
			"fail loudly, not defer to `?`'s absence")
	}
	if !strings.Contains(err.Error(), "'!'") {
		t.Errorf("error %q must be the strict error — `!` outranks `?`", err.Error())
	}
}

// A `?` field takes nothing from the deprecated-alias bridge (Strategy 3): its
// own mapping is the only reference allowed to decide it, exactly as for `!`.
func TestOptionalExplicitFieldTakesNothingFromTheDeprecatedBridge(t *testing.T) {
	spec := ActionInputSpec{
		Optional:   []string{"payload"},
		Deprecated: map[string]string{"payload_field": "payload"},
	}
	collected := map[string]interface{}{
		"legacy": map[string]interface{}{"payload": "via-the-bridge"},
	}
	config := map[string]interface{}{
		"payload?":      "input_data.spec.payload", // does not resolve
		"payload_field": "legacy.payload",          // the bridge COULD resolve this
	}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if raw := inputs.GetRaw("payload"); raw != nil {
		t.Fatalf("payload = %#v, want ABSENT: the deprecated bridge must not stand in for a "+
			"`?` reference that missed", raw)
	}

	// Control: without the marker the bridge fills the field — proving the
	// fixture exercises the bridge at all.
	control, err := ExtractActionInputs(collected, map[string]interface{}{
		"payload_field": "legacy.payload",
	}, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	if got := control.Get("payload"); got != "via-the-bridge" {
		t.Fatalf("CONTROL FAILED: unmarked payload = %q, want \"via-the-bridge\" — the bridge "+
			"no longer reaches this fixture, so the absence above proves nothing", got)
	}
}

// A `?` field takes nothing from the nested-object fallback (the current_page /
// site_record / input_data parent hunt) — implicit parent-object hunting is
// exactly what both markers forbid.
func TestOptionalExplicitFieldSkipsNestedObjectFallback(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"current_page", "description"}}
	collected := map[string]interface{}{
		"current_page": map[string]interface{}{"description": "the PAGE's description"},
	}
	config := map[string]interface{}{"description?": "input_data.spec.description"}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if raw := inputs.GetRaw("description"); raw != nil {
		t.Fatalf("description = %#v, want ABSENT: the nested-object fallback must not unwrap "+
			"current_page into a `?` field whose reference missed", raw)
	}

	// Control: without the marker the nested fallback fills description from
	// the current_page object — proving the fixture reaches that arm.
	control, err := ExtractActionInputs(collected, map[string]interface{}{}, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	if got := control.Get("description"); got != "the PAGE's description" {
		t.Fatalf("CONTROL FAILED: unmarked description = %q, want the current_page child — "+
			"the nested fallback no longer reaches this fixture, so the absence above "+
			"proves nothing", got)
	}
}

// The `?` spelling is recognised by unknown-config-key detection exactly when
// the base field is; an unrecognised base stays reported, in the spelling the
// author wrote.
func TestUnknownConfigKeysRecognisesTheOptionalExplicitSpelling(t *testing.T) {
	RegisterActionInputSpec("optional_explicit_marker_test_action", ActionInputSpec{
		Required:    []string{"asset_id"},
		CheckConfig: true,
	})

	unknown, checked := UnknownConfigKeys("optional_explicit_marker_test_action", map[string]interface{}{
		"asset_id?": "stored.asset_id",
		"mystery?":  "somewhere.else",
	})
	if !checked {
		t.Fatal("action opted in via CheckConfig but checked=false")
	}
	if len(unknown) != 1 || unknown[0] != "mystery?" {
		t.Fatalf("unknown = %v, want [mystery?]: asset_id? is the optional-explicit spelling of "+
			"a declared field (recognised); mystery? is not declared and must be reported as "+
			"written", unknown)
	}
}
