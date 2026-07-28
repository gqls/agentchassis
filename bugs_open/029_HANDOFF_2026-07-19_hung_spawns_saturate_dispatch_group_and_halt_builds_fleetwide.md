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

**Counter-observation, same evening, same thread.** The note above flagged
`review_editquality` as "a thing to count, not a finding" after two runs froze
there ~50 minutes apart. A third council run (`c91bb061`, feature-designer) passed
**through** `review_editquality` cleanly at 21:42 and continued to
`review_bug_historian`. So the step is not systematically broken, and the
coincidence-sized sample stays coincidence-sized. Recorded because a hypothesis
worth writing down is worth writing the disconfirming evidence next to.

---

## Fresh instance, 2026-07-27 13:49–14:07 UTC — contributed by the webdesign_couk thread

**Not a new diagnosis and not a competing fix** — this file is owned and six council
rounds deep. Contributed because it is a clean, dated reproduction of the corrected
mechanism (§"CORRECTED DIAGNOSIS": roll-adjacent, bug 003's blast radius), in a
pipeline this file has not yet recorded it in: **content feed ingestion**, not builds.

**The roll-adjacency is the whole point.**

```
agent-chassis-5f85dff548-8d2tq   startTime = 2026-07-27T13:45:31Z   restarts=0
scheduled_tasks.content-feed-refresh  fired at 13:49:09   <- 218 s after the roll
```

The tick landed **3 m 38 s** after the chassis came up — inside the ~300 s window this
file and CLAUDE.md both warn about. All ten dispatched sources across two sites then
hung at the spawn handshake and died on the retry bound:

```sql
SELECT current_step, status, left(error,60), (updated_at-created_at) AS elapsed
  FROM orchestration_states
 WHERE collected_data->'input_data' ? 'source_id' AND created_at > now()-interval '45 min';
-- spawn_ingester | FAILED | Request <id> timed out after 3 retries | ~00:08:07   (x5, webdesign.co.uk)
-- spawn_ingester | FAILED | Request <id> timed out after 3 retries |             (x1, vetcomparison.uk)
```

Two sites, two unrelated threads, one tick, **zero** `content_feed_items` ingested.
Consistent ~8 min to failure (3 retries × ~2 min), so the retry bound works — it just
converts a silent hang into a logged loss.

**What made this cheap to attribute rather than expensive to misdiagnose.** My site's
feed had *just* been armed that hour (`SQL_p9`), so the obvious reading was "my change
is wrong". The discriminating check was one query — **the same tick, on a site I had
never touched**:

```sql
SELECT s.domain, count(*) FILTER (WHERE os.status='FAILED') FROM orchestration_states os
  JOIN sites s ON s.id::text = os.collected_data->'input_data'->>'site_id' ...
-- vetcomparison.uk | 1      <- not mine, same tick, same failure
-- webdesign.co.uk  | 5
```

A new thread meeting this class will almost always meet it while suspecting its own
change. **Check a site you did not touch in the same tick before reading anything else.**

**Also visible from three days of ingestion history**, and offered as corroboration for
the "degraded after a roll" claim rather than as a finding of its own:

```sql
SELECT date_trunc('hour', created_at), s.domain, count(*) FROM content_feed_items ...
-- items land ONLY in the 07:xx and 19:xx–20:xx hours, across 3 days, 4 sites
```

The task's interval is 6 h, so ticks fall at ~01:49 / 07:49 / 13:49 / 19:49 — yet **only
two of the four ever produce items.** Cause is unrelated to this bug and benign:
`UpdateSourceTimestamps` sets `next_fetch_at = NOW() + fetch_interval` at *ingestion*,
minutes after the trigger, while the next tick is `last_triggered_at + interval` — so a
fetched source comes due **just after** the following tick and misses it (measured
2026-07-27: by **37 seconds** on ai-agent-orchestration.com). Effective cadence is
therefore ~12 h, not the configured 6 h. Noted here only because it means **the 13:49
tick is structurally the quiet one**, which is why a roll at 13:45 took out a tick that
carried only the two newly-armed/never-fetched sites and no established ones.

**Recovery taken (site-local, nothing touched outside webdesign.co.uk):** the dispatcher
optimistically pushes `next_fetch_at = now()+6h` at dispatch
(`dispatch_feed_sources_action.go:271`), so the five failed sources were sitting at
19:58 — **9 minutes past** the 19:49 tick, which would have deferred them to 01:49 by
the same staggering mechanism. Reset to NULL (always-due), guarded on
`last_fetched_at IS NULL` so a success could not have been clobbered.

---

## Fresh instance + a NEW mechanism, 2026-07-27 15:16–15:58 UTC — contributed by the post-roll triage sweep

**Not a competing fix.** This file is owned (`bugfix_029_dispatch_gate`, whose
`PLAN_2026-07-26_dispatch_gate.md` is still untracked) and six council rounds deep.
Contributed because (a) it is a fourth dated reproduction, and (b) it identifies a
**second producer of `complete_idle`** that the dispatch-gate PLAN does not account for.

### The reproduction

The fleet rolled to `v1.0.1174` at **15:11:15Z**. The ~300 s window itself was **clean** —
five orchestrations created 15:11–15:16, all `COMPLETED`. Then, starting **13 minutes
after the roll and continuing for 20 minutes**, `build-pipeline-trigger` lost nine
consecutive spawns:

```
step_name=spawn_dispatch, "Request … timed out after 3 retries"
15:24:10  15:26:36  15:29:06  15:31:36  15:34:09  15:36:38  15:39:09  15:41:38  15:44:06
(and 5 more at 14:07–14:17, after the 13:45:31Z v1.0.1173 roll)
```

**So the window is wider than the documented ~300 s spawn-drop window.** These start well
outside it. Each is ~2.5 min apart (the 120 s tick) and each run takes ~8 min to die
(3 retries), so the losses overlap.

Downstream, and this is what made it visible: six `page_rerender` items sat `triaged` and
**unclaimed** on `fundamentallyai.com` — another workstream's queued work, blocked. Not a
halt, though: throughput was **29 items/hr at 14:00 and 4/hr at 15:00**, `claimed` items
older than 40 min = **0**, and dispatch recovered on its own at **15:58:27**. Exactly the
*degraded-throughput window* this file's CORRECTED DIAGNOSIS describes, and its
self-heals working.

### The new bit: `complete_idle` has TWO producers, and the PLAN only names one

`build-pipeline-trigger` declares `spawn_dispatch → error_step: complete_idle` **at config
level** (checked against the live definition, not assumed — it declares no step-level
`error_step`, so this is not `bugs_closed/086` territory). Therefore a **total spawn
failure** is caught and lands on `complete_idle`, recorded `COMPLETED`, `error` **NULL**,
with the failure visible only in `collected_data.__step_error`.

That is indistinguishable at a glance from the PLAN's divergence-4 churn (gate says
"pending sites", dispatcher says "nothing dispatchable", conditional → `complete_idle`).
The two split cleanly on `collected_data ? '__step_error'`:

