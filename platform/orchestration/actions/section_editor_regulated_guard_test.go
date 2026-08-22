package actions

import (
	"os"
	"strings"
	"testing"
)

// TestEveryPersistingEditBranchIsGuarded is the structural answer to a council
// objection, rather than another one-off patch.
//
// The history is the argument. The regulated guard (CGV-033) was first wired into
// ONE branch of the persist switch. Round 2 of correlation aac38d5b objected that
// the section editor had no guard at all; round 3 objected that the fix covered
// one path and asked about others. Checking found THREE persisting branches —
// content_edit, rendered_html_transform and component_swap — of which two had the
// guard (one added by this change, one copied in by the lane that introduced that
// branch) and one had nothing.
//
// Patching the third fixes today. This test fixes tomorrow: a FOURTH branch that
// writes without the guard fails here, loudly, at the moment it is added.
//
// It scans the persist switch only — not the file — because a whole-file scan
// finds these words in comments and passes while the code is wrong. That is not
// hypothetical: the sibling test in render_sitemap_test.go was written that way,
// failed to catch its own mutation, and had to be rewritten.
func TestEveryPersistingEditBranchIsGuarded(t *testing.T) {
	src, err := os.ReadFile("section_editor_actions.go")
	if err != nil {
		t.Fatalf("cannot read section_editor_actions.go: %v", err)
	}
	body := string(src)

	// The persist switch is the one whose branches call updatePageComponent*.
	start := strings.Index(body, "// --- Persist the page_components row ---")
	if start < 0 {
		t.Fatal("cannot find the persist switch marker comment — if it was renamed, " +
			"re-anchor this test rather than deleting it")
	}
	end := strings.Index(body[start:], "if errors.Is(err, errComponentLocked)")
	if end < 0 {
		t.Fatal("cannot find the end of the persist switch")
	}
	sw := body[start : start+end]

	branches := 0
	for _, part := range strings.Split(sw, "case \"")[1:] {
		name := part[:strings.Index(part, "\"")]
		branches++
		if !strings.Contains(part, "refuseRegulatedIdentityEdit") {
			t.Errorf("persist branch %q writes page_components without the regulated-identity "+
				"guard. Every branch that persists LLM-authored HTML must call "+
				"refuseRegulatedIdentityEdit, or a regulated claim reaches a live page by that "+
				"route — the exact 'bypassed, re-run, or edited' gap CGV-033 exists to close.", name)
		}
		if !strings.Contains(part, "updatePageComponent") {
			t.Logf("note: branch %q does not call updatePageComponent* — if it no longer "+
				"persists, the guard requirement above may not apply to it", name)
		}
	}
	if branches < 3 {
		t.Errorf("expected at least 3 persist branches, found %d — the switch was restructured "+
			"and this test is no longer measuring what it thinks", branches)
	}
	t.Logf("persist branches checked: %d, all guarded", branches)
}
