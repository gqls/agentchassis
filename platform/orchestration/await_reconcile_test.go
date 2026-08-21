// FILE: platform/orchestration/await_reconcile_test.go
//
// bugs_open/343 — the advance decision cross-checks the awaited_requests table.
//
// "All responses are in" is decided from the AwaitedRequests JSONB map alone, and
// the table that drives the response consumer is never consulted. Where the two
// disagree, the orchestration advances past work it still owes and nothing
// records that it happened. These tests pin the two halves separately, because a
// detector and an enforcer in series can hide each other: a mutation that kills
// one must fail its own test while the other still passes.
package orchestration

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// outstandingRows builds the rows OutstandingAwaitedRequests scans, in the order
// of awaitedRequestColumns.
func outstandingRows(reqIDs ...string) *sqlmock.Rows {
	cols := []string{
		"request_id", "orchestration_id", "correlation_id", "step_id", "step_name",
		"retry_version", "target_agent_id", "target_agent_type",
		"responses_topic", "requests_topic", "sent_at", "timeout_at",
		"reply_to_request_id", "request_payload", "status", "processed_at",
	}
	rows := sqlmock.NewRows(cols)
	now := time.Now()
	for _, id := range reqIDs {
		rows.AddRow(
			id, "orch-1", "corr-1", "step-1", "process_item_iter_1_call_handler",
			0, "agent-1", "worker",
			"responses.topic", "requests.topic", now, now.Add(20*time.Minute),
			// reply_to_request_id is "" not NULL: the shared scanner reads it into
			// a plain string, and live rows agree — 0 of 34,808 are NULL
			// [MEASURED 2026-08-21]. A NULL fixture here tests a shape the table
			// does not produce, and fails on the scan instead of on the assertion.
			"", nil, "waiting", nil,
		)
	}
	return rows
}

// stateWithPlan builds the fresh state the reconcile runs against.
func stateWithPlan(enforce bool) *OrchestrationState {
	return &OrchestrationState{
		OrchestrationID: "orch-1",
		CurrentStep:     "process_items_loop",
		Status:          StatusExecutingStep,
		AwaitedRequests: map[string]*AwaitedRequest{},
		WorkflowPlan: models.WorkflowPlan{
			AwaitReconcileEnforce: enforce,
			Steps: map[string]models.Step{
				"process_items_loop": {Action: "loop", NextStep: "finish"},
			},
		},
	}
}

// DETECTION, flag OFF. The table disagrees; the divergence must be recorded, and
// the decision must stay byte-identical to today (allDone survives).
//
// Mutation 1: remove the OutstandingAwaitedRequests call → the expected SELECT
// goes unmet and this fails. Mutation 2: let detection change the decision with
// the flag off → the allDone assertion fails. The two mutations fail different
// assertions in the same test, so neither can hide behind the other.
func TestDivergenceIsDetectedAndTheFlagOffDecisionIsUnchanged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM awaited_requests").
		WithArgs("orch-1", "req-completing").
		WillReturnRows(outstandingRows("req-still-out"))
	// The durable breadcrumb: detection must record the divergence, not only log it.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	repo := NewStateRepository(db, zap.NewNop())
	state := stateWithPlan(false)

	got := s.reconcileAllDoneAgainstTable(context.Background(), repo, state, "req-completing", true)

	if !got {
		t.Error("with await_reconcile_enforce OFF the decision must be exactly today's - detection may not change what runs")
	}
	if state.Status != StatusExecutingStep {
		t.Errorf("status = %q, want it untouched (%q) with the flag off", state.Status, StatusExecutingStep)
	}
	if len(state.AwaitedRequests) != 0 {
		t.Errorf("the map was mutated with the flag off: %v", state.AwaitedRequests)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("divergence went unqueried or unrecorded: %v", err)
	}
}

// ENFORCEMENT, flag ON. The orchestration adopts the table's outstanding rows and
// re-parks instead of advancing.
//
// Mutation: delete the enforcement branch → this fails while the detection test
// above still passes.
func TestEnforceAdoptsTheTablesOutstandingRowsAndReparks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM awaited_requests").
		WithArgs("orch-1", "req-completing").
		WillReturnRows(outstandingRows("req-still-out"))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	repo := NewStateRepository(db, zap.NewNop())
	state := stateWithPlan(true)

	got := s.reconcileAllDoneAgainstTable(context.Background(), repo, state, "req-completing", true)

	if got {
		t.Fatal("with enforcement on, a table row still outstanding must stop the advance")
	}
	if state.Status != StatusAwaitingResponses {
		t.Errorf("status = %q, want %q after adopting", state.Status, StatusAwaitingResponses)
	}
	adopted, present := state.AwaitedRequests["req-still-out"]
	if !present {
		t.Fatalf("the outstanding row was not adopted into the map: %v", state.AwaitedRequests)
	}
	if adopted.StepName != "process_item_iter_1_call_handler" {
		t.Errorf("adopted entry lost its step name: %+v", adopted)
	}
}

// The clean path: table and map agree, so nothing is logged and the advance
// proceeds. This is the control — if it ever fails, detection is firing on
// ordinary traffic and the query needs a grace age, not deletion.
func TestNoDivergenceLeavesTheDecisionAloneAndRecordsNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM awaited_requests").
		WithArgs("orch-1", "req-completing").
		WillReturnRows(outstandingRows())
	// No INSERT expectation: recording a clean run is the failure.

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	repo := NewStateRepository(db, zap.NewNop())
	state := stateWithPlan(false)

	if got := s.reconcileAllDoneAgainstTable(context.Background(), repo, state, "req-completing", true); !got {
		t.Error("a clean cross-check must leave allDone true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the clean path wrote something it should not have: %v", err)
	}
}

// A cross-check must never become a new way to fail: a query error warns and
// lets the caller proceed exactly as today.
func TestQueryFailureFallsBackToTheInMemoryDecision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM awaited_requests").
		WithArgs("orch-1", "req-completing").
		WillReturnError(driver.ErrBadConn)

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	repo := NewStateRepository(db, zap.NewNop())

	// Even with enforcement ON, an unreadable table cannot block the advance.
	if got := s.reconcileAllDoneAgainstTable(context.Background(), repo, stateWithPlan(true), "req-completing", true); !got {
		t.Error("a failed cross-check changed the decision - detection must be best-effort in the failure direction")
	}
}

// allDone false short-circuits: there is nothing to cross-check, and the query
// must not run at all.
func TestNotAllDoneSkipsTheCrossCheckEntirely(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// No expectations at all.

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	repo := NewStateRepository(db, zap.NewNop())

	if got := s.reconcileAllDoneAgainstTable(context.Background(), repo, stateWithPlan(false), "req-completing", false); got {
		t.Error("the cross-check invented an allDone the caller did not have")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the cross-check queried when the caller was still awaiting: %v", err)
	}
}
