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

---

-- hero_url input mapping

-- Add hero_url to write_page_content input_mapping
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,input_mapping}',
        '{
            "db_sync": "db_sync",
            "site_record": "site_record",
            "current_page": "current_page",
            "reviewed_brief": "input_data.reviewed_brief",
            "style_collection": "style_collection",
            "hero_url": "hero_url",
            "logo_url": "logo_url",
            "brand_logo_url": "brand_logo_url"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify
SELECT default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config'->'input_mapping'
FROM agent_definitions
WHERE type = 'pageflow-builder';

--

-- Update write_page_content input_mapping with optional fields (? suffix)
-- Run this AFTER deploying the code change to support optional fields

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,input_mapping}',
        '{
            "db_sync": "db_sync",
            "site_record": "site_record",
            "current_page": "current_page",
            "reviewed_brief": "input_data.reviewed_brief",
            "style_collection": "style_collection",
            "hero_url?": "hero_url",
            "logo_url?": "logo_url",
            "brand_logo_url?": "brand_logo_url"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify
SELECT default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config'->'input_mapping' as input_mapping
FROM agent_definitions
WHERE type = 'pageflow-builder';


------

-- Integration: Add webdesign-agent to pageflow-builder
--
-- Requires:
-- 1. Spawn the agent (before it can be called)
-- 2. Call it after build_pages_loop, before trigger_site_deploy

-- Step 1: Add spawn step
-- Insert: ... → spawn_image_generator → spawn_webdesign_agent → generate_logo → ...
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_webdesign_agent}',
        '{
            "action": "spawn_agent",
            "config": {
                "role": "webdesigner",
                "agent_type": "webdesign-agent"
            },
            "description": "Spawn webdesign agent",
            "next_step": "generate_logo",
            "output_field": "webdesign_agent"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Update spawn_image_generator to point to spawn_webdesign_agent
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_image_generator,next_step}',
        '"spawn_webdesign_agent"'
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Step 2: Add apply_site_design step
-- Insert: ... → build_pages_loop → apply_site_design → trigger_site_deploy → ...
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_site_design}',
        '{
            "action": "call_agent",
            "config": {
                "agent_type": "webdesign-agent",
                "target_role": "webdesigner",
                "input_mapping": {
                    "site_id": "site_record.site_id",
                    "domain": "site_record.domain"
                },
                "timeout_seconds": 300
            },
            "description": "Generate and deploy site stylesheet",
            "next_step": "trigger_site_deploy",
            "output_field": "design_result"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Update build_pages_loop to point to apply_site_design
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,next_step}',
        '"apply_site_design"'
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'spawn_image_generator'->>'next_step' as after_spawn_img,
    default_config->'workflow'->'steps'->'spawn_webdesign_agent'->>'next_step' as after_spawn_design,
    default_config->'workflow'->'steps'->'build_pages_loop'->>'next_step' as after_build,
    default_config->'workflow'->'steps'->'apply_site_design'->>'next_step' as after_design
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Expected:
-- after_spawn_img     | spawn_webdesign_agent
-- after_spawn_design  | generate_logo
-- after_build         | apply_site_design
-- after_design        | trigger_site_deploy


------------

