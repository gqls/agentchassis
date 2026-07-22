-- PATCH_feature_designer_022_council_to_sonnet5.sql
-- 2026-07-22 (chat "diagnosis fixloop 5")
--
-- feature-designer's 5 reviewer seats were left on claude-sonnet-4-6 @ 3000 when
-- the worker steps (design / reframe / repropose) were moved to sonnet-5 @ 16000
-- in the D1 model migration -- a PARTIAL migration: the design generation was
-- upgraded, its own review council was not. Surfaced by
-- 102_LINT_council_seat_parity.py (a within-family-uniform divergence it does not
-- flag, but a cross-council question it reports). fix-proposer and council-gate
-- run their reviewer seats on sonnet-5 @ 8000; this brings feature-designer's
-- council to that same reviewer standard (16000 is the design step's ceiling, not
-- a reviewer's).
--
-- Surgical: only ai_service.model + ai_service.max_tokens on the 5 review_ seats.
-- provider / api_key_env_var and tolerate_truncation are untouched (all 5 seats
-- already carry tolerate_truncation=true, so the 019 protection survives the move).
-- Snapshot first (restorable from agent_definitions_backup); FOR UPDATE locks the
-- live row; a per-seat guard skips any seat not currently on sonnet-4-6, so a
-- re-run or a concurrent partial change cannot double-apply or clobber.

BEGIN;

SELECT snapshot_agent('feature-designer',
  'pre-update: council reviewer seats sonnet-4-6@3000 -> sonnet-5@8000 (missed D1 bump; PATCH_022)');

DO $$
DECLARE
  seat  text;
  seats text[] := ARRAY['review_editquality','review_bug_historian','review_guardian',
                        'review_guidelines','review_reuse_agent'];
  cfg   jsonb;
  moved int := 0;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type='feature-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  FOR UPDATE;

  IF cfg IS NULL THEN
    RAISE EXCEPTION 'no live feature-designer row';
  END IF;

  FOREACH seat IN ARRAY seats LOOP
    IF cfg #>> ARRAY['workflow','steps',seat,'config','ai_service','model'] IS DISTINCT FROM 'claude-sonnet-4-6' THEN
      RAISE NOTICE 'seat % not on claude-sonnet-4-6 (got %) -- skipping',
        seat, cfg #>> ARRAY['workflow','steps',seat,'config','ai_service','model'];
      CONTINUE;
    END IF;
    cfg := jsonb_set(cfg, ARRAY['workflow','steps',seat,'config','ai_service','model'], '"claude-sonnet-5"'::jsonb);
    cfg := jsonb_set(cfg, ARRAY['workflow','steps',seat,'config','ai_service','max_tokens'], '8000'::jsonb);
    moved := moved + 1;
  END LOOP;

  UPDATE agent_definitions SET default_config = cfg, updated_at = now()
  WHERE type='feature-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  RAISE NOTICE 'feature-designer council: % seat(s) moved to claude-sonnet-5 @ 8000', moved;
END $$;

COMMIT;
