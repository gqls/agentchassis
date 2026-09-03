-- 746_advertise_news_feed_enablement.sql
--
-- Enables the news feed for advertise.co.uk (site d991a5b8-428f-44c1-b3eb-e50f44326fd9):
-- authors content_features.news_feed into its current classification spec, and creates
-- the content_sources rows that fill /news/index.html. Owned by the news_feed_ingestion
-- lane (docs/agent_docs/docs024_key_docs_latest/news_feed_ingestion/), routed here by
-- three peers on 2026-09-02: portfolio_positioning (the owner's WebProNews ask),
-- bugs_open/444 (advertise's empty /news/ page is that bug's standing repro), and the
-- 444 plan's enablement contract ("feeds: author content_features.news_feed ... and/or
-- seed content_sources").
--
-- ── WHY BOTH HALVES, AND WHY BY HAND ──────────────────────────────────────────
--
-- The framework selects a site for news on ONE flag:
--   site_specs.aspect='classification' AND is_current
--   AND (data->'content_features'->'news_feed'->>'recommended')::boolean = true
-- read by content-feed-trigger.find_news_sites (6-hourly, LIVE — last completed
-- 2026-09-03 09:10Z) and by the seeder. advertise.co.uk's current classification row
-- (ec005136-e07d-4d5f-aad4-beef6ec02517, domain-research-classifier, 2026-09-02) carries
-- NO content_features key at all, and the site has 0 content_sources rows
-- [MEASURED 2026-09-03], so neither path has ever selected it.
--
-- The framework's own author of that key is evaluate_news_feed
-- (feed_news_recommendation_action.go). Its matcher reads classification.industry /
-- .site_type / .category and domain substrings; this site's signals are '' / editorial /
-- editorial / advertise.co.uk, and verticalNewsMap has no advertising, marketing or
-- media key, so matchVerticalNews returns nil and the action writes NOTHING on no-match.
-- Same wall, same remedy, same reasoning as idea.uk's hand-authored entry of 2026-08-25
-- (idea_uk_vm_site/sql/SQL_2026-08-25_arm_news_feed.sql). Its only carrier is
-- improvement-loop anyway.
--
-- The sources are inserted HERE rather than left to the seeder, for two reasons that
-- are both in seed_content_sources_action.go: (1) the seeder SKIPS source_type 'rss'
-- ("requires manual URL config"), so the owner's WebProNews feed can only ever arrive
-- by hand; and (2) the seeder returns early when the site has ANY active source, so once
-- the rss row exists it would never create the news_search rows the spec names. A spec
-- naming source_types the site will never get is a lie; this file creates exactly what
-- the seeder would have created — same names, same config shape, region='uk' as
-- regionForDomain() now sets for every .uk domain (commit 0a408f8db, live in the
-- v1.0.1358 roll) — plus the rss row the seeder cannot.
--
-- ── THE SOURCES ────────────────────────────────────────────────────────────────
--
-- rss: WebProNews, https://www.webpronews.com/feed/ — the owner's ask, 2026-09-02:
--   "record the webpronews.com rss feed details ... we could add it to the news
--   sources." Re-verified 2026-09-03 12:3xZ: 200, 1,076,370 bytes, 100 items, newest
--   pubDate 12:12Z the same day. The owner endorsed the FEED's content, not the old
--   Drupal advertise.co.uk's wholesale-import pattern (the classifier's own reading of
--   that site: "Articles imported wholesale from WebProNews — no original content").
--   Through this pipeline every item is scored by feed-triage against THIS spec
--   (feed_triage_actions.go: >= 50 relevant, 20–49 review, < 20 rejected, flagged →
--   rejected), so the feed's off-topic majority — sampled 2026-09-03: Anthropic
--   classifiers, FCC robocalls, Gemini 3.8, C# union types — never displays.
--
-- news_search × 5, region=uk: the site's own vertical_landscape spec names the news
--   stream it should carry ("ASA rulings, platform policy changes, and IAB UK data
--   releases provide a real news stream") and its content_direction requires named UK
--   sources (ASA, IAB UK, Ofcom, WARC). The five queries are anchored on those
--   institutions, journalism-shaped on purpose (idea.uk's lesson: institution/authority
--   queries return reporting, market-category queries return vendor pages), and carry
--   region='uk' so Firecrawl's country default of "US" does not apply — this is the
--   first UK site enabled since that fix rolled, and the first live exercise of it.
--
-- No api_news: no LLM-authored items on a site whose proposition is plain, honest
-- explanation of advertising (mission_brief). NB adding api_news to source_types later
-- does NOTHING while sources exist (seeder early-return above) — insert the row
-- directly, in seedAPINewsSources' shape.
--
-- What this file does NOT do: fetch, triage, render or deploy. Those are
-- content-feed-orchestrator's (seed → dispatch_sources → feed-ingester → feed-triage →
-- render_news_section → commit). The trigger's own predicate — reproduced in the
-- post-check below — selects the site on the next 6-hourly tick because every new
-- source has next_fetch_at NULL. /news/index.html (b1cd8ffb, page_type news-index,
-- deployed) fills via render_news_section's queueNewsPageRerenders(): the news-listing
-- component re-resolves query.news_archive from content_feed_items and re-renders from
-- content_data — no LLM, no HTML patching (bugs_open/027).
--
-- ── SCOPE: MEASURED, NOT GUESSED [MEASURED 2026-09-03] ────────────────────────
--   current classification specs for the site : 1 (ec005136…, no content_features)
--   content_sources rows for the site          : 0 (any is_active)
--   content_sources rows naming webpronews     : 0 fleet-wide
--   pages build_status='deployed' for the site : 22 (incl. /news/index.html)
--   fleet content_sources / site_specs         : captured at run time below
--
-- Idempotent by refusal: a re-run fails its own precondition ("already carries
-- content_features") rather than double-writing. Data only — no schema change.
-- Reversal: the UPPERCASE-suffixed sidecar of the same number (refuses if any ingested
-- item has been published to a page). Verify: 746_..._VERIFY.sql (read-only).

BEGIN;

-- Fleet-wide counts BEFORE, so the post-check can prove nothing outside this site moved.
CREATE TEMP TABLE _746_before ON COMMIT DROP AS
  SELECT (SELECT count(*) FROM content_sources) AS n_sources,
         (SELECT count(*) FROM site_specs)      AS n_specs;

-- ── PRECONDITIONS. DO/RAISE, not SELECT: a non-empty SELECT cannot stop a COMMIT. ──
DO $pre$
DECLARE
  n_cur     int;
  cur_id    uuid;
  has_cf    bool;
  n_src     int;
  n_wpn     int;
  n_deploy  int;
  news_pt   text;
BEGIN
  SELECT count(*), (array_agg(id))[1], COALESCE(bool_or(data ? 'content_features'), false)
    INTO n_cur, cur_id, has_cf
    FROM site_specs
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND aspect = 'classification' AND is_current = true;
  IF n_cur <> 1 THEN
    RAISE EXCEPTION '746: expected exactly 1 current classification spec for advertise.co.uk, found %', n_cur;
  END IF;
  IF cur_id <> 'ec005136-e07d-4d5f-aad4-beef6ec02517' THEN
    RAISE EXCEPTION '746: current classification spec is % — not the 2026-09-02 row this file was written against (ec005136…). The classifier re-ran; re-read the row before applying.', cur_id;
  END IF;
  IF has_cf THEN
    RAISE EXCEPTION '746: already applied — advertise.co.uk classification already carries content_features';
  END IF;

  SELECT count(*) INTO n_src FROM content_sources
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9';
  IF n_src <> 0 THEN
    RAISE EXCEPTION '746: advertise.co.uk already has % content_sources rows (any is_active) — the seeder is all-or-nothing and this file assumes zero; stop and read them', n_src;
  END IF;

  SELECT count(*) INTO n_wpn FROM content_sources
   WHERE config::text ILIKE '%webpronews%' OR name ILIKE '%webpronews%';
  IF n_wpn <> 0 THEN
    RAISE EXCEPTION '746: % content_sources rows already name webpronews fleet-wide — another lane added it; read them before duplicating', n_wpn;
  END IF;

  SELECT count(*) INTO n_deploy FROM pages
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9' AND build_status = 'deployed';
  IF n_deploy = 0 THEN
    RAISE EXCEPTION '746: no deployed page for advertise.co.uk — the trigger predicate requires one, so enabling now would select nothing';
  END IF;

  SELECT page_type INTO news_pt FROM pages
   WHERE id = 'b1cd8ffb-47bc-4a43-956c-7851a33c3a4a'
     AND site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9';
  IF news_pt IS DISTINCT FROM 'news-index' THEN
    RAISE EXCEPTION '746: /news/index.html (b1cd8ffb) page_type is % — separate_page=true below would be a lie', COALESCE(news_pt, '<missing>');
  END IF;
END
$pre$;

-- ── 1. Supersede the classifier's row and insert the enriched copy, the way ──
--      evaluate_news_feed does it (supersede + insert, never UPDATE in place).
WITH old AS (
  UPDATE site_specs
     SET is_current = false, superseded_at = now()
   WHERE id = 'ec005136-e07d-4d5f-aad4-beef6ec02517'
     AND site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND aspect = 'classification' AND is_current = true
  RETURNING site_id, data
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT
  old.site_id,
  'classification',
  old.data || jsonb_build_object(
    'content_features', jsonb_build_object(
      'news_feed', jsonb_build_object(
        'recommended', true,
        'separate_page', true,
        'source_types', jsonb_build_array('rss', 'news_search'),
        'vertical_keywords', jsonb_build_array(
          'Advertising Standards Authority rulings',
          'CAP Code advertising rules',
          'IAB UK digital advertising spend',
          'Advertising Association WARC expenditure report',
          'UK advertising industry news'
        ),
        'reason',
          'advertise.co.uk explains advertising to the people who pay for it (mission_brief '
          || '2026-09-02). Its own vertical_landscape names the news stream the site should '
          || 'carry — ASA rulings, platform policy changes and IAB UK data releases — and its '
          || 'content_direction requires named UK sources (ASA, IAB UK, Ofcom, WARC). '
          || 'Hand-authored 2026-09-03 by the news_feed_ingestion lane (migration 746): '
          || 'matchVerticalNews reads industry/site_type/category and domain substrings only; '
          || 'this site''s signals are empty / editorial / editorial / advertise.co.uk and the '
          || 'vertical table has no advertising or marketing key, so the automatic path returns '
          || 'nil and writes nothing — the same wall idea.uk hit on 2026-08-25. source_types: '
          || 'rss is the owner-endorsed WebProNews feed (2026-09-02) — endorsed for its content, '
          || 'not for the old Drupal site''s wholesale-import pattern; feed-triage scores every '
          || 'item against this spec and rejects below 20, holds 20–49 for review, so off-topic '
          || 'items never display. news_search is five UK-region queries (region=uk) anchored '
          || 'on the institutions the landscape names, because the WebProNews feed sampled '
          || '2026-09-03 is broad US tech/business, not UK advertising. No api_news: no '
          || 'LLM-authored items on a site whose proposition is plain, honest explanation. The '
          || 'seeder returns early while any source exists, so to add api_news later insert the '
          || 'row directly rather than editing source_types.'
      )
    )
  ),
  'authored',
  'news-feed-ingestion-lane',
  true,
  'news-feed-ingestion-lane',
  'Migration 746: enables the news feed. Adds content_features.news_feed only; every other '
  || 'key is the 2026-09-02 domain-research-classifier row (ec005136) verbatim, which stays '
  || 'in the table superseded. Reverse with the UPPERCASE-suffixed sidecar of the same number.'
FROM old;

-- ── 2. The sources: the rss row the seeder cannot create, and the five news_search ──
--      rows it would have created (seedNewsSearchSources' exact shape, region=uk).
--      No ON CONFLICT: the precondition asserted zero rows, so a collision here is a
--      concurrent write and must fail loudly, not vanish.
INSERT INTO content_sources (site_id, source_type, name, config) VALUES
  ('d991a5b8-428f-44c1-b3eb-e50f44326fd9', 'rss', 'WebProNews',
   '{"feed_url": "https://www.webpronews.com/feed/", "max_items": 15}'::jsonb),
  ('d991a5b8-428f-44c1-b3eb-e50f44326fd9', 'news_search', 'News Search: Advertising Standards Authority rulings',
   '{"query": "Advertising Standards Authority rulings", "num_results": 10, "region": "uk"}'::jsonb),
  ('d991a5b8-428f-44c1-b3eb-e50f44326fd9', 'news_search', 'News Search: CAP Code advertising rules',
   '{"query": "CAP Code advertising rules", "num_results": 10, "region": "uk"}'::jsonb),
  ('d991a5b8-428f-44c1-b3eb-e50f44326fd9', 'news_search', 'News Search: IAB UK digital advertising spend',
   '{"query": "IAB UK digital advertising spend", "num_results": 10, "region": "uk"}'::jsonb),
  ('d991a5b8-428f-44c1-b3eb-e50f44326fd9', 'news_search', 'News Search: Advertising Association WARC expenditure report',
   '{"query": "Advertising Association WARC expenditure report", "num_results": 10, "region": "uk"}'::jsonb),
  ('d991a5b8-428f-44c1-b3eb-e50f44326fd9', 'news_search', 'News Search: UK advertising industry news',
   '{"query": "UK advertising industry news", "num_results": 10, "region": "uk"}'::jsonb);

-- ── POST-CHECK. Every assertion is DO/RAISE (LANDMINES: bare SELECTs cannot stop a COMMIT). ──
DO $post$
DECLARE
  rec        bool;
  n_types    int;
  n_kw       int;
  same_rest  bool;
  n_src      int;
  n_rss      int;
  n_ns       int;
  n_ns_uk    int;
  n_pending  int;
  picked     int;
  before_src bigint;
  before_sp  bigint;
  after_src  bigint;
  after_sp   bigint;
BEGIN
  -- 1. the spec: flag set, both source types, five keywords, everything else verbatim
  SELECT (data->'content_features'->'news_feed'->>'recommended')::boolean,
         jsonb_array_length(data->'content_features'->'news_feed'->'source_types'),
         jsonb_array_length(data->'content_features'->'news_feed'->'vertical_keywords'),
         (data - 'content_features') = (SELECT data FROM site_specs WHERE id = 'ec005136-e07d-4d5f-aad4-beef6ec02517')
    INTO rec, n_types, n_kw, same_rest
    FROM site_specs
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND aspect = 'classification' AND is_current = true;
  IF rec IS DISTINCT FROM true THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: current classification does not read recommended=true';
  END IF;
  IF n_types <> 2 OR n_kw <> 5 THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: source_types=% keywords=% (want 2 / 5)', n_types, n_kw;
  END IF;
  IF NOT same_rest THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: the new classification row differs from ec005136 outside content_features';
  END IF;
  IF EXISTS (SELECT 1 FROM site_specs WHERE id = 'ec005136-e07d-4d5f-aad4-beef6ec02517' AND (is_current OR superseded_at IS NULL)) THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: the superseded classifier row is still current';
  END IF;

  -- 2. the sources: six, one rss with a feed_url, five news_search all region=uk
  SELECT count(*),
         count(*) FILTER (WHERE source_type = 'rss' AND config->>'feed_url' = 'https://www.webpronews.com/feed/'),
         count(*) FILTER (WHERE source_type = 'news_search'),
         count(*) FILTER (WHERE source_type = 'news_search' AND config->>'region' = 'uk'),
         count(*) FILTER (WHERE next_fetch_at IS NULL AND is_active)
    INTO n_src, n_rss, n_ns, n_ns_uk, n_pending
    FROM content_sources
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9';
  IF n_src <> 6 OR n_rss <> 1 OR n_ns <> 5 OR n_ns_uk <> 5 OR n_pending <> 6 THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: sources total=% rss=% news_search=% region_uk=% pending=% (want 6/1/5/5/6)',
      n_src, n_rss, n_ns, n_ns_uk, n_pending;
  END IF;

  -- 3. the trigger's own predicate (content-feed-trigger.find_news_sites, read from the
  --    live agent_definitions row 2026-09-03), reduced to this site: it must select it.
  SELECT count(*) INTO picked
    FROM sites s
    JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true
     AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
   WHERE s.id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed')
     AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true)
          OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true
                       AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW() + interval '3 hours')));
  IF picked <> 1 THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: the content-feed-trigger predicate would still not select advertise.co.uk (picked=%)', picked;
  END IF;

  -- 4. negative control: nothing outside this site moved
  SELECT n_sources, n_specs INTO before_src, before_sp FROM _746_before;
  SELECT count(*) INTO after_src FROM content_sources;
  SELECT count(*) INTO after_sp  FROM site_specs;
  IF after_src <> before_src + 6 OR after_sp <> before_sp + 1 THEN
    RAISE EXCEPTION '746 POST-CHECK FAILED: fleet counts moved by other than +6 sources / +1 spec (sources % -> %, specs % -> %)',
      before_src, after_src, before_sp, after_sp;
  END IF;

  RAISE NOTICE '746 POST-CHECK PASSED: advertise.co.uk recommended=true, 6 sources (1 rss WebProNews + 5 news_search region=uk), trigger predicate selects it, fleet +6/+1 only.';
END
$post$;

COMMIT;
