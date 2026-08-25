-- 619_..._HOLD_ROLLBACK.sql - put the four LLM audit seats back on the loop's path.
-- The inverse edge, not a snapshot restore. NOTE what this reinstates: every LLM
-- finding dispatches a page rewrite again (the owner's 2026-08-25 complaint).
DO $probe$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'improvement-loop'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,call_completeness_discovery,next_step}' = 'record_audit_pass'
    ) THEN
        RAISE EXCEPTION '619 ROLLBACK: not applied';
    END IF;
END $probe$;
BEGIN;
SELECT snapshot_agent('improvement-loop', '619_..._ROLLBACK: pre-restore');
UPDATE agent_definitions
   SET default_config = jsonb_set(jsonb_set(default_config,
         '{workflow,steps,call_completeness_discovery,next_step}', to_jsonb('spawn_design_audit'::text), true),
         '{workflow,steps,call_completeness_discovery,description}', to_jsonb('Run completeness checks (empty sections)'::text), true),
       updated_at = now()
 WHERE type = 'improvement-loop'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
DO $v$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type='improvement-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                   AND default_config #>> '{workflow,steps,call_completeness_discovery,next_step}' = 'spawn_design_audit') THEN
        RAISE EXCEPTION '619 ROLLBACK verify: edge not restored';
    END IF;
END $v$;
COMMIT;
