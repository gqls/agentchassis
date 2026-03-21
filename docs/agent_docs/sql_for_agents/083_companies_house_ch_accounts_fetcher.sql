-- ============================================================================
-- CH Accounts Fetch — Migration, Agent Definition & Scheduled Task
-- ============================================================================

-- Add accounts_fetched tracking column to ch_vet_companies
ALTER TABLE business_intel.ch_vet_companies
    ADD COLUMN IF NOT EXISTS accounts_fetched BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE business_intel.ch_vet_companies
    ADD COLUMN IF NOT EXISTS accounts_fetched_at TIMESTAMPTZ;

-- Ensure financial columns exist on companies_house_data
-- (they should from the original schema, but belt and braces)
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS accounts_date DATE;
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS accounts_type TEXT;
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS total_assets_gbp NUMERIC(12,2);
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS net_worth_gbp NUMERIC(12,2);
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS turnover_gbp NUMERIC(12,2);
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS profit_loss_gbp NUMERIC(12,2);
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS employee_count INTEGER;
ALTER TABLE business_intel.companies_house_data
    ADD COLUMN IF NOT EXISTS employee_count_band TEXT;

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
             'ch-accounts-fetcher',
             'CH Accounts Fetcher',
             'Fetches and parses Companies House filed accounts (iXBRL). Extracts net assets, total assets, employee count, turnover, profit/loss where available.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "fetch_accounts",
                     "steps": {
                         "fetch_accounts": {
                             "action": "ch_fetch_accounts",
                             "config": {
                                 "batch_size": 30,
                                 "delay_ms": 15000,
                                 "task_name": "ch-fetch-accounts"
                             },
                             "next_step": "complete",
                             "description": "Fetch filing history, download iXBRL, parse financial values. 30 per batch, 15s delay.",
                             "output_field": "accounts_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["accounts_result"]
                             },
                             "description": "Accounts fetch complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 3600
             }'::jsonb,
             true,
             '["companies-house", "enrichment", "accounts"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.902',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "enrichment", "accounts", "financial"]',
             '{"required": [], "optional": ["batch_size", "delay_ms"]}'::jsonb,
             '{"produces": {"accounts_result": "object - fetched, parsed, no_accounts, failed"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

-- Scheduled task — runs every 20 minutes. 30 companies per batch.
-- 3 API calls per company at 15s delay = ~22 minutes per batch.
-- Pre-query stops when all accounts are fetched.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-fetch-accounts',
             'Fetch and parse CH filed accounts (iXBRL) for enriched companies.',
             1200,      -- 20 minutes
             'ch-accounts-fetcher',
             'system.agent.business-intel.requests',
             'ch-enrichment',
             1,
             false,
             3600,      -- 60 min timeout
             'SELECT COUNT(*) as unfetched FROM business_intel.ch_vet_companies ch JOIN business_intel.companies_house_data chd ON chd.business_id = ch.matched_business_id WHERE ch.details_fetched = true AND ch.accounts_fetched = false HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

-- ============================================================================
-- Verification queries:
-- ============================================================================
--
-- Financial data coverage:
--   SELECT COUNT(*) FILTER (WHERE accounts_date IS NOT NULL) as has_accounts,
--          COUNT(*) FILTER (WHERE net_worth_gbp IS NOT NULL) as has_net_worth,
--          COUNT(*) FILTER (WHERE employee_count IS NOT NULL) as has_employees,
--          COUNT(*) FILTER (WHERE turnover_gbp IS NOT NULL) as has_turnover,
--          COUNT(*) as total
--   FROM business_intel.companies_house_data;
--
-- Sample financial data:
--   SELECT b.name, chd.net_worth_gbp, chd.total_assets_gbp, chd.employee_count,
--          chd.turnover_gbp, chd.accounts_type, chd.accounts_date
--   FROM business_intel.companies_house_data chd
--   JOIN business_intel.businesses b ON b.id = chd.business_id
--   WHERE chd.net_worth_gbp IS NOT NULL
--   ORDER BY chd.net_worth_gbp DESC
--   LIMIT 20;
-- ============================================================================

-- Scheduled task and agent definition for ch-accounts-fetcher
-- Run against clients_db as clients_user.
--
-- The agent definition may already exist from a previous session.
-- ON CONFLICT handles idempotent upsert.

