-- 739_loancash_repair_three_misstated_rules_ROLLBACK.sql
--
-- Cancels the four content_rewrite items migration 739 filed. It CANCELS rather than
-- DELETES: a work item that has been claimed or handled has history a delete would erase,
-- and 'cancelled' is a terminal status the dedup index already excludes, so a later re-file
-- of the same item_key is not blocked.
--
-- ⚠ REFUSES if any of the four has already been claimed or completed. Once page-build-handler
-- has acted, the served copy may already have changed, and cancelling the item does not put
-- the old wording back — that is a content decision, not a rollback.

BEGIN;

DO $$
DECLARE nacted int;
BEGIN
  SELECT count(*) INTO nacted FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 739)'
     AND (claimed_at IS NOT NULL OR completed_at IS NOT NULL OR status <> 'triaged');
  IF nacted <> 0 THEN
    RAISE EXCEPTION '739 ROLLBACK ABORT: % of the four items have already been claimed, completed or moved - the copy may already have changed; inspect before cancelling', nacted;
  END IF;
END $$;

UPDATE site_work_items
   SET status = 'cancelled', updated_at = now(),
       error = '739 ROLLBACK: cancelled before dispatch'
 WHERE created_by = 'loancash_couk_fca_validation lane (migration 739)'
   AND status = 'triaged';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 739)'
     AND status <> 'cancelled';
  IF n <> 0 THEN
    RAISE EXCEPTION '739 ROLLBACK VERIFY: expected 0 non-cancelled items, found %', n;
  END IF;
  RAISE NOTICE '739 ROLLBACK OK: 4 repair items cancelled';
END $$;

COMMIT;
