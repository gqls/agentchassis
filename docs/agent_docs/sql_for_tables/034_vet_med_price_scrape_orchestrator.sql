-- med_pricing_spawn_setup.sql
-- Sets up spawn-based execution for med pricing agents.
-- Business-intel receives triggers → spawns temporary pods → pods do the work → terminate.

-- ============================================================================
-- 1. Orchestrator definitions (thin wrappers that spawn + call workers)
-- ============================================================================

-- Price scrape orchestrator
INSERT INTO agent_definitions (
    type, name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-price-scrape-orchestrator',
             'Med Price Scrape Orchestrator',
             'Spawns a temporary pod to scrape veterinary medicine prices. Thin wrapper around med-price-collector.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.935',
             true, 'active',
             '{
                 "workflow": {
                     "initial_step": "spawn_scraper",
                     "steps": {
                         "spawn_scraper": {
                             "action": "spawn_agent",
                             "config": {
                                 "agent_type": "med-price-collector",
                                 "role": "price_scraper"
                             },
                             "next_step": "call_scraper",
                             "output_field": "scraper_spawn",
                             "description": "Spawn a temporary pod for price scraping"
                         },
                         "call_scraper": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "med-price-collector",
                                 "target_role": "price_scraper",
                                 "input_mapping": {
                                     "input_data": "input_data"
                                 },
                                 "timeout_seconds": 600
                             },
                             "next_step": "complete",
                             "output_field": "scrape_result",
                             "description": "Send scrape work to spawned pod and wait for completion"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Scrape run complete"
                         }
                     }
                 }
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[]'::jsonb,
             0
         ) ON CONFLICT (type) WHERE deleted_at IS NULL
    DO UPDATE SET
    default_config = EXCLUDED.default_config,
           description = EXCLUDED.description,
           updated_at = NOW();

-- URL discovery orchestrator
INSERT INTO agent_definitions (
    type, name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-url-discover-orchestrator',
             'Med URL Discover Orchestrator',
             'Spawns a temporary pod to discover product URLs from retailer category pages.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.935',
             true, 'active',
             '{
                 "workflow": {
                     "initial_step": "spawn_discoverer",
                     "steps": {
                         "spawn_discoverer": {
                             "action": "spawn_agent",
                             "config": {
                                 "agent_type": "med-url-discoverer",
                                 "role": "url_discoverer"
                             },
                             "next_step": "call_discoverer",
                             "output_field": "discoverer_spawn",
                             "description": "Spawn a temporary pod for URL discovery"
                         },
                         "call_discoverer": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "med-url-discoverer",
                                 "target_role": "url_discoverer",
                                 "input_mapping": {
                                     "input_data": "input_data"
                                 },
                                 "timeout_seconds": 600
                             },
                             "next_step": "complete",
                             "output_field": "discover_result",
                             "description": "Send discovery work to spawned pod and wait for completion"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Discovery run complete"
                         }
                     }
                 }
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[]'::jsonb,
             0
         ) ON CONFLICT (type) WHERE deleted_at IS NULL
    DO UPDATE SET
    default_config = EXCLUDED.default_config,
           description = EXCLUDED.description,
           updated_at = NOW();


-- ============================================================================
-- 2. Update worker definitions with env_vars for the spawned pod
--    FIRECRAWL_API_KEY comes via spawn_actions.go secretKeyRef (see Go patch below).
--    Non-secret config goes in env_vars here.
-- ============================================================================

UPDATE agent_definitions
SET env_vars = '[
    {"name": "FIRECRAWL_API_URL", "value": "https://api.firecrawl.dev/v2"},
    {"name": "OLLAMA_URL", "value": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"},
    {"name": "MED_PRICE_LLM_MODEL", "value": "mistral-small3.1"}
]'::jsonb,
    updated_at = NOW()
WHERE type IN ('med-price-collector', 'med-url-discoverer');


-- ============================================================================
-- 3. Update scheduled tasks to trigger orchestrators (not workers directly)
-- ============================================================================

UPDATE scheduled_tasks
SET agent_type = 'med-price-scrape-orchestrator',
    message_payload = '{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":20}}'::jsonb,
    updated_at = NOW()
WHERE name = 'med-scrape-prices';

