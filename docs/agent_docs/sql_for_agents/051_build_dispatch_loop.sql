-- build-dispatch-loop agent definition
-- One-item-at-a-time dispatcher. No loops, no sub_workflows.
--
-- Each invocation:
--   1. Loads next pending item (highest priority, deps satisfied)
--   2. Claims it
--   3. Spawns + calls the handler agent
--   4. Marks the item complete
--   5. Checks for remaining items
--   6. If remaining → spawns a fresh dispatch loop (separate orchestration, clean logs)
--   7. Completes
--
-- Triggered by: seed_build_queue or external scheduler
-- Input: input_data.site_id (required), input_data.domain (optional)
--
-- NOTE: requires backward-compatible patch to LoadWorkItemsAction
--       to add 'first_item' convenience field to output.

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-dispatch-loop',
             'Build Dispatch Loop',
             'Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows — each dispatch is a separate orchestration with clean logs.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "load_next_item",
                     "steps": {

                         "load_next_item": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_domain": "build",
                                 "max_items": 1
                             },
                             "next_step": "check_has_item",
                             "description": "Load highest-priority pending build work item",
                             "output_field": "pending"
                         },

                         "check_has_item": {
                             "action": "conditional",
                             "config": {
                                 "condition": "pending.has_items == true",
                                 "then_step": "claim_item",
                                 "else_step": "complete_empty"
                             },
                             "description": "Check if there is a work item to process"
                         },

                         "claim_item": {
                             "action": "claim_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id"
                             },
                             "next_step": "check_claimed",
                             "description": "Atomically claim the work item",
                             "output_field": "claim_result"
                         },

                         "check_claimed": {
                             "action": "conditional",
                             "config": {
                                 "condition": "claim_result.claimed == true",
                                 "then_step": "spawn_handler",
                                 "else_step": "complete_empty"
                             },
                             "description": "Verify claim succeeded (item may have been claimed by another instance)"
                         },

                         "spawn_handler": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "handler",
                                 "agent_type_field": "pending.first_item.handler_agent"
                             },
                             "next_step": "call_handler",
                             "description": "Spawn the handler agent for this work item",
                             "output_field": "handler_spawned"
                         },

                         "call_handler": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type_field": "pending.first_item.handler_agent",
                                 "target_role": "handler",
                                 "input_mapping": {
                                     "site_id":      "pending.first_item.site_id",
                                     "domain":       "pending.first_item.domain",
                                     "work_item_id": "pending.first_item.id",
                                     "item_type":    "pending.first_item.item_type",
                                     "spec":         "pending.first_item.spec"
                                 },
                                 "timeout_seconds": 300
                             },
                             "next_step": "mark_complete",
                             "description": "Call handler — it does its work and returns",
                             "output_field": "handler_result"
                         },

                         "mark_complete": {
                             "action": "complete_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id",
                                 "result": "handler_result"
                             },
                             "next_step": "check_remaining",
                             "description": "Mark work item as complete (dispatch loop owns lifecycle)",
                             "output_field": "item_completed"
                         },

                         "check_remaining": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_domain": "build",
                                 "max_items": 1
                             },
                             "next_step": "has_remaining",
                             "description": "Check if handler created follow-on items or other items are pending",
                             "output_field": "remaining"
                         },

                         "has_remaining": {
                             "action": "conditional",
                             "config": {
                                 "condition": "remaining.has_items == true",
                                 "then_step": "spawn_next_dispatch",
                                 "else_step": "complete"
                             },
                             "description": "If more items exist, spawn another dispatch invocation"
                         },

                         "spawn_next_dispatch": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "dispatcher",
                                 "agent_type": "build-dispatch-loop"
                             },
                             "next_step": "call_next_dispatch",
                             "description": "Spawn a fresh dispatch loop for remaining items",
                             "output_field": "next_dispatch_spawned"
                         },

                         "call_next_dispatch": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "build-dispatch-loop",
                                 "target_role": "dispatcher",
                                 "input_mapping": {
                                     "site_id": "input_data.site_id",
                                     "domain":  "input_data.domain"
                                 },
                                 "timeout_seconds": 600
                             },
                             "next_step": "complete",
                             "description": "Chain to next dispatch (separate orchestration, clean logs)",
                             "output_field": "next_dispatch_result"
                         },

                         "complete_empty": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": [],
                                 "success_message": "No pending work items for this site"
                             },
                             "description": "Queue empty — nothing to dispatch"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["handler_result", "item_completed", "next_dispatch_result"]
                             },
                             "description": "Dispatch invocation complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 900
             }'::jsonb,
             true,
             '["dispatch", "orchestration", "work-items"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["build", "dispatch", "orchestration"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id"], "optional": ["domain"], "description": "Receives site_id from seed_build_queue or external trigger."}'::jsonb,
             '{"produces": {"handler_result": "Result from the handler agent", "item_completed": "Work item marked complete", "next_dispatch_result": "Result from chained dispatch (if any)"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();


