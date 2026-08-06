-- SQL_2026-08-06b_classifier_cap_32000_to_64000.sql
-- bugs_open/183 — raise domain-research-classifier/classify_and_extract 32000 -> 64000.
-- OWNER-DIRECTED, 2026-08-06. Live config: takes effect immediately, no image rebuild.
--
-- WHY 64000 AND NOT 128000 (the model ceiling for claude-sonnet-4-6):
--  * The chassis does NOT stream. `platform/aiservice/anthropic.go:72` uses a single
--    http.Client with a 600s timeout, and its own comment records having hit
--    "Client.Timeout exceeded" at ~600,0xx ms. A cap the model can actually fill is
--    therefore bounded by wall-clock, not just by the API's limit.
--  * 64000 is ALREADY LIVE and exercised on this chassis: tool-recreation-handler/
--    recreate_tool, 77 calls in 90 days, peak output 11,888. So this is an existing
--    operating point, not a new one.
--  * 128000 would make this step the fleet's ONLY 128000 — the exact shape that made
--    this bug invisible for months (it was the fleet's only 6000). A singleton cap has
--    no sibling to compare against and nothing to notice when it drifts.
--
-- HEADROOM AFTER THIS CHANGE: observed max output on this step is 6,590 tokens
-- (2026-08-02, at the 16000 cap). 64000 is ~9.7x that. Regrowth toward the ceiling is
-- now ANNOUNCED rather than discovered by a burned site — `fleet-step-token-pressure`
-- (LCO-007) flags this step at p95 >= 85% or peak >= 95% of whatever cap is live.
--
-- SHADOWING CHECKED, NOT ASSUMED: this agent has no root `ai_service` block, so the
-- step-level max_tokens is the live value (MDL-039 / bugs_open/009 — a root block makes
-- step-level max_tokens dead config, and the read below would then be reporting a
-- number that never reaches the API). Re-verified in the same transaction.
--
-- The UPDATE is guarded on the current value so it cannot silently overwrite another
-- session's concurrent change, and jsonb_set touches ONLY this path.

BEGIN;

-- Guard 1: refuse if a root ai_service block exists (step value would be shadowed).
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type='domain-research-classifier' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND (default_config #> '{ai_service}') IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'ABORT: root ai_service block present — step-level max_tokens is shadowed (bugs_open/009). Fix the shadowing first.';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,classify_and_extract,config,ai_service,max_tokens}',
         '64000'::jsonb,
         false),           -- false: do NOT create the key if absent; a missing key means
                           -- the path is wrong and we must not invent config.
       updated_at = now()
 WHERE type='domain-research-classifier' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,classify_and_extract,config,ai_service,max_tokens}' = '32000';

-- Guard 2: exactly one row must have moved to 64000. A verify block of bare SELECTs
-- cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty result) — so RAISE.
DO $$
DECLARE v text;
BEGIN
  SELECT default_config #>> '{workflow,steps,classify_and_extract,config,ai_service,max_tokens}'
    INTO v FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v IS DISTINCT FROM '64000' THEN
    RAISE EXCEPTION 'ABORT: cap is % after update, expected 64000 (was it already changed by another session?)', COALESCE(v,'NULL');
  END IF;
END $$;

COMMIT;

-- Verify live (run separately; the step block is the live one only because guard 1 passed):
--   SELECT (default_config #> '{ai_service}') AS root_block_must_be_null,
--          default_config #>> '{workflow,steps,classify_and_extract,config,ai_service,max_tokens}' AS live_cap
--     FROM agent_definitions WHERE type='domain-research-classifier' AND is_active
--       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- Then confirm at the ARTEFACT, not the config: the next real run must log max_tokens=64000.
--   SELECT created_at, max_tokens, output_tokens, success FROM llm_call_log
--    WHERE step_name='classify_and_extract' ORDER BY created_at DESC LIMIT 3;
-- Until such a row exists, this is a config claim, not a proven one.
