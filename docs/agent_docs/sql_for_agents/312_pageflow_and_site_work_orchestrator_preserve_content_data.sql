-- 312_pageflow_and_site_work_orchestrator_preserve_content_data.sql
--
-- bugs_open/194 — `save_page_sections` takes the path to its structured section
-- data from a per-caller config key, `sections_metadata_field`. A caller that does
-- not set it falls through to the regex HTML-parse path, whose sections carry no
-- content_data, and the INSERT writes SQL NULL
-- (save_page_sections_action.go:658-687). content_data is the ONLY thing the
-- rerender path can regenerate a section from.
--
-- Seed 310 fixed the first of the three remaining instances (page-rebuild). This
-- seed fixes the other two that HAVE the data: pageflow-builder and
-- site-work-orchestrator. Census re-read live 2026-08-04 19:38Z — note the step is
-- nested inside a loop sub_workflow in four of the six callers, so a top-level
-- jsonb_each MISSES them and you need jsonb_path_query(default_config,'$.**.steps'):
--
--     page-build-handler        page_content.response.sections_metadata   preserved
--     page-rerender             rerender_sections.sections_metadata       preserved
--     page-rebuild              page_content.response.sections_metadata   preserved (310)
--     pageflow-builder          ABSENT                                    NULLed  <- this seed
--     site-work-orchestrator    ABSENT                                    NULLed  <- this seed
--     tool-recreation-handler   ABSENT                                    correct, see below
--
-- WHY THIS KEY RESOLVES FOR BOTH. Each of the two runs `write_page_content` — a
-- call_agent at `page-content-writer` with `output_field: page_content` — in the
-- SAME loop sub_workflow as its save step (measured, not assumed:
--   SELECT ad.type, s.value->'config'->>'agent_type', s.value->>'output_field'
--   FROM agent_definitions ad,
--   LATERAL jsonb_path_query(ad.default_config,'$.**.steps') AS steps,
--   LATERAL jsonb_each(steps) AS s(key,value)
--   WHERE s.key='write_page_content' AND ad.is_active
--     AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
--   -> pageflow-builder|page-content-writer|page_content
--   -> site-work-orchestrator|page-content-writer|page_content
--   -> page-rebuild|page-content-writer|page_content   [the one proven live by 310] )
-- So `page_content.response.sections_metadata` is byte-identical to the value
-- page-build-handler and page-rebuild already use, and the writer's compile_page
-- returns sections_metadata among its four output keys.
--
-- WHY tool-recreation-handler IS NOT TOUCHED — its [UNMEASURED] flag in the bug
-- file is now MEASURED and the answer is that it is not this defect. Its step
-- graph has no writer call at all: recreate_tool (execute_llm_prompt ->
-- tool_recreation), validate_tool (validate_page_content -> validation_result),
-- saving from validation_result.clean_html. It recreates a whole-page tool as one
-- HTML blob; there is no sections_metadata anywhere on that path and none is
-- expected. A NULL content_data there is the correct shape, and the rerender path
-- agrees with that reading — rerender_page_sections_action.go:318 exempts a
-- self-contained tool section from the missing-content escalation by design.
--
-- WHAT CHANGES AT RUNTIME FOR THESE TWO, DECLARED. The save switches from the HTML
-- parse of `assembled_page.html` to the structured metadata array. The persisted
-- section bytes then come from each entry's `rendered_html` rather than from the
-- re-parsed assembled page — near-identical by construction (the assembled page is
-- built FROM those same strings) but not guaranteed byte-identical, and slot names
-- come from `component_function` rather than the parsed data-component attribute.
-- Both agents are dormant — absent from agent_run_stats, which spans 2026-07-26 to
-- now across 95 agent types and does track orchestrator-shaped agents
-- (build-dispatch-loop 2352, content-feed-orchestrator 161) — so this changes no
-- traffic today. It is applied now because a dormant agent that wakes up should
-- wake up correct.
--
-- VERIFICATION on a live run: bugs_open/194's own query on the rebuilt page —
--   SELECT slot_name, length(rendered_html), length(content_data::text), updated_at
--   FROM page_components WHERE page_id='<uuid>' ORDER BY position;
-- content_data non-NULL on every row at the new run's updated_at. Disconfirming
-- outcome, stated in advance: if content_data is STILL NULL, the writer's reply is
-- not reaching the save under this path and the key name is not the fault — check
-- that `page_content.response` carries `sections_metadata` on THAT run first.

BEGIN;

