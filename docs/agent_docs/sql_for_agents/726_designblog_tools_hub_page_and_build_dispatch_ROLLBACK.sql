-- 726_designblog_tools_hub_page_and_build_dispatch_ROLLBACK.sql
-- Removes the tools-index plan row and cancels the build item IF it has not gone terminal.
-- Deliberately does NOT touch a pages row or any built components: if the page has already
-- been built, removing the plan row alone would strand it (the "page not in the plan gets
-- archived at next replan" hazard) — in that case decide the page's fate explicitly instead
-- of running this file.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index';
  IF n <> 0 THEN
    RAISE EXCEPTION '726R REFUSED: a tools-index pages row exists (the page has been built) — rolling back the plan row now would strand a live page; handle the page first';
  END IF;
END $$;

DELETE FROM site_plan_pages
 WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND name='tools-index';

UPDATE site_work_items
   SET status='cancelled'
 WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a'
   AND item_key='needs_page:tools-index'
   AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_pages WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND name='tools-index';
  IF n <> 0 THEN RAISE EXCEPTION '726R VERIFY: plan row still present'; END IF;
  RAISE NOTICE '726R OK: plan row removed, live build item (if any) cancelled.';
END $$;

COMMIT;
