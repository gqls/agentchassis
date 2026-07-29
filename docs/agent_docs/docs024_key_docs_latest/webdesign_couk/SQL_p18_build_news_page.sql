-- SQL_p18_build_news_page.sql — webdesign.co.uk, phase 2 W5 continuation
--
-- FILE THE BUILD ITEM the news page has been waiting for since 07-27.
--
-- WHY THIS EXISTS. SQL_p8 created the page row with build_status='planned',
-- saying "so the build pipeline picks it up and composes it". Two days later it
-- is still 'planned', because NOTHING WATCHES THE pages TABLE FOR PLANNED ROWS.
-- The two mechanisms that file build items both read a different store:
--   * reconcile_site_plan files needs_page for pages in site_plan_pages —
--     the news page is in `pages` only (hand-inserted), so it never qualifies;
--   * the content-gap planner files needs_content_page from its own plans.
-- The only things that ever touched the page were two feed-cycle
-- `page_rerender` items (07-27 19:58, 07-28 09:10) — both "complete", both
-- no-ops, because assemble republishes STORED html and a never-built page has
-- none (handoff §3 shape 4: a green status is not evidence).
--
-- WHAT THIS DOES.
-- 1. Widens sections from ["news-listing"] to the news-index archetype
--    ["hero","news-listing","call-to-action"] — the shape ALL THREE deployed
--    news-index pages use (gaswholesalers, robot-hands, relojistas), and what
--    defaultSectionsForPage() returns for the type. p8 copied the minimal note,
--    not the working shape. The build reads pages.sections via
--    load_page_sections_from_spec's fallback (the page is in no plan, so the
--    two authoritative sources return nothing).
-- 2. Files the operator-shaped needs_page item, copied from the completed
--    fundamentallyai.com precedent (source='operator', handler
--    page-build-handler, status 'triaged'). item_key follows the reconciler's
--    'needs_page:<page_name>' convention.
--
-- The nav row needs NO action here: refresh_nav_tables deleted it on 07-27
-- precisely because the page was not deployed, and rebuilds it from deployed
-- pages once the build lands (handoff §5 — the old ordering trap is gone).

\set ON_ERROR_STOP on

BEGIN;

UPDATE pages p
   SET sections = '["hero","news-listing","call-to-action"]'::jsonb,
       updated_at = NOW()
  FROM sites s
 WHERE s.id = p.site_id
   AND s.domain = 'webdesign.co.uk'
   AND p.name = 'news'
   AND p.page_type = 'news-index'
   AND p.build_status = 'planned'          -- refuse to touch it if a build won the race
   AND p.sections = '["news-listing"]'::jsonb;

INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, page_id, priority, handler_agent, status, created_by, item_key
)
SELECT s.id, 'operator', 'build', 'needs_page', 'medium',
       'Build news page /news/index.html — planned since 07-27, feed live with 25 on-topic items',
       jsonb_build_object('reason', 'not_built', 'page_name', 'news'),
       p.id, 40, 'page-build-handler', 'triaged', 'webdesign_couk_thread',
       'needs_page:news'
  FROM sites s
  JOIN pages p ON p.site_id = s.id AND p.name = 'news'
 WHERE s.domain = 'webdesign.co.uk'
ON CONFLICT DO NOTHING;

DO $verify$
DECLARE v_site uuid; v_sections jsonb; v_items int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT sections INTO v_sections FROM pages
     WHERE site_id = v_site AND name = 'news';
    IF v_sections <> '["hero","news-listing","call-to-action"]'::jsonb THEN
        RAISE EXCEPTION 'news page sections not archetype: %', v_sections;
    END IF;

    SELECT count(*) INTO v_items FROM site_work_items
     WHERE site_id = v_site AND item_key = 'needs_page:news'
       AND status NOT IN ('complete','cancelled','rejected');
    IF v_items <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 open needs_page:news item, found %', v_items;
    END IF;

    RAISE NOTICE 'news build item filed: sections=archetype, needs_page:news triaged for page-build-handler';
END $verify$;

COMMIT;
