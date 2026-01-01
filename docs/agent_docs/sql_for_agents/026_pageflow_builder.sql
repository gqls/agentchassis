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