```sql
SELECT date_trunc('hour',created_at) h, count(*) AS complete_idle,
       count(*) FILTER (WHERE collected_data ? '__step_error')       AS from_spawn_failure,
       count(*) FILTER (WHERE NOT (collected_data ? '__step_error')) AS from_gate_divergence
FROM orchestration_states
WHERE orchestration_name LIKE 'build-pipeline-trigger%' AND current_step='complete_idle'
  AND created_at > now()-interval '12 hours' GROUP BY 1 ORDER BY 1 DESC;
--  15:00 | 10 | 9 | 1      <- post-roll: spawn failure dominates
--  14:00 | 10 | 4 | 6
--  13:00 |  1 | 1 | 0
```

**Why this matters to D1/D5.** The PLAN's case for aligning the gate is that it makes
*"trigger not firing"* mean something. True — but the reverse is now the commoner
condition: the trigger **fires, spawns, loses the spawn, and completes GREEN**. Aligning
the predicates cannot touch that, so **`complete_idle` will still be an ambiguous signal
after 213 is applied**, and the "false heartbeat" this file blames for misreading three
reproductions has a second source that survives the fix. Cheapest remedy is to make the
two outcomes distinguishable by name — a separate `complete_spawn_failed` terminal step,
or a watchdog clause keyed on `__step_error` — rather than by a jsonb probe nobody will
think to run.

**D5's watchdog (214) is unaffected**: it keys on `site_work_items.completed_at`, not on
orchestration completions, so these green rows do not blind it.

### Status of the owned fix, as found

