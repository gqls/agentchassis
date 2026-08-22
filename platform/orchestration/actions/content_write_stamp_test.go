// FILE: platform/orchestration/actions/content_write_stamp_test.go
//
// bugs_open/355 A1. The load-bearing assertions, each mutation-proven at
// authoring time (the mutation and the failing test are named in the commit):
//
//	M1 delete the set_config Exec from stampedExecContext
//	   → TestStampedExecStampsInsideTheSameTransaction fails (unfulfilled
//	     expectation: the stamp), NOT by accident of ordering — sqlmock here
//	     runs with expectations IN ORDER, because "stamp somewhere nearby" is
//	     exactly the bug: it must precede the write INSIDE the tx.
//	M2 replace is_local=true with false in stampWriterSQL
//	   → same test fails on the argument mismatch (ExpectExec pins the SQL
//	     text) — the session-scoped form is the pgbouncer leak.
//	M3 make stampedExecContext return the write's error without running the
//	   write after a failed stamp → TestStampedExecStampFailureStillWrites
//	   fails (the write was never executed).
package actions

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const stampTestWrite = `UPDATE page_components SET content_data = $1 WHERE id = $2`

func TestStampedExecStampsInsideTheSameTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT set_config('application_name', $1, true)`)).
		WithArgs("action:test_writer").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(stampTestWrite)).
		WithArgs("{}", "pc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	res, err := stampedExecContext(context.Background(), db, "action:test_writer", stampTestWrite, "{}", "pc-1")
	if err != nil {
		t.Fatalf("stampedExecContext: %v", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("RowsAffected = %d, want 1 — the lock-refusal paths read this through the wrapper", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestStampedExecBeginFailureStillWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("pool exhausted"))
	// The write must still run — unstamped, on the plain connection.
	mock.ExpectExec(regexp.QuoteMeta(stampTestWrite)).
		WithArgs("{}", "pc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := stampedExecContext(context.Background(), db, "action:test_writer", stampTestWrite, "{}", "pc-1"); err != nil {
		t.Fatalf("attribution broke the write: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestStampedExecStampFailureStillWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT set_config('application_name', $1, true)`)).
		WithArgs("action:test_writer").
		WillReturnError(errors.New("connection reset"))
	mock.ExpectRollback()
	mock.ExpectExec(regexp.QuoteMeta(stampTestWrite)).
		WithArgs("{}", "pc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := stampedExecContext(context.Background(), db, "action:test_writer", stampTestWrite, "{}", "pc-1"); err != nil {
		t.Fatalf("attribution broke the write: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestStampedExecWriteErrorPassesThrough(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT set_config('application_name', $1, true)`)).
		WithArgs("action:test_writer").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(stampTestWrite)).
		WithArgs("{}", "pc-1").
		WillReturnError(errors.New("value too long"))
	mock.ExpectRollback()

	if _, err := stampedExecContext(context.Background(), db, "action:test_writer", stampTestWrite, "{}", "pc-1"); err == nil {
		t.Fatal("the write's own error must pass through, not be swallowed by the wrapper")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
