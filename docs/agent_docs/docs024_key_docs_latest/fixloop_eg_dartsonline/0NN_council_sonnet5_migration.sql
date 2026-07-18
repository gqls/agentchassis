-- 0NN_council_sonnet5_migration.sql — owner decision D1 (2026-07-18).
-- Move the fix-proposer council (proposer + all reviewers) from
-- claude-sonnet-4-6 to claude-sonnet-5, AND raise reviewer max_tokens to guard
-- against the Sonnet-5 adaptive-thinking-truncation trap.
--
-- PATCH-STYLE, idempotent, DO-loop over steps (survives roster growth — the
-- roster changed 5× in 18h). NEVER a whole-object re-seed (config-clobber
-- finding, multi_session_coordination/FINDING_2026-07-17_config_reseed_clobber).
--
-- WHY max_tokens too (not just the model): on Sonnet 5, omitting `thinking`
-- runs ADAPTIVE THINKING ON (4.6 ran it off), and thinking spends from the same
-- max_tokens budget as the output. Reviewer steps are at max_tokens=3000
-- (verdicts observed 275-1290 tokens at 4.6). A bare model swap would let a
-- reviewer think ~2k then have ~1k for its verdict JSON -> a TRUNCATED verdict
-- (the exact BUG A class, inside the reviewer of the BUG A fix; and CLAUDE.md's
-- `output_tokens == max_tokens means CUT` rule). So reviewers are raised to
-- >=8000. propose/reframe/repropose are already 8000 (observed <=2692 output;
-- 8000 holds thinking + output) — model-only there.
--
-- fix-proposer has NO root ai_service, so per-step ai_service IS read (BUG B —
-- root shadows step — does NOT bite here; the step value takes effect).
--
-- GATE: do NOT hand-patch council-gate. After this, run the roster mirror
--   python3 099_SYNC_gate_roster.py --apply
-- which deep-copies every review_*/gate_* step (config.ai_service included —
-- transform_step only rewrites error_step/input_fields/prompt_template), so the
-- model + max_tokens propagate to the gate's reviewers automatically.
--
-- DB config: LIVE IMMEDIATELY, no image. Back up before applying (below).

BEGIN;

-- Safety backup of the live fix-proposer row (whole object) before the patch.
CREATE TABLE IF NOT EXISTS bak_agentdef_fixproposer_sonnet5_20260718 AS
  SELECT *, now() AS backed_up_at FROM agent_definitions WHERE type='fix-proposer';

DO $do$
DECLARE
  stepname text;
  stepval  jsonb;
  new_model constant jsonb := '"claude-sonnet-5"'::jsonb;
  review_min_max constant int := 8000;
  cur_max int;
BEGIN
  FOR stepname, stepval IN
    SELECT s.key, s.value
    FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') s
    WHERE type='fix-proposer'
      AND s.value #> '{config,ai_service}' IS NOT NULL
  LOOP
    -- 1. model -> sonnet-5 on every LLM step (idempotent).
    UPDATE agent_definitions
      SET default_config = jsonb_set(default_config,
            ARRAY['workflow','steps',stepname,'config','ai_service','model'], new_model)
    WHERE type='fix-proposer'
      AND (default_config #> ARRAY['workflow','steps',stepname,'config','ai_service','model']) <> new_model;

    -- 2. reviewer steps: raise max_tokens to >= review_min_max (never lower).
    IF stepname LIKE 'review\_%' THEN
      cur_max := COALESCE((stepval #>> '{config,ai_service,max_tokens}')::int, 0);
      IF cur_max < review_min_max THEN
        UPDATE agent_definitions
          SET default_config = jsonb_set(default_config,
                ARRAY['workflow','steps',stepname,'config','ai_service','max_tokens'],
                to_jsonb(review_min_max))
        WHERE type='fix-proposer';
      END IF;
    END IF;
  END LOOP;

  UPDATE agent_definitions SET updated_at = now() WHERE type='fix-proposer';
END
$do$;

COMMIT;

-- Verify (run after):
--   SELECT s.key, s.value #>> '{config,ai_service,model}' AS model,
--          s.value #>> '{config,ai_service,max_tokens}' AS max_tokens
--   FROM agent_definitions, LATERAL jsonb_each(default_config #>'{workflow,steps}') s
--   WHERE type='fix-proposer' AND s.value #>'{config,ai_service}' IS NOT NULL
--   ORDER BY s.key;
--   -- every model = claude-sonnet-5; every review_* max_tokens >= 8000.
-- Then: python3 099_SYNC_gate_roster.py --apply   (propagates to council-gate)
-- Then: prove — fire 091 on e505f70f; confirm reviewers log model=claude-sonnet-5,
--       max_tokens=8000, and output_tokens < max_tokens (no truncated verdicts).
