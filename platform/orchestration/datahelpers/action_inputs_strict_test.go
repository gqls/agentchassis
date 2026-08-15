// FILE: platform/orchestration/datahelpers/action_inputs_strict_test.go
//
// RFC_029 §9 D3 (owner-delegated ruling, 2026-08-15) — the opt-in `!` STRICT
// marker on ExtractActionInputs' config surface: `"field!": "<reference>"` means
// explicit resolution only; if the reference does not resolve, the extraction
// fails loudly; the whole-tree search NEVER runs for the field. Opt-in with the
// unsafe (searching) default OFF, per the owner's 2026-08-02 §2 ruling.
//
// Also here: the observation-only WARN for the bugfix-213 lane's mapped-field
// class (a dotless config reference bypassed by the search), because the `!`
// marker is that class's opt-in remedy and the WARN is its measurement.
package datahelpers

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const bypassWarnMsg = "aggressive search: explicit single-segment mapping bypassed"

func observedExtract(t *testing.T, collected, config map[string]interface{}, spec ActionInputSpec) (*ActionInputs, error, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zapcore.WarnLevel)
	inputs, err := ExtractActionInputs(collected, config, spec, zap.New(core))
	return inputs, err, logs
}

// The happy path: a strict dotted reference resolves exactly like an unmarked
// one — the marker costs nothing where the mapping is sound.
func TestStrictMarkerResolvesItsExplicitReference(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"asset_id"}}
	collected := map[string]interface{}{
		"stored": map[string]interface{}{"asset_id": "a-real"},
		"decoy":  map[string]interface{}{"asset_id": "a-WRONG"},
	}
	config := map[string]interface{}{"asset_id!": "stored.asset_id"}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.Get("asset_id"); got != "a-real" {
		t.Fatalf("asset_id = %q, want \"a-real\": the strict reference must resolve exactly "+
			"as the unmarked one does", got)
	}
}

// THE MARKER'S WHOLE MEANING: when the strict reference cannot resolve, the
// extraction FAILS — it does not let the search find the same-named decoy that
// an unmarked field would happily take (the control below proves it would).
func TestStrictMarkerFailsLoudlyInsteadOfSearching(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"asset_id"}}
	collected := map[string]interface{}{
		"decoy": map[string]interface{}{"asset_id": "a-WRONG"},
	}

	_, err := ExtractActionInputs(collected,
		map[string]interface{}{"asset_id!": "stored.asset_id"}, spec, zap.NewNop())
	if err == nil {
		t.Fatal("strict field with an unresolvable reference returned nil error — the loud " +
			"failure is the marker's entire contract")
	}
	if !strings.Contains(err.Error(), "asset_id") || !strings.Contains(err.Error(), "'!'") {
		t.Errorf("strict error %q must name the field and the marker", err.Error())
	}

	// THE CONTROL, and it is not optional: the identical fixture WITHOUT the
	// marker must resolve the decoy (unique candidate → the search takes it).
	// If this stops passing, the fixture no longer exercises the search and the
	// assertion above proves nothing about the marker suppressing it.
	control, err := ExtractActionInputs(collected,
		map[string]interface{}{"asset_id": "stored.asset_id"}, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	if got := control.Get("asset_id"); got != "a-WRONG" {
		t.Fatalf("CONTROL FAILED: unmarked asset_id = %q, want \"a-WRONG\" (the search should "+
			"reach the decoy) — rebuild the fixture so the search CAN win before trusting "+
			"the strict assertion above", got)
	}
}

// A strict SINGLE-SEGMENT reference resolves via the explicit collected_data
// key — closing, for marked fields, the bugfix-213 class where a dotless
// mapping was silently outvoted by the search.
func TestStrictMarkerResolvesSingleSegmentReference(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"spec_data"}}
	collected := map[string]interface{}{
		"site_plan": map[string]interface{}{"pages": 3},
		"foreign":   map[string]interface{}{"spec_data": "not-this"},
	}
	config := map[string]interface{}{"spec_data!": "site_plan"}

	inputs, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.GetMap("spec_data"); got == nil || got["pages"] != 3 {
		t.Fatalf("spec_data = %#v, want the resolved site_plan object: a strict dotless string "+
			"is a single-segment REFERENCE, resolved explicitly, never a search result",
			inputs.GetRaw("spec_data"))
	}

	// And when the single-segment reference is absent, strict fails rather than
	// taking the foreign same-named key the search would find.
	delete(collected, "site_plan")
	if _, err := ExtractActionInputs(collected, config, spec, zap.NewNop()); err == nil {
		t.Fatal("strict single-segment reference with no matching collected_data key must fail, " +
			"not fall back to the foreign spec_data the search can see")
	}
}