-- current backup
427aa3e5-5ea2-4917-8d24-d751ebd283b2 | pageflow-builder | PageFlow Builder | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["site_record", "pages_built", "deployment_result"]}, "description": "Site build complete"}, "generate_logo": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true", "else_step": "check_hero_images", "then_step": "call_logo_generation"}, "description": "Check if logo needs to be generated"}, "spawn_planner": {"action": "spawn_agent", "config": {"role": "planner", "agent_type": "site-planner"}, "next_step": "spawn_content_writer", "description": "Spawn site planner agent", "output_field": "planner_agent"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "ensure_site_record", "description": "Spawn deployer agent", "output_field": "deployer_agent"}, "spawn_reviewer": {"action": "spawn_agent", "config": {"role": "reviewer", "agent_type": "content-reviewer"}, "next_step": "spawn_deployer", "description": "Spawn content reviewer agent", "output_field": "reviewer_agent"}, "store_site_plan": {"action": "update_site_content", "config": {"merge": true, "content_field": "site_plan", "site_id_field": "site_record.site_id"}, "next_step": "sync_pages_to_db", "description": "Store the site plan in sites.content_data", "output_field": "content_stored"}, "build_pages_loop": {"action": "loop", "config": {"mode": "sequential", "items_field": "pages_to_build.pages", "sub_workflow": {"steps": {"deploy_page": {"action": "git_commit", "config": {"page_field": "current_page", "domain_field": "site_record.domain", "content_field": "assembled_page.html"}, "next_step": "update_page_status", "description": "Commit page to git", "output_field": "page_deployed"}, "assemble_page": {"action": "assemble_page", "config": {"content_field": "page_content.response.page_html", "add_navigation": false}, "next_step": "deploy_page", "description": "Assemble full page HTML from components", "output_field": "assembled_page"}, "complete_page": {"action": "loop_complete", "description": "Page build complete"}, "update_page_status": {"action": "update_page_status", "config": {"status": "deployed", "commit_from": "page_deployed.commit_sha", "page_id_field": "current_page.id"}, "next_step": "complete_page", "description": "Mark page as deployed in database"}, "write_page_content": {"action": "call_agent", "config": {"agent_type": "page-content-writer", "target_role": "content_writer", "input_mapping": {"db_sync": "db_sync", "hero_url?": "hero_url", "logo_url?": "logo_url", "site_record": "site_record", "current_page": "current_page", "reviewed_brief": "input_data.reviewed_brief", "brand_logo_url?": "brand_logo_url", "style_collection": "style_collection"}, "timeout_seconds": 300}, "next_step": "review_page_content", "description": "Write content for this page", "output_field": "page_content"}, "review_page_content": {"action": "call_agent", "config": {"agent_type": "content-reviewer", "target_role": "reviewer", "input_mapping": {"current_page": "current_page", "page_content": "page_content", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 3900}, "next_step": "assemble_page", "description": "Review page content (HITL or auto-eval)", "output_field": "reviewed_content"}}, "start_step": "write_page_content"}, "item_variable": "current_page", "max_iterations": 20}, "next_step": "apply_site_design", "description": "Build each page: write → review → deploy", "output_field": "pages_built"}, "store_hero_asset": {"action": "store_asset", "config": {"purpose": "hero", "asset_type": "image", "data_field": "hero_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "hero_images.home", "origin_prompt_field": "site_plan.image_prompts.hero_home", "update_site_brand_assets": true}, "next_step": "deploy_hero_image", "description": "Store generated hero image", "output_field": "hero_stored"}, "store_logo_asset": {"action": "store_asset", "config": {"purpose": "brand_logo", "asset_type": "logo", "data_field": "logo_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "logo.primary", "origin_prompt_field": "site_plan.image_prompts.logo", "update_site_brand_assets": true}, "next_step": "check_hero_images", "description": "Store generated logo in assets table and site brand_assets", "output_field": "logo_stored"}, "sync_pages_to_db": {"action": "sync_pages_to_db", "config": {"input_fields": ["site_record", "site_plan"]}, "next_step": "check_assets_needed", "description": "Create page records from site plan", "output_field": "db_sync"}, "apply_site_design": {"action": "call_agent", "config": {"agent_type": "webdesign-agent", "target_role": "webdesigner", "input_mapping": {"domain": "site_record.domain", "site_id": "site_record.site_id"}, "timeout_seconds": 300}, "next_step": "trigger_site_deploy", "description": "Generate and deploy site stylesheet", "output_field": "design_result"}, "call_site_planner": {"action": "call_agent", "config": {"agent_type": "site-planner", "target_role": "planner", "input_mapping": {"input_data": "input_data", "site_record": "site_record", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 120}, "next_step": "store_reviewed_brief", "description": "Plan pages, select components, identify asset needs", "output_field": "site_plan"}, "check_hero_images": {"action": "conditional", "config": {"condition": "site_plan.needs_images == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "generate_hero_image"}, "description": "Check if hero images need to be generated"}, "deploy_hero_image": {"action": "deploy_image_asset", "config": {"purpose": "hero", "uri_field": "hero_result.image_uri", "domain_field": "site_record.domain", "output_mapping": {"files": "response.data.files", "domain": "response.data.domain", "deployed": "response.data.success", "repo_url": "response.data.repo_url", "image_url": "response.data.file_path"}}, "next_step": "select_style_collection", "description": "Download, optimize and deploy hero image to git", "output_field": "hero_deployed"}, "ensure_site_record": {"action": "ensure_site_record", "config": {"store_brief_in_content_data": true}, "next_step": "call_site_planner", "description": "Create or update site record in database", "output_field": "site_record"}, "get_pages_to_build": {"action": "get_pages_to_build", "config": {"include_all": false, "build_statuses": ["planned", "needs_rebuild"]}, "next_step": "build_pages_loop", "description": "Query pages from database that need content generation", "output_field": "pages_to_build"}, "update_site_status": {"action": "update_site_status", "config": {"status": "deployed", "deployed_at": "now", "site_id_field": "site_record.site_id"}, "next_step": "complete", "description": "Mark site as deployed", "output_field": "site_updated"}, "check_assets_needed": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true OR site_plan.needs_images == true OR site_plan.response.needs_logo == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "spawn_image_generator"}, "description": "Check if logo or images need to be generated"}, "generate_hero_image": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.hero_home", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_hero_asset", "description": "Generate hero image for home page", "output_field": "hero_result"}, "trigger_site_deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_mapping": {"pages_built": "pages_built", "site_record": "site_record"}, "timeout_seconds": 180}, "next_step": "update_site_status", "description": "Trigger Cloudflare deployment", "output_field": "deployment_result"}, "call_logo_generation": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.logo", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_logo_asset", "description": "Generate logo using image-generator agent", "output_field": "logo_result"}, "spawn_content_writer": {"action": "spawn_agent", "config": {"role": "content_writer", "agent_type": "page-content-writer"}, "next_step": "spawn_reviewer", "description": "Spawn content writer agent", "output_field": "content_writer_agent"}, "store_reviewed_brief": {"action": "update_site_content", "config": {"merge": true, "content_field": "input_data.reviewed_brief", "site_id_field": "site_record.site_id"}, "next_step": "store_site_plan", "description": "Store the reviewed brief in sites.content_data", "output_field": "brief_stored"}, "spawn_image_generator": {"action": "spawn_agent", "config": {"role": "image_generator", "agent_type": "image-generator"}, "next_step": "spawn_webdesign_agent", "description": "Spawn image generator agent for asset creation", "output_field": "image_generator_info"}, "spawn_webdesign_agent": {"action": "spawn_agent", "config": {"role": "webdesigner", "agent_type": "webdesign-agent"}, "next_step": "generate_logo", "description": "Spawn webdesign agent", "output_field": "webdesign_agent"}, "set_default_components": {"action": "update_site_defaults", "config": {"defaults": {"head": "head-seo-standard", "footer_from": "style_collection.footer_component_name", "header_from": "style_collection.header_component_name"}, "site_id_field": "site_record.site_id"}, "next_step": "get_pages_to_build", "description": "Set default head/header/footer components", "output_field": "defaults_set"}, "select_style_collection": {"action": "select_style_collection", "config": {"style_from": "site_plan.style_collection", "site_id_field": "site_record.site_id", "fallback_by_domain": true}, "next_step": "set_default_components", "description": "Choose style collection based on site plan", "output_field": "style_collection"}}, "start_step": "spawn_planner"}, "processing_mode": "orchestrator", "timeout_seconds": 900} | t         | 2025-12-22 17:46:51.419068+00 | 2026-01-30 09:00:48.346758+00 |            | ["orchestration", "website-builder", "component-based", "image-generation"] | docker.io/aqls/agent-chassis | v1.0.732  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |      14 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | ["website", "multipage", "component-based"] | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]} |           0 | f           | {"expects": {"reviewed_brief": "object - completed questionnaire answers", "input_data.domain": "string - the domain name", "input_data.objective": "string - what the site should achieve"}, "required": ["input_data.domain", "reviewed_brief"]} | {"produces": {"site_id": "uuid - the site record ID", "deploy_url": "string - the live site URL", "pages_built": "number - count of pages deployed"}}
(1 row)


        --

-- add save sections step

        -- Add save_sections step to pageflow-builder workflow
-- This saves rendered HTML to page_components.rendered_html for future rerender operations
--
-- The step runs AFTER deploy_page and BEFORE update_page_status

-- ============================================================
-- 1. Add save_sections step to the sub_workflow within build_pages_loop
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections}',
        '{
            "action": "save_page_sections",
            "config": {
                "html_field": "assembled_page.html",
                "page_name_field": "current_page.name",
                "site_id_field": "site_record.site_id"
            },
            "description": "Save rendered sections to page_components for rerender",
            "next_step": "update_page_status",
            "output_field": "save_result"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- ============================================================
-- 2. Update deploy_page to go to save_sections instead of update_page_status
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,deploy_page,next_step}',
        '"save_sections"'
                     )
WHERE type = 'pageflow-builder';

-- ============================================================
-- 3. Verify the update
-- ============================================================
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'deploy_page'->>'next_step' as deploy_next,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'save_sections'->>'action' as save_action,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'save_sections'->>'next_step' as save_next
FROM agent_definitions
WHERE type = 'pageflow-builder';


---

-- adding rerender step one page at a time

