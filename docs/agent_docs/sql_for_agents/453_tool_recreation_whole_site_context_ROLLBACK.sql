-- 453 ROLLBACK — restore tool-recreation-handler's load_related_context to the
-- pre-453 text (`… LEFT JOIN research_results … ORDER BY p.nav_order LIMIT 10`).
--
-- DB config is LIVE ON APPLY, so this file exists BEFORE the forward apply.
-- Same pre-state-gate shape as the forward migration: refuses unless the row
-- currently carries 453's LATERAL text, so re-running it cannot silently
-- revert a LATER change some other session made.
--
-- Hand-run only (uppercase sidecar suffix — the runner excludes it).

BEGIN;

SELECT snapshot_agent('tool-recreation-handler',
  '453_tool_recreation_whole_site_context_ROLLBACK: pre-revert');

DO $$
DECLARE q text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '453-ROLLBACK: expected exactly 1 live tool-recreation-handler row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_related_context,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('LEFT JOIN LATERAL' in q) = 0 THEN
    RAISE EXCEPTION '453-ROLLBACK: current query is not 453''s — refusing to revert someone else''s change: %', q;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_related_context,config,query}',
         to_jsonb('SELECT p.name, p.title, p.page_type, rr.summary FROM pages p LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = ''adoption_page'' WHERE p.site_id = $1 AND p.name != $2 ORDER BY p.nav_order LIMIT 10'::text)
       ),
       updated_at = now()
 WHERE id = '8701375f-81f7-4d92-ba39-c85f8489dada'
   AND type = 'tool-recreation-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_related_context,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF position('LIMIT 10' in q) = 0 OR position('LEFT JOIN research_results rr ON' in q) = 0 THEN
    RAISE EXCEPTION '453-ROLLBACK: restore failed — query is: %', q;
  END IF;
END $$;

COMMIT;
