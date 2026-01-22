-- ============================================================================
-- Convert call_agent steps from deprecated input_fields to explicit input_mapping
-- ============================================================================
-- This migration converts the legacy input_fields arrays to explicit input_mapping
-- objects which provide clearer source paths for data resolution.
--
-- Pattern:
--   BEFORE: "input_fields": ["field1", "field2"]
--   AFTER:  "input_mapping": { "field1": "field1", "field2": "nested.path.field2" }
--
-- Note: Some fields need paths like "input_data.reviewed_brief" because that's
-- where the fallback search actually finds them.
-- ============================================================================

-- ============================================================================
-- 1. PAGEFLOW-BUILDER - Main workflow steps
-- ============================================================================

-- 1a. review_page_content (inside build_pages_loop sub_workflow)
-- From: input_fields: ["current_page", "page_content", "reviewed_brief"]
-- Note: reviewed_brief is found at input_data.reviewed_brief per fallback logs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,review_page_content,config}',
        (default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'review_page_content'->'config')
            - 'input_fields'
            || '{"input_mapping": {"current_page": "current_page", "page_content": "page_content", "reviewed_brief": "input_data.reviewed_brief"}}'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- 1b. generate_hero_image
-- From: input_fields: ["site_plan", "reviewed_brief"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_hero_image,config}',
        (default_config->'workflow'->'steps'->'generate_hero_image'->'config')
            - 'input_fields'
            || '{"input_mapping": {"site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}}'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- 1c. trigger_site_deploy
-- From: input_fields: ["site_record", "pages_built"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,trigger_site_deploy,config}',
        (default_config->'workflow'->'steps'->'trigger_site_deploy'->'config')
            - 'input_fields'
            || '{"input_mapping": {"site_record": "site_record", "pages_built": "pages_built"}}'::jsonb
                     )
WHERE type = 'pageflow-builder';

-- 1d. call_logo_generation
-- From: input_fields: ["site_plan", "reviewed_brief"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_generation,config}',
        (default_config->'workflow'->'steps'->'call_logo_generation'->'config')
            - 'input_fields'
            || '{"input_mapping": {"site_plan": "site_plan", "reviewed_brief": "input_data.reviewed_brief"}}'::jsonb
                     )
WHERE type = 'pageflow-builder';


-- ============================================================================
-- 2. MULTIPAGE-WEBSITE-BUILDER
-- ============================================================================

-- 2a. call_strategist
-- From: input_fields: ["input_data", "site_record"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_strategist,config}',
        (default_config->'workflow'->'steps'->'call_strategist'->'config')
            - 'input_fields'
            || '{"input_mapping": {"input_data": "input_data", "site_record": "site_record"}}'::jsonb
                     )
WHERE type = 'multipage-website-builder';

-- 2b. deploy
-- From: input_fields: ["site_files", "input_data", "site_record"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy,config}',
        (default_config->'workflow'->'steps'->'deploy'->'config')
            - 'input_fields'
            || '{"input_mapping": {"site_files": "site_files", "input_data": "input_data", "site_record": "site_record"}}'::jsonb
                     )
WHERE type = 'multipage-website-builder';

-- 2c. generate_pages_loop substeps - generate_content
-- From: input_fields: ["current_page", "input_data", "page_plan"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_pages_loop,config,substeps,generate_content,config}',
        (default_config->'workflow'->'steps'->'generate_pages_loop'->'config'->'substeps'->'generate_content'->'config')
            - 'input_fields'
            || '{"input_mapping": {"current_page": "current_page", "input_data": "input_data", "page_plan": "page_plan"}}'::jsonb
                     )
WHERE type = 'multipage-website-builder';

-- 2d. generate_pages_loop substeps - create_html
-- From: input_fields: ["page_content", "current_page", "input_data", "db_sync", "page_plan"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_pages_loop,config,substeps,create_html,config}',
        (default_config->'workflow'->'steps'->'generate_pages_loop'->'config'->'substeps'->'create_html'->'config')
            - 'input_fields'
            || '{"input_mapping": {"page_content": "page_content", "current_page": "current_page", "input_data": "input_data", "db_sync": "db_sync", "page_plan": "page_plan"}}'::jsonb
                     )
