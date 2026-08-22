// FILE: cmd/config-key-audit/livedeclarations.go
//
// `--live-declaration-drift` — the half of bugs_open/363 that no Go test can hold.
//
// PHASE 1 (shipped 2026-08-22, `873575ecf`) moved the guards off append-only
// migration files and onto `platform/livespec`, so they compare **Go against the
// declaration**. That closed the half where a guard asserted something the
// checksum rule had already made impossible. It left the other half wide open:
// nothing compared the declaration against the **live database object**, so a
// migration could move production out from under livespec with no tell at all.
//
// WHY THIS CANNOT BE A GO TEST, and why it is not a pre-commit hook either:
//
//   - `go test` runs inside a `git archive HEAD` build context with no cluster.
//     A unit test that needs Postgres is not a unit test, and would be skipped
//     exactly where it matters.
//   - A pre-commit hook cannot gate live config, because AT COMMIT TIME THE
//     MIGRATION IS UNAPPLIED. The owner ruled on precisely this (RFC_006,
//     2026-08-02) and the answer there was a marker in code plus a DAILY JOB
//     against the real system. This is that job.
//
// So the check is after the fact, on a clock, by design — and its detection
// window (~24h) is stated in the bug rather than hidden.
//
// READ-ONLY. Every probe is a single SELECT, asserted as such by
// livespec_test.go. This binary already owns the direct-Postgres route because
// the service account has no pods/exec RBAC in this namespace, so a kubectl-based
// checker fails in a way that looks CLEAN.
//
// ⚠ WHAT IT STILL CANNOT SEE. A declaration compares live TEXT to required
// fragments. The live `claimed-item-timeout` `pre_query` carries a COMMENT that
// states a contract superseded on 2026-08-19 and names a deleted test — and this
// checker passes it, because the clause matches and it is the prose that lies.
// That is recorded in bugs_open/363 and in the LANDMINES entry, not papered over.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/livespec"
)

// emitLiveDeclarationDrift compares every livespec Declaration to its live object.
//
// EXIT CODES, and the middle one is the point:
//
//	0 — every declaration matched, and we can say how many we looked at
//	1 — at least one drift finding
//	2 — the check could not run: no database, an unreadable probe, a NULL or
//	    missing row, or an empty registry. NEVER a clean report.
//
// A checker that reports success when it could not look is the failure this
// estate keeps re-learning (most recently the render audit's blind pass). So
// every "cannot tell" path below exits 2 and says why on stderr.
func emitLiveDeclarationDrift(args []string) {
	report := false
	for _, a := range args {
		switch a {
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --live-declaration-drift: unknown flag %q\n", a)
			os.Exit(2)
		}
	}

	// An empty registry would otherwise sail through the loop and print a
	// triumphant "0 findings over 0 objects".
	if len(livespec.Declarations) == 0 {
		fmt.Fprintln(os.Stderr, "config-key-audit --live-declaration-drift: livespec holds ZERO declarations — "+
			"refusing to report clean over an empty set")
		os.Exit(2)
	}

	db, err := dbConn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --live-declaration-drift: %v\n", err)
		os.Exit(2)
	}
	if db == nil {
		// dbConn treats "no PG_CLIENTS_HOST" as the terminal case, not an error.
		// For an auditor whose whole job is to read production, it is an error.
		fmt.Fprintln(os.Stderr, "config-key-audit --live-declaration-drift: no database configured "+
			"(PG_CLIENTS_HOST unset) — this check exists to read the LIVE objects, so it cannot report clean")
		os.Exit(2)
	}
	defer db.Close()

	findings, probed, kinds := compareAllDeclarations(db)

	scope := describeScope(probed, kinds)
	for _, f := range findings {
		fmt.Println(f)
	}
	fmt.Printf("live-declaration-drift: %s; %d finding(s)\n", scope, len(findings))

	if report {
		// ONE row per run, INCLUDING clean runs, so a MISSING row reads as "the job
		// did not run" rather than "nothing is wrong". The body carries the SCOPE,
		// because a finding count alone cannot distinguish "looked at 6" from
		// "looked at 0".
		body := fmt.Sprintf("live-declaration-drift: %s; %d finding(s).", scope, len(findings))
		if len(findings) > 0 {
			body += "\n" + strings.Join(findings, "\n")
		}
		body += "\n\nPhase 1 (bugs_open/363) moved the Go guards onto platform/livespec; this is the phase-2 " +
			"tie to the live objects. A finding means the live object and its declaration have parted: fix " +
			"whichever is wrong, in the same commit as the migration that moved it."
		writeDocNote("live-declaration-drift", body, "live-declaration-drift", "cmd/config-key-audit")
	}

	if len(findings) > 0 {
		os.Exit(1)
	}
}

// compareAllDeclarations probes every declaration and returns the findings, the
// number of objects actually READ, and the count per kind.
//
// It exits 2 rather than returning on any "could not look" condition: a partial
// sweep that returns findings would report a smaller number than the truth and
// look like an improvement.
func compareAllDeclarations(db *sql.DB) (findings []string, probed int, kinds map[string]int) {
	kinds = map[string]int{}
	for _, d := range livespec.Declarations {
		var live sql.NullString
		switch err := db.QueryRow(d.ProbeSQL).Scan(&live); {
		case err == sql.ErrNoRows:
			fmt.Fprintf(os.Stderr, "config-key-audit --live-declaration-drift: %s: probe returned NO ROWS.\n"+
				"The live object named by this declaration does not exist — that is a bigger finding than drift, "+
				"and it must not be reported as clean.\n  probe: %s\n", d.Key, d.ProbeSQL)
			os.Exit(2)
		case err != nil:
			fmt.Fprintf(os.Stderr, "config-key-audit --live-declaration-drift: %s: probe failed: %v\n  probe: %s\n",
				d.Key, err, d.ProbeSQL)
			os.Exit(2)
		}
		if !live.Valid {
			fmt.Fprintf(os.Stderr, "config-key-audit --live-declaration-drift: %s: probe returned NULL.\n"+
				"A NULL is not an empty match — it means the column or object is not what this declaration "+
				"thinks it is.\n  probe: %s\n", d.Key, d.ProbeSQL)
			os.Exit(2)
		}

		probed++
		kinds[d.Kind]++

		switch d.Mode {
		case livespec.CountEqual:
			findings = append(findings, d.CompareCount(live.String)...)
		default:
			findings = append(findings, d.CompareFragments(live.String)...)
		}
	}
	return findings, probed, kinds
}

// describeScope renders what was actually looked at, e.g.
// "probed 5 live objects (2 scheduled_task, 2 trigger_fn, 1 trigger_bindings)".
func describeScope(probed int, kinds map[string]int) string {
	parts := make([]string, 0, len(kinds))
	for k := range kinds {
		parts = append(parts, k)
	}
	sort.Strings(parts)
	for i, k := range parts {
		parts[i] = fmt.Sprintf("%d %s", kinds[k], k)
	}
	if len(parts) == 0 {
		return "probed 0 live objects"
	}
	return fmt.Sprintf("probed %d live object(s) (%s)", probed, strings.Join(parts, ", "))
}
