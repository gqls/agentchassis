package actions

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tests for the completeness floors on the two remaining reconciliation deletes
// (bugs_closed/165 sites B and C) and for the shared reporting/refusal surface they
// added to prune_floor.go.
//
// The load-bearing one, if any are ever cut, is
// TestNavFloorAllowsAPageReHomedBetweenGroups. It pins the measured correction to
// what the bug file proposed: bugs_closed/165 asked for "per nav group" cohorts, and
// classifyPagesForNav RE-HOMES pages between groups as a matter of course
// (robot-hands.com's `tools` group holds a page the current classifier places in
// `utility`). A per-group cohort scores that legitimate re-homing as a 100% loss
// of one class and refuses the rebuild for ever. There is deliberately no
// per-group cohort, and this test is what stops one being added back.
//
// Every test that asserts a refusal was checked by breaking the guard and watching
// it fail. The package landmine — a test asserting a query is NOT issued passes
// vacuously against insertWorkItem, which swallows the mock's error — is why the
// refusal tests pin the INSERT's own arguments via expectRefusalItem rather than
// its existence. floorMockDB, floorParams and expectRefusalItem are shared with
// save_sections_prune_floor_test.go on purpose: three spellings of one mock is the
// same drift this file's subject exists to prevent.

// expectNavMeasurement registers the one nav measurement query with the counts the
// guard divides by. Registered so it SUCCEEDS: a measurement that errors sends the
// guard down its fail-closed path, which would make every test below pass for the
// wrong reason.
func expectNavMeasurement(mock sqlmock.Sqlmock, pagesStored, itemsStored int, domain string) {
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_nav_items WHERE site_id = $1")).
		WillReturnRows(sqlmock.NewRows([]string{"pages_stored", "items_stored", "domain"}).
			AddRow(pagesStored, itemsStored, domain))
}

func expectLinkMeasurement(mock sqlmock.Sqlmock, linksStored int) {
	mock.ExpectQuery(regexp.QuoteMeta("FROM link_registry WHERE source_page_id = $1")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(linksStored))
}

// ---------------------------------------------------------------------------
// The shared predicate — the guard must measure the population the loader reads
// ---------------------------------------------------------------------------

// navPageScopeSQL exists so the floor's denominator and loadPagesForNav's SELECT
// cannot drift apart. A guard measuring a different population from the one at
// risk reports a ratio for rows nobody was going to touch, which is worse than no
// guard because it looks like one. This pins that the loader actually uses it.
func TestLoadPagesForNavUsesTheSharedScopePredicate(t *testing.T) {
	db, mock := floorMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(navPageScopeSQL)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "title", "url", "nav_label", "nav_order", "page_type", "in_header", "in_footer",
		}))

	if _, err := loadPagesForNav(context.Background(), db, uuid.New(), zap.NewNop()); err != nil {
		t.Fatalf("loadPagesForNav: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("loadPagesForNav no longer uses navPageScopeSQL — the floor now measures a different population: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Site B — populate_nav_tables
// ---------------------------------------------------------------------------

func TestNavFloorAllowsAHealthyRebuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 31 /*pages stored*/, 17 /*items stored*/, "robot-hands.com")

	detail, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 31 /*pages loaded*/, 17 /*items to write*/)
	if err != nil {
		t.Fatalf("a healthy rebuild was refused: %v", err)
	}
	if detail["completeness_status"] != pruneStatusPassed {
		t.Fatalf("status = %v, want %q", detail["completeness_status"], pruneStatusPassed)
	}
	// The denominator must be published beside the count, not left to be inferred.
	if detail["pages_stored"] != 31 || detail["items_stored"] != 17 {
		t.Fatalf("the floor did not publish its denominators: %+v", detail)
	}
}

// THE MEASURED CORRECTION TO THE BUG FILE. robot-hands.com stores 17 nav items,
// one of which sits in a hand-created `tools` group; the current classifier places
// that same page in `utility`. Totals are unchanged — 17 in, 17 out — and the
// rebuild must pass. A per-group cohort (which bugs_closed/165 asked for) would read
// tools 0/1 = 0% and refuse this for ever.
func TestNavFloorAllowsAPageReHomedBetweenGroups(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 31, 17, "robot-hands.com")

	if _, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 31, 17); err != nil {
		t.Fatalf("a re-homing between nav groups was refused — a per-group cohort has been added back: %v", err)
	}
}

