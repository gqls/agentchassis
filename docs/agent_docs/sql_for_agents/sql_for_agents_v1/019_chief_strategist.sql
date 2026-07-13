


-- Fix chief-strategist to add parse_json step after LLM generation
-- Key insight: Point source_field to the MAP, not to .result
-- Your unwrapDeep will detect "result" key and parse it (Pattern 2)

UPDATE agent_definitions
SET default_config = jsonb_build_object(
        'workflow', jsonb_build_object(
                'steps', jsonb_build_object(
                        'generate_build_plan', jsonb_build_object(
                                'action', 'execute_llm_prompt',
                                'config', jsonb_build_object(
                                        'ai_service', jsonb_build_object(
                                                'model', 'claude-haiku-4-5-20251001',
                                                'provider', 'anthropic',
                                                'api_key_env_var', 'ANTHROPIC_API_KEY'
                                                      ),
                                        'input_data', jsonb_build_array('domain', 'objective', 'model'),
                                        'prompt_template', 'You are a Chief Marketing Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}.

Available Components: [header, hero, features, social_proof, pricing, faq, call_to_action, footer].

Based on the {{.model}} model, select the best sequence of components. Then for each component devise a plan for the copy structure, suggested copy and suggested graphics style that suits the objective {{ .objective }} and the marketing model {{ .model }}.

Output JSON: {"sections": ["component_name", ...], "component_details": {...}}'
                                          ),
                                'next_step', 'parse_plan',
                                'description', 'Create the Build Plan using LLM',
                                'output_field', 'build_plan_raw'
                                               ),
                        'parse_plan', jsonb_build_object(
                                'action', 'parse_json_field',
                                'config', jsonb_build_object(
                                        'source_field', 'build_plan_raw'  -- ← Point to map, not build_plan_raw.result
                                          ),                                     -- Your unwrapDeep handles the rest!
                                'next_step', 'complete',
                                'description', 'Parse JSON from LLM response using existing datahelpers',
                                'output_field', 'plan_data'  -- ← Clear name, avoids "sections.sections" confusion
                                      ),
                        'complete', jsonb_build_object(
                                'action', 'complete_workflow',
                                'description', 'Return parsed plan'
                                    )
                         ),
                'start_step', 'generate_build_plan'
                    ),
        'processing_mode', 'task',
        'timeout_seconds', 120
                     )
WHERE type = 'chief-strategist';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'parse_plan'->'config'->>'source_field' as parse_source,
    default_config->'workflow'->'steps'->'parse_plan'->>'output_field' as parse_output,
    default_config->'workflow'->'start_step' as start_step
FROM agent_definitions
WHERE type = 'chief-strategist';


----

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_build_plan,config}',
        '{
            "ai_service": {
                "provider": "anthropic",
                "model": "claude-haiku-4-5-20251001",
                "api_key_env_var": "ANTHROPIC_API_KEY"
            },
            "max_tokens": 8192,
            "input_data": ["domain", "objective", "model"],
            "prompt_template": "You are a Chief Marketing Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}.\n\nAvailable Components: [header, hero, features, social_proof, pricing, faq, call_to_action, footer].\n\nBased on the {{.model}} model, select the best sequence of components. Then for each component devise a plan for the copy structure, suggested copy and suggested graphics style that suits the objective {{ .objective }} and the marketing model {{ .model }}.\n\nIMPORTANT: You MUST complete the entire JSON structure. If approaching token limits, prioritize completing all JSON fields with brief descriptions rather than leaving structures incomplete.\n\nOutput ONLY valid JSON (no markdown fences): {\"sections\": [\"component_name\", ...], \"component_details\": {...}}"
        }'::jsonb
                     )
WHERE type = 'chief-strategist';

-=- move max tokens to config.ai_service

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_build_plan,config,prompt_template}',
        '"You are a Chief Marketing Strategist designing a landing page for {{.domain}}.\n\nObjective: {{.objective}}\nMarketing Model: {{.model}}\n\nAvailable Components: header, hero, features, social_proof, pricing, faq, call_to_action, footer\n\nTask:\n1. Select 6-8 components based on the {{.model}} model\n2. For EACH component provide: aida_stage, purpose, copy_structure, suggested_copy, graphics_style\n\nCRITICAL: Output complete, valid JSON. If approaching token limits, use concise descriptions but ensure ALL components have ALL required fields. Never leave JSON structures incomplete.\n\nOutput format (valid JSON only):\n{\n  \"sections\": [\"header\", \"hero\", \"features\", ...],\n  \"component_details\": {\n    \"header\": {\"aida_stage\": \"...\", \"purpose\": \"...\", \"copy_structure\": {...}, \"suggested_copy\": {...}, \"graphics_style\": {...}},\n    \"hero\": {...},\n    ...\n  }\n}"'::jsonb
                     )
WHERE type = 'chief-strategist';

--
-- ============================================================================
-- Add output_type to Agent Configs
-- This tells ai_actions.go whether to append JSON output instructions
-- ============================================================================



-- 2. CHIEF STRATEGIST - Outputs JSON build plan
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_build_plan,config,output_type}',
        '"json"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'chief-strategist'
  AND is_active = true;