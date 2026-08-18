-- ROLLBACK for 464 — removes the INITIALIZED arm, restoring the post-463 text
-- (i.e. the reaper keeps its RUNNING arm; only INITIALIZED goes).
--
-- Restores the text whose value md5 is 91ba970443ff2d237b53633d80e20904, captured
-- byte-exactly from the LIVE row on 2026-08-18 before 464 was applied.
--
-- Run by hand. The migration runner skips *_ROLLBACK.sql.
--
-- NOTE: rolling back re-opens the INITIALIZED half of the gap — measured at the
-- time, 2 rows idle 871 h and 10.7 h, the second still rising. They pin no Kafka
-- topics, so this is not a bugs_open/240 regression; the cost is permanently
-- non-terminal rows that never leave the active partial index.
--
-- Guards are the house idiom, as in 463's sidecar:
--   GUARD 1 three-way md5 pre-flight  <- 458_detected_item_promoter_..._ROLLBACK.sql
--   GUARD 2 EXECUTE the stored SQL    <- 210_report_pipeline_scheduled_tasks.sql
-- A pre_query is DATA to this file: LIKE assertions prove a needle present and
-- nothing about whether the SQL parses, which it does only when the reaper ticks.

BEGIN;

DO $do$
DECLARE live_md5 text;
BEGIN
  SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 IS NULL THEN
    RAISE EXCEPTION '464 ROLLBACK: no stale-orchestration-reaper row exists.';
  ELSIF live_md5 = '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE NOTICE '464 ROLLBACK: live text is 464 — removing the INITIALIZED arm.';
  ELSIF live_md5 = '91ba970443ff2d237b53633d80e20904' THEN
    RAISE NOTICE '464 ROLLBACK: already rolled back — this run is a no-op.';
  ELSE
    RAISE EXCEPTION '464 ROLLBACK REFUSED: reaper pre_query is neither 464''s text nor the post-463 pre-image (live md5 %). A THIRD edit has landed — blind restoration would revert it. Re-capture the live text and redo this by hand.', live_md5;
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
   AND md5(pre_query) = '421dfc4f9f74035d71f43a33c703cd44';

DO $do$
DECLARE q text; live_md5 text;
BEGIN
  SELECT md5(pre_query), pre_query INTO live_md5, q FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 <> '91ba970443ff2d237b53633d80e20904' THEN
    RAISE EXCEPTION '464 ROLLBACK: restored text is not byte-exact to the pre-464 pre-image (got %, want 91ba970443ff2d237b53633d80e20904). Do NOT commit it.', live_md5;
  END IF;

  IF q LIKE '%failed_initialized%' THEN
    RAISE EXCEPTION '464 ROLLBACK: the INITIALIZED arm is still present.';
  END IF;

  -- NEGATIVE CONTROL: the five arms that must SURVIVE, 463's RUNNING arm included.
  IF NOT (q LIKE '%dispatch loop idle for >30 min%'
      AND q LIKE '%stale AWAITING_RESPONSES for >90 min%'
      AND q LIKE '%stale EXECUTING_STEP for >4h%'
      AND q LIKE '%stale RUNNING for >4h%'
      AND q LIKE '%reap_stale_collection_tasks()%'
      AND q LIKE '%expired_awaited AS (%') THEN
    RAISE EXCEPTION '464 ROLLBACK: an arm that should have survived was lost.';
  END IF;

  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '464 ROLLBACK: restored pre_query does NOT execute (%) — a task with a broken gate never fires and looks merely idle. Aborting.', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '464 ROLLBACK: INITIALIZED arm removed; RUNNING arm intact; pre_query executes.';
END
$do$;

COMMIT;
