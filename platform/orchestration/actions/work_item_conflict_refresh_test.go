// FILE: platform/orchestration/actions/work_item_conflict_refresh_test.go
//
// bugs_open/091 candidate 1 — "refresh the open item instead of dropping the
// finding".
//
// The filed defect: `stale_evidence` is keyed per SITE while the finding is per
// FACT, so a second, DIFFERENT fact drifting while an earlier item is open hit
// ON CONFLICT DO NOTHING and vanished. The row stayed, describing the EARLIER
// drift, and it is the only thing a human ever reads. Measured on 2026-08-02,
// four of the five open stale_evidence items named the wrong facts — one naming
// a completely different fact from the one that had moved, one describing drift
// that no longer existed at all.
//
// Every test here exercises the CONFLICT branch. The happy path cannot show any
// of this: an insert that succeeds behaves identically under both policies,
// which is exactly why the defect survived a year of green builds.
package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func refreshableItem() workItem {
	return workItem{
		siteID:       uuid.New(),
		source:       "scheduled",
		pipeline:     "content",
		itemType:     "stale_evidence",
		severity:     "medium",
		summary:      "Evidence freshness (example.com): 1 fact(s) drifted outside tolerance",
		spec:         `{"drifted":[{"fact_id":"C4-agent-definitions-catalogue"}]}`,
		priority:     35,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    "evidence-freshness",
		itemKey:      "stale_evidence:" + uuid.New().String(),
		// Deliberately a DETECTED finding, so the anti-churn probe runs — the
		// refresh must compose with it, not sidestep it.
		recurrenceExpected: false,
	}
}

