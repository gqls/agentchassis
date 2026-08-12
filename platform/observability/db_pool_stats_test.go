package observability

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
)

// openIdlePool returns a *sql.DB that has never connected. sql.Open is lazy, so
// no database is required — and that is the point: DBStats is pool bookkeeping,
// readable without a server, which is why this test can assert the CONFIGURED
// size offline.
func openIdlePool(t *testing.T, maxOpen int) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", "postgres://never-dialled/db")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(maxOpen)
	return db
}

func gatherNames(t *testing.T, reg *prometheus.Registry) map[string]float64 {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	out := map[string]float64{}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			switch {
			case m.GetGauge() != nil:
				out[f.GetName()] = m.GetGauge().GetValue()
			case m.GetCounter() != nil:
				out[f.GetName()] = m.GetCounter().GetValue()
			}
		}
	}
	return out
}

// The reading bugs_closed/246 wanted and could not have: the pool's CONFIGURED
// size, from the pool itself rather than inferred from an environment variable.
// The 12 here is the production value of CHASSIS_DB_MAX_OPEN_CONNS; the assert
// is disconfirmable — before this collector existed the series did not exist at
// all, and re-sizing the pool moves the number.
func TestDBPoolStats_ReportsTheConfiguredPoolSize(t *testing.T) {
	reg := prometheus.NewRegistry()
	db := openIdlePool(t, 12)

	if err := RegisterDBPoolStats(reg, db, "clients_db"); err != nil {
		t.Fatalf("RegisterDBPoolStats: %v", err)
	}

	got := gatherNames(t, reg)
	if v, ok := got["go_sql_max_open_connections"]; !ok {
		t.Fatal("go_sql_max_open_connections absent — the pool's size is still unreadable")
	} else if v != 12 {
		t.Fatalf("configured pool size: want 12, got %v", v)
	}

	// The saturation instrument itself. Zero here is correct on an idle pool;
	// what matters is that the SERIES exists, because its absence is what made
	// every pool question in 246 answerable only as "unmeasurable".
	for _, want := range []string{"go_sql_wait_count_total", "go_sql_wait_duration_seconds_total"} {
		if _, ok := got[want]; !ok {
			t.Errorf("%s absent — pool contention would still be invisible", want)
		}
	}
}

// The re-size that caused 246 is now VISIBLE: a second owner shrinking the pool
// changes the reading rather than passing unnoticed.
func TestDBPoolStats_ASecondOwnerReSizingThePoolIsVisible(t *testing.T) {
	reg := prometheus.NewRegistry()
	db := openIdlePool(t, 12)
	if err := RegisterDBPoolStats(reg, db, "clients_db"); err != nil {
		t.Fatalf("RegisterDBPoolStats: %v", err)
	}
	if got := gatherNames(t, reg)["go_sql_max_open_connections"]; got != 12 {
		t.Fatalf("precondition: want 12, got %v", got)
	}

	db.SetMaxOpenConns(4) // exactly what NewMessageProcessor used to do

	if got := gatherNames(t, reg)["go_sql_max_open_connections"]; got != 4 {
		t.Fatalf("a re-size must be observable: want 4, got %v", got)
	}
}

// Registering twice must not be an error: callers should not have to track
// whether they have already instrumented a pool.
func TestDBPoolStats_RegistrationIsIdempotent(t *testing.T) {
	reg := prometheus.NewRegistry()
	db := openIdlePool(t, 12)

	if err := RegisterDBPoolStats(reg, db, "clients_db"); err != nil {
		t.Fatalf("first registration: %v", err)
	}
	if err := RegisterDBPoolStats(reg, db, "clients_db"); err != nil {
		t.Fatalf("second registration must be tolerated, got: %v", err)
	}
}

// agentbase leaves its handle nil whenever DatabaseURL is empty, so a nil pool
// reaches this function in production shapes. It must refuse rather than hand
// the collector a typed nil that panics on the first SCRAPE — i.e. later, on
// another goroutine, far from the mistake.
func TestDBPoolStats_NilPoolIsRefusedNotDeferredToAPanic(t *testing.T) {
	reg := prometheus.NewRegistry()

	err := RegisterDBPoolStats(reg, nil, "clients_db")
	if err == nil {
		t.Fatal("a nil *sql.DB must be refused")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("error should name the cause, got: %v", err)
	}

	// And nothing was registered, so a later real registration still succeeds.
	if err := RegisterDBPoolStats(reg, openIdlePool(t, 4), "clients_db"); err != nil {
		t.Fatalf("refusal must not poison the registry: %v", err)
	}
}
