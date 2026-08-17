-- ROLLBACK for 445 — restore tool-suggester's load_library_tools to 406's text
-- (gated query, description uncapped, LIMIT 30).
--
-- Scoped by id with a pre-state gate, the same shape as 445 itself: refuses
-- unless the row is the pinned one and currently carries 445's uncapped query,
-- so re-running it cannot silently undo a LATER change by another session.

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '445_tool_suggester_whole_library_ROLLBACK: pre-revert');

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}' INTO q
    FROM agent_definitions
   WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF q IS NULL THEN
    RAISE EXCEPTION '445-ROLLBACK: no live tool-suggester row at the pinned id';
  END IF;
  IF position('left(description, 200)' in q) = 0 THEN
    RAISE EXCEPTION '445-ROLLBACK: current query is not 445''s — refusing to revert someone else''s change: %', q;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_library_tools,config,query}',
         to_jsonb('SELECT id::text, function, display_name, category, description FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND html_template != '''' AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY display_name LIMIT 30'::text)
       ),
       updated_at = now()
 WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
   AND type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}' INTO q
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';
  IF position('LIMIT 30' in q) = 0 OR position('requires-backend' in q) = 0 THEN
    RAISE EXCEPTION '445-ROLLBACK: revert did not restore 406''s text: %', q;
  END IF;
END $$;

COMMIT;
