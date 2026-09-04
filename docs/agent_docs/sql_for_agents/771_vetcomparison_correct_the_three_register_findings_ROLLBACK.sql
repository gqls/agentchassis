-- 771_vetcomparison_correct_the_three_register_findings_ROLLBACK.sql
--
-- Cancels the nine correction items migration 771 filed, BEFORE they are claimed.
--
-- ⚠ IT CANNOT UNDO A COMPLETED EDIT. Once page-build-handler has run an item the page is already
-- changed, and this file will refuse rather than pretend. The pre-image of any completed edit is the
-- `op='delete'` row in `page_component_history` for that page (join on `page_id`, never
-- `component_id`) — that archive is what made the loancash restoration possible on 2026-09-03.
--
-- ⚠ AND CANCELLING LEAVES THREE KNOWN FACTUAL ERRORS ON A LIVE SITE: the CMA final report dated
-- November 2024 when it is 24 March 2026 (2 pages), the £21/£12.50 prescription caps stated as
-- settled when the draft brackets them as inflation-adjustable placeholders (7 pages), and "36
-- service categories" for what draft Schedule 1 defines as 36 services in 5 categories (8 pages).
-- Those are recorded in the site's evidence register (migration 759) either way, but recorded is not
-- repaired. If you are cancelling because a repair went wrong, say which, and prefer fixing forward.

BEGIN;

DO $$
DECLARE nacted int; ntotal int;
BEGIN
  SELECT count(*) FILTER (WHERE status NOT IN ('triaged','approved')), count(*)
    INTO nacted, ntotal
    FROM site_work_items
   WHERE created_by = 'bugfix_414 register-programme lane (migration 771)';
  IF ntotal IS DISTINCT FROM 9 THEN
    RAISE EXCEPTION '771 ROLLBACK ABORT: expected 9 items, found % - look before cancelling', coalesce(ntotal::text,'NULL');
  END IF;
  IF nacted IS DISTINCT FROM 0 THEN
    RAISE EXCEPTION '771 ROLLBACK ABORT: % item(s) are already claimed, complete or failed - a completed edit CANNOT be cancelled, only reverted from page_component_history. Inspect each before proceeding', nacted;
  END IF;
END $$;

UPDATE site_work_items
   SET status = 'cancelled',
       error = '771 ROLLBACK: cancelled before dispatch - the three recorded factual errors are left UNREPAIRED on the live site',
       updated_at = now()
 WHERE created_by = 'bugfix_414 register-programme lane (migration 771)'
   AND status IN ('triaged','approved');

DO $$
DECLARE nopen int;
BEGIN
  SELECT count(*) INTO nopen FROM site_work_items
   WHERE created_by = 'bugfix_414 register-programme lane (migration 771)'
     AND status IN ('triaged','approved');
  IF nopen IS DISTINCT FROM 0 THEN
    RAISE EXCEPTION '771 ROLLBACK VERIFY: % item(s) still open', coalesce(nopen::text,'NULL');
  END IF;
  RAISE NOTICE '771 ROLLBACK OK: 9 items cancelled - the three errors remain live and recorded-but-unrepaired';
END $$;

COMMIT;
