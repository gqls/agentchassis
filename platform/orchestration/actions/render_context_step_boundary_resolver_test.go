// FILE: platform/orchestration/actions/render_context_step_boundary_resolver_test.go
//
// RFC_029 §10.13 step 4, proven at the seam the conflict was recorded on.
//
// The surviving conflict class after v1.0.1315 was ONE shape: page-content-writer's
// generate_content step requests `current_page` (its prompt reads
// {{.current_page.title}} — it needs the page RECORD), and the resolver's
// whole-tree search found the record under input_data.current_page AND the page
// NAME STRING that build_render_context had filed under render_context.current_page
// / build_render_context.current_page. Different values → a
// RESOLVER_CONFLICTING_CANDIDATES row on every run (23 in the four hours after
// the roll; 100% of the class). The winner was always the record, pinned by
// bugs_open/306's tie-break, so no page was wrong — but a permanent false
// positive blocks step 5 (conflicts → refusal) for ever.
//
// These tests drive the real resolver (datahelpers.ExtractFields, the call
// execute_llm_prompt's input extraction makes) over a tree shaped like the live
// run, with the render context produced by the real renderCtxToMap, and read the
// instrument through its recorder hook. The control runs the PRE-fix shape and
// must still record a conflict: a test that cannot fail proves nothing, and this
// one fails if the instrument is ever silenced rather than the collision fixed.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// liveShapedPCWTree mirrors the depth-1 layout the conflict rows named:
// input_data.current_page (the page record) plus the build_render_context step's
// output filed under BOTH its output_field (render_context) and its step name.
func liveShapedPCWTree(renderCtx map[string]interface{}) map[string]interface{} {
	page := map[string]interface{}{
		"id": "0f0e3e4a-0000-4000-8000-0000000000aa", "name": "about",
		"title": "About us", "purpose": "who we are",
	}
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"site_id":      "6f1d8c2e-0000-4000-8000-00000000abcd",
			"domain":       "example.com",
			"current_page": page,
		},
		"render_context":       renderCtx,
		"build_render_context": renderCtx,
	}
}

func captureResolverFindings(t *testing.T) *[]datahelpers.ResolverFinding {
	t.Helper()
	got := &[]datahelpers.ResolverFinding{}
	datahelpers.SetResolverFindingRecorder(func(f datahelpers.ResolverFinding) { *got = append(*got, f) })
	t.Cleanup(func() { datahelpers.SetResolverFindingRecorder(nil) })
	return got
}

func conflictsFor(findings []datahelpers.ResolverFinding, field string) int {
	n := 0
	for _, f := range findings {
		if f.Code == datahelpers.ResolverFindingConflictingCandidates && f.Field == field {
			n++
		}
	}
	return n
}

// TestResolverSeesOnePageUnderCurrentPageAfterTheRename: with the render context
// that renderCtxToMap now emits, the only `current_page` candidates in the tree
// are the page record (twice — input_data and the ~unwrap hop to it), so the
// resolver returns the record and records NO conflict.
func TestResolverSeesOnePageUnderCurrentPageAfterTheRename(t *testing.T) {
	findings := captureResolverFindings(t)
	renderCtx := renderCtxToMap(&RenderContext{Domain: "example.com", CurrentPage: "about"})
	tree := liveShapedPCWTree(renderCtx)

	out := datahelpers.ExtractFields(tree, []string{"current_page"}, zap.NewNop())

	page, ok := out["current_page"].(map[string]interface{})
	if !ok {
		t.Fatalf("current_page resolved to %T (%v), want the page record map", out["current_page"], out["current_page"])
	}
	if page["title"] != "About us" {
		t.Errorf("current_page resolved to the wrong record: %v", page)
	}
	if n := conflictsFor(*findings, "current_page"); n != 0 {
		t.Errorf("resolver recorded %d RESOLVER_CONFLICTING_CANDIDATES for current_page, want 0 — "+
			"the page-name string is back under the same key as the page record: %+v", n, *findings)
	}
	if _, leaked := renderCtx["current_page"]; leaked {
		t.Errorf("renderCtxToMap still emits `current_page` — see TestStepOutputNeverCarriesCurrentPage")
	}
}

// TestResolverControl_PreRenameShapeStillConflicts is the control: the SAME tree
// with the render context as build_render_context used to write it (page name
// under `current_page`) must still draw a conflict row. If this stops failing
// the instrument went quiet, and the test above would be passing blind.
func TestResolverControl_PreRenameShapeStillConflicts(t *testing.T) {
	findings := captureResolverFindings(t)
	preRename := renderCtxToMap(&RenderContext{Domain: "example.com", CurrentPage: "about"})
	preRename["current_page"] = "about" // the old spelling, as stored trees before the roll carry it
	delete(preRename, "current_page_name")
	tree := liveShapedPCWTree(preRename)

	out := datahelpers.ExtractFields(tree, []string{"current_page"}, zap.NewNop())

	// PHASE 2 (2026-08-21): a conflict now resolves to NOTHING, so the pre-rename
	// shape yields no value at all. Before the flip this asserted the page RECORD
	// won on 306's tie-break. The control's load-bearing half is unchanged and is
	// the assertion below — that the shape still records a CONFLICT. If that ever
	// stops firing, the instrument has gone quiet and the test above is passing
	// blind, which is the only thing this control exists to catch.
	if out["current_page"] != nil {
		t.Errorf("pre-rename shape: Phase 2 refuses a conflict, so nothing should resolve; got %T",
			out["current_page"])
	}
	if n := conflictsFor(*findings, "current_page"); n == 0 {
		t.Fatalf("control failed: the pre-rename shape recorded no conflict — the instrument is silent, " +
			"so the passing test above cannot be trusted")
	}
}
