// FILE: platform/orchestration/actions/queue_evidence_base_rerenders_test.go
//
// bugs_open/427: queueEvidenceBasePageRerenders is the propagation half of
// "dates get corrected, not just added" — a human's edit to a fact (via the
// stale_evidence item) or a daily re-verification changes the register, and
// this is what tells a query.upcoming_events-consuming page to re-resolve.
// Cousin of queueNewsPageRerenders/queueDirectoryPageRerenders; these tests
// follow their shape rather than re-deriving it.

package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TestQueueEvidenceBasePageRerendersNoConsumersIsZero — the consumer lookup
// runs against DepEvidenceBase and an empty result queues nothing (no site in
// the fleet has an upcoming_events-consuming component yet, so this is the
// live case today).
func TestQueueEvidenceBasePageRerendersNoConsumersIsZero(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "component", "input_schema"}))

	queued := queueEvidenceBasePageRerenders(context.Background(), db, siteID, zap.NewNop())
	if queued != 0 {
		t.Errorf("queued = %d, want 0 for an empty consumer set", queued)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestQueueEvidenceBasePageRerendersQueuesForAConsumerPage is the positive
// control — without it, the zero-consumer test above would pass just as well
// if the function queued nothing under every circumstance.
func TestQueueEvidenceBasePageRerendersQueuesForAConsumerPage(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	upcomingEventsSchema := `{"fields":{"items":{"type":"array","source":"query.upcoming_events"}}}`

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "component", "input_schema"}).
			AddRow(pageID, "tool-fight-calendar", "/tools/fight-calendar/index.html", "boxingonline.com", "event-list", upcomingEventsSchema))

	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "refresh_evidence_base", "low", sqlmock.AnyArg(), pageID,
			specArgWith{`"reason":"section_data_resolved"`, `"page_name":"tool-fight-calendar"`, `"domain":"boxingonline.com"`},
			pageRerenderItemKey("tool-fight-calendar", siteID, "section_data_resolved"),
			sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	queued := queueEvidenceBasePageRerenders(context.Background(), db, siteID, zap.NewNop())
	if queued != 1 {
		t.Errorf("queued = %d, want 1", queued)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestQueueEvidenceBasePageRerendersDedupsAnOpenItem — a second call while an
// earlier reasoned item is still open (ON CONFLICT DO NOTHING, 0 rows
// affected) must not be double-counted as queued.
func TestQueueEvidenceBasePageRerendersDedupsAnOpenItem(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	upcomingEventsSchema := `{"fields":{"items":{"type":"array","source":"query.upcoming_events"}}}`

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "component", "input_schema"}).
			AddRow(pageID, "tool-fight-calendar", "/tools/fight-calendar/index.html", "boxingonline.com", "event-list", upcomingEventsSchema))

	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))

	queued := queueEvidenceBasePageRerenders(context.Background(), db, siteID, zap.NewNop())
	if queued != 0 {
		t.Errorf("queued = %d, want 0 for a dedup-suppressed insert", queued)
	}
}

// TestQueueEvidenceBasePageRerendersNilDBIsZero mirrors the other two
// producers' nil-safety (both check db == nil before querying).
func TestQueueEvidenceBasePageRerendersNilDBIsZero(t *testing.T) {
	if got := queueEvidenceBasePageRerenders(context.Background(), nil, uuid.New(), zap.NewNop()); got != 0 {
		t.Errorf("queued = %d, want 0 for a nil db", got)
	}
}
