-- 434_..._ROLLBACK.sql — surgical inverse of 434.
--
-- Restores the (dead) map-valued spec_data verbatim on the four steps and
-- removes the spec_paths/spec_literal keys 434 added. Guarded to abort if the
-- steps are not in 434's post state (i.e. someone has edited them since).
--
-- DELIBERATELY NOT REVERTED: the site_work_items backfill/re-arm. That was a
-- data repair from the rows' own columns; emptying specs again would only
-- re-break the items. If the re-armed items must be stopped, cancel them by id.

ROLLBACK;

BEGIN;

DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,spec_paths,page_id}' = 'tool_data.page_id'
      AND NOT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}') ? 'spec_data';
    IF n <> 1 THEN
        RAISE EXCEPTION '434-ROLLBACK: tool-auditor is not in 434''s post state (found %), aborting', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='internal-linker' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}') ? 'spec_paths';
    IF n <> 1 THEN
        RAISE EXCEPTION '434-ROLLBACK: internal-linker is not in 434''s post state (found %), aborting', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='component-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND (default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}') ? 'spec_paths';
    IF n <> 1 THEN
        RAISE EXCEPTION '434-ROLLBACK: component-quality-auditor is not in 434''s post state (found %), aborting', n;
    END IF;

    RAISE NOTICE '434-ROLLBACK: pre-flight OK';
END $$;

SELECT snapshot_agent('tool-auditor', '434-ROLLBACK: restoring map-valued spec_data on both loop create steps');
SELECT snapshot_agent('internal-linker', '434-ROLLBACK: restoring map-valued spec_data on create_rewrite_item');
SELECT snapshot_agent('component-quality-auditor', '434-ROLLBACK: restoring map-valued spec_data on create_work_item');

UPDATE agent_definitions
SET default_config =
    jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}',
            ((default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}') - 'spec_paths' - 'spec_literal')
            || '{"spec_data": {"check": "tool_auditor",
                               "issue": "current_finding.description",
                               "page_id": "tool_data.page_id",
                               "category": "current_finding.category",
                               "page_name": "tool_data.page_name",
                               "component_id": "tool_data.component_id",
                               "fix_suggestion": "current_finding.fix_suggestion"}}'::jsonb
        ),
        '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}',
        ((default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}') - 'spec_paths' - 'spec_literal')
        || '{"spec_data": {"check": "tool_auditor",
                           "issue": "current_finding.description",
                           "page_id": "tool_data.page_id",
                           "category": "current_finding.category",
                           "page_name": "tool_data.page_name",
                           "confidence": "current_finding.confidence",
                           "component_id": "tool_data.component_id",
                           "tool_function": "tool_data.function",
                           "fix_suggestion": "current_finding.fix_suggestion"}}'::jsonb
    ),
    updated_at = NOW()
WHERE type='tool-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}') ? 'spec_data';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}',
        ((default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}') - 'spec_paths' - 'spec_literal')
        || '{"spec_data": {"source": "internal-linker",
                           "page_name": "current_link.source_page",
                           "suggestion": "current_link.guidance",
                           "anchor_text": "current_link.anchor_text",
                           "link_target_url": "target_page.url",
                           "link_target_title": "target_page.title"}}'::jsonb
    ),
    updated_at = NOW()
WHERE type='internal-linker' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}') ? 'spec_data';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}',
        ((default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}') - 'spec_paths' - 'spec_literal')
        || '{"spec_data": {"function": "current_component.function",
                           "component_id": "current_component.component_id",
                           "quality_score": "current_component.quality_score",
                           "quality_issues": "current_component.quality_issues"}}'::jsonb
    ),
    updated_at = NOW()
WHERE type='component-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}') ? 'spec_data';

DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions ad,
         jsonb_each(ad.default_config->'workflow'->'steps') outer_steps,
         jsonb_each(coalesce(outer_steps.value->'config'->'sub_workflow'->'steps', '{}'::jsonb)) inner_steps
    WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
      AND jsonb_typeof(inner_steps.value->'config'->'spec_data') = 'object';
    IF n <> 4 THEN
        RAISE EXCEPTION '434-ROLLBACK: post-condition failed — expected the 4 map-valued spec_data steps back, found %', n;
    END IF;
    RAISE NOTICE '434-ROLLBACK: post-condition OK — pre-434 config restored';
END $$;

DELETE FROM schema_migrations
WHERE filename = '434_spec_data_map_is_never_read_four_producers_file_empty_specs.sql';

COMMIT;
