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

---
--

-- ============================================================
-- Fix tool-suggester: correct columns, parameterised queries,
-- .specs. paths, model upgrade
-- ============================================================

UPDATE agent_definitions
SET default_config = $cfg${
    "workflow": {
        "start_step": "read_specs",
        "processing_mode": "orchestrator",
        "timeout_seconds": 300,
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
                    "query": "SELECT p.name, p.url, p.title, p.page_type, p.build_status FROM pages p WHERE p.site_id = $1 AND p.status = 'active' ORDER BY p.nav_order",
                    "params": ["input_data.site_id"],
                    "output_format": "array"
                },
                "output_field": "pages",
                "next_step": "load_existing_tools",
                "description": "Load active pages for context"
            },
            "load_existing_tools": {
                "action": "query_database",
                "config": {
                    "query": "SELECT cc.function, cc.display_name, cc.category FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1 AND cc.component_level = 'tool' AND cc.is_active = true",
                    "params": ["input_data.site_id"],
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
                        "model": "claude-sonnet-4-6",
                        "max_tokens": 3000,
                        "api_key_env_var": "ANTHROPIC_API_KEY"
                    },
                    "input_fields": ["input_data", "site_specs", "pages", "existing_tools", "library_tools"],
                    "output_format": "json",
                    "prompt_template": "You are evaluating what interactive tools would genuinely help visitors to this website.\n\n## Site Context\nDomain: {{.input_data.domain}}\n{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}} — {{.site_specs.specs.identity.sub_industry}}\nCompany: {{.site_specs.specs.identity.company_name}}\nAbout: {{.site_specs.specs.identity.about_summary}}\nServices: {{if .site_specs.specs.identity.services}}{{range .site_specs.specs.identity.services}}\n  - {{.name}}: {{.description}}{{end}}{{end}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n{{if .site_specs.specs.classification}}Site Type: {{.site_specs.specs.classification.site_type}}{{end}}\n\n## Existing Pages\n{{if .pages}}{{range .pages}}- {{.name}} ({{.url}}): {{.title}}\n{{end}}{{else}}No pages loaded.{{end}}\n\n## Already Deployed Tools\n{{if .existing_tools}}{{range .existing_tools}}- {{.display_name}} ({{.function}})\n{{end}}{{else}}None deployed yet.{{end}}\n\n## Available in Library (can be forked and customised)\n{{if .library_tools}}{{range .library_tools}}- {{.display_name}} ({{.function}}): {{.description}}\n{{end}}{{else}}No library tools available.{{end}}\n\n## Your Task\n\nSuggest 2-5 interactive tools that would genuinely add value for this website. Think about what someone in this industry would actually use.\n\nExamples of good tool suggestions:\n- Gas wholesaler: unit converter (therms to kWh to MJ), quote estimator, service area checker\n- Photographer: booking calculator, print size guide, gallery filter\n- Accountant: tax deadline calendar, VAT calculator, expense categoriser\n- Restaurant: reservation widget, menu filter (allergies), tip calculator\n\nDO NOT suggest tools that are irrelevant to this industry (e.g. password strength checker for a gas wholesaler).\n\nFor each tool, decide whether it can be forked from the library (reference the library function name) or needs to be built from scratch.\n\nReturn ONLY valid JSON:\n{\n  \"reasoning\": \"Brief explanation of why these tools suit this industry and audience\",\n  \"suggestions\": [\n    {\n      \"name\": \"Human-readable tool name\",\n      \"function\": \"kebab-case-function-name\",\n      \"description\": \"What the tool does and why it helps visitors\",\n      \"priority\": 1,\n      \"target_page\": \"name of best page for this tool, or new for a dedicated tools page\",\n      \"library_source\": \"function name from library if forkable, or null if new\",\n      \"complexity\": \"simple\"\n    }\n  ]\n}"
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
                                "description": "Item created"
                            }
                        }
                    }
                },
                "output_field": "items_created",
                "next_step": "complete",
                "description": "Create work items for each suggestion"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["evaluation", "items_created"]
                },
                "description": "Tool evaluation complete"
            }
        }
    }
}$cfg$::jsonb,
    updated_at = NOW()
WHERE type = 'tool-suggester';


