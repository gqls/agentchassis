-- SQL_2026-07-29l_nav_drift_then_builds.sql
--
-- WHAT MEASUREMENT CHANGED HERE, and it corrects session 1's handoff caveat.
--
-- Session 1 closed saying: "the nav fix is DATA-ONLY. pages.in_header is correct
-- and the orphan rows are archived, but the live header still serves stale
-- chrome (bugs_open/117), so the three 404 links remain on the served pages
-- until a chrome rebuild runs."
--
-- That was the wrong diagnosis, and curl says so. The three dead links
-- (/shop.html, /brands.html, /guides.html) are NOT on every served page:
--
--   /index.html            clean      rebuilt 2026-07-29
--   /about.html            clean      rebuilt 2026-07-29
--   /blog/barrel-weight.html clean    rebuilt 2026-07-29
--   /blog/beginners.html   clean      rebuilt 2026-07-29
--   /shipping-returns.html clean      rebuilt 2026-07-29
--   /sale.html             3 dead links   last built 2026-07-28
--   /new-arrivals.html     3 dead links   last built 2026-07-26
--   /guides/index.html     3 dead links
--   /contact.html          3 dead links
--
-- The header is regenerated PER PAGE at build time, and GetNavItems already
-- prunes never-deployed targets — which is exactly why every page rebuilt today
-- came out clean without anybody fixing chrome. So this was never a stale-chrome
-- problem (bugs_open/117 does not apply here); it is four pages that have not
-- been rebuilt since the nav data was corrected.
--
-- But site_nav_items is still stale, and that IS a real gap: it holds the three
-- archived orphan rows and has never heard of guides-index. Rebuilding a page
-- right now would prune the dead links (good) and still not add Guides (bad).
-- So the order is load-bearing:
--
--   1. THIS FILE: one nav_drift item -> nav-updater, which re-runs
--      populate_nav_tables from the corrected `pages` rows and re-renders the
--      site components.
--   2. WAIT for it to complete.
--   3. THEN the page rebuilds (SQL ...m), so each page bakes in the right nav.
--
-- Doing 3 before 2 would rebuild seven pages against a nav table that does not
-- contain Guides, and every one of them would have to be built again.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
) VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'dartsonline-traffic-workstream',
  'build',
  'nav_drift',
  'high',
  'Nav tables still hold three archived orphan pages and have never heard of guides-index',
  jsonb_build_object(
    'orphan_type', 'nav_drift',
    'affected_pages', jsonb_build_array('guides-index','new-arrivals','sale'),
    'affected_count', 3,
    'fix', 'site_nav_items still lists /shop.html, /brands.html and /guides.html — rows for pages archived 2026-07-29 — and omits guides-index, which is now in_header. Re-run populate_nav_tables from the corrected pages rows and re-render the site components.',
    'note', 'Expected header after this runs, by nav_order: Guides (3), Start Here (4), Deals (5), plus the utility group (Home, About, Contact, Shipping & Returns). shop-index and brands-index are deliberately in_header=false: they are retail hubs with nothing to put in them until an affiliate feed exists.'
  ),
  20,
  'nav-updater',
  'triaged',
  'dartsonline-traffic-workstream',
  'nav_drift_rebuild:5fe8785b-223d-41a3-88ee-c07187622381',
  'auto'
)
ON CONFLICT DO NOTHING;

-- ── verification ───────────────────────────────────────────────────────────
-- Watch it drain (the `error` column holds the LAST failure and is NOT cleared
-- on re-claim, so a stale message there means nothing — read status):
--   SELECT status, attempt_count, updated_at FROM site_work_items
--   WHERE item_key='nav_drift_rebuild:5fe8785b-223d-41a3-88ee-c07187622381';
--
-- Then confirm the nav table itself, which is the artefact this is for:
--   SELECT g.group_key, i.label, i.url, i.position
--   FROM site_nav_items i JOIN site_nav_groups g ON g.id=i.group_id
--   WHERE i.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--   ORDER BY g.group_key, i.position;
-- Expect: no /shop.html, no /brands.html, no /guides.html; a Guides row
-- pointing at /guides/index.html.
