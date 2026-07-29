-- SQL — dartsonline.com: arm the news feed (owner decision D2, 2026-07-29)
--
-- THE GATE. `content-feed-trigger.find_news_sites` selects sites by:
--   site_specs.aspect='classification' AND is_current
--   AND (data->'content_features'->'news_feed'->>'recommended')::boolean = true
--   AND EXISTS (a deployed page)          -- dartsonline already passes this leg
-- dartsonline's `content_features` key does not exist, so it has never been selected.
--
-- WHY HAND-AUTHORED rather than letting the platform decide: the writer of that key is
-- `evaluate_news_feed` (feed_news_recommendation_action.go), whose `matchVerticalNews`
-- (:388) builds its signals from classification.industry / .site_type / .category and
-- domain substrings. It never reads `industry_tags`. This site's tags are the only place
-- the words darts-retail / competitive-sport / sports-equipment appear; its category and
-- site_type both read "ecommerce", which matches no key in verticalNewsMap. So the
-- automatic path cannot reach a decision here, and its only live caller is
-- improvement-loop.enrich_news_feed, whose scheduled task has been disabled since
-- 2026-05-02. Hand-authoring is the established precedent: relojistas.com
-- (source_agent 'relojistas-rebuild') and vetcomparison.uk ('sites-vetcomparison-thread')
-- both carry hand-written news_feed blocks.
--
-- vertical_keywords are DARTS-SPECIFIC on purpose. If `industry_tags` were fed to the
-- matcher, "sports-equipment" would token-match the generic "sports" entry, whose
-- keywords are "sports news / match results / tournament / league standings" — true of
-- darts and useless for it. A darts audience wants the PDC circuit by name.
--
-- source_types drives seeding: `seed_content_sources_action.go:216-227` deliberately
-- SKIPS 'rss' and 'scrape' ("requires manual URL config"), so listing rss here does not
-- create anything by itself — curated feeds are inserted by hand in a later step, AFTER
-- the first orchestrator run, because the seeder is all-or-nothing (:92-111): if ANY
-- active source already exists it skips seeding entirely and the search/api sources
-- would never be created.
--
-- separate_page=true means a page_type='news-index' page is expected;
-- render_news_section_action.go:216 gates data/news-archive.json on that page existing.
-- bugs_closed/141 (news-index can enter nav) is CLOSED and proven live on v1.0.1198,
-- so the canonical /news/index.html shape is safe to create.
--
-- ALSO CORRECTED HERE: `reasoning` and `detected_signals` were built on the Australian
-- and Portland conflation that produced the false identity (see
-- SQL_2026-07-29_identity_truth_reset.sql). They are research provenance rather than
-- rendered copy, but they are what a future planner would read to re-derive the site's
-- purpose, so leaving them would re-seed the same error. category/site_type/
-- recommended_builder are deliberately UNTOUCHED — they drive builder selection and the
-- site is still commerce-shaped (affiliate retail, owner decision D1).

BEGIN;

CREATE TABLE IF NOT EXISTS bak_darts_classification_20260729 AS
SELECT * FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND aspect = 'classification';

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'classification' AND is_current = true;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'classification',
  b.data || jsonb_build_object(
    'content_features', jsonb_build_object(
      'news_feed', jsonb_build_object(
        'recommended', true,
        'separate_page', true,
        'source_types', jsonb_build_array('rss', 'news_search', 'api_news'),
        'vertical_keywords', jsonb_build_array(
          'PDC darts news',
          'World Darts Championship',
          'Premier League Darts',
          'darts results and rankings',
          'darts equipment and barrel releases'
        ),
        'reason', 'Darts is an event-driven sport with a dense professional calendar '
          || '(PDC circuit, weekly Premier League nights, World Championship) and good '
          || 'published RSS supply. The audience returns for results and gear news, which '
          || 'is the same audience the buying guides serve. Hand-authored 2026-07-29: '
          || 'matchVerticalNews reads industry/site_type/category and domain substrings '
          || 'only, never industry_tags, so this site (category=ecommerce, tags carrying '
          || 'darts-retail/competitive-sport) can never match a vertical automatically.'
      )
    ),
    'reasoning', 'A specialist darts site. NOTE (corrected 2026-07-29): the original '
      || 'classification reasoned from the related .com.au entity and a Portland '
      || 'distribution address, and that conflation produced an identity belonging to '
      || 'other companies — see SQL_2026-07-29_identity_truth_reset.sql. Those signals '
      || 'describe OTHER businesses, not this one. What is true of this site: it is a '
      || 'UK-based, online-only darts publication holding no stock, publishing spec-first '
      || 'buying guides and darts news, and intended to carry an affiliate equipment feed.',
    'detected_signals', jsonb_build_array(
      'Specialist single-sport focus (darts) rather than general sporting goods',
      'Buying-guide content set already planned: tungsten percentage, barrel weight, '
        || 'shaft length, flight shapes, grip styles, steel-tip vs soft-tip',
      'Setup-builder tool planned — configure barrel, shaft and flight as one system',
      'Strong published news supply in the vertical (PDC circuit and darts press)',
      'CORRECTED 2026-07-29: signals previously listed here (Trustpilot page, Facebook '
        || 'presence, Portland sales address, brand stock lists) all belong to '
        || 'dartsonline.com.au or darts.com and were never evidence about this site.'
    )
  ),
  'authored',
  'dartsonline-traffic-workstream',
  true,
  'dartsonline-traffic-workstream',
  'Arms the news feed per owner decision D2 (aggregated feed + our own analysis) and '
    || 'corrects the reasoning/detected_signals that were derived from other companies. '
    || 'category/site_type/recommended_builder untouched. Prior rows in '
    || 'bak_darts_classification_20260729.'
FROM site_specs b
WHERE b.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND b.aspect = 'classification' AND b.is_current = false
ORDER BY b.created_at DESC LIMIT 1;

COMMIT;

-- Verify: the gate query itself, run for this site
SELECT s.domain,
       (ss.data->'content_features'->'news_feed'->>'recommended')::boolean AS recommended,
       ss.data->'content_features'->'news_feed'->>'separate_page' AS separate_page,
       ss.data->'content_features'->'news_feed'->'source_types' AS source_types,
       EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.build_status='deployed') AS has_deployed_page,
       ss.data->>'category' AS category_unchanged,
       ss.data::text ILIKE '%Portland%' AS mentions_portland
FROM sites s
JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect='classification' AND ss.is_current
WHERE s.id = '5fe8785b-223d-41a3-88ee-c07187622381';
