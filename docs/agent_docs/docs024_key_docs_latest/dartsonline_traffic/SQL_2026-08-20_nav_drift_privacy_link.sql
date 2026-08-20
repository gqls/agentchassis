-- SQL_2026-08-20_nav_drift_privacy_link.sql
--
-- WHY: /privacy.html went live this morning (built through the planner route, copy
-- verified verbatim 16/16 blocks against the approved draft) and **nothing links to
-- it**. Measured 2026-08-20:
--
--   curl https://dartsonline.com/ | grep -c 'href="/privacy.html"'   ->  0
--   site_nav_items                                                   ->  9 rows,
--     primary: Guides, News, Start Here, Deals
--     utility: Home, About, Contact, Setup Builder, Dart Weight Comparator
--     (no Privacy row; the Shipping & Returns row is already gone)
--
-- An orphan privacy page is close to useless for the job it exists for: an affiliate
-- network reviewer looks for the FOOTER LINK, and a crawler that cannot reach a page
-- from anywhere on the site has little reason to keep it. `pages.in_footer` is already
-- true for this row — the nav TABLE is what has not heard about it.
--
-- Also settles a loose end from 2026-08-16: shipping-returns was retired and retracted
-- today (page archived, file deleted in sites@2af7c17dd, live 404 confirmed), so this
-- run should also drop any nav row still pointing at it.
--
-- ⚠ THE TRAP TO GRADE THIS AGAINST (LANDMINES, and this lane paid for it on 08-16):
-- a nav_drift rebuild refreshes stored chrome on EVERY page and redeploys only SOME.
-- Measured here on 08-16: item `complete`, 0 of 25 pages stale in the DB, **19 of 25
-- SERVED pages still stale** — and the 6 that had redeployed were the homepage and the
-- two index pages, i.e. exactly what a person spot-checks. So grade at the served bytes
-- over the whole sitemap, never at the work item's status and never at pages.rendered_*.
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
  'Privacy page is live and orphaned — no nav row and no footer link on any page',
  jsonb_build_object(
    'orphan_type', 'nav_drift',
    'affected_pages', jsonb_build_array('privacy','shipping-returns'),
    'affected_count', 2,
    'fix', 'site_nav_items has no row for /privacy.html, which went live 2026-08-20 with in_footer=true. Re-run populate_nav_tables from the current pages rows and re-render the site components, so Privacy appears in the utility/footer group. The same run should drop any row still pointing at /shipping-returns.html, whose page was archived and whose file was retracted today.',
    'note', 'Expected footer group after this runs: Home, About, Contact, Privacy (plus the two tool links). Privacy is a legal/utility link, not a primary nav item — in_header stays false. Grade at the SERVED bytes across the sitemap, not at the item status: this exact item type refreshed stored chrome on all 25 pages here on 2026-08-16 and redeployed only 6 of them.'
  ),
  20,
  'nav-updater',
  'triaged',
  'dartsonline-traffic-workstream',
  'nav_drift_privacy_link:5fe8785b-223d-41a3-88ee-c07187622381',
  'auto'
)
ON CONFLICT DO NOTHING;

SELECT id, status, item_key FROM site_work_items
 WHERE item_key = 'nav_drift_privacy_link:5fe8785b-223d-41a3-88ee-c07187622381';
