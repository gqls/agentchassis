-- 464 — the stale-orchestration reaper gains an INITIALIZED arm.
--
-- Sibling of 463 (bugs_closed/294), and the SECOND half of the same defect:
-- the reaper's coverage is an ENUMERATION of statuses, not an invariant, so any
-- status nobody listed is never swept — not swept slowly, never. 463 closed
-- RUNNING; this closes INITIALIZED. The class itself is closed only by fix
-- candidate 2 (reap on the invariant), which remains the right long-term design
-- and is NOT what this file does. See LANDMINES.md, "The stale-orchestration
-- reaper enumerates STATUSES…".
--
-- Owner decision 2026-08-18: add the arm now (contained, identical shape to the
-- council-approved 463), and take the invariant rewrite as separate work.
--
-- ── WHY INITIALIZED IS DIFFERENT FROM RUNNING, AND WHY 4 h IS STILL SAFE ─────
-- RUNNING was licensed structurally: one writer, one caller, next write flips it
-- back, so the window is milliseconds BY CONSTRUCTION. **That argument does NOT
-- transfer.** INITIALIZED is a genuine WAITING state — a row is created in it
-- (state.go:734) and leaves it only when its first message is handled
-- (coordinator.go:741, 'case StatusInitialized' -> SetExecutingStep). Its
-- duration is Kafka delivery plus consumer pickup, i.e. it depends on the queue,
-- not on an in-process transition. So it had to be MEASURED, not reasoned about.
--
-- MEASURED 2026-08-18, fleet-wide, over every retained row that has left
-- INITIALIZED. Time in INITIALIZED is (processing_history->0->>'timestamp')
-- minus created_at — the first state_updated entry is the moment it left:
--
--     rows measured        5,736      (all agent types)
--     average                0.22 s
--     p99                    2.01 s
--     MAXIMUM                6.31 s   (page-content-writer)
--     rows over 5 minutes        0
--     rows over 1 hour           0
--     rows over 4 hours          0
--
-- **This could have come out otherwise** — had any agent legitimately waited
-- minutes for its spawn, over_5min would be non-zero. It is zero across 5,736
-- rows and every agent type. A 4 h threshold is ~2,280x the observed maximum.
-- Kept at 4 h to match the sibling arms rather than tuned tighter.
--
-- THE STUCK ROWS' SIGNATURE, for whoever reads a future one: both live examples
-- have last_activity == created_at exactly and processing_history = [] — zero
-- entries. That is a spawn that never arrived (cf. the ~300 s post-restart drop
-- window), not a slow one. A row that merely started slowly would carry history.
--
-- Population at apply time: 2 rows — generic/check_health idle 36.2 days, and
-- endpoint-health-checker/check_health idle 0.4 days AND RISING, i.e. the leak
-- is live, not historical. For scale on how abnormal that is,
-- endpoint-health-checker ran 956 times in the preceding 24 h with an average
-- total lifetime of 1.4 s and a maximum of 3.9 s.
--
-- Unlike RUNNING these rows pin no Kafka topics (INITIALIZED is absent from
-- getActiveOrchestrationTopics's protected set, cleanup_stale_topics.go:207-215),
-- so this does not feed bugs_open/240. The harm is a permanently non-terminal
-- row that masks real work and never leaves the active partial index.
--
-- ── GUARDS (the house idiom; council round 1 on 463 gated at HIGH for want of
-- ── them, and round 2 approved once they existed)
-- GUARD 1, three-way md5 pre-flight  <- 458_detected_item_promoter_..._ROLLBACK.sql
-- GUARD 2, EXECUTE the stored SQL    <- 210_report_pipeline_scheduled_tasks.sql
-- A pre_query is DATA to this migration: LIKE assertions prove a needle is
-- present and prove NOTHING about whether the assembled SQL parses. It parses
-- only when the reaper next ticks, so a typo would commit happily and take out
-- ALL SIX arms minutes later. Variation from 210: its gates are inert SELECTs,
-- the reaper's pre_query MUTATES, so the sentinel raise discards the effects.
--
-- Rollback sidecar: 464_reaper_initialized_arm_ROLLBACK.sql
-- ─────────────────────────────────────────────────────────────────────────────

BEGIN;

-- GUARD 1 (pre-flight): three distinguishable states.
DO $do$
DECLARE live_md5 text;
BEGIN
  SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 IS NULL THEN
    RAISE EXCEPTION '464: no stale-orchestration-reaper row exists.';
  ELSIF live_md5 = '91ba970443ff2d237b53633d80e20904' THEN
    RAISE NOTICE '464: live text is 463 — adding the INITIALIZED arm.';
  ELSIF live_md5 = '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE EXCEPTION '464: already applied (live text is 464) — nothing to do.';
  ELSE
    RAISE EXCEPTION '464 REFUSED: reaper pre_query is neither 463''s text nor 464''s (live md5 %). Another lane has edited it. Re-derive this migration from the live text rather than clobbering their change.', live_md5;
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
   AND md5(pre_query) = '91ba970443ff2d237b53633d80e20904';

-- GUARD 2 + verify (DO/RAISE — a SELECT cannot stop a COMMIT)
DO $do$
DECLARE q text; live_md5 text;
BEGIN
  SELECT md5(pre_query), pre_query INTO live_md5, q FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 <> '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE EXCEPTION '464: written text is not byte-exact to the intended 464 text (got %, want 421dfc4f9f74035d71f43a33c703cd44).', live_md5;
  END IF;

  IF NOT (q LIKE '%failed_initialized AS (%'
      AND q LIKE '%stale INITIALIZED for >4h%'
      AND q LIKE '%as initialized_failed%'
      AND q LIKE '%OR (SELECT COUNT(*) FROM failed_initialized) > 0%') THEN
    RAISE EXCEPTION '464: failed_initialized arm not fully wired (CTE / SELECT / HAVING)';
  END IF;

  -- NEGATIVE CONTROL: every pre-existing arm survived. Without this, a rewrite
  -- that dropped them would pass the assertion above.
  IF NOT (q LIKE '%dispatch loop idle for >30 min%'
      AND q LIKE '%stale AWAITING_RESPONSES for >90 min%'
      AND q LIKE '%stale EXECUTING_STEP for >4h%'
      AND q LIKE '%stale RUNNING for >4h%'
      AND q LIKE '%reap_stale_collection_tasks()%'
      AND q LIKE '%expired_awaited AS (%') THEN
    RAISE EXCEPTION '464: a pre-existing reaper arm was lost in the rewrite';
  END IF;

  -- GUARD 2: it must actually PARSE AND EXECUTE. Effects discarded by sentinel.
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '464: new pre_query does NOT execute (%) — a task with a broken gate never fires and looks merely idle. Aborting.', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '464: INITIALIZED arm added; all six arms present and the pre_query executes.';
END
$do$;

COMMIT;
