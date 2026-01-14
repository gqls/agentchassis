changed from multipage-website-builder version 3


UPDATE agent_definitions
SET
    type = 'pageflow-builder',
    display_name = 'PageFlow Builder',
    description = 'Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time.',
    version = 1,  -- Reset to v1 since it's a new type
    updated_at = NOW()
WHERE type = 'multipage-website-builder'
  AND version = 3;

-- Step 2: Verify the rename worked
SELECT
    id,
    type,
    version,
    display_name,
    description,
    is_active,
    status
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Step 3: Check that intake-orchestrator will find it (matches %-builder pattern)
SELECT
    type,
    display_name,
    description
FROM agent_definitions
WHERE type LIKE '%-builder'
  AND is_active = true
ORDER BY type;

-- Step 4: Verify there's no conflict with old multipage-website-builder entries
SELECT
    type,
    version,
    display_name,
    is_active
FROM agent_definitions
WHERE type LIKE '%multipage%' OR type LIKE '%pageflow%'
ORDER BY type, version;

---

-- ============================================================================
-- WORKFLOW UPDATE: Add get_pages_to_build step to pageflow-builder
-- Database: clients_db (agent_definitions table)
-- NOTE: Uses 'type' column (not 'agent_type')
-- ============================================================================

BEGIN;

-- 1. Update set_default_components to point to new step instead of build_pages_loop
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,set_default_components,next_step}',
        '"get_pages_to_build"'
                     )
WHERE type = 'pageflow-builder';

-- 2. Add the new get_pages_to_build step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,get_pages_to_build}',
        '{
            "action": "get_pages_to_build",
            "description": "Query pages from database that need content generation",
            "config": {
                "build_statuses": ["planned", "needs_rebuild"],
                "include_all": false
            },
            "output_field": "pages_to_build",
            "next_step": "build_pages_loop"
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- 3. Update build_pages_loop to use pages from the new step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,items_field}',
        '"pages_to_build.pages"'
                     )
WHERE type = 'pageflow-builder';

COMMIT;

-- Verify the changes
SELECT
    type,
    default_config->'workflow'->'steps'->'set_default_components'->>'next_step' as set_defaults_next,
    default_config->'workflow'->'steps'->'get_pages_to_build'->>'action' as new_step_action,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->>'items_field' as loop_items_field
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- ============================================================================
-- Fix pageflow-builder to pass reviewed_brief to page-content-writer
-- Database: clients_db
-- ============================================================================
--
-- The issue: build_pages_loop substep write_page_content calls page-content-writer
-- with input_fields: ["current_page", "site_record", "style_collection"]
-- but it's missing "reviewed_brief" which is needed for content generation
-- ============================================================================

BEGIN;

-- Update the write_page_content substep to include reviewed_brief
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,input_fields}',
        '["current_page", "site_record", "reviewed_brief", "style_collection"]'::jsonb
                     )
WHERE type = 'pageflow-builder';

COMMIT;

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config'->'input_fields' as input_fields
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Fix for pageflow-builder timeout when calling content-reviewer
--
-- Problem: The review_page_content step has timeout_seconds: 300 (5 minutes)
--          but content-reviewer's HITL step needs up to 3600 seconds (1 hour)
--          The parent times out before the child can complete HITL
--
-- Solution: Increase the call_agent timeout to be longer than the child's HITL timeout

-- First, let's see the current configuration
-- SELECT agent_type, config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'review_page_content'
-- FROM agent_definitions WHERE agent_type = 'pageflow-builder';

-- Update the timeout for review_page_content step from 300 to 3900 seconds (65 minutes)
-- This gives 5 extra minutes buffer over the HITL's 1-hour timeout

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,review_page_content,config,timeout_seconds}',
        '3900'::jsonb
             )
WHERE type = 'pageflow-builder';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'review_page_content'->'config'->'timeout_seconds' as review_timeout
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Also update the write_page_content step if needed (currently 120 seconds)
-- This may need to be longer if research-agent is involved
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,timeout_seconds}',
    '300'::jsonb
)
WHERE type = 'pageflow-builder';

--

Note: I also changed the content source from reviewed_content to page_content because:

reviewed_content is the review result (approved/issues/etc.)
page_content is the actual page HTML/sections from page-content-writer

