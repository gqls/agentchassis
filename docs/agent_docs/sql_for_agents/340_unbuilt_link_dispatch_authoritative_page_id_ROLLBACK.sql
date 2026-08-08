-- 340 ROLLBACK — removes the two keys 340 added. Safe in any order relative to
-- the chassis roll: without the mapping the handler input is absent and
-- load_page_record's authoritative path is simply never taken.

\set ON_ERROR_STOP on

BEGIN;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,page_id?}',
       updated_at = NOW()
 WHERE type = 'build-dispatch-loop'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,load_page_record,config,authoritative_page_id}',
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $post$
DECLARE v_n int;
BEGIN
    SELECT count(*) INTO v_n FROM agent_definitions
     WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}' ? 'page_id?'
         OR default_config #> '{workflow,steps,load_page_record,config}' ? 'authoritative_page_id');
    IF v_n <> 0 THEN
        RAISE EXCEPTION '340 ROLLBACK FAILED: % row(s) still carry a 340 key', v_n;
    END IF;
    RAISE NOTICE '340 ROLLBACK OK';
END $post$;

COMMIT;
