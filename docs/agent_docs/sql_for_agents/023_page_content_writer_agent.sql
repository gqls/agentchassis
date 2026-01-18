latest:
      page-content-writer Analysis
Workflow Flow:

spawn_research_agent → spawns researcher for potential use
load_page_components → loads component definitions for the page's sections
build_render_context → builds context from brief, site, and brand data
process_sections_loop → iterates over each section component
compile_page → compiles all sections into final page output
complete

Input Contract (expects):

current_page - object with name, title, sections[] (required)
site_record - object with site_id, domain (required)
reviewed_brief - object with company_name, services, about_us, etc (required)
style_collection - object with colors, component refs (optional)
brand_assets - object with logo, images (optional)

Output Contract (produces):

sections - array of {component_id, rendered_html, content_data}
page_name - string
research_ids - array of research result UUIDs used

Data Access Pattern:
The config paths use input_data.* prefix:

page_from: "input_data.current_page"
sections_from: "input_data.current_page.sections"
site: "input_data.site_record"
brief: "input_data.reviewed_brief"
' 
----------------

-- ===========================================================================
-- PAGE CONTENT WRITER AGENT
-- File: 045_page_content_writer_agent.sql
-- ===========================================================================
-- Writes content for a single page, section by section.
-- Uses templates for structure, LLM only for content that needs writing.
-- Can spawn research-agent for sections needing research.
-- ===========================================================================

