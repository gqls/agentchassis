package actions

import "testing"

// The containment property council 060bcc0a r2/r3 (guardian) asked to see
// asserted directly: with the step flag ON, a rerender dispatched for ANY
// reason other than the literal_markdown repair must NOT strip — the flag
// arms the capability, the dispatch's spec.reason scopes it. A regression
// here silently widens a scoped repair into a blanket mutation of the
// fleet's highest-volume pipeline (13,993 lifetime completes).
func TestShouldStripLiteralMarkdownGate(t *testing.T) {
	flagOn := map[string]interface{}{"strip_literal_markdown": true}

	for _, reason := range []string{
		"image_landed", "section_data_resolved", "cta_links_stale",
		"template_changed",                      // every other live check_rerender_mode reason
		"",                                      // reason absent entirely (a caller that omits spec.reason)
		"literal_markdown ", "LITERAL_MARKDOWN", // near-misses must not match
	} {
		if shouldStripLiteralMarkdown(flagOn, reason) {
			t.Errorf("flag ON + reason %q must NOT strip", reason)
		}
	}

	if !shouldStripLiteralMarkdown(flagOn, "literal_markdown") {
		t.Error("flag ON + reason literal_markdown MUST strip — the repair depends on it")
	}

	// The other half of the double gate: reason alone, capability not enabled.
	for _, cfg := range []map[string]interface{}{
		{},                                 // key absent (default OFF)
		{"strip_literal_markdown": false},  // explicit off
		{"strip_literal_markdown": "true"}, // string, not bool — config typo must fail closed
		{"strip_literal_markdown": 1},      // number — same
		{"strip_literal_markdown": interface{}(nil)}, // null
	} {
		if shouldStripLiteralMarkdown(cfg, "literal_markdown") {
			t.Errorf("config %v + reason literal_markdown must NOT strip (flag not a true bool)", cfg)
		}
	}
}
