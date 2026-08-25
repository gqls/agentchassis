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

## THE TWO CLOSURE QUERIES (added 2026-08-25 — they were in no doc until today)

The lane's headline census (36 → 48 anchors) and the acceptance fetch were both run from
scrollback and never written down, so the 2026-08-25 session had to **reconstruct** the first from
the Go predicates. That is exactly the gap this file exists to close. Both are below.

### (a) The stored blast-radius census — NOT the closure gate

Mirrors the live code, so it moves when the code moves:
`datahelpers.PageLinkRefusedPredicateFor("p")` (`platform/orchestration/datahelpers/links.go:386`),
`NeverDeployedPagePredicateFor` (`links.go:277`, the shipped-site escape),
`NormalizePagePath` (`links.go:215`) and `linkablePageStatusPredicate`
(`platform/orchestration/actions/prepare_link_context_action.go:54`).

```sql
WITH refused_pages AS (
  SELECT p.site_id, p.url,
         COALESCE(NULLIF(rtrim(regexp_replace(lower(btrim(split_part(split_part(p.url,'#',1),'?',1))),'index\.html$',''),'/'),''),'/') AS norm
  FROM pages p
  WHERE p.status NOT IN ('deleted','archived')
    AND ( (p.deployed_at IS NULL AND COALESCE(p.build_status,'')='planned'
           AND p.updated_at < NOW() - INTERVAL '48 hours')
       OR (p.deployed_at IS NULL AND COALESCE(p.build_status,'')='needs_rebuild'
           AND NOT EXISTS (SELECT 1 FROM page_components pc
                           WHERE pc.page_id=p.id AND COALESCE(pc.rendered_html,'')<>'')) )
),
shipped_sites AS (              -- ESCAPE 2: a site that has never shipped a page suppresses nothing
  SELECT p.site_id FROM pages p
  WHERE p.status NOT IN ('deleted','archived')
    AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed')
  GROUP BY 1
),
all_pages_norm AS (
  SELECT p.site_id, p.id,
         COALESCE(NULLIF(rtrim(regexp_replace(lower(btrim(split_part(split_part(p.url,'#',1),'?',1))),'index\.html$',''),'/'),''),'/') AS norm
  FROM pages p WHERE p.status NOT IN ('deleted','archived')
),
anchors AS (
  SELECT pg.site_id, pg.id AS ref_page_id, pg.url AS ref_url, pg.deployed_at AS ref_deployed,
         m[1] AS raw_href,
         COALESCE(NULLIF(rtrim(regexp_replace(lower(btrim(split_part(split_part(m[1],'#',1),'?',1))),'index\.html$',''),'/'),''),'/') AS norm
  FROM page_components pc
  JOIN pages pg ON pg.id = pc.page_id
  CROSS JOIN LATERAL regexp_matches(pc.rendered_html, 'href="([^"]*)"', 'g') AS m
  WHERE COALESCE(pc.rendered_html,'') <> ''
    AND pg.status NOT IN ('deleted','archived')
    AND m[1] LIKE '/%'
)
SELECT s.domain, a.ref_url, COALESCE(a.ref_deployed::text,'NEVER') AS ref_deployed,
       a.raw_href, count(*) AS hits
FROM anchors a
JOIN shipped_sites ss ON ss.site_id = a.site_id
JOIN refused_pages r  ON r.site_id = a.site_id AND r.norm = a.norm
JOIN sites s          ON s.id = a.site_id
GROUP BY 1,2,3,4 ORDER BY 1,2,4;
```

⚠ **It will NEVER reach zero, and it is not the bar.** Suppression is outbound-only *by design* —
the authored href stays in `page_components.rendered_html` so the link returns when the target
ships. Use this to enumerate the **population**; use (b) to decide.

⚠ **Swap `JOIN refused_pages` for `LEFT JOIN` + a CASE to get the kept/suppressed split**, and read
the kept count as the `bugs_open/313` control: if it collapses, something is stripping internal
links wholesale.

⚠ **Its absolute number is not comparable across sessions unless the SQL is byte-identical.** The
2026-08-25 run read 36/21/9 against 08-24's 48/28/16, but from a reconstruction — some of that gap
is targets that shipped, some is encoding. **A census reconstructed is a new instrument.**

### (b) The served census — THIS is the closure gate

Feed (a)'s `domain | ref_url | raw_href` rows into a cache-busted fetch. Two assertions per page,
plus a control per domain, all in one run:

```bash
# CONTROL FIRST: a parked domain 200s every path, which would make the whole census read clean
for d in <domains>; do
  printf '%-32s invented-url -> %s\n' "$d" \
    "$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 "https://${d}/definitely-not-a-real-page-328-$RANDOM.html")"
done   # every public domain must return 404; a non-resolving .internal host is OUT of the population

while IFS='|' read -r dom path needles; do
  body=$(curl -s --max-time 25 "https://${dom}${path}?cb=$RANDOM$RANDOM")
  total=$(printf '%s' "$body" | grep -o 'href="/[^"]*"' | wc -l)     # POSITIVE CONTROL
  for n in $needles; do
    printf '%-25s %-44s %-40s DEAD=%s internal_total=%s\n' "$dom" "$path" "$n" \
      "$(printf '%s' "$body" | grep -o "href=\"${n}\"" | wc -l)" "$total"
  done
done < pages.txt
```

⚠ **`DEAD=0` alone is not a pass.** Capture `internal_total` BEFORE the re-render and assert it is
materially unchanged after — otherwise a page that stopped emitting internal links altogether
scores as fixed. That is `bugs_open/313`'s failure mode and this platform has reached it.

⚠ **Re-confirm the targets are still unbuilt at the end.** If a target shipped, the link was
*validated*, not removed, and the run proves nothing about suppression.

## Dispatching the tail: 7 inserts and 1 re-arm (2026-08-25)

Mirror the canary row `b18a0287` exactly — `handler_agent='page-rerender'`, `status='triaged'`,
`priority=40`, `severity='medium'`, `pipeline='build'`, `approval_mode='auto'`, and
`spec = {domain, page_id, filename, page_name}` where `filename` is `ltrim(pages.url,'/')`.

⚠ **`spec.page_name` is mandatory** — a `page_rerender` dispatched without it throws away
everything it re-renders.

⚠ **A page_rerender parked at `status='deferred'` holds the `idx_swi_dedup` slot for ever, and your
INSERT for that page fails 23505.** `deferred` is not in the index's terminal-status array, and
`claim_work_item_action.go:102` selects only `('triaged','approved')` — so nothing will ever
dispatch it either. **Re-arm the existing row rather than inserting a duplicate**, and leave its
`source`/`created_by` alone; it is another producer's provenance. Check every slot first:

```sql
SELECT s.domain, p.url,
       COALESCE((SELECT w.status FROM site_work_items w
                 WHERE w.site_id=p.site_id
                   AND w.item_key='page_rerender_'||p.name||'_'||p.site_id||'_assemble'
                   AND w.status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
                                              'failed','unresolved','cancelled'])
                 LIMIT 1),'FREE') AS slot
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE (s.domain, p.url) IN ( ... );
```

⚠ **No dispatch within ~300s of a chassis pod restart** — the spawn is silently dropped. Check
`kubectl -n ai-persona-system get pods -l app=agent-chassis` for pod AGE immediately before firing.