// THE DEFECT THIS FILE EXISTS FOR. loadPagesForNav logs a warning and `continue`s
// past a row it cannot scan, so a partial read is silent and success-shaped. Here
// 10 of 31 pages came back; the nav rebuild must not be allowed to replace the
// stored nav on that basis.
func TestNavFloorRefusesAPartialPageRead(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 31, 17, "robot-hands.com")
	expectRefusalItem(mock)

	_, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 10 /*only 10 of 31 pages scanned*/, 6)
	if err == nil {
		t.Fatal("a 10-of-31 page read was allowed to rebuild the whole nav")
	}
	if !strings.Contains(err.Error(), "REFUSED") {
		t.Fatalf("refusal does not say so: %v", err)
	}
	if !strings.Contains(err.Error(), "pages seen") {
		t.Fatalf("refusal blames the wrong cohort: %v", err)
	}
	// The refusal must name its own remedy, or it gets overridden by someone
	// deleting the guard.
	if !strings.Contains(err.Error(), populateNavFloorKey) {
		t.Fatalf("refusal does not name the knob that resolves it: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the refusal did not reach site_work_items: %v", err)
	}
}

// The case the page-load cohort cannot see: every page loaded, and the classifier
// then placed almost none of them. bugs_open/149 A2 was exactly this shape — a
// URL-prefix rule that dropped every /tools/ page out of nav while the pages
// themselves loaded fine.
func TestNavFloorRefusesAClassifierCollapse(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 31, 17, "robot-hands.com")
	expectRefusalItem(mock)

	_, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 31 /*all pages loaded*/, 2 /*but almost nothing classified*/)
	if err == nil {
		t.Fatal("a rebuild that placed 2 of 17 items was allowed")
	}
	if !strings.Contains(err.Error(), "nav items") {
		t.Fatalf("refusal blames the wrong cohort: %v", err)
	}
}

// A NEW SITE must never be refused. Nothing is stored, so there is nothing to
// lose: prune_floor reads an empty cohort as fully confirmed rather than as 0/0.
func TestNavFloorAllowsAFirstBuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 12, 0 /*no nav yet*/, "newsite.com")

	if _, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 12, 9); err != nil {
		t.Fatalf("a site's first nav build was refused: %v", err)
	}
}

// ...but a first build off a HALF-READ page corpus is still caught, because the
// pages cohort is measured against pages, not against the empty nav.
func TestNavFloorCatchesATruncatedFirstBuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 40, 0, "newsite.com")
	expectRefusalItem(mock)

	_, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 5 /*5 of 40 pages*/, 3)
	if err == nil {
		t.Fatal("a first build that saw 5 of 40 pages was allowed")
	}
	if !strings.Contains(err.Error(), "pages seen") {
		t.Fatalf("refusal blames the wrong cohort: %v", err)
	}
}

// FAILS CLOSED. The measurement is the only thing standing between a half-blind
// page read and a stripped nav bar; when it cannot be taken, nothing is deleted.
func TestNavFloorFailsClosedWhenItCannotMeasure(t *testing.T) {
	db, mock := floorMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_nav_items WHERE site_id = $1")).
		WillReturnError(errors.New("connection reset"))
	expectRefusalItem(mock)

	_, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil), uuid.New(), 31, 17)
	if err == nil {
		t.Fatal("an unmeasurable floor allowed the rebuild")
	}
	if !strings.Contains(err.Error(), "could not be measured") {
		t.Fatalf("the failure does not say why: %v", err)
	}
}

