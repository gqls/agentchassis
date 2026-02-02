- NOTE: You'll need to also update the preceding step to point to render_site_components
-- and update render_site_components.next_step to point to the right step

-- ============================================================
-- 3. Create a standalone workflow for re-rendering site components
-- Useful when site info changes (email, company name, etc.)
-- ============================================================

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    status
) VALUES (
    'render-site-components',
    'Render Site Components',
    'Renders header, footer, and head components for a site and stores them in site_components table',
    'specialist',
    '{
        "workflow": {
            "start_step": "render_components",
            "steps": {
                "render_components": {
                    "action": "render_site_components",
                    "config": {
                        "input_fields": ["site_id", "domain"],
                        "slots": ["header", "footer", "head"],
                        "force_rerender": true
                    },
                    "description": "Render and store site components",
                    "output_field": "render_result",
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "success_message": "Site components rendered"
                    },
                    "description": "Complete workflow"
                }
            }
        }
    }',
    'active'
) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    updated_at = now();