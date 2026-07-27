-- SQL_p9_news_feed_classification_flag.sql — webdesign.co.uk, phase 2 W5 (part 2)
--
-- WHY THIS EXISTS: SQL_p8 created the sources, the page and the nav row, and the
-- HANDOFF then said "wait for the feed". The feed would have waited forever.
--
-- THE BLOCKER, found 2026-07-27 by reading the trigger rather than the clock.
-- `content-feed-refresh` (every 6h) targets agent `content-feed-trigger`, whose
-- FIRST step `find_news_sites` enumerates sites with this predicate:
--
--     JOIN site_specs ss ON ss.site_id = s.id
--        AND ss.aspect = 'classification' AND ss.is_current = true
--        AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
--
-- webdesign.co.uk HAS a current classification spec, but it has no
-- `content_features` key at all — the domain-research-classifier that wrote it on
-- 2026-07-25 did not evaluate news. So the predicate yields NULL, not true, and
-- the site is never enumerated. Verified before this change:
--
--     SELECT s.domain FROM sites s JOIN site_specs ss ON ... recommended = true;
--      ai-agent-orchestration.com / gaswholesalers.com / relojistas.com /
--      robot-hands.com / vetcomparison.uk        <- 5 rows, ours absent
--
-- Creating content_sources rows is NOT what arms a feed. The spec flag is.
--
-- PRECEDENT: robot-hands.com was armed exactly this way on 2026-07-08 — a new
-- superseding classification version adding content_features.news_feed, with
-- source='manual-recovery'. This follows that row's shape deliberately.
--
-- ---------------------------------------------------------------------------
-- THE ONE TRAP IN THIS FILE — read before changing vertical_keywords.
-- ---------------------------------------------------------------------------
-- content-feed-orchestrator runs `seed_sources` BEFORE `dispatch_sources`, and
-- `seed_content_sources` creates ONE content_source PER vertical keyword, named
--
--     fmt.Sprintf("News Search: %s", keyword)      (seed_content_sources_action.go:262)
--     ... INSERT ... ON CONFLICT (site_id, name) DO NOTHING
--
-- So the keywords below are NOT free text. They are chosen to be exactly the
-- name suffixes of the five sources SQL_p8 already created, so that the
-- auto-seeder collides on `idx_cs_site_name` and does nothing. Change one
-- character and the auto-seeder silently adds a SIXTH source whose query is the
-- bare keyword — quietly overriding the editorial queries SQL_p8 argued for.
--
-- The source `name` is a label; the search string lives in `config.query` and is
-- untouched by this file. That separation is why the collision is safe.
--
-- source_types is ["news_search"] ONLY, on purpose:
--   * "api_news" would create an extra xAI/Grok-backed "LLM News: webdesign.co.uk"
--     source we never chose (seed_content_sources_action.go:296);
--   * "rss" and "scrape" are skipped by the seeder anyway (they need manual URLs).
--
-- LATENT RISK, recorded not fixed: find_news_sites ends `ORDER BY s.domain
-- LIMIT 5` with no rotation, and this change makes webdesign.co.uk the SIXTH
-- recommended site — alphabetically LAST of the six. Whenever five or more sites
-- are due in the same tick, ours is the one dropped, every time. It is not biting
-- at the next tick (13:49 UTC: only relojistas and vetcomparison are due, so 3 of
-- 5 slots are used), because ingestion sets next_fetch_at = now()+6h a few
-- minutes AFTER the trigger, which staggers sites apart. But that is luck, not
-- design. See NOTES_webdesign_couk.md.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Supersede the current classification spec
-- ---------------------------------------------------------------------------
-- idx_site_specs_current is UNIQUE on (site_id, aspect) WHERE is_current, so the
-- old row must be retired before the new one lands. Same transaction.
UPDATE site_specs ss
   SET is_current = false, superseded_at = now()
  FROM sites s
 WHERE s.id = ss.site_id
   AND s.domain = 'webdesign.co.uk'
   AND ss.aspect = 'classification'
   AND ss.is_current = true;

