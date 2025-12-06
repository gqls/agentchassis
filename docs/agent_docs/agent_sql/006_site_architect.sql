-- 2. Fix site-architect
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "design",
                "steps": {
                    "design": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-haiku-4-5-20251001",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "domain_analysis"],
                            "prompt_template": "You are a site architect. Design a website structure based on:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nDomain Analysis: {{.domain_analysis}}\n\nCreate a JSON site architecture with:\n- page_structure: array of pages with name, purpose, and sections\n- navigation: main nav items\n- color_scheme: primary, secondary, accent colors (hex codes)\n- typography: heading and body font recommendations\n- layout_style: grid/flexbox preference\n- components_needed: array of UI components required\n\nReturn ONLY valid JSON, no markdown or explanation."
                        },
                        "output_field": "architecture_result",
                        "next_step": "complete",
                        "description": "Design site structure and architecture"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return site architecture"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'site-architect';

remove the top-level fields that are overriding the step config:
default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'
WHERE type IN ('site-architect', 'content-creator', 'html-developer');


UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "design",
                "steps": {
                    "design": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-haiku-4-5-20251001",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "domain_analysis"],
                            "prompt_template": "You are a site architect. Design a website structure based on:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nDomain Analysis: {{.domain_analysis}}\n\nCreate a JSON site architecture with:\n- page_structure: array of pages with name, purpose, and sections\n- navigation: main nav items\n- color_scheme: primary, secondary, accent colors (hex codes)\n- typography: heading and body font recommendations\n- layout_style: grid/flexbox preference\n- components_needed: array of UI components required\n\nReturn ONLY valid JSON, no markdown or explanation."
                        },
                        "output_field": "architecture_result",
                        "next_step": "complete",
                        "description": "Design site structure and architecture"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return site architecture"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'site-architect';

-- Site Architect
UPDATE agent_definitions
SET
    input_contract = '{
        "required": ["input_data", "domain_analysis"],
        "expects": {
            "domain_analysis": "object"
        }
    }'::jsonb,
    output_contract = '{
        "produces": "site_architecture",
        "format": {
            "type": "object",
            "description": "Site structure, pages, and component layout"
        }
    }'::jsonb
WHERE type = 'site-architect';

