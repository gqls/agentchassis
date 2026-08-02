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

// 2026-08-02, candidate 1: the three hand-rolled INSERTs are gone and all three
// arms now route through insertWorkItem, so execWroteARow (and its unit test)
// went with them. What replaces that test is the assertion that matters more —
// that the adoption did not quietly change behaviour.
//
// gapPlanWorkItem MUST set recurrenceExpected. Routing through the shared door
// brings the two-strike anti-churn with it, and a gap plan asking for a page to
// be built is an ACTION REQUEST, not a redetected defect: a completed
// predecessor means the request SUCCEEDED. Without the flag, adoption would
// newly suppress an item within 3h of a terminal predecessor and brand it
// 'unresolved' after two — bugs_open/024's regression, re-created by a fix for
// a different bug. This is cheap to assert and impossible to notice in review.
func TestGapPlanWorkItem_IsRecurrenceExpected(t *testing.T) {
	item := gapPlanWorkItem(uuid.New(), "needs_content_page", "s", "{}", nil, 40, "k", nil)
	if !item.recurrenceExpected {
		t.Error("gap-plan items must be recurrence-expected: they are action requests, " +
			"so a COMPLETE predecessor is a success, not a strike (bugs_open/024)")
	}
	if item.itemKey == "" {
		t.Error("a gap-plan item without an item_key is not deduped at all")
	}
}

// The parent link is the reason all three arms forked away from the shared
// helper in the first place — workItem had no field for it. If it silently stops
// being carried, the fork's whole justification comes back and nobody notices,
// because a dropped parent_item_id breaks nothing that errors.
func TestGapPlanWorkItem_CarriesTheParentLink(t *testing.T) {
	parent := uuid.New()
	item := gapPlanWorkItem(uuid.New(), "needs_content_page", "s", "{}", nil, 40, "k", &parent)
	if item.parentItemID == nil || *item.parentItemID != parent {
		t.Errorf("parentItemID = %v, want %v — the shared helper must carry the parent link, "+
			"which is the ONLY thing the hand-rolled INSERTs had that it did not", item.parentItemID, parent)
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
	// The item now goes through withWorkItemTx -> insertWorkItem, so it has its
	// own transaction. No anti-churn probe is expected: gap-plan items are
	// recurrence-expected, and sqlmock fails the test if one is issued.
	mock.ExpectBegin()
	// 0 rows affected — an OPEN item already holds this key.
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

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
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

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
	// 0 rows affected — a build item already holds this key. Own transaction now.
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
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
