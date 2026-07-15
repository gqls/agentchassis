-- 149: page-build-handler — no-op exits must not read as success
--
-- Context (2026-07-14, robot-hands false-completion investigation):
-- page-build-handler's complete_error step is a SUCCESS-labelled
-- complete_workflow. Two branches reach it without doing any work and
-- without flagging the work item:
--
--   check_has_ready_sections (ready_count == 0)  → complete_error
--   check_content_produced   (writer skipped)     → complete_error
--
-- The dispatch loop (build-dispatch-loop / site-work-orchestrator) calls
-- complete_work_item on every successful handler saga, so those no-ops got
-- stamped 'complete'. Live catch: gripper-detail's product sections
-- (pages.sections = [] → ready_count 0 → no-op) were marked complete on
-- 2026-07-10 and still served empty markup on 2026-07-14; two-strike then
-- parked the re-detections as non-dispatchable 'unresolved' zombies.
--
-- This migration inserts flag steps on both branches. The flag uses
-- update_work_item_status (skip_if_missing: true), so invocations without a
-- work item — plain build-pipeline calls on sectionless pages — pass through
-- unchanged. CompleteWorkItemAction's guard preserves needs_human_review, so
-- the dispatch loop's subsequent complete_work_item becomes a no-op.
--
-- Requires the chassis image with:
--   - update_work_item_status accepting needs_human_review + error_message
--   - the completion-verification gate (complete_work_item_verification.go),
--     which independently blocks false completions for empty_section items
--     even if this workflow change is missing (defence in depth).
--
-- Verify after applying:
--   SELECT default_config->'workflow'->'steps'->'check_has_ready_sections'->'config'->>'else_step',
--          default_config->'workflow'->'steps'->'check_content_produced'->'config'->>'else_step'
--   FROM agent_definitions WHERE type = 'page-build-handler' AND is_active;
--   -- expect: mark_no_ready_sections | mark_writer_skipped

BEGIN;

UPDATE agent_definitions
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(
          jsonb_set(
            default_config,
            '{workflow,steps,mark_no_ready_sections}',
            '{
              "action": "update_work_item_status",
              "config": {
                "status": "needs_human_review",
                "skip_if_missing": true,
                "error_message": "page-build-handler no-op: no sections ready to build (empty spec sections, or all sections deferred for missing data) — the target section was NOT rebuilt",
                "work_item_id_field": "input_data.work_item_id"
              },
              "next_step": "complete_error",
              "error_step": "complete_error",
              "description": "No ready sections — park the work item visibly instead of letting the dispatch loop stamp it complete"
            }'::jsonb
          ),
          '{workflow,steps,mark_writer_skipped}',
          '{
            "action": "update_work_item_status",
            "config": {
              "status": "needs_human_review",
              "skip_if_missing": true,
              "error_message": "page-build-handler no-op: content writer skipped this page — the target section was NOT rebuilt",
              "work_item_id_field": "input_data.work_item_id"
            },
            "next_step": "complete_error",
            "error_step": "complete_error",
            "description": "Writer skipped — park the work item visibly instead of letting the dispatch loop stamp it complete"
          }'::jsonb
        ),
        '{workflow,steps,check_has_ready_sections,config,else_step}',
        '"mark_no_ready_sections"'
      ),
      '{workflow,steps,check_content_produced,config,else_step}',
      '"mark_writer_skipped"'
    ),
    updated_at = NOW()
WHERE type = 'page-build-handler'
  AND is_active
  AND deleted_at IS NULL;

COMMIT;
