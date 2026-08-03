package actions

import (
	"strings"
	"testing"
)

// Branch coverage for resolveVisionImageRefs, the unwrap execute_vision_prompt
// stands on (council fee9d810: editquality's owed-proof note, and
// bug_historian's ask that the zero-renders guard fire on a genuinely EMPTY
// array and not just a missing key — the two are DIFFERENT arms, pinned
// separately here). The distinction matters because this platform has twice
// shipped unwraps that silently coerced a mismatch to empty (bugs 085/095);
// this helper's contract is loud on every mismatch shape:
//
//	missing key          → error ("nothing under")            — the audit never ran
//	non-array node       → error ("does not resolve")         — schema drift, incl.
//	                       query_database's jsonb-stringification landmine, which
//	                       delivers a STRING where the array should be
//	present, empty array → ([], nil)                          — the CALLER's
//	                       len==0 arm fails loud ("no renders"), a capture step
//	                       that ran and captured nothing
func TestResolveVisionImageRefs_MissingKeyIsLoud(t *testing.T) {
	_, err := resolveVisionImageRefs(map[string]interface{}{}, "render_audit")
	if err == nil || !strings.Contains(err.Error(), "nothing under") {
		t.Fatalf("missing key must be a loud error, got %v", err)
	}
}

func TestResolveVisionImageRefs_StringNodeIsLoudNotEmpty(t *testing.T) {
	// The jsonb-stringification shape: a query_database projection delivering
	// the renders as a JSON STRING. Coercing this to zero refs would report
	// "no renders" for an audit that captured plenty — it must name the
	// mismatch instead.
	_, err := resolveVisionImageRefs(map[string]interface{}{
		"render_audit": `[{"uri":"s3://bucket/render-sweep/a.png"}]`,
	}, "render_audit")
	if err == nil || !strings.Contains(err.Error(), "does not resolve") {
		t.Fatalf("a stringified renders array must be a loud error, got %v", err)
	}
}

func TestResolveVisionImageRefs_EmptyArrayIsZeroRefsNoError(t *testing.T) {
	refs, err := resolveVisionImageRefs(map[string]interface{}{
		"render_audit": map[string]interface{}{
			"response": map[string]interface{}{
				"renders": []interface{}{},
			},
		},
	}, "render_audit")
	if err != nil {
		t.Fatalf("an empty capture is the CALLER's fail-loud case, not a resolve error: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("want 0 refs, got %d", len(refs))
	}
}

func TestResolveVisionImageRefs_UnwrapsAllThreeShapes(t *testing.T) {
	render := map[string]interface{}{
		"uri": "s3://bucket/render-sweep/a.png", "profile": "mobile", "url": "https://x/p.html",
	}
	for name, collected := range map[string]map[string]interface{}{
		"direct array":      {"render_audit": []interface{}{render}},
		"object .renders":   {"render_audit": map[string]interface{}{"renders": []interface{}{render}}},
		"awaited .response": {"render_audit": map[string]interface{}{"response": map[string]interface{}{"renders": []interface{}{render}}}},
	} {
		refs, err := resolveVisionImageRefs(collected, "render_audit")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(refs) != 1 || refs[0].URI != "s3://bucket/render-sweep/a.png" ||
			refs[0].Profile != "mobile" || refs[0].PageURL != "https://x/p.html" {
			t.Fatalf("%s: refs mangled: %+v", name, refs)
		}
	}
}
