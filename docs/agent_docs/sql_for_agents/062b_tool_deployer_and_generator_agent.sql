-- ============================================================================
-- tool-deployer: forks a library tool to a site
-- Simplified workflow — no query_database steps, all logic in Go action
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, status,
    image_repository, image_tag,
    resources, default_config, input_contract, output_contract,
    domain_tags, agent_category, idle_timeout_seconds
) VALUES (
             'tool-deployer',
             'Tool Deployer',
             'Forks a library tool to a site. Creates the component fork, tool page, and page_component link. The page then flows through the normal render/deploy pipeline.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.862',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 180,
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "deploy_tool",
                             "output_field": "site_record"
                         },
                         "deploy_tool": {
                             "action": "deploy_tool_to_site",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "tool_component_id": "input_data.spec.tool_component_id",
                                 "nav_section": "Tools",
                                 "in_header": true,
                                 "in_footer": false
                             },
                             "next_step": "complete",
                             "error_step": "complete_error",
                             "description": "Fork library tool and create tool page",
                             "output_field": "deploy_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["deploy_result"]
                             }
                         },
                         "complete_error": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["deploy_result"],
                                 "success_message": "Tool deployment failed"
                             }
                         }
                     }
                 }
             }'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["spec"]}'::jsonb,
             '{"produces": {"deploy_result": "fork_id, page_id, page_url"}}'::jsonb,
             '["tools", "deployment", "fork"]'::jsonb,
             'specialist',
             120
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    display_name = EXCLUDED.display_name,
    image_tag = EXCLUDED.image_tag,
    idle_timeout_seconds = EXCLUDED.idle_timeout_seconds,
    status = EXCLUDED.status,
    updated_at = NOW();

-- ============================================================================
-- tool-generator: creates a tool from scratch via LLM
-- For when there's no library tool to fork
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, status,
    image_repository, image_tag,
    resources, default_config, input_contract, output_contract,
    domain_tags, agent_category, idle_timeout_seconds
) VALUES (
             'tool-generator',
             'Tool Generator',
             'Creates a new interactive tool from scratch via LLM. Generates self-contained HTML/JS/CSS based on the tool description, saves as a content_component, and creates a tool page. Used when no suitable library tool exists to fork.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.862',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 300,
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "load_brand_context",
                             "output_field": "site_record"
                         },
                         "load_brand_context": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "site_record.site_id"
                             },
                             "next_step": "generate_tool_html",
                             "error_step": "generate_tool_html",
                             "description": "Load brand context so the tool matches site style",
                             "output_field": "site_specs"
                         },
                         "generate_tool_html": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "site_record", "site_specs"],
                                 "output_format": "text",
                                 "prompt_template": "You are creating a self-contained interactive tool for a website.\n\n## Tool Requirements\nName: {{.input_data.spec.name}}\nFunction: {{.input_data.spec.function}}\nDescription: {{.input_data.spec.description}}\n{{if .input_data.spec.complexity}}Complexity: {{.input_data.spec.complexity}}{{end}}\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs}}{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}}\nCompany: {{.site_specs.specs.identity.company_name}}\nAudience: {{.site_specs.specs.identity.target_audience}}{{end}}{{end}}\n\n## Rules\n1. Output a single, self-contained HTML fragment\n2. Include all CSS in a single <style> block at the top\n3. Include all JavaScript in a single <script> block at the bottom\n4. Use CSS custom properties for colours: var(--color-primary), var(--color-secondary), var(--color-accent), var(--color-background), var(--color-surface), var(--color-text), var(--color-text-muted), var(--color-border)\n5. Never hardcode hex colour values — always use CSS variables\n6. Must work on mobile (responsive layout, touch-friendly inputs ≥44px)\n7. No external dependencies (no CDN links, no fetch calls)\n8. All calculations and logic must work client-side only\n9. Use semantic HTML (labels for inputs, fieldset for groups, proper headings)\n10. Keep it simple and functional — this is v1, it will be improved later\n11. Wrap the tool in a <div class=\"tool-container\"> with a sensible max-width\n12. Include a clear heading and brief instruction text\n\n## Structure\n```\n<style>\n  /* Layout and spacing only — colours from CSS variables */\n</style>\n\n<div class=\"tool-container\">\n  <h2>Tool Name</h2>\n  <p class=\"tool-description\">Brief instruction</p>\n  <!-- Interactive content -->\n</div>\n\n<script>\n  // Self-contained logic\n</script>\n```\n\nOutput ONLY the HTML fragment. No markdown fences. No explanation. Start directly with <style>."
                             },
                             "next_step": "save_tool",
                             "error_step": "complete_error",
                             "description": "LLM generates the tool HTML/JS/CSS",
                             "output_field": "generated_html"
                         },
                         "save_tool": {
                             "action": "create_tool_component",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "html_content": "generated_html.result",
                                 "function": "input_data.spec.function",
                                 "display_name": "input_data.spec.name",
                                 "description": "input_data.spec.description",
                                 "category": "interactive",
                                 "nav_section": "Tools",
                                 "in_header": true,
                                 "in_footer": false
                             },
                             "next_step": "complete",
                             "error_step": "complete_error",
                             "description": "Save generated HTML as component + create tool page",
                             "output_field": "create_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["create_result"]
                             }
                         },
                         "complete_error": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["generated_html"],
                                 "success_message": "Tool generation failed"
                             }
                         }
                     }
                 }
             }'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["spec"]}'::jsonb,
             '{"produces": {"create_result": "component_id, page_id, page_url"}}'::jsonb,
             '["tools", "generation", "llm"]'::jsonb,
             'specialist',
             120
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    display_name = EXCLUDED.display_name,
    image_tag = EXCLUDED.image_tag,
    idle_timeout_seconds = EXCLUDED.idle_timeout_seconds,
    status = EXCLUDED.status,
    updated_at = NOW();

-- ============================================================================
-- Unblock existing add_tool items
-- ============================================================================
UPDATE site_work_items
SET status = 'triaged'
WHERE item_type = 'add_tool'
  AND status = 'blocked';

-- ============================================================================
-- Verify
-- ============================================================================
SELECT type, display_name, status
FROM agent_definitions
WHERE type IN ('tool-deployer', 'tool-generator') AND deleted_at IS NULL;