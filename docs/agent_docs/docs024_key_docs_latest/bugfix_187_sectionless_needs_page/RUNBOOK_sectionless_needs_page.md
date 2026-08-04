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

## Grep an IMAGE before deploying it (caught v1.0.1247 shipping without the fix)

```bash
docker run --rm --entrypoint sh aqls/agent-chassis:<tag> -c \
  'cd /tmp && strings /app/agent-chassis > s.txt && printf "pos=%s skip=%s neg=%s\n" \
   $(grep -c declaredPageSections s.txt) $(grep -c skipped_sectionless_page s.txt) \
   $(grep -c toolPageDeclaredSections s.txt)'
# healthy fix image: pos=5 skip=3 neg=0 ; any pre-fix image: neg>0
```
Gotchas:
- The container root fs is read-only for sh — write the strings dump under the
  container's /tmp, not /.
- An image's CreatedAt being AFTER your commit proves nothing (v1.0.1247:
  built 08:55, commit 00:30, fix absent — a pinned/stale ref builds late and
  lies by recency). bugs_open/153's rule, at the image instead of the pod.
- Deploy gate: `make deploy-agent-chassis` refuses a tag not in the registry;
  push first with a DIRECT `docker push docker.io/aqls/agent-chassis:<tag>` —
  `push-backend` retags 13 other services' stale local images (177 lane
  lesson).
- Before `kubectl apply -k` on the chassis overlay: check for in-flight
  council rounds (a roll kills them):
  `SELECT count(*) FROM orchestration_states WHERE status IN
   ('EXECUTING_STEP','AWAITING_RESPONSES') AND updated_at > now() - interval
   '15 minutes' AND (current_step LIKE 'review%' OR current_step LIKE
   '%council%');`
