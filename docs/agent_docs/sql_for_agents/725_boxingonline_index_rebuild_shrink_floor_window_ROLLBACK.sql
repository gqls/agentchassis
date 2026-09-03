-- 725_boxingonline_index_rebuild_shrink_floor_window_ROLLBACK.sql
-- Closes the window opened by 725_..._HOLD.sql: removes section_shrink_floor from
-- page-build-handler's save_sections step config, restoring the fleet default (0.5).
-- Run the moment the boxingonline index rebuild item is TERMINAL (success or failure).
-- Refuses if the key is not the 0.1 this window wrote (another override would be live).
BEGIN;
DO $$
DECLARE n int; cur text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '725 ROLLBACK REFUSED: expected exactly 1 active page-build-handler row, found %', n;
  END IF;
  SELECT default_config #>> '{workflow,steps,save_sections,config,section_shrink_floor}' INTO cur
    FROM agent_definitions WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cur IS NULL THEN
    RAISE EXCEPTION '725 ROLLBACK: nothing to do — section_shrink_floor is already absent (window already closed)';
  END IF;
  IF cur IS DISTINCT FROM '0.1' THEN
    RAISE EXCEPTION '725 ROLLBACK REFUSED: section_shrink_floor is % not 0.1 — not this window''s value; investigate before removing', cur;
  END IF;
  PERFORM snapshot_agent('page-build-handler',
                         '725_..._ROLLBACK.sql: pre-close (floor 0.1 about to be removed)');
END $$;
UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,save_sections,config,section_shrink_floor}',
       updated_at = now()
 WHERE type='page-build-handler' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
DO $$
DECLARE v text;
BEGIN
  SELECT default_config #>> '{workflow,steps,save_sections,config,section_shrink_floor}' INTO v
    FROM agent_definitions WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v IS NOT NULL THEN
    RAISE EXCEPTION '725 ROLLBACK VERIFY FAILED: section_shrink_floor still reads %', v;
  END IF;
  RAISE NOTICE '725 WINDOW CLOSED at %: section_shrink_floor absent, default 0.5 governs', now();
END $$;
COMMIT;
