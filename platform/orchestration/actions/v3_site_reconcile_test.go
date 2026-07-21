package actions

import (
	"testing"

	"go.uber.org/zap"
)

// Tests for the re-plan preservation guard (bugs_open/001). The fixtures are
// shaped from the real idea.uk regression of 2026-07-14 (plan 32be2797 ->
// ff03bdef): a clean 9-page site where a re-plan re-composed "about" from
// [hero-about, info-card-grid, ...] down to a generic [hero, ...], dropped
// pages the LLM omitted, and left a catalogued page empty forever.
//
// The invariant these lock down: a re-plan must never silently redesign or drop
// an already-built page, and must still be able to compose a catalogued one.

// realised builds a pages-table row in the shape the load_existing_pages query
// returns. sections is passed as a JSON string because query_database
// stringifies jsonb — the production shape, not a convenience.
func realised(name, url, buildStatus, sectionsJSON string, locked bool) map[string]interface{} {
	return map[string]interface{}{
		"name":            name,
		"url":             url,
		"page_type":       "content",
		"title":           name,
		"build_status":    buildStatus,
		"adoption_locked": locked,
		"sections":        sectionsJSON,
	}
}

func llmPage(name, url string, sections ...string) map[string]interface{} {
	s := make([]interface{}, len(sections))
	for i, v := range sections {
		s[i] = v
	}
	return map[string]interface{}{
		"name": name, "url": url, "page_type": "content",
		"title": name, "sections": s,
	}
}

func sectionsOf(t *testing.T, pages []interface{}, name string) []string {
	t.Helper()
	for _, p := range pages {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if n, _ := pm["name"].(string); n != name {
			continue
		}
		raw, _ := pm["sections"].([]interface{})
		out := make([]string, 0, len(raw))
		for _, s := range raw {
			out = append(out, s.(string))
		}
		return out
	}
	t.Fatalf("page %q not present in plan", name)
	return nil
}

func hasPage(pages []interface{}, name string) bool {
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			if n, _ := pm["name"].(string); n == name {
				return true
			}
		}
	}
	return false
}

// The core regression: a BUILT page, under no adoption lock, re-proposed by the
// LLM under the same name with a generic composition. Before the fix the whole
// function was a no-op here and the LLM's version won.
func TestReconcile_BuiltPageCompositionSurvivesReplan(t *testing.T) {
	existing := []interface{}{
		realised("about", "/about.html", "deployed",
			`["hero-about","info-card-grid","call-to-action"]`, false),
	}
	llm := []interface{}{llmPage("about", "/about.html", "hero", "call-to-action")}

	got, _, _, _, snapped := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if snapped != 1 {
		t.Errorf("snapped_sections = %d, want 1", snapped)
	}
	want := []string{"hero-about", "info-card-grid", "call-to-action"}
	if s := sectionsOf(t, got, "about"); !equalStrings(s, want) {
		t.Errorf("about sections = %v, want %v (built page was re-composed)", s, want)
	}
}

// A built page the LLM omits entirely must be unioned back into the plan.
func TestReconcile_BuiltPageOmittedByLLMIsUnioned(t *testing.T) {
	existing := []interface{}{
		realised("report", "/report.html", "deployed",
			`["generic-text-block","info-card-grid"]`, false),
	}
	llm := []interface{}{llmPage("index", "/index.html", "hero")}

	got, unioned, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if unioned != 1 {
		t.Errorf("unioned = %d, want 1", unioned)
	}
	if !hasPage(got, "report") {
		t.Fatal("built page 'report' was dropped from the plan")
	}
	want := []string{"generic-text-block", "info-card-grid"}
	if s := sectionsOf(t, got, "report"); !equalStrings(s, want) {
		t.Errorf("report sections = %v, want %v", s, want)
	}
}

