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

-- Human Change Requests - Work Items
-- Apply after migration 071 (loop-based dispatch).
--
-- Sites:
--   gaswholesalers.com  5fe15466-4e2e-4ff2-981e-98c1b7074002
--   finetuning.uk       1368e337-dd1d-4799-bbb3-8221a1b79bcc

BEGIN;

-- Clear any previous human items that haven't started processing
DELETE FROM site_work_items
WHERE source = 'human'
  AND status IN ('detected', 'triaged')
  AND item_key LIKE 'human_%';

-- ============================================================================
-- GASWHOLESALERS.COM
-- ============================================================================

INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary, spec,
    priority, handler_agent, status, created_by, item_key
) VALUES
-- 1. Redesign CSS
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'needs_design', 'high',
    'Redesign: consistent color scheme, nav on one line, hero image visible',
    '{}'::jsonb,
    5, 'webdesign-agent', 'triaged', 'admin',
    'human_redesign_gaswholesalers'),
-- 2. Phone
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'content_edit', 'medium',
    'Update phone number to +44 (0) 7934 524 911',
    '{"edit_type": "content_edit", "page_name": "contact", "slot_name": "contact-info", "field_updates": {"phone": "+44 (0) 7934 524 911"}}'::jsonb,
    10, 'section-editor', 'triaged', 'admin',
    'human_phone_gaswholesalers'),
-- 3. Email
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'content_edit', 'medium',
    'Update email to gaswholesalers@contactforsales.com',
    '{"edit_type": "content_edit", "page_name": "contact", "slot_name": "contact-info", "field_updates": {"email": "gaswholesalers@contactforsales.com"}}'::jsonb,
    11, 'section-editor', 'triaged', 'admin',
    'human_email_gaswholesalers'),
-- 4. Remove hours
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'content_edit', 'medium',
    'Remove hours panel from contact page',
    '{"edit_type": "content_edit", "page_name": "contact", "slot_name": "contact-info", "field_updates": {"hours": "", "business_hours": "", "opening_hours": ""}}'::jsonb,
    12, 'section-editor', 'triaged', 'admin',
    'human_remove_hours_gaswholesalers'),
-- 5. Logo
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'needs_logo', 'medium',
    'Generate company logo for Gas Wholesalers',
    '{"purpose": "logo", "image_prompts": {"logo": "Professional logo for Gas Wholesalers, a wholesale fuel distribution company. Modern industrial design with fuel pipeline or gas flame iconography. Bold corporate typography. Dark blue and orange palette."}}'::jsonb,
    20, 'image-build-handler', 'triaged', 'admin',
    'human_logo_gaswholesalers'),
-- 6. Hero image
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'needs_hero_image', 'medium',
    'Generate hero image for Gas Wholesalers homepage',
    '{"purpose": "hero", "image_prompts": {"hero_home": "Professional industrial scene of fuel distribution terminal with large storage tanks and tanker trucks. Modern petroleum facility at dawn with dramatic lighting. Clean corporate photography style. Dark blue and orange tones."}}'::jsonb,
    21, 'image-build-handler', 'triaged', 'admin',
    'human_hero_gaswholesalers'),
-- 7. Rerender (runs last)
(   '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'human', 'build', 'needs_rerender', 'high',
    'Reassemble and deploy all pages after human change requests',
    '{"refresh_site_components": true}'::jsonb,
    99, 'rerender-pages', 'triaged', 'admin',
    'human_rerender_gaswholesalers');

-- Re-triage tool evaluation if it exists
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL, result = '{}'::jsonb
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND item_key = 'evaluate_tools_gaswholesalers'
  AND status NOT IN ('complete', 'claimed');

-- ============================================================================
-- FINETUNING.UK
-- ============================================================================

INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary, spec,
    priority, handler_agent, status, created_by, item_key
) VALUES
-- 1. Redesign
(   '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'human', 'build', 'needs_design', 'high',
    'Redesign: vibrant colors reflecting multi-agent AI platform managing thousands of domains',
    '{}'::jsonb,
    5, 'webdesign-agent', 'triaged', 'admin',
    'human_redesign_finetuning'),
-- 2. Email
(   '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'human', 'build', 'content_edit', 'medium',
    'Update email to finetuning@contactforsales.com',
    '{"edit_type": "content_edit", "page_name": "contact", "slot_name": "contact-info", "field_updates": {"email": "finetuning@contactforsales.com"}}'::jsonb,
    10, 'section-editor', 'triaged', 'admin',
    'human_email_finetuning'),
-- 3. Hero image
(   '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'human', 'build', 'needs_hero_image', 'medium',
    'Generate hero image - AI agents orchestrating web creation',
    '{"purpose": "hero", "image_prompts": {"hero_home": "Dynamic abstract visualization of interconnected AI agents working together to build websites. Glowing neural network nodes and data streams. Futuristic dark background with vivid accent colors - electric purple, cyan, and amber. Professional tech company aesthetic."}}'::jsonb,
    20, 'image-build-handler', 'triaged', 'admin',
    'human_hero_finetuning'),
-- 4. Logo
(   '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'human', 'build', 'needs_logo', 'medium',
    'Generate logo - leopardess or ant colony theme',
    '{"purpose": "logo", "image_prompts": {"logo": "Sleek modern logo for Finetuning, an AI agent orchestration platform. Abstract leopardess silhouette composed of interconnected nodes and data pathways. Minimalist geometric style. Electric purple and cyan on transparent background. Professional tech branding."}}'::jsonb,
    21, 'image-build-handler', 'triaged', 'admin',
    'human_logo_finetuning'),
-- 5. Rerender
(   '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'human', 'build', 'needs_rerender', 'high',
    'Reassemble and deploy all pages after human change requests',
    '{"refresh_site_components": true}'::jsonb,
    99, 'rerender-pages', 'triaged', 'admin',
    'human_rerender_finetuning');

COMMIT;

-- Verify
SELECT s.domain, wi.item_type, wi.handler_agent, wi.priority, wi.status
FROM site_work_items wi
         JOIN sites s ON s.id = wi.site_id
WHERE wi.source = 'human'
  AND wi.status = 'triaged'
ORDER BY s.domain, wi.priority;