-- ---------------------------------------------------------------------------
-- 2. New version = old data + content_features.news_feed
-- ---------------------------------------------------------------------------
-- The existing `data` is carried forward with `||` rather than retyped. Retyping
-- a 30-line classification to add one key is how a transcription error ships.
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current)
SELECT ss.site_id,
       'classification',
       ss.data || jsonb_build_object(
           'content_features', jsonb_build_object(
               'news_feed', jsonb_build_object(
                   'recommended',  true,
                   'separate_page', true,
                   'source_types',  jsonb_build_array('news_search'),
                   'vertical_keywords', jsonb_build_array(
                       'UK web design',
                       'AI web design tools',
                       'CSS and browsers',
                       'web accessibility UK',
                       'web design trends'
                   ),
                   'reason', 'Phase 2 news section, owner go-ahead 2026-07-27. '
                          || 'The keywords are the NAME SUFFIXES of the five sources SQL_p8 '
                          || 'created, so seed_content_sources collides on (site_id, name) and '
                          || 'no auto-generated source overrides the editorial queries.'
               )
           )
       ),
       'manual-recovery',
       'webdesign-couk-phase2',
       'Arms the news feed 2026-07-27. SQL_p8 created sources/page/nav but the '
       || 'content-feed-trigger enumerates on content_features.news_feed.recommended, '
       || 'which this spec lacked entirely — so the feed could never have fired. '
       || 'Shape copied from robot-hands.com 2026-07-08.',
       true
  FROM site_specs ss
  JOIN sites s ON s.id = ss.site_id
 WHERE s.domain = 'webdesign.co.uk'
   AND ss.aspect = 'classification'
   AND ss.is_current = false
 ORDER BY ss.superseded_at DESC NULLS LAST, ss.created_at DESC
 LIMIT 1;

-- ---------------------------------------------------------------------------
-- 3. Verify — assert the TRIGGER's own predicate, not a restatement of it
-- ---------------------------------------------------------------------------
DO $verify$
DECLARE
    v_site   uuid;
    v_cur    int;
    v_enum   int;
    v_src    int;
    v_lost   int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    -- exactly one current classification spec
    SELECT count(*) INTO v_cur FROM site_specs
     WHERE site_id = v_site AND aspect = 'classification' AND is_current;
    IF v_cur <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 current classification spec, found %', v_cur;
    END IF;

    -- nothing from the old spec was dropped by the merge
    SELECT count(*) INTO v_lost
      FROM (
        SELECT jsonb_object_keys(data) AS k FROM site_specs
         WHERE site_id = v_site AND aspect = 'classification' AND NOT is_current
         ORDER BY superseded_at DESC NULLS LAST, created_at DESC LIMIT 1
      ) old
     WHERE NOT EXISTS (
        SELECT 1 FROM site_specs cur, jsonb_object_keys(cur.data) ck
         WHERE cur.site_id = v_site AND cur.aspect = 'classification' AND cur.is_current
           AND ck = old.k
     );
    IF v_lost > 0 THEN
        RAISE EXCEPTION 'the merge dropped % key(s) from the previous classification', v_lost;
    END IF;

    -- THE REAL TEST: run the trigger's own enumeration predicate verbatim.
    SELECT count(*) INTO v_enum
      FROM sites s
      JOIN site_specs ss ON ss.site_id = s.id
                        AND ss.aspect = 'classification'
                        AND ss.is_current = true
                        AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
     WHERE s.domain = 'webdesign.co.uk'
       AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed')
       AND (NOT EXISTS (SELECT 1 FROM content_sources cs
                         WHERE cs.site_id = s.id AND cs.is_active = true)
            OR EXISTS (SELECT 1 FROM content_sources cs
                        WHERE cs.site_id = s.id AND cs.is_active = true
                          AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())));
    IF v_enum <> 1 THEN
        RAISE EXCEPTION 'find_news_sites still would NOT enumerate webdesign.co.uk (got %)', v_enum;
    END IF;

    -- the five editorial sources are untouched and unduplicated
    SELECT count(*) INTO v_src FROM content_sources
     WHERE site_id = v_site AND is_active AND source_type = 'news_search';
    IF v_src <> 5 THEN
        RAISE EXCEPTION 'expected the 5 SQL_p8 sources, found %', v_src;
    END IF;

    RAISE NOTICE 'webdesign.co.uk is now enumerable by content-feed-trigger; 5 sources intact';
    RAISE NOTICE 'NEXT: next tick ~13:49 UTC. Then content_feed_items must be non-zero BEFORE the page builds, and the page must build BEFORE chrome is re-rendered.';
END
$verify$;

COMMIT;
