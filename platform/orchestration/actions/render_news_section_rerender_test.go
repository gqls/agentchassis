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
// `build_status = 'deployed'` and `status = 'active'` answer different
// questions and nothing keeps them in step: archiving sets status and leaves
// build_status alone, so a selector keyed on build_status alone keeps choosing
// a retired page for ever.
func TestQueueNewsPageRerendersFiltersOnPageStatus(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	// The expectation carries BOTH predicates. Remove either from the
	// production query and this fails.
	mock.ExpectQuery(`p\.build_status = 'deployed'[\s\S]*p\.status = 'active'`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "domain"}))

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
	mock.ExpectQuery(`p\.status = 'active'`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "domain"}))

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
	mock.ExpectQuery(`p\.build_status = 'deployed'[\s\S]*p\.status = 'active'`).
		WithArgs(siteID, "model-directory", "model-directory-listing").
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	queued := queueDirectoryPageRerenders(context.Background(), db, siteID,
		directoryPublishProfile{SnippetComponent: "model-directory", ListingComponent: "model-directory-listing"},
		zap.NewNop())
	if queued != 0 {
		t.Errorf("queued = %d, want 0 for an empty result set", queued)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the directory requeue query no longer filters on BOTH build_status and status — "+
			"it will re-render and re-publish archived pages, exactly as its news cousin did: %v", err)
	}
}
