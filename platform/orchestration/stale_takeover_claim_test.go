// FILE: platform/orchestration/stale_takeover_claim_test.go
//
// bugs_open/329. Both takeover arms of handleOrchestrationStatus used to judge an
// orchestration "stuck" from the CALLER'S SNAPSHOT and then resume it, with nothing
// claiming the row. These tests drive the REAL arms against a mocked DB and assert
// on the traffic they actually produce.
//
// ⚠ READ THIS BEFORE WRITING ANOTHER CONCURRENCY TEST HERE. The obvious test is
// inverted. Exactly SIMULTANEOUS takers never double-executed: each one's write
// goes through UpdateStateWithVersion (UpdateState is a one-line wrapper for it —
// bugs_open/329 and bugs_closed/294 both assert the opposite and both are wrong),
// so the loser's CAS fails and its arm returns the error. A sync.WaitGroup
// start-line test therefore shows the BROKEN code passing. The defect is a
// check-then-act across TWO READS, and its disconfirming case is the SEQUENTIAL
// interleaving: taker A wins and refreshes last_activity; taker B arrives seconds
// later still holding a stale snapshot. That is what every test below sets up.
//
// ⚠ AND THE GUARDS IN SERIES. There are three layers around this defect: the
// chassis intake serialisation claim above it (agent-chassis only, keyed on the
// orchestration_id), these arms, and per-path CASes below (the work-item claim on
// the dispatch path). A test run on the dispatch path passes with the fix REVERTED,
// because the work-item CAS absorbs it. These tests call handleOrchestrationStatus
// directly — below the intake claim, above any action-level CAS — and sqlmock fails
// on any unexpected statement, so the expectation list IS the complete DB traffic of
// the span. There is no other guard in it to pass the test on this code's behalf.
//
// Written to reference only pre-existing symbols, so it compiles against unfixed
// HEAD and can be watched to FAIL there first:
//
//	scripts/verify-head-builds.sh --with platform/orchestration/stale_takeover_claim_test.go --test
package orchestration

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// staleRow is the row as the ARM's caller saw it: idle well past the threshold.
// takeoverRow is what a FRESH read returns once another actor has claimed and
// resumed it — status moved on, last_activity now, version bumped.
func takeoverRow(status string, currentlyExecuting interface{}, lastActivity time.Time, version int) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(freshRowColumns).AddRow(
		"orch-329", "site-work-orchestrator", "corr-1", "agent-1", "site-work-orchestrator",
		"owner", nil, "client-1",
		"requests.topic", "responses.topic",
		status, "step_a", []byte(`[]`), []byte(`{}`),
		currentlyExecuting, lastActivity, "pod-other", nil,
		[]byte(`{}`), []byte(`{}`), nil, []byte(`{}`),
		[]byte(`[]`), []byte(`{}`), []byte(`[]`), []byte(`{}`),
		0, nil, version, now, now,
		nil,
	)
}

// processingHistoryArgIndex is processing_history's zero-based position in
// UpdateStateWithVersion's UPDATE ($13, state.go:1031).
const processingHistoryArgIndex = 12

func staleSnapshot(status OrchestrationStatus, executing *string) *OrchestrationState {
	return &OrchestrationState{
		OrchestrationID:    "orch-329",
		CurrentStep:        "step_a",
		Status:             status,
		CurrentlyExecuting: executing,
		LastActivity:       time.Now().Add(-6 * time.Minute), // > StuckOrchestrationTimeout
		Version:            7,
	}
}

func newTestCoordinator(t *testing.T) (*SagaCoordinator, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	s := &SagaCoordinator{
		db:            db,
		logger:        zap.NewNop(),
		podName:       "pod-test",
		tracer:        types.NewTraceLogger(zap.NewNop()),
		retryCounters: map[string]int{},
	}
	return s, mock, func() { db.Close() }
}

func execCtxFor() *types.ExecutionContext {
	return &types.ExecutionContext{RequestID: "req-329", CorrelationID: "corr-1"}
}

// THE RACE, EXECUTING_STEP arm. The snapshot says stale; by the time we look, the
// row has been claimed and resumed by someone else. Nothing may be written and
// nothing may be executed.
//
// Against UNFIXED HEAD this fails: the arm calls ClearExecutingStep, which issues an
// UPDATE that sqlmock has not been told to expect, and the arm returns that error —
// naming, in the failure message, the resume that must not have happened.
func TestStaleExecutingStepTakeoverIsRefusedWhenTheFreshRowIsBeingDriven(t *testing.T) {
	s, mock, done := newTestCoordinator(t)
	defer done()

	executing := "step_a"
	// The claim's fresh read: another actor got there first.
	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-329").
		WillReturnRows(takeoverRow("EXECUTING_STEP", "step_a", time.Now(), 9))
	// and NO ExpectExec: a lost claim must write nothing.

	err := s.handleOrchestrationStatus(context.Background(),
		staleSnapshot(StatusExecutingStep, &executing), execCtxFor(), false)
	if err != nil {
		t.Fatalf("a lost takeover must return nil (the row is being driven), got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB traffic on a lost claim: %v", err)
	}
}

