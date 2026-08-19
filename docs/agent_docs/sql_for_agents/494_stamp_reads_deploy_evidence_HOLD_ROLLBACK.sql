-- 494_stamp_reads_deploy_evidence_HOLD_ROLLBACK.sql
--
-- Removes the deploy_result_field key from the three stamping steps. The key is
-- opt-in with the unsafe default OFF, so removing it restores today's behaviour
-- byte for byte — the guard simply does not run.
--
-- Surgical (a `#-` of one key each), deliberately NOT a whole-config restore:
-- other lanes edit these agents and a blob restore would revert their work too.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-rerender',  '494_ROLLBACK: pre-restore');
SELECT snapshot_agent('report-builder', '494_ROLLBACK: pre-restore');
SELECT snapshot_agent('section-editor', '494_ROLLBACK: pre-restore');

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,update_status,config,deploy_result_field}',
       updated_at = NOW()
 WHERE type IN ('page-rerender','report-builder')
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,update_page_status,config,deploy_result_field}',
       updated_at = NOW()
 WHERE type = 'section-editor'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type IN ('page-rerender','report-builder','section-editor')
       AND default_config::text LIKE '%deploy_result_field%';
    IF bad <> 0 THEN
        RAISE EXCEPTION '494 ROLLBACK verify: % row(s) still carry deploy_result_field', bad;
    END IF;
END $$;

COMMIT;
