-- 754: correct apis.uk/index's CURRENT plan rows to match the live page, so
--      the next page BUILD stops carrying a latent revert of the illustrated swap
--
-- Context (2026-09-03, diagnosed by the bugs_open/427 lane, verified live by the
-- apis.uk lane before writing this). The owner's shipped index page serves
-- illustrated-text-block x6; the CURRENT plan's site_plan_sections still name
-- generic-text-block x6. The page build's loader takes tier 1
-- (site_plan_sections, load_page_sections_from_spec_action.go:142-148) and
-- SYNCS THE WINNER DOWN over pages.sections (:558-570). The page is
-- build_status='needs_rebuild' with built_from_plan_version NULL, so
-- reconcile decideEmit returns "stale": a rebuild WILL happen, and without
-- this migration it flips pages.sections back to generic-text-block x6.
-- NOT the re-plan path: reconcilePlanWithRealised (v3_site_actions.go:7701-7724)
-- snaps realised sections onto a proposal, so a genuine re-plan is safe.
--
-- SHAPE: an in-place RENAME at pinned ids/orderings (migration 750's worked
-- example; precedent 154). NOT delete+insert: subject and assigned_fact_ids
-- are read positionally by ordering. Here all six rows carry subject NULL and
-- assigned_fact_ids NULL, and the rename touches neither -- the subject
-- BACKFILL for these six rows is deliberately NOT here: it is HELD pending
-- the phrasing spec from the prompts lane (apis NOTES 2026-09-03).
--
-- WHAT THIS DOES NOT TOUCH, and why it is safe:
--   * pages.sections        -- already correct (illustrated x6); after this,
--                              the loader's sync becomes a no-op (IS DISTINCT
--                              FROM false).
--   * page_components       -- 7 rows, lock_type='permanent'. Pairing is the
--                              SHARED relation (slot_pairing.go): arm 1 pairs
--                              slot_name='generic-text-block' with the OLD
--                              spelling, arm 4 pairs component_function=
--                              'illustrated-text-block' with the NEW one --
--                              the locks pair one-to-one under EITHER, so this
--                              rename cannot orphan them.
--   * site_plan_imagery     -- IMG-075 section rows key on scope_ref
--                              'index:1'..'index:6' (ordinals), not names.
--   * site_specs.site_plan  -- this site has NO such aspect (0 rows); creating
--                              one would hand tier 2 a permanent say (750's
--                              153-lesson).
--   * the served artefact   -- expected byte-for-byte unchanged. This removes
--                              a latent revert; any change to served bytes or
--                              page_components is a FINDING.
--
-- ALSO CLOSED: the open section_source_drift item for index
-- (626c9c5d-27f6-43e0-adda-f89c77daea16, needs_human_review) -- this
-- migration IS the human review's outcome, acting on the 427 lane's diagnosis.
--
-- RESIDUAL (stated, not buried): per-plan corrections cannot reach a FUTURE
-- plan row; the framework fix (typed action, architecture review) is the 427
-- lane's. And this file does NOT change build_status/built_from_plan_version:
-- whether apis.uk/index should remain 'needs_rebuild' is a separate decision
-- (the page is wedged between its locks and the prune_floor save guard --
-- apis NOTES 2026-09-02) and folding it in here would widen a drift fix into
-- a lifecycle decision.
--
-- Verify by hand after applying (the loader's own query):
--   SELECT sps.component_name FROM site_plan_sections sps
--     JOIN site_plans sp ON sp.id=sps.plan_id
--    WHERE sp.site_id='1c6f3424-9d05-4a18-963b-72541bc19dca'
--      AND sp.is_current AND sps.page_name='index' ORDER BY sps.ordering;
--   -- expect: hero, illustrated-text-block x6, site-footer

BEGIN;

DO $pre$
DECLARE
    v_site  uuid := '1c6f3424-9d05-4a18-963b-72541bc19dca';
    v_plan  uuid := '7d520a81-5c69-4d50-a577-e8bb69149b96';
    v_n     int;
    v_names jsonb;
    v_pages jsonb;
BEGIN
    SELECT count(*) INTO v_n FROM site_plans WHERE site_id = v_site;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '754 ABORT: expected exactly 1 site_plans row for apis.uk, found %. The no-superseded-plan premise no longer holds; re-derive.', v_n;
    END IF;
    PERFORM 1 FROM site_plans WHERE id = v_plan AND site_id = v_site AND is_current;
    IF NOT FOUND THEN
        RAISE EXCEPTION '754 ABORT: plan % is not apis.uk''s current plan any more. A re-plan has happened; this file targets a plan that no longer decides builds.', v_plan;
    END IF;

    SELECT jsonb_agg(component_name ORDER BY ordering) INTO v_names
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = 'index';
    IF v_names IS DISTINCT FROM '["hero","generic-text-block","generic-text-block","generic-text-block","generic-text-block","generic-text-block","generic-text-block","site-footer"]'::jsonb THEN
        RAISE EXCEPTION '754 ABORT: tier 1 reads %, not the drifted state this file corrects. Already applied, or edited since the census.', v_names;
    END IF;

    SELECT sections INTO v_pages FROM pages WHERE site_id = v_site AND name = 'index';
    IF v_pages IS DISTINCT FROM '["hero","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","site-footer"]'::jsonb THEN
        RAISE EXCEPTION '754 ABORT: pages.sections reads %, not the illustrated list this file aligns tier 1 to. The premise has moved; re-verify both stores.', v_pages;
    END IF;

    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = 'index' AND pc.lock_type = 'permanent';
    IF v_n <> 7 THEN
        RAISE EXCEPTION '754 ABORT: expected 7 permanently locked page_components on index, found %.', v_n;
    END IF;

    PERFORM 1 FROM site_work_items
     WHERE id = '626c9c5d-27f6-43e0-adda-f89c77daea16' AND status = 'needs_human_review';
    IF NOT FOUND THEN
        RAISE EXCEPTION '754 ABORT: the section_source_drift item is not in needs_human_review; someone is acting on it. Check who before applying.';
    END IF;
END
$pre$;

DO $upd$
DECLARE
    v_n int;
BEGIN
    UPDATE site_plan_sections
       SET component_name = 'illustrated-text-block'
     WHERE id IN ('57a87e9a-285f-4c2c-887d-6079e8f25cc6',
                  '067d1441-e7bf-4d5a-a29c-5e6d255cd7c7',
                  '8ad77de8-48a3-4849-94b2-44a179464766',
                  'd86d49b1-f534-4222-9fdd-b1b45514559b',
                  'a1835a4b-a207-4e61-af8e-5f2aff32a90e',
                  'c821c7e0-f4f1-4312-b3f8-0bdc25fe4431')
       AND component_name = 'generic-text-block';
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 6 THEN
        RAISE EXCEPTION '754 ABORT: rename touched % rows, not 6 -- the pinned ids have moved; nothing is committed.', v_n;
    END IF;

    UPDATE site_work_items
       SET status = 'complete',
           result = jsonb_build_object(
               'resolution', 'plan rows corrected in place to match the live page (migration 754); latent build revert removed',
               'corrected_by', 'apis.uk lane, on the 427 lane''s diagnosis',
               'migration', '754'),
           updated_at = NOW()
     WHERE id = '626c9c5d-27f6-43e0-adda-f89c77daea16' AND status = 'needs_human_review';
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '754 ABORT: drift-item close touched % rows, not 1.', v_n;
    END IF;
END
$upd$;

DO $post$
DECLARE
    v_site  uuid := '1c6f3424-9d05-4a18-963b-72541bc19dca';
    v_plan  uuid := '7d520a81-5c69-4d50-a577-e8bb69149b96';
    v_names jsonb;
    v_pages jsonb;
    v_n     int;
BEGIN
    SELECT jsonb_agg(component_name ORDER BY ordering) INTO v_names
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = 'index';
    IF v_names IS DISTINCT FROM '["hero","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","illustrated-text-block","site-footer"]'::jsonb THEN
        RAISE EXCEPTION '754 VERIFY FAILED: tier 1 reads %, expected the illustrated list.', v_names;
    END IF;
    SELECT count(*) INTO v_n FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = 'index'
       AND (subject IS NOT NULL OR assigned_fact_ids IS NOT NULL);
    IF v_n <> 0 THEN
        RAISE EXCEPTION '754 VERIFY FAILED: % rows gained a subject or fact list -- the rename must not touch them (backfill is HELD).', v_n;
    END IF;
    SELECT sections INTO v_pages FROM pages WHERE site_id = v_site AND name = 'index';
    IF v_pages IS DISTINCT FROM v_names THEN
        RAISE EXCEPTION '754 VERIFY FAILED: pages.sections (%) no longer matches tier 1 -- this file must not touch the cache.', v_pages;
    END IF;
    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = 'index' AND pc.lock_type = 'permanent';
    IF v_n <> 7 THEN
        RAISE EXCEPTION '754 VERIFY FAILED: locked component count moved to %.', v_n;
    END IF;
    RAISE NOTICE '754 applied: tier 1 now matches the live page; drift item closed; locks and cache untouched.';
END
$post$;

COMMIT;
