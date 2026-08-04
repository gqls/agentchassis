# RUNBOOK — bugfix_080_canonical_collisions

Commands that were hard to get right, with the gotcha attached. Update HERE when one changes.

## The collision census (the honest one)

Gotcha 1: 080's original survey filtered by `page_type` and missed its own instance — the
duplicate row is typed `section-index`, outside the filter. Key on URL shape / name stem instead.
Gotcha 2: do not use `build_status='deployed'` as liveness — `bugs_open/185` (28 shipped pages
carry another status). Compute has_shipped the way `datahelpers.NeverDeployedPagePredicate` does.

```sql
WITH u AS (
  SELECT p.id, s.domain, p.name, p.url, p.page_type, p.build_status, p.status,
         NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed') AS has_shipped,
         CASE WHEN p.url LIKE '%/index.html' THEN regexp_replace(p.url,'/index\.html$','')
              ELSE regexp_replace(p.url,'\.html$','') END AS path_key
  FROM pages p JOIN sites s ON s.id=p.site_id
)
SELECT domain, path_key, count(*) AS n,
       count(*) FILTER (WHERE status='active') AS active_n,
       string_agg(url||' ['||page_type||'/'||COALESCE(build_status,'-')||'/'||COALESCE(status,'-')||']',
                  E'\n    ' ORDER BY url) AS detail
FROM u GROUP BY domain, path_key HAVING count(*)>1 ORDER BY active_n DESC, 1,2;
```

2026-08-03 result: 6 groups; active_n=2 only for robot-hands `/news` and `/gripper-catalog`.

## Would canonicalisation re-key any existing row? (the surface-A safety proof)

`normaliseSlug` lowercases, spaces→dashes, strips directories and `.html`/`.htm`. A name it would
change matches:

```sql
SELECT s.domain, p.name, p.url FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.name ~ '[A-Z ]' OR p.name LIKE '%/%' OR p.name LIKE '%.html' OR p.name LIKE '%.htm';
-- 2026-08-03: 0 rows. Re-run after roll; still expect 0.
```

## Tool shape distribution (surface-B blast radius)

```sql
SELECT CASE
    WHEN name LIKE 'tool-%' AND url LIKE '/tools/%/index.html' THEN 'A canonical'
    WHEN name LIKE 'tool-%' AND url LIKE '/tools/tool-%' THEN 'B double-prefixed url'
    WHEN name NOT LIKE 'tool-%' AND url LIKE '/tools/%.html' THEN 'C deploy_tool shape'
    ELSE 'D other' END AS shape,
  count(*) AS rows,
  count(*) FILTER (WHERE NOT (deployed_at IS NULL AND COALESCE(build_status,'')<>'deployed')
                     AND status='active') AS live
FROM pages WHERE page_type='tool' GROUP BY 1 ORDER BY 1;
-- 2026-08-03: A=102/86, B=14/14, C=12/12, D=20/19
```

## Wire checks

```bash
# curl needs the sandbox disabled (network). Both sides of a pair, plus canonical tag:
curl -s -o /dev/null -w "%{http_code} %{size_download}\n" https://robot-hands.com/news.html
curl -s https://robot-hands.com/news.html | grep -oE '<link[^>]*rel="canonical"[^>]*>'
# robot-hands.com/sitemap.xml is a 404 (B2 NoSuchKey blob) — nothing corrects the dup for crawlers.
```

## DB access

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Check enablement (AFTER image is pod-verified — image → seed, never the reverse)

Model: `docs/agent_docs/sql_for_agents/296_enable_content_duplication_on_completeness_discovery.sql`
(snapshot + fences + neighbour assertions + rollback). Pod-verify first:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "page_canonical_collision"'   # want >=1
# negative control: a string the change REMOVED (expect 0) — see NOTES for the chosen needle.
```
