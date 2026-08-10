// FILE: platform/orchestration/actions/write_site_plan_imagery_wiring_test.go
//
// bugs_open/214 — the WIRING guard, driving the real WriteSitePlanAction.
//
// WHY THIS FILE IS SEPARATE FROM THE UNIT TESTS. The fifteen tests in
// write_site_plan_imagery_scope_test.go all call the resolution helpers
// DIRECTLY. Measured 2026-08-10 by mutation: delete the whole resolution block
// from WriteSitePlanAction and every one of them still passes. That is the
// exact failure this tree logged in WRONG_CALLS on the same day — "twelve
// passing tests, and the fix could be unwired without one turning red" — and a
// unit test can never close it, because the thing it must observe is the CALL,
// not the callee.
//
// So these tests drive the action end to end under sqlmock and assert on the
// value that reaches the INSERT bind. They fail if the resolution block is
// removed, bypassed, reordered after the insert, or fed the wrong side of the
// canonicalisation.
package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// writePlanParams builds the CollectedData shape the plan-builder's terminal
// step receives: one page the canonicaliser will RENAME (about -> about-index,
// the live gamesdesign shape), and imagery keyed by the planner's own spelling.
func writePlanParams(db *sql.DB, siteID uuid.UUID, imagery map[string]interface{}) ActionParams {
	return ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"target_site_id": siteID.String()},
			"page_plan": map[string]interface{}{
				"pages": []interface{}{
					map[string]interface{}{
						"name":      "about",
						"page_type": "section-index",
						"slug":      "about",
						"sections":  []interface{}{"hero", "content-block-about", "differentiators"},
					},
				},
			},
			"imagery": imagery,
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"target_site_id": "input_data.target_site_id",
		}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "44444444-4444-4444-4444-444444444444",
			StepName:        "write_site_plan",
		},
	}
}

// expectPlanWriteUpTo sets the expectations every run shares, up to and
// including the page/section inserts. Imagery expectations are added by the
// caller, because they are what each test is actually about.
func expectPlanWriteUpTo(mock sqlmock.Sqlmock, planID uuid.UUID) {
	mock.ExpectBegin()
	// No previous plan for this site: the supersede UPDATE finds nothing, which
	// the action tolerates via errors.Is(err, sql.ErrNoRows).
	mock.ExpectQuery("UPDATE site_plans").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("INSERT INTO site_plans").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(planID))
	mock.ExpectExec("INSERT INTO site_plan_pages").WillReturnResult(sqlmock.NewResult(1, 1))
	for i := 0; i < 3; i++ { // hero, content-block-about, differentiators
		mock.ExpectExec("INSERT INTO site_plan_sections").WillReturnResult(sqlmock.NewResult(1, 1))
	}
}

// TestWriteSitePlan_ImageryScopeRefIsCanonicalisedBeforeInsert is THE wiring
// guard. The bind pinned below is the canonical name — so if the resolution
// block is deleted from the action, the insert carries "about" and this fails.
func TestWriteSitePlan_ImageryScopeRefIsCanonicalisedBeforeInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, planID := uuid.New(), uuid.New()
	expectPlanWriteUpTo(mock, planID)

	// The assertion: scope_ref (bind 3) must be the CANONICAL name the plan
	// actually wrote to site_plan_pages, not the planner's raw key.
	mock.ExpectExec("INSERT INTO site_plan_imagery").
		WithArgs(planID, "page", "about-index", "hero_about", "hero", "an about hero",
			sqlmock.AnyArg(), sqlmock.AnyArg(), 0, "llm").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO site_plan_imagery").
		WithArgs(planID, "section", "about-index:2", "icon_no_ads", "icon", "no ads",
			sqlmock.AnyArg(), sqlmock.AnyArg(), 0, "llm").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := writePlanParams(db, siteID, map[string]interface{}{
		"pages": map[string]interface{}{
			"about": []interface{}{
				map[string]interface{}{"key": "hero_about", "kind": "hero", "prompt": "an about hero"},
			},
		},
		"sections": map[string]interface{}{
			"about:2": []interface{}{
				map[string]interface{}{"key": "icon_no_ads", "kind": "icon", "prompt": "no ads"},
			},
		},
	})

	out, err := WriteSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, _ := out.(map[string]interface{})
	if got := res["imagery_refs_canonicalised"]; got != 2 {
		t.Errorf("imagery_refs_canonicalised = %v, want 2", got)
	}
	if got := res["imagery_refs_unresolved"]; got != 0 {
		t.Errorf("imagery_refs_unresolved = %v, want 0", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations — the site_plan_imagery bind is the wiring assertion.\n"+
			"A mismatch reading scope_ref=%q means the resolution block in WriteSitePlanAction is not running: %v",
			"about", err)
	}
}

