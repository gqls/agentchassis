# RUNBOOK — parked findings release

Every command here was needed to get something right on 2026-09-04. Gotchas are attached to the
command, not kept separately.

`PSQL` below means:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## 1. Census — the population, split the way that matters

⚠ **`deferred` means two different things.** Empty `handler_agent` = the record-mode / roadmap
shelf (this lane). **Named** `handler_agent` = `bugs_open/396`'s park (a different lane, 60 of its
rows carry another lane's live release condition). **Any query that does not split on
`handler_agent` is answering a different question from the one it looks like it is answering.**

```sql
SELECT CASE WHEN COALESCE(handler_agent,'')='' THEN 'shelf (this lane)' ELSE 'named handler (396)' END,
       COALESCE(spec->>'filing_mode','(none)'), (spec ? 'release_recipe'), count(*)
FROM site_work_items WHERE status='deferred' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

Per site:
```sql
SELECT s.domain, count(*) AS parked, min(w.created_at)::date AS oldest, max(w.created_at)::date AS newest
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.status='deferred' AND w.spec->>'filing_mode'='record'
GROUP BY 1 ORDER BY 2 DESC;
```

## 2. THE CHECK TO RUN BEFORE ANY RELEASE — how many rows door 5 will swallow

```sql
SELECT COALESCE(spec->>'origin','(unstamped)') AS origin, count(*)
FROM site_work_items WHERE status='deferred' AND spec->>'filing_mode'='record' GROUP BY 1;
```
`model_opinion` rows are **held permanently at `detected`** by `detected-item-promoter`'s door 5.
Releasing them with the row's own documented recipe changes the status and dispatches nothing.

Confirm the door is still live and verbatim (config drifts; this is a `scheduled_tasks` row, not code):
```sql
SELECT name, enabled, (pre_query LIKE '%origin_ok%') AS door5_live FROM scheduled_tasks
WHERE name='detected-item-promoter';
```
⚠ `scheduled_tasks` has **no `schedule` column** — the tick is `interval_seconds` (900 for the
promoter, 30 for `build-pipeline-trigger`), and no `last_run_at` — it is `last_triggered_at` /
`last_completed_at`.

## 3. Simulate all five doors before releasing anything

This is the "what would actually flow" query. Run it per wave — the answer moves as pairs succeed
and fail.

```sql
WITH parked AS (
  SELECT id, item_type, pipeline, spec->>'routed_handler' AS h,
         COALESCE(spec->>'origin','')='model_opinion' AS is_opinion
  FROM site_work_items WHERE status='deferred' AND spec->>'filing_mode'='record'
    AND spec->>'routed_handler' IS NOT NULL
), hist AS (
  SELECT item_type, handler_agent,
         count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
         count(*) FILTER (WHERE status='failed') AS f
  FROM (SELECT item_type,handler_agent,status FROM site_work_items
        UNION ALL SELECT item_type,handler_agent,status FROM site_work_items_archive) x
  GROUP BY 1,2
)
SELECT p.item_type, p.h AS handler, count(*) AS parked,
       bool_and(p.pipeline IN ('build','content','design')) AS pipe_ok,
       COALESCE(max(hi.c),0) AS ever_good, COALESCE(max(hi.f),0) AS ever_failed,
       COALESCE(max(hi.c),0) > 0 AS known_good,
       bool_and(NOT p.is_opinion) AS origin_ok
FROM parked p LEFT JOIN hist hi ON hi.item_type=p.item_type AND hi.handler_agent=p.h
GROUP BY 1,2 ORDER BY 3 DESC;
```
⚠ **`known_good` and `floor_ok` read `site_work_items` UNION `site_work_items_archive`.** The live
table is a rolling window — a pair's history is mostly in the archive, so a check that reads only
the live table under-counts and will tell you a proven pair is unproven.

## 4. Confirm door 5 empirically (migration `629`'s own direction-1 recipe)

Safe *because* the row is designed to be held — a held row cannot dispatch. **Not yet run.**

1. Insert ONE synthetic row: `status='detected'`, a pair from §3 with `known_good`, a named live
   handler, `spec.origin='model_opinion'`, and a summary saying what it is and who made it.
2. Wait **≥ 2 promoter ticks (≥ 1,800 s)**.
3. Assert it is **still `detected`**. In the same window assert ≥ 1 *natural* promotion happened
   (any unstamped row promoted, `triaged_at` in the window) — otherwise you have proved only that
   the promoter did not run.
4. Close it by hand: `status='cancelled'`, `result` noting it was the verification row.

⚠ **Never a synthetic *promotable* control** — it would dispatch real work at a real site.

## 5. Release (per row, the documented recipe)

```sql
UPDATE site_work_items SET status = spec->>'routed_status', handler_agent = spec->>'routed_handler',
       updated_at = now()
 WHERE id = <id> AND status = 'deferred' AND spec->>'filing_mode' = 'record';
```
⚠ Guarded on `filing_mode='record'` deliberately: a **rule-3b** park (`capability_gap`, no
`filing_mode`, no `routed_handler`) must not be released — no handler can do it.
⚠ **The recipe leaves `created_at` alone**, so a released row keeps its August date and sorts to
the **FRONT** of `find_dispatchable_site`'s `ORDER BY MIN(created_at) ASC` — ahead of today's work.

## 6. Watch a release through

```sql
-- promotion (900 s ticks, LIMIT 20 candidates per tick ⇒ 80/hour ceiling)
SELECT status, count(*) FROM site_work_items WHERE id = ANY(<ids>) GROUP BY 1;
-- what the doors refused, and why — the promoter writes its held reasons
SELECT name, last_triggered_at, last_completed_at FROM scheduled_tasks WHERE name='detected-item-promoter';
```
⚠ **A `complete` work item is not a repaired artefact.** Check the page. `output_tokens ==
max_tokens` means the completion was CUT and a fragment may have been persisted as success.

## 7. Fleet state before releasing into it

```sql
SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
 WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
```
⚠ **Filter by `pod_name`, not the `service` column** (other pods share the image). ⚠ It is a
**two-hour window**, not a history — it lists survivors, so cross-check against
`kubectl -n ai-persona-system get pods -l app=agent-chassis` before concluding a roll is mid-flight.
That is exactly how the "roll has not landed" hold turned out to be already cleared.
⚠ **No dispatch within ~300 s of a chassis restart** — the spawn is silently dropped.
