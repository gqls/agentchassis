-- 725_boxingonline_index_rebuild_shrink_floor_window_HOLD.sql
--
-- A TEMPORARY, OWNER-AUTHORISED WINDOW on the page-build-handler's save_sections step:
-- section_shrink_floor 0.5 (the default) -> 0.1, for ONE rebuild of boxingonline.com
-- /index.html, restored by the companion ROLLBACK the moment that rebuild is terminal.
--
-- WHY. bugs_open/425 §2: the rerender path does not execute the fixed list-item producer, so
-- the home page's cards get their decks only from a BUILD-path rebuild (the route that fixed
-- guides-index on 2026-09-02). The owner authorised that rebuild (2026-09-02 ~21:31Z, "Fire it
-- now"); it FAILED three times (21:51, ~22:3x, 23:25Z) on save_page_sections' content-loss
-- guard: "SECTION SHRINK REFUSED for page "index" — call-to-action 1116->167 chars of VISIBLE
-- text (15% kept, floor 50%)". The boxingonline session then read the defended block: a
-- ~1.1k-char call-to-action that narrates the news feed and walks the reader through four tools
-- in full sentences — item 1 of the owner's 2026-08-31 critique verbatim. The refusal is
-- protecting the copy the owner rejected; the shrink IS the repair. A percentage floor cannot
-- tell padding from substance (save_sections_shrink_guard.go's own calibration note: the guard
-- would refuse exactly one legitimate tightening rewrite a week, and this is that one).
--
-- WHY THIS SHAPE AND NOT ANOTHER. (a) The floor is read ONLY from the saving step's config
-- (save_sections_shrink_guard.go:246, single_slot_floors.go:112) — there is no per-item or
-- per-section override, and the guard's own text names the step key as the escape hatch.
-- (b) page-build-handler's save_sections and page-rerender's save_sections are SEPARATE step
-- configs (measured at the live rows 2026-09-03 09:3xZ) — this touches builds only; the ~192
-- queued page_rerenders are unaffected. (c) 0.1 rather than 0: a 167/1116 = 15% keep passes;
-- a total wipe (0%) is still refused. (d) Blast radius: open needs_page fleet-wide at apply time
-- = 0 (measured); completion rate ~4/hour; the window is minutes. Any OTHER build that saves in
-- the window is subject to the 10% floor — the ROLLBACK is run by the monitor at the target
-- item's terminal state, not by hand later, and this file's verify block records the time.
--
-- OWNER RULING 2026-09-03 ~09:37Z: "Scoped override, one run" chosen over leaving it to the
-- copy lane. Applied BY HAND (HOLD; never by an unscoped runner --apply). Companion:
-- 725_..._ROLLBACK.sql — MUST run when the rebuild item is terminal, success or failure.
BEGIN;
DO $$
DECLARE n int; cur text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '725 REFUSED: expected exactly 1 active page-build-handler row, found %', n;
  END IF;
  SELECT default_config #>> '{workflow,steps,save_sections,config,section_shrink_floor}' INTO cur
    FROM agent_definitions WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cur IS NOT NULL THEN
    RAISE EXCEPTION '725 REFUSED: section_shrink_floor already set (%) on page-build-handler save_sections — window already open or a prior override is live; do not stack', cur;
  END IF;
  IF (SELECT default_config #> '{workflow,steps,save_sections,config}' FROM agent_definitions
        WHERE type='page-build-handler' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) IS NULL THEN
    RAISE EXCEPTION '725 REFUSED: workflow.steps.save_sections.config not found on page-build-handler — step renamed?';
  END IF;
  PERFORM snapshot_agent('page-build-handler',
                         '725_boxingonline_index_rebuild_shrink_floor_window_HOLD.sql: pre-window (floor default 0.5)');
END $$;
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
                                  '{workflow,steps,save_sections,config,section_shrink_floor}',
                                  '0.1'::jsonb, true),
       updated_at = now()
 WHERE type='page-build-handler' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
DO $$
DECLARE v text;
BEGIN
  SELECT default_config #>> '{workflow,steps,save_sections,config,section_shrink_floor}' INTO v
    FROM agent_definitions WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v IS DISTINCT FROM '0.1' THEN
    RAISE EXCEPTION '725 VERIFY FAILED: section_shrink_floor reads % after update', v;
  END IF;
  RAISE NOTICE '725 WINDOW OPEN at %: page-build-handler save_sections.section_shrink_floor = 0.1', now();
END $$;
COMMIT;
