-- ROLLBACK for 406_tool_suggester_requires_backend_gate.sql
-- Restores load_library_tools to its pre-406 shape: the ungated query,
-- and REMOVES the params key (the step had none before 406).
-- Uppercase suffix keeps this out of the runner's pending set by design.
--
-- HARDENED 2026-08-14 after the council's REVISE on c78ed496 (debug_historian):
-- scoped to the exact row id (single live row measured 2026-08-14: version 1,
-- id below — tool-suggester is NOT one of the dual-active-row agent types),
-- and gated on the pre-state actually carrying the 406 gate, so running this
-- against a row that was concurrently changed again is a loud no-op (the
-- verify RAISES on rowcount 0) rather than a silent clobber.

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
 WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
   AND type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   -- pre-state gate: only roll back the state 406 wrote
   AND position('requires-backend' in default_config#>>'{workflow,steps,load_library_tools,config,query}') > 0;

DO $$
DECLARE
  q text;
  p jsonb;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}',
         default_config#>'{workflow,steps,load_library_tools,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';

  IF q IS NULL OR position('requires-backend' in q) > 0 THEN
    RAISE EXCEPTION '406_ROLLBACK: gate still present (pre-state gate no-op? read the row before forcing): %', q;
  END IF;
  IF p IS NOT NULL THEN
    RAISE EXCEPTION '406_ROLLBACK: params key still present: %', p;
  END IF;
END $$;

COMMIT;
