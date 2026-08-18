-- ROLLBACK for 463 — restores the stale-orchestration-reaper pre_query to its
-- pre-463 text (the RUNNING arm removed, all other arms untouched).
--
-- Captured byte-exactly from the LIVE row on 2026-08-18 before 463 was applied
-- (length 2120, value md5 66891c14ef5026700185cbd5c7a945ed) and verified by md5
-- against that row — not retyped from a migration file, so this restores what
-- was actually running, which is the only thing a rollback should restore.
--
-- Run by hand, deliberately. The migration runner skips *_ROLLBACK.sql.
--
-- NOTE: rolling back re-opens bugs_open/294 — a row that stops in RUNNING
-- becomes immortal again and pins two Kafka topics for ever.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- TWO GUARDS, ADDED 2026-08-18 AFTER COUNCIL ROUND 1 (corr 860d87d9) RETURNED
-- REVISE ON EXACTLY THIS FILE'S CLASS OF RISK. Both are copyable: ANY migration
-- that rewrites a scheduled_tasks.pre_query wants them.
--
-- GUARD 1 — CONCURRENCY. The UPDATE is gated on the md5 of the text we expect to
-- be replacing. scheduled_tasks is edited by more than one lane, and an ungated
-- full-text rewrite silently CLOBBERS whatever another session put there. Gated,
-- a drifted row makes this a 0-row no-op, which the verify block then turns into
-- a loud abort. (Do NOT gate on updated_at: measured 2026-08-18, it moves for
-- scheduler stamping with pre_query unchanged. md5 of the text is the only
-- check that answers the question actually being asked.)
--
-- GUARD 2 — FUNCTIONAL PARSE CHECK, the objection debug_historian gated on at
-- HIGH severity. A pre_query is DATA to this migration: substring assertions
-- (pre_query LIKE '%...%') prove a needle is present, and prove NOTHING about
-- whether the assembled SQL parses. It parses only when the reaper next ticks —
-- so a typo here commits happily and breaks THE WHOLE REAPER (all five arms,
-- not just the edited one) minutes later, with no earlier signal. The guard
-- below EXECUTEs the text we just wrote inside a sub-block, then raises a
-- sentinel to discard the effects, so a syntax error aborts the migration.
--
-- The guard was PROVEN IN BOTH DIRECTIONS before being trusted (2026-08-18):
-- the live text passed, and a deliberately corrupted copy ('failed_running AS
-- ((((') was CAUGHT. A guard only ever observed passing is indistinguishable
-- from a guard that cannot fire.
-- ─────────────────────────────────────────────────────────────────────────────

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
 WHERE name = 'stale-orchestration-reaper'
   -- GUARD 1: only roll back the exact text 463 installed.
   AND md5(pre_query) = '91ba970443ff2d237b53633d80e20904';

-- GUARD 2 + verify (DO/RAISE — a SELECT cannot stop a COMMIT)
DO $do$
DECLARE q text;
BEGIN
  -- Did GUARD 1 refuse? Then the live row is not what we expected: abort loudly
  -- rather than leave the operator believing a rollback happened.
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND md5(pre_query) = '66891c14ef5026700185cbd5c7a945ed') <> 1 THEN
    RAISE EXCEPTION '294/463 ROLLBACK REFUSED: the reaper pre_query is not the text 463 installed (md5 mismatch) — another lane has edited it since. Re-capture the live text and redo this rollback by hand rather than clobbering their change.';
  END IF;

  -- the RUNNING arm is gone and every other arm survived
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'stale-orchestration-reaper'
         AND pre_query NOT LIKE '%failed_running%'
         AND pre_query LIKE '%dispatch loop idle for >30 min%'
         AND pre_query LIKE '%stale AWAITING_RESPONSES for >90 min%'
         AND pre_query LIKE '%stale EXECUTING_STEP for >4h%'
         AND pre_query LIKE '%reap_stale_collection_tasks()%'
         AND pre_query LIKE '%expired_awaited AS (%') <> 1 THEN
    RAISE EXCEPTION '294/463 ROLLBACK: reaper pre_query not restored cleanly';
  END IF;

  -- GUARD 2: the restored text must actually PARSE AND EXECUTE, not merely
  -- contain the right substrings. Effects are discarded by the sentinel raise.
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'stale-orchestration-reaper';
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '294/463 ROLLBACK: restored pre_query does NOT execute (%) — the reaper would break at its next tick. Aborting.', SQLERRM;
    END IF;
  END;
END
$do$;

COMMIT;
