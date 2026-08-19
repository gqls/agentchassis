-- ROLLBACK for 466 — drops the status foreign key and the vocabulary table, and
-- restores both pre_queries to 465's literal status lists.
--
-- Run by hand. The migration runner skips *_ROLLBACK.sql.
--
-- ⚠ WHAT ROLLING BACK COSTS: `orchestration_states.status` goes back to having NO
-- constraint of any kind, so any string becomes writable again — including the
-- blank status that `coordinator.go` wrote on a failed-persist path before the fix
-- in the same commit. The reaper and database-cleanup go back to carrying literal
-- terminal/pausable lists in two places, so adding a status once again means
-- finding every copy by hand. The class fix (465) itself SURVIVES a rollback of
-- this file — the invariant arm stays, it just reads literals again.
--
-- ORDER MATTERS: the FK must be dropped before the table it references.
--
-- Guards: md5 pre-flight per task (458's form) and an EXECUTE parse check on both
-- restored texts (210's form).

BEGIN;

DO $do$
DECLARE r_md5 text; c_md5 text;
BEGIN
  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 = 'd9a9b83b5cff6799c4acad4b1b78bede' AND c_md5 = '8659aa16ebc8e0fda15650e3eb50c634' THEN
    RAISE NOTICE '466 ROLLBACK: pre_queries already reverted — continuing to drop the FK/table if present.';
  ELSIF r_md5 = '4b53e549449c15defc386f867f0fdf49' AND c_md5 = 'c26ccf49e38f5df181006c1d132f19e4' THEN
    RAISE NOTICE '466 ROLLBACK: live texts are 466 — reverting both.';
  ELSE
    RAISE EXCEPTION '466 ROLLBACK REFUSED: live texts are neither 466 nor its pre-image (reaper %, cleanup %). A THIRD edit has landed — re-capture and redo by hand.', r_md5, c_md5;
  END IF;
END
$do$;

-- FK first: the table cannot be dropped while it is referenced.
ALTER TABLE orchestration_states DROP CONSTRAINT IF EXISTS fk_orchestration_states_status;
DROP TABLE IF EXISTS orchestration_status_vocabulary;

UPDATE scheduled_tasks
   SET pre_query = $PQR$
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
failed_orphaned AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: orphaned non-terminal (was ' || status || ') for >4h with no awaited requests; step=' || COALESCE(current_step, '(none)'),
        updated_at = NOW()
    WHERE status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED')
      AND status NOT IN ('PAUSED_FOR_HUMAN', 'PAUSED_FOR_HUMAN_INPUT')
      AND COALESCE(awaited_requests, '{}'::jsonb) = '{}'::jsonb
      AND last_activity < NOW() - INTERVAL '4 hours'
      AND orchestration_id NOT IN (SELECT orchestration_id FROM failed_dispatch)
      AND orchestration_id NOT IN (SELECT orchestration_id FROM failed_orchs)
      AND orchestration_id NOT IN (SELECT orchestration_id FROM failed_wedged)
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
    (SELECT COUNT(*) FROM failed_orphaned)::text as orphaned_failed,
    (SELECT reset_count FROM reset_tasks)::text as tasks_reset,
    (SELECT parked_count FROM reset_tasks)::text as tasks_parked,
    (SELECT COUNT(*) FROM expired_awaited)::text as awaited_expired
HAVING
    (SELECT COUNT(*) FROM failed_dispatch) > 0
    OR (SELECT COUNT(*) FROM failed_orchs) > 0
    OR (SELECT COUNT(*) FROM failed_wedged) > 0
    OR (SELECT COUNT(*) FROM failed_orphaned) > 0
    OR (SELECT reset_count + parked_count FROM reset_tasks) > 0
    OR (SELECT COUNT(*) FROM expired_awaited) > 0
$PQR$,
       updated_at = now()
 WHERE name = 'stale-orchestration-reaper'
   AND md5(pre_query) = '4b53e549449c15defc386f867f0fdf49';

UPDATE scheduled_tasks
   SET pre_query = $PQC$
    -- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)
    WITH deleted_errors AS (
        DELETE FROM agent_error_log
        WHERE (resolved = true AND occurred_at < NOW() - INTERVAL '14 days')
           OR (resolved = false AND occurred_at < NOW() - INTERVAL '30 days')
        RETURNING id
    ),
 
    -- 2. Clean orchestration_state_audit (keep last 100k rows)
    deleted_audit AS (
        DELETE FROM orchestration_state_audit
        WHERE id < (
            SELECT COALESCE(MAX(id) - 100000, 0)
            FROM orchestration_state_audit
        )
        RETURNING id
    ),
 
    -- 3. Clean completed/failed orchestration_states (> 24 hours)
    -- CASCADE will clean orchestration_requests, input_requests, pending_requests
    deleted_orchestrations AS (
        DELETE FROM orchestration_states
        WHERE status IN ('COMPLETED', 'FAILED')
          AND updated_at < NOW() - INTERVAL '24 hours'
        RETURNING orchestration_id
    ),
 
    -- 4. Clean stale orchestrations stuck in EXECUTING_STEP (> 4 hours)
    -- These are leftovers from pod restarts that never completed
    deleted_stale AS (
        DELETE FROM orchestration_states
        WHERE status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED')
          AND status NOT IN ('PAUSED_FOR_HUMAN', 'PAUSED_FOR_HUMAN_INPUT')
          AND updated_at < NOW() - INTERVAL '24 hours'
        RETURNING orchestration_id
    ),
 
    -- 5. Clean orphan composition palettes
    --    Adopted palettes not referenced by any css_themes row, with a
    --    24h grace period to avoid racing in-flight installs. Seed
    --    palettes are NEVER touched (origin filter).
    deleted_orphan_palettes AS (
        DELETE FROM palettes p
        WHERE p.origin = 'adopted'
          AND p.source_site_id IS NOT NULL
          AND p.forked_at < NOW() - INTERVAL '24 hours'
          AND NOT EXISTS (
              SELECT 1 FROM css_themes t WHERE t.palette_id = p.id
          )
        RETURNING id
    ),
 
    -- 6. Clean orphan composition typography_sets
    --    Same shape as palettes. Seed typography_sets (sans-modern,
    --    serif-editorial, etc.) stay even when no active theme currently
    --    references them — they're library resources for future forks.
    deleted_orphan_typography AS (
        DELETE FROM typography_sets ts
        WHERE ts.origin = 'adopted'
          AND ts.source_site_id IS NOT NULL
          AND ts.forked_at < NOW() - INTERVAL '24 hours'
          AND NOT EXISTS (
              SELECT 1 FROM css_themes t WHERE t.typography_set_id = ts.id
          )
        RETURNING id
    )
 
    SELECT
        (SELECT COUNT(*) FROM deleted_errors)::text as errors_deleted,
        (SELECT COUNT(*) FROM deleted_audit)::text as audit_deleted,
        (SELECT COUNT(*) FROM deleted_orchestrations)::text as orchestrations_deleted,
        (SELECT COUNT(*) FROM deleted_stale)::text as stale_deleted,
        (SELECT COUNT(*) FROM deleted_orphan_palettes)::text as orphan_palettes_deleted,
        (SELECT COUNT(*) FROM deleted_orphan_typography)::text as orphan_typography_deleted
    -- Always returns a row so scheduler marks task as executed
$PQC$,
       updated_at = now()
 WHERE name = 'database-cleanup'
   AND md5(pre_query) = 'c26ccf49e38f5df181006c1d132f19e4';

DO $do$
DECLARE q text; r_md5 text; c_md5 text;
BEGIN
  IF to_regclass('orchestration_status_vocabulary') IS NOT NULL THEN
    RAISE EXCEPTION '466 ROLLBACK: the vocabulary table still exists';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname='fk_orchestration_states_status') THEN
    RAISE EXCEPTION '466 ROLLBACK: the status foreign key still exists';
  END IF;
  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 <> 'd9a9b83b5cff6799c4acad4b1b78bede' THEN RAISE EXCEPTION '466 ROLLBACK: reaper not byte-exact (got %)', r_md5; END IF;
  IF c_md5 <> '8659aa16ebc8e0fda15650e3eb50c634' THEN RAISE EXCEPTION '466 ROLLBACK: cleanup not byte-exact (got %)', c_md5; END IF;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  IF q LIKE '%orchestration_status_vocabulary%' THEN
    RAISE EXCEPTION '466 ROLLBACK: the reaper still references the dropped table — it would fail at its next tick';
  END IF;
  -- NEGATIVE CONTROL: 465's class fix must SURVIVE this rollback
  IF q NOT LIKE '%failed_orphaned AS (%' THEN
    RAISE EXCEPTION '466 ROLLBACK: 465s invariant arm was lost — this file should only revert the vocabulary';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '466 ROLLBACK: reaper does NOT execute (%)', SQLERRM; END IF;
  END;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='database-cleanup';
  IF q LIKE '%orchestration_status_vocabulary%' THEN
    RAISE EXCEPTION '466 ROLLBACK: database-cleanup still references the dropped table';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '466 ROLLBACK: database-cleanup does NOT execute (%)', SQLERRM; END IF;
  END;

  RAISE NOTICE '466 ROLLBACK: FK and vocabulary dropped; both sweeps back on literals; 465 invariant intact; both execute.';
END
$do$;

COMMIT;
