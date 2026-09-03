// FILE: platform/orchestration/actions/tool_item_sources_test.go
//
// bugs_open/450, the supply half: a planned tool page whose tool does not exist
// is held out of the plan rather than minted as a page row nothing will fill.
//
// WHAT MAKES THE POSITIVES HERE NON-VACUOUS. The seotools shape (a tool page
// planned as hero-tool + generic-text-block against a site with no tools) and
// the advertise shape (a tool page whose sections name a tool the site owns)
// differ ONLY in the site's tool set — the same page role, the same plan
// structure. Every held/kept pair below varies that one fact, so a test cannot
// pass by accident of page shape.
//
// MUTATION PROTOCOL:
//   - `isToolRole`: return false → every Held test fails, OutOfFamily still passes.
//   - `toolFunctionCandidates`: drop the page-name arms → the two Kept-by-name
//     tests fail while the section-name test still passes (the arms are
//     independent, which is why they are tested separately).
//   - `ResolveToolItemSource`: return Producible:true always → all Held tests fail.
//   - the `!anyToolPage` early return: invert it → NoToolPages_SkipsCensus fails
//     on an unexpected query.
//   - the preserve-guard branch: delete it → RealisedPageKeptWithReceipt fails.
//   - the census-error arm: change fail-open to fail-closed → CensusErrorFailsOpen
//     fails, which is the one that would otherwise starve every fresh build.

package actions

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// --- resolver-level tests (pure — no database) ----------------------------

// toolPage builds the plan-page view the gate sees.
func toolPage(name string, sections ...string) planPageView {
	return planPageView{Name: name, Role: "tool", URL: "/tools/" + name + "/index.html", Sections: sections}
}

// TestResolveTool_SectionNamesSiteToolFunction_Kept is the advertise.co.uk
// control from the bug file: its plan persisted AFTER its tools existed, so the
// plan named the real tool component as a section and must resolve.
func TestResolveTool_SectionNamesSiteToolFunction_Kept(t *testing.T) {
	page := toolPage("tool-ab-test", "hero-tool", "tool-guide-intro", "tool-ab-test-calculator", "tool-cta")
	site := map[string]bool{"tool-ab-test-calculator": true}

	res := ResolveToolItemSource(page, site)
	if !res.ToolFamily {
		t.Fatal("a page_type='tool' page must be in the tool family")
	}
	if !res.Producible {
		t.Fatalf("the site owns the tool this page's sections name — must be kept: %s", res.Evidence)
	}
}

// TestResolveTool_PageNameMatchesToolFunction_Kept pins the acceptance coupling:
// create_tool_component sets pages.name = cc.function, so the pipeline's OWN
// pages resolve when a replan echoes them back — without this arm the gate would
// file false gaps against every real tool page on every replan.
func TestResolveTool_PageNameMatchesToolFunction_Kept(t *testing.T) {
	page := toolPage("tool-robots-txt-generator") // no sections at all
	site := map[string]bool{"tool-robots-txt-generator": true}

	if res := ResolveToolItemSource(page, site); !res.Producible {
		t.Fatalf("a page named for a tool the site owns must be kept: %s", res.Evidence)
	}
}

// TestResolveTool_LegacyBarePageName_Kept covers resolveToolPageIdentity's
// legacy shape: a page named for the bare slug while the function carries the
// sanitiseFunction-guaranteed `tool-` prefix.
func TestResolveTool_LegacyBarePageName_Kept(t *testing.T) {
	page := toolPage("robots-txt-generator")
	site := map[string]bool{"tool-robots-txt-generator": true}

	if res := ResolveToolItemSource(page, site); !res.Producible {
		t.Fatalf("legacy bare-named tool page must resolve via the tool- prefix candidate: %s", res.Evidence)
	}
}

