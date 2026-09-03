// Tests for direction 2 of features_open/035 P1 — the ancestor recompose that
// runs at the tail of apply_section_edit.
//
// WHY THESE AND NOT A HAPPY PATH. This code was written on 2026-08-31, reviewed
// over three council rounds, committed, and then sat with NO CALLER and NO TEST
// until 2026-09-03 — long enough for the linker to eliminate it and for a binary
// probe to read its absence as a missing commit (editorial_design_uplift/
// HANDOFF_2026-09-02 §9). So the cases below are the ones that discriminate:
// what the write REFUSES, what a refusal is REPORTED as, and that the wiring
// exists at all.
//
// Mutation-proved at authoring time, each against the case aimed at it:
//
//	M1  drop `AND `+datahelpers.NotRemovedSQL     → predicate case fails (tombstone)
//	M2  drop `AND `+pageComponentAgentWritableSQL → predicate case fails (lock)
//	M3  `return true, nil` instead of `n > 0, nil` → refusal case fails
//	M4  delete the call from ApplySectionEditAction → wiring case fails
//
// Two more added 2026-09-03 after council cab931b1 came back APPROVED with
// advisories, both run and both killed:
//
//	M5  `return true, nil` in the rowsErr branch   → fail-closed case fails
//	M6  AgentWritableSQLFor → "(true)"             → predicate-parity case fails
package actions

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The predicates are hardcoded rather than referenced from their constants, for
// the reason section_editor_tombstone_guard_test.go gives: asserting against the
// constant would let a vacuous edit to the constant pass its own test.
const (
	recomposeTombstonePredicate = "build_status IS DISTINCT FROM 'removed'"
	recomposeLockPredicate      = "locked_at IS NULL"
)

// A recompose reaches rows nobody chose — the ANCESTORS of the row an operator
// or agent actually named. ApplySectionEditAction checks the lock and the
// tombstone on its own target and on nothing above it, so these two clauses in
// this statement are the only thing standing between a child edit and a rewrite
// of a human-locked parent.
//
// Note what did NOT catch the original unguarded form: TestNoHandSpelledTombstonePredicate
// finds a WRONG spelling of the predicate, never a MISSING one, and the writer
// coverage test asks about the floors, not about locks. Absence has no detector,
// which is why it needs a pin.
func TestRecomposedAncestorWriteCarriesTheTombstoneAndLockPredicates(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// No ExpectBegin, so stampedExecContext's BeginTx fails and it falls back to
	// the plain single ExecContext — exactly one captured statement, the same
	// arrangement the section editor's own guard tests rely on.
	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := writeRecomposedAncestor(context.Background(), db, uuid.New(), "<article>x</article>"); err != nil {
		t.Fatalf("writeRecomposedAncestor: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected exactly 1 statement, got %d: %v", len(queries), queries)
	}

	stmt := queries[0]
	whereAt := strings.Index(stmt, "WHERE")
	if whereAt < 0 {
		t.Fatalf("statement has no WHERE clause: %s", stmt)
	}
	for _, p := range []struct{ name, needle, why string }{
		{"tombstone", recomposeTombstonePredicate,
			"a recomposed 'removed' ancestor is un-retired and republished (bugs_open/360)"},
		{"lock", recomposeLockPredicate,
			"a human-locked ancestor is rewritten by a path no human pointed at it (bugs_open/058)"},
	} {
		at := strings.Index(stmt, p.needle)
		if at < 0 {
			t.Errorf("statement is missing the %s predicate — %s:\n%s", p.name, p.why, stmt)
		} else if at < whereAt {
			// The predicate text appearing only ABOVE the WHERE is a comment, not
			// enforcement — the distinction this whole file exists to keep.
			t.Errorf("%s predicate appears BEFORE the WHERE clause (comment, not enforcement):\n%s", p.name, stmt)
		}
	}
}