-- ============================================================
-- Fix tool-improver: parameterised query, correct columns,
-- .specs. paths, model upgrade
-- ============================================================

UPDATE agent_definitions
SET default_config = $cfg2${
    "workflow": {
        "start_step": "load_tool",
        "processing_mode": "orchestrator",
        "timeout_seconds": 300,
        "steps": {
            "load_tool": {
                "action": "query_database",
                "config": {
                    "query": "SELECT cc.id::text as component_id, cc.function, cc.display_name, cc.html_template, cc.description, cc.semantic_tags::text as tags, cc.is_dark_section, p.url as page_url, p.id::text as page_id, p.name as page_name FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1",
                    "params": ["input_data.component_id"],
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
                        "model": "claude-sonnet-4-6",
                        "max_tokens": 8000,
                        "api_key_env_var": "ANTHROPIC_API_KEY"
                    },
                    "input_fields": ["input_data", "tool_data", "site_specs"],
                    "output_format": "text",
                    "prompt_template": "You are improving an interactive tool component on a website.\n\n## Issue to Fix\n{{.input_data.issue}}\n\n## Tool Info\nName: {{.tool_data.display_name}}\nFunction: {{.tool_data.function}}\nPage: {{.tool_data.page_name}} ({{.tool_data.page_url}})\nDark section: {{.tool_data.is_dark_section}}\n\n## Current HTML\n{{.tool_data.html_template}}\n\n{{if .tool_data.description}}## Component Description\n{{.tool_data.description}}{{end}}\n\n## Site Brand Context\n{{if .site_specs.specs.webdesign}}Colors: {{.site_specs.specs.webdesign}}{{end}}\n{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}}{{end}}\n\n## Rules\n1. Fix the specific issue described above\n2. Keep the core functionality intact\n3. Use CSS custom properties (var(--color-primary) etc) for colours — never hardcode hex values\n4. Ensure the tool works on mobile (min-width considerations, touch targets)\n5. Keep all interactive JavaScript working\n6. The output replaces html_template — it must be a complete, self-contained HTML fragment\n7. Include inline <style> for layout-only CSS (colours come from the site stylesheet)\n8. Include inline <script> for any JavaScript\n9. Do not add external dependencies\n\nOutput ONLY the improved HTML fragment. No markdown fences. No explanation. Start directly with the HTML."
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
    }
}$cfg2$::jsonb,
    updated_at = NOW()
WHERE type = 'tool-improver';


-- ============================================================
-- Verify
-- ============================================================
SELECT type, status,
       default_config->'workflow'->'steps'->'load_pages'->'config'->>'query' as pages_query,
    default_config->'workflow'->'steps'->'load_pages'->'config'->'params' as pages_params
FROM agent_definitions
WHERE type = 'tool-suggester';

SELECT type, status,
       default_config->'workflow'->'steps'->'load_tool'->'config'->>'query' as tool_query,
    default_config->'workflow'->'steps'->'load_tool'->'config'->'params' as tool_params
FROM agent_definitions
WHERE type = 'tool-improver';

---
-- improve tool suggestion - no tools if not needed

-- ============================================================
-- Patch tool-suggester: fix routing and strengthen prompt
-- ============================================================
-- The create_items_loop currently sends all suggestions to
-- tool-deployer, but tools with library_source=null need
-- tool-generator. Also, the LLM needs clearer permission
-- to return zero suggestions.
--
-- This replaces the create_items_loop with a smarter version
-- that uses conditional routing. It also adds a
-- "no suggestions" path that completes cleanly.
--
-- NOTE: This must run AFTER fix_tool_agent_definitions.sql
-- since it patches the same agent.

-- We need to update the prompt_template to add the zero-suggestions instruction,
-- and update the loop to handle routing.

-- The cleanest approach: use a Go action for the routing logic.
-- For now, we split into two loops — one for library forks, one for novel tools.
-- But that's complex in workflow JSON. Instead, we add a note to the prompt
-- telling the LLM to set handler_agent in the suggestion itself.

-- Pragmatic fix: update the prompt to tell the LLM to include
-- tool_component_id when suggesting a library tool. Then the
-- create_work_item action passes spec_data (the whole suggestion)
-- which already contains tool_component_id for library tools.
-- tool-deployer reads tool_component_id from spec.
-- For novel tools (no tool_component_id), we change handler_agent
-- to tool-generator.

-- HOWEVER — the loop sub_workflow can't conditionally set handler_agent.
-- So the simplest robust fix: change handler_agent in the loop to a
-- Go action that examines the suggestion and creates the right item.
-- This keeps the workflow simple.

-- For this session, we do the minimal safe fix:
-- 1. Update prompt to be clear about zero suggestions
-- 2. Update prompt to include tool ID when suggesting library tools
-- 3. Keep routing to tool-deployer for now (novel tools will fail
--    gracefully since tool-deployer checks for tool_component_id)
-- 4. Document the routing fix as next step

-- The prompt patch (applied to existing tool-suggester definition):
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(
                'You are evaluating what interactive tools would genuinely help visitors to this website.

        ## Site Context
        Domain: {{.input_data.domain}}
        {{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}} — {{.site_specs.specs.identity.sub_industry}}
        Company: {{.site_specs.specs.identity.company_name}}
        About: {{.site_specs.specs.identity.about_summary}}
        Services: {{if .site_specs.specs.identity.services}}{{range .site_specs.specs.identity.services}}
          - {{.name}}: {{.description}}{{end}}{{end}}
        Target Audience: {{.site_specs.specs.identity.target_audience}}{{end}}
        {{if .site_specs.specs.classification}}Site Type: {{.site_specs.specs.classification.site_type}}{{end}}

        ## Existing Pages
        {{if .pages}}{{range .pages}}- {{.name}} ({{.url}}): {{.title}}
        {{end}}{{else}}No pages loaded.{{end}}

        ## Already Deployed Tools
        {{if .existing_tools}}{{range .existing_tools}}- {{.display_name}} ({{.function}})
        {{end}}{{else}}None deployed yet.{{end}}

        ## Available in Library (can be forked and customised)
        {{if .library_tools}}{{range .library_tools}}- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}
        {{end}}{{else}}No library tools available.{{end}}

        ## Your Task

        Suggest 0-5 interactive tools that would genuinely add value for this website. Think about what someone in this industry would actually use day-to-day.

        IMPORTANT RULES:
        1. If NO tools would genuinely help this audience, return an EMPTY suggestions array. Not every site needs tools. A gas wholesaler does not need a password strength checker. A veterinary practice does not need an A/B test calculator.
        2. Only suggest tools that are directly relevant to this specific industry and audience. Generic "nice to have" tools that any site could use are NOT good suggestions.
        3. Prefer library tools when they fit. For library tools, include the id field in tool_component_id.
        4. When no library tool fits but a custom tool would genuinely help, suggest it with library_source: null and tool_component_id: null. Be specific about what it should do.

        Examples of GOOD suggestions:
        - Gas wholesaler: unit converter (therms to kWh to MJ), delivery cost estimator
        - Photographer: booking calculator, print size guide
        - Accountant: VAT calculator, tax deadline calendar
        - Restaurant: tip calculator, allergen filter

        Examples of BAD suggestions (do not do this):
        - Password checker for a gas wholesaler (irrelevant to their audience)
        - A/B test calculator for a restaurant (their visitors are diners, not marketers)
        - Meme generator for an accountancy firm (not professional)

        Return ONLY valid JSON:
        {
          "reasoning": "Brief explanation of why these tools suit (or do not suit) this industry",
          "suggestions": [
            {
              "name": "Human-readable tool name",
              "function": "tool-kebab-case-name",
              "description": "What the tool does and why it helps visitors",
              "priority": 1,
              "target_page": "name of best page, or new for a dedicated tools page",
              "library_source": "function name from library if forkable, or null if new",
              "tool_component_id": "uuid from library if forkable, or null if new",
              "complexity": "simple"
            }
          ]
        }'::text
        ),
        updated_at = NOW()
            WHERE type = 'tool-suggester';


-- Verify the prompt was updated
SELECT LENGTH(default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template') as prompt_len
FROM agent_definitions
WHERE type = 'tool-suggester';

---

-- ============================================================
-- Patch tool-suggester: write evaluation to site_specs
-- ============================================================
-- Adds a save_tool_spec step between suggest_tools and
-- check_has_suggestions. The LLM evaluation (reasoning,
-- suggestions, rejected tools) is persisted as site_specs
-- aspect = 'tools'. This gives auditors a contract to check
-- against and makes tool decisions visible in the admin panel.
--
-- Also updates the LLM prompt to include rejected_tools in
-- the JSON output so the spec records what was considered
-- and why it was rejected.
--
-- Run AFTER fix_tool_agent_definitions.sql (which fixes
-- columns, params, paths) and patch_tool_suggester_prompt.sql
-- (which strengthens the prompt).

-- Step 1: Add save_tool_spec step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,save_tool_spec}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "site_record.site_id",
                "spec_data": "evaluation.result",
                "aspect": "tools",
                "source": "tool-suggester",
                "source_agent": "tool-suggester",
                "notes": "LLM tool evaluation — reasoning, recommendations, and rejections"
            },
            "output_field": "tool_spec_written",
            "next_step": "check_has_suggestions",
            "error_step": "check_has_suggestions",
            "description": "Persist tool evaluation to site_specs for auditor visibility"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'tool-suggester';

