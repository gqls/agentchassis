UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "create_content",
                "steps": {
                    "create_content": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-sonnet-4-5-20250514",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "domain_analysis", "site_architecture"],
                            "prompt_template": "You are a professional copywriter. Create website content based on:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nPersuasion Model: {{.input_data.model}}\nDomain Analysis: {{.domain_analysis}}\nSite Architecture: {{.site_architecture}}\n\nCreate compelling content for each section in the architecture. Use the specified persuasion model.\n\nReturn a JSON object with:\n- hero: headline, subheadline, cta_text\n- sections: array of section objects with title, content, and any relevant data\n- meta: page_title, meta_description\n- footer: company info, links\n\nReturn ONLY valid JSON, no markdown or explanation."
                        },
                        "output_field": "content_result",
                        "next_step": "complete",
                        "description": "Create website content"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return website content"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'content-creator';

remove the top-level fields that are overriding the step config:
default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'
WHERE type IN ('site-architect', 'content-creator', 'html-developer');

haiku

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "create_content",
                "steps": {
                    "create_content": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-haiku-4-5-20251001",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "domain_analysis", "site_architecture"],
                            "prompt_template": "You are a professional copywriter. Create website content based on:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nPersuasion Model: {{.input_data.model}}\nDomain Analysis: {{.domain_analysis}}\nSite Architecture: {{.site_architecture}}\n\nCreate compelling content for each section. Use the specified persuasion model.\n\nReturn a JSON object with:\n- hero: headline, subheadline, cta_text\n- sections: array of section objects with title, content\n- meta: page_title, meta_description\n- footer: company info, links\n\nReturn ONLY valid JSON, no markdown or explanation."
                        },
                        "output_field": "content_result",
                        "next_step": "complete",
                        "description": "Create website content"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return website content"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'content-creator';

-- Content Creator
UPDATE agent_definitions
SET
    input_contract = '{
        "required": ["input_data", "domain_analysis", "site_architecture"],
        "expects": {
            "domain_analysis": "object",
            "site_architecture": "object"
        }
    }'::jsonb,
    output_contract = '{
        "produces": "site_content",
        "format": {
            "type": "object",
            "description": "Content data for all site sections"
        }
    }'::jsonb
WHERE type = 'content-creator';
