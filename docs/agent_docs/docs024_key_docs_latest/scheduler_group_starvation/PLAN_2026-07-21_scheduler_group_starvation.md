# PLAN — scheduler concurrency-group starvation (bugs_open/048)

**Started 2026-07-21.** Owner: this thread (bugfix 048). The diagnosis was already
filed by the bugfix-006 thread in `bugs_open/048_HANDOFF_2026-07-20_...md`; this
workstream is the **fix + roll + live verification**.

## The bug (verified against live code + DB, 2026-07-21)

`cmd/scheduler/main.go`. A scheduled task whose `pre_query` returns zero rows
("nothing to do") took its concurrency group's only slot and never released it,
and — because it never stamped `last_triggered_at` on that path — stayed pinned at
the head of `loadDueTasks`' `ORDER BY last_triggered_at ASC NULLS FIRST` queue and
re-won the slot on the very next tick. Result: the entire `maintenance` group
(`feasibility-recheck`, `database-cleanup`, `stale-work-item-reaper`,
`work-item-archiver`, all `max_concurrent=1`) did not run for **79 days**, silently.

Grounded facts:
- `feasibility-recheck` is the head of `maintenance` (oldest `last_triggered_at`,
  2026-05-02 04:17:06). It is `fire_message=f` (CTE-only); its pre-query
  `UPDATE site_work_items SET status='triaged' WHERE status='blocked' … RETURNING`
  returns rows only when something is blocked. `blocked` count is **0** right now,
  so it returns 0 rows every tick → no-rows `continue` → no stamp → stays head.
- `work-item-archiver` is `fire_message=t` with **no** pre_query; it is starved
  purely because the slot is always taken by the head.
- `thunder-reaper` (group `thunder-lifecycle`, group of one) is the same no-rows
  path: it never records liveness, so its `last_triggered_at` is a dead signal.

## Decision — fix candidates (1)+(2) together, as the handoff recommends

**(1) Don't claim the slot until the task commits to work.** Moved
`inFlight[group]++` from the top of the loop (before the pre-query) to just before
the fire/CTE-complete branch. A no-op, an errored pre-query, or a merge error now
leaves the slot free for its group-mates. This is cleaner than decrementing on each
early-exit `continue` (three of them) and cannot leak on a path added later.

**(2) Stamp `last_triggered_at` + `last_completed_at` on a successful no-op.** New
`stampCompleted` helper, shared by the no-op path, the CTE-only path and the fired
path (was two copy-pasted UPDATEs). A no-op task rotates to the back of the queue
like any completed run, so it cannot pin itself at the head; and `last_triggered_at`
now means "we looked", which every operator reading that column assumes.

**Why both:** (2) alone breaks the self-perpetuation for the *no-rows* path; (1)
alone leaves the head-of-queue pin for any *other* early exit and does not let
group-mates run in the same tick. Together they fully un-starve the group.

**Error paths stay bare `continue` (no stamp), deliberately.** A genuinely failing
pre-query or merge should retry next tick and stay visible via its stuck timestamp
+ warn logs. Moving the slot claim below them means they no longer starve the group,
which was the only harm.

### Not done (scoped out, noted as follow-ups)

- **Candidate (3), the liveness alert** (`enabled AND last_triggered_at < now() -
  interval_seconds*N`) is a separate monitor, related to `bugs_open/044` (capability
  exists, nothing routes to it). Not built here.
- **Unit test.** The scheduler is a single `package main` with `runTick` coupled to
  live DB + Kafka I/O and **zero** existing tests. A meaningful regression test needs
  the decision logic extracted behind injectable seams — a refactor that widens the
  blast radius on a fleet-critical binary. Deferred; verification is the live-pod
  check below. Recorded as an accepted gap.

## Build / deploy (scheduler is its own binary)

- `cmd/scheduler` → image `aqls/kafka-scheduler`. Needs a **scheduler** build+roll,
  not a chassis one.
- Built at fresh tag **v1.0.1146** (pod was on v1.0.1144; v1.0.1145 already taken by
  the bugs_open/047 release). Tag passed on the command line — the makefile
  `IMAGE_TAG` line was NOT edited (it belonged to another thread's working tree; the
  shared-line sweep trap). `make quick-scheduler-update IMAGE_TAG=v1.0.1146`.

## Verify (live, against the pod — see RUNBOOK)

All four `maintenance` tasks' `last_triggered_at` must advance from 2026-05-02 to
now within ~one tick of each other, and `blocked=0` must **not** stop them. The
`stale-work-item-reaper` backlog (1 item at fix time) must drain.
