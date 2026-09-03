-- ROLLBACK for 750.
--
-- Restores the CURRENT plan's site_plan_sections rows for
-- boxingonline.com/tool-fight-calendar to the pre-750 state, and re-opens the
-- section_source_drift item 750 closed.
--
-- ⚠ AFTER THIS RUNS THE PAGE IS BACK IN LATENT-REVERT STATE: tier 1 will again
-- name [hero-tool, generic-text-block, advertising] while the page is
-- [hero-tool, event-list], so the next page BUILD deletes event-list and
-- undoes bugs_open/427's fix. check_section_source_drift will re-file within a
-- day (idx_swi_dedup frees the key on close, so a re-file is a genuine signal
-- rather than a duplicate). That is the CORRECT behaviour for a rollback of
-- this migration, not a defect.
--
-- The advertising row is restored with its ORIGINAL id and created_at, so a
-- restore is a restore and not a look-alike row with a new identity and
-- today's timestamp. Its assigned_fact_ids is restored to '[]' (NOT NULL --
-- the two are different instructions to the section writer).
--
-- NOTE: this deliberately writes NO pages.sections statement. A reader of
-- 719/727/728's rollbacks will expect one; there is none because 750 never
-- touched the cache. Reaching for one here would break the alignment in the
-- opposite direction and empty a live page's section list.

BEGIN;

DO $pre$
DECLARE
    v_site  uuid := 'd2aa5206-73bc-4707-a69c-2702c1eb9152';
    v_page  text := 'tool-fight-calendar';
    v_plan  uuid;
    v_names jsonb;
BEGIN
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;
    IF v_plan IS NULL THEN
        RAISE EXCEPTION '750-ROLLBACK ABORT: no is_current plan for boxingonline.com';
    END IF;
    IF v_plan <> 'bba66eda-2eae-459f-9e37-896efc9d079c' THEN
        RAISE EXCEPTION '750-ROLLBACK ABORT: the current plan is now %, not the plan 750 edited. A re-plan has happened; rolling back would write into a plan 750 never touched.', v_plan;
    END IF;

    SELECT jsonb_agg(component_name ORDER BY ordering) INTO v_names
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero-tool","event-list"]'::jsonb THEN
        RAISE EXCEPTION '750-ROLLBACK ABORT: tier 1 reads %, not the post-750 state [hero-tool,event-list]. Nothing to roll back, or someone has edited since.', v_names;
    END IF;
END
$pre$;

UPDATE site_plan_sections
   SET component_name = 'generic-text-block'
 WHERE id = 'd74518a8-03f9-4054-bd88-517aeae5f623'
   AND ordering = 1
   AND component_name = 'event-list';

INSERT INTO site_plan_sections
    (id, plan_id, page_name, ordering, component_name, assigned_fact_ids, subject, created_at)
VALUES
    ('16a18d39-e8a8-4f79-9abc-f041dfe40665',
     'bba66eda-2eae-459f-9e37-896efc9d079c',
     'tool-fight-calendar', 2, 'advertising', '[]'::jsonb, NULL,
     '2026-08-31 12:36:55.617418+00');

UPDATE site_work_items
   SET status = 'needs_human_review', result = NULL, updated_at = NOW()
 WHERE site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
   AND item_key = 'section_source_drift:tool-fight-calendar';

DO $post$
DECLARE
    v_names jsonb;
BEGIN
    SELECT jsonb_agg(component_name ORDER BY ordering) INTO v_names
      FROM site_plan_sections
     WHERE plan_id = 'bba66eda-2eae-459f-9e37-896efc9d079c' AND page_name = 'tool-fight-calendar';
    IF v_names IS DISTINCT FROM '["hero-tool","generic-text-block","advertising"]'::jsonb THEN
        RAISE EXCEPTION '750-ROLLBACK VERIFY FAILED: tier 1 reads %, expected the pre-750 triple.', v_names;
    END IF;
END
$post$;

COMMIT;
