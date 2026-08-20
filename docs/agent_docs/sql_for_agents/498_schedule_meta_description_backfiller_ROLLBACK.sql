-- 498 ROLLBACK — unschedule the meta-description backfiller
--
-- Deletes the scheduled_tasks row. The AGENT is untouched and stays hand-runnable
-- via scripts/backfill-meta-descriptions.sh, so this stops the automation without
-- removing the capability.
--
-- Prefer `UPDATE scheduled_tasks SET enabled = false WHERE name='meta-description-backfill'`
-- if you only want to pause it — that keeps the row, its history and its timestamps.
-- This file is for removing it properly.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'meta-description-backfill';
  IF n <> 1 THEN
    RAISE EXCEPTION '498 ROLLBACK: expected exactly 1 meta-description-backfill task, found %', n;
  END IF;
END $$;

DELETE FROM scheduled_tasks WHERE name = 'meta-description-backfill';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'meta-description-backfill';
  IF n <> 0 THEN
    RAISE EXCEPTION '498 ROLLBACK VERIFY: % rows remain', n;
  END IF;
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '498 ROLLBACK VERIFY: the agent should still exist and be live, found % rows', n;
  END IF;
  RAISE NOTICE '498 ROLLBACK OK — unscheduled; the agent remains live and hand-runnable';
END $$;

COMMIT;
