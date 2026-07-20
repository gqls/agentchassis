-- R4e (2026-07-20) — archive the tool page that was never built, and close the
-- two incomplete_page_group items that tracked both of them.
--
-- WHY THIS FILE EXISTS: r4 removed the dead card from
-- `page_components.content_data->'items'`. The re-render at 17:10:23 **put it
-- back** — `items` regenerated to all five tool pages, and the live homepage
-- still carried `/tools/robot-payload-budget-calculator/index.html`, a 404.
--
-- Same lesson as the CTA resolver correction recorded in NOTES: **the tool list
-- is a DERIVED field.** It is regenerated from the site's tool pages on render,
-- so editing the rendered items array cannot hold. The listing includes the page
-- because the page still EXISTS — `build_status='planned'`, never built.
--
-- The house fix for "a row that should not be listed" on this site is R6's:
-- `pages.status='archived'`. R6 archived six dead article rows and the hub then
-- listed exactly the three real guides, so the listings demonstrably respect it.
--
-- STRUCTURAL NOTE (not fixed here): a listing that advertises pages which were
-- never built is how these 404 CTAs arose in the first place. The durable fix is
-- for the tool-list query to filter on `build_status`, not for each site to
-- archive its unbuilt rows by hand. Recorded in /bugs_open/023's addendum
-- territory — same "derived field, authored edit cannot hold" family.
--
-- tool-matchmatrix is deliberately NOT archived: it is now real, live and 200,
-- and it SHOULD be listed.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Archive the unbuilt tool page.
-- ---------------------------------------------------------------------------
UPDATE pages
   SET status = 'archived', updated_at = now()
 WHERE site_id = :'site'
   AND name = 'tool-robot-payload-budget-calculator'
   AND build_status = 'planned';

-- ---------------------------------------------------------------------------
-- 2. Close the two incomplete_page_group items (open since 2026-07-14, both
--    sitting in needs_human_review — a queue with no consumer, /bugs_open/033).
--    matchmatrix: genuinely resolved, the page is built and live.
--    payload-budget: wont_fix, the page is archived rather than built.
-- ---------------------------------------------------------------------------
UPDATE site_work_items
   SET status = 'complete', completed_at = now(), updated_at = now(),
       resolution_path = 'Tool built and deployed live at /tools/matchmatrix/index.html '
                      || '(gqls/sites 0a6dc426); see robot_hands/HANDOFF_2026-07-20'
 WHERE site_id = :'site'
   AND item_type = 'incomplete_page_group'
   AND summary ILIKE '%tool-matchmatrix%'
   AND status = 'needs_human_review';

UPDATE site_work_items
   SET status = 'wont_fix', updated_at = now(),
       resolution_path = 'Page archived rather than built (owner ruling 2026-07-20: remove the '
                      || 'card). Nothing user-visible references it; see robot_hands/HANDOFF_2026-07-20'
 WHERE site_id = :'site'
   AND item_type = 'incomplete_page_group'
   AND summary ILIKE '%tool-robot-payload-budget-calculator%'
   AND status = 'needs_human_review';

-- ---------------------------------------------------------------------------
-- 3. Re-render index so the tool list regenerates WITHOUT the archived page.
--    handler_agent is set here — omitting it is what left the r4d batch sitting
--    at `triaged`, unclaimed and unroutable, with no error (see r4d's own
--    correction block).
-- ---------------------------------------------------------------------------
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'robot-hands-r4e-archive', 'page_rerender', 'medium',
       'Rerender index — regenerate tool list without the archived payload-budget page',
       'triaged', 'session-2026-07-20-robot-hands', 'build', 20, now(),
       'page-rerender',
       'page_rerender_index_r4e_' || p.site_id::text,
       jsonb_build_object('domain','robot-hands.com','reason','cta_links_stale',
                          'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p
WHERE p.site_id = :'site' AND p.name = 'index';

-- ---------------------------------------------------------------------------
-- VERIFY
-- ---------------------------------------------------------------------------
\echo '--- tool pages and their status (payload-budget must be archived) ---'
SELECT name, build_status, coalesce(status,'-') AS status
FROM pages WHERE site_id = :'site' AND name LIKE 'tool-%' AND name NOT LIKE '%-guide'
ORDER BY name;

\echo '--- the two tracking items (expect complete + wont_fix, none left in review) ---'
SELECT status, left(summary,60) FROM site_work_items
WHERE site_id = :'site' AND item_type = 'incomplete_page_group';

COMMIT;
