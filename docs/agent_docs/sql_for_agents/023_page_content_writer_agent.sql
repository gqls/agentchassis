
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

--

-- backup before changing prompt

5946a27b-38ab-41e8-8b49-7bc1a4b626b8 | page-content-writer | Page Content Writer | Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content. | specialist | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_field": "page_content"}}, "compile_page": {"action": "compile_page_sections", "config": {"page_from": "input_data.current_page", "inject_footer": true, "inject_header": true, "sections_from": "processed_sections", "include_research_ids": true}, "next_step": "complete", "description": "Compile all sections into page output", "output_field": "page_content"}, "build_render_context": {"action": "build_render_context", "config": {"sources": {"page": "input_data.current_page", "site": "input_data.site_record", "brief": "input_data.reviewed_brief", "style": "input_data.style_collection", "assets": "brand_assets", "db_sync": "input_data.db_sync", "available_pages": "db_sync.pages"}, "site_id_field": "input_data.site_record.site_id"}, "next_step": "process_sections_loop", "description": "Build render context from brief, site, and brand data", "output_field": "render_context"}, "load_page_components": {"action": "load_page_section_components", "config": {"page_from": "input_data.current_page", "sections_from": "input_data.current_page.sections", "include_templates": true, "include_input_schema": true}, "next_step": "build_render_context", "description": "Load component definitions for this page's sections", "output_field": "section_components"}, "prepare_link_context": {"action": "prepare_link_context", "config": {"enabled": true, "pages_field": "db_sync.pages", "max_links_per_section": 3}, "next_step": "load_page_components", "description": "Prepare available pages context for internal linking", "output_field": "link_context"}, "spawn_research_agent": {"action": "spawn_agent", "config": {"role": "researcher", "agent_type": "research-agent", "await_response": true}, "next_step": "prepare_link_context", "description": "Spawn research agent in case sections need research", "output_field": "researcher_info"}, "process_sections_loop": {"action": "loop", "config": {"loop_var": "current_section", "iterate_over": "section_components.components", "sub_workflow": {"steps": {"render_section": {"action": "render_component", "config": {"output_html": true, "content_from": "generated_content.result", "context_from": "render_context", "component_from": "current_section"}, "description": "Render LLM-generated content into component template", "output_field": "section_output"}, "call_researcher": {"action": "call_agent", "config": {"agent_type": "research-agent", "target_role": "researcher", "input_mapping": {"site_record": "input_data.site_record", "reviewed_brief": "input_data.reviewed_brief", "current_section": "current_section"}, "timeout_seconds": 90}, "next_step": "generate_content", "description": "Research topic for this section", "output_field": "research_result"}, "generate_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 2000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["current_section", "render_context", "reviewed_brief", "current_page", "link_context"], "output_format": "json", "prompt_template": "Write content for the {{.current_section.category}} section of {{.current_page.title}}.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n{{if .link_context.link_constraint_text}}\n{{.link_context.link_constraint_text}}\n\n{{end}}\n## Section Requirements\nComponent: {{.current_section.name}}\nFunction: {{.current_section.category}}\nPurpose: {{.current_section.description}}\n\n## Data Schema Required\n{{.current_section.input_schema}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON with these EXACT field names (use the ones that apply to this component type):\n\n### For Hero/Banner sections:\n```json\n{\n  \"headline\": \"Your Compelling Main Headline\",\n  \"subheadline\": \"Supporting text that expands on the headline\",\n  \"primary_cta\": \"Get Started\",\n  \"primary_cta_url\": \"/contact.html\",\n  \"secondary_cta\": \"Learn More\",\n  \"secondary_cta_url\": \"/about.html\"\n}\n```\n\n### For Feature/Services sections:\n```json\n{\n  \"headline\": \"Section Headline\",\n  \"subheadline\": \"Brief introduction\",\n  \"features\": [\n    {\"name\": \"Feature Name\", \"description\": \"Feature description\", \"icon\": \"icon-name\"},\n    {\"name\": \"Feature 2\", \"description\": \"Description 2\", \"icon\": \"icon-name\"}\n  ]\n}\n```\n\n### For CTA/Call-to-Action sections:\n```json\n{\n  \"headline\": \"Ready to Get Started?\",\n  \"subheadline\": \"Contact us today\",\n  \"primary_cta\": \"Contact Us\",\n  \"primary_cta_url\": \"/contact.html\"\n}\n```\n\n### For Text/Content sections:\n```json\n{\n  \"heading\": \"Section Heading\",\n  \"content\": \"Paragraph content here...\"\n}\n```\n\nRules:\n- Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided\n- Only create internal links to pages listed in the Internal Links section above"}, "next_step": "render_section", "description": "Generate content for this section", "output_field": "generated_content"}, "check_render_mode": {"action": "conditional", "config": {"condition": "current_section.render_mode == 'agent' OR current_section.needs_llm == true", "else_step": "render_from_template", "then_step": "check_needs_research"}, "description": "Check if section needs LLM or just template"}, "check_needs_research": {"action": "conditional", "config": {"condition": "current_section.needs_research == true", "else_step": "generate_content", "then_step": "call_researcher"}, "description": "Check if section needs research first"}, "render_from_template": {"action": "render_component", "config": {"output_html": true, "content_from": "render_context", "context_from": "render_context", "component_from": "current_section"}, "description": "Render section from template with brief data only", "output_field": "section_output"}}, "start_step": "check_render_mode"}, "max_iterations": 15}, "next_step": "compile_page", "description": "Process each section - template render or LLM generate", "output_field": "processed_sections"}}, "start_step": "spawn_research_agent"}, "processing_mode": "task", "timeout_seconds": 180, "link_constraints": {"enabled": true, "max_internal_links_per_section": 3}} | t         | 2025-12-22 17:47:17.609605+00 | 2026-01-31 17:51:01.881339+00 |            | ["content-generation", "template-rendering", "research"] | docker.io/aqls/agent-chassis | v1.0.737  |         | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | []          | {}                     |           0 | f           | {"optional": ["reviewed_brief", "style_collection", "db_sync", "generated_images"], "required": ["current_page", "site_record"]} | {"produces": ["page_html", "metadata", "seo_data"]}
(1 row)

-- Update page-content-writer prompt to:
-- 1. Provide contact info explicitly
-- 2. Forbid hallucination of emails/phones
-- 3. Require proper HTML formatting (<p> tags for body text)

-- The prompt_template is deeply nested, so we need to update it carefully
-- Path: default_config -> workflow -> steps -> process_sections_loop -> config -> sub_workflow -> steps -> generate_content -> config -> prompt_template

-- ============================================================
-- 1. Update the prompt template
-- ============================================================
-- =============================================================
-- Update page-content-writer prompt template
--
-- Adds:
--   - Official Contact Information section with brief variables
--   - Contact Info JSON example for contact sections
--   - Strict rules about not inventing contact details
--   - HTML paragraph wrapping guidance
--   - Text/Content/About section type
--
-- Preserves:
--   - link_context block (was accidentally dropped in earlier draft)
--   - Research findings block
--   - All existing JSON examples
--
-- Path: workflow.steps.process_sections_loop.config.sub_workflow
--        .steps.generate_content.config.prompt_template
-- =============================================================

-- Verify current prompt exists at expected path
SELECT
    length(default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template') as current_prompt_length,
    left(default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template', 80) as prompt_start
FROM agent_definitions
WHERE type = 'page-content-writer';

-- Apply updated prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb('Write content for the {{.current_section.category}} section of {{.current_page.title}}.

## Company Context
Company: {{.render_context.company_name}}
Industry: {{.render_context.industry}}
Tone: {{.render_context.tone}}
Target Audience: {{.render_context.target_audience}}
Services: {{.reviewed_brief.services}}
Tagline: {{.render_context.tagline}}

## Official Contact Information (USE ONLY THESE - DO NOT INVENT)
Email: {{.reviewed_brief.contact_email}}
Phone: {{.reviewed_brief.contact_phone}}
Location: {{.reviewed_brief.headquarters}}

{{if .link_context.link_constraint_text}}
## Internal Linking
{{.link_context.link_constraint_text}}

{{end}}
## Section Requirements
Component: {{.current_section.name}}
Function: {{.current_section.category}}
Purpose: {{.current_section.description}}

## Data Schema Required
{{.current_section.input_schema}}

{{if .research_result}}
## Research Findings
{{.research_result.response.summary}}

Sources:
{{range $index, $src := .research_result.response.sources}}
- [{{$index}}] {{$src.title}} ({{$src.domain}})
{{end}}
{{end}}

## Task
Write compelling, specific content for this section.

Return JSON with these EXACT field names (use the ones that apply to this component type):

### For Hero/Banner sections:
```json
{
  "headline": "Your Compelling Main Headline",
  "subheadline": "Supporting text that expands on the headline",
  "primary_cta": "Get Started",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "Learn More",
  "secondary_cta_url": "/about.html"
}
```

### For Feature/Services sections:
```json
{
  "headline": "Section Headline",
  "subheadline": "Brief introduction",
  "features": [
    {"name": "Feature Name", "description": "Feature description", "icon": "icon-name"},
    {"name": "Feature 2", "description": "Description 2", "icon": "icon-name"}
  ]
}
```

### For CTA/Call-to-Action sections:
```json
{
  "headline": "Ready to Get Started?",
  "subheadline": "Contact us today",
  "primary_cta": "Contact Us",
  "primary_cta_url": "/contact.html"
}
```

### For Text/Content/About sections:
```json
{
  "heading": "Section Heading",
  "content": "<p>First paragraph of content here.</p><p>Second paragraph here.</p><p>Third paragraph if needed.</p>"
}
```

### For Contact Info sections:
```json
{
  "heading": "Contact Us",
  "email": "USE_BRIEF_EMAIL",
  "phone": "USE_BRIEF_PHONE",
  "hours": "Monday-Friday 9am-5pm"
}
```

## STRICT RULES:
1. Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)
2. No placeholder text like [Your Company] or Lorem ipsum
3. Be specific to this company and industry
4. Professional but engaging tone matching the brief
5. Include source citations [0], [1] if research was provided
6. NEVER invent contact information - use ONLY the email and phone provided in Official Contact Information above
7. For body text content, ALWAYS wrap paragraphs in <p> tags - never output raw unwrapped text
8. For "content" fields that contain multiple paragraphs, use proper HTML: <p>Paragraph 1</p><p>Paragraph 2</p>
9. If contact email or phone is empty in the brief, do NOT make one up - omit it or use a generic "Contact us" link
10. Only create internal links to pages listed in the Internal Linking section above'::text)
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- Verify the update
SELECT
    length(default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template') as new_prompt_length,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%Official Contact Information%' as has_contact_section,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%link_context%' as has_link_context,
    version
FROM agent_definitions
WHERE type = 'page-content-writer';

--

-- phone details in contact

-- =============================================================
-- Fix: Contact phone missing from rendered pages
--
-- ROOT CAUSE:
-- 1. reviewed_brief may not have contact_phone (required=false
--    in briefing questionnaire)
-- 2. Site planner adds contact_phone to its output
-- 3. store_site_plan merges phone into DB content_data, BUT
--    this happens AFTER ensure_site_record already captured
--    a stale snapshot of site_record
-- 4. page-content-writer gets stale site_record + incomplete
--    reviewed_brief → render_context has email but no phone
--
-- FIX (3 parts):
-- A. Pass site_plan to page-content-writer as an additional
--    input (workflow change in pageflow-builder)
-- B. Add site_plan as a source in build_render_context
--    (workflow change in page-content-writer)
-- C. Update LLM prompt to use render_context.email/phone
--    instead of reviewed_brief.contact_* (prompt change in
--    page-content-writer)
-- =============================================================


-- =============================================================
-- PART A: Add site_plan to pageflow-builder's write_page_content
--         input_mapping
-- =============================================================

-- Current input_mapping:
-- "db_sync": "db_sync",
-- "site_record": "site_record",
-- "current_page": "current_page",
-- "reviewed_brief": "input_data.reviewed_brief",
-- "style_collection": "style_collection"
--
-- Adding: "site_plan": "site_plan"

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,input_mapping,site_plan}',
        '"site_plan"'
                     ),
    updated_at = now()
WHERE type = 'pageflow-builder';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config'->'input_mapping' as write_page_input_mapping
FROM agent_definitions
WHERE type = 'pageflow-builder';


-- =============================================================
-- PART B: Add site_plan as a source in page-content-writer's
--         build_render_context step
-- =============================================================

-- Current sources:
-- "page": "input_data.current_page",
-- "site": "input_data.site_record",
-- "brief": "input_data.reviewed_brief",
-- "style": "input_data.style_collection",
-- "assets": "brand_assets",
-- "db_sync": "input_data.db_sync",
-- "available_pages": "db_sync.pages"
--
-- Adding: "plan": "input_data.site_plan"

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_render_context,config,sources,plan}',
        '"input_data.site_plan"'
                     ),
    updated_at = now()
