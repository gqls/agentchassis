// FILE: platform/orchestration/actions/audit_tone_routes_to_copy_editor_test.go
//
// A `tone` audit finding is a STYLISTIC adjustment to copy that already exists.
// It must reach an agent that EDITS that copy, not one that regenerates the page.
//
// Why this is a test and not a comment. Until 2026-08-24 `tone` filed `tone_shift`
// at page-build-handler, whose save path DELETEs and re-INSERTs the component row,
// so a resolver-sourced key that fails to resolve on that run is destroyed rather
// than left alone. The recorded cost, quoted from plan_sections_action.go's
// carry-stored comment (bugs_open/238): finetuning.uk's homepage "lost all 11 of
// its non-llm URL keys to one tone_shift and served five <img src=""> plus six
// vanished controls". Asking for a better sentence emptied five image tags.
//
// copy-editor (CQ-024) is the surgical alternative — `field_updates` on named
// fields of named components — and structurally cannot write to a page at all:
// migration 447 RAISEs if a page-writing step is added to it. So this route files
// an auto-PROPOSAL that parks for a human, never an auto-edit (owner decision D2).

package actions

import (
	"testing"

	"github.com/google/uuid"
)

// pageRegeneratingHandlers are handlers whose save path replaces the component
// row wholesale. A stylistic finding must never be routed at one of these.
var pageRegeneratingHandlers = map[string]string{
	"page-build-handler": "save_page_sections DELETEs and re-INSERTs the row (bugs_open/238)",
}

func TestToneFindingRoutesToCopyEditorNotAPageRegenerator(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()
	pages := map[string]pageInfo{"index": {ID: pageID, Name: "index"}}

	c := classifyFinding(auditFinding{
		Category:    "tone",
		Page:        "index",
		Description: "the homepage opens by describing itself rather than addressing the reader",
		Severity:    "medium",
	}, pages, siteID, "content-quality-auditor")

	if c.ItemType != "needs_copy_edit" {
		t.Errorf("tone routed to item_type %q, want needs_copy_edit", c.ItemType)
	}
	if c.HandlerAgent != "copy-editor" {
		t.Errorf("tone routed to handler %q, want copy-editor", c.HandlerAgent)
	}
	if why, isRegen := pageRegeneratingHandlers[c.HandlerAgent]; isRegen {
		t.Errorf("tone routed at %q, which regenerates the page: %s. A stylistic finding "+
			"must be an EDIT, not a rebuild", c.HandlerAgent, why)
	}

	// copy-editor's dispatched entry path reads the page off the work item
	// (migration 579), so a route that loses the page id cannot be dispatched.
	if c.PageID == nil || *c.PageID != pageID {
		t.Error("tone finding must carry the page id — copy-editor's dispatched entry " +
			"path reads page_id off the work item and has nothing else to go on")
	}
}

// TestToneRerouteIsScopedToTone is the control: the change must not have moved
// the neighbouring content categories, which legitimately DO want a rebuild.
func TestToneRerouteIsScopedToTone(t *testing.T) {
	siteID := uuid.New()
	pages := map[string]pageInfo{"index": {ID: uuid.New(), Name: "index"}}

	for _, tc := range []struct{ category, wantType, wantHandler string }{
		{"gap", "needs_content_page", "page-build-handler"},
		{"content", "content_rewrite", "page-build-handler"},
		{"differentiation", "content_rewrite", "page-build-handler"},
		{"structure", "content_rewrite", "page-build-handler"},
	} {
		c := classifyFinding(auditFinding{
			Category: tc.category, Page: "index", Description: "x", Severity: "medium",
		}, pages, siteID, "content-quality-auditor")
		if c.ItemType != tc.wantType || c.HandlerAgent != tc.wantHandler {
			t.Errorf("category %q routed to %s/%s, want %s/%s — the tone reroute must not "+
				"have moved this", tc.category, c.ItemType, c.HandlerAgent, tc.wantType, tc.wantHandler)
		}
	}
}
