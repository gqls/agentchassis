-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================


-- page-content-writer
-- Writes content for a single page, may spawn research-agent internally
UPDATE agent_definitions
SET input_contract = '{
    "required": ["current_page", "site_record"],
    "optional": ["reviewed_brief", "style_collection", "db_sync", "generated_images"]
}'::jsonb,
    output_contract = '{
    "produces": ["page_html", "metadata", "seo_data"]
}'::jsonb
WHERE type = 'page-content-writer';

--
use input mapping

-- Update write_page_content step in pageflow-builder to use input_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config}',
        '{
            "agent_type": "page-content-writer",
            "target_role": "content_writer",
            "input_mapping": {
                "current_page": "current_page",
                "site_record": "site_record",
                "reviewed_brief": "reviewed_brief",
                "style_collection": "style_collection",
                "db_sync": "db_sync"
            },
            "timeout_seconds": 300
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- Verify the change
SELECT
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config' as write_page_content_config
FROM agent_definitions
WHERE type = 'pageflow-builder';


--- updating call site planner to use input_mapping

-- Update call_site_planner step to use input_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_site_planner,config}',
        '{
            "agent_type": "site-planner",
            "target_role": "planner",
            "input_mapping": {
                "input_data": "input_data",
                "site_record": "site_record",
                "reviewed_brief": "reviewed_brief"
            },
            "timeout_seconds": 120
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'call_site_planner'->'config' as call_site_planner_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

---

-- Update call_site_planner step with CORRECT paths
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_site_planner,config}',
        '{
            "agent_type": "site-planner",
            "target_role": "planner",
            "input_mapping": {
                "input_data": "input_data",
                "site_record": "site_record",
                "reviewed_brief": "input_data.reviewed_brief"
            },
            "timeout_seconds": 120
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

--

fix

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config}',
        '{
            "agent_type": "page-content-writer",
            "target_role": "content_writer",
            "input_mapping": {
                "current_page": "current_page",
                "site_record": "site_record",
                "reviewed_brief": "input_data.reviewed_brief",
                "style_collection": "style_collection",
                "db_sync": "db_sync"
            },
            "timeout_seconds": 300
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';



===

clients_db=# SELECT
                 type,
                 default_config->'workflow'->'steps'->'build_pages_loop'->>'next_step' as current_next_step,
                 default_config->'workflow'->'steps'->'trigger_site_deploy'->>'action' as trigger_step_exists
             FROM agent_definitions
             WHERE type = 'pageflow-builder';
type       | current_next_step  | trigger_step_exists
------------------+--------------------+---------------------
 pageflow-builder | update_site_status | call_agent



UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,next_step}',
        '"trigger_site_deploy"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify the fix
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->>'next_step' as build_pages_loop_next,
    default_config->'workflow'->'steps'->'trigger_site_deploy'->>'next_step' as trigger_site_deploy_next,
    default_config->'workflow'->'steps'->'update_site_status'->>'next_step' as update_site_status_next
FROM agent_definitions
WHERE type = 'pageflow-builder';


    type       | build_pages_loop_next | trigger_site_deploy_next | update_site_status_next
------------------+-----------------------+--------------------------+-------------------------
    pageflow-builder | trigger_site_deploy   | update_site_status       | complete