WHERE type = 'page-content-writer';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'build_render_context'->'config'->'sources' as render_context_sources
FROM agent_definitions
WHERE type = 'page-content-writer';


-- =============================================================
-- PART C: Update LLM prompt to use render_context for contact
--         info instead of reviewed_brief
--
-- The render_context is built from ALL sources (brief + site +
-- plan + style), so it has the most complete data. The prompt
-- should reference render_context.email and render_context.phone
-- which are guaranteed to be populated from whichever source
-- has them.
--
-- Also add site_plan to the generate_content input_fields so
-- it's available in the prompt template context.
-- =============================================================

-- First: add site_plan to generate_content input_fields
-- Current: ["current_section", "render_context", "reviewed_brief", "current_page", "link_context"]
-- New:     ["current_section", "render_context", "reviewed_brief", "current_page", "link_context", "site_plan"]

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}',
        '["current_section", "render_context", "reviewed_brief", "current_page", "link_context", "site_plan"]'::jsonb
                     ),
    updated_at = now()
WHERE type = 'page-content-writer';

-- Now update the prompt template to use render_context for contact info
-- We change:
--   Email: {{.reviewed_brief.contact_email}}
--   Phone: {{.reviewed_brief.contact_phone}}
--   Location: {{.reviewed_brief.headquarters}}
-- To:
--   Email: {{.render_context.email}}
--   Phone: {{.render_context.phone}}
--   Location: {{.reviewed_brief.headquarters}}
--
-- NOTE: We keep reviewed_brief.headquarters since render_context
-- doesn't extract that (it's not a field on RenderContext).
-- We also add a fallback note about the site_plan source.

-- Get current prompt to verify the exact text we're replacing
SELECT
    substring(
            default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
        FROM '## Official Contact.*?(?=\n\n)'
    ) as current_contact_block
FROM agent_definitions
WHERE type = 'page-content-writer';

-- Update the contact info section in the prompt
-- The exact text to replace (from our previous update in file 48):
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                        E'## Official Contact Information (USE ONLY THESE - DO NOT INVENT)\nEmail: {{.reviewed_brief.contact_email}}\nPhone: {{.reviewed_brief.contact_phone}}\nLocation: {{.reviewed_brief.headquarters}}',
                        E'## Official Contact Information (USE ONLY THESE - DO NOT INVENT)\nEmail: {{.render_context.email}}\nPhone: {{.render_context.phone}}\nLocation: {{.reviewed_brief.headquarters}}'
                )
        )
                     ),
    updated_at = now()
WHERE type = 'page-content-writer';

-- Verify the prompt now references render_context for contact
SELECT
    substring(
            default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
        FROM '## Official Contact.*?(?=\n\n)'
    ) as updated_contact_block
FROM agent_definitions
WHERE type = 'page-content-writer';


-- =============================================================
-- VERIFICATION: Show all changes made
-- =============================================================

-- 1. Pageflow-builder: write_page_content now passes site_plan
SELECT 'pageflow-builder input_mapping' as check,
    default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config'->'input_mapping' as value
FROM agent_definitions WHERE type = 'pageflow-builder'
UNION ALL
-- 2. Page-content-writer: build_render_context includes plan source
SELECT 'page-content-writer sources' as check,
    default_config->'workflow'->'steps'->'build_render_context'->'config'->'sources' as value
FROM agent_definitions WHERE type = 'page-content-writer'
UNION ALL
-- 3. Page-content-writer: generate_content input_fields includes site_plan
SELECT 'generate_content input_fields' as check,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->'input_fields' as value
FROM agent_definitions WHERE type = 'page-content-writer';

-- don't inject header and footer here

-- Corrected: target page-content-writer, not content-creator
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,compile_page,config,inject_header}',
                'false'::jsonb
        ),
        '{workflow,steps,compile_page,config,inject_footer}',
        'false'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

--

-- <head> fixes

-- ============================================================================
-- FIX: Move inject_head from pageflow-builder's assemble_page step
--      to page-content-writer's compile_page step
-- ============================================================================
-- ROOT CAUSE: CompilePageSectionsAction (page-content-writer) injects header
-- and footer but defers head injection to AssemblePageAction (pageflow-builder).
-- The placeholder <head> from buildPageHTML() travels across agent boundaries
-- and can get corrupted by cleanHTMLStructure's dedup logic, ending up inside
-- <body>. Fix: inject head at the same point as header/footer.
-- ============================================================================

-- Step 1: Add inject_head: true to page-content-writer's compile_page step
-- Current config: {"page_from": "input_data.current_page", "inject_footer": true, "inject_header": true, "sections_from": "processed_sections", "include_research_ids": true}
-- After:          + "inject_head": true
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,compile_page,config,inject_head}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

---

-- tone down the case studies and social proofs

