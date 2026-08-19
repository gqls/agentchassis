package datahelpers

// RFC_029 §10.13 step 3 (owner ruling 2026-08-18): ensureCoreFields recovers
// current_page / current_section / render_context ONLY when they were requested,
// while domain / objective / model stay UNCONDITIONAL. These tests pin both
// halves, because the pre-gate suite pinned neither: the gate changed behaviour
// and every existing datahelpers test still passed — which is exactly why a
// behaviour nobody declared was able to carry 63% of an observation window.
//
// Mutation proof: remove `&& requested("current_page")` from the gate →
// TestUnrequestedPageFieldsAreNotInjected FAILS. Add a `requested("domain")`
// guard to the domain block → TestBusinessFieldsStayUnconditional FAILS.

import (
	"testing"

	"go.uber.org/zap"
)

// A tree where every one of the six core fields is findable by the whole-tree
// search, none of them at the root — so presence in the result can ONLY come
// from ensureCoreFields' recovery.
func coreFieldsFixture() map[string]interface{} {
	return map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.co.uk"},
		"brief":       map[string]interface{}{"objective": "sell widgets", "model": "claude-x"},
		"loop_state": map[string]interface{}{
			"current_page":    map[string]interface{}{"name": "about"},
			"current_section": map[string]interface{}{"id": "hero"},
			"render_context":  map[string]interface{}{"theme": "dark"},
		},
	}
}

func TestUnrequestedPageFieldsAreNotInjected(t *testing.T) {
	// Request something unrelated; the three page-ish fields must NOT appear.
	got := ExtractFields(coreFieldsFixture(), []string{"brief"}, zap.NewNop())

	for _, f := range []string{"current_page", "current_section", "render_context"} {
		if _, present := got[f]; present {
			t.Errorf("%s was injected although nothing requested it — the RFC_029 §10.13 step-3 gate is open", f)
		}
	}
}

func TestBusinessFieldsStayUnconditional(t *testing.T) {
	// Same call: domain / objective / model MUST still be recovered — 39/10/6
	// live steps read them from templates without declaring them.
	got := ExtractFields(coreFieldsFixture(), []string{"brief"}, zap.NewNop())

	if got["domain"] != "example.co.uk" {
		t.Errorf("domain = %v, want example.co.uk — the unconditional recovery must survive the gate", got["domain"])
	}
	if got["objective"] != "sell widgets" {
		t.Errorf("objective = %v, want \"sell widgets\"", got["objective"])
	}
	if got["model"] != "claude-x" {
		t.Errorf("model = %v, want claude-x", got["model"])
	}
}

func TestRequestedPageFieldIsStillRecoveredByTheFallback(t *testing.T) {
	// render_context is requested but NOT findable by the special-case path
	// (there is none for it) — it must still arrive via ensureCoreFields'
	// fallback, exactly as before the gate. The gate closes the UNREQUESTED
	// path only.
	got := ExtractFields(coreFieldsFixture(), []string{"render_context"}, zap.NewNop())

	rc, _ := got["render_context"].(map[string]interface{})
	if rc["theme"] != "dark" {
		t.Fatalf("render_context = %v, want the recovered map — a REQUESTED page-ish field must still be recovered", got["render_context"])
	}
	// And the unrequested siblings still stay out.
	for _, f := range []string{"current_page", "current_section"} {
		if _, present := got[f]; present {
			t.Errorf("%s leaked in while only render_context was requested", f)
		}
	}
}

func TestRequestedCurrentPageViaSpecialCaseIsUnchanged(t *testing.T) {
	// current_page has its own special-case extraction ahead of ensureCoreFields;
	// requesting it must behave exactly as before (found, populated).
	got := ExtractFields(coreFieldsFixture(), []string{"current_page"}, zap.NewNop())
	cp, _ := got["current_page"].(map[string]interface{})
	if cp["name"] != "about" {
		t.Fatalf("current_page = %v, want the page map", got["current_page"])
	}
}

func TestGateThroughExtractActionInputs_DispatchLoopShape(t *testing.T) {
	// The production case that motivated the gate: a step whose spec declares
	// only work_item_id (claim_work_item), in a collected_data tree holding a
	// stale current_page from a previous loop iteration. Before the gate,
	// ExtractActionInputs' Values carried that stale page (a value nobody read);
	// after it, the page is absent and work_item_id is untouched.
	spec := ActionInputSpec{Required: []string{"work_item_id"}}
	data := map[string]interface{}{
		"current_item":   map[string]interface{}{"id": "item-7"},
		"handler_result": map[string]interface{}{"input_data": map[string]interface{}{"current_page": map[string]interface{}{"reason": "stale"}}},
	}
	config := map[string]interface{}{"work_item_id": "current_item.id"}

	inputs, err := ExtractActionInputs(data, config, spec, zap.NewNop())
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if inputs.Get("work_item_id") != "item-7" {
		t.Fatalf("work_item_id = %q, want item-7", inputs.Get("work_item_id"))
	}
	if _, present := inputs.Values["current_page"]; present {
		t.Fatalf("current_page = %v was injected into a step that declares only work_item_id — the gate must close this", inputs.Values["current_page"])
	}
}
