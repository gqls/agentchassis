# RUNBOOK — the work-item terminal-write contract (bugs_open/307)

Every command here was hard to get right once. The gotcha is attached to each.

## The config census — the nested walk, and why the obvious one returns nothing

`steps` is a JSON **object keyed by step name**, nested under `workflow`. A top-level
`jsonb_each` undercounts by **100%** (not partially — zero rows), and
`jsonb_path_query(...,'$.**.steps')` piped into `jsonb_array_elements` also returns zero,
because the result is an object, not an array. Both mistakes are in `WRONG_CALLS.md`.

```sql
WITH live AS (SELECT type, default_config FROM agent_definitions
              WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
arrs AS (SELECT l.type, arr FROM live l, LATERAL jsonb_path_query(l.default_config,'$.**.steps') arr),
walked AS (SELECT type, e.key AS step_name, e.value AS step
           FROM arrs, LATERAL jsonb_each(arr) e WHERE jsonb_typeof(arr)='object'
           UNION ALL
           SELECT type, COALESCE(s->>'name','(arr)'), s
           FROM arrs, LATERAL jsonb_array_elements(arr) s WHERE jsonb_typeof(arr)='array')
SELECT type, step_name, step->>'action' AS action, step->'config'->>'status' AS status
FROM walked WHERE step->>'action' IN ('update_work_item_status','fail_work_item')
ORDER BY action, type, step_name;
```
Keep the array arm even though it returns 0 rows today: without it, a future array-shaped
`steps` would be silently invisible.

## The failure population — and the three ways to miscount it

```sql
-- died BEFORE exhausting its budget (the defect's own signature)
SELECT count(*) FILTER (WHERE attempt_count < max_attempts) AS died_early, count(*) AS all_failed
FROM site_work_items WHERE status='failed' AND updated_at >= now()-interval '14 days';
```
- Filter on **`updated_at`**, never `completed_at`: `failed` rows carry **no**
  `completed_at` at all (LANDMINES) — a `completed_at` filter returns zero and reads as
  "no failures".
- `site_work_items` is a **~7-day window**; the `work-item-archiver` moves terminal rows
  to `site_work_items_archive`. Any *lifetime* claim must `UNION ALL` the archive.
- There are **two** terminal success statuses (`complete` **and** `verified`) — always
  `GROUP BY status` before filtering on it.

## Proving the guard and the ladder without touching production

`FailWorkItemAction` had **no test at all** before this work, so there is no fixture to
copy. The tests use sqlmock. Two traps:
- `update_work_item_status`'s existing tests pin a **4-argument** `ExpectExec`. sqlmock
  matches on argument count, so adding a bound parameter to that UPDATE fails them — they
  are updated in the same commit, deliberately, not worked around.
- A mock's own bookkeeping cannot prove a negative. To show the guard really guards,
  **mutate the guard list and watch the test fail** (`MEMORY: mutate-the-code-to-prove-the-guard`).

## Reading a `scheduled_tasks.pre_query` without consuming its rotation

Running one by hand **advances the rotation** (LANDMINES). Always:
```sql
BEGIN; <the pre_query>; ROLLBACK;
```
And to see which tasks touch our table at all — read the `pre_query`, not the `ILIKE`:
```sql
SELECT name, enabled, interval_seconds FROM scheduled_tasks
WHERE pre_query ILIKE '%site_work_items%' ORDER BY enabled DESC, name;
```
Five of them UPDATE `site_work_items.status`: `claimed-item-timeout` (120s),
`feasibility-recheck` (600s), `detected-item-promoter` (900s), `stale-work-item-reaper`
(3600s), `held-pair-canary-escalation`.

## The burst detector, run by hand against a past window

```sql
SELECT date_trunc('hour', occurred_at) AS h, count(*), count(DISTINCT domain) AS domains,
       count(DISTINCT agent_type) AS types,
       left(regexp_replace(regexp_replace(regexp_replace(error_message,
         '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}','<uuid>','g'),
         '"[^"]*"','<q>','g'), '[0-9]+','<n>','g'), 120) AS sig
FROM agent_error_log WHERE occurred_at BETWEEN '2026-08-17 13:00Z' AND '2026-08-17 17:00Z'
GROUP BY 1, sig HAVING count(*) >= 10 ORDER BY 1;
```
Use **`domain`** (0% NULL), never `site_id` (89.8% NULL in this table), and never
`error_code` (`ClassifyError` labels the git 404 `LLM_API_ERROR`).

