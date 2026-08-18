-- ROLLBACK for 470 — restores the stale-orchestration-reaper pre_query to its
-- pre-470 text (the INITIALIZED arm removed, every other arm untouched).
--
-- Captured byte-exactly from the LIVE row on 2026-08-18 before 470 was applied
-- (length 2582, md5 91ba970443ff2d237b53633d80e20904) and verified by md5 against
-- that row — not retyped from a migration file, so this restores what was
-- actually running, which is the only thing a rollback should restore. That text
-- is 463's, i.e. the reaper with arms for AWAITING_RESPONSES x2, EXECUTING_STEP
-- and RUNNING.
--
-- Run by hand, deliberately. The migration runner skips *_ROLLBACK.sql.
--
-- NOTE: rolling back re-opens bugs_open/310 — a row that stops in INITIALIZED
-- becomes permanent again, reaped by nothing and pruned by nothing. It does NOT
-- re-open 294 (the RUNNING arm is in the restored text and stays live), and it
-- releases no Kafka topics either way, because INITIALIZED pins none.
--
-- GUARD 1 — CONCURRENCY, three-way. Distinguishes the benign repeat run from the
-- dangerous third-edit case, so an operator under incident pressure is not told
-- "someone else edited it" when the truth is "you already ran this".
--     live == 470's text -> roll back (the expected case)
--     live == pre-470    -> ALREADY rolled back; clean no-op
--     anything else      -> a THIRD edit landed; refuse, do not clobber it
-- (Not updated_at: it moves for scheduler stamping with pre_query unchanged.)
--
-- GUARD 2 — FUNCTIONAL PARSE CHECK. A pre_query is DATA to this migration;
-- substring assertions prove nothing about whether the assembled SQL parses, and
-- it parses only when the reaper next ticks — so a typo here commits happily and
-- takes out ALL the arms minutes later with no earlier signal. The reaper's
-- pre_query MUTATES, so the sentinel raise discards the effects rather than
-- reaping rows for real.

BEGIN;

-- GUARD 1 (pre-flight)
DO $do$
DECLARE live_md5 text;
BEGIN
  SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 IS NULL THEN
    RAISE EXCEPTION '310/470 ROLLBACK: no stale-orchestration-reaper row exists.';
  ELSIF live_md5 = '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE NOTICE '310/470 ROLLBACK: live text is 470 — rolling back.';
  ELSIF live_md5 = '91ba970443ff2d237b53633d80e20904' THEN
    RAISE NOTICE '310/470 ROLLBACK: already rolled back — this run is a no-op.';
  ELSE
    RAISE EXCEPTION '310/470 ROLLBACK REFUSED: the reaper pre_query is neither 470''s text nor the pre-470 pre-image (live md5 %). A THIRD edit has landed since 470 — blind restoration would revert it. Re-capture the live text and redo this rollback by hand.', live_md5;
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
 WHERE name = 'stale-orchestration-reaper'
   -- 0 rows on the already-rolled-back path, which GUARD 1 declared benign.
   AND md5(pre_query) = '421dfc4f9f74035d71f43a33c703cd44';

-- GUARD 2 + verify
DO $do$
DECLARE q text; live_md5 text;
BEGIN
  SELECT md5(pre_query), pre_query INTO live_md5, q FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 <> '91ba970443ff2d237b53633d80e20904' THEN
    RAISE EXCEPTION '310/470 ROLLBACK: restored text is not byte-exact to the captured pre-470 pre-image (got %, want 91ba970443ff2d237b53633d80e20904). Do NOT commit it.', live_md5;
  END IF;

  -- the INITIALIZED arm is gone and every other arm survived
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query NOT LIKE '%failed_initialized%'
         AND pre_query LIKE '%dispatch loop idle for >30 min%'
         AND pre_query LIKE '%stale AWAITING_RESPONSES for >90 min%'
         AND pre_query LIKE '%stale EXECUTING_STEP for >4h%'
         AND pre_query LIKE '%stale RUNNING for >4h%'
         AND pre_query LIKE '%reap_stale_collection_tasks()%'
         AND pre_query LIKE '%expired_awaited AS (%') <> 1 THEN
    RAISE EXCEPTION '310/470 ROLLBACK: reaper pre_query not restored cleanly';
  END IF;

  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '310/470 ROLLBACK: restored pre_query does NOT execute (%) — a task with a broken gate never fires and looks merely idle. Aborting.', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '310/470 ROLLBACK: pre_query restored byte-exactly and proven to execute.';
END
$do$;

COMMIT;
