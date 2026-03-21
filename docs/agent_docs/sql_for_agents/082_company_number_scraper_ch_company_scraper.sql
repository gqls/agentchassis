-- ============================================================================
-- CH Company Number Scraper — Migration, Agent Definition & Scheduled Task
-- ============================================================================

-- Add column to businesses table for scraped company numbers
-- NULL = not yet scraped, '' = scraped but not found, '12345678' = found
ALTER TABLE business_intel.businesses
    ADD COLUMN IF NOT EXISTS company_number_scraped VARCHAR(10);

-- Index for quick lookup of unscraped businesses
CREATE INDEX IF NOT EXISTS idx_bi_businesses_company_number_scraped
    ON business_intel.businesses (company_number_scraped)
    WHERE company_number_scraped IS NULL AND verification_status = 'verified';

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
             'ch-company-scraper',
             'CH Company Number Scraper',
             'Scrapes business websites for Companies House registration numbers. Extracts company numbers from footers/about pages using regex patterns. Generic across all verticals.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "scrape",
                     "steps": {
                         "scrape": {
                             "action": "ch_scrape_company_number",
                             "config": {
                                 "batch_size": 100,
                                 "delay_ms": 1000,
                                 "request_timeout_sec": 10,
                                 "min_confidence": 0.40,
                                 "vertical_slug": "veterinary",
                                 "task_name": "ch-scrape-company-number"
                             },
                             "next_step": "complete",
                             "description": "Scrape website footers for company registration numbers",
                             "output_field": "scrape_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["scrape_result"]
                             },
                             "description": "Scraping complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["companies-house", "web-scraping", "matching"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.900',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "web-scraping"]',
             '{"required": [], "optional": ["batch_size", "delay_ms", "vertical_slug"]}'::jsonb,
             '{"produces": {"scrape_result": "object - scraped, found, matched, failed"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

-- Scheduled task — runs every 10 minutes processing 100 websites per batch.
-- At 1 req/sec, each batch takes ~100 seconds. Pre-query stops when all are scraped.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-scrape-company-number',
             'Scrape business websites for company registration numbers. 100 per batch, 1 req/sec.',
             600,       -- 10 minutes
             'ch-company-scraper',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             600,       -- 10 min timeout
             'SELECT COUNT(*) as unscraped FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id WHERE bv.slug = ''veterinary'' AND b.verification_status = ''verified'' AND b.confidence_score >= 0.40 AND b.website_url IS NOT NULL AND b.website_url != '''' AND b.company_number_scraped IS NULL AND b.business_type NOT ILIKE ''%directory%'' HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

---
-- forget it, we'll use the regex in the verifier
-- Remove the unused scraper agent and task
DELETE FROM scheduled_tasks WHERE name = 'ch-scrape-company-number';
DELETE FROM agent_definitions WHERE type = 'ch-company-scraper';