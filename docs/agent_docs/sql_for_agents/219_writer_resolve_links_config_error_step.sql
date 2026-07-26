-- ============================================================================
-- 219_writer_resolve_links_config_error_step.sql — bugs_open/086, contained fix
--
-- WHY. page-content-writer's resolve_links step already declares
--   "error_step": "select_sections"
-- at STEP level — the author's explicit statement that a failed link-resolve is
-- survivable. It has never once taken effect: convertToWorkflowPlan
-- (platform/messaging/processor.go) builds models.Step field by field and did
-- not copy error_step, so no persisted workflow_plan has ever carried one
-- (measured: 0 of 14,209 plan steps over three days). The config-level twin
-- DOES survive, because the whole config map is copied verbatim, and
-- routeToErrorStepOrFail falls back to it (coordinator.go:3234).
--
-- So this adds the twin, inside config, for this one step. Effect: when the
-- link resolver times out or errors, the writer continues at select_sections
-- instead of dying. All 32 resolve_links failures ever logged were fatal; 30 of
-- them are "timed out after 3 retries" — roughly one a day over the last month.
--
-- This is the contained alternative the council gate named when it vetoed the
-- fleet-wide Go fix (corr 88ef6d08-ca87-4c90-b682-f85f1e6036f1): one agent,
-- config-only, live immediately, no image roll, on the battle-tested path.
-- The Go fix is kept (owner ruling 2026-07-26) and remains the durable answer;
-- this seed is safe alongside it — the coordinator prefers the step-level field
-- and falls back to this one, and both name the same target.
--
-- REVERT:
--   UPDATE agent_definitions
--   SET default_config = default_config #- '{workflow,steps,resolve_links,config,error_step}'
--   WHERE type='page-content-writer' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- ============================================================================

DO $$
DECLARE
  step_level text;
  cfg_level  text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'resolve_links'->>'error_step',
         default_config->'workflow'->'steps'->'resolve_links'->'config'->>'error_step'
    INTO step_level, cfg_level
  FROM agent_definitions
  WHERE type='page-content-writer' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF step_level IS NULL THEN
    RAISE EXCEPTION 'resolve_links has no step-level error_step — the anchor this seed mirrors is gone; re-read the step before applying';
  END IF;
  IF cfg_level IS NOT NULL THEN
    RAISE NOTICE 'config.error_step already present (%) — nothing to do', cfg_level;
    RETURN;
  END IF;

  -- mirror the step-level value rather than hardcoding it, so the two cannot drift
  UPDATE agent_definitions
  SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,resolve_links,config,error_step}',
        to_jsonb(step_level))
  WHERE type='page-content-writer' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  RAISE NOTICE 'resolve_links config.error_step set to % (mirrors the step-level declaration)', step_level;
END $$;

-- verify: both levels present, identical, and the rest of the step untouched
SELECT default_config->'workflow'->'steps'->'resolve_links'->>'error_step'          AS step_level,
       default_config->'workflow'->'steps'->'resolve_links'->'config'->>'error_step' AS config_level,
       default_config->'workflow'->'steps'->'resolve_links'->>'next_step'            AS next_step,
       default_config->'workflow'->'steps'->'resolve_links'->'config'->>'agent_type' AS callee
FROM agent_definitions
WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