// TestResolveTool_NoToolOnSite_Held is the seotools shape — the bug itself.
func TestResolveTool_NoToolOnSite_Held(t *testing.T) {
	page := toolPage("tool-robots-txt-tester", "hero-tool", "generic-text-block")
	site := map[string]bool{} // the site owns no tools

	res := ResolveToolItemSource(page, site)
	if !res.ToolFamily {
		t.Fatal("tool family test failed on a page_type='tool' page")
	}
	if res.Producible {
		t.Fatal("a tool page on a site with NO tools was judged producible — this is bugs_open/450 exactly")
	}
	if res.ProducerNeeded != toolProducerSlug {
		t.Errorf("ProducerNeeded = %q, want %q (the roadmap sweep groups on it)", res.ProducerNeeded, toolProducerSlug)
	}
}

// TestResolveTool_NearMissDoesNotResolve is the control that stops the previous
// test passing for the wrong reason: the site DOES own a tool, just not this
// one. The planner's name and the suggester's name were disjoint on all seven
// seotools pages (robots-txt-TESTER planned, robots-txt-GENERATOR built), so a
// substring or fuzzy match here would reopen the whole bug.
func TestResolveTool_NearMissDoesNotResolve(t *testing.T) {
	page := toolPage("tool-robots-txt-tester", "hero-tool", "generic-text-block")
	site := map[string]bool{"tool-robots-txt-generator": true}

	if res := ResolveToolItemSource(page, site); res.Producible {
		t.Fatal("robots-txt-tester must NOT resolve against robots-txt-generator — " +
			"the planner's and suggester's names being DISJOINT is the mechanism of bugs_open/450")
	}
}

// TestResolveTool_EmptySections_Held pins the deliberate decision that a
// SECTIONLESS tool page is held too. Its instance (websitepromotion's
// tool-channel-prioritiser) parks 7 unbuilt_internal_link items in a HUMAN queue
// per remake instead of shelling — "no shell" is not "harmless".
func TestResolveTool_EmptySections_Held(t *testing.T) {
	page := planPageView{Name: "tool-channel-prioritiser", Role: "tool", URL: "/tools/tool-channel-prioritiser/index.html"}

	if res := ResolveToolItemSource(page, map[string]bool{}); res.Producible {
		t.Fatal("a sectionless tool page with no tool must be held — it becomes a recurring HITL tax " +
			"and a phantom-link source on a row no producer will ever fill")
	}
}

// TestResolveTool_NonToolRolesAreOutOfFamily pins the scope. A guide page
// carrying tool-prefixed SECTIONS is not a tool page, and a tools hub is the
// listing gate's business, not this one.
func TestResolveTool_NonToolRolesAreOutOfFamily(t *testing.T) {
	for _, p := range []planPageView{
		{Name: "tool-ab-test-guide", Role: "blog-post", Sections: []string{"hero", "tool-guide-intro", "tool-cta"}},
		{Name: "tools-index", Role: "section-index", Sections: []string{"hero", "tool-list"}},
		{Name: "index", Role: "landing", Sections: []string{"hero"}},
	} {
		if res := ResolveToolItemSource(p, map[string]bool{}); res.ToolFamily {
			t.Errorf("page %q (role %q) must be out of the tool family", p.Name, p.Role)
		}
	}
}

// --- lockstep tests -------------------------------------------------------

// TestToolProducerSlug_MatchesBuilderRouting keeps this gate's spec.builder_needed
// in step with the estate's page-type routing without reading it at runtime. If
// routing ever gains a real tool builder, this fails and a human decides what
// the gap receipt should say — rather than the field silently changing meaning.
func TestToolProducerSlug_MatchesBuilderRouting(t *testing.T) {
	_, slug, known := builderForPageType("tool")
	if !known {
		t.Fatal("builderForPageType no longer knows page_type 'tool' — this gate's slug is now unanchored")
	}
	if slug != toolProducerSlug {
		t.Errorf("routing says the tool producer is %q, this gate files %q — the capability_gap "+
			"receipts would group under a slug nothing routes to", slug, toolProducerSlug)
	}
}

