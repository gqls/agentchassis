package actions

// The plan guard, asserted at the action boundary (council trail da3f2d9b,
// bug_historian objection; owner decision 2026-07-31 to build rather than
// defer). page_components is downstream of the plan stores, so a slot whose
// repetition the effective plan source itself specifies must never be deleted
// below the specified count.
//
// Assertion style per the standing landmine on vacuous not-issued assertions:
// EFFECTS, never call-absence. The skip tests register only the SELECTs and the
// commit; if the action attempted the DELETE anyway, the unregistered Exec
// would fail the mock, which fails the call, which fails the test. The result
// map's plan_specified_repetition entry is the positive effect asserted.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// Long enough that NormaliseSectionText clears the 80-char floor — the guard
// tests are about the PLAN, so the content must be unambiguously in-remit.
const dupBlobA = `{"headline":"The rules are simple","body":"Holding your position is not. Every day one provocation drops into the arena and you defend it against an opponent on a clock, with a judge at the end."}`
const dupBlobB = `{"headline":"A different section entirely","body":"This one exists so the delete-everything refusal guard has no reason to fire while the plan guard is what is under test here."}`

func dedupeParams(db *sql.DB, pageID uuid.UUID) ActionParams {
	return ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{}, // the action dereferences it unconditionally
		StepConfig: models.Step{
			Config: map[string]interface{}{},
		},
		CollectedData: map[string]interface{}{
			"page_id": pageID.String(),
		},
	}
}

func sectionColumns() []string {
	return []string{"id", "position", "slot_name", "content_data"}
}