BEGIN;

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    is_active,
    status,
    version,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    input_contract,
    output_contract,
    default_config
) VALUES (
             'page-content-writer',
             'Page Content Writer',
             'Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content.',
             'specialist',
             true,
             'active',
             1,
             '["content-generation", "template-rendering", "research"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.575',
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "error": "system.errors.{type}",
                 "process": "system.agent.{type}.process",
                 "response": "system.responses.{type}"
             }'::jsonb,
             '{
                 "port": 8080,
                 "liveness_path": "/health",
                 "readiness_path": "/ready",
                 "initial_delay_seconds": 15
             }'::jsonb,
             -- Input contract
             '{
                 "expects": {
                     "current_page": "object with name, title, sections[]",
                     "site_record": "object with site_id, domain",
                     "reviewed_brief": "object with company_name, services, about_us, etc",
                     "style_collection": "object with colors, component refs",
                     "brand_assets": "object with logo, images (optional)"
                 },
                 "required": ["current_page", "site_record", "reviewed_brief"]
             }'::jsonb,
             -- Output contract
             '{
                 "produces": {
                     "page_name": "string",
                     "sections": "array of {component_id, rendered_html, content_data}",
                     "research_ids": "array of research result UUIDs used"
                 }
             }'::jsonb,
             -- Workflow
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 180,
                 "workflow": {
                     "start_step": "spawn_research_agent",
                     "steps": {
                         "spawn_research_agent": {
                             "action": "spawn_agent",
                             "description": "Spawn research agent in case sections need research",
                             "config": {
                                 "role": "researcher",
                                 "agent_type": "research-agent"
                             },
                             "next_step": "load_page_components",
                             "output_field": "researcher_info"
                         },

                         "load_page_components": {
                             "action": "load_page_section_components",
                             "description": "Load component definitions for this page''s sections",
                             "config": {
                                 "sections_from": "current_page.sections",
                                 "include_templates": true,
                                 "include_input_schema": true
                             },
                             "next_step": "build_render_context",
                             "output_field": "section_components"
                         },

                         "build_render_context": {
                             "action": "build_render_context",
                             "description": "Build render context from brief, site, and brand data",
                             "config": {
                                 "sources": {
                                     "brief": "reviewed_brief",
                                     "site": "site_record",
                                     "style": "style_collection",
                                     "page": "current_page",
                                     "assets": "brand_assets"
                                 }
                             },
                             "next_step": "process_sections_loop",
                             "output_field": "render_context"
                         },

                         "process_sections_loop": {
                             "action": "loop",
                             "description": "Process each section - template render or LLM generate",
                             "config": {
                                 "loop_var": "current_section",
                                 "iterate_over": "section_components",
                                 "max_iterations": 15,
                                 "substeps": {
                                     "check_render_mode": {
                                         "action": "conditional",
                                         "description": "Check if section needs LLM or just template",
                                         "config": {
                                             "condition": "current_section.render_mode == ''agent'' OR current_section.needs_llm == true",
                                             "then_step": "check_needs_research",
                                             "else_step": "render_from_template"
                                         }
                                     },

                                     "check_needs_research": {
                                         "action": "conditional",
                                         "description": "Check if section needs research first",
                                         "config": {
                                             "condition": "current_section.needs_research == true",
                                             "then_step": "call_researcher",
                                             "else_step": "generate_content"
                                         }
                                     },

                                     "call_researcher": {
                                         "action": "call_agent",
                                         "description": "Research topic for this section",
                                         "config": {
                                             "target_role": "researcher",
                                             "agent_type": "research-agent",
                                             "input_fields": ["current_section", "reviewed_brief", "site_record"],
                                             "timeout_seconds": 90
                                         },
                                         "next_step": "generate_content",
                                         "output_field": "research_result"
                                     },

                                     "generate_content": {
                                         "action": "execute_llm_prompt",
                                         "description": "Generate content for this section",
                                         "config": {
                                             "ai_service": {
                                                 "provider": "anthropic",
                                                 "model": "claude-sonnet-4-5-20250514",
                                                 "max_tokens": 2000,
                                                 "api_key_env_var": "ANTHROPIC_API_KEY"
                                             },
                                             "input_fields": ["current_section", "render_context", "research_result", "reviewed_brief"],
                                             "output_format": "json",
                                             "prompt_template": "Write content for the {{current_section.function}} section of {{current_page.title}}.\n\n## Company Context\nCompany: {{render_context.company_name}}\nIndustry: {{render_context.industry}}\nTone: {{render_context.tone}}\nTarget Audience: {{render_context.target_audience}}\nServices: {{reviewed_brief.services}}\n\n## Section Requirements\nComponent: {{current_section.name}}\nPurpose: {{current_section.description}}\n\n## Data Schema Required\n{{current_section.input_schema}}\n\n{{#if research_result}}\n## Research Findings\n{{research_result.summary}}\n\nSources:\n{{#each research_result.sources}}\n- [{{@index}}] {{this.title}} ({{this.domain}})\n{{/each}}\n{{/if}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON matching the input_schema. Example:\n```json\n{\n  \"headline\": \"Your Compelling Headline\",\n  \"body\": \"Engaging body content...\",\n  \"cta_text\": \"Call to Action\"\n}\n```\n\nRules:\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided"
                                         },
                                         "next_step": "render_section",
                                         "output_field": "generated_content"
                                     },

                                     "render_from_template": {
                                         "action": "render_component",
                                         "description": "Render section from template with brief data only",
                                         "config": {
                                             "component_from": "current_section",
                                             "context_from": "render_context",
                                             "output_html": true
                                         },
                                         "output_field": "section_output"
                                     },

                                     "render_section": {
                                         "action": "render_component",
                                         "description": "Render LLM-generated content into component template",
                                         "config": {
                                             "component_from": "current_section",
                                             "content_from": "generated_content",
                                             "context_from": "render_context",
                                             "output_html": true
                                         },
                                         "output_field": "section_output"
                                     }
                                 }
                             },
                             "next_step": "compile_page",
                             "output_field": "processed_sections"
                         },

                         "compile_page": {
                             "action": "compile_page_sections",
                             "description": "Compile all sections into page output",
                             "config": {
                                 "sections_from": "processed_sections",
                                 "page_from": "current_page",
                                 "include_research_ids": true
                             },
                             "next_step": "complete",
                             "output_field": "page_content"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_field": "page_content"
                             }
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              version = EXCLUDED.version,
                              default_config = EXCLUDED.default_config,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = now();

COMMIT;

----------------------

-- FIX: page-content-writer template syntax
-- =========================================
-- Issues:
-- 1. Missing dot prefixes on variable references
-- 2. Handlebars syntax ({{#each}}, {{#if}}, {{this.}}, {{@index}}) needs Go template syntax

-- Fix generate_content step (inside process_sections_loop substeps)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,substeps,generate_content,config,prompt_template}',
        '"Write content for the {{.current_section.function}} section of {{.current_page.title}}.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\n\n## Section Requirements\nComponent: {{.current_section.name}}\nPurpose: {{.current_section.description}}\n\n## Data Schema Required\n{{.current_section.input_schema}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.summary}}\n\nSources:\n{{range $index, $src := .research_result.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON matching the input_schema. Example:\n```json\n{\n  \"headline\": \"Your Compelling Headline\",\n  \"body\": \"Engaging body content...\",\n  \"cta_text\": \"Call to Action\"\n}\n```\n\nRules:\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided"'
                     )
WHERE type = 'page-content-writer';

-- Verify page-content-writer fix
SELECT 'page-content-writer' as agent,
       'generate_content' as step,
       CASE
           WHEN default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'config'->>'prompt_template' LIKE '%{{.current_section.%'
           AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'config'->>'prompt_template' LIKE '%{{if .research_result}}%'
    AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'config'->>'prompt_template' LIKE '%{{range%'
    THEN 'FIXED'
    ELSE 'NEEDS FIX'
END as status
FROM agent_definitions
WHERE type = 'page-content-writer';

    ---

amend paths to include input-data

-- Update page-content-writer workflow to use correct paths
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_components,config,page_from}',
        '"input_data.current_page"'
                     )
WHERE type = 'page-content-writer';

-- Also update sections_from if it exists
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_components,config,sections_from}',
        '"input_data.current_page.sections"'
                     )
