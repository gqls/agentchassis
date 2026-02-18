/*vet-pipeline-orchestrator:
  spawn_sweep_orchestrator (area-sweep-orchestrator)
  call_agent → waits for it to finish
  promote_candidates (local action)
  spawn_batch_processor (vet-batch-processor)
  call_agent → waits for it to finish
  complete

  Sources (many)                    Queue (one)              Processor (one)
─────────────                    ──────────              ──────────────
area sweep → promote ─┐
manual entry ─────────┤
bought list import ───┼──→  collection_tasks  ──→  vet-batch-processor
API submission ───────┤         (pending)            reads & claims tasks
future sources ───────┘
  */

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


---

-- vet_pipeline_updates.sql
--
-- Three agent definition updates:
--
-- 1. area-sweep-orchestrator: convert from dispatch (fire-and-forget) to
--    spawn+loop so the pipeline can call it and wait for completion.
--
-- 2. vet-batch-processor: add continue_on_error to its existing loop
--    so one failed verification doesn't kill the batch.
--
-- 3. vet-pipeline-orchestrator: thin coordinator that calls the two
--    child orchestrators in sequence with promote_candidates between.
--
-- Also fixes the query_template town→postcode issue if not already done.


-- =====================================================================
-- Fix: query_template town → postcode in vet-practice-verifier
-- (idempotent — only updates if still using town)
-- =====================================================================
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


-- =====================================================================
-- 1. area-sweep-orchestrator: dispatch → spawn+loop
--
-- Before: load_areas → dispatch_discoverers (fire-and-forget) → complete
-- After:  load_areas → spawn_discoverer → sweep_loop → complete
--
-- The loop calls the spawned area-sweep-discoverer per district.
-- continue_on_error: true so one failed district doesn't stop the batch.
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 3600,
    "workflow": {
        "start_step": "load_areas",
        "steps": {

            "load_areas": {
                "action": "load_unswept_areas",
                "config": {
                    "input_fields": ["limit", "country", "business_type", "area_code"]
                },
                "output_field": "unswept_areas",
                "next_step": "check_areas",
                "description": "Load un-swept postcode districts from DB"
            },

            "check_areas": {
                "action": "conditional",
                "config": {
                    "condition": "unswept_areas.count > 0",
                    "then_step": "spawn_discoverer",
                    "else_step": "complete"
                },
                "description": "Skip if no areas to sweep"
            },

            "spawn_discoverer": {
                "action": "spawn_agent",
                "config": {
                    "role": "discoverer",
                    "agent_type": "area-sweep-discoverer"
                },
                "output_field": "discoverer_agent",
                "next_step": "sweep_loop",
                "description": "Spawn a single area sweep discoverer"
            },

            "sweep_loop": {
                "action": "loop",
                "config": {
                    "items_field": "unswept_areas.areas",
                    "item_variable": "current_area",
                    "max_iterations": 200,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_discoverer",
                        "steps": {
                            "call_discoverer": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "area-sweep-discoverer",
                                    "target_role": "discoverer",
                                    "input_mapping": {
                                        "district_code": "current_area.district_code",
                                        "area_name": "current_area.area_name",
                                        "search_area_id": "current_area.search_area_id",
                                        "business_type": "unswept_areas.business_type"
                                    },
                                    "timeout_seconds": 120
                                },
                                "output_field": "sweep_result",
                                "description": "Call discoverer for this district"
                            }
                        }
                    }
                },
                "output_field": "sweep_results",
                "next_step": "complete",
                "description": "Loop through districts, calling discoverer for each"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["unswept_areas", "sweep_results"]
                },
                "description": "Area sweep batch complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'area-sweep-orchestrator';


-- =====================================================================
-- 2. vet-batch-processor: add continue_on_error to loop
--
-- Only change: continue_on_error: true added to the process_batch loop.
-- Everything else (load_batch, spawn_verifier, etc.) stays the same.
-- =====================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_batch,config,continue_on_error}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';


