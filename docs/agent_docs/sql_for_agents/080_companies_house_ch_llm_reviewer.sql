-- ============================================================================
-- CH LLM Match Reviewer — Agent Definition & Scheduled Task
-- ============================================================================
-- Reviews ambiguous CH matches using LLM judgment.
-- Runs after ch-local-match populates pending_llm_review entries.
-- Uses claude-haiku-4-5 for cost-effective batch review.
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-llm-reviewer',
             'CH LLM Match Reviewer',
             'Reviews ambiguous Companies House matches using LLM judgment. Processes pending_llm_review entries from ch_vet_companies, classifies each as confirmed, rejected, or uncertain.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "review",
                     "steps": {
                         "review": {
                             "action": "ch_llm_review",
                             "config": {
                                 "llm_batch_size": 15,
                                 "task_name": "ch-llm-review"
                             },
                             "next_step": "complete",
                             "description": "Review pending matches using LLM. Batches 15 pairs per call.",
                             "output_field": "review_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["review_result"]
                             },
                             "description": "Review complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["companies-house", "matching", "llm-review"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.895',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "matching", "veterinary", "llm"]',
             '{"required": [], "optional": ["llm_batch_size"]}'::jsonb,
             '{"produces": {"review_result": "object - reviewed, confirmed, rejected, uncertain"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

-- Scheduled task — disabled by default. Run after ch-local-match.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-llm-review',
             'LLM review of ambiguous CH matches. Runs after local matching populates pending_llm_review entries.',
             86400,     -- daily
             'ch-llm-reviewer',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             300,       -- 5 min timeout
             'SELECT COUNT(*) as pending FROM business_intel.ch_vet_companies WHERE match_method = ''pending_llm_review'' HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();


---

-- ============================================================================
-- CH LLM Match Reviewer — Agent Definition & Scheduled Task
-- ============================================================================
-- Reviews ambiguous CH matches using LLM judgment.
-- ai_service is at the top level of default_config per guidelines.
-- Industry-specific context is in the step config, keeping the Go action generic.
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-llm-reviewer',
             'CH LLM Match Reviewer',
             'Reviews ambiguous Companies House matches using LLM judgment. Processes pending_llm_review entries from ch_vet_companies, classifies each as confirmed, rejected, or uncertain.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "review",
                     "steps": {
                         "review": {
                             "action": "ch_llm_review",
                             "config": {
                                 "llm_batch_size": 15,
                                 "task_name": "ch-llm-review",
                                 "industry_name": "veterinary practice",
                                 "industry_context": "- Vets4Pets: All branches registered at SK9 3RN (Pets at Home HQ, Cheadle). Format: \"[LOCATION] VETS4PETS LIMITED\". Only match if the practice is actually a Vets4Pets branch (group_name contains Vets4Pets).\n- IVC Evidensia / Linnaeus: Many registered at YO30 4UZ (York HQ) or B90 4BN. Only match if the practice is part of IVC/Linnaeus/Linnaeus (Mars).\n- CVS Group: Registered at various HQ addresses. Only match if practice group is CVS.\n- Medivet: Registered at central London addresses. Only match if practice group is Medivet.\n- Many independent vet practices are registered at their accountant or solicitor address, which may be in a different town. The distinctive practice name is the key signal, not geography.\n- Common vet name patterns: \"[Name] Veterinary Surgery/Centre/Clinic/Practice/Hospital\" registered as \"[NAME] VETERINARY [TYPE] LIMITED\".",
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-haiku-4-5",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 2000,
                                     "temperature": 0.0
                                 }
                             },
                             "next_step": "complete",
                             "description": "Review pending matches using LLM. Batches 15 pairs per call.",
                             "output_field": "review_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["review_result"]
                             },
                             "description": "Review complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["companies-house", "matching", "llm-review"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.896',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "matching", "veterinary", "llm"]',
             '{"required": [], "optional": ["llm_batch_size", "industry_name", "industry_context"]}'::jsonb,
             '{"produces": {"review_result": "object - reviewed, confirmed, rejected, uncertain"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

-- Scheduled task — disabled by default. Run after ch-local-match.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-llm-review',
             'LLM review of ambiguous CH matches. Runs after local matching populates pending_llm_review entries.',
             86400,     -- daily
             'ch-llm-reviewer',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             300,       -- 5 min timeout
             'SELECT COUNT(*) as pending FROM business_intel.ch_vet_companies WHERE match_method = ''pending_llm_review'' HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