WHERE type = 'multipage-website-builder';


-- ============================================================================
-- 3. INTAKE-ORCHESTRATOR
-- ============================================================================

-- 3a. call_classifier
-- From: input_fields: ["input_data", "available_builders"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_classifier,config}',
        (default_config->'workflow'->'steps'->'call_classifier'->'config')
            - 'input_fields'
            || '{"input_mapping": {"input_data": "input_data", "available_builders": "available_builders"}}'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- 3b. call_briefer
-- From: input_fields: ["input_data", "classification", "confirmed_type", "questionnaire"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_briefer,config}',
        (default_config->'workflow'->'steps'->'call_briefer'->'config')
            - 'input_fields'
            || '{"input_mapping": {"input_data": "input_data", "classification": "classification", "confirmed_type": "confirmed_type", "questionnaire": "questionnaire"}}'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- 3c. call_builder
-- From: input_fields: ["input_data", "classification", "confirmed_type", "brief_data", "reviewed_brief", "questionnaire"]
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_builder,config}',
        (default_config->'workflow'->'steps'->'call_builder'->'config')
            - 'input_fields'
            || '{"input_mapping": {"input_data": "input_data", "classification": "classification", "confirmed_type": "confirmed_type", "brief_data": "brief_data", "reviewed_brief": "reviewed_brief", "questionnaire": "questionnaire"}}'::jsonb
                     )
WHERE type = 'intake-orchestrator';


-- ============================================================================
-- 4. PAGE-CONTENT-WRITER (sub_workflow call_researcher)
-- ============================================================================

-- 4a. call_researcher inside process_sections_loop
-- From: input_fields: ["current_section", "reviewed_brief", "site_record"]
-- Note: reviewed_brief comes from input_data.reviewed_brief
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,call_researcher,config}',
        (default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'call_researcher'->'config')
            - 'input_fields'
            || '{"input_mapping": {"current_section": "current_section", "reviewed_brief": "input_data.reviewed_brief", "site_record": "input_data.site_record"}}'::jsonb
                     )
WHERE type = 'page-content-writer';


-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- Check pageflow-builder updates
SELECT
    'pageflow-builder' as agent,
    'review_page_content' as step,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'review_page_content'->'config' as config
FROM agent_definitions WHERE type = 'pageflow-builder'
UNION ALL
SELECT
    'pageflow-builder',
    'generate_hero_image',
    default_config->'workflow'->'steps'->'generate_hero_image'->'config'
FROM agent_definitions WHERE type = 'pageflow-builder'
UNION ALL
SELECT
    'pageflow-builder',
    'trigger_site_deploy',
    default_config->'workflow'->'steps'->'trigger_site_deploy'->'config'
FROM agent_definitions WHERE type = 'pageflow-builder'
UNION ALL
SELECT
    'pageflow-builder',
    'call_logo_generation',
    default_config->'workflow'->'steps'->'call_logo_generation'->'config'
FROM agent_definitions WHERE type = 'pageflow-builder';

-- Check multipage-website-builder updates
SELECT
    'multipage-website-builder' as agent,
    'call_strategist' as step,
    default_config->'workflow'->'steps'->'call_strategist'->'config' as config
FROM agent_definitions WHERE type = 'multipage-website-builder'
UNION ALL
SELECT
    'multipage-website-builder',
    'deploy',
    default_config->'workflow'->'steps'->'deploy'->'config'
FROM agent_definitions WHERE type = 'multipage-website-builder';

-- Check intake-orchestrator updates
SELECT
    'intake-orchestrator' as agent,
    'call_classifier' as step,
    default_config->'workflow'->'steps'->'call_classifier'->'config' as config
FROM agent_definitions WHERE type = 'intake-orchestrator'
UNION ALL
SELECT
    'intake-orchestrator',
    'call_briefer',
    default_config->'workflow'->'steps'->'call_briefer'->'config'
FROM agent_definitions WHERE type = 'intake-orchestrator'
UNION ALL
SELECT
    'intake-orchestrator',
    'call_builder',
    default_config->'workflow'->'steps'->'call_builder'->'config'
FROM agent_definitions WHERE type = 'intake-orchestrator';