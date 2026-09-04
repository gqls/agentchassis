# RUNBOOK — farmerinsurance.uk

Site id: `99cae989-2413-430d-b026-59dfeeb638c0`. DB:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## 1. Serve-side census, with the control that makes it informative
**Always fetch the invented path in the same breath.** A parked/marketplace domain 200s every
path, and a census with no 404 control has reversed an architectural conclusion before
(LANDMINES). Farmer is NOT parked — verified 2026-09-04 — but re-prove it each session:

```bash
curl -sS -o /dev/null -w "root %{http_code}\n"     https://farmerinsurance.uk/
curl -sS -o /dev/null -w "control %{http_code}\n"  https://farmerinsurance.uk/this-path-does-not-exist-9f3a.html   # must be 404
```

Page list to drive a crawl from (the DB is the source of truth for what SHOULD serve):
```sql
SELECT url, page_type, build_status, deployed_at
FROM pages WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND status='active' ORDER BY url;
```
Crawler used on 2026-09-04 (fetch each page, collect every `href`/`src` on the same host, then
status-check each target): scratchpad `farmer/crawl.py`. ⚠ **Use `curl -o /dev/null` for the
link check, not a text read** — an image target makes a text-mode `subprocess.run` throw
`UnicodeDecodeError` and abort the whole census mid-way (hit this first time).

## 2. The queue, and how to read it without being misled
```sql
SELECT item_type, status, count(*), min(created_at)::timestamp(0) AS oldest
FROM site_work_items
WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0'
  AND status NOT IN ('complete','cancelled','rejected','failed')
GROUP BY 1,2 ORDER BY 3 DESC;
```
⚠ **A count here is not a count of live defects.** Measured 2026-09-04: 58% of the open rows
were false (their cause was fixed after filing) or moot (the page they judge is archived).
Two joins that separate the three cases in one query:

```sql
-- which PAGE a row judges, and whether that page is still live
SELECT swi.item_type, swi.status, p.url, p.status AS page_status, p.build_status
FROM site_work_items swi LEFT JOIN pages p ON p.id = swi.page_id
WHERE swi.site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND swi.status='detected'
ORDER BY swi.item_type, p.status, p.url;

-- unbuilt_internal_link rows carry their TARGET in the item_key's 5th field
SELECT split_part(item_key,':',5) AS target_href, count(*)
FROM site_work_items
WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0'
  AND item_type='unbuilt_internal_link' AND status IN ('unresolved','failed')
GROUP BY 1 ORDER BY 2 DESC;   -- 2026-09-04: 104 rows, ONE target (/claims.html), all false
```

## 3. Specs — read the live row, never a seed file
```sql
SELECT aspect, created_by, created_at::timestamp(0) FROM site_specs
WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND is_current ORDER BY aspect;

SELECT jsonb_pretty(data) FROM site_specs
WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND is_current AND aspect='classification';
```
⚠ The column is `data`, not `content`. ⚠ `classification` is REWRITTEN by two enrichment agents
(`evaluate_news_feed`, `evaluate_directory_features`) on most build passes — its history has 15+
superseded rows. The `content_features` block is where both write; the rest is the
domain-research-classifier's.

## 4. Which sites carry a given directory kind (the fleet control for §3's defect)
```sql
SELECT s.domain, ss.data->'content_features'->'health_insurer_directory'->>'kind'
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE ss.is_current AND ss.aspect='classification'
  AND ss.data->'content_features' ? 'health_insurer_directory';
```
Swap the key for `mortgage_lender_directory` / `savings_provider_directory` to census the other
kinds. Map of signal → kind: `verticalDirectoryMap` in
`platform/orchestration/actions/feed_directory_recommendation_action.go`.

## 5. What is parked for the owner
```sql
SELECT p.url, swi.item_type, swi.created_at::timestamp(0), left(swi.summary,80)
FROM site_work_items swi LEFT JOIN pages p ON p.id=swi.page_id
WHERE swi.site_id='99cae989-2413-430d-b026-59dfeeb638c0'
  AND swi.status='needs_human_review' ORDER BY swi.item_type, p.url;
```

## 6. Peer lanes to contribute to rather than compete with
`loanzy_uk_example_site` (route, council, growth posture) · `copy_quality_two_stage` (the 14
proposals) · `bugfix_316_news_feed_ordering` (news region) · `bugfix_206_directory_build_handler`
(entity directories) · `staged_component_build` + offer-analysis (carousels) · `lendzy_co_uk`
(evidence-register method). Message by session name via SendMessage; several are live.