-- Add save_sections step to pageflow-builder workflow
-- This saves rendered HTML to page_components.rendered_html for future rerender operations
--
-- The step runs AFTER deploy_page and BEFORE update_page_status

-- ============================================================
-- 1. Add save_sections step to the sub_workflow within build_pages_loop
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections}',
        '{
            "action": "save_page_sections",
            "config": {
                "html_field": "assembled_page.html",
                "page_name_field": "current_page.name",
                "site_id_field": "site_record.site_id"
            },
            "description": "Save rendered sections to page_components for rerender",
            "next_step": "update_page_status",
            "output_field": "save_result"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- ============================================================
-- 2. Update deploy_page to go to save_sections instead of update_page_status
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,deploy_page,next_step}',
        '"save_sections"'
                     )
WHERE type = 'pageflow-builder';

-- ============================================================
-- 3. Verify the update
-- ============================================================
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'deploy_page'->>'next_step' as deploy_next,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'save_sections'->>'action' as save_action,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'save_sections'->>'next_step' as save_next
FROM agent_definitions
WHERE type = 'pageflow-builder';


-- backup before changing contract
427aa3e5-5ea2-4917-8d24-d751ebd283b2 | pageflow-builder | PageFlow Builder | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["site_record", "pages_built", "deployment_result"]}, "description": "Site build complete"}, "generate_logo": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true", "else_step": "check_hero_images", "then_step": "call_logo_generation"}, "description": "Check if logo needs to be generated"}, "spawn_planner": {"action": "spawn_agent", "config": {"role": "planner", "agent_type": "site-planner"}, "next_step": "spawn_content_writer", "description": "Spawn site planner agent", "output_field": "planner_agent"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "ensure_site_record", "description": "Spawn deployer agent", "output_field": "deployer_agent"}, "spawn_reviewer": {"action": "spawn_agent", "config": {"role": "reviewer", "agent_type": "content-reviewer"}, "next_step": "spawn_deployer", "description": "Spawn content reviewer agent", "output_field": "reviewer_agent"}, "store_site_plan": {"action": "update_site_content", "config": {"merge": true, "content_field": "site_plan", "site_id_field": "site_record.site_id"}, "next_step": "sync_pages_to_db", "description": "Store the site plan in sites.content_data", "output_field": "content_stored"}, "build_pages_loop": {"action": "loop", "config": {"mode": "sequential", "items_field": "pages_to_build.pages", "sub_workflow": {"steps": {"deploy_page": {"action": "git_commit", "config": {"page_field": "current_page", "domain_field": "site_record.domain", "content_field": "assembled_page.html"}, "next_step": "save_sections", "description": "Commit page to git", "output_field": "page_deployed"}, "assemble_page": {"action": "assemble_page", "config": {"content_field": "page_content.response.page_html", "add_navigation": false}, "next_step": "deploy_page", "description": "Assemble full page HTML from components", "output_field": "assembled_page"}, "complete_page": {"action": "loop_complete", "description": "Page build complete"}, "save_sections": {"action": "save_page_sections", "config": {"html_field": "assembled_page.html", "site_id_field": "site_record.site_id", "page_name_field": "current_page.name"}, "next_step": "update_page_status", "description": "Save rendered sections to page_components for rerender", "output_field": "save_result"}, "update_page_status": {"action": "update_page_status", "config": {"status": "deployed", "commit_from": "page_deployed.commit_sha", "page_id_field": "current_page.id"}, "next_step": "complete_page", "description": "Mark page as deployed in database"}, "write_page_content": {"action": "call_agent", "config": {"agent_type": "page-content-writer", "target_role": "content_writer", "input_mapping": {"db_sync": "db_sync", "hero_url?": "hero_url", "logo_url?": "logo_url", "site_record": "site_record", "current_page": "current_page", "reviewed_brief": "input_data.reviewed_brief", "brand_logo_url?": "brand_logo_url", "style_collection": "style_collection"}, "timeout_seconds": 300}, "next_step": "review_page_content", "description": "Write content for this page", "output_field": "page_content"}, "review_page_content": {"action": "call_agent", "config": {"agent_type": "content-reviewer", "target_role": "reviewer", "input_mapping": {"site_record": "site_record", "current_page": "current_page", "page_content": "page_content", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 3900}, "next_step": "check_review_approved", "description": "Review page content (HITL or auto-eval)", "output_field": "reviewed_content"}, "check_review_approved": {"action": "conditional", "config": {"condition": "reviewed_content.review_result.approved == true OR reviewed_content.approved == true", "else_step": "complete_page", "then_step": "assemble_page"}, "description": "Check if content was approved - skip deploy if rejected"}}, "start_step": "write_page_content"}, "item_variable": "current_page", "max_iterations": 20}, "next_step": "apply_site_design", "description": "Build each page: write → review → deploy", "output_field": "pages_built"}, "store_hero_asset": {"action": "store_asset", "config": {"purpose": "hero", "asset_type": "image", "data_field": "hero_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "hero_images.home", "origin_prompt_field": "site_plan.image_prompts.hero_home", "update_site_brand_assets": true}, "next_step": "deploy_hero_image", "description": "Store generated hero image", "output_field": "hero_stored"}, "store_logo_asset": {"action": "store_asset", "config": {"purpose": "brand_logo", "asset_type": "logo", "data_field": "logo_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "logo.primary", "origin_prompt_field": "site_plan.image_prompts.logo", "update_site_brand_assets": true}, "next_step": "check_hero_images", "description": "Store generated logo in assets table and site brand_assets", "output_field": "logo_stored"}, "sync_pages_to_db": {"action": "sync_pages_to_db", "config": {"input_fields": ["site_record", "site_plan"]}, "next_step": "check_assets_needed", "description": "Create page records from site plan", "output_field": "db_sync"}, "apply_site_design": {"action": "call_agent", "config": {"agent_type": "webdesign-agent", "target_role": "webdesigner", "input_mapping": {"domain": "site_record.domain", "site_id": "site_record.site_id"}, "timeout_seconds": 300}, "next_step": "trigger_site_deploy", "description": "Generate and deploy site stylesheet", "output_field": "design_result"}, "call_site_planner": {"action": "call_agent", "config": {"agent_type": "site-planner", "target_role": "planner", "input_mapping": {"input_data": "input_data", "site_record": "site_record", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 120}, "next_step": "store_reviewed_brief", "description": "Plan pages, select components, identify asset needs", "output_field": "site_plan"}, "check_hero_images": {"action": "conditional", "config": {"condition": "site_plan.needs_images == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "generate_hero_image"}, "description": "Check if hero images need to be generated"}, "deploy_hero_image": {"action": "deploy_image_asset", "config": {"purpose": "hero", "uri_field": "hero_result.image_uri", "domain_field": "site_record.domain", "output_mapping": {"files": "response.data.files", "domain": "response.data.domain", "deployed": "response.data.success", "repo_url": "response.data.repo_url", "image_url": "response.data.file_path"}}, "next_step": "select_style_collection", "description": "Download, optimize and deploy hero image to git", "output_field": "hero_deployed"}, "ensure_site_record": {"action": "ensure_site_record", "config": {"store_brief_in_content_data": true}, "next_step": "call_site_planner", "description": "Create or update site record in database", "output_field": "site_record"}, "get_pages_to_build": {"action": "get_pages_to_build", "config": {"include_all": false, "build_statuses": ["planned", "needs_rebuild"]}, "next_step": "build_pages_loop", "description": "Query pages from database that need content generation", "output_field": "pages_to_build"}, "update_site_status": {"action": "update_site_status", "config": {"status": "deployed", "deployed_at": "now", "site_id_field": "site_record.site_id"}, "next_step": "complete", "description": "Mark site as deployed", "output_field": "site_updated"}, "check_assets_needed": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true OR site_plan.needs_images == true OR site_plan.response.needs_logo == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "spawn_image_generator"}, "description": "Check if logo or images need to be generated"}, "generate_hero_image": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.hero_home", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_hero_asset", "description": "Generate hero image for home page", "output_field": "hero_result"}, "trigger_site_deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_mapping": {"pages_built": "pages_built", "site_record": "site_record"}, "timeout_seconds": 180}, "next_step": "update_site_status", "description": "Trigger Cloudflare deployment", "output_field": "deployment_result"}, "call_logo_generation": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.logo", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_logo_asset", "description": "Generate logo using image-generator agent", "output_field": "logo_result"}, "spawn_content_writer": {"action": "spawn_agent", "config": {"role": "content_writer", "agent_type": "page-content-writer"}, "next_step": "spawn_reviewer", "description": "Spawn content writer agent", "output_field": "content_writer_agent"}, "store_reviewed_brief": {"action": "update_site_content", "config": {"merge": true, "content_field": "input_data.reviewed_brief", "site_id_field": "site_record.site_id"}, "next_step": "store_site_plan", "description": "Store the reviewed brief in sites.content_data", "output_field": "brief_stored"}, "spawn_image_generator": {"action": "spawn_agent", "config": {"role": "image_generator", "agent_type": "image-generator"}, "next_step": "spawn_webdesign_agent", "description": "Spawn image generator agent for asset creation", "output_field": "image_generator_info"}, "spawn_webdesign_agent": {"action": "spawn_agent", "config": {"role": "webdesigner", "agent_type": "webdesign-agent"}, "next_step": "generate_logo", "description": "Spawn webdesign agent", "output_field": "webdesign_agent"}, "set_default_components": {"action": "update_site_defaults", "config": {"defaults": {"head": "head-seo-standard", "footer_from": "style_collection.footer_component_name", "header_from": "style_collection.header_component_name"}, "site_id_field": "site_record.site_id"}, "next_step": "get_pages_to_build", "description": "Set default head/header/footer components", "output_field": "defaults_set"}, "select_style_collection": {"action": "select_style_collection", "config": {"style_from": "site_plan.style_collection", "site_id_field": "site_record.site_id", "fallback_by_domain": true}, "next_step": "set_default_components", "description": "Choose style collection based on site plan", "output_field": "style_collection"}}, "start_step": "spawn_planner"}, "processing_mode": "orchestrator", "timeout_seconds": 900} | t         | 2025-12-22 17:46:51.419068+00 | 2026-01-31 22:23:28.542796+00 |            | ["orchestration", "website-builder", "component-based", "image-generation"] | docker.io/aqls/agent-chassis | v1.0.739  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |      17 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | ["website", "multipage", "component-based"] | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]} |           0 | f           | {"expects": {"reviewed_brief": "object - completed questionnaire answers", "input_data.domain": "string - the domain name", "input_data.objective": "string - what the site should achieve"}, "required": ["input_data.domain", "reviewed_brief"]} | {"produces": {"site_id": "uuid - the site record ID", "deploy_url": "string - the live site URL", "pages_built": "number - count of pages deployed"}}
(1 row)

