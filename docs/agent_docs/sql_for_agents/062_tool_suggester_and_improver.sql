-- Migration: 062_tool_suggester_and_improver.sql
--
-- Two new agents for the tool lifecycle:
--
--   tool-suggester  — LLM evaluates what tools would benefit a site,
--                     creates add_tool work items for each recommendation.
--                     Handler for "evaluate_tools" work items.
--
--   tool-improver   — LLM fixes/improves an existing deployed tool
--                     based on an issue description (rendering, mobile,
--                     UX, accessibility, etc). Handler for "improve_tool"
--                     work items.
--
-- Uses dollar-quoting for default_config to avoid single-quote
-- escaping issues in embedded SQL queries and prompt text.

-- ============================================================================
-- 1. tool-suggester
-- ============================================================================
INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag,
    resources, topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             'tool-suggester',
             'Tool Suggester',
             'Evaluates what interactive tools would benefit a site based on its industry, services, audience, and existing pages. Uses LLM judgment — not limited to library catalogue. Creates add_tool work items for each recommendation.',
             'specialist',
             $cfg${
        "workflow": {
            "start_step": "read_specs",
             "steps": {
                "read_specs": {
                    "action": "read_site_spec",
             "config": {
                        "site_id": "input_data.site_id"
                    },
             "output_field": "site_specs",
             "next_step": "load_pages",
             "description": "Load all site specs (identity, classification, brand_dna)"
                },
             "load_pages": {
                    "action": "query_database",
             "config": {
                        "query": "SELECT p.name, p.slug, p.purpose, p.status FROM pages p WHERE p.site_id = '{{.input_data.site_id}}' AND p.status IN ('deployed', 'planned') ORDER BY p.sort_order",
             "output_format": "array"
                    },
             "output_field": "pages",
             "next_step": "load_existing_tools",
             "description": "Load deployed and planned pages"
                },
             "load_existing_tools": {
                    "action": "query_database",
             "config": {
                        "query": "SELECT cc.function, cc.display_name, cc.category FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE p.site_id = '{{.input_data.site_id}}' AND cc.component_level = 'tool' AND cc.is_active = true",
             "output_format": "array"
                    },
             "output_field": "existing_tools",
             "next_step": "load_library_tools",
             "description": "Check what tools are already deployed"
                },
             "load_library_tools": {
                    "action": "query_database",
             "config": {
                        "query": "SELECT id::text, function, display_name, category, description FROM content_components WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true AND html_template != '' ORDER BY display_name LIMIT 30",
             "output_format": "array"
                    },
             "output_field": "library_tools",
             "next_step": "suggest_tools",
             "description": "Load available library tools for reference"
                },
             "suggest_tools": {
                    "action": "execute_llm_prompt",
             "config": {
                        "ai_service": {
                            "provider": "anthropic",
             "model": "claude-sonnet-4-5",
             "max_tokens": 3000,
             "api_key_env_var": "ANTHROPIC_API_KEY"
                        },
             "input_fields": ["input_data", "site_specs", "pages", "existing_tools", "library_tools"],
             "output_format": "json",
             "prompt_template": "You are evaluating what interactive tools would genuinely help visitors to this website.\n\n## Site Context\nDomain: {{.input_data.domain}}\nIndustry: {{if .site_specs.identity}}{{.site_specs.identity.industry}} — {{.site_specs.identity.sub_industry}}{{end}}\nCompany: {{if .site_specs.identity}}{{.site_specs.identity.company_name}}{{end}}\nAbout: {{if .site_specs.identity}}{{.site_specs.identity.about_summary}}{{end}}\nServices: {{if .site_specs.identity}}{{range .site_specs.identity.services}}\n  - {{.name}}: {{.description}}{{end}}{{end}}\nSite Type: {{if .site_specs.classification}}{{.site_specs.classification.site_type}}{{end}}\nTarget Audience: {{if .site_specs.identity}}{{.site_specs.identity.target_audience}}{{end}}\n\n## Existing Pages\n{{range .pages}}- {{.name}} ({{.slug}}): {{.purpose}}\n{{end}}\n\n## Already Deployed Tools\n{{if .existing_tools}}{{range .existing_tools}}- {{.display_name}} ({{.function}})\n{{end}}{{else}}None deployed yet.{{end}}\n\n## Available in Library (can be forked and customised)\n{{range .library_tools}}- {{.display_name}} ({{.function}}): {{.description}}\n{{end}}\n\n## Your Task\n\nSuggest 2-5 interactive tools that would genuinely add value for this website. Think about what someone in this industry would actually use.\n\nExamples of good tool suggestions:\n- Gas wholesaler: unit converter (therms to kWh to MJ), quote estimator, service area checker\n- Photographer: booking calculator, print size guide, gallery filter\n- Accountant: tax deadline calendar, VAT calculator, expense categoriser\n- Restaurant: reservation widget, menu filter (allergies), tip calculator\n\nDO NOT suggest tools that are irrelevant to this industry (e.g. password strength checker for a gas wholesaler).\n\nFor each tool, decide whether it can be forked from the library (reference the library function name) or needs to be built from scratch.\n\nReturn ONLY valid JSON:\n{\n  \"reasoning\": \"Brief explanation of why these tools suit this industry and audience\",\n  \"suggestions\": [\n    {\n      \"name\": \"Human-readable tool name\",\n      \"function\": \"kebab-case-function-name\",\n      \"description\": \"What the tool does and why it helps visitors\",\n      \"priority\": 1,\n      \"target_page\": \"slug of best page for this tool, or new for a dedicated tools page\",\n      \"library_source\": \"function name from library if forkable, or null if new\",\n      \"complexity\": \"simple\"\n    }\n  ]\n}"
                    },
             "output_field": "evaluation",
             "next_step": "check_has_suggestions",
             "description": "LLM evaluates what tools would benefit this site"
                },
             "check_has_suggestions": {
                    "action": "conditional",
             "config": {
                        "condition": "evaluation.result.suggestions != null",
             "then_step": "create_items_loop",
             "else_step": "complete"
                    },
             "description": "Check if LLM produced any suggestions"
                },
             "create_items_loop": {
                    "action": "loop",
             "config": {
                        "items_field": "evaluation.result.suggestions",
             "item_variable": "current_suggestion",
             "mode": "sequential",
             "max_iterations": 10,
             "sub_workflow": {
                            "start_step": "create_tool_item",
             "steps": {
                                "create_tool_item": {
                                    "action": "create_work_item",
             "config": {
                                        "site_id": "input_data.site_id",
             "item_type": "add_tool",
             "handler_agent": "tool-deployer",
             "item_domain": "build",
             "severity": "low",
             "source": "tool-suggester",
             "summary": "current_suggestion.name",
             "priority": 130,
             "item_key_prefix": "add_tool",
             "spec_data": "current_suggestion"
                                    },
             "output_field": "item_created",
             "next_step": "done",
             "description": "Create add_tool work item for this suggestion"
                                },
             "done": {
                                    "action": "loop_complete",
             "description": "Suggestion item created"
                                }
                            }
                        }
                    },
             "output_field": "items_created",
             "next_step": "complete",
             "description": "Create work items for each tool suggestion"
                },
             "complete": {
                    "action": "complete_workflow",
             "config": {
                        "output_fields": ["evaluation", "items_created"]
                    },
             "description": "Tool evaluation complete"
                }
            }
        },
             "processing_mode": "orchestrator",
             "timeout_seconds": 180
    }$cfg$::jsonb,
             true,
             '["llm", "tool-evaluation", "work-item-creation"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.818',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             'analyst',
             'experimental',
             '["tools", "evaluation", "improvement"]'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["work_item_id"], "description": "Receives site_id + domain from dispatch loop. Evaluates what tools would benefit the site."}'::jsonb,
             '{"produces": {"evaluation": "LLM evaluation with reasoning and suggestions array", "items_created": "Loop result of created add_tool work items"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       updated_at = NOW();


