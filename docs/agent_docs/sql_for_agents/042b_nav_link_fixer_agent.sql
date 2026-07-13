-- ============================================================================
-- 2. NAV LINK FIXER AGENT
-- Fixes broken navigation links in site_components (header/footer).
--
-- Strategy:
--   1. Load the site record
--   2. Fix the content_component template if it uses #{{.slug}} pattern
--   3. Force re-render site_components (header + footer)
--   4. If pages are deployed, trigger a rerender of all pages
--      (because header/footer are assembled into every page)
--
-- The template fix is done via update_component_template action.
-- The re-render uses the existing render_site_components action with force=true.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics,
    health_config, env_vars, version,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire,
    usage_count, is_snapshot, input_contract, output_contract
) VALUES (
             'nav-link-fixer',
             'Nav Link Fixer',
             'Fixes broken navigation links in header/footer. Updates content_component templates to use {{.url}} instead of #{{.slug}}, then force re-renders site_components.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {

                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "fix_nav_templates",
                             "description": "Load site record from database",
                             "output_field": "site_record"
                         },

                         "fix_nav_templates": {
                             "action": "fix_nav_link_templates",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "patterns": [
                                     {"find": "href=\"#{{.slug}}\"", "replace": "href=\"{{.url}}\""},
                                     {"find": "href=\"#{{ .slug }}\"", "replace": "href=\"{{ .url }}\""},
                                     {"find": "href=\"#{{.name}}\"", "replace": "href=\"{{.url}}\""}
                                 ]
                             },
                             "next_step": "rerender_site_components",
                             "description": "Fix header/footer templates: replace #slug links with proper URLs",
                             "output_field": "template_fix_result"
                         },

                         "rerender_site_components": {
                             "action": "render_site_components",
                             "config": {
                                 "slots": ["header", "footer"],
                                 "domain_field": "site_record.domain",
                                 "site_id_field": "site_record.site_id",
                                 "force_rerender": true
                             },
                             "next_step": "complete",
                             "description": "Force re-render header and footer with corrected templates",
                             "output_field": "rerender_result"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["template_fix_result", "rerender_result"]
                             },
                             "description": "Nav fix complete. Pages need rerender to pick up new header/footer."
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["nav-fix", "template-repair", "site-component-rerender"]'::jsonb,
             'docker.io/aqls/agent-chassis', 'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb, 1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator', 'experimental',
             '["maintenance", "nav", "template-fix"]'::jsonb, '{}'::jsonb,
             0, false,
             '{"required": ["site_id"], "optional": ["domain", "spec", "work_item_id"]}'::jsonb,
             '{"produces": {"template_fix_result": "templates updated count", "rerender_result": "slots re-rendered"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();

