# SUMMARY — scheduler concurrency-group starvation (bugs_open/048), 2026-07-21

**What we're trying to do.** Get four dormant "maintenance" scheduled jobs running
again, and fix the scheduler bug that stopped them, so the class can't recur silently.

**Where we've come from.** The bug was found and diagnosed on 2026-07-20 by another
thread (the bugfix-006 investigation) as a side-effect of a different question, and
filed as `bugs_open/048`. Symptom: four jobs in the `maintenance` concurrency group
had not run since 2 May — 79 days — with nothing erroring and `enabled` still true.

**What we've done.** Verified the diagnosis against live code and DB, then applied the
handoff's recommended two-part fix to `cmd/scheduler/main.go`: (1) the shared
concurrency slot is no longer claimed until a job commits to real work, so a job that
finds nothing to do leaves the slot free; and (2) a job that runs and finds nothing
now records that it ran, so it rotates to the back of the queue instead of pinning
itself at the head forever. Built the scheduler (its own binary) at v1.0.1146 and
rolled it.

**Where we are now.** Fixed AND live, verified on the running pod: with the
discriminating condition present (nothing blocked), the previously-jammed head job
now records its no-op runs and cycles on its interval; all four maintenance jobs have
left May; every one of the 14 enabled scheduled jobs is healthy; and the one small
backlog that had accrued cleared. `bugs_open/048` is closed and moved to
`/bugs_closed/`. The fix commit is in mainline history, so any future image carries it.

**Where we're going.** Nothing further on this bug. Two deliberate non-goals left as
follow-ups, both belonging elsewhere: a liveness alert for silently-dormant scheduled
tasks (tracked with `bugs_open/044`, "capability exists, nothing routes to it"), and a
regression test, which would require extracting the scheduler's decision logic from
its DB/Kafka I/O — a refactor of a fleet-critical binary that we chose not to bundle
with the fix.
