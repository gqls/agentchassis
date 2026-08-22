-- 556_news_feed_capacity_both_caps_to_10_ROLLBACK.sql
--
-- Returns BOTH caps to 5: the query LIMIT and process_sites.max_iterations.
--
-- ⚠ REVERT BOTH OR NEITHER. Reverting only the query LIMIT leaves the loop at 10
-- and the query at 5, which is harmless but pointless; reverting only the loop
-- leaves the query returning 10 rows the loop will not process, which is the
-- silent no-op state migration 556 exists to prevent — and the cap census would
-- keep reporting "under the cap" while throughput had halved.
--
-- ⚠ THIS REINSTATES A KNOWN SHORTFALL. At cap 5 the queue supplies 20
-- site-refreshes/day against 42 demanded, and seven sites that this change put
-- on their configured 6-hour cadence go back to being served roughly every other
-- run. Do this only to relieve cost or load pressure, and say so.
--
-- It does NOT touch the due-ordering from migration 554 — the fairness fix stays
-- either way, and the guard below refuses if it has gone.

SELECT snapshot_agent('content-feed-trigger',
       'migration 556 ROLLBACK: pre-revert (both caps 10 -> 5)');

BEGIN;

DO $$
DECLARE q text; iters jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query',
           default_config->'workflow'->'steps'->'process_sites'->'config'->'max_iterations'
      INTO q, iters
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF q IS DISTINCT FROM $post$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$post$ OR iters IS DISTINCT FROM to_jsonb(10) THEN
        RAISE EXCEPTION 'MIGRATION 556 ROLLBACK: the live state is not the one 556 installed (query match %, max_iterations %) — refusing to revert someone else''s change.',
            (q IS NOT DISTINCT FROM $post$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$post$), COALESCE(iters::text,'ABSENT');
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,find_news_sites,config,query}',
            to_jsonb($pre$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5$pre$::text),
            false),
        '{workflow,steps,process_sites,config,max_iterations}',
        to_jsonb(5),
        false),
    updated_at = now()
WHERE type = 'content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text; iters jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query',
           default_config->'workflow'->'steps'->'process_sites'->'config'->'max_iterations'
      INTO q, iters
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF q NOT LIKE '%LIMIT 5' OR iters IS DISTINCT FROM to_jsonb(5) THEN
        RAISE EXCEPTION 'MIGRATION 556 ROLLBACK: both caps did not return to 5 (tail %, max_iterations %).', right(q,40), COALESCE(iters::text,'ABSENT');
    END IF;
    IF q NOT LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5' THEN
        RAISE EXCEPTION 'MIGRATION 556 ROLLBACK: the due-ordering from 554 was lost.';
    END IF;
END $$;

COMMIT;
