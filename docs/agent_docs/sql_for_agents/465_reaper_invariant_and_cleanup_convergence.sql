-- 465 — the reaper stops ENUMERATING statuses and starts reaping on an INVARIANT,
--       and database-cleanup is converged onto the same rule.
--
-- Closes the CLASS behind bugs_closed/294 (RUNNING) and bugs_closed/310
-- (INITIALIZED). Those were two instances of one defect: **the reaper's coverage
-- was a LIST of statuses, so any status nobody listed was never swept — not swept
-- slowly, NEVER.** 463 and 464 added the two missing entries. This removes the
-- need for a third.
--
-- Owner decision 2026-08-19: build the invariant sweep AND converge
-- database-cleanup, which the 310 lane found has the SAME blind spot — two
-- independent enumerations, both missing INITIALIZED, neither aware of the other.
--
-- ── WHAT REPLACES WHAT ──────────────────────────────────────────────────────
-- REAPER: `failed_running` and `failed_initialized` are DELETED and replaced by
-- one `failed_orphaned` arm:
--     non-terminal  AND  not legitimately pausable  AND  awaiting nothing  AND  stale
-- The three specific arms (`failed_dispatch` 30 min, `failed_orchs` 90 min,
-- `failed_wedged` 4 h) STAY — their thresholds are deliberate business policy, not
-- accidents — and `failed_orphaned` excludes the rows they already took, using the
-- same `NOT IN (SELECT ... FROM <cte>)` pattern `failed_orchs` already used.
-- **That exclusion is load-bearing, not tidiness:** two data-modifying CTEs
-- updating the same row in one statement is undefined behaviour in Postgres.
--
-- DATABASE-CLEANUP: its rule 4 deleted `EXECUTING_STEP`/`AWAITING_RESPONSES`
-- outright at 4 h — the same clock as the reaper, so it raced the reaper and could
-- DELETE a row before the reaper wrote why it died, destroying the only record.
-- It now uses the same invariant at **24 h**, i.e. strictly behind the reaper's
-- 4 h, so the normal path is: reaper marks FAILED with a reason at 4 h -> rule 3
-- deletes FAILED at 24 h. Rule 4 becomes a true backstop for anything the reaper
-- somehow missed, rather than a competitor.
--
-- ── THE INVERSION, STATED PLAINLY BECAUSE IT IS THE WHOLE POINT ─────────────
-- This does not abolish enumeration; it INVERTS which side is enumerated, and
-- that changes the failure mode from silent to loud:
--   BEFORE — forget to list a non-terminal status  ->  rows live FOR EVER, silently.
--   AFTER  — forget to list a new TERMINAL status  ->  those rows get FAILED, visibly.
-- A wrong-and-visible failure is recoverable; a silent leak is what produced 294
-- and 310. **If you add a terminal status, add it to BOTH lists in this file.**
--
-- ── `PAUSED_FOR_HUMAN` — A REAL TRAP THIS FILE HAS TO STEP AROUND ───────────
-- The invariant would reap a legitimately-paused orchestration, so pausable
-- statuses are excluded. Measured 2026-08-19, and it is worse than it looks:
--   * Go declares `StatusPausedForHuman = "PAUSED_FOR_HUMAN"` (state.go:28) and
--     **nothing writes it** — grep returns the declaration and no other hit.
--   * Four SQL sites protect `'PAUSED_FOR_HUMAN_INPUT'` — a DIFFERENT string
--     (`cleanup_stale_topics.go:209,213`, `topic_cleanup_handler.go:85`,
--     `dashboard_handlers.go:353`). Nothing writes that either.
--   * `SELECT ... WHERE status ILIKE 'PAUSED%'` returns **0 rows**, all-history.
--   * There is **no CHECK constraint** on `orchestration_states.status`, so any
--     string is accepted and neither spelling is enforced.
-- So the pause capability is declared in five places, implemented in none, and the
-- two halves DISAGREE ON THE SPELLING. **BOTH spellings are excluded below**, so
-- whichever one a future implementer picks, this sweep will not pre-empt it.
--
-- ── EQUIVALENCE, MEASURED ON A MATRIX THAT COULD HAVE FAILED ────────────────
-- Old arms vs new invariant, scratch rows in a rolled-back transaction:
--   RUNNING >4h            old t  new t   <- preserved
--   INITIALIZED >4h        old t  new t   <- preserved
--   WEIRD_NEW_STATE >4h    old f  new t   <- THE POINT: the unlisted status
--   PAUSED_FOR_HUMAN >4h   old f  new f   <- the guard holds
--   RUNNING fresh          old f  new f   <- negative control
--   AWAITING_RESPONSES >4h WITH awaited entries: old f new f  <- still waiting
--   CANCELLED >4h          old f  new f   <- terminal
-- A strict superset: every existing behaviour preserved, exactly one case added.
--
-- Guards are the house idiom (three-way md5 pre-flight <- 458; EXECUTE-the-stored-
-- SQL <- 210). BOTH pre_queries are parse-checked, because this file rewrites two.
-- Rollback sidecar: 465_reaper_invariant_and_cleanup_convergence_ROLLBACK.sql

