// FILE: platform/orchestration/actions/work_item_created_honesty_test.go
//
// bugs_open/091 candidate 2 — "stop reporting a write that did not happen".
//
// Three separate actions reported that a work item had been created when what
// they actually knew was that nothing had errored:
//
//	refresh_evidence_base_action.go  res.WorkItemCreated = true   (in the else of an err check)
//	apply_gap_plan_action.go         "item_created": true         (over ON CONFLICT DO NOTHING)
//	apply_gap_plan_action.go         "item_created": true         (retype arm, same statement shape)
//
// The filed case is the first: the stale_evidence item is keyed per SITE, so a
// second, different fact drifting while an earlier item is still open dedups to
// nothing — and the run said it had raised it. The only trace of the truth was a
// log line carrying inserted=false, in a pod replaced four minutes later.
//
// These tests exercise the SUPPRESSED branch, because the reporting bug is
// invisible on the happy path — an insert that succeeds reports true either way,
// which is precisely why this survived.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TestExecWroteARow covers the helper the two apply_gap_plan sites now use.
// The ON CONFLICT DO NOTHING case — 0 rows affected — is the one that used to be
// reported as a creation.
func TestExecWroteARow(t *testing.T) {
	if !execWroteARow(sqlmock.NewResult(0, 1)) {
		t.Error("a row WAS written and execWroteARow said no")
	}
	if execWroteARow(sqlmock.NewResult(0, 0)) {
		t.Error("ON CONFLICT DO NOTHING wrote nothing and execWroteARow said yes — " +
			"this is the bugs_open/091 report shape")
	}
	// Defensive branch: a driver that cannot answer degrades to the previous,
	// optimistic answer rather than inventing a false negative.
	if !execWroteARow(nil) {
		t.Error("a nil Result should degrade to the previous behaviour, not report a false negative")
	}
}

// TestApplyNewPage_ReportsItemCreatedFalseWhenDeduped is the regression on the
// gap-planner arm: the page upserts, the work item hits ON CONFLICT DO NOTHING,
// and the action must now say so.
func TestApplyNewPage_ReportsItemCreatedFalseWhenDeduped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// 0 rows affected — an item already holds this key.
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))

	res, err := applyNewPage(context.Background(), db,
		newPagePlan("recovery-waterfall", ""), uuid.New(), "example.com", nil, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}
	m := res.(map[string]interface{})
	if m["item_created"] != false {
		t.Errorf("item_created = %v, want false — nothing was written", m["item_created"])
	}
	// The page itself DID upsert, so the action still succeeded.
	if m["applied"] != true {
		t.Errorf("applied = %v, want true — the page was created even though the item deduped", m["applied"])
	}
}

// TestApplyNewPage_ReportsItemCreatedTrueWhenWritten is the other half: the
// honest report must still say true when a row really was written, or the fix
// has just moved the lie in the opposite direction.
func TestApplyNewPage_ReportsItemCreatedTrueWhenWritten(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := applyNewPage(context.Background(), db,
		newPagePlan("recovery-waterfall", ""), uuid.New(), "example.com", nil, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}
	if m := res.(map[string]interface{}); m["item_created"] != true {
		t.Errorf("item_created = %v, want true — a row was written", m["item_created"])
	}
}

// TestApplyRetypeExisting_ReportsItemCreatedFalseWhenDeduped covers the arm the
// council's editquality seat noted was unnamed: the retype path has the same
// hardcoded-true over the same ON CONFLICT DO NOTHING statement, and
// execWroteARow being unit-tested does not prove this caller USES it.
func TestApplyRetypeExisting_ReportsItemCreatedFalseWhenDeduped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	itemID := uuid.New()

	// Same sequence as TestApplyRetypeExisting_HappyPath — the authorising spec
	// comes from the ORIGINAL work item, written deterministically by the
	// discovery check, because this arm is fail-closed and does not trust the
	// LLM plan. The ONE difference is the work-item INSERT affecting 0 rows.
	mock.ExpectQuery("SELECT spec FROM site_work_items").
		WithArgs(itemID, siteID).
		WillReturnRows(sqlmock.NewRows([]string{"spec"}).AddRow(retypeItemSpecJSON))
	mock.ExpectExec("UPDATE pages").
		WithArgs(uuid.MustParse(retypeCandidateID), "news-index", sqlmock.AnyArg(), siteID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 0 rows affected — a build item already holds this key.
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// markOriginalComplete
	mock.ExpectExec("UPDATE site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := applyRetypeExisting(context.Background(), db,
		retypePlan("noticias-index"), siteID, &itemID, zap.NewNop())
	if err != nil {
		t.Fatalf("applyRetypeExisting: %v", err)
	}
	res := got.(map[string]interface{})
	if res["applied"] != true {
		t.Fatalf("applied=%v want true — the re-type itself succeeded (reason=%v)",
			res["applied"], res["reason"])
	}
	if res["item_created"] != false {
		t.Errorf("item_created = %v, want false — nothing was written", res["item_created"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
