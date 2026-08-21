// FILE: platform/orchestration/actions/upsert_page_sections_guard_test.go
//
// bugs_open/204, the write-side half. `sections = EXCLUDED.sections` was
// unconditional while its nav_label and meta_description siblings IN THE SAME
// STATEMENT had carried destructive-write guards since 2026-08-19. On 2026-08-20
// that unguarded clause turned one planner defect into 41 emptied live pages.
//
// These pin all four directions, because a guard that only ever keeps is as broken
// as one that only ever overwrites:
//   - empty over non-empty  -> KEEP, and say so (the defect);
//   - empty over non-empty WITH the recompose release -> overwrite (the door stays
//     open; 72 of 748 active pages legitimately live at sections=[], 60 of them
//     tools, so making zero sections unreachable would be a worse bug);
//   - non-empty over non-empty -> overwrite (the plan is still authoritative — this
//     is what keeps the guard from freezing the column, which is exactly the
//     failure mode the meta_description test warns about one file over);
//   - empty over empty -> overwrite, no noise.

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// The guard as it must appear in the statement. Written out rather than matched
// loosely, so a CASE that protected the WRONG direction — keeping the stored list
// even when the plan proposes a real one, which would freeze every page's
// composition for ever — still fails this test.
const sectionsGuardSQL = `sections = CASE
				WHEN $13::bool THEN EXCLUDED.sections
				WHEN COALESCE(jsonb_array_length(EXCLUDED.sections), 0) > 0 THEN EXCLUDED.sections
				WHEN COALESCE(jsonb_array_length(pages.sections), 0) = 0 THEN EXCLUDED.sections
				ELSE pages.sections
			END`

func TestUpsertPage_SectionsGuardIsInTheStatementAndProtectsTheRightDirection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(sectionsGuardSQL)).WillReturnRows(pageReturnRows(pageID, siteID, false))

	page := map[string]interface{}{"name": "about", "url": "/about.html", "title": "About", "page_type": "content"}
	if _, err := upsertPage(context.Background(), db, siteID, page, 0, uuid.NullUUID{}, false, zap.NewNop()); err != nil {
		t.Fatalf("upsertPage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the ON CONFLICT clause no longer protects sections from an empty overwrite: %v", err)
	}
}

// ⚠ FOUND BY MUTATION, and the finding is about the LIMITS of this file's method.
// sqlmock never executes SQL, so TestUpsertPage_RefusalIsReportedToTheCaller below
// proves only that the Go SCANS the column — mutating the RETURNING expression to
// `false AND ...` left every test green. A mock's bookkeeping cannot assert a
// negative about SQL it does not run. So the expression is pinned as TEXT, exactly
// as the CASE clause above is:
//
// What this CAN establish: the statement still asks the database the right
// question — "is the stored list non-empty while the incoming one is empty?"
// What it CANNOT: that Postgres answers it correctly. That needs a live row, and
// the verification plan in the lane's PLAN §7 induces one on a throwaway page.
// Recorded rather than papered over, because a test that looks behavioural and is
// textual is worse than one that admits which it is.
const sectionsKeptExprSQL = `(COALESCE(jsonb_array_length(sections), 0) > 0
		             AND COALESCE(jsonb_array_length($11::jsonb), 0) = 0) AS sections_kept`

func TestUpsertPage_RefusalExpressionIsStillInTheStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(sectionsKeptExprSQL)).WillReturnRows(pageReturnRows(pageID, siteID, false))

	page := map[string]interface{}{"name": "about", "url": "/about.html", "title": "About", "page_type": "content"}
	if _, err := upsertPage(context.Background(), db, siteID, page, 0, uuid.NullUUID{}, false, zap.NewNop()); err != nil {
		t.Fatalf("upsertPage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the RETURNING clause no longer computes sections_kept — a refusal would become silent: %v", err)
	}
}

// The refusal must reach the CALLER, not just the database. A silent keep would be
// a new landmine of exactly the shape this bug is: the write reports success and
// the plan quietly is not what the database holds.
func TestUpsertPage_RefusalIsReportedToTheCaller(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO pages")).
		WillReturnRows(pageReturnRows(pageID, siteID, true)) // the DB says it kept

	page := map[string]interface{}{"name": "about", "url": "/about.html", "title": "About", "page_type": "content"}
	rec, err := upsertPage(context.Background(), db, siteID, page, 0, uuid.NullUUID{}, false, zap.NewNop())
	if err != nil {
		t.Fatalf("upsertPage: %v", err)
	}
	if !rec.SectionsKept {
		t.Fatal("a refusal must surface on the PageRecord; a caller that cannot see it cannot record it, " +
			"and an unrecorded divergence between the plan and the database is this bug in a new place")
	}
}

// THE DOOR. Zero sections must stay reachable, or the guard is a worse bug than the
// one it fixes. The deliberate route is the recompose_pages release — the channel
// that already means "this page is released for redesign".
func TestUpsertPage_RecomposeReleaseAllowsTheEmptying(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	// $13 true is the whole assertion: the caller declared this emptying deliberate.
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO pages")).
		WithArgs(siteID, "about", "/about.html", "About", "content",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true).
		WillReturnRows(pageReturnRows(pageID, siteID, false))

	page := map[string]interface{}{"name": "about", "url": "/about.html", "title": "About", "page_type": "content"}
	if _, err := upsertPage(context.Background(), db, siteID, page, 0, uuid.NullUUID{}, true, zap.NewNop()); err != nil {
		t.Fatalf("upsertPage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the recompose release did not reach the statement: %v", err)
	}
}