---

-- update include error

-- build-dispatch-loop agent definition
-- One-item-at-a-time dispatcher. No loops, no sub_workflows.
--
-- Each invocation:
--   1. Loads next pending item (highest priority, deps satisfied)
--   2. Claims it
--   3. Spawns + calls the handler agent
--   4a. On success: marks the item complete
--   4b. On failure: marks the item failed (increments attempt_count, resets to triaged if retries remain)
--   5. Checks for remaining items
--   6. If remaining → spawns a fresh dispatch loop (separate orchestration, clean logs)
--   7. Completes
--
-- Error handling: uses error_step routing so handler failures don't kill the chain.
-- The dispatch loop continues to the next item regardless of individual handler outcomes.
--
-- Triggered by: seed_build_queue or external scheduler (heartbeat)
-- Input: input_data.site_id (required), input_data.domain (optional)
--
-- NOTE: LoadWorkItemsAction already has 'first_item' convenience field.
-- NOTE: Requires coordinator.go with error_step routing support.

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-dispatch-loop',
             'Build Dispatch Loop',
             'Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows — each dispatch is a separate orchestration with clean logs.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "load_next_item",
                     "steps": {

                         "load_next_item": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_domain": "build",
                                 "max_items": 1
                             },
                             "next_step": "check_has_item",
                             "description": "Load highest-priority pending build work item",
                             "output_field": "pending"
                         },

                         "check_has_item": {
                             "action": "conditional",
                             "config": {
                                 "condition": "pending.has_items == true",
                                 "then_step": "claim_item",
                                 "else_step": "complete_empty"
                             },
                             "description": "Check if there is a work item to process"
                         },

                         "claim_item": {
                             "action": "claim_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id"
                             },
                             "next_step": "check_claimed",
                             "description": "Atomically claim the work item",
                             "output_field": "claim_result"
                         },

                         "check_claimed": {
                             "action": "conditional",
                             "config": {
                                 "condition": "claim_result.claimed == true",
                                 "then_step": "spawn_handler",
                                 "else_step": "complete_empty"
                             },
                             "description": "Verify claim succeeded (item may have been claimed by another instance)"
                         },

                         "spawn_handler": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "handler",
                                 "agent_type_field": "pending.first_item.handler_agent",
                                 "error_step": "mark_failed"
                             },
                             "next_step": "call_handler",
                             "description": "Spawn the handler agent for this work item",
                             "output_field": "handler_spawned"
                         },

                         "call_handler": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type_field": "pending.first_item.handler_agent",
                                 "target_role": "handler",
                                 "input_mapping": {
                                     "site_id":      "pending.first_item.site_id",
                                     "domain":       "pending.first_item.domain",
                                     "work_item_id": "pending.first_item.id",
                                     "item_type":    "pending.first_item.item_type",
                                     "spec":         "pending.first_item.spec"
                                 },
                                 "timeout_seconds": 300,
                                 "error_step": "mark_failed"
                             },
                             "next_step": "mark_complete",
                             "description": "Call handler — it does its work and returns",
                             "output_field": "handler_result"
                         },

                         "mark_complete": {
                             "action": "complete_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id",
                                 "result": "handler_result"
                             },
                             "next_step": "check_remaining",
                             "description": "Mark work item as complete (dispatch loop owns lifecycle)",
                             "output_field": "item_completed"
                         },

                         "mark_failed": {
                             "action": "fail_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id",
                                 "error_message": "Handler agent failed or could not be spawned"
                             },
                             "next_step": "check_remaining",
                             "description": "Mark work item as failed — increments attempt count, resets to triaged if retries remain",
                             "output_field": "item_failed"
                         },

                         "check_remaining": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_domain": "build",
                                 "max_items": 1
                             },
                             "next_step": "has_remaining",
                             "description": "Check if handler created follow-on items or other items are pending",
                             "output_field": "remaining"
                         },

                         "has_remaining": {
                             "action": "conditional",
                             "config": {
                                 "condition": "remaining.has_items == true",
                                 "then_step": "spawn_next_dispatch",
                                 "else_step": "complete"
                             },
                             "description": "If more items exist, spawn another dispatch invocation"
                         },

                         "spawn_next_dispatch": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "dispatcher",
                                 "agent_type": "build-dispatch-loop"
                             },
                             "next_step": "call_next_dispatch",
                             "description": "Spawn a fresh dispatch loop for remaining items",
                             "output_field": "next_dispatch_spawned"
                         },

                         "call_next_dispatch": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "build-dispatch-loop",
                                 "target_role": "dispatcher",
                                 "input_mapping": {
                                     "site_id": "input_data.site_id",
                                     "domain":  "input_data.domain"
                                 },
                                 "timeout_seconds": 600,
                                 "error_step": "complete"
                             },
                             "next_step": "complete",
                             "description": "Chain to next dispatch (separate orchestration, clean logs). On failure, current item was already processed — heartbeat picks up remaining.",
                             "output_field": "next_dispatch_result"
                         },

                         "complete_empty": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": [],
                                 "success_message": "No pending work items for this site"
                             },
                             "description": "Queue empty — nothing to dispatch"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["handler_result", "item_completed", "item_failed", "next_dispatch_result"]
                             },
                             "description": "Dispatch invocation complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 900
             }'::jsonb,
             true,
             '["dispatch", "orchestration", "work-items"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["build", "dispatch", "orchestration"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id"], "optional": ["domain"], "description": "Receives site_id from seed_build_queue or external trigger."}'::jsonb,
             '{"produces": {"handler_result": "Result from the handler agent", "item_completed": "Work item marked complete", "item_failed": "Work item marked failed (if error_step triggered)", "next_dispatch_result": "Result from chained dispatch (if any)"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();