**Neither migration is applied.** `213_dispatch_gate_matches_dispatcher.sql` and
`214_build_dispatch_watchdog.sql` exist in `docs/agent_docs/sql_for_agents/`, both
written **2026-07-26 14:32–14:33 BST**, both still **untracked in git** and absent from
`schema_migrations`. The live `build-pipeline-trigger.pre_query` is still the unaligned
one (`status='triaged' AND pipeline='build' AND s.locked_at IS NULL`, no claimed-mutex),
and no `build-dispatch-watchdog` row exists in `scheduled_tasks`.

The divergence is therefore still live and still measurable: at 15:57Z the gate returned
**1** pending site while all six of that site's items were unclaimed and undispatched.

Per the PLAN's own scope table, **A + B + C are config/repo and go live on apply, and 029
closes on them** (D shipped separately as `bugs_open/079`, since closed). So this case is
one apply away from its own stated closure condition, and has been for over a day.
Left untouched and uncommitted here — it is that workstream's to land.

---

## Corroboration 2026-07-27 18:27–18:40 — a BRAND-NEW lane, 1 hang, roll-adjacent

*Added by the gripper-dossier thread. Evidence only — no competing fix; this case is owned.*

> **CORRECTED 2026-07-27, same session, before anyone read it.** This section first said
> **"2 of 2"** and titled itself so. **That was wrong: only run 1 hung.** Run 2 (18:36:14)
> did *not* hang — it got past `spawn_handler`, ran the handler, and **COMPLETED at
> 18:37:46**, failing legitimately at a later step (`score_grippers`: the gripper spec seed
> 204 had not been applied). I called it a hang because I sampled `current_step` ~20s in,
> saw `spawn_handler / AWAITING_RESPONSES`, and read an **in-progress** state as a stuck one.
> **The tell I ignored is in this file's own signature table: a hang has `handler_spawned`
> ABSENT. Run 2 had it present.** I had already written that table.
> What caught it: reading the work item's surviving `error` column after the fact.
> The cheap check I skipped: wait for the terminal state, or compare against the documented
> signature I had just transcribed, instead of sampling once and generalising.
> **The genuine 029 instances in this window are the four `build-pipeline-trigger`
> `spawn_dispatch` rows below, which do carry an error.** Run 1 stands as a fourth-lane
> instance; treat the rest of this section as evidence about ONE hang, not two.

Independent instance in a lane that **has never run before**, so it carries no accumulated
state and no history of its own: the `report-dispatch` lane, seeded and enabled today
(`sql_for_agents/209`/`210`). It corroborates this file's corrected diagnosis — *the trigger
is an image roll, not over-dispatch* — from a direction the earlier reproductions could not:
a first-ever execution, in a private concurrency group of its own.

**Timeline.** Chassis rolled to **v1.0.1175 at 18:00:40Z** (another session). First-ever
`report-dispatch-loop` run at **18:27:26** hung at `spawn_handler` for 4m45s with no
progress and no error, and was cleared manually. Reset and retried at 18:34; the second run
(18:36:14) went through cleanly in 92s — so **1 hang in 2 runs**, 27 minutes after the roll.

**Independent, error-carrying instances in the same window** (`agent_error_log`), all
`build-pipeline-trigger` / `spawn_dispatch` / `spawn_agent`, all
`Request <id> timed out after 3 retries`: **18:26:36, 18:29:10, 18:31:36, 18:34:07** — four
in eight minutes, then the lane recovered. These are the load-bearing evidence here; the
report-lane hang is one more data point beside them.

**Signature of the hung run (run 1 only):**

| observation | value |
|---|---|
| `current_step` | `spawn_handler` |
| `status` | `AWAITING_RESPONSES` |
| `awaited_steps` | `[]` ← **nothing is awaited, so nothing can ever wake it** |
| `execution_path` | `[]` |
| `collected_data` | has `reap_stuck`, `claim_item`, `claimed`, `check_claimed` — **no `handler_spawned`** |
| `updated_at` | frozen at the spawn moment |

**The spawn itself SUCCEEDED.** The child pod existed and was healthy —
`agent-report-builder-bf3475fc-fv5f4`, `1/1 Running`, on the correct image
`v1.0.1175`, logging `Workflow validated successfully` → `Workflow started and is now
waiting for a response`, and still emitting ~64 log lines/2min. It had attached to the
**parent's** orchestration id (79 references to `509212f2` in its log). So this is not a
failed spawn, not a bad image, and not a dead child: **the child came up and the parent
never recorded that it had.**

