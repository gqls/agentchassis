package datahelpers

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

// Regression tests for bugs_open/042.
//
// Every config-reading branch of ExtractActionInputs asserts config[field].(string)
// and treats the result as a REFERENCE to resolve against collectedData. Numeric and
// boolean config values therefore never reached the action at all: they failed the
// type assertion and the call site's Go fallback won, silently, while the config read
// as though it were live. Strategy 5 takes non-string scalars as literals.
//
// The defect hid because a seeded value happened to equal the code default (72 == 72),
// so config and behaviour agreed. These tests deliberately use a configured value that
// DIFFERS from the fallback, because a test that reuses the default proves nothing.

func TestExtractActionInputs_NumericConfigLiteralReachesAction(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"max_age_hours", "max_items"}}
	config := map[string]interface{}{
		"max_age_hours": float64(720), // as JSON unmarshals it
		"max_items":     float64(30),
	}

	got, err := ExtractActionInputs(map[string]interface{}{}, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fallbacks deliberately differ from the configured values.
	if v := got.GetInt("max_age_hours", 72); v != 720 {
		t.Errorf("max_age_hours: got %d, want 720 (fallback 72 means config was dropped)", v)
	}
	if v := got.GetInt("max_items", 6); v != 30 {
		t.Errorf("max_items: got %d, want 30 (fallback 6 means config was dropped)", v)
	}
}

func TestExtractActionInputs_IntAndJSONNumberAndBoolLiterals(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"plain_int", "json_num", "flag"}}
	config := map[string]interface{}{
		"plain_int": 15,
		"json_num":  json.Number("21"),
		"flag":      true,
	}

	got, err := ExtractActionInputs(map[string]interface{}{}, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := got.GetInt("plain_int", 1); v != 15 {
		t.Errorf("plain_int: got %d, want 15", v)
	}
	if !got.Has("json_num") {
		t.Error("json.Number literal should be carried through")
	}
	if v := got.GetBool("flag", false); v != true {
		t.Errorf("flag: got %v, want true", v)
	}
}

// A string in config is a REFERENCE, and must keep resolving as one. Strategy 5 must
// not disturb this — it is the behaviour the whole config format depends on.
func TestExtractActionInputs_StringReferencesStillResolve(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"site_id", "spec_data"}}
	config := map[string]interface{}{
		"site_id":   "input_data.site_id", // dot-path reference
		"spec_data": "site_plan",          // single-segment reference
	}
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{"site_id": "abc-123"},
		"site_plan":  map[string]interface{}{"pages": 4},
	}

	got, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := got.Get("site_id"); v != "abc-123" {
		t.Errorf("dot-path reference broke: got %q, want %q", v, "abc-123")
	}
	if !got.Has("spec_data") {
		t.Error("single-segment reference broke: spec_data not resolved from collectedData")
	}
}

// A string that fails to resolve must stay ABSENT rather than becoming its own value.
// Taking it literally would turn a broken reference into a silent literal and mask real
// wiring bugs — which is why Strategy 5 is restricted to non-strings.
func TestExtractActionInputs_UnresolvedStringIsNotTakenAsLiteral(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"domain"}}
	config := map[string]interface{}{"domain": "vetcomparison.uk"}

	got, err := ExtractActionInputs(map[string]interface{}{}, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Has("domain") {
		t.Errorf("unresolved string was taken as a literal (%v); it must stay absent so a broken reference stays visible",
			got.Values["domain"])
	}
}

// A resolved value must win over the config literal — Strategy 5 only fills gaps.
func TestExtractActionInputs_ResolvedValueWinsOverLiteral(t *testing.T) {
	spec := ActionInputSpec{Optional: []string{"max_items"}}
	config := map[string]interface{}{"max_items": float64(30)}
	collected := map[string]interface{}{"max_items": float64(99)}

	got, err := ExtractActionInputs(collected, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := got.GetInt("max_items", 0); v != 99 {
		t.Errorf("collectedData value should win: got %d, want 99", v)
	}
}
