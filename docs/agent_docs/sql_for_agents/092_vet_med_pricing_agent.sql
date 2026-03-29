-- med_pricing_agent_and_task.sql
-- Agent definition + scheduled task for med price collection.
-- Runs on the business-intel pod (same pattern as ch-* agents).

-- ============================================================================
-- Agent definition — med-price-collector
-- ============================================================================
INSERT INTO agent_definitions (
    type, version, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'med-price-collector', 1,
             'Med Price Collector',
             'Scrapes veterinary medicine prices from online pharmacies',
             'data-driven',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "processing_mode": "task",
                 "workflow": {
                     "start_step": "scrape_prices",
                     "steps": {
                         "scrape_prices": {
                             "action": "med_scrape_prices",
                             "description": "Scrape prices from retailer product pages",
                             "config": {
                                 "batch_size": 20,
                                 "delay_ms": 2000,
                                 "task_name": "med-scrape-prices"
                             },
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete price collection run"
                         }
                     }
                 }
             }'::jsonb,
             '["med_pricing", "scraping", "firecrawl"]'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       capabilities = EXCLUDED.capabilities,
                                       updated_at = NOW();

-- ============================================================================
-- Scheduled task — runs every 2 days, disabled initially for manual testing
-- ============================================================================
INSERT INTO scheduled_tasks (
    name, target_agent_type, target_topic, input_data,
    interval_seconds, concurrency_group, max_concurrent, timeout_seconds,
    enabled, fire_message
) VALUES (
             'med-scrape-prices',
             'med-price-collector',
             'system.agent.business-intel.requests',
             '{"agent_type": "med-price-collector"}'::jsonb,
             172800,              -- 2 days
             'med-collection',    -- own concurrency group, separate from ch-enrichment
             1,
             600,                 -- 10 min timeout
             false,               -- disabled until we verify extraction works
             true
         )
    ON CONFLICT (name) DO UPDATE SET
    target_agent_type = EXCLUDED.target_agent_type,
                              target_topic = EXCLUDED.target_topic,
                              input_data = EXCLUDED.input_data,
                              interval_seconds = EXCLUDED.interval_seconds,
                              concurrency_group = EXCLUDED.concurrency_group,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              updated_at = NOW();