-- Update page-content-writer prompt_template to prevent fabricated testimonials,
-- case studies, statistics, and fake people.
--
-- Changes to STRICT RULES section:
-- - Added rules 11-16 covering testimonials, case studies, statistics
-- - Modified rule 3 to distinguish "specific to the industry" from "invent specifics"

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
                E'Write content for the {{.current_section.category}} section of {{.current_page.title}}.\n\n'
        || E'## Company Context\n'
        || E'Company: {{.render_context.company_name}}\n'
        || E'Industry: {{.render_context.industry}}\n'
        || E'Tone: {{.render_context.tone}}\n'
        || E'Target Audience: {{.render_context.target_audience}}\n'
        || E'Services: {{.reviewed_brief.services}}\n'
        || E'Tagline: {{.render_context.tagline}}\n\n'
        || E'## Official Contact Information (USE ONLY THESE - DO NOT INVENT)\n'
        || E'Email: {{.render_context.email}}\n'
        || E'Phone: {{.render_context.phone}}\n'
        || E'Location: {{.reviewed_brief.headquarters}}\n\n'
        || E'{{if .link_context.link_constraint_text}}\n'
        || E'## Internal Linking\n'
        || E'{{.link_context.link_constraint_text}}\n\n'
        || E'{{end}}\n'
        || E'## Section Requirements\n'
        || E'Component: {{.current_section.name}}\n'
        || E'Function: {{.current_section.category}}\n'
        || E'Purpose: {{.current_section.description}}\n\n'
        || E'## Data Schema Required\n'
        || E'{{.current_section.input_schema}}\n\n'
        || E'{{if .research_result}}\n'
        || E'## Research Findings\n'
        || E'{{.research_result.response.summary}}\n\n'
        || E'Sources:\n'
        || E'{{range $index, $src := .research_result.response.sources}}\n'
        || E'- [{{$index}}] {{$src.title}} ({{$src.domain}})\n'
        || E'{{end}}\n'
        || E'{{end}}\n\n'
        || E'## Task\n'
        || E'Write compelling content for this section that is relevant to the company''s industry and services.\n\n'
        || E'Return JSON with these EXACT field names (use the ones that apply to this component type):\n\n'
        || E'### For Hero/Banner sections:\n'
        || E'```json\n'
        || E'{\n'
        || E'  "headline": "Your Compelling Main Headline",\n'
        || E'  "subheadline": "Supporting text that expands on the headline",\n'
        || E'  "primary_cta": "Get Started",\n'
        || E'  "primary_cta_url": "/contact.html",\n'
        || E'  "secondary_cta": "Learn More",\n'
        || E'  "secondary_cta_url": "/about.html"\n'
        || E'}\n'
        || E'```\n\n'
        || E'### For Feature/Services sections:\n'
        || E'```json\n'
        || E'{\n'
        || E'  "headline": "Section Headline",\n'
        || E'  "subheadline": "Brief introduction",\n'
        || E'  "features": [\n'
        || E'    {"name": "Feature Name", "description": "Feature description", "icon": "icon-name"},\n'
        || E'    {"name": "Feature 2", "description": "Description 2", "icon": "icon-name"}\n'
        || E'  ]\n'
        || E'}\n'
        || E'```\n\n'
        || E'### For CTA/Call-to-Action sections:\n'
        || E'```json\n'
        || E'{\n'
        || E'  "headline": "Ready to Get Started?",\n'
        || E'  "subheadline": "Contact us today",\n'
        || E'  "primary_cta": "Contact Us",\n'
        || E'  "primary_cta_url": "/contact.html"\n'
        || E'}\n'
        || E'```\n\n'
        || E'### For Text/Content/About sections:\n'
        || E'```json\n'
        || E'{\n'
        || E'  "heading": "Section Heading",\n'
        || E'  "content": "<p>First paragraph of content here.</p><p>Second paragraph here.</p><p>Third paragraph if needed.</p>"\n'
        || E'}\n'
        || E'```\n\n'
        || E'### For Contact Info sections:\n'
        || E'```json\n'
        || E'{\n'
        || E'  "heading": "Contact Us",\n'
        || E'  "email": "USE_BRIEF_EMAIL",\n'
        || E'  "phone": "USE_BRIEF_PHONE",\n'
        || E'  "hours": "Monday-Friday 9am-5pm"\n'
        || E'}\n'
        || E'```\n\n'
        || E'### For Testimonial/Social Proof sections:\n'
        || E'Use the company''s own voice to describe their approach and values. Do NOT create fake testimonials.\n'
        || E'```json\n'
        || E'{\n'
        || E'  "headline": "What We Stand For",\n'
        || E'  "subheadline": "Our commitment to our clients",\n'
        || E'  "testimonials": [\n'
        || E'    {"quote": "A statement about the company''s approach or philosophy in their own voice - NOT attributed to a fake person", "name": "", "title": ""},\n'
        || E'    {"quote": "Another statement about values or working approach", "name": "", "title": ""}\n'
        || E'  ]\n'
        || E'}\n'
        || E'```\n'
        || E'When the site owner adds real testimonials later, these placeholders will be replaced.\n\n'
        || E'### For Case Study sections:\n'
        || E'Describe the types of work the company does and the outcomes they aim for. Do NOT invent specific clients, projects, or results.\n'
        || E'```json\n'
        || E'{\n'
        || E'  "headline": "Our Work",\n'
        || E'  "subheadline": "How we help our clients",\n'
        || E'  "case_studies": [\n'
        || E'    {"title": "Type of project or service area", "description": "What this involves and what clients can expect", "result": "The kind of outcomes typically achieved"},\n'
        || E'    {"title": "Another service area", "description": "Description of approach", "result": "Typical outcomes"}\n'
        || E'  ]\n'
        || E'}\n'
        || E'```\n\n'
        || E'## STRICT RULES:\n'
        || E'1. Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)\n'
        || E'2. No placeholder text like [Your Company] or Lorem ipsum\n'
        || E'3. Write content that is relevant to this company''s industry and services - but do NOT invent specific achievements, metrics, or outcomes that have not actually happened\n'
        || E'4. Professional but engaging tone matching the brief\n'
        || E'5. Include source citations [0], [1] if research was provided\n'
        || E'6. NEVER invent contact information - use ONLY the email and phone provided in Official Contact Information above\n'
        || E'7. For body text content, ALWAYS wrap paragraphs in <p> tags - never output raw unwrapped text\n'
        || E'8. For "content" fields that contain multiple paragraphs, use proper HTML: <p>Paragraph 1</p><p>Paragraph 2</p>\n'
        || E'9. If contact email or phone is empty in the brief, do NOT make one up - omit it or use a generic "Contact us" link\n'
        || E'10. Only create internal links to pages listed in the Internal Linking section above\n'
        || E'11. NEVER invent fake people, client names, or attributed quotes. No "Sarah Mitchell, Founder" or "James Chen, Marketing Director". If the brief does not include real testimonials, write company philosophy statements instead and leave name/title fields empty\n'
        || E'12. NEVER invent specific statistics, percentages, or metrics. No "300% increase" or "47 minutes" or "42-page report". Describe the type of outcome without fabricating numbers\n'
        || E'13. NEVER invent fake case studies with named businesses. No "Brighton Physiotherapy Clinic" or "Independent Mortgage Broker" as if they were real clients. Describe service categories and typical outcomes instead\n'
        || E'14. For testimonial sections: write 2-3 statements in the company''s own voice about their values, approach, or commitment. These serve as placeholder content the site owner will replace with real testimonials\n'
        || E'15. For case study sections: describe the types of problems the company solves and the approach they take, without inventing specific clients or results\n'
        || E'16. It is ALWAYS better to be honest and general than specific and fabricated. A real visitor will trust "We help businesses improve their online presence" far more than a fake testimonial from an invented person'
        )
                     )
WHERE type = 'page-content-writer';

-- Verify the update
SELECT length(
               default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
       ) as prompt_length,
       substring(
               default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
    from 'NEVER invent fake people.*?empty'
) as rule_11_check
FROM agent_definitions
WHERE type = 'page-content-writer';

--

give access to site_specs to get tone of voice fetch
     -- Add read_site_spec step to page-content-writer
-- Insert between spawn_research_agent and prepare_link_context

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_research_agent,next_step}',
        '"load_site_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_site_specs}',
        '{
            "action": "read_site_spec",
            "config": {
                "site_id": "input_data.site_record.site_id"
            },
            "next_step": "prepare_link_context",
            "description": "Load all site specs for content direction and identity context",
            "output_field": "site_specs"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;
```

Then the LLM prompt template needs to reference the specs. The `generate_content` step's prompt already has `input_fields: ["current_section", "render_context", "reviewed_brief", ...]`. Add `"site_specs"` to that list, and add to the prompt:
```
{{if .site_specs.specs.content_direction}}
## Content Direction (from site spec — follow this closely)
Voice: {{.site_specs.specs.content_direction.voice}}
Emphasis: {{.site_specs.specs.content_direction.emphasis}}
Avoid phrases: {{.site_specs.specs.content_direction.avoid_phrases}}
Social proof style: {{.site_specs.specs.content_direction.social_proof_style}}
{{end}}

{{if .site_specs.specs.identity.target_audience}}
## Target Audience
{{.site_specs.specs.identity.target_audience}}
{{end}}

---

  -- make aware of content brief from admin

  -- Update page-content-writer prompt template to include content_brief
--
-- Inserts a conditional block before "## Data Schema Required" that uses
-- admin-edited content briefs when present on page_components.
--
-- STEP 1: Check what the current prompt looks like around the anchor point
-- STEP 1: Check the anchor exists and see the escaping pattern
SELECT position('Data Schema Required' in default_config::text) as anchor_pos,
       substring(default_config::text from position('Data Schema Required' in default_config::text) - 80 for 160) as around_anchor
FROM agent_definitions WHERE type = 'page-content-writer';

-- STEP 2: Apply the update (only if not already applied)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '## Data Schema Required',
        '{{if .current_section.content_brief}}## Admin Content Brief (follow these instructions closely)\n{{if .current_section.content_brief.purpose}}Brief Purpose: {{.current_section.content_brief.purpose}}\n{{end}}{{if .current_section.content_brief.tone_direction}}Brief Tone: {{.current_section.content_brief.tone_direction}}\n{{end}}{{if .current_section.content_brief.section_guidance}}Brief Guidance: {{.current_section.content_brief.section_guidance}}\n{{end}}{{end}}\n## Data Schema Required'
                     )::jsonb
WHERE type = 'page-content-writer'
  AND default_config::text NOT LIKE '%content_brief%';

-- STEP 3: Verify
SELECT
    type,
    default_config::text LIKE '%content_brief%' as has_brief_block
FROM agent_definitions
WHERE type = 'page-content-writer';


--

-- Change load_page_components to use the section plan's ready list
-- instead of the raw page sections
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_components,config,sections_from}',
        '"input_data.section_plan.ready_names"'
                     )
WHERE type = 'page-content-writer';

-- Verify
SELECT default_config->'workflow'->'steps'->'load_page_components'->'config'->>'sections_from'
FROM agent_definitions
WHERE type = 'page-content-writer';

---
-- double header problem

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        default_config,
                        '{workflow,steps,compile_page,config,inject_header}',
                        'false'
                ),
                '{workflow,steps,compile_page,config,inject_footer}',
                'false'
        ),
        '{workflow,steps,compile_page,config,inject_head}',
        'false'
                     )
WHERE type = 'page-content-writer';

---
-- single field for voice
-- ============================================================================
-- Simplify content writer template — one field instead of four
-- ============================================================================
-- The old template read four hardcoded fields:
--   voice, emphasis, avoid_phrases, social_proof_style
-- The new template reads one field that contains everything:
--   formatted
-- The Go action formats whatever the spec contains into readable text.
-- ============================================================================

DO $$
DECLARE
current_prompt text;
    new_prompt text;
    old_block text;
    new_block text;
BEGIN
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
INTO current_prompt
FROM agent_definitions
WHERE type = 'page-content-writer';

old_block := '## Content Direction (from site spec — follow this closely)
{{if .site_specs.specs.content_direction}}
Voice: {{.site_specs.specs.content_direction.voice}}
Emphasis: {{.site_specs.specs.content_direction.emphasis}}
Avoid these phrases: {{.site_specs.specs.content_direction.avoid_phrases}}
Social proof approach: {{.site_specs.specs.content_direction.social_proof_style}}
{{end}}';

    new_block := '## Content Direction (from site spec — follow this closely)
{{if .site_specs.specs.content_direction}}{{if .site_specs.specs.content_direction.formatted}}
{{.site_specs.specs.content_direction.formatted}}
{{end}}{{end}}';

    IF position(old_block in current_prompt) > 0 THEN
        new_prompt := replace(current_prompt, old_block, new_block);

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(new_prompt)
                     )
WHERE type = 'page-content-writer';

RAISE NOTICE 'Content writer updated: 4 fields replaced with 1 formatted field.';
ELSE
        RAISE NOTICE 'Old block not found — content writer may already be updated or the template has changed.';
END IF;
END $$;

---

-- 072 fix: target the actual nested prompt location in page-content-writer

BEGIN;

-- Step 1: Add rewrite_guidance to generate_content input_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}',
        '["current_section", "render_context", "reviewed_brief", "current_page", "link_context", "site_plan", "site_specs", "existing_content", "build_mode", "rewrite_guidance"]'::jsonb
                     )
