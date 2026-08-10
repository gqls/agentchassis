-- 373_provocation_generator_batch_size.sql
--
-- Drop the generator's batch size from 8 to 4.
--
-- WHY — AND THE REASON IS NOT THE ONE 372 GAVE
-- Migration 372 set `ai_service.max_tokens = 8000` on the generate step after a
-- truncation. The next run truncated again, at **output_tokens=2048** — the
-- provider client's hardcoded fallback (`platform/aiservice/anthropic.go:109`),
-- i.e. the configured 8000 never reached the API.
--
-- The cause is in the Go, not the config. `max_tokens` is honoured only when it
-- is passed in the OPTIONS map to `GenerateText` (`anthropic.go:147`), and
-- `ExecuteAIStepAction` is what normally builds that map from
-- `ai_service.max_tokens` (`ai_actions.go:358-364`). Both provocation actions
-- call `GenerateText` directly with an EMPTY options map
-- (`provocation_generator_action.go:453`, `provocation_gate_action.go:830`), so
-- they bypass that builder entirely and always run at 2048. This is exactly the
-- class `bugs_open/205` counted — steps whose output budget nobody actually set,
-- discovered only when something large hit the cliff. The gate has been getting
-- away with it because a verdict is small; the generator asks for eight
-- documents and could not.
--
-- 372's 8000 is therefore INERT until that Go change ships, and is deliberately
-- left in place: it is correct, it is what the step wants, and it starts working
-- the moment the fix rolls. What this file changes is the half that can take
-- effect today.
--
-- WHY 4 IS A SIZING DECISION AND NOT JUST A WORKAROUND
-- 4 candidates at the rules' 900-character body limit is ~3.6k characters, ~950
-- tokens — comfortable inside 2048 with room for the model to overshoot. It is
-- also the better shape independently of the bug: several small batches drawn
-- against a `recentTitles` list that grows between runs produce more varied
-- candidates than one large batch generated in a single breath, and one bad
-- batch costs less. When the Go fix lands, raise this deliberately with a run to
-- justify it — do not assume 8 was right just because it was first.
--
-- Idempotent; safe to re-run.

BEGIN;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,generate,config,count}',
         '4'::jsonb,
         true
       ),
       updated_at = now()
 WHERE type = 'provocation-generator-manual'
   AND is_active AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

DO $$
DECLARE c text; m text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'generate'->'config'->>'count',
         default_config->'workflow'->'steps'->'generate'->'config'->'ai_service'->>'model'
    INTO c, m
    FROM agent_definitions
   WHERE type='provocation-generator-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

  IF c IS DISTINCT FROM '4' THEN
    RAISE EXCEPTION 'count is % on the generate step, expected 4', c;
  END IF;
  -- Same guard as 372: jsonb_set with create_missing invents a branch on a typo,
  -- and the result looks like a working config with the model quietly gone.
  IF m IS NULL THEN
    RAISE EXCEPTION 'the generate step has no model after the update — jsonb_set created a new branch';
  END IF;

  RAISE NOTICE 'generate step count = 4 (fits the 2048 hardcoded cap until the options-map fix ships), model still present';
END $$;

COMMIT;
