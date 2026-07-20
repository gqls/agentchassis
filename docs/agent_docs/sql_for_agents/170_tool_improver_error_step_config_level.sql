-- 170_tool_improver_error_step_config_level.sql — put error_step where the
-- workflow plan actually keeps it.
--
-- CORRECTS migration 169. 169 set the refusal route as a TOP-LEVEL key on the
-- step:
--     workflow.steps.update_component.error_step = 'refuse_mangled_write'
-- because pkg/models/contracts.go declares `ErrorStep string
-- json:"error_step,omitempty"` and routeToErrorStepOrFail prefers it
-- ("Check step-level first (parallel to NextStep) — preferred location").
--
-- That reasoning was from the struct. The running system disagrees.
--
-- PROVEN 2026-07-20 by an end-to-end test on a scratch component
-- (orchestration 5176d76e-1d71-47a0-a535-b381715c05f4): the guard correctly
-- refused a 12-char replacement of a 12,051-char component and hard-errored,
-- but the orchestration went straight to FAILED at update_component instead of
-- routing. Reading the stored plan for that very run shows why:
--
--   workflow_plan #> '{steps,update_component}'  -> NO error_step key at all
--   workflow_plan #> '{steps,append_note,config}' -> "error_step": "complete"  (survives)
--
-- Both steps exist in the same stored plan, and the plan is NOT stale (it
-- contains 169's refuse_mangled_write and note_refusal steps). So the
-- top-level key is dropped on the way into workflow_plan while the
-- config-level key round-trips. routeToErrorStepOrFail then finds neither
-- step.ErrorStep nor step.Config["error_step"], and falls through to
-- failWorkflow.
--
-- This is the durable invariant from 016b, met head-on: "A config key read on a
-- different path than it's set is a silent no-op, not an error." The evidence
-- was already in front of me — every OTHER step in this same workflow
-- (append_note, compose_note) sets error_step INSIDE config. Copy the working
-- example; do not reason from the struct.
--
-- Fix: set error_step inside config as well. The top-level key is left in place
-- — harmless, and if a future chassis does preserve it, both agree.
--
-- Consequence until this is applied: the guard still protects the component
-- (that half is live and proven), but the work item does NOT reach
-- needs_human_review and NO refusal note is written — the orchestration just
-- fails. Half of "refuse, and fail honestly".

BEGIN;

SELECT snapshot_agent('tool-improver', '170_tool_improver_error_step_config_level: pre-update');

DO $$
DECLARE
  n int := 0;
BEGIN
  -- Idempotency gate: re-running must be a 0-row no-op.
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='tool-improver' AND is_active
    AND default_config #>> '{workflow,steps,update_component,config,error_step}' = 'refuse_mangled_write';
  IF n > 0 THEN
    RAISE NOTICE '170: already applied (% row(s) carry config-level error_step) — no-op', n;
    RETURN;
  END IF;

  -- Pre-condition: 169 must have landed (the target step must exist).
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='tool-improver' AND is_active
    AND default_config #> '{workflow,steps,refuse_mangled_write}' IS NOT NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '170: expected exactly 1 active tool-improver carrying 169''s refuse_mangled_write step, found % — apply 169 first', n;
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,update_component,config,error_step}',
        '"refuse_mangled_write"'::jsonb,
        true)
  WHERE type='tool-improver' AND is_active
    AND default_config #>> '{workflow,steps,update_component,config,error_step}' IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '170: expected to update 1 tool-improver row, updated %', n;
  END IF;

  -- Verify the write landed on the path the plan actually keeps.
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='tool-improver' AND is_active
    AND default_config #>> '{workflow,steps,update_component,config,error_step}' = 'refuse_mangled_write';
  IF n <> 1 THEN
    RAISE EXCEPTION '170: post-update verification failed, matched % rows', n;
  END IF;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## error_step must live in a step''s config, not at the step''s top level
Observed 2026-07-20: the bugs_open/012 write guard correctly refused a collapsed rewrite on a scratch component (12,051 -> 12 chars) and hard-errored, but the orchestration went to FAILED at update_component rather than routing to the refusal path migration 169 had wired.
Root cause: 169 set workflow.steps.update_component.error_step at the step''s TOP LEVEL (following pkg/models/contracts.go, where ErrorStep is a declared field that routeToErrorStepOrFail checks first). The stored workflow_plan for the failing run has NO error_step on that step, while append_note''s CONFIG-level error_step survives in the same plan — so the top-level key is dropped building the plan and the config-level key round-trips. routeToErrorStepOrFail finds neither and calls failWorkflow.
Fix: migration 170 sets error_step inside update_component.config as well.
Transferable: a config key read on a different path than it''s set is a silent no-op, not an error — and every other step in that same workflow already showed the working shape. Copy the working example rather than reasoning from the struct.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '170_tool_improver_error_step_config_level'
);

COMMIT;
