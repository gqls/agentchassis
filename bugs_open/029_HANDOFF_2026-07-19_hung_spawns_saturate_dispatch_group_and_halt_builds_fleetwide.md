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

---

## SECOND REPRODUCTION 2026-07-20 — and the trigger is worse than this file says

Reproduced by the robot-hands thread ~25 minutes after the **v1.0.1140 image roll**
(pod `agent-chassis-5567d99bd6-5snzn`, started `2026-07-20T17:58:20Z`). Recovery was
this file's own SQL, unmodified, and it worked exactly as documented.

**The important difference: it was not self-inflicted this time.** The original write-up
attributes it to a thread hammering `build-dispatch-loop` to hurry a slow batch, and calls
it "self-inflicted by a reasonable action". Today nobody dispatched anything. The hangs
appeared **one per minute, on the minute, starting at 17:59 — sixty seconds after the pod
came up** — and simply accumulated until the pool was full:

```
AWAITING_RESPONSES | build-pipeline-trigger | age 18:50 | idle 18:30 | generic-orchestrate-0720-1759
AWAITING_RESPONSES | build-pipeline-trigger | age 17:50 | idle 17:31 | ...-1800
AWAITING_RESPONSES | build-pipeline-trigger | age 16:50 | idle 16:31 | ...-1801
   ... one per minute through ...-1805, 8 in total, against a pool of 8
```

Every one idle for essentially its entire life — they never did anything. So the trigger
is not "a thread over-dispatches"; it is **the scheduler firing normally into a chassis
that has just restarted**. CLAUDE.md's "no orchestration dispatch within ~300s of a pod
(re)start — the spawn is silently dropped" describes the same window from the caller's
side; what this shows is that the *scheduler's own* spawns are dropped too, and unlike a
thread's one-off dispatch, the scheduler keeps firing every 30s and each dropped spawn
**leaves a permanent AWAITING_RESPONSES row**. Eight minutes of that fills the pool.

**Therefore: every chassis image roll halts builds fleet-wide until a human notices.**
That is a much stronger claim than the original file makes, and it means the reaping fix
is not an edge-case nicety — it is on the critical path of every deploy. Until it lands,
**checking for this should be part of the post-roll routine**, not something you discover
25 minutes later while debugging your own stuck item.

Measured impact of this instance: **zero work items completed fleet-wide between the roll
(17:58:20Z) and the recovery (18:23Z)**, and the first completion afterwards was the item
whose stall led me here.

### Measurement trap — do not use `updated_at` for this

`SELECT count(*) ... WHERE status='complete' AND updated_at > <roll>` returns **0 even
after recovery**, because `site_work_items.updated_at` is not maintained (`/bugs_open/035`).
I reported the halt from that column first and it happened to agree with the truth for the
wrong reason. **Use `completed_at`:**

```sql
-- the halt, measured correctly
SELECT count(*) FROM site_work_items
 WHERE status='complete' AND completed_at > '<roll>' AND completed_at < '<recovery>';   -- 0
SELECT count(*) FROM site_work_items
 WHERE status='complete' AND completed_at > '<recovery>';                               -- 1, immediately
```

The independent evidence that does not depend on either column is the one to lead with:
**N hung `build-pipeline-trigger` rows in `AWAITING_RESPONSES` against `max_concurrent`**,
which is directly observable and is the mechanism itself.

### Leave other threads' orchestrations alone

