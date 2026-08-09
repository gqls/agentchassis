-- 342_page_build_handler_save_sections_follows_page_record.sql
--
-- Completes bugs_open/220: page-build-handler's save_sections step was the ONE
-- step still resolving its page from `input_data.spec.page_name` while every
-- sibling (load_spec_sections, load_existing_content, call_content_writer's
-- current_page, deploy_page) follows `page_record` — the page load_page_record
-- loaded, which since mig 340 honours the work item's authoritative page_id.
--
-- PROVEN LIVE 2026-08-09 on the 220 acceptance run (dartsonline, corr 110acf5a):
-- item 338deb27 — the writer wrote the TARGET's sections (grip-styles, from the
-- target's plan via load_spec_sections: page_record.name) and save_sections then
-- saved them ONTO THE CONTAINER (beginners, from spec.page_name), replacing that
-- page's stored content_data with another page's copy at 10:00:56Z. Sibling item
-- a8327624 was stopped only by the content-regression floor (2,520 chars of
-- brands-index copy vs brand-comparison's 11,914). A split-brain saga: deploy
-- honestly skipped the target ("no component rows yet") because the rows had
-- been saved to the wrong page.
--
-- THE FIX: save_sections resolves its page from page_record.name — the same
-- source as every other step, so the saga has ONE page identity end-to-end.
-- Blast radius: for every consistent dispatch (the 116-lane census: every live
-- item type except unbuilt_internal_link) page_record.name == spec.page_name,
-- so nothing changes; where load_page_record fell back from a non-page name
-- ("new page needed") to the id, page_record.name is the REAL page name, which
-- is strictly more correct than the marker string. Config-only; live on apply;
-- effective with the already-rolled chassis (v carrying migs 340/341's Go).
--
-- ROLLBACK: 342_..._ROLLBACK.sql (restore the spec path).

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,save_sections,config}' INTO v_cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_cfg IS NULL THEN
        RAISE EXCEPTION '342: no active page-build-handler save_sections step found';
    END IF;
    IF v_cfg #>> '{page_name_field}' IS DISTINCT FROM 'input_data.spec.page_name' THEN
        RAISE EXCEPTION '342: save_sections.page_name_field is not the expected input_data.spec.page_name (already applied, or drifted): %', v_cfg;
    END IF;
END $pre$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,save_sections,config,page_name_field}',
           '"page_record.name"'::jsonb
       ),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $post$
DECLARE v_val text;
BEGIN
    SELECT default_config #>> '{workflow,steps,save_sections,config,page_name_field}' INTO v_val
      FROM agent_definitions
     WHERE type = 'page-build-handler'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_val IS DISTINCT FROM 'page_record.name' THEN
        RAISE EXCEPTION '342 FAILED: page_name_field is % after update', v_val;
    END IF;
    RAISE NOTICE '342 OK: save_sections follows page_record.name — one page identity end-to-end';
END $post$;

COMMIT;
