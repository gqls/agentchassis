package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Covers persistSectionSkips (bugs_open/040 skip-not-recorded): plan_sections'
// on_missing=skip_section outcomes are merged durably into
// pages.suppressed_sections — skipped names added, names that planned ready
// removed — so the 040 partial-build guard stops counting a legitimately
// data-gated section as a build shortfall on every rebuild.
//
// sqlmock cannot evaluate the merge SQL itself (a correlated set-op subquery);
// the merge semantics were validated against the live DB (see the RUNBOOK probe
// in bugs_open/040's 2026-07-24 diagnosis block). What this test pins is the Go
// contract: which parameters are sent (ready = $3 removals, skipped = $4
// additions, JSON-encoded), that errors are warn-not-fail (no panic, no error
// escapes), and that the caller-side JSON encoding of empty slices stays "[]"
// (an accidental "null" would make jsonb_array_length($4) NULL and disarm the
// WHERE tail's no-op guard).

func TestPersistSectionSkips_SendsReadyAndSkippedAsJSON(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectExec("UPDATE pages SET suppressed_sections").
		WithArgs(siteID, "index", `["hero","features"]`, `["testimonials"]`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	persistSectionSkips(context.Background(), db, siteID, "index",
		[]string{"hero", "features"}, []string{"testimonials"}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Empty slices must encode as "[]", never "null": jsonb_array_length('null')
// is an error/NULL and would break the no-op WHERE guard. json.Marshal on a
// non-nil empty slice yields []; on a nil slice it yields null — the call site
// always allocates with make(), and this test pins the nil-safety of the
// helper itself too.
func TestPersistSectionSkips_EmptySlicesEncodeAsEmptyJSONArrays(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectExec("UPDATE pages SET suppressed_sections").
		WithArgs(siteID, "contact", `[]`, `[]`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	persistSectionSkips(context.Background(), db, siteID, "contact",
		[]string{}, []string{}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// Warn-not-fail: a DB error must not escape (persistence failure must not
// break the build — same fail-open convention as the 040 guard's own checks).
func TestPersistSectionSkips_DBErrorDoesNotPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectExec("UPDATE pages SET suppressed_sections").
		WithArgs(siteID, "index", `[]`, `["testimonials"]`).
		WillReturnError(context.DeadlineExceeded)

	// Must return normally.
	persistSectionSkips(context.Background(), db, siteID, "index",
		nil, []string{"testimonials"}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
