// FILE: platform/orchestration/actions/retract_page_graph_test.go
//
// The retraction graph guards (bugs_open/098, owner directive 2026-08-03).
//
// These are all NEGATIVE properties — "a linked page is not retracted", "a nav
// row is retired" — and a negative is exactly what a broken query also produces.
// So every guard here is proved by CONTRAST: the same call, the same fixture,
// changing only the thing the guard reads. A test that only asserts "0 paths
// returned" would pass against a query that always returns nothing.
//
// The SQL itself cannot be checked by the compiler and is not checked here
// either — it was PREPAREd against the live schema and then RUN, with a
// positive control on a page that IS linked (robot-hands /contact.html →
// 4 body refs, 2 chrome, 1 nav) and the retraction target (→ 0/0/0/0). Both
// runs are recorded in the workstream RUNBOOK. What these tests own is the
// DECISION taken on the query's answer, which is where the behaviour lives.
package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newRetractMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// TestDeactivateNavItemsUsesInactiveNotArchived. 098's own repair note records
// the convention trap: the pre-existing non-live nav rows fleet-wide are
// 'inactive'. 'archived' is equally excluded by every reader's `= 'active'`
// filter and so would LOOK correct, while being invisible to anyone querying by
// convention. Asserting the literal is the only way that stays true.
func TestDeactivateNavItemsUsesInactiveNotArchived(t *testing.T) {
	db, mock := newRetractMockDB(t)
	mock.ExpectExec(`UPDATE site_nav_items`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	n, err := deactivateNavItems(context.Background(), db,
		[]inboundRef{{NavItem: "a", Source: "nav"}, {NavItem: "b", Source: "nav"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("deactivateNavItems: %v", err)
	}
	if n != 2 {
		t.Errorf("retired %d rows, want 2", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	// The literal the query must contain. sqlmock matches on a regex, so this
	// asserts the statement really sets 'inactive' — a change to 'archived'
	// fails here rather than silently shipping a second spelling.
	db2, mock2 := newRetractMockDB(t)
	mock2.ExpectExec(`SET status = 'inactive'`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := deactivateNavItems(context.Background(), db2,
		[]inboundRef{{NavItem: "a"}}, zap.NewNop()); err != nil {
		t.Fatalf("deactivateNavItems (literal check): %v", err)
	}
	if err := mock2.ExpectationsWereMet(); err != nil {
		t.Errorf("the UPDATE no longer sets 'inactive': %v", err)
	}
}

// TestDeactivateNavItemsNoopsOnEmpty — the common case is a retraction with no
// nav rows at all, and it must not issue an UPDATE with an empty id list (which
// would be a full-table predicate away from a disaster). sqlmock proves the
// absence: any ExecContext at all fails an empty expectation set.
func TestDeactivateNavItemsNoopsOnEmpty(t *testing.T) {
	db, mock := newRetractMockDB(t)
	n, err := deactivateNavItems(context.Background(), db, nil, zap.NewNop())
	if err != nil || n != 0 {
		t.Fatalf("empty nav set: got (%d, %v), want (0, nil)", n, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a statement was issued for an empty nav set: %v", err)
	}
}

// TestAuditRetractionSkipsQueriesForEmptyBatch — the same shape one level up.
func TestAuditRetractionSkipsQueriesForEmptyBatch(t *testing.T) {
	db, mock := newRetractMockDB(t)
	g, err := auditRetraction(context.Background(), db, uuid.New(), nil, zap.NewNop())
	if err != nil {
		t.Fatalf("auditRetraction: %v", err)
	}
	if g == nil || len(g.Editorial) != 0 || len(g.Nav) != 0 || len(g.Stranded) != 0 {
		t.Errorf("empty batch produced findings: %+v", g)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("queries issued for an empty batch: %v", err)
	}
}

// TestAuditRetractionSplitsEditorialFromNav is the load-bearing distinction:
// a body-copy or chrome reference BLOCKS a retraction (repairing prose is a
// content decision), a nav row does NOT (it is a pointer, and leaving it is
// the relojistas defect). If these ever collapse into one list, either
// retractions start rewriting copy or they start leaving dead nav.
func TestAuditRetractionSplitsEditorialFromNav(t *testing.T) {
	db, mock := newRetractMockDB(t)
	siteID := uuid.New()
	pageID := uuid.New().String()

	mock.ExpectQuery(`FROM unnest`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "slot", "url"}).
			AddRow("how-it-works", "hero", "/gone.html"))
	mock.ExpectQuery(`site_components`).
		WillReturnRows(sqlmock.NewRows([]string{"slot", "url"}).
			AddRow("footer", "/gone.html"))
	mock.ExpectQuery(`site_nav_items`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "url"}).
			AddRow("nav-1", "Gone", "/gone.html"))
	mock.ExpectQuery(`WITH outbound`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "status"}))

	g, err := auditRetraction(context.Background(), db, siteID,
		[]retractionCandidate{{PageID: pageID, URL: "/gone.html", Status: "archived"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("auditRetraction: %v", err)
	}

	if len(g.Editorial) != 2 {
		t.Errorf("editorial refs = %d, want 2 (one body, one chrome): %+v", len(g.Editorial), g.Editorial)
	}
	var sawPage, sawChrome bool
	for _, r := range g.Editorial {
		switch r.Source {
		case "page":
			sawPage = true
		case "chrome":
			sawChrome = true
		case "nav":
			t.Errorf("a nav row was classified as EDITORIAL — it would block a retraction instead of being retired: %+v", r)
		}
	}
	if !sawPage || !sawChrome {
		t.Errorf("editorial must carry both body and chrome refs, got %+v", g.Editorial)
	}
	if len(g.Nav) != 1 || g.Nav[0].Source != "nav" {
		t.Errorf("nav refs = %+v, want exactly one nav-sourced ref", g.Nav)
	}
}

// TestAuditRetractionReportsStrandedTargets — the third obligation. A target
// that loses its last referrer is REPORTED, never repaired: what to do about a
// page nothing links to is a site-shaping decision that
// discovery_checks/check_orphan_pages.go already owns.
func TestAuditRetractionReportsStrandedTargets(t *testing.T) {
	db, mock := newRetractMockDB(t)
	mock.ExpectQuery(`FROM unnest`).WillReturnRows(sqlmock.NewRows([]string{"name", "slot", "url"}))
	mock.ExpectQuery(`site_components`).WillReturnRows(sqlmock.NewRows([]string{"slot", "url"}))
	mock.ExpectQuery(`site_nav_items`).WillReturnRows(sqlmock.NewRows([]string{"id", "label", "url"}))
	mock.ExpectQuery(`WITH outbound`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "status"}).
			AddRow("only-child", "/only-child.html", "active"))

	g, err := auditRetraction(context.Background(), db, uuid.New(),
		[]retractionCandidate{{PageID: uuid.New().String(), URL: "/parent.html", Status: "archived"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("auditRetraction: %v", err)
	}
	if len(g.Stranded) != 1 || g.Stranded[0].URL != "/only-child.html" {
		t.Fatalf("stranded = %+v, want the one target that loses its last referrer", g.Stranded)
	}
	// And it must NOT have been promoted into a blocking list — reporting is
	// the whole point.
	if len(g.Editorial) != 0 {
		t.Errorf("a stranded OUTBOUND target must not block the retraction: %+v", g.Editorial)
	}
}

// TestAuditRetractionPropagatesQueryErrors. A failed audit must fail the
// retraction, never be treated as "nothing found" — guessing clear on an error
// would delete a live-linked page and report success, which is the
// silent-success shape this whole bug is about.
func TestAuditRetractionPropagatesQueryErrors(t *testing.T) {
	db, mock := newRetractMockDB(t)
	mock.ExpectQuery(`FROM unnest`).WillReturnError(sql.ErrConnDone)

	if _, err := auditRetraction(context.Background(), db, uuid.New(),
		[]retractionCandidate{{PageID: uuid.New().String(), URL: "/x.html"}}, zap.NewNop()); err == nil {
		t.Fatal("a failing inbound scan must fail the audit, not report a clean graph")
	}
}
