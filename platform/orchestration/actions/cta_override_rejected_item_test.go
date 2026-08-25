// FILE: platform/orchestration/actions/cta_override_rejected_item_test.go
//
// The follow-up accepted from the CTA-override council round's bug_historian
// objection (noted_rebuild NOTES, 2026-08-25): a REFUSED header_cta_url
// override logged only a Warn, so the owner's explicit request could silently
// degrade for ever. emitCTAOverrideRejectedItem now files an owner-visible
// needs_human_review item; these tests pin the two properties that make it
// worth having.
//
// Like hitl_refresh_adoption_test.go, THESE TESTS DRIVE THE REAL EMITTER —
// calling writeWorkItem directly with the desired policy passes regardless of
// what the emitter does, which is the vacuous-test trap that file documents.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The key is per SITE (one override slot per site) while the finding names the
// CURRENT refused value, so the emitter must refresh an open row rather than
// drop the new finding (bugs_open/184, the bugs_closed/091 class): an owner who
// edits a refused override to a second, also-refused value must not be shown an
// open item describing the first for ever. An emitter reverted to insertWorkItem
// issues no UPDATE, leaving the expectation unmet — which is what lets this fail.
func TestEmitCTAOverrideRejectedItem_RefreshesTheOpenRow(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectConflictThenRefresh(mock)

	emitCTAOverrideRejectedItem(context.Background(), db, uuid.New(),
		"/tools/nonexistent.html", "/contact/index.html", zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the finding was dropped instead of refreshing the open row "+
			"(bugs_open/184, the bugs_closed/091 class): %v", err)
	}
}

// First refusal on a site: a new row must actually be written and committed.
// The emitter is fire-and-forget (log-and-continue on error, deliberately —
// a refused override must never fail the render), so the mock's script is the
// only witness that the write happened at all.
func TestEmitCTAOverrideRejectedItem_InsertsWhenNoOpenRow(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	// anti-churn probe: keyed, recurrenceExpected false, no terminal siblings
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1)) // fresh row: no refresh needed
	mock.ExpectCommit()

	emitCTAOverrideRejectedItem(context.Background(), db, uuid.New(),
		"/tools/nonexistent.html", "", zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the refusal was not durably recorded: %v", err)
	}
}
