-- 474 ROLLBACK — restore both writer rows' pre-474 config from the backup table.

BEGIN;

UPDATE agent_definitions ad
   SET default_config = b.default_config
  FROM _backup_474_writer_strip b
 WHERE ad.id = b.id;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (
         (type = 'page-content-writer' AND
          (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}')::text = 'true')
         OR
         (type = 'section-editor' AND
          (default_config #> '{workflow,steps,apply_edit,config,strip_literal_markdown}')::text = 'true')
       )
  ) THEN
    RAISE EXCEPTION '474 ROLLBACK FAILED: a strip flag survived';
  END IF;
  RAISE NOTICE '474 ROLLBACK OK';
END $$;

COMMIT;