WHERE type = 'page-content-writer';

-- again
-- ============================================================================
-- Fix page-content-writer workflow to use input_data paths
-- Database: clients_db
-- ============================================================================
--
-- The issue: call_agent passes fields under input_data.{field}
-- But actions expect them at root level
-- This updates the workflow config to use correct paths
-- ============================================================================

BEGIN;

-- 1. Fix load_page_components to look in input_data
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_components,config,page_from}',
        '"input_data.current_page"'
                     )
WHERE type = 'page-content-writer';

-- 2. Fix build_render_context sources to use input_data paths
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_render_context,config,sources}',
        '{
            "page": "input_data.current_page",
            "site": "input_data.site_record",
            "brief": "input_data.reviewed_brief",
            "style": "input_data.style_collection",
            "assets": "brand_assets"
        }'::jsonb
                     )
WHERE type = 'page-content-writer';

-- 3. Fix compile_page to use input_data path for page
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,compile_page,config,page_from}',
        '"input_data.current_page"'
                     )
WHERE type = 'page-content-writer';

-- 4. Fix the LLM prompt template references (inside process_sections_loop substeps)
-- The prompt_template uses {{.current_page.title}} which needs to work
-- This is harder to fix via SQL - the action-level fix handles this better

COMMIT;

-- Verify the changes
SELECT
    type,
    default_config->'workflow'->'steps'->'load_page_components'->'config'->>'page_from' as load_page_from,
    default_config->'workflow'->'steps'->'build_render_context'->'config'->'sources'->>'page' as render_page_from,
    default_config->'workflow'->'steps'->'compile_page'->'config'->>'page_from' as compile_page_from
FROM agent_definitions
WHERE type = 'page-content-writer';


---

fix path

-- ============================================================================
-- Fix page-content-writer process_sections_loop iterate_over path
-- Database: clients_db
-- ============================================================================
--
-- The issue: load_page_components returns {components: [...], count: N}
-- But iterate_over points to "section_components" (the whole object)
-- Should point to "section_components.components" (the array)
-- ============================================================================

BEGIN;

-- Fix the iterate_over path in process_sections_loop
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,iterate_over}',
        '"section_components.components"'
                     )
WHERE type = 'page-content-writer';

COMMIT;

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->>'iterate_over' as iterate_over
FROM agent_definitions
WHERE type = 'page-content-writer';

--
-- Update the generate_content substep to include current_page
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,substeps,generate_content,config,input_fields}',
        '["current_section", "render_context", "research_result", "reviewed_brief", "current_page"]'::jsonb
                     )
WHERE type = 'page-content-writer';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'config'->'input_fields' as input_fields
FROM agent_definitions
WHERE type = 'page-content-writer';

--

-- change paths

-- ============================================
-- Fix page-content-writer prompt template
-- The prompt references research_result.summary and research_result.sources
-- Need to change to research_result.response.summary and research_result.response.sources
-- ============================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,substeps,generate_content,config,prompt_template}',
        to_jsonb('Write content for the {{.current_section.function}} section of {{.current_page.title}}.

## Company Context
Company: {{.render_context.company_name}}
Industry: {{.render_context.industry}}
Tone: {{.render_context.tone}}
Target Audience: {{.render_context.target_audience}}
Services: {{.reviewed_brief.services}}

## Section Requirements
Component: {{.current_section.name}}
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

Return JSON matching the input_schema. Example:
```json
{
  "headline": "Your Compelling Headline",
  "body": "Engaging body content...",
  "cta_text": "Call to Action"
}
```

Rules:
- No placeholder text like [Your Company] or Lorem ipsum
- Be specific to this company and industry
- Professional but engaging tone matching the brief
- Include source citations [0], [1] if research was provided'::text)
                     )