// The second defect: a catalogued page with sections=[] must be composable. The
// non-empty gate is what allows this — preserving emptiness made "re-plan to
// compose the missing pages" structurally impossible.
func TestReconcile_EmptyCataloguedPageAcceptsLLMComposition(t *testing.T) {
	existing := []interface{}{
		realised("index", "/index.html", "deployed", `["hero","features"]`, false),
		realised("tool-audience-check", "/tools/audience-check.html", "planned", `[]`, false),
	}
	llm := []interface{}{
		llmPage("index", "/index.html", "hero", "features"),
		llmPage("tool-audience-check", "/tools/audience-check.html", "tool-hero", "tool-embed"),
	}

	got, _, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	want := []string{"tool-hero", "tool-embed"}
	if s := sectionsOf(t, got, "tool-audience-check"); !equalStrings(s, want) {
		t.Errorf("catalogued page sections = %v, want %v (composition was blocked)", s, want)
	}
}

// bugs_open/050: a DEPLOYED page with sections=[] renders through another
// subsystem (a tool page, a blog-index) — its emptiness is a positive statement,
// not an absence awaiting composition. A re-plan that proposes a generic layout
// under the SAME NAME (Pass B2) must be forced back to empty, not allowed to
// inject sections onto a built page.
func TestReconcile_DeployedEmptyPageStaysEmpty_PassB2(t *testing.T) {
	existing := []interface{}{
		realised("tool-embed-x", "/tools/x.html", "deployed", `[]`, false),
	}
	llm := []interface{}{llmPage("tool-embed-x", "/tools/x.html", "hero", "features")}

	got, _, _, _, snapped := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if snapped != 1 {
		t.Errorf("snapped_sections = %d, want 1 (LLM layout should have been forced empty)", snapped)
	}
	if s := sectionsOf(t, got, "tool-embed-x"); len(s) != 0 {
		t.Errorf("deployed sectionless page sections = %v, want [] (a re-plan injected a layout)", s)
	}
}

// bugs_open/050: the same deployed sectionless page reached via Pass B (the LLM
// reuses its URL under a DIFFERENT name). The realised emptiness is authoritative
// — the snapped-back page must stay empty, dropping the LLM's proposed sections.
func TestReconcile_DeployedEmptyPageStaysEmpty_PassB(t *testing.T) {
	existing := []interface{}{
		realised("tool-embed-x", "/tools/x.html", "deployed", `[]`, false),
	}
	llm := []interface{}{llmPage("tool-embed-renamed", "/tools/x.html", "hero")}

	got, _, _, renamed, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if renamed != 1 {
		t.Errorf("snapped_rename = %d, want 1", renamed)
	}
	if !hasPage(got, "tool-embed-x") {
		t.Fatal("renamed page was not snapped back to the realised identity")
	}
	if s := sectionsOf(t, got, "tool-embed-x"); len(s) != 0 {
		t.Errorf("deployed sectionless page sections = %v, want [] (Pass B carried an LLM layout)", s)
	}
}

// bugs_open/050: a NOT-deployed catalogued page (adoption-locked, so preserved,
// and per bugs_open/051 on the site's first plan) with sections=[] has never been
// composed. A re-plan that proposes the SAME NAME (Pass B2 fall-through) must be
// allowed to compose it — take the LLM's sections.
func TestReconcile_NotDeployedEmptyPageTakesLLMSections_PassB2(t *testing.T) {
	existing := []interface{}{
		realised("catalog-tool", "/cat.html", "planned", `[]`, true),
	}
	llm := []interface{}{llmPage("catalog-tool", "/cat.html", "hero", "body")}

	got, _, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	want := []string{"hero", "body"}
	if s := sectionsOf(t, got, "catalog-tool"); !equalStrings(s, want) {
		t.Errorf("catalogued page sections = %v, want %v (composition was blocked)", s, want)
	}
}

