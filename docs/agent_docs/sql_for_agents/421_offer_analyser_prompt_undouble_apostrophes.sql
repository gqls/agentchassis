-- 421_offer_analyser_prompt_undouble_apostrophes.sql
--
-- Cosmetic-but-real defect in migration 408, found by this lane's own review
-- pass on 2026-08-15: the offer-analyser prompt was authored inside a
-- $prompt$…$prompt$ dollar-quoted string, where '' is NOT an escape — it is
-- two literal apostrophes. The author wrote '' out of single-quote habit, so
-- the LIVE prompt says "site''s", "operator''s", "reader''s" etc.
-- [MEASURED 2026-08-15]: 10 doubled pairs, 0 legitimate lone apostrophes
-- (20 apostrophe characters total, all in pairs), in both the file and the
-- live row (byte-equal, 5,494 chars each).
--
-- Why it is worth a migration and not a shrug: both live runs handled it fine,
-- but `lead_with[].point` is DESIGNED to be a sentence a page could open with,
-- and the copy_quality_two_stage lane is invited to feed it toward writers —
-- a model mimicking the doubled style would put "site''s" one hop from
-- production copy. Ten characters of fix closes that path.
--
-- 408 itself is NOT edited: it is applied and ledger-recorded, and its
-- checksum must keep matching the ledger. This file is the correction.
--
-- Deliberately NOT re-proven with a live LLM run: the edit removes ten
-- characters and changes no instruction; the verify block re-asserts every
-- load-bearing line 408's own guards check (none contains an apostrophe, so
-- the replace cannot touch them), and a re-fire would file ~5 more
-- non-parkable work items on some site for no information.
--
-- ROLLBACK (421_…_ROLLBACK.sql): exact, because lone-apostrophe count is 0 —
-- re-double every apostrophe and you have 408's byte-identical prompt back.

BEGIN;

DO $guard$
DECLARE
  p text;
  pairs int;
BEGIN
  SELECT default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
    INTO p
  FROM agent_definitions
  WHERE type = 'offer-analyser' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF p IS NULL THEN
    RAISE EXCEPTION 'guard: no live offer-analyser prompt (408 not applied, or the agent was retired)';
  END IF;

  pairs := (length(p) - length(replace(p, $$''$$, ''))) / 2;
  IF pairs = 0 THEN
    RAISE EXCEPTION 'guard: 0 doubled pairs — already applied';
  END IF;
  IF pairs <> 10 OR length(p) <> 5494 THEN
    RAISE EXCEPTION 'guard: expected the 408 prompt exactly (10 pairs, 5494 chars), found % pairs, % chars — the prompt has been edited since; re-derive this fix rather than applying it blind', pairs, length(p);
  END IF;
END $guard$;

UPDATE agent_definitions SET
  default_config = jsonb_set(
    default_config,
    '{workflow,steps,run_offer_analysis,config,prompt}',
    to_jsonb(replace(
      default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
      $$''$$, $$'$$))
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

  IF position($$''$$ in p) > 0 THEN
    RAISE EXCEPTION 'verify: doubled apostrophes survive';
  END IF;
  IF length(p) <> 5484 THEN
    RAISE EXCEPTION 'verify: expected 5484 chars (5494 − 10), found %', length(p);
  END IF;
  -- 408's own load-bearing guards, re-asserted after the edit:
  IF p NOT LIKE '%HONESTY CONSTRAINT%' OR p NOT LIKE '%NO visitor behaviour data%' THEN
    RAISE EXCEPTION 'verify: the honesty constraint was damaged by the replace';
  END IF;
  IF p NOT LIKE '%gap, content, differentiation, structure, cta, nav_restructure, tone%' THEN
    RAISE EXCEPTION 'verify: the closed category vocabulary was damaged by the replace';
  END IF;
  IF p NOT LIKE '%inputs_missing%' OR p NOT LIKE '%degraded%' THEN
    RAISE EXCEPTION 'verify: the degradation instruction was damaged by the replace';
  END IF;
END $verify$;

COMMIT;