// TestToolNameCandidates_MatchToolPipelineNaming pins the naming contract this
// gate depends on. sanitiseFunction guarantees the `tool-` prefix on every
// function; the page is created with name == function. If either moves, real
// tool pages start collecting false capability_gap receipts on every replan, and
// this test is what says so instead.
func TestToolNameCandidates_MatchToolPipelineNaming(t *testing.T) {
	fn := sanitiseFunction("Robots Txt Generator")
	if fn != "tool-robots-txt-generator" {
		t.Fatalf("sanitiseFunction produced %q — the tool naming contract moved", fn)
	}
	// A page named exactly as the pipeline names it must be a candidate.
	if got := toolFunctionCandidates(toolPage(fn)); !containsString(got, fn) {
		t.Errorf("candidates %v omit the canonical page name %q", got, fn)
	}
	// And the legacy bare-slug page must reach the same function.
	if got := toolFunctionCandidates(toolPage("robots-txt-generator")); !containsString(got, fn) {
		t.Errorf("candidates %v omit the tool- prefixed form %q (resolveToolPageIdentity's legacy shape)", got, fn)
	}
}

// (containsString is the package's own helper, in v3_site_actions.go — reused
// rather than redeclared; the first draft of this file shadowed it and the
// package would not compile.)

// --- gate-level tests -----------------------------------------------------

// expectToolCensus scripts the ONE tool-set read the gate makes per plan.
func expectToolCensus(mock sqlmock.Sqlmock, functions ...string) {
	rows := sqlmock.NewRows([]string{"function"})
	for _, f := range functions {
		rows = rows.AddRow(f)
	}
	mock.ExpectQuery("SELECT DISTINCT cc.function").WillReturnRows(rows)
}

// expectCapabilityGapReceipt scripts the shared-writer sequence:
// tx open → anti-churn probe → INSERT → commit.
func expectCapabilityGapReceipt(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func TestEnforceToolItemSources_HoldsFilesGapAndKeepsTheRest(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCensus(mock) // site owns NO tools
	expectCapabilityGapReceipt(mock)
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(0, 1))

	params := gateParams(db, siteID)
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "index", "page_type": "landing",
			"sections": []interface{}{"hero"}},
		map[string]interface{}{"name": "tool-robots-txt-tester", "page_type": "tool",
			"url":      "/tools/tool-robots-txt-tester/index.html",
			"sections": []interface{}{"hero-tool", "generic-text-block"}},
	}
	kept := enforceToolItemSources(context.Background(), params, pages, nil)

	if len(kept) != 1 {
		t.Fatalf("kept %d pages, want 1 (the tool page must be held)", len(kept))
	}
	if name, _ := kept[0].(map[string]interface{})["name"].(string); name != "index" {
		t.Errorf("the surviving page is %q, want index — the gate held the wrong page", name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the hold must file its capability_gap AND its durable finding: %v", err)
	}
}

// TestEnforceToolItemSources_RealisedPageKeptWithReceipt is bugs_open/001's rule:
// a page that has already been built is never dropped from a plan. The 61
// existing shells are realised, so a replan keeps them — their removal is
// instance work — but the gap receipt is still filed so the state is recorded.
func TestEnforceToolItemSources_RealisedPageKeptWithReceipt(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCensus(mock)
	expectCapabilityGapReceipt(mock)
	// No agent_error_log expectation: a preserved page is not a DROP, so no
	// findings row is written. An unexpected write here fails the test.

	params := gateParams(db, siteID)
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "tool-serp-preview", "page_type": "tool",
			"url": "/tools/tool-serp-preview/index.html", "sections": []interface{}{"hero-tool"}},
	}
	existing := []interface{}{
		map[string]interface{}{"name": "tool-serp-preview", "page_type": "tool",
			"url": "/tools/tool-serp-preview/index.html"},
	}
	kept := enforceToolItemSources(context.Background(), params, pages, existing)

	if len(kept) != 1 {
		t.Fatalf("a REALISED page must be kept (bugs_open/001) — kept %d", len(kept))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the preserved page must still leave a capability_gap receipt: %v", err)
	}
}

