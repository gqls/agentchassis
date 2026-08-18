-- ═══════════════════════════════════════════════════════════════════════════
-- ⛔ SUPERSEDED 2026-08-18 — DO NOT RUN. This is a DUPLICATE of
--    464_reaper_initialized_arm.sql, which is what actually shipped.
--
-- Two lanes worked the bugs_open/310 INITIALIZED gap in parallel without seeing
-- each other. 464 was applied by hand at 18:43:03Z, owner-approved, and is the
-- row recorded in schema_migrations; the reaper fired on its own tick at
-- 18:45:53Z and failed both stranded rows. INITIALIZED is now 0 fleet-wide.
--
-- The two files' payloads are BYTE-IDENTICAL (md5 421dfc4f9f74035d71f43a33c703cd44,
-- 3,068 bytes), so this file is redundant rather than wrong. Renamed with an
-- uppercase suffix so the migration runner's SIDECAR_RE ('_[A-Z][A-Z0-9_]*\.sql$',
-- run-migrations.sh:65) excludes it from --apply while still listing it. It is NOT
-- in schema_migrations. If it were ever run, GUARD 1 would take the
-- "already present — this run is a no-op" branch.
--
-- ⚠ AND ITS REASONING IS PARTLY WRONG — see bugs_open/310 "CORRECTION 2".
-- The header below argues that 463's licence transfers because INITIALIZED, like
-- RUNNING, is an inter-step transition of milliseconds. It does NOT transfer.
-- RUNNING is transient by construction; INITIALIZED is a genuine WAITING state
-- that sits on a Kafka queue, so its duration is measurable, not derivable — and
-- measured it is avg 0.22s / p99 2.01s / MAX 6.31s over 5,736 rows, not
-- milliseconds. 4h remains safe (~2,280x the observed max) but is licensed by that
-- measurement. Re-run the p99 query, do not re-read coordinator.go.
-- ═══════════════════════════════════════════════════════════════════════════

-- 470 — the stale-orchestration reaper gains an INITIALIZED arm.
--
-- bugs_open/310: a row that stops in `INITIALIZED` is unreachable by EVERY
-- automated path — nothing reaps it and nothing prunes it — so it persists for
-- ever. This is bugs_closed/294 one status over; 463 closed RUNNING and did not
-- touch this. Unlike RUNNING it pins no Kafka topics, so it is NOT a bugs_open/240
-- contributor: the harm is a permanent row, one silently-lost unit of work per
-- occurrence, and a status invisible to every operator surface.
--
-- THE THREE LOCKS, all verified in the code at HEAD on 2026-08-18:
--   1. THE REAPER has arms for AWAITING_RESPONSES (30/90 min), EXECUTING_STEP
--      (4 h) and RUNNING (4 h, from 463). No INITIALIZED arm.  <-- this file
--   2. THE PRUNER (`database-cleanup`) deletes COMPLETED/FAILED >24 h and
--      EXECUTING_STEP/AWAITING_RESPONSES >4 h. INITIALIZED is in neither clause,
--      which is why a row survives from 2026-07-13.
--   3. handleOrchestrationStatus (coordinator.go:741) HAS a `case
--      StatusInitialized` and it works — but it runs only when an inbound Kafka
--      message arrives for that orchestration, and for a stranded row none ever
--      will. The reader is live and undriven, not missing.
--   (+ TimeoutMonitor is dormant estate-wide — NewOrchestratorHelper
--      (helpers.go:581) has ZERO callers, and it is NewTimeoutMonitor's only
--      caller (helpers.go:587). Re-verified by grep here, not cited from 294.
--      Scope: "no caller in this repository", not "uncallable".)
--
-- WHY A 4 h THRESHOLD IS SAFE — the licence is the CODE, not a census.
-- Same reasoning 463 settled on, and for the same reason: a census of a status
-- this rare reads ~0 in every band and so cannot discriminate "nothing healthy
-- lives here" from "I sampled at a quiet moment".
--   * INITIALIZED has EXACTLY ONE LIVE WRITER fleet-wide — state.go:734, the
--     $11 status parameter of StateRepository.CreateInitialState. (grep
--     StatusInitialized across the repo returns four hits: the constant, that
--     writer, the coordinator.go:741 reader, and helpers.go:615 — which is
--     inside the dormant OrchestratorHelper and never fires.)
--   * That writer is reached from coordinator.go:149 (getOrCreateState), and
--     the VERY NEXT THING its caller does, at coordinator.go:165, is
--     handleOrchestrationStatus -> case StatusInitialized -> SetExecutingStep,
--     flipping the row to EXECUTING_STEP.
--   So the INITIALIZED window is: one GetState, three log lines and a map
--   assignment, in ONE process inside ONE message handler. MILLISECONDS. It is
--   an inter-step transition, never a durable healthy state. 4 h is ~7 orders of
--   magnitude of headroom.
--
-- Kept at 4 h (not tighter) to match the sibling failed_wedged and failed_running
-- arms and to stay conservative on live config that takes effect on save.
--
-- WHY THE REAPER AND NOT A GO FIX: the producer is a process dying inside that
-- window, so no in-process recovery can exist — the process that would do it is
-- gone. An external sweeper is the correct primary fix, not merely the cheapest.
--
-- Statuses are disjoint, so no NOT IN exclusion against the sibling CTEs is needed.
--
-- Idempotent: GUARD 1 makes a replay after a direct psql apply a declared no-op.
-- Rollback sidecar: 470_reaper_initialized_arm_ROLLBACK.sql.
--
-- BOTH GUARDS ARE PRESENT IN THIS FILE FROM THE START, unlike 463 — whose council
-- round (corr 860d87d9) gated at HIGH because substring assertions prove a needle
-- is present and prove NOTHING about whether the assembled SQL parses, and whose
-- editquality seat then logged the advisory that 463 itself still lacked the
-- check. A stored pre_query parses only when the task next ticks, so a typo
-- commits happily and takes out ALL FIVE arms minutes later with no earlier
-- signal. House idiom: 458_..._ROLLBACK.sql (md5 branch),
-- 210_report_pipeline_scheduled_tasks.sql (EXECUTE check). The one deliberate
-- variation from 210: its gate SELECTs are inert, whereas this pre_query MUTATES,
-- so the sentinel raise below discards the effects.
--
-- PROVEN IN BOTH DIRECTIONS before this file was written (2026-08-18): the new
-- text passed the EXECUTE check, a corrupted copy ('failed_initialized AS ((((')
-- was CAUGHT with `syntax error at or near "UPDATE"`, and a control confirmed the
-- two live INITIALIZED rows were untouched by either run.