WHERE type = 'page-content-writer';

-- Verify page-content-writer
SELECT
    type,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'config'->'prompt_template' as prompt
FROM agent_definitions
WHERE type = 'page-content-writer';

--
add await_response

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_research_agent,config,await_response}',
        'true'
                     )
WHERE type = 'page-content-writer';

--

content_from change generated_content.result
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,substeps,render_section,config,content_from}',
        '"generated_content.result"'
                     )
WHERE type = 'page-content-writer';

---

-- path changes render_from_template

-- ============================================================================
-- WORKFLOW FIX 1: Add content_from to render_from_template
-- ============================================================================
-- This ensures non-LLM templates get brand/site data from render_context

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,substeps,render_from_template,config}',
        '{
            "output_html": true,
            "context_from": "render_context",
            "content_from": "render_context",
            "component_from": "current_section"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- ============================================================================
-- WORKFLOW FIX 2: Update LLM prompt to use correct field names
-- ============================================================================
-- Templates expect: headline, subheadline, primary_cta, primary_cta_url
-- Old prompt example used: headline, body, cta_text

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,substeps,generate_content,config,prompt_template}',
        '"Write content for the {{.current_section.function}} section of {{.current_page.title}}.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Section Requirements\nComponent: {{.current_section.name}}\nFunction: {{.current_section.function}}\nPurpose: {{.current_section.description}}\n\n## Data Schema Required\n{{.current_section.input_schema}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON with these EXACT field names (use the ones that apply to this component type):\n\n### For Hero/Banner sections:\n```json\n{\n  \"headline\": \"Your Compelling Main Headline\",\n  \"subheadline\": \"Supporting text that expands on the headline\",\n  \"primary_cta\": \"Get Started\",\n  \"primary_cta_url\": \"/contact.html\",\n  \"secondary_cta\": \"Learn More\",\n  \"secondary_cta_url\": \"/about.html\"\n}\n```\n\n### For Feature/Services sections:\n```json\n{\n  \"headline\": \"Section Headline\",\n  \"subheadline\": \"Brief introduction\",\n  \"features\": [\n    {\"name\": \"Feature Name\", \"description\": \"Feature description\", \"icon\": \"icon-name\"},\n    {\"name\": \"Feature 2\", \"description\": \"Description 2\", \"icon\": \"icon-name\"}\n  ]\n}\n```\n\n### For CTA/Call-to-Action sections:\n```json\n{\n  \"headline\": \"Ready to Get Started?\",\n  \"subheadline\": \"Contact us today\",\n  \"primary_cta\": \"Contact Us\",\n  \"primary_cta_url\": \"/contact.html\"\n}\n```\n\n### For Text/Content sections:\n```json\n{\n  \"heading\": \"Section Heading\",\n  \"content\": \"Paragraph content here...\"\n}\n```\n\nRules:\n- Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided"'
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- ============================================================================
-- VERIFY ALL CHANGES
-- ============================================================================

-- Check render_from_template has content_from
SELECT type,
       default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'render_from_template'->'config' as render_config
FROM agent_definitions
WHERE type = 'page-content-writer';

-- Check prompt has correct field examples
SELECT type,
       substring(default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'config'->>'prompt_template' from 1 for 300) as prompt_preview
FROM agent_definitions
WHERE type = 'page-content-writer';

