-- 742_evidence_register_absence_check_ROLLBACK.sql
--
-- Removes the `evidence-register-absence` scheduled task.
--
-- It does NOT touch the `missing_evidence_register` work items the task has already filed:
-- those are findings, and a finding survives the retirement of the detector that found it.
-- If you want them gone too, cancel them deliberately and say why — but read them first,
-- because each one names a deployed site that is asserting figures with nothing checking them.
--
-- ⚠ Disabling is usually the right move rather than deleting: `UPDATE scheduled_tasks SET
-- enabled = false WHERE name = 'evidence-register-absence'`. Deleting loses the pre_query,
-- which is the whole implementation.

BEGIN;

DO $$
DECLARE n int; nitems int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF n <> 1 THEN
    RAISE EXCEPTION '742 ROLLBACK ABORT: expected exactly 1 task row, found %', n;
  END IF;
  SELECT count(*) INTO nitems FROM site_work_items WHERE item_type = 'missing_evidence_register';
  RAISE NOTICE '742 ROLLBACK: removing the task; LEAVING % missing_evidence_register item(s) in place (findings outlive their detector)', nitems;
END $$;

DELETE FROM scheduled_tasks WHERE name = 'evidence-register-absence';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF n <> 0 THEN
    RAISE EXCEPTION '742 ROLLBACK VERIFY: expected 0 task rows, found %', n;
  END IF;
  RAISE NOTICE '742 ROLLBACK OK: evidence-register-absence removed';
END $$;

COMMIT;