WHERE type = 'page-content-writer';

-- Step 2: Add rewrite_guidance block to the prompt
DO $$
DECLARE
v_prompt text;
    v_new_prompt text;
BEGIN
SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
INTO v_prompt
FROM agent_definitions WHERE type = 'page-content-writer';

IF v_prompt IS NULL THEN
        RAISE NOTICE '072: prompt not found at expected path';
        RETURN;
END IF;

    IF v_prompt LIKE '%rewrite_guidance%' THEN
        RAISE NOTICE '072: rewrite_guidance already in prompt';
        RETURN;
END IF;

    v_new_prompt := replace(
        v_prompt,
        '## Section Requirements',
        E'{{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT — incorporate this into the content)\n{{.rewrite_guidance}}\n{{end}}\n\n## Section Requirements'
    );

    IF v_new_prompt = v_prompt THEN
        RAISE NOTICE '072: Could not find ## Section Requirements insertion point';
        RETURN;
END IF;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(v_new_prompt)
                     )
WHERE type = 'page-content-writer';

RAISE NOTICE '072: Prompt updated successfully';
END $$;

COMMIT;

-- Verify
SELECT
    default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}' LIKE '%rewrite_guidance%' as has_prompt_block,
    default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}' LIKE '%rewrite_guidance%' as has_input_field
FROM agent_definitions WHERE type = 'page-content-writer';
-- Expected: true, true


-----
-- major rewrite prompt and json workflow

-- ============================================================================
-- Step 3b: Update page-content-writer workflow
-- ============================================================================
--
-- BEFORE running this: take a snapshot for revert safety.
--
--   SELECT snapshot_agent('page-content-writer');
--   SELECT * FROM agent_snapshots WHERE type = 'page-content-writer';
--
-- AFTER running this: restart agent-chassis deployment if it caches definitions.
--
--   kubectl -n ai-persona-system rollout restart deployment/agent-chassis
--
-- IF SOMETHING IS WRONG: revert via the standard function.
--
--   SELECT revert_agent('page-content-writer');
--
-- ============================================================================
-- Workflow changes (vs. previous version):
--   1. Removed `load_page_components` step. Its job is now done upstream by
--      `plan_sections` (in page-build-handler), which populates the
--      `section_plan.sections_ready` array with full per-section data.
--
--   2. `prepare_link_context.next_step` redirected from `load_page_components`
--      to `build_render_context` (closing the gap left by removing step 1).
--
--   3. `process_sections_loop.iterate_over` changed from
--      `section_components.components` to `input_data.section_plan.sections_ready`.
--      Each iteration's `current_section` is now a sectionPlanItem carrying
--      `name`, `function`, `status`, `resolved_data`, `llm_fields`,
--      `llm_field_specs`, `component_id`, and the nested `component` map.
--
--   4. `render_section.config.component_from` changed from `current_section`
--      to `current_section.component` (where html_template, render_mode, etc.
--      now live). `render_from_template.config.component_from` likewise.
--
--   5. `render_section.config.merge_with` added (NEW): pointed at
--      `current_section.resolved_data`. RenderComponentAction merges this on
--      top of LLM output before render, so query-resolved items, static
--      fallbacks, and other authoritative data land in both the rendered
--      HTML and the persisted content_data. `render_from_template` gets the
--      same merge_with so its template-only path also benefits from
--      resolved_data overlay.
--
--   6. `check_render_mode` and `check_needs_research` conditionals updated
--      to read `current_section.component.render_mode`,
--      `current_section.component.needs_llm`,
--      `current_section.component.needs_research` (the fields moved into the
--      nested component map).
--
--   7. `generate_content.config.prompt_template` rewritten:
--      - The previous prompt dumped the full input_schema as text plus
--        ~100 lines of fallback JSON examples covering hero/feature/CTA/
--        text/contact/testimonial/case-study shapes. The LLM was asked to
--        pick the right shape and fill it in — leaving room for fabrication
--        of items, urls, and labels that should come from the database.
--      - The new prompt iterates `current_section.llm_field_specs` and asks
--        for ONLY the fields whose `source: llm` is declared in the schema.
--        Each field shown with type, required flag, and `llm_guidance` from
--        the component definition. The output format is a JSON object whose
--        example shape is rendered from the same llm_field_specs.
--      - All non-fabrication strict rules retained in full (rules 1-17).
--      - All upstream context blocks retained: Company Context, Contact
--        Info, Internal Linking, Content Direction, Rewrite Guidance, Admin
--        Brief, Research Findings, Existing Content for Recreate mode.
-- ============================================================================

BEGIN;

-- Confirm snapshot exists before applying (catch the case where snapshot was forgotten)
DO $check$
DECLARE
snap_count integer;
BEGIN
SELECT COUNT(*) INTO snap_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_snapshot = true
  AND deleted_at IS NULL;
IF snap_count = 0 THEN
        RAISE EXCEPTION 'No snapshot exists for page-content-writer. Run: SELECT snapshot_agent(''page-content-writer''); first.';
END IF;
    RAISE NOTICE 'Snapshot check OK: % snapshot row(s) exist for page-content-writer', snap_count;
END;
$check$;

-- Apply the workflow change
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        $JSON${
  "start_step": "spawn_research_agent",
        "timeout_seconds": 1200,
        "steps": {
    "spawn_research_agent": {
      "action": "spawn_agent",
        "config": {
        "role": "researcher",
        "agent_type": "research-agent",
        "await_response": true
    },
      "next_step": "load_site_specs",
      "description": "Spawn research agent in case sections need research",
      "output_field": "researcher_info"
    },
    "load_site_specs": {
      "action": "read_site_spec",
      "config": {
        "site_id": "input_data.site_record.site_id"
      },
      "next_step": "prepare_link_context",
      "error_step": "prepare_link_context",
      "description": "Load all site specs for content direction, identity, and design context",
      "output_field": "site_specs"
    },
    "prepare_link_context": {
      "action": "prepare_link_context",
      "config": {
        "enabled": true,
        "pages_field": "db_sync.pages",
        "max_links_per_section": 3
      },
      "next_step": "build_render_context",
      "description": "Prepare available pages context for internal linking",
      "output_field": "link_context"
    },
    "build_render_context": {
      "action": "build_render_context",
      "config": {
        "sources": {
          "page": "input_data.current_page",
          "plan": "input_data.site_plan",
          "site": "input_data.site_record",
          "brief": "input_data.reviewed_brief",
          "style": "input_data.style_collection",
          "assets": "brand_assets",
          "db_sync": "input_data.db_sync",
          "available_pages": "db_sync.pages"
        },
        "site_id_field": "input_data.site_record.site_id"
      },
      "next_step": "process_sections_loop",
      "description": "Build render context from brief, site, and brand data",
      "output_field": "render_context"
    },
    "process_sections_loop": {
      "action": "loop",
      "config": {
        "loop_var": "current_section",
        "iterate_over": "input_data.section_plan.sections_ready",
        "max_iterations": 15,
        "sub_workflow": {
          "start_step": "check_render_mode",
          "steps": {
            "check_render_mode": {
              "action": "conditional",
              "config": {
                "condition": "current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true",
                "then_step": "check_needs_research",
                "else_step": "render_from_template"
              },
              "description": "Check if section needs LLM or just template"
            },
            "check_needs_research": {
              "action": "conditional",
              "config": {
                "condition": "current_section.component.needs_research == true",
                "then_step": "call_researcher",
                "else_step": "generate_content"
              },
              "description": "Check if section needs research first"
            },
            "call_researcher": {
              "action": "call_agent",
              "config": {
                "agent_type": "research-agent",
                "target_role": "researcher",
                "input_mapping": {
                  "site_record": "input_data.site_record",
                  "reviewed_brief": "input_data.reviewed_brief",
                  "current_section": "current_section"
                },
                "timeout_seconds": 90
              },
              "next_step": "generate_content",
              "description": "Research topic for this section",
              "output_field": "research_result"
            },
            "generate_content": {
              "action": "execute_llm_prompt",
              "config": {
                "ai_service": {
                  "model": "claude-sonnet-4-6",
                  "provider": "anthropic",
                  "max_tokens": 2000,
                  "api_key_env_var": "ANTHROPIC_API_KEY"
                },
                "input_fields": [
                  "current_section",
                  "render_context",
                  "reviewed_brief",
                  "current_page",
                  "link_context",
                  "site_plan",
                  "site_specs",
                  "existing_content",
                  "build_mode",
                  "rewrite_guidance"
                ],
                "output_format": "json",
                "prompt_template": "Write content for the {{.current_section.name}} section of {{.current_page.title}}.\n\n## Language\nWrite all output content in the same language as the existing content, brief, and site specs in this prompt. Match the register, idioms, and conventions of that language.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Official Contact Information (USE ONLY THESE - DO NOT INVENT)\nEmail: {{.render_context.email}}\nPhone: {{.render_context.phone}}\nLocation: {{.reviewed_brief.headquarters}}\n\n{{if .link_context.link_constraint_text}}\n## Internal Linking\n{{.link_context.link_constraint_text}}\n\n{{end}}\n## Content Direction (from site spec \u2014 follow this closely)\n{{if .site_specs.specs.content_direction}}{{if .site_specs.specs.content_direction.formatted}}\n{{.site_specs.specs.content_direction.formatted}}\n{{end}}{{end}}\n{{if .site_specs.specs.identity.target_audience}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}\n{{end}}\n{{if .site_specs.specs.identity.key_differentiators}}\nKey Differentiators: {{.site_specs.specs.identity.key_differentiators}}\n{{end}}\n{{if .site_specs.specs.design_intent.imagery_direction}}\nImagery Direction: {{.site_specs.specs.design_intent.imagery_direction}}\n{{end}}\n\n{{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT \u2014 incorporate this into the content)\n{{.rewrite_guidance}}\n{{end}}\n\n{{if .current_section.component.content_brief}}## Admin Content Brief (follow these instructions closely)\n{{if .current_section.component.content_brief.purpose}}Brief Purpose: {{.current_section.component.content_brief.purpose}}\n{{end}}{{if .current_section.component.content_brief.tone_direction}}Brief Tone: {{.current_section.component.content_brief.tone_direction}}\n{{end}}{{if .current_section.component.content_brief.section_guidance}}Brief Guidance: {{.current_section.component.content_brief.section_guidance}}\n{{end}}{{end}}\n\n## What To Write\nWrite the following fields for the {{.current_section.name}} section. Each field's purpose is described in plain English \u2014 translate the intent into the output language as needed.\n\n{{range .current_section.llm_field_specs}}\n- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}\n{{end}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n{{if .existing_content}}{{if .existing_content.has_existing}}\n## EXISTING CONTENT \u2014 Recreate Mode\nThis page is being adopted from an existing site. Below is the original content from the page.\nYour task: find the content relevant to this section and adapt it to fit the writing fields listed above.\n\nPrioritise preserving the original meaning and information. Adapt the structure to match the required field names.\nIf the existing content does not have material relevant to this section, write fresh content as you normally would.\n\nOriginal page content:\n{{.existing_content.raw_markdown}}\n{{end}}{{end}}\n\n## Output Format\nReturn a JSON object with exactly these keys:\n\n{\n{{range $i, $f := .current_section.llm_field_specs}}{{if $i}},\n{{end}}  \"{{$f.name}}\": \"...\"{{end}}\n}\n\n## STRICT RULES:\n1. Use the EXACT field names shown in \"What To Write\" (these are technical identifiers \u2014 do not translate them).\n2. Return a JSON object with exactly the keys listed in \"What To Write\". Do not add any keys not in that list.\n3. Return ONLY the JSON object \u2014 no commentary, no markdown wrapper around the JSON.\n4. No placeholder text like [Your Company] or Lorem ipsum, in any language.\n5. Write content that is relevant to this company's industry and services \u2014 but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.\n6. Professional but engaging tone matching the brief.\n7. Include source citations [0], [1] if research was provided.\n8. NEVER invent contact information \u2014 use ONLY the email and phone provided in Official Contact Information above.\n9. For body text content, ALWAYS wrap paragraphs in <p> tags \u2014 never output raw unwrapped text.\n10. For rich_text or content fields that contain multiple paragraphs, use proper HTML: <p>Paragraph 1</p><p>Paragraph 2</p>.\n11. If contact email or phone is empty in the brief, do NOT make one up \u2014 omit it or write a generic contact-us link in the output language.\n12. Only create internal links to pages listed in the Internal Linking section above.\n13. NEVER invent fake people, client names, or attributed quotes. If the brief does not include real testimonials, write company philosophy statements instead and leave name/title fields empty.\n14. NEVER invent specific statistics, percentages, or metrics. Describe the type of outcome without fabricating numbers.\n15. NEVER invent fake case studies with named businesses presented as real clients. Describe service categories and typical outcomes instead.\n16. For testimonial sections: write 2-3 statements in the company's own voice about their values, approach, or commitment. These serve as placeholder content the site owner will replace with real testimonials.\n17. For case study sections: describe the types of problems the company solves and the approach they take, without inventing specific clients or results.\n18. It is ALWAYS better to be honest and general than specific and fabricated. A real visitor will trust a general statement of capability more than a fabricated testimonial.\n"
              },
              "next_step": "render_section",
              "description": "Generate content for this section's writing fields only",
              "output_field": "generated_content"
            },
            "render_section": {
              "action": "render_component",
              "config": {
                "output_html": true,
                "content_from": "generated_content.result",
                "merge_with": "current_section.resolved_data",
                "context_from": "render_context",
                "component_from": "current_section.component"
              },
              "description": "Render merged content into component template",
              "output_field": "section_output"
            },
            "render_from_template": {
              "action": "render_component",
              "config": {
                "output_html": true,
                "content_from": "render_context",
                "merge_with": "current_section.resolved_data",
                "context_from": "render_context",
                "component_from": "current_section.component"
              },
              "description": "Render section from template with resolved data (no LLM)",
              "output_field": "section_output"
            }
          }
        }
      },
      "next_step": "compile_page",
      "description": "Process each section \u2014 template render or LLM generate, then merge resolved data",
      "output_field": "processed_sections"
    },
    "compile_page": {
      "action": "compile_page_sections",
      "config": {
        "page_from": "input_data.current_page",
        "inject_head": false,
        "inject_footer": false,
        "inject_header": false,
        "sections_from": "processed_sections",
        "include_research_ids": true
      },
      "next_step": "complete",
      "description": "Compile all sections into page output",
      "output_field": "page_content"
    },
    "complete": {
      "action": "complete_workflow",
      "config": {
        "output_field": "page_content"
      }
    }
  }
}$JSON$::jsonb
),
    updated_at = now()
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Verify exactly one row was updated
DO $verify$
DECLARE
updated_count integer;
    step_count integer;
    has_load_step boolean;
    has_merge_with boolean;