// THE REGRESSION. An OPEN item holds the key, the insert writes nothing, and the
// finding used to be lost with it. It must now update that item's description.
//
// The assertion that matters most is `Inserted == false`. `work_item_created` is
// set from it, and if a refresh were allowed to set it true this fix would have
// re-created 091's original lie (a reported creation that never happened) inside
// 091's own fix — which is precisely what `ON CONFLICT … DO UPDATE` would have
// done, because RowsAffected counts an update.
func TestWriteWorkItem_RefreshOnConflict_UpdatesTheOpenItem(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	item := refreshableItem()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	// 0 rows: an OPEN item already holds this key.
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// The refresh, carrying THIS finding's summary and spec to the open row. It is
	// a QUERY now (RETURNING status), and the two status lists arrive as PARAMETERS
	// rather than interpolated text — council 8e7357ae, constitution seat.
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE site_work_items")).
		WithArgs(item.siteID, item.itemKey, item.summary, item.spec,
			sqlTextArrayLiteral(workItemTerminalStatuses),
			sqlTextArrayLiteral(workItemHeldStatuses)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("needs_human_review"))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, item, refreshOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.Inserted {
		t.Error("Inserted must stay FALSE for a refresh — work_item_created is set from it, " +
			"and a refresh that reports a creation is the bug this fixes, re-created")
	}
	if !w.Refreshed {
		t.Error("the open item was not refreshed; the second finding is lost again")
	}
	if !w.Recorded() {
		t.Error("Recorded() must be true — a durable record now describes this finding")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The default policy must be untouched. ~20 call sites use insertWorkItem and
// none of them asked for a refresh; if the conflict branch started issuing an
// UPDATE for them, this seam would have changed behaviour for callers who never
// opted in. sqlmock is the assertion: an unexpected UPDATE fails the test.
func TestWriteWorkItem_DropOnConflict_IssuesNoUpdate(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// NO ExpectExec for an UPDATE: issuing one fails this test.

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	// THIS TEST IS NOT VACUOUS, AND THAT WAS CHALLENGED — three council seats on
	// 8e7357ae (editquality, guardian, debug_historian) all asked whether a negative
	// sqlmock assertion means anything here, given the anti-churn probe a few lines
	// away in the SAME function swallows its own error and would let such a test pass
	// whatever happened. The answer is that refreshOpenWorkItem PROPAGATES its error
	// (only sql.ErrNoRows is treated as "nothing matched"), so an unexpected query
	// surfaces as a returned error rather than being absorbed — which the t.Fatalf
	// below turns into a failure. Verified by mutation, not by reading: forcing the
	// default policy down the refresh path fails this test and TestInsertWorkItem_
	// CannotRefresh. The err check is therefore load-bearing; do not soften it.
	w, err := writeWorkItem(context.Background(), tx, refreshableItem(), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.Inserted || w.Refreshed || w.Recorded() {
		t.Errorf("default policy must record nothing on conflict, got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// insertWorkItem is the door every existing caller uses. It must be incapable of
// refreshing whatever the item says — the policy is a parameter it cannot pass.
func TestInsertWorkItem_CannotRefresh(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Again: no UPDATE may be issued.

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	inserted, err := insertWorkItem(context.Background(), tx, refreshableItem(), zap.NewNop())
	if err != nil {
		t.Fatalf("insertWorkItem: %v", err)
	}
	if inserted {
		t.Error("insertWorkItem reported a creation over a conflict")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A successful insert must not then also run the refresh. Belt-and-braces on the
// ordering, because "insert, then update anyway" is a plausible mis-edit that no
// caller would ever notice: the row would be correct either way, and the second
// statement would just be silently wasted work on every keyed insert in the fleet.
func TestWriteWorkItem_RefreshOnConflict_NoUpdateWhenTheInsertWon(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, refreshableItem(), refreshOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if !w.Inserted || w.Refreshed {
		t.Errorf("a clean insert must report Inserted only, got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The race the predicate exists to lose safely. Nothing locks the row between
// the failed insert and the update, so the open item may go terminal — or a
// handler may claim it — in between. The UPDATE then matches nothing, and the
// caller must be told NOTHING was recorded rather than being handed a refresh
// that did not happen. An honest false is what makes the caller's warning fire
// and the next sweep insert cleanly.
func TestWriteWorkItem_RefreshOnConflict_ReportsNothingWhenTheRowIsUnavailable(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// The refresh matches no row: it went terminal, or a handler holds it.
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE site_work_items")).
		WillReturnRows(sqlmock.NewRows([]string{"status"}))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, refreshableItem(), refreshOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.Recorded() {
		t.Errorf("nothing was written and nothing was updated; Recorded() must be false, got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A keyless item has no conflict to resolve and no row a refresh could find.
// Without this guard the UPDATE's predicate would be `item_key = ”`, which
// matches nothing today but is one schema default away from matching everything
// on a site. Assert the statement is never issued at all.
func TestWriteWorkItem_RefreshOnConflict_KeylessItemIssuesNoUpdate(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	item := refreshableItem()
	item.itemKey = ""

	mock.ExpectBegin()
	// No anti-churn probe either: that block is keyed too.
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, item, refreshOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.Recorded() {
		t.Errorf("a keyless item cannot be refreshed, got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The guards live in the UPDATE's own predicate, so assert the predicate. A mock
// cannot show that a claimed row was spared — it returns whatever it is told —
// so the statement text is the only thing that can carry this claim, and these
// two clauses are the whole of the safety argument:
//
//   - terminal statuses: a completed item is never resurrected by a refresh;
//   - held statuses: a handler that read the spec when it claimed the row would
//     never see a change made underneath it.
//
// Both lists are interpolated from their single source in work_items_common.go —
// the same lockstep obligation idx_swi_dedup has, and the reason a status added
// to one list cannot silently go missing from this statement.
func TestRefreshStatement_GuardsTerminalAndHeldRows(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE site_work_items")).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("triaged"))
	mock.MatchExpectationsInOrder(true)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := writeWorkItem(context.Background(), tx, refreshableItem(), refreshOnConflict, zap.NewNop()); err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}

	// The statement that actually ran — not a copy re-derived in the test, which
	// would only prove the test agrees with itself.
	refreshSQL := refreshOpenWorkItemSQL()

	// The lists are PARAMETERS now, so the statement carries their placeholders and
	// the VALUES are asserted through the literal builder the code passes.
	for _, clause := range []string{"$5::text[]", "$6::text[]"} {
		if !strings.Contains(refreshSQL, clause) {
			t.Errorf("refresh predicate lost %s — one of the two guards is gone", clause)
		}
	}
	termLit := sqlTextArrayLiteral(workItemTerminalStatuses)
	heldLit := sqlTextArrayLiteral(workItemHeldStatuses)
	for _, st := range workItemTerminalStatuses {
		if !strings.Contains(termLit, st) {
			t.Errorf("terminal status %q missing — a %s item would be resurrected by a refresh", st, st)
		}
	}
	for _, st := range workItemHeldStatuses {
		if !strings.Contains(heldLit, st) {
			t.Errorf("held status %q missing — a handler holding the row would have its "+
				"spec changed underneath it", st)
		}
	}
	if strings.Contains(refreshSQL, "'complete'") {
		t.Error("status list is interpolated into the statement text again — it must be a parameter")
	}
	if !strings.Contains(refreshSQL, "summary") || !strings.Contains(refreshSQL, "spec") {
		t.Error("the refresh must carry the finding's description")
	}
	for _, forbidden := range []string{"status =", "priority =", "handler_agent =", "severity ="} {
		if strings.Contains(refreshSQL, forbidden) {
			t.Errorf("the refresh must not write %q — a human may have moved it on the open row", forbidden)
		}
	}
}