// The statement SUCCEEDS whether or not its predicates admitted the row, so the
// row count is the only thing that distinguishes "recomposed" from "refused".
// Reporting a refusal as success is a green result over stale bytes — this
// feature's own defect class, arriving inside its fix.
func TestRecomposedAncestorWriteReportsARefusalRatherThanSuccess(t *testing.T) {
	cases := []struct {
		name         string
		rowsAffected int64
		want         bool
	}{
		{"locked_or_tombstoned_row_admits_nothing", 0, false},
		{"row_took_the_write", 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))

			written, err := writeRecomposedAncestor(context.Background(), db, uuid.New(), "<article>x</article>")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if written != tc.want {
				t.Errorf("rows affected %d: written=%v, want %v — a refused write reported as done leaves the page serving a parent that embeds the pre-edit child",
					tc.rowsAffected, written, tc.want)
			}
		})
	}
}

// COUNCIL cab931b1, debug_historian (medium): the first version of this function
// returned `true` when RowsAffected() itself errored, on the reasoning that the
// statement had succeeded. That re-opens the exact door the predicates and the row
// count were added to close, on the one path where the driver says it cannot tell
// us. The asymmetry decides it — a false STALE costs a log line; a false WRITTEN
// costs a page serving a parent that embeds the pre-edit child, silently and
// indefinitely. This case pins the corrected direction.
//
// Mutation: `return true, nil` in the rowsErr branch -> this case goes red.
func TestRecomposedAncestorWriteFailsClosedWhenTheDriverCannotReportRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewErrorResult(errors.New("driver cannot report rows affected")))

	written, err := writeRecomposedAncestor(context.Background(), db, uuid.New(), "<article>x</article>")
	if err != nil {
		t.Fatalf("a RowsAffected failure must not become an error: %v", err)
	}
	if written {
		t.Error("an unreportable row count was treated as WRITTEN — the ancestor may hold pre-edit bytes " +
			"and the caller will not know; fail closed and report it stale (council cab931b1, debug_historian)")
	}
}

// COUNCIL cab931b1, guardian (medium) and reuse_agent (medium): this write is a
// SECOND guarded-write style for page_components sitting beside the edit path's
// own, and pageComponentAgentWritableSQL is called with an empty alias — so the
// two questions are "does the helper degrade to always-true on an empty alias"
// and "can the two writers drift apart". One test answers both: the predicate
// text this statement carries must be BYTE-IDENTICAL to the one the sibling
// writer carries, and must not be a tautology.
//
// It is a parity test rather than an assertion against the constant, because
// asserting against the constant would let a vacuous edit to the constant pass
// its own test (section_editor_tombstone_guard_test.go's reason, applied here).
func TestAncestorAndChildWritesCarryTheIDENTICALLockPredicate(t *testing.T) {
	capture := func(run func(db *sql.DB)) string {
		var queries []string
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		defer db.Close()
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))
		run(db)
		if len(queries) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(queries))
		}
		return queries[0]
	}

	ancestor := capture(func(db *sql.DB) {
		_, _ = writeRecomposedAncestor(context.Background(), db, uuid.New(), "<article>x</article>")
	})
	child := capture(func(db *sql.DB) {
		_ = updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<p>x</p>", nil)
	})

	// The rendered predicate, taken from the code rather than retyped — if the
	// helper ever degrades, both strings degrade together and the tautology check
	// below is what catches it.
	pred := pageComponentAgentWritableSQL("")

	// A tautology would satisfy every "is the predicate present" test while
	// guarding nothing. Two independent tells: it must name the lock column, and
	// it must not be one of the shapes that admits every row.
	if !strings.Contains(pred, "locked_at") {
		t.Fatalf("the agent-writable predicate does not mention locked_at, so it cannot be enforcing a lock: %q", pred)
	}
	for _, tautology := range []string{"true", "TRUE", "1=1", "1 = 1"} {
		if strings.TrimSpace(pred) == tautology || strings.TrimSpace(pred) == "("+tautology+")" {
			t.Fatalf("pageComponentAgentWritableSQL(\"\") rendered a tautology %q — every write is admitted "+
				"and both writers' lock guards are inert", pred)
		}
	}

	if !strings.Contains(ancestor, pred) {
		t.Errorf("the ancestor write does not carry the shared lock predicate:\n%s", ancestor)
	}
	if !strings.Contains(child, pred) {
		t.Fatalf("the CHILD write no longer carries it either — this test has gone blind rather than the "+
			"ancestor having drifted:\n%s", child)
	}
}