BEGIN
SELECT COUNT(*) INTO updated_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);
IF updated_count != 1 THEN
        RAISE EXCEPTION 'Expected 1 active page-content-writer row, found %', updated_count;
END IF;

SELECT
    jsonb_object_keys(default_config->'workflow'->'steps') = ANY(ARRAY['load_page_components']),
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'render_section'->'config' ? 'merge_with',
    (SELECT COUNT(*) FROM jsonb_object_keys(default_config->'workflow'->'steps'))
INTO has_load_step, has_merge_with, step_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

IF has_load_step THEN
        RAISE EXCEPTION 'load_page_components step still present after update — workflow did not apply cleanly';
END IF;
    IF NOT has_merge_with THEN
        RAISE EXCEPTION 'merge_with config not present on render_section — workflow did not apply cleanly';
END IF;
    IF step_count != 7 THEN
        RAISE EXCEPTION 'Expected 7 top-level workflow steps, found %', step_count;
END IF;

    RAISE NOTICE 'Workflow updated successfully: 7 top-level steps, load_page_components removed, merge_with present';
END;
$verify$;

COMMIT;

-- Final inspection
SELECT
    type,
    is_active,
    updated_at,
    jsonb_object_keys(default_config->'workflow'->'steps') AS workflow_step
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

---

-- ============================================================================
-- Step 3b: Update page-content-writer workflow
-- ============================================================================
--
-- BEFORE running this: take a snapshot for revert safety.
--
--   SELECT snapshot_agent('page-content-writer');
--   SELECT * FROM agent_snapshots WHERE type = 'page-content-writer';
--
-- AFTER running this: restart agent-chassis deployment if it caches definitions.
--
--   kubectl -n ai-persona-system rollout restart deployment/agent-chassis
--
-- IF SOMETHING IS WRONG: revert via the standard function.
--
--   SELECT revert_agent('page-content-writer');
--
-- ============================================================================
-- Workflow changes (vs. previous version):
--   1. Removed `load_page_components` step. Its job is now done upstream by
--      `plan_sections` (in page-build-handler), which populates the
--      `section_plan.sections_ready` array with full per-section data.
--
--   2. `prepare_link_context.next_step` redirected from `load_page_components`
--      to `build_render_context` (closing the gap left by removing step 1).
--
--   3. `process_sections_loop.iterate_over` changed from
--      `section_components.components` to `input_data.section_plan.sections_ready`.
--      Each iteration's `current_section` is now a sectionPlanItem carrying
--      `name`, `function`, `status`, `resolved_data`, `llm_fields`,
--      `llm_field_specs`, `component_id`, and the nested `component` map.
--
--   4. `render_section.config.component_from` changed from `current_section`
--      to `current_section.component` (where html_template, render_mode, etc.
--      now live). `render_from_template.config.component_from` likewise.
--
--   5. `render_section.config.merge_with` added (NEW): pointed at
--      `current_section.resolved_data`. RenderComponentAction merges this on
--      top of LLM output before render, so query-resolved items, static
--      fallbacks, and other authoritative data land in both the rendered
--      HTML and the persisted content_data. `render_from_template` gets the
--      same merge_with so its template-only path also benefits from
--      resolved_data overlay.
--
--   6. `check_render_mode` and `check_needs_research` conditionals updated
--      to read `current_section.component.render_mode`,
--      `current_section.component.needs_llm`,
--      `current_section.component.needs_research` (the fields moved into the
--      nested component map).
--
--   7. `generate_content.config.prompt_template` rewritten:
--      - The previous prompt dumped the full input_schema as text plus
--        ~100 lines of fallback JSON examples covering hero/feature/CTA/
--        text/contact/testimonial/case-study shapes. The LLM was asked to
--        pick the right shape and fill it in — leaving room for fabrication
--        of items, urls, and labels that should come from the database.
--      - The new prompt iterates `current_section.llm_field_specs` and asks
--        for ONLY the fields whose `source: llm` is declared in the schema.
--        Each field shown with type, required flag, and `llm_guidance` from
--        the component definition. The output format is a JSON object whose
--        example shape is rendered from the same llm_field_specs.
--      - All non-fabrication strict rules retained in full (rules 1-17).
--      - All upstream context blocks retained: Company Context, Contact
--        Info, Internal Linking, Content Direction, Rewrite Guidance, Admin
--        Brief, Research Findings, Existing Content for Recreate mode.
-- ============================================================================

BEGIN;

-- Confirm snapshot exists before applying (catch the case where snapshot was forgotten)
DO $check$
DECLARE
snap_count integer;
BEGIN
SELECT COUNT(*) INTO snap_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_snapshot = true
  AND deleted_at IS NULL;
IF snap_count = 0 THEN
        RAISE EXCEPTION 'No snapshot exists for page-content-writer. Run: SELECT snapshot_agent(''page-content-writer''); first.';
END IF;
    RAISE NOTICE 'Snapshot check OK: % snapshot row(s) exist for page-content-writer', snap_count;
END;
$check$;

