-- ROLLBACK for 465 — restores the ENUMERATED reaper arms and database-cleanup's
-- 4h status-list rule, i.e. the state left by 463 + 464.
--
-- Both texts captured byte-exactly from the LIVE rows on 2026-08-19 before 465
-- was applied (reaper md5 421dfc4f9f74035d71f43a33c703cd44, cleanup md5 be78af6b2d6df6f6e92087126d01dc35).
--
-- Run by hand. The migration runner skips *_ROLLBACK.sql.
--
-- ⚠ WHAT ROLLING BACK COSTS: the reaper returns to reaping a LIST of statuses,
-- so the next status nobody lists becomes immortal again — the class behind
-- bugs_closed/294 and bugs_closed/310 re-opens. The RUNNING and INITIALIZED arms
-- come back, so those two instances stay covered; it is the general case that is
-- lost. database-cleanup also returns to deleting EXECUTING_STEP/AWAITING_RESPONSES
-- at 4h, which races the reaper on the same clock and can delete a row before the
-- reaper records why it died.
--
-- Guards: three-way md5 pre-flight per task (458's form) and an EXECUTE parse
-- check on both restored texts (210's form).

BEGIN;

DO $do$
DECLARE r_md5 text; c_md5 text;
BEGIN
  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 = '421dfc4f9f74035d71f43a33c703cd44' AND c_md5 = 'be78af6b2d6df6f6e92087126d01dc35' THEN
    RAISE NOTICE '465 ROLLBACK: already rolled back — this run is a no-op.';
  ELSIF r_md5 = 'd9a9b83b5cff6799c4acad4b1b78bede' AND c_md5 = '8659aa16ebc8e0fda15650e3eb50c634' THEN
    RAISE NOTICE '465 ROLLBACK: live texts are 465 — reverting both.';
  ELSE
    RAISE EXCEPTION '465 ROLLBACK REFUSED: the live texts are neither 465 nor its pre-image (reaper %, cleanup %). A THIRD edit has landed — re-capture and redo by hand rather than clobbering it.', r_md5, c_md5;
  END IF;
END
$do$;

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
$PQR$,
       updated_at = now()
 WHERE name = 'stale-orchestration-reaper'
   AND md5(pre_query) = 'd9a9b83b5cff6799c4acad4b1b78bede';

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
        WHERE status IN ('EXECUTING_STEP', 'AWAITING_RESPONSES')
          AND updated_at < NOW() - INTERVAL '4 hours'
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
   AND md5(pre_query) = '8659aa16ebc8e0fda15650e3eb50c634';

DO $do$
DECLARE q text; r_md5 text; c_md5 text;
BEGIN
  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 <> '421dfc4f9f74035d71f43a33c703cd44' THEN RAISE EXCEPTION '465 ROLLBACK: reaper not byte-exact (got %, want 421dfc4f9f74035d71f43a33c703cd44)', r_md5; END IF;
  IF c_md5 <> 'be78af6b2d6df6f6e92087126d01dc35' THEN RAISE EXCEPTION '465 ROLLBACK: cleanup not byte-exact (got %, want be78af6b2d6df6f6e92087126d01dc35)', c_md5; END IF;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  IF q LIKE '%failed_orphaned%' THEN RAISE EXCEPTION '465 ROLLBACK: the invariant arm is still present'; END IF;
  IF NOT (q LIKE '%failed_running AS (%' AND q LIKE '%failed_initialized AS (%'
      AND q LIKE '%dispatch loop idle for >30 min%' AND q LIKE '%stale EXECUTING_STEP for >4h%') THEN
    RAISE EXCEPTION '465 ROLLBACK: the enumerated arms were not restored';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '465 ROLLBACK: reaper does NOT execute (%)', SQLERRM; END IF;
  END;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='database-cleanup';
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '465 ROLLBACK: cleanup does NOT execute (%)', SQLERRM; END IF;
  END;

  RAISE NOTICE '465 ROLLBACK: enumerated arms restored; both pre_queries execute.';
END
$do$;

COMMIT;
