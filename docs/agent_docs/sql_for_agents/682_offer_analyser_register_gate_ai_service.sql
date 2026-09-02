-- FILE: 682_offer_analyser_register_gate_ai_service.sql
--
-- Gives the producer register gate a model to call. WITHOUT THIS THE GATE IS
-- LIVE AND INERT.
--
-- ⚠ THIS IS THE DISPOSITION OF A COUNCIL "MISSING" ITEM THAT PREDICTED IT.
-- `llm_reliability` on the approved round (4054f4d9) wrote: "Plan does not state
-- max_tokens or thinking config for the new LLM call... nor confirm whether the
-- target agent has a root ai_service block that would shadow any step-level
-- config (MDL-039) — worth confirming at review-application time."
--
-- Confirmed at application time, and it came out badly:
--   [VERIFIED 2026-09-02, live row] offer-analyser has NO root `ai_service` block
--   at all. Its only model config sits on the `run_offer_analysis` step. So
--   resolveAIServiceConfig(agentConfig, stepConfig, "repair_ordering_register")
--   overlays root (absent) + this step's own block (absent) + runtime (absent)
--   and returns an EMPTY map.
--
-- The action handles that correctly and SAFELY — it returns
-- "no ai_service configuration resolvable", keeps every point, and records the
-- reason against each one — so the failure is loud and costs nothing. But it
-- repairs NOTHING, and a reader seeing the step present and the run green would
-- have every reason to think the gate was working.
--
-- ⚠ THAT IS THE `params.StorageClient` SHAPE, EXACTLY: a capability with no live
-- caller has an untested dependency on its ENVIRONMENT, and the first real call
-- is what finds it. "No live call yet" means the deployment contract is
-- UNVERIFIED, not merely that the thing is unused. Caught here by asking the
-- live row before the first run rather than after it.
--
-- MDL-039 (root shadowing a step block) does NOT apply: there is no root block to
-- shadow, so this step-level block is the sole source and cannot be overlaid.
--
-- Model choice is deliberately NOT a decision made here: it mirrors the agent's
-- own existing block (`run_offer_analysis`), so the repair runs on the same model
-- that wrote the text. `max_tokens` is the one difference and is set DOWN to
-- 2000, matching the action's own default and its stated reason — the answer is a
-- handful of restated sentences, and a section-sized ceiling only buys room for
-- the model to write an essay the judge would then reject.
--
-- Config-only, live immediately, no image dependency. Rollback: 682_..._ROLLBACK.

BEGIN;

DO $$
DECLARE v_cfg jsonb; v_root jsonb;
BEGIN
  SELECT default_config INTO v_cfg FROM agent_definitions
   WHERE type='offer-analyser' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF NOT (v_cfg->'workflow'->'steps' ? 'repair_ordering_register') THEN
    RAISE EXCEPTION 'repair_ordering_register step is absent — apply 681 first';
  END IF;
  IF v_cfg->'workflow'->'steps'->'repair_ordering_register'->'config' ? 'ai_service' THEN
    RAISE EXCEPTION 'the step already carries an ai_service block — another session has done this; read the live row';
  END IF;
  -- ⚠ If a root block APPEARS later, this step-level one would overlay it rather
  -- than be shadowed by it (root is applied first, step second). Recorded so a
  -- future reader knows the precedence rather than guessing: step WINS.
  v_root := v_cfg->'ai_service';
  IF v_root IS NOT NULL AND v_root <> 'null'::jsonb THEN
    RAISE NOTICE 'NOTE: a root ai_service block now exists (%). This step block overlays it.', v_root;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,repair_ordering_register,config,ai_service}',
         jsonb_build_object(
           'provider',        'anthropic',
           'model',           'claude-sonnet-4-6',
           'max_tokens',      2000,
           'api_key_env_var', 'ANTHROPIC_API_KEY'
         ),
         true)
 WHERE type='offer-analyser' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE v jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'repair_ordering_register'->'config'->'ai_service'
    INTO v FROM agent_definitions
   WHERE type='offer-analyser' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v IS NULL THEN RAISE EXCEPTION 'ai_service block was not written'; END IF;
  IF v->>'model' IS NULL OR v->>'provider' IS NULL OR v->>'api_key_env_var' IS NULL THEN
    RAISE EXCEPTION 'ai_service block is incomplete: %', v;
  END IF;
  -- An EMPTY string in any of these is the overlay-blanking shape
  -- checkOverlayRequiredKeys exists to catch; refuse to create one.
  IF v->>'model' = '' OR v->>'provider' = '' OR v->>'api_key_env_var' = '' THEN
    RAISE EXCEPTION 'ai_service carries an empty required key, which silently blanks a resolved config: %', v;
  END IF;
  RAISE NOTICE 'offer-analyser: register gate can now resolve a model (%/%)', v->>'provider', v->>'model';
END $$;

COMMIT;
