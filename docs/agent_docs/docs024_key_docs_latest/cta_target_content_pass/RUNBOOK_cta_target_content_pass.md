# RUNBOOK — cta_target_content_pass

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Population (the homogeneity census — re-run before each phase, the fleet moves)

```sql
WITH targets AS (
  SELECT s.domain, p.name AS page,
         COALESCE(pc.content_data->>'cta_url', pc.content_data->>'primary_cta_url') AS target
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  WHERE pc.slot_name IN ('hero','call-to-action') AND p.status='active'
    AND COALESCE(pc.content_data->>'cta_url', pc.content_data->>'primary_cta_url') LIKE '/%'
),
modal AS (
  SELECT domain, target, count(*) AS n,
         row_number() OVER (PARTITION BY domain ORDER BY count(*) DESC) AS rk
  FROM targets GROUP BY domain, target
)
SELECT t.domain, count(DISTINCT t.page) AS pages_total, m.target, m.n AS rows_on_modal
FROM targets t JOIN modal m ON m.domain=t.domain AND m.rk=1
GROUP BY t.domain, m.target, m.n HAVING m.n >= 6 ORDER BY m.n DESC;
```
Baseline 2026-08-15: 16 sites ≥6; finetuning.uk 39, aao 36, gaswholesalers 28.

## The two-step recipe (per site — see PLAN for why this shape)

1. Generate the tool list for guidance: `SELECT name, url, title FROM pages
   WHERE site_id='<site>' AND status='active' AND url LIKE '/tools/%';`
   (plus hub pages if wanted). Put it IN the `content_guidance` text.
2. `content_rewrite` items, `mode=edit_live` (load-bearing, bugs_open/178),
   labels-only guidance, one per target page.
3. After completes: `page_rerender` with `reason='cta_links_stale'` per page
   (label-match writes the url the new wording names). Verify as a matched
   pair (url keys + hrefs + prose sample) — the 268 lane's RUNBOOK has the
   invariant diff.

## Gotchas inherited from the 268 lane (all measured, none folklore)

- Dispatch serves the fleet's OLDEST eligible item (mig 284): a fresh item
  waits hours behind backlog. The 268 lane backdated its own items'
  `created_at` — if you do, SAY SO in NOTES (synthetic timestamps).
- A `failed` item is not failed work (`bugs_open/274`, ~15k instances) —
  verify at the row and the live page.
- The claims floor can refuse a rewrite whose section carries a banned
  claim — that is a guard; fix the copy first (268 D1 worked example).
- Label-match ties on incidental words (`bugs_closed/253`); self-link gap +
  double-target quirk (`bugs_open/248` tail, 2026-08-15).
- `content_guidance` reached the writer on page-build-handler items
  (proven 2026-08-15, 268 D1) — but `bugs_open/271` says some path has no
  reader for it; re-verify if using any other handler/item type.
