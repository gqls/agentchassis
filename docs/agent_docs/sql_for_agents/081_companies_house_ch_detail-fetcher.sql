-- ============================================================================
-- CH Detail Fetch — Agent Definition & Scheduled Task
-- ============================================================================
-- Fetches officers, PSC, and profile data from CH API for confirmed matches.
-- Processes a batch per invocation with rate limiting (~2 companies/minute).
-- Stores enrichment data in companies_house_data, marks ch_vet_companies
-- rows as details_fetched = true.
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
             'ch-detail-fetcher',
             'CH Detail Fetcher',
             'Fetches detailed company data (profile, officers, PSC) from CH API for confirmed matches. Derives succession risk signals and stores in companies_house_data.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "fetch",
                     "steps": {
                         "fetch": {
                             "action": "ch_detail_fetch",
                             "config": {
                                 "batch_size": 50,
                                 "delay_ms": 15000,
                                 "fetch_officers": true,
                                 "fetch_psc": true,
                                 "task_name": "ch-detail-fetch"
                             },
                             "next_step": "complete",
                             "description": "Fetch profile, officers, PSC for confirmed matches. 50 per batch, 15s delay between API calls.",
                             "output_field": "fetch_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["fetch_result"]
                             },
                             "description": "Detail fetch complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 3600
             }'::jsonb,
             true,
             '["companies-house", "enrichment", "detail-fetch"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.897',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "enrichment", "veterinary"]',
             '{"required": [], "optional": ["batch_size", "delay_ms", "fetch_officers", "fetch_psc"]}'::jsonb,
             '{"produces": {"fetch_result": "object - fetched, failed, total_remaining"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

-- Scheduled task — runs every 20 minutes until all confirmed matches are fetched.
-- 50 companies per batch × 3 API calls × 15s delay = ~37 minutes per batch.
-- But batch_size of 50 with 15s delay = ~12.5 minutes, leaving headroom.
-- Concurrency group prevents overlap with matching tasks.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-detail-fetch',
             'Fetch CH profile, officers, PSC for confirmed matches. Rate-limited API calls.',
             1200,      -- 20 minutes
             'ch-detail-fetcher',
             'system.agent.business-intel.requests',
             'ch-enrichment',
             1,
             false,
             3600,      -- 60 min timeout (50 companies × ~45s each)
             'SELECT COUNT(*) as unfetched FROM business_intel.ch_vet_companies WHERE matched_business_id IS NOT NULL AND details_fetched = false AND match_method NOT IN (''pending_llm_review'', ''llm_uncertain'') HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