---
-- path fix

-- build-dispatch-loop agent definition
-- One-item-at-a-time dispatcher. No loops, no sub_workflows.
--
-- Each invocation:
--   1. Loads next pending item (highest priority, deps satisfied)
--   2. Claims it
--   3. Spawns + calls the handler agent
--   4a. On success: marks the item complete
--   4b. On failure: marks the item failed (increments attempt_count, resets to triaged if retries remain)
--   5. Checks for remaining items
--   6. If remaining → spawns a fresh dispatch loop (separate orchestration, clean logs)
--   7. Completes
--
-- Error handling: uses error_step routing so handler failures don't kill the chain.
-- The dispatch loop continues to the next item regardless of individual handler outcomes.
--
-- Triggered by: seed_build_queue or external scheduler (heartbeat)
-- Input: input_data.site_id (required), input_data.domain (optional)
--
-- NOTE: LoadWorkItemsAction already has 'first_item' convenience field.
-- NOTE: Requires coordinator.go with error_step routing support.

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-dispatch-loop',
             'Build Dispatch Loop',
             'Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows — each dispatch is a separate orchestration with clean logs.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "load_next_item",
                     "steps": {

                         "load_next_item": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_domain": "build",
                                 "max_items": 1
                             },
                             "next_step": "check_has_item",
                             "description": "Load highest-priority pending build work item",
                             "output_field": "pending"
                         },

                         "check_has_item": {
                             "action": "conditional",
                             "config": {
                                 "condition": "pending.has_items == true",
                                 "then_step": "claim_item",
                                 "else_step": "complete_empty"
                             },
                             "description": "Check if there is a work item to process"
                         },

                         "claim_item": {
                             "action": "claim_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id"
                             },
                             "next_step": "check_claimed",
                             "description": "Atomically claim the work item",
                             "output_field": "claim_result"
                         },

                         "check_claimed": {
                             "action": "conditional",
                             "config": {
                                 "condition": "claim_result.claimed == true",
                                 "then_step": "spawn_handler",
                                 "else_step": "complete_empty"
                             },
                             "description": "Verify claim succeeded (item may have been claimed by another instance)"
                         },

                         "spawn_handler": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "handler",
                                 "agent_type_field": "pending.first_item.handler_agent",
                                 "error_step": "mark_failed"
                             },
                             "next_step": "call_handler",
                             "description": "Spawn the handler agent for this work item",
                             "output_field": "handler_spawned"
                         },

                         "call_handler": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type_field": "pending.first_item.handler_agent",
                                 "target_role": "handler",
                                 "input_mapping": {
                                     "site_id":      "pending.first_item.site_id",
                                     "domain":       "input_data.domain",
                                     "work_item_id": "pending.first_item.id",
                                     "item_type":    "pending.first_item.item_type",
                                     "spec":         "pending.first_item.spec"
                                 },
                                 "timeout_seconds": 300,
                                 "error_step": "mark_failed"
                             },
                             "next_step": "mark_complete",
                             "description": "Call handler — it does its work and returns",
                             "output_field": "handler_result"
                         },

                         "mark_complete": {
                             "action": "complete_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id",
                                 "result": "handler_result"
                             },
                             "next_step": "check_remaining",
                             "description": "Mark work item as complete (dispatch loop owns lifecycle)",
                             "output_field": "item_completed"
                         },

                         "mark_failed": {
                             "action": "fail_work_item",
                             "config": {
                                 "work_item_id": "pending.first_item.id",
                                 "error_message": "Handler agent failed or could not be spawned"
                             },
                             "next_step": "check_remaining",
                             "description": "Mark work item as failed — increments attempt count, resets to triaged if retries remain",
                             "output_field": "item_failed"
                         },

                         "check_remaining": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_domain": "build",
                                 "max_items": 1
                             },
                             "next_step": "has_remaining",
                             "description": "Check if handler created follow-on items or other items are pending",
                             "output_field": "remaining"
                         },

                         "has_remaining": {
                             "action": "conditional",
                             "config": {
                                 "condition": "remaining.has_items == true",
                                 "then_step": "spawn_next_dispatch",
                                 "else_step": "complete"
                             },
                             "description": "If more items exist, spawn another dispatch invocation"
                         },

                         "spawn_next_dispatch": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "dispatcher",
                                 "agent_type": "build-dispatch-loop"
                             },
                             "next_step": "call_next_dispatch",
                             "description": "Spawn a fresh dispatch loop for remaining items",
                             "output_field": "next_dispatch_spawned"
                         },

                         "call_next_dispatch": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "build-dispatch-loop",
                                 "target_role": "dispatcher",
                                 "input_mapping": {
                                     "site_id": "input_data.site_id",
                                     "domain":  "input_data.domain"
                                 },
                                 "timeout_seconds": 600,
                                 "error_step": "complete"
                             },
                             "next_step": "complete",
                             "description": "Chain to next dispatch (separate orchestration, clean logs). On failure, current item was already processed — heartbeat picks up remaining.",
                             "output_field": "next_dispatch_result"
                         },

                         "complete_empty": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": [],
                                 "success_message": "No pending work items for this site"
                             },
                             "description": "Queue empty — nothing to dispatch"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["handler_result", "item_completed", "item_failed", "next_dispatch_result"]
                             },
                             "description": "Dispatch invocation complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 900
             }'::jsonb,
             true,
             '["dispatch", "orchestration", "work-items"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["build", "dispatch", "orchestration"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id"], "optional": ["domain"], "description": "Receives site_id from seed_build_queue or external trigger."}'::jsonb,
             '{"produces": {"handler_result": "Result from the handler agent", "item_completed": "Work item marked complete", "item_failed": "Work item marked failed (if error_step triggered)", "next_dispatch_result": "Result from chained dispatch (if any)"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();


