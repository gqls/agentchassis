-- ROLLBACK for 463 — restores the stale-orchestration-reaper pre_query to its
-- pre-463 text (the RUNNING arm removed, all other arms untouched).
--
-- Captured byte-exactly from the LIVE row on 2026-08-18 before 463 was applied
-- (length 2120), not retyped from a migration file — so this restores what was
-- actually running, which is the only thing a rollback should restore.
--
-- Run by hand, deliberately. The migration runner skips *_ROLLBACK.sql.
--
-- NOTE: rolling back re-opens bugs_open/294 — a row that stops in RUNNING
-- becomes immortal again and pins two Kafka topics for ever.

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
    (SELECT reset_count FROM reset_tasks)::text as tasks_reset,
    (SELECT parked_count FROM reset_tasks)::text as tasks_parked,
    (SELECT COUNT(*) FROM expired_awaited)::text as awaited_expired
HAVING
    (SELECT COUNT(*) FROM failed_dispatch) > 0
    OR (SELECT COUNT(*) FROM failed_orchs) > 0
    OR (SELECT COUNT(*) FROM failed_wedged) > 0
    OR (SELECT reset_count + parked_count FROM reset_tasks) > 0
    OR (SELECT COUNT(*) FROM expired_awaited) > 0
$PQ$,
       updated_at = now()
 WHERE name = 'stale-orchestration-reaper';

DO $do$
BEGIN
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query NOT LIKE '%failed_running%'
         AND pre_query LIKE '%dispatch loop idle for >30 min%'
         AND pre_query LIKE '%stale AWAITING_RESPONSES for >90 min%'
         AND pre_query LIKE '%stale EXECUTING_STEP for >4h%'
         AND pre_query LIKE '%reap_stale_collection_tasks()%') <> 1 THEN
    RAISE EXCEPTION '294/463 ROLLBACK: reaper pre_query not restored cleanly';
  END IF;
END
$do$;

COMMIT;