-- Fix assemble_page config: use "content_field" instead of "content_from"
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config}',
        '{
          "content_field": "page_content",
          "add_navigation": false
        }'::jsonb
             )
WHERE type = 'pageflow-builder';

--

-- FILE: fix_pageflow_assemble_page.sql
-- Fix assemble_page to use correct content path
--
-- Issue: Config had content_field: "page_content"
--        But actual HTML is at page_content.page_html
--
-- The page-content-writer returns:
--   { "page_html": "<!DOCTYPE html>..." }
--
-- This is stored at collected_data["page_content"]
-- So the full path to HTML is: page_content.page_html
--
-- Note: The loop's propagateIterationOutputs should copy page_content_0 → page_content
-- before assemble_page runs, so we don't need the _0 suffix in the path.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,content_field}',
        '"page_content.page_html"'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'assemble_page'->'config' as assemble_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- ==============================================================================
-- ADDITIONAL: Add reviewed_content to propagateIterationOutputs list
-- ==============================================================================
-- In platform/orchestration/loop_actions.go, around line 38551, add to commonOutputFields:
--
-- commonOutputFields := []string{
--     "page_content",
--     "page_html",
--     "content_result",
--     "html_result",
--     "site_content",
--     "site_architecture",
--     "reviewed_content",    // ADD THIS
--     "assembled_page",      // ADD THIS
--     "page_deployed",       // ADD THIS
-- }
--
-- This ensures all substep outputs in the build_pages_loop are propagated.

==

fix for deployer agent too:

-- FILE: fix_git_commit_config.sql
-- Alternative fix: Change workflow config to use supported fields
--
-- Current config uses:
--   "html_from": "assembled_page.html",
--   "page_from": "current_page"
--
-- But extractFilesForGit supports:
--   - files_field: path to map of filename -> content
--   - content_field: path to single HTML string (saves as index.html)
--
-- For single page commits in a loop, we need a different approach.
-- The actual structure is:
--   assembled_page: { html: "...", assembled_at: "..." }
--   current_page: { name: "index", url: "/index.html", ... }
--
-- OPTION A: Use content_field (but always saves as index.html - not ideal for multi-page)
-- OPTION B: Apply the code fix to support html_from + page_from pattern (recommended)

-- If you want a quick workaround using content_field (works but always uses index.html):
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,deploy_page,config}',
        '{
          "content_field": "assembled_page.html",
          "site_id_from": "site_record.site_id"
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- NOTE: The content_field approach saves as "index.html" which works for index page
-- but won't work correctly for other pages (services.html, about.html, etc.)
-- The recommended fix is to apply fix_extract_files_for_git.go instead.

-- Verify
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'deploy_page'->'config' as deploy_config
FROM agent_definitions
WHERE type = 'pageflow-builder';

==

-- FILE: fix_pageflow_assemble_page.sql
-- Fix pageflow-builder config field mismatches
--
-- ISSUES IDENTIFIED:
--
-- 1. assemble_page step:
--    - Config uses "content_from" but AssemblePageAction expects "content_field"
--    - Config uses "page_from", "site_id_from" - not used by action
--
-- 2. deploy_page (git_commit) step:
--    - Config uses "html_from" but GitCommitAction expects "content_field"
--    - Config uses "page_from", "site_id_from" - not used by action
--    - GitCommitAction extracts domain via "domain_field" (defaults to "domain")
--
-- Data flow for page content:
--   page-content-writer outputs: {page_html, page_name, sections, research_ids}
--   complete_workflow wraps in collectedData
--   pageflow-builder stores at output_field "page_content"
--   Path to HTML: page_content.page_content.page_html
--
-- After assemble_page runs:
--   assembled_page = {html: "<full page>", assembled_at: "timestamp"}
--   Path to HTML: assembled_page.html