-- Apply the workflow change
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        $JSON${
  "start_step": "spawn_research_agent",
        "timeout_seconds": 1200,
        "steps": {
    "spawn_research_agent": {
      "action": "spawn_agent",
        "config": {
        "role": "researcher",
        "agent_type": "research-agent",
        "await_response": true
    },
      "next_step": "load_site_specs",
      "description": "Spawn research agent in case sections need research",
      "output_field": "researcher_info"
    },
    "load_site_specs": {
      "action": "read_site_spec",
      "config": {
        "site_id": "input_data.site_record.site_id"
      },
      "next_step": "prepare_link_context",
      "error_step": "prepare_link_context",
      "description": "Load all site specs for content direction, identity, and design context",
      "output_field": "site_specs"
    },
    "prepare_link_context": {
      "action": "prepare_link_context",
      "config": {
        "enabled": true,
        "pages_field": "db_sync.pages",
        "max_links_per_section": 3
      },
      "next_step": "build_render_context",
      "description": "Prepare available pages context for internal linking",
      "output_field": "link_context"
    },
    "build_render_context": {
      "action": "build_render_context",
      "config": {
        "sources": {
          "page": "input_data.current_page",
          "plan": "input_data.site_plan",
          "site": "input_data.site_record",
          "brief": "input_data.reviewed_brief",
          "style": "input_data.style_collection",
          "assets": "brand_assets",
          "db_sync": "input_data.db_sync",
          "available_pages": "db_sync.pages"
        },
        "site_id_field": "input_data.site_record.site_id"
      },
      "next_step": "process_sections_loop",
      "description": "Build render context from brief, site, and brand data",
      "output_field": "render_context"
    },
    "process_sections_loop": {
      "action": "loop",
      "config": {
        "loop_var": "current_section",
        "iterate_over": "input_data.section_plan.sections_ready",
        "max_iterations": 15,
        "sub_workflow": {
          "start_step": "check_render_mode",
          "steps": {
            "check_render_mode": {
              "action": "conditional",
              "config": {
                "condition": "current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true",
                "then_step": "check_needs_research",
                "else_step": "render_from_template"
              },
              "description": "Check if section needs LLM or just template"
            },
            "check_needs_research": {
              "action": "conditional",
              "config": {
                "condition": "current_section.component.needs_research == true",
                "then_step": "call_researcher",
                "else_step": "generate_content"
              },
              "description": "Check if section needs research first"
            },
            "call_researcher": {
              "action": "call_agent",
              "config": {
                "agent_type": "research-agent",
                "target_role": "researcher",
                "input_mapping": {
                  "site_record": "input_data.site_record",
                  "reviewed_brief": "input_data.reviewed_brief",
                  "current_section": "current_section"
                },
                "timeout_seconds": 90
              },
              "next_step": "generate_content",
              "description": "Research topic for this section",
              "output_field": "research_result"
            },
            "generate_content": {
              "action": "execute_llm_prompt",
              "config": {
                "ai_service": {
                  "model": "claude-sonnet-4-6",
                  "provider": "anthropic",
                  "max_tokens": 2000,
                  "api_key_env_var": "ANTHROPIC_API_KEY"
                },
                "input_fields": [
                  "current_section",
                  "render_context",
                  "reviewed_brief",
                  "current_page",
                  "link_context",
                  "site_plan",
                  "site_specs",
                  "existing_content",
                  "build_mode",
                  "rewrite_guidance"
                ],
                "output_format": "json",
                "prompt_template": "Write content for the {{.current_section.name}} section of {{.current_page.title}}.\n\n## Language\nWrite all output content in the same language as the existing content, brief, and site specs in this prompt. Match the register, idioms, and conventions of that language.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Official Contact Information (USE ONLY THESE - DO NOT INVENT)\nEmail: {{.render_context.email}}\nPhone: {{.render_context.phone}}\nLocation: {{.reviewed_brief.headquarters}}\n\n{{if .link_context.link_constraint_text}}\n## Internal Linking\n{{.link_context.link_constraint_text}}\n\n{{end}}\n## Content Direction (from site spec \u2014 follow this closely)\n{{if .site_specs.specs.content_direction}}{{if .site_specs.specs.content_direction.formatted}}\n{{.site_specs.specs.content_direction.formatted}}\n{{end}}{{end}}\n{{if .site_specs.specs.identity.target_audience}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}\n{{end}}\n{{if .site_specs.specs.identity.key_differentiators}}\nKey Differentiators: {{.site_specs.specs.identity.key_differentiators}}\n{{end}}\n{{if .site_specs.specs.design_intent.imagery_direction}}\nImagery Direction: {{.site_specs.specs.design_intent.imagery_direction}}\n{{end}}\n\n{{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT \u2014 incorporate this into the content)\n{{.rewrite_guidance}}\n{{end}}\n\n{{if .current_section.component.content_brief}}## Admin Content Brief (follow these instructions closely)\n{{if .current_section.component.content_brief.purpose}}Brief Purpose: {{.current_section.component.content_brief.purpose}}\n{{end}}{{if .current_section.component.content_brief.tone_direction}}Brief Tone: {{.current_section.component.content_brief.tone_direction}}\n{{end}}{{if .current_section.component.content_brief.section_guidance}}Brief Guidance: {{.current_section.component.content_brief.section_guidance}}\n{{end}}{{end}}\n\n## What To Write\nWrite the following fields for the {{.current_section.name}} section. Each field's purpose is described in plain English \u2014 translate the intent into the output language as needed.\n\n{{range .current_section.llm_field_specs}}\n- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}\n{{end}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n{{if .existing_content}}{{if .existing_content.has_existing}}\n## EXISTING CONTENT \u2014 Recreate Mode\nThis page is being adopted from an existing site. Below is the original content from the page.\nYour task: find the content relevant to this section and adapt it to fit the writing fields listed above.\n\nPrioritise preserving the original meaning and information. Adapt the structure to match the required field names.\nIf the existing content does not have material relevant to this section, write fresh content as you normally would.\n\nOriginal page content:\n{{.existing_content.raw_markdown}}\n{{end}}{{end}}\n\n## Output Format\nReturn a JSON object with exactly these keys:\n\n{\n{{range $i, $f := .current_section.llm_field_specs}}{{if $i}},\n{{end}}  \"{{$f.name}}\": \"...\"{{end}}\n}\n\n## STRICT RULES:\n1. Use the EXACT field names shown in \"What To Write\" (these are technical identifiers \u2014 do not translate them).\n2. Return a JSON object with exactly the keys listed in \"What To Write\". Do not add any keys not in that list.\n3. Return ONLY the JSON object \u2014 no commentary, no markdown wrapper around the JSON.\n4. No placeholder text like [Your Company] or Lorem ipsum, in any language.\n5. Write content that is relevant to this company's industry and services \u2014 but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.\n6. Professional but engaging tone matching the brief.\n7. Include source citations [0], [1] if research was provided.\n8. NEVER invent contact information \u2014 use ONLY the email and phone provided in Official Contact Information above.\n9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.\n10. For fields of type `rich_text` or `content` that contain multiple paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags: <p>Paragraph 1</p><p>Paragraph 2</p>.\n11. If contact email or phone is empty in the brief, do NOT make one up \u2014 omit it or write a generic contact-us link in the output language.\n12. Only create internal links to pages listed in the Internal Linking section above.\n13. NEVER invent fake people, client names, or attributed quotes. If the brief does not include real testimonials, write company philosophy statements instead and leave name/title fields empty.\n14. NEVER invent specific statistics, percentages, or metrics. Describe the type of outcome without fabricating numbers.\n15. NEVER invent fake case studies with named businesses presented as real clients. Describe service categories and typical outcomes instead.\n16. For testimonial sections: write 2-3 statements in the company's own voice about their values, approach, or commitment. These serve as placeholder content the site owner will replace with real testimonials.\n17. For case study sections: describe the types of problems the company solves and the approach they take, without inventing specific clients or results.\n18. It is ALWAYS better to be honest and general than specific and fabricated. A real visitor will trust a general statement of capability more than a fabricated testimonial.\n"
              },
              "next_step": "render_section",
              "description": "Generate content for this section's writing fields only",
              "output_field": "generated_content"
            },
            "render_section": {
              "action": "render_component",
              "config": {
                "output_html": true,
                "content_from": "generated_content.result",
                "merge_with": "current_section.resolved_data",
                "context_from": "render_context",
                "component_from": "current_section.component"
              },
              "description": "Render merged content into component template",
              "output_field": "section_output"
            },
            "render_from_template": {
              "action": "render_component",
              "config": {
                "output_html": true,
                "content_from": "render_context",
                "merge_with": "current_section.resolved_data",
                "context_from": "render_context",
                "component_from": "current_section.component"
              },
              "description": "Render section from template with resolved data (no LLM)",
              "output_field": "section_output"
            }
          }
        }
      },
      "next_step": "compile_page",
      "description": "Process each section \u2014 template render or LLM generate, then merge resolved data",
      "output_field": "processed_sections"
    },
    "compile_page": {
      "action": "compile_page_sections",
      "config": {
        "page_from": "input_data.current_page",
        "inject_head": false,
        "inject_footer": false,
        "inject_header": false,
        "sections_from": "processed_sections",
        "include_research_ids": true
      },
      "next_step": "complete",
      "description": "Compile all sections into page output",
      "output_field": "page_content"
    },
    "complete": {
      "action": "complete_workflow",
      "config": {
        "output_field": "page_content"
      }
    }
  }
}$JSON$::jsonb
),
    updated_at = now()
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Verify exactly one row was updated
DO $verify$
DECLARE
updated_count integer;
    step_count integer;
    has_load_step boolean;
    has_merge_with boolean;
BEGIN
SELECT COUNT(*) INTO updated_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);
IF updated_count != 1 THEN
        RAISE EXCEPTION 'Expected 1 active page-content-writer row, found %', updated_count;
END IF;

SELECT
    jsonb_object_keys(default_config->'workflow'->'steps') = ANY(ARRAY['load_page_components']),
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'render_section'->'config' ? 'merge_with',
    (SELECT COUNT(*) FROM jsonb_object_keys(default_config->'workflow'->'steps'))
INTO has_load_step, has_merge_with, step_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

IF has_load_step THEN
        RAISE EXCEPTION 'load_page_components step still present after update — workflow did not apply cleanly';
END IF;
    IF NOT has_merge_with THEN
        RAISE EXCEPTION 'merge_with config not present on render_section — workflow did not apply cleanly';
END IF;
    IF step_count != 7 THEN
        RAISE EXCEPTION 'Expected 7 top-level workflow steps, found %', step_count;
END IF;

    RAISE NOTICE 'Workflow updated successfully: 7 top-level steps, load_page_components removed, merge_with present';
END;
$verify$;

COMMIT;

-- Final inspection
SELECT
    type,
    is_active,
    updated_at,
    jsonb_object_keys(default_config->'workflow'->'steps') AS workflow_step
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

---


-- double p fix

