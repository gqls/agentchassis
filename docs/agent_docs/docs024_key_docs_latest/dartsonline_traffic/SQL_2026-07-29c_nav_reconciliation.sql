-- SQL — dartsonline.com nav reconciliation (2026-07-29)
--
-- THE LIVE DEFECT: 3 of the 5 header links 404. Verified on the wire and in the DB:
--   Shop   -> /shop.html    (page row build_status='planned', deployed_at NULL)
--   Brands -> /brands.html  (same)
--   Guides -> /guides.html  (same)
-- This is the open fixloop finding `silent:nav_linked_never_built:5fe8785b`.
--
-- WHAT THIS IS NOT: a missing-pages problem. Those three rows are ORPHANS from
-- superseded site plans. Verified against the CURRENT plan (site_plans.is_current):
-- it contains shop-index / brands-index / guides-index (role section-index) and does
-- NOT contain shop / brands / guides. The hubs replaced the landings; the old rows
-- were left behind still flagged in_header AND in_footer.
--
-- WHY NO CODE FIX: the framework already prunes this. `GetNavItems`
-- (platform/orchestration/actions/nav_tables.go:215-240) drops nav items whose target
-- page has never been deployed, and logs
--   "GetNavItems: dropped nav items whose target page has never been deployed"
-- Fleet-wide measurement, run before concluding anything (the whole point — a local
-- symptom is not evidence of a generic defect):
--   SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
--   WHERE p.in_header AND p.deployed_at IS NULL AND p.status IN ('active','deployed','pending')
--   GROUP BY 1;   -->  dartsonline.com | 4    (ONE site, no other site affected)
-- So the generic fix is already in the binary; dartsonline is serving STALE STORED
-- CHROME (bugs_open/117 — chrome is a stored artefact no page re-render regenerates).
-- The repair is data + a chrome rebuild, not a platform change.
--
-- NOTE the prune alone would merely DELETE Shop/Brands/Guides from the nav. That is
-- honest but poorer. Pointing nav at the real hubs keeps the destinations and is why
-- this file sets in_header on the hubs rather than just clearing the orphans.
--
-- `archived` is the house status value (20 rows fleet-wide use it); archiving also
-- removes these rows from loadPagesForNav, whose filter is
-- `status IN ('active','deployed','pending')` (populate_nav_tables_action.go:240-245).

BEGIN;

CREATE TABLE IF NOT EXISTS bak_darts_pages_nav_20260729 AS
SELECT id, name, url, page_type, status, build_status, in_header, in_footer, nav_label, nav_order
FROM pages WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

-- 1. Archive the three orphan landing rows (absent from the current plan, never deployed)
UPDATE pages SET status = 'archived', in_header = false, in_footer = false, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND name IN ('shop', 'brands', 'guides')
  AND build_status = 'planned'
  AND deployed_at IS NULL;

-- 2. The setup-builder IS in the current plan but is not built yet. Keep the row
--    active; keep it OUT of nav until it serves 200 (rail: never link a non-200).
UPDATE pages SET in_header = false, in_footer = false, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND name = 'tool-setup-builder';

-- 3. Put the real hubs into the header with clean labels and sane order.
--    guides-index is ALREADY deployed, so it survives the GetNavItems prune from the
--    next chrome render onwards. shop-index/brands-index are build_status='needs_rebuild'
--    and will be pruned until they deploy — which is the correct behaviour, not a bug:
--    they appear the moment they are real.
UPDATE pages SET in_header = true, nav_label = 'Shop',   nav_order = 20, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'shop-index';

UPDATE pages SET in_header = true, nav_label = 'Brands', nav_order = 30, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'brands-index';

UPDATE pages SET in_header = true, nav_label = 'Guides', nav_order = 40, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'guides-index';

-- 4. Clean the polluted nav_labels. Whole <title> strings leaked into nav_label on
--    every blog post plus sale and shipping-returns — e.g.
--    "Tungsten Percentage Explained — 80% vs 90% vs 95% | Darts Online".
--    A nav label is a couple of words; take the part before the first '|' or '—'
--    and trim. Blog posts are Tier-4 (never primary nav) but the label is also used
--    by listings and breadcrumbs, so it is worth being right.
UPDATE pages
SET nav_label = btrim(split_part(split_part(nav_label, '|', 1), '—', 1)),
    updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND (nav_label LIKE '%|%' OR nav_label LIKE '%—%');

COMMIT;

-- Verify: no active in_header page may lack a deployed_at, except the two hubs we
-- are about to build (shop-index, brands-index).
SELECT name, url, page_type, status, build_status, in_header, nav_label, nav_order,
       (deployed_at IS NOT NULL) AS deployed
FROM pages
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND (in_header = true OR status = 'archived')
ORDER BY in_header DESC, nav_order;