-- ==============================================================================
-- FIX 1: assemble_page config
-- ==============================================================================
-- AssemblePageAction expects:
--   content_field: string - path to HTML content
--   add_navigation: bool - whether to add nav (optional)

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config}',
        '{
            "content_field": "page_content.page_content.page_html",
            "add_navigation": false
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- ==============================================================================
-- FIX 2: deploy_page (git_commit) config
-- ==============================================================================
-- GitCommitAction expects:
--   content_field: string - path to HTML (creates single file as index.html or {page}.html)
--   domain_field: string - path to domain (defaults to "domain")
--   page_field: string - path to current page (to get filename)
--   files_field: string - path to files map (for multipage)
--
-- For per-page deployment, we need content_field and page context

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,deploy_page,config}',
        '{
            "content_field": "assembled_page.html",
            "page_field": "current_page",
            "domain_field": "site_record.domain"
        }'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- ==============================================================================
-- FIX 3: Verify all sub_workflow steps have expected fields
-- ==============================================================================

-- Check write_page_content (call_agent)
-- Expects: agent_type, target_role, input_fields, timeout_seconds
-- Current config looks correct - no changes needed

-- Check review_page_content (call_agent)
-- Current config looks correct - no changes needed

-- Check update_page_status
-- Needs to check what UpdatePageStatusAction expects

-- ==============================================================================
-- VERIFY FIXES
-- ==============================================================================

SELECT
    type,
    jsonb_pretty(
            default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'
    ) as sub_workflow_steps
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- ==============================================================================
-- ADDITIONAL Go Code Changes (if needed)
-- ==============================================================================
--
-- If GitCommitAction doesn't handle per-page filenames, update extractFilesForGit:
--
-- Add handling for page_field to determine filename:
--   pageField, _ := config["page_field"].(string)
--   if pageField != "" {
--     pageData := extractFieldValue(data, pageField, logger)
--     if pageName, ok := pageData["name"].(string); ok {
--       filename = pageName + ".html"
--     }
--   }
--
-- This would create about.html, contact.html, etc. instead of always index.html
-- ==============================================================================


-- change path for page_content.html

-- Update the assemble_page content_field path in pageflow-builder workflow
-- Changes: page_content.page_content.page_html -> page_content.page_html

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,content_field}',
        '"page_content.page_html"'::jsonb
             )
WHERE type = 'pageflow-builder';

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'assemble_page'->'config'->'content_field' as content_field
FROM agent_definitions
WHERE type = 'pageflow-builder';

==

-- path fix

-- Check current path
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'assemble_page'->'config'->'content_field' as content_field
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Update to include .response
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,content_field}',
        '"page_content.response.page_html"'::jsonb
                     )
WHERE type = 'pageflow-builder';

---
-- change to page_id_field
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config}',
        '{"status": "deployed", "commit_from": "page_deployed.commit_sha", "page_id_field": "current_page.id"}'
                     )
WHERE type = 'pageflow-builder';

---

increase timeout