-- Step 2: Point suggest_tools → save_tool_spec (instead of → check_has_suggestions)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,next_step}',
        '"save_tool_spec"'
                     ),
    updated_at = NOW()
WHERE type = 'tool-suggester';

-- Step 3: Update the prompt to include rejected_tools in the output
-- The prompt already allows 0 suggestions. We add rejected_tools so
-- the spec records what was considered and explicitly rejected.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(
                'You are evaluating what interactive tools would genuinely help visitors to this website.

        ## Site Context
        Domain: {{.input_data.domain}}
        {{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}} — {{.site_specs.specs.identity.sub_industry}}
        Company: {{.site_specs.specs.identity.company_name}}
        About: {{.site_specs.specs.identity.about_summary}}
        Services: {{if .site_specs.specs.identity.services}}{{range .site_specs.specs.identity.services}}
          - {{.name}}: {{.description}}{{end}}{{end}}
        Target Audience: {{.site_specs.specs.identity.target_audience}}{{end}}
        {{if .site_specs.specs.classification}}Site Type: {{.site_specs.specs.classification.site_type}}{{end}}

        ## Existing Pages
        {{if .pages}}{{range .pages}}- {{.name}} ({{.url}}): {{.title}}
        {{end}}{{else}}No pages loaded.{{end}}

        ## Already Deployed Tools
        {{if .existing_tools}}{{range .existing_tools}}- {{.display_name}} ({{.function}})
        {{end}}{{else}}None deployed yet.{{end}}

        ## Available in Library (can be forked and customised)
        {{if .library_tools}}{{range .library_tools}}- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}
        {{end}}{{else}}No library tools available.{{end}}

        ## Your Task

        Evaluate what interactive tools would genuinely add value for this website. Think about what someone in this industry would actually use day-to-day.

        IMPORTANT RULES:
        1. If NO tools would genuinely help this audience, return an EMPTY suggestions array. Not every site needs tools. A gas wholesaler does not need a password strength checker. A veterinary practice does not need an A/B test calculator.
        2. Only suggest tools that are directly relevant to this specific industry and audience. Generic "nice to have" tools that any site could use are NOT good suggestions.
        3. Prefer library tools when they fit. For library tools, include the id field in tool_component_id.
        4. When no library tool fits but a custom tool would genuinely help, suggest it with library_source: null and tool_component_id: null. Be specific about what it should do.
        5. Also list tools from the library that you considered but rejected, with a reason. This helps auditors understand the decision.

        Examples of GOOD suggestions:
        - Gas wholesaler: unit converter (therms to kWh to MJ), delivery cost estimator
        - Photographer: booking calculator, print size guide
        - Accountant: VAT calculator, tax deadline calendar
        - Restaurant: tip calculator, allergen filter
        - Mortgage broker: stamp duty calculator, affordability calculator, repayment calculator

        Examples of BAD suggestions (do not do this):
        - Password checker for a gas wholesaler (irrelevant to their audience)
        - A/B test calculator for a restaurant (their visitors are diners, not marketers)
        - Meme generator for an accountancy firm (not professional)

        Return ONLY valid JSON:
        {
          "reasoning": "Brief explanation of why these tools suit (or do not suit) this industry",
          "suggestions": [
            {
              "name": "Human-readable tool name",
              "function": "tool-kebab-case-name",
              "description": "What the tool does and why it helps visitors",
              "priority": 1,
              "target_page": "name of best page, or new for a dedicated tools page",
              "library_source": "function name from library if forkable, or null if new",
              "tool_component_id": "uuid from library if forkable, or null if new",
              "complexity": "simple"
            }
          ],
          "rejected_tools": [
            {
              "function": "tool-function-name",
              "reason": "Why this tool is not appropriate for this site"
            }
          ]
        }'::text
        ),
        updated_at = NOW()
            WHERE type = 'tool-suggester';


