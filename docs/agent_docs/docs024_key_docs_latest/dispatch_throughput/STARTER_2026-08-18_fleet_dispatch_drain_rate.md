# STARTER NOTE 2026-08-18 — why the drain rate is what it is, and the four levers

**Not a workstream yet.** This is a measured starting position for whoever picks the
question up. If you do, create the standing five in this directory first (CLAUDE.md,
"Working docs"), and read §6 before asserting any cause.

**Raised by:** the idea.uk lane's observation *"drain rate is ~1 item per 3 hours on
idea.uk"* (`idea_uk_vm_site/HANDOFF_2026-08-16_continue_here.md` §6). That figure is
real but it is a **queue-position** number, not a throughput number — see §1.

All figures below `[MEASURED 2026-08-18 ~16:00 UTC]` unless marked otherwise.

---

## 1. First correction: idea.uk is not draining at 1 item / 3 hours

idea.uk's completions by day: **08-14: 42 · 08-15: 61 · 08-16: 1 · 08-17: 42 · 08-18: 10.**
On 08-17 — the day the 3-hour figure was taken — its claims fell:

| hour (UTC) | claims |
|---|---|
| 15:00 | 1 |
| 16:00–18:00 | **0** |
| 19:00 | 25 |
| 20:00 | 9 |
| 21:00 | 4 |