-- Check current input_contract
SELECT type, input_contract FROM agent_definitions WHERE type = 'pageflow-builder';

clients_db=# SELECT type, input_contract FROM agent_definitions WHERE type = 'pageflow-builder';
type       |                                                                                                                   input_contract
------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 pageflow-builder | {"expects": {"reviewed_brief": "object - completed questionnaire answers", "input_data.domain": "string - the domain name", "input_data.objective": "string - what the site should achieve"}, "required": ["input_data.domain", "reviewed_brief"]}
(1 row)

-- Update to accept input_data as object
UPDATE agent_definitions
SET input_contract = jsonb_set(
        COALESCE(input_contract, '{}'),
        '{required}',
        '["input_data", "reviewed_brief"]'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET input_contract = '{
    "expects": {
        "reviewed_brief": "object - completed questionnaire answers",
        "input_data.domain": "string - the domain name",
        "input_data.objective": "string - what the site should achieve"
    },
    "required": ["input_data", "reviewed_brief"]
}'::jsonb,
version = version + 1,
updated_at = NOW()
WHERE type = 'pageflow-builder';


-- backup before major refactor render and site components, site area components etc
427aa3e5-5ea2-4917-8d24-d751ebd283b2 | pageflow-builder | PageFlow Builder | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["site_record", "pages_built", "deployment_result"]}, "description": "Site build complete"}, "generate_logo": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true", "else_step": "check_hero_images", "then_step": "call_logo_generation"}, "description": "Check if logo needs to be generated"}, "spawn_planner": {"action": "spawn_agent", "config": {"role": "planner", "agent_type": "site-planner"}, "next_step": "spawn_content_writer", "description": "Spawn site planner agent", "output_field": "planner_agent"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "ensure_site_record", "description": "Spawn deployer agent", "output_field": "deployer_agent"}, "spawn_reviewer": {"action": "spawn_agent", "config": {"role": "reviewer", "agent_type": "content-reviewer"}, "next_step": "spawn_deployer", "description": "Spawn content reviewer agent", "output_field": "reviewer_agent"}, "store_site_plan": {"action": "update_site_content", "config": {"merge": true, "content_field": "site_plan", "site_id_field": "site_record.site_id"}, "next_step": "sync_pages_to_db", "description": "Store the site plan in sites.content_data", "output_field": "content_stored"}, "build_pages_loop": {"action": "loop", "config": {"mode": "sequential", "items_field": "pages_to_build.pages", "sub_workflow": {"steps": {"deploy_page": {"action": "git_commit", "config": {"page_field": "current_page", "domain_field": "site_record.domain", "content_field": "assembled_page.html"}, "next_step": "save_sections", "description": "Commit page to git", "output_field": "page_deployed"}, "assemble_page": {"action": "assemble_page", "config": {"content_field": "page_content.response.page_html", "add_navigation": false}, "next_step": "deploy_page", "description": "Assemble full page HTML from components", "output_field": "assembled_page"}, "complete_page": {"action": "loop_complete", "description": "Page build complete"}, "save_sections": {"action": "save_page_sections", "config": {"html_field": "assembled_page.html", "site_id_field": "site_record.site_id", "page_name_field": "current_page.name"}, "next_step": "update_page_status", "description": "Save rendered sections to page_components for rerender", "output_field": "save_result"}, "update_page_status": {"action": "update_page_status", "config": {"status": "deployed", "commit_from": "page_deployed.commit_sha", "page_id_field": "current_page.id"}, "next_step": "complete_page", "description": "Mark page as deployed in database"}, "write_page_content": {"action": "call_agent", "config": {"agent_type": "page-content-writer", "target_role": "content_writer", "input_mapping": {"db_sync": "db_sync", "hero_url?": "hero_url", "logo_url?": "logo_url", "site_record": "site_record", "current_page": "current_page", "reviewed_brief": "input_data.reviewed_brief", "brand_logo_url?": "brand_logo_url", "style_collection": "style_collection"}, "timeout_seconds": 300}, "next_step": "review_page_content", "description": "Write content for this page", "output_field": "page_content"}, "review_page_content": {"action": "call_agent", "config": {"agent_type": "content-reviewer", "target_role": "reviewer", "input_mapping": {"site_record": "site_record", "current_page": "current_page", "page_content": "page_content", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 3900}, "next_step": "check_review_approved", "description": "Review page content (HITL or auto-eval)", "output_field": "reviewed_content"}, "check_review_approved": {"action": "conditional", "config": {"condition": "reviewed_content.review_result.approved == true OR reviewed_content.approved == true", "else_step": "complete_page", "then_step": "assemble_page"}, "description": "Check if content was approved - skip deploy if rejected"}}, "start_step": "write_page_content"}, "item_variable": "current_page", "max_iterations": 20}, "next_step": "apply_site_design", "description": "Build each page: write → review → deploy", "output_field": "pages_built"}, "store_hero_asset": {"action": "store_asset", "config": {"purpose": "hero", "asset_type": "image", "data_field": "hero_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "hero_images.home", "origin_prompt_field": "site_plan.image_prompts.hero_home", "update_site_brand_assets": true}, "next_step": "deploy_hero_image", "description": "Store generated hero image", "output_field": "hero_stored"}, "store_logo_asset": {"action": "store_asset", "config": {"purpose": "brand_logo", "asset_type": "logo", "data_field": "logo_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "logo.primary", "origin_prompt_field": "site_plan.image_prompts.logo", "update_site_brand_assets": true}, "next_step": "check_hero_images", "description": "Store generated logo in assets table and site brand_assets", "output_field": "logo_stored"}, "sync_pages_to_db": {"action": "sync_pages_to_db", "config": {"input_fields": ["site_record", "site_plan"]}, "next_step": "check_assets_needed", "description": "Create page records from site plan", "output_field": "db_sync"}, "apply_site_design": {"action": "call_agent", "config": {"agent_type": "webdesign-agent", "target_role": "webdesigner", "input_mapping": {"domain": "site_record.domain", "site_id": "site_record.site_id"}, "timeout_seconds": 300}, "next_step": "trigger_site_deploy", "description": "Generate and deploy site stylesheet", "output_field": "design_result"}, "call_site_planner": {"action": "call_agent", "config": {"agent_type": "site-planner", "target_role": "planner", "input_mapping": {"input_data": "input_data", "site_record": "site_record", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 120}, "next_step": "store_reviewed_brief", "description": "Plan pages, select components, identify asset needs", "output_field": "site_plan"}, "check_hero_images": {"action": "conditional", "config": {"condition": "site_plan.needs_images == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "generate_hero_image"}, "description": "Check if hero images need to be generated"}, "deploy_hero_image": {"action": "deploy_image_asset", "config": {"purpose": "hero", "uri_field": "hero_result.image_uri", "domain_field": "site_record.domain", "output_mapping": {"files": "response.data.files", "domain": "response.data.domain", "deployed": "response.data.success", "repo_url": "response.data.repo_url", "image_url": "response.data.file_path"}}, "next_step": "select_style_collection", "description": "Download, optimize and deploy hero image to git", "output_field": "hero_deployed"}, "ensure_site_record": {"action": "ensure_site_record", "config": {"store_brief_in_content_data": true}, "next_step": "call_site_planner", "description": "Create or update site record in database", "output_field": "site_record"}, "get_pages_to_build": {"action": "get_pages_to_build", "config": {"include_all": false, "build_statuses": ["planned", "needs_rebuild"]}, "next_step": "build_pages_loop", "description": "Query pages from database that need content generation", "output_field": "pages_to_build"}, "update_site_status": {"action": "update_site_status", "config": {"status": "deployed", "deployed_at": "now", "site_id_field": "site_record.site_id"}, "next_step": "complete", "description": "Mark site as deployed", "output_field": "site_updated"}, "check_assets_needed": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true OR site_plan.needs_images == true OR site_plan.response.needs_logo == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "spawn_image_generator"}, "description": "Check if logo or images need to be generated"}, "generate_hero_image": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.hero_home", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_hero_asset", "description": "Generate hero image for home page", "output_field": "hero_result"}, "trigger_site_deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_mapping": {"pages_built": "pages_built", "site_record": "site_record"}, "timeout_seconds": 180}, "next_step": "update_site_status", "description": "Trigger Cloudflare deployment", "output_field": "deployment_result"}, "call_logo_generation": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.logo", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_logo_asset", "description": "Generate logo using image-generator agent", "output_field": "logo_result"}, "spawn_content_writer": {"action": "spawn_agent", "config": {"role": "content_writer", "agent_type": "page-content-writer"}, "next_step": "spawn_reviewer", "description": "Spawn content writer agent", "output_field": "content_writer_agent"}, "store_reviewed_brief": {"action": "update_site_content", "config": {"merge": true, "content_field": "input_data.reviewed_brief", "site_id_field": "site_record.site_id"}, "next_step": "store_site_plan", "description": "Store the reviewed brief in sites.content_data", "output_field": "brief_stored"}, "spawn_image_generator": {"action": "spawn_agent", "config": {"role": "image_generator", "agent_type": "image-generator"}, "next_step": "spawn_webdesign_agent", "description": "Spawn image generator agent for asset creation", "output_field": "image_generator_info"}, "spawn_webdesign_agent": {"action": "spawn_agent", "config": {"role": "webdesigner", "agent_type": "webdesign-agent"}, "next_step": "generate_logo", "description": "Spawn webdesign agent", "output_field": "webdesign_agent"}, "set_default_components": {"action": "update_site_defaults", "config": {"defaults": {"head": "head-seo-standard", "footer_from": "style_collection.footer_component_name", "header_from": "style_collection.header_component_name"}, "site_id_field": "site_record.site_id"}, "next_step": "get_pages_to_build", "description": "Set default head/header/footer components", "output_field": "defaults_set"}, "select_style_collection": {"action": "select_style_collection", "config": {"style_from": "site_plan.style_collection", "site_id_field": "site_record.site_id", "fallback_by_domain": true}, "next_step": "set_default_components", "description": "Choose style collection based on site plan", "output_field": "style_collection"}}, "start_step": "spawn_planner"}, "processing_mode": "orchestrator", "timeout_seconds": 900} | t         | 2025-12-22 17:46:51.419068+00 | 2026-02-02 14:17:17.129046+00 |            | ["orchestration", "website-builder", "component-based", "image-generation"] | docker.io/aqls/agent-chassis | v1.0.742  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |      20 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | ["website", "multipage", "component-based"] | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]} |           0 | f           | {"expects": {"reviewed_brief": "object - completed questionnaire answers", "input_data.domain": "string - the domain name", "input_data.objective": "string - what the site should achieve"}, "required": ["input_data", "reviewed_brief"]} | {"produces": {"site_id": "uuid - the site record ID", "deploy_url": "string - the live site URL", "pages_built": "number - count of pages deployed"}}
(1 row)

