-- ============================================================================
-- CH Bulk Collection — Workflow & Scheduled Task
-- ============================================================================
-- This sets up a simple workflow for bulk-collecting all SIC 75000 companies
-- from Companies House into the local ch_vet_companies table.
--
-- The collection runs on the business-intel pod via a separate agent definition
-- (ch-collector) spawned as a K8s job. This keeps collection separate from
-- the ongoing enrichment workflow.
-- ============================================================================

-- Agent definition for collection
-- This is a minimal workflow: collect → complete.
-- The action handles all pagination internally.
INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags
) VALUES (
             gen_random_uuid(),
             'ch-collector',
             'CH Vet Company Collector',
             'Bulk collects all Companies House companies with SIC 75000 (veterinary activities) into local mirror table. Run periodically to keep the mirror fresh.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "collect",
                     "steps": {
                         "collect": {
                             "action": "ch_bulk_collect",
                             "config": {
                                 "sic_code": "75000",
                                 "company_status": "active",
                                 "page_size": 100,
                                 "delay_ms": 2000
                             },
                             "next_step": "report",
                             "description": "Paginate through CH advanced search and store all SIC 75000 companies",
                             "output_field": "collection_result"
                         },
                         "report": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT COUNT(*) as total_companies, COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched, COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched FROM business_intel.ch_vet_companies WHERE company_status = ''active''",
                                 "output_format": "object"
                             },
                             "next_step": "complete",
                             "description": "Report collection stats",
                             "output_field": "stats"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["collection_result", "stats"]
                             },
                             "description": "Collection complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 1800
             }'::jsonb,
             true,
             '["companies-house", "collection", "bulk-data"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.889',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "collection", "veterinary"]'
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              image_tag = EXCLUDED.image_tag,
                              updated_at = NOW();

-- Scheduled task for periodic refresh (monthly)
-- Disabled by default — enable after first manual run validates the data.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled,
    pre_query
) VALUES (
             'ch-vet-collect',
             'Monthly refresh of local CH vet companies mirror (SIC 75000)',
             2592000,  -- 30 days
             'ch-collector',
             'system.agent.business-intel.requests',
             'ch-collection',
             1,
             false,    -- disabled until first manual run validates
             'SELECT ''ready'' as status'  -- always ready when enabled
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              updated_at = NOW();

-- ============================================================================
-- To run the first collection manually:
-- ============================================================================
-- Option 1: Enable the scheduled task and wait for it to fire
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'ch-vet-collect';
--
-- Option 2: The collection runs on the business-intel pod since it has the
-- COMPANIES_HOUSE_API_KEY and database access. The ch-collector agent type
-- will be spawned as a K8s job by the remote-job-spawner.
--
-- After collection, verify:
--   SELECT COUNT(*) FROM business_intel.ch_vet_companies;
--   SELECT company_status, COUNT(*) FROM business_intel.ch_vet_companies GROUP BY company_status;
--   SELECT postcode_prefix, COUNT(*) FROM business_intel.ch_vet_companies GROUP BY postcode_prefix ORDER BY COUNT(*) DESC LIMIT 20;
-- ============================================================================


---

--

-- ============================================================================
-- CH Bulk Collection — Workflow & Scheduled Task
-- ============================================================================
-- This sets up a simple workflow for bulk-collecting all SIC 75000 companies
-- from Companies House into the local ch_vet_companies table.
--
-- The collection runs on the business-intel pod via a separate agent definition
-- (ch-collector) spawned as a K8s job. This keeps collection separate from
-- the ongoing enrichment workflow.
-- ============================================================================

-- Agent definition for collection
-- Workflow: collect → report → notify_scheduler → complete
-- The action handles all pagination internally.
INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-collector',
             'CH Vet Company Collector',
             'Bulk collects all Companies House companies with SIC 75000 (veterinary activities) into local mirror table. Run periodically to keep the mirror fresh.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "collect",
                     "steps": {
                         "collect": {
                             "action": "ch_bulk_collect",
                             "config": {
                                 "sic_code": "75000",
                                 "company_status": "active",
                                 "page_size": 100,
                                 "delay_ms": 2000
                             },
                             "next_step": "report",
                             "description": "Paginate through CH advanced search and store all SIC 75000 companies",
                             "output_field": "collection_result"
                         },
                         "report": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT COUNT(*) as total_companies, COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched, COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched FROM business_intel.ch_vet_companies WHERE company_status = ''active''",
                                 "output_format": "object"
                             },
                             "next_step": "notify_scheduler",
                             "description": "Report collection stats",
                             "output_field": "stats"
                         },
                         "notify_scheduler": {
                             "action": "query_database",
                             "config": {
                                 "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''ch-vet-collect''",
                                 "output_format": "object"
                             },
                             "next_step": "complete",
                             "description": "Tell scheduler this execution finished"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["collection_result", "stats"]
                             },
                             "description": "Collection complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 1800
             }'::jsonb,
             true,
             '["companies-house", "collection", "bulk-data"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.890',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "collection", "veterinary"]',
             '{"required": [], "optional": ["start_from"]}'::jsonb,
             '{"produces": {"collection_result": "object - total_collected, total_new, total_hits, pages_processed", "stats": "object - total_companies, matched, unmatched"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              image_tag = EXCLUDED.image_tag,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();

-- Scheduled task for periodic refresh (monthly)
-- Disabled by default — enable after first manual run validates the data.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-vet-collect',
             'Monthly refresh of local CH vet companies mirror (SIC 75000)',
             2592000,  -- 30 days
             'ch-collector',
             'system.agent.business-intel.requests',
             'ch-collection',
             1,
             false,    -- disabled until first manual run validates
             1800,     -- 30 min timeout — must be < interval_seconds
             'SELECT ''ready'' as status'  -- always ready when enabled
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              updated_at = NOW();

-- ============================================================================
-- To run the first collection manually:
-- ============================================================================
-- Option 1: Enable the scheduled task and wait for it to fire
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'ch-vet-collect';
--
-- Option 2: The collection runs on the business-intel pod since it has the
-- COMPANIES_HOUSE_API_KEY and database access. The ch-collector agent type
-- will be spawned as a K8s job by the remote-job-spawner.
--
-- After collection, verify:
--   SELECT COUNT(*) FROM business_intel.ch_vet_companies;
--   SELECT company_status, COUNT(*) FROM business_intel.ch_vet_companies GROUP BY company_status;
--   SELECT postcode_prefix, COUNT(*) FROM business_intel.ch_vet_companies GROUP BY postcode_prefix ORDER BY COUNT(*) DESC LIMIT 20;
-- ============================================================================