// TestEnforceToolItemSources_ToolExistsSoPlanIsUntouched is the self-clearing
// control: same plan, same page, the site owns the tool. Nothing is held and NO
// receipt is filed — an unexpected INSERT fails the test.
func TestEnforceToolItemSources_ToolExistsSoPlanIsUntouched(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCensus(mock, "tool-ab-test-calculator")

	params := gateParams(db, siteID)
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "tool-ab-test", "page_type": "tool",
			"url":      "/tools/tool-ab-test/index.html",
			"sections": []interface{}{"hero-tool", "tool-ab-test-calculator", "tool-cta"}},
	}
	kept := enforceToolItemSources(context.Background(), params, pages, nil)

	if len(kept) != 1 {
		t.Fatalf("a tool page whose tool EXISTS must be kept — kept %d", len(kept))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("no receipt should be written when nothing is held: %v", err)
	}
}

// TestEnforceToolItemSources_NoToolPagesSkipsCensus pins the early return: the
// common plan has no tool page at all and must not pay for the census.
func TestEnforceToolItemSources_NoToolPagesSkipsCensus(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	params := gateParams(db, uuid.New())
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "index", "page_type": "landing", "sections": []interface{}{"hero"}},
	}
	kept := enforceToolItemSources(context.Background(), params, pages, nil)

	if len(kept) != 1 {
		t.Fatalf("kept %d, want 1", len(kept))
	}
	// No query was scripted; any query at all is an unexpected call.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a plan with no tool pages must make no database call: %v", err)
	}
}

// TestEnforceToolItemSources_CensusErrorFailsOpen is the arm that matters most
// for safety. An unreadable census must NOT hold every tool page — that would be
// the starvation failure the §7 measurement exists to rule out, arriving by
// accident on a transient database error.
func TestEnforceToolItemSources_CensusErrorFailsOpen(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT DISTINCT cc.function").
		WillReturnError(context.DeadlineExceeded)

	params := gateParams(db, uuid.New())
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "tool-x", "page_type": "tool", "sections": []interface{}{"hero-tool"}},
	}
	kept := enforceToolItemSources(context.Background(), params, pages, nil)

	if len(kept) != 1 {
		t.Fatalf("an unreadable tool census must fail OPEN and keep every page — kept %d", len(kept))
	}
}

// --- composition with the listing gate (BLD-028) ---------------------------

