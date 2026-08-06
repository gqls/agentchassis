// FILE: platform/orchestration/agenterrors/agenterrors_test.go
//
// The writer's contract, pinned: byte-compatible 13-argument INSERT (every
// sqlmock expectation in the estate that pinned that arity must keep passing),
// defaulting, and — the RFC_012 half — that failures are COUNTED, not
// swallowed: recorded-but-lost must never read as recorded.
package agenterrors

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestWrite_CanonicalThirteenArgShape(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			"site", "domain", "item", "orch",
			"agent", "id", "pod", "step", "action",
			"msg", "CODE", "warning", `{"k":"v"}`,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok := Write(context.Background(), db, zap.NewNop(), Entry{
		SiteID: "site", Domain: "domain", WorkItemID: "item", OrchestrationID: "orch",
		AgentType: "agent", AgentID: "id", PodName: "pod", StepName: "step", Action: "action",
		ErrorMessage: "msg", ErrorCode: "CODE", Severity: "warning",
		Context: map[string]interface{}{"k": "v"},
	})
	if !ok {
		t.Fatal("Write returned false on a successful insert")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestWrite_DefaultsSeverityAndClassifiesCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// A nil Context marshals to JSON "null", NOT "{}" — json.Marshal(nil map)
	// returns []byte("null"), so the historical writer's contextJSON==nil guard
	// never fired for a nil map. Byte-compatibility means pinning the real
	// behaviour, not the intended one.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "TIMEOUT", "error", "null",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	Write(context.Background(), db, zap.NewNop(), Entry{
		AgentType:    "agent",
		ErrorMessage: "context deadline exceeded while waiting",
		// Severity, ErrorCode, Context all unset.
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("defaults not applied as expected: %v", err)
	}
}

func TestWrite_FailureReturnsFalseNotSilence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnError(errors.New("connection reset"))

	if Write(context.Background(), db, zap.NewNop(), Entry{AgentType: "a", ErrorMessage: "m"}) {
		t.Fatal("Write returned true on a failed insert — recorded-but-lost must not read as recorded")
	}
}

func TestWrite_NilDBIsFalse(t *testing.T) {
	if Write(context.Background(), nil, zap.NewNop(), Entry{}) {
		t.Fatal("nil DB must report false, not pretend the row landed")
	}
}

// TestRecordFindings_CountsLossesInsteadOfSwallowingThem is the RFC_012
// contract: (attempted, recorded) diverge exactly when a row is lost, so the
// caller can put the difference in its audit.
func TestRecordFindings_CountsLossesInsteadOfSwallowingThem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"first refused", "X_REFUSED", "warning", sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnError(errors.New("disk full"))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"the audit", "X_AUDIT", "info", sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	base := Entry{SiteID: "s", Domain: "d", AgentType: "agent", Action: "x"}
	attempted, recorded := RecordFindings(context.Background(), db, zap.NewNop(), base, []Finding{
		{ErrorCode: "X_REFUSED", Severity: "warning", Message: "first refused"},
		{ErrorCode: "X_REFUSED", Severity: "warning", Message: "second refused"},
		{ErrorCode: "X_AUDIT", Severity: "info", Message: "the audit"},
	})
	if attempted != 3 || recorded != 2 {
		t.Fatalf("got (attempted=%d, recorded=%d), want (3, 2)", attempted, recorded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestRecordFindings_BaseIdentityIsNotMutated: the base entry is copied per
// finding — a helper that mutated it would leak one finding's code into the
// next caller's row.
func TestRecordFindings_BaseIdentityIsNotMutated(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	base := Entry{AgentType: "agent", ErrorCode: "", Severity: ""}
	RecordFindings(context.Background(), db, zap.NewNop(), base, []Finding{
		{ErrorCode: "A", Severity: "info", Message: "m"},
	})
	if base.ErrorCode != "" || base.Severity != "" {
		t.Fatalf("base entry was mutated: %+v", base)
	}
}

func TestClassifyError_SpotChecks(t *testing.T) {
	cases := map[string]string{
		"context deadline exceeded":     "TIMEOUT",
		"connection refused by peer":    "CONNECTION_ERROR",
		"failed to parse response body": "PARSE_ERROR",
		"something else entirely":       "UNKNOWN",
	}
	for msg, want := range cases {
		if got := ClassifyError(msg); got != want {
			t.Errorf("ClassifyError(%q) = %q, want %q", msg, got, want)
		}
	}
}
