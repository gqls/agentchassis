-- 691_uk_news_search_region_default_ROLLBACK.sql
--
-- Withdraws migration 691's backfill: strips the 'region' key back out of
-- content_sources.config for UK news_search sources.
--
-- ⚠ SCOPED, NOT BLANKET. This removes 'region' only from rows matching the
-- exact predicate 691 wrote under (source_type='news_search', domain suffix
-- '.uk', current value = 'uk'). It does NOT touch:
--   - a .uk news_search row an operator has since edited to a different
--     region value on purpose (config->>'region' <> 'uk') — read as intent,
--     not drift, and left alone;
--   - the Go-code half (SeedContentSourcesAction's seed-time default) — that
--     is a code change and reverts only with an image rebuild/roll, not SQL.
-- Rolling back the data alone without also reverting the Go change means any
-- newly-seeded .uk source will still pick up region='uk' going forward; this
-- file only undoes the one-time backfill of the 26 pre-existing rows.
--
-- Idempotent: safe to run when the key is already absent.

BEGIN;

DO $rb$
DECLARE
  n int;
BEGIN
  UPDATE content_sources cs
     SET config = cs.config - 'region'
    FROM sites s
   WHERE s.id = cs.site_id
     AND cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND cs.config->>'region' = 'uk';
  GET DIAGNOSTICS n = ROW_COUNT;
  RAISE NOTICE '691 ROLLBACK: removed region key from % row(s)', n;
END
$rb$;

-- ── POST-CHECK ─────────────────────────────────────────────────────────
DO $rbpost$
DECLARE
  remaining int;
BEGIN
  SELECT count(*) INTO remaining
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND cs.config->>'region' = 'uk';
  IF remaining <> 0 THEN
    RAISE EXCEPTION '691 ROLLBACK FAILED: % rows still carry region=uk', remaining;
  END IF;
  RAISE NOTICE '691 ROLLBACK COMPLETE: no UK news_search row carries region=uk.';
END
$rbpost$;

COMMIT;
