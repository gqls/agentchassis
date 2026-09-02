# RUNBOOK — bugfix 443 fallback-tier subjects

All psql via:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## The censuses (each figure goes stale by ADDITION — re-run before quoting)

Plan-less real sites + damage proxy (counts empty-sections pages correctly — the
`CROSS JOIN LATERAL … GROUP BY` shape silently DROPS them; `FILTER (WHERE EXISTS …)` does not):

```sql
WITH planless AS (
  SELECT s.id, s.domain FROM sites s
  WHERE NOT EXISTS (SELECT 1 FROM site_plans sp WHERE sp.site_id = s.id AND sp.is_current)
    AND s.domain NOT LIKE 'pool-%.internal'
)
SELECT pl.domain, count(*) AS deployed_pages,
       count(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM (SELECT elem FROM jsonb_array_elements_text(p.sections) elem
                        GROUP BY elem HAVING count(*) > 1) r)) AS pages_with_repeated_type
FROM pages p JOIN planless pl ON p.site_id = pl.id
WHERE p.deployed_at IS NOT NULL
GROUP BY pl.domain ORDER BY 2 DESC;
```

Gotcha: `build_status='deployed'` is a strict subset of `deployed_at IS NOT NULL` — the
17-page gap is `needs_rebuild`, the pages most likely to re-render next (bug file correction).

## Which tier serves a page (per PAGE, not per site)

Site-level "does an aspect exist" MISLEADS — the aspect must name the page WITH sections.
Guard every jsonb length/elements call with `jsonb_typeof(...)='array'` or a scalar row errors
the whole query:

```sql
SELECT CASE WHEN EXISTS (
  SELECT 1 FROM site_specs ss
  CROSS JOIN LATERAL jsonb_array_elements(ss.data->'pages') pg
  WHERE ss.site_id = :site_id AND ss.aspect='site_plan' AND ss.is_current
    AND jsonb_typeof(ss.data->'pages')='array'
    AND pg->>'name' = :page_name
    AND jsonb_typeof(pg->'sections')='array' AND jsonb_array_length(pg->'sections') > 0
) THEN 'TIER2' ELSE 'TIER3-or-4' END;
-- tier 1 first, separately: site_plan_sections rows for the current plan + page_name.
```

## Serve check (invented-URL control per domain — a parked domain 200s everything)

```bash
code=$(curl -s -o page.html -w "%{http_code}" --max-time 20 "https://<domain>/<page>.html")
curl -s -o /dev/null -w "%{http_code}\n" "https://<domain>/zz-invented-control-443.html"  # must 404
grep -o '<h2[^>]*>[^<]*</h2>' page.html | sed 's/<[^>]*>//g'
```

## Mechanism state (read the LIVE row, not the seed)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'plan_sections'->'config')
FROM agent_definitions
WHERE type='page-build-handler' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL;
-- must carry section_facts AND section_subjects (639 applied 2026-09-02).
```

641 state: `grep -n "^-- APPLIED" docs/agent_docs/sql_for_agents/641_*.sql` — no line = held.

## Backfill template (lane work, after the roll; D8)

```sql
-- subjects must be SAME LENGTH and SAME ORDER as pages.sections, one entry per slot,
-- null for slots that need none. Verify alignment in the same statement you write:
UPDATE pages SET section_subjects = :subjects::jsonb
WHERE site_id = :site_id AND name = :page
  AND jsonb_array_length(sections) = jsonb_array_length(:subjects::jsonb);
-- 0 rows updated = misaligned = fix the array, do not force it.
```

## Detector read-back

```sql
SELECT created_at, context->>'page', context->>'component', error_message
FROM agent_error_log WHERE error_code='REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT'
ORDER BY created_at DESC LIMIT 20;
```