// TestWriteSitePlan_UnresolvedImageryRefIsWrittenVerbatimAndRecorded pins both
// halves of the miss path through the real action: the row is NOT dropped (it
// keeps today's exact behaviour), and the durable record is written AFTER the
// commit — so a rolled-back plan write cannot leave error rows describing
// imagery that was never persisted.
func TestWriteSitePlan_UnresolvedImageryRefIsWrittenVerbatimAndRecorded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, planID := uuid.New(), uuid.New()
	expectPlanWriteUpTo(mock, planID)

	// Written verbatim — an unresolvable ref must never be silently dropped.
	mock.ExpectExec("INSERT INTO site_plan_imagery").
		WithArgs(planID, "page", "team", "hero_team", "hero", "a team hero",
			sqlmock.AnyArg(), sqlmock.AnyArg(), 0, "llm").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// ...and only THEN the durable record. Ordering is part of the assertion:
	// sqlmock is ordered by default, so an INSERT expected after ExpectCommit
	// fails if the code logs before committing.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"write_site_plan", sqlmock.AnyArg(), "IMAGERY_SCOPE_REF_UNRESOLVED", "warning",
			sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := writePlanParams(db, siteID, map[string]interface{}{
		"pages": map[string]interface{}{
			"team": []interface{}{
				map[string]interface{}{"key": "hero_team", "kind": "hero", "prompt": "a team hero"},
			},
		},
	})

	out, err := WriteSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, _ := out.(map[string]interface{})
	if got := res["imagery_refs_unresolved"]; got != 1 {
		t.Errorf("imagery_refs_unresolved = %v, want 1", got)
	}
	if got := res["imagery_written"]; got != 1 {
		t.Errorf("imagery_written = %v, want 1 — the row must survive, not be dropped", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations (the agent_error_log INSERT after ExpectCommit is the durable-record-and-timing assertion): %v", err)
	}
}

// TestWriteSitePlan_AlreadyCanonicalImageryRefIsUntouched is the no-regression
// guard at the seam: a planner that already keys imagery by the plan's real
// page name must see byte-identical behaviour, and must not trip the miss path.
func TestWriteSitePlan_AlreadyCanonicalImageryRefIsUntouched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, planID := uuid.New(), uuid.New()
	expectPlanWriteUpTo(mock, planID)

	mock.ExpectExec("INSERT INTO site_plan_imagery").
		WithArgs(planID, "page", "about-index", "hero_about", "hero", "h",
			sqlmock.AnyArg(), sqlmock.AnyArg(), 0, "llm").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// No agent_error_log INSERT expected: a resolved ref records nothing.

	params := writePlanParams(db, siteID, map[string]interface{}{
		"pages": map[string]interface{}{
			"about-index": []interface{}{
				map[string]interface{}{"key": "hero_about", "kind": "hero", "prompt": "h"},
			},
		},
	})

	out, err := WriteSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, _ := out.(map[string]interface{})
	if got := res["imagery_refs_canonicalised"]; got != 0 {
		t.Errorf("imagery_refs_canonicalised = %v, want 0 — an already-correct ref must not be rewritten", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
