-- SQL_p16_ai_in_design_source.sql — webdesign.co.uk
--
-- Owner direction 2026-07-28 (evening), three parts:
--   1. "I can accept empty feeds if they're genuinely empty."
--   2. "Modern AI design will have different pertinent topics, we need to be aware
--      of them. I want to include coverage on that AI influence."
--   3. "Fewer articles on topic is better."
--
-- So: repurpose the dead source rather than pad it, and point it at AI's influence
-- on design.
--
-- WHY `web platform standards` IS THE ONE TO REPURPOSE. It returned **0** items on
-- the first recency-windowed tick, and that is not a tuning failure — W3C/WHATWG
-- specifications do not publish inside a 30-day window, they move yearly. It is
-- also largely REDUNDANT with `CSS and browsers`, which is the healthiest source
-- we have (8 items: Safari Technology Preview 248, Firefox 153/154, Chrome 150).
-- Browser releases are the same subject at a cadence that actually exists. So the
-- slot was carrying no information and duplicating a live one.
--
-- AND WIDENING ITS WINDOW WOULD NOT HAVE HELPED — checked, not assumed.
-- `feed_actions.go:878` hard-codes a 30-day maximum age, so `time_range: year`
-- would fetch a year of material only for the writer to discard everything past
-- day 30. `month` is the only window that is neither wasteful nor lossy.
--
-- THE QUERY FOLLOWS THE PRINCIPLE THE FAILURES TAUGHT, not another guess.
-- Three rounds of retuning produced one rule: queries built from SECTOR-NEUTRAL
-- vocabulary return other sectors' content, and queries built from DOMAIN NOUNS
-- return ours.
--     failed: "AI website builder design tools"      -> vendor landing pages
--             ("builder", "tools" are product categories)
--     failed: "UK web design industry"               -> agency ranking listicles
--     failed: "design agency acquisition merger industry report"
--                                                    -> cross-sector M&A: Amgen
--             layoffs, Kroger-Giant Eagle, an insurance broker that matched only
--             on the word "Creative"
--     worked: "CSS new features browser support", "typeface release design system
--             open source", "web accessibility WCAG UK regulations"
-- So the new query carries the AI qualifier plus design-practice nouns, and
-- NO product-category words (builder/tool/platform/app) and NO transaction words
-- (industry/report/market/acquisition). Those two families are what failed.
--
-- PURGING THE UNDATED ITEMS, and this is a policy change worth stating plainly.
-- 53 of the 78 rows have `source_published_at IS NULL`. Every one predates
-- bugs_closed/127 — they are plain web-search results captured when the feed could
-- not ask for news at all, and they are exactly the material the owner has now
-- ruled against ("fewer articles on topic is better"). Leaving them means the news
-- page curates from a pool that is 68% pre-fix junk.
--
-- There is a second, sharper reason. The age filter is GUARDED:
--
--     if publishedAt != nil {
--         if age > 30*24*time.Hour { ... skip }
--
-- so an item with **no date is never age-checked at all** and is written whatever
-- its age. Undated rows are therefore not merely unverifiable as news — they are
-- the one shape that structurally evades the freshness rule this site now depends
-- on. Deleting them is consistent, not merely tidy.
--
-- Feed items are a regenerable cache, not a record. The 25 dated rows survive.

\set ON_ERROR_STOP on

BEGIN;

-- 1. spec first: keywords must move WITH the source name or seed_content_sources
--    creates a sixth source carrying the bare keyword as its query.
UPDATE site_specs sp
   SET is_current = false, superseded_at = NOW(), updated_at = NOW()
  FROM sites s
 WHERE sp.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND sp.aspect = 'classification' AND sp.is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT s.id, 'classification',
       jsonb_set(prev.data, '{content_features,news_feed,vertical_keywords}',
         $kw$["CSS and browsers",
              "web accessibility UK",
              "AI in design",
              "design industry moves",
              "typography and design systems"]$kw$::jsonb, true),
       'manual', 'webdesign-couk-news-ai',
       'Owner 2026-07-28: cover AI influence on design; empty feeds acceptable if genuinely empty; fewer on-topic articles beat more off-topic ones. Retires "web platform standards" (0 items on the first windowed tick — specs move yearly, not monthly, and it duplicated "CSS and browsers").',
       true, 'webdesign-couk-news-ai'
  FROM sites s
  JOIN LATERAL (
        SELECT data FROM site_specs
         WHERE site_id = s.id AND aspect = 'classification'
         ORDER BY superseded_at DESC NULLS FIRST, created_at DESC LIMIT 1
       ) prev ON true
 WHERE s.domain = 'webdesign.co.uk';

-- 2. drop the retired source's cached items (all 9 are undated pre-fix results)
DELETE FROM content_feed_items cfi
 USING content_sources cs, sites s
 WHERE cfi.source_id = cs.id AND cs.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cs.name = 'News Search: web platform standards';

-- 3. repurpose in place (rename preserves the FK for any future items)
UPDATE content_sources cs
   SET name   = 'News Search: AI in design',
       config = jsonb_build_object(
                  'query', 'AI generative design workflow designers',
                  'num_results', 10,
                  'time_range', 'month'),
       last_fetched_at = NULL, next_fetch_at = NULL, error_count = 0, last_error = NULL,
       updated_at = NOW()
  FROM sites s
 WHERE cs.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cs.name = 'News Search: web platform standards';

-- 4. purge every undated row fleet-of-this-site-wide (see header for the two reasons)
DELETE FROM content_feed_items cfi
 USING sites s
 WHERE cfi.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cfi.source_published_at IS NULL;

DO $verify$
DECLARE v_names text[]; v_keys text[]; v_items int; v_undated int; v_noquery int;
BEGIN
    SELECT array_agg(replace(cs.name,'News Search: ','') ORDER BY replace(cs.name,'News Search: ',''))
      INTO v_names
      FROM content_sources cs JOIN sites s ON s.id = cs.site_id
     WHERE s.domain = 'webdesign.co.uk';

    SELECT array_agg(kw.keyword ORDER BY kw.keyword) INTO v_keys
      FROM site_specs sp
      JOIN sites s ON s.id = sp.site_id
      CROSS JOIN LATERAL jsonb_array_elements_text(
             sp.data->'content_features'->'news_feed'->'vertical_keywords') AS kw(keyword)
     WHERE s.domain = 'webdesign.co.uk' AND sp.aspect = 'classification' AND sp.is_current;

    IF v_names IS DISTINCT FROM v_keys THEN
        RAISE EXCEPTION 'name/keyword lockstep BROKEN — a sixth auto-seeded source will appear. names=% keys=%', v_names, v_keys;
    END IF;

    SELECT count(*) FILTER (WHERE COALESCE(cs.config->>'query','') = ''),
           count(*)
      INTO v_noquery, v_items
      FROM content_sources cs JOIN sites s ON s.id = cs.site_id
     WHERE s.domain = 'webdesign.co.uk';
    IF v_noquery > 0 THEN RAISE EXCEPTION '% source(s) lost their query', v_noquery; END IF;
    IF v_items <> 5 THEN RAISE EXCEPTION 'expected 5 sources, found %', v_items; END IF;

    SELECT count(*), count(*) FILTER (WHERE source_published_at IS NULL)
      INTO v_items, v_undated
      FROM content_feed_items cfi JOIN sites s ON s.id = cfi.site_id
     WHERE s.domain = 'webdesign.co.uk';
    IF v_undated <> 0 THEN RAISE EXCEPTION '% undated rows survived the purge', v_undated; END IF;

    IF (SELECT data->'content_features'->'news_feed'->>'recommended'
          FROM site_specs sp JOIN sites s ON s.id=sp.site_id
         WHERE s.domain='webdesign.co.uk' AND sp.aspect='classification' AND sp.is_current) <> 'true' THEN
        RAISE EXCEPTION 'news_feed.recommended lost — the site would drop out of find_news_sites';
    END IF;

    RAISE NOTICE 'lockstep holds on all 5; % dated items remain, 0 undated.', v_items;
    RAISE NOTICE 'NOT VERIFIED: whether "AI in design" returns journalism, and whether round 3 of "design industry moves" is on-topic. Read the TITLES after the next tick — a count is not a verdict.';
END
$verify$;

COMMIT;
