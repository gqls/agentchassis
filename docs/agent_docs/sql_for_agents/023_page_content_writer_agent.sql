
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