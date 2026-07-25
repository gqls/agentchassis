# 021 FEATURE — a paved road for operator-driven bulk page rebuilds

**Raised:** 2026-07-25, from the fundamentallyai.com stage-2 rollout: landing new
components + a new copy voice across five pages needed the pages rebuilt, and
there is no first-class way to ask for that. Hand-driving it collided with two
immune-system mechanisms tuned to stop runaway loops, which treat a deliberate
rebuild as recurrence and quietly park it.
**Status:** specified, not built. **Depends on:** `bugs_open/070` (the reaper
predicate — that fix is a prerequisite, not part of this feature).

## The gap

An operator with a legitimate reason to rebuild a set of pages — a new component
placed in the site plan, a rewritten writer prompt, a corrected evidence base —
has no supported operation for it. What we actually did, five times:

```sql
UPDATE pages SET build_status='needs_rebuild' WHERE site_id=$1 AND name=$2;
UPDATE site_work_items SET status='triaged' WHERE id=$3;  -- an OLD row, reused
```

That second statement is the problem. There is no "rebuild this page" item type,
so the operator reaches for whatever historic row is closest — here
`needs_page:index`, created 2026-07-20 and long since satisfied. Consequences,
all observed:

- **The row's own description becomes a lie.** `needs_page:capabilities` still
  reads `Build capabilities page (not_built)` while the page is deployed. Anyone
  reading the queue sees a build that never happened rather than a rebuild that
  was asked for.
- **It races the stale-item reaper and usually loses** (`bugs_open/070`).
  `stale-work-item-reaper` (live, hourly) parks any `triaged` build item whose
  **`created_at`** is 48h+ old and that has not been claimed. A re-queued 2026-07-20
  row is *born* stale, so the operator's minutes-old request is eligible for
  reaping immediately. It survives only if `build-pipeline-trigger` (every 120s)
  claims it first.
- **Batches lose that race systematically.** Queue four pages at once and the
  first is claimed within ~2 minutes; the rest wait behind it [INFERRED: the
  builder appears to be single-flight per site — the trigger's own pre-query only
  filters `sites.locked_at IS NULL`, so the serialisation is downstream and I did
  not trace it]. A page build takes tens of minutes, so siblings sit exposed for
  longer than the reaper's hour and get parked. Sequential queueing works
  because each item is exposed for ~2 minutes. This is the whole of the
  "batch re-queues get parked" behaviour recorded in this workstream's NOTES.
- **The summary is mutated in place, cumulatively.** Three rows on this site now
  carry `[stale: triaged 48h+] [stale: triaged 48h+] Build … (not_built)` —
  prefixed once per reaping, permanently.

## What already exists (and is dormant)

Most of the road is built. **Do not build a second one.** Verified live
2026-07-25:

| piece | state |
|---|---|
| `maintenance_queue` table, with `task_type`, `payload`, `reason`, `requested_by` | exists; **2 rows ever, newest 2026-02-18** |
| `maintenance-triage` agent (`scan_and_queue` → `check_dry_run` → `prepare_rebuild_dispatches` → `check_has_rebuilds` → `spawn_rebuilder` → `rebuild_loop`) | `is_active=true`, full workflow present |
| `PrepareRebuildDispatchesAction` — claims `page_rebuild` tasks, flags pages `needs_rebuild`, groups dispatches one-per-site (`platform/orchestration/actions/maintenance_actions.go:426`, `flagPagesForRebuild` at `:960`) | code present |
| a `scheduled_tasks` row that fires `maintenance-triage` | **none** — nothing drives it |

So `page_rebuild` was designed as a maintenance-queue task type with an explicit
`requested_by` field, wired end to end, and then left without a trigger. Its
scan side looks for `stale_pages`/`missing_content`/`orphan_nav` on a 30-day
threshold — a cadence concern, not an operator one, which is probably why no
operator path was ever added.

## What this feature is