## Building and proving the deploy

```bash
make build-agent-chassis           # builds from committed HEAD, not the tree
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <our-commit> <the stamped sha>   # "did it ship?" is a query
```
An empty grep means "not in range" (it is a startup line and scrolls), not "unstamped".
Read the stamp **per service**, and bump `IMAGE_TAG` for every build or the node serves
its cached binary.

## Council submission (migrations are IN scope since 2026-08-19)

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json   # tests admission free
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
```
`plan` is an **object** (`summary`/`edits`/`grounded_in`/`risks`), `risks` is a **single
string**, `operation` ∈ `modify|add|remove|config_change` (`create` is refused), and
`edits[].file` is one repo-relative path. Find the run by payload, not by the printed id:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```
Budget ~30 minutes: the council takes 2–5, the dispatch queues behind the fleet.

## Post-roll acceptance poll (added 2026-08-21) — run BOTH halves together, or a zero lies

```sql
-- demand side: did any failure event occur since the fix went live?
SELECT count(*) FROM agent_error_log WHERE occurred_at >= '2026-08-20 16:09Z';
-- supply side: new-path evidence. ⚠ THE 341 CARVE-OUT IS LOAD-BEARING: the
-- claimed-item-timeout task (bugs_open/341) is the one remaining divergent writer and its
-- fair-weather terminal rows would otherwise read as a 307 regression.
SELECT id, item_type, status, attempt_count, max_attempts, retry_after, handled_by, left(error,120)
FROM site_work_items
WHERE (retry_after IS NOT NULL OR (updated_at >= '2026-08-20 16:09Z' AND status='failed'))
  AND (error IS NULL OR error NOT LIKE 'Claim timed out%');
```
Readings: zero+zero = quiet traffic, wait. Demand non-zero + supply zero terminal-failures =
the bleed has stopped (GOOD — that is the 08-21 morning reading: 288 vs 0). Any post-roll
`failed` at `attempt_count < max_attempts` outside the carve-out = **ladder bypass = FAIL,
stop and investigate**. ⚠ `attempt_count > 0` alone is NOT ladder evidence — parked/complete
`update_work_item_status` writes still increment it (§8 residual).

## The close canary (owner-authorised 2026-08-21) — three arms nothing natural demands

One synthetic row, real path end-to-end; precedents 299 (`pool-web-tech.internal`) and 302.
Insert on a pool site: `item_type='content_rewrite'`, `status='triaged'`,
`handler_agent='page-build-handler'`, `item_key='canary_307_close_20260820'`, `max_attempts=3`,
spec naming a NONEXISTENT page (the "not found" text matches no transient needle in either
classifier; one item cannot trip the burst conjunction — 1 domain, 1 agent type).

1. Attempt 1 fails → assert `triaged` / `attempt_count=1` / `retry_after ≈ +30m`
   (`reaper_policies` `__default__` backoff 30 — READ IT FIRST; a prediction that cannot miss
   proves nothing) / claim columns cleared. **Record the stamped value BEFORE shortening it.**
2. Shorten `retry_after` to `now()`; when the row is next claimed, immediately flip it to
   `wont_fix`. When the failure write lands: assert SKIP — status still `wont_fix`,
   `attempt_count` still 1, and the pod logs the line at `work_item_failure_ladder.go:412`
   (`skipped — a deliberate status is already recorded`). Race lost = the write landed first as
   attempt 2 with `retry_after ≈ +60m` — that is the scaling observation instead; retry the
   flip next cycle.
3. Flip back to `triaged`, shorten; attempt 2 → `2` / `≈ +60m`. Shorten; attempt 3 →
   **terminal `failed` at 3 of 3** (the §5(c) honesty case, live).
4. **Teardown:** DELETE the canary row and verify 0 (before the ~7-day archiver sees it, so the
   promoter's `content_rewrite`/page-build-handler ratio is not polluted); check the immune
   sweep filed nothing off the canary's failures (cancel + note if it did). `agent_error_log`
   rows remain — harmless. No dispatch within ~300s of a chassis pod (re)start.

Fallback if pool sites don't qualify for build dispatch (read the dispatcher predicate as
COLUMN TEXT only — executing a `pre_query` advances its rotation): hand-fire the handler
orchestration carrying `work_item_id` in `input_data` (the 313 lane's fire-script pattern);
exit 0 proves nothing — verify in `orchestration_states` by correlation.