-- Step 4: Also add ensure_site_record before read_specs so save_tool_spec
-- has site_record.site_id available. Check if it already exists.
-- (fix_tool_agent_definitions.sql starts at read_specs, not ensure_site_record)

-- First check current start_step
SELECT default_config->'workflow'->>'start_step' as current_start
FROM agent_definitions WHERE type = 'tool-suggester';

-- Add ensure_site_record step and update start_step
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,ensure_site_record}',
                '{
                    "action": "ensure_site_record",
                    "config": {},
                    "next_step": "read_specs",
                    "output_field": "site_record"
                }'::jsonb
        ),
        '{workflow,start_step}',
        '"ensure_site_record"'
                     ),
    updated_at = NOW()
WHERE type = 'tool-suggester'
  AND NOT (default_config->'workflow'->'steps' ? 'ensure_site_record');


-- ============================================================
-- Verify
-- ============================================================

-- Check the workflow step order
SELECT
    key as step,
    value->>'next_step' as next_step,
    value->>'action' as action
FROM agent_definitions,
    jsonb_each(default_config->'workflow'->'steps')
WHERE type = 'tool-suggester'
ORDER BY
    CASE key
    WHEN 'ensure_site_record' THEN 1
    WHEN 'read_specs' THEN 2
    WHEN 'load_pages' THEN 3
    WHEN 'load_existing_tools' THEN 4
    WHEN 'load_library_tools' THEN 5
    WHEN 'suggest_tools' THEN 6
    WHEN 'save_tool_spec' THEN 7
    WHEN 'check_has_suggestions' THEN 8
    WHEN 'create_items_loop' THEN 9
    WHEN 'complete' THEN 10
