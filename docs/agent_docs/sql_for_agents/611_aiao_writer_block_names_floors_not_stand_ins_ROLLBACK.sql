-- 611_aiao_writer_block_names_floors_not_stand_ins_ROLLBACK.sql
-- Restores the exact pre-611 evidence_base row (557-era writer_block, NNN+ exemplars
-- included) from migration_backups. Sidecar: never applied by the runner.

BEGIN;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect='evidence_base' AND is_current
  AND created_by='611_migration';

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT '2a8ebf9c-20a2-4c39-b191-840b012371da', 'evidence_base',
       mb.old_value->'data',
       ss.source, ss.source_agent,
       'restored by 611_ROLLBACK from migration_backups (pre-611 byte-exact copy)',
       true, '611_rollback'
FROM migration_backups mb
JOIN site_specs ss ON ss.id::text = mb.target_id
WHERE mb.migration_name='611_aiao_writer_block_names_floors_not_stand_ins';

DO $$
DECLARE n int; wb text;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '611_ROLLBACK: expected 1 current row, found %', n; END IF;
  SELECT data->>'writer_block' INTO wb FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF position('NNN+ AI agents' in wb) = 0 THEN
    RAISE EXCEPTION '611_ROLLBACK: restored row does not carry the 557-era block';
  END IF;
END $$;

COMMIT;
