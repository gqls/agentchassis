// FILE: platform/orchestration/datahelpers/action_inputs_defaulted_test.go
//
// Pins the provenance half of bugs_open/248 finding (b).
//
// The defect these guard against is not a wrong value — it is a value that is
// RIGHT and unattributable. Defaults are written into Values before Strategy 1/2
// run, and every later strategy skips a field that already holds a value, so a
// field WITH a default can only ever be set by a Strategy 0 dot-path that
// resolves. A consumer reading Values sees "hero" and cannot tell whether a
// caller chose it. deploy_image_asset shipped every work-item-dispatched asset —
// logos included — under the hero geometry and extension for months because of
// exactly that, while every mechanical signal read green.
//
// So the property worth pinning is that the map DISCRIMINATES. A test that only
// checked "purpose == hero" would have passed throughout the defect.

package datahelpers

import (
	"testing"

	"go.uber.org/zap"
)

func defaultedSpec() ActionInputSpec {
	return ActionInputSpec{
		Optional: []string{"purpose", "asset_key"},
		Defaults: map[string]interface{}{"purpose": "hero"},
	}
}

// TestDefaultedMarksASpecDefault — nobody supplied purpose, so the value is the
// spec's and must say so.
func TestDefaultedMarksASpecDefault(t *testing.T) {
	in, err := ExtractActionInputs(
		map[string]interface{}{"input_data": map[string]interface{}{}},
		map[string]interface{}{"purpose": "input_data.purpose"},
		defaultedSpec(), zap.NewNop())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if got := in.Get("purpose"); got != "hero" {
		t.Fatalf("purpose = %q, want the default %q", got, "hero")
	}
	if !in.WasDefaulted("purpose") {
		t.Errorf("WasDefaulted(purpose) = false, want true.\n" +
			"Nothing in collected_data supplied a purpose, so the value is the spec's. Without this " +
			"distinction a consumer cannot refuse to act on a default, which is bugs_open/248 (b): " +
			"every asset deployed as a hero because 'hero' looked like an instruction.")
	}
}

// TestDefaultedIsClearedWhenACallerSuppliesTheValue is named in
// ExtractActionInputs' own comment: Strategy 0 is the only strategy that can
// overwrite a default, so it is the only place provenance is cleared. If a
// future strategy gains that power, this is where the invariant breaks.
func TestDefaultedIsClearedWhenACallerSuppliesTheValue(t *testing.T) {
	in, err := ExtractActionInputs(
		map[string]interface{}{
			"input_data": map[string]interface{}{"purpose": "logo"},
		},
		map[string]interface{}{"purpose": "input_data.purpose"},
		defaultedSpec(), zap.NewNop())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if got := in.Get("purpose"); got != "logo" {
		t.Fatalf("purpose = %q, want the caller's %q", got, "logo")
	}
	if in.WasDefaulted("purpose") {
		t.Errorf("WasDefaulted(purpose) = true after a caller supplied %q.\n"+
			"A stale 'defaulted' mark is worse than none: a consumer would override a value the "+
			"caller deliberately chose.", "logo")
	}
}

// TestWasDefaultedIsFalseForAFieldWithNoDefault — the no-op case, which is the
// one that goes unchecked. asset_key has no default, so it must never be marked
// however it was resolved. (This asymmetry IS the bug's shape: asset_key had no
// default and so could be found by the recursive search, while purpose could not.)
func TestWasDefaultedIsFalseForAFieldWithNoDefault(t *testing.T) {
	in, err := ExtractActionInputs(
		map[string]interface{}{
			"input_data": map[string]interface{}{"spec": map[string]interface{}{"asset_key": "logo"}},
		},
		map[string]interface{}{}, defaultedSpec(), zap.NewNop())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if in.WasDefaulted("asset_key") {
		t.Errorf("WasDefaulted(asset_key) = true, but the spec declares no default for it")
	}
	if in.WasDefaulted("nonexistent_field") {
		t.Errorf("WasDefaulted on an unknown field = true, want false (nil-safe zero value)")
	}
}

// TestWasDefaultedIsNilSafe — the action's error path constructs an ActionInputs
// literal with no Defaulted map (deploy_image_asset does exactly this when
// extraction fails and it falls back to legacy extraction). Reading a nil map
// must not panic, or a fallback path becomes a crash.
func TestWasDefaultedIsNilSafe(t *testing.T) {
	bare := &ActionInputs{Values: map[string]interface{}{}}
	if bare.WasDefaulted("purpose") {
		t.Errorf("WasDefaulted on a nil Defaulted map = true, want false")
	}
}