-- ============================================================================
-- Step 3b: Update page-content-writer workflow
-- ============================================================================
--
-- BEFORE running this: take a snapshot for revert safety.
--
--   SELECT snapshot_agent('page-content-writer');
--   SELECT * FROM agent_snapshots WHERE type = 'page-content-writer';
--
-- AFTER running this: restart agent-chassis deployment if it caches definitions.
--
--   kubectl -n ai-persona-system rollout restart deployment/agent-chassis
--
-- IF SOMETHING IS WRONG: revert via the standard function.
--
--   SELECT revert_agent('page-content-writer');
--
-- ============================================================================
-- Workflow changes (vs. previous version):
--   1. Removed `load_page_components` step. Its job is now done upstream by
--      `plan_sections` (in page-build-handler), which populates the
--      `section_plan.sections_ready` array with full per-section data.
--
--   2. `prepare_link_context.next_step` redirected from `load_page_components`
--      to `build_render_context` (closing the gap left by removing step 1).
--
--   3. `process_sections_loop.iterate_over` changed from
--      `section_components.components` to `input_data.section_plan.sections_ready`.
--      Each iteration's `current_section` is now a sectionPlanItem carrying
--      `name`, `function`, `status`, `resolved_data`, `llm_fields`,
--      `llm_field_specs`, `component_id`, and the nested `component` map.
--
--   4. `render_section.config.component_from` changed from `current_section`
--      to `current_section.component` (where html_template, render_mode, etc.
--      now live). `render_from_template.config.component_from` likewise.
--
--   5. `render_section.config.merge_with` added (NEW): pointed at
--      `current_section.resolved_data`. RenderComponentAction merges this on
--      top of LLM output before render, so query-resolved items, static
--      fallbacks, and other authoritative data land in both the rendered
--      HTML and the persisted content_data. `render_from_template` gets the
--      same merge_with so its template-only path also benefits from
--      resolved_data overlay.
--
--   6. `check_render_mode` and `check_needs_research` conditionals updated
--      to read `current_section.component.render_mode`,
--      `current_section.component.needs_llm`,
--      `current_section.component.needs_research` (the fields moved into the
--      nested component map).
--
--   7. `generate_content.config.prompt_template` rewritten:
--      - The previous prompt dumped the full input_schema as text plus
--        ~100 lines of fallback JSON examples covering hero/feature/CTA/
--        text/contact/testimonial/case-study shapes. The LLM was asked to
--        pick the right shape and fill it in — leaving room for fabrication
--        of items, urls, and labels that should come from the database.
--      - The new prompt iterates `current_section.llm_field_specs` and asks
--        for ONLY the fields whose `source: llm` is declared in the schema.
--        Each field shown with type, required flag, and `llm_guidance` from
--        the component definition. The output format is a JSON object whose
--        example shape is rendered from the same llm_field_specs.
--      - All non-fabrication strict rules retained in full (rules 1-17).
--      - All upstream context blocks retained: Company Context, Contact
--        Info, Internal Linking, Content Direction, Rewrite Guidance, Admin
--        Brief, Research Findings, Existing Content for Recreate mode.
-- ============================================================================

BEGIN;

-- Confirm snapshot exists before applying (catch the case where snapshot was forgotten)
DO $check$
DECLARE
snap_count integer;
BEGIN
SELECT COUNT(*) INTO snap_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_snapshot = true
  AND deleted_at IS NULL;
IF snap_count = 0 THEN
        RAISE EXCEPTION 'No snapshot exists for page-content-writer. Run: SELECT snapshot_agent(''page-content-writer''); first.';
END IF;
    RAISE NOTICE 'Snapshot check OK: % snapshot row(s) exist for page-content-writer', snap_count;
END;
$check$;

-- Apply the workflow change
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        $JSON${
  "start_step": "spawn_research_agent",
        "timeout_seconds": 1200,
        "steps": {
    "spawn_research_agent": {
      "action": "spawn_agent",
        "config": {
        "role": "researcher",
        "agent_type": "research-agent",
        "await_response": true
    },
      "next_step": "load_site_specs",
      "description": "Spawn research agent in case sections need research",
      "output_field": "researcher_info"
    },
    "load_site_specs": {
      "action": "read_site_spec",
      "config": {
        "site_id": "input_data.site_record.site_id"
      },
      "next_step": "prepare_link_context",
      "error_step": "prepare_link_context",
      "description": "Load all site specs for content direction, identity, and design context",
      "output_field": "site_specs"
    },
    "prepare_link_context": {
      "action": "prepare_link_context",
      "config": {
        "enabled": true,
        "pages_field": "db_sync.pages",
        "max_links_per_section": 3
      },
      "next_step": "build_render_context",
      "description": "Prepare available pages context for internal linking",
      "output_field": "link_context"
    },
    "build_render_context": {
      "action": "build_render_context",
      "config": {
        "sources": {
          "page": "input_data.current_page",
          "plan": "input_data.site_plan",
          "site": "input_data.site_record",
          "brief": "input_data.reviewed_brief",
          "style": "input_data.style_collection",
          "assets": "brand_assets",
          "db_sync": "input_data.db_sync",
          "available_pages": "db_sync.pages"
        },
        "site_id_field": "input_data.site_record.site_id"
      },
      "next_step": "process_sections_loop",
      "description": "Build render context from brief, site, and brand data",
      "output_field": "render_context"
    },
    "process_sections_loop": {
      "action": "loop",
      "config": {
        "loop_var": "current_section",
        "iterate_over": "input_data.section_plan.sections_ready",
        "max_iterations": 15,
        "sub_workflow": {
          "start_step": "check_render_mode",
          "steps": {
            "check_render_mode": {
              "action": "conditional",
              "config": {
                "condition": "current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true",
                "then_step": "check_needs_research",
                "else_step": "render_from_template"
              },
              "description": "Check if section needs LLM or just template"
            },
            "check_needs_research": {
              "action": "conditional",
              "config": {
                "condition": "current_section.component.needs_research == true",
                "then_step": "call_researcher",
                "else_step": "generate_content"
              },
              "description": "Check if section needs research first"
            },
            "call_researcher": {
              "action": "call_agent",
              "config": {
                "agent_type": "research-agent",
                "target_role": "researcher",
                "input_mapping": {
                  "site_record": "input_data.site_record",
                  "reviewed_brief": "input_data.reviewed_brief",
                  "current_section": "current_section"
                },
                "timeout_seconds": 90
              },
              "next_step": "generate_content",
              "description": "Research topic for this section",
              "output_field": "research_result"
            },
            "generate_content": {
              "action": "execute_llm_prompt",
              "config": {
                "ai_service": {
                  "model": "claude-sonnet-4-6",
                  "provider": "anthropic",
                  "max_tokens": 2000,
                  "api_key_env_var": "ANTHROPIC_API_KEY"
                },
                "input_fields": [
                  "current_section",
                  "render_context",
                  "reviewed_brief",
                  "current_page",
                  "link_context",
                  "site_plan",
                  "site_specs",
                  "existing_content",
                  "build_mode",
                  "rewrite_guidance"
                ],
                "output_format": "json",
                "prompt_template": "Write content for the {{.current_section.name}} section of {{.current_page.title}}.\n\n## Language\nWrite all output content in the same language as the existing content, brief, and site specs in this prompt. Match the register, idioms, and conventions of that language.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Official Contact Information (USE ONLY THESE - DO NOT INVENT)\nEmail: {{.render_context.email}}\nPhone: {{.render_context.phone}}\nLocation: {{.reviewed_brief.headquarters}}\n\n{{if .link_context.link_constraint_text}}\n## Internal Linking\n{{.link_context.link_constraint_text}}\n\n{{end}}\n## Content Direction (from site spec \u2014 follow this closely)\n{{if .site_specs.specs.content_direction}}{{if .site_specs.specs.content_direction.formatted}}\n{{.site_specs.specs.content_direction.formatted}}\n{{end}}{{end}}\n{{if .site_specs.specs.identity.target_audience}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}\n{{end}}\n{{if .site_specs.specs.identity.key_differentiators}}\nKey Differentiators: {{.site_specs.specs.identity.key_differentiators}}\n{{end}}\n{{if .site_specs.specs.design_intent.imagery_direction}}\nImagery Direction: {{.site_specs.specs.design_intent.imagery_direction}}\n{{end}}\n\n{{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT \u2014 incorporate this into the content)\n{{.rewrite_guidance}}\n{{end}}\n\n{{if .current_section.component.content_brief}}## Admin Content Brief (follow these instructions closely)\n{{if .current_section.component.content_brief.purpose}}Brief Purpose: {{.current_section.component.content_brief.purpose}}\n{{end}}{{if .current_section.component.content_brief.tone_direction}}Brief Tone: {{.current_section.component.content_brief.tone_direction}}\n{{end}}{{if .current_section.component.content_brief.section_guidance}}Brief Guidance: {{.current_section.component.content_brief.section_guidance}}\n{{end}}{{end}}\n\n## What To Write\nWrite the following fields for the {{.current_section.name}} section. Each field's purpose is described in plain English \u2014 translate the intent into the output language as needed.\n\n{{range .current_section.llm_field_specs}}\n- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}\n{{end}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n{{if .existing_content}}{{if .existing_content.has_existing}}\n## EXISTING CONTENT \u2014 Recreate Mode\nThis page is being adopted from an existing site. Below is the original content from the page.\nYour task: find the content relevant to this section and adapt it to fit the writing fields listed above.\n\nPrioritise preserving the original meaning and information. Adapt the structure to match the required field names.\nIf the existing content does not have material relevant to this section, write fresh content as you normally would.\n\nOriginal page content:\n{{.existing_content.raw_markdown}}\n{{end}}{{end}}\n\n## Output Format\nReturn a JSON object with exactly these keys:\n\n{\n{{range $i, $f := .current_section.llm_field_specs}}{{if $i}},\n{{end}}  \"{{$f.name}}\": \"...\"{{end}}\n}\n\n## STRICT RULES:\n1. Use the EXACT field names shown in \"What To Write\" (these are technical identifiers \u2014 do not translate them).\n2. Return a JSON object with exactly the keys listed in \"What To Write\". Do not add any keys not in that list.\n3. Return ONLY the JSON object \u2014 no commentary, no markdown wrapper around the JSON.\n4. No placeholder text like [Your Company] or Lorem ipsum, in any language.\n5. Write content that is relevant to this company's industry and services \u2014 but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.\n6. Professional but engaging tone matching the brief.\n7. Include source citations [0], [1] if research was provided.\n8. NEVER invent contact information \u2014 use ONLY the email and phone provided in Official Contact Information above.\n9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.\n10. For fields of type `rich_text` or `content` that contain multiple paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags: <p>Paragraph 1</p><p>Paragraph 2</p>.\n11. If contact email or phone is empty in the brief, do NOT make one up \u2014 omit it or write a generic contact-us link in the output language.\n12. Only create internal links to pages listed in the Internal Linking section above.\n13. NEVER invent fake people, client names, or attributed quotes. If the brief does not include real testimonials, write company philosophy statements instead and leave name/title fields empty.\n14. NEVER invent specific statistics, percentages, or metrics. Describe the type of outcome without fabricating numbers.\n15. NEVER invent fake case studies with named businesses presented as real clients. Describe service categories and typical outcomes instead.\n16. For testimonial sections: write 2-3 statements in the company's own voice about their values, approach, or commitment. These serve as placeholder content the site owner will replace with real testimonials.\n17. For case study sections: describe the types of problems the company solves and the approach they take, without inventing specific clients or results.\n18. It is ALWAYS better to be honest and general than specific and fabricated. A real visitor will trust a general statement of capability more than a fabricated testimonial.\n"
              },
              "next_step": "render_section",
              "description": "Generate content for this section's writing fields only",
              "output_field": "generated_content"
            },
            "render_section": {
              "action": "render_component",
              "config": {
                "output_html": true,
                "content_from": "generated_content.result",
                "merge_with": "current_section.resolved_data",
                "context_from": "render_context",
                "component_from": "current_section.component"
              },
              "description": "Render merged content into component template",
              "output_field": "section_output"
            },
            "render_from_template": {
              "action": "render_component",
              "config": {
                "output_html": true,
                "content_from": "render_context",
                "merge_with": "current_section.resolved_data",
                "context_from": "render_context",
                "component_from": "current_section.component"
              },
              "description": "Render section from template with resolved data (no LLM)",
              "output_field": "section_output"
            }
          }
        }
      },
      "next_step": "compile_page",
      "description": "Process each section \u2014 template render or LLM generate, then merge resolved data",
      "output_field": "processed_sections"
    },
    "compile_page": {
      "action": "compile_page_sections",
      "config": {
        "page_from": "input_data.current_page",
        "inject_head": false,
        "inject_footer": false,
        "inject_header": false,
        "sections_from": "processed_sections",
        "include_research_ids": true
      },
      "next_step": "complete",
      "description": "Compile all sections into page output",
      "output_field": "page_content"
    },
    "complete": {
      "action": "complete_workflow",
      "config": {
        "output_field": "page_content"
      }
    }
  }
}$JSON$::jsonb
),
    updated_at = now()
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Verify exactly one row was updated
DO $verify$
DECLARE
updated_count integer;
    step_count integer;
    has_load_step boolean;
    has_merge_with boolean;
