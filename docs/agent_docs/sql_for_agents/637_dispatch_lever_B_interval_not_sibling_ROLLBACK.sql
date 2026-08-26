-- 637 ROLLBACK — restore the 584 state (sibling enabled, original at interval 60).
-- Guarded; a replay is a 0-row no-op. After running this, re-edit the 584 VERIFY's
-- lever assertion (it encodes ruling B) — they move in lockstep.
BEGIN;
UPDATE scheduled_tasks SET interval_seconds = 60, updated_at = NOW()
 WHERE name = 'build-pipeline-trigger' AND interval_seconds = 30;
UPDATE scheduled_tasks SET enabled = true, updated_at = NOW()
 WHERE name = 'build-pipeline-trigger-2' AND NOT enabled;
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%' AND enabled AND interval_seconds = 60;
  IF n <> 2 THEN RAISE EXCEPTION '637 ROLLBACK post-check: expected 2 enabled rows at interval 60, found %', n; END IF;
END $$;
COMMIT;