END;

-- above didn't work

---

-- ============================================================
-- Fix: tool-suggester prompt + ensure_site_record
-- Single DO block to avoid client parsing issues
-- ============================================================

DO $$
DECLARE
prompt_text TEXT;
    current_config JSONB;
BEGIN
    -- Build the prompt in a variable (no quoting issues)
    prompt_text := E'You are evaluating what interactive tools would genuinely help visitors to this website.\n\n'
        || E'## Site Context\n'
        || E'Domain: {{.input_data.domain}}\n'
        || E'{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}} \u2014 {{.site_specs.specs.identity.sub_industry}}\n'
        || E'Company: {{.site_specs.specs.identity.company_name}}\n'
        || E'About: {{.site_specs.specs.identity.about_summary}}\n'
        || E'Services: {{if .site_specs.specs.identity.services}}{{range .site_specs.specs.identity.services}}\n'
        || E'  - {{.name}}: {{.description}}{{end}}{{end}}\n'
        || E'Target Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n'
        || E'{{if .site_specs.specs.classification}}Site Type: {{.site_specs.specs.classification.site_type}}{{end}}\n\n'
        || E'## Existing Pages\n'
        || E'{{if .pages}}{{range .pages}}- {{.name}} ({{.url}}): {{.title}}\n'
        || E'{{end}}{{else}}No pages loaded.{{end}}\n\n'
        || E'## Already Deployed Tools\n'
        || E'{{if .existing_tools}}{{range .existing_tools}}- {{.display_name}} ({{.function}})\n'
        || E'{{end}}{{else}}None deployed yet.{{end}}\n\n'
        || E'## Available in Library (can be forked and customised)\n'
        || E'{{if .library_tools}}{{range .library_tools}}- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}\n'
        || E'{{end}}{{else}}No library tools available.{{end}}\n\n'
        || E'## Your Task\n\n'
        || E'Evaluate what interactive tools would genuinely add value for this website. Think about what someone in this industry would actually use day-to-day.\n\n'
        || E'IMPORTANT RULES:\n'
        || E'1. If NO tools would genuinely help this audience, return an EMPTY suggestions array. Not every site needs tools.\n'
        || E'2. Only suggest tools that are directly relevant to this specific industry and audience.\n'
        || E'3. Prefer library tools when they fit. For library tools, include the id field in tool_component_id.\n'
        || E'4. When no library tool fits but a custom tool would genuinely help, suggest it with library_source: null and tool_component_id: null.\n'
        || E'5. Also list tools from the library that you considered but rejected, with a reason.\n\n'
        || E'Examples of GOOD suggestions:\n'
        || E'- Gas wholesaler: unit converter (therms to kWh to MJ), delivery cost estimator\n'
        || E'- Photographer: booking calculator, print size guide\n'
        || E'- Accountant: VAT calculator, tax deadline calendar\n'
        || E'- Mortgage broker: stamp duty calculator, affordability calculator, repayment calculator\n\n'
        || E'Examples of BAD suggestions (do not do this):\n'
        || E'- Password checker for a gas wholesaler (irrelevant to their audience)\n'
        || E'- A/B test calculator for a restaurant (their visitors are diners, not marketers)\n'
        || E'- Meme generator for an accountancy firm (not professional)\n\n'
        || E'Return ONLY valid JSON:\n'
        || E'{\n'
        || E'  "reasoning": "Brief explanation of why these tools suit (or do not suit) this industry",\n'
        || E'  "suggestions": [\n'
        || E'    {\n'
        || E'      "name": "Human-readable tool name",\n'
        || E'      "function": "tool-kebab-case-name",\n'
        || E'      "description": "What the tool does and why it helps visitors",\n'
        || E'      "priority": 1,\n'
        || E'      "target_page": "name of best page, or new for a dedicated tools page",\n'
        || E'      "library_source": "function name from library if forkable, or null if new",\n'
        || E'      "tool_component_id": "uuid from library if forkable, or null if new",\n'
        || E'      "complexity": "simple"\n'
        || E'    }\n'
        || E'  ],\n'
        || E'  "rejected_tools": [\n'
        || E'    {\n'
        || E'      "function": "tool-function-name",\n'
        || E'      "reason": "Why this tool is not appropriate for this site"\n'
        || E'    }\n'
        || E'  ]\n'
        || E'}';

    -- Load current config