// The write must announce itself, or page_component_history attributes the
// ancestor rewrite to the connection's socket like every pre-A1 row did
// (bugs_open/355). This is the one case that needs the transaction to SUCCEED,
// because that is the only path on which the stamp is issued at all.
func TestRecomposedAncestorWriteAnnouncesItsWriter(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(".*").WithArgs(contentWriterRecomposeAncestors).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	written, err := writeRecomposedAncestor(context.Background(), db, uuid.New(), "<article>x</article>")
	if err != nil {
		t.Fatalf("writeRecomposedAncestor: %v", err)
	}
	if !written {
		t.Fatal("expected the write to be reported as taken")
	}
	if len(queries) != 2 {
		t.Fatalf("expected the stamp and the write, got %d statements: %v", len(queries), queries)
	}
	if !strings.Contains(queries[0], "set_config") || !strings.Contains(queries[0], "application_name") {
		t.Errorf("first statement is not the writer stamp — the archive row will carry the socket identity (bugs_open/355 A1): %s", queries[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The cost claim, pinned rather than asserted in prose. On today's population
// (0 of 2,249 page_components carry a parent_instance_id, 2026-08-31) every edit
// takes this path, so "one indexed SELECT and nothing else" is the whole of what
// this feature charges the live edit path — and RFC_022's "unsafe side OFF by
// default" is exactly this: inert until a row opts in.
func TestRecomposeAncestorsIsInertOnATopLevelRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	child := uuid.New()
	mock.ExpectQuery("parent_instance_id").WithArgs(child).
		WillReturnRows(sqlmock.NewRows([]string{"parent_instance_id"}).AddRow(nil))

	stale := recomposeAncestors(context.Background(), ActionParams{DB: db}, db, child, uuid.New(), zap.NewNop())
	if len(stale) != 0 {
		t.Errorf("a top-level row has no ancestors, so nothing can be stale; got %v", stale)
	}
	// ExpectationsWereMet is the load-bearing half: sqlmock fails on any
	// statement beyond the ones expected, so this asserts that NOTHING else ran —
	// no render, no write, no second read.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or exceeded expectations: %v", err)
	}
}

// The wiring itself. A source scan is a weak instrument and the weakness is
// stated: it proves the call is WRITTEN, not that it EXECUTES, and the
// behavioural half is the inert-row case above.
//
// Comments are stripped first, and that is not a detail — this file, and the
// call site's own comment block, both name the function repeatedly. Without the
// strip the scan would pass on prose, which is the vacuous-pass trap that makes
// source-scanning tests worthless (LANDMINES: "a source-scan test makes your
// COMMENTS load-bearing").
func TestApplySectionEditWiresTheAncestorRecompose(t *testing.T) {
	src, err := os.ReadFile("section_editor_actions.go")
	if err != nil {
		t.Fatalf("read section_editor_actions.go: %v", err)
	}
	code := stripGoComments(string(src))

	if !strings.Contains(code, "recomposeAncestors(ctx, params, params.DB") {
		t.Fatal("ApplySectionEditAction does not call recomposeAncestors — a child edit then leaves every ancestor " +
			"embedding the pre-edit bytes, which is what the page actually serves (features_open/035 P1 direction 2). " +
			"The function existed uncalled from 2026-08-31 to 2026-09-03; this test is why it cannot happen again.")
	}

	// Order is the other half of the contract: the recompose must precede the
	// reassembly, or assemblePage reads the ancestors back BEFORE they are
	// refreshed and returns the stale parent as the page to deploy.
	recomposeAt := strings.Index(code, "recomposeAncestors(ctx, params, params.DB")
	assembleAt := strings.Index(code, "assemblePage(ctx, params.DB")
	if assembleAt < 0 {
		t.Fatal("could not find the assemblePage call — this test's ordering assertion has gone blind, " +
			"which is worse than it failing")
	}
	if recomposeAt > assembleAt {
		t.Error("the ancestor recompose runs AFTER assemblePage, so the reassembled HTML carries the stale parent")
	}

	// The result key, because a stale ancestor that is only logged is gone within
	// minutes and cannot be queried out of collected_data later.
	if !strings.Contains(code, `"stale_ancestor_slots"`) {
		t.Error("the action does not publish stale_ancestor_slots — an ancestor it could not refresh " +
			"would exist only in a pod log that rotates")
	}
}