// A guard with no exit is a defect in a guard's costume. 0 disables it, and the
// result says so rather than reading like a floor that looked and was content.
func TestNavFloorZeroDisablesAndSaysSo(t *testing.T) {
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 31, 17, "robot-hands.com")

	detail, err := enforceNavPruneFloor(context.Background(),
		floorParams(db, map[string]interface{}{populateNavFloorKey: 0}),
		uuid.New(), 2 /*would otherwise refuse*/, 1)
	if err != nil {
		t.Fatalf("a disabled floor still refused: %v", err)
	}
	if detail["completeness_status"] != pruneStatusDisabled {
		t.Fatalf("status = %v, want %q", detail["completeness_status"], pruneStatusDisabled)
	}
}

func TestNavFloorHonoursConfiguredRatio(t *testing.T) {
	// 20 of 31 pages = 65%. Passes the default 0.5, refused at 0.9.
	db, mock := floorMockDB(t)
	expectNavMeasurement(mock, 31, 17, "robot-hands.com")
	if _, err := enforceNavPruneFloor(context.Background(), floorParams(db, nil),
		uuid.New(), 20, 17); err != nil {
		t.Fatalf("65%% was refused at the default floor: %v", err)
	}

	db2, mock2 := floorMockDB(t)
	expectNavMeasurement(mock2, 31, 17, "robot-hands.com")
	expectRefusalItem(mock2)
	if _, err := enforceNavPruneFloor(context.Background(),
		floorParams(db2, map[string]interface{}{populateNavFloorKey: 0.9}),
		uuid.New(), 20, 17); err == nil {
		t.Fatal("65% was allowed at a configured floor of 0.9")
	}
}

// ---------------------------------------------------------------------------
// Site C — extract_and_sync_links
// ---------------------------------------------------------------------------

func TestLinkRegistryFloorAllowsAHealthyResync(t *testing.T) {
	db, mock := floorMockDB(t)
	expectLinkMeasurement(mock, 24)

	detail, err := enforceLinkRegistryFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", 23)
	if err != nil {
		t.Fatalf("a healthy resync was refused: %v", err)
	}
	if detail["links_stored"] != 24 || detail["links_to_write"] != 23 {
		t.Fatalf("the floor did not publish its denominator: %+v", detail)
	}
}

func TestLinkRegistryFloorRefusesATruncatedExtraction(t *testing.T) {
	db, mock := floorMockDB(t)
	expectLinkMeasurement(mock, 40)
	expectRefusalItem(mock)

	// The contract is an ERROR, uniform with sites A and B. Round 2 made this one
	// return a bool and skip instead; four council seats ruled that a workaround for
	// the loop's missing per-substep error routing (now bugs_open/173) rather than a
	// fix, and a bool is ignorable where an error is not.
	_, err := enforceLinkRegistryFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", 3)
	if err == nil {
		t.Fatal("a 3-of-40 extraction was allowed to replace the page's links")
	}
	if !strings.Contains(err.Error(), "links") || !strings.Contains(err.Error(), linkRegistryFloorKey) {
		t.Fatalf("refusal is not actionable: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the refusal did not reach site_work_items: %v", err)
	}
}

// INERT TODAY, BY CONSTRUCTION — link_registry holds zero rows fleet-wide, so
// every cohort reads Stored=0 and the floor allows every prune until the insert
// half starts working. That is the correct behaviour (a class appearing for the
// first time must never refuse a prune) and it is pinned so nobody "fixes" it
// into a guard that blocks the table's first ever write.
func TestLinkRegistryFloorAllowsTheFirstSyncIntoAnEmptyTable(t *testing.T) {
	db, mock := floorMockDB(t)
	expectLinkMeasurement(mock, 0)

	if _, err := enforceLinkRegistryFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", 12); err != nil {
		t.Fatalf("the first ever link sync was refused: %v", err)
	}
}

func TestLinkRegistryFloorFailsClosedWhenItCannotMeasure(t *testing.T) {
	db, mock := floorMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM link_registry WHERE source_page_id = $1")).
		WillReturnError(errors.New("connection reset"))
	expectRefusalItem(mock)

	if _, err := enforceLinkRegistryFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", 12); err == nil {
		t.Fatal("an unmeasurable floor allowed the resync")
	}
}

// ---------------------------------------------------------------------------
// The shared surface added to prune_floor.go
// ---------------------------------------------------------------------------

