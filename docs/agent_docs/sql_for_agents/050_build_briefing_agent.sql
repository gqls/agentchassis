-- build-briefing-agent agent definition
-- Handler for: needs_briefing work items
-- Pipeline position: after domain-research-classifier, before site-planner
--
-- Distinct from existing briefing-agent (v1) which is used by intake-orchestrator
-- and receives questionnaire directly as input. This version reads from site_specs.
--
-- Receives from dispatch loop:
--   input_data.site_id       — UUID of the site
--   input_data.domain        — domain name
--   input_data.work_item_id  — the work item being processed
--
-- Reads from site_specs:
--   identity       — company info from domain research
--   classification — site type, recommended builder
--
-- Outputs to site_specs:
--   aspect "briefing" — answered questionnaire fields
--
-- Creates next work item:
--   needs_site_plan → site-planner

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-briefing-agent',
             'Build Briefing Agent',
             'Handler-mode briefing agent for the build dispatch pipeline. Reads domain research from site_specs, fetches the target builder questionnaire, uses LLM to answer all briefing questions, writes answers to site_specs, chains to site-planner.',
             'specialist',
             '{
                 "workflow": {
                     "start_step": "read_specs",
                     "steps": {

                         "read_specs": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "aspect": "all"
                             },
                             "next_step": "fetch_questionnaire",
                             "description": "Load identity and classification from site_specs",
                             "output_field": "site_specs"
                         },

                         "fetch_questionnaire": {
                             "action": "fetch_agent_questionnaire",
                             "config": {
                                 "agent_type_field": "site_specs.classification.recommended_builder"
                             },
                             "next_step": "answer_briefing",
                             "description": "Fetch briefing questionnaire for the target builder agent",
                             "output_field": "questionnaire"
                         },

                         "answer_briefing": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "site_specs", "questionnaire"],
                                 "output_format": "json",
                                 "prompt_template": "You are completing a website briefing questionnaire. Use ONLY the research data provided — do not invent information you do not have.\n\nDomain: {{.input_data.domain}}\n\n## Research Data (from domain analysis)\n{{.site_specs}}\n\n## Questionnaire to Answer\n{{.questionnaire}}\n\n## Instructions\nAnswer every question in the questionnaire using the research data above.\n\nFor each question:\n- Use the exact field name as the JSON key\n- If the research data contains the answer, use it\n- If the research data is partial, provide what we have and note gaps\n- For json_array fields, return proper JSON arrays\n- For boolean fields, return true/false\n- For select fields, choose the most appropriate option\n- If we have NO data for a field and it is not required, use null\n- If we have NO data for a required field, provide a reasonable placeholder based on the domain name and industry and mark it with \"[INFERRED]\" prefix\n\nAlso include these synthesis fields:\n- \"company_name\": from identity research\n- \"industry\": from identity research  \n- \"tagline\": from identity research (null if not found)\n- \"tone\": inferred from industry and site type\n- \"site_type\": from classification\n- \"recommended_builder\": from classification\n- \"briefing_confidence\": 0.0-1.0 (how much of the questionnaire we could answer from research)\n- \"gaps\": [\"list of fields we could not answer from research\"]\n\nReturn ONLY valid JSON with all questionnaire field names as keys."
                             },
                             "next_step": "write_briefing_spec",
                             "description": "LLM answers all briefing questions using research data",
                             "output_field": "briefing_answers"
                         },

                         "write_briefing_spec": {
                             "action": "write_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "spec_data": "briefing_answers.result",
                                 "aspect": "briefing",
                                 "source": "build-briefing-agent",
                                 "source_agent": "build-briefing-agent",
                                 "source_item_id": "input_data.work_item_id"
                             },
                             "next_step": "create_next_item",
                             "description": "Persist briefing answers to site_specs",
                             "output_field": "briefing_written"
                         },

                         "create_next_item": {
                             "action": "create_work_item",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_type": "needs_site_plan",
                                 "handler_agent": "site-planner",
                                 "item_domain": "build",
                                 "severity": "high",
                                 "source": "build-briefing-agent",
                                 "summary": "Site plan needed after briefing complete",
                                 "item_key_prefix": "site_plan",
                                 "priority": 15
                             },
                             "next_step": "complete",
                             "description": "Create work item for site planning stage",
                             "output_field": "next_item_created"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["briefing_answers", "briefing_written", "next_item_created"]
                             },
                             "description": "Briefing complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 180
             }'::jsonb,
             true,
             '["briefing", "questionnaire", "llm"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'specialist',
             'experimental',
             '["briefing", "questionnaire", "research-synthesis"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id"], "optional": ["domain", "work_item_id"], "description": "Receives site_id from dispatch loop. Reads identity and classification from site_specs."}'::jsonb,
             '{"produces": {"briefing_written": "site_spec write result for briefing aspect", "next_item_created": "work item created for site-planner"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();

