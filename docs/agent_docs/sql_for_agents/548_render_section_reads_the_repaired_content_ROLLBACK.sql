-- ROLLBACK for 548 — point render_section back at the writer's raw output.
--
-- ⚠ This returns the gate to REPORTING repairs it does not deliver: the action
-- still rewrites, still marks `status: repaired` and `hits_after`, and the
-- renderer goes back to reading the unpatched map. That is the exact state
-- measured on 2026-08-21 and is worse than having no gate, because the marker
-- reads like success. If you want the repair OFF, roll back 517 (which makes it
-- report `repair_unavailable`, loudly) or 509 (which removes the step).

BEGIN;

SELECT snapshot_agent('page-content-writer', '548_ROLLBACK: pre-rollback');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,content_from}',
         '"generated_content.result"'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,content_from}' = 'copy_gate.result';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,content_from}' = 'generated_content.result';
  IF n <> 1 THEN
    RAISE EXCEPTION '548 ROLLBACK FAILED: content_from is not back at generated_content.result (n=%)', n;
  END IF;
  RAISE NOTICE '548 ROLLBACK OK — ⚠ the gate now reports repairs it does not deliver';
END $$;

COMMIT;
