package actions

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tests for the completeness floor on save_page_sections' reconciliation delete
// (bugs_open/165 site A). Two of them exist because a measurement caught me
// getting the design wrong, and they are the ones to keep if any are ever cut:
//
//   - TestPageSectionFloorPassesAHealthyRebuildOfALockedPage pins the locked-row
//     correction. My first denominator counted planned slots a lock makes
//     unwritable, so a PERFECT rebuild of idea.uk/index.html (6 planned, 4 locked)
//     scored 2/6 and would have been refused. Locked pages are precisely the ones a
//     human curated; refusing every rebuild of them is how a guard gets deleted.
//   - TestPageSectionFloorCatchesTheRatchetTheRowCohortMisses is the whole reason
//     there are two cohorts. Once a truncating writer has cut a page to two rows,
//     the row cohort reads 2/2 = 100% for ever and the damage is the new baseline.
//
// Every test that asserts a refusal was checked by breaking the guard and watching
// it fail; the package landmine (a test asserting a query is NOT issued passes
// vacuously against insertWorkItem, which swallows the mock's error) is why the
// refusal tests pin the INSERT's own arguments rather than its existence.

func floorMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// floorParams builds the ActionParams the guard reads: step config (for the
// floor) and the DB. Nothing else in ActionParams is touched by this path.
func floorParams(db *sql.DB, config map[string]interface{}) ActionParams {
	p := ActionParams{DB: db, Logger: zap.NewNop()}
	p.StepConfig.Config = config
	return p
}

// expectMeasurement registers the one measurement query with the four counts the
// guard divides by. Registered so it SUCCEEDS: a measurement that errors sends
// the guard down its fail-closed path, which would make every test below pass for
// the wrong reason.
func expectMeasurement(mock sqlmock.Sqlmock, writable, locked, planned, suppressed int) {
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages p WHERE p.id = $1")).
		WillReturnRows(sqlmock.NewRows([]string{"writable", "locked", "planned", "suppressed"}).
			AddRow(writable, locked, planned, suppressed))
}

