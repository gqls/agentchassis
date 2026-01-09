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