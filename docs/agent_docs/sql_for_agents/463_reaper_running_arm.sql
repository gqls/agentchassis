-- 463 — the stale-orchestration reaper gains a RUNNING arm.
--
-- bugs_open/294: a row that stops in `RUNNING` is unreachable by EVERY recovery
-- path, and pins two Kafka topics for ever (a direct contributor to
-- bugs_open/240, the kafka-scheduler OOM). Three independent locks on the door,
-- all verified in the code at HEAD on 2026-08-18:
--
--   1. THE REAPER has arms for AWAITING_RESPONSES (30 min / 90 min) and
--      EXECUTING_STEP (4 h) only. No RUNNING arm.  <-- this file closes it
--   2. TimeoutMonitor (platform/orchestration/helpers.go) keys entirely on
--      entries in AwaitedRequests — MonitorRequest is started per awaited
--      request, and every handler indexes state.AwaitedRequests[requestID]. A
--      row with none has no monitor goroutine at all, at any age.
--   3. handleOrchestrationStatus (coordinator.go:740-796) has NO `case
--      StatusRunning`. A message arriving for such a row falls to
--      `default: return fmt.Errorf("unknown orchestration status: %s", ...)`,
--      so even message-driven recovery actively rejects it.
--   (+ failWorkflow (coordinator.go:3718) declines to fail a RUNNING row on the
--      optimistic-lock path — correct for a race measured in seconds, and it
--      also means the path that noticed the error will not clear it.)
--
-- WHY A 4 h THRESHOLD IS SAFE — the licence is the CODE, not a census.
-- bugs_open/294 licensed its threshold with an age census (49 rows, all >4h, 0
-- in every younger band). Re-run 2026-08-18 before writing this file, that
-- census returns 0 rows in EVERY band including >4h — the 289 sweep cleared it
-- and 289's fix stopped the producer. A census that reads 0 everywhere cannot
-- discriminate "nothing healthy lives in RUNNING" from "I sampled at a quiet
-- moment", so it no longer licenses anything and is NOT the evidence here.
--
-- The code does, and it does not expire:
--   * `RUNNING` is written on EXACTLY ONE LINE fleet-wide —
--     state.go:1428, in StateRepository.ClearExecutingStep, which flips
--     EXECUTING_STEP -> RUNNING. (grep StatusRunning across platform/ internal/
--     pkg/ returns one assignment; the only other Go/SQL `'RUNNING'` literals
--     are READERS — monitoring.go, cleanup_stale_topics.go, the two admin
--     handlers — plus the unrelated Thunder vendor enum.)
--   * ClearExecutingStep has exactly ONE caller — coordinator.go:765, the
--     stuck-orchestration takeover — and the next thing that caller does is
--     GetState then continueExecution.
--   * continueExecution's main loop opens with SetExecutingStep
--     (coordinator.go:868), flipping the row straight back to EXECUTING_STEP.
--   So the RUNNING window is: one GetState + a circuit-breaker check + a
--   max-age check. MILLISECONDS. It is an inter-step transition, never a
--   durable healthy state. 4 h is ~7 orders of magnitude of headroom.
--
-- Kept at 4 h (not tighter) to match the sibling `failed_wedged` arm and to stay
-- deliberately conservative on live config that takes effect the instant it is
-- saved. The code above would licence minutes if topic pressure ever demands it.
--
-- WHY THE REAPER AND NOT A GO FIX: if the pod dies between ClearExecutingStep
-- and SetExecutingStep, no in-process recovery can exist — the process that
-- would do it is gone. An external sweeper is the only thing that can reach this
-- state, so this is the correct primary fix, not merely the cheapest one.
--
-- Statuses are disjoint, so no NOT IN exclusion against the sibling CTEs is
-- needed (failed_dispatch/failed_orchs are AWAITING_RESPONSES, failed_wedged is
-- EXECUTING_STEP).
--
-- Idempotent: a same-text UPDATE, so a migration-runner replay after a direct
-- psql apply is a no-op. Rollback sidecar: 463_reaper_running_arm_ROLLBACK.sql.

BEGIN;

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
    (SELECT reset_count FROM reset_tasks)::text as tasks_reset,
    (SELECT parked_count FROM reset_tasks)::text as tasks_parked,
    (SELECT COUNT(*) FROM expired_awaited)::text as awaited_expired
HAVING
    (SELECT COUNT(*) FROM failed_dispatch) > 0
    OR (SELECT COUNT(*) FROM failed_orchs) > 0
    OR (SELECT COUNT(*) FROM failed_wedged) > 0
    OR (SELECT COUNT(*) FROM failed_running) > 0
    OR (SELECT reset_count + parked_count FROM reset_tasks) > 0
    OR (SELECT COUNT(*) FROM expired_awaited) > 0
$PQ$,
       updated_at = now()
 WHERE name = 'stale-orchestration-reaper';

-- ── Verify (DO/RAISE — a SELECT cannot stop a COMMIT) ───────────────────────
DO $do$
BEGIN
  -- the task exists and was actually rewritten
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper') <> 1 THEN
    RAISE EXCEPTION '294/463: stale-orchestration-reaper row missing';
  END IF;

  -- the new arm is present, in all three places it has to be
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query LIKE '%failed_running AS (%'
         AND pre_query LIKE '%stale RUNNING for >4h%'
         AND pre_query LIKE '%as running_failed%'
         AND pre_query LIKE '%OR (SELECT COUNT(*) FROM failed_running) > 0%') <> 1 THEN
    RAISE EXCEPTION '294/463: failed_running arm not fully wired (CTE / SELECT / HAVING)';
  END IF;

  -- NEGATIVE CONTROL: the pre-existing arms must all survive. Without this a
  -- clumsy rewrite that dropped them would pass the assertion above.
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query LIKE '%dispatch loop idle for >30 min%'
         AND pre_query LIKE '%stale AWAITING_RESPONSES for >90 min%'
         AND pre_query LIKE '%stale EXECUTING_STEP for >4h%'
         AND pre_query LIKE '%reap_stale_collection_tasks()%'
         AND pre_query LIKE '%expired_awaited AS (%') <> 1 THEN
    RAISE EXCEPTION '294/463: a pre-existing reaper arm was lost in the rewrite';
  END IF;
END
$do$;

COMMIT;