SELECT default_config INTO current_config
FROM agent_definitions WHERE type = 'tool-suggester';

-- Update prompt
current_config := jsonb_set(
        current_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(prompt_text)
    );

    -- Add ensure_site_record if missing
    IF NOT (current_config->'workflow'->'steps' ? 'ensure_site_record') THEN
        current_config := jsonb_set(
            current_config,
            '{workflow,steps,ensure_site_record}',
            '{"action": "ensure_site_record", "config": {}, "next_step": "read_specs", "output_field": "site_record"}'::jsonb
        );
        current_config := jsonb_set(
            current_config,
            '{workflow,start_step}',
            '"ensure_site_record"'
        );
        RAISE NOTICE 'Added ensure_site_record step';
ELSE
        RAISE NOTICE 'ensure_site_record already exists';
END IF;

    -- Write back
UPDATE agent_definitions
SET default_config = current_config, updated_at = NOW()
WHERE type = 'tool-suggester';

RAISE NOTICE 'tool-suggester updated: prompt + ensure_site_record';
END;
$$;

-- Verify step chain
SELECT
    key as step,
    value->>'next_step' as next_step,
    value->>'action' as action
FROM agent_definitions,
    jsonb_each(default_config->'workflow'->'steps')
WHERE type = 'tool-suggester'
ORDER BY
    CASE key
    WHEN 'ensure_site_record' THEN 1
    WHEN 'read_specs' THEN 2
    WHEN 'load_pages' THEN 3
    WHEN 'load_existing_tools' THEN 4
    WHEN 'load_library_tools' THEN 5
    WHEN 'suggest_tools' THEN 6
    WHEN 'save_tool_spec' THEN 7
    WHEN 'check_has_suggestions' THEN 8
    WHEN 'create_items_loop' THEN 9
    WHEN 'complete' THEN 10
    WHEN 'complete_empty' THEN 11
END;

