// FILE: platform/orchestration/park_arrival_check_test.go
//
// bug 343 (silent post-abandonment freeze) — the park's arrival check, keyed on request IDENTITY.
//
// The defect these tests pin: the check used to ask "is there a response marker
// under this step name?", which on a loop step is true of every iteration after
// the first. A re-registration of an already-answered step name therefore read as
// "the reply beat the park", and the park returned success WITHOUT persisting —
// leaving a row in awaited_requests, nothing in the AwaitedRequests map, and an
// orchestration sitting in EXECUTING_STEP that only the 4-hour stale reaper would
// ever notice.
//
// Each test names the mutation that must fail it. They drive the REAL
// persistAwaitingStateWithRetry and processAwaitResponse against a mocked DB, so
// a change that bypasses the call site fails them too.
package orchestration

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// THE WEDGE REGRESSION TEST. A marker under this step name records an EARLIER
// request's reply; the request now being parked has never been answered, so the
// park must persist normally.
//
// Mutation that must fail this: revert the check to keying on the bare "response"
// marker (presence, not identity) — the park then skips, issues no UPDATE, and
// sqlmock's unmet expectation fails the test. That is the wedge, reproduced.
func TestParkProceedsOnStaleMarkerFromAnEarlierRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	updateArgs := make([]driver.Value, updateArgCount)
	for i := range updateArgs {
		updateArgs[i] = sqlmock.AnyArg()
	}

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"deploy_logo_image":{"response":{"data":{"success":true}},"response_request_id":"req-logo-0"}}`))
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(updateArgs...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewStateRepository(db, zap.NewNop())
	outcome, err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop())
	if err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if outcome != parkPersisted {
		t.Fatalf("outcome = %v, want parkPersisted: a marker naming a DIFFERENT request is stale residue, not an arrival", outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the park skipped on another request's marker - this is the bug 343 wedge: %v", err)
	}
}

// The genuine beat-the-park race: the marker names THIS request, so the response
// consumer has already applied the reply and owns the continuation.
//
// Mutation that must fail this: make the skip branch park anyway → an unexpected
// UPDATE, which sqlmock rejects.
func TestParkSkipsWhenThisRequestHasAlreadyBeenAnswered(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"deploy_logo_image":{"response":{"data":{"success":true}},"response_request_id":"req-logo-1"}}`))
	// Deliberately no ExpectExec: any UPDATE here is the failure.

	repo := NewStateRepository(db, zap.NewNop())
	outcome, err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop())
	if err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if outcome != parkSkippedReplyArrived {
		t.Fatalf("outcome = %v, want parkSkippedReplyArrived", outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("park wrote after a reply to THIS request had landed: %v", err)
	}
}

// The legacy compatibility branch: a marker written by an image that predates
// awaitedResponseIDMarker carries no id, identity is unrecoverable, and
// treat-as-arrived is today's behaviour — the safe reading while a mixed fleet is
// live.
//
// Mutation that must fail this: treat a no-id marker as stale and park — an
// unexpected UPDATE. (Do not "fix" the branch that way: an old pod's genuine
// arrival would be double-driven.)
func TestParkTreatsALegacyMarkerWithNoIDAsArrived(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"deploy_logo_image":{"response":{"data":{"success":true}}}}`))

	repo := NewStateRepository(db, zap.NewNop())
	outcome, err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop())
	if err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if outcome != parkSkippedReplyArrived {
		t.Fatalf("outcome = %v, want parkSkippedReplyArrived (legacy no-id marker)", outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("park wrote on a legacy marker: %v", err)
	}
}

// A marker recorded for a request on a DIFFERENT step name must not be consulted
// at all: the check reads the container under the parking request's own step.
func TestParkIgnoresMarkersOnOtherSteps(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	updateArgs := make([]driver.Value, updateArgCount)
	for i := range updateArgs {
		updateArgs[i] = sqlmock.AnyArg()
	}

	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"some_other_step":{"response":{"ok":true},"response_request_id":"req-logo-1"}}`))
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(updateArgs...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewStateRepository(db, zap.NewNop())
	outcome, err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop())
	if err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if outcome != parkPersisted {
		t.Fatalf("outcome = %v, want parkPersisted", outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// The case BETWEEN the two the council's editquality seat named in round 1
// (correlation RESUBMIT_CORR=cc782778): a container IS present under the awaited
// step name — the dispatching action's own carried-across result — but holds no
// response marker at all. This is the ordinary shape of a first park on a step
// that computed something before dispatching, and it must persist.
//
// It earns its own test because the two guards in front of the switch cover each
// other IN SERIES: with `exists` removed, a nil map still yields hasResponse
// false and continues; with `hasResponse` removed, a missing key still fails
// `exists` and continues. Removing either alone therefore changes nothing
// observable, which is exactly the "a mutation that passes may have hit a guard
// in series" shape — so the only mutation that proves this branch is removing
// BOTH, and this is the test that then fails.
//
// Mutation that must fail this: delete both the `if !exists` and the
// `if _, hasResponse := ...` guards. The switch then reaches `default` with no
// id, reads it as a legacy arrival, and skips a park that has never been
// answered — a fresh await silently abandoned on its FIRST attempt, which is
// worse than the bug this file exists to fix.
func TestParkPersistsWhenTheStepHasDataButNoResponseMarker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	updateArgs := make([]driver.Value, updateArgCount)
	for i := range updateArgs {
		updateArgs[i] = sqlmock.AnyArg()
	}

	// The action's own result is under the awaited step name; no reply has landed.
	mock.ExpectQuery("FROM orchestration_states").
		WithArgs("orch-1").
		WillReturnRows(freshRow(`{"deploy_logo_image":{"image_url":"/assets/images/logo.png","asset_id":"a-1"}}`))
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(updateArgs...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewStateRepository(db, zap.NewNop())
	outcome, err := persistAwaitingStateWithRetry(context.Background(), parkingState(), repo, zap.NewNop())
	if err != nil {
		t.Fatalf("park failed: %v", err)
	}
	if outcome != parkPersisted {
		t.Fatalf("outcome = %v, want parkPersisted: a step's own result is not an arrived reply", outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a first park was skipped because the step already held unrelated data: %v", err)
	}
}
