-- SQL_p21_news_listing_is_news.sql — webdesign.co.uk
--
-- OWNER, 2026-07-29: "The news page is not an Updates and Change log page,
-- it's news." He was reading the NEWS-LISTING section, not the hero: SQL_p19
-- fixed the hero (which had the same fault, a changelog framing) and I did not
-- check the section below it. The listing's own stored headline still said:
--
--     headline    "Updates and changelog"
--     subheadline "A record of new tools, feature additions and fixes."
--
-- under which sit browser releases, CSS news and design-industry stories from
-- live third-party sources. The copy describes a product changelog this site
-- does not have and has never had.
--
-- MY MISS, recorded: p19's verify block asserted the HERO's text and the CTA's
-- buttons and passed. One component's copy was checked; its sibling on the same
-- page, with the same defect, was not. Reading the page as a VISITOR (all of its
-- text, in order) is what found it — the same check that found p19's faults, not
-- run widely enough.
--
-- Also sets the hero image. `/assets/images/hero.jpg` was referenced by the
-- hero's content_data since the build but returned 404 — the file never existed
-- (two 'hero' assets sat as undeployed_asset work items since 07-26). The real
-- illustration is now committed to gqls/sites; this script only ensures the
-- stored markup keeps pointing at it. Both fields, per the standing rule.

\set ON_ERROR_STOP on

BEGIN;

UPDATE page_components pc
   SET content_data = pc.content_data
         || jsonb_build_object(
              'headline', 'Web design news',
              'subheadline', 'Browser releases, CSS, accessibility and design industry stories, gathered from live sources and refreshed twice a day.'),
       rendered_html = replace(replace(pc.rendered_html,
         'Updates and changelog',
         'Web design news'),
         'A record of new tools, feature additions and fixes.',
         'Browser releases, CSS, accessibility and design industry stories, gathered from live sources and refreshed twice a day.'),
       updated_at = NOW()
  FROM pages p, sites s, content_components cc
 WHERE p.id = pc.page_id AND s.id = p.site_id AND cc.id = pc.component_id
   AND s.domain = 'webdesign.co.uk' AND p.name = 'news' AND cc.function = 'news-listing';

DO $verify$
DECLARE v_site uuid; v_cd text; v_html text;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT pc.content_data->>'headline', pc.rendered_html INTO v_cd, v_html
      FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
      JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = 'news' AND cc.function = 'news-listing';

    IF v_cd <> 'Web design news' THEN
        RAISE EXCEPTION 'listing content_data headline not updated: %', v_cd;
    END IF;
    IF position('Updates and changelog' in v_html) > 0
       OR position('feature additions and fixes' in v_html) > 0 THEN
        RAISE EXCEPTION 'changelog copy still present in rendered_html';
    END IF;
    IF position('Web design news' in v_html) = 0 THEN
        RAISE EXCEPTION 'new headline absent from rendered_html';
    END IF;

    -- the whole-page check this file exists because I skipped: NO component on
    -- the news page may describe the site's own releases.
    IF EXISTS (
        SELECT 1 FROM page_components pc
          JOIN pages p ON p.id = pc.page_id
         WHERE p.site_id = v_site AND p.name = 'news'
           AND (pc.rendered_html ILIKE '%changelog%'
                OR pc.rendered_html ILIKE '%release notes%')
    ) THEN
        RAISE EXCEPTION 'a news-page component still carries changelog framing';
    END IF;

    RAISE NOTICE 'news listing reads as news; no changelog framing anywhere on the page';
END $verify$;

COMMIT;
