# 048 — a scheduled task whose pre-query finds nothing holds its concurrency group's only slot forever, starving every other task in that group

**Filed 2026-07-20** by the bugfix-006 thread, found while answering the open question in
`bugs_open/006` §C (where does the claim-timeout sweep live). **Status: OPEN.**
**Severity: silent, long-running.** Nothing errors, nothing alerts, `enabled` stays `true`, and the
admin pipeline view shows the tasks as healthy. Four maintenance tasks have simply not run since
**2026-05-02 — 79 days.**

**Not a duplicate of `029`.** 029 is hung *orchestrations* occupying real in-flight slots in the
`dispatch` group; there the task fires and is legitimately refused a slot, and `last_triggered_at`
keeps advancing. Here the task is *selected and runs*, finds nothing to do, and the slot it took is
never given back — because the accounting happens before the work does. Same family
(concurrency-group starvation), different mechanism, different fix.

## Observed

```sql
SELECT name, interval_seconds, enabled, last_triggered_at, now()-last_triggered_at AS since_last
FROM scheduled_tasks WHERE enabled = true ORDER BY last_triggered_at NULLS FIRST;
```
```
 feasibility-recheck        |   600 | t | 2026-05-02 04:17:06 | 79 days 15:19:59
 database-cleanup           |  3600 | t | 2026-05-02 04:17:38 | 79 days 15:19:27
 stale-work-item-reaper     |  3600 | t | 2026-05-02 04:18:06 | 79 days 15:18:59
 work-item-archiver         | 86400 | t | 2026-05-02 12:52:37 | 79 days 06:44:28
 thunder-reaper             |   900 | t | 2026-06-19 08:12:34 | 31 days 11:24:31
 ... 8 further tasks, all firing on schedule (claimed-item-timeout last fired <1s before this query)
```

The first four all share **`concurrency_group='maintenance'`, `max_concurrent=1`**, and three of them
stopped within **60 seconds of each other**. The eight healthy tasks are spread across other groups.
So this is not "the scheduler is down" — it is per-group, and it is total.

## Mechanism

`cmd/scheduler/main.go`. Three things combine:

**1. The slot is taken before the work is attempted.** `runTick` increments the in-memory group
counter *and then* runs the pre-query:

```go
// main.go:171-184  — slot claimed here
if task.ConcurrencyGroup.Valid {
    if current >= task.MaxConcurrent { continue }
    inFlight[group] = current + 1          // <-- taken
}
// main.go:188-199  — work decided here
if task.PreQuery.Valid && ... {
    dynamicData, err := runPreQuery(...)
    if dynamicData == nil {
        logger.Info("Pre-query returned no rows, skipping task", ...)
        continue                            // <-- slot never released, timestamp never written
    }
}
```

**2. "Nothing to do" is the normal steady state for a reaper.** `feasibility-recheck`'s pre-query
promotes `blocked` work items back to `triaged` and ends `HAVING COUNT(*) > 0`, so it returns **zero
rows whenever there is nothing blocked**. There is currently nothing blocked, and there has not been
for 79 days:
```sql
SELECT count(*) FROM site_work_items WHERE status='blocked';  -- 0
```
So every tick, forever, it takes the `maintenance` slot and immediately bails.

**3. The failure is self-perpetuating, because of the sort order.** `loadDueTasks`
(`main.go:270`) orders `last_triggered_at ASC NULLS FIRST`. A task that bails on the no-rows path
never updates `last_triggered_at` — so it stays permanently at the **head** of its group and wins the
slot again on the very next tick. It cannot lose its position by failing; failing is what keeps it
there.

Net effect per tick: `feasibility-recheck` takes the only `maintenance` slot and does nothing;
`database-cleanup`, `stale-work-item-reaper` and `work-item-archiver` are then skipped with
`"Skipping task — concurrency group at max"`. Repeat every tick for 79 days.

`thunder-reaper` is the same bug in a group of one (`thunder-lifecycle`): its pre-query is
`... WHERE status='running' AND running_since < ... LIMIT 1`, which returns no rows when no Thunder
instance is over its cap. It starves only itself, but it also never records that it ran — so its
`last_triggered_at` is not a usable liveness signal either.

### Hypothesis that was tested and REFUTED

First guess was a leaked slot in `countInFlight` — i.e. a task stuck "in flight" in the database
forever. That is **wrong**: `countInFlight` (`main.go:296-307`) already guards with
`last_triggered_at + timeout_seconds > NOW()`, so a DB-level slot self-heals. The leak is not in the
DB accounting at all; it is the **in-memory** `inFlight[group]` increment within a single tick,
which is never rolled back on the early-`continue` paths. Recording this because the DB-side
accounting looks correct and will absorb an investigation that starts there.

## Consequence (measured, not inferred)

`stale-work-item-reaper` marks build-pipeline items `unresolved` after 48h triaged. It has a real
backlog it should have cleared:
```sql
SELECT count(*) FROM site_work_items
WHERE status='triaged' AND pipeline='build'
  AND created_at < now() - interval '48 hours' AND claimed_at IS NULL;   -- 11
```
`database-cleanup` and `work-item-archiver` have likewise not run in 79 days; their backlogs are
**[UNMEASURED]** — whoever takes this should size them before deciding urgency, since an archiver
that has not run since May may be the more expensive one.

## Fix candidates

1. **Release the slot on every early exit** (smallest, most direct). Decrement `inFlight[group]` on
   the no-rows path, the pre-query-error path, and the merge-error path — or, cleaner, don't take the
   slot until the task is actually about to fire. A task that did no work did not consume capacity,
   and the counter should say so.
2. **Stamp `last_triggered_at` even on a no-op tick** (or add a separate `last_evaluated_at`). This
   breaks the self-perpetuation independently of (1): a task that ran and found nothing has *run*,
   and should go to the back of the queue like everything else. It also makes
   `last_triggered_at` mean "we looked", which is what every operator reading that column assumes.
   Doing (2) alone fixes the starvation; doing (1) alone leaves the head-of-queue pin in place for
   any *other* early-exit path.
3. **Alert on the invariant.** `enabled=true AND last_triggered_at < now() - (interval_seconds * N)`
   is a trivially checkable liveness condition and would have caught this in May. Related to `044`
   (capability exists, nothing routes to it) — same "silent dormancy is undetectable" theme.

Recommend **(1) + (2) together**, then (3) as the check that stops the class recurring.

## How to verify a fix

Before: `SELECT name, last_triggered_at FROM scheduled_tasks WHERE concurrency_group='maintenance';`
— three of four stuck in May. After a scheduler image roll, all four should advance within one
`interval_seconds` of each other, and `blocked_items` staying 0 must **not** stop them. Confirm the
11-item `stale-work-item-reaper` backlog drains.

Note the scheduler is its own binary (`cmd/scheduler`), so this needs a scheduler build+roll, not a
chassis one — verify against the running pod, not the tag.

## Where

- `cmd/scheduler/main.go:171-184` (slot taken), `:188-199` (early exits), `:270` (sort order),
  `:296-307` (DB accounting — correct, not the bug).
- `scheduled_tasks` rows: `feasibility-recheck`, `thunder-reaper` (the two heads-of-group);
  `database-cleanup`, `stale-work-item-reaper`, `work-item-archiver` (the starved).
- Related: `bugs_open/029` (same family, hung spawns), `bugs_open/044` (silent dormancy undetectable),
  `bugs_open/006` §C (where this was found).
