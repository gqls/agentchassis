// FILE: platform/orchestration/actions/write_audit_findings_record_retraction_test.go
//
// Controls for the record-mode DEFAULT retraction gate (recordModeSilenceRule).
// Each kills a specific wrong version: a default that never runs (the corrected
// ADDENDUM 1 claim — verdict rows parked for ever); one that swallows
// dispatch-mode rows of ungated types (the deliberately-inert posture
// TestAuditRetraction_UngatedItemTypeIsInert pins); one that double-gates a
// type that already has a measured rule.

package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// expectRecordTypesLoad pins loadRecordModeItemTypes with no record rows — the
// historical shape every pre-existing retraction test now passes through.
func expectRecordTypesLoad(mock sqlmock.Sqlmock, siteID uuid.UUID) {
	mock.ExpectQuery("SELECT DISTINCT item_type").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_type"}))
}

const recordSpec = `{"audit_source":"visual-design-audit","filing_mode":"record","routed_handler":"page-build-handler"}`
const dispatchSpec = `{"audit_source":"visual-design-audit","page_name":"index"}`

// A record row of an UNGATED type retracts on the third silence under the
// default rule. This is the test that fails against the pre-fix code, where
// the gates map's single entry meant a content_rewrite verdict could never
// self-clear.
func TestRecordRetraction_ThirdSilenceRetractsAnUngatedVerdict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()
	key := "visual-design-audit_content_rewrite_index_x"

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).WillReturnRows(candidateRows())
	mock.ExpectQuery("SELECT DISTINCT item_type").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_type"}).AddRow("content_rewrite"))
	mock.ExpectQuery("item_type = 'content_rewrite'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().AddRow(itemID, key, "deferred", recordSpec,
			`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`))
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs("visual-design-audit", argJSONContains{"3 consecutive runs"}, siteID,
			"content_rewrite", key, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "content_rewrite:record")
	if r.Retracted != 1 {
		t.Fatalf("want the verdict retracted under the default rule; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// A DISPATCH-mode row of the same ungated type in the same candidate set is
// OUT OF SCOPE for the default gate — no streak write, no retraction. This is
// the control that keeps TestAuditRetraction_UngatedItemTypeIsInert true for
// the population it was written about.
func TestRecordRetraction_DispatchRowsOfUngatedTypesStayInert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	recordID, dispatchID := uuid.New(), uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).WillReturnRows(candidateRows())
	mock.ExpectQuery("SELECT DISTINCT item_type").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_type"}).AddRow("content_rewrite"))
	mock.ExpectQuery("item_type = 'content_rewrite'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().
			AddRow(recordID, "k1", "deferred", recordSpec, `{}`).
			AddRow(dispatchID, "k2", "failed", dispatchSpec, `{}`))
	// Exactly ONE streak write — the record row's first silence. The dispatch
	// row gets neither a bump nor a retraction; an expectation for a second
	// write would go unmet, and an unexpected one fails the ordered mock.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(recordID, 1, "visual-design-audit").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "content_rewrite:record")
	if r.StreaksBumped != 1 || r.Retracted != 0 || r.Candidates != 1 {
		t.Fatalf("want 1 in-scope candidate bumped once, dispatch row untouched; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// A type that HAS a measured gate is skipped by the default pass: record rows
// of dark_section_audit are judged once, under the measured rule, never twice.
func TestRecordRetraction_GatedTypesAreNotDoubleJudged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().AddRow(itemID, "k", "deferred", recordSpec, `{}`))
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(itemID, 1, "visual-design-audit").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// The record-types pass sees dark_section_audit and SKIPS it — no second
	// candidates query for the type may follow.
	mock.ExpectQuery("SELECT DISTINCT item_type").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_type"}).AddRow("dark_section_audit"))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if _, doubled := out.(map[string]interface{})["retraction"].(map[string]interface{})["dark_section_audit:record"]; doubled {
		t.Fatalf("gated type was judged a second time under the default rule")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Round 04a3ce1f's gating objection (editquality, true round 2): content_rewrite
// has FIVE producers ([MEASURED 2026-08-26] reader 36, content-quality 23,
// site-review 19, brief-fidelity 15, offer 12 record rows on one item_type), so
// the multi-producer property must be a PINNED CONTROL, not an assertion. One
// seat's silence may only ever judge its own rows: another seat's record row of
// the SAME type, and a row with a FOREIGN (non-seat) audit_source that merely
// copied the filing_mode key, are both out of scope — no streak write, no
// retraction. This is also the answer to the "self-service enrolment" worry
// (guardian, architecture): carrying filing_mode='record' does NOT enrol a row;
// enrolment requires BOTH the key AND spec.audit_source naming the seat whose
// run is judging — and only write_audit_findings seats run this pass at all.
func TestRecordRetraction_AnotherSeatsVerdictsAreNotMySilences(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	mineID := uuid.New()
	// The running seat is visual-design-audit (auditParams). Three candidates of
	// ONE item_type: mine (in scope), the reader seat's (another producer), and a
	// hypothetical non-seat producer that copied the key (foreign audit_source).
	readerSpec := `{"audit_source":"reader-experience-audit","filing_mode":"record","routed_handler":"page-build-handler"}`
	foreignSpec := `{"audit_source":"some-future-producer","filing_mode":"record"}`

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).WillReturnRows(candidateRows())
	mock.ExpectQuery("SELECT DISTINCT item_type").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_type"}).AddRow("content_rewrite"))
	mock.ExpectQuery("item_type = 'content_rewrite'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().
			AddRow(mineID, "mine", "deferred", recordSpec, `{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`).
			AddRow(uuid.New(), "readers", "deferred", readerSpec, `{"retraction":{"silent_runs":2,"audit_source":"reader-experience-audit"}}`).
			AddRow(uuid.New(), "foreign", "deferred", foreignSpec, `{}`))
	// Exactly ONE write: MY row's third silence retracts it. The reader's row —
	// same type, same streak — and the foreign row get nothing; an unexpected
	// second write fails the ordered mock, and their absence from the retract
	// call's args is the whole point.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs("visual-design-audit", argJSONContains{"3 consecutive runs"}, siteID,
			"content_rewrite", "mine", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "content_rewrite:record")
	if r.Retracted != 1 || r.Candidates != 1 || r.StreaksBumped != 0 {
		t.Fatalf("want exactly my own row judged (1 candidate, 1 retracted); got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
