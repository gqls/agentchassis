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

---

-- fix path - make refresh_site_components optional

-- Fix: make refresh_site_components optional in dispatch loop input_mapping
-- Run against the deployed agent definition

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_handler,config,input_mapping}',
        '{
            "current_page": "pending.first_item.spec",
            "domain": "input_data.domain",
            "item_type": "pending.first_item.item_type",
            "refresh_site_components?": "pending.first_item.spec.refresh_site_components",
            "site_id": "pending.first_item.site_id",
            "spec": "pending.first_item.spec",
            "work_item_id": "pending.first_item.id"
        }'::jsonb
                     )
WHERE type = 'build-dispatch-loop';

-- backup

clients_db=# SELECT * FROM agent_definitions WHERE type = 'build-dispatch-loop';
id                  |        type         |    display_name     |                                                                            description                                                                            |   category   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  default_config                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | is_active |          created_at           |          updated_at           | deleted_at |                capabilities                 |       image_repository       | image_tag | command |                                           resources                                            |                                                       topics                                                       |                                            health_config                                            | env_vars | version | previous_version_id | task_workflow | orchestrator_workflow | orchestration_workflow |                delegation_preferences                 | agent_category |    status    |              domain_tags               | briefing_questionnaire | usage_count | is_snapshot |                                                         input_contract                                                          |                                                                                                                    output_contract
--------------------------------------+---------------------+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+--------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-----------+-------------------------------+-------------------------------+------------+---------------------------------------------+------------------------------+-----------+---------+------------------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------------+-----------------------------------------------------------------------------------------------------+----------+---------+---------------------+---------------+-----------------------+------------------------+-------------------------------------------------------+----------------+--------------+----------------------------------------+------------------------+-------------+-------------+---------------------------------------------------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 099b51e0-6dd0-4856-8f82-805a379e8b1d | build-dispatch-loop | Build Dispatch Loop | Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows — each dispatch is a separate orchestration with clean logs. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["handler_result", "item_completed", "item_failed", "next_dispatch_result"]}, "description": "Dispatch invocation complete"}, "claim_item": {"action": "claim_work_item", "config": {"work_item_id": "pending.first_item.id"}, "next_step": "check_claimed", "description": "Atomically claim the work item", "output_field": "claim_result"}, "mark_failed": {"action": "fail_work_item", "config": {"work_item_id": "pending.first_item.id", "error_message": "Handler agent failed or could not be spawned"}, "next_step": "check_remaining", "description": "Mark work item as failed — increments attempt count, resets to triaged if retries remain", "output_field": "item_failed"}, "call_handler": {"action": "call_agent", "config": {"error_step": "mark_failed", "target_role": "handler", "input_mapping": {"spec": "pending.first_item.spec", "domain": "input_data.domain", "site_id": "pending.first_item.site_id", "item_type": "pending.first_item.item_type", "current_page": "pending.first_item.spec", "work_item_id": "pending.first_item.id", "refresh_site_components?": "pending.first_item.spec.refresh_site_components"}, "timeout_seconds": 300, "agent_type_field": "pending.first_item.handler_agent"}, "next_step": "mark_complete", "description": "Call handler — it does its work and returns", "output_field": "handler_result"}, "check_claimed": {"action": "conditional", "config": {"condition": "claim_result.claimed == true", "else_step": "complete_empty", "then_step": "spawn_handler"}, "description": "Verify claim succeeded (item may have been claimed by another instance)"}, "has_remaining": {"action": "conditional", "config": {"condition": "remaining.has_items == true", "else_step": "complete", "then_step": "spawn_next_dispatch"}, "description": "If more items exist, spawn another dispatch invocation"}, "mark_complete": {"action": "complete_work_item", "config": {"result": "handler_result", "work_item_id": "pending.first_item.id"}, "next_step": "check_remaining", "description": "Mark work item as complete (dispatch loop owns lifecycle)", "output_field": "item_completed"}, "spawn_handler": {"action": "spawn_agent", "config": {"role": "handler", "error_step": "mark_failed", "agent_type_field": "pending.first_item.handler_agent"}, "next_step": "call_handler", "description": "Spawn the handler agent for this work item", "output_field": "handler_spawned"}, "check_has_item": {"action": "conditional", "config": {"condition": "pending.has_items == true", "else_step": "complete_empty", "then_step": "claim_item"}, "description": "Check if there is a work item to process"}, "complete_empty": {"action": "complete_workflow", "config": {"output_fields": [], "success_message": "No pending work items for this site"}, "description": "Queue empty — nothing to dispatch"}, "load_next_item": {"action": "load_work_items", "config": {"site_id": "input_data.site_id", "max_items": 1, "item_domain": "build"}, "next_step": "check_has_item", "description": "Load highest-priority pending build work item", "output_field": "pending"}, "check_remaining": {"action": "load_work_items", "config": {"site_id": "input_data.site_id", "max_items": 1, "item_domain": "build"}, "next_step": "has_remaining", "description": "Check if handler created follow-on items or other items are pending", "output_field": "remaining"}, "call_next_dispatch": {"action": "call_agent", "config": {"agent_type": "build-dispatch-loop", "error_step": "complete", "target_role": "dispatcher", "input_mapping": {"domain": "input_data.domain", "site_id": "input_data.site_id"}, "timeout_seconds": 600}, "next_step": "complete", "description": "Chain to next dispatch (separate orchestration, clean logs). On failure, current item was already processed — heartbeat picks up remaining.", "output_field": "next_dispatch_result"}, "spawn_next_dispatch": {"action": "spawn_agent", "config": {"role": "dispatcher", "agent_type": "build-dispatch-loop"}, "next_step": "call_next_dispatch", "description": "Spawn a fresh dispatch loop for remaining items", "output_field": "next_dispatch_spawned"}}, "start_step": "load_next_item"}, "processing_mode": "orchestrator", "timeout_seconds": 900} | t         | 2026-02-24 16:56:54.875139+00 | 2026-02-28 19:04:08.894085+00 |            | ["dispatch", "orchestration", "work-items"] | docker.io/aqls/agent-chassis | v1.0.817  |         | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} | coordinator    | experimental | ["build", "dispatch", "orchestration"] | {}                     |           0 | f           | {"optional": ["domain"], "required": ["site_id"], "description": "Receives site_id from seed_build_queue or external trigger."} | {"produces": {"item_failed": "Work item marked failed (if error_step triggered)", "handler_result": "Result from the handler agent", "item_completed": "Work item marked complete", "next_dispatch_result": "Result from chained dispatch (if any)"}}
(1 row)


