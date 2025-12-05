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