-- =====================================================================
-- 3. vet-pipeline-orchestrator: thin coordinator
--
-- Spawns and calls the two child orchestrators in sequence.
-- promote_candidates runs between sweeps and verification.
--
-- Flow:
--   spawn sweep-orch → call (waits for all sweeps) →
--   promote_candidates →
--   spawn batch-processor → call (waits for all verifications) →
--   complete
--
-- NOTE: promote_candidates action needs the collection_task INSERT added
--       It should promote high-confidence discovery_candidates to the
--       businesses table. Until implemented, this step will fail
--       gracefully or can be commented out.
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 7200,
    "workflow": {
        "start_step": "spawn_sweep_orch",
        "steps": {

            "spawn_sweep_orch": {
                "action": "spawn_agent",
                "config": {
                    "role": "sweep-coordinator",
                    "agent_type": "area-sweep-orchestrator"
                },
                "output_field": "sweep_orch",
                "next_step": "run_sweeps",
                "description": "Spawn the area sweep orchestrator"
            },

            "run_sweeps": {
                "action": "call_agent",
                "config": {
                    "agent_type": "area-sweep-orchestrator",
                    "target_role": "sweep-coordinator",
                    "input_mapping": {
                        "limit": "input_data.sweep_limit",
                        "area_code": "input_data.area_code",
                        "country": "input_data.country",
                        "business_type": "input_data.business_type"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "sweep_result",
                "next_step": "promote_candidates",
                "description": "Call sweep orchestrator and wait for completion"
            },

            "promote_candidates": {
                "action": "promote_candidates",
                "config": {
                    "vertical_slug": "veterinary",
                    "promote_limit": 500
                },
                "output_field": "promotion_result",
                "next_step": "spawn_batch_processor",
                "description": "Promote high-confidence candidates to businesses"
            },

            "spawn_batch_processor": {
                "action": "spawn_agent",
                "config": {
                    "role": "batch-verifier",
                    "agent_type": "vet-batch-processor"
                },
                "output_field": "batch_orch",
                "next_step": "run_verification",
                "description": "Spawn the batch verification orchestrator"
            },

            "run_verification": {
                "action": "call_agent",
                "config": {
                    "agent_type": "vet-batch-processor",
                    "target_role": "batch-verifier",
                    "input_mapping": {
                        "batch_size": "input_data.verify_batch_size",
                        "vertical_slug": "input_data.vertical_slug"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "verification_result",
                "next_step": "complete",
                "description": "Call batch processor and wait for completion"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["sweep_result", "promotion_result", "verification_result"]
                },
                "description": "Pipeline run complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';

--

-- data position bugs, input_mappers are compulsory

-- vet_pipeline_updates.sql
--
-- Three agent definition updates:
--
-- 1. area-sweep-orchestrator: convert from dispatch (fire-and-forget) to
--    spawn+loop so the pipeline can call it and wait for completion.
--
-- 2. vet-batch-processor: add continue_on_error to its existing loop
--    so one failed verification doesn't kill the batch.
--
-- 3. vet-pipeline-orchestrator: thin coordinator that calls the two
--    child orchestrators in sequence with promote_candidates between.
--
-- Also fixes the query_template town→postcode issue if not already done.


-- =====================================================================
-- Fix: query_template town → postcode in vet-practice-verifier
-- (idempotent — only updates if still using town)
-- =====================================================================
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


-- =====================================================================
-- 1. area-sweep-orchestrator: dispatch → spawn+loop
--
-- Before: load_areas → dispatch_discoverers (fire-and-forget) → complete
-- After:  load_areas → spawn_discoverer → sweep_loop → complete
--
-- The loop calls the spawned area-sweep-discoverer per district.
-- continue_on_error: true so one failed district doesn't stop the batch.
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 3600,
    "workflow": {
        "start_step": "load_areas",
        "steps": {

            "load_areas": {
                "action": "load_unswept_areas",
                "config": {
                    "input_fields": ["limit", "country", "business_type", "area_code"]
                },
                "output_field": "unswept_areas",
                "next_step": "check_areas",
                "description": "Load un-swept postcode districts from DB"
            },

            "check_areas": {
                "action": "conditional",
                "config": {
                    "condition": "unswept_areas.count > 0",
                    "then_step": "spawn_discoverer",
                    "else_step": "complete"
                },
                "description": "Skip if no areas to sweep"
            },

            "spawn_discoverer": {
                "action": "spawn_agent",
                "config": {
                    "role": "discoverer",
                    "agent_type": "area-sweep-discoverer"
                },
                "output_field": "discoverer_agent",
                "next_step": "sweep_loop",
                "description": "Spawn a single area sweep discoverer"
            },

            "sweep_loop": {
                "action": "loop",
                "config": {
                    "items_field": "unswept_areas.areas",
                    "item_variable": "current_area",
                    "max_iterations": 200,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_discoverer",
                        "steps": {
                            "call_discoverer": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "area-sweep-discoverer",
                                    "target_role": "discoverer",
                                    "input_mapping": {
                                        "district_code": "current_area.district_code",
                                        "area_name": "current_area.area_name",
                                        "search_area_id": "current_area.search_area_id",
                                        "business_type": "unswept_areas.business_type"
                                    },
                                    "timeout_seconds": 120
                                },
                                "output_field": "sweep_result",
                                "description": "Call discoverer for this district"
                            }
                        }
                    }
                },
                "output_field": "sweep_results",
                "next_step": "complete",
                "description": "Loop through districts, calling discoverer for each"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["unswept_areas", "sweep_results"]
                },
                "description": "Area sweep batch complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'area-sweep-orchestrator';


-- =====================================================================
-- 2. vet-batch-processor: add continue_on_error to loop
--
-- Only change: continue_on_error: true added to the process_batch loop.
-- Everything else (load_batch, spawn_verifier, etc.) stays the same.
-- =====================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_batch,config,continue_on_error}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';


-- =====================================================================
-- 3. vet-pipeline-orchestrator: thin coordinator
--
-- Spawns and calls the two child orchestrators in sequence.
-- promote_candidates runs between sweeps and verification.
--
-- Flow:
--   spawn sweep-orch → call (waits for all sweeps) →
--   promote_candidates →
--   spawn batch-processor → call (waits for all verifications) →
--   complete
--
-- NOTE: promote_candidates action needs the collection_task INSERT added
--       It should promote high-confidence discovery_candidates to the
--       businesses table. Until implemented, this step will fail
--       gracefully or can be commented out.
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 7200,
    "workflow": {
        "start_step": "spawn_sweep_orch",
        "steps": {

            "spawn_sweep_orch": {
                "action": "spawn_agent",
                "config": {
                    "role": "sweep-coordinator",
                    "agent_type": "area-sweep-orchestrator"
                },
                "output_field": "sweep_orch",
                "next_step": "run_sweeps",
                "description": "Spawn the area sweep orchestrator"
            },

            "run_sweeps": {
                "action": "call_agent",
                "config": {
                    "agent_type": "area-sweep-orchestrator",
                    "target_role": "sweep-coordinator",
                    "input_mapping": {
                        "limit": "input_data.limit"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "sweep_result",
                "next_step": "promote_candidates",
                "description": "Call sweep orchestrator and wait for completion"
            },

            "promote_candidates": {
                "action": "promote_candidates",
                "config": {
                    "vertical_slug": "veterinary",
                    "promote_limit": 500
                },
                "output_field": "promotion_result",
                "next_step": "spawn_batch_processor",
                "description": "Promote high-confidence candidates to businesses"
            },

            "spawn_batch_processor": {
                "action": "spawn_agent",
                "config": {
                    "role": "batch-verifier",
                    "agent_type": "vet-batch-processor"
                },
                "output_field": "batch_orch",
                "next_step": "run_verification",
                "description": "Spawn the batch verification orchestrator"
            },

            "run_verification": {
                "action": "call_agent",
                "config": {
                    "agent_type": "vet-batch-processor",
                    "target_role": "batch-verifier",
                    "input_mapping": {
                        "batch_size": "input_data.verify_limit"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "verification_result",
                "next_step": "complete",
                "description": "Call batch processor and wait for completion"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["sweep_result", "promotion_result", "verification_result"]
                },
                "description": "Pipeline run complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';

---

-- business vertical and country are no longer optional

         -- vet_pipeline_updates.sql
--
-- Vertical-agnostic pipeline definitions.
--
-- Vet-specific values (country, business_type, vertical_slug) are defined
-- in the workflow configs, not hardcoded in Go. A seaweed pipeline would
-- have different agent definitions with different values, same Go code.
--
-- Values flow: pipeline input_data → call_agent input_mapping → child input_data
-- Local actions (promote_candidates) read from their step config directly.
--
-- CLI invocation must include: country, business_type, vertical_slug
-- e.g. {"limit": 50, "country": "GB", "business_type": "veterinary practice",
--        "vertical_slug": "veterinary", "verify_limit": 100, "promote_limit": 500}


-- =====================================================================
-- Fix: query_template town → postcode in vet-practice-verifier
-- (idempotent — only updates if still using town)
-- =====================================================================
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


-- =====================================================================
-- 1. area-sweep-orchestrator: dispatch → spawn+loop
--
-- load_areas now requires country and business_type from input_data
-- (passed down by the pipeline's call_agent input_mapping).
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 3600,
    "workflow": {
        "start_step": "load_areas",
        "steps": {

            "load_areas": {
                "action": "load_unswept_areas",
                "config": {
                    "input_fields": ["limit", "country", "business_type", "area_code"]
                },
                "output_field": "unswept_areas",
                "next_step": "check_areas",
                "description": "Load un-swept postcode districts from DB"
            },

            "check_areas": {
                "action": "conditional",
                "config": {
                    "condition": "unswept_areas.count > 0",
                    "then_step": "spawn_discoverer",
                    "else_step": "complete"
                },
                "description": "Skip if no areas to sweep"
            },

            "spawn_discoverer": {
                "action": "spawn_agent",
                "config": {
                    "role": "discoverer",
                    "agent_type": "area-sweep-discoverer"
                },
                "output_field": "discoverer_agent",
                "next_step": "sweep_loop",
                "description": "Spawn a single area sweep discoverer"
            },

            "sweep_loop": {
                "action": "loop",
                "config": {
                    "items_field": "unswept_areas.areas",
                    "item_variable": "current_area",
                    "max_iterations": 200,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_discoverer",
                        "steps": {
                            "call_discoverer": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "area-sweep-discoverer",
                                    "target_role": "discoverer",
                                    "input_mapping": {
                                        "district_code": "current_area.district_code",
                                        "area_name": "current_area.area_name",
                                        "search_area_id": "current_area.search_area_id",
                                        "business_type": "unswept_areas.business_type"
                                    },
                                    "timeout_seconds": 120
                                },
                                "output_field": "sweep_result",
                                "description": "Call discoverer for this district"
                            }
                        }
                    }
                },
                "output_field": "sweep_results",
                "next_step": "complete",
                "description": "Loop through districts, calling discoverer for each"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["unswept_areas", "sweep_results"]
                },
                "description": "Area sweep batch complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'area-sweep-orchestrator';


-- =====================================================================
-- 2. vet-batch-processor: add continue_on_error to loop
-- =====================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_batch,config,continue_on_error}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';


-- =====================================================================
-- 3. vet-pipeline-orchestrator: thin coordinator
--
-- Vet-specific values are:
--   - Passed to children via input_mapping from input_data
--   - Set as static config on local actions (promote_candidates)
--
-- CLI must provide: country, business_type, vertical_slug
-- Optional: limit (default 50), verify_limit, promote_limit, area_code
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 7200,
    "workflow": {
        "start_step": "spawn_sweep_orch",
        "steps": {

            "spawn_sweep_orch": {
                "action": "spawn_agent",
                "config": {
                    "role": "sweep-coordinator",
                    "agent_type": "area-sweep-orchestrator"
                },
                "output_field": "sweep_orch",
                "next_step": "run_sweeps",
                "description": "Spawn the area sweep orchestrator"
            },

            "run_sweeps": {
                "action": "call_agent",
                "config": {
                    "agent_type": "area-sweep-orchestrator",
                    "target_role": "sweep-coordinator",
                    "input_mapping": {
                        "limit": "input_data.limit",
                        "country": "input_data.country",
                        "business_type": "input_data.business_type"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "sweep_result",
                "next_step": "promote_candidates",
                "description": "Call sweep orchestrator and wait for completion"
            },

            "promote_candidates": {
                "action": "promote_candidates",
                "config": {
                    "vertical_slug": "veterinary",
                    "business_type": "veterinary_practice",
                    "country": "GB",
                    "promote_limit": 500
                },
                "output_field": "promotion_result",
                "next_step": "spawn_batch_processor",
                "description": "Promote high-confidence candidates to businesses"
            },

            "spawn_batch_processor": {
                "action": "spawn_agent",
                "config": {
                    "role": "batch-verifier",
                    "agent_type": "vet-batch-processor"
                },
                "output_field": "batch_orch",
                "next_step": "run_verification",
                "description": "Spawn the batch verification orchestrator"
            },

            "run_verification": {
                "action": "call_agent",
                "config": {
                    "agent_type": "vet-batch-processor",
                    "target_role": "batch-verifier",
                    "input_mapping": {
                        "batch_size": "input_data.verify_limit",
                        "vertical_slug": "input_data.vertical_slug"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "verification_result",
                "next_step": "complete",
                "description": "Call batch processor and wait for completion"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["sweep_result", "promotion_result", "verification_result"]
                },
                "description": "Pipeline run complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';

--

-- make country, business etc not optional in initial message, area code is optional '?' parameter mapping

-- vet_pipeline_updates.sql
--
-- Vertical-agnostic pipeline definitions.
--
-- Vet-specific values (country, business_type, vertical_slug) are defined
-- in the workflow configs, not hardcoded in Go. A seaweed pipeline would
-- have different agent definitions with different values, same Go code.
--
-- Values flow: pipeline input_data → call_agent input_mapping → child input_data
-- Local actions (promote_candidates) read from their step config directly.
--
-- CLI invocation must include: country, business_type, vertical_slug
-- e.g. {"limit": 50, "country": "GB", "business_type": "veterinary practice",
--        "vertical_slug": "veterinary", "verify_limit": 100, "promote_limit": 500}


-- =====================================================================
-- Fix: query_template town → postcode in vet-practice-verifier
-- (idempotent — only updates if still using town)
-- =====================================================================
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


-- =====================================================================
-- 1. area-sweep-orchestrator: dispatch → spawn+loop
--
-- load_areas now requires country and business_type from input_data
-- (passed down by the pipeline's call_agent input_mapping).
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 3600,
    "workflow": {
        "start_step": "load_areas",
        "steps": {

            "load_areas": {
                "action": "load_unswept_areas",
                "config": {
                    "input_fields": ["limit", "country", "business_type", "area_code"]
                },
                "output_field": "unswept_areas",
                "next_step": "check_areas",
                "description": "Load un-swept postcode districts from DB"
            },

            "check_areas": {
                "action": "conditional",
                "config": {
                    "condition": "unswept_areas.count > 0",
                    "then_step": "spawn_discoverer",
                    "else_step": "complete"
                },
                "description": "Skip if no areas to sweep"
            },

            "spawn_discoverer": {
                "action": "spawn_agent",
                "config": {
                    "role": "discoverer",
                    "agent_type": "area-sweep-discoverer"
                },
                "output_field": "discoverer_agent",
                "next_step": "sweep_loop",
                "description": "Spawn a single area sweep discoverer"
            },

            "sweep_loop": {
                "action": "loop",
                "config": {
                    "items_field": "unswept_areas.areas",
                    "item_variable": "current_area",
                    "max_iterations": 200,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_discoverer",
                        "steps": {
                            "call_discoverer": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "area-sweep-discoverer",
                                    "target_role": "discoverer",
                                    "input_mapping": {
                                        "district_code": "current_area.district_code",
                                        "area_name": "current_area.area_name",
                                        "search_area_id": "current_area.search_area_id",
                                        "business_type": "unswept_areas.business_type"
                                    },
                                    "timeout_seconds": 120
                                },
                                "output_field": "sweep_result",
                                "description": "Call discoverer for this district"
                            }
                        }
                    }
                },
                "output_field": "sweep_results",
                "next_step": "complete",
                "description": "Loop through districts, calling discoverer for each"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["unswept_areas", "sweep_results"]
                },
                "description": "Area sweep batch complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'area-sweep-orchestrator';


-- =====================================================================
-- 2. vet-batch-processor: add continue_on_error to loop
-- =====================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_batch,config,continue_on_error}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';


-- =====================================================================
-- 3. vet-pipeline-orchestrator: thin coordinator
--
-- Vet-specific values are:
--   - Passed to children via input_mapping from input_data
--   - Set as static config on local actions (promote_candidates)
--
-- CLI must provide: country, business_type, vertical_slug
-- Optional: limit (default 50), verify_limit, promote_limit, area_code
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 7200,
    "workflow": {
        "start_step": "spawn_sweep_orch",
        "steps": {

            "spawn_sweep_orch": {
                "action": "spawn_agent",
                "config": {
                    "role": "sweep-coordinator",
                    "agent_type": "area-sweep-orchestrator"
                },
                "output_field": "sweep_orch",
                "next_step": "run_sweeps",
                "description": "Spawn the area sweep orchestrator"
            },

            "run_sweeps": {
                "action": "call_agent",
                "config": {
                    "agent_type": "area-sweep-orchestrator",
                    "target_role": "sweep-coordinator",
                    "input_mapping": {
                        "limit": "input_data.limit",
                        "country": "input_data.country",
                        "business_type": "input_data.business_type",
                        "area_code?": "input_data.area_code"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "sweep_result",
                "next_step": "promote_candidates",
                "description": "Call sweep orchestrator and wait for completion"
            },

            "promote_candidates": {
                "action": "promote_candidates",
                "config": {
                    "vertical_slug": "veterinary",
                    "business_type": "veterinary_practice",
                    "country": "GB",
                    "promote_limit": 500
                },
                "output_field": "promotion_result",
                "next_step": "spawn_batch_processor",
                "description": "Promote high-confidence candidates to businesses"
            },

            "spawn_batch_processor": {
                "action": "spawn_agent",
                "config": {
                    "role": "batch-verifier",
                    "agent_type": "vet-batch-processor"
                },
                "output_field": "batch_orch",
                "next_step": "run_verification",
                "description": "Spawn the batch verification orchestrator"
            },

            "run_verification": {
                "action": "call_agent",
                "config": {
                    "agent_type": "vet-batch-processor",
                    "target_role": "batch-verifier",
                    "input_mapping": {
                        "batch_size": "input_data.verify_limit",
                        "vertical_slug": "input_data.vertical_slug"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "verification_result",
                "next_step": "complete",
                "description": "Call batch processor and wait for completion"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["sweep_result", "promotion_result", "verification_result"]
                },
                "description": "Pipeline run complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';


--

-- vet_pipeline_updates.sql
--
-- Vertical-agnostic pipeline definitions.
--
-- Vet-specific values (country, business_type, vertical_slug) are defined
-- in the workflow configs, not hardcoded in Go. A seaweed pipeline would
-- have different agent definitions with different values, same Go code.
--
-- Values flow: pipeline input_data → call_agent input_mapping → child input_data
-- Local actions (promote_candidates) read from their step config directly.
--
-- CLI invocation must include: country, business_type, vertical_slug
-- e.g. {"limit": 50, "country": "GB", "business_type": "veterinary practice",
--        "vertical_slug": "veterinary", "verify_limit": 100, "promote_limit": 500}


-- =====================================================================
-- Fix: query_template town → postcode in vet-practice-verifier
-- (idempotent — only updates if still using town)
-- =====================================================================
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


-- =====================================================================
-- 1. area-sweep-orchestrator: dispatch → spawn+loop
--
-- load_areas now requires country and business_type from input_data
-- (passed down by the pipeline's call_agent input_mapping).
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 3600,
    "workflow": {
        "start_step": "load_areas",
        "steps": {

            "load_areas": {
                "action": "load_unswept_areas",
                "config": {
                    "input_fields": ["limit", "country", "business_type", "area_code"]
                },
                "output_field": "unswept_areas",
                "next_step": "check_areas",
                "description": "Load un-swept postcode districts from DB"
            },

            "check_areas": {
                "action": "conditional",
                "config": {
                    "condition": "unswept_areas.count > 0",
                    "then_step": "spawn_discoverer",
                    "else_step": "complete"
                },
                "description": "Skip if no areas to sweep"
            },

            "spawn_discoverer": {
                "action": "spawn_agent",
                "config": {
                    "role": "discoverer",
                    "agent_type": "area-sweep-discoverer"
                },
                "output_field": "discoverer_agent",
                "next_step": "sweep_loop",
                "description": "Spawn a single area sweep discoverer"
            },

            "sweep_loop": {
                "action": "loop",
                "config": {
                    "items_field": "unswept_areas.areas",
                    "item_variable": "current_area",
                    "max_iterations": 200,
                    "continue_on_error": true,
                    "sub_workflow": {
                        "start_step": "call_discoverer",
                        "steps": {
                            "call_discoverer": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "area-sweep-discoverer",
                                    "target_role": "discoverer",
                                    "input_mapping": {
                                        "district_code": "current_area.district_code",
                                        "area_name": "current_area.area_name",
                                        "search_area_id": "current_area.search_area_id",
                                        "business_type": "unswept_areas.business_type"
                                    },
                                    "timeout_seconds": 120
                                },
                                "output_field": "sweep_result",
                                "description": "Call discoverer for this district"
                            }
                        }
                    }
                },
                "output_field": "sweep_results",
                "next_step": "complete",
                "description": "Loop through districts, calling discoverer for each"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["unswept_areas", "sweep_results"]
                },
                "description": "Area sweep batch complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'area-sweep-orchestrator';


-- =====================================================================
-- 2. vet-batch-processor: add continue_on_error to loop
-- =====================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_batch,config,continue_on_error}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';


-- =====================================================================
-- 3. vet-pipeline-orchestrator: thin coordinator
--
-- Vet-specific values are:
--   - Passed to children via input_mapping from input_data
--   - Set as static config on local actions (promote_candidates)
--
-- CLI must provide: country, business_type, vertical_slug
-- Optional: limit (default 50), verify_limit, promote_limit, area_code
-- =====================================================================
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "orchestrator",
    "timeout_seconds": 7200,
    "workflow": {
        "start_step": "spawn_sweep_orch",
        "steps": {

            "spawn_sweep_orch": {
                "action": "spawn_agent",
                "config": {
                    "role": "sweep-coordinator",
                    "agent_type": "area-sweep-orchestrator"
                },
                "output_field": "sweep_orch",
                "next_step": "run_sweeps",
                "description": "Spawn the area sweep orchestrator"
            },

            "run_sweeps": {
                "action": "call_agent",
                "config": {
                    "agent_type": "area-sweep-orchestrator",
                    "target_role": "sweep-coordinator",
                    "input_mapping": {
                        "limit": "input_data.limit",
                        "country": "input_data.country",
                        "business_type": "input_data.business_type",
                        "area_code?": "input_data.area_code"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "sweep_result",
                "next_step": "promote_candidates",
                "description": "Call sweep orchestrator and wait for completion"
            },

            "promote_candidates": {
                "action": "promote_candidates",
                "config": {
                    "vertical_slug": "veterinary",
                    "business_type": "veterinary_practice",
                    "country": "GB",
                    "promote_limit": 500
                },
                "output_field": "promotion_result",
                "next_step": "ensure_tasks",
                "description": "Promote high-confidence candidates to businesses"
            },

            "ensure_tasks": {
                "action": "ensure_collection_tasks",
                "config": {
                    "vertical_slug": "veterinary",
                    "task_type": "initial_verification",
                    "task_priority": 5
                },
                "output_field": "ensure_result",
                "next_step": "spawn_batch_processor",
                "description": "Backfill collection_tasks for any pending businesses missing them"
            },

            "spawn_batch_processor": {
                "action": "spawn_agent",
                "config": {
                    "role": "batch-verifier",
                    "agent_type": "vet-batch-processor"
                },
                "output_field": "batch_orch",
                "next_step": "run_verification",
                "description": "Spawn the batch verification orchestrator"
            },

            "run_verification": {
                "action": "call_agent",
                "config": {
                    "agent_type": "vet-batch-processor",
                    "target_role": "batch-verifier",
                    "input_mapping": {
                        "batch_size": "input_data.verify_limit",
                        "vertical_slug": "input_data.vertical_slug"
                    },
                    "timeout_seconds": 3600
                },
                "output_field": "verification_result",
                "next_step": "complete",
                "description": "Call batch processor and wait for completion"
            },

            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["sweep_result", "promotion_result", "ensure_result", "verification_result"]
                },
                "description": "Pipeline run complete"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';


---


increase timeouts

-- 1. Increase timeout for sweep step (12 hours)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_sweeps,config,timeout_seconds}',
        '43200'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';

-- 2. Also increase for verification step (6 hours)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_verification,config,timeout_seconds}',
        '21600'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-pipeline-orchestrator';