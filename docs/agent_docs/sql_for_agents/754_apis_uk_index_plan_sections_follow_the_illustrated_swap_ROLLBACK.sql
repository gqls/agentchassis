-- ROLLBACK for 754. Restores the six plan rows to generic-text-block and
-- re-opens the drift item.
--
-- !! AFTER THIS RUNS THE PAGE IS BACK IN LATENT-REVERT STATE: the next page
-- BUILD flips pages.sections back to generic-text-block x6 (the state 754
-- removed). check_section_source_drift will re-file; that is correct
-- behaviour for this rollback, not a defect. No pages.sections statement here
-- on purpose: 754 never touched the cache (750_ROLLBACK's lesson).

BEGIN;

DO $pre$
DECLARE
    v_plan  uuid;
    v_names jsonb;
BEGIN
    SELECT id INTO v_plan FROM site_plans
     WHERE site_id = '1c6f3424-9d05-4a18-963b-72541bc19dca' AND is_current;
    IF v_plan IS DISTINCT FROM '7d520a81-5c69-4d50-a577-e8bb69149b96' THEN
        RAISE EXCEPTION '754-ROLLBACK ABORT: current plan is %, not the plan 754 edited. A re-plan happened; rolling back would write into a plan 754 never touched.', v_plan;
    END IF;
    SELECT jsonb_agg(component_name ORDER BY ordering) INTO v_names
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = 'index';
    IF v_names IS DISTINCT FROM '["hero","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","site-footer"]'::jsonb THEN
        RAISE EXCEPTION '754-ROLLBACK ABORT: tier 1 reads %, not the post-754 state. Nothing to roll back, or edited since.', v_names;
    END IF;
END
$pre$;

DO $upd$
DECLARE
    v_n int;
BEGIN
    UPDATE site_plan_sections
       SET component_name = 'generic-text-block'
     WHERE id IN ('57a87e9a-285f-4c2c-887d-6079e8f25cc6',
                  '067d1441-e7bf-4d5a-a29c-5e6d255cd7c7',
                  '8ad77de8-48a3-4849-94b2-44a179464766',
                  'd86d49b1-f534-4222-9fdd-b1b45514559b',
                  'a1835a4b-a207-4e61-af8e-5f2aff32a90e',
                  'c821c7e0-f4f1-4312-b3f8-0bdc25fe4431')
       AND component_name = 'illustrated-text-block';
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 6 THEN
        RAISE EXCEPTION '754-ROLLBACK ABORT: rename touched % rows, not 6.', v_n;
    END IF;

    UPDATE site_work_items
       SET status = 'needs_human_review', result = NULL, updated_at = NOW()
     WHERE id = '626c9c5d-27f6-43e0-adda-f89c77daea16' AND status = 'complete';
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '754-ROLLBACK ABORT: drift-item reopen touched % rows, not 1.', v_n;
    END IF;
END
$upd$;

COMMIT;
