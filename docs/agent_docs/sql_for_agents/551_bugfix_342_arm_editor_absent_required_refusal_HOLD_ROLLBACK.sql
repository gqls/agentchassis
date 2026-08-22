-- 551_bugfix_342_arm_editor_absent_required_refusal_HOLD_ROLLBACK.sql
--
-- Removes refuse_absent_required_fields from section-editor's apply_edit step.
-- Unset means pre-551 behaviour byte for byte (fail-open reader); the snapshot
-- taken by 551 remains available for a full restore.

BEGIN;

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,apply_edit,config,refuse_absent_required_fields}',
    version    = version + 1,
    updated_at = now()
WHERE type='section-editor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    armed text;
BEGIN
    SELECT default_config#>>'{workflow,steps,apply_edit,config,refuse_absent_required_fields}' INTO armed
    FROM agent_definitions
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF armed IS NOT NULL THEN
        RAISE EXCEPTION 'ROLLBACK 551: key still present (%).', armed;
    END IF;
END $$;

COMMIT;
