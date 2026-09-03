-- 746_advertise_news_feed_enablement_VERIFY.sql
--
-- Proves migration 746 is live for advertise.co.uk and reports what the pipeline has
-- done with it since. Read-only — ends in ROLLBACK, writes nothing.
--
-- Run:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--         -f - < docs/agent_docs/sql_for_agents/746_advertise_news_feed_enablement_VERIFY.sql

BEGIN;

DO $v$
DECLARE
  rec        bool;
  n_src      int;
  n_rss      int;
  n_ns_uk    int;
  n_fetched  int;
  n_err      int;
  picked     int;
  n_items    int;
  n_relevant int;
  n_review   int;
  n_rejected int;
  n_ingested int;
BEGIN
  -- structure (RAISE on failure)
  SELECT (data->'content_features'->'news_feed'->>'recommended')::boolean INTO rec
    FROM site_specs
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9' AND aspect = 'classification' AND is_current = true;
  IF rec IS DISTINCT FROM true THEN
    RAISE EXCEPTION '746 VERIFY FAILED: current classification does not read recommended=true';
  END IF;

  SELECT count(*),
         count(*) FILTER (WHERE source_type = 'rss' AND config->>'feed_url' = 'https://www.webpronews.com/feed/'),
         count(*) FILTER (WHERE source_type = 'news_search' AND config->>'region' = 'uk'),
         count(*) FILTER (WHERE last_fetched_at IS NOT NULL),
         COALESCE(sum(error_count), 0)
    INTO n_src, n_rss, n_ns_uk, n_fetched, n_err
    FROM content_sources
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9' AND is_active = true;
  IF n_src <> 6 OR n_rss <> 1 OR n_ns_uk <> 5 THEN
    RAISE EXCEPTION '746 VERIFY FAILED: active sources total=% rss=% news_search(region=uk)=% (want 6/1/5)', n_src, n_rss, n_ns_uk;
  END IF;

  -- progress (NOTICE only — these legitimately change over time)
  SELECT count(*) INTO picked
    FROM sites s
    JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true
     AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
   WHERE s.id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed')
     AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true)
          OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true
                       AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW() + interval '3 hours')));

  SELECT count(*),
         count(*) FILTER (WHERE status = 'relevant'),
         count(*) FILTER (WHERE status = 'review'),
         count(*) FILTER (WHERE status = 'rejected'),
         count(*) FILTER (WHERE status = 'ingested')
    INTO n_items, n_relevant, n_review, n_rejected, n_ingested
    FROM content_feed_items
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9';

  RAISE NOTICE '746 VERIFY PASSED: recommended=true; 6 active sources (1 rss + 5 news_search region=uk); % of 6 fetched at least once, error_count sum %; trigger would select the site now: % (1 = due, 0 = not due yet, both fine after a fetch); feed items: % total = % relevant / % review / % rejected / % ingested(unscored).',
    n_fetched, n_err, picked, n_items, n_relevant, n_review, n_rejected, n_ingested;
END
$v$;

ROLLBACK;
