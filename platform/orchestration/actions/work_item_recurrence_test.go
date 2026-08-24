package actions

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

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
//
// The arg COUNT is load-bearing even though most args are AnyArg: sqlmock fails
// on a length mismatch. That is why insertWorkItem appends parent_item_id ($17)
// only for the callers that set one — an unconditional column failed twenty
// tests in eight files that had nothing to do with the change
// (bugs_open/091 candidate 1, 2026-08-02).
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

// Within-cycle: a fix lands, its re-render completes, a second fix lands an
// hour later. The brake used to drop that second request ENTIRELY — no row, no
// error, Inserted:false, indistinguishable from a genuine dedup.
//
// REWRITTEN 2026-08-24 (bugs_open/326 option D, owner ruling "D + E now").
// Formerly ..._DropsItem, and it asserted the drop. The row is now written at
// the caller's own status with retry_after set to the REMAINDER of the window,
// so the brake holds the key quiet exactly as long as before while the request
// survives it. The old drop is pinned as the KILL SWITCH's behaviour in
// anti_churn_deferral_test.go, not deleted.
func TestInsertWorkItem_WithinCycleWindow_DefersRatherThanDrops(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	// 1 terminal row, 1h old — inside the window, so 2h of it remain.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(1, 1.0))
	expectDeferredInsert(mock,
		"Re-render page after tool improvement", "triaged",
		withinOf(time.Now().Add(2*time.Hour), 5*time.Minute))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, baseItem(), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if !w.Inserted {
		t.Fatal("within-cycle arrival must be RECORDED and delayed, never dropped (bugs_open/326)")
	}
	if !w.Deferred {
		t.Error("Deferred must be true: the caller cannot otherwise tell this row will not dispatch yet")
	}
	if got := time.Until(w.DeferredUntil); got > 2*time.Hour+5*time.Minute || got < 2*time.Hour-5*time.Minute {
		t.Errorf("DeferredUntil is %v away, want ~2h (the REMAINDER of a %vh window entered at 1h)",
			got.Round(time.Minute), antiChurnWindowHours)
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
