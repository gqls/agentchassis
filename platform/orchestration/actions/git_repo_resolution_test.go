package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// resolveGitRepoNameDB is the deploy-target selector for every git commit. The DB
// fallback exists because most workflows never load the site record, and without it a
// VM-hosted site's artefacts silently land in the default "sites" repo (→ B2).
func TestResolveGitRepoNameDB(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	newMockDB := func(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		t.Cleanup(func() { db.Close() })
		return db, mock
	}

	t.Run("explicit step config wins over everything", func(t *testing.T) {
		db, _ := newMockDB(t) // no expectations: DB must not be queried
		got := resolveGitRepoNameDB(ctx, db,
			map[string]interface{}{"repo_name": "explicit-repo"},
			map[string]interface{}{"site_record": map[string]interface{}{"github_repo": "vm-sites"}},
			"relojistas.com", logger)
		if got != "explicit-repo" {
			t.Fatalf("got %q, want explicit-repo", got)
		}
	})

	t.Run("collected site_record.github_repo wins over DB", func(t *testing.T) {
		db, _ := newMockDB(t) // no expectations: DB must not be queried
		got := resolveGitRepoNameDB(ctx, db,
			map[string]interface{}{},
			map[string]interface{}{"site_record": map[string]interface{}{"github_repo": "vm-sites"}},
			"relojistas.com", logger)
		if got != "vm-sites" {
			t.Fatalf("got %q, want vm-sites", got)
		}
	})

	t.Run("DB fallback used when site record absent (the misroute fix)", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT github_repo FROM sites").
			WithArgs("relojistas.com").
			WillReturnRows(sqlmock.NewRows([]string{"github_repo"}).AddRow("vm-sites"))
		got := resolveGitRepoNameDB(ctx, db, map[string]interface{}{}, map[string]interface{}{}, "relojistas.com", logger)
		if got != "vm-sites" {
			t.Fatalf("got %q, want vm-sites", got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NULL github_repo in DB falls back to default", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT github_repo FROM sites").
			WithArgs("plain-site.com").
			WillReturnRows(sqlmock.NewRows([]string{"github_repo"}).AddRow(nil))
		got := resolveGitRepoNameDB(ctx, db, map[string]interface{}{}, map[string]interface{}{}, "plain-site.com", logger)
		if got != "sites" {
			t.Fatalf("got %q, want sites", got)
		}
	})

	t.Run("unknown domain falls back to default", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT github_repo FROM sites").
			WithArgs("nosuch.com").
			WillReturnRows(sqlmock.NewRows([]string{"github_repo"}))
		got := resolveGitRepoNameDB(ctx, db, map[string]interface{}{}, map[string]interface{}{}, "nosuch.com", logger)
		if got != "sites" {
			t.Fatalf("got %q, want sites", got)
		}
	})

	t.Run("DB error falls back to default rather than failing the commit", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT github_repo FROM sites").
			WithArgs("relojistas.com").
			WillReturnError(sql.ErrConnDone)
		got := resolveGitRepoNameDB(ctx, db, map[string]interface{}{}, map[string]interface{}{}, "relojistas.com", logger)
		if got != "sites" {
			t.Fatalf("got %q, want sites", got)
		}
	})

	t.Run("nil DB and empty domain are safe", func(t *testing.T) {
		if got := resolveGitRepoNameDB(ctx, nil, map[string]interface{}{}, map[string]interface{}{}, "relojistas.com", logger); got != "sites" {
			t.Fatalf("nil db: got %q, want sites", got)
		}
		db, _ := newMockDB(t)
		if got := resolveGitRepoNameDB(ctx, db, map[string]interface{}{}, map[string]interface{}{}, "", logger); got != "sites" {
			t.Fatalf("empty domain: got %q, want sites", got)
		}
	})
}
