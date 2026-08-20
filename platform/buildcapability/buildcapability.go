// FILE: platform/buildcapability/buildcapability.go
//
// What a running binary can do, written somewhere a human or a script can read
// it. RFC_040 (owner-ratified 2026-08-20, SCOPED TO THIS HALF ONLY).
//
// ── WHY THIS EXISTS ──────────────────────────────────────────────────────────
//
// Go changes are inert until an image is rebuilt and rolled; DB config is live
// immediately. So every change whose config half names a behaviour in its code
// half has a window where the config is ahead of the binary, and the question
// "is the new code actually live yet?" has to be answerable BEFORE the config
// lands. Measured 2026-08-19 over docs/agent_docs/sql_for_agents: 32 migrations
// assert a binary precondition in prose; ZERO can verify one.
//
// The estate's documented answer — read the `build provenance` startup line,
// then `git merge-base --is-ancestor <commit> <stamp>` — frequently cannot be
// performed at all:
//
//   - it is a STARTUP line, so it scrolls. Measured: gone from a FULL
//     `kubectl logs` three hours after a roll.
//   - buildinfo.GitCommit is ONE string, not an ancestry, so grepping the
//     binary for your own commit returns ABSENT for a binary that certainly
//     contains it. Two lanes have now been burned by exactly this
//     (bugs_open/215 on v1.0.1288; bugs_open/299 on v1.0.1316).
//
// What works instead is asking for the CAPABILITY rather than the commit: is
// THIS check name, THIS config key, in the artefact that is running? That
// question has no shelf life — the answer is in the binary for as long as the
// binary runs. This package moves that answer out of the pod, where only
// `kubectl exec` can reach it, and into the one place every migration, CronJob
// and council seat already reads.
//
// ── WHAT THIS PACKAGE DELIBERATELY DOES NOT DO ───────────────────────────────
//
// It does not assert. There is no assert_live_capability() and no migration
// calling one — the owner scoped RFC_040 to the recording half on 2026-08-20,
// and the reason is not timidity: a fail-closed assertion helper with exactly
// ONE caller is a mechanism nobody exercises, which is this estate's own
// documented failure mode (it is why the 2026-07-29 ruling declined to REQUIRE
// default-OFF switches). Make the fact durable first; let the second real
// demand shape the assertion. A future author adding the helper should be able
// to name two migrations that want it.
//
// It also knows nothing about what a capability IS. Callers pass the lists in.
// That is what keeps this package importable from anywhere: `actions` imports
// `discovery_checks`, so a package that reached into either would inherit that
// direction and could never be imported by both.
package buildcapability

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/pkg/buildinfo"
)

// KindProvenance is the sentinel kind every Record call writes, whatever else
// it is given.
//
// It exists so that "this pod reported in" is a QUERYABLE fact, distinct from
// "this pod has no capabilities of the kind you asked about". Without it, a
// service that registers no discovery checks is indistinguishable from a
// service that never wrote at all — and those two must never look alike,
// because the whole value of this table is that an absence means something.
// RetentionWindow is how long a pod's rows survive without a Touch.
//
// It MUST be comfortably longer than any caller's Touch interval, or a live
// long-lived pod would have its own rows pruned while it is still running — the
// exact staleness-in-reverse this table exists to avoid. Two hours against the
// 15-minute TouchInterval below leaves eight missed heartbeats of slack.
const RetentionWindow = "2 hours"

// TouchInterval is the cadence a long-lived caller should call Touch at.
// Ephemeral per-job pods do NOT need it: they die well inside RetentionWindow
// and are supposed to age out.
const TouchInterval = 15 * time.Minute

const KindProvenance = "build"

// NameProvenance is the sentinel row's name. Together with KindProvenance it
// gives one row per pod carrying the commit, always.
const NameProvenance = "provenance"

// Set is one enumerable registry the binary carries.
//
// Kind is the vocabulary ("discovery_check", "action"); Names is whatever the
// registry's own listing function returned. Both are the caller's words — this
// package does not validate them against anything, because there is nothing
// authoritative to validate against that would not re-create the import
// direction described in the package comment.
type Set struct {
	Kind  string
	Names []string
}

