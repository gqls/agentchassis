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
