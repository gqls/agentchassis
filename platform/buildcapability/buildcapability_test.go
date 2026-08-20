// FILE: platform/buildcapability/buildcapability_test.go
//
// These tests are written against the failure modes RFC_040 names, not just the
// happy path. The ones that earn their place are the NEGATIVES: a table whose
// absence is meant to be meaningful is only trustworthy if it refuses to write
// rows that would make an absence ambiguous.
package buildcapability

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestRecordWritesProvenanceSentinelEvenWithNoSets(t *testing.T) {
	// The load-bearing case. A service that registers no discovery checks must
	// still be distinguishable from a service that never wrote at all —
	// otherwise "no rows" means both "nothing registered" and "never reported",
	// and every future reader of this table is reading an ambiguity.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1", sqlmock.AnyArg(), KindProvenance, NameProvenance).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := Record(context.Background(), db, "agent-chassis", "pod-1"); err != nil {
		t.Fatalf("Record with no sets: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the sentinel row was not written: %v", err)
	}
}

func TestRecordWritesEveryNameOfEverySet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// sentinel, then the two sets in order
	for _, pair := range [][2]string{
		{KindProvenance, NameProvenance},
		{"discovery_check", "cta_nonpage_destination"},
		{"discovery_check", "misdirected_cta"},
		{"action", "resolve_internal_links"},
	} {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO service_binary_capabilities")).
			WithArgs("agent-chassis", "pod-1", sqlmock.AnyArg(), pair[0], pair[1]).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err = Record(context.Background(), db, "agent-chassis", "pod-1",
		Set{Kind: "discovery_check", Names: []string{"cta_nonpage_destination", "misdirected_cta"}},
		Set{Kind: "action", Names: []string{"resolve_internal_links"}},
	)
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not every capability was written: %v", err)
	}
}

func TestRecordRefusesAnUnattributableRow(t *testing.T) {
	// An empty service or pod would key a row on "", which would satisfy a
	// presence check for EVERY service at once — a false positive that looks
	// exactly like the thing this table is for. Refusing is the whole point,
	// and it must refuse BEFORE touching the database.
	for _, tc := range []struct{ name, service, pod string }{
		{"no service", "", "pod-1"},
		{"no pod", "agent-chassis", ""},
		{"neither", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			// No expectations set at all: any statement reaching the DB fails.

			err = Record(context.Background(), db, tc.service, tc.pod, Set{Kind: "action", Names: []string{"x"}})
			if err == nil {
				t.Fatal("expected a refusal, got nil — an unattributable row would have been written")
			}
			if !strings.Contains(err.Error(), "required") {
				t.Errorf("error should say what was missing, got: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("it must refuse before touching the DB: %v", err)
			}
		})
	}
}

func TestRecordRollsBackWhenAnInsertFails(t *testing.T) {
	// A half-written capability list is worse than none: it would report a
	// binary as carrying some of what it carries, and a reader cannot tell a
	// truncated list from a short one.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1", sqlmock.AnyArg(), KindProvenance, NameProvenance).
		WillReturnError(errBoom{})
	mock.ExpectRollback()

	if err := Record(context.Background(), db, "agent-chassis", "pod-1"); err == nil {
		t.Fatal("expected the insert failure to surface")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected a rollback: %v", err)
	}
}

func TestRecordSkipsEmptyKindsAndNames(t *testing.T) {
	// A row with an empty kind or name is unqueryable — it can never be asked
	// for by a reader, so writing it only adds noise that looks like data.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1", sqlmock.AnyArg(), KindProvenance, NameProvenance).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1", sqlmock.AnyArg(), "action", "real").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = Record(context.Background(), db, "agent-chassis", "pod-1",
		Set{Kind: "", Names: []string{"orphan"}},             // no kind → skipped entirely
		Set{Kind: "action", Names: []string{"", "real", ""}}, // empty names → skipped
	)
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unqueryable rows were written: %v", err)
	}
}

func TestRecordRefusesNilDB(t *testing.T) {
	if err := Record(context.Background(), nil, "agent-chassis", "pod-1"); err == nil {
		t.Fatal("expected an error for a nil db handle")
	}
	if err := Touch(context.Background(), nil, "agent-chassis", "pod-1"); err == nil {
		t.Fatal("expected an error for a nil db handle")
	}
}

func TestTouchRefreshesOnlyThisPod(t *testing.T) {
	// Scoped to (service, pod): one pod's heartbeat must never make another
	// pod's stale rows look current — that is the staleness hole this
	// mechanism must not reintroduce into itself.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE service_binary_capabilities SET last_seen_at")).
		WithArgs("agent-chassis", "pod-1").
		WillReturnResult(sqlmock.NewResult(0, 3))

	if err := Touch(context.Background(), db, "agent-chassis", "pod-1"); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Touch did not scope to this pod: %v", err)
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }

func TestRecordPrunesStaleRowsInTheSameTransaction(t *testing.T) {
	// The prune is what stops this table being a leak: the chassis binary also
	// runs as EPHEMERAL per-job pods, and measured over its first 3h40m live it
	// wrote 75,827 rows across 191 pods (24 MB) of which 109 pods were already
	// dead. Without this DELETE the table grows without bound.
	//
	// It needs its own assertion because Record deliberately IGNORES the prune's
	// error (losing the write would be worse than an oversized table) — so a
	// prune that silently stopped running would leave every other test green.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_binary_capabilities WHERE service")).
		WithArgs("agent-chassis", "pod-1").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO service_binary_capabilities")).
		WithArgs("agent-chassis", "pod-1", sqlmock.AnyArg(), KindProvenance, NameProvenance).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// the retention prune, keyed on last_seen_at and carrying the window
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_binary_capabilities WHERE last_seen_at")).
		WithArgs(RetentionWindow).WillReturnResult(sqlmock.NewResult(0, 42))
	mock.ExpectCommit()

	if err := Record(context.Background(), db, "agent-chassis", "pod-1"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the retention prune did not run: %v", err)
	}
}

func TestRetentionWindowExceedsTouchInterval(t *testing.T) {
	// The ordering is load-bearing: if a live pod's heartbeat were slower than
	// the retention window, the prune would sweep the rows of a pod that is
	// still serving — staleness in reverse, and just as misleading as the stale
	// rows this window exists to remove.
	d, err := time.ParseDuration("2h")
	if err != nil {
		t.Fatal(err)
	}
	if RetentionWindow != "2 hours" {
		t.Fatalf("RetentionWindow changed to %q — re-check this invariant by hand", RetentionWindow)
	}
	if TouchInterval >= d {
		t.Fatalf("TouchInterval (%s) must stay comfortably under RetentionWindow (%s), or live pods get pruned", TouchInterval, d)
	}
	if d/TouchInterval < 4 {
		t.Errorf("only %d heartbeats fit in the retention window — too little slack for a missed beat", d/TouchInterval)
	}
}
