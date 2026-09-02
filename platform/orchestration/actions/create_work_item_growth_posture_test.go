// FILE: platform/orchestration/actions/create_work_item_growth_posture_test.go
//
// The growth-posture guard in CreateWorkItemAction (owner decision 5,
// 2026-08-31): a growth-gated item filed against a site whose
// settings->maintenance_profile->>growth_posture is 'hold' is born in the
// record shape — status 'deferred' ($12), handler_agent '' ($11) — which the
// detected-item-promoter cannot score (its CTE excludes handler-less rows
// before any door runs; write_audit_findings_filing_mode_test proves that
// shape unpromotable). The spec ($7) carries the held marker, the original
// handler, and the release recipe, so release stays a one-UPDATE human verb.
//
// The bypass (source 'owner-request') and the gated-type set are proven on
// the PURE half (datahelpers.GrowthGateApplies) — asserting "the posture
// query never ran" through sqlmock is vacuous here, because the guard fails
// open on an unexpected-query error and the un-held outcome would pass either
// way.

package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func expectGrowthPostureRead(mock sqlmock.Sqlmock, siteID, posture string) {
	mock.ExpectQuery(`growth_posture`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"posture"}).AddRow(posture))
}

// Mutation proof: delete the guard block in CreateWorkItemAction and this
// fails on $12 ('triaged' instead of 'deferred') and $11.
func TestCreateWorkItem_GrowthHold_AddToolBornHeld(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New().String()

	expectGrowthPostureRead(mock, siteID, "hold")
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"[growth held] Add tool: Livestock Value Estimator", // $6 summary
			`{"growth_handler":"tool-generator","growth_held":true,"growth_release_recipe":"owner release: UPDATE site_work_items SET status='detected', handler_agent=spec-\u003e\u003e'growth_handler' WHERE id='\u003cthis row\u003e'"}`, // $7 spec (json.Marshal HTML-escapes > and <)
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"",         // $11 handler_agent — the promoter cannot score this row
			"deferred", // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"site_id": siteID},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":       "input_data.site_id",
			"item_type":     "add_tool",
			"handler_agent": "tool-generator",
			"source":        "tool-suggester",
			"summary":       "Add tool: Livestock Value Estimator",
		}},
	}
	if _, err := CreateWorkItemAction(context.Background(), params); err != nil {
		t.Fatalf("held filing must succeed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Posture 'open' — and, by the same arm, ANY value other than 'hold' — leaves
// the filing byte-identical to a world with no switch.
func TestCreateWorkItem_GrowthOpenOrUnknownPosture_Untouched(t *testing.T) {
	for _, posture := range []string{"open", "review"} {
		db, mock := newInsertMock(t)
		siteID := uuid.New().String()

		expectGrowthPostureRead(mock, siteID, posture)
		mock.ExpectBegin()
		expectHandlerRegisteredProbe(mock, "tool-generator", true)
		expectInsertWithSummaryAndStatus(mock, "Add tool: X", "triaged")
		mock.ExpectCommit()

		params := ActionParams{
			Context:          context.Background(),
			DB:               db,
			Logger:           zap.NewNop(),
			ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
			CollectedData: map[string]interface{}{
				"input_data": map[string]interface{}{"site_id": siteID},
			},
			StepConfig: models.Step{Config: map[string]interface{}{
				"site_id":       "input_data.site_id",
				"item_type":     "add_tool",
				"handler_agent": "tool-generator",
				"source":        "tool-suggester",
				"summary":       "Add tool: X",
			}},
		}
		if _, err := CreateWorkItemAction(context.Background(), params); err != nil {
			t.Fatalf("posture %q: open filing must succeed: %v", posture, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("posture %q: unmet expectations: %v", posture, err)
		}
		db.Close()
	}
}

// The pure half: exactly the two chain HEADS consult the posture, and an
// explicit owner request never does. Mutation proof both ways: add a type to
// GrowthGatedItemTypes or drop the bypass and one of these arms fails.
func TestGrowthGateApplies_TypeSetAndOwnerBypass(t *testing.T) {
	cases := []struct {
		itemType, source string
		want             bool
	}{
		{"add_tool", "tool-suggester", true},
		{"evaluate_tools", "discovery", true},
		{"add_tool", "owner-request", false},          // the owner asking is not growth to refuse
		{"evaluate_tools", "owner-request", false},    // bypass is source-, not type-shaped
		{"needs_content_page", "tool-generator", false}, // downstream of the heads — dies with them
		{"needs_rerender", "discovery", false},
		{"content_rewrite", "site-review", false}, // audit growth is record-mode's, not this gate's
	}
	for _, c := range cases {
		if got := datahelpers.GrowthGateApplies(c.itemType, c.source); got != c.want {
			t.Errorf("GrowthGateApplies(%q, %q) = %v, want %v", c.itemType, c.source, got, c.want)
		}
	}
}