// A `deleted: 0` is ambiguous between "nothing to delete" and "we refused", which
// is why the status is emitted rather than inferred. All three values are pinned
// because the refused one is the case a reader acts on.
func TestPruneFloorDetailReportsEachStatusDistinctly(t *testing.T) {
	cases := []struct {
		name    string
		floor   float64
		cohorts []pruneCohort
		want    string
	}{
		{"passed", 0.5, []pruneCohort{{Label: "x", Confirmed: 9, Stored: 10}}, pruneStatusPassed},
		{"refused", 0.5, []pruneCohort{{Label: "x", Confirmed: 1, Stored: 10}}, pruneStatusRefused},
		{"disabled", 0, []pruneCohort{{Label: "x", Confirmed: 1, Stored: 10}}, pruneStatusDisabled},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := evaluatePruneFloor(c.floor, c.cohorts)
			d := pruneFloorDetail(v, true, "reason", map[string]interface{}{"extra": 7})
			if d["completeness_status"] != c.want {
				t.Fatalf("status = %v, want %q", d["completeness_status"], c.want)
			}
			if d["extra"] != 7 {
				t.Fatalf("caller's own numbers were dropped from the detail block: %+v", d)
			}
			if d["completeness_cohorts"] == nil {
				t.Fatalf("the cohorts behind the decision were not reported: %+v", d)
			}
		})
	}
}

// A refusal nobody can see is the 034/076 shape. This pins the routing that makes
// it actionable: status 'needs_human_review' with no handler agent, because
// whether a corpus genuinely halved is a human judgement.
func TestPruneRefusalWorkItemRoutesToHumanReview(t *testing.T) {
	db, mock := floorMockDB(t)
	expectRefusalItem(mock)

	err := emitPruneRefusalWorkItem(context.Background(), db, pruneRefusal{
		SiteID:   uuid.New(),
		Source:   "populate_nav_tables",
		Pipeline: "build",
		ItemType: "nav_rebuild_refused_incomplete",
		ItemKey:  "nav_rebuild_refused_incomplete:x",
		Subject:  "site x",
		Summary:  "refused",
		Reason:   "reason",
		Fix:      "fix",
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("emit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the refusal did not reach site_work_items with the expected routing: %v", err)
	}
}

// A nil DB must not panic and must not pretend to have recorded anything.
func TestPruneRefusalWorkItemIsANoOpWithoutADB(t *testing.T) {
	if err := emitPruneRefusalWorkItem(context.Background(), nil, pruneRefusal{}, zap.NewNop()); err != nil {
		t.Fatalf("nil DB should be a silent no-op, got %v", err)
	}
}

// THE REVERTED CONTRACT, pinned. Round 2 of council c69e935a made this guard
// return (detail, bool) and skip the sync; four seats ruled that a workaround for
// the loop's missing per-substep error routing (now bugs_open/173) rather than a
// fix, and round 3 reverted it. The `editquality` seat then noted, fairly, that
// nothing pinned the revert itself — so this does.
//
// Two assertions, and the second is the one that matters. A refusal must return a
// non-nil ERROR (a bool is ignorable where an error is not), and it must return a
// NIL detail map, so a caller that ignores the error cannot find usable-looking
// numbers and carry on. The compile-time signature check is what actually fails if
// anyone re-introduces the bool.
func TestLinkRegistryFloorRefusalIsAnErrorNotASkippableFlag(t *testing.T) {
	// Compile-time: the contract is (map, error), identical to sites A and B.
	var _ func(context.Context, ActionParams, uuid.UUID, uuid.UUID, string, int) (map[string]interface{}, error) = enforceLinkRegistryFloor

	db, mock := floorMockDB(t)
	expectLinkMeasurement(mock, 40)
	expectRefusalItem(mock)

	detail, err := enforceLinkRegistryFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", 3)
	if err == nil {
		t.Fatal("a refusal returned no error — the round-2 skip contract has been reinstated")
	}
	if detail != nil {
		t.Fatalf("a refusal returned a non-nil detail map: a caller ignoring the error would find usable numbers and proceed: %+v", detail)
	}
}
