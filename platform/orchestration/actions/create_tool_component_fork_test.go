// FILE: platform/orchestration/actions/create_tool_component_fork_test.go
//
// RFC_036 §9.3: when a LIBRARY tool (component_level='tool', forked_from IS
// NULL, is_active) already claims the generated tool's function, the new
// site-specific row must be born as a FORK of it — forked_from = the library
// row's id — because "this row is forkable by other sites" and "this row
// blocks any rebuild of that tool" are the same partial-index predicate
// (idx_cc_tool_function_unique): a bare INSERT dies on SQLSTATE 23505 at
// save_tool while the work item reports complete. The section-level sibling
// of this collision is bugs_open/311 (CLC-020, resolved differently on
// measured grounds — RFC_036 §11).
//
// Mutation-proof construction: each test pins the INSERT's forked_from
// argument via WithArgs, so deleting the lookup wiring flips $10 from the
// library id to nil (or vice versa) and the argument match fails. The
// INSERT returns an error deliberately — the walk stops there because
// everything after it (page creation, nav, linking) is the adopt tests'
// territory; the arg assertion has already happened by the time the error
// returns.

package actions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const forkTestLibraryID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

// forkPreamble walks to just before the library-claim lookup.
func forkPreamble(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
}

// insertArgsWithFork pins name ($2), function ($4) and forked_from ($10);
// everything else is AnyArg. fork may be a string id or nil.
func expectComponentInsert(mock sqlmock.Sqlmock, fork interface{}) {
	mock.ExpectExec("INSERT INTO content_components").
		WithArgs(sqlmock.AnyArg(), "tool-aspect-ratio-webdesign-co-uk", sqlmock.AnyArg(),
			"tool-aspect-ratio", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), fork).
		WillReturnError(errors.New("stop-here: fork test ends at the INSERT"))
}

// A library claim forces the fork: forked_from = the library row's id.
func TestCreateToolComponent_LibraryClaimForcesFork(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	forkPreamble(mock)
	mock.ExpectQuery(`forked_from IS NULL`).WithArgs("tool-aspect-ratio").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(forkTestLibraryID))
	expectComponentInsert(mock, forkTestLibraryID)

	_, err = CreateToolComponentAction(context.Background(), adoptTestParams(db, nil))
	if err == nil || !strings.Contains(err.Error(), "stop-here") {
		t.Fatalf("walk must end at the INSERT with the deliberate error; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations (forked_from must carry the library id): %v", err)
	}
}

// No library claim: forked_from stays NULL — the pre-§9.3 write, unchanged.
func TestCreateToolComponent_NoLibraryClaimStaysBase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	forkPreamble(mock)
	mock.ExpectQuery(`forked_from IS NULL`).WithArgs("tool-aspect-ratio").
		WillReturnError(sql.ErrNoRows)
	expectComponentInsert(mock, nil)

	_, err = CreateToolComponentAction(context.Background(), adoptTestParams(db, nil))
	if err == nil || !strings.Contains(err.Error(), "stop-here") {
		t.Fatalf("walk must end at the INSERT with the deliberate error; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations (forked_from must be NULL): %v", err)
	}
}

// A lookup READ error fails towards today's behaviour — no fork, INSERT
// proceeds, and a real collision is refused loudly by the unique index
// rather than guessed at. (The Warn branch, pinned so it stays fail-open.)
func TestCreateToolComponent_LookupErrorFailsOpenToBase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	forkPreamble(mock)
	mock.ExpectQuery(`forked_from IS NULL`).WithArgs("tool-aspect-ratio").
		WillReturnError(errors.New("transient census failure"))
	expectComponentInsert(mock, nil)

	_, err = CreateToolComponentAction(context.Background(), adoptTestParams(db, nil))
	if err == nil || !strings.Contains(err.Error(), "stop-here") {
		t.Fatalf("walk must end at the INSERT with the deliberate error; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations (fail-open must not fork): %v", err)
	}
}
