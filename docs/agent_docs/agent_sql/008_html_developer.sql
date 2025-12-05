UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "develop",
                "steps": {
                    "develop": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-sonnet-4-5-20250514",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "prompt_template": "You are an expert HTML/CSS developer. Create a complete, production-ready webpage based on:\n\nSite Architecture: {{.site_architecture}}\nContent: {{.site_content}}\n\nCreate a single HTML file with:\n- Embedded CSS in a <style> tag\n- Responsive design (mobile-first)\n- Modern, clean aesthetic\n- Semantic HTML5 elements\n- The color scheme and typography from the architecture\n- All content properly placed\n\nReturn ONLY the complete HTML document, no explanation. Start with <!DOCTYPE html>."
                        },
                        "output_field": "html_result",
                        "next_step": "complete",
                        "description": "Develop HTML/CSS"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return developed HTML"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'html-developer';

remove the top-level fields that are overriding the step config:
default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'
WHERE type IN ('site-architect', 'content-creator', 'html-developer');

--

claude-haiku-4-5-20251001

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "develop",
                "steps": {
                    "develop": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-sonnet-4-5-20250514",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "prompt_template": "You are an expert HTML/CSS developer. Create a complete, production-ready webpage based on:\n\nSite Architecture: {{.site_architecture}}\nContent: {{.site_content}}\n\nCreate a single HTML file with:\n- Embedded CSS in a <style> tag\n- Responsive design (mobile-first)\n- Modern, clean aesthetic\n- Semantic HTML5 elements\n- The color scheme and typography from the architecture\n- All content properly placed\n\nReturn ONLY the complete HTML document, no explanation. Start with <!DOCTYPE html>."
                        },
                        "output_field": "html_result",
                        "next_step": "complete",
                        "description": "Develop HTML/CSS"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return developed HTML"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'html-developer';