UPDATE scheduled_tasks
SET agent_type = 'med-url-discover-orchestrator',
    message_payload = '{"action":"orchestrate","config":{"agent_type":"med-url-discover-orchestrator"},"input_data":{}}'::jsonb,
    updated_at = NOW()
WHERE name = 'med-discover-urls';


-- ============================================================================
-- Verify
-- ============================================================================
-- SELECT type, name FROM agent_definitions WHERE type LIKE 'med-%' AND is_active = true;
-- SELECT name, agent_type, target_topic FROM scheduled_tasks WHERE name LIKE 'med-%';


-- med_pricing_spawn_setup.sql (FIXED)
-- Checked against: \d agent_definitions, \d scheduled_tasks

-- ============================================================================
-- 1. Orchestrator definitions (thin wrappers that spawn + call workers)
-- ============================================================================

-- Price scrape orchestrator
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-price-scrape-orchestrator',
             'Med Price Scrape Orchestrator',
             'Spawns a temporary pod to scrape veterinary medicine prices.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.935',
             true, 'active',
             '{
                 "workflow": {
                     "initial_step": "spawn_scraper",
                     "steps": {
                         "spawn_scraper": {
                             "action": "spawn_agent",
                             "config": {
                                 "agent_type": "med-price-collector",
                                 "role": "price_scraper"
                             },
                             "next_step": "call_scraper",
                             "output_field": "scraper_spawn",
                             "description": "Spawn a temporary pod for price scraping"
                         },
                         "call_scraper": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "med-price-collector",
                                 "target_role": "price_scraper",
                                 "input_mapping": {
                                     "input_data": "input_data"
                                 },
                                 "timeout_seconds": 600
                             },
                             "next_step": "complete",
                             "output_field": "scrape_result",
                             "description": "Send scrape work to spawned pod and wait"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Scrape run complete"
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

-- URL discovery orchestrator
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-url-discover-orchestrator',
             'Med URL Discover Orchestrator',
             'Spawns a temporary pod to discover product URLs from retailer category pages.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.935',
             true, 'active',
             '{
                 "workflow": {
                     "initial_step": "spawn_discoverer",
                     "steps": {
                         "spawn_discoverer": {
                             "action": "spawn_agent",
                             "config": {
                                 "agent_type": "med-url-discoverer",
                                 "role": "url_discoverer"
                             },
                             "next_step": "call_discoverer",
                             "output_field": "discoverer_spawn",
                             "description": "Spawn a temporary pod for URL discovery"
                         },
                         "call_discoverer": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "med-url-discoverer",
                                 "target_role": "url_discoverer",
                                 "input_mapping": {
                                     "input_data": "input_data"
                                 },
                                 "timeout_seconds": 600
                             },
                             "next_step": "complete",
                             "output_field": "discover_result",
                             "description": "Send discovery work to spawned pod and wait"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Discovery run complete"
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
-- 2. Update worker definitions with env_vars for the spawned pod
-- ============================================================================

UPDATE agent_definitions
SET env_vars = '[
    {"name": "FIRECRAWL_API_URL", "value": "https://api.firecrawl.dev/v2"},
    {"name": "OLLAMA_URL", "value": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"},
    {"name": "MED_PRICE_LLM_MODEL", "value": "mistral-small3.1"}
]'::jsonb,
    updated_at = NOW()
WHERE type IN ('med-price-collector', 'med-url-discoverer');


-- ============================================================================
-- 3. Update scheduled tasks to trigger orchestrators
--    Columns: target_agent_type, target_topic, input_data
-- ============================================================================

UPDATE scheduled_tasks
SET target_agent_type = 'med-price-scrape-orchestrator',
    input_data = '{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":20}}'::jsonb,
    updated_at = NOW()
WHERE name = 'med-scrape-prices';

UPDATE scheduled_tasks
SET target_agent_type = 'med-url-discover-orchestrator',
    input_data = '{"action":"orchestrate","config":{"agent_type":"med-url-discover-orchestrator"},"input_data":{}}'::jsonb,
    updated_at = NOW()
WHERE name = 'med-discover-urls';


-- ============================================================================
-- Verify
-- ============================================================================
SELECT type, display_name, status FROM agent_definitions WHERE type LIKE 'med-%' AND is_active = true ORDER BY type;
SELECT name, target_agent_type, target_topic, enabled FROM scheduled_tasks WHERE name LIKE 'med-%';