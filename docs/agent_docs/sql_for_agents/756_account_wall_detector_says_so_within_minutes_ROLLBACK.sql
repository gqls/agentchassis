-- 756_account_wall_detector_says_so_within_minutes_ROLLBACK.sql — removes the detector task.
-- Notes it wrote stay: they are history. After this, an account wall is once again invisible
-- until someone reads a failed run.
BEGIN;
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='account-wall-detector') THEN
    RAISE EXCEPTION '756 ROLLBACK REFUSED: account-wall-detector absent — not applied (or already rolled back).';
  END IF;
END $$;
DELETE FROM scheduled_tasks WHERE name='account-wall-detector';
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='account-wall-detector') THEN RAISE EXCEPTION '756 ROLLBACK VERIFY: task survived'; END IF;
  RAISE NOTICE '756 ROLLBACK OK: account-wall-detector removed. REMINDER: the account wall is now silent again.';
END $$;
COMMIT;