// expectLoadAndResolve registers the two queries every scenario starts with:
// the FOR UPDATE section load and the page -> (site_id, name) resolve.
func expectLoadAndResolve(mock sqlmock.Sqlmock, pageID, siteID uuid.UUID, rows *sqlmock.Rows) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, position, COALESCE`).
		WithArgs(pageID).
		WillReturnRows(rows)
	mock.ExpectQuery(`SELECT site_id, name FROM pages`).
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id", "name"}).
			AddRow(siteID.String(), "index"))
}

func planTableRows(components ...string) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"component_name"})
	for _, c := range components {
		r.AddRow(c)
	}
	return r
}

// Plan (table) specifies the component TWICE -> both rows are plan-specified,
// nothing is deleted, and the result says so. This is the webdesign.co.uk
// shape (info-card-grid x2 on index) with byte-identical content.
func TestRemoveDuplicates_PlanSpecifiedRepetitionIsSkipped(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	pageID, siteID := uuid.New(), uuid.New()

	expectLoadAndResolve(mock, pageID, siteID, sqlmock.NewRows(sectionColumns()).
		AddRow(uuid.New().String(), 1, "info-card-grid", dupBlobA).
		AddRow(uuid.New().String(), 2, "info-card-grid", dupBlobA))
	mock.ExpectQuery(`site_plan_sections`).
		WithArgs(siteID, "index").
		WillReturnRows(planTableRows("info-card-grid", "info-card-grid"))
	mock.ExpectCommit() // the no-op branch commits; no DELETE is registered

	got, err := RemoveDuplicatePageSectionsAction(context.Background(), dedupeParams(db, pageID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := got.(map[string]interface{})
	if m["changed"] != false || m["removed"] != 0 {
		t.Errorf("plan-specified repetition must not be deleted; got %v", m)
	}
	skipped, _ := m["plan_specified_repetition"].([]map[string]interface{})
	if len(skipped) != 1 || skipped[0]["slot_name"] != "info-card-grid" ||
		skipped[0]["planned_count"] != 2 || skipped[0]["plan_source"] != "site_plan_sections" {
		t.Errorf("skip must be REPORTED with slot, count and source; got %v", m["plan_specified_repetition"])
	}
	if !strings.Contains(m["detail"].(string), "plan-specified") {
		t.Errorf("detail must say why nothing was deleted; got %q", m["detail"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// Plan specifies the component ONCE -> the duplicate is not plan-specified and
// the deletion proceeds. The guard must not disable the repair.
func TestRemoveDuplicates_PlanCountOneStillDeletes(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	pageID, siteID := uuid.New(), uuid.New()
	keepID, dropID, otherID := uuid.New(), uuid.New(), uuid.New()

	expectLoadAndResolve(mock, pageID, siteID, sqlmock.NewRows(sectionColumns()).
		AddRow(keepID.String(), 1, "hero-about", dupBlobA).
		AddRow(dropID.String(), 2, "hero-about", dupBlobA).
		AddRow(otherID.String(), 3, "differentiators", dupBlobB))
	mock.ExpectQuery(`site_plan_sections`).
		WithArgs(siteID, "index").
		WillReturnRows(planTableRows("hero-about", "differentiators"))
	mock.ExpectExec(`DELETE FROM page_components`).
		WithArgs(pageID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE page_components`).
		WithArgs(pageID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery(`SELECT count`).
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectCommit()

	got, err := RemoveDuplicatePageSectionsAction(context.Background(), dedupeParams(db, pageID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := got.(map[string]interface{})
	if m["changed"] != true || m["removed"] != 1 {
		t.Errorf("plan-count 1 must not protect a duplicate; got %v", m)
	}
	if skipped, _ := m["plan_specified_repetition"].([]map[string]interface{}); len(skipped) != 0 {
		t.Errorf("nothing should be reported plan-specified here; got %v", skipped)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// A site on the ASPECT path (no current-plan table rows) must be guarded by
// site_specs.site_plan — a table-only guard is blind on most of the fleet
// (310 of 1,026 slot-named rows resolve to a table entry, measured 2026-07-31).
func TestRemoveDuplicates_AspectPathIsGuardedToo(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	pageID, siteID := uuid.New(), uuid.New()

	expectLoadAndResolve(mock, pageID, siteID, sqlmock.NewRows(sectionColumns()).
		AddRow(uuid.New().String(), 1, "generic-text-block", dupBlobA).
		AddRow(uuid.New().String(), 2, "generic-text-block", dupBlobA))
	mock.ExpectQuery(`site_plan_sections`).
		WithArgs(siteID, "index").
		WillReturnRows(planTableRows()) // table silent
	mock.ExpectQuery(`SELECT data FROM site_specs`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow(`{"pages":[{"name":"index","sections":["generic-text-block","generic-text-block"]}]}`))
	mock.ExpectCommit()

	got, err := RemoveDuplicatePageSectionsAction(context.Background(), dedupeParams(db, pageID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := got.(map[string]interface{})
	if m["changed"] != false {
		t.Errorf("aspect-specified repetition must not be deleted; got %v", m)
	}
	skipped, _ := m["plan_specified_repetition"].([]map[string]interface{})
	if len(skipped) != 1 || skipped[0]["plan_source"] != "site_specs.site_plan" {
		t.Errorf("skip must name the aspect as the source; got %v", m["plan_specified_repetition"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// A site on the CACHE path (neither table nor aspect) must be guarded by
// pages.sections — a full rebuild of such a page reads exactly that store.
func TestRemoveDuplicates_CachePathIsGuardedToo(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	pageID, siteID := uuid.New(), uuid.New()

	expectLoadAndResolve(mock, pageID, siteID, sqlmock.NewRows(sectionColumns()).
		AddRow(uuid.New().String(), 1, "info-card-grid", dupBlobA).
		AddRow(uuid.New().String(), 2, "info-card-grid", dupBlobA))
	mock.ExpectQuery(`site_plan_sections`).
		WithArgs(siteID, "index").
		WillReturnRows(planTableRows())
	mock.ExpectQuery(`SELECT data FROM site_specs`).
		WithArgs(siteID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT sections FROM pages`).
		WithArgs(siteID, "index").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).
			AddRow(`["info-card-grid","info-card-grid"]`))
	mock.ExpectCommit()

	got, err := RemoveDuplicatePageSectionsAction(context.Background(), dedupeParams(db, pageID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := got.(map[string]interface{})
	if m["changed"] != false {
		t.Errorf("cache-specified repetition must not be deleted; got %v", m)
	}
	skipped, _ := m["plan_specified_repetition"].([]map[string]interface{})
	if len(skipped) != 1 || skipped[0]["plan_source"] != "pages.sections" {
		t.Errorf("skip must name pages.sections as the source; got %v", m["plan_specified_repetition"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// An unreadable plan store FAILS CLOSED: the action errors instead of deleting
// as if the page were unplanned.
func TestRemoveDuplicates_UnreadablePlanStoreFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	pageID, siteID := uuid.New(), uuid.New()

	expectLoadAndResolve(mock, pageID, siteID, sqlmock.NewRows(sectionColumns()).
		AddRow(uuid.New().String(), 1, "hero-about", dupBlobA).
		AddRow(uuid.New().String(), 2, "hero-about", dupBlobA))
	mock.ExpectQuery(`site_plan_sections`).
		WithArgs(siteID, "index").
		WillReturnError(fmt.Errorf("connection reset"))

	_, err = RemoveDuplicatePageSectionsAction(context.Background(), dedupeParams(db, pageID))
	if err == nil {
		t.Fatal("an unreadable plan store must abort the delete, not permit it")
	}
	if !strings.Contains(err.Error(), "refusing to delete") {
		t.Errorf("error should state the refusal; got %v", err)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Error(merr)
	}
}
