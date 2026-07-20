package actions

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Covers the anti-churn block in insertWorkItem (the "two-strike rule").
//
// The rule reads a repeat item_key as "we keep fixing this and it keeps coming
// back". That holds for a DETECTED DEFECT. It is wrong for an ACTION REQUEST,
// where a terminal predecessor means the request SUCCEEDED — and where the
// consequences were severe: two successful re-renders poisoned a shared
// item_key, so later re-renders were born 'unresolved' and never dispatched,
// and durable template fixes never reached the live page (bugs_open/024).
//
// workItem.recurrenceExpected opts an item out of the heuristics (never out of
// dedup, which the unique index still enforces).

func newInsertMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return db, mock
}

// expectInsertWithSummaryAndStatus asserts the row is written with a specific
// summary ($6) and status ($12) — the two fields the anti-churn block rewrites.
// Everything else is positional noise for this test's purposes.
func expectInsertWithSummaryAndStatus(mock sqlmock.Sqlmock, summary, status string) {
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			summary, // $6
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			status, // $12
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func baseItem() workItem {
	return workItem{
		siteID:    uuid.New(),
		source:    "tool-improver",
		pipeline:  "build",
		itemType:  "needs_rerender",
		severity:  "medium",
		summary:   "Re-render page after tool improvement",
		spec:      "{}",
		status:    "triaged",
		createdBy: "tool-improver",
		itemKey:   "rerender_tool_fix_gamesdesign.co.uk",
	}
}

// A recurrence-expected item must NOT run the anti-churn probe at all. If it
// did, two prior COMPLETE (i.e. successful) re-renders would either suppress
// the insert outright or brand the row 'unresolved' at birth.
func TestInsertWorkItem_RecurrenceExpected_SkipsAntiChurnEntirely(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	// Deliberately NO ExpectQuery for the anti-churn probe: if insertWorkItem
	// issues it, sqlmock fails the test with "call to Query was not expected".
	// The row must land dispatchable and with its summary untouched.
	expectInsertWithSummaryAndStatus(mock,
		"Re-render page after tool improvement", "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	item := baseItem()
	item.recurrenceExpected = true

	inserted, err := insertWorkItem(context.Background(), tx, item, zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if !inserted {
		t.Fatal("recurrence-expected item was suppressed; it must always be inserted")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The regression itself: two COMPLETE predecessors are two SUCCESSES, but the
// rule counts them as failed attempts. Without the opt-in the row is born
// 'unresolved' — non-dispatchable — which is how the fix loop silently died.
func TestInsertWorkItem_TwoCompletedPredecessors_PoisonTheKey(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	// 2 terminal rows, newest 5h old (past the 3h within-cycle suppression).
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(2, 5.0))
	// The row is branded at birth: status 'unresolved' (non-dispatchable) and a
	// summary claiming two failed attempts — when both predecessors SUCCEEDED.
	expectInsertWithSummaryAndStatus(mock,
		"[unresolved after 2 attempts] Re-render page after tool improvement",
		"unresolved")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	inserted, err := insertWorkItem(context.Background(), tx, baseItem(), zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if !inserted {
		t.Fatal("expected the item to be inserted (labelled), not suppressed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Within-cycle suppression drops the item ENTIRELY (returns false, no row).
// For an action request that is the most damaging branch: a fix lands, its
// re-render completes, a second fix lands an hour later — and the re-render
// that would deploy it is silently discarded.
func TestInsertWorkItem_WithinCycleSuppression_DropsItem(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	// 1 terminal row, 1h old — inside the 3h window.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(1, 1.0))
	// No INSERT expected: the item is suppressed.

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	inserted, err := insertWorkItem(context.Background(), tx, baseItem(), zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if inserted {
		t.Fatal("expected within-cycle suppression to drop the item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Same suppression window, but recurrence-expected: the item must survive.
func TestInsertWorkItem_RecurrenceExpected_SurvivesWithinCycleWindow(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	item := baseItem()
	item.recurrenceExpected = true

	inserted, err := insertWorkItem(context.Background(), tx, item, zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if !inserted {
		t.Fatal("recurrence-expected item must not be suppressed within the cycle window")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Default behaviour is unchanged for detected defects with a clean history.
func TestInsertWorkItem_NoPriorTerminalItems_InsertsNormally(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	item := baseItem()
	item.itemType = "empty_section" // a genuine detected defect
	item.recurrenceExpected = false

	inserted, err := insertWorkItem(context.Background(), tx, item, zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if !inserted {
		t.Fatal("expected a clean-history item to insert")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// An item with no item_key never enters the anti-churn block at all.
func TestInsertWorkItem_NoItemKey_SkipsAntiChurn(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	item := baseItem()
	item.itemKey = ""

	inserted, err := insertWorkItem(context.Background(), tx, item, zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if !inserted {
		t.Fatal("expected keyless item to insert")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
