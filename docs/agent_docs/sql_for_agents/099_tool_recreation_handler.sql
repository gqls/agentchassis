-- tool-recreation-handler agent definition
-- Routes: dispatch loop spawns this for work items with handler_agent = 'tool-recreation-handler'
-- Purpose: Recreate interactive tools/games from crawled rawHtml source code
--
-- Workflow:
--   ensure_site_record → load_page_record → check_page_found
--     → load_site_specs → load_existing_content → load_related_context
--     → analyze_tool (LLM: functional spec from source + context)
--     → recreate_tool (LLM/Opus: produce working HTML/CSS/JS)
--     → check_completeness (verify not truncated, strip marker)
--     → validate_tool → save_page_sections → update_page_status
--     → spawn_rerender → deploy_page → complete

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    container_image, image_tag,
    resource_config, topic_config, health_config,
    env_vars, version,
    delegation_config, role, status, tags,
    questionnaire, priority, is_singleton,
    input_contract, output_contract, timeout_seconds
) VALUES (
             gen_random_uuid(),
             'tool-recreation-handler',
             'Tool Recreation Handler',
             'Recreates interactive tools, games, and applications from crawled source code. Two-stage: analyses the tool purpose and function, then generates working replacement code. Used for adoption of sites with JavaScript-heavy interactive pages.',
             'specialist',
             -- default_config (workflow definition)
             $cfg${
  "workflow": {
    "start_step": "ensure_site_record",
             "processing_mode": "orchestrator",
             "timeout_seconds": 2400,
             "steps": {

      "ensure_site_record": {
        "action": "ensure_site_record",
             "config": { "store_brief_in_content_data": false },
             "next_step": "load_page_record",
             "description": "Load site record for context",
             "output_field": "site_record"
      },

             "load_page_record": {
        "action": "load_page_record",
             "config": {
          "site_id": "site_record.site_id",
             "page_name": "input_data.page_name"
        },
             "next_step": "check_page_found",
             "description": "Load page record from DB",
             "output_field": "page_record"
      },

             "check_page_found": {
        "action": "conditional",
             "config": {
          "condition": "page_record.found == true",
             "then_step": "load_site_specs",
             "else_step": "complete_error"
        },
             "description": "Check if page exists in DB"
      },

             "load_site_specs": {
        "action": "read_site_spec",
             "config": { "site_id": "site_record.site_id" },
             "next_step": "load_existing_content",
             "error_step": "load_existing_content",
             "description": "Load identity, archetype, content_direction, design specs",
             "output_field": "site_specs"
      },

             "load_existing_content": {
        "action": "load_existing_content",
             "config": {
          "mode": "input_data.spec.mode",
             "page_id": "page_record.id",
             "site_id": "site_record.site_id",
             "page_name": "page_record.name"
        },
             "next_step": "load_related_context",
             "error_step": "load_related_context",
             "description": "Load rawHtml and markdown from adoption crawl research_results",
             "output_field": "existing_content"
      },

             "load_related_context": {
        "action": "query_database",
             "config": {
          "query": "SELECT p.name, p.title, p.page_type, rr.summary FROM pages p LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = 'adoption_page' WHERE p.site_id = $1 AND p.name != $2 ORDER BY p.nav_order LIMIT 10",
             "params": ["site_record.site_id", "page_record.name"],
             "output_format": "array"
        },
             "next_step": "analyze_tool",
             "error_step": "analyze_tool",
             "description": "Load other pages on the site for context about how this tool fits",
             "output_field": "related_pages"
      },

             "analyze_tool": {
        "action": "execute_llm_prompt",
             "config": {
          "ai_service": {
            "provider": "anthropic",
             "model": "claude-sonnet-4-6",
             "api_key_env_var": "ANTHROPIC_API_KEY",
             "max_tokens": 8000
          },
             "temperature": 0.2,
             "input_fields": ["site_record", "site_specs", "page_record", "existing_content", "related_pages", "input_data"],
             "output_format": "json",
             "prompt_template": "You are a senior software analyst specialising in interactive web applications.\n\nYou are examining an existing interactive tool or game from a website that is being adopted into our system. Your job is to produce a detailed functional specification so that a developer can recreate this tool as a working, self-contained HTML/CSS/JavaScript application.\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity}}Company: {{.site_specs.specs.identity.company_name}}\nIndustry: {{.site_specs.specs.identity.industry}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n{{if .site_specs.specs.site_archetype}}## Site Archetype\nLabel: {{.site_specs.specs.site_archetype.label}}\n{{if .site_specs.specs.site_archetype.purpose}}Purpose: {{.site_specs.specs.site_archetype.purpose}}{{end}}\n{{if .site_specs.specs.site_archetype.audience}}Audience: {{.site_specs.specs.site_archetype.audience}}{{end}}\n{{if .site_specs.specs.site_archetype.interaction_patterns}}Interaction Patterns: {{.site_specs.specs.site_archetype.interaction_patterns}}{{end}}\n{{if .site_specs.specs.site_archetype.constraints}}Constraints (things we must NOT change): {{.site_specs.specs.site_archetype.constraints}}{{end}}{{end}}\n\n{{if .site_specs.specs.content_direction.formatted}}## Content Direction\n{{.site_specs.specs.content_direction.formatted}}{{end}}\n\n## This Page\nPage name: {{.page_record.name}}\nPage title: {{.page_record.title}}\nPage type: {{.page_record.page_type}}\n\n{{if .input_data.spec.interactive_features}}## Interactive Features Identified During Adoption\n{{range .input_data.spec.interactive_features}}- Name: {{.name}}\n  Type: {{.type}}\n  Description: {{.description}}\n  Page: {{.page}}\n{{end}}{{end}}\n\n## Other Pages on This Site (for context)\n{{if .related_pages}}{{range .related_pages}}- {{.name}} ({{.page_type}}): {{.title}}\n{{end}}{{else}}No other pages loaded.{{end}}\n\n## Original Page Content (Markdown)\n{{if .existing_content.raw_markdown}}{{.existing_content.raw_markdown}}{{else}}No markdown content available.{{end}}\n\n## Original Page Source Code (HTML)\n{{if .existing_content.existing_content.raw_html}}{{.existing_content.existing_content.raw_html}}{{else}}No rawHtml source available — analysis will be based on markdown only.{{end}}\n\n## Your Task\n\nAnalyse this interactive tool/game thoroughly. Produce a JSON functional specification that captures:\n\n1. **What it does** — the purpose, the problem it solves for the user\n2. **Who uses it** — the target user persona and their goals when using this tool\n3. **How it works** — step by step interaction flow from the user's perspective\n4. **Technical implementation** — what JavaScript patterns, algorithms, or libraries it uses\n5. **Visual design** — layout structure, color usage, typography, responsive behaviour\n6. **Data model** — what inputs the user provides, what outputs are generated, any state management\n7. **Edge cases** — what happens with empty input, extreme values, errors\n8. **External dependencies** — any CDN libraries, APIs, or assets it relies on\n9. **How it connects to the site** — what content references this tool, why it exists on this site\n10. **Improvement opportunities** — what could make this tool more useful, more polished, or more educational for the target audience (but keep the core functionality identical)\n\nReturn ONLY valid JSON:\n{\n  \"tool_name\": \"Human-readable name\",\n  \"tool_type\": \"calculator|simulator|game|visualisation|editor|other\",\n  \"purpose\": \"What this tool does and why it matters to the target user\",\n  \"target_user\": \"Who uses this and what they're trying to accomplish\",\n  \"interaction_flow\": [\n    \"Step 1: User does X\",\n    \"Step 2: Tool responds with Y\"\n  ],\n  \"technical_spec\": {\n    \"algorithms\": [\"Description of key algorithms or logic\"],\n    \"state_management\": \"How state is tracked (variables, objects, etc)\",\n    \"event_handling\": [\"Click handlers, input listeners, timers, etc\"],\n    \"rendering\": \"How the UI updates — DOM manipulation, canvas, SVG, etc\"\n  },\n  \"visual_spec\": {\n    \"layout\": \"Description of the layout structure\",\n    \"color_scheme\": \"Colors used and their purpose\",\n    \"typography\": \"Font choices and sizing\",\n    \"responsive\": \"How it adapts to different screen sizes\",\n    \"animations\": \"Any transitions or animations\"\n  },\n  \"data_model\": {\n    \"inputs\": [{\"name\": \"field_name\", \"type\": \"number|text|select\", \"description\": \"What this controls\", \"default\": \"default value\", \"constraints\": \"min/max/validation\"}],\n    \"outputs\": [{\"name\": \"output_name\", \"type\": \"text|chart|table|visual\", \"description\": \"What this shows\"}],\n    \"internal_state\": [\"Key state variables and what they track\"]\n  },\n  \"edge_cases\": [\"What happens when...\"],\n  \"dependencies\": {\n    \"external_libs\": [\"Library name and version if identifiable\"],\n    \"apis\": [\"Any external API calls\"],\n    \"assets\": [\"Images, fonts, or other assets needed\"]\n  },\n  \"site_context\": \"How this tool fits the site's purpose and what other content references it\",\n  \"improvement_notes\": [\"Specific, actionable improvements that would enhance the tool for its target audience\"]\n}"
        },
             "next_step": "recreate_tool",
             "error_step": "complete_error",
             "description": "Analyse the original tool to produce a detailed functional specification",
             "output_field": "tool_analysis"
      },

             "recreate_tool": {
        "action": "execute_llm_prompt",
             "config": {
          "ai_service": {
            "provider": "anthropic",
             "model": "claude-opus-4-6",
             "api_key_env_var": "ANTHROPIC_API_KEY",
             "max_tokens": 64000
          },
             "temperature": 0.1,
             "input_fields": ["site_record", "site_specs", "page_record", "existing_content", "tool_analysis", "input_data"],
             "output_format": "text",
             "prompt_template": "You are a senior frontend developer with deep expertise in building interactive web tools, games, simulators, and data visualisations. You write production-quality code that works correctly on the first deploy. You have extensive experience in the domain of {{.site_specs.specs.identity.industry}}.\n\nYou are recreating an interactive tool for the website {{.site_record.domain}}.\n\n## Functional Specification\nThis specification was produced by analysing the original tool. Follow it precisely.\n\n{{.tool_analysis.result | toJSON}}\n\n## Design Context\n{{if .site_specs.specs.design}}Original site design:\n{{.site_specs.specs.design}}{{end}}\n{{if .site_specs.specs.site_archetype.visual_character}}Visual character: {{.site_specs.specs.site_archetype.visual_character}}{{end}}\n\n## Original Source Code (REFERENCE ONLY)\nStudy this carefully to understand the implementation. Your recreation should achieve the same functionality but with clean, well-structured code.\n\n{{if .existing_content.existing_content.raw_html}}{{.existing_content.existing_content.raw_html}}{{else}}No source code available — build from the functional specification above.{{end}}\n\n## Requirements\n\n### Mandatory\n1. The tool MUST be fully functional — every button, input, calculation, animation, and interaction must work\n2. All JavaScript must be embedded in the HTML file (no external JS files)\n3. All CSS must be embedded in the HTML file (no external CSS files) \n4. The output is the INNER content of the page — do NOT include <html>, <head>, <body>, or <!DOCTYPE> tags. Start with the tool's heading and content. The site header, footer, navigation, and base CSS will be injected by the deployment system\n5. Use CSS custom properties where possible (--primary-color, --bg-color, etc) so the site theme can override them\n6. The tool must be responsive — usable on mobile screens\n7. No placeholder functions — every function must have a complete implementation\n8. No TODO comments — everything must be finished\n9. No fake data or dummy outputs — calculations must be mathematically correct\n10. Error handling for all user inputs — graceful behaviour with empty, zero, negative, or extreme values\n\n### Code Quality\n- Use modern JavaScript (ES6+) — const/let, arrow functions, template literals, destructuring\n- Meaningful variable and function names\n- Comments explaining non-obvious logic (algorithms, formulas, game mechanics)\n- Clean separation: tool heading/description at the top, then the interactive widget, then any help text\n- If the tool uses canvas or complex rendering, ensure proper cleanup and no memory leaks\n\n### Visual Quality\n- Match the original tool's visual style as closely as possible\n- Use the site's colour palette via CSS custom properties\n- Smooth transitions and feedback for user interactions\n- Loading states where calculations take time\n- Clear labels on all inputs and outputs\n\n### Structure\nProduce the HTML in this structure:\n```\n<div class=\"tool-page\">\n  <div class=\"tool-header\">\n    <h1>Tool Name</h1>\n    <p class=\"tool-description\">What this tool does and why it's useful</p>\n  </div>\n  <div class=\"tool-container\">\n    <!-- The interactive tool itself -->\n  </div>\n  <div class=\"tool-info\">\n    <!-- Optional: how to use it, what the results mean, etc -->\n  </div>\n</div>\n<style>\n  /* All CSS here, using custom properties */\n</style>\n<script>\n  /* All JavaScript here, in an IIFE or module pattern to avoid globals */\n</script>\n```\n\n## Output\nReturn ONLY the HTML content. No markdown code fences. No explanation text before or after. Just the HTML starting with <div class=\"tool-page\">.\n\n### Size Management\nIf the tool is very complex, prioritise:\n1. All core functionality working correctly\n2. All user interactions responding properly\n3. Clean, readable code\n4. Visual polish and animations (can be simplified if needed)\n\nNever sacrifice working functionality for visual effects. If you must simplify, simplify CSS animations and visual flourishes, not logic.\n\n### Completion Marker\nYou MUST end your output with this exact comment on its own line after the closing </script> tag:\n<!-- tool-recreation-complete -->\n\nIf your output does not end with this marker, it means the code was truncated and is incomplete."
        },
             "next_step": "check_completeness",
             "error_step": "complete_error",
             "description": "Generate working HTML/CSS/JS for the interactive tool",
             "output_field": "tool_recreation"
      },

             "check_completeness": {
        "action": "check_tool_completeness",
             "config": {
          "html_field": "tool_recreation.result"
        },
             "next_step": "validate_tool",
             "error_step": "complete_error",
             "description": "Verify output is complete and not truncated, strip completion marker",
             "output_field": "completeness_check"
      },

             "validate_tool": {
        "action": "validate_page_content",
             "config": {
          "domain": "site_record.domain",
             "html_field": "completeness_check.clean_html"
        },
             "next_step": "save_sections",
             "error_step": "save_sections",
             "description": "Check for obvious issues — cross-site contamination, empty output",
             "output_field": "validation_result"
      },

             "save_sections": {
        "action": "save_page_sections",
             "config": {
          "html_field": "validation_result.clean_html",
             "site_id_field": "site_record.site_id",
             "page_name_field": "page_record.name"
        },
             "next_step": "update_status",
             "error_step": "complete_error",
             "description": "Persist generated tool HTML to page_components",
             "output_field": "sections_saved"
      },

             "update_status": {
        "action": "update_page_status",
             "config": {
          "status": "deployed",
             "site_id_field": "site_record.site_id",
             "page_name_field": "page_record.name"
        },
             "next_step": "spawn_rerender",
             "description": "Mark page as deployed",
             "output_field": "status_updated"
      },

             "spawn_rerender": {
        "action": "spawn_agent",
             "config": {
          "role": "page_renderer",
             "agent_type": "page-rerender"
        },
             "next_step": "deploy_page",
             "description": "Spawn page-rerender for assembly and deploy",
             "output_field": "rerender_agent"
      },

             "deploy_page": {
        "action": "call_agent",
             "config": {
          "target_role": "page_renderer",
             "input_mapping": {
            "site_id": "site_record.site_id",
             "page_id": "page_record.id",
             "domain": "site_record.domain"
          },
             "timeout_seconds": 120
        },
             "next_step": "complete",
             "error_step": "complete_error",
             "description": "Assemble page from stored components and deploy to git",
             "output_field": "deploy_result"
      },

             "complete": {
        "action": "complete_workflow",
             "config": {
          "output_fields": ["tool_analysis", "sections_saved", "deploy_result"]
        },
             "description": "Tool recreation complete"
      },

             "complete_error": {
        "action": "complete_workflow",
             "config": {
          "output_fields": ["tool_analysis", "tool_recreation", "site_record"],
             "success_message": "Tool recreation completed with errors"
        },
             "description": "Tool recreation completed with errors"
      }
    }
  }
}$cfg$::jsonb,
             true,  -- is_active
             '["interactive", "tools", "games", "code-generation"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.941',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'specialist',
             'active',
             '["build", "interactive", "handler"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"optional": ["page_name", "page_id", "sections"], "required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"tool_analysis": "Functional specification of the tool", "deploy_result": "git commit result", "sections_saved": "save result with page_id"}}'::jsonb,
             180
         );

