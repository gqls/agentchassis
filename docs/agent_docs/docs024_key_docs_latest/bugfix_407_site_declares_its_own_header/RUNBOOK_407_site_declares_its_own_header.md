# RUNBOOK — bugs_open/407

Every command that was hard to get right, with its gotcha attached. Change commands HERE.

---

## 1. THE DISCRIMINATOR — is a page absent because of the TIER, or because the nav is stale?

`bugs_open/407` §4 says this could not be built and its own attempt
(`pages.updated_at > max(site_nav_items.updated_at)`) fails because `updated_at` is bumped by
any re-render. **The way past it is not a better timestamp.** `classifyPagesForNav` is
deterministic Go over `pages`, so REPLAY it in SQL and diff the expected primary rank against
the stored nav. A page absent from stored primary is then exactly one of three things.

### (a) The cheap screen — necessary, and it comes from the mechanism

The tier table only decides anything when there is COMPETITION for slots. Below the cap there is
no competition, so tier cannot be the cause.

```sql
WITH cap AS (   -- ⚠ read the cap from LIVE config. Hardcoding 8 asserts the very
                -- fleet-wide constant this bug is about.
  SELECT COALESCE((jsonb_path_query_first(default_config,
           '$.** ? (@.action == "populate_nav_tables")')->'config'->>'max_header_items')::int, 8) AS n
  FROM agent_definitions
  WHERE type='nav-updater' AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL),
absent AS (   -- the bug's own §6 population
  SELECT p.id, p.site_id, s.domain, p.name
  FROM pages p JOIN sites s ON s.id=p.site_id
  WHERE p.in_header AND p.status='active' AND p.build_status='deployed'
    AND p.url NOT LIKE '/tools/%' AND p.url NOT LIKE '/blog/%' AND p.url NOT LIKE '/guides/%'
    AND NOT EXISTS (SELECT 1 FROM site_nav_items ni JOIN site_nav_groups ng ON ng.id=ni.group_id
                     WHERE ni.site_id=p.site_id AND ng.group_type='primary'
                       AND ni.status='active' AND ni.url=p.url)),
stored AS (
  SELECT ni.site_id, count(*) AS primary_items
  FROM site_nav_items ni JOIN site_nav_groups ng ON ng.id=ni.group_id
  WHERE ng.group_type='primary' AND ni.status='active' GROUP BY 1)
SELECT a.domain, a.name, COALESCE(st.primary_items,0) AS primary_items, cap.n AS cap,
       CASE WHEN COALESCE(st.primary_items,0) >= cap.n THEN 'TIER/CAP (407)'
            ELSE 'NOT 407 — nav below cap' END AS verdict
FROM absent a LEFT JOIN stored st ON st.site_id=a.site_id CROSS JOIN cap
ORDER BY 1,2;
```

⚠ "At cap" is necessary and, **on today's data**, sufficient. Its named failure mode: a site at
cap whose nav is ALSO stale would be classified 407 wrongly. (b) closes that.

### (b) The mechanism replay — the authoritative per-page verdict

The full query is long; it is reproduced in the lane NOTES for 2026-08-26. Its shape:
re-implement `classifyPagesForNav`'s buckets and tiers as SQL CASE expressions over
`pages`, `row_number()` the candidates by `(tier, nav_order, created_at)`, and read the rank
of each absent page.

⚠ Three caveats travel with it and must not be dropped:
- the Go sort is `sort.Slice`, which is **UNSTABLE**, so pages tied on `(tier, nav_order)` at the
  cap boundary have no guaranteed order. None of today's verdicts sits on such a tie (ranks 9–12
  against a cap of 8 are all strict).
- it is a **measurement tool, not a monitor** — it must be edited in lockstep if the Go lists
  change.
- post-fix it needs a `declared` CTE ranked ahead of the tier sort, and the literal cap replaced
  by `COALESCE((sp.data->'chrome'->>'max_header_items')::int, cap.n)`.