// A spec Default is NOT an acceptable stand-in for a strict reference that
// failed: the author demanded the mapping resolve, and a Default silently
// standing in is the silent-substitution shape the marker refuses.
func TestStrictMarkerWithDefaultStillFailsWhenUnresolved(t *testing.T) {
	spec := ActionInputSpec{
		Optional: []string{"purpose"},
		Defaults: map[string]interface{}{"purpose": "hero"},
	}
	collected := map[string]interface{}{}
	config := map[string]interface{}{"purpose!": "asset_row.purpose"}

	_, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err == nil {
		t.Fatal("strict field fell back to its spec Default — a Default is a value, not a " +
			"resolution of the author's explicit reference")
	}
}

// The strict spelling is recognised by unknown-config-key detection exactly
// when the base field is; an unrecognised base stays reported, in the spelling
// the author wrote.
func TestUnknownConfigKeysRecognisesTheStrictSpelling(t *testing.T) {
	RegisterActionInputSpec("strict_marker_test_action", ActionInputSpec{
		Required:    []string{"asset_id"},
		CheckConfig: true,
	})

	unknown, checked := UnknownConfigKeys("strict_marker_test_action", map[string]interface{}{
		"asset_id!": "stored.asset_id",
		"mystery!":  "somewhere.else",
	})
	if !checked {
		t.Fatal("action opted in via CheckConfig but checked=false")
	}
	if len(unknown) != 1 || unknown[0] != "mystery!" {
		t.Fatalf("unknown = %v, want [mystery!]: asset_id! is the strict spelling of a declared "+
			"field (recognised); mystery! is not declared and must be reported as written", unknown)
	}
}

// OBSERVATION ONLY (RFC_029, the bugfix-213 contribution): an UNMARKED dotless
// reference that the search outvotes emits the bypass WARN — the measurement
// that decides whether the mapped-field class gets its own Phase 2 remedy.
// Behaviour itself must be unchanged in Phase 1: the searched value still wins.
func TestSingleSegmentMappingBypassWarns(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"payload"}}
	collected := map[string]interface{}{
		"handler_output": map[string]interface{}{"ok": true},
		"sibling":        map[string]interface{}{"payload": map[string]interface{}{"foreign": true}},
	}
	// The author mapped payload to the handler_output envelope — dotless, so
	// Strategy 0 skips it and the whole-tree search runs first.
	config := map[string]interface{}{"payload": "handler_output"}

	inputs, err, logs := observedExtract(t, collected, config, spec)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}

	got := inputs.GetMap("payload")
	if got == nil || got["foreign"] != true {
		t.Fatalf("payload = %#v, want the sibling's foreign value: Phase 1 must observe the "+
			"bypass, not change it — if this moved, the observation window's baseline is broken",
			inputs.GetRaw("payload"))
	}
	if n := logs.FilterMessage(bypassWarnMsg).Len(); n != 1 {
		t.Fatalf("bypass WARN fired %d times, want exactly 1: the 213 class is invisible "+
			"without it", n)
	}

	// Control: when the search's answer AGREES with the mapping (it found the
	// mapped envelope itself), there is no bypass and no WARN.
	agreeing := map[string]interface{}{
		"nested": map[string]interface{}{"payload": map[string]interface{}{"same": true}},
		"stash":  map[string]interface{}{"payload": map[string]interface{}{"same": true}},
	}
	_, err, logs = observedExtract(t, agreeing,
		map[string]interface{}{"payload": "missing_key"}, spec)
	if err != nil {
		t.Fatalf("control ExtractActionInputs: %v", err)
	}
	if n := logs.FilterMessage(bypassWarnMsg).Len(); n != 0 {
		t.Errorf("bypass WARN fired %d times when the mapped key does not exist — the WARN is "+
			"for a mapping the search OUTVOTED, not for one that could never resolve", n)
	}
}