Same as the first reproduction: a live `webdesign-agent` orchestration (idle 2:50) was
present and was **not** touched. The recovery SQL's `agent_type IN
('build-dispatch-loop','build-pipeline-trigger')` clause is what keeps it safe — do not
broaden it to "everything AWAITING_RESPONSES".

### THIRD OCCURRENCE 2026-07-21 — confirms "every roll", and the recovery only buys a window under v1.0.1144

Reproduced on the **v1.0.1144** roll (pod `agent-chassis-59c675c4f-pxr9f`, started
`2026-07-21T08:47:45Z`). This is the third roll in a row that halted builds fleet-wide, which
settles the "every image roll triggers it" claim from the second occurrence — it is not
occasional.

Measured ~73 min post-roll: **37 triaged build items waiting, 1 completed in the previous
20 minutes**, newest `build-pipeline-trigger` stuck at `AWAITING_RESPONSES / call_dispatch`.

**New detail — the one-shot cancel recovery only buys a brief window here.** After cancelling
the six `build-*` orchestrations older than 15 min, exactly **one** work item completed, then
throughput stalled again as fresh triggers re-hung at `call_dispatch` and re-filled the pool.
So the documented recovery is *relief, not repair* under this build — it clears the backlog
for one dispatch and no more. Re-running it is a treadmill; **do not** loop the cancel
(and per the original warning, never loop a *dispatch* to "hurry" it — that manufactures more
hung rows). The durable fix is still the reaper this file's "Why it survives" section calls
for; until it lands, a roll needs a human to babysit dispatch or accept a backlog.

**Diagnostic trap I nearly fell into, recorded because it is the useful part.** The
`AWAITING_RESPONSES` count is a bad halt signal on its own — I watched it go 16 → 24 → 13 and
first read "re-accumulating", then "clearing". Neither was the signal. The newest triggers
were *completing* `complete_idle` in between the hung ones, so the raw count mixes healthy and
dead rows. The signals that actually decide it: **(a)** does a *new* trigger reach
`COMPLETED`, and **(b)** `completed_at` throughput against the `triaged` backlog. Here (a) was
intermittently yes and (b) was ~1/20min against 37 waiting — i.e. degraded, not idle. I
almost recorded "queue is empty, all fine" off the `complete_idle` rows; the 37-item backlog
query is what refuted it. Check throughput-vs-backlog, not the pool census.

---

## CORRECTED DIAGNOSIS 2026-07-21 (bugfix-029 thread) — the title mechanism is wrong; this is bug 003's blast radius, not a concurrency-group bug

Read the whole file above first — the *observations* (degraded builds after a roll,
recovery via cancel, throughput-vs-backlog signal) are all real and reproduce. What is
wrong is the **named cause**: "hung spawns saturate the `dispatch` concurrency group".
The `dispatch` concurrency group cannot be saturated by hung orchestrations, and nothing
about the scheduler's slot accounting is involved. The evidence is direct, from current
`HEAD` code and the live DB; every figure below has its check inline.

### Why the concurrency group is NOT the mechanism (verified against HEAD)

`cmd/scheduler/main.go` computes group occupancy from `scheduled_tasks`, **never from
`orchestration_states`**:
- `countInFlight` (`:327`) counts *enabled scheduled_tasks* whose
  `last_completed_at < last_triggered_at AND last_triggered_at + timeout > NOW()`. A hung
  *orchestration* has no bearing on it. And it self-heals: the `+ timeout > NOW()` guard
  ages any DB-level slot out on its own (this is exactly the "leaked slot" hypothesis
  `bugs_open/048` tested and **refuted**, 048 §"REFUTED").
- The fire path stamps `last_completed_at = NOW()` **immediately** after producing
  (`stampCompleted`, `:253`) — fire-and-forget. So a scheduled task is out of
  `countInFlight` the instant it fires; it never holds its slot pending the orchestration.
- Live roster of the `dispatch` group: **`build-pipeline-trigger` is the only enabled
  member** (`improvement-sweep`, `intent-collection` are `enabled=false`), `max_concurrent=8`.
  One enabled member against a pool of 8 that it vacates on every fire **cannot be refused a
  slot**. Verify:
  ```sql
  SELECT name, enabled, max_concurrent FROM scheduled_tasks WHERE concurrency_group='dispatch';
  -- build-pipeline-trigger | t | 8   ;  improvement-sweep | f | 2 ;  intent-collection | f | 2
  ```
  I have watched `build-pipeline-trigger` fire and COMPLETE hundreds of times an hour with
  hung rows present the whole time (265 COMPLETED in one 6h window while 1–2 sat
  AWAITING_RESPONSES). It is never blocked.

> **This wrong premise has already propagated.** `bugs_open/048` opens by distinguishing
> itself from 029 and, in doing so, restates 029's cause as fact: *"029 is hung
> orchestrations occupying real in-flight slots in the `dispatch` group … the task fires
> and is legitimately refused a slot."* It is not, and it never was. 048's *own* mechanism
> (in-memory `inFlight[group]` pinning the `maintenance` group's head-of-queue) is real and
> correctly diagnosed — but its one-line characterisation of 029 should not be trusted.
> That is the cost of a confident cause in a handoff: the next thread builds on it.

### What actually halts/degrades builds (verified)

The gate is **work-item claiming**, one layer up from the scheduler:

1. `build-pipeline-trigger`'s workflow step **`find_dispatchable_site`** (live
   `agent_definitions`) picks a site to dispatch with:
   ```sql
   ... WHERE wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
       AND NOT EXISTS (SELECT 1 FROM site_work_items active
                       WHERE active.site_id = wi.site_id AND active.status = 'claimed')
   ```
   The `NOT EXISTS (... status='claimed')` is a **per-site mutex**: a site with even one
   `claimed` item is not dispatchable.
2. `build-dispatch-loop` **claims** each item before working it (`claim_work_item_action.go:98`
   sets `status='claimed'`). If the loop then hangs in `AWAITING_RESPONSES` — because its
   handler spawn was dropped (the 300s post-roll window) or lost (**`bugs_open/003`**, spawn
   loses child response) — the item is orphaned in `claimed`.
3. That single orphaned `claimed` item removes its whole site from `find_dispatchable_site`,
   and (if it was the last un-claimed triaged work anywhere) empties the scheduled task's
   `pre_query` (`EXISTS ... status='triaged'`) so the scheduler stops firing the trigger at
   all. **That is the 1st reproduction's "from 13:25 not one build-pipeline-trigger
   orchestration was created"** — not a refused slot; an empty pre-query because everything
   dispatchable had been claimed by dead loops.

So the whole chain is downstream of **bug 003**. 029 is bug 003's *fleet-wide symptom on the
build path*, exactly as this file's header always said ("the consequence half of 003") — but
the consequence is delivered through claimed-item starvation, not through the scheduler.

### It self-heals — bounded, not permanent (the piece the file omits)

Two guards already live in the DB (config, no image roll) bound every hang:
- **`claimed-item-timeout`** (scheduled task, enabled, 120s tick) resets any `claimed` item
  older than **40 min** back to `triaged`/`failed` (clearing `claimed_by/claimed_at`), and
  auto-completes claims >15 min whose artifact is provably done. This is the load-bearing
  un-wedge — it is what returns a starved site to dispatchable. Evidence it fires:
  ```sql
  SELECT left(error,42), count(*) FROM site_work_items
   WHERE error LIKE 'Claim timed out%' OR error LIKE 'Auto-completed%' GROUP BY 1;
  -- 'Claim timed out — handler pod likely died' | 396   (newest today)
  -- 'Auto-completed: work verified done…'        |  93
  -- 'Claim timed out (attempts exhausted)'        |  25
  ```
- **`stale-orchestration-reaper`** (enabled, 180s tick) fails the hung `build-dispatch-loop`
  at 30 min / any `AWAITING_RESPONSES` at 90 min / `EXECUTING_STEP` at 4h, and expires
  `awaited_requests`. (These 30/90-min AWAITING clauses were committed **2026-04-03**,
  `7c9d7b67d` — i.e. they were already live during *all three* reproductions. The reaper the
  "third occurrence" section calls for as the missing durable fix **already exists**.)

So the maximum unattended window is ~40 min (claimed-item-timeout), not "until a human
notices". The 1st/2nd reproductions were manually recovered at ~22–25 min — *before* the
auto-reset would have fired — which is why they read as a permanent halt. The manual
`UPDATE orchestration_states SET status='CANCELLED'` does **not** release claims (no trigger
does: only `claimed-item-timeout`, `fail_work_item` on a live loop, or admin requeue), so the
one-tick "recovery" was the pre-query finding other, un-starved work — not the cancel
un-wedging anything. That is also why the third occurrence saw the cancel "buy one dispatch"
and then re-stall: it never addressed the claims or the ongoing spawn loss.

### The honest current picture

At 2026-07-21 15:49Z the fleet is **healthy**: 0 `claimed` items older than 40 min, builds
dispatching (`find_dispatchable_site` returns sites), `claimed-item-timeout` recycling
normally. 029 is therefore **not an active permanent outage** — it is a *degraded-throughput
window after a roll*, whose depth and duration are set by the **rate of bug-003 spawn loss**
vs the ~40-min recycle. When spawn loss is heavy right after a roll (the third occurrence),
the guards can't keep pace and builds crawl (1 completion / 20 min vs 37 waiting); once the
loss rate drops, they catch up. Reaping AWAITING_RESPONSES faster would **not** fix this,
because the driver is new losses, not old rows.

### Where the durable fix belongs

Route it to **`bugs_open/003` F2/F3** (parent-driven await timeout + retry, migration 180),
owned by the bugfix_003 workstream — that is the only change that stops loops hanging in the
first place. Two *optional* config-only mitigations for the window, neither a real fix:
tighten `claimed-item-timeout`'s 40-min reset (risk: prematurely reclaiming a legitimately
slow 1200s handler → duplicate work — do not touch without measuring the tail), or add a
post-roll dispatch check to the deploy routine. **Do not** add another orchestration reaper —
the reaper is live and is not the lever.

### My own near-miss (recorded per the working-docs rule)

Mid-diagnosis I had concluded "nothing releases orphaned `claimed` items → sites wedge
forever" and was about to write it here as the residual. It is **false** — I had read the
reaper, `load_work_items`, `fail_work_item` and the triggers, but skipped the
`claimed-item-timeout` scheduled task, which is the whole self-heal. Caught by reading the
`090` trigger script's own comment ("resets any claim older than 40 minutes"). The lesson is
the CLAUDE.md one exactly: the failure was not-looking at one task, and confidence was no
protection. Logged in `WRONG_CALLS.md`.

*Independent diagnosis-loop verification of this correction was offered but not run — the
model is grounded on current-HEAD code + live queries (all cited above) and is corroborated
by three independent reproductions; the durable fix (003 F2/F3) is unchanged either way.*

---

## Fresh instance, 2026-07-26 19:23–19:54 UTC — contributed by the bugfix_077 thread

Not a new diagnosis and not a competing fix (`who-owns.py 029` → dispatch_queue_serialisation).
Recording it here because it is a **council-gate** instance rather than a build one, and
because the discriminator is unusually clean.

A council submission (`fix_correlation_id 346500db-89ca-47f3-bc5a-e1c099d6f4f8`,
orchestration `5c6e4fa1-0b84-4211-bcd1-8271df8539f7`) reached the first review step and
stopped dead:

```
created_at 19:23:14.548 | updated_at 19:23:15.295 | status EXECUTING_STEP
current_step / currently_executing = review_editquality
awaited_steps = []        error = NULL
```

Frozen for **31 minutes and counting** — one second of life, then nothing. Note
`awaited_steps` is EMPTY while `currently_executing` is set: the row is holding the step
lock with nothing recorded as outstanding, so a reaper keyed on awaited responses has
nothing to key on.

**The discriminator — this is not queue latency.** A *different* correlation
(`569241fb`) started **after** mine at 19:27:52 and had **COMPLETED twice** by 19:49:07,
and was on a third round at 19:51. The fleet advanced 10 orchestrations in the five
minutes before this was written. So the lane was healthy throughout, and one run in it
was simply lost.

**It is not a singleton, and the step repeats.** Fleet-wide at 19:54, seven orchestrations
were frozen >20 min mid-step — and **two of them at `review_editquality`**: mine, and
`f4610451` from 18:34:11 (also alive for 7 seconds, then frozen). Same step, ~50 minutes
apart, different submissions.

```sql
SELECT current_step, count(*) FROM orchestration_states
WHERE status IN ('EXECUTING_STEP','AWAITING_RESPONSES')
  AND updated_at < now() - interval '20 minutes'
