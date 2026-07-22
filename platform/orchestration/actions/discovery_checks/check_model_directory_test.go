// FILE: platform/orchestration/actions/discovery_checks/check_model_directory_test.go
//
// The property that matters: these checks must stay silent until BOTH a
// site has opted in AND the global registry actually has publishable data —
// raising a work item for an empty section would push content-gap-planner
// to build a section with nothing to show. Each gate is tested in isolation
// so a future change to one can't silently widen or narrow the other.

package discovery_checks

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestMissingModelDirectorySectionCheck_NotOptedIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).WillReturnError(sql.ErrNoRows)

	res, err := (&MissingModelDirectorySectionCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items when the site hasn't opted in, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectorySectionCheck_OptedInButRegistryEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM directory_claims`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	res, err := (&MissingModelDirectorySectionCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Opted in, but discovery hasn't produced anything yet — must wait, not
	// build an empty section.
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items with an empty registry, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectorySectionCheck_RaisesFindingWhenDataExistsButNoSection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM directory_claims`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM page_components`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	res, err := (&MissingModelDirectorySectionCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("expected exactly one work item, got %d", len(res.WorkItems))
	}
	item := res.WorkItems[0]
	if item.ItemType != "missing_model_directory_section" {
		t.Errorf("item_type = %q, want missing_model_directory_section", item.ItemType)
	}
	if item.HandlerAgent != "content-gap-planner" {
		t.Errorf("handler = %q, want content-gap-planner", item.HandlerAgent)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectorySectionCheck_SilentWhenSectionAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM directory_claims`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM page_components`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	res, err := (&MissingModelDirectorySectionCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items when the section already exists, got %d", len(res.WorkItems))
	}
}

func TestMissingModelDirectoryPageCheck_RequiresSeparatePageFlag(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// recommended=true but separate_page is absent (defaults false) — must
	// stay silent, same as missing_news_page's identical gate.
	specJSON := `{"content_features":{"model_directory":{"recommended":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))

	res, err := (&MissingModelDirectoryPageCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("expected no work items without separate_page=true, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMissingModelDirectoryPageCheck_RaisesFinding(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	specJSON := `{"content_features":{"model_directory":{"recommended":true,"separate_page":true}}}`
	mock.ExpectQuery(`SELECT data FROM site_specs`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(specJSON)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM directory_claims`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM pages`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT domain FROM sites`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("ai-agent-orchestration.com"))

	res, err := (&MissingModelDirectoryPageCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("expected exactly one work item, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].ItemType != "missing_model_directory_page" {
		t.Errorf("item_type = %q, want missing_model_directory_page", res.WorkItems[0].ItemType)
	}
}
