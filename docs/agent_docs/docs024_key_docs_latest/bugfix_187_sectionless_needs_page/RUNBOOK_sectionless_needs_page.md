# RUNBOOK — bugfix 187

Commands that were hard to get right, with the gotcha attached.

## The census (from the bug file; re-run before quoting any count)

```sql
SELECT source, item_type, status, count(*), max(created_at)::date AS newest
FROM site_work_items
WHERE error LIKE '%no sections ready to build%'
GROUP BY 1,2,3 ORDER BY 4 DESC;
```
Gotcha: this error-text match also catches `tool_content` (closed 177) and
`needs_content_page` rows — filter `item_type='needs_page'` for THIS bug's
population.

## Per-row triage — join the page BY NAME, not by page_id

```sql
WITH parked AS (
  SELECT w.*, split_part(w.item_key,':',2) AS page_name
  FROM site_work_items w
  WHERE w.error LIKE '%no sections ready to build%'
    AND w.item_type='needs_page' AND w.status='needs_human_review'
)
SELECT k.source, left(k.id::text,8), k.page_name, k.created_at::date,
       p.status, COALESCE(jsonb_array_length(p.sections),0) AS decl,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS slots,
       (SELECT count(*) FROM site_plan_sections sps
          JOIN site_plans pl ON pl.id=sps.plan_id AND pl.is_current
         WHERE pl.site_id=k.site_id AND sps.page_name=k.page_name) AS plan_rows
FROM parked k
JOIN pages p ON p.site_id=k.site_id AND p.name=k.page_name
ORDER BY k.source, k.created_at;
```
Gotchas, each one cost a wrong first query:
- **27/28 items carry `page_id` NULL** — a LEFT JOIN on `w.page_id` says "no
  page" while the page exists. Join `pages` on `(site_id, name)`.
- `site_plan_sections` has NO page_id — it keys `(plan_id, page_name)`; get
  the current plan via `site_plans.is_current`.
- `site_specs` has no `site_plan` column — the plan lives in the row WHERE
  `aspect='site_plan'`, payload in `data`.

## Revalidator coverage (the map, not the docs)

`platform/orchestration/actions/revalidate_review_queue_action.go:149` —
`reviewRevalidators` map. An item_type absent from the map = 'unknown' =
stamped and left. Do not trust any doc's claim about coverage; read the map.
