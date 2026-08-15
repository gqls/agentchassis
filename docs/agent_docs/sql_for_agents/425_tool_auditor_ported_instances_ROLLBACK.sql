-- 425 ROLLBACK — reverse of 425_tool_auditor_ported_instances.sql (bugs_open/281).
-- Restores the component-only load_tool, the html_template prompt source, the
-- confidence-split loop start, and removes the per-page key/identity fields.
-- Guarded on the migrated shape; a second run or a never-applied 425 is a no-op.
-- Alternative restore: the agent_definitions_backup row 425 took (snapshot_reason
-- '425: bugs_open/281 …'; check it holds the PRE-change load_tool query).
--
-- Verify after applying (expect one row, params = ["input_data.component_id"],
-- start_step = check_confidence, no check_target_class step):
--   SELECT default_config #> '{workflow,steps,load_tool,config,params}',
--          default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,start_step}',
--          default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps}' ? 'check_target_class'
--   FROM agent_definitions WHERE type='tool-auditor' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('tool-auditor', '425_ROLLBACK: pre-reversal');

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,load_tool,config,query}',
            to_jsonb(
                'SELECT cc.id::text AS component_id, cc.function, cc.display_name, cc.html_template, COALESCE(pc.rendered_html, '''') AS rendered_html, cc.description, p.id::text AS page_id, p.name AS page_name, p.url AS page_url, COALESCE(p.build_status, '''') AS build_status FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1'::text
            )
        ),
        '{workflow,steps,load_tool,config,params}',
        '["input_data.component_id"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-auditor' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id", "input_data.spec.page_id"]'::jsonb;

UPDATE agent_definitions
SET default_config = replace(default_config::text, '{{.tool_data.source_html}}', '{{.tool_data.html_template}}')::jsonb,
    updated_at = NOW()
WHERE type = 'tool-auditor' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%{{.tool_data.source_html}}%';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,check_target_class}',
        '{workflow,steps,create_items_loop,config,sub_workflow,start_step}',
        '"check_confidence"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-auditor' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,start_step}' = 'check_target_class';

UPDATE agent_definitions
SET default_config = default_config
        #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,page_id}'
        #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,item_key_suffix_field}'
        #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,page_id}'
        #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,item_key_suffix_field}'
        #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,spec_data,page_id}'
        #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,spec_data,page_name}',
    updated_at = NOW()
WHERE type = 'tool-auditor' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DELETE FROM schema_migrations WHERE filename = '425_tool_auditor_ported_instances.sql';

SELECT default_config #> '{workflow,steps,load_tool,config,params}' AS params,
       default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,start_step}' AS start_step,
       default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps}' ? 'check_target_class' AS gate_present
FROM agent_definitions
WHERE type = 'tool-auditor' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

COMMIT;
