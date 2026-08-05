// FILE: platform/orchestration/actions/component_library_cta_url_fallback_test.go
//
// Guards the fix for the phantom-CTA bug found live on 2026-08-05: a hero (or
// any of the other ctaFieldNames components) whose LLM-authored content_data
// carries cta_text but no cta_url — because resolve_internal_links couldn't
// match a real page and correctly left it unset — rendered a real anchor
// pointing at a hardcoded "/contact.html" default anyway, defeating the
// template's own `{{if and .cta_text .cta_url}}` guard. Correct-or-absent
// (LNK-005) says the button must not render at all in that case; these tests
// mutate the fix back to the old default to confirm they'd have caught it.

package actions

import "testing"

func TestContextToInterfaceMapLeavesCTAUrlAbsentWhenUnresolved(t *testing.T) {
	ctx := fullyPopulatedRenderContext()
	ctx.CTAUrl = "" // resolver left it unset — the live-bug precondition
	ctx.ContentData = map[string]interface{}{
		"cta_text": "Read the tungsten percentage guide", // LLM-authored, real
		// deliberately no cta_url — never resolved to a real page
	}

	got := contextToInterfaceMap(ctx)

	if v, _ := got["cta_url"].(string); v != "" {
		t.Errorf("cta_url = %q, want empty — an unresolved CTA must stay "+
			"correct-or-absent, not fall back to a fabricated destination "+
			"paired with a real cta_text (%q)", v, got["cta_text"])
	}
	if v, _ := got["cta_text"].(string); v != "Read the tungsten percentage guide" {
		t.Fatalf("cta_text = %q, want the ContentData value — fixture broken", v)
	}
}

func TestContextToMapLeavesCTAUrlAbsentWhenUnresolved(t *testing.T) {
	ctx := fullyPopulatedRenderContext()
	ctx.CTAUrl = ""
	ctx.ContentData = map[string]interface{}{
		"cta_text": "Read the tungsten percentage guide",
	}

	got := contextToMap(ctx)

	if v := got["cta_url"]; v != "" {
		t.Errorf("cta_url = %q, want empty — same correct-or-absent contract "+
			"as contextToInterfaceMap (bugs_open/109: the two paths must not "+
			"diverge on what they default)", v)
	}
}

// TestContextToInterfaceMapStillDefaultsCTAUrlWhenStructFieldIsSet pins the
// non-regression half: a genuinely resolved ctx.CTAUrl (the struct field, not
// ContentData) must still pass through unchanged. The fix removes the
// fallback, not the field.
func TestContextToInterfaceMapStillPassesThroughAResolvedCTAUrl(t *testing.T) {
	ctx := fullyPopulatedRenderContext()
	ctx.CTAUrl = "/blog/tungsten-guide.html"

	got := contextToInterfaceMap(ctx)

	if v, _ := got["cta_url"].(string); v != "/blog/tungsten-guide.html" {
		t.Errorf("cta_url = %q, want the resolved struct value passed through", v)
	}
}
