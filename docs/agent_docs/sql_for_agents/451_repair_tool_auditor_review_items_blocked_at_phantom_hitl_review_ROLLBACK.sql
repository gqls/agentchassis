-- ROLLBACK for 451_repair_tool_auditor_review_items_blocked_at_phantom_hitl_review.sql
-- Restores every repair_291-stamped row to the pre-repair shape (blocked at the
-- phantom handler, error text restored) and removes the stamp. Keys on the stamp,
-- so it cannot touch rows the repair did not write (the 442 mechanism).

BEGIN;

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items WHERE result ? 'repair_291';
  IF n = 0 THEN
    RAISE EXCEPTION '451_ROLLBACK: no repair_291-stamped rows — nothing to roll back';
  END IF;
END $$;

UPDATE site_work_items
   SET status        = 'blocked',
       handler_agent = 'hitl-review',
       error         = 'Handler agent not registered: hitl-review',
       result        = result - 'repair_291',
       updated_at    = now()
 WHERE result ? 'repair_291';

DO $$
DECLARE
  leftover int;
BEGIN
  SELECT count(*) INTO leftover FROM site_work_items WHERE result ? 'repair_291';
  IF leftover > 0 THEN
    RAISE EXCEPTION '451_ROLLBACK: % row(s) still stamped after rollback', leftover;
  END IF;
END $$;

COMMIT;