-- ============================================================================
-- 2. tool-improver
-- ============================================================================
INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag,
    resources, topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             'tool-improver',
             'Tool Improver',
             'Incrementally improves deployed tools. Loads current HTML, applies LLM-driven fixes for rendering, mobile, UX, or accessibility issues, saves updated HTML, and triggers page re-render.',
             'specialist',
             $cfg${
        "workflow": {
            "start_step": "load_tool",
             "steps": {
                "load_tool": {
                    "action": "query_database",
             "config": {
                        "query": "SELECT cc.id::text as component_id, cc.function, cc.display_name, cc.html_template, cc.description, cc.semantic_tags::text as tags, cc.is_dark_section, p.slug as page_slug, p.id::text as page_id, p.name as page_name FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = '{{.input_data.component_id}}' AND cc.is_active = true LIMIT 1",
             "output_format": "object"
                    },
             "output_field": "tool_data",
             "next_step": "check_tool_found",
             "description": "Load the tool component and its page context"
                },
             "check_tool_found": {
                    "action": "conditional",
             "config": {
                        "condition": "tool_data.component_id != null",
             "then_step": "load_brand_context",
             "else_step": "complete_not_found"
                    },
             "description": "Verify the tool component exists"
                },
             "complete_not_found": {
                    "action": "complete_workflow",
             "config": {
                        "output_fields": ["tool_data"],
             "status": "skipped",
             "reason": "Tool component not found"
                    },
             "description": "Tool not found — skip"
                },
             "load_brand_context": {
                    "action": "read_site_spec",
             "config": {
                        "site_id": "input_data.site_id"
                    },
             "output_field": "site_specs",
             "next_step": "improve_tool",
             "description": "Load brand context so improvements match site style"
                },
             "improve_tool": {
                    "action": "execute_llm_prompt",
             "config": {
                        "ai_service": {
                            "provider": "anthropic",
             "model": "claude-sonnet-4-5",
             "max_tokens": 8000,
             "api_key_env_var": "ANTHROPIC_API_KEY"
                        },
             "input_fields": ["input_data", "tool_data", "site_specs"],
             "output_format": "text",
             "prompt_template": "You are improving an interactive tool component on a website.\n\n## Issue to Fix\n{{.input_data.issue}}\n\n## Tool Info\nName: {{.tool_data.display_name}}\nFunction: {{.tool_data.function}}\nPage: {{.tool_data.page_name}} ({{.tool_data.page_slug}})\nDark section: {{.tool_data.is_dark_section}}\n\n## Current HTML\n{{.tool_data.html_template}}\n\n{{if .tool_data.description}}## Component Description\n{{.tool_data.description}}{{end}}\n\n## Site Brand Context\n{{if .site_specs.webdesign}}Colors: {{.site_specs.webdesign}}{{end}}\n{{if .site_specs.identity}}Industry: {{.site_specs.identity.industry}}{{end}}\n\n## Rules\n1. Fix the specific issue described above\n2. Keep the core functionality intact\n3. Use CSS custom properties (var(--color-primary) etc) for colours — never hardcode hex values\n4. Ensure the tool works on mobile (min-width considerations, touch targets)\n5. Keep all interactive JavaScript working\n6. The output replaces html_template — it must be a complete, self-contained HTML fragment\n7. Include inline <style> for layout-only CSS (colours come from the site stylesheet)\n8. Include inline <script> for any JavaScript\n9. Do not add external dependencies\n\nOutput ONLY the improved HTML fragment. No markdown fences. No explanation. Start directly with the HTML."
                    },
             "output_field": "improved_html",
             "next_step": "update_component",
             "description": "LLM fixes the tool based on the reported issue"
                },
             "update_component": {
                    "action": "update_component_html",
             "config": {
                        "component_id": "tool_data.component_id",
             "html_field": "improved_html.result",
             "create_version": true
                    },
             "output_field": "update_result",
             "next_step": "create_rerender_item",
             "description": "Save improved HTML back to content_components"
                },
             "create_rerender_item": {
                    "action": "create_work_item",
             "config": {
                        "site_id": "input_data.site_id",
             "page_id": "tool_data.page_id",
             "item_type": "needs_rerender",
             "handler_agent": "rerender-pages",
             "item_domain": "build",
             "severity": "medium",
             "source": "tool-improver",
             "summary": "Re-render page after tool improvement",
             "priority": 50,
             "item_key_prefix": "rerender_tool_fix"
                    },
             "output_field": "rerender_item",
             "next_step": "complete",
             "description": "Trigger page re-render to deploy the improved tool"
                },
             "complete": {
                    "action": "complete_workflow",
             "config": {
                        "output_fields": ["tool_data", "improved_html", "update_result", "rerender_item"]
                    },
             "description": "Tool improvement complete"
                }
            }
        },
             "processing_mode": "orchestrator",
             "timeout_seconds": 300
    }$cfg$::jsonb,
             true,
             '["llm", "tool-improvement", "component-editing"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.818',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             'specialist',
             'experimental',
             '["tools", "improvement", "rendering"]'::jsonb,
             '{"required": ["site_id", "component_id", "issue"], "optional": ["work_item_id", "domain", "check"], "description": "Receives component_id + issue description. Loads current HTML, applies fix, triggers re-render."}'::jsonb,
             '{"produces": {"improved_html": "Updated HTML fragment", "update_result": "DB update confirmation", "rerender_item": "Created needs_rerender work item"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       updated_at = NOW();

