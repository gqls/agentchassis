-- ROLLBACK for 406_tool_suggester_requires_backend_gate.sql
-- Restores load_library_tools to its pre-406 shape: the ungated query,
-- and REMOVES the params key (the step had none before 406).
-- Uppercase suffix keeps this out of the runner's pending set by design.

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '406_tool_suggester_requires_backend_gate_ROLLBACK: pre-rollback');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config #- '{workflow,steps,load_library_tools,config,params}',
         '{workflow,steps,load_library_tools,config,query}',
         to_jsonb('SELECT id::text, function, display_name, category, description FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND html_template != '''' ORDER BY display_name LIMIT 30'::text)
       ),
       updated_at = now()
 WHERE type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  q text;
  p jsonb;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}',
         default_config#>'{workflow,steps,load_library_tools,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) > 0 THEN
    RAISE EXCEPTION '406_ROLLBACK: gate still present in query: %', q;
  END IF;
  IF p IS NOT NULL THEN
    RAISE EXCEPTION '406_ROLLBACK: params key still present: %', p;
  END IF;
END $$;

COMMIT;
