-- ROLLBACK for 509 — take the copy gate's repair step back out of
-- page-content-writer's section loop (bugs_open/305).
--
-- ORDER IS LOAD-BEARING: re-point the writer FIRST, then delete the step. The
-- other order leaves generate_content pointing at a step that no longer exists,
-- which is the same broken chain 509's _HOLD suffix exists to prevent.
--
-- This rolls back the CONFIG only. The Go action, the scanner and the default-ON
-- counting annotations stay: they change no output on their own, so there is
-- nothing to undo there, and the counting is what tells you whether removing the
-- repair made the copy worse.

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '509_ROLLBACK: pre-rollback (bugs_open/305)');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}',
         '"render_section"'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}' = 'rewrite_negations';

UPDATE agent_definitions
   SET default_config = (default_config
         #- '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,copy_gate_annotate}')
         #- '{workflow,steps,compile_page,config,copy_gate_annotate}',
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations}',
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}' = 'render_section';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}' = 'render_section'
     AND (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations}') IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '509 ROLLBACK FAILED: expected the chain back at generate_content -> render_section with the step removed, got % rows', n;
  END IF;
  RAISE NOTICE '509 ROLLBACK OK';
END $$;

COMMIT;