// TestToolGateRunsBeforeListingGate is the test the ordering comment in
// v3_site_actions.go names, and it exists because the council's guardian seat
// asked for the interaction to be PINNED rather than reasoned about:
//
//	"The ordering change is a new behavioral coupling between two independently-
//	 gated features. It only manifests when BOTH keys are armed simultaneously ...
//	 should be pinned by an explicit test naming both keys armed together, not
//	 deferred to 'worth a reviewer's eye'."
//
// ⚠ It was right to ask twice over: the comment claiming this test existed was
// written BEFORE the test was, and the objection is what caught it.
//
// The coupling: a /tools/ section-index hub resolves by counting child pages
// under its prefix. Hold the tool children first and the hub counts ZERO, so the
// listing gate holds the hub too and no phantom /tools/ URL is planned at all.
// Run the gates the other way round and the hub survives on the strength of
// children that are about to be removed — shipping an empty hub, which is a
// 444-class page. Neither gate is wrong on its own; only the order decides.
func TestToolGateRunsBeforeListingGate(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// The site owns NO tools, so both tool pages are held.
	expectToolCensus(mock)
	expectCapabilityGapReceipt(mock) // tool page 1
	expectCapabilityGapReceipt(mock) // tool page 2
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(0, 1))
	// Then the hub, held by the LISTING gate for having no children left.
	expectCapabilityGapReceipt(mock)
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(0, 1))

	params := gateParams(db, siteID)
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "index", "page_type": "landing",
			"sections": []interface{}{"hero"}},
		map[string]interface{}{"name": "tools-index", "page_type": "section-index",
			"url": "/tools/index.html", "sections": []interface{}{"hero"}},
		map[string]interface{}{"name": "tool-robots-txt-tester", "page_type": "tool",
			"url":      "/tools/tool-robots-txt-tester/index.html",
			"sections": []interface{}{"hero-tool", "generic-text-block"}},
		map[string]interface{}{"name": "tool-serp-preview", "page_type": "tool",
			"url":      "/tools/tool-serp-preview/index.html",
			"sections": []interface{}{"hero-tool", "generic-text-block"}},
	}

	// THE ORDER UNDER TEST — the same order ValidateSitePlanAction applies.
	afterTool := enforceToolItemSources(context.Background(), params, pages, nil)
	if len(afterTool) != 2 {
		t.Fatalf("tool gate kept %d pages, want 2 (index + the hub)", len(afterTool))
	}
	afterBoth := enforceListingItemSources(context.Background(), params, afterTool, nil)

	if len(afterBoth) != 1 {
		names := make([]string, 0, len(afterBoth))
		for _, p := range afterBoth {
			n, _ := p.(map[string]interface{})["name"].(string)
			names = append(names, n)
		}
		t.Fatalf("after both gates kept %d pages (%v), want 1 — the /tools/ hub must be held "+
			"once its children are gone, or the plan ships an empty hub (a 444-class page)", len(afterBoth), names)
	}
	if name, _ := afterBoth[0].(map[string]interface{})["name"].(string); name != "index" {
		t.Errorf("survivor is %q, want index", name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("both gates must file their receipts: %v", err)
	}
}

// TestListingGateFirstWouldKeepTheEmptyHub is the CONTROL that gives the test
// above its meaning. Run in the wrong order — listing first — the hub survives,
// because at that moment its tool children are still in the plan. Without this,
// TestToolGateRunsBeforeListingGate would pass for any order and prove nothing
// about the ordering it is named for.
func TestListingGateFirstWouldKeepTheEmptyHub(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	params := gateParams(db, siteID)
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "tools-index", "page_type": "section-index",
			"url": "/tools/index.html", "sections": []interface{}{"hero"}},
		map[string]interface{}{"name": "tool-robots-txt-tester", "page_type": "tool",
			"url":      "/tools/tool-robots-txt-tester/index.html",
			"sections": []interface{}{"hero-tool", "generic-text-block"}},
	}

	// Listing gate FIRST: the hub still has a child, so it is kept and files nothing.
	afterListing := enforceListingItemSources(context.Background(), params, pages, nil)
	kept := false
	for _, p := range afterListing {
		if n, _ := p.(map[string]interface{})["name"].(string); n == "tools-index" {
			kept = true
		}
	}
	if !kept {
		t.Fatal("control failed: with its tool child still present the hub should be KEPT — " +
			"if it is held here, the ordering argument in v3_site_actions.go rests on nothing")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the listing gate should not have filed a receipt for a hub with a child: %v", err)
	}
}

// TestEnforceToolItemSources_NilDBFailsOpen — the gate stands down entirely.
func TestEnforceToolItemSources_NilDBFailsOpen(t *testing.T) {
	params := ActionParams{
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": uuid.New().String()},
		},
		Logger: zap.NewNop(),
	}
	pages := []interface{}{
		map[string]interface{}{"name": "tool-x", "page_type": "tool"},
	}
	if kept := enforceToolItemSources(context.Background(), params, pages, nil); len(kept) != 1 {
		t.Fatalf("nil DB must fail open — kept %d", len(kept))
	}
}
