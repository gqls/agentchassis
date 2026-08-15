-- 426 ROLLBACK — reverse of 426_tool_improver_instance_pinned_load.sql (bugs_open/281).
-- Restores the component-only load_tool and the function-derived delivery slot.
-- Guarded on the migrated shape; a second run or a never-applied 426 is a no-op.
-- Alternative restore: the agent_definitions_backup row 426 took.
--
-- Verify after applying (expect params = ["input_data.component_id"], slot_name_path
-- = tool_data.function, key_suffix = update_result.component_id):
--   SELECT default_config #> '{workflow,steps,load_tool,config,params}',
--          default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}',
--          default_config #>> '{workflow,steps,create_rerender_item,config,item_key_suffix_field}'
--   FROM agent_definitions WHERE type='tool-improver' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('tool-improver', '426_ROLLBACK: pre-reversal');

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,load_tool,config,query}',
            to_jsonb(
                'SELECT cc.id::text as component_id, cc.function, cc.display_name, cc.html_template, cc.description, cc.semantic_tags::text as tags, cc.is_dark_section, p.url as page_url, p.id::text as page_id, p.name as page_name FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1'::text
            )
        ),
        '{workflow,steps,load_tool,config,params}',
        '["input_data.component_id"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-improver' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id", "input_data.spec.page_id"]'::jsonb;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}',
            '"tool_data.function"'::jsonb
        ),
        '{workflow,steps,create_rerender_item,config,item_key_suffix_field}',
        '"update_result.component_id"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-improver' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}' = 'tool_data.slot_name';

DELETE FROM schema_migrations WHERE filename = '426_tool_improver_instance_pinned_load.sql';

SELECT default_config #> '{workflow,steps,load_tool,config,params}' AS params,
       default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}' AS slot_name_path,
       default_config #>> '{workflow,steps,create_rerender_item,config,item_key_suffix_field}' AS key_suffix
FROM agent_definitions
WHERE type = 'tool-improver' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

COMMIT;
