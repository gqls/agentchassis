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