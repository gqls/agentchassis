-- 063: color-variable-fixer agent
-- Replaces hardcoded hex colors in component inline styles with CSS variables.
-- Fixes both content_components.html_template (permanent) and
-- page_components.rendered_html (immediate). No content regeneration needed.
--
-- Dispatched by improvement loop when hardcoded_section_colors finding detected.
-- Returns needs_rerender: true so pages get reassembled.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, image_repository, image_tag, resources, topics, health_config,
    env_vars, version, delegation_preferences, status, domain_tags,
    input_contract, output_contract
) VALUES (
             'color-variable-fixer',
             'Color Variable Fixer',
             'Replaces hardcoded hex colors in component inline <style> blocks with CSS variable references. Fixes templates (permanent) and rendered HTML (immediate). No content changes.',
             'specialist',
             '{
                 "workflow": {
                     "start_step": "fix_colors",
                     "steps": {
                         "fix_colors": {
                             "action": "fix_hardcoded_colors",
                             "config": {
                                 "fix_templates": true,
                                 "fix_rendered": true
                             },
                             "next_step": "complete",
                             "description": "Replace hardcoded hex backgrounds with var(--color-primary) etc",
                             "output_field": "fix_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["fix_result"]
                             },
                             "description": "Return fix results"
                         }
                     }
                 },
                 "processing_mode": "task",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["maintenance", "css", "quality-fix"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.811',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'experimental',
             '["maintenance", "css", "quality"]'::jsonb,
             '{"required": ["site_id"], "optional": []}'::jsonb,
             '{"produces": {"templates_fixed": "int", "rendered_fixed": "int", "needs_rerender": "bool"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       display_name = EXCLUDED.display_name;