--
5946a27b-38ab-41e8-8b49-7bc1a4b626b8 | page-content-writer | Page Content Writer | Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content. | specialist | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_field": "page_content"}}, "compile_page": {"action": "compile_page_sections", "config": {"page_from": "input_data.current_page", "sections_from": "processed_sections", "include_research_ids": true}, "next_step": "complete", "description": "Compile all sections into page output", "output_field": "page_content"}, "build_render_context": {"action": "build_render_context", "config": {"sources": {"page": "input_data.current_page", "site": "input_data.site_record", "brief": "input_data.reviewed_brief", "style": "input_data.style_collection", "assets": "brand_assets"}}, "next_step": "process_sections_loop", "description": "Build render context from brief, site, and brand data", "output_field": "render_context"}, "load_page_components": {"action": "load_page_section_components", "config": {"page_from": "input_data.current_page", "sections_from": "input_data.current_page.sections", "include_templates": true, "include_input_schema": true}, "next_step": "build_render_context", "description": "Load component definitions for this page's sections", "output_field": "section_components"}, "spawn_research_agent": {"action": "spawn_agent", "config": {"role": "researcher", "agent_type": "research-agent", "await_response": true}, "next_step": "load_page_components", "description": "Spawn research agent in case sections need research", "output_field": "researcher_info"}, "process_sections_loop": {"action": "loop", "config": {"loop_var": "current_section", "substeps": {"render_section": {"action": "render_component", "config": {"output_html": true, "content_from": "generated_content.result", "context_from": "render_context", "component_from": "current_section"}, "description": "Render LLM-generated content into component template", "output_field": "section_output"}, "call_researcher": {"action": "call_agent", "config": {"agent_type": "research-agent", "target_role": "researcher", "input_fields": ["current_section", "reviewed_brief", "site_record"], "timeout_seconds": 90}, "next_step": "generate_content", "description": "Research topic for this section", "output_field": "research_result"}, "generate_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 2000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["current_section", "render_context", "research_result", "reviewed_brief", "current_page"], "output_format": "json", "prompt_template": "Write content for the {{.current_section.function}} section of {{.current_page.title}}.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Section Requirements\nComponent: {{.current_section.name}}\nFunction: {{.current_section.function}}\nPurpose: {{.current_section.description}}\n\n## Data Schema Required\n{{.current_section.input_schema}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON with these EXACT field names (use the ones that apply to this component type):\n\n### For Hero/Banner sections:\n```json\n{\n  \"headline\": \"Your Compelling Main Headline\",\n  \"subheadline\": \"Supporting text that expands on the headline\",\n  \"primary_cta\": \"Get Started\",\n  \"primary_cta_url\": \"/contact.html\",\n  \"secondary_cta\": \"Learn More\",\n  \"secondary_cta_url\": \"/about.html\"\n}\n```\n\n### For Feature/Services sections:\n```json\n{\n  \"headline\": \"Section Headline\",\n  \"subheadline\": \"Brief introduction\",\n  \"features\": [\n    {\"name\": \"Feature Name\", \"description\": \"Feature description\", \"icon\": \"icon-name\"},\n    {\"name\": \"Feature 2\", \"description\": \"Description 2\", \"icon\": \"icon-name\"}\n  ]\n}\n```\n\n### For CTA/Call-to-Action sections:\n```json\n{\n  \"headline\": \"Ready to Get Started?\",\n  \"subheadline\": \"Contact us today\",\n  \"primary_cta\": \"Contact Us\",\n  \"primary_cta_url\": \"/contact.html\"\n}\n```\n\n### For Text/Content sections:\n```json\n{\n  \"heading\": \"Section Heading\",\n  \"content\": \"Paragraph content here...\"\n}\n```\n\nRules:\n- Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided"}, "next_step": "render_section", "description": "Generate content for this section", "output_field": "generated_content"}, "check_render_mode": {"action": "conditional", "config": {"condition": "current_section.render_mode == 'agent' OR current_section.needs_llm == true", "else_step": "render_from_template", "then_step": "check_needs_research"}, "description": "Check if section needs LLM or just template"}, "check_needs_research": {"action": "conditional", "config": {"condition": "current_section.needs_research == true", "else_step": "generate_content", "then_step": "call_researcher"}, "description": "Check if section needs research first"}, "render_from_template": {"action": "render_component", "config": {"output_html": true, "content_from": "render_context", "context_from": "render_context", "component_from": "current_section"}, "description": "Render section from template with brief data only", "output_field": "section_output"}}, "iterate_over": "section_components.components", "max_iterations": 15}, "next_step": "compile_page", "description": "Process each section - template render or LLM generate", "output_field": "processed_sections"}}, "start_step": "spawn_research_agent"}, "processing_mode": "task", "timeout_seconds": 180} | t         | 2025-12-22 17:47:17.609605+00 | 2026-01-14 22:08:07.685662+00 |            | ["content-generation", "template-rendering", "research"] | docker.io/aqls/agent-chassis | v1.0.673  |         | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | []          | {}                     |           0 | f           | {"expects": {"site_record": "object with site_id, domain", "brand_assets": "object with logo, images (optional)", "current_page": "object with name, title, sections[]", "reviewed_brief": "object with company_name, services, about_us, etc", "style_collection": "object with colors, component refs"}, "required": ["current_page", "site_record", "reviewed_brief"]} | {"produces": {"sections": "array of {component_id, rendered_html, content_data}", "page_name": "string", "research_ids": "array of research result UUIDs used"}}
(1 row)

