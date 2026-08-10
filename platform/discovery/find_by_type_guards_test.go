// FILE: platform/discovery/find_by_type_guards_test.go
package discovery

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// FindByType is the ONE query standing between a dispatched message and the
// agent it named, and until bugs_open/239 it carried neither of the two guards
// every sibling loader has. Both halves are pinned here:
//
//  1. the predicate, because a missing guard is invisible — the query still
//     returns a row, just the wrong one (an active snapshot is stored at
//     version+1000, so it OUTRANKS the live definition under ORDER BY version
//     DESC, and a soft-deleted row is still is_active);
//  2. the not-found sentinel, because the caller now has to tell "this agent
//     does not exist" (terminal: refuse) from "the lookup faulted" (transient:
//     re-attempt), and it used to answer both with a bare fmt.Errorf.

func definitionRows(agentType string, workflow []byte) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "type", "display_name", "description", "category",
		"default_config", "workflow", "capabilities", "briefing_questionnaire",
		"usage_count", "version", "is_snapshot",
	}).AddRow(
		"def-1", agentType, agentType, "", "",
		[]byte(`{}`), workflow, []byte(`[]`), nil,
		0, 1, false,
	)
}

// TestFindByType_QueryCarriesTheSiblingGuards asserts the predicate itself.
// sqlmock matches the expectation as a regexp against the SQL actually issued,
// so an absent clause fails the test rather than silently changing which row
// production picks.
func TestFindByType_QueryCarriesTheSiblingGuards(t *testing.T) {
	for _, tc := range []struct {
		name       string
		minVersion int
	}{
		{"unversioned", 0},
		{"version-floored", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery("deleted_at IS NULL").
				WillReturnRows(definitionRows("page-build-handler", []byte(`{"start_step":"go"}`)))

			d := NewAgentDefinitionDiscovery(db)
			if _, err := d.FindByType(context.Background(), "page-build-handler", tc.minVersion, zap.NewNop()); err != nil {
				t.Fatalf("FindByType: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("query lacks the deleted_at guard: %v", err)
			}
		})
	}
}

func TestFindByType_QueryExcludesSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name       string
		minVersion int
	}{
		{"unversioned", 0},
		{"version-floored", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery("is_snapshot = false").
				WillReturnRows(definitionRows("page-build-handler", []byte(`{"start_step":"go"}`)))

			d := NewAgentDefinitionDiscovery(db)
			if _, err := d.FindByType(context.Background(), "page-build-handler", tc.minVersion, zap.NewNop()); err != nil {
				t.Fatalf("FindByType: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("query lacks the snapshot guard — an active snapshot at version+1000 would win: %v", err)
			}
		})
	}
}

// TestFindByType_NoRowsIsTheSentinel — the terminal arm.
func TestFindByType_NoRowsIsTheSentinel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM agent_definitions").WillReturnError(sql.ErrNoRows)

	d := NewAgentDefinitionDiscovery(db)
	_, err = d.FindByType(context.Background(), "no-such-agent", 0, zap.NewNop())
	if err == nil {
		t.Fatal("a missing definition returned no error")
	}
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false — the caller cannot tell a miss from a fault", err)
	}
	// The type must survive into the message: it is what the FAILED
	// orchestration row's owner column and the operator's grep both key on.
	if got := err.Error(); got == "" || !stderrors.Is(err, ErrAgentDefinitionNotFound) {
		t.Fatalf("sentinel not wrapped: %v", got)
	}
}

// TestFindByType_DriverFaultIsNotTheSentinel — the transient arm. This is the
// distinction bugs_open/239 turned on: a bad connection used to be answered
// exactly like an unknown agent, and the message was silently run as the
// consuming pod's own no-op instead of being re-attempted.
func TestFindByType_DriverFaultIsNotTheSentinel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM agent_definitions").WillReturnError(fmt.Errorf("driver: bad connection"))

	d := NewAgentDefinitionDiscovery(db)
	_, err = d.FindByType(context.Background(), "page-build-handler", 0, zap.NewNop())
	if err == nil {
		t.Fatal("a driver fault returned no error")
	}
	if IsNotFound(err) {
		t.Fatal("a transient driver fault classified as 'no such agent' — the message would be refused for ever")
	}
}

// TestFindByType_NullWorkflowIsNotAScanError pins the fix for the trap found
// while testing the refusal path: `default_config->'workflow'` is SQL NULL for
// any definition without that key, and scanning NULL into a json.RawMessage
// errors. That made a permanent config fault ("this agent has no workflow")
// arrive as a lookup error, i.e. as something worth retrying. The row must come
// back cleanly with an empty Workflow so the caller can refuse it terminally.
func TestFindByType_NullWorkflowIsNotAScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM agent_definitions").WillReturnRows(definitionRows("page-build-handler", nil))

	d := NewAgentDefinitionDiscovery(db)
	result, err := d.FindByType(context.Background(), "page-build-handler", 0, zap.NewNop())
	if err != nil {
		t.Fatalf("a NULL workflow column produced an error instead of an empty workflow: %v", err)
	}
	if len(result.Workflow) != 0 {
		t.Fatalf("Workflow = %q, want empty", string(result.Workflow))
	}
}
