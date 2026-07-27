-- SQL_p8_news_section.sql — webdesign.co.uk, phase 2 W5
--
-- Enable a news section on THIS SITE ONLY, using the pipeline that already
-- exists. Owner go-ahead given 2026-07-27.
--
-- SCOPE — READ THIS BEFORE ASSUMING IT MEANS MORE. The owner asked for "a news
-- section". That is NOT `features_open/005`, which is a fleet programme to
-- onboard ~37 new pool domains and arm a shared content pool, parked by the
-- owner with *"we're not quite ready to onboard more domains."* Nothing here
-- touches a pool, onboards a domain, or alters another site. One site, its own
-- sources, its own page.
--
-- WHAT ALREADY EXISTS (found, not built — this project's most expensive mistake
-- was filing a bug about a capability that had been in production for weeks):
--   * components `news-listing`, `latest-news`, `blog-listing`, all active;
--   * `content_sources` / `content_items` / `content_feed_items` tables;
--   * scheduled task `content-feed-refresh`, every 6h, enabled;
--   * the working shape, copied from robot-hands.com: a page with
--     `page_type='news-index'` whose sections include `news-listing`.
--
-- SOURCE TYPE. The fleet's dominant type is `news_search` (13 of 25 active) —
-- `config: {"query": "...", "num_results": N}`. Chosen over `rss` deliberately:
-- a query states the site's topic focus and degrades gracefully, whereas a
-- curated RSS list silently rots as feeds move and quietly encodes an editorial
-- position nobody reviewed.
--
-- THE QUERIES ARE AN EDITORIAL CHOICE AND THE OWNER SHOULD PRUNE THEM. They are
-- written to match the Phase 2 brief — UK, today's problems, visual craft, AI in
-- web design — and are listed here in full so they can be argued with. They are
-- not tuned; nobody has seen what they return yet. Expect to revise after the
-- first fetch.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Sources
-- ---------------------------------------------------------------------------
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval, is_active)
SELECT s.id, 'news_search', v.name,
       jsonb_build_object('query', v.query, 'num_results', 10),
       '06:00:00'::interval, true
FROM sites s
CROSS JOIN (VALUES
    ('News Search: UK web design',        'UK web design industry'),
    ('News Search: AI web design tools',  'AI website builder design tools'),
    ('News Search: CSS and browsers',     'CSS new features browser support'),
    ('News Search: web accessibility UK', 'web accessibility WCAG UK regulations'),
    ('News Search: web design trends',    'web design visual trends typography colour')
) AS v(name, query)
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (
      SELECT 1 FROM content_sources cs
       WHERE cs.site_id = s.id AND cs.name = v.name
  );

-- ---------------------------------------------------------------------------
-- 2. The news page
-- ---------------------------------------------------------------------------
-- rebuild_policy='generic' ON PURPOSE, unlike this site's other 97 pages. The
-- news listing is machine-maintained from content_items and must be free to
-- re-render as the feed changes; an owned page would freeze it. build_status
-- 'planned' so the build pipeline picks it up and composes it.
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description,
                   sections, rebuild_policy, build_status,
                   in_header, in_footer, nav_label, nav_order)
SELECT s.id, 'news', '/news/index.html',
       'News | webdesign.co.uk',
       'news-index', 'active',
       'What is changing in web design — browsers, CSS, accessibility rules and AI tooling, with a UK slant.',
       '["news-listing"]'::jsonb,
       'generic', 'planned',
       true, true, 'News', 40
FROM sites s
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (
      SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.name = 'news'
  );

-- ---------------------------------------------------------------------------
-- 3. Nav — added only because the page now exists.
-- ---------------------------------------------------------------------------
-- A nav entry pointing at an unbuilt page is bugs_open/049's exact shape, so
-- this runs after the page row and the page must be BUILT before the chrome is
-- re-rendered. See the verify block.
INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
SELECT s.id, g.id, 'News', p.url, p.id, 'page_link', p.nav_order, 'active'
FROM sites s
JOIN site_nav_groups g ON g.site_id = s.id AND g.group_key = 'primary'
JOIN pages p ON p.site_id = s.id AND p.name = 'news'
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (
      SELECT 1 FROM site_nav_items ni WHERE ni.site_id = s.id AND ni.url = p.url
  );

DO $verify$
DECLARE v_site uuid; v_src int; v_page int; v_nav int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT count(*) INTO v_src FROM content_sources
     WHERE site_id = v_site AND is_active;
    IF v_src < 5 THEN RAISE EXCEPTION 'expected 5 active sources, found %', v_src; END IF;

    SELECT count(*) INTO v_page FROM pages
     WHERE site_id = v_site AND name = 'news' AND page_type = 'news-index';
    IF v_page <> 1 THEN RAISE EXCEPTION 'news page missing or duplicated (%)', v_page; END IF;

    SELECT count(*) INTO v_nav FROM site_nav_items ni
      JOIN site_nav_groups g ON g.id = ni.group_id
     WHERE ni.site_id = v_site AND g.group_key = 'primary';
    IF v_nav <> 4 THEN RAISE EXCEPTION 'expected 4 primary nav items, found %', v_nav; END IF;

    -- Nothing outside this site may have been touched.
    IF EXISTS (SELECT 1 FROM sites WHERE status = 'pool' AND updated_at > now() - interval '2 minutes') THEN
        RAISE EXCEPTION 'a pool site was modified — this script must touch one site only';
    END IF;

    RAISE NOTICE 'news enabled for webdesign.co.uk: 5 sources, page /news/index.html (planned), nav item added';
    RAISE NOTICE 'NEXT: content-feed-refresh (every 6h) populates content_items; then the page must BUILD before chrome is re-rendered, or the nav points at a 404';
END
$verify$;

COMMIT;
