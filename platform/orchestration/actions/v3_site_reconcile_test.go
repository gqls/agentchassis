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

// bugs_open/037: a needs_rebuild page still holds its intended composition in
// pages.sections, and every writer of that status means "re-render as planned",
// never "recompose from scratch". A re-plan re-proposing it under the same name
// with a different composition must snap back — the dartsonline index case
// (2026-07-20: needs_rebuild, lost `differentiators` + `content-listing` to the
// LLM's proposal because the guard only protected `deployed`).
func TestReconcile_NeedsRebuildPageCompositionSurvivesReplan(t *testing.T) {
	existing := []interface{}{
		realised("index", "/index.html", "needs_rebuild",
			`["hero","category-listing","product-grid","differentiators","call-to-action","testimonials","content-listing"]`, false),
	}
	llm := []interface{}{llmPage("index", "/index.html",
		"hero", "product-grid", "category-listing", "features", "call-to-action", "testimonials")}

	got, _, _, _, snapped := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if snapped != 1 {
		t.Errorf("snapped_sections = %d, want 1 (needs_rebuild page was re-composed)", snapped)
	}
	want := []string{"hero", "category-listing", "product-grid", "differentiators", "call-to-action", "testimonials", "content-listing"}
	if s := sectionsOf(t, got, "index"); !equalStrings(s, want) {
		t.Errorf("needs_rebuild index sections = %v, want %v", s, want)
	}
}

// bugs_open/037: a needs_rebuild page the LLM omits must be unioned back, not
// dropped — the same protection a deployed page gets.
func TestReconcile_NeedsRebuildPageOmittedByLLMIsUnioned(t *testing.T) {
	existing := []interface{}{
		realised("contact", "/contact.html", "needs_rebuild",
			`["hero-contact","contact-form","call-to-action"]`, false),
	}
	llm := []interface{}{llmPage("index", "/index.html", "hero")}

	got, unioned, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if unioned != 1 || !hasPage(got, "contact") {
		t.Fatalf("needs_rebuild page dropped: unioned=%d present=%v", unioned, hasPage(got, "contact"))
	}
	want := []string{"hero-contact", "contact-form", "call-to-action"}
	if s := sectionsOf(t, got, "contact"); !equalStrings(s, want) {
		t.Errorf("contact sections = %v, want %v", s, want)
	}
}

// bugs_open/037 x 050 interaction — the load-bearing test for the design choice.
// A needs_rebuild page with EMPTY sections is NOT necessarily rendered elsewhere:
// it may be genuinely awaiting composition (dartsonline brands-index:
// needs_rebuild, 0 sections, 0 components). Bringing needs_rebuild into the
// preserved MEMBERSHIP set must NOT force such a page back to empty the way a
// DEPLOYED sectionless page is (that would block its composition forever). This
// passes only because the empty-gate stays keyed on realisedPageIsBuilt
// (== deployed) while membership uses realisedPageCompositionIsPreserved; a naive
// "widen realisedPageIsBuilt to include needs_rebuild" would fail it by
// force-emptying brands-index.
func TestReconcile_NeedsRebuildEmptyPageIsStillComposable(t *testing.T) {
	existing := []interface{}{
		realised("index", "/index.html", "deployed", `["hero","features"]`, false),
		realised("brands-index", "/brands.html", "needs_rebuild", `[]`, false),
	}
	llm := []interface{}{
		llmPage("index", "/index.html", "hero", "features"),
		llmPage("brands-index", "/brands.html", "category-listing"),
	}

	got, _, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	want := []string{"category-listing"}
	if s := sectionsOf(t, got, "brands-index"); !equalStrings(s, want) {
		t.Errorf("needs_rebuild empty page sections = %v, want %v (composition blocked or force-emptied)", s, want)
	}
}

// The membership predicate itself: needs_rebuild joins deployed; planned and a
// missing column do not (the latter is the safe-degradation fallback that lets
// the Go change and migration 173 land in either order).
func TestRealisedPageCompositionIsPreserved(t *testing.T) {
	cases := []struct {
		row  map[string]interface{}
		want bool
	}{
		{map[string]interface{}{"build_status": "deployed"}, true},
		{map[string]interface{}{"build_status": "needs_rebuild"}, true},
		{map[string]interface{}{"build_status": "planned"}, false},
		{map[string]interface{}{}, false},
	}
	for _, c := range cases {
		if got := realisedPageCompositionIsPreserved(c.row); got != c.want {
			t.Errorf("realisedPageCompositionIsPreserved(%v) = %v, want %v", c.row, got, c.want)
		}
	}
}

// features_open/012 (bugs_open/037 fix step 4): recomposePagesFromSpec reads the
// explicit redesign list from the trigger spec at input_data.spec.recompose_pages.
func TestRecompose_ReadsSpecList(t *testing.T) {
	cd := map[string]interface{}{
		"input_data": map[string]interface{}{
			"spec": map[string]interface{}{
				"recompose_pages": []interface{}{"index", "about"},
			},
		},
	}
	set := recomposePagesFromSpec(cd, zap.NewNop())
	if len(set) != 2 || !set["index"] || !set["about"] {
		t.Errorf("recompose set = %v, want {index, about}", set)
	}
	if s := recomposePagesFromSpec(map[string]interface{}{}, zap.NewNop()); s != nil {
		t.Errorf("expected nil for absent recompose_pages, got %v", s)
	}
}

// features_open/012: filterOutRecomposePages releases only the named realised
// pages, leaving the rest of the preserve set intact.
func TestRecompose_FilterReleasesOnlyNamedPages(t *testing.T) {
	existing := []interface{}{
		realised("index", "/index.html", "deployed", `["hero","features"]`, false),
		realised("contact", "/contact.html", "needs_rebuild", `["hero-contact"]`, false),
	}
	kept := filterOutRecomposePages(existing, map[string]bool{"index": true}, zap.NewNop())
	if hasPage(kept, "index") {
		t.Error("recompose page 'index' was not released from the realised set")
	}
	if !hasPage(kept, "contact") {
		t.Error("non-recompose page 'contact' was wrongly released")
	}
}

// features_open/012 end-to-end: a page named in recompose_pages has its LLM
// composition honoured (the guard is released for it), while a peer NOT named is
// still preserved. Discriminating: without the filter, 'index' would be snapped
// back to its three realised sections.
func TestRecompose_EndToEnd_NamedPageIsRedesignedPeerIsPreserved(t *testing.T) {
	existing := []interface{}{
		realised("index", "/index.html", "deployed", `["hero","category-listing","features"]`, false),
		realised("about", "/about.html", "deployed", `["hero-about","about-body"]`, false),
	}
	existing = filterOutRecomposePages(existing, map[string]bool{"index": true}, zap.NewNop())

	llm := []interface{}{
		llmPage("index", "/index.html", "hero", "testimonials"), // redesigned
		llmPage("about", "/about.html", "hero", "about-body"),   // LLM tries to genericise it
	}
	got, _, _, _, _ := reconcilePlanWithRealised(llm, existing, zap.NewNop())

	if s := sectionsOf(t, got, "index"); !equalStrings(s, []string{"hero", "testimonials"}) {
		t.Errorf("recomposed index sections = %v, want [hero testimonials] (LLM should govern)", s)
	}
	if s := sectionsOf(t, got, "about"); !equalStrings(s, []string{"hero-about", "about-body"}) {
		t.Errorf("preserved about sections = %v, want [hero-about about-body] (guard should hold for an unnamed page)", s)
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
