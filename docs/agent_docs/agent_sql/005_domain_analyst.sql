-- 1. Fix domain-analyst
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "analyze",
                "steps": {
                    "analyze": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-haiku-4-5-20251001",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data"],
                            "prompt_template": "You are a domain analyst. Analyze this website project:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nPersuasion Model: {{.input_data.model}}\n\nProvide a JSON analysis with:\n- target_audience: who this site is for\n- key_value_propositions: array of main benefits to highlight\n- tone: recommended writing tone\n- competitive_advantages: array of what makes this unique\n- recommended_sections: array of suggested page sections\n- industry: detected industry vertical\n- keywords: array of SEO keywords\n\nReturn ONLY valid JSON, no markdown or explanation."
                        },
                        "output_field": "analysis_result",
                        "next_step": "complete",
                        "description": "Analyze domain and business requirements"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return domain analysis"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'domain-analyst';

remove the top-level fields that are overriding the step config:
default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'
WHERE type = 'domain-analyst';


UPDATE agent_definitions
SET
    updated_at = now(),
            default_config = jsonb_set(
                    default_config,
                    '{workflow}',
                    '{
                        "start_step": "analyze",
                        "steps": {
                            "analyze": {
                                "action": "execute_llm_prompt",
                                "config": {
                                    "ai_service": {
                                        "model": "claude-haiku-4-5-20251001",
                                        "provider": "anthropic",
                                        "api_key_env_var": "ANTHROPIC_API_KEY"
                                    },
                                    "input_fields": ["input_data"],
                                    "prompt_template": "You are a domain analyst. Analyze this website project:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nPersuasion Model: {{.input_data.model}}\n\nProvide a JSON analysis with:\n- target_audience: who this site is for\n- key_value_propositions: array of main benefits to highlight\n- tone: recommended writing tone\n- competitive_advantages: array of what makes this unique\n- recommended_sections: array of suggested page sections\n- industry: detected industry vertical\n- keywords: array of SEO keywords\n\nReturn ONLY valid JSON, no markdown or explanation."
                                },
                                "output_field": "analysis_result",
                                "next_step": "complete",
                                "description": "Analyze domain and business requirements"
                            },
                            "complete": {
                                "action": "complete_workflow",
                                "description": "Return domain analysis"
                            }
                        }
                    }'::jsonb
                             )
WHERE type = 'domain-analyst';

-- Domain Analyst
UPDATE agent_definitions
SET
    input_contract = '{
        "required": ["input_data"],
        "expects": {
            "domain": "string",
            "objective": "string"
        }
    }'::jsonb,
    output_contract = '{
        "produces": "domain_analysis",
        "format": {
            "type": "object",
            "description": "Analysis of domain name and business objective"
        }
    }'::jsonb
WHERE type = 'domain-analyst';

