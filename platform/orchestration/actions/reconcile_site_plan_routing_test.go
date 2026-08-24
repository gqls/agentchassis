// FILE: platform/orchestration/actions/reconcile_site_plan_routing_test.go
//
// bugs_open/206 closure hardening — the PRODUCER half.
//
// builder_routing_test.go pins the pure decision. These tests pin that
// reconcile_site_plan actually ASKS it, which is the half that was broken:
// the emit hardcoded 'page-build-handler' for every page and minted five
// doomed items across three sites (measured 2026-08-24), every one no-op'ing
// with "no sections ready to build", burning an attempt and parking
// needs_human_review — including garden-tools.uk's brand-directory-index,
// an entity-directory page whose builder had been live since 08-08.
//
// Each test asserts the mechanism's EFFECT — the INSERT's own arguments —
// never the absence of a call. This package has a live landmine on exactly
// that: an assertion written as "the query is not issued" passes vacuously
// because the error sqlmock raises is swallowed downstream. Here a WithArgs
// mismatch makes the Exec fail, which fails the action, which fails the test
// (each is mutation-proven; see the per-test comments).

package actions

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// reconcileRoutingFixture drives ReconcileSitePlanAction over exactly ONE plan
// page, with every query it makes registered, so the only interesting output
// is the work-item INSERT the emit chooses.
//
// pageType is what `pages` holds for the page; role is what the plan holds.
// buildStatus 'planned' makes decideEmit return "not_built" — the candidate
// path — which is the state all five real parked items were in.
type reconcileRoutingFixture struct {
	pageName    string
	role        string
	pageType    string
	buildStatus string
}

