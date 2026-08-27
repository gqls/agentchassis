-- 662_design_critique_vision_falls_back_to_claude.sql — model-trial leg B for the 018 critic.
--
-- WHY: the owner's recorded decision (vigilant PLAN_2026-08-02 Phase 2) was "trial both
-- models — try Gemini first and see if it works better; revisit if not". The evidence is now
-- in: Gemini produced ONE good report (2026-08-26, 16 before-captures) and then failed the
-- SAME call REPRODUCIBLY — 400 INVALID_ARGUMENT "Unable to process input image" on runs
-- 0eff246f AND a21e0c3e (2026-08-27), after the hero batch made the full-page captures
-- taller. The pipeline has no downscaling anywhere (approved plan §cost envelope warned
-- this); Anthropic's API resizes oversized images server-side, Gemini's rejects them.
-- This flips the critique step's ai_service to the 317-proven anthropic/claude-sonnet-5
-- vision config. Reversible: snapshot below + the inverse jsonb_set in the comment at foot.
-- The Gemini datum is recorded in SQ-003 and the leopardess RUNNING_NOTES 2026-08-27.

BEGIN;

SELECT snapshot_agent('design-critique-agent', '662_design_critique_vision_falls_back_to_claude.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,critique,config,ai_service}',
    '{"provider": "anthropic", "model": "claude-sonnet-5", "api_key_env_var": "ANTHROPIC_API_KEY", "max_tokens": 6000}'::jsonb)
WHERE type = 'design-critique-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE prov text; mdl text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'critique'->'config'->'ai_service'->>'provider',
         default_config->'workflow'->'steps'->'critique'->'config'->'ai_service'->>'model'
    INTO prov, mdl
  FROM agent_definitions
  WHERE type='design-critique-agent' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF prov <> 'anthropic' OR mdl <> 'claude-sonnet-5' THEN
    RAISE EXCEPTION 'critique ai_service is %/% (want anthropic/claude-sonnet-5)', prov, mdl;
  END IF;
END $$;

COMMIT;

-- ROLLBACK (manual): restore from the snapshot taken above, or inverse jsonb_set:
--   UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--     '{workflow,steps,critique,config,ai_service}',
--     '{"provider":"gemini","model":"gemini-pro-latest","api_key_env_var":"GEMINI_API_KEY","max_tokens":6000}'::jsonb)
--   WHERE type='design-critique-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
