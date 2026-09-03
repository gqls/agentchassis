// FILE: cmd/config-key-audit/fleetdb.go
//
// The DIRECT-POSTGRES half of this binary, added 2026-08-08 for RFC_012 (d)'s
// standing CronJob (the owner ruling was "(d) as a STANDING check, ONLINE if the
// framework allows").
//
// WHY A SECOND ROUTE TO THE SAME DATA. Every mode in this binary reads its fleet
// export from stdin, and a session at a terminal produces that with
// `kubectl exec -i postgres-clients-0 -- psql`. A CronJob CANNOT: the
// ai-persona-app service account has no pods/exec RBAC in this namespace, a
// constraint component-render-check, single-owner-carriers-check and
// bugs-open-staleness-sweep all hit before this did. Worse, a kubectl-only tool
// fails there in a way that looks like a clean run unless you are watching exit
// codes — which is the precise failure this check exists to prevent, one level up.
//
// PG_CLIENTS_HOST being set is how the job declares itself, the same convention
// component-render-check and check_placeholder_fallbacks.py use. Unset, nothing
// here changes and the stdin path is untouched.
//
// THE TWO ROUTES SHARE ONE DECODER ON PURPOSE. loadLiveAgentsFromDB runs the SAME
// jsonb_agg query the wrapper script runs and hands the bytes to
// decodeLiveAgents, so the DB path and the stdin path cannot drift in how they
// parse a fleet, only in how they fetch one. Two hand-written decoders agreeing
// with each other is bugs_open/144, and this file is not going to add a third.
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// fleetExportQuery is the export, and it is deliberately character-for-character
// the query in scripts/audit-shared-output-fields.sh. If you change one, change
// both — a divergence here means the scheduled check and the hand-run check
// answer different questions while reporting the same name.
const fleetExportQuery = `
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';`

// dbConn returns a direct Postgres handle when PG_CLIENTS_HOST is set, and
// (nil, nil) otherwise — "no DB configured" is not an error, it is the terminal
// case. It refuses to guess a connection when the host is set but the password
// is not, because a job that silently cannot connect is the thing this file is
// here to avoid.
func dbConn() (*sql.DB, error) {
	host := os.Getenv("PG_CLIENTS_HOST")
	if host == "" {
		return nil, nil
	}
	pw := os.Getenv("CLIENTS_DB_PASSWORD")
	if pw == "" {
		return nil, fmt.Errorf("PG_CLIENTS_HOST is set but CLIENTS_DB_PASSWORD is not — " +
			"refusing to guess a connection")
	}
	port := os.Getenv("PG_CLIENTS_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("PG_CLIENTS_USER")
	if user == "" {
		user = "clients_user"
	}
	dbname := os.Getenv("PG_CLIENTS_DB")
	if dbname == "" {
		dbname = "clients_db"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pw, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

// loadLiveAgentsFromDB fetches the fleet export directly. A NULL aggregate (an
// empty fleet) is returned as an explicit error rather than as zero agents: the
// callers all refuse to print a clean report over an empty fleet, and this keeps
// that refusal on one side of the boundary instead of relying on each of them.
func loadLiveAgentsFromDB(db *sql.DB, caller string) ([]liveAgent, int, error) {
	return loadLiveAgentsFromDBWithQuery(db, caller, fleetExportQuery)
}

// loadLiveAgentsFromDBWithQuery is the same loader for a mode whose export needs
// a projection the shared query does not carry (--template-input-fields needs
// the agent-level prompt_template). The NULL guard and the decoder stay on this
// side of the boundary for every caller — a mode that hand-rolled its own
// QueryRow would be re-deciding "is an empty fleet clean?" for itself, which is
// the drift this file's header is about.
func loadLiveAgentsFromDBWithQuery(db *sql.DB, caller, query string) ([]liveAgent, int, error) {
	var raw []byte
	if err := db.QueryRow(query).Scan(&raw); err != nil {
		return nil, 0, fmt.Errorf("config-key-audit %s: fleet export query failed: %w", caller, err)
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, 0, fmt.Errorf("config-key-audit %s: fleet export returned NULL — "+
			"no live workflow definitions matched, refusing to report over an empty fleet", caller)
	}
	return decodeLiveAgents(raw, caller)
}

// writeDocNote records a run in doc_notes on EVERY run, clean or not.
//
// This is the convention check_placeholder_fallbacks.py established and
// component-render-check follows, and the reason is the whole point of putting a
// check on a schedule: a check that only speaks when it fails is
// INDISTINGUISHABLE FROM ONE THAT HAS STOPPED RUNNING. A missing row must mean
// "the job did not run", never "nothing is wrong" — bugs_open/140 is what that
// ambiguity costs.
//
// Best-effort, and deliberately so: a failure to RECORD must never become a
// failure to REPORT, so this only warns. The exit code still carries the finding.
func writeDocNote(subjectKey, body, category, source string) {
	db, err := dbConn()
	if err != nil || db == nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "(could not record run: %v)\n", err)
		}
		return
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
	                  VALUES ('pipeline', $1, $2, jsonb_build_array($3::text), $4)`,
		subjectKey, body, category, source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "(could not write doc_notes: %v)\n", err)
	}
}
