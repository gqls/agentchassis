-- 474 ROLLBACK — surgical inverse: delete exactly the two flag keys 474 added.
-- Intervening migrations on either row survive (NOT a restore-from-snapshot).

BEGIN;

SELECT snapshot_agent('page-content-writer', '474_ROLLBACK: pre-revert');
SELECT snapshot_agent('section-editor', '474_ROLLBACK: pre-revert');

UPDATE agent_definitions
   SET default_config = default_config
         #- '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}',
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config
         #- '{workflow,steps,apply_edit,config,strip_literal_markdown}',
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (
         (type = 'page-content-writer' AND
          default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}' IS NOT NULL)
         OR
         (type = 'section-editor' AND
          default_config #> '{workflow,steps,apply_edit,config,strip_literal_markdown}' IS NOT NULL)
       )
  ) THEN
    RAISE EXCEPTION '474 ROLLBACK FAILED: a strip flag survived';
  END IF;
  RAISE NOTICE '474 ROLLBACK OK';
END $$;

COMMIT;
