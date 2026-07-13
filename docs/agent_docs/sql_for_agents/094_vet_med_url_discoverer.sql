-- med_url_discovery_agent_and_task.sql
-- Agent definition + scheduled task for URL discovery.
-- Runs on the business-intel pod.

-- ============================================================================
-- Agent definition — med-url-discoverer
-- ============================================================================
INSERT INTO agent_definitions (
    type, version, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'med-url-discoverer', 1,
             'Med URL Discoverer',
             'Discovers product URLs by scraping retailer category pages',
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
                     "start_step": "discover_urls",
                     "steps": {
                         "discover_urls": {
                             "action": "med_discover_urls",
                             "description": "Scrape category pages to discover product URLs",
                             "config": {
                                 "delay_ms": 3000,
                                 "task_name": "med-discover-urls"
                             },
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete URL discovery run"
                         }
                     }
                 }
             }'::jsonb,
             '["med_pricing", "url_discovery", "firecrawl"]'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       capabilities = EXCLUDED.capabilities,
                                       updated_at = NOW();

-- ============================================================================
-- Scheduled task — weekly, disabled initially for manual testing
-- ============================================================================
INSERT INTO scheduled_tasks (
    name, target_agent_type, target_topic, input_data,
    interval_seconds, concurrency_group, max_concurrent, timeout_seconds,
    enabled, fire_message
) VALUES (
             'med-discover-urls',
             'med-url-discoverer',
             'system.agent.business-intel.requests',
             '{"agent_type": "med-url-discoverer"}'::jsonb,
             604800,              -- weekly
             'med-collection',
             1,
             600,                 -- 10 min timeout
             false,               -- disabled until tested
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

-- ============================================================================
-- Registry entry (add to registry.go in MED_PRICING section)
-- ============================================================================
-- "med_discover_urls": {
--     Handler:     MedDiscoverURLsAction,
--     Category:    "med_pricing",
--     Description: "Discover product URLs from retailer category pages",
--     IsLocal:     true,
-- },