// ===========================================================================
// WORKFLOW SQL UPDATE — add inject_head to assemble_page step config
                            // ===========================================================================
                            // The pageflow-builder s build_pages_loop sub_workflow has:
//
//   "assemble_page": {
//       "action": "assemble_page",
//       "config": {
//           "content_field": "page_content.response.page_html",
//           "add_navigation": false
//       },
//       ...
//   }
//
// Add "inject_head": true to the config.
// Note: NOT adding inject_header/inject_footer here because the
// page-content-writer already includes header/footer in its output.
// ===========================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,inject_head}',
        'true'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- backup
id                     | 427aa3e5-5ea2-4917-8d24-d751ebd283b2
type                   | pageflow-builder
display_name           | PageFlow Builder
description            | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time.
category               | orchestrator
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["site_record", "pages_built", "deployment_result"]}, "description": "Site build complete"}, "generate_logo": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true", "else_step": "check_hero_images", "then_step": "call_logo_generation"}, "description": "Check if logo needs to be generated"}, "spawn_planner": {"action": "spawn_agent", "config": {"role": "planner", "agent_type": "site-planner"}, "next_step": "spawn_content_writer", "description": "Spawn site planner agent", "output_field": "planner_agent"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "ensure_site_record", "description": "Spawn deployer agent", "output_field": "deployer_agent"}, "spawn_reviewer": {"action": "spawn_agent", "config": {"role": "reviewer", "agent_type": "content-reviewer"}, "next_step": "spawn_deployer", "description": "Spawn content reviewer agent", "output_field": "reviewer_agent"}, "store_site_plan": {"action": "update_site_content", "config": {"merge": true, "content_field": "site_plan", "site_id_field": "site_record.site_id"}, "next_step": "sync_pages_to_db", "description": "Store the site plan in sites.content_data", "output_field": "content_stored"}, "build_pages_loop": {"action": "loop", "config": {"mode": "sequential", "items_field": "pages_to_build.pages", "sub_workflow": {"steps": {"deploy_page": {"action": "git_commit", "config": {"page_field": "current_page", "domain_field": "site_record.domain", "content_field": "assembled_page.html"}, "next_step": "save_sections", "description": "Commit page to git", "output_field": "page_deployed"}, "assemble_page": {"action": "assemble_page", "config": {"inject_head": true, "content_field": "page_content.response.page_html", "add_navigation": false}, "next_step": "deploy_page", "description": "Assemble full page HTML from components", "output_field": "assembled_page"}, "complete_page": {"action": "loop_complete", "description": "Page build complete"}, "save_sections": {"action": "save_page_sections", "config": {"html_field": "assembled_page.html", "site_id_field": "site_record.site_id", "page_name_field": "current_page.name"}, "next_step": "update_page_status", "description": "Save rendered sections to page_components for rerender", "output_field": "save_result"}, "update_page_status": {"action": "update_page_status", "config": {"status": "deployed", "commit_from": "page_deployed.commit_sha", "page_id_field": "current_page.id"}, "next_step": "complete_page", "description": "Mark page as deployed in database"}, "write_page_content": {"action": "call_agent", "config": {"agent_type": "page-content-writer", "target_role": "content_writer", "input_mapping": {"db_sync": "db_sync", "hero_url?": "hero_url", "logo_url?": "logo_url", "site_plan": "site_plan", "site_record": "site_record", "current_page": "current_page", "reviewed_brief": "input_data.reviewed_brief", "brand_logo_url?": "brand_logo_url", "style_collection": "style_collection"}, "timeout_seconds": 300}, "next_step": "review_page_content", "description": "Write content for this page", "output_field": "page_content"}, "review_page_content": {"action": "call_agent", "config": {"agent_type": "content-reviewer", "target_role": "reviewer", "input_mapping": {"site_record": "site_record", "current_page": "current_page", "page_content": "page_content", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 3900}, "next_step": "check_review_approved", "description": "Review page content (HITL or auto-eval)", "output_field": "reviewed_content"}, "check_review_approved": {"action": "conditional", "config": {"condition": "reviewed_content.review_result.approved == true OR reviewed_content.approved == true", "else_step": "complete_page", "then_step": "assemble_page"}, "description": "Check if content was approved - skip deploy if rejected"}}, "start_step": "write_page_content"}, "item_variable": "current_page", "max_iterations": 20}, "next_step": "apply_site_design", "description": "Build each page: write → review → deploy", "output_field": "pages_built"}, "store_hero_asset": {"action": "store_asset", "config": {"purpose": "hero", "asset_type": "image", "data_field": "hero_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "hero_images.home", "origin_prompt_field": "site_plan.image_prompts.hero_home", "update_site_brand_assets": true}, "next_step": "deploy_hero_image", "description": "Store generated hero image", "output_field": "hero_stored"}, "store_logo_asset": {"action": "store_asset", "config": {"purpose": "brand_logo", "asset_type": "logo", "data_field": "logo_result.image_url", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "logo.primary", "origin_prompt_field": "site_plan.image_prompts.logo", "update_site_brand_assets": true}, "next_step": "check_hero_images", "description": "Store generated logo in assets table and site brand_assets", "output_field": "logo_stored"}, "sync_pages_to_db": {"action": "sync_pages_to_db", "config": {"input_fields": ["site_record", "site_plan"]}, "next_step": "check_assets_needed", "description": "Create page records from site plan", "output_field": "db_sync"}, "apply_site_design": {"action": "call_agent", "config": {"agent_type": "webdesign-agent", "target_role": "webdesigner", "input_mapping": {"domain": "site_record.domain", "site_id": "site_record.site_id"}, "timeout_seconds": 300}, "next_step": "trigger_site_deploy", "description": "Generate and deploy site stylesheet", "output_field": "design_result"}, "call_site_planner": {"action": "call_agent", "config": {"agent_type": "site-planner", "target_role": "planner", "input_mapping": {"input_data": "input_data", "site_record": "site_record", "reviewed_brief": "input_data.reviewed_brief"}, "timeout_seconds": 120}, "next_step": "store_reviewed_brief", "description": "Plan pages, select components, identify asset needs", "output_field": "site_plan"}, "check_hero_images": {"action": "conditional", "config": {"condition": "site_plan.needs_images == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "generate_hero_image"}, "description": "Check if hero images need to be generated"}, "deploy_hero_image": {"action": "deploy_image_asset", "config": {"purpose": "hero", "uri_field": "hero_result.image_uri", "domain_field": "site_record.domain", "output_mapping": {"files": "response.data.files", "domain": "response.data.domain", "deployed": "response.data.success", "repo_url": "response.data.repo_url", "image_url": "response.data.file_path"}}, "next_step": "select_style_collection", "description": "Download, optimize and deploy hero image to git", "output_field": "hero_deployed"}, "ensure_site_record": {"action": "ensure_site_record", "config": {"store_brief_in_content_data": true}, "next_step": "call_site_planner", "description": "Create or update site record in database", "output_field": "site_record"}, "get_pages_to_build": {"action": "get_pages_to_build", "config": {"include_all": false, "build_statuses": ["planned", "needs_rebuild"]}, "next_step": "build_pages_loop", "description": "Query pages from database that need content generation", "output_field": "pages_to_build"}, "update_site_status": {"action": "update_site_status", "config": {"status": "deployed", "deployed_at": "now", "site_id_field": "site_record.site_id"}, "next_step": "complete", "description": "Mark site as deployed", "output_field": "site_updated"}, "check_assets_needed": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true OR site_plan.needs_images == true OR site_plan.response.needs_logo == true OR site_plan.response.needs_images == true", "else_step": "select_style_collection", "then_step": "spawn_image_generator"}, "description": "Check if logo or images need to be generated"}, "generate_hero_image": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.hero_home", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_hero_asset", "description": "Generate hero image for home page", "output_field": "hero_result"}, "trigger_site_deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_mapping": {"pages_built": "pages_built", "site_record": "site_record"}, "timeout_seconds": 180}, "next_step": "update_site_status", "description": "Trigger Cloudflare deployment", "output_field": "deployment_result"}, "call_logo_generation": {"action": "call_agent", "config": {"agent_type": "image-generator", "target_role": "image_generator", "input_mapping": {"prompt": "site_plan.response.image_prompts.logo", "site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}, "output_mapping": {"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}, "timeout_seconds": 120}, "next_step": "store_logo_asset", "description": "Generate logo using image-generator agent", "output_field": "logo_result"}, "spawn_content_writer": {"action": "spawn_agent", "config": {"role": "content_writer", "agent_type": "page-content-writer"}, "next_step": "spawn_reviewer", "description": "Spawn content writer agent", "output_field": "content_writer_agent"}, "store_reviewed_brief": {"action": "update_site_content", "config": {"merge": true, "content_field": "input_data.reviewed_brief", "site_id_field": "site_record.site_id"}, "next_step": "store_site_plan", "description": "Store the reviewed brief in sites.content_data", "output_field": "brief_stored"}, "spawn_image_generator": {"action": "spawn_agent", "config": {"role": "image_generator", "agent_type": "image-generator"}, "next_step": "spawn_webdesign_agent", "description": "Spawn image generator agent for asset creation", "output_field": "image_generator_info"}, "spawn_webdesign_agent": {"action": "spawn_agent", "config": {"role": "webdesigner", "agent_type": "webdesign-agent"}, "next_step": "generate_logo", "description": "Spawn webdesign agent", "output_field": "webdesign_agent"}, "set_default_components": {"action": "update_site_defaults", "config": {"defaults": {"head": "head-seo-standard", "footer_from": "style_collection.footer_component_name", "header_from": "style_collection.header_component_name"}, "site_id_field": "site_record.site_id"}, "next_step": "get_pages_to_build", "description": "Set default head/header/footer components", "output_field": "defaults_set"}, "select_style_collection": {"action": "select_style_collection", "config": {"style_from": "site_plan.style_collection", "site_id_field": "site_record.site_id", "fallback_by_domain": true}, "next_step": "set_default_components", "description": "Choose style collection based on site plan", "output_field": "style_collection"}}, "start_step": "spawn_planner"}, "processing_mode": "orchestrator", "timeout_seconds": 900}
is_active              | t
created_at             | 2025-12-22 17:46:51.419068+00
updated_at             | 2026-02-04 10:15:48.506128+00
deleted_at             |
capabilities           | ["orchestration", "website-builder", "component-based", "image-generation"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.746
command                |
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 20
previous_version_id    |
task_workflow          |
orchestrator_workflow  |
orchestration_workflow |
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         |
status                 | active
domain_tags            | ["website", "multipage", "component-based"]
briefing_questionnaire | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"reviewed_brief": "object - completed questionnaire answers", "input_data.domain": "string - the domain name", "input_data.objective": "string - what the site should achieve"}, "required": ["input_data", "reviewed_brief"]}
output_contract        | {"produces": {"site_id": "uuid - the site record ID", "deploy_url": "string - the live site URL", "pages_built": "number - count of pages deployed"}}



