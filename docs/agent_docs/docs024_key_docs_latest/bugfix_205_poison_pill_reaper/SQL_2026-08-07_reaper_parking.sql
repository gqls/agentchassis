-- SQL 2026-08-07 — bug 205: the stale-orchestration-reaper parks a task after 5
-- stale-claim resets instead of resurrecting it for ever.
--
-- No repo seed defines this scheduled_tasks row; the live row is the source of
-- truth. BACK UP THE ROW FIRST (see the \copy line below, run separately), then
-- apply this file. The verify step uses DO/RAISE, not bare SELECTs — a SELECT
-- returning rows cannot stop a COMMIT (LANDMINE, RFC_006 lane).
--
-- Semantics: retry_count counts REAPER RESETS of a stale in_progress claim (it
-- counted nothing before — no code path incremented it). On what would be the
-- 5th reset, the task parks as status='failed' with an error_message naming this
-- bug, instead of returning to 'pending'. Earlier resets back off re-eligibility
-- via scheduled_for (the claim in LoadBusinessBatchAction already honours it).
-- COALESCE guards NULL retry_count (default is 0 but the column is nullable).
--
-- Known residuals, deliberate:
--  * The reaper still resets tasks whose claiming orchestration is alive but
--    slow (>20 min); each steal now at least counts toward the 5. Verifier
--    attempts complete in minutes, so a 20-minute attempt is already pathological.
--  * A reset to 'pending' can in principle collide with the partial unique index
--    idx_collection_tasks_unique_pending if a duplicate pending row exists for
--    the same (business_id, task_type). That hazard is in the CURRENT query too;
--    unchanged here.
--
-- Backup (run BEFORE this file, from the repo root):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
--     -c "SELECT pre_query FROM scheduled_tasks WHERE name='stale-orchestration-reaper'" \
--     > docs/agent_docs/docs024_key_docs_latest/bugfix_205_poison_pill_reaper/BACKUP_2026-08-07_reaper_pre_query.sql

BEGIN;

DO $do$
DECLARE
  n int;
BEGIN
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
    UPDATE business_intel.collection_tasks
    SET status = CASE WHEN COALESCE(retry_count,0) >= 4 THEN 'failed' ELSE 'pending' END,
        retry_count = COALESCE(retry_count,0) + 1,
        started_at = NULL,
        orchestration_id = NULL,
        error_message = CASE WHEN COALESCE(retry_count,0) >= 4
            THEN 'reaper: parked after ' || (COALESCE(retry_count,0) + 1)::text || ' stale-claim resets (bugs_open/205)'
            ELSE error_message END,
        scheduled_for = CASE WHEN COALESCE(retry_count,0) >= 4 THEN scheduled_for
            ELSE NOW() + make_interval(mins => 20 * (COALESCE(retry_count,0) + 1)) END
    WHERE status = 'in_progress'
      AND started_at < NOW() - INTERVAL '20 minutes'
    RETURNING id, (COALESCE(retry_count,0) >= 5) AS parked
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
    (SELECT COUNT(*) FROM reset_tasks WHERE NOT parked)::text as tasks_reset,
    (SELECT COUNT(*) FROM reset_tasks WHERE parked)::text as tasks_parked,
    (SELECT COUNT(*) FROM expired_awaited)::text as awaited_expired
HAVING
    (SELECT COUNT(*) FROM failed_dispatch) > 0
    OR (SELECT COUNT(*) FROM failed_orchs) > 0
    OR (SELECT COUNT(*) FROM failed_wedged) > 0
    OR (SELECT COUNT(*) FROM reset_tasks) > 0
    OR (SELECT COUNT(*) FROM expired_awaited) > 0
$PQ$,
         updated_at = now()
   WHERE name = 'stale-orchestration-reaper';

  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 scheduled_tasks row updated, got %', n;
  END IF;
END
$do$;

-- Verify the new text landed (DO/RAISE so a miss aborts the COMMIT)
DO $do$
BEGIN
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name='stale-orchestration-reaper'
         AND pre_query LIKE '%tasks_parked%'
         AND pre_query LIKE '%bugs_open/205%') <> 1 THEN
    RAISE EXCEPTION 'reaper pre_query does not carry the parking branch';
  END IF;
END
$do$;

COMMIT;

-- ── SEED (run AFTER the above, as its own statement) ────────────────────────
-- The 33 measured loopers have ~50 real dispatches each but retry_count=0
-- (nothing ever counted). Seed them to 4 so the NEW mechanism parks each on its
-- next stale cycle — this exercises the parking branch live (behavioural proof)
-- rather than hand-writing 'failed'. Their true attempt count is recorded in
-- bugs_open/205; the parked message will honestly say '5 stale-claim resets'
-- counting only from this seed.
--
-- UPDATE business_intel.collection_tasks ct
--    SET retry_count = 4
--  WHERE ct.status IN ('in_progress','pending')
--    AND ct.task_type = 'initial_verification'
--    AND ct.id IN (
--      SELECT (os.collected_data->'input_data'->>'task_id')::uuid
--        FROM orchestration_states os
--       WHERE os.owner_agent_type='vet-practice-verifier'
--         AND os.status='FAILED'
--         AND os.created_at > now() - interval '7 days'
--       GROUP BY 1 HAVING count(*) >= 10);