---
-- fixed

-- tool-recreation-handler agent definition
-- Fixed column names to match actual schema

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag,
    resources, topics, health_config,
    env_vars, version,
    delegation_preferences, agent_category, status, domain_tags,
    briefing_questionnaire, usage_count, is_snapshot,
    input_contract, output_contract, idle_timeout_seconds
) VALUES (
             gen_random_uuid(),
             'tool-recreation-handler',
             'Tool Recreation Handler',
             'Recreates interactive tools, games, and applications from crawled source code. Two-stage: analyses the tool purpose and function, then generates working replacement code. Used for adoption of sites with JavaScript-heavy interactive pages.',
             'specialist',
             $cfg${
  "workflow": {
    "start_step": "ensure_site_record",
             "processing_mode": "orchestrator",
             "timeout_seconds": 2400,
             "steps": {

      "ensure_site_record": {
        "action": "ensure_site_record",
             "config": { "store_brief_in_content_data": false },
             "next_step": "load_page_record",
             "description": "Load site record for context",
             "output_field": "site_record"
      },

             "load_page_record": {
        "action": "load_page_record",
             "config": {
          "site_id": "site_record.site_id",
             "page_name": "input_data.page_name"
        },
             "next_step": "check_page_found",
             "description": "Load page record from DB",
             "output_field": "page_record"
      },

             "check_page_found": {
        "action": "conditional",
             "config": {
          "condition": "page_record.found == true",
             "then_step": "load_site_specs",
             "else_step": "complete_error"
        },
             "description": "Check if page exists in DB"
      },

             "load_site_specs": {
        "action": "read_site_spec",
             "config": { "site_id": "site_record.site_id" },
             "next_step": "load_existing_content",
             "error_step": "load_existing_content",
             "description": "Load identity, archetype, content_direction, design specs",
             "output_field": "site_specs"
      },

             "load_existing_content": {
        "action": "load_existing_content",
             "config": {
          "mode": "input_data.spec.mode",
             "page_id": "page_record.id",
             "site_id": "site_record.site_id",
             "page_name": "page_record.name"
        },
             "next_step": "load_related_context",
             "error_step": "load_related_context",
             "description": "Load rawHtml and markdown from adoption crawl research_results",
             "output_field": "existing_content"
      },

             "load_related_context": {
        "action": "query_database",
             "config": {
          "query": "SELECT p.name, p.title, p.page_type, rr.summary FROM pages p LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = 'adoption_page' WHERE p.site_id = $1 AND p.name != $2 ORDER BY p.nav_order LIMIT 10",
             "params": ["site_record.site_id", "page_record.name"],
             "output_format": "array"
        },
             "next_step": "analyze_tool",
             "error_step": "analyze_tool",
             "description": "Load other pages on the site for context about how this tool fits",
             "output_field": "related_pages"
      },

             "analyze_tool": {
        "action": "execute_llm_prompt",
             "config": {
          "ai_service": {
            "provider": "anthropic",
             "model": "claude-sonnet-4-6",
             "api_key_env_var": "ANTHROPIC_API_KEY",
             "max_tokens": 8000
          },
             "temperature": 0.2,
             "input_fields": ["site_record", "site_specs", "page_record", "existing_content", "related_pages", "input_data"],
             "output_format": "json",
             "prompt_template": "You are a senior software analyst specialising in interactive web applications.\n\nYou are examining an existing interactive tool or game from a website that is being adopted into our system. Your job is to produce a detailed functional specification so that a developer can recreate this tool as a working, self-contained HTML/CSS/JavaScript application.\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity}}Company: {{.site_specs.specs.identity.company_name}}\nIndustry: {{.site_specs.specs.identity.industry}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n{{if .site_specs.specs.site_archetype}}## Site Archetype\nLabel: {{.site_specs.specs.site_archetype.label}}\n{{if .site_specs.specs.site_archetype.purpose}}Purpose: {{.site_specs.specs.site_archetype.purpose}}{{end}}\n{{if .site_specs.specs.site_archetype.audience}}Audience: {{.site_specs.specs.site_archetype.audience}}{{end}}\n{{if .site_specs.specs.site_archetype.interaction_patterns}}Interaction Patterns: {{.site_specs.specs.site_archetype.interaction_patterns}}{{end}}\n{{if .site_specs.specs.site_archetype.constraints}}Constraints (things we must NOT change): {{.site_specs.specs.site_archetype.constraints}}{{end}}{{end}}\n\n{{if .site_specs.specs.content_direction.formatted}}## Content Direction\n{{.site_specs.specs.content_direction.formatted}}{{end}}\n\n## This Page\nPage name: {{.page_record.name}}\nPage title: {{.page_record.title}}\nPage type: {{.page_record.page_type}}\n\n{{if .input_data.spec.interactive_features}}## Interactive Features Identified During Adoption\n{{range .input_data.spec.interactive_features}}- Name: {{.name}}\n  Type: {{.type}}\n  Description: {{.description}}\n  Page: {{.page}}\n{{end}}{{end}}\n\n## Other Pages on This Site (for context)\n{{if .related_pages}}{{range .related_pages}}- {{.name}} ({{.page_type}}): {{.title}}\n{{end}}{{else}}No other pages loaded.{{end}}\n\n## Original Page Content (Markdown)\n{{if .existing_content.raw_markdown}}{{.existing_content.raw_markdown}}{{else}}No markdown content available.{{end}}\n\n## Original Page Source Code (HTML)\n{{if .existing_content.existing_content.raw_html}}{{.existing_content.existing_content.raw_html}}{{else}}No rawHtml source available — analysis will be based on markdown only.{{end}}\n\n## Your Task\n\nAnalyse this interactive tool/game thoroughly. Produce a JSON functional specification that captures:\n\n1. **What it does** — the purpose, the problem it solves for the user\n2. **Who uses it** — the target user persona and their goals when using this tool\n3. **How it works** — step by step interaction flow from the user's perspective\n4. **Technical implementation** — what JavaScript patterns, algorithms, or libraries it uses\n5. **Visual design** — layout structure, color usage, typography, responsive behaviour\n6. **Data model** — what inputs the user provides, what outputs are generated, any state management\n7. **Edge cases** — what happens with empty input, extreme values, errors\n8. **External dependencies** — any CDN libraries, APIs, or assets it relies on\n9. **How it connects to the site** — what content references this tool, why it exists on this site\n10. **Improvement opportunities** — what could make this tool more useful, more polished, or more educational for the target audience (but keep the core functionality identical)\n\nReturn ONLY valid JSON:\n{\n  \"tool_name\": \"Human-readable name\",\n  \"tool_type\": \"calculator|simulator|game|visualisation|editor|other\",\n  \"purpose\": \"What this tool does and why it matters to the target user\",\n  \"target_user\": \"Who uses this and what they're trying to accomplish\",\n  \"interaction_flow\": [\n    \"Step 1: User does X\",\n    \"Step 2: Tool responds with Y\"\n  ],\n  \"technical_spec\": {\n    \"algorithms\": [\"Description of key algorithms or logic\"],\n    \"state_management\": \"How state is tracked (variables, objects, etc)\",\n    \"event_handling\": [\"Click handlers, input listeners, timers, etc\"],\n    \"rendering\": \"How the UI updates — DOM manipulation, canvas, SVG, etc\"\n  },\n  \"visual_spec\": {\n    \"layout\": \"Description of the layout structure\",\n    \"color_scheme\": \"Colors used and their purpose\",\n    \"typography\": \"Font choices and sizing\",\n    \"responsive\": \"How it adapts to different screen sizes\",\n    \"animations\": \"Any transitions or animations\"\n  },\n  \"data_model\": {\n    \"inputs\": [{\"name\": \"field_name\", \"type\": \"number|text|select\", \"description\": \"What this controls\", \"default\": \"default value\", \"constraints\": \"min/max/validation\"}],\n    \"outputs\": [{\"name\": \"output_name\", \"type\": \"text|chart|table|visual\", \"description\": \"What this shows\"}],\n    \"internal_state\": [\"Key state variables and what they track\"]\n  },\n  \"edge_cases\": [\"What happens when...\"],\n  \"dependencies\": {\n    \"external_libs\": [\"Library name and version if identifiable\"],\n    \"apis\": [\"Any external API calls\"],\n    \"assets\": [\"Images, fonts, or other assets needed\"]\n  },\n  \"site_context\": \"How this tool fits the site's purpose and what other content references it\",\n  \"improvement_notes\": [\"Specific, actionable improvements that would enhance the tool for its target audience\"]\n}"
        },
             "next_step": "recreate_tool",
             "error_step": "complete_error",
             "description": "Analyse the original tool to produce a detailed functional specification",
             "output_field": "tool_analysis"
      },

             "recreate_tool": {
        "action": "execute_llm_prompt",
             "config": {
          "ai_service": {
            "provider": "anthropic",
             "model": "claude-opus-4-6",
             "api_key_env_var": "ANTHROPIC_API_KEY",
             "max_tokens": 64000
          },
             "temperature": 0.1,
             "input_fields": ["site_record", "site_specs", "page_record", "existing_content", "tool_analysis", "input_data"],
             "output_format": "text",
             "prompt_template": "You are a senior frontend developer with deep expertise in building interactive web tools, games, simulators, and data visualisations. You write production-quality code that works correctly on the first deploy. You have extensive experience in the domain of {{.site_specs.specs.identity.industry}}.\n\nYou are recreating an interactive tool for the website {{.site_record.domain}}.\n\n## Functional Specification\nThis specification was produced by analysing the original tool. Follow it precisely.\n\n{{.tool_analysis.result | toJSON}}\n\n## Design Context\n{{if .site_specs.specs.design}}Original site design:\n{{.site_specs.specs.design}}{{end}}\n{{if .site_specs.specs.site_archetype.visual_character}}Visual character: {{.site_specs.specs.site_archetype.visual_character}}{{end}}\n\n## Original Source Code (REFERENCE ONLY)\nStudy this carefully to understand the implementation. Your recreation should achieve the same functionality but with clean, well-structured code.\n\n{{if .existing_content.existing_content.raw_html}}{{.existing_content.existing_content.raw_html}}{{else}}No source code available — build from the functional specification above.{{end}}\n\n## Requirements\n\n### Mandatory\n1. The tool MUST be fully functional — every button, input, calculation, animation, and interaction must work\n2. All JavaScript must be embedded in the HTML file (no external JS files)\n3. All CSS must be embedded in the HTML file (no external CSS files) \n4. The output is the INNER content of the page — do NOT include <html>, <head>, <body>, or <!DOCTYPE> tags. Start with the tool's heading and content. The site header, footer, navigation, and base CSS will be injected by the deployment system\n5. Use CSS custom properties where possible (--primary-color, --bg-color, etc) so the site theme can override them\n6. The tool must be responsive — usable on mobile screens\n7. No placeholder functions — every function must have a complete implementation\n8. No TODO comments — everything must be finished\n9. No fake data or dummy outputs — calculations must be mathematically correct\n10. Error handling for all user inputs — graceful behaviour with empty, zero, negative, or extreme values\n\n### Code Quality\n- Use modern JavaScript (ES6+) — const/let, arrow functions, template literals, destructuring\n- Meaningful variable and function names\n- Comments explaining non-obvious logic (algorithms, formulas, game mechanics)\n- Clean separation: tool heading/description at the top, then the interactive widget, then any help text\n- If the tool uses canvas or complex rendering, ensure proper cleanup and no memory leaks\n\n### Visual Quality\n- Match the original tool's visual style as closely as possible\n- Use the site's colour palette via CSS custom properties\n- Smooth transitions and feedback for user interactions\n- Loading states where calculations take time\n- Clear labels on all inputs and outputs\n\n### Structure\nProduce the HTML in this structure:\n```\n<div class=\"tool-page\">\n  <div class=\"tool-header\">\n    <h1>Tool Name</h1>\n    <p class=\"tool-description\">What this tool does and why it's useful</p>\n  </div>\n  <div class=\"tool-container\">\n    <!-- The interactive tool itself -->\n  </div>\n  <div class=\"tool-info\">\n    <!-- Optional: how to use it, what the results mean, etc -->\n  </div>\n</div>\n<style>\n  /* All CSS here, using custom properties */\n</style>\n<script>\n  /* All JavaScript here, in an IIFE or module pattern to avoid globals */\n</script>\n```\n\n## Output\nReturn ONLY the HTML content. No markdown code fences. No explanation text before or after. Just the HTML starting with <div class=\"tool-page\">.\n\n### Size Management\nIf the tool is very complex, prioritise:\n1. All core functionality working correctly\n2. All user interactions responding properly\n3. Clean, readable code\n4. Visual polish and animations (can be simplified if needed)\n\nNever sacrifice working functionality for visual effects. If you must simplify, simplify CSS animations and visual flourishes, not logic.\n\n### Completion Marker\nYou MUST end your output with this exact comment on its own line after the closing </script> tag:\n<!-- tool-recreation-complete -->\n\nIf your output does not end with this marker, it means the code was truncated and is incomplete."
        },
             "next_step": "check_completeness",
             "error_step": "complete_error",
             "description": "Generate working HTML/CSS/JS for the interactive tool",
             "output_field": "tool_recreation"
      },

             "check_completeness": {
        "action": "check_tool_completeness",
             "config": {
          "html_field": "tool_recreation.result"
        },
             "next_step": "validate_tool",
             "error_step": "complete_error",
             "description": "Verify output is complete and not truncated, strip completion marker",
             "output_field": "completeness_check"
      },

             "validate_tool": {
        "action": "validate_page_content",
             "config": {
          "domain": "site_record.domain",
             "html_field": "completeness_check.clean_html"
        },
             "next_step": "save_sections",
             "error_step": "save_sections",
             "description": "Check for obvious issues — cross-site contamination, empty output",
             "output_field": "validation_result"
      },

             "save_sections": {
        "action": "save_page_sections",
             "config": {
          "html_field": "validation_result.clean_html",
             "site_id_field": "site_record.site_id",
             "page_name_field": "page_record.name"
        },
             "next_step": "update_status",
             "error_step": "complete_error",
             "description": "Persist generated tool HTML to page_components",
             "output_field": "sections_saved"
      },

             "update_status": {
        "action": "update_page_status",
             "config": {
          "status": "deployed",
             "site_id_field": "site_record.site_id",
             "page_name_field": "page_record.name"
        },
             "next_step": "spawn_rerender",
             "description": "Mark page as deployed",
             "output_field": "status_updated"
      },

             "spawn_rerender": {
        "action": "spawn_agent",
             "config": {
          "role": "page_renderer",
             "agent_type": "page-rerender"
        },
             "next_step": "deploy_page",
             "description": "Spawn page-rerender for assembly and deploy",
             "output_field": "rerender_agent"
      },

             "deploy_page": {
        "action": "call_agent",
             "config": {
          "target_role": "page_renderer",
             "input_mapping": {
            "site_id": "site_record.site_id",
             "page_id": "page_record.id",
             "domain": "site_record.domain"
          },
             "timeout_seconds": 120
        },
             "next_step": "complete",
             "error_step": "complete_error",
             "description": "Assemble page from stored components and deploy to git",
             "output_field": "deploy_result"
      },

             "complete": {
        "action": "complete_workflow",
             "config": {
          "output_fields": ["tool_analysis", "sections_saved", "deploy_result"]
        },
             "description": "Tool recreation complete"
      },

             "complete_error": {
        "action": "complete_workflow",
             "config": {
          "output_fields": ["tool_analysis", "tool_recreation", "site_record"],
             "success_message": "Tool recreation completed with errors"
        },
             "description": "Tool recreation completed with errors"
      }
    }
  }
}$cfg$::jsonb,
             true,
             '["interactive", "tools", "games", "code-generation"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.941',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'specialist',
             'active',
             '["build", "interactive", "handler"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"optional": ["page_name", "page_id", "sections"], "required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"tool_analysis": "Functional specification of the tool", "deploy_result": "git commit result", "sections_saved": "save result with page_id"}}'::jsonb,
             180
         );

