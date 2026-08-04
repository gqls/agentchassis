// FILE: platform/orchestration/actions/enrich_fingerprint_shape_test.go
//
// bugs_open/192, second instance. This step is wired with
// output_field: design_fingerprint — the same key extract_design_fingerprint
// wrote — so whatever it returns REPLACES the fingerprint wholesale
// (coordinator.go storeActionResult). Its two success paths always returned
// the fingerprint; its two EARLY-OUTS returned a status stub, so a present but
// unparseable fingerprint was destroyed and every downstream consumer was
// handed {"status": "invalid_fingerprint"} where a fingerprint belongs.
//
// It was found by the fleet census the council's bug_historian seat asked for,
// not by any reported failure — so there is no production symptom to regress
// against, and these cases ARE the regression evidence. They fail on the
// pre-fix code, which is the only property that makes them worth having.
package actions

import (
	"context"
	"reflect"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func enrichFingerprintParams(collected map[string]interface{}) ActionParams {
	return ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    collected,
	}
}

// assertNotAStatusStub is the contract: because output_field names this step's
// own input, the return value must never be a bare {"status": …} envelope —
// that does not replace the fingerprint with a worse fingerprint, it replaces
// it with something that is not a fingerprint at all.
func assertNotAStatusStub(t *testing.T, got interface{}) {
	t.Helper()
	m, ok := got.(map[string]interface{})
	if !ok {
		return // not a map at all — cannot be the stub
	}
	if _, hasStatus := m["status"]; hasStatus && len(m) == 1 {
		t.Fatalf("bugs_open/192 regression: returned a status stub %v, which REPLACES "+
			"design_fingerprint (output_field) with a non-fingerprint", m)
	}
}

// A present-but-unparseable fingerprint is still the caller's data. Replacing
// it with a stub destroys it; this is the case that loses real information.
func TestEnrichFingerprint_NonMapFingerprint_ReturnedUnchanged(t *testing.T) {
	original := "a fingerprint that did not survive whatever produced it"
	collected := map[string]interface{}{
		"design_fingerprint": original,
	}

	got, err := EnrichFingerprintWithCSSAction(context.Background(), enrichFingerprintParams(collected))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNotAStatusStub(t, got)
	if !reflect.DeepEqual(got, original) {
		t.Errorf("an unparseable fingerprint must come back byte-identical, not be replaced.\n got: %v\nwant: %v", got, original)
	}
}

// No fingerprint at all: nothing to preserve, but the step must not INVENT one.
// Writing {"status": "no_fingerprint"} makes design_fingerprint a truthy object
// for every downstream consumer asking "is there a fingerprint?".
func TestEnrichFingerprint_NoFingerprint_DoesNotInventOne(t *testing.T) {
	got, err := EnrichFingerprintWithCSSAction(context.Background(), enrichFingerprintParams(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNotAStatusStub(t, got)
	if got != nil {
		t.Errorf("with no fingerprint to enrich the step must not write a value, got %v", got)
	}
}

// The ordinary path, kept so the two negatives above cannot be satisfied by an
// action that simply stopped working: a real fingerprint with no CSS to apply
// still comes back as a fingerprint, annotated in place rather than wrapped.
func TestEnrichFingerprint_NoCSSContent_ReturnsThePlanAnnotatedInPlace(t *testing.T) {
	collected := map[string]interface{}{
		"design_fingerprint": map[string]interface{}{
			"palette":     []interface{}{"#111111"},
			"type_scale":  "major-third",
			"source_urls": []interface{}{"https://example.com"},
		},
	}

	got, err := EnrichFingerprintWithCSSAction(context.Background(), enrichFingerprintParams(collected))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNotAStatusStub(t, got)

	fp, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("return value is not a fingerprint map: %T", got)
	}
	if _, nested := fp["design_fingerprint"]; nested {
		t.Error("bugs_open/192 regression: the fingerprint is nested under its own key")
	}
	if fp["type_scale"] != "major-third" {
		t.Errorf("existing fingerprint fields must survive, type_scale = %v", fp["type_scale"])
	}
	if fp["css_enrichment"] != "no_css_content_found" {
		t.Errorf("css_enrichment = %v, want the in-place no_css_content_found marker", fp["css_enrichment"])
	}
}
