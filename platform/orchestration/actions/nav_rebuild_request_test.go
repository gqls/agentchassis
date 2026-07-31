package actions

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Written after council round 4486f1a9, which APPROVED the change and raised two
// objections this file answers. Neither is a behaviour change — both are claims
// about existing machinery that the submission asserted and the seats could not
// verify from outside:
//
//   - `guardian` (medium): "I can't confirm from SQL that `recurrenceExpected`
//     exists on the workItem struct with the semantics claimed; if it doesn't … the
//     third rebuild request goes terminal silently, which is exactly the failure
//     mode being engineered around."
//   - `editquality` (low): "the two-writer removal / RequestNavRebuild path has no
//     test in this plan."
//
// The guardian's objection is the interesting one because the failure it names is
// SILENT: a request branded `unresolved` is terminal and never dispatched, and
// nothing in the log distinguishes that from a request that was correctly
// coalesced.
//
// My first attempt at pinning it was itself a silent pass — see the long comment on
// TestNavRebuildRequestSkipsTheTwoStrikeRule. The short version: it asserted the
// ABSENCE of the two-strike query, sqlmock turns an unexpected query into an error,
// insertWorkItem swallows that error, and so the test stayed green with
// `recurrenceExpected: false`. Every test here is now checked by breaking the thing
// it guards and watching it fail, because on this seam a green test was the default
// rather than the finding.

func navRebuildMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func navRebuildTestRequest() NavRebuildRequest {
	return NavRebuildRequest{
		SiteID:      uuid.New(),
		PageID:      uuid.New(),
		PageName:    "tool-drop-rate-tuner",
		PageURL:     "/tools/tool-drop-rate-tuner.html",
		InHeader:    true,
		InFooter:    true,
		RequestedBy: "tool-deployer",
	}
}

// TestNavRebuildRequestSkipsTheTwoStrikeRule is the guardian's objection, pinned.
//
// insertWorkItem's two-strike block runs a COUNT over prior terminal items with the
// same item_key and brands the third as `unresolved` — terminal, never dispatched.
// It is guarded by `if item.itemKey != "" && !item.recurrenceExpected`.
//
// THE FIRST VERSION OF THIS TEST WAS WORTHLESS AND PASSED, which is why it is
// written the awkward way it now is. That version registered only Begin/Exec/Commit
// and treated the ABSENCE of the COUNT query as the assertion, on the reasoning
// that sqlmock errors on an unexpected query. It does — and insertWorkItem
// SWALLOWS that error (`if err == nil && terminalCount > 0`), so the branding never
// happens, the INSERT proceeds, and every registered expectation is still met.
// Setting `recurrenceExpected: false` left the test green. The mock environment was
// masking the exact difference the test existed to detect.
//
// So this asserts the mechanism's EFFECT instead: supply a two-strike history that
// WOULD brand the item, then require the INSERT to carry `status = 'triaged'`. A
// WithArgs mismatch makes the Exec fail, which makes RequestNavRebuild return
// false, which fails the test. Proven by setting `recurrenceExpected: false`:
// the INSERT then carries `unresolved` and this test fails.
//
// `ExpectationsWereMet` is deliberately NOT asserted here: on the correct code path
// the COUNT query is never issued, so that expectation is legitimately unused, and
// requiring it to be met would invert the test. The status argument is the assertion.
func TestNavRebuildRequestSkipsTheTwoStrikeRule(t *testing.T) {
	ctx := context.Background()
	db, mock := navRebuildMockDB(t)
	req := navRebuildTestRequest()

	mock.MatchExpectationsInOrder(false)

	// A history that WOULD trip the two-strike rule: 2 prior terminal items, the
	// newest 100 hours old (so the <3h within-cycle suppression does not apply).
	// Registered so that if the flag is ever dropped the query SUCCEEDS and the
	// branding actually happens — rather than erroring and being swallowed.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age_hours"}).AddRow(2, 100.0))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(navRebuildInsertArgsRequiringStatus("triaged")...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if got := RequestNavRebuild(ctx, db, req, zap.NewNop()); !got {
		t.Fatalf("RequestNavRebuild returned false: the INSERT did not carry status='triaged'. " +
			"With a 2-strike history that means insertWorkItem branded it 'unresolved' — " +
			"i.e. recurrenceExpected is no longer set, and the third nav rebuild request " +
			"for a site would be born terminal and never dispatched")
	}
}

// navRebuildInsertArgsRequiringStatus matches insertWorkItem's 16-column INSERT
// while pinning only the status column ($12), so the test breaks on a status change
// and not on every unrelated column addition.
//
// Column order, from insertWorkItem: site_id, source, pipeline, item_type,
// severity, summary, spec, page_id, component_id, priority, handler_agent, STATUS,
// created_by, item_key, batch_id, depends_on.
func navRebuildInsertArgsRequiringStatus(status string) []driver.Value {
	const cols = 16
	const statusIdx = 11 // 0-based; $12

	args := make([]driver.Value, cols)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[statusIdx] = status
	return args
}

// TestNavRebuildRequestCoalescesRatherThanDuplicating pins the editquality seat's
// question: a second request while one is still open must be a no-op, not a
// duplicate and not an error. insertWorkItem's ON CONFLICT ... DO NOTHING makes
// that a zero-rows-affected insert, and the caller must read that as "already
// covered" rather than as failure — a tool build must not report a problem
// because the site already has a rebuild pending.
func TestNavRebuildRequestCoalescesRatherThanDuplicating(t *testing.T) {
	ctx := context.Background()
	db, mock := navRebuildMockDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WillReturnResult(sqlmock.NewResult(0, 0)) // ON CONFLICT DO NOTHING
	mock.ExpectCommit()

	if got := RequestNavRebuild(ctx, db, navRebuildTestRequest(), zap.NewNop()); got {
		t.Fatalf("RequestNavRebuild returned true when the insert affected no rows — a coalesced request must report false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB interactions: %v", err)
	}
}

// TestNavRebuildRequestNeverBreaksItsCaller. The request is emitted from inside a
// tool build that has ALREADY succeeded — the page row, the component and the
// content work item are committed by the time it runs. A nav request that
// propagated a DB error would therefore fail a build that worked, which is a worse
// outcome than a missing footer link. Same policy as the cross-link emitter.
func TestNavRebuildRequestNeverBreaksItsCaller(t *testing.T) {
	ctx := context.Background()

	t.Run("insert fails", func(t *testing.T) {
		db, mock := navRebuildMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		if got := RequestNavRebuild(ctx, db, navRebuildTestRequest(), zap.NewNop()); got {
			t.Fatalf("returned true after an insert error")
		}
	})

	t.Run("nil db", func(t *testing.T) {
		if got := RequestNavRebuild(ctx, nil, navRebuildTestRequest(), zap.NewNop()); got {
			t.Fatalf("returned true with a nil DB")
		}
	})

	t.Run("no site id", func(t *testing.T) {
		db, _ := navRebuildMockDB(t)
		req := navRebuildTestRequest()
		req.SiteID = uuid.Nil
		// No expectations registered: it must not touch the DB at all.
		if got := RequestNavRebuild(ctx, db, req, zap.NewNop()); got {
			t.Fatalf("returned true with a nil site_id")
		}
	})
}
