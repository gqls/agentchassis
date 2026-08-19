-- ROLLBACK for 481. Restores the tool-generator prompt_template from the snapshot
-- migration 481 took. Safe to run only if 481 ran; it RAISEs otherwise rather than
-- silently doing nothing.
BEGIN;
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.tables
   WHERE table_name='mig_481_tool_generator_prompt_backup';
  IF n = 0 THEN RAISE EXCEPTION '481 ROLLBACK: no snapshot table - 481 never ran'; END IF;
  SELECT count(*) INTO n FROM mig_481_tool_generator_prompt_backup;
  IF n = 0 THEN RAISE EXCEPTION '481 ROLLBACK: snapshot table is EMPTY - nothing to restore'; END IF;
END $$;

UPDATE agent_definitions a
SET default_config = jsonb_set(a.default_config,
      '{workflow,steps,generate_tool_html,config,prompt_template}',
      to_jsonb(b.prompt_template), false),
    updated_at = now()
FROM (SELECT DISTINCT ON (agent_id) agent_id, prompt_template
      FROM mig_481_tool_generator_prompt_backup ORDER BY agent_id, taken_at DESC) b
WHERE a.id = b.agent_id;

DO $$
DECLARE tmpl text;
BEGIN
  SELECT default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}' INTO tmpl
  FROM agent_definitions WHERE type='tool-generator' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tmpl LIKE '%15. Never report success%' THEN
    RAISE EXCEPTION '481 ROLLBACK: rule 15 still present - restore did not take';
  END IF;
END $$;
COMMIT;
