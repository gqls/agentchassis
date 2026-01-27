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

-- images - content_mapping
-- Fix pageflow-builder workflow to add output_mapping for image generation steps
-- This flattens the nested response data so store_hero_asset and store_logo_asset can find image_url

-- First, let's see the current state
SELECT type, version,
       jsonb_pretty(default_config->'workflow'->'steps'->'generate_hero_image') as generate_hero_image,
       jsonb_pretty(default_config->'workflow'->'steps'->'call_logo_generation') as call_logo_generation
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Update generate_hero_image step to add output_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_hero_image,config,output_mapping}',
        '{
            "image_uri": "generate.response.image_uri",
            "image_url": "generate.response.image_url",
            "prompt": "generate.response.prompt",
            "generated_at": "generate.response.generated_at"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Update call_logo_generation step to add output_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_generation,config,output_mapping}',
        '{
            "image_uri": "generate.response.image_uri",
            "image_url": "generate.response.image_url",
            "prompt": "generate.response.prompt",
            "generated_at": "generate.response.generated_at"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify the changes
SELECT type, version,
       jsonb_pretty(default_config->'workflow'->'steps'->'generate_hero_image'->'config') as hero_config,
       jsonb_pretty(default_config->'workflow'->'steps'->'call_logo_generation'->'config') as logo_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- image paths

-- Fix pageflow-builder workflow to use input_mapping for prompt instead of template syntax
-- This is the proper way to pass dynamic data to child agents

-- First, check current state of the image generation steps
SELECT type, version,
       jsonb_pretty(default_config->'workflow'->'steps'->'generate_hero_image'->'config') as hero_config,
       jsonb_pretty(default_config->'workflow'->'steps'->'call_logo_generation'->'config') as logo_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Update generate_hero_image: remove prompt from config, add it to input_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
    -- First remove the prompt key from config
        jsonb_set(
                default_config,
                '{workflow,steps,generate_hero_image,config}',
                (default_config->'workflow'->'steps'->'generate_hero_image'->'config') - 'prompt'
        ),
    -- Then update input_mapping to include prompt
        '{workflow,steps,generate_hero_image,config,input_mapping}',
        '{
            "prompt": "site_plan.image_prompts.hero_home",
            "site_plan": "site_plan",
            "reviewed_brief": "input_data.reviewed_brief"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Update call_logo_generation: remove prompt from config, add it to input_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
    -- First remove the prompt key from config
        jsonb_set(
                default_config,
                '{workflow,steps,call_logo_generation,config}',
                (default_config->'workflow'->'steps'->'call_logo_generation'->'config') - 'prompt'
        ),
    -- Then update input_mapping to include prompt
        '{workflow,steps,call_logo_generation,config,input_mapping}',
        '{
            "prompt": "site_plan.image_prompts.logo",
            "site_plan": "site_plan",
            "reviewed_brief": "input_data.reviewed_brief"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify the changes
SELECT type, version,
       jsonb_pretty(default_config->'workflow'->'steps'->'generate_hero_image'->'config') as hero_config,
       jsonb_pretty(default_config->'workflow'->'steps'->'call_logo_generation'->'config') as logo_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Expected result for hero_config:
-- {
--     "agent_type": "image-generator",
--     "target_role": "image_generator",
--     "input_mapping": {
--         "prompt": "site_plan.image_prompts.hero_home",
--         "site_plan": "site_plan",
--         "reviewed_brief": "input_data.reviewed_brief"
--     },
--     "timeout_seconds": 120
-- }
-- Note: no "prompt": "{{...}}" in config anymore

--
-- fixing recursive image path error
-- Fix the input_mapping paths to include .response

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_hero_image,config,input_mapping}',
        '{
            "prompt": "site_plan.response.image_prompts.hero_home",
            "site_plan": "site_plan",
            "reviewed_brief": "input_data.reviewed_brief"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_generation,config,input_mapping}',
        '{
            "prompt": "site_plan.response.image_prompts.logo",
            "site_plan": "site_plan",
            "reviewed_brief": "input_data.reviewed_brief"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify
SELECT type, version,
       default_config->'workflow'->'steps'->'generate_hero_image'->'config'->'input_mapping' as hero_input_mapping,
       default_config->'workflow'->'steps'->'call_logo_generation'->'config'->'input_mapping' as logo_input_mapping
FROM agent_definitions
WHERE type = 'pageflow-builder';


--url paths

-- Add output_mapping to deploy_hero_image step
-- This flattens the git adapter response so hero_deployed.image_url is accessible
--
-- The git adapter now returns:
-- {
--   "success": true,
--   "file_path": "/assets/images/hero.jpg",
--   "files": ["/assets/images/hero.jpg"],
--   ...
-- }
--
-- Without output_mapping, this is stored as:
--   hero_deployed.response.data.file_path
--
-- With output_mapping, we flatten it to:
--   hero_deployed.image_url (mapped from response.data.file_path)

-- First check current state
SELECT type, version,
       jsonb_pretty(default_config->'workflow'->'steps'->'deploy_hero_image') as deploy_hero,
       jsonb_pretty(default_config->'workflow'->'steps'->'deploy_logo_image') as deploy_logo
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Update deploy_hero_image to add output_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_hero_image,config,output_mapping}',
        '{
            "image_url": "response.data.file_path",
            "deployed": "response.data.success",
            "files": "response.data.files",
            "repo_url": "response.data.repo_url",
            "domain": "response.data.domain"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Check if deploy_logo_image step exists and update it too
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_logo_image,config,output_mapping}',
        '{
            "image_url": "response.data.file_path",
            "deployed": "response.data.success",
            "files": "response.data.files",
            "repo_url": "response.data.repo_url",
            "domain": "response.data.domain"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder'
  AND default_config->'workflow'->'steps'->'deploy_logo_image' IS NOT NULL;

-- Verify the changes
SELECT type, version,
       jsonb_pretty(default_config->'workflow'->'steps'->'deploy_hero_image'->'config') as hero_config,
       jsonb_pretty(default_config->'workflow'->'steps'->'deploy_logo_image'->'config') as logo_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Expected hero_config after update:
-- {
--     "purpose": "hero",
--     "uri_field": "hero_result.image_uri",
--     "domain_field": "site_record.domain",
--     "output_mapping": {
--         "image_url": "response.data.file_path",
--         "deployed": "response.data.success",
--         "files": "response.data.files",
--         "repo_url": "response.data.repo_url",
--         "domain": "response.data.domain"
--     }
-- }
--
-- After this, hero_deployed will contain:
-- {
--     "image_url": "/assets/images/hero.jpg",
--     "deployed": true,
--     "files": ["/assets/images/hero.jpg"],
--     "repo_url": "https://github.com/gqls/sites",
--     "domain": "leopardessconsulting.co.uk",
--     "response_received_at": "...",
--     "response_status": "complete"
-- }
--
-- Then BuildRenderContextAction can find hero_deployed.image_url