So the site waited ~3.5 hours for its **turn** and then took 38 items in three hours.
This is the shape `HANDOFF_2026-08-16` §7 already warns about ("a queue depth is not a
prediction"). The right question is not "why is each item slow" — median handler runtime
is **36 seconds** — it is **"why does a site wait hours for a turn"**. §2 and §3 answer that.

**Handler runtime, all sites, 7 days, n=6461 completes:** p50 **36 s**, p90 **121 s**,
max 2,637,367 s. The mean (1272 s) is worthless here — one stuck row dominates it. Quote
the p50/p90.

---

## 2. The binding constraint: the fleet dispatches ONE site at a time

`scheduled_tasks` row `build-pipeline-trigger`: `interval_seconds=60`,
`concurrency_group='dispatch'`, **`max_concurrent=8`**, `timeout_seconds=300`, enabled.

**`max_concurrent: 8` is dead config.** Two independent reasons, both in
`cmd/scheduler/main.go`:

1. **The group cap counts TASK ROWS, not executions.** `countInFlight` (`main.go:414`)
   is `SELECT concurrency_group, COUNT(*) FROM scheduled_tasks WHERE … GROUP BY 1`, and
   the cap test is `inFlight[group] >= task.MaxConcurrent` (`main.go:181`). The `dispatch`
   group has **one enabled row** (`intent-collection` and `improvement-sweep` are
   `enabled=f`), so group occupancy can only ever be 0 or 1. A cap of 8 against a
   population of 1 can never bind.
2. **A per-task single-flight guard.** `loadDueTasks` (`main.go:368`) will not re-fire a
   `fire_message=true` task until `last_completed_at >= last_triggered_at` **or**
   `last_triggered_at + timeout_seconds <= NOW()`. One execution of the row at a time,
   full stop.

**Measured, at the artefact:** `build-pipeline-trigger` orchestrations, per minute,
25-minute window — **17 runs, never 2 in the same minute**, with gaps at 15:41, 15:45,
15:48, 15:51, 15:54, 15:57, 16:00, 16:03. Over 6 hours: 152 productive runs
(`current_step='complete'`, **avg 218 s**), 64 idle runs (`complete_idle`, avg 1 s),
2 `AWAITING_RESPONSES`, 2 `FAILED`.

So the real tick is **~218 seconds, not the configured 60**, because the guard holds the
next fire until the run finishes.

**What one run does:** `find_dispatchable_site` returns **one** site (`… ORDER BY
wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`, migration 284), then
`build-dispatch-loop` loads **≤5** items (`load_items.max_items: 5`) and processes them
**serially** (`process_item.max_iterations: 5`).

**Ceiling arithmetic:** 5 items / 218 s ≈ **1.4 items/min ≈ 83 items/hour, fleet-wide,
across all 21+ sites with work.** Measured fleet completions over the last 48 h: **50–180
per hour.** The arithmetic and the artefact agree.

### The control that makes this disconfirmable

`count(DISTINCT site_id)` of claims **per minute** over 6 hours reads **1** in nearly
every populated minute. If concurrent dispatch existed it would read >1. ⚠ **Do not use a
5-minute bucket** — that reads 2–6 distinct sites and proves nothing, because 5 ticks × 1
site each produces exactly that. I nearly recorded the 5-minute number as evidence of
concurrency; it is evidence of nothing.

**Live corroboration, last 90 minutes:** `webdesign.co.uk` took ~45 claims (all
`page_rerender`); `adversecreditmortgage.co.uk` — 41 `triaged` items, eligible the whole
time — took **zero**.

---

## 3. The second constraint: service order is strict fleet-wide FIFO by item age

Migration 284 (`sql_for_agents/284_dispatch_selects_oldest_waiting_site.sql`) made the
selector pick the site owning the **globally oldest eligible item**. That is
starvation-free by construction and it was the right fix for what it fixed. Its cost is
the thing the idea.uk lane felt: **a newly filed item queues behind every older eligible
item on every other site.** idea.uk filed its items at 15:34/15:41 on 08-17 — the newest
in the fleet — so it went to the back of ~150 older rows. 284 says so explicitly, and
records cross-site priority as deliberately unimplemented.

---

## 4. The levers, ranked by what closes the door

### L1 — make the trigger actually concurrent (biggest, ~8× ceiling)
- **(a) Code.** Count in-flight *executions* per task and honour `max_concurrent` per
  task, not per group-row-count. Needs a `cmd/scheduler` build + roll (its own binary,
  image `aqls/kafka-scheduler`, **not** a chassis roll). This adds authority to a shared
  fleet-critical seam → **council gate**, and per the 2026-08-02 §2 ruling it should ship
  as an **opt-in field with the unsafe default OFF**, not as a behaviour change to every
  task row.
- **(b) Config-only sibling rows** — add `build-pipeline-trigger-2…8` to the `dispatch`
  group, same target agent/topic/input. Each row gets its own single-flight guard, and
  `find_dispatchable_site`'s `NOT EXISTS (… status='claimed' … same site)` keeps per-site
  serialisation intact, so the per-site safety property is preserved. **⚠ BLOCKED AS-IS,
  and this is the trap:** both agents' `notify_scheduler` steps hardcode
  `UPDATE scheduled_tasks … WHERE name = 'build-pipeline-trigger'`. Sibling rows would
  never be stamped, so their guard would fall through to `timeout_seconds` and they would
  fire every **300 s**, not 60. Make the stamped name per-row config first, or measure
  what you actually get.
- Two executions can pick the same site in the same instant. The atomic claim handles it
  (`claim_work_item` → `check_claim` → `done`), so the cost is a wasted spawn, not a
  double-dispatch. Verify, don't assume.

### L2 — raise the per-turn batch (cheap, config-only, live immediately)
`build-dispatch-loop`: `load_items.max_items: 5` **and** `process_item.max_iterations: 5`
— both must move together. At p50 36 s/item, 5 items ≈ 3 min, which is the measured 218 s
average run. **Raising it to 10 pushes the run past the trigger's `timeout_seconds=300`**,
after which the scheduler re-fires while the run is still live. That overlap may be
exactly the parallelism you want or exactly `bugs_open/029`; establish which **before**
turning it up, and move `scheduled_tasks.timeout_seconds` in step. The other ceiling is
`call_dispatch.timeout_seconds: 900` in `build-pipeline-trigger`.

### L3 — cross-site fairness (owner decision, not a session's)
Round-robin over sites, or age-plus-priority, instead of strict FIFO. This is the direct
answer to "my site got nothing for 3 hours" and it is the one lever that changes per-site
**latency** rather than fleet **throughput**. Migration 284 states the shape and says an
aging scheme needs an owner-agreed scale constant. **Ask; do not pick a constant.**

### L4 — lowering `interval_seconds` below 60: don't bother
Runs average 218 s and the single-flight guard binds, so the interval is not the
constraint. Changing it would produce no measurable difference and would look like a fix.

---

## 5. For idea.uk specifically, today: there is nothing to drain

| status | count |
|---|---|
| triaged / approved / claimed | **0** |
| detected | 31 |
| deferred | 40 |
| needs_human_review | 37 |
| failed | 14 |

All 31 `detected` rows carry **`handler_agent = ''`** (`head_essentials_missing`,
`canonical_mismatch`, `dead_internal_link_live`, `image_url_404`, `placeholder_contact`).
The promoter's routability guard (`workItemRoutableSQL`, `work_items_common.go:321`)
refuses to promote a row the claim path would refuse — correctly, per `bugs_closed/284`.
Fleet-wide: **698 `detected` rows, 681 of them unroutable**; `head_essentials_missing`
alone is **606 rows across 21 sites**.

**That is `bugs_open/083`, and it is ACTIVELY OWNED** — `scripts/who-owns.py` puts it on
the `bugfix_277_required_fields_repair` lane (25 commits/14 d, council-approved rounds and
a `detected-item-promoter` shipped on 08-17). **Contribute into the bug file; do not
compete.** So no dispatch change will make idea.uk drain today: dispatch is not what is
holding those 31 rows.

---

## 6. Read these before asserting a cause

- **`bugs_open/029`** (hung spawns saturate the `dispatch` group, OPEN) states the
  mechanism as *"the pool is 8 … six dead build-\* orchestrations plus live traffic is
  enough to close it"*. **That does not match what `countInFlight` reads today** — it
  counts `scheduled_tasks` rows, and a hung *orchestration* is not one. The observed
  consequence (builds halt fleet-wide) is the same, but today's route to it is the
  **per-task** guard: a run whose `last_completed_at` never advances blocks the next fire
  for `timeout_seconds`. **[INFERRED — I have not read the 2026-07-19 version of
  `main.go`; the 048 fix landed 07-21, after 029 was filed.]** Establish which before
  building on 029: the two mechanisms have different fixes.
- **A live `needs_diagnosis`, filed today 12:14** (status `failed`), describes a
  `build-dispatch-loop` freezing at `process_item_iter_N_spawn_handler` after its parent's
  final retry replay. Under single-flight that is not one site's problem — one wedged loop
  costs the **whole fleet** up to 300 s. Same family; check it before filing anything new.
- **The 2026-07-31 owner ruling applies here.** "The fleet dispatches one site at a time"
  is a cross-cutting structural claim about shared infrastructure. Run
  `090_TRIGGER_needs_diagnosis_v1.sh` before writing it into a `bugs_open/` file, or state
  plainly what first-hand verification you substituted.

---

## 7. Recipes — measure at the artefact, before and after

```sql
-- (1) Is the trigger single-flight? Expect at most ONE run per minute.
SELECT date_trunc('minute',created_at) m, count(*) runs,
       round(avg(EXTRACT(epoch FROM updated_at-created_at)),1) avg_s
FROM orchestration_states
WHERE workflow_plan->'steps' ? 'find_dispatchable_site'
  AND created_at > now()-interval '25 minutes'
GROUP BY 1 ORDER BY 1 DESC;

-- (2) How many sites are served at once? PER MINUTE, never per 5 minutes (§2).
SELECT date_trunc('minute',claimed_at) m, count(*) claims, count(DISTINCT site_id) sites
FROM site_work_items WHERE claimed_at > now()-interval '6 hours'
GROUP BY 1 ORDER BY 1 DESC LIMIT 30;

-- (3) Fleet throughput, and how concentrated it is.
SELECT date_trunc('hour', completed_at) hr, count(*) done, count(DISTINCT site_id) sites
FROM site_work_items WHERE completed_at > now()-interval '48 hours' AND status='complete'
GROUP BY 1 ORDER BY 1 DESC;

-- (4) Per-site latency -- the number L3 is meant to move.
SELECT s.domain, count(*) n,
       round(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM wi.claimed_at-wi.triaged_at))/60) p50_wait_min,
       round(percentile_cont(0.9) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM wi.claimed_at-wi.triaged_at))/60) p90_wait_min
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE wi.claimed_at > now()-interval '7 days' AND wi.triaged_at IS NOT NULL
GROUP BY 1 ORDER BY 3 DESC;

-- (5) Handler runtime -- quote p50/p90, NEVER the mean (one stuck row owns it).
SELECT count(*) n,
       round(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM completed_at-claimed_at))) p50_s,
       round(percentile_cont(0.9) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM completed_at-claimed_at))) p90_s
FROM site_work_items
WHERE status='complete' AND completed_at > now()-interval '7 days' AND claimed_at IS NOT NULL;
```

**Config to read before changing anything** (live rows, not the seed files):
```sql
SELECT name, interval_seconds, concurrency_group, max_concurrent, timeout_seconds, enabled
FROM scheduled_tasks WHERE concurrency_group='dispatch';

SELECT default_config #> '{workflow,steps,load_items,config}'   AS load_items,
       default_config #> '{workflow,steps,process_item,config,max_iterations}' AS max_iters
FROM agent_definitions WHERE type='build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Code to read:** `cmd/scheduler/main.go` — `loadDueTasks`, `countInFlight`, and the cap
test around `main.go:181`. `platform/orchestration/actions/load_work_item_actions.go:707`
(the dispatchable predicate) and `work_items_common.go:321` (routability).
