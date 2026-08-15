package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// The shared-component fence in UpdateComponentHTMLAction (bugs_open/281).
//
// The write it guards fans out: step 4 flips every placement of the component
// to build_status='pending'. For a shared SECTION component that is a fleet
// incident (the ported-page wrapper: one row behind ~115 pages on two sites,
// rewritten by a tool fix on 2026-08-05 and again on 2026-08-14). The fence is
// proven by MUTATION here, not by bookkeeping: a refused write must never
// reach the UPDATE (sqlmock fails the test on any unexpected statement), a
// permitted write must.

const fenceComponentID = "a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef"

// newFenceMock wires the two reads every path performs: the component load
// (with component_level) and the placement census.
func newFenceMock(t *testing.T, level string, pages, sites int) (sqlmock.Sqlmock, ActionParams, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	mock.ExpectQuery(`SELECT html_template, function, name, COALESCE\(component_level, ''\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "function", "name", "component_level"}).
			AddRow(`<section class="ported-page" data-component="ported-page">{{.body}}</section>`, "ported-page", "Ported Page", level))
	mock.ExpectQuery(`SELECT count\(DISTINCT pc.page_id\), count\(DISTINCT p.site_id\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"pages", "sites"}).AddRow(pages, sites))
	params := ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"component_id": "input_data.component_id",
			"html_field":   "input_data.new_html",
		}},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"component_id": fenceComponentID,
				"new_html":     `<section class="ported-page" data-component="ported-page"><style>.x{}</style>{{.body}}</section>`,
			},
		},
	}
	return mock, params, func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("sqlmock expectations: %v", err)
		}
		db.Close()
	}
}

// A shared section-level component is refused BEFORE the UPDATE — the mock has
// no UPDATE expectation, so reaching one fails the test. The refusal is
// recorded (agent_error_log INSERT) and is a hard error, never a silent success.
func TestUpdateComponentHTML_RefusesSharedNonToolComponent(t *testing.T) {
	mock, params, done := newFenceMock(t, "section", 115, 2)
	defer done()
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := UpdateComponentHTMLAction(context.Background(), params)
	if err == nil {
		t.Fatal("shared section component write must be REFUSED; got nil error")
	}
	if !strings.Contains(err.Error(), "placed on 115 pages across 2 sites") {
		t.Errorf("refusal must name the blast radius; got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "allow_shared_component_write") {
		t.Errorf("refusal must name the opt-in; got %q", err.Error())
	}
}

// The opt-in in STEP CONFIG lets a deliberate fleet-wide template change
// through: the UPDATE and the placement flip both run.
func TestUpdateComponentHTML_OptInAllowsSharedWrite(t *testing.T) {
	mock, params, done := newFenceMock(t, "section", 115, 2)
	defer done()
	params.StepConfig.Config["allow_shared_component_write"] = true
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE page_components`).WillReturnResult(sqlmock.NewResult(0, 115))

	if _, err := UpdateComponentHTMLAction(context.Background(), params); err != nil {
		t.Fatalf("opted-in shared write must proceed; got %v", err)
	}
}

// A per-site tool fork on one page — the ordinary case — is untouched by the
// fence and reaches the UPDATE exactly as before.
func TestUpdateComponentHTML_SinglePlacementForkProceeds(t *testing.T) {
	mock, params, done := newFenceMock(t, "tool", 1, 1)
	defer done()
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE page_components`).WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := UpdateComponentHTMLAction(context.Background(), params); err != nil {
		t.Fatalf("single-placement fork write must proceed; got %v", err)
	}
}

// A tool fork placed on two sites is the established pattern
// (tool-llm-cost-calculator, five successful edits): WARN, do not refuse.
func TestUpdateComponentHTML_MultiSiteForkProceeds(t *testing.T) {
	mock, params, done := newFenceMock(t, "tool", 2, 2)
	defer done()
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE page_components`).WillReturnResult(sqlmock.NewResult(0, 2))

	if _, err := UpdateComponentHTMLAction(context.Background(), params); err != nil {
		t.Fatalf("multi-site fork write must proceed (warn only); got %v", err)
	}
}

// A fence that cannot look must not wave the write through.
func TestUpdateComponentHTML_CensusFailureFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT html_template, function, name, COALESCE\(component_level, ''\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "function", "name", "component_level"}).
			AddRow("<div>{{.body}}</div>", "ported-page", "Ported Page", "section"))
	mock.ExpectQuery(`SELECT count\(DISTINCT pc.page_id\), count\(DISTINCT p.site_id\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(context.DeadlineExceeded)

	params := ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"component_id": "input_data.component_id",
			"html_field":   "input_data.new_html",
		}},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"component_id": fenceComponentID,
				"new_html":     "<div><style>.x{}</style>{{.body}}</div>",
			},
		},
	}
	_, err = UpdateComponentHTMLAction(context.Background(), params)
	if err == nil || !strings.Contains(err.Error(), "placement census failed") {
		t.Fatalf("census failure must fail closed; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// The guardian seat's narrowing (council 360ae540): the census is fail-closed
// ONLY on the non-tool path. A tool fork whose census errors keeps the
// pre-fence behaviour — warn and write — so the common path gains no new
// failure mode.
func TestUpdateComponentHTML_CensusFailureOnToolForkProceeds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT html_template, function, name, COALESCE\(component_level, ''\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "function", "name", "component_level"}).
			AddRow("<div>{{.body}}</div>", "tool-x", "Tool X", "tool"))
	mock.ExpectQuery(`SELECT count\(DISTINCT pc.page_id\), count\(DISTINCT p.site_id\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(context.DeadlineExceeded)
	mock.ExpectExec(`UPDATE content_components`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE page_components`).WillReturnResult(sqlmock.NewResult(0, 1))

	params := ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"component_id": "input_data.component_id",
			"html_field":   "input_data.new_html",
		}},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"component_id": fenceComponentID,
				"new_html":     "<div><style>.x{}</style>{{.body}}</div>",
			},
		},
	}
	if _, err := UpdateComponentHTMLAction(context.Background(), params); err != nil {
		t.Fatalf("tool fork with a failed census must still write (advisory path); got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}
