-- 637 — dispatch lever: OWNER RULING B (2026-08-26) — interval, not sibling
--
-- AUTHORISATION: owner, 2026-08-26, on the options table in
-- dispatch_throughput/README_where_we_are.md (2026-08-25 entry):
--   "Decision - B as you suggest and we can do C when we have the governor"
-- B = retire the sibling row, set the ORIGINAL row's interval_seconds 60 -> 30.
-- C (interval 25, ~3x) is GATED on the D4 spend governor — do not lower interval
-- below 30 before it exists; the 584 VERIFY asserts this.
--
-- MECHANISM (established 2026-08-25, council-approved r3 corr db9b7cbf; bugs_open/398):
-- the scheduler is fire-and-forget (stampCompleted sets BOTH stamps at fire), so a row's
-- turn rate = interval_seconds rounded up to the 30 s tick, unconditionally. The sibling
-- (584) fired ~1 s after the original and co-picked the SAME site 94% of the time (first
-- claim lands p50 17.7 s after loop spawn), losing 39% of claim attempts for ~+10-15%.
-- Interval 30 gives a fire every ~60 s, spaced >= 30 s: the previous turn's claim is
-- visible (p90 24.2 s), so successive turns steer to DISTINCT sites with full batches.
--
-- EXPECTED EFFECT, MEASURABLE AT THE ARTEFACT (~30 min of data; RUNBOOK "Concurrency
-- meters" section):
--   fire cadence on build-pipeline-trigger: p50 90 s -> ~60 s
--   build-pipeline-trigger-2 last_triggered_at: frozen at the disable time
--   claims lost / (lost+won): 39% -> single digits
--   avg items claimed per loaded loop: 2.6 -> toward 5 (backlog permitting)
--   turn rate: ~80/h (2 rows, phase-locked) -> ~60/h (1 row, spaced)
--
-- The sibling row is DISABLED, not deleted — rollback is one re-enable + one interval
-- restore (637_..._ROLLBACK.sql, both guarded). Its input_data.task_name stays correct,
-- and 583's parameterised stamps mean a re-enabled sibling stamps its own row.
--
-- RERUN-SAFE (582's council advisory, debug_historian 2026-08-25): both UPDATEs are
-- gated on the pre-state, so a replay is a 0-row no-op and the post-check still holds.

BEGIN;

-- Pre-flight: the pair must agree on the dispatch gate before we touch the lever —
-- changing interval on a desynced pair would bake the desync into the surviving row.
DO $$
DECLARE n int;
BEGIN
  SELECT count(DISTINCT (md5(coalesce(pre_query,'')), target_agent_type, target_topic, fire_message))
    INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 1 THEN
    RAISE EXCEPTION '637 pre-flight: trigger rows disagree on gate config (% distinct) — reconcile before moving the lever (LANDMINES 2026-08-25, sibling parity)', n;
  END IF;
END $$;

UPDATE scheduled_tasks SET interval_seconds = 30, updated_at = NOW()
 WHERE name = 'build-pipeline-trigger' AND interval_seconds = 60;

UPDATE scheduled_tasks SET enabled = false, updated_at = NOW()
 WHERE name = 'build-pipeline-trigger-2' AND enabled;

-- Post-check: the sanctioned state under ruling B, asserted so the COMMIT cannot
-- succeed on a partial apply (a block of SELECTs cannot stop a COMMIT — DO/RAISE can).
DO $$
DECLARE n int; iv int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%' AND enabled;
  IF n <> 1 THEN
    RAISE EXCEPTION '637 post-check: % enabled trigger rows, expected exactly 1 (ruling B)', n;
  END IF;
  SELECT interval_seconds INTO iv FROM scheduled_tasks WHERE name = 'build-pipeline-trigger' AND enabled;
  IF iv IS DISTINCT FROM 30 THEN
    RAISE EXCEPTION '637 post-check: enabled row interval_seconds = %, expected 30 (ruling B; < 30 is C and is gated on the D4 governor)', iv;
  END IF;
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'build-pipeline-trigger-2' AND NOT enabled AND input_data->>'task_name' = 'build-pipeline-trigger-2';
  IF n <> 1 THEN
    RAISE EXCEPTION '637 post-check: sibling row missing or identity broken — rollback path would be unsafe';
  END IF;
END $$;

COMMIT;
