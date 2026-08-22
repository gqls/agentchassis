-- 554_news_feed_trigger_orders_by_the_schedule_not_the_alphabet_ROLLBACK.sql
--
-- Restores `content-feed-trigger.find_news_sites` to the alphabetical ordering
-- it carried before migration 554 — VERBATIM, as captured from live
-- agent_definitions on 2026-08-22 and saved alongside the lane docs at
-- docs/agent_docs/docs024_key_docs_latest/bugfix_316_news_feed_ordering/
-- PREFIX_find_news_sites_query_2026-08-22.sql.
--
-- ⚠ ROLLING THIS BACK REINSTATES bugs_open/316. The alphabetical sort is the
-- defect: measured over five consecutive cap-hitting runs, the alphabetically
-- last of nine eligible sites was selected ZERO times while continuously due and
-- reached 419% of its own configured cadence. Do this only to unblock something
-- worse, and re-open the bug if you do.
--
-- The guards mirror 554's: exactly one live row, and the CURRENT state must be
-- the one 554 installed. If someone has since changed the ordering again, this
-- aborts rather than reverting their work to a defect.

SELECT snapshot_agent('content-feed-trigger',
       'migration 554 ROLLBACK: pre-revert (restoring the pre-316 alphabetical ordering)');

BEGIN;

DO $$
DECLARE
    live_rows int;
    q         text;
BEGIN
    SELECT count(*) INTO live_rows
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF live_rows <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 554 ROLLBACK: expected exactly 1 live content-feed-trigger row, found %.', live_rows;
    END IF;

    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
      INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF q IS NULL OR q NOT LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5' THEN
        RAISE EXCEPTION 'MIGRATION 554 ROLLBACK: the live query is not the one 554 installed — refusing to revert someone else''s change to a known defect. Live tail is: %', right(COALESCE(q,'(null)'), 60);
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,find_news_sites,config,query}',
        to_jsonb($q$SELECT DISTINCT s.id::text as site_id, s.domain FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW()))) ORDER BY s.domain LIMIT 5$q$::text),
        false),
    updated_at = now()
WHERE type = 'content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    q text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
      INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF q NOT LIKE '%ORDER BY s.domain LIMIT 5' THEN
        RAISE EXCEPTION 'MIGRATION 554 ROLLBACK: the alphabetical ordering was not restored. Live tail is: %', right(q, 60);
    END IF;
END $$;

COMMIT;
