-- ============================================================================
-- page-build-handler: stop laundering a save_page_sections FAILURE into a
-- `complete` success.
-- ============================================================================
--
-- BUG (diagnosed 2026-06-15, gamesdesign index):
--   save_page_sections has a content-regression guard that `return nil, err`s
--   when the new content's stripped text < existingTextLen/4 (existing>200) —
--   a correct refusal to overwrite a rich page with a thin one. But in
--   page-build-handler the `save_sections` step has `error_step: complete_error`,
--   and `complete_error` is a `complete_workflow` (success exit). So the guard's
--   genuine error is swallowed: the work item goes `complete`, the deploy
--   re-renders the STALE stored components, git commits stale HTML, and the
--   operator sees success. index reproduced this 3× (06-13/14/15), each run
--   committing the 06-06 components with the phantom hero CTAs.
--
--   It never reaches agent_error_log either (that's errors only; the WARN
--   "CONTENT REGRESSION BLOCKED" is stdout-only, and the workflow exits success).
--
-- FIX (structural, reuses the existing pattern):
--   Route save_sections' error to a NEW `mark_save_failed` step — a
--   `fail_work_item` (the same action `mark_needs_review` already uses) that sets
--   the WORK ITEM to a non-terminal `needs_human_review` status with a
--   save-specific message — instead of to `complete_error`. The work-item status
--   is the durable signal the operator sees; marking it needs_human_review makes
--   a blocked/failed save VISIBLE instead of a silent `complete`.
--
--   Why needs_human_review (not failed): a regression block is NOT a transient —
--   re-queuing reproduces it. It needs a human to look at WHY the writer returned
--   thin content (the separate, still-open content question). needs_human_review
--   is non-terminal, so it holds (and the dedup index keeps a same-key re-queue
--   from silently re-running it). `failed` would be wrong (terminal, re-runnable).
--
-- SCOPE: this fixes the SILENT-COMPLETION (visibility) only. It does NOT change
--   the deploy-proceeds-after-no-write behaviour (separate item — gate deploy on
--   sections_saved>0), and does NOT address WHY index's recreate/preserve writer
--   returns thin content (separate content investigation). Those are deliberately
--   out of scope here.
--
-- NB on a DIFFERENT failure mode that SHOULD still pass through:
--   save_page_sections also has benign success-exits (skipped: no DB / no page /
--   "no sections found") that `return ... nil` (a result, not an error) and do
--   NOT trigger error_step — they flow to next_step (update_status) normally.
--   Those are unaffected by this change; only the actual ERROR return (the
--   regression guard, and any future save error) now routes to mark_save_failed.
--   That is the intended behaviour: a true save failure must be visible; a
--   legitimate skip continues.
--
-- Snapshot + reversible.
-- ============================================================================

-- 0. Snapshot
SELECT snapshot_agent('page-build-handler',
  'before routing save_sections error to mark_save_failed (silent-completion fix 2026-06-15)');

-- 1. Add the new mark_save_failed step (fail_work_item → needs_human_review),
--    then repoint save_sections.error_step to it.
--    Two chained jsonb_set on default_config.workflow.steps.
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,mark_save_failed}',
            jsonb_build_object(
                'action', 'fail_work_item',
                'config', jsonb_build_object(
                    'work_item_id', 'input_data.work_item_id',
                    'error_message', 'Page save failed (e.g. content-regression guard blocked an overwrite, or save error) — needs human review. The deployed page may be stale.',
                    'status_override', 'needs_human_review'
                ),
                'next_step', 'complete_error',
                'description', 'Mark the work item needs_human_review when save_page_sections returns an error (e.g. content-regression block). Makes a blocked/failed save VISIBLE instead of a silent complete. Flows to complete_error afterward so the workflow terminates cleanly, but the work item carries the review status.',
                'output_field', 'save_failure_result'
            ),
            true
        ),
        '{workflow,steps,save_sections,error_step}',
        '"mark_save_failed"',
        true
    ),
    updated_at = now()
WHERE type = 'page-build-handler';

-- 2. Verify:
--    a) save_sections.error_step now = mark_save_failed (both the config-level and
--       step-level error_step exist on this step in the row; update the step-level
--       one that the engine reads — confirm which the engine honours, see note).
--    b) mark_save_failed exists and is a fail_work_item with needs_human_review.
SELECT
  default_config #>> '{workflow,steps,save_sections,error_step}'              AS save_sections_error_step,
  default_config #>> '{workflow,steps,save_sections,config,error_step}'       AS save_sections_config_error_step,
  default_config #>> '{workflow,steps,mark_save_failed,action}'              AS msf_action,
  default_config #>> '{workflow,steps,mark_save_failed,config,status_override}' AS msf_status,
  default_config #>> '{workflow,steps,mark_save_failed,next_step}'           AS msf_next
FROM agent_definitions
WHERE type = 'page-build-handler';
-- Expect: save_sections_error_step=mark_save_failed; msf_action=fail_work_item;
--         msf_status=needs_human_review; msf_next=complete_error.

-- ============================================================================
-- IMPORTANT — TWO error_step locations on save_sections (READ BEFORE DEPLOY):
-- The deployed row has error_step in BOTH places on the save_sections step:
--   steps.save_sections.error_step            = "complete_error"   (step-level)
--   steps.save_sections.config.error_step     = "complete_error"   (config-level)
-- Step 1 above repoints the STEP-LEVEL one. Most chassis workflow engines read
-- the step-level error_step; but this engine ALSO carries error_step inside many
-- steps' config (e.g. plan_sections, load_spec_sections), so it is not certain
-- which the engine honours for routing.
--
-- DO NOT GUESS. Before relying on this, confirm which error_step the engine reads
-- for routing (grep the chassis for how it resolves error_step — step.ErrorStep vs
-- config["error_step"]). If it reads the CONFIG one, also repoint that:
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--       '{workflow,steps,save_sections,config,error_step}', '"mark_save_failed"', true),
--       updated_at = now()
--   WHERE type = 'page-build-handler';
-- Safest: set BOTH to mark_save_failed once confirmed they don't conflict. The
-- verification query above surfaces both so you can see their current values.
-- ============================================================================

-- Revert: SELECT revert_agent('page-build-handler');
