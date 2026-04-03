-- med_url_map_agent.sql
-- Agent definition for Firecrawl /map based URL discovery.
-- Broader than med-url-discoverer (which targets category pages).

-- Registry entry (add to registry.go):
-- "med_map_urls": {Handler: MedMapURLsAction, Category: "med_pricing", Description: "Discover product URLs via Firecrawl /map endpoint (site-wide)", IsLocal: true},

-- Agent definition
INSERT INTO agent_definitions (
    type, name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-url-mapper',
             'Med URL Mapper',
             'Discovers product URLs across retailer sites using Firecrawl /map endpoint. Broader than category-page discovery.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.935',
             true, 'experimental',
             '{
                 "workflow": {
                     "steps": {
                         "map_urls": {
                             "action": "med_map_urls",
                             "config": {},
                             "next_step": "complete",
                             "timeout_seconds": 300
                         }
                     },
                     "initial_step": "map_urls"
                 }
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[]'::jsonb,
             0
         ) ON CONFLICT (type) WHERE deleted_at IS NULL
    DO UPDATE SET
    description = EXCLUDED.description,
           default_config = EXCLUDED.default_config,
           updated_at = NOW();

-- Optional scheduled task (disabled by default)
INSERT INTO scheduled_tasks (
    name, agent_type, schedule_cron, is_enabled,
    target_topic, message_payload,
    description
) VALUES (
             'med-map-urls',
             'med-url-mapper',
             '0 3 * * 1',  -- weekly Monday 3am
             false,
             'system.agent.business-intel.requests',
             '{"action":"orchestrate","config":{"agent_type":"med-url-mapper"},"input_data":{}}'::jsonb,
             'Weekly site-wide URL discovery via Firecrawl /map'
         ) ON CONFLICT (name) DO NOTHING;