BEGIN;

-- GUARD 1 (pre-flight): three-way on md5 of the TEXT. scheduled_tasks is edited
-- by more than one lane and this is a full-text rewrite, so an ungated UPDATE
-- silently CLOBBERS whatever another session put there. Do NOT gate on
-- updated_at: measured 2026-08-18 it moves for scheduler stamping with pre_query
-- unchanged, so it answers a different question than the one being asked.
DO $do$
DECLARE live_md5 text;
BEGIN
  SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 IS NULL THEN
    RAISE EXCEPTION '310/470: no stale-orchestration-reaper row exists.';
  ELSIF live_md5 = '91ba970443ff2d237b53633d80e20904' THEN
    RAISE NOTICE '310/470: live text is 463 (RUNNING arm) — adding the INITIALIZED arm.';
  ELSIF live_md5 = '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE NOTICE '310/470: INITIALIZED arm already present — this run is a no-op.';
  ELSE
    RAISE EXCEPTION '310/470 REFUSED: the reaper pre_query is neither 463''s text nor 470''s (live md5 %). Another edit has landed; blind rewriting would revert it. Re-capture the live text and rebuild this migration by hand.', live_md5;
  END IF;
END
$do$;

UPDATE scheduled_tasks
   SET pre_query = $PQ$
    WITH failed_dispatch AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: dispatch loop idle for >30 min',
        updated_at = NOW()
    WHERE status = 'AWAITING_RESPONSES'
      AND owner_agent_type = 'build-dispatch-loop'
      AND last_activity < NOW() - INTERVAL '30 minutes'
    RETURNING orchestration_id, owner_agent_type, current_step
),
failed_orchs AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: stale AWAITING_RESPONSES for >90 min',
        updated_at = NOW()
    WHERE status = 'AWAITING_RESPONSES'
      AND last_activity < NOW() - INTERVAL '90 minutes'
      AND orchestration_id NOT IN (SELECT orchestration_id FROM failed_dispatch)
    RETURNING orchestration_id, owner_agent_type, current_step
),
failed_wedged AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: stale EXECUTING_STEP for >4h; step=' || COALESCE(current_step, '(none)'),
        updated_at = NOW()
    WHERE status = 'EXECUTING_STEP'
      AND last_activity < NOW() - INTERVAL '4 hours'
    RETURNING orchestration_id, owner_agent_type, current_step
),
failed_running AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: stale RUNNING for >4h; step=' || COALESCE(current_step, '(none)'),
        updated_at = NOW()
    WHERE status = 'RUNNING'
      AND last_activity < NOW() - INTERVAL '4 hours'
    RETURNING orchestration_id, owner_agent_type, current_step
),
failed_initialized AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: stale INITIALIZED for >4h; step=' || COALESCE(current_step, '(none)'),
        updated_at = NOW()
    WHERE status = 'INITIALIZED'
      AND last_activity < NOW() - INTERVAL '4 hours'
    RETURNING orchestration_id, owner_agent_type, current_step
),
reset_tasks AS (
    SELECT * FROM business_intel.reap_stale_collection_tasks()
),
expired_awaited AS (
    UPDATE awaited_requests
    SET status = 'expired',
        processed_at = NOW()
    WHERE status = 'waiting'
      AND timeout_at < NOW() - INTERVAL '5 minutes'
    RETURNING request_id
)
SELECT
    (SELECT COUNT(*) FROM failed_dispatch)::text as dispatch_failed,
    (SELECT COUNT(*) FROM failed_orchs)::text as orchs_failed,
    (SELECT COUNT(*) FROM failed_wedged)::text as wedged_failed,
    (SELECT COUNT(*) FROM failed_running)::text as running_failed,
    (SELECT COUNT(*) FROM failed_initialized)::text as initialized_failed,
    (SELECT reset_count FROM reset_tasks)::text as tasks_reset,
    (SELECT parked_count FROM reset_tasks)::text as tasks_parked,
    (SELECT COUNT(*) FROM expired_awaited)::text as awaited_expired