BEGIN;

DO $do$
DECLARE r_md5 text; c_md5 text;
BEGIN
  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 IS NULL OR c_md5 IS NULL THEN
    RAISE EXCEPTION '465: a target scheduled_tasks row is missing (reaper %, cleanup %)', r_md5, c_md5;
  END IF;
  IF r_md5 = 'd9a9b83b5cff6799c4acad4b1b78bede' AND c_md5 = '8659aa16ebc8e0fda15650e3eb50c634' THEN
    RAISE EXCEPTION '465: already applied — nothing to do.';
  END IF;
  IF r_md5 <> '421dfc4f9f74035d71f43a33c703cd44' THEN
    RAISE EXCEPTION '465 REFUSED: reaper pre_query is not the post-464 text (live md5 %). Another lane has edited it; re-derive from the live text rather than clobbering.', r_md5;
  END IF;
  IF c_md5 <> 'be78af6b2d6df6f6e92087126d01dc35' THEN
    RAISE EXCEPTION '465 REFUSED: database-cleanup pre_query is not the expected text (live md5 %). Another lane has edited it; re-derive from the live text.', c_md5;
  END IF;
  RAISE NOTICE '465: both targets match their expected pre-images — converging.';
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
   AND md5(pre_query) = '421dfc4f9f74035d71f43a33c703cd44';

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
   AND md5(pre_query) = 'be78af6b2d6df6f6e92087126d01dc35';

DO $do$
DECLARE q text; r_md5 text; c_md5 text;
BEGIN
  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 <> 'd9a9b83b5cff6799c4acad4b1b78bede' THEN RAISE EXCEPTION '465: reaper text not byte-exact (got %, want d9a9b83b5cff6799c4acad4b1b78bede)', r_md5; END IF;
  IF c_md5 <> '8659aa16ebc8e0fda15650e3eb50c634' THEN RAISE EXCEPTION '465: cleanup text not byte-exact (got %, want 8659aa16ebc8e0fda15650e3eb50c634)', c_md5; END IF;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  IF NOT (q LIKE '%failed_orphaned AS (%'
      AND q LIKE '%orphaned non-terminal (was%'
      AND q LIKE '%as orphaned_failed%'
      AND q LIKE '%OR (SELECT COUNT(*) FROM failed_orphaned) > 0%'
      AND q LIKE '%PAUSED_FOR_HUMAN%' AND q LIKE '%PAUSED_FOR_HUMAN_INPUT%') THEN
    RAISE EXCEPTION '465: failed_orphaned arm not fully wired (CTE / SELECT / HAVING / pause guard)';
  END IF;
  -- the enumerated arms this replaces must be GONE
  IF q LIKE '%failed_running%' OR q LIKE '%failed_initialized%' THEN
    RAISE EXCEPTION '465: the enumerated RUNNING/INITIALIZED arms are still present — they would double-update rows the invariant also matches.';
  END IF;
  -- NEGATIVE CONTROL: the deliberate business-policy arms must SURVIVE
  IF NOT (q LIKE '%dispatch loop idle for >30 min%'
      AND q LIKE '%stale AWAITING_RESPONSES for >90 min%'
      AND q LIKE '%stale EXECUTING_STEP for >4h%'
      AND q LIKE '%reap_stale_collection_tasks()%'
      AND q LIKE '%expired_awaited AS (%') THEN
    RAISE EXCEPTION '465: a pre-existing reaper arm was lost in the rewrite';
  END IF;
  -- the invariant MUST exclude the three specific arms, or CTEs collide
  IF NOT (q LIKE '%NOT IN (SELECT orchestration_id FROM failed_dispatch)%'
      AND q LIKE '%NOT IN (SELECT orchestration_id FROM failed_orchs)%'
      AND q LIKE '%NOT IN (SELECT orchestration_id FROM failed_wedged)%') THEN
    RAISE EXCEPTION '465: failed_orphaned does not exclude the specific arms — two CTEs could update one row (undefined in Postgres).';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '465: reaper pre_query does NOT execute (%)', SQLERRM; END IF;
  END;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='database-cleanup';
  IF NOT (q LIKE '%NOT IN (%COMPLETED%FAILED%CANCELLED%'
      AND q LIKE '%24 hours%' AND q LIKE '%PAUSED_FOR_HUMAN%') THEN
    RAISE EXCEPTION '465: database-cleanup rule 4 not converged onto the invariant';
  END IF;
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '465: a pre-existing database-cleanup rule was lost in the rewrite';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '465: database-cleanup pre_query does NOT execute (%)', SQLERRM; END IF;
  END;

  RAISE NOTICE '465: reaper converged onto the invariant; database-cleanup converged behind it at 24h; both execute.';
END
$do$;

COMMIT;
