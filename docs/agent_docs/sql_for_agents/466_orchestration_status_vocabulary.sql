-- 466 — `orchestration_states.status` gets a SINGLE SOURCE OF TRUTH, enforced.
--
-- Owner decisions 2026-08-19, after the architecture seat raised it on 465's
-- council round (`1c212b15`): (1) a vocabulary TABLE with a FOREIGN KEY, not a
-- CHECK; (2) rely on the constraint alone — no parity test, no cron; (3) scope to
-- `orchestration_states` only, leaving the 8+ other status columns on their
-- inline CHECKs; (4) DELETE the dead pause vocabulary rather than implement it.
--
-- ── THE PROBLEM 465 LEFT ────────────────────────────────────────────────────
-- 465 closed the CLASS behind bugs_closed/294 and /310 by reaping on an invariant
-- instead of a status list — but that invariant still carried literal sets
-- ('COMPLETED','FAILED','CANCELLED') and the pausable pair, in TWO pre_queries.
-- So "which statuses are terminal" was still an enumeration, just a smaller and
-- better-placed one. Adding a status still meant finding every copy by hand —
-- the same shape, one level up, that produced 294 and 310 in the first place.
--
-- ── WHY A TABLE AND NOT A CHECK, given 8+ tables here use CHECKs ────────────
-- A CHECK would be a SECOND copy of the list and could drift from the table; the
-- FK *is* the table, so the two cannot disagree. The estate's own precedent for
-- moving reaper literals into declared rows is `reaper_policies` (migration 335,
-- RFC_018), and this follows it. The remaining consumers are DB-resident SQL in
-- `scheduled_tasks.pre_query`, which no Go-side constant list can reach — the
-- `sqlInList(workItemTerminalStatuses)` idiom works only for Go-built queries.
--
-- ── WHY `NOT IN (terminal or pausable)` AND NOT `IN (non-terminal)` ─────────
-- They are equivalent while the FK holds, and they FAIL DIFFERENTLY if it is ever
-- dropped. `NOT IN (terminal…)` treats an unknown status as reapable — loud, and
-- the row is cleaned within 4h. `IN (non-terminal…)` would treat it as NOT
-- reapable — silent, immortal, i.e. exactly bug 294 again. The formulation that
-- degrades safely is the one to write down.
--
-- ── THE SEED MUST COVER WHAT GO CAN *WRITE*, NOT WHAT THE TABLE HOLDS ───────
-- Only 5 statuses exist in the table right now; INITIALIZED and RUNNING are
-- absent because 463/464/465 cleaned them — yet both are written constantly.
-- Seeding from `SELECT DISTINCT status` would have made **every new orchestration
-- fail at INSERT**. The seed below is derived from the Go writers instead,
-- enumerated by grep, and the pre-FK guard proves no existing row is excluded.
--
-- ── A LATENT BUG THE FK FORCED OUT ─────────────────────────────────────────
-- `coordinator.go` set `state.Status = ""` as a "reset" when
-- `persistAwaitingStateWithRetry` failed. That is not a member of any vocabulary,
-- and the state IS persisted afterwards: returning false sends control back to
-- `continueExecution`, which — seeing a status that is not AWAITING_RESPONSES —
-- falls through to `saveStepResultWithRetry`. So a blank status could reach this
-- column. 0 rows have ever held one, so the path is rare, but it is real. Fixed in
-- the same commit to restore the prior status. **ORDERING, NAMED HONESTLY: that Go
-- fix is inert until the next roll, while this FK is live immediately.** In the
-- window between, that path would raise an FK violation instead of writing a blank
-- status. That is a louder failure on a path that has ALREADY failed its persist,
-- so it is acceptable — but it is a real, named window, not an absence of one.
--
-- Rollback sidecar: 466_orchestration_status_vocabulary_ROLLBACK.sql

BEGIN;

