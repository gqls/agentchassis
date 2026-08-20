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

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

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

// This test used to assert the same contract against contextToMap, the string
// map the deleted regex fallback rendered from (bugs_open/260 deleted both).
// Asserting it against a function that no longer exists would have been the
// cheapest possible rework and the emptiest: the property that matters was
// never "the map has no fabricated value", it was "no fabricated destination
// reaches a rendered page". So it now asserts that at the SEAM, which is the
// only surviving surface and the one a live page actually goes through.
//
// It also covers bugs_open/203's other two members (primary_cta_url ->
// /contact.html, secondary_cta_url -> /about.html), which lived in the same
// deleted function: nothing else in the suite would notice their return.
func TestRenderedOutputNeverFabricatesAnUnresolvedCTADestination(t *testing.T) {
	ctx := fullyPopulatedRenderContext()
	ctx.CTAUrl = ""
	ctx.ContentData = map[string]interface{}{
		"cta_text": "Read the tungsten percentage guide",
	}

	const tmpl = `<section>
{{if and .cta_text .cta_url}}<a class="cta" href="{{.cta_url}}">{{.cta_text}}</a>{{end}}
<a class="primary" href="{{.primary_cta_url}}">go</a>
<a class="secondary" href="{{.secondary_cta_url}}">go</a>
</section>`

	out := mustRender(t, tmpl, ctx, zap.NewNop())

	for _, fabricated := range []string{"/contact.html", "/about.html"} {
		if strings.Contains(out, fabricated) {
			t.Errorf("rendered output contains the fabricated destination %q — "+
				"correct-or-absent (LNK-005, bugs_open/203) means an unresolved "+
				"CTA renders no destination at all:\n%s", fabricated, out)
		}
	}
	if strings.Contains(out, `class="cta"`) {
		t.Errorf("the guarded CTA anchor rendered with no resolved cta_url:\n%s", out)
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
