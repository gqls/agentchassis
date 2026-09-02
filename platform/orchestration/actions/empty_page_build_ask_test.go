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
	// must file the needs_content_page ask (stable key) and NOT a rerender.
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM page_components`).
		WithArgs(emptyPage).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WithArgs(siteID, sqlmock.AnyArg(), emptyPage, sqlmock.AnyArg(),
			"needs_content_page:"+emptyPage.String()).
		WillReturnResult(sqlmock.NewResult(0, 1))

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

	// Filed: RowsAffected 1 -> true, with the stable per-page dedup key.
	db, mock, _ := sqlmock.New()
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WithArgs(siteID, sqlmock.AnyArg(), pageID, sqlmock.AnyArg(),
			"needs_content_page:"+pageID.String()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !fileBuildAskForEmptyPage(context.Background(), db, siteID, pageID, "p", nil, zap.NewNop()) {
		t.Error("RowsAffected=1 must report filed=true")
	}
	db.Close()

	// Deduped: RowsAffected 0 (idx_swi_dedup conflict) -> false, no error.
	db2, mock2, _ := sqlmock.New()
	mock2.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if fileBuildAskForEmptyPage(context.Background(), db2, siteID, pageID, "p", nil, zap.NewNop()) {
		t.Error("a deduped insert must report filed=false")
	}
	db2.Close()

	// Insert failure must never fail the skip: false, no panic.
	db3, mock3, _ := sqlmock.New()
	mock3.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnError(errors.New("boom"))
	if fileBuildAskForEmptyPage(context.Background(), db3, siteID, pageID, "p", nil, zap.NewNop()) {
		t.Error("an insert error must report filed=false")
	}
	db3.Close()
	_ = mock
}
