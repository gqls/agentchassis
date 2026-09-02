-- 691_uk_news_search_region_default.sql
--
-- bugs_open/316 rider (owner ask, CONTRIB_2026-08-31_from_loanzy_lane_owner_
-- wants_uk_news_default_for_uk_tlds...md): "The news is from America. I'd
-- like it to be UK news for all .co.uk and .uk sites, perhaps as a flag with
-- a UK default." Backfills content_sources.config->>'region' = 'uk' onto
-- every EXISTING news_search source belonging to a .uk/.co.uk site.
--
-- Companion to a Go change in the same commit: SeedContentSourcesAction now
-- sets this same key for every NEWLY seeded news_search source on a .uk
-- domain. That insert is `ON CONFLICT (site_id, name) DO NOTHING`, so it can
-- never touch a source that already exists — without this backfill the
-- owner's ask would only take effect for sites created after the roll, not
-- "all .co.uk and .uk sites" as asked.
--
-- Mechanism this key feeds (verified against the providers' own API docs,
-- 2026-09-02, not assumed): WebSearchAction -> web-search-adapter ->
-- providers.SearchOptions.Region -> FirecrawlProvider sets payload["country"],
-- ScrapingBeeProvider sets country_code. PRIMARY_SEARCH_PROVIDER=firecrawl is
-- live (checked against the running deployment), and Firecrawl's own default
-- when the key is absent is "US" — the literal, load-bearing cause of the
-- complaint, not a downstream ranking effect. Full design:
-- docs/agent_docs/docs024_key_docs_latest/news_feed_ingestion/PLAN_2026-09-02_news_feed_ingestion.md
--
-- ── SCOPE: MEASURED, NOT GUESSED [MEASURED 2026-09-02] ──
--
-- content_sources rows with source_type='news_search' whose site's domain
-- ends in '.uk' (a plain suffix match — '.co.uk' ends in '.uk' too, so no
-- separate branch is needed, the same reasoning as isBlockedDomain's own
-- suffix check in platform/orchestration/actions/helpers.go):
--
--   sites carrying at least one such source : 6  (farmerinsurance.uk,
--     idea.uk, loanandmortgagecalculator.co.uk, mortgagecalculator.co.uk,
--     remortgagecalculator.uk, webdesign.co.uk)
--   matching content_sources rows           : 26 of 52 news_search rows fleet-wide
--   rows already carrying a 'region' key    : 0
--
-- Idempotent: the WHERE clause excludes any row that already carries a
-- 'region' key, so a re-run — or the migration runner's own doomed-
-- transaction probe — reports "already applied" rather than double-writing.
--
-- Data only. No schema change (config is jsonb; no ALTER TABLE, no new
-- column, no index — the existing news_search dispatch path already reads
-- source_type first via idx_cs_site_type before this key is ever consulted).

BEGIN;

-- ── PRECONDITION: the population must match the census this file was ─────
-- written against, or must already be fully applied. ─────────────────────
DO $pre$
DECLARE
  pending int;
BEGIN
  SELECT count(*) INTO pending
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND NOT (cs.config ? 'region');

  IF pending = 0 THEN
    RAISE EXCEPTION '691: already applied - every UK news_search source already carries a region key';
  END IF;

  IF pending <> 26 THEN
    RAISE EXCEPTION '691: expected exactly 26 unbackfilled UK news_search rows (the 2026-09-02 census), found %. The population may have moved — re-run the census in PLAN_2026-09-02_news_feed_ingestion.md before applying.', pending;
  END IF;
END
$pre$;

-- ── APPLY ──────────────────────────────────────────────────────────────
DO $apply$
DECLARE
  n int;
BEGIN
  UPDATE content_sources cs
     SET config = cs.config || jsonb_build_object('region', 'uk')
    FROM sites s
   WHERE s.id = cs.site_id
     AND cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND NOT (cs.config ? 'region');
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 26 THEN
    RAISE EXCEPTION '691: expected to update exactly 26 rows, updated %', n;
  END IF;
END
$apply$;

-- ── POST-CHECK. A verify block of bare SELECTs cannot stop a COMMIT ──────
-- (ON_ERROR_STOP ignores a non-empty result set) — every assertion here is
-- DO/RAISE (LANDMINES).
DO $post$
DECLARE
  remaining   int;
  wrong_value int;
BEGIN
  SELECT count(*) INTO remaining
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND NOT (cs.config ? 'region');
  IF remaining <> 0 THEN
    RAISE EXCEPTION '691 POST-CHECK FAILED: % UK news_search rows still have no region key', remaining;
  END IF;

  SELECT count(*) INTO wrong_value
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND cs.config->>'region' IS DISTINCT FROM 'uk';
  IF wrong_value <> 0 THEN
    RAISE EXCEPTION '691 POST-CHECK FAILED: % UK news_search rows carry a region key with the wrong value', wrong_value;
  END IF;

  -- Negative control: a non-.uk site's news_search config must be untouched.
  IF EXISTS (
    SELECT 1 FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND NOT (lower(s.domain) LIKE '%.uk')
     AND cs.config ? 'region'
  ) THEN
    RAISE EXCEPTION '691 POST-CHECK FAILED: a non-UK site''s news_search source carries a region key — this migration wrote outside its measured scope';
  END IF;

  RAISE NOTICE '691 POST-CHECK PASSED: 26 UK news_search rows carry region=uk, zero non-UK rows touched.';
END
$post$;

COMMIT;