-- ============================================================
-- 1. Agent definition (ch-accounts-fetcher)
-- ============================================================
INSERT INTO agent_definitions (
    type, version, name, description, category,
    default_config, enabled, tags, container_image, image_tag,
    resource_limits, topics, health_check,
    dependencies, priority,
    orchestration_config, role, status, capabilities
) VALUES (
             'ch-accounts-fetcher', 1,
             'CH Accounts Fetcher',
             'Fetches and parses Companies House filed accounts (iXBRL). Extracts net assets, total assets, employee count, turnover, profit/loss where available.',
             'data-driven',
             jsonb_build_object(
                     'workflow', jsonb_build_object(
                     'start_step', 'fetch_accounts',
                     'steps', jsonb_build_object(
                             'fetch_accounts', jsonb_build_object(
                                     'action', 'ch_fetch_accounts',
                                     'description', 'Fetch filing history, download iXBRL, parse financial values. 30 per batch, 15s delay.',
                                     'config', jsonb_build_object(
                                             'batch_size', 30,
                                             'delay_ms', 15000,
                                             'task_name', 'ch-fetch-accounts'
                                               ),
                                     'output_field', 'accounts_result',
                                     'next_step', 'complete'
                                               ),
                             'complete', jsonb_build_object(
                                     'action', 'complete_workflow',
                                     'config', jsonb_build_object(
                                             'output_fields', jsonb_build_array('accounts_result')
                                               ),
                                     'description', 'Accounts fetch complete'
                                         )
                              )
                                 ),
                     'processing_mode', 'orchestrator',
                     'timeout_seconds', 3600
             ),
             true,
             ARRAY['companies-house', 'enrichment', 'accounts'],
             'docker.io/aqls/agent-chassis',
             'v1.0.902',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             ARRAY[]::text[],
             1,  -- priority (same as other CH agents, below business-intel at 3)
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'specialist',
             'experimental',
             ARRAY['companies-house', 'enrichment', 'accounts', 'financial']
         )
    ON CONFLICT (type, version)
DO UPDATE SET
    name = EXCLUDED.name,
           description = EXCLUDED.description,
           default_config = EXCLUDED.default_config,
           enabled = EXCLUDED.enabled,
           tags = EXCLUDED.tags,
           image_tag = EXCLUDED.image_tag,
           updated_at = NOW();

-- ============================================================
-- 2. Scheduled task (ch-fetch-accounts)
-- ============================================================
-- Runs every 20 minutes (same cadence as ch-detail-fetch).
-- pre_query prevents triggering when there's nothing to fetch.
-- Uses ch-enrichment concurrency group (same as ch-detail-fetch).
INSERT INTO scheduled_tasks (
    name, agent_type, schedule, enabled,
    concurrency_group, pre_query,
    config
) VALUES (
             'ch-fetch-accounts',
             'ch-accounts-fetcher',
             '*/20 * * * *',
             false,  -- start disabled, enable after migration verified
             'ch-enrichment',
             'SELECT EXISTS(
                 SELECT 1 FROM business_intel.ch_vet_companies ch
                 JOIN business_intel.companies_house_data chd ON chd.business_id = ch.matched_business_id
                 WHERE ch.details_fetched = true
                   AND ch.accounts_fetched = false
                   AND ch.match_method NOT IN (''pending_llm_review'', ''llm_uncertain'')
                 LIMIT 1
             ) AS has_work',
             jsonb_build_object(
                     'agent_type', 'ch-accounts-fetcher'
             )
         )
    ON CONFLICT (name)
DO UPDATE SET
    agent_type = EXCLUDED.agent_type,
           schedule = EXCLUDED.schedule,
           concurrency_group = EXCLUDED.concurrency_group,
           pre_query = EXCLUDED.pre_query,
           config = EXCLUDED.config,
           updated_at = NOW();

---

-- Scheduled task for ch-fetch-accounts
-- Run against clients_db as clients_user.
--
-- The ch-accounts-fetcher agent definition already exists (ea99a5f5-...).
-- This just creates the scheduled task to trigger it.
--
-- The scheduler fires to the business-intel pod's requests topic.
-- input_data.config.agent_type tells selectWorkflow which agent definition to load.

INSERT INTO scheduled_tasks (
    name,
    target_agent_type,
    target_topic,
    input_data,
    pre_query,
    interval_seconds,
    concurrency_group,
    max_concurrent,
    timeout_seconds,
    enabled,
    fire_message
) VALUES (
             'ch-fetch-accounts',
             'ch-accounts-fetcher',
             'system.agent.business-intel.requests',
             '{"config": {"agent_type": "ch-accounts-fetcher"}}'::jsonb,
             'SELECT 1 AS has_work
              FROM business_intel.ch_vet_companies ch
              JOIN business_intel.companies_house_data chd ON chd.business_id = ch.matched_business_id
              WHERE ch.details_fetched = true
                AND ch.accounts_fetched = false
                AND ch.match_method NOT IN (''pending_llm_review'', ''llm_uncertain'')
              LIMIT 1',
             1200,           -- every 20 min, same as ch-detail-fetch
             'ch-enrichment', -- same concurrency group as ch-detail-fetch
             1,
             3600,           -- 1 hour timeout
             false,          -- start disabled, enable after migration verified
             true
         )
    ON CONFLICT (name)
DO UPDATE SET
    target_agent_type = EXCLUDED.target_agent_type,
           target_topic = EXCLUDED.target_topic,
           input_data = EXCLUDED.input_data,
           pre_query = EXCLUDED.pre_query,
           interval_seconds = EXCLUDED.interval_seconds,
           concurrency_group = EXCLUDED.concurrency_group,
           max_concurrent = EXCLUDED.max_concurrent,
           timeout_seconds = EXCLUDED.timeout_seconds,
           fire_message = EXCLUDED.fire_message,
           updated_at = NOW();

