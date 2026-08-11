// FILE: platform/messaging/processor_pool_ownership_test.go
//
// bugs_open/246. NewMessageProcessor used to call SetMaxOpenConns/SetMaxIdleConns/
// SetConnMaxLifetime on the *sql.DB it was PASSED. A *sql.DB is a pool object, so
// that re-sized the caller's pool rather than creating its own, silently discarding
// the value agentbase had just read from CHASSIS_DB_MAX_OPEN_CONNS.
//
// These tests call the REAL constructor on purpose. Every other test in this package
// builds &MessageProcessor{...} as a struct literal, which is convenient and which
// structurally CANNOT catch a defect that lives in the constructor — the whole class
// of bug was invisible to the existing suite.
package messaging

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// openUnconnectedPool builds a pool object without contacting a database.
// sql.Open only parses the DSN and constructs the pool; it does not dial.
func openUnconnectedPool(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", "postgres://u:p@127.0.0.1:1/none")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// newProcessorWithPool constructs a processor the way agentbase does, with the
// collaborators it does not dereference left nil. Verified against the constructor:
// producer, orchestrator, validator and initializer are only stored, and
// orchestration.NewStateRepository stores its arguments without touching them.
func newProcessorWithPool(t *testing.T, db *sql.DB) {
	t.Helper()
	// DATABASE_URL must be unset for this case: it is NOT set on chassis pods, and
	// a fixture that sets it exercises a shape production does not have.
	t.Setenv("DATABASE_URL", "")
	_ = NewMessageProcessor(
		"test-agent", "test-id", "test-role",
		db,
		nil, // producer
		nil, // orchestrator
		nil, // validator
		zap.NewNop(),
		nil, // initializer
	)
}

// TestConstructorDoesNotResizeTheCallersPool is the regression proper.
//
// The sub-tests are a matched pair on purpose. Asserting only that 12 survives
// would also pass against an assertion that can never fail; the second case sets a
// different value so a probe that always reports the expected number is caught.
func TestConstructorDoesNotResizeTheCallersPool(t *testing.T) {
	for _, tc := range []struct {
		name string
		size int
	}{
		{"operator configured 12, as the live chassis does", 12},
		{"control: a different value must also survive", 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openUnconnectedPool(t)

			// What agentbase does before constructing the processor.
			db.SetMaxOpenConns(tc.size)
			if got := db.Stats().MaxOpenConnections; got != tc.size {
				t.Fatalf("precondition: pool did not accept its size: got %d, want %d", got, tc.size)
			}

			newProcessorWithPool(t, db)

			if got := db.Stats().MaxOpenConnections; got != tc.size {
				t.Errorf("NewMessageProcessor re-sized the pool it was handed: got %d, want %d.\n"+
					"The pool belongs to whoever opened it (agentbase); this constructor must not "+
					"call SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime on the `db` PARAMETER. "+
					"See bugs_open/246.", got, tc.size)
			}
		})
	}
}

// TestConstructorSizesThePoolItOpensItself is the other half of the same rule.
//
// sqlDB is opened INSIDE the constructor, so sizing it is the constructor's job.
// It was previously left at Go's zero value, which means UNLIMITED — an unbounded
// client pool behind a transaction-mode pgbouncer with max_client_conn = 200.
func TestConstructorSizesThePoolItOpensItself(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@127.0.0.1:1/none")

	handed := openUnconnectedPool(t)
	handed.SetMaxOpenConns(12)

	p := NewMessageProcessor(
		"test-agent", "test-id", "test-role",
		handed,
		nil, nil, nil,
		zap.NewNop(),
		nil,
	)

	if p.sqlDB == nil {
		t.Fatal("DATABASE_URL was set to a parseable DSN but sqlDB was not opened")
	}
	if got := p.sqlDB.Stats().MaxOpenConnections; got <= 0 {
		t.Errorf("the pool this constructor opens is unbounded: MaxOpenConnections = %d. "+
			"Go's zero value is 0 = unlimited; a pool we open must be sized. See bugs_open/246.", got)
	}

	// And the handed-in pool is still untouched even on this branch.
	if got := handed.Stats().MaxOpenConnections; got != 12 {
		t.Errorf("the caller's pool was re-sized while opening our own: got %d, want 12", got)
	}
}