// THE RACE, RUNNING arm. This arm previously wrote NOTHING before resuming, so two
// arrivals seconds apart both resumed.
func TestStaleRunningResumeIsRefusedWhenTheFreshRowIsBeingDriven(t *testing.T) {
	s, mock, done := newTestCoordinator(t)
	defer done()

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-329").
		WillReturnRows(takeoverRow("EXECUTING_STEP", "step_a", time.Now(), 9))

	err := s.handleOrchestrationStatus(context.Background(),
		staleSnapshot(StatusRunning, nil), execCtxFor(), false)
	if err != nil {
		t.Fatalf("a lost takeover must return nil, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB traffic on a lost claim: %v", err)
	}
}

// THE VERSION RACE, through ExecuteWithOptimisticLocking's retry: our claim's CAS
// affects 0 rows because another taker's write landed first, we re-read, and the
// fresh row is no longer stale — so we stand down rather than retrying blind.
func TestSimultaneousTakersOnlyOneClaims(t *testing.T) {
	s, mock, done := newTestCoordinator(t)
	defer done()

	executing := "step_a"
	args := make([]driver.Value, updateArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}

	// Attempt 1: still stale on the read, but the CAS loses.
	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-329").
		WillReturnRows(takeoverRow("EXECUTING_STEP", "step_a", time.Now().Add(-6*time.Minute), 7))
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows => optimistic lock failure
	// Attempt 2: re-read shows the winner's refreshed row. Predicate now false.
	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-329").
		WillReturnRows(takeoverRow("EXECUTING_STEP", "step_a", time.Now(), 8))

	err := s.handleOrchestrationStatus(context.Background(),
		staleSnapshot(StatusExecutingStep, &executing), execCtxFor(), false)
	if err != nil {
		t.Fatalf("losing the version race then finding the row fresh must return nil, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// POSITIVE CONTROL. Without this, a fix that simply refuses every takeover passes
// all three tests above. A genuinely abandoned row must still be claimed and resumed.
//
// errStopHere is returned from continueExecution's FIRST database read, which proves
// the claim was won and execution was entered without letting the test run the whole
// workflow engine.
func TestStaleTakeoverClaimsAndResumesWhenNobodyElseHas(t *testing.T) {
	s, mock, done := newTestCoordinator(t)
	defer done()

	executing := "step_a"
	errStopHere := errors.New("sentinel: continueExecution was entered")

	var history string
	args := make([]driver.Value, updateArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[processingHistoryArgIndex] = argCaptor{dest: &history}

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-329").
		WillReturnRows(takeoverRow("EXECUTING_STEP", "step_a", time.Now().Add(-6*time.Minute), 7))
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1)) // the claim is won
	mock.ExpectQuery("FROM orchestration_states").
		WillReturnError(errStopHere)

	err := s.handleOrchestrationStatus(context.Background(),
		staleSnapshot(StatusExecutingStep, &executing), execCtxFor(), false)
	if err == nil {
		t.Fatal("expected the claim to be won and continueExecution entered; got nil (nothing resumed)")
	}
	if !strings.Contains(err.Error(), errStopHere.Error()) {
		t.Fatalf("expected to reach continueExecution's first read, got: %v", err)
	}
	if !strings.Contains(history, "stale_takeover_claimed") {
		t.Fatalf("the claim must leave a durable needle in processing_history; got: %s", history)
	}
}

// NEGATIVE CONTROL, both arms. A fresh row must be taken over by NEITHER, or the
// tests above pass for a fix that refuses everything AND for one that claims
// everything. No DB traffic at all is the assertion.
func TestFreshRowsAreNeverTakenOver(t *testing.T) {
	executing := "step_a"
	cases := []struct {
		name      string
		status    OrchestrationStatus
		executing *string
	}{
		{"executing_step", StatusExecutingStep, &executing},
		{"running", StatusRunning, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, done := newTestCoordinator(t)
			defer done()

			fresh := staleSnapshot(tc.status, tc.executing)
			fresh.LastActivity = time.Now() // not stale

			if err := s.handleOrchestrationStatus(context.Background(), fresh, execCtxFor(), false); err != nil {
				t.Fatalf("a fresh row must be left alone, got: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("a fresh row must produce NO database traffic: %v", err)
			}
		})
	}
}
