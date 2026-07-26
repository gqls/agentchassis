package actions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Chrome renders onto EVERY page of a site, so a nav item whose target was
// never built is a site-wide 404. That is bugs_open/049 mechanism 2: one
// gaswholesalers utility item ("Pricing Framework" -> /fuel-pricing-framework.html,
// needs_rebuild, deployed_at NULL) produced 28 broken anchors, one per page.
//
// The subtlety these tests exist to protect is that the OBVIOUS fix is wrong.
// Filtering on build_status='deployed' drops the 34 fleet pages that are
// needs_rebuild but were deployed once and still serve their old artefact.
// Only "never deployed" predicts a 404 — see datahelpers.NeverDeployedPagePredicate.
func TestGetNavItemsFetchableVisibility(t *testing.T) {
	ctx := context.Background()
	siteID := uuid.New()

	newMockDB := func(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		t.Cleanup(func() { db.Close() })
		return db, mock
	}

	navRows := func(pairs ...string) *sqlmock.Rows {
		r := sqlmock.NewRows([]string{"label", "url"})
		for i := 0; i+1 < len(pairs); i += 2 {
			r.AddRow(pairs[i], pairs[i+1])
		}
		return r
	}

	pageRows := func(urls ...string) *sqlmock.Rows {
		r := sqlmock.NewRows([]string{"url"})
		for _, u := range urls {
			r.AddRow(u)
		}
		return r
	}

	// Matches the fetchable-page query and NOTHING that keys on build_status
	// alone. If someone "simplifies" the predicate to build_status='deployed',
	// every test using this expectation fails to match and the suite goes red.
	fetchableQuery := regexp.QuoteMeta("NOT (deployed_at IS NULL")

	observed := func() (*zap.Logger, *observer.ObservedLogs) {
		core, logs := observer.New(zapcore.WarnLevel)
		return zap.New(core), logs
	}

	// THE REGRESSION PIN. aao /tools.html is needs_rebuild with deployed_at set
	// and returns 200 live; gaswholesalers /fuel-pricing-framework.html is
	// needs_rebuild with deployed_at NULL and returns 404. Same build_status,
	// opposite outcome. A build_status filter would delete the working one.
	t.Run("a needs_rebuild page that was deployed once STAYS in the nav", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger, logs := observed()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "Tools", "/tools.html", "About", "/about.html"))
		// The fetchable set is what the predicate returns: all three, because
		// "needs_rebuild but deployed once" is fetchable.
		mock.ExpectQuery(fetchableQuery).WillReturnRows(
			pageRows("/index.html", "/tools.html", "/about.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(got) != 3 {
			t.Fatalf("expected all 3 items kept, got %d: %+v", len(got), got)
		}
		if n := logs.Len(); n != 0 {
			t.Fatalf("expected no warnings when nothing is dropped, got %d: %+v", n, logs.All())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// The defect itself.
	t.Run("a never-deployed target is dropped and named in a warning", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger, logs := observed()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "Pricing Framework", "/fuel-pricing-framework.html"))
		mock.ExpectQuery(fetchableQuery).WillReturnRows(pageRows("/index.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupUtility}, NavFetchableOnly, 0, logger)
		if len(got) != 1 || got[0].URL != "/index.html" {
			t.Fatalf("expected only /index.html kept, got %+v", got)
		}
		if logs.Len() != 1 {
			t.Fatalf("expected exactly 1 warning, got %d: %+v", logs.Len(), logs.All())
		}
		// The dropped URL must be in the log, or the operator cannot tell which
		// link vanished from a live site.
		if !strings.Contains(fmt.Sprint(logs.All()[0].ContextMap()["dropped_urls"]), "fuel-pricing-framework") {
			t.Fatalf("warning does not name the dropped URL: %+v", logs.All()[0].ContextMap())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// THE FIRST-BUILD GUARD. Do not delete this: RenderSiteComponentsAction
	// skips re-rendering once rendered_html is non-empty, so a nav emptied
	// during a first build would persist indefinitely.
	//
	// This test caught a real defect in the first draft of the fix. The guard
	// was written as "if nothing survived, degrade" — but loadFetchablePageSet
	// always injects the site root, so on a first build the "Home" item DID
	// survive, the guard never fired, and the other items were dropped: a
	// brand-new site would have frozen a one-item header into its chrome. The
	// guard therefore keys on the site having no deployed pages, which is the
	// thing actually being detected, rather than on the surviving count, which
	// is a proxy that the root injection breaks.
	t.Run("a site with no deployed pages keeps its UNFILTERED nav", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger, logs := observed()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "About", "/about.html", "Contact", "/contact.html"))
		mock.ExpectQuery(fetchableQuery).WillReturnRows(pageRows()) // nothing deployed yet

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(got) != 3 {
			t.Fatalf("first build must keep all 3 items, got %d: %+v", len(got), got)
		}
		if logs.Len() != 1 {
			t.Fatalf("expected exactly 1 warning, got %d: %+v", logs.Len(), logs.All())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// An infrastructure failure must never empty a live site's chrome.
	t.Run("a failed fetchable lookup degrades to the unfiltered nav", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger, logs := observed()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "About", "/about.html"))
		mock.ExpectQuery(fetchableQuery).WillReturnError(errors.New("connection reset"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(got) != 2 {
			t.Fatalf("expected both items on loader failure, got %d: %+v", len(got), got)
		}
		if logs.Len() != 1 {
			t.Fatalf("expected exactly 1 warning, got %d: %+v", logs.Len(), logs.All())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// NavAllItems is the planning/prompt path and must cost nothing extra.
	t.Run("NavAllItems issues no fetchable query at all", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger := zap.NewNop()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "Shop", "/shop.html"))
		// No ExpectQuery for pages — ExpectationsWereMet proves none ran.

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavAllItems, 0, logger)
		if len(got) != 2 {
			t.Fatalf("NavAllItems must not filter, got %d: %+v", len(got), got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// Only page links are checked against the pages table.
	t.Run("external and anchor items survive without a pages row", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger := zap.NewNop()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "Partner", "https://example.org/", "Top", "#top"))
		mock.ExpectQuery(fetchableQuery).WillReturnRows(pageRows("/index.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(got) != 3 {
			t.Fatalf("non-page items must survive, got %d: %+v", len(got), got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// maxItems means "N usable items". A SQL LIMIT capped BEFORE filtering, so
	// an unfetchable item inside the first N silently shortened the nav.
	t.Run("maxItems caps after filtering, not before", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger := zap.NewNop()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(navRows(
			"A", "/a.html", "Dead1", "/dead1.html", "B", "/b.html", "Dead2", "/dead2.html",
			"C", "/c.html", "D", "/d.html", "E", "/e.html", "F", "/f.html"))
		mock.ExpectQuery(fetchableQuery).WillReturnRows(
			pageRows("/a.html", "/b.html", "/c.html", "/d.html", "/e.html", "/f.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 6, logger)
		if len(got) != 6 {
			t.Fatalf("expected 6 usable items, got %d: %+v", len(got), got)
		}
		for _, item := range got {
			if strings.HasPrefix(item.URL, "/dead") {
				t.Fatalf("unfetchable item survived the cap: %+v", got)
			}
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// Parity with check_phantom_internal_links, which also injects the root.
	t.Run("the site root is always fetchable", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger := zap.NewNop()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Home", "/index.html", "About", "/about.html"))
		// The pages query returns no index row at all.
		mock.ExpectQuery(fetchableQuery).WillReturnRows(pageRows("/about.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(got) != 2 {
			t.Fatalf("root must survive an absent index row, got %+v", got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// The /bugs_open/053 gate and this filter must compose: a pre-nav-table
	// site takes the pages fallback AND still gets its dead links filtered.
	t.Run("pages fallback is filtered too", func(t *testing.T) {
		db, mock := newMockDB(t)
		logger, logs := observed()

		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(navRows())
		mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).AddRow(false)) // pre-nav-table site
		mock.ExpectQuery("COALESCE\\(nav_label").WillReturnRows(
			navRows("Home", "/index.html", "Ghost", "/ghost.html"))
		mock.ExpectQuery(fetchableQuery).WillReturnRows(pageRows("/index.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
		if len(got) != 1 || got[0].URL != "/index.html" {
			t.Fatalf("expected the fallback to be filtered, got %+v", got)
		}
		if logs.Len() != 1 {
			t.Fatalf("expected exactly 1 warning, got %d: %+v", logs.Len(), logs.All())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}