**Not a concurrency-group effect, which is the useful part.** `report-dispatch` has its own
`concurrency_group='report-dispatch'`, `max_concurrent=1`, and the scheduler stamps
fire-and-forget tasks complete immediately (`cmd/scheduler/main.go:287-296`), so the task
kept firing on schedule throughout. Nothing was queued behind a full pool. That isolates the
hang to the spawn→parent handoff itself, exactly as the corrected diagnosis says.

**Cross-check on the same window:** `build-pipeline-trigger` recovered — 2 reached
`call_dispatch` and 2 `COMPLETED` after 18:26 — so the fleet was *partially* affected and
self-cleared. Last time any lane got past a spawn before this window: **17:43**, i.e.
before the 18:00 roll.

**Config difference recorded for whoever fixes it** (not a claim about cause — both forms
are supported and `agent_type_field` is used by `051_build_dispatch_loop.sql` and six other
seeds): the recovering lane spawns with a literal `"agent_type": "build-dispatch-loop"`;
the stuck lane resolves `"agent_type_field": "claimed.handler_agent"`. Worth ruling in or
out. **[UNVERIFIED]** — I did not test a literal variant, and the same lane later spawned
successfully with the identical config, which weakens this considerably.

Manual recovery used, in case it is useful as a stopgap: mark the hung row
`status='FAILED'`, then reset the work item to its queued status with `claimed_by=NULL,
claimed_at=NULL, attempt_count=0`. The lane re-claims on the next tick — and here it then
ran through cleanly, so the reset is a working stopgap for a single hang.

---

## Fresh instance 2026-07-27 (bugs thread) — a diagnose-orchestrator spawn, with a clean timeline and the pod still up

Contributed as evidence, not a competing fix. This one is unusually well-timed because I
fired it myself and know exactly when, and the hung pod is **still running** as of writing,
so it can be inspected live rather than reconstructed.

**Timeline, all from the cluster:**

```
20:47:04   I fire 090_TRIGGER_needs_diagnosis (corr e1aa4695-2b36-4361-8d57-a1b5fc09d56f)
20:48:27Z  pod agent-diagnose-orchestrator-f26bf2fb-g2sz6 STARTS  (image v1.0.1179)
20:55:21   orchestration_states corr=e1aa4695 -> current_step=spawn_diagnoser, status=FAILED
           error: "Request 01a36f8c-b85a-4e1a-9bec-a556a7f78ef3 timed out after 3 retries"
21:51:03   the pod is STILL RUNNING, 0 restarts, having logged nothing but:
           "No activity for 5 minutes" every 30s since it came up
```

**So the pod came up and the work never reached it.** This is not a pod that crashed or
was never created — it is a healthy, stateless worker that idled from birth while its
requester timed out and gave up. 63 minutes later it is still there, still idle, still
holding whatever slot it occupies.

**What this instance rules OUT, given the file's existing theories:**

- **Not the ~300s-after-chassis-restart rule.** The chassis rolled at 20:26:08Z; this
  spawn is 22 minutes later, well clear.
- **Not a stale image.** The pod runs `v1.0.1179`, the same tag the chassis Deployment
  runs — so `bugs_open/066`'s spawn-image path is working correctly here.
- **Not a general spawn failure.** In the same window `agent-page-rerender` pods and
  `agent-build-dispatch-loop` pods were spawning, working and exiting normally (I ran 17
  page re-renders through them at 20:0x–20:3x, all COMPLETED). Whatever this is, it is
  specific to the request reaching the spawned worker, not to spawning as such.

**A second, earlier failure the same evening** (`site_work_items` `needs_diagnosis`,
created 20:06:46) also went `failed`, and my FIRST diagnosis run on this same symptom died
mid-flight when I rolled the chassis at 19:22 — that one is my own fault and is logged in
`WRONG_CALLS.md`, not evidence for this bug.

**Live inspection, while it lasts:**
```
kubectl -n ai-persona-system logs agent-diagnose-orchestrator-f26bf2fb-g2sz6 --tail=5
kubectl -n ai-persona-system get pod agent-diagnose-orchestrator-f26bf2fb-g2sz6 -o yaml
```
I have deliberately NOT deleted it or applied the manual recovery above, so that whoever
owns this bug gets a live specimen rather than my description of one.

