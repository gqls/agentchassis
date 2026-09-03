-- 743_loancash_restore_content_dropped_by_the_739_rewrite_ROLLBACK.sql
--
-- Cancels the two restoration items 743 filed. Refuses once either has been claimed, because
-- by then the page may already have been edited and cancelling the item does not undo that.
--
-- ⚠ Cancelling these leaves the site WITHOUT its identity disclaimer on two pages, which is
-- the regression 743 exists to repair. Do not run this to tidy the queue.

BEGIN;

DO $$
DECLARE nacted int;
BEGIN
  SELECT count(*) INTO nacted FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)'
     AND (claimed_at IS NOT NULL OR completed_at IS NOT NULL OR status <> 'triaged');
  IF nacted <> 0 THEN
    RAISE EXCEPTION '743 ROLLBACK ABORT: % item(s) already claimed or completed - the pages may already be edited', nacted;
  END IF;
END $$;

UPDATE site_work_items
   SET status = 'cancelled', updated_at = now(),
       error = '743 ROLLBACK: cancelled before dispatch - the disclaimer regression is left UNREPAIRED'
 WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)'
   AND status = 'triaged';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)' AND status <> 'cancelled';
  IF n <> 0 THEN
    RAISE EXCEPTION '743 ROLLBACK VERIFY: expected 0 non-cancelled, found %', n;
  END IF;
  RAISE NOTICE '743 ROLLBACK OK: 2 restoration items cancelled - disclaimer still missing on two pages';
END $$;

COMMIT;
