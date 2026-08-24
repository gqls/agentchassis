// FILE: platform/orchestration/actions/anti_churn_deferral_test.go
//
// bugs_open/326, option D (owner ruling 2026-08-24, "D + E now"): the anti-churn
// brake's WITHIN-CYCLE arm may DELAY a repeat request; it may no longer destroy
// it. The TWO-STRIKE arm is deliberately unchanged — that asymmetry IS the
// ruling (option D over option A, RFC_048 §6b), because a third of the rows it
// parks are detectors re-finding a fault whose fixer lies about completing
// (bugs_open/352's class), and deferring those re-dispatches a futile fix.
//
// ⚠ THE VACUOUS-PASS TRAP THIS FILE IS WRITTEN AGAINST (LANDMINES.md, footprint
// load_work_item_actions.go): "A test asserting a query is NOT issued passes
// VACUOUSLY against insertWorkItem — it swallows the error the mock raises."
// Every test asserts an EFFECT — a pinned INSERT argument, a returned outcome
// value — never the absence of a call. The kill-switch cases expect NO insert
// and are non-vacuous for the opposite reason, stated at the callsite:
// writeWorkItem RETURNS the Exec error rather than swallowing it.
//
// bugs_open/333's owned-page door also lives in writeWorkItem, BEFORE the
// brake. It is guarded on `item.pageID != nil && *item.pageID != uuid.Nil`, and
// baseItem carries no pageID, so the door never fires here and its probes need
// no scripting — stated so nobody adds expectWorkItemDoorStandsDown "for
// safety" and then wonders why the expectation goes unmet.

package actions

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// timeWithin matches a time.Time argument inside [want±slack]. A bare AnyArg in
// the retry_after position would accept a zero deferral — passing while the
// brake does nothing. The window is what makes the assertion mean something.
type timeWithin struct {
	want  time.Time
	slack time.Duration
}

func withinOf(want time.Time, slack time.Duration) timeWithin {
	return timeWithin{want: want, slack: slack}
}

func (m timeWithin) Match(v driver.Value) bool {
	got, ok := v.(time.Time)
	if !ok {
		return false
	}
	d := got.Sub(m.want)
	if d < 0 {
		d = -d
	}
	return d <= m.slack
}

// expectDeferredInsert pins the SEVENTEEN-argument insert a deferral produces:
// the caller's own summary at $6, the caller's own status at $12, retry_after
// at $17 inside the expected window. The arg COUNT is load-bearing: sqlmock
// fails on a length mismatch, so dropping the conditional column append fails
// this even if every value assertion were relaxed.
func expectDeferredInsert(mock sqlmock.Sqlmock, summary, status string, retryAfter timeWithin) {
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			summary, // $6  — unbranded
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			status, // $12 — the caller's own
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			retryAfter, // $17 — the deferral
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// The window REMAINDER at its boundaries — the property a flat interval would
// silently lose. An arrival late in the window waits the little that is left,
// so the total quiet period per key is exactly what it always was.
func TestWithinCycleDeferral_WaitsTheRemainderNotAFlatInterval(t *testing.T) {
	for _, tc := range []struct {
		ageHours float64
		wantWait time.Duration
	}{
		{0.1, 174 * time.Minute}, // just after a terminal sibling: nearly the whole window
		{1.0, 2 * time.Hour},
		{2.99, 36 * time.Second}, // the case a flat 3h interval would get 300× wrong
	} {
		t.Run(fmt.Sprintf("age=%.2fh", tc.ageHours), func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
				WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(1, tc.ageHours))
			expectDeferredInsert(mock, "Re-render page after tool improvement", "triaged",
				withinOf(time.Now().Add(tc.wantWait), 2*time.Minute))

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			w, err := writeWorkItem(context.Background(), tx, baseItem(), dropOnConflict, zap.NewNop())
			if err != nil {
				t.Fatalf("writeWorkItem: %v", err)
			}
			if !w.Deferred || w.PriorAttempts != 1 {
				t.Fatalf("want Deferred with PriorAttempts=1, got %+v", w)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations: %v", err)
			}
		})
	}
}