-- 063_dispatch_loop_remove_self_chaining.sql
--
-- Problem: build-dispatch-loop spawns a copy of itself to process the
-- next work item. The child's response frequently gets lost (Kafka topic
-- retention expiry, pod restarts), leaving the parent stuck at
-- spawn_next_dispatch in AWAITING_RESPONSES forever.
--
-- Fix: after processing one item, loop back to load_next_item within
-- the same orchestration. No inter-orchestration messaging, no lost responses.
--
-- Changes:
--   - has_remaining.then_step: spawn_next_dispatch → load_next_item
--   - mark_failed.next_step: check_remaining → load_next_item
--     (skip the separate check_remaining load, just reload from the top)
--   - Remove steps: spawn_next_dispatch, call_next_dispatch
--   - Remove output_field: next_dispatch_result from complete step
--
-- The workflow now processes ALL pending items for a site in one
-- orchestration run, then completes. Much simpler failure mode.

UPDATE agent_definitions
SET default_config = $cfg${
    "workflow": {
        "start_step": "load_next_item",
        "steps": {
            "load_next_item": {
                "action": "load_work_items",
                "config": {
                    "site_id": "input_data.site_id",
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
                    "else_step": "complete"
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
                    "else_step": "complete"
                },
                "description": "Verify claim succeeded (item may have been claimed by another instance)"
            },
            "spawn_handler": {
                "action": "spawn_agent",
                "config": {
                    "role": "handler",
                    "error_step": "mark_failed",
                    "agent_type_field": "pending.first_item.handler_agent"
                },
                "next_step": "call_handler",
                "description": "Spawn the handler agent for this work item",
                "output_field": "handler_spawned"
            },
            "call_handler": {
                "action": "call_agent",
                "config": {
                    "target_role": "handler",
                    "error_step": "mark_failed",
                    "input_mapping": {
                        "site_id": "pending.first_item.site_id",
                        "domain": "input_data.domain",
                        "item_type": "pending.first_item.item_type",
                        "work_item_id": "pending.first_item.id",
                        "spec": "pending.first_item.spec",
                        "current_page": "pending.first_item.spec",
                        "refresh_site_components?": "pending.first_item.spec.refresh_site_components"
                    },
                    "timeout_seconds": 300,
                    "agent_type_field": "pending.first_item.handler_agent"
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
                "next_step": "load_next_item",
                "description": "Mark work item complete, then loop back for next item",
                "output_field": "item_completed"
            },
            "mark_failed": {
                "action": "fail_work_item",
                "config": {
                    "work_item_id": "pending.first_item.id",
                    "error_message": "Handler agent failed or could not be spawned"
                },
                "next_step": "load_next_item",
                "description": "Mark work item failed, then loop back for next item",
                "output_field": "item_failed"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["handler_result", "item_completed", "item_failed"],
                    "success_message": "All pending work items processed"
                },
                "description": "Queue empty — all items processed or no items found"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}$cfg$::jsonb,
    updated_at = NOW()
WHERE type = 'build-dispatch-loop';


-- Verify
SELECT type,
       default_config->'workflow'->'start_step' as start_step,
       jsonb_object_keys(default_config->'workflow'->'steps') as steps
FROM agent_definitions
WHERE type = 'build-dispatch-loop';

----

-- updated to use loop in workflow

-- Migration 071: Switch build-dispatch-loop from step-chaining to loop action
--
-- Problem: Step-chaining (mark_complete → load_next_item) processes only one
-- work item per trigger. The loop action is proven in maintenance-triage and
-- pageflow-builder.
--
-- Change: Load all dispatchable items upfront, iterate with sub_workflow.
-- Items with unmet dependencies are excluded by load_work_items' query.
-- Newly unblocked items are picked up by the next scheduler trigger.
--
-- Variable name change:
--   OLD: pending.first_item.*
--   NEW: current_item.*  (set by loop's item_variable)
--
-- Outer workflow: 4 steps (load → check → loop → complete)
-- Sub_workflow:   7 steps (claim → check → spawn → call → mark_complete/mark_failed → done)

BEGIN;

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "start_step": "load_items",
    "steps": {
      "load_items": {
        "action": "load_work_items",
        "config": {
          "site_id": "input_data.site_id",
          "max_items": 50
        },
        "next_step": "check_has_items",
        "output_field": "pending",
        "description": "Load all dispatchable items (dependency-filtered, priority-ordered)"
      },
      "check_has_items": {
        "action": "conditional",
        "config": {
          "condition": "pending.has_items == true",
          "then_step": "process_items",
          "else_step": "complete"
        },
        "description": "Any items to process?"
      },
      "process_items": {
        "action": "loop",
        "config": {
          "items_field": "pending.items",
          "item_variable": "current_item",
          "max_iterations": 50,
          "continue_on_error": true,
          "sub_workflow": {
            "start_step": "claim",
            "steps": {
              "claim": {
                "action": "claim_work_item",
                "config": { "work_item_id": "current_item.id" },
                "next_step": "check_claim",
                "output_field": "claim_result",
                "description": "Atomically claim item"
              },
              "check_claim": {
                "action": "conditional",
                "config": {
                  "condition": "claim_result.claimed == true",
                  "then_step": "spawn_handler",
                  "else_step": "done"
                },
                "description": "Skip if already claimed by another instance"
              },
              "spawn_handler": {
                "action": "spawn_agent",
                "config": {
                  "role": "handler",
                  "agent_type_field": "current_item.handler_agent",
                  "error_step": "mark_failed"
                },
                "next_step": "call_handler",
                "output_field": "handler_spawned",
                "description": "Spawn handler (dynamic type per item)"
              },
              "call_handler": {
                "action": "call_agent",
                "config": {
                  "target_role": "handler",
                  "error_step": "mark_failed",
                  "timeout_seconds": 300,
                  "input_mapping": {
                    "site_id": "current_item.site_id",
                    "domain": "input_data.domain",
                    "work_item_id": "current_item.id",
                    "item_type": "current_item.item_type",
                    "spec": "current_item.spec",
                    "current_page": "current_item.spec",
                    "refresh_site_components?": "current_item.spec.refresh_site_components"
                  }
                },
                "next_step": "mark_complete",
                "output_field": "handler_result",
                "description": "Call handler agent"
              },
              "mark_complete": {
                "action": "complete_work_item",
                "config": {
                  "work_item_id": "current_item.id",
                  "result": "handler_result"
                },
                "next_step": "done",
                "output_field": "item_completed",
                "description": "Mark item complete"
              },
              "mark_failed": {
                "action": "fail_work_item",
                "config": {
                  "work_item_id": "current_item.id",
                  "error_message": "Handler failed"
                },
                "next_step": "done",
                "output_field": "item_failed",
                "description": "Mark item failed, continue to next"
              },
              "done": {
                "action": "loop_complete",
                "description": "Item done"
              }
            }
          }
        },
        "next_step": "complete",
        "output_field": "items_processed",
        "description": "Process each item: claim → spawn → call → mark"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": ["pending", "items_processed"]
        },
        "description": "Dispatch complete"
      }
    }
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 1800
}'::jsonb,
updated_at = NOW()
WHERE type = 'build-dispatch-loop';

-- Verify
SELECT type,
       default_config #>> '{workflow,start_step}' AS start,
    default_config #>> '{workflow,steps,process_items,action}' AS loop_action,
    default_config #>> '{workflow,steps,process_items,config,item_variable}' AS item_var
FROM agent_definitions
WHERE type = 'build-dispatch-loop';

COMMIT;


---

-- maybe same loop fix

-- Migration 071: Switch build-dispatch-loop from step-chaining to loop action
--
-- Problem: Step-chaining (mark_complete → load_next_item) processes only one
-- work item per trigger. The loop action is proven in maintenance-triage and
-- pageflow-builder.
--
-- Change: Load all dispatchable items upfront, iterate with sub_workflow.
-- Items with unmet dependencies are excluded by load_work_items' query.
-- Newly unblocked items are picked up by the next scheduler trigger.
--
-- Variable name change:
--   OLD: pending.first_item.*
--   NEW: current_item.*  (set by loop's item_variable)
--
-- Also adds optional input_mapping fields for section-editor compatibility:
--   edit_type?, page_name?, slot_name?, field_updates?, etc.
-- These are ? suffixed so they're silently skipped for non-section-editor handlers.
--
-- Outer workflow: 4 steps (load → check → loop → complete)
-- Sub_workflow:   7 steps (claim → check → spawn → call → mark_complete/mark_failed → done)

BEGIN;

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "start_step": "load_items",
    "steps": {
      "load_items": {
        "action": "load_work_items",
        "config": {
          "site_id": "input_data.site_id",
          "max_items": 50
        },
        "next_step": "check_has_items",
        "output_field": "pending",
        "description": "Load all dispatchable items (dependency-filtered, priority-ordered)"
      },
      "check_has_items": {
        "action": "conditional",
        "config": {
          "condition": "pending.has_items == true",
          "then_step": "process_items",
          "else_step": "complete"
        },
        "description": "Any items to process?"
      },
      "process_items": {
        "action": "loop",
        "config": {
          "items_field": "pending.items",
          "item_variable": "current_item",
          "max_iterations": 50,
          "continue_on_error": true,
          "sub_workflow": {
            "start_step": "claim",
            "steps": {
              "claim": {
                "action": "claim_work_item",
                "config": { "work_item_id": "current_item.id" },
                "next_step": "check_claim",
                "output_field": "claim_result",
                "description": "Atomically claim item"
              },
              "check_claim": {
                "action": "conditional",
                "config": {
                  "condition": "claim_result.claimed == true",
                  "then_step": "spawn_handler",
                  "else_step": "done"
                },
                "description": "Skip if already claimed by another instance"
              },
              "spawn_handler": {
                "action": "spawn_agent",
                "config": {
                  "role": "handler",
                  "agent_type_field": "current_item.handler_agent",
                  "error_step": "mark_failed"
                },
                "next_step": "call_handler",
                "output_field": "handler_spawned",
                "description": "Spawn handler (dynamic type per item)"
              },
              "call_handler": {
                "action": "call_agent",
                "config": {
                  "target_role": "handler",
                  "error_step": "mark_failed",
                  "timeout_seconds": 300,
                  "input_mapping": {
                    "site_id": "current_item.site_id",
                    "domain": "input_data.domain",
                    "work_item_id": "current_item.id",
                    "item_type": "current_item.item_type",
                    "spec": "current_item.spec",
                    "current_page": "current_item.spec",
                    "refresh_site_components?": "current_item.spec.refresh_site_components",
                    "edit_type?": "current_item.spec.edit_type",
                    "page_name?": "current_item.spec.page_name",
                    "slot_name?": "current_item.spec.slot_name",
                    "field_updates?": "current_item.spec.field_updates",
                    "replacement_content_data?": "current_item.spec.replacement_content_data",
                    "new_component_function?": "current_item.spec.new_component_function"
                  }
                },
                "next_step": "mark_complete",
                "output_field": "handler_result",
                "description": "Call handler agent"
              },
              "mark_complete": {
                "action": "complete_work_item",
                "config": {
                  "work_item_id": "current_item.id",
                  "result": "handler_result"
                },
                "next_step": "done",
                "output_field": "item_completed",
                "description": "Mark item complete"
              },
              "mark_failed": {
                "action": "fail_work_item",
                "config": {
                  "work_item_id": "current_item.id",
                  "error_message": "Handler failed"
                },
                "next_step": "done",
                "output_field": "item_failed",
                "description": "Mark item failed, continue to next"
              },
              "done": {
                "action": "loop_complete",
                "description": "Item done"
              }
            }
          }
        },
        "next_step": "complete",
        "output_field": "items_processed",
        "description": "Process each item: claim → spawn → call → mark"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": ["pending", "items_processed"]
        },
        "description": "Dispatch complete"
      }
    }
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 1800
}'::jsonb,
updated_at = NOW()
WHERE type = 'build-dispatch-loop';

-- Verify
SELECT type,
       default_config #>> '{workflow,start_step}' AS start,
    default_config #>> '{workflow,steps,process_items,action}' AS loop_action,
    default_config #>> '{workflow,steps,process_items,config,item_variable}' AS item_var
FROM agent_definitions
WHERE type = 'build-dispatch-loop';

COMMIT;

---

-- image handling now deploys

-- Migration 071: Switch build-dispatch-loop from step-chaining to loop action
--
-- Problem: Step-chaining (mark_complete → load_next_item) processes only one
-- work item per trigger. The loop action is proven in maintenance-triage and
-- pageflow-builder.
--
-- Change: Load all dispatchable items upfront, iterate with sub_workflow.
-- Items with unmet dependencies are excluded by load_work_items' query.
-- Newly unblocked items are picked up by the next scheduler trigger.
--
-- Variable name change:
--   OLD: pending.first_item.*
--   NEW: current_item.*  (set by loop's item_variable)
--
-- Also adds optional input_mapping fields for section-editor compatibility:
--   edit_type?, page_name?, slot_name?, field_updates?, etc.
-- These are ? suffixed so they're silently skipped for non-section-editor handlers.
--
-- Outer workflow: 4 steps (load → check → loop → complete)
-- Sub_workflow:   7 steps (claim → check → spawn → call → mark_complete/mark_failed → done)

BEGIN;

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "start_step": "load_items",
    "steps": {
      "load_items": {
        "action": "load_work_items",
        "config": {
          "site_id": "input_data.site_id",
          "max_items": 50
        },
        "next_step": "check_has_items",
        "output_field": "pending",
        "description": "Load all dispatchable items (dependency-filtered, priority-ordered)"
      },
      "check_has_items": {
        "action": "conditional",
        "config": {
          "condition": "pending.has_items == true",
          "then_step": "process_items",
          "else_step": "complete"
        },
        "description": "Any items to process?"
      },
      "process_items": {
        "action": "loop",
        "config": {
          "items_field": "pending.items",
          "item_variable": "current_item",
          "max_iterations": 50,
          "continue_on_error": true,
          "sub_workflow": {
            "start_step": "claim",
            "steps": {
              "claim": {
                "action": "claim_work_item",
                "config": { "work_item_id": "current_item.id" },
                "next_step": "check_claim",
                "output_field": "claim_result",
                "description": "Atomically claim item"
              },
              "check_claim": {
                "action": "conditional",
                "config": {
                  "condition": "claim_result.claimed == true",
                  "then_step": "spawn_handler",
                  "else_step": "done"
                },
                "description": "Skip if already claimed by another instance"
              },
              "spawn_handler": {
                "action": "spawn_agent",
                "config": {
                  "role": "handler",
                  "agent_type_field": "current_item.handler_agent",
                  "error_step": "mark_failed"
                },
                "next_step": "call_handler",
                "output_field": "handler_spawned",
                "description": "Spawn handler (dynamic type per item)"
              },
              "call_handler": {
                "action": "call_agent",
                "config": {
                  "target_role": "handler",
                  "error_step": "mark_failed",
                  "timeout_seconds": 300,
                  "input_mapping": {
                    "site_id": "current_item.site_id",
                    "domain": "input_data.domain",
                    "work_item_id": "current_item.id",
                    "item_type": "current_item.item_type",
                    "spec": "current_item.spec",
                    "current_page": "current_item.spec",
                    "refresh_site_components?": "current_item.spec.refresh_site_components",
                    "edit_type?": "current_item.spec.edit_type",
                    "page_name?": "current_item.spec.page_name",
                    "slot_name?": "current_item.spec.slot_name",
                    "field_updates?": "current_item.spec.field_updates",
                    "replacement_content_data?": "current_item.spec.replacement_content_data",
                    "new_component_function?": "current_item.spec.new_component_function"
                  }
                },
                "next_step": "mark_complete",
                "output_field": "handler_result",
                "description": "Call handler agent"
              },
              "mark_complete": {
                "action": "complete_work_item",
                "config": {
                  "work_item_id": "current_item.id",
                  "result": "handler_result"
                },
                "next_step": "done",
                "output_field": "item_completed",
                "description": "Mark item complete"
              },
              "mark_failed": {
                "action": "fail_work_item",
                "config": {
                  "work_item_id": "current_item.id",
                  "error_message": "Handler failed"
                },
                "next_step": "done",
                "output_field": "item_failed",
                "description": "Mark item failed, continue to next"
              },
              "done": {
                "action": "loop_complete",
                "description": "Item done"
              }
            }
          }
        },
        "next_step": "complete",
        "output_field": "items_processed",
        "description": "Process each item: claim → spawn → call → mark"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": ["pending", "items_processed"]
        },
        "description": "Dispatch complete"
      }
    }
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 1800
}'::jsonb,
updated_at = NOW()
WHERE type = 'build-dispatch-loop';