`[MEASURED 2026-08-26]` the two instruments agree: **5 of 6 are 407; `idea.uk/report` is
`page_type='tool'`, barred by `neverPrimaryTypes`, and no rebuild would ever place it.**

## 2. Post-fix acceptance — must return 0 rows, per DECLARED site

```sql
SELECT s.domain, slot.name, slot.ord
FROM sites s
JOIN site_specs sp ON sp.site_id = s.id AND sp.aspect='site_config' AND sp.is_current
CROSS JOIN LATERAL jsonb_array_elements_text(sp.data->'chrome'->'header_slots')
           WITH ORDINALITY AS slot(name, ord)
JOIN pages p ON p.site_id = s.id AND lower(p.name) = lower(slot.name)
            AND p.status IN ('active','deployed','pending')
WHERE lower(p.name) NOT IN ('404','sitemap','robots')
  AND lower(p.name) !~ '^(privacy|terms|cookie|disclaimer|legal)'
  AND slot.ord <= COALESCE((sp.data->'chrome'->>'max_header_items')::int, 8)
  AND NOT EXISTS (SELECT 1 FROM site_nav_items ni JOIN site_nav_groups ng ON ng.id=ni.group_id
                   WHERE ni.site_id=s.id AND ng.group_type='primary'
                     AND ni.status='active' AND ni.page_id = p.id);
```

Keyed on `page_id`, which `insertNavItem` stores — not on URL.

## 3. ⚠ VERIFY AT THE SERVED PAGE, AND ANCHOR ON `<nav>` — NEVER ON `<header>`

The bug's §6 warns that nav-updater's last step only FILES re-render items, so the tables can be
correct for an hour while every served header is stale, and that `pages.rendered_header` is NULL
site-wide on some sites.

```bash
curl -s "https://<domain>/" | tr -d '\n' \
  | grep -oE '<nav[^>]*>.*</nav>' | head -1 \
  | grep -o 'href="[^"]*"'
```

**A rendered page contains SEVERAL `<header>` elements**, because components carry their own.
Matching `<header>…</header>` on idea.uk returned `<header class="info-card-grid__header">` — a
420-byte card-grid header with no links in it — and every verdict taken off it was worthless.
The tell: a site chrome with zero `href` is not a chrome. Print the hrefs you found.

## 4. Reading what the declaration actually did

The step returns its own outcome; read that rather than inferring from the nav tables.

```sql
SELECT collected_data->'refresh_nav_tables'->>'nav_declaration_source',   -- default|site_config|invalid
       collected_data->'refresh_nav_tables'->'declared_slots',
       collected_data->'refresh_nav_tables'->'declared_missing',
       collected_data->'refresh_nav_tables'->'declared_ineligible',
       collected_data->'refresh_nav_tables'->'declared_flag_disagreed',
       collected_data->'refresh_nav_tables'->>'max_header_items_effective'
  FROM orchestration_states
 WHERE created_at > now() - interval '30 minutes'
   AND collected_data ? 'refresh_nav_tables'
 ORDER BY created_at DESC LIMIT 3;
```

`nav_declaration_source = 'invalid'` means the spec holds a shape the reader could not use —
fix the spec rather than leaving it half-read.

## 5. Where the declaration lives, and the two homes that look right and are not

```sql
SELECT s.domain, sp.data->'chrome' FROM site_specs sp JOIN sites s ON s.id=sp.site_id
 WHERE sp.aspect='site_config' AND sp.is_current AND sp.data ? 'chrome';
```

⚠ **`site_nav_items` AND `site_nav_groups` are BOTH deleted for the site on every rebuild**
(`populate_nav_tables_action.go:160,163`) — neither can hold a declaration. ⚠ `pages.in_header`
and `pages.nav_order` are rewritten by `sync_pages_to_db`'s upsert on every re-plan. ⚠ A site
with **no current `site_config` row** gets nothing from an `UPDATE … WHERE aspect='site_config'`:
it matches nothing and reports `UPDATE 0`, which reads exactly like success. `[MEASURED
2026-08-26]` 33 of 51 sites have one, so 18 do not — check before seeding, and INSERT if absent.
