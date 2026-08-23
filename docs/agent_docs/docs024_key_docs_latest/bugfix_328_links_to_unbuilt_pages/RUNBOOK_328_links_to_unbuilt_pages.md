# RUNBOOK — bugfix 328

Every query/command that was hard to get right, with its gotcha attached.

## Is the bug still live? (the artefact, not the plan)

```bash
for u in https://loanzy.uk/ https://loanzy.uk/your-rights.html https://loanzy.uk/guides/index.html \
         https://mortgagecalculator.co.uk/scorecard-simulator.html; do
  printf "%-60s %s\n" "$u" "$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 "$u?cb=$RANDOM")"
done
```

⚠ **Cache-bust every fetch** (`?cb=$RANDOM`). A CDN 200 on a retracted page and a cached 404 on
a page that shipped ten minutes ago are both states this fleet reaches.

The other half — the anchor must still be SERVED, not merely stored:

```bash
curl -s --max-time 20 "https://loanzy.uk/?cb=$RANDOM" | grep -o 'href="[^"]*"' | sort | uniq -c | sort -rn
```

⚠ Do **not** substitute a `page_components.rendered_html` query for this. Stored and served
diverge in both directions on this platform, and the link-repair unlink arm is one of the
mechanisms that makes them diverge (LANDMINES: "a dead internal link is REPAIRED into orphaned
prose" — the stored HTML holds a well-formed anchor while the wire shows bare words).

## The population

```sql
-- who is affected, and for how long
SELECT s.domain, count(*) AS items, count(DISTINCT w.page_id) AS distinct_targets,
       min(w.created_at)::date AS oldest
FROM site_work_items w JOIN sites s ON s.id = w.site_id
WHERE w.item_type = 'unbuilt_internal_link'
  AND w.status NOT IN ('complete','verified','cancelled','rejected','wont_fix')
GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **`w.page_id` on this item type is the TARGET page, not the linking page** — the check files
it that way deliberately (`check_phantom_internal_links.go`, the `actionablePageID` switch), and
the linking page is named only inside `summary`. Reading it as the referrer inverts every count.

```sql
-- WHY they are parked: the load-bearing column is `error`, not `status`
SELECT status,
       CASE WHEN error ILIKE '%no sections ready to build%' THEN 'no sections ready'
            WHEN error ILIKE '%content validation failed%'  THEN 'validation blockers'
            WHEN error ILIKE '%SECTION COMPONENT FLOOR%'    THEN 'component floor'
            WHEN error ILIKE '%AI endpoint unav%'           THEN 'AI endpoint'
            WHEN error IS NULL OR error = ''                THEN '(none)'
            ELSE 'other' END AS why,
       count(*), min(created_at)::date AS oldest, max(created_at)::date AS newest
FROM site_work_items WHERE item_type = 'unbuilt_internal_link'
GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠ `needs_human_review` reads like "never dispatched". It is not: check `triaged_at`,
`handled_by` and `attempt_count` in the same breath. Every parked row here had the handler run
against it and fail.

```sql
-- how many pages fleet-wide would 404 if linked
SELECT build_status, count(*), count(*) FILTER (WHERE deployed_at IS NOT NULL) AS ever_deployed
FROM pages WHERE COALESCE(status,'') NOT IN ('deleted','archived') GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **`build_status` is not the predicate.** 38 of the 52 `needs_rebuild` pages HAVE shipped and
serve their previous artefact. The estate's one definition of "has been served" is
`datahelpers.NeverDeployedPagePredicate` (`deployed_at`-based) and it has a lockstep test
family; anything that spells the question a second way is the drift this platform keeps
re-finding.

## Ownership, before routing any work

```bash
./scripts/who-owns.py 328     # advisory, ~0.3s, reads COMMITS — blind to a session mid-fix
git status --porcelain | grep -E "link_repair|validate_page_content|chrome_link_policy"
```

Both are lagging. Re-run at every phase boundary.
