-- 421_offer_analyser_prompt_undouble_apostrophes_ROLLBACK.sql
--
-- Restores 408's byte-identical prompt by re-doubling every apostrophe.
-- Exact ONLY because the fixed prompt's apostrophes are all ex-pairs
-- (lone count was 0 before 421) — the guard refuses if that stops being true.

BEGIN;

DO $guard$
DECLARE
  p text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
    INTO p
  FROM agent_definitions
  WHERE type = 'offer-analyser' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF p IS NULL THEN
    RAISE EXCEPTION 'guard: no live offer-analyser prompt';
  END IF;
  IF position($$''$$ in p) > 0 THEN
    RAISE EXCEPTION 'guard: prompt already carries doubled apostrophes — 421 not applied, or already rolled back';
  END IF;
  IF length(p) <> 5484 THEN
    RAISE EXCEPTION 'guard: expected the post-421 prompt (5484 chars), found % — the prompt has been edited since; a blind re-double would corrupt it', length(p);
  END IF;
END $guard$;

UPDATE agent_definitions SET
  default_config = jsonb_set(
    default_config,
    '{workflow,steps,run_offer_analysis,config,prompt}',
    to_jsonb(replace(
      default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
      $$'$$, $$''$$))
  ),
  updated_at = now()
WHERE type = 'offer-analyser' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $verify$
DECLARE
  p text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
    INTO p
  FROM agent_definitions
  WHERE type = 'offer-analyser' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF length(p) <> 5494 THEN
    RAISE EXCEPTION 'rollback verify: expected 408''s 5494 chars back, found %', length(p);
  END IF;
END $verify$;

COMMIT;
