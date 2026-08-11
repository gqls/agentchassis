-- ROLLBACK for 380_build_dispatch_loop_maps_purpose_top_level.sql
-- Removes the optional top-level purpose mapping from build-dispatch-loop's
-- call_handler step, restoring the pre-380 dispatch shape (undeployed_asset
-- deploys fall back to the spec Default "hero" — the bugs_open/231 state).

BEGIN;

UPDATE agent_definitions
SET default_config = default_config
        #- '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,purpose?}',
    updated_at = now()
WHERE type = 'build-dispatch-loop'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

DO $$
DECLARE remaining int;
BEGIN
    SELECT count(*) INTO remaining
    FROM agent_definitions
    WHERE type = 'build-dispatch-loop' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,purpose?}' IS NOT NULL;
    IF remaining <> 0 THEN
        RAISE EXCEPTION '380 ROLLBACK verify FAILED: % rows still carry the purpose? mapping', remaining;
    END IF;
    RAISE NOTICE '380 ROLLBACK verify OK: purpose? mapping removed';
END $$;

COMMIT;
