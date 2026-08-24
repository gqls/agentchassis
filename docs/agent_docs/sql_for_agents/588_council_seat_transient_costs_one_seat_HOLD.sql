-- 588_council_seat_transient_costs_one_seat_HOLD.sql
--
-- bugs_open/243-anthropic-cap. Council-Reviewed: 82f07fa6-1c42-46ad-bdf6-1d58892c44a7
--
-- WHAT: (a) a council review seat whose LLM call errors costs ONE SEAT instead of
-- the whole round; (b) the claude endpoint is re-probed every 60s instead of every
-- 3600s.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- ⚠ _HOLD — DO NOT LET THE MIGRATION RUNNER APPLY THIS. APPLY BY HAND, AND ONLY
--   AFTER THE GO HALF IS LIVE ON BOTH REPLICAS. THE ORDERING IS THE SAFETY
--   PROPERTY, NOT PAPERWORK.
--
-- Why: repointing error_step makes a failed seat's result field merely ABSENT,
-- and diagnose_council_decide counts an absent field as an ABSTENTION. Its own
-- comment (diagnose_council_decide_action.go, the raw == nil branch) says why
-- that is wrong: "an abstention is a seat the relevance filter skipped, which is
-- information; an unreadable seat is an opinion we were owed and lost...
-- Conflating them would let a lost opinion read as a considered non-objection."
-- An `unreadable` seat DOWNGRADES an approval to revise; an abstention does not.
--
-- So applying this BEFORE the Go half is live does not merely fail to help — it
-- converts a LOUD failure (the round dies, you resubmit) into a SILENT one (the
-- round approves with a seat nobody heard from). That is a worse state than the
-- bug this fixes.
--
-- THE GATE, and run BOTH controls in the same breath — a grep that matches
-- everything proves nothing (LANDMINES: "a control that matches everything"):
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--   kubectl -n ai-persona-system exec $POD -- grep -ac __step_errors /proc/1/exe   # MUST be >= 1
--   kubectl -n ai-persona-system exec $POD -- grep -ac __step_errors_absent_control /proc/1/exe  # MUST be 0
-- Repeat for EVERY replica, not one: a partial roll leaves the old binary serving
-- some rounds. Then apply this file by hand and record it with --record-only.
-- ─────────────────────────────────────────────────────────────────────────────

BEGIN;

-- (a) Each seat's error_step becomes that seat's OWN next_step, so a failed seat
-- is skipped and the chain continues to the next reviewer.
--
-- The filter is on error_step='complete_invalid' with the two terminals named as
-- exceptions, NOT on a `review_%` name pattern. That is deliberate and it is the
-- editquality seat's objection on this round: a name filter would silently leave
-- a future gate_* seat routing to complete_invalid, which is the exact failure
-- this migration exists to close. Measured 2026-08-24: the 19 steps carrying
-- error_step='complete_invalid' are the 17 review_* seats plus exactly these two.
--
-- persist_submission and council_decide KEEP complete_invalid on purpose: if the
-- submission cannot be persisted there is nothing to review, and if aggregation
-- fails there is no verdict. Neither is a reviewer's opinion.
UPDATE agent_definitions ad
SET default_config = jsonb_set(
      ad.default_config,
      ARRAY['workflow', 'steps', s.k, 'config', 'error_step'],
      to_jsonb(s.v->>'next_step')
    ),
    updated_at = NOW()
FROM LATERAL jsonb_each(ad.default_config #> '{workflow,steps}') AS s(k, v)
WHERE ad.type = 'council-gate'
  AND ad.is_active
  AND COALESCE(ad.is_snapshot, false) = false
  AND ad.deleted_at IS NULL
  AND s.v->'config'->>'error_step' = 'complete_invalid'
  AND s.k NOT IN ('persist_submission', 'council_decide')
  AND COALESCE(s.v->>'next_step', '') <> '';

-- (b) The claude endpoint's re-probe interval. 3600s meant one refused call
-- stopped fleet-wide work-item dispatch for up to an hour (measured 60m25s on
-- 2026-08-17). 60s is the value the cpu-ollama row in this same table already
-- carries, so it is an established setting for this mechanism rather than a
-- number picked for this migration.
--
-- NOTE this bounds the damage; it does not remove the mechanism. The prober
-- CANNOT detect a usage cap at all — pingClaude returns healthy for any non-auth
-- status and the cap is a 400 — so for this condition it is a timer that clears
-- the flag, not a health check. The removal is the Go half's symmetric writer.
UPDATE ai_endpoint_health
   SET check_interval_seconds = 60, updated_at = NOW()
 WHERE name = 'claude';

-- Verify INSIDE the transaction. A block of bare SELECTs cannot stop a COMMIT —
-- ON_ERROR_STOP ignores a non-empty result — so this must RAISE.
DO $$
DECLARE
  still_invalid int;
  probe_interval int;
BEGIN
  SELECT count(*) INTO still_invalid
    FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k, v)
   WHERE type = 'council-gate' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND v->'config'->>'error_step' = 'complete_invalid'
     AND k NOT IN ('persist_submission', 'council_decide');
  IF still_invalid > 0 THEN
    RAISE EXCEPTION '% reviewer seat(s) still route error_step to complete_invalid', still_invalid;
  END IF;

  -- The two terminals must NOT have been swept up.
  IF (SELECT count(*) FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k, v)
       WHERE type = 'council-gate' AND is_active
         AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
         AND k IN ('persist_submission', 'council_decide')
         AND v->'config'->>'error_step' = 'complete_invalid') <> 2 THEN
    RAISE EXCEPTION 'persist_submission/council_decide must KEEP complete_invalid — one of them was rewritten';
  END IF;

  SELECT check_interval_seconds INTO probe_interval FROM ai_endpoint_health WHERE name = 'claude';
  IF probe_interval IS DISTINCT FROM 60 THEN
    RAISE EXCEPTION 'claude check_interval_seconds is %, expected 60', probe_interval;
  END IF;
END $$;

COMMIT;
