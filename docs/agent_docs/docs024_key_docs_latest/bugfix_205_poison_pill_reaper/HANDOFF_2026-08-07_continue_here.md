# HANDOFF 2026-08-07 — bug 205: COLD-START for the next session

> **SUPERSEDED 2026-08-08 — the four decisions below were RULED AND EXECUTED the
> same day (NOTES tail, commit `b6e70cd70`). NOTHING REMAINS on this lane.**
> Cap 8000 live and proven (the poisoned record verified on its FIRST attempt,
> `output_tokens=3135`); the 32 cancelled; per-type ceilings + the shared park
> function live via migration `sql_for_agents/335` (`reaper_policies` +
> `business_intel.reap_stale_collection_tasks()`, RFC_018, register SCH-024).
> Re-verified live 2026-08-08 ~15:30Z by a later session: collection_tasks =
> 606 cancelled / 2528 completed, zero pending/in_progress/failed; reaper stamp
> advancing; live pre_query calls the function. Sole watch-item: the framework
> WARN names whichever of the 7 remaining uncapped steps runs first — size that
> step when it fires. One fresh trap paid for on re-verification:
> **`reaper_policies` lives in schema `public`, NOT `business_intel`** — a
> schema-qualified probe reads "relation does not exist" and looks like the
> migration never applied; query it unqualified, as the LANDMINES check spells
> it. The traps section below remains valid.

**State: fixed AND live in full. Nothing is burning. What remains is executing
whichever of the four owner decisions land.** Read PLAN → NOTES tail → this file.

## What is live and proven (do not re-prove unless something looks wrong)

- **Reaper parking (config)**: `stale-orchestration-reaper.pre_query` counts each
  stale-claim reset in `collection_tasks.retry_count`, backs off via
  `scheduled_for`, parks at the 5th as `status='failed'` naming bugs_open/205.
  Applied 2026-08-07 01:26Z; **behaviourally proven** — 33 seeded loopers parked
  in one pass at 01:40:48Z; still parked and pipeline quiet at 08:17Z.
  Backup + mechanically-generated ROLLBACK sit beside this file.
- **Go halves, pod-proven on `v1.0.1262` (both replicas, 08:17Z)**:
  `ensure_collection_tasks.go` blocks re-tasking a `'failed'` business;
  `ai_actions.go` WARNs (`max_tokens not configured at any level`) when a step
  falls to the hardcoded 2048 (`anthropic.go:109`). WARN unfired so far —
  correct: the only ACTIVE uncapped step is the parked verifier.
- **Council APPROVED round 2**, corr `2db88f8f-11ea-47ed-b37d-35a6096d5597`.
  Commits: `d1eb3a6b5` (Go+register+landmine), `8ab75a9fb`, `2c3041f7d`,
  `d3184d6bd`. Concept register SCH-014 corrected; LANDMINES entry synced.

## The four decisions (plain-prose versions in README_where_we_are)

1. **Step cap for `vet-practice-verifier/extract_and_reconcile` (~8000).**
   If YES: set `default_config->workflow->steps->extract_and_reconcile->config->
   ai_service->max_tokens = 8000` on the live `agent_definitions` row (nested
   path — see the LANDMINES entry on `config.ai_service.max_tokens` vs
   `config.max_tokens`; assert the row count). Then un-park ONLY task
   `ea489aed-d770-4962-8866-b0313d5a0dc0` (RUNBOOK has the un-park UPDATE) and
   watch one cycle: expect a batch within 5 min, verifier SUCCESS, task
   `completed`. bugs_closed/067 says sweep the agent's OTHER steps' caps while
   in there.
2. **The 32 scrape-failing parked tasks.** Un-park (bounded — they re-park after
   5), investigate the scrape errors first (samples in NOTES: external API
   refusals / "URL failed to load"), or cancel (574-row precedent). No cost
   while parked.
3. **Per-task_type park ceiling** — only if the owner wants it; today the table
   is single-task_type (censused 2026-08-07). Would be a small pre_query change.
4. **Shared reaper-accounting mechanism** (architecture seat's note; 2nd/3rd
   hand-rolled instance). RFC-shaped, architecture track, not urgent.

## Verification snippets (fresh session)

```sql
-- parked set intact?
SELECT status, count(*) FROM business_intel.collection_tasks GROUP BY 1;
-- loop still dead? (absence of runs + presence of advancing stamps = parked-quiet)
SELECT count(*) FROM orchestration_states WHERE owner_agent_type='vet-practice-verifier'
  AND created_at > now() - interval '6 hours';
SELECT name, last_triggered_at FROM scheduled_tasks
 WHERE name IN ('vet-batch-verify','stale-orchestration-reaper');
```

## Traps already paid for (do not re-pay)

- `collection_tasks.started_at` is naive UTC; check `SELECT now()` before calling
  a row stale (WRONG_CALLS 2026-08-07).
- The cap lives at `config.ai_service.max_tokens`; `config.max_tokens` reads NULL
  for every row and a wrong-path `jsonb_set` is a silent no-op (LANDMINES).
- A council config-change edit needs a repo-relative `file` (the apply script),
  target described in the rationale — prose paths fail validation client-side.
- A landmine synced mid-task reads as prior art AGAINST your own submission —
  date-stamp premise claims, or sync after the round.
- Count truncations from `error_message ILIKE '%stop_reason=max_tokens%'`, never
  from `output_tokens` (NULL on truncated rows).
