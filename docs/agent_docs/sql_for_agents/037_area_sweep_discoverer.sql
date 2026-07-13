-- ==========================================================================
-- Area Sweep Discoverer agent definition
-- Searches for businesses in a postcode district and inserts unknown ones
-- as discovery candidates.
--
-- Input: { "district_code": "BT4", "area_name": "Belfast", "search_area_id": "uuid" }
-- ==========================================================================

INSERT INTO agent_definitions (
    type, name, description, processing_mode, default_config,
    is_active, tags, container_image, container_version,
    resource_config, topic_config, health_check_config,
    dependencies, replicas, orchestration_config,
    agent_class, lifecycle_stage, capabilities,
    input_schema, output_schema
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
                             "next_step": "process_results",
                             "description": "Search for vet practices in this postcode district",
                             "output_field": "search_results"
                         },
                         "process_results": {
                             "action": "process_area_sweep",
                             "config": {
                                 "input_fields": ["district_code", "area_name", "search_area_id", "search_results"]
                             },
                             "next_step": "complete",
                             "description": "Check results against known businesses and create discovery candidates",
                             "output_field": "sweep_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["search_results", "sweep_result"]
                             },
                             "description": "Area sweep complete"
                         }
                     }
                 },
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-haiku-4-5",
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["data-collection", "discovery", "veterinary", "geographic-sweep"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.759',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}',
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}',
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}',
             '[]',
             1,
             NULL, NULL, NULL,
             'specialist',
             'experimental',
             '["veterinary", "business-intelligence", "discovery"]',
             '{"required": ["district_code"], "optional": ["area_name", "search_area_id"]}',
             '{"produces": {"search_results": "object - raw search results", "sweep_result": "object - candidates found and tracking update"}}'
         )
    ON CONFLICT (type) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              tags = EXCLUDED.tags,
                              input_schema = EXCLUDED.input_schema,
                              output_schema = EXCLUDED.output_schema,
                              updated_at = NOW();

-- Verify
SELECT type, name,
       default_config->'workflow'->'start_step' as start_step,
       jsonb_object_keys(default_config->'workflow'->'steps') as steps
FROM agent_definitions
WHERE type = 'area-sweep-discoverer';

--

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


-- revision
-- 2. area-sweep-discoverer
-- Searches one postcode district, processes results into discovery candidates.
-- Called by orchestrator with: {"district_code": "BT4", "area_name": "Belfast", ...}
-- Also callable standalone with just a district_code.

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
                                 "num_results": 10,
                                 "query_template": "{{.input_data.business_type}} {{.input_data.district_code}} {{.input_data.area_name}} UK"
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

