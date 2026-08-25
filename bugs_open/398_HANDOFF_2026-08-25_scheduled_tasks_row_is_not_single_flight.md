# 398 — a `scheduled_tasks` row is not a single-flight slot: the scheduler stamps BOTH timestamps at fire, so every row-level concurrency control (`max_concurrent`, `timeout_seconds`, the per-row guard) and every agent-side `notify_scheduler` stamp is inert — and the code comments, the register and a shipped throughput lever all say the opposite

**Filed 2026-08-25** by the dispatch_throughput lane (session "throughput"), found while answering
the council's round-2 objection on migrations 582–584 (corr `db9b7cbf`).
**Status: OPEN — behaviour is deliberate (commit `892a289e9`, 2026-03-17), the DOCUMENTATION of it
is the defect, and one lever was shipped on the wrong reading.** Nothing is broken at runtime;
what is broken is every claim built on the guard.

**Verification statement (owner ruling 2026-07-31):** this file asserts a structural property of
the platform and did NOT go through the 090 loop. Substituted, and stated plainly: (a) the fire
path read in full (`cmd/scheduler/main.go` `runTick`, `stampCompleted`), not a grep; (b) the live
stamps on two rows read 39 s after a fire (`last_triggered_at = last_completed_at` to the
microsecond, runs still executing); (c) a self-overlap census over 24.5 h of `orchestration_states`
(361 / 322 pairs on two rows). A 090 on the OPPOSITE claim was filed by this lane on 2026-08-19
and died at its verdict step (`max_tokens`, run `a16b82cd`, class `bugs_open/183`); a re-file would
face the same cap. Three independent artefact-level confirmations are the substitute.

## Symptom (what a reader sees)

- `loadDueTasks`' doc comment: *"returns enabled tasks whose interval has elapsed AND whose previous
  execution has completed (or timed out). This prevents the same task from spawning multiple
  concurrent pods."* — false since 2026-03-17.
- `scheduled_tasks.max_concurrent` (default 1, `build-pipeline-trigger` = 8) and `timeout_seconds`
  read as live concurrency controls. They never bind for a `fire_message=true` row.
- `build-pipeline-trigger` and `build-dispatch-loop` carry `notify_scheduler` /
  `notify_scheduler_idle` steps that `UPDATE scheduled_tasks SET last_completed_at = NOW()`. They
  read as the completion handshake. They are inert.
- WDS-002, the dispatch_throughput STARTER/PLAN/RESEARCH (2026-08-18/19) and the rationale of
  migrations 582–584 (2026-08-24) all state "one execution of the row at a time" and built a
  sibling-row concurrency lever on it. Corrected visibly 2026-08-25 in each.

## Mechanism (read, not inferred)

`cmd/scheduler/main.go` `runTick`, after `fireTrigger` publishes the Kafka message:

```go
// Update timestamps. For fire-and-forget tasks, mark completed immediately
// so the concurrency slot opens for the next tick. The message has been
// published — we don't wait for the orchestration to finish.
if err := stampCompleted(ctx, db, task.ID); err != nil {
```
```go
func stampCompleted(...) { `UPDATE scheduled_tasks SET last_triggered_at = NOW(), last_completed_at = NOW(), updated_at = NOW() WHERE id = $1` }
```

So immediately after every fire: `last_completed_at >= last_triggered_at` (the guard passes),
`last_completed_at < last_triggered_at` is false (`countInFlight` = 0), and the `timeout_seconds`
fallback is unreachable. The row is due again at `last_triggered_at + interval_seconds`, rounded up
to the next 30 s tick (`TICK_INTERVAL_SECONDS`, default 30). History: `892a289e9` 2026-03-17 *"for
fire and forget tasks dont wait for a response, send complete immediately"* wrote the both-stamps
UPDATE inline; `dc2e4b61a` 2026-07-21 (`bugs_open/048`) named it `stampCompleted` and reused it for
the no-op pre-query path, leaving the `loadDueTasks` comment as it was.

## Evidence `[MEASURED 2026-08-25 16:1x–16:3xZ]`

