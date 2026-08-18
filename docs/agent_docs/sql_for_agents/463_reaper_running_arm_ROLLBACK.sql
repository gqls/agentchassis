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
-- TWO GUARDS. Council round 1 (corr 860d87d9) gated at HIGH on this file's class
-- of risk; round 2 APPROVED, and both remaining advisories are folded in here.
-- BOTH GUARDS FOLLOW AN EXISTING HOUSE IDIOM rather than being invented —
-- reuse_agent's round-2 note was right that I authored them from scratch first:
--   * the three-way md5 branch is 458_detected_item_promoter_..._ROLLBACK.sql;
--   * the EXECUTE-the-stored-SQL check is 210_report_pipeline_scheduled_tasks.sql,
--     whose header puts the danger better than I did: "a pre_query with a typo
--     fails silently at tick time (the task simply never fires), which is the
--     hardest kind of dead pipeline to notice."
--
-- GUARD 1 — CONCURRENCY, three-way (458's form). scheduled_tasks is edited by
-- more than one lane and an ungated full-text rewrite silently CLOBBERS whatever
-- another session put there. Three distinguishable states, because round 2's
-- editquality objection was that a single "someone else edited it" message
-- misdirects an operator under incident pressure when the real cause is a benign
-- repeat run:
--     live == 463's text   -> roll back (the expected case)
--     live == pre-463 text -> ALREADY rolled back; this run is a clean no-op
--     anything else        -> a THIRD edit landed; refuse, do not clobber it
-- (Do NOT gate on updated_at: measured 2026-08-18 it moves for scheduler
-- stamping with pre_query unchanged. md5 of the TEXT is the only check that
-- answers the question being asked.)
--
-- GUARD 2 — FUNCTIONAL PARSE CHECK, the objection debug_historian gated on. A
-- pre_query is DATA to this migration: substring assertions (pre_query LIKE
-- '%...%') prove a needle is present and prove NOTHING about whether the
-- assembled SQL parses. It parses only when the reaper next ticks — so a typo
-- here commits happily and takes out the reaper (all five arms, not just the
-- edited one) minutes later, with no earlier signal.
-- Variation from 210, deliberately: 210 EXECUTEs gate SELECTs, which are inert.
-- The reaper's pre_query MUTATES (it is a chain of UPDATE ... RETURNING CTEs),
-- so executing it plainly would reap rows for real. The sentinel raise below
-- discards the effects while still proving the text parses and runs.
--
-- The guard was PROVEN IN BOTH DIRECTIONS before being trusted (2026-08-18):
-- the live text passed, and a deliberately corrupted copy ('failed_running AS
-- ((((') was CAUGHT. A guard only ever observed passing is indistinguishable
-- from a guard that cannot fire.
-- ─────────────────────────────────────────────────────────────────────────────

BEGIN;

-- GUARD 1 (pre-flight): decide which of the three states we are in.
DO $do$
DECLARE live_md5 text;
BEGIN
  SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 IS NULL THEN
    RAISE EXCEPTION '294/463 ROLLBACK: no stale-orchestration-reaper row exists.';
  ELSIF live_md5 = '91ba970443ff2d237b53633d80e20904' THEN
    RAISE NOTICE '294/463 ROLLBACK: live text is 463 — rolling back.';
  ELSIF live_md5 = '66891c14ef5026700185cbd5c7a945ed' THEN
    RAISE NOTICE '294/463 ROLLBACK: already rolled back — this run is a no-op.';
  ELSE
    RAISE EXCEPTION '294/463 ROLLBACK REFUSED: the reaper pre_query is neither 463''s text nor the pre-463 pre-image (live md5 %). A THIRD edit has landed since 463 — blind restoration would revert it. Re-capture the live text and redo this rollback by hand.', live_md5;
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
   -- 0 rows on the already-rolled-back path, which GUARD 1 has declared benign.
   AND md5(pre_query) = '91ba970443ff2d237b53633d80e20904';

-- GUARD 2 + verify (DO/RAISE — a SELECT cannot stop a COMMIT)
DO $do$
DECLARE q text; live_md5 text;
BEGIN
  SELECT md5(pre_query), pre_query INTO live_md5, q FROM scheduled_tasks
   WHERE name = 'stale-orchestration-reaper';

  IF live_md5 <> '66891c14ef5026700185cbd5c7a945ed' THEN
    RAISE EXCEPTION '294/463 ROLLBACK: restored text is not byte-exact to the captured pre-463 pre-image (got %, want 66891c14ef5026700185cbd5c7a945ed). Do NOT commit it.', live_md5;
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
  -- contain the right substrings. Effects discarded by the sentinel raise.
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '294/463 ROLLBACK: restored pre_query does NOT execute (%) — a task with a broken gate never fires and looks merely idle. Aborting.', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '294/463 ROLLBACK: pre_query restored byte-exactly and proven to execute.';
END
$do$;

COMMIT;
