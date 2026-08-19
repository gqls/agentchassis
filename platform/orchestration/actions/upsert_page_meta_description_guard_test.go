// FILE: platform/orchestration/actions/upsert_page_meta_description_guard_test.go
//
// bugs_open/320, mechanism M2 — `upsertPage`'s ON CONFLICT clause used to read
//
//	meta_description = EXCLUDED.meta_description
//
// three lines below a `nav_label` clause that WAS guarded
// (`COALESCE(NULLIF(pages.nav_label,''), EXCLUDED.nav_label)`). Since
// `metaDescription` defaults to "" whenever the incoming page map omits the key —
// and `build-site-planner`'s page object never carried the key at all until
// migration 485 — **every replan of an existing page wrote a blank over whatever
// description was there.**
//
// It is not a hypothetical. `site_snapshots.pages_snapshot` carries the column,
// and four robot-hands.com pages holding 97/120/169/329 characters on 2026-04-10
// read 0 today, all of them `built_from_plan_version IS NOT NULL`.
//
// THE WHOLE PACKAGE'S TESTS PASSED BOTH BEFORE AND AFTER THE FIX, which is the
// reason this file exists: the behaviour was entirely untested, so nothing would
// have objected if a later edit reverted it. The guarded neighbour is exactly
// what made the unguarded line read as deliberate — a reviewer sees a COALESCE
// two lines up and assumes the difference is meant.
package actions

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TestUpsertPage_MetaDescriptionIsProtectedFromABlank asserts the SQL itself,
// following the local precedent in record_vision_finding_action_test.go: the
// property is a shape of the statement, and a behavioural assertion through
// sqlmock cannot see which side of the COALESCE won.
//
// MUTATION THAT KILLS IT (verified by running it): restore
// `meta_description = EXCLUDED.meta_description`. The regexp stops matching and
// sqlmock reports an unexpected query.
func TestUpsertPage_MetaDescriptionIsProtectedFromABlank(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The guard, as it must appear in the statement. Written out in full rather
	// than as a loose "COALESCE" match, so a COALESCE that protected the WRONG
	// direction (existing over incoming, which would freeze the column for ever)
	// would still fail this test.
	guard := regexp.QuoteMeta(
		"meta_description = COALESCE(NULLIF(EXCLUDED.meta_description, ''), pages.meta_description)")

	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery(guard).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "site_id", "name", "url", "title", "page_type",
			"nav_label", "nav_order", "in_header", "in_footer", "status",
		}).AddRow(pageID, siteID, "about", "/about.html", "About", "content",
			"About", 10, true, true, "active"),
	)

	// The page map deliberately OMITS meta_description — this is the exact input
	// that caused the damage, and after migration 485 it is also what an older
	// plan replayed against a newer page still looks like.
	page := map[string]interface{}{
		"name":      "about",
		"url":       "/about.html",
		"title":     "About",
		"page_type": "content",
	}

	if _, err := upsertPage(context.Background(), db, siteID, page, 0,
		uuid.NullUUID{}, zap.NewNop()); err != nil {
		t.Fatalf("upsertPage: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the ON CONFLICT clause no longer protects meta_description from a blank: %v", err)
	}
}

// TestUpsertPage_StillBindsAnEmptyStringForAMissingKey pins the OTHER half of the
// mechanism, so the fix cannot be misread as "the empty default was removed".
// It was not: `GetStringField(page,"meta_description","")` still binds "", and
// that is fine now precisely BECAUSE the SQL refuses to write it over a value.
// If a later edit made the default non-empty instead, this test says so.
//
// MUTATION THAT KILLS IT: change the default in upsertPage to anything non-blank.
func TestUpsertPage_StillBindsAnEmptyStringForAMissingKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	// $10 is meta_description in the INSERT's argument order (site_id, name, url,
	// title, page_type, nav_label, nav_order, in_header, in_footer,
	// meta_description, sections, built_from_plan_version).
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO pages")).
		WithArgs(siteID, "about", "/about.html", "About", "content",
			"About", 10, true, true, "", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "site_id", "name", "url", "title", "page_type",
			"nav_label", "nav_order", "in_header", "in_footer", "status",
		}).AddRow(pageID, siteID, "about", "/about.html", "About", "content",
			"About", 10, true, true, "active"))

	page := map[string]interface{}{
		"name":      "about",
		"url":       "/about.html",
		"title":     "About",
		"page_type": "content",
		"nav_label": "About",
		"nav_order": 10,
	}

	if _, err := upsertPage(context.Background(), db, siteID, page, 0,
		uuid.NullUUID{}, zap.NewNop()); err != nil {
		t.Fatalf("upsertPage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("argument binding changed: %v", err)
	}
}

var _ = sql.ErrNoRows // keep database/sql referenced if the file is trimmed later