-- 1a: Redirect set_default_components to render_site_components
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,set_default_components,next_step}',
        '"render_site_components"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- 1b: Add the render_site_components step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,render_site_components}',
        '{
            "action": "render_site_components",
            "config": {
                "slots": ["header", "footer", "head"],
                "force_rerender": false
            },
            "next_step": "get_pages_to_build",
            "description": "Render and store site-level components for rerender",
            "output_field": "site_components_rendered"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';


====

-- add logo image changes

-- ============================================================================
-- Logo Pipeline Fix: Workflow + Template Updates
-- ============================================================================
-- Issue 1: Add deploy_logo_image step (logo was stored in S3 but never committed to git)
-- Issue 2: Fix store_logo_asset.config.purpose from "brand_logo" to "logo"
--          ("brand_logo" isn't in ImagePurposes, fell to default 1200x800 jpg instead of 400x400 png)
-- Issue 5: Update header templates to show logo image when available
-- ============================================================================


-- ============================================================================
-- FIX 1+2: Pageflow-builder workflow — add deploy_logo_image, fix purpose
-- ============================================================================

-- Step A: Fix store_logo_asset — change purpose to "logo" and next_step to "deploy_logo_image"
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset}',
        '{
          "action": "store_asset",
          "config": {
            "purpose": "logo",
            "asset_type": "logo",
            "data_field": "logo_result.image_url",
            "origin_type": "generated",
            "site_id_field": "site_record.site_id",
            "brand_asset_key": "logo.primary",
            "origin_prompt_field": "site_plan.image_prompts.logo",
            "update_site_brand_assets": true
          },
          "next_step": "deploy_logo_image",
          "description": "Store generated logo in assets table and site brand_assets",
          "output_field": "logo_stored"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Step B: Add deploy_logo_image step (modelled on existing deploy_hero_image)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_logo_image}',
        '{
          "action": "deploy_image_asset",
          "config": {
            "purpose": "logo",
            "uri_field": "logo_result.image_uri",
            "domain_field": "site_record.domain"
          },
          "next_step": "check_hero_images",
          "description": "Download, optimize and deploy logo image to git",
          "output_field": "logo_deployed"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';



