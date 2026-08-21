// FILE: platform/orchestration/loop_skip_persists_test.go
//
// bugs_open/343 P2 — the loop-skip advance persists with the same optimistic-lock
// retry discipline as every other advance, and the awaited row goes terminal only
// once that advance is durable.
//
// The defect: skipToNextLoopIteration mutated the caller's state and wrote it with
// ONE unretried repo.UpdateState. An optimistic-lock failure there returned an
// error and lost the advance — and on the async path it lost the awaited-map
// delete with it, while the table row had ALREADY been marked terminal, so the
// reply's redelivery was eaten by the processed_at duplicate guard. A lost
// continuation, on exactly the error path that carried all 31 of 343's observed
// terminal outcomes.
//
// This is filed as a PATH repair: verified as a path, UNVERIFIED as ever having
// fired. Nothing here claims it is the wedge's mechanism.
package orchestration

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// loopSkipPlan is a two-iteration loop whose iter_0 substep has failed.
func loopSkipPlan() models.WorkflowPlan {
	return models.WorkflowPlan{
		Steps: map[string]models.Step{
			"process_items_iter_0_call_handler": {Action: "call_agent", NextStep: "process_items_iter_0_save"},
			"process_items_iter_0_save":         {Action: "save"},
			"process_items_iter_1_call_handler": {Action: "call_agent", NextStep: "process_items_iter_1_save"},
			"process_items_iter_1_save":         {Action: "save"},
			"process_items_complete":            {Action: "noop", Config: map[string]interface{}{"total_iterations": float64(2)}},
		},
	}
}

func loopSkipState() *OrchestrationState {
	return &OrchestrationState{
		OrchestrationID: "orch-1",
		CurrentStep:     "process_items_iter_0_call_handler",
		Status:          StatusExecutingStep,
		Version:         7,
		CollectedData:   map[string]interface{}{},
		AwaitedRequests: map[string]*AwaitedRequest{
			"req-dead": {RequestID: "req-dead", StepName: "process_items_iter_0_call_handler"},
		},
		WorkflowPlan: loopSkipPlan(),
	}
}

// loopSkipRow is the DB copy the retry reloads.
func loopSkipRow(version int) *sqlmock.Rows {
	now := time.Now()
	plan := `{"steps":{` +
		`"process_items_iter_0_call_handler":{"action":"call_agent","next_step":"process_items_iter_0_save"},` +
		`"process_items_iter_0_save":{"action":"save"},` +
		`"process_items_iter_1_call_handler":{"action":"call_agent","next_step":"process_items_iter_1_save"},` +
		`"process_items_iter_1_save":{"action":"save"},` +
		`"process_items_complete":{"action":"noop","config":{"total_iterations":2}}}}`
	return sqlmock.NewRows(freshRowColumns).AddRow(
		"orch-1", "build-dispatch-loop", "corr-1", "agent-1", "build-dispatch-loop",
		"owner", nil, "client-1",
		"requests.topic", "responses.topic",
		"EXECUTING_STEP", "process_items_iter_0_call_handler", []byte(`[]`),
		[]byte(`{"req-dead":{"request_id":"req-dead","step_name":"process_items_iter_0_call_handler"}}`),
		nil, now, "pod-1", nil,
		[]byte(`{}`), []byte(`{}`), nil, []byte(plan),
		[]byte(`[]`), []byte(`{}`), []byte(`[]`), []byte(`{}`),
		0, nil, version, now, now,
		nil,
	)
}

func collectedDataCaptor(dest *string) []driver.Value {
	args := make([]driver.Value, updateArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[collectedDataArgIndex] = argCaptor{dest: dest}
	return args
}

// THE REGRESSION TEST. One optimistic-lock failure must not lose the advance.
//
// Mutation that must fail this: restore the single unretried repo.UpdateState —
// the first failure returns an error, no second UPDATE is issued, the expectation
// goes unmet, and the awaited row is never marked complete. That is the lost
// continuation, reproduced.
func TestLoopSkipSurvivesAnOptimisticLockFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	var persisted string

	// Attempt 1: the row moved under us.
	mock.ExpectExec("UPDATE orchestration_states").
		WillReturnError(errors.New("optimistic lock failure: state was modified by another process"))
	// The retry reloads...
	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(loopSkipRow(8))
	// ...re-applies, and writes. THIS is what the old code never did.
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(collectedDataCaptor(&persisted)...).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Only now may the awaited row be retired.
	mock.ExpectExec("UPDATE awaited_requests").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	state := loopSkipState()

	// continueExecution runs after the persist and will fail against this mock;
	// the advance itself is what this test is about, so the error is not fatal.
	_ = s.skipToNextLoopIterationWithAwaited(context.Background(), state, "req-dead", "child never answered", zap.NewNop())

	if state.CurrentStep != "process_items_iter_1_call_handler" {
		t.Errorf("current step = %q, want the next iteration's first substep - the advance was lost", state.CurrentStep)
	}
	if _, present := state.AwaitedRequests["req-dead"]; present {
		t.Error("the resolved awaited request came back from the reloaded row - the delete must be re-applied on every attempt")
	}
	if persisted == "" {
		t.Fatal("no collected_data was captured: the retry never issued its UPDATE")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the loop skip did not persist through an optimistic-lock failure: %v", err)
	}
}