---

-- tool-recreation-handler agent definition
-- Fixed column names to match actual schema

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag,
    resources, topics, health_config,
    env_vars, version,
    delegation_preferences, agent_category, status, domain_tags,
    briefing_questionnaire, usage_count, is_snapshot,
    input_contract, output_contract, idle_timeout_seconds
) VALUES (
             gen_random_uuid(),
             'tool-recreation-handler',
             'Tool Recreation Handler',
             'Recreates interactive tools, games, and applications from crawled source code. Two-stage: analyses the tool purpose and function, then generates working replacement code. Used for adoption of sites with JavaScript-heavy interactive pages.',
             'specialist',
             $cfg${
  "workflow": {
    "start_step": "ensure_site_record",
             "processing_mode": "orchestrator",
             "timeout_seconds": 2400,
             "steps": {

      "ensure_site_record": {
        "action": "ensure_site_record",
             "config": { "store_brief_in_content_data": false },
             "next_step": "load_page_record",
             "description": "Load site record for context",
             "output_field": "site_record"
      },

             "load_page_record": {
        "action": "load_page_record",
             "config": {
          "site_id": "site_record.site_id",
             "page_name": "input_data.page_name"
        },
             "next_step": "check_page_found",
             "description": "Load page record from DB",
             "output_field": "page_record"
      },

             "check_page_found": {
        "action": "conditional",
             "config": {
          "condition": "page_record.found == true",
             "then_step": "load_site_specs",
             "else_step": "complete_error"
        },
             "description": "Check if page exists in DB"
      },

             "load_site_specs": {
        "action": "read_site_spec",
             "config": { "site_id": "site_record.site_id" },
             "next_step": "load_existing_content",
             "error_step": "load_existing_content",
             "description": "Load identity, archetype, content_direction, design specs",
             "output_field": "site_specs"
      },

             "load_existing_content": {
        "action": "load_existing_content",
             "config": {
          "mode": "input_data.spec.mode",
             "page_id": "page_record.id",
             "site_id": "site_record.site_id",
             "page_name": "page_record.name"
        },
             "next_step": "load_related_context",
             "error_step": "load_related_context",
             "description": "Load rawHtml and markdown from adoption crawl research_results",
             "output_field": "existing_content"
      },

             "load_related_context": {
        "action": "query_database",
             "config": {
          "query": "SELECT p.name, p.title, p.page_type, rr.summary FROM pages p LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = 'adoption_page' WHERE p.site_id = $1 AND p.name != $2 ORDER BY p.nav_order LIMIT 10",
             "params": ["site_record.site_id", "page_record.name"],
             "output_format": "array"
        },
             "next_step": "analyze_tool",
             "error_step": "analyze_tool",
             "description": "Load other pages on the site for context about how this tool fits",
             "output_field": "related_pages"
      },

             "analyze_tool": {
        "action": "execute_llm_prompt",
             "config": {
          "ai_service": {
            "provider": "anthropic",
             "model": "claude-sonnet-4-6",
             "api_key_env_var": "ANTHROPIC_API_KEY",
             "max_tokens": 8000
          },
             "temperature": 0.2,
             "input_fields": ["site_record", "site_specs", "page_record", "existing_content", "related_pages", "input_data"],
             "output_format": "json",
             "prompt_template": "You are a senior software analyst specialising in interactive web applications.\n\nYou are examining an existing interactive tool or game from a website that is being adopted into our system. Your job is to produce a detailed functional specification so that a developer can recreate this tool as a working, self-contained HTML/CSS/JavaScript application.\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity}}Company: {{.site_specs.specs.identity.company_name}}\nIndustry: {{.site_specs.specs.identity.industry}}\nTarget Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n{{if .site_specs.specs.site_archetype}}## Site Archetype\nLabel: {{.site_specs.specs.site_archetype.label}}\n{{if .site_specs.specs.site_archetype.purpose}}Purpose: {{.site_specs.specs.site_archetype.purpose}}{{end}}\n{{if .site_specs.specs.site_archetype.audience}}Audience: {{.site_specs.specs.site_archetype.audience}}{{end}}\n{{if .site_specs.specs.site_archetype.interaction_patterns}}Interaction Patterns: {{.site_specs.specs.site_archetype.interaction_patterns}}{{end}}\n{{if .site_specs.specs.site_archetype.constraints}}Constraints (things we must NOT change): {{.site_specs.specs.site_archetype.constraints}}{{end}}{{end}}\n\n{{if .site_specs.specs.content_direction.formatted}}## Content Direction\n{{.site_specs.specs.content_direction.formatted}}{{end}}\n\n## This Page\nPage name: {{.page_record.name}}\nPage title: {{.page_record.title}}\nPage type: {{.page_record.page_type}}\n\n{{if .input_data.spec.interactive_features}}## Interactive Features Identified During Adoption\n{{range .input_data.spec.interactive_features}}- Name: {{.name}}\n  Type: {{.type}}\n  Description: {{.description}}\n  Page: {{.page}}\n{{end}}{{end}}\n\n## Other Pages on This Site (for context)\n{{if .related_pages}}{{range .related_pages}}- {{.name}} ({{.page_type}}): {{.title}}\n{{end}}{{else}}No other pages loaded.{{end}}\n\n## Original Page Content (Markdown)\n{{if .existing_content.raw_markdown}}{{.existing_content.raw_markdown}}{{else}}No markdown content available.{{end}}\n\n## Original Page Source Code (HTML)\n{{if .existing_content.existing_content.raw_html}}{{.existing_content.existing_content.raw_html}}{{else}}No rawHtml source available — analysis will be based on markdown only.{{end}}\n\n## Your Task\n\nAnalyse this interactive tool/game thoroughly. Produce a JSON functional specification that captures:\n\n1. **What it does** — the purpose, the problem it solves for the user\n2. **Who uses it** — the target user persona and their goals when using this tool\n3. **How it works** — step by step interaction flow from the user's perspective\n4. **Technical implementation** — what JavaScript patterns, algorithms, or libraries it uses\n5. **Visual design** — layout structure, color usage, typography, responsive behaviour\n6. **Data model** — what inputs the user provides, what outputs are generated, any state management\n7. **Edge cases** — what happens with empty input, extreme values, errors\n8. **External dependencies** — any CDN libraries, APIs, or assets it relies on\n9. **How it connects to the site** — what content references this tool, why it exists on this site\n10. **Improvement opportunities** — what could make this tool more useful, more polished, or more educational for the target audience (but keep the core functionality identical)\n\nReturn ONLY valid JSON:\n{\n  \"tool_name\": \"Human-readable name\",\n  \"tool_type\": \"calculator|simulator|game|visualisation|editor|other\",\n  \"purpose\": \"What this tool does and why it matters to the target user\",\n  \"target_user\": \"Who uses this and what they're trying to accomplish\",\n  \"interaction_flow\": [\n    \"Step 1: User does X\",\n    \"Step 2: Tool responds with Y\"\n  ],\n  \"technical_spec\": {\n    \"algorithms\": [\"Description of key algorithms or logic\"],\n    \"state_management\": \"How state is tracked (variables, objects, etc)\",\n    \"event_handling\": [\"Click handlers, input listeners, timers, etc\"],\n    \"rendering\": \"How the UI updates — DOM manipulation, canvas, SVG, etc\"\n  },\n  \"visual_spec\": {\n    \"layout\": \"Description of the layout structure\",\n    \"color_scheme\": \"Colors used and their purpose\",\n    \"typography\": \"Font choices and sizing\",\n    \"responsive\": \"How it adapts to different screen sizes\",\n    \"animations\": \"Any transitions or animations\"\n  },\n  \"data_model\": {\n    \"inputs\": [{\"name\": \"field_name\", \"type\": \"number|text|select\", \"description\": \"What this controls\", \"default\": \"default value\", \"constraints\": \"min/max/validation\"}],\n    \"outputs\": [{\"name\": \"output_name\", \"type\": \"text|chart|table|visual\", \"description\": \"What this shows\"}],\n    \"internal_state\": [\"Key state variables and what they track\"]\n  },\n  \"edge_cases\": [\"What happens when...\"],\n  \"dependencies\": {\n    \"external_libs\": [\"Library name and version if identifiable\"],\n    \"apis\": [\"Any external API calls\"],\n    \"assets\": [\"Images, fonts, or other assets needed\"]\n  },\n  \"site_context\": \"How this tool fits the site's purpose and what other content references it\",\n  \"improvement_notes\": [\"Specific, actionable improvements that would enhance the tool for its target audience\"]\n}"
        },
             "next_step": "recreate_tool",
             "error_step": "complete_error",
             "description": "Analyse the original tool to produce a detailed functional specification",
             "output_field": "tool_analysis"
      },

             "recreate_tool": {
        "action": "execute_llm_prompt",
             "config": {
          "ai_service": {
            "provider": "anthropic",
             "model": "claude-opus-4-6",
             "api_key_env_var": "ANTHROPIC_API_KEY",
             "max_tokens": 64000
          },
             "temperature": 0.1,
             "input_fields": ["site_record", "site_specs", "page_record", "existing_content", "tool_analysis", "input_data"],
             "output_format": "text",
             "prompt_template": "You are a senior frontend developer with deep expertise in building interactive web tools, games, simulators, and data visualisations. You write production-quality code that works correctly on the first deploy. You have extensive experience in the domain of {{.site_specs.specs.identity.industry}}.\n\nYou are recreating an interactive tool for the website {{.site_record.domain}}.\n\n## Functional Specification\nThis specification was produced by analysing the original tool. Follow it precisely.\n\n{{.tool_analysis.result | toJSON}}\n\n## Design Context\n{{if .site_specs.specs.design}}Original site design:\n{{.site_specs.specs.design}}{{end}}\n{{if .site_specs.specs.site_archetype.visual_character}}Visual character: {{.site_specs.specs.site_archetype.visual_character}}{{end}}\n\n## Original Source Code (REFERENCE ONLY)\nStudy this carefully to understand the implementation. Your recreation should achieve the same functionality but with clean, well-structured code.\n\n{{if .existing_content.existing_content.raw_html}}{{.existing_content.existing_content.raw_html}}{{else}}No source code available — build from the functional specification above.{{end}}\n\n## Requirements\n\n### Mandatory\n1. The tool MUST be fully functional — every button, input, calculation, animation, and interaction must work\n2. All JavaScript must be embedded in the HTML file (no external JS files)\n3. All CSS must be embedded in the HTML file (no external CSS files) \n4. The output is the INNER content of the page — do NOT include <html>, <head>, <body>, or <!DOCTYPE> tags. Start with the tool's heading and content. The site header, footer, navigation, and base CSS will be injected by the deployment system\n5. Use CSS custom properties where possible (--primary-color, --bg-color, etc) so the site theme can override them\n6. The tool must be responsive — usable on mobile screens\n7. No placeholder functions — every function must have a complete implementation\n8. No TODO comments — everything must be finished\n9. No fake data or dummy outputs — calculations must be mathematically correct\n10. Error handling for all user inputs — graceful behaviour with empty, zero, negative, or extreme values\n\n### Code Quality\n- Use modern JavaScript (ES6+) — const/let, arrow functions, template literals, destructuring\n- Meaningful variable and function names\n- Comments explaining non-obvious logic (algorithms, formulas, game mechanics)\n- Clean separation: tool heading/description at the top, then the interactive widget, then any help text\n- If the tool uses canvas or complex rendering, ensure proper cleanup and no memory leaks\n\n### Visual Quality\n- Match the original tool's visual style as closely as possible\n- Use the site's colour palette via CSS custom properties\n- Smooth transitions and feedback for user interactions\n- Loading states where calculations take time\n- Clear labels on all inputs and outputs\n\n### Structure\nProduce the HTML in this structure:\n```\n<div class=\"tool-page\">\n  <div class=\"tool-header\">\n    <h1>Tool Name</h1>\n    <p class=\"tool-description\">What this tool does and why it's useful</p>\n  </div>\n  <div class=\"tool-container\">\n    <!-- The interactive tool itself -->\n  </div>\n  <div class=\"tool-info\">\n    <!-- Optional: how to use it, what the results mean, etc -->\n  </div>\n</div>\n<style>\n  /* All CSS here, using custom properties */\n</style>\n<script>\n  /* All JavaScript here, in an IIFE or module pattern to avoid globals */\n</script>\n```\n\n## Output\nReturn ONLY the HTML content. No markdown code fences. No explanation text before or after. Just the HTML starting with <div class=\"tool-page\">.\n\n### Size Management\nIf the tool is very complex, prioritise:\n1. All core functionality working correctly\n2. All user interactions responding properly\n3. Clean, readable code\n4. Visual polish and animations (can be simplified if needed)\n\nNever sacrifice working functionality for visual effects. If you must simplify, simplify CSS animations and visual flourishes, not logic.\n\n### Completion Marker\nYou MUST end your output with this exact comment on its own line after the closing </script> tag:\n<!-- tool-recreation-complete -->\n\nIf your output does not end with this marker, it means the code was truncated and is incomplete."
        },
             "next_step": "check_completeness",
             "error_step": "complete_error",
             "description": "Generate working HTML/CSS/JS for the interactive tool",
             "output_field": "tool_recreation"
      },

             "check_completeness": {
        "action": "check_tool_completeness",
             "config": {
          "html_field": "tool_recreation.result"
        },
             "next_step": "save_training_data",
             "error_step": "complete_error",
             "description": "Verify output is complete and not truncated, strip completion marker",
             "output_field": "completeness_check"
      },

             "save_training_data": {
        "action": "save_tool_training_data",
             "config": {},
             "next_step": "validate_tool",
             "error_step": "validate_tool",
             "description": "Save source+spec+recreation triple for model fine-tuning. Non-blocking.",
             "output_field": "training_data_saved"
      },

             "validate_tool": {
        "action": "validate_page_content",
             "config": {
          "domain": "site_record.domain",
             "html_field": "completeness_check.clean_html"
        },
             "next_step": "save_sections",
             "error_step": "save_sections",
             "description": "Check for obvious issues — cross-site contamination, empty output",
             "output_field": "validation_result"
      },

             "save_sections": {
        "action": "save_page_sections",
             "config": {
          "html_field": "validation_result.clean_html",
             "site_id_field": "site_record.site_id",
             "page_name_field": "page_record.name"
        },
             "next_step": "update_status",
             "error_step": "complete_error",
             "description": "Persist generated tool HTML to page_components",
             "output_field": "sections_saved"
      },

             "update_status": {
        "action": "update_page_status",
             "config": {
          "status": "deployed",
             "site_id_field": "site_record.site_id",
             "page_name_field": "page_record.name"
        },
             "next_step": "spawn_rerender",
             "description": "Mark page as deployed",
             "output_field": "status_updated"
      },

             "spawn_rerender": {
        "action": "spawn_agent",
             "config": {
          "role": "page_renderer",
             "agent_type": "page-rerender"
        },
             "next_step": "deploy_page",
             "description": "Spawn page-rerender for assembly and deploy",
             "output_field": "rerender_agent"
      },

             "deploy_page": {
        "action": "call_agent",
             "config": {
          "target_role": "page_renderer",
             "input_mapping": {
            "site_id": "site_record.site_id",
             "page_id": "page_record.id",
             "domain": "site_record.domain"
          },
             "timeout_seconds": 120
        },
             "next_step": "complete",
             "error_step": "complete_error",
             "description": "Assemble page from stored components and deploy to git",
             "output_field": "deploy_result"
      },

             "complete": {
        "action": "complete_workflow",
             "config": {
          "output_fields": ["tool_analysis", "sections_saved", "deploy_result", "training_data_saved"]
        },
             "description": "Tool recreation complete"
      },

             "complete_error": {
        "action": "complete_workflow",
             "config": {
          "output_fields": ["tool_analysis", "tool_recreation", "site_record"],
             "success_message": "Tool recreation completed with errors"
        },
             "description": "Tool recreation completed with errors"
      }
    }
  }
}$cfg$::jsonb,
             true,
             '["interactive", "tools", "games", "code-generation"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.941',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'specialist',
             'active',
             '["build", "interactive", "handler"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"optional": ["page_name", "page_id", "sections"], "required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"tool_analysis": "Functional specification of the tool", "deploy_result": "git commit result", "sections_saved": "save result with page_id"}}'::jsonb,
             180
         );