HAVING
    (SELECT COUNT(*) FROM failed_dispatch) > 0
    OR (SELECT COUNT(*) FROM failed_orchs) > 0
    OR (SELECT COUNT(*) FROM failed_wedged) > 0
    OR (SELECT COUNT(*) FROM failed_running) > 0
    OR (SELECT COUNT(*) FROM failed_initialized) > 0
    OR (SELECT reset_count + parked_count FROM reset_tasks) > 0
    OR (SELECT COUNT(*) FROM expired_awaited) > 0
$PQ$,
       updated_at = now()
 WHERE name = 'stale-orchestration-reaper'
   -- 0 rows on the already-applied path, which GUARD 1 has declared benign.
   AND md5(pre_query) = '91ba970443ff2d237b53633d80e20904';

-- ── GUARD 2 + verify (DO/RAISE — a SELECT cannot stop a COMMIT) ─────────────
DO $do$
DECLARE q text; live_md5 text;
BEGIN
  SELECT md5(pre_query), pre_query INTO live_md5, q FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 <> '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE EXCEPTION '310/470: post-update text is not byte-exact (got %, want 421dfc4f9f74035d71f43a33c703cd44). Do NOT commit it.', live_md5;
  END IF;

  -- the new arm is wired in all three places it has to be
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query LIKE '%failed_initialized AS (%'
         AND pre_query LIKE '%stale INITIALIZED for >4h%'
         AND pre_query LIKE '%as initialized_failed%'
         AND pre_query LIKE '%OR (SELECT COUNT(*) FROM failed_initialized) > 0%') <> 1 THEN
    RAISE EXCEPTION '310/470: failed_initialized arm not fully wired (CTE / SELECT / HAVING)';
  END IF;

  -- NEGATIVE CONTROL: every pre-existing arm must survive. Without this, a
  -- rewrite that dropped them would pass the assertion above.
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query LIKE '%dispatch loop idle for >30 min%'
         AND pre_query LIKE '%stale AWAITING_RESPONSES for >90 min%'
         AND pre_query LIKE '%stale EXECUTING_STEP for >4h%'
         AND pre_query LIKE '%stale RUNNING for >4h%'
         AND pre_query LIKE '%reap_stale_collection_tasks()%'
         AND pre_query LIKE '%expired_awaited AS (%') <> 1 THEN
    RAISE EXCEPTION '310/470: a pre-existing reaper arm was lost in the rewrite';
  END IF;

  -- GUARD 2: the stored text must actually PARSE AND EXECUTE, not merely contain
  -- the right substrings. Effects discarded by the sentinel raise.
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '310/470: stored pre_query does NOT execute (%) — a task with a broken gate never fires and merely looks idle. Aborting.', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '310/470: INITIALIZED arm wired, all prior arms intact, pre_query proven to execute.';
END
$do$;

COMMIT;
