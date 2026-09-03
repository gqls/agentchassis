-- 745_loancash_restore_disclaimer_third_page_ROLLBACK.sql
--
-- Cancels the item 745 filed. Refuses once it has been claimed, because the page may already
-- be edited. ⚠ Cancelling leaves check-your-lender-is-authorised without the site-identity
-- disclaimer, which is the regression 745 exists to repair.

BEGIN;

DO $$
DECLARE nacted int;
BEGIN
  SELECT count(*) INTO nacted FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 745)'
     AND (claimed_at IS NOT NULL OR completed_at IS NOT NULL OR status <> 'triaged');
  IF nacted <> 0 THEN
    RAISE EXCEPTION '745 ROLLBACK ABORT: the item is already claimed or completed - the page may already be edited';
  END IF;
END $$;

UPDATE site_work_items
   SET status = 'cancelled', updated_at = now(),
       error  = '745 ROLLBACK: cancelled before dispatch - the disclaimer regression is left UNREPAIRED'
 WHERE created_by = 'loancash_couk_fca_validation lane (migration 745)' AND status = 'triaged';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 745)' AND status <> 'cancelled';
  IF n <> 0 THEN
    RAISE EXCEPTION '745 ROLLBACK VERIFY: expected 0 non-cancelled, found %', n;
  END IF;
  RAISE NOTICE '745 ROLLBACK OK: item cancelled - disclaimer still missing on that page';
END $$;

COMMIT;