--
-- update conditional whether to use llm for e.g. hero section

-- Fix the check_render_mode conditional in page-content-writer
-- Change from broken expression syntax to working condition_field format

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_render_mode,config}',
        '{
            "condition_field": "current_section.needs_llm",
            "conditions": {
                "true": "check_needs_research"
            },
            "default": "render_from_template"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer';

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'check_render_mode'->'config' as check_render_mode_config
FROM agent_definitions
WHERE type = 'page-content-writer';



==============================

backup before changes:
id                     | 5946a27b-38ab-41e8-8b49-7bc1a4b626b8
type                   | page-content-writer
display_name           | Page Content Writer
description            | Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content.
category               | specialist
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_field": "page_content"}}, "compile_page": {"action": "compile_page_sections", "config": {"page_from": "input_data.current_page", "sections_from": "processed_sections", "include_research_ids": true}, "next_step": "complete", "description": "Compile all sections into page output", "output_field": "page_content"}, "build_render_context": {"action": "build_render_context", "config": {"sources": {"page": "input_data.current_page", "site": "input_data.site_record", "brief": "input_data.reviewed_brief", "style": "input_data.style_collection", "assets": "brand_assets"}}, "next_step": "process_sections_loop", "description": "Build render context from brief, site, and brand data", "output_field": "render_context"}, "load_page_components": {"action": "load_page_section_components", "config": {"page_from": "input_data.current_page", "sections_from": "input_data.current_page.sections", "include_templates": true, "include_input_schema": true}, "next_step": "build_render_context", "description": "Load component definitions for this page's sections", "output_field": "section_components"}, "spawn_research_agent": {"action": "spawn_agent", "config": {"role": "researcher", "agent_type": "research-agent", "await_response": true}, "next_step": "load_page_components", "description": "Spawn research agent in case sections need research", "output_field": "researcher_info"}, "process_sections_loop": {"action": "loop", "config": {"loop_var": "current_section", "substeps": {"render_section": {"action": "render_component", "config": {"output_html": true, "content_from": "generated_content.result", "context_from": "render_context", "component_from": "current_section"}, "description": "Render LLM-generated content into component template", "output_field": "section_output"}, "call_researcher": {"action": "call_agent", "config": {"agent_type": "research-agent", "target_role": "researcher", "input_fields": ["current_section", "reviewed_brief", "site_record"], "timeout_seconds": 90}, "next_step": "generate_content", "description": "Research topic for this section", "output_field": "research_result"}, "generate_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 2000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["current_section", "render_context", "research_result", "reviewed_brief", "current_page"], "output_format": "json", "prompt_template": "Write content for the {{.current_section.function}} section of {{.current_page.title}}.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Section Requirements\nComponent: {{.current_section.name}}\nFunction: {{.current_section.function}}\nPurpose: {{.current_section.description}}\n\n## Data Schema Required\n{{.current_section.input_schema}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON with these EXACT field names (use the ones that apply to this component type):\n\n### For Hero/Banner sections:\n```json\n{\n  \"headline\": \"Your Compelling Main Headline\",\n  \"subheadline\": \"Supporting text that expands on the headline\",\n  \"primary_cta\": \"Get Started\",\n  \"primary_cta_url\": \"/contact.html\",\n  \"secondary_cta\": \"Learn More\",\n  \"secondary_cta_url\": \"/about.html\"\n}\n```\n\n### For Feature/Services sections:\n```json\n{\n  \"headline\": \"Section Headline\",\n  \"subheadline\": \"Brief introduction\",\n  \"features\": [\n    {\"name\": \"Feature Name\", \"description\": \"Feature description\", \"icon\": \"icon-name\"},\n    {\"name\": \"Feature 2\", \"description\": \"Description 2\", \"icon\": \"icon-name\"}\n  ]\n}\n```\n\n### For CTA/Call-to-Action sections:\n```json\n{\n  \"headline\": \"Ready to Get Started?\",\n  \"subheadline\": \"Contact us today\",\n  \"primary_cta\": \"Contact Us\",\n  \"primary_cta_url\": \"/contact.html\"\n}\n```\n\n### For Text/Content sections:\n```json\n{\n  \"heading\": \"Section Heading\",\n  \"content\": \"Paragraph content here...\"\n}\n```\n\nRules:\n- Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided"}, "next_step": "render_section", "description": "Generate content for this section", "output_field": "generated_content"}, "check_render_mode": {"action": "conditional", "config": {"condition": "current_section.render_mode == 'agent' OR current_section.needs_llm == true", "else_step": "render_from_template", "then_step": "check_needs_research"}, "description": "Check if section needs LLM or just template"}, "check_needs_research": {"action": "conditional", "config": {"condition": "current_section.needs_research == true", "else_step": "generate_content", "then_step": "call_researcher"}, "description": "Check if section needs research first"}, "render_from_template": {"action": "render_component", "config": {"output_html": true, "content_from": "render_context", "context_from": "render_context", "component_from": "current_section"}, "description": "Render section from template with brief data only", "output_field": "section_output"}}, "iterate_over": "section_components.components", "max_iterations": 15}, "next_step": "compile_page", "description": "Process each section - template render or LLM generate", "output_field": "processed_sections"}}, "start_step": "spawn_research_agent"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-12-22 17:47:17.609605+00
updated_at             | 2026-01-17 17:45:08.029181+00
deleted_at             |
capabilities           | ["content-generation", "template-rendering", "research"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.681
command                |
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    |
task_workflow          |
orchestrator_workflow  |
orchestration_workflow |
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         |
status                 | active
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"site_record": "object with site_id, domain", "brand_assets": "object with logo, images (optional)", "current_page": "object with name, title, sections[]", "reviewed_brief": "object with company_name, services, about_us, etc", "style_collection": "object with colors, component refs"}, "required": ["current_page", "site_record", "reviewed_brief"]}
output_contract        | {"produces": {"sections": "array of {component_id, rendered_html, content_data}", "page_name": "string", "research_ids": "array of research result UUIDs used"}}



