-- ROLLBACK for 446 — restore 445's UNMARKED left(description, 200).
-- Pre-state gated: refuses unless the row currently carries 446's marker, so
-- re-running cannot silently revert a later change by another session.

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '446_tool_suggester_mark_truncated_descriptions_ROLLBACK: pre-revert');

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}' INTO q
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF q IS NULL OR position('[…truncated]' in q) = 0 THEN
    RAISE EXCEPTION '446-ROLLBACK: current query is not 446''s — refusing to revert someone else''s change: %', q;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_library_tools,config,query}',
         to_jsonb('SELECT id::text, function, display_name, category, left(description, 200) AS description FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND html_template != '''' AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY display_name'::text)
       ),
       updated_at = now()
 WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
   AND type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;