GROUP BY 1 ORDER BY 2 DESC;
--  review_editquality 2 | call_content_writer 1 | call_dispatch 1
--  complete 1 | review_debug_historian 1 | review_reuse_agent 1
```

**Whether `review_editquality` is special is UNTESTED** — two hits in one evening is a
coincidence-sized sample, and it is also the *first* review seat every submission reaches,
so it gets the most attempts. Flagging it as a thing to count, not a finding.

**What this cost, and the thread-level lesson.** This was the *second* lost dispatch for
one change in one evening. The first vanished with **no row at all**, published 2–4 minutes
after a chassis pod restart (`startTime 18:35:07Z`) — CLAUDE.md's documented ~300s drop
window, which nothing in the 097 trigger warns about: it prints a correlation id and a
cheerful lane-depth report either way. The resubmission landed instantly ("Lane is clear,
LAG 1") and then hung here instead. **A third round was not fired** — spending another
full council into a lane that is demonstrably losing runs buys a likely-identical hang, so
`bugs_closed/077` records that no verdict was ever returned and carries **no
`Council-Reviewed:` trailer** (the trailer is earned by an APPROVED verdict only).

Two cheap things that would have changed the evening, both outside this bug's remit:

1. **Check the chassis pod's `startTime` BEFORE publishing anything that spends credits**,
   not after waiting on it — one `kubectl get pods` command.
2. The 097 trigger could refuse, or at least warn, when the chassis has been up < ~5
   minutes. It already queries the lane; it does not query the pod.
