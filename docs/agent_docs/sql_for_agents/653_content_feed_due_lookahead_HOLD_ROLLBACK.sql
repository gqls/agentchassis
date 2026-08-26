-- 653_content_feed_due_lookahead_HOLD_ROLLBACK.sql
--
-- Removes the half-cadence due look-ahead from find_news_sites, restoring the
-- 556 post-image (bare cs.next_fetch_at <= NOW()).
--
-- ⚠ THIS REINSTATES bugs_open/410's PHASE LOCK at the site level: every news
-- site whose sources are all 6-hourly returns to a 12 h effective cadence
-- under a 6 h label, silently, with every run still COMPLETED. Do this only to
-- relieve cost or load pressure, and say so where the 410 file will be found.
-- Note the Go half (feedSourceDuePredicate) stays in the rolled binary either
-- way; with only this rollback applied, sites are admitted on the bare-NOW()
-- rule and their sources fetched with the look-ahead — the phase lock returns
-- because admission is the outer gate.

SELECT snapshot_agent('content-feed-trigger',
       'migration 653 ROLLBACK: remove due look-ahead');

BEGIN;

DO $$
DECLARE q text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF q IS DISTINCT FROM $post$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW() + COALESCE((SELECT make_interval(secs => interval_seconds / 2.0) FROM scheduled_tasks WHERE name = 'content-feed-refresh'), interval '3 hours'))))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$post$ THEN
        RAISE EXCEPTION 'MIGRATION 653 ROLLBACK: the live query is not the one 653 installed — refusing to revert someone else''s change.';
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,find_news_sites,config,query}',
        to_jsonb($pre$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$pre$::text),
        false),
    updated_at = now()
WHERE type = 'content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF q LIKE '%make_interval%' THEN
        RAISE EXCEPTION 'MIGRATION 653 ROLLBACK: the look-ahead is still present.';
    END IF;
    IF q NOT LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10' THEN
        RAISE EXCEPTION 'MIGRATION 653 ROLLBACK: the 554 ordering / 556 capacity tail was lost.';
    END IF;
END $$;

COMMIT;
