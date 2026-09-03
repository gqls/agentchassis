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
-- ⚠ the timestamp column is occurred_at, NOT created_at (this file said created_at until
-- 2026-09-03 and the query errored); context key is page_name; domain column is often blank —
-- join sites on site_id for the domain.
SELECT ael.occurred_at, s.domain, ael.context->>'page_name', ael.context->>'component',
       ael.context->>'repeats', ael.context->>'without_subject'
FROM agent_error_log ael LEFT JOIN sites s ON s.id = ael.site_id
WHERE ael.error_code='REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT'
ORDER BY ael.occurred_at DESC LIMIT 20;
```

## Deploy verification (debug_historian advisory, council b7c59309)

Trust the POD, not the tag/config/tests. After the roll, capability-probe the chassis binary
with a present-control AND an absent-control in the same breath:

```bash
P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
# Discriminating literals for dbb218a41 — verified 2026-09-02 by the finetuning session
# against the live pod: `subjects_attached` and `facts_attached` are 0 in the pre-fix
# binary and absent from git at dbb218a41^. Do NOT probe `section_subjects` as the
# shipped-signal: it is the RAILS literal (born 35905c547, already rolled) and returns 1
# on a binary that has never seen this fix. Same for `without_subject` (pre-existing in
# write_site_plan_action.go).
kubectl -n ai-persona-system exec $P -- grep -ac 'subjects_attached' /proc/1/exe                         # >0 = dbb218a41 shipped
kubectl -n ai-persona-system exec $P -- grep -ac 'REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT' /proc/1/exe  # >0 = shipped (second discriminator)
kubectl -n ai-persona-system exec $P -- grep -ac 'section_subjects' /proc/1/exe                          # >0 (present control — RAILS literal, pre-existing by design)
kubectl -n ai-persona-system exec $P -- grep -ac 'STRING_THAT_MUST_NOT_EXIST_443' /proc/1/exe            # 0 (absent control)
```

Then read the build provenance stamp and `git merge-base --is-ancestor <this fix's commit> <stamp>`.
