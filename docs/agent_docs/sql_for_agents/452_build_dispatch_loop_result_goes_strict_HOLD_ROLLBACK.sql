-- 452 ROLLBACK — restore the un-marked spellings on build-dispatch-loop
-- process_item.mark_complete and drop the added error_step.
-- Forward jsonb transform fenced on `result!` being present.

BEGIN;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           (default_config
              #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,result!}'
              #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,work_item_id!}'
              #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,error_step}'),
           '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,result}',
           COALESCE(default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,result!}',
                    '"handler_result"'::jsonb),
           true),
         '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,work_item_id}',
         COALESCE(default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,work_item_id!}',
                  '"current_item.id"'::jsonb),
         true),
       updated_at = now()
 WHERE type = 'build-dispatch-loop'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}' ? 'result!';

DO $$
DECLARE cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}'
      INTO cfg
      FROM agent_definitions
     WHERE type='build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF (cfg ? 'result!') OR NOT (cfg ? 'result') OR (cfg ? 'error_step') THEN
        RAISE EXCEPTION '452 ROLLBACK: not restored — config is %', cfg;
    END IF;
END $$;

COMMIT;