-- ============================================================================
-- Verification queries
-- ============================================================================

-- Check workflow has deploy_logo_image step
SELECT
    type,
    default_config->'workflow'->'steps'->'store_logo_asset'->'config'->>'purpose' AS store_purpose,
    default_config->'workflow'->'steps'->'store_logo_asset'->>'next_step' AS store_next,
    default_config->'workflow'->'steps'->'deploy_logo_image'->>'action' AS deploy_action,
    default_config->'workflow'->'steps'->'deploy_logo_image'->'config'->>'purpose' AS deploy_purpose,
    default_config->'workflow'->'steps'->'deploy_logo_image'->>'next_step' AS deploy_next,
    default_config->'workflow'->'steps'->'deploy_logo_image'->>'output_field' AS deploy_output
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Check header templates have logo_url support
SELECT name,
       html_template LIKE '%logo_url%' AS has_logo_url,
       html_template LIKE '%.logo-img%' AS has_logo_img_css
FROM content_components
WHERE name IN ('header-professional-dark', 'header-minimal-light', 'header-bold-gradient');

---


-- Fix render_site_components step in pageflow-builder to use site_id_field
-- The action was looking for site_id at root level but it's at site_record.site_id

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,render_site_components,config}',
        '{
            "slots": ["header", "footer", "head"],
            "force_rerender": false,
            "site_id_field": "site_record.site_id",
            "domain_field": "site_record.domain"
        }'::jsonb
             )