--
pageflow-builder before changes:
         427aa3e5-5ea2-4917-8d24-d751ebd283b2 | pageflow-builder | PageFlow Builder | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["site_record", "pages_built", "deployment_result"]}, "description": "Site build complete"}, "generate_logo": {"action": "conditional", "config": {"condition": "site_plan.needs_logo == true", "else_step": "check_hero_images", "then_step": "call_logo_generation"}, "description": "Check if logo needs to be generated"}, "spawn_planner": {"action": "spawn_agent", "config": {"role": "planner", "agent_type": "site-planner"}, "next_step": "spawn_content_writer", "description": "Spawn site planner agent", "output_field": "planner_agent"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "ensure_site_record", "description": "Spawn deployer agent", "output_field": "deployer_agent"}, "spawn_reviewer": {"action": "spawn_agent", "config": {"role": "reviewer", "agent_type": "content-reviewer"}, "next_step": "spawn_deployer", "description": "Spawn content reviewer agent", "output_field": "reviewer_agent"}, "store_site_plan": {"action": "update_site_content", "config": {"merge": true, "content_field": "site_plan", "site_id_field": "site_record.site_id"}, "next_step": "sync_pages_to_db", "description": "Store the site plan in sites.content_data", "output_field": "content_stored"}, "build_pages_loop": {"action": "loop", "config": {"mode": "sequential", "items_field": "pages_to_build.pages", "sub_workflow": {"steps": {"deploy_page": {"action": "git_commit", "config": {"page_field": "current_page", "domain_field": "site_record.domain", "content_field": "assembled_page.html"}, "next_step": "update_page_status", "description": "Commit page to git", "output_field": "page_deployed"}, "assemble_page": {"action": "assemble_page", "config": {"content_field": "page_content.response.page_html", "add_navigation": false}, "next_step": "deploy_page", "description": "Assemble full page HTML from components", "output_field": "assembled_page"}, "complete_page": {"action": "loop_complete", "description": "Page build complete"}, "update_page_status": {"action": "update_page_status", "config": {"status": "deployed", "commit_from": "page_deployed.commit_sha", "page_id_field": "current_page.id"}, "next_step": "complete_page", "description": "Mark page as deployed in database"}, "write_page_content": {"action": "call_agent", "config": {"agent_type": "page-content-writer", "target_role": "content_writer", "input_fields": ["current_page", "site_record", "reviewed_brief", "style_collection"], "timeout_seconds": 300}, "next_step": "review_page_content", "description": "Write content for this page", "output_field": "page_content"}, "review_page_content": {"action": "call_agent", "config": {"agent_type": "content-reviewer", "target_role": "reviewer", "input_fields": ["current_page", "page_content", "reviewed_brief"], "timeout_seconds": 3900}, "next_step": "assemble_page", "description": "Review page content (HITL or auto-eval)", "output_field": "reviewed_content"}}, "start_step": "write_page_content"}, "item_variable": "current_page", "max_iterations": 20}, "next_step": "trigger_site_deploy", "description": "Build each page: write → review → deploy", "output_field": "pages_built"}, "store_hero_asset": {"action": "store_asset", "config": {"purpose": "hero", "url_field": "hero_result.image_url", "asset_type": "image", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "hero_images.home", "origin_prompt_field": "site_plan.image_prompts.hero_home", "update_site_brand_assets": true}, "next_step": "select_style_collection", "description": "Store generated hero image", "output_field": "hero_stored"}, "store_logo_asset": {"action": "store_asset", "config": {"purpose": "brand_logo", "url_field": "logo_result.image_url", "asset_type": "logo", "origin_type": "generated", "site_id_field": "site_record.site_id", "brand_asset_key": "logo.primary", "origin_prompt_field": "site_plan.image_prompts.logo", "update_site_brand_assets": true}, "next_step": "check_hero_images", "description": "Store generated logo in assets table and site brand_assets", "output_field": "logo_stored"}, "sync_pages_to_db": {"action": "sync_pages_to_db", "config": {"input_fields": ["site_record", "site_plan"]}, "next_step": "check_assets_needed", "description": "Create page records from site plan", "output_field": "pages_synced"}, "call_site_planner": {"action": "call_agent", "config": {"agent_type": "site-planner", "target_role": "planner", "input_fields": ["input_data", "site_record", "reviewed_brief"], "timeout_seconds": 120}, "next_step": "store_reviewed_brief", "description": "Plan pages, select components, identify asset needs", "output_field": "site_plan"}, "check_hero_images": {"action": "conditional", "config": {"condition": "site_plan.validated_plan.needs_images == true", "else_step": "select_style_collection", "then_step": "generate_hero_image"}, "description": "Check if hero images need to be generated"}, "ensure_site_record": {"action": "ensure_site_record", "config": {"store_brief_in_content_data": true}, "next_step": "call_site_planner", "description": "Create or update site record in database", "output_field": "site_record"}, "get_pages_to_build": {"action": "get_pages_to_build", "config": {"include_all": false, "build_statuses": ["planned", "needs_rebuild"]}, "next_step": "build_pages_loop", "description": "Query pages from database that need content generation", "output_field": "pages_to_build"}, "update_site_status": {"action": "update_site_status", "config": {"status": "deployed", "deployed_at": "now", "site_id_field": "site_record.site_id"}, "next_step": "complete", "description": "Mark site as deployed", "output_field": "site_updated"}, "check_assets_needed": {"action": "conditional", "config": {"condition": "site_plan.validated_plan.needs_logo == true OR site_plan.validated_plan.needs_images == true", "else_step": "select_style_collection", "then_step": "spawn_image_generator"}, "description": "Check if logo or images need to be generated"}, "generate_hero_image": {"action": "call_agent", "config": {"prompt": "{{site_plan.image_prompts.hero_home}}", "agent_type": "image-generator", "target_role": "image_generator", "input_fields": ["site_plan", "reviewed_brief"], "timeout_seconds": 120}, "next_step": "store_hero_asset", "description": "Generate hero image for home page", "output_field": "hero_result"}, "trigger_site_deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_fields": ["site_record", "pages_built"], "timeout_seconds": 180}, "next_step": "update_site_status", "description": "Trigger Cloudflare deployment", "output_field": "deployment_result"}, "call_logo_generation": {"action": "call_agent", "config": {"prompt": "{{site_plan.image_prompts.logo}}", "agent_type": "image-generator", "target_role": "image_generator", "input_fields": ["site_plan", "reviewed_brief"], "timeout_seconds": 120}, "next_step": "store_logo_asset", "description": "Generate logo using image-generator agent", "output_field": "logo_result"}, "spawn_content_writer": {"action": "spawn_agent", "config": {"role": "content_writer", "agent_type": "page-content-writer"}, "next_step": "spawn_reviewer", "description": "Spawn content writer agent", "output_field": "content_writer_agent"}, "store_reviewed_brief": {"action": "update_site_content", "config": {"merge": true, "content_field": "input_data.reviewed_brief", "site_id_field": "site_record.site_id"}, "next_step": "store_site_plan", "description": "Store the reviewed brief in sites.content_data", "output_field": "brief_stored"}, "spawn_image_generator": {"action": "spawn_agent", "config": {"role": "image_generator", "agent_type": "image-generator"}, "next_step": "generate_logo", "description": "Spawn image generator agent for asset creation", "output_field": "image_generator_info"}, "set_default_components": {"action": "update_site_defaults", "config": {"defaults": {"head": "head-seo-standard", "footer_from": "style_collection.footer_component_name", "header_from": "style_collection.header_component_name"}, "site_id_field": "site_record.site_id"}, "next_step": "get_pages_to_build", "description": "Set default head/header/footer components", "output_field": "defaults_set"}, "select_style_collection": {"action": "select_style_collection", "config": {"style_from": "site_plan.style_collection", "site_id_field": "site_record.site_id", "fallback_by_domain": true}, "next_step": "set_default_components", "description": "Choose style collection based on site plan", "output_field": "style_collection"}}, "start_step": "spawn_planner"}, "processing_mode": "orchestrator", "timeout_seconds": 900} | t         | 2025-12-22 17:46:51.419068+00 | 2026-01-13 21:06:34.477623+00 |            | ["orchestration", "website-builder", "component-based", "image-generation"] | docker.io/aqls/agent-chassis | v1.0.667  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | ["website", "multipage", "component-based"] | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]} |           0 | f           | {"expects": {"reviewed_brief": "object - completed questionnaire answers", "input_data.domain": "string - the domain name", "input_data.objective": "string - what the site should achieve"}, "required": ["input_data.domain", "reviewed_brief"]} | {"produces": {"site_id": "uuid - the site record ID", "deploy_url": "string - the live site URL", "pages_built": "number - count of pages deployed"}}
