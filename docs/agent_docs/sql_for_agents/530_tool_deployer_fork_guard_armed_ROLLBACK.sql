-- 530 ROLLBACK — disarm the fork-path guard (absent = record-only = pre-521).
BEGIN;
UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,deploy_tool,config,enforce_instance_scope}',
    updated_at = now()
WHERE type='tool-deployer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
COMMIT;
