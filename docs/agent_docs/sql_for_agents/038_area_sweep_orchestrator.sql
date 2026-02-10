-- ==========================================================================
-- Agent definitions for geographic area sweep discovery
-- ==========================================================================

-- 1. area-sweep-orchestrator
-- Loads un-swept postcode districts from DB, dispatches discoverer agents.
-- Kick off with: {"limit": 50, "country": "GB"}

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, domain_tags,
    agent_category, status,
    input_contract, output_contract
) VALUES (
             'area-sweep-orchestrator',
             'Area Sweep Orchestrator',
             'Loads un-swept postcode districts from the search_areas table and dispatches area-sweep-discoverer agents for each one.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "load_areas",
                     "steps": {
                         "load_areas": {
                             "action": "load_unswept_areas",
                             "config": {
                                 "input_fields": ["limit", "country", "business_type", "area_code"]
                             },
                             "output_field": "unswept_areas",
                             "next_step": "dispatch_discoverers",
                             "description": "Load un-swept postcode districts from DB"
                         },
                         "dispatch_discoverers": {
                             "action": "dispatch_area_discoverers",
                             "config": {
                                 "input_fields": ["unswept_areas"]
                             },
                             "output_field": "dispatch_result",
                             "next_step": "complete",
                             "description": "Send a Kafka message per district to trigger discoverers"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["unswept_areas", "dispatch_result"]
                             },
                             "description": "Sweep batch dispatched"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["discovery", "geographic-sweep", "veterinary"]'::jsonb,
             '["veterinary", "business-intelligence", "discovery"]'::jsonb,
             'coordinator',
             'experimental',
             '{"required": [], "optional": ["limit", "country", "business_type", "area_code"]}'::jsonb,
             '{"produces": {"unswept_areas": "object - areas loaded and stats", "dispatch_result": "object - how many dispatched"}}'::jsonb
         )
    ON CONFLICT (type) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();


-- 2. area-sweep-discoverer
-- Searches one postcode district, processes results into discovery candidates.
-- Called by orchestrator with: {"district_code": "BT4", "area_name": "Belfast", "search_area_id": "uuid"}

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, domain_tags,
    agent_category, status,
    input_contract, output_contract
) VALUES (
             'area-sweep-discoverer',
             'Area Sweep Discoverer',
             'Searches for veterinary practices within a UK postcode district and creates discovery candidates for unknown businesses.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "search_area",
                     "steps": {
                         "search_area": {
                             "action": "web_search",
                             "config": {
                                 "num_results": 10
                             },
                             "input_map": {
                                 "query": "veterinary practice {{.district_code}} {{.area_name}} UK"
                             },
                             "output_field": "search_results",
                             "next_step": "process_results",
                             "description": "Search for vet practices in this postcode district"
                         },
                         "process_results": {
                             "action": "process_area_sweep",
                             "config": {
                                 "input_fields": ["district_code", "area_name", "search_area_id", "search_results"]
                             },
                             "output_field": "sweep_result",
                             "next_step": "complete",
                             "description": "Check results against known businesses, insert candidates"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["search_results", "sweep_result"]
                             },
                             "description": "District sweep complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["discovery", "web-search", "veterinary"]'::jsonb,
             '["veterinary", "business-intelligence", "discovery"]'::jsonb,
             'specialist',
             'experimental',
             '{"required": ["district_code"], "optional": ["area_name", "search_area_id", "business_type"]}'::jsonb,
             '{"produces": {"search_results": "object - raw search results", "sweep_result": "object - candidates found counts"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();


-- Verify
SELECT type, display_name, status,
       default_config->'workflow'->'start_step' as start_step
FROM agent_definitions
WHERE type IN ('area-sweep-orchestrator', 'area-sweep-discoverer');