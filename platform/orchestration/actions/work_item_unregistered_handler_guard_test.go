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

// Covers the write-door routability guard (bugs_open/291): a work item born in
// a dispatchable status whose handler_agent names no registered agent is
// demoted to 'blocked' at the door, with the same error text claim would write
// one tick later. tool-auditor filed 14 real findings at 'hitl-review' — an
// agent that has NEVER existed — and every one was claimed, flipped to
// blocked, and pinned a dedup slot that silently dropped the auditor's later
// findings for the same page.
//
// The probe expectation in these tests is BUILT FROM the shared renderer
// (workItemHandlerRegisteredSQL), the same coupling proof
// TestClaimAndPromoterAskTheSameQuestion uses: hand-write the probe SQL in
// writeWorkItem (e.g. add `AND is_active`) and the expectation stops matching.

// expectHandlerRegisteredProbe asserts the door runs claim's own registration
// predicate, byte-identical, and scripts its answer.
func expectHandlerRegisteredProbe(mock sqlmock.Sqlmock, handler string, registered bool) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT " + workItemHandlerRegisteredSQL("$1"))).
		WithArgs(handler).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(registered))
}

// expectInsertWithStatusAndError asserts the row is written blocked ($12) WITH
// the conditional 17th error argument — the same conditional-append idiom as
// parent_item_id, so callers that never trip the guard keep their sixteen args.
func expectInsertWithStatusAndError(mock sqlmock.Sqlmock, status, errText string) {
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			status, // $12
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			errText, // $17 — only present because the guard fired
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func guardItem(status, handler string) workItem {
	return workItem{
		siteID:       uuid.New(),
		source:       "tool-auditor",
		pipeline:     "build",
		itemType:     "needs_human_review",
		severity:     "low",
		summary:      "None of the four number inputs have associated labels",
		spec:         "{}",
		status:       status,
		handlerAgent: handler,
		createdBy:    "tool-auditor",
		itemKey:      "audit_review_" + uuid.NewString(),
		// Isolate the guard from the anti-churn block; the guard runs after it
		// either way (it must see the FINAL status, including 'unresolved').
		recurrenceExpected: true,
	}
}

// The bug's exact shape: born dispatchable at a handler that does not exist.
// The row must land — Inserted true, the finding is durable — but at 'blocked'
// with claim's error text. Mutation proof: delete the guard block in
// writeWorkItem and this fails on $12 ('triaged' instead of 'blocked').
func TestWriteWorkItem_UnregisteredHandler_BornBlocked(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	expectHandlerRegisteredProbe(mock, "hitl-review", false)
	expectInsertWithStatusAndError(mock, "blocked",
		"Handler agent not registered: hitl-review")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx,
		guardItem("triaged", "hitl-review"), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if !w.Inserted {
		t.Fatal("a demoted item must still be INSERTED — losing the finding is the failure mode the demotion exists to avoid")
	}
	if !w.BornBlocked {
		t.Fatal("BornBlocked must report the demotion to the caller")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A registered handler passes through untouched: same status, the classic
// sixteen arguments, no error column.
func TestWriteWorkItem_RegisteredHandler_Untouched(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	expectHandlerRegisteredProbe(mock, "tool-improver", true)
	expectInsertWithSummaryAndStatus(mock,
		"None of the four number inputs have associated labels", "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx,
		guardItem("triaged", "tool-improver"), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if !w.Inserted || w.BornBlocked {
		t.Fatalf("registered handler must insert undemoted, got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The scope pin: parked, deferred and pre-promotion statuses are NEVER probed,
// and neither is an empty handler at any status. Each of these shapes is a
// deliberate platform idiom that a wider guard would demote to blocked,
// recreating bugs_closed/284 inside 291's fix:
//   - needs_human_review + 'human-review': the checkpoint pseudo-handler
//   - deferred + 'tool-builder': capability_gap naming an UNBUILT builder
//   - detected + a real handler: judged later, at promotion
//   - triaged + empty handler: that shape is CHECK 443's territory, not ours
//
// Mutation proof: widen workItemStatusRequiresRegisteredHandler (or drop the
// handler != "" condition) and the tripwire below demotes the item, failing
// the INSERT match on $12; the function-level pin one test down fails by name.
func TestWriteWorkItem_ParkedAndPrePromotionShapes_NeverProbed(t *testing.T) {
	cases := []struct {
		name    string
		status  string
		handler string
	}{
		{"parked_pseudo_handler", "needs_human_review", "human-review"},
		{"deferred_unbuilt_builder", "deferred", "tool-builder"},
		{"detected_judged_at_promotion", "detected", "page-build-handler"},
		{"empty_handler_is_check_443s_job", "triaged", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()

			// TRIPWIRE, not an expectation to be met: the probe is scripted to
			// answer "not registered", so a build that wrongly probes this
			// shape demotes the item and the INSERT below fails on $12.
			// A bare no-probe-expectation cannot catch that mutation — the
			// guard's own probe-failure fall-through swallows sqlmock's
			// "unexpected query" error and inserts normally (learned the hard
			// way: the widened-set mutation PASSED that version of this test).
			// Unordered matching + no ExpectationsWereMet: the tripwire is
			// deliberately unconsumed when the code is correct.
			mock.MatchExpectationsInOrder(false)
			mock.ExpectBegin()
			expectHandlerRegisteredProbe(mock, tc.handler, false)
			expectInsertWithSummaryAndStatus(mock,
				"None of the four number inputs have associated labels", tc.status)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}

			w, err := writeWorkItem(context.Background(), tx,
				guardItem(tc.status, tc.handler), dropOnConflict, zap.NewNop())
			if err != nil {
				t.Fatalf("writeWorkItem: %v", err)
			}
			if !w.Inserted || w.BornBlocked {
				t.Fatalf("shape must insert untouched, got %+v", w)
			}
		})
	}
}

// The scope pin in its most legible, mutation-sensitive form: the trigger set
// is EXACTLY CHECK swi_no_handlerless_promotable's status list. Widen or
// narrow workItemStatusRequiresRegisteredHandler and this fails by name.
func TestStatusRequiresRegisteredHandler_ExactlyCheck443sList(t *testing.T) {
	requires := map[string]bool{
		"triaged": true, "approved": true, "claimed": true,
		"detected": false, "deferred": false, "needs_human_review": false,
		"blocked": false, "pending_review": false, "complete": false,
		"failed": false, "unresolved": false, "cancelled": false, "": false,
	}
	for status, want := range requires {
		if got := workItemStatusRequiresRegisteredHandler(status); got != want {
			t.Errorf("workItemStatusRequiresRegisteredHandler(%q) = %v, want %v", status, got, want)
		}
	}
}

// A failed probe must not take the write down with it: log, skip the guard,
// write the item as asked — claim's branch remains the backstop. (Same posture
// claim itself takes on a failed handler read.)
func TestWriteWorkItem_ProbeFailure_FallsThroughToClaimBackstop(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT " + workItemHandlerRegisteredSQL("$1"))).
		WithArgs("hitl-review").
		WillReturnError(context.DeadlineExceeded)
	expectInsertWithSummaryAndStatus(mock,
		"None of the four number inputs have associated labels", "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx,
		guardItem("triaged", "hitl-review"), dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem must not fail on a probe error: %v", err)
	}
	if !w.Inserted || w.BornBlocked {
		t.Fatalf("probe failure must fall through undemoted, got %+v", w)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// create_work_item action: the relaxed handler validation (bugs_open/291)
// ---------------------------------------------------------------------------

// A PARKED item may omit the handler: status needs_human_review + no
// handler_agent is the platform's HITL idiom (migration 217). Before 291 this
// action refused the shape outright, which is what forced tool-auditor's
// config to name SOME handler — and the name it picked had never existed.
func TestCreateWorkItemAction_ParkedItemMayOmitHandler(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.MatchExpectationsInOrder(false)
	mock.ExpectBegin()
	expectInsertWithSummaryAndStatus(mock, "needs_human_review", "needs_human_review")
	mock.ExpectCommit()

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"site_id": uuid.New().String()},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":   "input_data.site_id",
			"item_type": "needs_human_review",
			"status":    "needs_human_review",
			"source":    "tool-auditor",
		}},
	}
	out, err := CreateWorkItemAction(context.Background(), params)
	if err != nil {
		t.Fatalf("a parked item with no handler must be accepted: %v", err)
	}
	result, _ := out.(map[string]interface{})
	if result["born_blocked"] != false {
		t.Errorf("a parked item never trips the door guard, got %+v", result)
	}
}

// A DISPATCHABLE item (or one whose omitted status would default to triaged)
// still requires a handler — handler-less dispatchable is CHECK 443's
// forbidden shape and could only ever become claim's blocked.
func TestCreateWorkItemAction_DispatchableItemOmittingHandlerIsRefused(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	_ = mock

	for _, status := range []string{"", "triaged", "approved"} {
		cfg := map[string]interface{}{
			"site_id":   "input_data.site_id",
			"item_type": "needs_rerender",
			"source":    "tool-improver",
		}
		if status != "" {
			cfg["status"] = status
		}
		params := ActionParams{
			Context:          context.Background(),
			DB:               db,
			Logger:           zap.NewNop(),
			ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
			CollectedData: map[string]interface{}{
				"input_data": map[string]interface{}{"site_id": uuid.New().String()},
			},
			StepConfig: models.Step{Config: cfg},
		}
		_, err := CreateWorkItemAction(context.Background(), params)
		if err == nil {
			t.Fatalf("status %q with no handler must be refused", status)
		}
		if !strings.Contains(err.Error(), "handler_agent config is required") {
			t.Fatalf("error must name the missing key, got: %v", err)
		}
	}
}
