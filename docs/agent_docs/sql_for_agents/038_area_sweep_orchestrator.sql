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