// Record replaces this pod's rows with the capabilities it currently carries.
//
// DELETE-then-INSERT for this pod only, in one transaction: a pod's capability
// list is a snapshot of one binary, so a partial overlap between an old and a
// new roll would be a lie about both. Scoping the delete to (service, pod_name)
// is what keeps one service's restart from disturbing another's rows — and pod
// names are unique per roll, so the previous pod's rows age out via
// last_seen_at rather than being deleted by a successor that is not them.
//
// The caller is expected to treat a returned error as NON-FATAL. A capability
// registry that can stop a service starting is a worse bargain than the problem
// it solves, so this returns the error and lets the caller log and continue;
// it never panics and never blocks.
func Record(ctx context.Context, db *sql.DB, service, podName string, sets ...Set) error {
	if db == nil {
		return fmt.Errorf("buildcapability: nil db handle")
	}
	if service == "" || podName == "" {
		// Refuse rather than write a row keyed on "". An unattributable row is
		// worse than no row: it would satisfy a presence check for every
		// service at once.
		return fmt.Errorf("buildcapability: service and podName are both required (got service=%q pod=%q)", service, podName)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("buildcapability: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM service_binary_capabilities WHERE service = $1 AND pod_name = $2`,
		service, podName); err != nil {
		return fmt.Errorf("buildcapability: clear previous rows: %w", err)
	}

	// The sentinel first, so the commit is recorded even if every set below is
	// empty (see KindProvenance).
	rows := []Set{{Kind: KindProvenance, Names: []string{NameProvenance}}}
	rows = append(rows, sets...)

	for _, s := range rows {
		if s.Kind == "" {
			continue // a nameless vocabulary cannot be queried; skip rather than invent one
		}
		for _, name := range s.Names {
			if name == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO service_binary_capabilities
				     (service, pod_name, git_commit, kind, name)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (service, pod_name, kind, name)
				 DO UPDATE SET last_seen_at = now(), git_commit = EXCLUDED.git_commit`,
				service, podName, buildinfo.GitCommit, s.Kind, name); err != nil {
				return fmt.Errorf("buildcapability: insert %s/%s: %w", s.Kind, name, err)
			}
		}
	}

	// PRUNE, in the same transaction. Added 2026-08-20 after the first live run
	// exposed the defect this mechanism had in its own design.
	//
	// The chassis binary does NOT only run in long-lived Deployments. It also runs
	// as EPHEMERAL PER-JOB pods (agent-page-rerender-*, agent-page-build-handler-*,
	// agent-build-dispatch-loop-*, agent-site-publisher-* …), each of which starts,
	// writes ~400 rows, and dies. Measured in the first 3h40m live: 75,827 rows
	// across 191 distinct pods — 24 MB — while only 82 pods existed at that moment,
	// i.e. 109 of the reporting pods were ALREADY DEAD with their rows frozen. That
	// extrapolates to roughly half a million rows and ~160 MB per day, unbounded.
	//
	// So a retention window is not a nicety here, it is the difference between a
	// registry and a leak. Rows are pruned on the SAME PATH that creates them, which
	// is what makes it self-limiting without a scheduler: every pod start pays a
	// tiny, bounded cost to clear what the dead ones left. The window must exceed
	// the Touch interval of any long-lived caller (see Touch) or a live pod's rows
	// would be pruned out from under it — that ordering is the load-bearing part.
	//
	// A failed prune is deliberately IGNORED rather than returned: the registry
	// being slightly too large is strictly better than it being absent, and
	// returning here would roll back the write we just made.
	_, _ = tx.ExecContext(ctx,
		`DELETE FROM service_binary_capabilities WHERE last_seen_at < now() - $1::interval`,
		RetentionWindow)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("buildcapability: commit: %w", err)
	}
	return nil
}

// Touch refreshes last_seen_at for this pod's rows.
//
// ⚠ THIS IS THE PART THAT KEEPS THE TABLE HONEST, and it is the caller's job to
// call it periodically. Without a refresh, a dead pod's rows sit there looking
// current, and a future reader would conclude a binary is running that is not —
// which is the SAME class of error this whole mechanism exists to end,
// reintroduced by the mechanism itself. Any reader must therefore filter on
// last_seen_at, and any service that records must also touch.
func Touch(ctx context.Context, db *sql.DB, service, podName string) error {
	if db == nil {
		return fmt.Errorf("buildcapability: nil db handle")
	}
	_, err := db.ExecContext(ctx,
		`UPDATE service_binary_capabilities SET last_seen_at = now()
		  WHERE service = $1 AND pod_name = $2`, service, podName)
	if err != nil {
		return fmt.Errorf("buildcapability: touch: %w", err)
	}
	return nil
}
