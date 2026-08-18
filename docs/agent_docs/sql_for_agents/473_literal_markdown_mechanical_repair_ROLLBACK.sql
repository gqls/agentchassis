-- 473 ROLLBACK — restore page-rerender's pre-473 config from the backup table.
-- Restores BOTH halves (condition + strip flag) in one write.

BEGIN;

UPDATE agent_definitions ad
   SET default_config = b.default_config
  FROM _backup_473_page_rerender b
 WHERE ad.id = b.id;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
           LIKE '%literal_markdown%'
  ) THEN
    RAISE EXCEPTION '473 ROLLBACK FAILED: condition still carries literal_markdown';
  END IF;
  RAISE NOTICE '473 ROLLBACK OK';
END $$;

COMMIT;
