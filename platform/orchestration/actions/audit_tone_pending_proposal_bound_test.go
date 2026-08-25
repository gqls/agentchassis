// The tone route's convergence bound (copy_quality_two_stage handoff
// 2026-08-25, item 3): while a page has an open needs_copy_edit or an
// un-reviewed copy_edit_proposed, a new tone finding must NOT file another.
//
// Why a test and not a comment. idx_swi_dedup looks like it already bounds
// this and does not: it bounds CONCURRENT needs_copy_edit rows only — the
// slot frees when the copy-editor run completes, while the un-reviewed
// proposal it parked is a different item_type holding no dedup slot at all.
// And the dedup key embeds audit_source, so two auditors could file the same
// page in parallel. Stage 2 keeps proposing on repeat runs over one page
// (run 5 re-edited 2 of the 3 components run 4 had just changed), so without
// this bound an auto-dispatched loop accumulates un-reviewed proposals for
// ever. The bound drains when a human acts on the parked proposal — owner
// decision D2's posture: the human is the rate limiter.
//
// The wiring test drives the WHOLE action, deliberately: a helper-level test
// alone still passes when the call site is deleted, which is the mock-
// bookkeeping-cannot-assert-a-negative shape this estate has recorded.

package actions

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var toneFinding = map[string]interface{}{
	"category":    "tone",
	"severity":    "medium",
	"description": "the homepage opens by describing itself rather than addressing the reader",
	"page":        "index",
}

// TestToneBound_WithholdsWhileAProposalIsPending: a tone finding on a page
// that already has a pending copy-edit item or proposal is withheld, counted,
// and reported — not filed. The insert path must never be reached (sqlmock
// fails the test on an unexpected Begin/insert for the finding).
func TestToneBound_WithholdsWhileAProposalIsPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID, pageID := uuid.New(), uuid.New()

	// loadSitePages must return the page or tone classifies down a different rule.
	mock.ExpectQuery("FROM pages").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "page_type", "sections"}).
			AddRow(pageID, "index", "", "[]"))
	// The batch-level blocked-keys load.
	mock.ExpectQuery("SELECT item_key FROM site_work_items").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))
	// The producer-scoped blocked check for this finding: not blocked.
	mock.ExpectQuery(`spec->>'audit_source'`).
		WithArgs(siteID, "needs_copy_edit", sqlmock.AnyArg(), "visual-design-audit").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// THE BOUND: the page already carries a pending copy-edit item/proposal.
	mock.ExpectQuery(`item_type IN \('needs_copy_edit', 'copy_edit_proposed'\)`).
		WithArgs(siteID, pageID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Nothing else for this finding — no dedup check, no insert. The trailing
	// silence-retraction pass still runs; give it its empty round.
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_key", "status", "spec", "result"}))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(),
		auditParams(db, siteID, []interface{}{toneFinding}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	if m["items_created"] != 0 {
		t.Errorf("want items_created 0, got %v — the bound did not withhold", m["items_created"])
	}
	// The receipt must be asserted, not just logged: a withheld finding that
	// reports as an ordinary skip is invisible to whoever reads the result.
	if m["items_skipped_pending_proposal"] != 1 {
		t.Errorf("want items_skipped_pending_proposal 1, got %v", m["items_skipped_pending_proposal"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestPendingCopyEditBound_QueryShapeAndBothArms pins the helper's SQL to the
// CANONICAL terminal-status list (idx_swi_dedup ↔ workItemTerminalStatuses is
// one contract; a hand-rolled status list here is the drift that produced
// SQLSTATE 42P10 fleet-wide once already) and exercises both arms.
func TestPendingCopyEditBound_QueryShapeAndBothArms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID, pageID := uuid.New(), uuid.New()

	q := `(?s)item_type IN \('needs_copy_edit', 'copy_edit_proposed'\)\s+AND status NOT IN \(` +
		regexp.QuoteMeta(sqlInList(workItemTerminalStatuses)) + `\)`

	mock.ExpectQuery(q).WithArgs(siteID, pageID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	pending, err := pendingCopyEditForPage(context.Background(), db, siteID, pageID)
	if err != nil || !pending {
		t.Fatalf("want pending=true nil error, got %v %v", pending, err)
	}

	mock.ExpectQuery(q).WithArgs(siteID, pageID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	pending, err = pendingCopyEditForPage(context.Background(), db, siteID, pageID)
	if err != nil || pending {
		t.Fatalf("want pending=false nil error, got %v %v", pending, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
