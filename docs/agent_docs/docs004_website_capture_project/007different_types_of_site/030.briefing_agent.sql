-- ============================================================================
-- BRIEFING AGENT - INSERT (new agent definition)
-- Executes briefing questionnaires via LLM inference or HITL collection
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'briefing-agent',
           'Briefing Agent',
           'Executes briefing questionnaires - either via LLM inference or HITL collection',
           'data-collection',
           '{
             "workflow": {
               "start_step": "check_mode",
               "steps": {
                 "check_mode": {
                   "action": "evaluate_condition",
                   "config": {
                     "condition_field": "input_data.hitl_mode",
                     "conditions": {
                       "interactive": "collect_via_hitl",
                       "auto": "infer_via_llm"
                     },
                     "default": "infer_via_llm"
                   },
                   "next_step": "infer_via_llm",
                   "description": "Determine if briefing should be interactive or auto-inferred"
                 },
                 "infer_via_llm": {
                   "action": "execute_llm_prompt",
                   "config": {
                     "ai_service": {
                       "provider": "anthropic",
                       "model": "claude-haiku-4-5-20251001",
                       "api_key_env_var": "ANTHROPIC_API_KEY",
                       "max_tokens": 4000
                     },
                     "input_fields": ["input_data", "questionnaire"],
                     "output_field": "brief_answers",
                     "prompt_template": "You are completing a briefing questionnaire for a website project.\n\nProject Info:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Site Type: {{.input_data.site_type}}\n- Industry: {{.input_data.detected_industry}}\n\nQuestionnaire to complete:\n{{.questionnaire}}\n\nBased on the domain name, objective, and your knowledge of the industry, provide thoughtful answers to each question in the questionnaire.\n\nFor questions you cannot confidently answer from the available information, provide reasonable defaults appropriate for the industry.\n\nReturn your answers as a JSON object where keys match the field names in the questionnaire:\n{\n  \"field_name_1\": \"your answer\",\n  \"field_name_2\": \"your answer\",\n  ...\n}\n\nReturn ONLY valid JSON."
                   },
                   "next_step": "complete"
                 },
                 "collect_via_hitl": {
                   "action": "request_human_input",
                   "config": {
                     "request_type": "questionnaire",
                     "questionnaire_field": "questionnaire",
                     "timeout_seconds": 86400,
                     "message": "Please complete the briefing questionnaire for this project"
                   },
                   "output_field": "brief_answers",
                   "next_step": "complete",
                   "description": "Collect answers via human-in-the-loop"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return the completed brief"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 120
           }'::jsonb,
           true,
           '["briefing", "questionnaire", "llm", "hitl"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              updated_at = now();