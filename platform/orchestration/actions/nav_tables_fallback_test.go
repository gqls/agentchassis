package actions

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// GetNavItems overloads a zero-row nav-table result. It can mean "this site
// predates the nav tables" (fall back to the pages table) or "this nav-table
// site has no items in the requested group" (a truthful empty answer). Before
// the /bugs_open/053 fix the fallback ran on both, so a site with a full nav
// but no legal pages had its footer's legal slot filled with every footer page.
// The gate is siteHasAnyNavItems: fall back ONLY when the site has no
// site_nav_items rows at all.
func TestGetNavItemsFallbackGate(t *testing.T) {
	logger := zap.NewNop()
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

	// Scenario A — the bug: a nav-table site (robot-hands) with zero legal items.
	// The empty tables answer is truthful; the gate finds nav rows exist, so NO
	// fallback runs and the legal slot is empty.
	t.Run("nav-table site with no legal items returns empty, no fallback", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(navRows()) // 0 legal rows
		mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).AddRow(true)) // site has other nav rows
		// No ExpectQuery("FROM pages") — the fallback must NOT run.

		got := GetNavItems(ctx, db, siteID, []string{NavGroupLegal}, false, 0, logger)
		if len(got) != 0 {
			t.Fatalf("expected 0 legal items, got %d: %+v", len(got), got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// Scenario B — the regression guard: a nav-table site (leopardess) WITH legal
	// items. They come straight from the tables path; the gate is never consulted.
	t.Run("nav-table site with legal items returns them, no gate/fallback", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(
			navRows("Privacy", "/privacy.html", "Terms", "/terms.html"))
		// No EXISTS gate, no pages fallback expected.

		got := GetNavItems(ctx, db, siteID, []string{NavGroupLegal}, false, 0, logger)
		if len(got) != 2 || got[0].Label != "Privacy" || got[1].URL != "/terms.html" {
			t.Fatalf("expected the 2 real legal items, got %+v", got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// Scenario C — a genuinely pre-nav-table site: no nav rows at all. The gate
	// returns false and the pages fallback IS the source of nav.
	t.Run("pre-nav-table site (no rows) falls back to pages", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(navRows())
		mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectQuery("FROM pages").WillReturnRows(
			navRows("Home", "/index.html", "About", "/about.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, false, 0, logger)
		if len(got) != 2 {
			t.Fatalf("expected 2 fallback items, got %d: %+v", len(got), got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	// Scenario A-loud — the anomaly guard (bugs_open/053 council revision). A
	// request that INCLUDES the primary group must never be empty on a nav-table
	// site; if it is, primary nav is missing (a sync/write defect), so returning
	// empty is logged LOUD (Warn), not silently — the silent-empty-render class
	// (016b §9). A lone legal group returning empty stays quiet (Debug).
	t.Run("primary-including empty on a nav-table site warns; legal-only empty stays quiet", func(t *testing.T) {
		run := func(groups []string) *observer.ObservedLogs {
			core, logs := observer.New(zapcore.DebugLevel)
			db, mock := newMockDB(t)
			mock.ExpectQuery("FROM site_nav_items ni").WillReturnRows(navRows()) // 0 rows for the group
			mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
				sqlmock.NewRows([]string{"exists"}).AddRow(true)) // site has nav rows elsewhere
			got := GetNavItems(ctx, db, siteID, groups, false, 0, zap.New(core))
			if len(got) != 0 {
				t.Fatalf("groups %v: expected empty, got %+v", groups, got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
			return logs
		}

		// primary-including request → exactly one Warn, no fallback
		if n := run([]string{NavGroupPrimary}).FilterLevelExact(zapcore.WarnLevel).Len(); n != 1 {
			t.Fatalf("primary-only: expected 1 Warn about missing primary nav, got %d", n)
		}
		if n := run([]string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}).FilterLevelExact(zapcore.WarnLevel).Len(); n != 1 {
			t.Fatalf("footer (primary+utility+legal): expected 1 Warn, got %d", n)
		}
		// lone legal request → no Warn (legitimately-empty group)
		if n := run([]string{NavGroupLegal}).FilterLevelExact(zapcore.WarnLevel).Len(); n != 0 {
			t.Fatalf("legal-only: expected 0 Warn (empty is legitimate), got %d", n)
		}
	})

	// Scenario D — older deployment without the nav tables at all. Both the tables
	// query and the existence gate error with "does not exist"; the gate treats
	// that as pre-nav-table and the pages fallback runs.
	t.Run("nav tables absent (does not exist) falls back to pages", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("FROM site_nav_items ni").
			WillReturnError(errors.New(`relation "site_nav_items" does not exist`))
		mock.ExpectQuery("SELECT EXISTS").
			WillReturnError(errors.New(`relation "site_nav_items" does not exist`))
		mock.ExpectQuery("FROM pages").WillReturnRows(
			navRows("Home", "/index.html"))

		got := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, false, 0, logger)
		if len(got) != 1 {
			t.Fatalf("expected 1 fallback item, got %d: %+v", len(got), got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

// ---------------------------------------------------------------------------
// navPriorityTier — the directory registers rank as hubs, not leftovers
// ---------------------------------------------------------------------------
//
// The property: a directory page must sort with the content hubs, not behind
// them. Tier decides who survives max_header_items truncation, so a tier-3
// directory is one newly-promoted page away from silently falling out of a
// header the owner deliberately put it in. Tested rather than trusted because
// the failure is invisible — nothing errors, the link just stops being there.
func TestNavPriorityTierRanksDirectoryPagesAsHubs(t *testing.T) {
	for _, name := range []string{"model-directory", "adoption-tracker", "protocol-tracker"} {
		if got := navPriorityTier(name, name); got != 2 {
			t.Errorf("navPriorityTier(%q) = %d, want 2 (content hub)", name, got)
		}
	}

	// Guard the boundaries the same call decides, so a future edit to the maps
	// cannot quietly promote or demote the surrounding cases either.
	if got := navPriorityTier("index", "content"); got != 1 {
		t.Errorf("navPriorityTier(index) = %d, want 1", got)
	}
	if got := navPriorityTier("blog", "blog-index"); got != 2 {
		t.Errorf("navPriorityTier(blog) = %d, want 2", got)
	}
	if got := navPriorityTier("some-landing-page", "content"); got != 3 {
		t.Errorf("navPriorityTier(some-landing-page) = %d, want 3 (secondary)", got)
	}
}
