-- ROLLBACK for 760 — remove `gripper-spec-sheet` from the CURRENT plan's section
-- rows for robot-hands.com/gripper-catalog and close the gap it leaves.
--
-- ⚠ THIS ROLLBACK DOES NOT UNDO THE VISIBLE EFFECT, and that asymmetry is the
-- whole reason to read this header before running it. 760 only changes the PLAN.
-- The plan reaches a visitor only via a page REBUILD, and a rebuild composes
-- `page_components` and rewrites `pages.sections`. So:
--
--   * Run this BEFORE any rebuild  → a true rollback. Nothing ever rendered.
--   * Run this AFTER a rebuild     → the plan reverts, but the page keeps the
--                                    restored section until it is rebuilt AGAIN.
--                                    You have re-created `bugs_open/469`'s exact
--                                    divergence, pointing the other way, and the
--                                    next build will destroy the section a
--                                    second time. If that is what you want, say
--                                    so and drive the rebuild deliberately.
--
-- The pre-check below refuses rather than guessing which case you are in.

BEGIN;

DO $pre$
DECLARE
    v_site   uuid := '00ff3af5-dad8-4770-9f70-3edc267a3c92';
    v_page   text := 'gripper-catalog';
    v_plan   uuid;
    v_names  jsonb;
    v_orders int[];
    v_n      int;
BEGIN
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;
    IF v_plan IS NULL THEN
        RAISE EXCEPTION '760R ABORT: no is_current plan for robot-hands.com';
    END IF;

    SELECT jsonb_agg(component_name ORDER BY ordering), array_agg(ordering ORDER BY ordering)
      INTO v_names, v_orders
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero","generic-text-block","gripper-spec-sheet","info-card-grid","call-to-action"]'::jsonb
       OR v_orders IS DISTINCT FROM ARRAY[0,1,2,3,4] THEN
        RAISE EXCEPTION '760R ABORT: tier 1 is % at orderings %, not the post-760 five-section list. 760 did not run, has already been rolled back, or something else has edited the plan since.', v_names, v_orders;
    END IF;

    -- Refuse if the page has since rebuilt WITH the section: reverting the plan
    -- then leaves a live section no store names, which is 469's defect inverted.
    SELECT count(*) INTO v_n
      FROM page_components pc
      JOIN pages p ON p.id = pc.page_id
      LEFT JOIN content_components cc ON cc.id = pc.component_id
     WHERE p.site_id = v_site AND p.name = v_page
       AND COALESCE(pc.build_status,'') <> 'removed'
       AND (COALESCE(pc.slot_name,'') = 'gripper-spec-sheet'
            OR COALESCE(cc.function,'') = 'gripper-spec-sheet');
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760R ABORT: the page has REBUILT and carries % live gripper-spec-sheet row(s). Reverting the plan now re-creates bugs_open/469''s divergence and the next build destroys the section again. Decide deliberately: either drive a rebuild in the same window, or leave 760 in place.', v_n;
    END IF;
END
$pre$;

DELETE FROM site_plan_sections
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true)
   AND page_name = 'gripper-catalog'
   AND ordering = 2
   AND component_name = 'gripper-spec-sheet';

-- Close the ordering gap, via the same +1000 park so the non-deferrable UNIQUE
-- index (plan_id, page_name, ordering) cannot raise an order-dependent conflict.
UPDATE site_plan_sections
   SET ordering = ordering + 1000
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true)
   AND page_name = 'gripper-catalog'
   AND ordering >= 3;

UPDATE site_plan_sections
   SET ordering = ordering - 1001
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true)
   AND page_name = 'gripper-catalog'
   AND ordering >= 1000;

DO $post$
DECLARE
    v_site   uuid := '00ff3af5-dad8-4770-9f70-3edc267a3c92';
    v_page   text := 'gripper-catalog';
    v_plan   uuid;
    v_names  jsonb;
    v_orders int[];
BEGIN
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;
    SELECT jsonb_agg(component_name ORDER BY ordering), array_agg(ordering ORDER BY ordering)
      INTO v_names, v_orders
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero","generic-text-block","info-card-grid","call-to-action"]'::jsonb
       OR v_orders IS DISTINCT FROM ARRAY[0,1,2,3] THEN
        RAISE EXCEPTION '760R VERIFY FAILED: tier 1 reads % at orderings %, expected the original four-section list at [0,1,2,3].', v_names, v_orders;
    END IF;
    -- Same rows, moved back — not replacements.
    IF NOT EXISTS (SELECT 1 FROM site_plan_sections WHERE id = '8654754d-9e17-490f-80a3-4826ea628d1b'
                     AND component_name = 'info-card-grid' AND ordering = 2)
       OR NOT EXISTS (SELECT 1 FROM site_plan_sections WHERE id = '0f34d9f1-68a5-4fa2-8d88-4e8223f2650c'
                     AND component_name = 'call-to-action' AND ordering = 3) THEN
        RAISE EXCEPTION '760R VERIFY FAILED: the shifted rows are not the original rows moved back.';
    END IF;
END
$post$;

COMMIT;
