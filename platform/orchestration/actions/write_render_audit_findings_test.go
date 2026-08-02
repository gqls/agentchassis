package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// The assertions here pin EFFECTS (which INSERT ran, with which item_type,
// status, handler and key), never the absence of a query — a vacuous guard
// passes against insertWorkItem because it swallows sqlmock errors
// (work_item_recurrence_test.go's lesson).

func renderAuditCollected(siteID uuid.UUID, payload map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"site_id": siteID.String(),
		// The coordinator stores an awaited adapter response under
		// output_field.response — the unwrap path is part of what's under test.
		"render_audit": map[string]interface{}{
			"response":          payload,
			"response_status":   "complete",
			"response_received": "2026-08-02T00:00:00Z",
		},
	}
}

func TestWriteRenderAuditFindings_FilesFirmContrastSkipsOverImage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// No locked components on the site.
	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))

	mock.ExpectBegin()
	// Two-strike pre-check for the ONE firm finding (the over_image one must
	// never reach here — a second pre-check would fail ExpectationsWereMet).
	mock.ExpectQuery("INTERVAL '7 days'").
		WithArgs(siteID, "contrast_failure:/pricing.html#h2.card-title").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	// The effect: one INSERT with the routed shape.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "render-audit", "build", "contrast_failure", "high",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			60, "css-patch-agent", "detected", sqlmock.AnyArg(),
			"contrast_failure:/pricing.html#h2.card-title", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-1",
			"contrast": []map[string]interface{}{
				{
					"url": "https://example.com/pricing.html", "tag": "h2",
					"class": "card-title muted", "text": "Plans",
					"fg": "#111111", "bg": "#0f0f0f",
					"ratio": 1.2, "need": 4.5, "font_px": 20, "over_image": false,
				},
				{
					"url": "https://example.com/index.html", "tag": "p",
					"class": "hero-sub", "fg": "#ffffff", "bg": "#888888",
					"ratio": 2.9, "need": 4.5, "over_image": true,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["inserted"] != 1 || m["over_image_reported"] != 1 {
		t.Fatalf("want inserted=1 over_image_reported=1, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestWriteRenderAuditFindings_LockedCulpritIsSkippedAndCounted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// A locked component whose markup carries the finding's class token.
	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).
			AddRow(`<section><h2 class="card-title">Locked</h2></section>`))
	// No pre-check, no INSERT — the tx opens and commits empty. Asserting the
	// commit (an effect) rather than "no insert happened" keeps this non-vacuous:
	// an unexpected INSERT fails ExpectationsWereMet loudly.
	mock.ExpectBegin()
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-2",
			"contrast": []map[string]interface{}{
				{
					"url": "https://example.com/about.html", "tag": "h2",
					"class": "card-title", "fg": "#222222", "bg": "#111111",
					"ratio": 1.1, "need": 4.5, "over_image": false,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["skipped_locked"] != 1 || m["inserted"] != 0 {
		t.Fatalf("want skipped_locked=1 inserted=0, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestWriteRenderAuditFindings_BrokenImagesAttributedAndNot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	assetID := uuid.New()

	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))

	// hero.jpg resolves to an assets row; ghost.jpg does not.
	mock.ExpectQuery("FROM assets").
		WithArgs(siteID, "hero.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(assetID.String()))
	mock.ExpectQuery("FROM assets").
		WithArgs(siteID, "ghost.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectBegin()
	mock.ExpectQuery("INTERVAL '7 days'").
		WithArgs(siteID, "undeployed_asset:"+assetID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "render-audit", "build", "undeployed_asset", "medium",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			60, "asset-deployer", "detected", sqlmock.AnyArg(),
			"undeployed_asset:"+assetID.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: renderAuditCollected(siteID, map[string]interface{}{
			"run_id": "run-3",
			"broken_images": []map[string]interface{}{
				{"url": "https://example.com/index.html", "src": "/assets/images/hero.jpg"},
				{"url": "https://example.com/index.html", "src": "/assets/images/ghost.jpg"},
			},
		}),
	})
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["inserted"] != 1 || m["unattributed_images"] != 1 {
		t.Fatalf("want inserted=1 unattributed_images=1, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestWriteRenderAuditFindings_AbsentAuditIsAnError(t *testing.T) {
	// Absent ≠ malformed ≠ clean: a run whose audit never landed must FAIL the
	// step, not report zero findings written.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	_, err = WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"site_id": uuid.New().String()},
	})
	if err == nil || !strings.Contains(err.Error(), "has not run") {
		t.Fatalf("want a loud 'has not run' error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no queries should have run: %v", err)
	}
}

func TestWriteRenderAuditFindings_StillAwaitedIsAnError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	_, err = WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id": uuid.New().String(),
			// The await-signal shape the action's own request step returns,
			// with no .response yet.
			"render_audit": map[string]interface{}{"success": true, "request_id": "r-1"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "awaited or failed") {
		t.Fatalf("want a loud still-awaited error, got %v", err)
	}
}