=====================================================
--

-- Fix: Add start_step to page-content-writer's process_sections_loop
--
-- Problem: The loop's substeps have no explicit start_step, and the auto-detection
-- fails because conditional steps (check_render_mode, check_needs_research) use
-- config.then_step/else_step instead of the standard next_step field.
--
-- This causes buildSubstepOrder() to pick a non-deterministic starting step
-- since multiple steps appear to have no incoming references.
--
-- Solution: Add explicit start_step: "check_render_mode" to the substeps config

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,start_step}',
        '"check_render_mode"'
                     )
WHERE type = 'page-content-writer';

-- Verify the start_step fix
SELECT
    type,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->>'start_step' as loop_start_step
FROM agent_definitions
WHERE type = 'page-content-writer';

-- ============================================================================
-- DIAGNOSTIC QUERY: Check what the conditional currently looks like
-- ============================================================================
SELECT
    jsonb_pretty(
            default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'check_render_mode'
    ) as check_render_mode_config
FROM agent_definitions
WHERE type = 'page-content-writer';

-- ============================================================================
-- OPTIONAL: If the string expression conditional still doesn't work,
-- convert check_render_mode to Format 2 (condition_field + conditions map)
--
-- NOTE: This changes the logic to check ONLY needs_llm (not render_mode)
-- since Format 2 can only check one field at a time.
-- ============================================================================

-- Uncomment below if you want to try Format 2 instead:
/*
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,process_sections_loop,config,substeps,check_render_mode}',
    '{
        "action": "evaluate_condition",
        "config": {
            "condition_field": "current_section.needs_llm",
            "conditions": {
                "true": "check_needs_research"
            },
            "default": "render_from_template"
        },
        "description": "Check if section needs LLM or just template"
    }'::jsonb
)
WHERE type = 'page-content-writer';
*/

-- ============================================================================
-- VERIFY: Check the full loop config after fixes
-- ============================================================================
SELECT
    type,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->>'start_step' as start_step,
    jsonb_array_length(
    COALESCE(
    (SELECT jsonb_agg(key) FROM jsonb_object_keys(
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'
    ) AS key),
    '[]'::jsonb
    )
    ) as substep_count
FROM agent_definitions
WHERE type = 'page-content-writer';