-- ── 1. the vocabulary ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS orchestration_status_vocabulary (
    status      character varying(50) PRIMARY KEY,
    is_terminal boolean NOT NULL,
    is_pausable boolean NOT NULL DEFAULT false,
    written_by  text,
    notes       text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE orchestration_status_vocabulary IS
  'Single source of truth for orchestration_states.status, enforced by FK.
   is_terminal: the orchestration is finished and must never be reaped.
   is_pausable: it may legitimately wait indefinitely (nothing qualifies today).
   The stale-orchestration-reaper and database-cleanup both read this table
   rather than carrying literal status lists. Adding a status = one INSERT here;
   without it the FK refuses the write. bugs_closed/294, /310, migration 465.';

INSERT INTO orchestration_status_vocabulary (status, is_terminal, is_pausable, written_by, notes) VALUES
  ('INITIALIZED',        false, false, 'StateRepository.CreateInitialState (state.go:734)',        'Row creation. Left when the first inbound message is handled (coordinator.go case StatusInitialized).'),
  ('RUNNING',            false, false, 'StateRepository.ClearExecutingStep (state.go:1428)',       'The inter-step gap during a stuck-orchestration takeover. Milliseconds by construction; bugs_closed/294.'),
  ('EXECUTING_STEP',     false, false, 'StateRepository.SetExecutingStep (state.go:1346)',         'A step is running.'),
  ('AWAITING_RESPONSES', false, false, 'processAwaitResponse (coordinator.go)',                    'Waiting on entries in awaited_requests. The reaper spares it while that map is non-empty.'),
  ('COMPLETED',          true,  false, 'coordinator.go completeWorkflow',                          'Terminal. database-cleanup deletes after 24h.'),
  ('FAILED',             true,  false, 'coordinator.go failWorkflow + the stale-orchestration-reaper', 'Terminal. database-cleanup deletes after 24h.'),
  ('CANCELLED',          true,  false, 'NO GO WRITER — hand-written only',                          'Measured 2026-08-19: 24 rows, oldest 2026-07-19, all with non-empty awaited_requests. Written by direct UPDATE, not by any code path. Terminal by convention (cleanup_stale_topics excludes it from topic protection).')
ON CONFLICT (status) DO NOTHING;

-- ── 2. prove the seed covers reality BEFORE the FK can reject anything ──────
DO $do$
DECLARE missing text;
BEGIN
  SELECT string_agg(DISTINCT s.status, ', ') INTO missing
  FROM orchestration_states s
  LEFT JOIN orchestration_status_vocabulary v ON v.status = s.status
  WHERE v.status IS NULL;
  IF missing IS NOT NULL THEN
    RAISE EXCEPTION '466 REFUSED: these live statuses are not in the seeded vocabulary and the FK would reject them: %. Add them (with is_terminal set correctly) before applying.', missing;
  END IF;
  RAISE NOTICE '466: every live status is covered by the vocabulary.';
END
$do$;

-- ── 3. the enforcement ──────────────────────────────────────────────────────
ALTER TABLE orchestration_states
  DROP CONSTRAINT IF EXISTS fk_orchestration_states_status;
ALTER TABLE orchestration_states
  ADD CONSTRAINT fk_orchestration_states_status
  FOREIGN KEY (status) REFERENCES orchestration_status_vocabulary (status);

-- ── 4. both sweeps read the vocabulary instead of literals ──────────────────
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
    WHERE status NOT IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal OR is_pausable)
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
        WHERE status NOT IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal OR is_pausable)
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
   AND md5(pre_query) = '8659aa16ebc8e0fda15650e3eb50c634';

-- ── 5. verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────
DO $do$
DECLARE q text; r_md5 text; c_md5 text; n int;
BEGIN
  IF to_regclass('orchestration_status_vocabulary') IS NULL THEN
    RAISE EXCEPTION '466: vocabulary table missing';
  END IF;
  SELECT count(*) INTO n FROM orchestration_status_vocabulary;
  IF n <> 7 THEN RAISE EXCEPTION '466: expected 7 vocabulary rows, found %', n; END IF;
  SELECT count(*) INTO n FROM orchestration_status_vocabulary WHERE is_terminal;
  IF n <> 3 THEN RAISE EXCEPTION '466: expected 3 terminal statuses, found %', n; END IF;
  -- nothing is pausable today; the column exists so a future pause status needs
  -- no sweep change at all. Asserted so a stray true is caught, not assumed.
  SELECT count(*) INTO n FROM orchestration_status_vocabulary WHERE is_pausable;
  IF n <> 0 THEN RAISE EXCEPTION '466: expected 0 pausable statuses, found %', n; END IF;

  SELECT count(*) INTO n FROM pg_constraint
   WHERE conname = 'fk_orchestration_states_status' AND contype = 'f';
  IF n <> 1 THEN RAISE EXCEPTION '466: the status foreign key was not created'; END IF;

  SELECT md5(pre_query) INTO r_md5 FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  SELECT md5(pre_query) INTO c_md5 FROM scheduled_tasks WHERE name='database-cleanup';
  IF r_md5 <> '4b53e549449c15defc386f867f0fdf49' THEN RAISE EXCEPTION '466: reaper text not byte-exact (got %)', r_md5; END IF;
  IF c_md5 <> 'c26ccf49e38f5df181006c1d132f19e4' THEN RAISE EXCEPTION '466: cleanup text not byte-exact (got %)', c_md5; END IF;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
  IF q LIKE '%PAUSED_FOR_HUMAN%' OR q LIKE '%''CANCELLED''%' THEN
    RAISE EXCEPTION '466: the reaper still carries literal status lists';
  END IF;
  IF q NOT LIKE '%orchestration_status_vocabulary%' THEN
    RAISE EXCEPTION '466: the reaper does not read the vocabulary';
  END IF;
  -- NEGATIVE CONTROL: the arms that must survive
  IF NOT (q LIKE '%failed_orphaned AS (%' AND q LIKE '%dispatch loop idle for >30 min%'
      AND q LIKE '%stale AWAITING_RESPONSES for >90 min%' AND q LIKE '%stale EXECUTING_STEP for >4h%'
      AND q LIKE '%reap_stale_collection_tasks()%' AND q LIKE '%expired_awaited AS (%') THEN
    RAISE EXCEPTION '466: a pre-existing reaper arm was lost';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '466: reaper pre_query does NOT execute (%)', SQLERRM; END IF;
  END;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='database-cleanup';
  IF q NOT LIKE '%orchestration_status_vocabulary%' THEN
    RAISE EXCEPTION '466: database-cleanup does not read the vocabulary';
  END IF;
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '466: a pre-existing database-cleanup rule was lost';
  END IF;
  BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '466: database-cleanup does NOT execute (%)', SQLERRM; END IF;
  END;

  RAISE NOTICE '466: vocabulary seeded (7 statuses, 3 terminal, 0 pausable), FK enforced, both sweeps read it, both execute.';
END
$do$;

COMMIT;
