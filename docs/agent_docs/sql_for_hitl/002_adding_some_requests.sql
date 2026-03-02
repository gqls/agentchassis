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