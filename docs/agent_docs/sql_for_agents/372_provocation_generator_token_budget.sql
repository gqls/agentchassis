-- 372_provocation_generator_token_budget.sql
--
-- Give the generator's `generate` step an explicit `max_tokens` of 8000.
--
-- WHY, WITH THE MEASUREMENT
-- The first real generation round (orchestration c49adc3f-ea8c-481c-9734-2760fc81373c,
-- 2026-08-10 18:18Z) failed with:
--
--   generator call failed: response truncated: stop_reason=max_tokens
--
-- `platform/aiservice/anthropic.go:109` defaults `max_tokens` to **2048** when the
-- step config does not set it, and migration 371 did not set it. A batch of 8
-- candidates needs roughly 8 × (title + teaser + a body the rules cap at 900
-- characters) ≈ 6.7k characters ≈ ~1.8k tokens before JSON overhead — so 2048 was
-- always going to cut the reply, and the size of the ask made that certain rather
-- than unlucky. 8000 is the fleet's most common explicit value (50 steps) and
-- leaves ~3.5× headroom at the configured count.
--
-- WHAT THIS IS NOT: a workaround for a truncation. The platform did exactly the
-- right thing — `anthropic.go:246` turns `stop_reason=max_tokens` into an ERROR
-- rather than returning the partial, so `parseGenerated` never saw a cut JSON
-- array and no half-written provocation reached the pool. CLAUDE.md's rule
-- ("`output_tokens == max_tokens` means the completion was CUT, not finished")
-- is enforced in the client here, and this is what it looks like when it works.
-- Raising the budget is the correct response to a *correctly reported* cut.
--
-- WHY A NEW FILE RATHER THAN EDITING 371: 371 is recorded as applied, and its
-- insert is `WHERE NOT EXISTS`, so re-running it would change nothing. Migrations
-- are forward-only.
--
-- Idempotent; safe to re-run.

BEGIN;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,generate,config,ai_service,max_tokens}',
         '8000'::jsonb,
         true
       ),
       updated_at = now()
 WHERE type = 'provocation-generator-manual'
   AND is_active AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

DO $$
DECLARE mt text; cnt int;
BEGIN
  SELECT count(*) INTO cnt FROM agent_definitions
   WHERE type='provocation-generator-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF cnt <> 1 THEN RAISE EXCEPTION 'expected exactly 1 active generator agent, found %', cnt; END IF;

  SELECT default_config->'workflow'->'steps'->'generate'->'config'->'ai_service'->>'max_tokens'
    INTO mt FROM agent_definitions
   WHERE type='provocation-generator-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF mt IS DISTINCT FROM '8000' THEN
    RAISE EXCEPTION 'max_tokens is % on the generate step, expected 8000 — jsonb_set addressed the wrong path', mt;
  END IF;

  -- The sibling path must still be intact: jsonb_set with create_missing=true will
  -- happily invent a whole branch if any key above it is misspelled, leaving a
  -- plausible-looking config whose model is now unset. Assert a key that was
  -- already there.
  IF (SELECT default_config->'workflow'->'steps'->'generate'->'config'->'ai_service'->>'model'
        FROM agent_definitions
       WHERE type='provocation-generator-manual' AND is_active
         AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false) IS NULL THEN
    RAISE EXCEPTION 'the generate step has no model after the update — jsonb_set created a new branch instead of editing the existing one';
  END IF;

  RAISE NOTICE 'generate step max_tokens = 8000, model still present';
END $$;

COMMIT;
