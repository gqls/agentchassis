-- 732_designblog_tools_hub_pages_row_and_refile_ROLLBACK.sql
-- Removes what 732 created — REFUSES once the page has been built (page_components exist),
-- because then the honest teardown is a deliberate retraction, not a rollback.

BEGIN;

DO $$
DECLARE n int; pid uuid;
BEGIN
  SELECT id INTO pid FROM pages
   WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index';
  IF pid IS NOT NULL THEN
    SELECT count(*) INTO n FROM page_components WHERE page_id = pid;
    IF n > 0 THEN
      RAISE EXCEPTION '732R REFUSED: tools-index has % component(s) — it has been built; retract deliberately, do not roll back', n;
    END IF;
  END IF;
END $$;

DELETE FROM pages
 WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index';

DELETE FROM site_plan_sections
 WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND page_name='tools-index';

UPDATE site_work_items
   SET status='cancelled'
 WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a'
   AND item_key='needs_page:tools-index'
   AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index';
  IF n <> 0 THEN RAISE EXCEPTION '732R VERIFY: pages row still present'; END IF;
  RAISE NOTICE '732R OK: pages row + plan sections removed, live item (if any) cancelled. 726 plan row untouched — run 726_ROLLBACK for that.';
END $$;

COMMIT;