(1 row)

-- deploy fix and input and output contracts

-- 001_fix_pageflow_skip_deploy.sql
-- Step 1.2: Make build_pages_loop go directly to update_site_status
-- This skips the failing deployer-agent call
--
-- Background: Pages are already committed individually in the loop via git_commit action.
-- The separate deployer-agent call after the loop is redundant and fails because
-- it expects data at different paths than pageflow-builder provides.

BEGIN;

-- Change build_pages_loop to skip trigger_site_deploy
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,next_step}',
        '"update_site_status"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'build_pages_loop'->>'next_step' as loop_next_step
FROM agent_definitions
WHERE type = 'pageflow-builder';

COMMIT;

-- ROLLBACK if needed:
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(default_config, '{workflow,steps,build_pages_loop,next_step}', '"trigger_site_deploy"')
-- WHERE type = 'pageflow-builder';

---

-- not sure about this one:
-- 008_fix_pageflow_builder_data_paths.sql
--
-- PROBLEM: call_agent wraps responses in metadata structure
--
-- When pageflow-builder calls site-planner via call_agent, the result looks like:
--   site_plan: {
--     response: { validated_plan: { needs_logo: true, ... } },  <-- actual data
--     request_id: "...",
--     action_sent: "...",
--     ...call_agent metadata...
--   }
--
-- But workflows expect direct access:
--   site_plan.validated_plan.needs_logo  (FAILS - validated_plan not at top level)
--
-- FIX: Update all paths to include .response where call_agent results are used
--
-- Log evidence:
--   "path":"site_plan.validated_plan.needs_logo"
--   "available_keys":["response","request_id","action_sent","await_response",...]
--   "part":"validated_plan" <- NOT FOUND because it's under "response"
--
-- NOTE: content-reviewer's review_mode issue is NOT a bug - it correctly
-- defaults to auto_eval when review_mode is not specified

