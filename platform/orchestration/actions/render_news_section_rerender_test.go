// FILE: platform/orchestration/actions/render_news_section_rerender_test.go
//
// queueNewsPageRerenders must not resurrect retired pages (bugs_open/098).
//
// WHY THIS TEST IS SHAPED LIKE THIS. The defect was an ABSENT predicate, and an
// absent predicate has no output of its own to assert — the function returns a
// count of queued items either way, and on a site with no archived news page it
// returns the identical count with and without the fix. So the assertion is on
// the STATEMENT: the query must carry a p.status filter. sqlmock matches the SQL
// it is given, so deleting the filter fails this test, which is the only thing
// that distinguishes "the guard is there" from "the guard was never needed here".
//
// The live evidence this exists for: robot-hands.com/learning-center-index is
// status='archived', build_status='deployed', and was re-rendered and
// re-committed to the deploy repo twice a day (six page_rerender items raised by
// this function between 2026-07-31 and 2026-08-03). It also made a retraction
// self-undoing — delete the file and the next news refresh republishes it.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TestQueueNewsPageRerendersFiltersOnPageStatus is the regression guard.
// The artifact-liveness predicate and `status = 'active'` answer different
// questions and nothing keeps them in step: archiving sets status and leaves
// the build columns alone, so a selector keyed on artifact liveness alone
// keeps choosing a retired page for ever.
//
// UPDATED 2026-08-03 (bugs_open/185 tranche 2, same day as this file was
// written): the artifact-liveness half changed spelling from
// `p.build_status = 'deployed'` to NOT(datahelpers.NeverDeployedPagePredicate),
// which additionally covers a needs_rebuild page still serving its previous
// artefact. The regex pins the NEW spelling's distinctive prefix
// (`NOT (p.deployed_at IS NULL`) so a revert to the narrow form — or the loss
// of the status filter — both still fail here. The INTENT is unchanged from
// this test's original: BOTH predicates present, or an archived page gets
// re-rendered and re-published (bugs_open/098).
func TestQueueNewsPageRerendersFiltersOnPageStatus(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	// The expectation carries BOTH predicates. Remove either from the
	// production query and this fails.
	//
	// > RELOCATED 2026-08-25 (RFC_052). The SQL these two tests guard is no
	// > longer written in render_news_section_html.go — that emitter now calls
	// > queryresolve.ConsumerPages, and the predicates live in
	// > queryresolve.consumerSQL. THE TEST STAYS HERE ANYWAY, driving the real
	// > emitter, because what bugs_open/098 cost was an archived page being
	// > re-rendered BY THIS FUNCTION; a guard that only checks the shared helper
	// > would still pass if this call site were rewired to something weaker.
	// > queryresolve/consumers_dependency_test.go asserts the same thing at the
	// > destination — both, deliberately, because a relocating fix that is only
	// > measured at its new home leaves the old one reading green either way.
	// >
	// > The status spelling CHANGED with the move: `IN ('active','deployed')`
	// > rather than `= 'active'`, because the shared lookup deliberately mirrors
	// > the resolvers' own set (a consumer page is chosen by the rule its items
	// > are chosen by). Same behaviour today — `pages.status` holds only
	// > `active` and `archived` — and `archived` is excluded either way, which
	// > is the whole point of the test.
	// Order-agnostic on purpose: BOTH predicates must be present, in either
	// order. The shared lookup happens to emit status first; pinning that would
	// make a harmless reorder look like the bugs_open/098 regression.
	mock.ExpectQuery(`(NOT \(p\.deployed_at IS NULL[\s\S]*p\.status IN \('active', 'deployed'\)|p\.status IN \('active', 'deployed'\)[\s\S]*NOT \(p\.deployed_at IS NULL)`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "component", "input_schema"}))

	queued := queueNewsPageRerenders(context.Background(), db, siteID, zap.NewNop())
	if queued != 0 {
		t.Errorf("queued = %d, want 0 for an empty result set", queued)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the news requeue query no longer filters on BOTH build_status and status — "+
			"an archived page will be re-rendered and re-published (bugs_open/098): %v", err)
	}
}

// TestQueueNewsPageRerendersOrderIndependence — the same assertion with the two
// predicates in the opposite order, so the test pins the PRESENCE of the status
// filter rather than one particular formatting of the WHERE clause. Without
// this, a harmless reorder would look like a regression.
func TestQueueNewsPageRerendersOrderIndependence(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()
	mock.ExpectQuery(`p\.status IN \('active', 'deployed'\)`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "component", "input_schema"}))

	_ = queueNewsPageRerenders(context.Background(), db, siteID, zap.NewNop())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no p.status filter in the news requeue query: %v", err)
	}
}

// TestQueueDirectoryPageRerendersFiltersOnPageStatus — the cousin. Its own
// comment calls it "cousin of queueNewsPageRerenders", which is exactly how a
// reader concludes the two behave alike; they did not, and this keeps them
// honest. Latent rather than live (0 non-active pages carry a directory
// component fleet-wide, measured 2026-08-03), so this test is the only thing
// that would catch the filter being dropped again.
func TestQueueDirectoryPageRerendersFiltersOnPageStatus(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()
	// RELOCATED 2026-08-25 with the news cousin (RFC_052) — see the note there
	// for why the test stays at this call site as well as at the destination.
	// The lookup is now dependency-scoped, so the profile's SOURCES are what
	// select pages and the component names no longer reach the query: the
	// WithArgs is siteID alone.
	mock.ExpectQuery(`(NOT \(p\.deployed_at IS NULL[\s\S]*p\.status IN \('active', 'deployed'\)|p\.status IN \('active', 'deployed'\)[\s\S]*NOT \(p\.deployed_at IS NULL)`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "component", "input_schema"}))

	queued := queueDirectoryPageRerenders(context.Background(), db, siteID,
		directoryPublishProfile{
			SnippetComponent: "model-directory", ListingComponent: "model-directory-listing",
			SnippetSource: "model_directory", ListingSource: "model_directory_full",
		},
		zap.NewNop())
	if queued != 0 {
		t.Errorf("queued = %d, want 0 for an empty result set", queued)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the directory requeue query no longer filters on BOTH build_status and status — "+
			"it will re-render and re-publish archived pages, exactly as its news cousin did: %v", err)
	}
}
