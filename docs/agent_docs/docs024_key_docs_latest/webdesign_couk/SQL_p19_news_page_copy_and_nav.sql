-- SQL_p19_news_page_copy_and_nav.sql — webdesign.co.uk
--
-- The news page BUILT AND DEPLOYED (07:44 UTC, ~3 min after SQL_p18's item was
-- claimed; 200 on the wire, listing renders real items server-side). Two
-- defects found by reading the ARTEFACT, not the status — instances the
-- handoff's rule predicted:
--
-- 1. THE HERO DESCRIBES A DIFFERENT PAGE. The LLM writer produced "New tools
--    and site updates / Read the release notes for new utilities and technical
--    guides. Modifications to existing tools are documented here." The page is
--    an INDUSTRY news feed (browser releases, CSS, accessibility, design
--    industry); nothing on it is a release note about our own tools. No
--    invented figures this time — an invented PURPOSE.
-- 2. THE CTA HAS NO BUTTONS. The writer supplied `primary_cta`/`secondary_cta`
--    text but not `primary_cta_url`/`secondary_cta_url`, and the template gates
--    each button on `{{if and .primary_cta .primary_cta_url}}` — so "Browse the
--    full list of tools" renders above an EMPTY .cta-buttons div. A promise
--    with no way to take it (bugs_closed/023's cousin: the pairing nobody
--    checked).
--
-- Both fixes write content_data AND rendered_html (handoff §3 shape 4: the
-- assemble path republishes STORED html; the scoped path re-renders from
-- content_data — writing both makes either path correct).
--
-- Copy choices: headline matches the page's own meta_description phrasing
-- ("What is changing in web design"); no apostrophes anywhere, so hand-written
-- rendered_html cannot disagree with Go html/template's escaping on a later
-- re-render. CTA targets are the two nav destinations, both live 200s.
--
-- 3. NAV. site_nav_items has no News row (refresh_nav_tables deleted it 07-27
--    while the page was undeployed — correctly). The page is now deployed with
--    in_header=true, so the orphan_pages discovery check would eventually file
--    nav_drift; robot-hands has such an item sitting at 'detected', so
--    "eventually" is not a schedule. File it at 'triaged' ourselves — the
--    documented chrome-publish path (handoff §8, bugs_open/117): nav-updater
--    rebuilds site_nav_items from deployed pages, re-renders chrome, and queues
--    the page re-renders itself. Shape copied from the completed gamesdesign
--    operator precedent.

\set ON_ERROR_STOP on

BEGIN;

-- 1. Hero: right description of the page, both fields
UPDATE page_components pc
   SET content_data = pc.content_data
         || jsonb_build_object(
              'headline', 'What is changing in web design',
              'subheadline', 'Browser releases, new CSS, accessibility rules and design industry moves, collected automatically from live sources.'),
       rendered_html = replace(replace(pc.rendered_html,
         'New tools and site updates',
         'What is changing in web design'),
         'Read the release notes for new utilities and technical guides. Modifications to existing tools are documented here.',
         'Browser releases, new CSS, accessibility rules and design industry moves, collected automatically from live sources.'),
       updated_at = NOW()
  FROM pages p, sites s, content_components cc
 WHERE p.id = pc.page_id AND s.id = p.site_id AND cc.id = pc.component_id
   AND s.domain = 'webdesign.co.uk' AND p.name = 'news' AND cc.function = 'hero';

-- 2. CTA: give the buttons their URLs, both fields
UPDATE page_components pc
   SET content_data = pc.content_data
         || jsonb_build_object(
              'primary_cta_url', '/tools/index.html',
              'secondary_cta_url', '/learn/index.html'),
       rendered_html = regexp_replace(pc.rendered_html,
         '<div class="cta-buttons">\s*</div>',
         '<div class="cta-buttons">' || chr(10)
           || '            <a href="/tools/index.html" class="cta-btn cta-btn-primary">Browse all tools</a>' || chr(10)
           || '            <a href="/learn/index.html" class="cta-btn cta-btn-secondary">Read the guides</a>' || chr(10)
           || '        </div>'),
       updated_at = NOW()
  FROM pages p, sites s, content_components cc
 WHERE p.id = pc.page_id AND s.id = p.site_id AND cc.id = pc.component_id
   AND s.domain = 'webdesign.co.uk' AND p.name = 'news' AND cc.function = 'call-to-action';

-- 3. Publish the chrome: nav_drift -> nav-updater
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT s.id, 'operator', 'build', 'nav_drift', 'medium',
       'Rebuild nav + chrome: news page deployed 07-29, News row absent from site_nav_items',
       jsonb_build_object(
         'fix', 'The news page (/news/index.html) deployed 07-29 07:44 with in_header=true, but refresh_nav_tables deleted its nav row on 07-27 while the page was unbuilt. Rebuild nav tables from deployed pages, re-render chrome, let nav-updater queue the page re-renders.',
         'check', 'orphan_pages',
         'orphan_type', 'nav_drift',
         'affected_pages', jsonb_build_array('news')),
       20, 'nav-updater', 'triaged', 'webdesign_couk_thread',
       'nav_drift:' || s.id || ':news-page-deployed'
  FROM sites s
 WHERE s.domain = 'webdesign.co.uk'
ON CONFLICT DO NOTHING;

DO $verify$
DECLARE v_site uuid; v_hero text; v_cta text; v_hero_cd text; v_btns int; v_items int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT pc.content_data->>'headline', pc.rendered_html
      INTO v_hero_cd, v_hero
      FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
      JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = 'news' AND cc.function = 'hero';
    IF v_hero_cd <> 'What is changing in web design' THEN
        RAISE EXCEPTION 'hero content_data headline not updated: %', v_hero_cd;
    END IF;
    IF position('What is changing in web design' in v_hero) = 0
       OR position('release notes' in v_hero) > 0 THEN
        RAISE EXCEPTION 'hero rendered_html not rewritten';
    END IF;

    SELECT pc.rendered_html INTO v_cta
      FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
      JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = 'news' AND cc.function = 'call-to-action';
    v_btns := (length(v_cta) - length(replace(v_cta, 'cta-btn-', ''))) / length('cta-btn-');
    IF v_btns < 2 THEN
        RAISE EXCEPTION 'CTA buttons not inserted (found % cta-btn- occurrences)', v_btns;
    END IF;

    SELECT count(*) INTO v_items FROM site_work_items
     WHERE site_id = v_site AND item_key LIKE 'nav_drift:%:news-page-deployed'
       AND status NOT IN ('complete','cancelled','rejected');
    IF v_items <> 1 THEN
        RAISE EXCEPTION 'expected 1 open nav_drift item, found %', v_items;
    END IF;

    RAISE NOTICE 'p19 applied: hero reframed, CTA buttons present (both fields), nav_drift triaged';
END $verify$;

COMMIT;
