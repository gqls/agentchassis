// Pins bugs_open/360's race-free half: every page_components UPDATE the
// section editor issues must carry the tombstone predicate
// (pageComponentNotRemovedSQL), because all three persist branches promote
// build_status to 'approved' — an UPDATE reaching a 'removed' row would
// UN-RETIRE it, and the next rerender then publishes the retired content
// again. Measured 2026-08-21: four literal_markdown transform edits
// resurrected four retired ported slots on webdesign.co.uk and the pages
// publicly served two stacked tools for ~19 h.
//
// The captured-SQL assertion style follows
// section_editor_html_only_persist_test.go: sqlmock reports whatever row
// count the test configures, so this pins the STATEMENT the code sends, not
// the database's evaluation of it — the predicate text in the WHERE clause IS
// the contract. The advisory tombstone gate in ApplySectionEditAction is
// messaging; these WHERE clauses are the enforcement, which is why the pin is
// here. Mutation-proved at authoring time: deleting the predicate from any
// one statement fails that statement's case (the three cases are independent,
// so a guard surviving in a SIBLING statement cannot mask the mutation).
//
// The final case pins the refusal SURFACE: zero rows affected must come back
// as errComponentLocked (the skip-convertible sentinel), never as success —
// otherwise a tombstone-refused write would report the edit as applied.

package actions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestSectionEditorWritesCarryTombstonePredicate(t *testing.T) {
	// Hardcoded on purpose — asserting against the const would let a vacuous
	// edit to pageComponentNotRemovedSQL pass its own test.
	const predicate = "COALESCE(build_status, 'pending') <> 'removed'"

	cases := []struct {
		name string
		run  func(db *sql.DB) error
	}{
		{"after_edit_html_only", func(db *sql.DB) error {
			return updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<p>x</p>", nil)
		}},
		{"after_edit_with_content_data", func(db *sql.DB) error {
			return updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<p>x</p>",
				map[string]interface{}{"body": "x"})
		}},
		{"component_swap", func(db *sql.DB) error {
			return updatePageComponentSwap(context.Background(), db, uuid.New(), uuid.New(),
				"slot", "<p>x</p>", map[string]interface{}{"body": "x"})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var queries []string
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			// No ExpectBegin: stampedExecContext's BeginTx fails and it falls
			// back to the plain single ExecContext, same as the html-only
			// persist tests rely on — exactly one captured statement.
			mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

			if err := tc.run(db); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if len(queries) != 1 {
				t.Fatalf("expected exactly 1 statement, got %d", len(queries))
			}
			stmt := queries[0]
			whereAt := strings.Index(stmt, "WHERE")
			if whereAt < 0 {
				t.Fatalf("statement has no WHERE clause: %s", stmt)
			}
			predAt := strings.Index(stmt, predicate)
			if predAt < 0 {
				t.Errorf("statement is missing the tombstone predicate — a 'removed' row is writable and an automated edit can resurrect it (bugs_open/360): %s", stmt)
			} else if predAt < whereAt {
				t.Errorf("tombstone predicate appears BEFORE the WHERE clause (comment, not enforcement): %s", stmt)
			}
		})
	}
}

func TestSectionEditorWriteZeroRowsSurfacesAsSkipSentinel(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

	err = updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<p>x</p>", nil)
	if !errors.Is(err, errComponentLocked) {
		t.Fatalf("a write refused by the WHERE predicates must surface as errComponentLocked (the skip-convertible sentinel), got: %v", err)
	}
}