---
-- page content writer expects current_page data path not spec - add both

-- Add current_page mapping for page-content-writer compatibility
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_handler,config,input_mapping,current_page}',
        '"pending.first_item.spec"'
                     ),
    updated_at = now()
WHERE type = 'build-dispatch-loop';

-- Verify
SELECT default_config->'workflow'->'steps'->'call_handler'->'config'->'input_mapping'
FROM agent_definitions
WHERE type = 'build-dispatch-loop';


-- path fixes

-- 056_build_pipeline_fixes.sql
--
-- Fixes for the automated build pipeline post-mortem (gaswholesalers.com):
--
-- Problem 1: page-content-writer returns HTML but doesn't save to page_components
--   Fix: New page-build-handler agent wraps writer with save_sections (055_page_build_handler.sql)
--   Fix: WriteBuildItemsAction Go patch changes handler_agent to "page-build-handler"
--
-- Problem 2: needs_design had domain "design", dispatch loop filters domain "build"
--   Fix: WriteBuildItemsAction Go patch changes domain to "build"
--   Fix: Below — retroactively fix any existing needs_design items
--
-- Problem 3: No rerender/deploy step in the pipeline
--   Fix: WriteBuildItemsAction Go patch adds needs_rerender work item
--   Fix: Below — dispatch loop input_mapping passes refresh_site_components
--
-- Problem 4: Empty CSS variables (no :root block)
--   Fix: webdesign-agent will now be dispatched (needs_design domain fixed)
--   Fix: rerender-pages refreshes site components including CSS head