| what | value |
|---|---|
| live scheduler | `kafka-scheduler:v1.0.1337` (pods up 09:27Z) |
| stamps 39 s after a fire | `last_triggered_at = last_completed_at = 16:22:20.956331` on `build-pipeline-trigger`; `16:22:22.027404` on `-2`; `in_flight_per_guard = f` on both |
| trigger run duration | p50 97 s, p90 242 s (COMPLETED, 24.5 h) |
| fire cadence per row | p50 90 s, p90 91 s (interval 60 + 30 s tick) |
| **self-overlap pairs, one row** | **361** (`build-pipeline-trigger`), **322** (`-2`), min gap **0.25 s** |
| alive dispatch loops per minute | 1–8 (mode 1–2); Little's law mean 1.65 |
| enabled `fire_message=true` rows | **40** of 50 enabled (all fire-and-forget) |

## Blast radius

Every enabled `fire_message=true` task (40 as of 2026-08-25) re-fires at its interval whether or
not its previous orchestration is alive. For most that is the intended fire-and-forget. Where it
matters: any task whose author set `max_concurrent`, `timeout_seconds` or a `concurrency_group`
believing they bound something (they bound only CTE-only `fire_message=false` tasks and the
no-op path); any agent workflow carrying a `notify_scheduler` stamp (two, both dispatch); and any
capacity reasoning built on "one at a time" — which is how this lane priced a lever at ~2× that
delivers ~+10–15% (the second row's fire lands ~1 s after the first and co-picks the same site 94%
of the time; NOTES 2026-08-25 §5).

## Root cause

A deliberate design change (03-17) that updated the fire path and not the three places that
described the old contract: the `loadDueTasks` comment, the agents' `notify_scheduler` steps, and
the `max_concurrent`/`timeout_seconds` columns' implied semantics. Readers of the guard (this lane,
twice) then inferred the writer's behaviour from the reader's code — WDS-001's own trap, "don't
infer writers from readers", on a different table.

## Fix candidates, ordered by what makes the wrong state unrepresentable

1. **Per-task executions (the D9 fork, owner-deferred 2026-08-21):** record each fire in an
   executions table and have `loadDueTasks`/`countInFlight` count live executions. Makes
   `max_concurrent` real, makes "N concurrent turns" a per-row number instead of a row count, and
   makes the sibling-row idiom unnecessary. Platform code + image roll + council.
2. **Make the contract honest in the code (cheap, no behaviour change):** rewrite the
   `loadDueTasks` comment to state fire-and-forget; add a doc comment on `max_concurrent` /
   `timeout_seconds` in the seed (`052`) saying they bind only for `fire_message=false`; remove or
   comment the inert `notify_scheduler` steps in the two dispatch agents (they cost a
   `query_database` round-trip per run — 777 + 548 per row per day for nothing). Council-scope
   (`cmd/`? no — `cmd/scheduler` is not in `council-scope.sh`; check before assuming).
3. **Document `interval_seconds` as the rate knob** (done: WDS-002, RUNBOOK, LANDMINES 2026-08-25)
   and retire the sibling row in its favour — owner decision pending (README 2026-08-25, options
   A–D).

## How to verify (any of these, no symptom needed)

```sql
-- non-zero on ONE row = not a slot
WITH t AS (SELECT orchestration_id, collected_data->'input_data'->>'task_name' tn, created_at s, updated_at e
           FROM orchestration_states WHERE owner_agent_type='build-pipeline-trigger' AND created_at > now()-interval '24 hours')
SELECT tn, count(*) FROM t a JOIN t b ON a.tn=b.tn AND a.orchestration_id<b.orchestration_id AND a.s<b.e AND b.s<a.e GROUP BY 1;
-- stamps equal while a run is alive
SELECT name, last_triggered_at = last_completed_at both_stamped_at_fire FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
```
`grep -n 'last_completed_at' cmd/scheduler/main.go` — the first hit is the writer.

## Related

`bugs_open/048` (the 07-21 refactor) · WDS-002 (corrected 2026-08-25) · LANDMINES 2026-08-25 (two
entries: this, and the sibling-parity trap) · `WRONG_CALLS.md` 2026-08-25 · 016b §9 (this date) ·
`docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/NOTES_dispatch_throughput.md` 2026-08-25 §5 ·
`sql_for_agents/584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql`.