BEGIN
SELECT COUNT(*) INTO updated_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);
IF updated_count != 1 THEN
        RAISE EXCEPTION 'Expected 1 active page-content-writer row, found %', updated_count;
END IF;

SELECT
    jsonb_object_keys(default_config->'workflow'->'steps') = ANY(ARRAY['load_page_components']),
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'render_section'->'config' ? 'merge_with',
    (SELECT COUNT(*) FROM jsonb_object_keys(default_config->'workflow'->'steps'))
INTO has_load_step, has_merge_with, step_count
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

IF has_load_step THEN
        RAISE EXCEPTION 'load_page_components step still present after update — workflow did not apply cleanly';
END IF;
    IF NOT has_merge_with THEN
        RAISE EXCEPTION 'merge_with config not present on render_section — workflow did not apply cleanly';
END IF;
    IF step_count != 7 THEN
        RAISE EXCEPTION 'Expected 7 top-level workflow steps, found %', step_count;
END IF;

    RAISE NOTICE 'Workflow updated successfully: 7 top-level steps, load_page_components removed, merge_with present';
END;
$verify$;

COMMIT;

-- Final inspection
SELECT
    type,
    is_active,
    updated_at,
    jsonb_object_keys(default_config->'workflow'->'steps') AS workflow_step
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

--

-- prompt change for item keys to unbug item_key, item_type overlap
--  the ## What To Write and ## Output Format blocks should now reference item_fields:

clients_db=# SELECT snapshot_agent('page-content-writer', 'item_key, item_type unbug');
NOTICE:  Snapshot captured: type=page-content-writer, source_version=2, source_id=5946a27b-38ab-41e8-8b49-7bc1a4b626b8, reason=item_key, item_type unbug
snapshot_agent
5946a27b-38ab-41e8-8b49-7bc1a4b626b8

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
                replace(
                        replace(
                                default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
                                $old_wtw$- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}$old_wtw$,
        $new_wtw$- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}{{if .item_fields}} — each item is an object with exactly these fields: {{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}$new_wtw$
      ),
                        $old_out$  "{{$f.name}}": "..."$old_out$,
                        $new_out$  "{{$f.name}}": {{if $f.item_fields}}[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}"{{$k}}": "..."{{end}} }]{{else}}"..."{{end}}$new_out$
    )
        ),
        false
                     ),
    updated_at = now()
WHERE type = 'page-content-writer';

SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
FROM agent_definitions WHERE type = 'page-content-writer';

clients_db=# SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
             FROM agent_definitions WHERE type = 'page-content-writer';
?column?
Write content for the {{.current_section.name}} section of {{.current_page.title}}.

## Language
Write all output content in the same language as the existing content, brief, and site specs in this prompt. Match the register, idioms, and conventions of that language.

## Company Context
Company: {{.render_context.company_name}}
Industry: {{.render_context.industry}}
Tone: {{.render_context.tone}}
Target Audience: {{.render_context.target_audience}}
Services: {{.reviewed_brief.services}}
Tagline: {{.render_context.tagline}}

## Official Contact Information (USE ONLY THESE - DO NOT INVENT)
Email: {{.render_context.email}}
Phone: {{.render_context.phone}}
Location: {{.reviewed_brief.headquarters}}

{{if .link_context.link_constraint_text}}
## Internal Linking
{{.link_context.link_constraint_text}}

{{end}}
## Content Direction (from site spec — follow this closely)
{{if .site_specs.specs.content_direction}}{{if .site_specs.specs.content_direction.formatted}}
{{.site_specs.specs.content_direction.formatted}}
{{end}}{{end}}
{{if .site_specs.specs.identity.target_audience}}
Target Audience: {{.site_specs.specs.identity.target_audience}}
{{end}}
{{if .site_specs.specs.identity.key_differentiators}}
Key Differentiators: {{.site_specs.specs.identity.key_differentiators}}
{{end}}
{{if .site_specs.specs.design_intent.imagery_direction}}
Imagery Direction: {{.site_specs.specs.design_intent.imagery_direction}}
{{end}}

{{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT — incorporate this into the content)
{{.rewrite_guidance}}
{{end}}

{{if .current_section.component.content_brief}}## Admin Content Brief (follow these instructions closely)
{{if .current_section.component.content_brief.purpose}}Brief Purpose: {{.current_section.component.content_brief.purpose}}
{{end}}{{if .current_section.component.content_brief.tone_direction}}Brief Tone: {{.current_section.component.content_brief.tone_direction}}
{{end}}{{if .current_section.component.content_brief.section_guidance}}Brief Guidance: {{.current_section.component.content_brief.section_guidance}}
{{end}}{{end}}

## What To Write
Write the following fields for the {{.current_section.name}} section. Each field's purpose is described in plain English — translate the intent into the output language as needed.

{{range .current_section.llm_field_specs}}
- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}{{if .item_fields}} — each item is an object with exactly these fields: {{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}
{{end}}

{{if .research_result}}
## Research Findings
{{.research_result.response.summary}}

Sources:
{{range $index, $src := .research_result.response.sources}}
- [{{$index}}] {{$src.title}} ({{$src.domain}})
{{end}}
{{end}}

{{if .existing_content}}{{if .existing_content.has_existing}}
## EXISTING CONTENT — Recreate Mode
This page is being adopted from an existing site. Below is the original content from the page.
Your task: find the content relevant to this section and adapt it to fit the writing fields listed above.

Prioritise preserving the original meaning and information. Adapt the structure to match the required field names.
If the existing content does not have material relevant to this section, write fresh content as you normally would.

Original page content:
{{.existing_content.raw_markdown}}
{{end}}{{end}}

## Output Format
Return a JSON object with exactly these keys:

{
{{range $i, $f := .current_section.llm_field_specs}}{{if $i}},
{{end}}  "{{$f.name}}": {{if $f.item_fields}}[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}"{{$k}}": "..."{{end}} }]{{else}}"..."{{end}}{{end}}
}

## STRICT RULES:
1. Use the EXACT field names shown in "What To Write" (these are technical identifiers — do not translate them).
2. Return a JSON object with exactly the keys listed in "What To Write". Do not add any keys not in that list.
3. Return ONLY the JSON object — no commentary, no markdown wrapper around the JSON.
4. No placeholder text like [Your Company] or Lorem ipsum, in any language.
5. Write content that is relevant to this company's industry and services — but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.
6. Professional but engaging tone matching the brief.
7. Include source citations [0], [1] if research was provided.
8. NEVER invent contact information — use ONLY the email and phone provided in Official Contact Information above.
9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.
10. For fields of type `rich_text` or `content` that contain multiple paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags: <p>Paragraph 1</p><p>Paragraph 2</p>.
11. If contact email or phone is empty in the brief, do NOT make one up — omit it or write a generic contact-us link in the output language.
12. Only create internal links to pages listed in the Internal Linking section above.
13. NEVER invent fake people, client names, or attributed quotes. If the brief does not include real testimonials, write company philosophy statements instead and leave name/title fields empty.
14. NEVER invent specific statistics, percentages, or metrics. Describe the type of outcome without fabricating numbers.
15. NEVER invent fake case studies with named businesses presented as real clients. Describe service categories and typical outcomes instead.
16. For testimonial sections: write 2-3 statements in the company's own voice about their values, approach, or commitment. These serve as placeholder content the site owner will replace with real testimonials.
17. For case study sections: describe the types of problems the company solves and the approach they take, without inventing specific clients or results.
18. It is ALWAYS better to be honest and general than specific and fabricated. A real visitor will trust a general statement of capability more than a fabricated testimonial.

(1 row)
clients_db=#