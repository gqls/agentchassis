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


-- adding images

               -- ============================================================
-- PART 1: Add deploy_hero_image step to pageflow-builder workflow
-- ============================================================

-- Add the new step after store_hero_asset
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_hero_image}',
        '{
            "action": "deploy_image_asset",
            "config": {
                "purpose": "hero",
                "uri_field": "hero_result.image_uri",
                "domain_field": "site_record.domain"
            },
            "next_step": "select_style_collection",
            "description": "Download, optimize and deploy hero image to git",
            "output_field": "hero_deployed"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Update store_hero_asset to flow into deploy_hero_image
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,next_step}',
        '"deploy_hero_image"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';


---

-- Update the conditional to check the correct path
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_hero_images,config,condition}',
        '"site_plan.needs_images == true OR site_plan.response.needs_images == true"'
                     )
WHERE type = 'pageflow-builder';

-- Also fix check_assets_needed which has the same issue
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_assets_needed,config,condition}',
        '"site_plan.needs_logo == true OR site_plan.needs_images == true OR site_plan.response.needs_logo == true OR site_plan.response.needs_images == true"'
                     )
WHERE type = 'pageflow-builder';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'check_hero_images'->'config'->>'condition' as hero_condition,
    default_config->'workflow'->'steps'->'check_assets_needed'->'config'->>'condition' as assets_condition
FROM agent_definitions
WHERE type = 'pageflow-builder';