// expectRefusalItem requires the refusal to reach site_work_items with the
// routing that makes it actionable — status 'needs_human_review' (no handler
// agent can decide whether a page genuinely halved). Pinning the argument is the
// assertion: a mismatch fails the Exec, and this test then fails on the missing
// commit rather than passing silently.
func expectRefusalItem(mock sqlmock.Sqlmock) {
	const cols = 16
	const statusIdx = 11 // 0-based; $12

	args := make([]driver.Value, cols)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[statusIdx] = "needs_human_review"

	// No two-strike COUNT is registered: the emitter sets recurrenceExpected, so
	// insertWorkItem's `if item.itemKey != "" && !item.recurrenceExpected` guard
	// means that query is never issued on the correct path. The case where it IS
	// issued — the flag dropped — is pinned separately and deliberately, by
	// TestPageSectionRefusalSurvivesATwoStrikeHistory.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func plainSections(n int) []SectionData {
	out := make([]SectionData, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, SectionData{
			ComponentName: fmt.Sprintf("slot-%d", i),
			ComponentID:   uuid.New().String(),
			HTML:          fmt.Sprintf("<section class=\"section\"><h2>Section %d</h2><p>Real copy.</p></section>", i),
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// projectedSectionInserts — the numerator both cohorts divide by
// ---------------------------------------------------------------------------

func TestProjectedSectionInsertsCountsPlainSections(t *testing.T) {
	if got := projectedSectionInserts(plainSections(5), nil); got != 5 {
		t.Fatalf("projected = %d, want 5", got)
	}
}

// A locked slot swallows the incoming section that matches it: the locked copy
// stands and the fresh copy is discarded, so it is not content this save writes.
func TestProjectedSectionInsertsExcludesLockedSlots(t *testing.T) {
	sections := plainSections(4)
	locked := []*lockedPageRow{
		{id: uuid.New(), slot: "slot-0"},
		{id: uuid.New(), slot: "slot-2"},
	}
	if got := projectedSectionInserts(sections, locked); got != 2 {
		t.Fatalf("projected = %d, want 2 (4 sections, 2 swallowed by locks)", got)
	}
}

// The simulation must not consume the REAL locked rows — the insert loop runs
// afterwards and relies on that bookkeeping. If the guard mutated them, every
// locked slot would look already-consumed and the locked copy would be
// overwritten: the guard would cause the 058 defect it sits next to.
func TestProjectedSectionInsertsDoesNotConsumeTheRealLockedRows(t *testing.T) {
	locked := []*lockedPageRow{{id: uuid.New(), slot: "slot-0"}}
	_ = projectedSectionInserts(plainSections(3), locked)
	if locked[0].consumed {
		t.Fatal("the real locked row was marked consumed by the floor's simulation")
	}
}

// One locked row may swallow only ONE incoming section, matching the insert
// loop's own rule — otherwise a page with a duplicate slot name would have its
// numerator silently under-counted and a healthy save refused.
func TestProjectedSectionInsertsOneLockSwallowsOneSection(t *testing.T) {
	sections := []SectionData{
		{ComponentName: "hero", ComponentID: uuid.New().String(), HTML: "<section>a</section>"},
		{ComponentName: "hero", ComponentID: uuid.New().String(), HTML: "<section>b</section>"},
	}
	locked := []*lockedPageRow{{id: uuid.New(), slot: "hero"}}
	if got := projectedSectionInserts(sections, locked); got != 1 {
		t.Fatalf("projected = %d, want 1 (one lock consumes one of two 'hero' sections)", got)
	}
}

// An unresolvable empty stub is refused by the insert loop (bugs_open/039), so
// counting it as confirmed content would inflate the numerator — permissive in
// the one direction this guard must not be wrong in.
func TestProjectedSectionInsertsExcludesUnresolvableStubs(t *testing.T) {
	sections := append(plainSections(2), SectionData{
		ComponentName: "orphan",
		ComponentID:   "", // resolves to no component
		HTML:          `<section class="section section--generic"><h2></h2><div></div></section>`,
	})
	if got := projectedSectionInserts(sections, nil); got != 2 {
		t.Fatalf("projected = %d, want 2 (the empty generic stub is not written)", got)
	}
}

// sectionIsUnresolvableStub is shared with the insert loop; a generic block that
// DID receive content must never match, or a real section stops counting.
func TestSectionIsUnresolvableStubIgnoresAGenericBlockWithContent(t *testing.T) {
	withCopy := SectionData{HTML: `<section class="section section--generic"><h2>Our approach</h2><p>Words.</p></section>`}
	if sectionIsUnresolvableStub(withCopy) {
		t.Fatal("a generic block carrying visible copy was treated as an empty stub")
	}
	resolved := SectionData{ComponentID: uuid.New().String(), HTML: `<section class="section section--generic"></section>`}
	if sectionIsUnresolvableStub(resolved) {
		t.Fatal("a section with a resolvable component_id was treated as an unresolvable stub")
	}
}

// ---------------------------------------------------------------------------
// the denominators
// ---------------------------------------------------------------------------

// The plan denominator is planned − suppressed − locked. The locked term is the
// one that was wrong first: see the file header.
func TestMeasurePageSectionCompletenessExcludesLockedAndSuppressedFromThePlan(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 2 /*writable*/, 4 /*locked*/, 7 /*planned*/, 1 /*suppressed*/)

	m, err := measurePageSectionCompleteness(context.Background(), db, uuid.New(), 2)
	if err != nil {
		t.Fatalf("measure: %v", err)
	}
	if m.Planned != 2 {
		t.Fatalf("planned denominator = %d, want 2 (7 planned − 1 suppressed − 4 locked)", m.Planned)
	}
	if len(m.Cohorts) != 2 {
		t.Fatalf("cohorts = %d, want 2", len(m.Cohorts))
	}
	if m.Cohorts[0].Stored != 2 || m.Cohorts[1].Stored != 2 {
		t.Fatalf("cohort denominators = %d/%d, want 2/2", m.Cohorts[0].Stored, m.Cohorts[1].Stored)
	}
}

// More locks than the plan mentions is representable (a lock outlives a replan),
// and a negative denominator would make the cohort's ratio nonsense.
func TestMeasurePageSectionCompletenessClampsThePlanAtZero(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 1, 9, 2, 0)

	m, err := measurePageSectionCompleteness(context.Background(), db, uuid.New(), 1)
	if err != nil {
		t.Fatalf("measure: %v", err)
	}
	if m.Planned != 0 {
		t.Fatalf("planned = %d, want 0 (clamped, not negative)", m.Planned)
	}
	// An empty cohort has nothing to lose and must never refuse a save.
	if got := m.Cohorts[1].ratio(); got != 1 {
		t.Fatalf("empty plan cohort ratio = %v, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// the guard
// ---------------------------------------------------------------------------

func TestPageSectionFloorAllowsAHealthyRebuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 6, 0, 6, 0)

	detail, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "index", plainSections(6), nil)
	if err != nil {
		t.Fatalf("healthy rebuild refused: %v", err)
	}
	if detail["completeness_status"] != "passed" {
		t.Fatalf("status = %v, want passed", detail["completeness_status"])
	}
	// The numbers are reported on a PASS too: "sections_saved: 6" without its
	// denominator is the alarm presented as output.
	if detail["writable_rows"] != 6 || detail["sections_projected"] != 6 {
		t.Fatalf("pass detail lost its numbers: %+v", detail)
	}
}

// THE CORRECTION, PINNED. 6 planned, 4 locked, 2 writable: a perfect rebuild
// hands over all 6 sections, 4 are swallowed by locks, 2 are written. Both
// cohorts must read 2/2. With the locked rows left in the plan denominator this
// scores 2/6 = 33% and refuses — which is what the first draft did, and what the
// live measurement of idea.uk/index.html caught.
func TestPageSectionFloorPassesAHealthyRebuildOfALockedPage(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 2 /*writable*/, 4 /*locked*/, 6 /*planned*/, 0)

	locked := []*lockedPageRow{
		{id: uuid.New(), slot: "slot-0"},
		{id: uuid.New(), slot: "slot-1"},
		{id: uuid.New(), slot: "slot-2"},
		{id: uuid.New(), slot: "slot-3"},
	}
	if _, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "index", plainSections(6), locked); err != nil {
		t.Fatalf("a healthy rebuild of a page with 4 locked slots was refused: %v", err)
	}
}