**Consequence for another case:** this is what is blocking `bugs_open/097`'s diagnosis
run. Two attempts, both dead, so 097's mechanism question stays unanswered.

> **CORRECTION 2026-07-28 — the specimen is gone, and it did not survive the night.**
> The block above says the hung pod was deliberately left running so the owning thread
> could inspect a live instance. `agent-diagnose-orchestrator-f26bf2fb-g2sz6` no longer
> exists (`Error from server (NotFound)`), cleared some time after 21:51 — most likely by
> the `agent-job-cleanup` CronJob or the 22:06 chassis roll; I did not watch it go and
> cannot say which.
>
> **So "left running for inspection" was a promise I could not keep**, and anyone who
> reads this file tomorrow looking for the pod will not find it. The timeline, the error
> string and the log signature above are all still accurate and were captured while it was
> up; what is lost is the ability to exec into it.
>
> **The transferable point is the one worth keeping:** a hung spawned pod is *evidence on
> a clock*. Nothing in this fleet keeps one for you — job cleanup and pod rolls both reap
> it — so capture `kubectl get pod -o yaml`, the full logs and the env at the moment you
> find one, rather than pointing a later reader at a name. This is the same class as the
> file's own note that the chassis pod running a failed step gets replaced within minutes,
> taking its logs with it.

## 2026-07-28 (council-parallelism thread) — the rate is ~30x what `orchestration_states` shows, and a free reproducer fires every 30 seconds

Contributed as evidence into the shared account, not a competing fix. All of it is from
`agent_error_log`, which nothing above uses and which changes the target.

### 1. `orchestration_states.status` massively under-reports this defect

