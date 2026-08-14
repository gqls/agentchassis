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
// producer, orchestrator, validator and initializer are only stored.
func newProcessorWithPool(t *testing.T, db *sql.DB, databaseURL string) {
	t.Helper()
	// DATABASE_URL is NOT set on chassis pods. Since bugs_open/259 the constructor
	// does not read it at all, which is why the table below drives BOTH values: the
	// caller's pool must survive whatever this variable says.
	t.Setenv("DATABASE_URL", databaseURL)
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
		name        string
		size        int
		databaseURL string
	}{
		{"operator configured 12, as the live chassis does", 12, ""},
		{"control: a different value must also survive", 9, ""},
		// The production shape is DATABASE_URL unset, so the two cases above are
		// the ones that matter. This third case is the regression guard left by
		// bugs_open/259: the constructor used to open a SECOND pool from this
		// variable, and the sibling test that covered its sizing went with it. A
		// re-added opener that reaches for the caller's pool is caught here.
		{"DATABASE_URL set: the constructor must still open and touch nothing", 12, "postgres://u:p@127.0.0.1:1/none"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openUnconnectedPool(t)

			// What agentbase does before constructing the processor.
			db.SetMaxOpenConns(tc.size)
			if got := db.Stats().MaxOpenConnections; got != tc.size {
				t.Fatalf("precondition: pool did not accept its size: got %d, want %d", got, tc.size)
			}

			newProcessorWithPool(t, db, tc.databaseURL)

			if got := db.Stats().MaxOpenConnections; got != tc.size {
				t.Errorf("NewMessageProcessor re-sized the pool it was handed: got %d, want %d.\n"+
					"The pool belongs to whoever opened it (agentbase); this constructor must not "+
					"call SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime on the `db` PARAMETER. "+
					"See bugs_open/246.", got, tc.size)
			}
		})
	}
}

// TestConstructorSizesThePoolItOpensItself was HERE, and bugs_open/259 removed it
// along with its subject. It asserted that the second handle the constructor opened
// from DATABASE_URL was opened and was sized (Go's zero value for MaxOpenConns is
// 0 = UNLIMITED, so an unsized pool behind a transaction-mode pgbouncer was the
// hazard it guarded). There is no longer a pool for this constructor to open, so
// the test could only have been kept by keeping the defect.
//
// This is a genuine coverage REMOVAL, not a replacement — recorded here rather than
// letting the suite quietly shrink. The half that could be salvaged is the
// DATABASE_URL-set case in the table above, which still asserts the caller's pool
// survives; nothing now watches the sizing of a pool this file no longer opens.
