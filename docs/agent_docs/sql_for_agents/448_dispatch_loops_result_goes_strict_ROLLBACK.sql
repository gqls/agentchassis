-- 448 ROLLBACK — restore the un-marked spellings on diagnose/report-dispatch-loop
-- mark_complete and drop the added error_step.
-- Forward jsonb transform fenced on `result!` being present (does not depend on
-- the backup table; a hand restore from agent_definitions_backup must use the
-- snapshot that HOLDS the old key — see the 417 re-run lesson).

BEGIN;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           (default_config
              #- '{workflow,steps,mark_complete,config,result!}'
              #- '{workflow,steps,mark_complete,config,work_item_id!}'
              #- '{workflow,steps,mark_complete,config,error_step}'),
           '{workflow,steps,mark_complete,config,result}',
           COALESCE(default_config #> '{workflow,steps,mark_complete,config,result!}',
                    '"handler_result"'::jsonb),
           true),
         '{workflow,steps,mark_complete,config,work_item_id}',
         COALESCE(default_config #> '{workflow,steps,mark_complete,config,work_item_id!}',
                  '"claimed.work_item_id"'::jsonb),
         true),
       updated_at = now()
 WHERE type IN ('diagnose-dispatch-loop','report-dispatch-loop')
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,mark_complete,config}' ? 'result!';

DO $$
DECLARE r record; cfg jsonb;
BEGIN
    FOR r IN SELECT type, default_config FROM agent_definitions
              WHERE type IN ('diagnose-dispatch-loop','report-dispatch-loop')
                AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    LOOP
        cfg := r.default_config #> '{workflow,steps,mark_complete,config}';
        IF (cfg ? 'result!') OR NOT (cfg ? 'result') OR (cfg ? 'error_step') THEN
            RAISE EXCEPTION '448 ROLLBACK: % not restored — config is %', r.type, cfg;
        END IF;
    END LOOP;
END $$;

COMMIT;