1. **An operator entry point** — a trigger script beside the other
   `docs/.../fixloop_eg_dartsonline/0NN_*` scripts that inserts one
   `maintenance_queue` row per requested page:
   `task_type='page_rebuild'`, `requested_by='<session/operator>'`,
   `reason='<why>'`, `payload={"pages":[...],"intent":"recompose|rerender"}`,
   then dispatches `maintenance-triage` directly rather than waiting on a
   cadence. Prints the correlation id so the run is findable by payload (the
   council-gate lesson: never look it up by the printed id alone).
2. **Sequencing that respects the builder's own limits.** One page in flight per
   site, next queued on completion of the last — what the ad-hoc driver in this
   workstream ended up doing by hand. This belongs in `rebuild_loop`, which
   already exists as a step, so the operator asks for five pages and gets five
   sequential builds rather than one build and four parked rows.
3. **A rebuild item that says it is a rebuild.** Whether that is a new
   `item_type='page_rebuild'` or a re-used `page_rerender` with an explicit
   intent field, the row must be **freshly created** (never a resurrected
   historic row) and its summary must name the trigger — "Rebuild capabilities:
   new hero-card-carousel in plan + writer-voice v2". This is what makes
   `bugs_open/070` a prerequisite rather than a blocker: a fresh row is not
   born stale, so the reaper is correct to leave it alone.
4. **Intent: recompose vs re-render.** `features_open/012` (explicit
   `recompose_pages` intent, LIVE v1.0.1149) already draws this distinction for
   re-planning; a bulk rebuild needs the same switch. "Fill the plan's components
   with fresh content" and "re-render existing content through fixed templates"
   are different operations, and the operator knows which one they want.
   Reuse 012's vocabulary rather than inventing a third.
5. **Dry-run first.** `maintenance-triage` already has a `check_dry_run` branch.
   The operator path should default to reporting *which pages would rebuild and
   why* — the same discipline `bugs_open/044`'s inventory report kept
   (`dry_run` MUST stay).

## Why it is worth building

The workaround is not merely awkward, it is **silently lossy**: a parked
`unresolved` row looks like a considered decision, not a dropped request. Every
operator who rebuilds pages by hand rediscovers this, and the evidence they
leave behind (a mutated summary on a row about something else) actively
misleads the next reader. Meanwhile the queue mechanism designed for exactly
this has sat unused for five months.

It also unblocks the rest of the site-quality automation set: `016`
(brief-fidelity audit), `017` (component adoption) and `018` (design critic) all
end in "…and then the affected pages get rebuilt". Without a paved road, each
one either grows its own dispatch code or hands the operator the same broken
manual recipe.

## Grounding (queries re-run 2026-07-25, re-run again at build time)

```sql
-- the reaper, live, keying on row age
SELECT pre_query FROM scheduled_tasks WHERE name='stale-work-item-reaper';
-- the claim cadence it races
SELECT interval_seconds FROM scheduled_tasks WHERE name='build-pipeline-trigger';  -- 120
-- the dormant road
SELECT task_type, status, count(*), max(created_at) FROM maintenance_queue GROUP BY 1,2;
SELECT type, is_active FROM agent_definitions
 WHERE type='maintenance-triage' AND deleted_at IS NULL;
SELECT name FROM scheduled_tasks WHERE target_agent_type='maintenance-triage';  -- 0 rows
-- the damage pattern
SELECT status, attempt_count, summary FROM site_work_items
 WHERE item_type='needs_page' AND summary LIKE '[stale%';
```

## Relates to

- `bugs_open/070` — the reaper predicate. **Fix first**; this feature assumes a
  freshly-created row is safe from reaping.
- `features_open/012` — explicit recompose intent (LIVE): the intent vocabulary.
- `bugs_open/044` — dormant-agents inventory: `maintenance-triage` is a live
  instance of exactly what that report looks for (active definition, no trigger,
  no runs). Worth checking whether 044's report already flags it.
- `bugs_closed/048` — the same reaper starving its concurrency group. Different
  defect, same task; read it before touching the row.
- `bugs_open/029`/`030` — dispatch queue serialisation. The single-flight
  behaviour inferred above may be the same mechanism; do not assume, trace it.
