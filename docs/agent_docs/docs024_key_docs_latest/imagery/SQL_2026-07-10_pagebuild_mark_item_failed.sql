-- SQL_2026-07-10_pagebuild_mark_item_failed.sql
--
-- Fix for the "no-op complete" anomaly (running notes Turns 10/15, thread A):
-- page-build-handler's step-level error routing pointed straight at
-- complete_error, which is a SUCCESS-labelled complete_workflow — so a real
-- step failure (observed: content-writer child dying on a Kafka
-- "topic partition not found" reply error) completed the orchestration as
-- success and the dispatcher stamped the work item 'complete' with no error.
-- CompleteWorkItemAction's guard already preserves flagged/terminal statuses;
-- the established pattern (mark_needs_review, mark_no_sections) is "flag the
-- item BEFORE completing". This migration extends that pattern to real
-- errors:
--   1. Adds a mark_item_failed step (update_work_item_status → status
--      'failed', attempt_count bumped, skip_if_missing for non-item runs).
--   2. Repoints all 8 step-level "error_step": "complete_error" references to
--      mark_item_failed, which then proceeds to complete_error.
-- Benign branches keep their existing flag-first behaviour. Failed items are
-- now VISIBLE and queryable instead of silently complete.
--
-- Workflow-config-only: no code deploy required; live for the next page build.

\set ON_ERROR_STOP on

-- ── Backup (outside transaction) ──
CREATE TABLE IF NOT EXISTS agent_def_page_build_handler_backup_20260710_mark_item_failed AS
SELECT * FROM agent_definitions
WHERE type = 'page-build-handler' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

SELECT count(*) AS backup_rows
FROM agent_def_page_build_handler_backup_20260710_mark_item_failed;

BEGIN;

SELECT snapshot_agent('page-build-handler',
                      'route step errors through mark_item_failed before complete_error (no-op-complete fix)');

-- 1. Repoint all error_step references (text-level on the JSON, then cast back).
UPDATE agent_definitions
SET default_config = replace(default_config::text,
        '"error_step": "complete_error"',
        '"error_step": "mark_item_failed"')::jsonb,
    updated_at = now()
WHERE type = 'page-build-handler' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 2. Add the mark_item_failed step. Its own failure path goes straight to
--    complete_error so a broken status update cannot loop.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,mark_item_failed}',
        '{
          "action": "update_work_item_status",
          "config": {
            "status": "failed",
            "work_item_id_field": "input_data.work_item_id",
            "skip_if_missing": true
          },
          "next_step": "complete_error",
          "error_step": "complete_error",
          "description": "Real step error — mark the work item failed (visible, attempt counted) before completing via the error path"
        }'::jsonb),
    updated_at = now()
WHERE type = 'page-build-handler' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- ── Verify ──
DO $verify$
DECLARE
    cfg text;
    n_mark int;
    n_complete_err int;
BEGIN
    SELECT default_config::text INTO cfg
    FROM agent_definitions
    WHERE type = 'page-build-handler' AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false);

    n_mark := (length(cfg) - length(replace(cfg, '"error_step": "mark_item_failed"', '')))
              / length('"error_step": "mark_item_failed"');
    n_complete_err := (length(cfg) - length(replace(cfg, '"error_step": "complete_error"', '')))
              / length('"error_step": "complete_error"');

    IF n_mark <> 8 THEN
        RAISE EXCEPTION 'expected 8 error_step→mark_item_failed pointers, got %', n_mark;
    END IF;
    -- The only remaining complete_error pointer is mark_item_failed's own.
    IF n_complete_err <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 remaining error_step→complete_error (inside mark_item_failed), got %', n_complete_err;
    END IF;
    IF position('"mark_item_failed"' IN cfg) = 0 THEN
        RAISE EXCEPTION 'mark_item_failed step not found after update';
    END IF;

    RAISE NOTICE 'page-build-handler error routing fixed: 8 pointers repointed, mark_item_failed step added';
END
$verify$;

COMMIT;