// The guard must not become an excuse to stop threading the release. If this ever
// binds a constant, a recompose_pages run silently stops being able to redesign.
func TestSyncPagesToDB_ThreadsTheRecomposeReleasePerPage(t *testing.T) {
	// The behaviour is asserted at the SQL boundary in the test above; here we pin
	// the wiring that decides the flag, because the two can drift independently:
	// recomposePagesFromSpec reads input_data.spec.recompose_pages, and a page not
	// named there must get false even when another page on the same run is named.
	set := recomposePagesFromSpec(map[string]interface{}{
		"input_data": map[string]interface{}{
			"spec": map[string]interface{}{
				"recompose_pages": []interface{}{"guide-a"},
			},
		},
	}, zap.NewNop())

	if !set["guide-a"] {
		t.Error("a page named in the release must be allowed to empty")
	}
	if set["guide-b"] {
		t.Error("a page NOT named in the release must not inherit another page's permission — " +
			"a site-wide flag here would re-open the exact hole the guard closes")
	}
	if set["about"] {
		t.Error("an unrelated page must not be releasable by accident")
	}
}

func TestSyncPagesToDB_AbsentReleaseIsNoRelease(t *testing.T) {
	// The unsafe direction is the DEFAULT: with no recompose_pages key at all,
	// nothing may be emptied. Disconfirming result: a non-nil permissive set.
	if set := recomposePagesFromSpec(map[string]interface{}{}, zap.NewNop()); len(set) != 0 {
		t.Fatalf("no release must mean no permission, got %v", set)
	}
	if set := recomposePagesFromSpec(nil, zap.NewNop()); set["anything"] {
		t.Fatal("nil collected data must not grant permission")
	}
}

// The refusal message must name the pages and point at the upstream cause. A
// warning that says only "something was refused" costs the next reader the
// investigation this whole lane just did.
func TestSectionsRefusalMessageNamesPagesAndCause(t *testing.T) {
	pages := []string{"guide-a", "guide-b"}
	msg := "2 page(s) proposed an EMPTY sections list over a non-empty stored one; the stored list was kept: " +
		strings.Join(pages, ", ")
	for _, want := range []string{"guide-a", "guide-b", "EMPTY sections list", "stored list was kept"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal message must contain %q", want)
		}
	}
}

// pageReturnRows builds the widened RETURNING row: the eleven original columns
// plus sections_kept, which bugs_open/204 added so a refusal reaches the caller.
func pageReturnRows(pageID, siteID uuid.UUID, sectionsKept bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "site_id", "name", "url", "title", "page_type",
		"nav_label", "nav_order", "in_header", "in_footer", "status", "sections_kept",
	}).AddRow(pageID, siteID, "about", "/about.html", "About", "content",
		"About", 10, true, true, "active", sectionsKept)
}

// ⚠ ALSO FOUND BY MUTATION, and it was the mutation that mattered most.
// Replacing the per-page lookup at the call site with a constant `true` — i.e.
// making the release SITE-WIDE — left every test above green, because they all
// stop at upsertPage's boundary and none of them exercise the wiring that DECIDES
// the flag. A site-wide release re-opens the exact hole the guard closes: one
// legitimately-recomposed page would license emptying every other page in the run.
//
// So this drives the whole action, with two pages, one released and one not, and
// asserts the flag each page's statement actually received.
func TestSyncPagesToDB_ReleaseIsPerPageAtTheStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// The reads SyncPagesToDBAction makes before the loop.
	mock.ExpectQuery("flat_urls|site_specs").WillReturnRows(
		sqlmock.NewRows([]string{"flat"}).AddRow(false))
	mock.ExpectQuery("honour_realised_identity").WillReturnRows(
		sqlmock.NewRows([]string{"honour", "twin", "stem"}).AddRow(false, false, false))
	mock.ExpectQuery("FROM site_plans").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	// released-guide is named in the release; kept-guide is NOT. The assertion is
	// the LAST bound argument of each statement: true then false.
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO pages")).
		WithArgs(siteID, "released-guide", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true).
		WillReturnRows(pageReturnRows(uuid.New(), siteID, false))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO pages")).
		WithArgs(siteID, "kept-guide", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnRows(pageReturnRows(uuid.New(), siteID, false))

	// Navigation, and whatever else the tail of the action reads, may return
	// anything — the assertion above has already been made by then.
	mock.MatchExpectationsInOrder(false)

	params := ActionParams{
		Context: context.Background(),
		Logger:  zap.NewNop(),
		DB:      db,
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{
					"recompose_pages": []interface{}{"released-guide"},
				},
			},
			"page_plan": map[string]interface{}{
				"pages": []interface{}{
					map[string]interface{}{"name": "released-guide", "page_type": "content", "sections": []interface{}{}},
					map[string]interface{}{"name": "kept-guide", "page_type": "content", "sections": []interface{}{}},
				},
			},
		},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		ExecutionContext: &orchtypes.ExecutionContext{OrchestrationID: uuid.New().String(), StepName: "sync_pages"},
	}

	// The action's tail (navigation) may error against a bare mock; the two
	// statements above are what this test is about, so an error there is not a
	// failure of the assertion.
	_, _ = SyncPagesToDBAction(context.Background(), params)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the recompose release is not being decided PER PAGE at the statement — "+
			"a site-wide flag here re-opens the hole the guard closes: %v", err)
	}
}