// The row cohort's own case: twelve stored, two written.
func TestPageSectionFloorRefusesATruncatedRebuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 12, 0, 12, 0)
	expectRefusalItem(mock)

	_, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", plainSections(2), nil)
	if err == nil {
		t.Fatal("a 2-of-12 rebuild was allowed to replace the page")
	}
	if !strings.Contains(err.Error(), "REFUSED") {
		t.Fatalf("refusal does not say so: %v", err)
	}
	// The refusal must name its own remedy, or it gets overridden by someone
	// deleting the guard.
	if !strings.Contains(err.Error(), savePageSectionsFloorKey) {
		t.Fatalf("refusal does not name the knob that resolves it: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the refusal did not reach site_work_items: %v", err)
	}
}

// THE RATCHET, and the reason for a second cohort. The page has already been cut
// to 2 rows by an earlier truncation, so the row cohort reads 2/2 = 100% and sees
// nothing wrong. The plan still says 12. Without the plan cohort this save is
// waved through and the damage becomes permanent.
func TestPageSectionFloorCatchesTheRatchetTheRowCohortMisses(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 2 /*writable — already damaged*/, 0, 12 /*planned*/, 0)
	expectRefusalItem(mock)

	_, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "services", plainSections(2), nil)
	if err == nil {
		t.Fatal("the row cohort's 100% hid a 2-of-12 page: the plan cohort did not fire")
	}
	if !strings.Contains(err.Error(), "planned sections") {
		t.Fatalf("refusal blames the wrong cohort: %v", err)
	}
}

// FAILS CLOSED. The measurement is the only thing standing between a half-blind
// writer and a stripped page; when it cannot be taken the save must not proceed.
func TestPageSectionFloorFailsClosedWhenItCannotMeasure(t *testing.T) {
	db, mock := floorMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages p WHERE p.id = $1")).
		WillReturnError(errors.New("connection reset"))
	expectRefusalItem(mock)

	_, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "index", plainSections(6), nil)
	if err == nil {
		t.Fatal("an unmeasurable floor allowed the save")
	}
	if !strings.Contains(err.Error(), "could not be measured") {
		t.Fatalf("refusal does not say the measurement failed: %v", err)
	}
}

// A guard with no exit is a defect in a guard's costume: 0 disables it, and the
// result says so rather than reporting a pass the floor never granted.
func TestPageSectionFloorZeroDisablesAndSaysSo(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 12, 0, 12, 0)

	detail, err := enforcePageSectionFloor(context.Background(),
		floorParams(db, map[string]interface{}{savePageSectionsFloorKey: 0}),
		uuid.New(), uuid.New(), "services", plainSections(1), nil)
	if err != nil {
		t.Fatalf("floor=0 refused the save: %v", err)
	}
	if detail["completeness_status"] != "floor_disabled" {
		t.Fatalf("status = %v, want floor_disabled — a disabled guard must not read as a pass",
			detail["completeness_status"])
	}
}