BEGIN;

-- First, verify current paths
SELECT
    type,
    default_config->'workflow'->'steps'->'check_assets_needed'->'config'->>'condition' as check_assets_condition,
    default_config->'workflow'->'steps'->'check_hero_images'->'config'->>'condition' as check_hero_condition,
    default_config->'workflow'->'steps'->'generate_logo'->'config'->>'condition' as generate_logo_condition
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Fix check_assets_needed conditional
-- BEFORE: site_plan.validated_plan.needs_logo == true OR site_plan.validated_plan.needs_images == true
-- AFTER:  site_plan.response.validated_plan.needs_logo == true OR site_plan.response.validated_plan.needs_images == true
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_assets_needed,config,condition}',
        '"site_plan.response.validated_plan.needs_logo == true OR site_plan.response.validated_plan.needs_images == true"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Fix check_hero_images conditional
-- BEFORE: site_plan.validated_plan.needs_images == true
-- AFTER:  site_plan.response.validated_plan.needs_images == true
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_hero_images,config,condition}',
        '"site_plan.response.validated_plan.needs_images == true"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Fix generate_logo conditional (if it exists with similar path)
-- BEFORE: site_plan.needs_logo == true
-- AFTER:  site_plan.response.needs_logo == true
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_logo,config,condition}',
        '"site_plan.response.needs_logo == true"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Fix store_hero_asset and store_logo_asset paths for image_prompts
-- These reference site_plan.image_prompts.hero_home and site_plan.image_prompts.logo
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,config,origin_prompt_field}',
        '"site_plan.response.image_prompts.hero_home"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset,config,origin_prompt_field}',
        '"site_plan.response.image_prompts.logo"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Fix select_style_collection path
-- BEFORE: site_plan.style_collection
-- AFTER:  site_plan.response.style_collection
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,select_style_collection,config,style_from}',
        '"site_plan.response.style_collection"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Fix generate_hero_image and call_logo_generation prompt paths
-- These use template syntax {{site_plan.image_prompts.hero_home}}
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_hero_image,config,prompt}',
        '"{{site_plan.response.image_prompts.hero_home}}"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_generation,config,prompt}',
        '"{{site_plan.response.image_prompts.logo}}"'
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- Verify changes
SELECT
    type,
    default_config->'workflow'->'steps'->'check_assets_needed'->'config'->>'condition' as check_assets_fixed,
    default_config->'workflow'->'steps'->'check_hero_images'->'config'->>'condition' as check_hero_fixed,
    default_config->'workflow'->'steps'->'select_style_collection'->'config'->>'style_from' as style_from_fixed
FROM agent_definitions
WHERE type = 'pageflow-builder';

COMMIT;

-- ROLLBACK (restore original paths):
/*
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,check_assets_needed,config,condition}',
    '"site_plan.validated_plan.needs_logo == true OR site_plan.validated_plan.needs_images == true"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,check_hero_images,config,condition}',
    '"site_plan.validated_plan.needs_images == true"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,generate_logo,config,condition}',
    '"site_plan.needs_logo == true"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,store_hero_asset,config,origin_prompt_field}',
    '"site_plan.image_prompts.hero_home"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,store_logo_asset,config,origin_prompt_field}',
    '"site_plan.image_prompts.logo"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,select_style_collection,config,style_from}',
    '"site_plan.style_collection"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,generate_hero_image,config,prompt}',
    '"{{site_plan.image_prompts.hero_home}}"')
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,call_logo_generation,config,prompt}',
    '"{{site_plan.image_prompts.logo}}"')
WHERE type = 'pageflow-builder';
*/