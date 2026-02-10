-- ==========================================================================
-- Agent Definition: vet-pipeline-orchestrator
-- Database: clients_db
-- ==========================================================================
--
-- Rolling pipeline: each run advances work from previous runs.
--
-- Step 1 (load_areas): Load unswept postcode districts
-- Step 2 (dispatch_sweepers): Fire-and-forget Kafka messages to area-sweep-discoverer agents
-- Step 3 (promote): Move discovery_candidates → businesses (from PREVIOUS sweeps)
-- Step 4 (dispatch_verifiers): Fire-and-forget Kafka messages to vet-practice-verifier agents
-- Step 5 (complete): Done
--
-- Usage:
--   Send one Kafka message to system.agent.generic.requests with:
--   {"action":"orchestrate","config":{"agent_type":"vet-pipeline-orchestrator"},"input_data":{}}
--
--   Optional input_data fields:
--     area_code:       Filter sweep to specific area (e.g. "BT" for Belfast)
--     sweep_limit:     Max areas to sweep (default 50)
--     verify_limit:    Max businesses to verify (default 100)
--     promote_limit:   Max candidates to promote (default 500)
--     delay_ms:        Delay between Kafka dispatches in ms (default 200)
-- ==========================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config, env_vars,
    version, delegation_preferences,
    agent_category, status, domain_tags,
    briefing_questionnaire, usage_count, is_snapshot,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'vet-pipeline-orchestrator',
             'Vet Pipeline Orchestrator',
             'Runs the full vet discovery pipeline: sweep areas → promote candidates → dispatch verifiers. Uses rolling execution — each run advances work from previous runs.',
             'orchestrator',

             -- default_config with workflow
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
                             "next_step": "dispatch_sweepers",
                             "description": "Load postcode districts that need sweeping"
                         },
                         "dispatch_sweepers": {
                             "action": "dispatch_area_discoverers",
                             "config": {
                                 "input_fields": ["unswept_areas"]
                             },
                             "output_field": "sweep_dispatch_result",
                             "next_step": "promote",
                             "description": "Fire-and-forget: send one Kafka message per district"
                         },
                         "promote": {
                             "action": "promote_candidates",
                             "config": {
                                 "input_fields": ["promote_limit", "vertical_slug"]
                             },
                             "output_field": "promote_result",
                             "next_step": "dispatch_verifiers",
                             "description": "Move pending discovery candidates into businesses table"
                         },
                         "dispatch_verifiers": {
                             "action": "dispatch_verifiers",
                             "config": {
                                 "input_fields": ["verify_limit", "vertical_slug", "delay_ms"]
                             },
                             "output_field": "verify_dispatch_result",
                             "next_step": "complete",
                             "description": "Fire-and-forget: send one Kafka message per pending business"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["unswept_areas", "sweep_dispatch_result", "promote_result", "verify_dispatch_result"]
                             },
                             "description": "Pipeline run complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300
             }'::jsonb,

             true,  -- is_active
             '["pipeline", "discovery", "verification", "veterinary"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.770',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,  -- env_vars
             1,            -- version
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',    -- agent_category
             'experimental',   -- status
             '["veterinary", "business-intelligence", "pipeline"]'::jsonb,
             '{}'::jsonb,      -- briefing_questionnaire
             0,                -- usage_count
             false,            -- is_snapshot
             '{"optional": ["area_code", "sweep_limit", "verify_limit", "promote_limit", "delay_ms", "vertical_slug", "country", "business_type"], "required": []}'::jsonb,
             '{"produces": {"unswept_areas": "object - areas loaded", "sweep_dispatch_result": "object - sweep dispatches", "promote_result": "object - candidates promoted", "verify_dispatch_result": "object - verifier dispatches"}}'::jsonb
         )
    ON CONFLICT (type) WHERE deleted_at IS NULL
    DO UPDATE SET
    display_name = EXCLUDED.display_name,
           description = EXCLUDED.description,
           default_config = EXCLUDED.default_config,
           is_active = EXCLUDED.is_active,
           capabilities = EXCLUDED.capabilities,
           input_contract = EXCLUDED.input_contract,
           output_contract = EXCLUDED.output_contract,
           updated_at = NOW();