// A page with nothing stored and nothing planned has nothing to lose. A first
// build must never be refused by a guard about deletion.
func TestPageSectionFloorAllowsAFirstBuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 0, 0, 0, 0)

	if _, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "new-page", plainSections(3), nil); err != nil {
		t.Fatalf("a first build was refused: %v", err)
	}
}

// A truncated FIRST build is still caught, by the plan cohort alone — the case
// the row cohort structurally cannot see, and 63 pages currently sit in exactly
// this state (a plan, no rows).
func TestPageSectionFloorCatchesATruncatedFirstBuild(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 0 /*no rows yet*/, 0, 8 /*planned*/, 0)
	expectRefusalItem(mock)

	if _, err := enforcePageSectionFloor(context.Background(), floorParams(db, nil),
		uuid.New(), uuid.New(), "new-page", plainSections(1), nil); err == nil {
		t.Fatal("a 1-of-8 first build was allowed")
	}
}

// THE FOUR-SEAT OBJECTION, PINNED. Council round a54172b6 raised this as medium
// from editquality, tooling_provenance, debug_historian and guardian: a new
// emitter reusing a per-page item_key through insertWorkItem inherits the
// two-strike rule, and the third refusal on a chronically-thin page is born
// `unresolved` — terminal, and off the human-review queue this row exists to
// reach. Worse, a refusal arriving within 3h of a terminal predecessor is dropped
// outright. Both are silent, and both hit exactly when the page needs looking at.
//
// So the emitter sets recurrenceExpected, and this test is the proof. It supplies
// a history that WOULD brand the item — 2 prior terminal items, newest 100h old,
// so the <3h within-cycle suppression does not fire instead and mask the result —
// and requires the INSERT to still carry `needs_human_review`.
//
// Registered so the COUNT SUCCEEDS, which is the whole trick: insertWorkItem does
// `if err == nil && terminalCount > 0`, so an UNregistered query errors, the error
// is swallowed, the branding never happens, and the test passes green whichever
// way the flag is set. That is the vacuous-mock landmine on this exact helper.
// Verified by clearing the flag: the INSERT then carries `unresolved`, the
// WithArgs mismatch fails the Exec, and this test fails.
//
// IT NOW GUARDS ALL THREE CALL SITES. Site A shipped a private emitter hours
// before sites B and C landed the same routing as the shared
// emitPruneRefusalWorkItem; the private copy is retired, and this test points at
// the shared function, so clearing the flag there fails this test on behalf of
// page_components, site_nav_items/groups AND link_registry. The sibling test
// TestPruneRefusalWorkItemRoutesToHumanReview pins the ROUTING but cannot pin the
// flag: with recurrenceExpected set the COUNT is never issued, so a test that
// does not supply a branding history passes either way.
func TestPageSectionRefusalSurvivesATwoStrikeHistory(t *testing.T) {
	db, mock := floorMockDB(t)
	mock.MatchExpectationsInOrder(false)

	const cols = 16
	const statusIdx = 11 // 0-based; $12
	args := make([]driver.Value, cols)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[statusIdx] = "needs_human_review"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age_hours"}).AddRow(2, 100.0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := emitPruneRefusalWorkItem(context.Background(), db,
		savePageSectionsRefusal(uuid.New(), uuid.New(), "services", "refused: too few sections"),
		zap.NewNop()); err != nil {
		t.Fatalf("the third refusal on a page did not reach site_work_items as needs_human_review: %v", err)
	}
}

// The floor honours step config, so a workflow that knows its pages shrink can
// resolve it without anyone editing Go.
func TestPageSectionFloorHonoursConfiguredRatio(t *testing.T) {
	db, mock := floorMockDB(t)
	expectMeasurement(mock, 10, 0, 10, 0)

	detail, err := enforcePageSectionFloor(context.Background(),
		floorParams(db, map[string]interface{}{savePageSectionsFloorKey: 0.2}),
		uuid.New(), uuid.New(), "services", plainSections(3), nil)
	if err != nil {
		t.Fatalf("3-of-10 refused at a configured floor of 0.2: %v", err)
	}
	if detail["completeness_from_config"] != true {
		t.Fatalf("the result does not record that the floor came from config: %+v", detail)
	}
}
