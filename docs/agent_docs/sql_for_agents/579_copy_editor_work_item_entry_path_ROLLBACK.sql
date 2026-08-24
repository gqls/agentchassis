-- ROLLBACK for 579_copy_editor_work_item_entry_path.sql
-- Restores copy-editor's default_config from the snapshot 579 took.
BEGIN;
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions_backup
   WHERE type='copy-editor' AND snapshot_reason='579_copy_editor_work_item_entry_path';
  IF n < 1 THEN RAISE EXCEPTION 'no 579 snapshot to restore from'; END IF;
END $$;

UPDATE agent_definitions a SET default_config = b.default_config, updated_at = now()
  FROM (SELECT DISTINCT ON (id) id, default_config FROM agent_definitions_backup
         WHERE type='copy-editor' AND snapshot_reason='579_copy_editor_work_item_entry_path'
         ORDER BY id, taken_at DESC) b
 WHERE a.id = b.id;

DO $$
DECLARE s text;
BEGIN
  SELECT default_config->'workflow'->>'start_step' INTO s FROM agent_definitions
   WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF s <> 'ensure_site_record' THEN RAISE EXCEPTION 'rollback did not restore start_step, got %', s; END IF;
END $$;
COMMIT;
