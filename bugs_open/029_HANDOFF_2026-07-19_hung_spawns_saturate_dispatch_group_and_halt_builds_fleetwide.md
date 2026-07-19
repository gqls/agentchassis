# 029 — hung spawns saturate the `dispatch` concurrency group and silently halt ALL builds fleet-wide

**Filed 2026-07-19** (relojistas thread). **Status: OPEN.** Fleet-wide outage class. Nothing
errors, nothing alerts, no site reports a failure — builds simply stop happening
**everywhere**, and the scheduler keeps firing every 30 seconds into a full pool.

This is the *consequence* half of `bugs_open/003` (spawn lost child response). 003 explains
why an individual orchestration hangs. 029 is what those hangs do to the platform once a few
accumulate: they take the whole build pipeline down and leave no trace that says so.

## Observed

Between 12:52 and 13:28 on 2026-07-19, twelve orchestrations entered `AWAITING_RESPONSES`
waiting on a spawned child that never returned, and stayed there:

| agent | count | stuck at |
|---|---|---|
| `diagnose-orchestrator` | 5 | `spawn_diagnoser` / `call_diagnoser` |
| `build-dispatch-loop` | 4 | `process_item_iter_N_spawn_handler` |
| `build-pipeline-trigger` | 2 | `spawn_dispatch` |
| (unnamed) | 1 | `call_planner` |

**From 13:25:09 onward, not one `build-pipeline-trigger` or `build-dispatch-loop`
orchestration was created** — despite `scheduled_tasks.last_triggered_at` advancing every 30
seconds as designed (13:42:17 at time of diagnosis). Meanwhile `council-gate`,
`endpoint-health-checker` and other groups continued normally, which is what makes this so
easy to misread as "my site is stuck" rather than "the build pipeline is down".

Note the cross-session character: the `diagnose-orchestrator` hangs belong to a different
session entirely. **Any thread's hung spawn stalls every other thread's builds.**

## Mechanism

`scheduled_tasks` row `build-pipeline-trigger`: `interval_seconds=30`,
`concurrency_group='dispatch'`, **`max_concurrent=8`**.

Its `pre_query` correctly finds work:
```sql
SELECT COUNT(*)::text AS pending_sites FROM sites s
WHERE s.locked_at IS NULL
  AND EXISTS (SELECT 1 FROM site_work_items wi
              WHERE wi.site_id = s.id AND wi.status='triaged'
                AND wi.pipeline='build' AND wi.attempt_count < wi.max_attempts)
HAVING COUNT(*) > 0;
```
So the trigger fires, sees pending work, and is then refused a slot because the `dispatch`
group is full of orchestrations that will never complete. Hung orchestrations are counted as
running forever: **nothing ages them out, times them out, or reaps them.**

The pool is 8. Six dead build-* orchestrations plus live traffic is enough to close it.

## Proof (the recovery is the proof)

Cancelling only the six hung `build-dispatch-loop` / `build-pipeline-trigger` orchestrations
older than 15 minutes:

```sql
UPDATE orchestration_states SET status='CANCELLED'
 WHERE status='AWAITING_RESPONSES' AND updated_at < now() - interval '15 minutes'
   AND initial_request_data->'config'->>'agent_type'
       IN ('build-dispatch-loop','build-pipeline-trigger');
-- UPDATE 6
```

A new `build-pipeline-trigger` orchestration appeared within one scheduler tick
(13:47:51, `EXECUTING_STEP / spawn_dispatch`) after 22 minutes of complete silence. The
other session's five `diagnose-orchestrator` hangs were deliberately left in place — they
belong to that thread to decide about, and clearing the build-* six alone was sufficient.

## Why it is dangerous out of proportion to its cause

- **Silent.** No failed work item, no error status, no alert. `site_work_items` rows sit at
  `triaged` looking perfectly healthy — correct `pipeline`, `approval_mode='auto'`,
  `attempt_count=0` — which is exactly what a *dispatchable* item looks like.
- **Fleet-wide from a single site's problem.** The group is global. One thread hammering one
  site can stop builds for every site.
- **Self-inflicted by a reasonable action.** I caused four of the six by dispatching
  `build-dispatch-loop` repeatedly to hurry a slow batch. That is an obvious thing to try and
  there is nothing warning against it (now recorded in the relojistas runbook).
- **Diagnosis leads you everywhere except here.** I checked pod health, pod age, the kcat
  vanish trap, Kafka consumer lag, `sites.locked_at`, and the work item's dispatch fields —
  all fine — before thinking to look at the scheduler's concurrency group.

## Fix candidates

1. **Reap stale orchestrations.** Any `AWAITING_RESPONSES` older than a threshold (15–30 min)
   is dead by definition — no child is coming. Mark it `FAILED`/`CANCELLED` and free the
   slot. Smallest change, addresses the class rather than one agent, and is the one that
   would have prevented this outage outright.
2. **Don't count stale orchestrations toward `max_concurrent`.** Same effect without mutating
   state — apply an age cut-off in whatever query computes group occupancy.
3. **Time out the await itself.** The real fix for `003`: a spawn that gets no response
   within N seconds should fail its parent rather than wait forever. Most correct, most work.
4. **Alert on a saturated group.** If a scheduled task's `pre_query` returns work but it has
   been refused a slot for N consecutive ticks, that is an outage and should say so. Does not
   fix anything, but converts a silent halt into a visible one.

Recommend 1 (or 2) immediately as the cheap guard, with 3 as the proper fix for `003`.

## How to verify a fix

Deliberately hang enough spawns to fill the group, then confirm that
`build-pipeline-trigger` orchestrations resume being created without manual intervention.
Check `orchestration_states` creation timestamps, not the scheduler's `last_triggered_at` —
the scheduler goes on ticking happily throughout, which is precisely the trap.

## Diagnostic query worth keeping

```sql
-- "is the build pipeline actually running, or just being triggered?"
SELECT initial_request_data->'config'->>'agent_type' AS agent,
       count(*), max(created_at) AS newest
  FROM orchestration_states
 WHERE created_at > now() - interval '30 minutes'
 GROUP BY 1 ORDER BY 2 DESC;
-- plus the giveaway:
SELECT initial_request_data->'config'->>'agent_type' AS agent, count(*)
  FROM orchestration_states
 WHERE status='AWAITING_RESPONSES' AND updated_at < now() - interval '15 minutes'
 GROUP BY 1;
```
If the second returns rows and the first shows a build agent whose `newest` is old, this is
the bug.

## Related

- `bugs_open/003` — spawn lost child response. The cause; this is the blast radius.
- `bugs_open/028` — filed the same day from the same batch: a build no-op reporting
  `complete` and deploying. Both are silent-failure defects in the build path.
