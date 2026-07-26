package actions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// bugs_open/076 residual: the CONSUMER half of tolerate-a-truncation was
// recorded only by zap.Warn, and a pod log dies with its pod. These tests pin
// the durable half — one agent_error_log row per damaged seat, best-effort, and
// never at the cost of a decision that has already been persisted.

// runRecord drives recordTruncationDegradation against a mock DB and returns the
// context JSON of every row it wrote, in order.
func runRecord(t *testing.T, damage []truncationDegradation, execErr error) []map[string]interface{} {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	captured := make([]string, len(damage))
	codes := make([]string, len(damage))
	for i := range damage {
		e := mock.ExpectExec(`INSERT INTO agent_error_log`).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				captureArg{got: &codes[i]}, sqlmock.AnyArg(), captureArg{got: &captured[i]},
			)
		if execErr != nil {
			e.WillReturnError(execErr)
		} else {
			e.WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "11111111-1111-1111-1111-111111111111",
			StepName:        "council_decide",
		},
	}

	recordTruncationDegradation(context.Background(), params, "corr-1", "revise", damage, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}

	out := make([]map[string]interface{}, 0, len(captured))
	for i, c := range captured {
		if codes[i] != "TRUNCATION_DEGRADED_REVIEW" {
			t.Errorf("row %d: error_code = %q, want TRUNCATION_DEGRADED_REVIEW", i, codes[i])
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(c), &m); err != nil {
			t.Fatalf("row %d: context is not JSON (%v): %s", i, err, c)
		}
		out = append(out, m)
	}
	return out
}

// One row per damaged seat, each naming which branch damaged it. The branch is
// the point: "salvaged" and "lost" are different amounts of loss, and a query
// that cannot tell them apart cannot tell a nuisance from an incident.
func TestRecordTruncationDegradationWritesOneRowPerSeat(t *testing.T) {
	rows := runRecord(t, []truncationDegradation{
		{Field: "review_editquality", Reviewer: "edit-quality", Verdict: "object", Objections: 2, Branch: "producer_marker"},
		{Field: "review_guardian", Reviewer: "guardian", Verdict: "approve", Objections: 0, Branch: "salvaged_from_invalid_json"},
		{Field: "review_bug_historian", Branch: "unsalvageable_invalid_json"},
	}, nil)

	if len(rows) != 3 {
		t.Fatalf("wrote %d rows, want 3", len(rows))
	}
	wantBranch := []string{"producer_marker", "salvaged_from_invalid_json", "unsalvageable_invalid_json"}
	wantField := []string{"review_editquality", "review_guardian", "review_bug_historian"}
	for i, r := range rows {
		if got := r["branch"]; got != wantBranch[i] {
			t.Errorf("row %d: branch = %v, want %s", i, got, wantBranch[i])
		}
		if got := r["review_field"]; got != wantField[i] {
			t.Errorf("row %d: review_field = %v, want %s", i, got, wantField[i])
		}
		if got := r["council_decision"]; got != "revise" {
			t.Errorf("row %d: council_decision = %v, want revise", i, got)
		}
		if got := r["correlation_id"]; got != "corr-1" {
			t.Errorf("row %d: correlation_id = %v, want corr-1", i, got)
		}
	}
}

// The overwhelmingly common case is a clean round. It must cost nothing —
// no row, no query, no allocation of an empty batch.
func TestRecordTruncationDegradationWritesNothingOnACleanRound(t *testing.T) {
	if rows := runRecord(t, nil, nil); len(rows) != 0 {
		t.Fatalf("wrote %d rows for an undamaged council, want 0", len(rows))
	}
}

// Best-effort is the whole contract: this runs AFTER the council report is
// durable, so a failing insert must not panic and must not propagate. If this
// test ever needs changing to let an error escape, the caller is wrong, not this.
func TestRecordTruncationDegradationSurvivesAFailingInsert(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked on a failing insert: %v", r)
		}
	}()
	runRecord(t, []truncationDegradation{
		{Field: "review_guardian", Branch: "producer_marker"},
	}, errors.New("connection reset"))
}

// A nil DB is reachable in tests and in any future caller that has not wired one;
// the guard must come before the loop, not inside it.
func TestRecordTruncationDegradationHandlesNilDB(t *testing.T) {
	params := ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{},
	}
	recordTruncationDegradation(context.Background(), params, "corr", "approved",
		[]truncationDegradation{{Field: "review_guardian", Branch: "producer_marker"}}, zap.NewNop())
}

// The message a human reads in agent_error_log has to say WHICH seat and HOW,
// because that row is the whole reason this residual was worth closing.
func TestRecordTruncationDegradationMessageNamesSeatAndBranch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	var msg string
	mock.ExpectExec(`INSERT INTO agent_error_log`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), captureArg{got: &msg},
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{},
	}
	recordTruncationDegradation(context.Background(), params, "c", "revise",
		[]truncationDegradation{{Field: "review_editquality", Branch: "producer_marker"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
	if !strings.Contains(msg, "review_editquality") || !strings.Contains(msg, "producer_marker") {
		t.Errorf("error_message does not name the seat and the branch: %q", msg)
	}
}