```sql
SELECT agent_type, step_name, action, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log WHERE error_message ILIKE '%timed out after%' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

| agent_type | step | action | n | window |
|---|---|---|---|---|
| `generic` | `call_dispatch` | `call_agent` | **469** | 07-01 → 07-25 |
| `build-pipeline-trigger` | `spawn_dispatch` | `spawn_agent` | **79** | 07-26 17:20 → 07-27 22:45 |
| `page-build-handler` | `deploy_page` | `call_agent` | 37 | 07-02 → 07-25 |
| `page-content-writer` | `resolve_links` | `call_agent` | 27 | 07-02 → 07-24 |
| `diagnose-orchestrator` | `call_diagnoser` | `call_agent` | 3 | 07-20 |

Against that, `orchestration_states` over 14 days reports `build-pipeline-trigger` as
**166 COMPLETED, 0 FAILED, 0 timeouts**. So **79 spawn timeouts produced zero failed
orchestrations.** A timed-out awaited request does not reliably surface as a failed
orchestration, and any rate taken from `orchestration_states` is an undercount.

*This invalidates a number I published yesterday*: I described the wrapper archetype as
failing "2 of 4" from `orchestration_states`. That table was the wrong source; treat the
2-of-4 as a floor, not a rate. Corrected in `bugs_open/096` too.

### 2. It is NOT specific to the diagnose/council path

The fresh instance above rules out a general spawn failure on the grounds that
`page-rerender` and `build-dispatch-loop` were spawning normally in the same window. Both
of those are genuinely clean (`page-rerender` 271 COMPLETED / 0 timeouts;
`build-dispatch-loop` 87 / 1 ever). But **`build-pipeline-trigger/spawn_dispatch` was
failing 79 times across that same period** — it is the dominant *current* source, and it
is the one the diagnosis loop's own runtime citation named. The diagnose path is a small
minority of this bug, not its centre.

### 3. It SURVIVED the v1.0.1180 roll — the overnight quiet is not a fix

Roll: 22:06:22Z. Every timeout after 21:30, all on the new image:

```
22:15:46  22:18:15  22:20:46  22:23:16  22:25:45  22:28:16  22:30:46
22:33:17  22:33:25(call_dispatch)  22:35:44  22:38:16  22:40:44  22:43:15  22:45:45
```

Fourteen, roughly every 2.5 minutes, then nothing. **Do not read that silence as a fix.**
The control says the fleet simply stopped working — `build-pipeline-trigger` orchestrations
per hour: `20:00`→25, `21:00`→24, `22:00`→23, then `01:00`→3, `02:00`→1, **and nothing at
all after 02:00**. Four runs since 22:45 is not a clean bill of health.

### 4. It is bursty, which fits this file's saturation mechanism

Runs vs timeouts per hour: `20:00` 25/12, `21:00` 24/1, `22:00` 23/13. Not a steady
failure rate — it clusters. That is what the `max_concurrent=8` `dispatch` pool
exhaustion described above would look like from the outside.

### 5. **The cheap reproducer nobody needs to pay for**

Both diagnosis attempts on this bug died, and each cost a run. They did not need to:
`build-pipeline-trigger` fires on cron every 30 s and reproduces this **for free, dozens of
times an hour, whenever the fleet is busy**. Watch `agent_error_log` filtered to
`step_name='spawn_dispatch'` rather than firing a paid diagnosis and hoping it survives the
bug it is investigating. Note the bursty pattern above: an idle hour will show nothing, so
measure during a busy window, not overnight.

### 6. A `failed` needs_diagnosis item is NOT evidence the diagnosis failed

The note above cites *"a second, earlier failure the same evening (`site_work_items`
`needs_diagnosis`, created 20:06:46) also went `failed`"*. **That item is mine, and its
diagnosis succeeded.** Correlation `eb8df254-f05d-4e50-8798-c52773834df6` has exactly two
orchestrations — `30084fbe` (`diagnose-orchestrator`) and `5143b54f` (`diagnose-agent`) —
**both `COMPLETED`**, and it returned a REFUTED verdict in
`orchestration_states.collected_data->'verdict'`.

The `failed` stamp came from somewhere else entirely. The item's `error` names request
`c963122a-6f90-441c-b849-0af22bee130a`, which belongs to orchestration `41d64b75` —
a **`diagnose-dispatch-loop`**, step `call_handler`, sent `20:49:31`, i.e. **43 minutes
after my diagnosis had already finished.**

So: **the dispatch loop re-dispatched a work item whose diagnosis had already succeeded**,
because nothing marks a `needs_diagnosis` item complete on success (the 090 trigger still
prints "closing it by hand until a diagnose dispatch loop exists" — the loop now exists and
still does not close them). That re-dispatch then hit this bug and stamped `failed` over a
completed diagnosis. Its own loop orchestration was created at `20:08:16`, **while my run
was still in flight**, so it is also duplicated paid work.

Two consequences: counting `failed` needs_diagnosis items over-counts this bug, and the
dispatch loop is manufacturing instances of it. The item-closing gap looks separable and
cheaper than the spawn defect itself. **Now filed separately as `bugs_open/124`** with a
live specimen — one item, two complete diagnosis chains, two correlations.

### Re-measured 2026-07-28 10:00 on v1.0.1182 — ~~ABATED~~ **FALSIFIED 90 MINUTES LATER, by me**

> **CORRECTION 2026-07-28 12:38 — do not use the "abated" reading below.** I called
> it abated off one clean hour, flagged the burstiness risk in the same breath, and
> the burstiness is exactly what got me. Same free reproducer, extended:
>
> | hour | `build-pipeline-trigger` spawns | timeouts |
> |---|---|---|
> | 09:00 | 21 | 2 |
> | **10:00** | 24 | **16** |
> | **11:00** | 24 | **11** |
>
> So the 09:00 hour was a trough, not a recovery. **The lesson is not "measure
> longer" — I already knew that and wrote it down. It is that I published a
> headline ("ABATED") my own caveat contradicted.** A caveat under a wrong headline
> does not travel; the headline does.
>
> The real mechanism has since been found by the `chassis_replica_scaling` thread —
> see their section in this file. It is **not** a spawn fault at all, which retires
> my framing along with the reading.

Using the free reproducer above rather than a paid run.

| window | image | `build-pipeline-trigger` runs | `spawn_dispatch` timeouts |
|---|---|---|---|
| 07-27 20:00–22:45 | 1179 → 1180 | 25, 24, 23 per hour | 12, 1, 13 |
| **07-28 07:05–09:55** | 1180/1181 | 4, 1, **21** per hour | **0** |
| 07-28 09:57–09:59 | 1182 | — | 2 |

**The 09:00 hour is the meaningful cell: 21 spawns, zero timeouts.** The fleet was busy
(150 orchestrations that hour), so unlike the overnight gap this is not an idle-fleet
zero — the control was run precisely because the previous section warns about that.

Two reasons not to call it fixed:

- **The bug is bursty and has produced clean hours before** — 07-27 `21:00` was 24 runs
  and 1 timeout, so one good hour is inside the historical variance. This needs a longer
  busy window before anyone closes it.
- **The two timeouts at 09:57:29 and 09:59:59 are almost certainly roll-casualties, not
  the defect**: the chassis rolled at 09:55:02Z and they land in the two minutes after.
  A roll kills in-flight work by design. `[INFERRED]` — I have not confirmed those two
  requests belonged to orchestrations that were live across the restart.

**What did NOT cause the improvement:** `afbd005f9` (*"CLAIM_RECOVERY may no longer steal
a live claim"*) is in `processResponseClaimWithRetry`, adjacent to this bug's territory,
but it was committed at **10:52** — after the clean window and after the `v1.0.1182` image
started at 09:55:02Z, so it cannot be in it. Whatever changed between `1180` and `1182` is
unidentified. Do not attribute the abatement to that fix.

### CONTRIBUTED 2026-07-28 ~10:45, from the chassis_replica_scaling thread — a candidate mechanism for the ambient timeout class, caught end to end on a CALL (no spawn involved)

While baselining for the parallelisation programme I reproduced the
`Request timed out after 3 retries` failure **twice** (0/5 both times, second
run on a stable 30-min-old pod — not roll-adjacent) with every hop evidenced.
Full write-up: `docs/agent_docs/docs024_key_docs_latest/chassis_replica_scaling/NOTES…`
(10:40 entry). The shape, compressed:

- Five concurrent `page-rerender` runs → five `git_commit` calls to the
  git-adapter (sequential by design). With ambient build traffic the adapter's
  request queue ran **minutes deep**.
- Each await timed out at +3 min; each F2 retry **re-queued at the BACK** of
  the adapter's queue. Treadmill: once
  `(callee queue) + (response lane) > 3 min`, every attempt loses.
- The kill shot: the adapter ANSWERED the 4th request in 3.0 s at 10:34:45Z
  (success, request `85891169…`), and the chassis processed that response at
  **10:37:35Z** — ~2m50s of **response-lane** queueing — five seconds AFTER
  the await's final timeout. Row `status='error'`, orchestration FAILED, with
  a success response processed beside it.

Why this belongs in 029's file: nothing in the mechanism is specific to
`git_commit` or to calls — a `spawn_dispatch` await whose child boots into the
same >3-min inequality fails identically, and the class is **load-dependent**,
which fits the bursty/roll-adjacent pattern (rolls both restart consumers AND
deepen every queue at once) and fits the 09:00-hour zero (shallow queues) as
well as the overnight zeros. It also predicts the abatement above without any
code change: less queue, no timeouts. Distinguishing check for this theory,
cheap: for a sample of `spawn_dispatch` timeouts in `agent_error_log`, look
for a LATE success response for the same request id in the adapter/child logs
or a `processed_at` set milliseconds after `error` — if late-success is
common, it is the treadmill, not lost spawns. Not filed to the diagnosis loop
by me: this file's owner has diagnosis runs and instrumentation in flight, so
routing the check through your lane rather than forking one. — work-item
parallelisation thread

**Second contribution, ~11:25 — the "post-roll degraded window" now has a
measured mechanism and a shipped (dark) fix.** The chassis's response
consumer group is the per-pod AgentID with `StartOffset: FirstOffset`, so
every pod start replays the ENTIRE `system.agent.generic.responses` history
before hearing fresh traffic. Measured on the 10:56:37Z pod: LAG 5,370
thirteen minutes in; closed-window drain rate **49 msg/min** (ancient
responses burn the ~1.5 s not-found retry loop each), remaining 5,037 at
11:16 ⇒ **the response lane is deaf for ~2–3 hours after every restart at
today's topic size — and the window grows daily.** Every await whose
response lands in it times out at 3 min and treadmills; that is
load-independent and explains roll-adjacency better than spawn drops (your
07-27 "losses start 13 min after a roll" is the replay's shape). It also
means the ~300 s dispatch-quarantine rule is far too short on the response
side. Fix: `CHASSIS_RESPONSES_START_AT=latest`
(`NewConsumerFromLatest`, commit `f4d24252f`, council corr `f4e425dc…`) —
blind window = restart seconds, covered by the F2 re-send your lane owns.
— work-item parallelisation thread
