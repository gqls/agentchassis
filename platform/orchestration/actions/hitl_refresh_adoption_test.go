// FILE: platform/orchestration/actions/hitl_refresh_adoption_test.go
//
// bugs_open/184 — the three sibling call sites of the bugs_closed/091 class.
//
// Each files a HITL-terminal work item whose item_key is coarser than its
// finding: `citation_unverified:<site>` over a list of rejected candidates, and
// `directory_citation_unverified` / `stale_directory_claim`, which are keyed by a
// CONSTANT so one row stands for the entire model directory. Under the default
// conflict policy every finding after the first was discarded and the open row
// went on describing the first batch it ever saw.
//
// WHY THESE TESTS EXIST, given the change is one identifier per site: the defect
// is INVISIBLE on the happy path. An insert that succeeds behaves identically
// under both policies, so nothing — not a green suite, not a clean run, not the
// action's own result — distinguishes a switched call site from an unswitched one.
// Only the STATEMENTS the emitter issues can tell them apart.
//
// THESE TESTS DRIVE THE REAL EMITTERS, and that is deliberate and load-bearing.
// A first draft called writeWorkItem directly with refreshOnConflict and asserted
// the outcome; it passed, it read convincingly, and it was worthless — reverting a
// call site to insertWorkItem did not fail it, because it never touched the call
// sites at all. A test named for a call site must CALL it.
package actions

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// expectConflictThenRefresh scripts one emitter's write: the INSERT hits an open
// row (0 rows affected) and the refresh must therefore run. An emitter still on
// insertWorkItem issues no refresh, leaving the UPDATE expectation unmet — which
// is exactly what makes this test able to fail.
func expectConflictThenRefresh(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0)) // an OPEN row already holds the key
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE site_work_items")).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("needs_human_review"))
	mock.ExpectCommit()
}

// evidence_citations.go — key is per SITE, finding is a list of rejected
// candidates that differs every research pass.
func TestCreateCitationFailuresItem_RefreshesTheOpenRow(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectConflictThenRefresh(mock)

	params := ActionParams{DB: db, AgentType: "evidence-researcher"}
	err := createCitationFailuresItem(context.Background(), params, uuid.New(),
		[]map[string]interface{}{{"claim": "a"}, {"claim": "b"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("createCitationFailuresItem: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the finding was dropped instead of refreshing the open row "+
			"(bugs_open/184, the bugs_closed/091 class): %v", err)
	}
}

// directory_claims.go — key is a CONSTANT, so one row stands for the whole
// directory. The coarsest instance of the class in the tree.
func TestCreateDirectoryCitationFailuresItem_RefreshesTheOpenRow(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectConflictThenRefresh(mock)

	params := ActionParams{DB: db, AgentType: "directory-researcher"}
	err := createDirectoryCitationFailuresItem(context.Background(), params,
		[]rejectedDirectoryClaim{{Slug: "a"}, {Slug: "b"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("createDirectoryCitationFailuresItem: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the finding was dropped instead of refreshing the open row "+
			"(bugs_open/184): %v", err)
	}
}

// directory_claims.go — the sharpest of the three: its SUMMARY carries a count
// ("%d claim(s) changed verification status"), so a frozen row shows a number
// that was true once. A count is what a human reads without opening the spec.
func TestCreateStaleDirectoryClaimItem_RefreshesTheOpenRow(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectConflictThenRefresh(mock)

	params := ActionParams{DB: db, AgentType: "directory-freshness"}
	err := createStaleDirectoryClaimItem(context.Background(), params,
		[]directoryClaimRefresh{
			{Slug: "a", Outcome: "citation_lost"},
			{Slug: "b", Outcome: "fetch_error"},
		}, zap.NewNop())
	if err != nil {
		t.Fatalf("createStaleDirectoryClaimItem: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the finding was dropped instead of refreshing the open row "+
			"(bugs_open/184): %v", err)
	}
}

// The counterfactual, so the three tests above are known to be measuring
// something rather than passing on a mock that would accept anything: under the
// DEFAULT policy the identical sequence records nothing and issues no refresh.
// If this ever fails, the default has changed underneath every caller that never
// opted in — a far bigger event than 184.
func TestDefaultPolicyStillDropsTheFinding(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// No UPDATE expected. writeWorkItem PROPAGATES an unexpected-query error
	// (only sql.ErrNoRows means "nothing matched"), so issuing one fails this
	// test rather than being absorbed — unlike the anti-churn probe above it.

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, workItem{
		siteID: uuid.New(), itemType: "citation_unverified", spec: "{}",
		status: "needs_human_review", itemKey: "citation_unverified:x",
	}, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.Recorded() {
		t.Errorf("the default policy must still drop, got %+v — if this changed, every "+
			"caller that never opted in has changed with it", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