-- ============================================================================
-- Fix 1: Fix dispatch loop input_mapping
--   - domain: use input_data.domain (site domain like gaswholesalers.com)
--     NOT pending.first_item.domain (work item namespace like "build")
--   - refresh_site_components: passthrough for rerender-pages
--   - current_page: alias for page-content-writer compatibility
-- ============================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_handler,config,input_mapping}',
        '{
            "site_id":      "pending.first_item.site_id",
            "domain":       "input_data.domain",
            "work_item_id": "pending.first_item.id",
            "item_type":    "pending.first_item.item_type",
            "spec":         "pending.first_item.spec",
            "current_page": "pending.first_item.spec",
            "refresh_site_components": "pending.first_item.spec.refresh_site_components"
        }'::jsonb
                     ),
    updated_at = now()
WHERE type = 'build-dispatch-loop';

-- ============================================================================
-- Fix 2: Retroactively fix any existing needs_design items with wrong domain
-- (for sites that may have been built before the Go patch is deployed)
-- ============================================================================
UPDATE site_work_items
SET domain = 'build'
WHERE item_type = 'needs_design'
  AND domain = 'design'
  AND status IN ('detected', 'triaged', 'approved');

-- ============================================================================
-- Fix 3: Retroactively update handler_agent for pending content page items
-- (for sites that may have been built before the Go patch is deployed)
-- ============================================================================
UPDATE site_work_items
SET handler_agent = 'page-build-handler'
WHERE item_type = 'needs_content_page'
  AND handler_agent = 'page-content-writer'
  AND status IN ('detected', 'triaged', 'approved');

-- ============================================================================
-- Verify
-- ============================================================================
-- SELECT type,
--        default_config->'workflow'->'steps'->'call_handler'->'config'->'input_mapping' as mapping
-- FROM agent_definitions
-- WHERE type = 'build-dispatch-loop';