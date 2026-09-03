-- 742_page_rerender_routing_reason_write_door_HOLD_ROLLBACK.sql
--
-- Drops the write door. Leaves 741's read door alone — the two are independently
-- reversible, which is the whole benefit of the split (guardian [medium], round
-- 56047b18).
--
-- ⚠ WHAT ROLLING BACK COSTS: hand-written migrations can again INSERT a page_rerender
-- item with an out-of-vocabulary `spec.routing_reason`. If 741 is still applied that key
-- is REFUSED at the gate rather than silently assembled, so the loud half survives; if
-- neither is applied, bugs_open/440 is fully reopened.
--
-- WHEN YOU WOULD RUN IT: the constraint is rejecting INSERTs it should not. Run this
-- FIRST, because it distinguishes the two causes and they have opposite fixes:
--
--   SELECT DISTINCT spec->>'routing_reason' FROM site_work_items
--    WHERE item_type='page_rerender' AND spec->>'routing_reason' IS NOT NULL;
--
-- A value that SHOULD be in the vocabulary -> do NOT roll back; add it to
-- RerenderSectionReasons and cut one migration regenerating both this constraint and the
-- gate clause. A legitimate NON-routing value -> it belongs in spec.reason (free prose,
-- never validated, owner ruling D4) and the producer is what needs fixing, not this.

BEGIN;

SET LOCAL lock_timeout = '5s';

ALTER TABLE site_work_items
  DROP CONSTRAINT IF EXISTS chk_page_rerender_routing_reason_vocabulary;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint
              WHERE conname='chk_page_rerender_routing_reason_vocabulary'
                AND conrelid='site_work_items'::regclass) THEN
    RAISE EXCEPTION '742 ROLLBACK VERIFY FAILED: the constraint is still present';
  END IF;
  RAISE NOTICE '742 ROLLED BACK: write door removed. 741''s read door (if applied) still refuses an unknown routing key at the gate.';
END $$;

COMMIT;