-- Verify
SELECT type,
       default_config #>> '{workflow,start_step}' AS start,
    default_config #>> '{workflow,steps,process_items,action}' AS loop_action,
    default_config #>> '{workflow,steps,process_items,config,item_variable}' AS item_var
FROM agent_definitions
WHERE type = 'build-dispatch-loop';

COMMIT;

--

-- increase memory

-- 1. Increase memory for build-dispatch-loop
UPDATE agent_definitions
SET resources = '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "512Mi"}}'::jsonb,
    updated_at = NOW()
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

--

-- small batches

-- Batch processing: load 5 at a time, process, reload
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_items,config,max_items}',
        '5'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Change process_items.next_step from "complete" to "load_items"
-- This creates the batch loop: load 5 → process → load 5 → process → ... → no items → complete
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_items,next_step}',
        '"load_items"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Also reduce max_iterations in the loop to match batch size
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_items,config,max_iterations}',
        '5'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Reset the stuck claimed items
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, updated_at = NOW()
WHERE status = 'claimed';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'load_items'->'config'->>'max_items' as batch_size,
    default_config->'workflow'->'steps'->'process_items'->>'next_step' as after_batch,
    default_config->'workflow'->'steps'->'process_items'->'config'->>'max_iterations' as loop_max
FROM agent_definitions
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Expected: batch_size=5, after_batch=load_items, loop_max=5
```

The flow becomes:
```
load_items (max 5)
→ check_has_items → no items → complete
→ check_has_items → has items → process_items (loop over 5)
→ claim → spawn → call → mark complete (×5)
→ load_items (next batch of 5)
→ check_has_items → ...