func (f reconcileRoutingFixture) run(t *testing.T, expectInsert func(sqlmock.Sqlmock)) error {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	planID := uuid.New()
	mock.MatchExpectationsInOrder(false)

	// 1. resolve current plan
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM site_plans")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))

	// 2. plan pages / realised pages / plan sections / open items
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_plan_pages")).
		WillReturnRows(sqlmock.NewRows([]string{"name", "role", "title", "nav_order"}).
			AddRow(f.pageName, f.role, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WillReturnRows(sqlmock.NewRows([]string{"name", "build_status", "built_from_plan_version", "rebuild_policy", "page_type"}).
			AddRow(f.pageName, f.buildStatus, nil, "generic", f.pageType))
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_plan_sections")).
		WillReturnRows(sqlmock.NewRows([]string{"page_name", "component_name"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT item_key FROM site_work_items")).
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))

	// 3. the transaction: the INSERT under test, then the trailing writes.
	// The needs_rerender item is emitted only when a page item was, so it is
	// registered unconditionally and simply goes unused on the gap/review
	// paths — ExpectationsWereMet is deliberately not asserted (an unused
	// expectation here is correct, not a finding; the assertion is the
	// pinned argument on the INSERT under test).
	mock.ExpectBegin()
	expectInsert(mock)
	mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'needs_rerender'")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sites SET last_reconciled_at")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err = ReconcileSitePlanAction(context.Background(), ActionParams{
		DB:     db,
		Logger: zap.NewNop(),
		CollectedData: map[string]interface{}{
			"target_site_id": siteID.String(),
		},
		StepConfig:       models.Step{},
		ExecutionContext: &orchtypes.ExecutionContext{Action: "execute"},
	})
	return err
}

// reconcileInsertArgsPinning matches the 13-column needs_page / capability_gap
// INSERT while pinning only the named positions, so the test breaks on a
// routing change and not on every unrelated column addition.
//
// Column order, from the emit: site_id, summary, spec, priority,
// handler_agent, item_key, batch_id — the literals (source, pipeline,
// item_type, severity, status, created_by) are inline in the SQL, so a change
// to THOSE shows up as a query-text mismatch instead.
func reconcileInsertArgs(cols int, pin map[int]driver.Value) []driver.Value {
	args := make([]driver.Value, cols)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	for idx, v := range pin {
		args[idx] = v
	}
	return args
}

// An entity-directory page must reach directory-build-handler. This is
// garden-tools.uk's brand-directory-index, parked since 2026-08-23 on the
// hardcoded handler while its builder was live.
//
// Mutation-proven: restoring the hardcoded 'page-build-handler' in the emit
// makes the handler_agent argument mismatch, the Exec fail, and this test fail.
func TestReconcileRoutesEntityDirectoryToItsBuilder(t *testing.T) {
	f := reconcileRoutingFixture{
		pageName: "brand-directory-index", role: "entity-directory",
		pageType: "entity-directory", buildStatus: "planned",
	}
	err := f.run(t, func(mock sqlmock.Sqlmock) {
		// $5 is handler_agent (0-based index 4) in the needs_page INSERT.
		mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'needs_page'")).
			WithArgs(reconcileInsertArgs(7, map[int]driver.Value{4: "directory-build-handler"})...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	})
	if err != nil {
		t.Fatalf("reconcile failed — the needs_page INSERT did not carry "+
			"handler_agent='directory-build-handler', so an entity-directory page is "+
			"still being routed to the generic builder that no-ops on it: %v", err)
	}
}

// A section-index page must reach the same builder — the route two live hand
// re-routes proved (guides-index 2026-08-08, practice 2026-08-24) before any
// map knew it. loanzy.uk's guides-index is the parked real case.
func TestReconcileRoutesSectionIndexToItsBuilder(t *testing.T) {
	f := reconcileRoutingFixture{
		pageName: "guides-index", role: "section-index",
		pageType: "section-index", buildStatus: "planned",
	}
	err := f.run(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'needs_page'")).
			WithArgs(reconcileInsertArgs(7, map[int]driver.Value{4: "directory-build-handler"})...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	})
	if err != nil {
		t.Fatalf("reconcile failed — section-index did not route to directory-build-handler: %v", err)
	}
}

// A page whose type has NO builder must produce a deferred capability_gap and
// NOT a needs_page. Asserted by the gap INSERT's own arguments: if the emit
// fell through to the needs_page branch instead, that query text is not
// registered, the Exec errors, and the action fails.
//
// The real case is dartsonline.com's brand-detail, which burned an attempt and
// sat in needs_human_review from 2026-07-20 to at least 2026-08-24.
func TestReconcileFilesCapabilityGapForUnbuildableType(t *testing.T) {
	f := reconcileRoutingFixture{
		pageName: "brand-detail", role: "entity-page",
		pageType: "entity-page", buildStatus: "planned",
	}
	err := f.run(t, func(mock sqlmock.Sqlmock) {
		// $4 is item_key (0-based 3) in the gap INSERT — handler_agent is an
		// inline '' literal, NOT a bind parameter, so a regression that
		// re-introduces a bound handler_agent changes the arity and fails
		// here (5 args expected, 6 supplied). Round-2 bug_historian HIGH.
		mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'capability_gap'")).
			WithArgs(reconcileInsertArgs(5, map[int]driver.Value{
				3: "capability_gap:entity-page:brand-detail",
			})...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	})
	if err != nil {
		t.Fatalf("reconcile failed — an entity-page did not produce a deferred "+
			"capability_gap naming entity-page-builder; it is still minting a "+
			"needs_page that cannot succeed: %v", err)
	}
}

// The page-ownership guard must stay AHEAD of type routing: a tool-ROLE page
// emits owned_page_review (no handler) even though its page_type would
// otherwise route. Routing a tool page to a generic builder is the vonc arena
// clobber (TP-004) — the guard this change must not reorder.
func TestReconcileOwnedRoleGuardOutranksTypeRouting(t *testing.T) {
	f := reconcileRoutingFixture{
		pageName: "tool-arena", role: "tool",
		pageType: "section-index", buildStatus: "planned",
	}
	err := f.run(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'owned_page_review'")).
			WithArgs(reconcileInsertArgs(5, map[int]driver.Value{
				3: "owned_page_review:tool-arena",
			})...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	})
	if err != nil {
		t.Fatalf("reconcile failed — a tool-role page no longer emits owned_page_review; "+
			"type routing has been moved ahead of the ownership guard: %v", err)
	}
}

// Guards the fallback that makes the routing work for a page with no `pages`
// row yet (decideEmit "missing"): the plan's role carries the same vocabulary,
// measured 2026-08-24 across every current-plan page. Without the fallback such
// a page routes as an unknown type to the generic builder — the original bug,
// re-created for exactly the pages that have never been built.
func TestReconcileFallsBackToPlanRoleWhenNoPagesRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	planID := uuid.New()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM site_plans")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_plan_pages")).
		WillReturnRows(sqlmock.NewRows([]string{"name", "role", "title", "nav_order"}).
			AddRow("brand-directory-index", "entity-directory", nil, nil))
	// No pages row at all — the "missing" case.
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WillReturnRows(sqlmock.NewRows([]string{"name", "build_status", "built_from_plan_version", "rebuild_policy", "page_type"}))
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_plan_sections")).
		WillReturnRows(sqlmock.NewRows([]string{"page_name", "component_name"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT item_key FROM site_work_items")).
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'needs_page'")).
		WithArgs(reconcileInsertArgs(7, map[int]driver.Value{4: "directory-build-handler"})...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'needs_rerender'")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sites SET last_reconciled_at")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if _, err := ReconcileSitePlanAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		CollectedData:    map[string]interface{}{"target_site_id": siteID.String()},
		StepConfig:       models.Step{},
		ExecutionContext: &orchtypes.ExecutionContext{Action: "execute"},
	}); err != nil {
		t.Fatalf("reconcile failed — a page with no pages row did not route on its plan role, "+
			"so a never-built typed page still goes to the generic builder: %v", err)
	}
}

// A sanity check that the fixture's own plumbing cannot mask a routing change:
// an ordinary content page must still route to page-build-handler. If this
// fails while the others pass, the extraction changed an existing route.
func TestReconcileLeavesOrdinaryContentPagesOnTheGenericBuilder(t *testing.T) {
	f := reconcileRoutingFixture{
		pageName: "about", role: "content",
		pageType: "content", buildStatus: "planned",
	}
	err := f.run(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(regexp.QuoteMeta("'reconcile_site_plan', 'build', 'needs_page'")).
			WithArgs(reconcileInsertArgs(7, map[int]driver.Value{4: "page-build-handler"})...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	})
	if err != nil {
		t.Fatalf("reconcile failed — an ordinary content page no longer routes to "+
			"page-build-handler; the extraction changed an existing route: %v", err)
	}
}

var _ = sql.ErrNoRows // keep the database/sql import honest if assertions change
