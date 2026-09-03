package actions

// bugs_open/315 (reopened 2026-09-02): a page with ZERO component rows
// collected NINE completed page_rerender items while serving an empty <main>,
// because the 0-component skip completes the item and nothing routes the page
// to the BUILD queue. Two guards close it — the producer stops filing
// rerenders for such pages and converts them to a deduped needs_content_page
// ask (this file's action-level test), and the consumer files the same ask on
// any skip that still reaches it (helper tests below; same item_key, so both
// doors converge on ONE open build ask).
//
// MUTATION-PROVED at authoring: removing the !hasComponents conversion in
// CreateRerenderItemsAction fails TestEmptyPageRerenderConvertsToBuildAsk
// (sqlmock: unexpected page_rerender INSERT for the empty page); removing the
// helper's INSERT fails TestFileBuildAskForEmptyPage.

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

func TestEmptyPageRerenderConvertsToBuildAsk(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	emptyPage := uuid.New()
	builtPage := uuid.New()

	// Page A has 0 component rows: the EXISTS probe answers false, the action
	// must file the needs_content_page ask (via writeWorkItem, in its own tx)
	// and NOT a rerender.
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM page_components`).
		WithArgs(emptyPage).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectBegin() // the shared filer runs writeWorkItem in its own tx
	expectBuildAskWrite(mock, emptyPage, 1)
	mock.ExpectCommit()

	// Page B has components: EXISTS true, the normal page_rerender item files.
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM page_components`).
		WithArgs(builtPage).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := CreateRerenderItemsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": "site_id",
			"domain":  "domain",
		}},
		CollectedData: map[string]interface{}{
			"site_id": siteID.String(),
			"domain":  "example.com",
			"rerender_pages": map[string]interface{}{
				"pages": []interface{}{
					map[string]interface{}{"page_id": emptyPage.String(), "name": "roi-estimator", "filename": "roi-estimator.html"},
					map[string]interface{}{"page_id": builtPage.String(), "name": "about", "filename": "about.html"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateRerenderItemsAction: %v", err)
	}
	result, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result shape: %T", out)
	}
	if result["items_created"] != 1 {
		t.Errorf("items_created = %v, want 1 (the empty page must not get a rerender item)", result["items_created"])
	}
	converted, _ := result["empty_pages_converted_to_build_asks"].([]string)
	if len(converted) != 1 || converted[0] != "roi-estimator" {
		t.Errorf("empty_pages_converted_to_build_asks = %v, want [roi-estimator] — the conversion must be in the RESULT, not only a pod log", result["empty_pages_converted_to_build_asks"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectation not met (the conversion INSERT or the EXISTS probe did not run): %v", err)
	}
}

func TestFileBuildAskForEmptyPage(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()
	logger := zaptest.NewLogger(t)

	// Filed: the writeWorkItem INSERT reports a new row -> true.
	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	expectBuildAskWrite(mock, pageID, 1)
	mock.ExpectCommit()
	if !fileBuildAskForEmptyPage(context.Background(), db, siteID, pageID, "p", nil, logger) {
		t.Error("a new row must report filed=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
	db.Close()

	// Deduped (idx_swi_dedup conflict, no row back) -> false, no error.
	db2, mock2, _ := sqlmock.New()
	mock2.ExpectBegin()
	expectBuildAskWrite(mock2, pageID, 0)
	mock2.ExpectCommit()
	if fileBuildAskForEmptyPage(context.Background(), db2, siteID, pageID, "p", nil, logger) {
		t.Error("a deduped write must report filed=false")
	}
	db2.Close()

	// A door/write failure must never fail the skip: false, no panic.
	db3, mock3, _ := sqlmock.New()
	mock3.ExpectBegin()
	mock3.ExpectQuery(`SELECT COALESCE\(pages.rebuild_policy`).
		WillReturnError(errors.New("boom"))
	if fileBuildAskForEmptyPage(context.Background(), db3, siteID, pageID, "p", nil, zap.NewNop()) {
		t.Error("a write error must report filed=false")
	}
	db3.Close()
}

// expectBuildAskWrite encodes writeWorkItem's query sequence for THIS item
// shape (generic page, page-build-handler, status triaged): the policy-door
// rebuild_policy read, then the INSERT. rows=1 -> a new row; rows=0 -> the
// dedup conflict swallowed by the policy. If writeWorkItem grows a door that
// queries before inserting, this helper is the one place to teach it.
func expectBuildAskWrite(mock sqlmock.Sqlmock, pageID uuid.UUID, rows int64) {
	mock.ExpectQuery(`SELECT COALESCE\(pages.rebuild_policy`).
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", false))
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnResult(sqlmock.NewResult(0, rows))
}
