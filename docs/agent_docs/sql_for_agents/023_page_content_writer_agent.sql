
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

-- Find the current template
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->'prompt_template'
FROM agent_definitions
WHERE type = 'page-content-writer';

-- Update the template to use category instead of function
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(replace(
                default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                '{{.current_section.function}}',
                '{{.current_section.category}}'
                 ))
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';


====
-- navigation fixes

-- Update page-content-writer build_render_context to include db_sync and site_id_field
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,build_render_context,config,sources,db_sync}',
                '"input_data.db_sync"'
        ),
        '{workflow,steps,build_render_context,config,site_id_field}',
        '"input_data.site_record.site_id"'
                     )
WHERE type = 'page-content-writer';


--- link constraints
-- Update page-content-writer to receive available pages for linking
--
-- Problem: Content writer hallucinates links to non-existent pages
-- Solution: Pass list of available pages in input_mapping and include in prompt
--
-- The content writer already receives db_sync which contains pages info.
-- We update the prompt to:
-- 1. List available pages
-- 2. Instruct to ONLY link to these pages
--
-- Content/link suggestions are handled separately by maintenance workflow.

-- ============================================================
-- 1. Check current prompt template (this is informational)
-- ============================================================
-- The current prompt is in the execute_llm_prompt step
-- We need to add available pages context

-- ============================================================
-- 2. Update the process_sections_loop LLM prompt
-- ============================================================
-- First, let's see the current structure
-- The content writer has a "process_single_section" step with execute_llm_prompt

-- We need to add a preamble about available pages

-- ============================================================
-- 3. Update render_context building to include available pages
-- ============================================================
-- The build_render_context action should include pages list

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_render_context,config,sources,available_pages}',
        '"db_sync.pages"'
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- ============================================================
-- 4. Create a simplified update to add link constraints to prompts
-- ============================================================
-- Rather than modifying complex nested prompt templates via SQL,
-- we'll add a system instruction that the LLM prompt action can include

-- Add a link_constraints field to the workflow config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{link_constraints}',
        '{
            "enabled": true,
            "max_internal_links_per_section": 3
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- ============================================================
-- Note: The actual prompt update requires Go code changes
-- ============================================================
-- The execute_llm_prompt action needs to:
-- 1. Check for link_constraints in config
-- 2. Extract available pages from render_context.available_pages
-- 3. Prepend constraint text to the prompt
--
-- Example prompt addition:
--
-- ## Available Pages for Internal Links
-- You may ONLY create links to these pages:
-- {{range .available_pages}}
-- - {{.url}}: {{.title}}
-- {{end}}
--
-- DO NOT invent page URLs. If mentioning a topic without a page,
-- do not create a link for it.

-- ============================================================
-- VERIFY
-- ============================================================
SELECT type,
       default_config->'link_constraints' as link_constraints,
       default_config->'workflow'->'steps'->'build_render_context'->'config'->'sources' as render_sources
FROM agent_definitions
WHERE type = 'page-content-writer';


-- Update page-content-writer to prepare link context
--
-- Adds a step that runs BEFORE content generation to:
-- 1. Extract available pages from db_sync
-- 2. Build constraint text for prompt inclusion
-- 3. Store in link_context for use by LLM prompts
--
-- The constraint text is then available as {{.link_context.link_constraint_text}}
-- in prompt templates.

-- ============================================================
-- 1. Add prepare_link_context step
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,prepare_link_context}',
        '{
            "action": "prepare_link_context",
            "config": {
                "enabled": true,
                "pages_field": "db_sync.pages",
                "max_links_per_section": 3
            },
            "description": "Prepare available pages context for internal linking",
            "next_step": "load_page_components",
            "output_field": "link_context"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- ============================================================
-- 2. Update spawn_research_agent to go to prepare_link_context
-- ============================================================
-- Current flow: spawn_research_agent → load_page_components
-- New flow: spawn_research_agent → prepare_link_context → load_page_components

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_research_agent,next_step}',
        '"prepare_link_context"'
                     )
WHERE type = 'page-content-writer';

-- ============================================================
-- 3. Add link_context to input_fields for LLM prompts
-- ============================================================
-- The process_single_section step (or similar) needs to include link_context

-- First, let's check what the section processing step is called
-- and add link_context to its input_fields

-- For the section content generation, add link_context to inputs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_section_content,config,input_fields}',
        '["render_context", "current_section", "section_component", "link_context"]'::jsonb
                     )
WHERE type = 'page-content-writer'
  AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps' ? 'generate_section_content';

-- ============================================================
-- 4. Note on prompt template update
-- ============================================================
-- The LLM prompt template for section generation needs to include:
--
-- {{if .link_context.link_constraint_text}}
-- {{.link_context.link_constraint_text}}
-- {{end}}
--
-- This should be added near the beginning of the prompt, before
-- the main content generation instructions.
--
-- Example prompt structure:
-- """
-- {{if .link_context.link_constraint_text}}
-- {{.link_context.link_constraint_text}}
--
-- ---
-- {{end}}
--
-- Generate content for the {{.current_section.name}} section...
-- """

-- ============================================================
-- VERIFY
-- ============================================================
SELECT type,
       default_config->'workflow'->'start_step' as start_step,
       default_config->'workflow'->'steps'->'spawn_research_agent'->>'next_step' as after_spawn,
    default_config->'workflow'->'steps'->'prepare_link_context'->>'next_step' as after_link_context
FROM agent_definitions
WHERE type = 'page-content-writer';