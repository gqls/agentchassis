-- ROLLBACK for 779_pay_success_and_cancel_pages.sql
--
-- Removes the two work items. It deliberately does NOT delete any page the
-- framework may already have built from them: by the time you are rolling back,
-- /pay/success/ may be live and serving a real customer who has just paid, and
-- deleting it silently restores the 404 this migration existed to remove.
--
-- If the pages were built and you genuinely want them gone, retire them the way
-- the estate retires any page — archive the `pages` row (status='archived') so
-- the record survives — and check FIRST whether anything still sends buyers
-- there: `grep -n 'pay/success' internal/auth-service/billing/stripe.go`.
-- While stripe.go still mints those URLs, removing the pages is a regression,
-- not a rollback.

\set ON_ERROR_STOP on
BEGIN;

DELETE FROM site_work_items
 WHERE item_key IN ('needs_content_page:pay-success:2026-09-04',
                    'needs_content_page:pay-cancel:2026-09-04');

DO $$
DECLARE n int; n_pages int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE item_key IN ('needs_content_page:pay-success:2026-09-04',
                      'needs_content_page:pay-cancel:2026-09-04');
  IF n <> 0 THEN
    RAISE EXCEPTION 'rollback incomplete: % pay-page work item(s) remain', n;
  END IF;

  SELECT count(*) INTO n_pages FROM pages p JOIN sites s ON s.id = p.site_id
   WHERE s.domain = 'webdesign.uk' AND p.url LIKE '/pay/%' AND p.status = 'active';
  IF n_pages > 0 THEN
    RAISE NOTICE 'NOTE: % /pay/ page(s) are ALREADY BUILT and are LEFT SERVING on purpose. stripe.go still sends buyers there; read this file''s header before removing them.', n_pages;
  END IF;
END $$;

COMMIT;