// The awaited row must NOT be marked terminal when the advance never persisted:
// leaving it claimable is what lets the redelivery still drive the continuation,
// instead of being eaten by the processed_at duplicate guard.
//
// Mutation that must fail this, VERIFIED rather than assumed: retire the row on
// the error return path too (one added MarkAwaitedRequestComplete before the
// non-lock `return`). That kills THIS test and no other.
//
// ⚠ It is NOT the test that catches marking-before-persist. That mutation is
// killed by TestLoopSkipSurvivesAnOptimisticLockFailure and
// TestLoopSkipCleanPathWritesOnceThenRetiresTheRow, whose ordered expectations
// see the awaited UPDATE arrive out of turn. Recorded because the first draft of
// this comment claimed the wrong mutation and the claim survived a passing run —
// a test comment naming a mutation nobody has actually applied is a claim about
// behaviour, not the behaviour.
func TestAwaitedRowIsNotRetiredWhenTheAdvanceDoesNotPersist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// A non-lock error is terminal: one attempt, no retry, no mark.
	mock.ExpectExec("UPDATE orchestration_states").
		WillReturnError(errors.New("connection reset by peer"))
	// This expectation MUST REMAIN UNMET. Asserting the absence positively is the
	// only way to see it: MarkAwaitedRequestComplete's error is deliberately
	// swallowed (it logs a Warn and continues), so a mock rejection is invisible
	// to the code under test and ExpectationsWereMet() would read clean whether
	// or not the call happened. A mock's own bookkeeping cannot assert a negative
	// unless you give it something to leave undone.
	mock.ExpectExec("UPDATE awaited_requests").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	state := loopSkipState()

	err = s.skipToNextLoopIterationWithAwaited(context.Background(), state, "req-dead", "child never answered", zap.NewNop())
	if err == nil {
		t.Fatal("a failed persist must be reported, not swallowed")
	}
	unmet := mock.ExpectationsWereMet()
	if unmet == nil {
		t.Fatal("the awaited row WAS retired although the advance never persisted - a redelivery will now be eaten by the processed_at duplicate guard and the continuation is lost")
	}
	if !strings.Contains(unmet.Error(), "awaited_requests") {
		t.Errorf("something other than the awaited-row mark went unmet, so this test is not proving what it claims: %v", unmet)
	}
}

// The clean path: no lock contention, one write, then the mark. This is the
// control — if it fails, the retry loop has broken the ordinary case.
func TestLoopSkipCleanPathWritesOnceThenRetiresTheRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	var persisted string
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(collectedDataCaptor(&persisted)...).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE awaited_requests").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	state := loopSkipState()
	_ = s.skipToNextLoopIterationWithAwaited(context.Background(), state, "req-dead", "child never answered", zap.NewNop())

	if state.CurrentStep != "process_items_iter_1_call_handler" {
		t.Errorf("current step = %q, want the next iteration's first substep", state.CurrentStep)
	}
	if state.Status != StatusExecutingStep {
		t.Errorf("status = %q, want %q", state.Status, StatusExecutingStep)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("clean path: %v", err)
	}
}

// The synchronous caller has no awaited request in play, so nothing on
// awaited_requests may be touched at all.
func TestSyncLoopSkipTouchesNoAwaitedRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE orchestration_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// No awaited_requests expectation.

	s := &SagaCoordinator{db: db, logger: zap.NewNop()}
	state := loopSkipState()
	_ = s.skipToNextLoopIteration(context.Background(), state, "step failed", zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the synchronous path touched an awaited row: %v", err)
	}
}
