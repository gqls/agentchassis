-- 517 — bugs_open/305: give the `rewrite_negations` step its own `ai_service`,
-- because it had none and the repair was therefore INERT the moment 509 applied.
--
-- MEASURED, on the first page built after 509 (orchestration 8ce1ebc0, iteration
-- 1, 2026-08-21 10:31Z): the step ran, found 3 constructions, selected 1 for
-- repair — and returned
--   {"status":"repair_unavailable","error":"no ai_service configuration resolvable",
--    "hits_before":3,"targets":1,"rewritten":[],"hits_after":3}
-- i.e. the gate was LIVE AND BLIND: detecting correctly, repairing nothing.
--
-- WHY IT HAPPENED, and it is the mirror image of the trap the council's
-- llm_reliability seat warned about (MDL-039, "root ai_service shadows step
-- config"). `resolveAIServiceConfig` looks at the agent's ROOT block and at
-- `workflow.steps.<currentStep>.config.ai_service`. page-content-writer has NO
-- root `ai_service` — its model lives on the `generate_content` STEP — and
-- `currentStep` for a loop substep is `process_sections_loop_iter_N_rewrite_negations`,
-- which is not a top-level workflow step, so neither lookup can find anything.
-- The seat was right that this action's config resolution needed stating
-- explicitly; it just turned out to fail in the opposite direction.
--
-- WHY NOT FALL BACK TO THE SIBLING STEP'S BLOCK IN CODE: reaching across to
-- `generate_content.config.ai_service` would work today and would be invisible
-- machinery — a step whose model comes from a different step's config is exactly
-- the kind of thing nobody finds when the model changes. Declared config, where
-- a reviewer of THIS step can see it.
--
-- max_tokens 2000 is deliberate and small: the answer is a handful of replacement
-- sentences, and a section-sized ceiling would only buy room for the model to
-- write an essay this gate would then reject. ⚠ If a future model runs adaptive
-- thinking out of this budget the call yields no text, the repair fails CLOSED
-- (splices nothing) and says `repair_unavailable` — loud, not silent, which is
-- the whole reason that status exists.
--
-- Config-only: live the moment it applies, no image, no roll. Anchored on the
-- step's action and needle-gated, so a re-run is a 0-row no-op and a moved path
-- raises rather than minting an orphan key.

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '517_rewrite_negations_declares_its_own_ai_service.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service}',
         '{
            "model": "claude-sonnet-5",
            "provider": "anthropic",
            "max_tokens": 2000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          }'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,action}' = 'rewrite_negations'
   AND (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service}') IS NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,model}' = 'claude-sonnet-5'
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,provider}' = 'anthropic';
  IF n <> 1 THEN
    RAISE EXCEPTION '517 FAILED: expected exactly 1 page-content-writer with an ai_service on rewrite_negations, got %', n;
  END IF;
  RAISE NOTICE '517 OK: rewrite_negations declares its own ai_service';
END $$;

COMMIT;
