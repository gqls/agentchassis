-- 769 ROLLBACK — remove the two copywriter directory components. Refuses if a page already uses
-- them, because deleting a component a page_components row points at orphans that page silently.
BEGIN;
DO $r$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
   WHERE cc.function IN ('copywriter-directory','copywriter-directory-listing');
  IF n > 0 THEN RAISE EXCEPTION '769 ROLLBACK REFUSED: % page_components row(s) use these components', n; END IF;
  DELETE FROM content_components WHERE function IN ('copywriter-directory','copywriter-directory-listing');
  GET DIAGNOSTICS n = ROW_COUNT;
  RAISE NOTICE '769 ROLLBACK: removed % component(s)', n;
END $r$;
COMMIT;
