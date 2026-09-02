// FILE: platform/orchestration/actions/growth_posture_door_test.go
//
// The growth-posture DOOR in writeWorkItem (owner decision 5, 2026-08-31,
// WDS-020): a growth-gated item filed against a site whose
// settings->maintenance_profile->>growth_posture is 'hold' is written in the
// record shape — status 'deferred' ($12), handler_agent '' ($11) — which the
// detected-item-promoter cannot score (its CTE excludes handler-less rows
// before any door runs; write_audit_findings_filing_mode_test proves that
// shape unpromotable). The spec ($7) carries the held marker, the original
// handler, and the release recipe, so release stays a one-UPDATE human verb.
//
// The door sits at writeWorkItem — the seam EVERY filing crosses
// (insertWorkItem wraps it) — so these tests drive writeWorkItem directly,
// the same harness as the unregistered-handler door beside it. The bypass
// (source 'owner-request') and the gated-type set are proven on the PURE half
// (datahelpers.GrowthGateApplies); the non-gated case is additionally proven
// at the door in ORDERED mock mode, where an unexpected posture query is a
// hard mismatch.

package actions

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func expectGrowthPostureRead(mock sqlmock.Sqlmock, siteID, posture string) {
	mock.ExpectQuery(`growth_posture`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"posture"}).AddRow(posture))
}

func growthItem(siteID uuid.UUID) workItem {
	return workItem{
		siteID:       siteID,
		source:       "tool-suggester",
		pipeline:     "build",
		itemType:     "add_tool",
		severity:     "low",
		summary:      "Add tool: Livestock Value Estimator",
		spec:         `{"name":"Livestock Value Estimator"}`,
		status:       "triaged",
		handlerAgent: "tool-generator",
		createdBy:    "tool-suggester",
		itemKey:      "add_tool:" + uuid.NewString(),
	}
}

// Mutation proof: delete the applyGrowthPostureDoor call in writeWorkItem and
// this fails on $12 ('triaged' instead of 'deferred') and $11.
func TestWriteWorkItem_GrowthHold_BornHeld(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectBegin()
	expectGrowthPostureRead(mock, siteID.String(), "hold")
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"[growth held] Add tool: Livestock Value Estimator", // $6 summary
			`{"growth_handler":"tool-generator","growth_held":true,"growth_release_recipe":"owner release: UPDATE site_work_items SET status='detected', handler_agent=spec-\u003e\u003e'growth_handler' WHERE id='\u003cthis row\u003e'","name":"Livestock Value Estimator"}`, // $7 spec — producer keys preserved, growth keys beside them (json.Marshal HTML-escapes > and <)
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"",         // $11 handler_agent — the promoter cannot score this row
			"deferred", // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	w, err := writeWorkItem(context.Background(), tx, growthItem(siteID), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("a held filing must succeed — holding FILES, it does not refuse: %v", err)
	}
	if !w.Inserted {
		t.Fatal("the held row must be INSERTED — losing the signal is what filing-not-skipping exists to avoid")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Posture 'open' — and, by the same arm, ANY value other than 'hold' — leaves
// the filing byte-identical to a world with no door.
func TestWriteWorkItem_GrowthOpenOrUnknownPosture_Untouched(t *testing.T) {
	for _, posture := range []string{"open", "review"} {
		db, mock := newInsertMock(t)
		siteID := uuid.New()

		mock.ExpectBegin()
		expectGrowthPostureRead(mock, siteID.String(), posture)
		expectHandlerRegisteredProbe(mock, "tool-generator", true)
		expectInsertWithSummaryAndStatus(mock, "Add tool: Livestock Value Estimator", "triaged")

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		if _, err := writeWorkItem(context.Background(), tx, growthItem(siteID), dropOnConflict, zap.NewNop()); err != nil {
			t.Fatalf("posture %q: open filing must succeed: %v", posture, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("posture %q: unmet expectations: %v", posture, err)
		}
		db.Close()
	}
}

// A posture read ERROR fails open: the item files exactly as it would with no
// door at all. An opt-in hold must not stop fleet growth by breaking — the
// symmetric test to the owned-page door's probe-failure backstop.
func TestWriteWorkItem_GrowthPostureReadError_FailsOpen(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`growth_posture`).
		WithArgs(siteID.String()).
		WillReturnError(fmt.Errorf("connection reset"))
	expectHandlerRegisteredProbe(mock, "tool-generator", true)
	expectInsertWithSummaryAndStatus(mock, "Add tool: Livestock Value Estimator", "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := writeWorkItem(context.Background(), tx, growthItem(siteID), dropOnConflict, zap.NewNop()); err != nil {
		t.Fatalf("fail-open filing must succeed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A NON-GATED item type never pays the posture read: the classic probe-and-
// insert sequence with no growth query expected, in ORDERED mock mode so an
// extra query is a hard mismatch rather than a silently-tolerated call.
func TestWriteWorkItem_NonGatedType_NeverProbed(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New()

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	expectHandlerRegisteredProbe(mock, "tool-generator", true)
	expectInsertWithSummaryAndStatus(mock, "Re-render page after tool improvement", "triaged")

	item := growthItem(siteID)
	item.itemType = "needs_rerender" // not growth-gated
	item.summary = "Re-render page after tool improvement"

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop()); err != nil {
		t.Fatalf("non-gated filing must succeed with no posture read: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
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
		{"add_tool", "owner-request", false},            // the owner asking is not growth to refuse
		{"evaluate_tools", "owner-request", false},      // bypass is source-, not type-shaped
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