// bugs_open/050: the same not-deployed catalogued page reached via Pass B (the
// LLM proposes its URL under a DIFFERENT name). Keep the realised identity but
// take the LLM's sections so the first-plan page can finally be composed.
func TestReconcile_NotDeployedEmptyPageTakesLLMSections_PassB(t *testing.T) {
	existing := []interface{}{
		realised("catalog-page", "/cat.html", "planned", `[]`, true),
	}
	llm := []interface{}{llmPage("catalog-fresh", "/cat.html", "hero", "body")}

	got, _, _, renamed, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if renamed != 1 {
		t.Errorf("snapped_rename = %d, want 1", renamed)
	}
	if !hasPage(got, "catalog-page") {
		t.Fatal("renamed page was not snapped back to the realised identity")
	}
	want := []string{"hero", "body"}
	if s := sectionsOf(t, got, "catalog-page"); !equalStrings(s, want) {
		t.Errorf("catalogued page sections = %v, want %v (Pass B blocked composition)", s, want)
	}
}

// A genuinely from-scratch build — nothing built, nothing locked — must still
// leave the LLM plan completely untouched.
func TestReconcile_FromScratchBuildIsUntouched(t *testing.T) {
	existing := []interface{}{
		realised("draft", "/draft.html", "planned", `[]`, false),
	}
	llm := []interface{}{llmPage("index", "/index.html", "hero")}

	got, unioned, dropped, renamed, snapped := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if unioned+dropped+renamed+snapped != 0 {
		t.Errorf("expected a no-op, got unioned=%d dropped=%d renamed=%d snapped=%d",
			unioned, dropped, renamed, snapped)
	}
	if len(got) != 1 || !hasPage(got, "index") {
		t.Errorf("plan was modified on a from-scratch build: %v", got)
	}
}

// Degradation guarantee: on a chassis whose load_existing_pages query does not
// yet SELECT build_status, the widened set must collapse back to the
// adoption-locked set rather than misbehave. This is what makes the Go change
// and the query change safe to land in either order.
func TestReconcile_MissingBuildStatusFallsBackToLockedOnly(t *testing.T) {
	noStatus := map[string]interface{}{
		"name": "about", "url": "/about.html", "page_type": "content",
		"adoption_locked": false,
		"sections":        `["hero-about","info-card-grid"]`,
	}
	llm := []interface{}{llmPage("about", "/about.html", "hero")}

	got, unioned, _, _, snapped := reconcilePlanWithRealised(llm, []interface{}{noStatus}, zap.NewNop())

	if snapped != 0 || unioned != 0 {
		t.Errorf("expected pre-fix no-op without build_status, got snapped=%d unioned=%d", snapped, unioned)
	}
	if s := sectionsOf(t, got, "about"); !equalStrings(s, []string{"hero"}) {
		t.Errorf("about sections = %v, want [hero] (should be untouched)", s)
	}
}

// An adoption-locked page must keep behaving exactly as before the widening.
func TestReconcile_AdoptionLockedStillPreserved(t *testing.T) {
	existing := []interface{}{
		realised("guide-basics", "/guides/basics.html", "planned",
			`["guide-hero","guide-body"]`, true),
	}
	llm := []interface{}{llmPage("index", "/index.html", "hero")}

	got, unioned, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if unioned != 1 || !hasPage(got, "guide-basics") {
		t.Errorf("adoption-locked page not preserved: unioned=%d pages=%v", unioned, got)
	}
}

// Truncation must never evict a built page, however many pages the LLM invents.
func TestTruncate_KeepsBuiltPagesOverInventedOnes(t *testing.T) {
	built := []interface{}{
		realised("index", "/index.html", "deployed", `["hero"]`, false),
		realised("about", "/about.html", "deployed", `["hero-about"]`, false),
	}
	pages := []interface{}{
		llmPage("invented-1", "/i1.html", "hero"),
		llmPage("invented-2", "/i2.html", "hero"),
		llmPage("index", "/index.html", "hero"),
		llmPage("about", "/about.html", "hero"),
	}

	got := truncatePreservingRealised(pages, built, 2, zap.NewNop())

	if len(got) != 2 {
		t.Fatalf("truncated to %d pages, want 2", len(got))
	}
	if !hasPage(got, "index") || !hasPage(got, "about") {
		t.Errorf("truncation evicted a built page: %v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
