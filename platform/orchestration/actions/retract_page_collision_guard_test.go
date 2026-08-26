// FILE: platform/orchestration/actions/retract_page_collision_guard_test.go
//
// REGRESSION COVER FOR THE HOISTED COLLISION GUARD — written because the
// council's guardian seat asked for it and was right (corr 6ce98a66, medium):
//
//	"retract_page_deployment_action.go is a live production retraction pipeline.
//	 The plan delegates its collision-guard logic to the new
//	 datahelpers.ActivePageFilePaths 'verbatim,' but no test in this plan directly
//	 asserts the hoisted function's output matches the pre-refactor query for
//	 retraction's own guard-3 scenarios (only the NEW check's collision test
//	 exercises it). A refactor of a load-bearing write-path guard with zero direct
//	 regression coverage on the modified consumer is a containable but real gap."
//
// The gap was real. The 359 lane's own test exercises the guard from the
// DETECTOR's side, where the failure is a false positive on a flag-only item. On
// THIS side the same guard decides whether a live page's HTML file gets DELETED
// from the deploy repo, and the two directions are not equally forgiving.
//
// These assert the three properties guard 3 actually rests on — each stated as
// the damage it prevents, so a later reader can tell a deliberate change from a
// regression:
//
//  1. the map is keyed by the DERIVED FILE PATH, not the url string, so "/foo/"
//     and "/foo/index.html" collide as the one file they really are. The weaker
//     url-equality test finds no collisions on this estate and would let the
//     retraction delete a live page's artefact.
//  2. the lifecycle arm is present and the BUILD arm is ABSENT. An active page's
//     path is protected before it has ever deployed; adding a shipped arm here
//     would narrow the protected set and let a retraction delete the path a live
//     page is about to publish to.
//  3. a url that designates no file of its own (fragment, query, off-origin)
//     claims NO path. If it claimed one, an unrelated archived page sharing that
//     path would be wrongly refused — and idea.uk's "/tools.html#audience-check"
//     is a live row of exactly that shape, where "/tools.html" belongs to a
//     different page.
package actions

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestRetractionCollisionGuardKeysOnTheDerivedFilePath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(`FROM pages`).WithArgs(siteID).WillReturnRows(
		sqlmock.NewRows([]string{"name", "url"}).
			// the collision that matters: a directory url and its index document
			AddRow("learning-center", "/learning-center/index.html").
			AddRow("home", "/").
			// designates no file of its own — must claim nothing
			AddRow("tool-audience-check", "/tools.html#audience-check").
			AddRow("params", "/search.html?q=1").
			AddRow("offsite", "https://elsewhere.test/x.html").
			AddRow("blank", ""))

	got, err := loadActivePageFilePaths(context.Background(), db, siteID)
	if err != nil {
		t.Fatalf("loadActivePageFilePaths: %v", err)
	}

	// PROPERTY 1 — derived path, so a directory url is claimed as its index file.
	if owner, ok := got["learning-center/index.html"]; !ok || owner != "learning-center" {
		t.Fatalf("an active page's DERIVED file path must be claimed (an archived /learning-center/ "+
			"resolves to the same file and must be refused); got %v", got)
	}
	if owner, ok := got["index.html"]; !ok || owner != "home" {
		t.Fatalf("the site root must claim index.html; got %v", got)
	}

	// PROPERTY 3 — a url designating no file of its own claims nothing. If it
	// claimed "tools.html", an archived page legitimately owning /tools.html
	// would be refused retraction for ever.
	if owner, ok := got["tools.html"]; ok {
		t.Fatalf("a FRAGMENT url must claim no path — /tools.html belongs to a different page; "+
			"it was claimed by %q", owner)
	}
	if _, ok := got["search.html"]; ok {
		t.Fatal("a QUERY url must claim no path")
	}
	if len(got) != 2 {
		t.Fatalf("exactly the two resolvable urls should claim a path; got %d: %v", len(got), got)
	}
}

// TestRetractionCollisionGuardHasTheLifecycleArmAndNoBuildArm pins the predicate
// SHAPE, because the damage from getting it wrong is invisible in any output: a
// narrowed set silently un-protects live pages, and the retraction then deletes
// their artefacts. Asserted on the SQL the mock is asked for, which is the only
// place the shape is observable without a database.
func TestRetractionCollisionGuardHasTheLifecycleArmAndNoBuildArm(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// The lifecycle arm MUST be there…
	mock.ExpectQuery(`status = 'active'`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}).AddRow("home", "/"))

	if _, err := loadActivePageFilePaths(context.Background(), db, siteID); err != nil {
		t.Fatalf("the query must carry the lifecycle arm (status = 'active'): %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("lifecycle arm missing from the collision-guard query: %v", err)
	}

	// …and the BUILD arm must NOT be, because an active page's path is protected
	// before it has ever deployed. This is asserted against the source rather
	// than the mock: a negative about SQL text cannot be expressed as an
	// expectation, only as an absence.
	src := mustReadSource(t, "../datahelpers/active_page_file_paths.go")
	for _, banned := range []string{"PageHasShippedPredicateFor", "deployed_at", "build_status"} {
		if regexp.MustCompile(`(?m)^[^/]*\b` + regexp.QuoteMeta(banned) + `\b`).MatchString(src) {
			t.Fatalf("ActivePageFilePaths must NOT carry a build-axis arm (%s): an active page's file "+
				"path is protected BEFORE it deploys, and narrowing this set lets a retraction delete "+
				"the path a live page is about to publish to", banned)
		}
	}
}

// mustReadSource reads a source file relative to this package, for the assertions
// that can only be expressed as the ABSENCE of a string.
func mustReadSource(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Clean(rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}
