-- 691_uk_news_search_region_default_VERIFY.sql
--
-- Proves migration 691's backfill is live and scoped correctly. Read-only —
-- ends in ROLLBACK, writes nothing.
--
-- Run:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--         -f - < docs/agent_docs/sql_for_agents/691_uk_news_search_region_default_VERIFY.sql

BEGIN;

DO $v$
DECLARE
  uk_total       int;
  uk_with_region int;
  non_uk_with_region int;
BEGIN
  SELECT count(*) INTO uk_total
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk';

  SELECT count(*) INTO uk_with_region
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND lower(s.domain) LIKE '%.uk'
     AND cs.config->>'region' = 'uk';

  SELECT count(*) INTO non_uk_with_region
    FROM content_sources cs
    JOIN sites s ON s.id = cs.site_id
   WHERE cs.source_type = 'news_search'
     AND NOT (lower(s.domain) LIKE '%.uk')
     AND cs.config ? 'region';

  IF uk_total = 0 THEN
    RAISE EXCEPTION '691 VERIFY: no UK news_search rows found at all — either the population moved to zero or this is the wrong database';
  END IF;

  IF uk_with_region <> uk_total THEN
    RAISE EXCEPTION '691 VERIFY FAILED: % of % UK news_search rows carry region=uk (want all of them)', uk_with_region, uk_total;
  END IF;

  IF non_uk_with_region <> 0 THEN
    RAISE EXCEPTION '691 VERIFY FAILED: % non-UK news_search rows carry a region key — scope leaked', non_uk_with_region;
  END IF;

  RAISE NOTICE '691 VERIFY PASSED: %/% UK news_search rows carry region=uk, 0 non-UK rows touched.', uk_with_region, uk_total;
END
$v$;

ROLLBACK;
