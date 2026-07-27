-- SQL_p14_retune_news_queries.sql — webdesign.co.uk
--
-- Owner directive 2026-07-27: "aim the news queries at the industry and any new
-- big developments in design and web design."
--
-- WHAT THE FIRST INGESTION ACTUALLY TAUGHT US. 50 items landed on the 19:49 tick,
-- 10 per source, and the split was not subtle:
--
--   WORKED   "CSS new features browser support"        -> Can I Use, MDN, real feature news
--   WORKED   "web accessibility WCAG UK regulations"   -> WCAG 2.2, UK public-sector duty, law
--   FAILED   "AI website builder design tools"         -> vendor landing pages ("Free AI Website Builder")
--   FAILED   "UK web design industry"                  -> 9/10 "Top Web Design Agencies in the UK 2026"
--   FAILED   "web design visual trends typography colour" -> "Top 10 Trends 2026" listicles
--
-- The rule that falls out: **a query that names a TECHNOLOGY or an AUTHORITY
-- returns journalism; a query that names a MARKET CATEGORY returns marketing.**
-- "builder", "tools", "trends" and the exact phrase "UK web design" are the terms
-- vendors and agencies buy and optimise for, so a search engine hands back their
-- landing pages and ranking listicles. The two that worked named CSS/browsers and
-- WCAG/UK-regulations — things that standards bodies and journalists write about.
--
-- AND ONE OF THEM WAS A COMPLIANCE RISK, not merely weak. The standing rail on
-- this site is **never publish comparative rankings of named agencies** — different
-- risk class, and it destroys the neutrality the buying-design section depends on.
-- "UK web design industry" was returning almost nothing else. Retiring it is the
-- point of this file, not a side effect.
--
-- THE LOCKSTEP LANDMINE — the reason this file touches two tables.
-- `content-feed-orchestrator` runs `seed_sources` BEFORE `dispatch_sources`, and
-- `seed_content_sources` creates one source per `vertical_keyword`, named
-- "News Search: <keyword>", ON CONFLICT (site_id, name) DO NOTHING. The keywords
-- are therefore load-bearing: they exist to COLLIDE with the editorial source names
-- so the auto-seeder no-ops. Rename a source without renaming its keyword and the
-- seeder helpfully creates a SIXTH source carrying the bare keyword as its query —
-- diluting the editorial ones and silently reintroducing exactly the generic
-- phrasing this file exists to remove. So names and keywords move together, and the
-- verify block below asserts the two sets are identical.
--
-- RENAME, DO NOT DELETE AND RECREATE. `content_feed_items.source_id` is a foreign
-- key; renaming preserves it and the 20 good items from the two kept sources.
--
-- THE 30 ITEMS FROM THE RETIRED QUERIES ARE DELETED. They are agency ranking
-- listicles and vendor landing pages sitting in the exact table a news page would
-- curate from. Leaving them is leaving the rail-breaking material in place and
-- trusting a later triage step to refuse it. They are a regenerable cache, not a
-- record. The 20 items from the two KEPT sources are untouched.

\set ON_ERROR_STOP on

BEGIN;

-- 1. supersede the classification spec (unique partial index on site_id+aspect
--    WHERE is_current, so the old row must be closed before the new one lands)
UPDATE site_specs sp
   SET is_current = false, superseded_at = NOW(), updated_at = NOW()
  FROM sites s
 WHERE sp.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND sp.aspect = 'classification' AND sp.is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT s.id, 'classification',
       jsonb_set(prev.data, '{content_features,news_feed}', $nf${
         "reason": "Phase 2 news section. Retuned 2026-07-27 on the owner directive to aim at the industry and big developments, after the first ingestion showed market-category queries return vendor landing pages and agency ranking listicles while technology/authority queries return journalism. The keywords are the NAME SUFFIXES of the five editorial sources, so seed_content_sources collides on (site_id, name) and cannot add a sixth generic source.",
         "recommended": true,
         "source_types": ["news_search"],
         "separate_page": true,
         "vertical_keywords": [
           "CSS and browsers",
           "web accessibility UK",
           "web platform standards",
           "design industry moves",
           "typography and design systems"
         ]
       }$nf$::jsonb, true),
       'manual', 'webdesign-couk-news-retune',
       'Owner directive: aim the news at the industry and new big developments in design and web design. Retires "UK web design" (rail: never publish comparative rankings of named agencies), "AI web design tools" (vendor landing pages) and "web design trends" (listicles).',
       true, 'webdesign-couk-news-retune'
  FROM sites s
  JOIN LATERAL (
        SELECT data FROM site_specs
         WHERE site_id = s.id AND aspect = 'classification'
         ORDER BY superseded_at DESC NULLS FIRST, created_at DESC LIMIT 1
       ) prev ON true
 WHERE s.domain = 'webdesign.co.uk';