WHERE type = 'pageflow-builder';


---- - recognise hero image prompt
-- Fix 5: Update pageflow-builder workflow to use input_mapping instead of template syntax for image prompts
-- The old "prompt": "{{site_plan.image_prompts.hero_home}}" doesn't resolve
-- Need to use input_mapping with dot-path: "prompt": "site_plan.response.image_prompts.hero_home"
--
-- NOTE: This requires careful update of the workflow JSON. Run this query to find the agent:
-- SELECT agent_type, config FROM agent_definitions WHERE agent_type = 'pageflow-builder';
--
-- Then update the workflow.steps.generate_hero_image and workflow.steps.call_logo_generation
-- to use input_mapping instead of prompt field.
--
-- The fix should change from:
--   "generate_hero_image": {
--     "config": {
--       "prompt": "{{site_plan.image_prompts.hero_home}}",
--       "input_fields": ["site_plan", "reviewed_brief"]
--     }
--   }
--
-- To:
--   "generate_hero_image": {
--     "config": {
--       "input_mapping": {
--         "prompt": "site_plan.response.image_prompts.hero_home",
--         "site_plan": "site_plan",
--         "reviewed_brief": "input_data.reviewed_brief"
--       }
--     }
--   }
--
-- Similar change needed for call_logo_generation step.
--
-- This is a jsonb update, so it needs to be done carefully. Here's the update:

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,generate_hero_image,config}',
                '{
                    "agent_type": "image-generator",
                    "target_role": "image_generator",
                    "input_mapping": {
                        "prompt": "site_plan.response.image_prompts.hero_home",
                        "site_plan": "site_plan",
                        "reviewed_brief": "input_data.reviewed_brief"
                    },
                    "output_mapping": {
                        "prompt": "generate.response.prompt",
                        "image_uri": "generate.response.image_uri",
                        "image_url": "generate.response.image_url",
                        "generated_at": "generate.response.generated_at"
                    },
                    "timeout_seconds": 120
                }'::jsonb
        ),
        '{workflow,steps,call_logo_generation,config}',
        '{
            "agent_type": "image-generator",
            "target_role": "image_generator",
            "input_mapping": {
                "prompt": "site_plan.response.image_prompts.logo",
                "site_plan": "site_plan",
                "reviewed_brief": "input_data.reviewed_brief"
            },
            "output_mapping": {
                "prompt": "generate.response.prompt",
                "image_uri": "generate.response.image_uri",
                "image_url": "generate.response.image_url",
                "generated_at": "generate.response.generated_at"
            },
            "timeout_seconds": 120
        }'::jsonb
             ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- nav changes
-- Workflow change: Insert populate_nav_tables step into pageflow-builder
-- and multipage-website-builder workflows.
--
-- CURRENT step chains:
--   pageflow-builder:           sync_pages_to_db → check_assets_needed → ...
--   multipage-website-builder:  sync_pages_to_db → generate_pages_loop → ...
--
-- NEW step chains:
--   pageflow-builder:           sync_pages_to_db → populate_nav → check_assets_needed → ...
--   multipage-website-builder:  sync_pages_to_db → populate_nav → generate_pages_loop → ...

-- =========================================================================
-- 1. pageflow-builder: change sync_pages_to_db.next_step and add populate_nav
-- =========================================================================

-- Change sync_pages_to_db next_step from check_assets_needed to populate_nav
UPDATE agent_definitions
SET workflow = jsonb_set(
        workflow,
        '{workflow,steps,sync_pages_to_db,next_step}',
        '"populate_nav"'
               )
WHERE type = 'pageflow-builder' AND is_active = true;

-- Add the populate_nav step
UPDATE agent_definitions
SET workflow = jsonb_set(
        workflow,
        '{workflow,steps,populate_nav}',
        '{
            "action": "populate_nav_tables",
            "config": {
                "site_id_field": "site_record.site_id",
                "max_header_items": 6
            },
            "next_step": "check_assets_needed",
            "description": "Populate navigation tables from page plan",
            "output_field": "nav_data"
        }'::jsonb
               )
WHERE type = 'pageflow-builder' AND is_active = true;

--
-- nav changes
-- Update max_header_items from 6 to 8 in deployed workflows
-- Run this on the live DB since the original 002 SQL already ran with 6.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,populate_nav,config,max_header_items}',
        '8'
                     )
WHERE type = 'pageflow-builder' AND is_active = true;

