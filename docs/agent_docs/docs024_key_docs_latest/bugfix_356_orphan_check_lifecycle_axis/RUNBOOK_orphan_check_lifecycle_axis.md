# RUNBOOK — bugfix 356 (orphan check takes the build axis only)

Every command that was hard to get right, with its gotcha attached.

## Read a completed work item's real outcome

⚠ **The outcome is nested under `response`, not at the top of `result`.**
`result->'target_page'` returns NULL on every row and reads as "unreadable" — that is
misstep 1 in NOTES, and it silently turned 17 no-target rows into 34 unreadable ones.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
SELECT
  count(*) AS complete_total,
  count(*) FILTER (WHERE (result->'response'->'target_page'->>'count')::int = 0) AS no_target,
  count(*) FILTER (WHERE (result->'response'->'target_page'->>'count')::int > 0) AS with_target,
  count(*) FILTER (WHERE result->'response'->'target_page' IS NULL)              AS unreadable
FROM site_work_items WHERE handler_agent='internal-linker' AND status='complete';
SQL
```
⚠ The `unreadable` bucket is real and is **not** this bug — it is `bugs_closed/287`'s
`result`-is-the-spawn-record trap. Report it as its own column; folding it into either
other bucket misstates the finding.

To find the path on a table you have not read before, dump one row rather than guessing:
```bash
… -c "SELECT jsonb_pretty(result) FROM site_work_items WHERE handler_agent='internal-linker' AND status='complete' ORDER BY completed_at DESC LIMIT 1"
```

## Join a work item to the page it names

⚠ **`check_orphan_pages` does not set the `page_id` COLUMN** — its three `WorkItemSpec`
literals omit the field, so the id lives only in `spec->>'page_id'` (as TEXT, because it is
scanned `p.id::text`). A join on `w.page_id` returns nothing and looks like "no page".

```sql
FROM site_work_items w LEFT JOIN pages p ON p.id = (w.spec->>'page_id')::uuid
```

## The fleet census — run the producer's OWN predicate, split by lifecycle

This is the measurement that decides the bug. It runs `findOrphanPagesSQL` verbatim and
groups by `pages.status` and routing branch, so the archived rows are visible as the
producer sees them.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
WITH orphans AS (
  SELECT p.id, p.site_id, p.name, p.status, p.page_type,
         COALESCE(p.in_header,false) OR COALESCE(p.in_footer,false) AS nav_flagged
  FROM pages p
  WHERE NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed')   -- PageHasShippedPredicateFor("p")
    AND p.url IS NOT NULL AND p.url != ''
    AND p.name NOT IN ('index','home')
    AND (COALESCE(p.page_type,'content') NOT IN ('blog-index','tool')
         OR COALESCE(p.in_header,false) OR COALESCE(p.in_footer,false))
    AND NOT EXISTS (SELECT 1 FROM site_nav_items sni WHERE sni.site_id=p.site_id AND sni.url=p.url AND sni.status='active')
    AND NOT EXISTS (SELECT 1 FROM site_nav_items sni WHERE sni.page_id=p.id AND sni.status='active')
    AND NOT EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id=p.site_id AND sc.rendered_html IS NOT NULL AND sc.rendered_html LIKE '%'||p.url||'%')
    AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN pages p2 ON pc.page_id=p2.id
                    WHERE p2.site_id=p.site_id AND p2.id != p.id AND pc.rendered_html IS NOT NULL AND pc.rendered_html LIKE '%'||p.url||'%')
)
SELECT status AS page_status,
       CASE WHEN page_type='blog-post' THEN 'blog'
            WHEN nav_flagged THEN 'nav_drift'
            ELSE 'needs_internal_links' END AS routing,
       count(*) AS pages, count(DISTINCT site_id) AS sites
FROM orphans GROUP BY 1,2 ORDER BY 1,2;
SQL
```

⚠ **`PageHasShippedPredicateFor` is `NOT (deployed_at IS NULL AND build_status <> 'deployed')`
— spell it exactly.** Writing `build_status='deployed'` instead loses every page flagged
`needs_rebuild` after a real deploy and gives a different, wrong population
(`bugs_closed/037`, `bugs_closed/185`).

**This is the DISCONFIRMING measurement**, which is why it is worth its length: it is grouped
by `status`, so it could have come out with zero archived rows. It did not.

## Prove the three remedy paths disagree with the producer

Each is a separate read, and the point is that all three already carry a lifecycle arm:

```bash
# 1. needs_internal_links -> internal-linker  (LIVE CONFIG, not the seed)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT s.value->'config'->>'query'
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.type='internal-linker' AND a.is_active AND COALESCE(a.is_snapshot,false)=false
    AND a.deleted_at IS NULL AND s.key='load_target_page'"
# expect: ... AND p.status = 'active' ...
```
⚠ Read the LIVE row, never `sql_for_agents/101_internal_linker.sql` — the seed records what
the agent WAS (memory: *the seed is not the system*).

```bash
# 2. nav_drift -> nav-updater
grep -n "navPageScopeSQL *=" platform/orchestration/actions/nav_prune_floor.go
#    status IN ('active', 'deployed', 'pending')

# 3. orphan_blog_posts -> rerender-pages
sed -n '108,112p' platform/orchestration/actions/rebuild_blog_listing_action.go
#    AND p.status IN ('active', 'deployed')
```

## Measure the recurrence (why this is a loop, not an incident)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
SELECT w.item_key, count(*) AS times_filed,
       min(w.created_at)::date AS first, max(w.created_at)::date AS last, p.status AS page_status
FROM site_work_items w LEFT JOIN pages p ON p.id=(w.spec->>'page_id')::uuid
WHERE w.handler_agent='internal-linker'
GROUP BY w.item_key, p.status HAVING count(*) > 1 ORDER BY 2 DESC LIMIT 12;
SQL
```
⚠ Re-filing is EXPECTED and is not itself the bug: `idx_swi_dedup` excludes terminal
statuses, so a completed finding legitimately re-raises next rotation. The bug is that the
finding is **unsatisfiable**, so the re-raise never ends.

## Auditing which discovery checks carry a lifecycle arm

⚠ **Do not grep-count `status = 'active'`.** Every one of these checks joins
`site_nav_items sni ... AND sni.status='active'`, which matches the grep and says nothing
about the PAGE row. Measured 2026-08-22: a naive `grep -c` scores `check_orphan_pages.go`
at **2** — the file with the defect. Read the SQL and ask which table the predicate binds to.

## The 090 diagnosis for this lane

```
intake  4480a3a7-b4cd-4026-828c-5297878dfb7f
run     7bac4520-651d-41f9-aa98-f4721c49902f      <- artifacts are written under THIS
```
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
  WHERE correlation_id='7bac4520-651d-41f9-aa98-f4721c49902f' ORDER BY created_at"
```
⚠ A missing row is latency, not a dropped dispatch — do not re-trigger.
