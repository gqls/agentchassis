// FILE: platform/orchestration/actions/deploy_tool_identity_test.go
//
// bugs_open/080 / TL-010: deploy_tool used to hand-roll `<bare>` @
// /tools/<bare>.html while create_tool_component writes the canonical
// `tool-<bare>` @ /tools/<bare>/index.html — two surfaces disagreeing on the
// same logical page's identity. resolveToolPageIdentity converges them.
//
// The important one is ExistingLegacyRowKeepsItsIdentity: 12 live rows carry
// the legacy shape (measured 2026-08-03), and writing the canonical name while
// a legacy row stands would not conflict on (site_id, name) — it would INSERT
// A SECOND ROW, the exact duplicate-page failure the bug is about.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestResolveToolPageIdentity_NewToolGetsCanonicalShape(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT name, url FROM pages").
		WithArgs(sqlmock.AnyArg(), "ab-test-calculator", "tool-ab-test-calculator").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"})) // no existing row

	name, url, err := resolveToolPageIdentity(context.Background(), db, uuid.New(), "tool-ab-test-calculator", false)
	if err != nil {
		t.Fatalf("resolveToolPageIdentity: %v", err)
	}
	if name != "tool-ab-test-calculator" || url != "/tools/ab-test-calculator/index.html" {
		t.Errorf("got (%q, %q), want canonical (tool-ab-test-calculator, /tools/ab-test-calculator/index.html)", name, url)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestResolveToolPageIdentity_ExistingLegacyRowKeepsItsIdentity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// A live legacy-shape row: name=<bare>, flat URL. It must be reused
	// verbatim — not re-keyed, not re-urled.
	mock.ExpectQuery("SELECT name, url FROM pages").
		WithArgs(sqlmock.AnyArg(), "password-entropy", "tool-password-entropy").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}).
			AddRow("password-entropy", "/tools/password-entropy.html"))

	name, url, err := resolveToolPageIdentity(context.Background(), db, uuid.New(), "tool-password-entropy", false)
	if err != nil {
		t.Fatalf("resolveToolPageIdentity: %v", err)
	}
	if name != "password-entropy" || url != "/tools/password-entropy.html" {
		t.Errorf("got (%q, %q), want the existing row's identity untouched", name, url)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestResolveToolPageIdentity_ExistingDoublePrefixedURLIsNotMoved(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The 14-row middle shape: canonical name, legacy flat URL. The name
	// matches, so the URL must stay where it is — a redeploy is not a move.
	mock.ExpectQuery("SELECT name, url FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}).
			AddRow("tool-loot-table-balancer", "/tools/tool-loot-table-balancer.html"))

	name, url, err := resolveToolPageIdentity(context.Background(), db, uuid.New(), "tool-loot-table-balancer", false)
	if err != nil {
		t.Fatalf("resolveToolPageIdentity: %v", err)
	}
	if name != "tool-loot-table-balancer" || url != "/tools/tool-loot-table-balancer.html" {
		t.Errorf("got (%q, %q), want the stored identity untouched", name, url)
	}
}

func TestResolveToolPageIdentity_FlatSiteNewToolGetsFlatShape(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT name, url FROM pages").
		WithArgs(sqlmock.AnyArg(), "ab-test-calculator", "tool-ab-test-calculator").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"})) // no existing row

	name, url, err := resolveToolPageIdentity(context.Background(), db, uuid.New(), "tool-ab-test-calculator", true)
	if err != nil {
		t.Fatalf("resolveToolPageIdentity: %v", err)
	}
	if name != "tool-ab-test-calculator" || url != "/tools/ab-test-calculator.html" {
		t.Errorf("got (%q, %q), want flat (tool-ab-test-calculator, /tools/ab-test-calculator.html)", name, url)
	}
}

func TestResolveToolPageIdentity_FlagIrrelevantWhenRowExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// An existing NESTED row on a site later flagged flat: the stored
	// identity still wins — the flag shapes new synthesis only, it never
	// moves a live page.
	mock.ExpectQuery("SELECT name, url FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}).
			AddRow("tool-password-entropy", "/tools/password-entropy/index.html"))

	name, url, err := resolveToolPageIdentity(context.Background(), db, uuid.New(), "tool-password-entropy", true)
	if err != nil {
		t.Fatalf("resolveToolPageIdentity: %v", err)
	}
	if name != "tool-password-entropy" || url != "/tools/password-entropy/index.html" {
		t.Errorf("got (%q, %q), want the stored nested identity untouched by the flat flag", name, url)
	}
}

func TestResolveToolPageIdentity_EmptyFunctionRefused(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	if _, _, err := resolveToolPageIdentity(context.Background(), db, uuid.New(), "", false); err == nil {
		t.Fatal("empty tool function must be refused, not hand-rolled")
	}
}