// THE ONE TWO-STRIKE SUB-CASE THAT CHANGES: a two-striker arriving INSIDE the
// window used to hit the early-return drop before the brand — vanishing
// entirely. It now falls through to arm B's existing disposition: branded
// 'unresolved', RECORDED like every other two-striker. Not deferred — that
// would be option A, which the guardian vetoed and the owner declined.
func TestTwoStrikerInsideWindow_IsBrandedNotVanished(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(2, 1.0))
	// SIXTEEN args — no retry_after — and arm B's brand, exactly as for a
	// two-striker outside the window.
	expectInsertWithSummaryAndStatus(mock,
		"[unresolved after 2 attempts] Re-render page after tool improvement",
		"unresolved")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	w, err := writeWorkItem(context.Background(), tx, baseItem(), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if !w.Inserted {
		t.Fatal("a two-striker inside the window must be RECORDED (branded), never vanish")
	}
	if w.Deferred {
		t.Error("and it must NOT be deferred — that is option A, which was vetoed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The disarm lever must restore the pre-326 drop EXACTLY, on both sub-cases —
// an untested lever is a lever nobody can safely pull.
func TestAntiChurn_KillSwitch_RestoresTheDropExactly(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count int
		age   float64
	}{
		{"first-striker inside the window", 1, 1.0},
		{"two-striker inside the window", 2, 1.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DISABLE_ANTI_CHURN_DEFERRAL", "1")
			db, mock := newInsertMock(t)
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
				WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(tc.count, tc.age))
			// No INSERT registered. NOT a vacuous absence assertion: if the
			// mutation under test (deleting or inverting the switch) makes the
			// code insert here, sqlmock returns "call to ExecQuery … was not
			// expected", and writeWorkItem RETURNS that error ("insert failed
			// for %s: %w") rather than logging past it — the Fatalf below fires.
			// The landmine's swallowing case is the COUNT probe, whose error is
			// deliberately ignored; the INSERT's is not.

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			w, err := writeWorkItem(context.Background(), tx, baseItem(), dropOnConflict, zap.NewNop())
			if err != nil {
				t.Fatalf("writeWorkItem: %v — an INSERT was issued with the kill switch set", err)
			}
			if w.Inserted || w.Deferred {
				t.Fatalf("legacy within-cycle path must drop the item: got %+v", w)
			}
		})
	}
}

// THE BUG'S OWN REQUIRED NEGATIVE CONTROL: a genuine duplicate while an OPEN
// row holds the key must still dedup — concurrency protection lives entirely in
// idx_swi_dedup via ON CONFLICT, and neither the deferral nor recurrenceExpected
// may weaken it. Two operators submitting one domain at once still make one build.
func TestWriteWorkItem_OpenHolderStillDedups_WhateverTheBrakeDoes(t *testing.T) {
	for _, tc := range []struct {
		name               string
		recurrenceExpected bool
		probe              bool
	}{
		{"recurrence-expected item", true, false},
		{"ordinary item", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()

			mock.ExpectBegin()
			if tc.probe {
				// No terminal siblings: the brake has nothing to act on, so
				// this isolates the dedup arm.
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
					WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
			}
			// RowsAffected 0 == ON CONFLICT DO NOTHING fired: an OPEN row
			// holds the key. This outcome must survive the fix.
			mock.ExpectExec("INSERT INTO site_work_items").
				WillReturnResult(sqlmock.NewResult(0, 0))

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			item := baseItem()
			item.recurrenceExpected = tc.recurrenceExpected

			w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
			if err != nil {
				t.Fatalf("writeWorkItem: %v", err)
			}
			if w.Inserted {
				t.Fatal("a second OPEN item for one (site_id, item_key) must never be created")
			}
			if w.Deferred {
				t.Error("a dedup against an open holder is not a deferral: the work IS queued")
			}
		})
	}
}

// A key with no terminal siblings is byte-identical to before: the SIXTEEN-arg
// insert, nothing deferred, DeferredUntil zero. This is the regression guard
// for the ~20 lanes whose sqlmock expectations pin the argument count.
func TestWriteWorkItem_NoTerminalSiblings_IsUnchanged(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	expectInsertWithSummaryAndStatus(mock,
		"Re-render page after tool improvement", "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	w, err := writeWorkItem(context.Background(), tx, baseItem(), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if !w.Inserted || w.Deferred || w.PriorAttempts != 0 {
		t.Errorf("untouched path changed shape: %+v", w)
	}
	if !w.DeferredUntil.IsZero() {
		t.Errorf("DeferredUntil must stay zero when nothing was deferred, got %v", w.DeferredUntil)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
