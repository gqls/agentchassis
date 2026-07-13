-- med_url_mapper_setup.sql
-- Sets up the /map based URL discoverer for broad site-wide discovery.
-- Particularly useful for VioVet where category-page scraping misses products.

-- ============================================================================
-- 1. Agent definition for map-based discovery
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
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
             'docker.io/aqls/agent-chassis', 'v1.0.942',
             true, 'experimental',
             '{
                 "workflow": {
                     "start_step": "map_urls",
                     "steps": {
                         "map_urls": {
                             "action": "med_map_urls",
                             "config": {},
                             "next_step": "complete",
                             "timeout_seconds": 300
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Map discovery complete"
                         }
                     }
                 }
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[
                 {"name": "FIRECRAWL_API_URL", "value": "https://api.firecrawl.dev/v2"}
             ]'::jsonb,
             0
         ) ON CONFLICT (type, version)
DO UPDATE SET
    default_config = EXCLUDED.default_config,
                  description = EXCLUDED.description,
                  env_vars = EXCLUDED.env_vars,
                  updated_at = NOW();


-- ============================================================================
-- 2. Map-based discovery orchestrator (spawns temporary pod)
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-url-map-orchestrator',
             'Med URL Map Orchestrator',
             'Spawns a temporary pod to discover product URLs via Firecrawl /map endpoint.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.942',
             true, 'active',
             '{
                 "workflow": {
                     "start_step": "spawn_mapper",
                     "steps": {
                         "spawn_mapper": {
                             "action": "spawn_agent",
                             "config": {
                                 "agent_type": "med-url-mapper",
                                 "role": "url_mapper"
                             },
                             "next_step": "call_mapper",
                             "output_field": "mapper_spawn",
                             "description": "Spawn a temporary pod for URL mapping"
                         },
                         "call_mapper": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "med-url-mapper",
                                 "target_role": "url_mapper",
                                 "input_mapping": {
                                     "input_data": "input_data"
                                 },
                                 "timeout_seconds": 600
                             },
                             "next_step": "complete",
                             "output_field": "map_result",
                             "description": "Send map work to spawned pod and wait"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Map discovery complete"
                         }
                     }
                 }
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[]'::jsonb,
             0
         ) ON CONFLICT (type, version)
DO UPDATE SET
    default_config = EXCLUDED.default_config,
                  description = EXCLUDED.description,
                  updated_at = NOW();


-- ============================================================================
-- Registry entry needed in registry.go:
-- ============================================================================
-- "med_map_urls": {Handler: MedMapURLsAction, Category: "med_pricing", Description: "Discover product URLs via Firecrawl /map endpoint (site-wide)", IsLocal: true},