-- 2. drop the cached items from the three retired queries, BY source_id (stable
--    across the rename below)
DELETE FROM content_feed_items cfi
 USING content_sources cs, sites s
 WHERE cfi.source_id = cs.id AND cs.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cs.name IN ('News Search: UK web design',
                   'News Search: AI web design tools',
                   'News Search: web design trends');

-- 3. retune the three, in place. Each names a thing that PUBLISHES — a standards
--    body, a corporate event, a release — rather than a product category.
UPDATE content_sources cs
   SET name   = 'News Search: web platform standards',
       config = jsonb_build_object('query', 'web platform standards W3C WHATWG specification', 'num_results', 10),
       last_fetched_at = NULL, next_fetch_at = NULL, error_count = 0, last_error = NULL,
       updated_at = NOW()
  FROM sites s
 WHERE cs.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cs.name = 'News Search: AI web design tools';

UPDATE content_sources cs
   SET name   = 'News Search: design industry moves',
       config = jsonb_build_object('query', 'design agency acquisition merger industry report', 'num_results', 10),
       last_fetched_at = NULL, next_fetch_at = NULL, error_count = 0, last_error = NULL,
       updated_at = NOW()
  FROM sites s
 WHERE cs.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cs.name = 'News Search: UK web design';

UPDATE content_sources cs
   SET name   = 'News Search: typography and design systems',
       config = jsonb_build_object('query', 'typeface release design system open source', 'num_results', 10),
       last_fetched_at = NULL, next_fetch_at = NULL, error_count = 0, last_error = NULL,
       updated_at = NOW()
  FROM sites s
 WHERE cs.site_id = s.id AND s.domain = 'webdesign.co.uk'
   AND cs.name = 'News Search: web design trends';

DO $verify$
DECLARE v_names text[]; v_keys text[]; v_items int; v_srcs int;
BEGIN
    -- ORDER BY must repeat the EXPRESSION. `ORDER BY 1` inside array_agg is the
    -- constant 1, not a positional reference, so it silently does not sort — this
    -- verify block failed its own first run on two identical sets in different
    -- orders. The transaction rolled back, which is the block working.
    SELECT array_agg(replace(cs.name, 'News Search: ', '')
                     ORDER BY replace(cs.name, 'News Search: ', '')), count(*)
      INTO v_names, v_srcs
      FROM content_sources cs JOIN sites s ON s.id = cs.site_id
     WHERE s.domain = 'webdesign.co.uk';

    SELECT array_agg(k ORDER BY k) INTO v_keys
      FROM site_specs sp JOIN sites s ON s.id = sp.site_id,
           jsonb_array_elements_text(sp.data->'content_features'->'news_feed'->'vertical_keywords') k
     WHERE s.domain = 'webdesign.co.uk' AND sp.aspect = 'classification' AND sp.is_current;

    IF v_srcs <> 5 THEN
        RAISE EXCEPTION 'expected 5 sources, found %', v_srcs;
    END IF;
    -- THE load-bearing assertion: drift here summons a sixth auto-seeded source
    IF v_names IS DISTINCT FROM v_keys THEN
        RAISE EXCEPTION 'source names and vertical_keywords have DRIFTED. names=% keys=%', v_names, v_keys;
    END IF;

    IF (SELECT data->'content_features'->'news_feed'->>'recommended'
          FROM site_specs sp JOIN sites s ON s.id=sp.site_id
         WHERE s.domain='webdesign.co.uk' AND sp.aspect='classification' AND sp.is_current) <> 'true' THEN
        RAISE EXCEPTION 'news_feed.recommended lost — the site would drop out of find_news_sites entirely';
    END IF;

    SELECT count(*) INTO v_items
      FROM content_feed_items cfi JOIN sites s ON s.id = cfi.site_id
     WHERE s.domain = 'webdesign.co.uk';

    RAISE NOTICE 'names and keywords match on all 5. % cached items remain (the two kept sources).', v_items;
    RAISE NOTICE 'The three retuned sources are re-armed (next_fetch_at NULL) and will refetch on the next tick.';
    RAISE NOTICE 'NOT VERIFIED YET: whether the new queries return journalism. Read the titles after the next ingestion before building the news page.';
END
$verify$;

COMMIT;
