-- 596_claude_probe_interval_60s.sql
--
-- bugs_open/243-anthropic-cap. Council-Reviewed: 82f07fa6-1c42-46ad-bdf6-1d58892c44a7
--
-- Cuts the claude endpoint's re-probe interval from 3600s to 60s.
--
-- SPLIT OUT of 588_council_seat_transient_costs_one_seat_HOLD.sql on 2026-08-24, at the
-- owner's instruction to apply this half now. The split is MECHANICAL — this statement and
-- its guard are byte-identical to the ones the council reviewed as part of that file's edit;
-- no new behaviour, hence the Council-Reviewed trailer above rather than a fresh round.
--
-- WHY THE SPLIT WAS NEEDED, and it is not tidiness: 588 is `_HOLD` because its OTHER half
-- (repointing the council seats' error_step) is unsafe until the `__step_errors` WRITER is
-- live in the chassis, and that writer is still uncommitted — blocked on the bugs_open/354
-- lane's untracked callee. Probed on v1.0.1334, both replicas, 2026-08-24:
-- `step-error record capped at` = 0 (writer ABSENT), `__step_errors` = 1 (reader present and
-- failing closed). This half has NO such dependency, so holding it hostage to the other
-- would have left a measured 60-minute fleet stall in place for no reason.
--
-- WHAT IT BUYS: `claim_work_item` gates EVERY work-item claim fleet-wide on
-- ai_endpoint_health.healthy, and until chassis v1.0.1334 the only writer of `true` was this
-- probe. One refused LLM call therefore stopped ALL dispatch until the next tick — measured
-- 60m25s on 2026-08-17 while 93 of 99 live Anthropic calls in the same window succeeded.
-- 3600s -> 60s bounds that worst case to a MINUTE OR TWO — see the correction below.
--
-- ⚠ CORRECTED 2026-08-24, same day, AFTER MEASURING (the SQL applied is unchanged; only this
-- claim was wrong). "About a minute" overstates it. The probe fires when the scheduled task
-- `ai-endpoint-health-check` ticks AND the endpoint's own check_interval_seconds has elapsed
-- since last_checked. That task's own interval_seconds is ALSO 60, so the two compose and the
-- observed cadence is ~90s, not 60s: measured ticks 16:22:38 -> 16:24:12 -> 16:25:44, gaps of
-- 94s and 92s. So the honest bound is roughly ONE TO TWO MINUTES, phase-dependent — still a
-- ~39x improvement on 3600s, which is the point, but do not quote "one minute".
-- If a tighter bound is ever wanted, the lever is this row at 30s (the value gpu-ollama
-- carries) against the task's 60s tick, NOT lowering this further on its own.
--
-- 60 is not a number picked for this migration: it is the value the `cpu-ollama` row in this
-- same table already carries, so it is an established setting for this mechanism.
--
-- ⚠ THIS BOUNDS THE DAMAGE; IT DOES NOT REMOVE THE MECHANISM, and it is NOT the main fix.
-- `pingClaude` returns healthy for ANY non-auth status (its `default:` arm) and an Anthropic
-- usage cap arrives as a 400 — so for this condition the probe is a TIMER that clears the
-- flag, not a health check, and it reports healthy whether or not we are still being refused.
-- The real removal is the symmetric writer that shipped in chassis v1.0.1334 (register
-- MDL-044): a successful live call now clears the flag, which is a signal that can actually
-- observe recovery. This migration is the backstop for the case where there is no traffic
-- at all to do the clearing.
--
-- COST, stated rather than assumed: 1,440 probe calls/day instead of 24, each
-- `claude-haiku-4-5-20251001` with `max_tokens: 1`, against ~750-1,200 real fleet calls/day.
-- [UNMEASURED] as a billing figure.

BEGIN;

UPDATE ai_endpoint_health
   SET check_interval_seconds = 60, updated_at = NOW()
 WHERE name = 'claude';

-- Verify INSIDE the transaction and RAISE. A block of bare SELECTs cannot stop a COMMIT —
-- ON_ERROR_STOP ignores a non-empty result set.
DO $$
DECLARE
  n int;
  iv int;
BEGIN
  SELECT count(*) INTO n FROM ai_endpoint_health WHERE name = 'claude';
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly one claude row in ai_endpoint_health, found %', n;
  END IF;

  SELECT check_interval_seconds INTO iv FROM ai_endpoint_health WHERE name = 'claude';
  IF iv IS DISTINCT FROM 60 THEN
    RAISE EXCEPTION 'claude check_interval_seconds is %, expected 60', iv;
  END IF;
END $$;

COMMIT;