SELECT snapshot_agent('pageflow-builder',
    'pre-update: bugs_open/194 — save_sections never mapped sections_metadata, so every save NULLed content_data');
SELECT snapshot_agent('site-work-orchestrator',
    'pre-update: bugs_open/194 — save_sections never mapped sections_metadata, so every save NULLed content_data');

-- updated_at is set explicitly: there is NO trigger on agent_definitions, so the
-- column is current only if the seed sets it, and a stale one reads as "nobody has
-- touched this row" to the next session.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,sections_metadata_field}',
        '"page_content.response.sections_metadata"'::jsonb,
        true),
    updated_at = NOW()
WHERE type = 'pageflow-builder'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,sections_metadata_field}',
        '"page_content.response.sections_metadata"'::jsonb,
        true),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg      jsonb;
    expected CONSTANT text := 'page_content.response.sections_metadata';
BEGIN
    -- pageflow-builder ------------------------------------------------------
    SELECT default_config #> '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections}'
    INTO cfg FROM agent_definitions
    WHERE type = 'pageflow-builder'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '194/312: pageflow-builder has no build_pages_loop.save_sections step';
    END IF;

    -- IS DISTINCT FROM, not <>: a missing jsonb path is NULL and `NULL <> 'x'` is
    -- NULL, so a plain <> against an absent key can never fire.
    IF cfg #>> '{config,sections_metadata_field}' IS DISTINCT FROM expected THEN
        RAISE EXCEPTION '194/312: pageflow-builder sections_metadata_field is %, expected %',
            COALESCE(cfg #>> '{config,sections_metadata_field}', '<NULL>'), expected;
    END IF;

    -- the pre-existing keys must survive: this is an ADD, not a rewrite
    IF cfg #>> '{config,html_field}'         IS DISTINCT FROM 'assembled_page.html'
       OR cfg #>> '{config,site_id_field}'   IS DISTINCT FROM 'site_record.site_id'
       OR cfg #>> '{config,page_name_field}' IS DISTINCT FROM 'current_page.name' THEN
        RAISE EXCEPTION '194/312: a pageflow-builder save_sections config key was disturbed — %', cfg->'config';
    END IF;

    IF cfg->>'action' IS DISTINCT FROM 'save_page_sections'
       OR cfg->>'next_step' IS DISTINCT FROM 'update_page_status' THEN
        RAISE EXCEPTION '194/312: pageflow-builder save_sections step shape changed';
    END IF;

    -- site-work-orchestrator ------------------------------------------------
    SELECT default_config #> '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections}'
    INTO cfg FROM agent_definitions
    WHERE type = 'site-work-orchestrator'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '194/312: site-work-orchestrator has no build_items_loop.save_sections step';
    END IF;

    IF cfg #>> '{config,sections_metadata_field}' IS DISTINCT FROM expected THEN
        RAISE EXCEPTION '194/312: site-work-orchestrator sections_metadata_field is %, expected %',
            COALESCE(cfg #>> '{config,sections_metadata_field}', '<NULL>'), expected;
    END IF;

    -- NOTE the third key differs from pageflow-builder's: this loop iterates work
    -- ITEMS, so the page name lives at current_item.spec.name. Asserted verbatim
    -- rather than copied from the sibling above, which is exactly the copy this
    -- whole bug is about.
    IF cfg #>> '{config,html_field}'         IS DISTINCT FROM 'assembled_page.html'
       OR cfg #>> '{config,site_id_field}'   IS DISTINCT FROM 'site_record.site_id'
       OR cfg #>> '{config,page_name_field}' IS DISTINCT FROM 'current_item.spec.name' THEN
        RAISE EXCEPTION '194/312: a site-work-orchestrator save_sections config key was disturbed — %', cfg->'config';
    END IF;

    IF cfg->>'action' IS DISTINCT FROM 'save_page_sections'
       OR cfg->>'next_step' IS DISTINCT FROM 'update_page_status' THEN
        RAISE EXCEPTION '194/312: site-work-orchestrator save_sections step shape changed';
    END IF;

    RAISE NOTICE '194/312 OK — sections_metadata_field added to both callers, existing keys intact';
END $$;

COMMIT;

-- ROLLBACK if needed:
--   UPDATE agent_definitions SET default_config = default_config
--     #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,sections_metadata_field}'
--   WHERE type='pageflow-builder' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   UPDATE agent_definitions SET default_config = default_config
--     #- '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,sections_metadata_field}'
--   WHERE type='site-work-orchestrator' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
