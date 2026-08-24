-- ROLLBACK for 587 (bugs_open/352) — restore the withdrawn rows to the status
-- they held before the migration cancelled them.
--
-- This is possible only because 587 recorded `result.pre_352_status` on every
-- row it touched. It restores exactly those rows and nothing else: a row
-- cancelled by any other means carries no `cancelled_by = 'migration_587'`
-- marker and is left alone.
--
-- ⚠ THE COLLISION, AND WHY THIS SKIPS RATHER THAN FAILS. Cancelling a row FREED
-- its idx_swi_dedup slot (migration 157: cancelled is terminal for dedup). If a
-- later render audit has since filed a NEW, non-terminal row for the same
-- (site_id, item_key) — which is the whole point of freeing the slot —
-- restoring this row to a non-terminal status would put two live rows in that
-- slot and violate the index. So the restore is guarded by NOT EXISTS, and the
-- rows it declines to restore are REPORTED rather than silently passed over: a
-- rollback that quietly restores 40 of 73 and says "done" is worse than one that
-- fails, because the gap is invisible.
--
-- The declined rows need no action. Their successor is a BETTER row — same
-- defect, verified selector — so the correct outcome is to leave the successor
-- in place and keep the withdrawal.

BEGIN;

DO $$
DECLARE
    to_restore integer;
    blocked    integer;
BEGIN
    SELECT count(*) INTO to_restore
      FROM site_work_items w
     WHERE w.status = 'cancelled'
       AND w.result->>'cancelled_by' = 'migration_587'
       AND NOT EXISTS (
             SELECT 1 FROM site_work_items o
              WHERE o.site_id  = w.site_id
                AND o.item_key = w.item_key
                AND o.id      <> w.id
                AND o.status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled'));

    SELECT count(*) INTO blocked
      FROM site_work_items w
     WHERE w.status = 'cancelled'
       AND w.result->>'cancelled_by' = 'migration_587'
       AND EXISTS (
             SELECT 1 FROM site_work_items o
              WHERE o.site_id  = w.site_id
                AND o.item_key = w.item_key
                AND o.id      <> w.id
                AND o.status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled'));

    IF to_restore = 0 AND blocked = 0 THEN
        RAISE EXCEPTION '587 ROLLBACK: no rows carry cancelled_by = migration_587. Nothing to undo — check you are on the right database before forcing anything.';
    END IF;

    RAISE NOTICE '587 ROLLBACK: restoring % row(s); % SKIPPED because a live successor already holds the dedup slot (this is the expected outcome once a render audit has re-filed them, and those rows should stay withdrawn).', to_restore, blocked;
END $$;

UPDATE site_work_items w
   SET status     = COALESCE(w.result->>'pre_352_status', 'detected'),
       updated_at = now(),
       result     = (w.result - 'cancelled_by' - 'pre_352_status' - 'cancelled_at' - 'reason')
                    || jsonb_build_object('rolled_back_from_587_at', now())
 WHERE w.status = 'cancelled'
   AND w.result->>'cancelled_by' = 'migration_587'
   AND NOT EXISTS (
         SELECT 1 FROM site_work_items o
          WHERE o.site_id  = w.site_id
            AND o.item_key = w.item_key
            AND o.id      <> w.id
            AND o.status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled'));

COMMIT;
