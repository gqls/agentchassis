-- 455 ROLLBACK — put the spawn record back on every row 455 repaired.
--
-- Reversible from the row itself: 455 preserved the overwritten value under
-- `_replaced_spawn_record`, so this needs no backup table and no orchestration join.
-- Fenced on both provenance keys being present, so it can only touch 455's own rows
-- and a replay matches nothing. `updated_at` is preserved for the same reason 455
-- preserved it (bugs_closed/213: the stale reaper keys on it).

BEGIN;

UPDATE site_work_items
   SET result = (result -> '_replaced_spawn_record'),
       updated_at = updated_at
 WHERE result ? '_replaced_spawn_record'
   AND result ->> '_repaired_by' LIKE '455_repair_287_spawn_record_results%'
   AND jsonb_typeof(result -> '_replaced_spawn_record') = 'object';

DO $$
DECLARE leftover int;
BEGIN
    SELECT count(*) INTO leftover FROM site_work_items
     WHERE result ? '_replaced_spawn_record'
       AND result ->> '_repaired_by' LIKE '455_repair_287_spawn_record_results%';
    IF leftover > 0 THEN
        RAISE EXCEPTION '455 ROLLBACK: % rows still carry 455 provenance keys', leftover;
    END IF;
END $$;

COMMIT;
