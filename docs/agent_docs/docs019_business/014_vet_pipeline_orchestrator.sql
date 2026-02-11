/*vet-pipeline-orchestrator:
  spawn_sweep_orchestrator (area-sweep-orchestrator)
  call_agent → waits for it to finish
  promote_candidates (local action)
  spawn_batch_processor (vet-batch-processor)
  call_agent → waits for it to finish
  complete*/

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

-- ==========================================================================
-- Agent Definition: vet-pipeline-orchestrator
-- Database: clients_db
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
             'Runs the full vet discovery pipeline: sweep areas, promote candidates, dispatch verifiers. Uses rolling execution so each run advances work from previous runs.',
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
             true,
             '["pipeline", "discovery", "verification", "veterinary"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.770',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["veterinary", "business-intelligence", "pipeline"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"optional": ["area_code", "sweep_limit", "verify_limit", "promote_limit", "delay_ms", "vertical_slug", "country", "business_type"], "required": []}'::jsonb,
             '{"produces": {"unswept_areas": "object - areas loaded", "sweep_dispatch_result": "object - sweep dispatches", "promote_result": "object - candidates promoted", "verify_dispatch_result": "object - verifier dispatches"}}'::jsonb
         )
    ON CONFLICT (type, version)
DO UPDATE SET
    display_name = EXCLUDED.display_name,
           description = EXCLUDED.description,
           default_config = EXCLUDED.default_config,
           is_active = EXCLUDED.is_active,
           capabilities = EXCLUDED.capabilities,
           input_contract = EXCLUDED.input_contract,
           output_contract = EXCLUDED.output_contract,
           updated_at = NOW();

---

going for separate actions spawned
          -- vet_pipeline_orchestrator.sql
--
-- Updated pipeline workflow: spawn+loop instead of fire-and-forget dispatch.
--
-- Flow:
--   load_unswept_areas
--   → spawn_sweeper (one agent)
--   → sweep_areas (loop: call_agent per district, continue_on_error)
--   → promote_candidates
--   → load_pending_businesses (new action to get business list for loop)
--   → spawn_verifier (one agent)
--   → verify_businesses (loop: call_agent per business, continue_on_error)
--   → complete
--
-- The sweeper and verifier are each ONE spawned agent, called repeatedly
-- via the loop. Topics are handled by the framework. Failures in individual
-- iterations are skipped (continue_on_error: true) and logged.
--
-- NOTE: the dispatch_area_discoverers and dispatch_verifiers actions are no
-- longer used by this workflow. They can be kept for manual/CLI use.

-- First update the query_template fix (town → postcode) if not already done
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_practice,config,query_template}',
        '"{{.business_record.business.name}} {{.business_record.business.postcode}} veterinary practice"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier'
  AND default_config->'workflow'->'steps'->'search_practice'->'config'->>'query_template'
    LIKE '%town%';

-- Now update the pipeline orchestrator workflow
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 7200,
    "workflow": {
        "start_step": "load_areas",
        "steps": {

            "load_areas": {
                "action": "load_unswept_areas",
                "config": {
                    "input_fields": ["limit", "country", "business_type", "area_code"]
                },
                "output_field": "unswept_areas",
                "next_step": "spawn_sweeper",
                "description": "Load un-swept postcode districts from DB"
            },

            "spawn_sweeper": {
                "action": "spawn_agent",
                "config": {
                    "role": "sweeper",
                    "agent_type": "area-sweep-discoverer"
                },
                "output_field": "sweeper_agent",
                "next_step": "sweep_areas",
                "description": "Spawn a single area sweep agent"
            },

            "sweep_areas": {
                "action": "loop",
                "config": {
                    "items_field": "unswept_areas.areas",
                    "item_variable": "current_area",
                    "max_iterations": 200,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_sweeper",
                        "steps": {
                            "call_sweeper": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "area-sweep-discoverer",
                                    "target_role": "sweeper",
                                    "input_mapping": {
                                        "district_code": "current_area.district_code",
                                        "area_name": "current_area.area_name",
                                        "search_area_id": "current_area.search_area_id",
                                        "business_type": "unswept_areas.business_type"
                                    },
                                    "timeout_seconds": 120
                                },
                                "output_field": "sweep_result",
                                "description": "Call sweeper for this district"
                            }
                        }
                    }
                },
                "output_field": "sweep_results",
                "next_step": "promote_candidates",
                "description": "Loop through districts, calling sweeper for each"
            },

            "promote_candidates": {
                "action": "promote_discovery_candidates",
                "config": {
                    "input_fields": ["vertical_slug", "min_confidence"],
                    "vertical_slug": "veterinary",
                    "min_confidence": 0.6
                },
                "output_field": "promotion_result",
                "next_step": "load_pending_businesses",
                "description": "Promote high-confidence candidates to businesses"
            },

            "load_pending_businesses": {
                "action": "load_pending_verifications",
                "config": {
                    "input_fields": ["verify_limit", "vertical_slug"],
                    "verify_limit": 100,
                    "vertical_slug": "veterinary"
                },
                "output_field": "pending_businesses",
                "next_step": "check_pending",
                "description": "Load businesses needing verification"
            },

            "check_pending": {
                "action": "conditional",
                "config": {
                    "condition": "pending_businesses.count > 0",
                    "then_step": "spawn_verifier",
                    "else_step": "complete"
                },
                "description": "Skip verification if nothing pending"
            },

            "spawn_verifier": {
                "action": "spawn_agent",
                "config": {
                    "role": "verifier",
                    "agent_type": "vet-practice-verifier"
                },
                "output_field": "verifier_agent",
                "next_step": "verify_businesses",
                "description": "Spawn a single verifier agent"
            },

            "verify_businesses": {
                "action": "loop",
                "config": {
                    "items_field": "pending_businesses.businesses",
                    "item_variable": "current_business",
                    "max_iterations": 100,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_verifier",
                        "steps": {
                            "call_verifier": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "vet-practice-verifier",
                                    "target_role": "verifier",
                                    "input_mapping": {
                                        "business_id": "current_business.id"
                                    },
                                    "timeout_seconds": 300
                                },
                                "output_field": "verify_result",
                                "description": "Call verifier for this business"
                            }
                        }
                    }
                },
                "output_field": "verify_results",
                "next_step": "complete",
                "description": "Loop through businesses, calling verifier for each"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["unswept_areas", "sweep_results", "promotion_result", "pending_businesses", "verify_results"]
                },
                "description": "Pipeline run complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';