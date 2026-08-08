-- 335 — reaper accounting becomes a SHARED MECHANISM: a policy table where each
-- task type declares its own park ceiling, and one function that does the
-- count/backoff/park logic the reapers had been hand-writing per queue.
--
-- Owner decisions 2026-08-08 (bugs_open/205, decisions 3+4): "each task type can
-- declare its own ceiling" + "go ahead with the shared reaper-accounting
-- mechanism". Design + migration path for the OTHER reapers:
-- docs024_key_docs_latest/architecture_review/RFC_018_reaper_accounting_as_a_shared_mechanism.md
--
-- What this replaces: the reset_tasks CTE hand-written into the
-- stale-orchestration-reaper's pre_query (applied 2026-08-07,
-- bugfix_205_poison_pill_reaper/SQL_2026-08-07_reaper_parking.sql). Behaviour is
-- IDENTICAL for the existing population (park on the 5th reset, 20-min staleness,
-- 20-min linear backoff) — the numbers move from literals into a declared policy
-- row, and undeclared task types get the same values as documented defaults.
--
-- NOTE the workstream's ROLLBACK_2026-08-07_reaper_pre_query.sql restores the
-- PRE-205 unconditional reset (it predates this file); the sidecar
-- 335_..._ROLLBACK.sql restores the 2026-08-07 parking CTE instead.
--
-- Idempotent throughout (IF NOT EXISTS / OR REPLACE / ON CONFLICT / same-text
-- UPDATE), so a migration-runner replay after a direct psql apply is a no-op.

BEGIN;

-- ── 1. The policy vocabulary ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS reaper_policies (
    queue               text NOT NULL,
    item_type           text NOT NULL,
    park_after          integer NOT NULL CHECK (park_after >= 1),
    backoff_minutes     integer NOT NULL DEFAULT 20 CHECK (backoff_minutes >= 0),
    stale_after_minutes integer NOT NULL DEFAULT 20 CHECK (stale_after_minutes >= 1),
    notes               text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (queue, item_type)
);

COMMENT ON TABLE reaper_policies IS
  'Per-(queue, item_type) reap-and-park policy. A reaper consults this instead of
   hardcoding its numbers; an undeclared item_type gets the consuming function''s
   documented defaults (park_after 5, backoff 20m, stale 20m). RFC_018.';

INSERT INTO reaper_policies (queue, item_type, park_after, backoff_minutes, stale_after_minutes, notes)
VALUES ('business_intel.collection_tasks', 'initial_verification', 5, 20, 20,
        'bugs_open/205: vet verification. Ceiling chosen 2026-08-07; moved from a pre_query literal to this row 2026-08-08 (owner decision 3).')
ON CONFLICT (queue, item_type) DO NOTHING;

-- ── 2. The shared executor for the collection_tasks queue ───────────────────
-- Contract (RFC_018): a queue opts in by having status / started_at /
-- retry_count / scheduled_for / error_message and a type column. This is the
-- first consumer; the second queue either adopts the same columns and takes a
-- near-copy of this 30-line body, or at that point the body is generalised with
-- dynamic SQL — deliberately not before there are two real consumers to test.
CREATE OR REPLACE FUNCTION business_intel.reap_stale_collection_tasks()
RETURNS TABLE (reset_count integer, parked_count integer)
LANGUAGE sql AS $fn$
WITH eligible AS (
    SELECT ct.id,
           COALESCE(ct.retry_count, 0) + 1 AS next_rc,
           COALESCE(p.park_after, 5)       AS park_after,
           COALESCE(p.backoff_minutes, 20) AS backoff_minutes
      FROM business_intel.collection_tasks ct
      LEFT JOIN reaper_policies p
        ON p.queue = 'business_intel.collection_tasks'
       AND p.item_type = ct.task_type
     WHERE ct.status = 'in_progress'
       AND ct.started_at < NOW() - make_interval(mins => COALESCE(p.stale_after_minutes, 20))
),
updated AS (
    UPDATE business_intel.collection_tasks ct
       SET status           = CASE WHEN e.next_rc >= e.park_after THEN 'failed' ELSE 'pending' END,
           retry_count      = e.next_rc,
           started_at       = NULL,
           orchestration_id = NULL,
           error_message    = CASE WHEN e.next_rc >= e.park_after
                 THEN 'reaper: parked after ' || e.next_rc::text
                      || ' stale-claim resets (park_after=' || e.park_after::text
                      || ', reaper_policies; bugs_open/205)'
                 ELSE ct.error_message END,
           scheduled_for    = CASE WHEN e.next_rc >= e.park_after THEN ct.scheduled_for
                 ELSE NOW() + make_interval(mins => e.backoff_minutes * e.next_rc) END
      FROM eligible e
     WHERE ct.id = e.id
 RETURNING (e.next_rc >= e.park_after) AS parked
)
SELECT count(*) FILTER (WHERE NOT parked)::integer AS reset_count,
       count(*) FILTER (WHERE parked)::integer     AS parked_count
  FROM updated;
$fn$;

-- ── 3. Rewire the reaper's pre_query to call it ─────────────────────────────
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

-- ── 4. Verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────
DO $do$
BEGIN
  IF to_regclass('reaper_policies') IS NULL THEN
    RAISE EXCEPTION 'reaper_policies table missing';
  END IF;
  IF (SELECT count(*) FROM reaper_policies
       WHERE queue='business_intel.collection_tasks' AND item_type='initial_verification') <> 1 THEN
    RAISE EXCEPTION 'seed policy row missing';
  END IF;
  IF (SELECT count(*) FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
       WHERE n.nspname='business_intel' AND p.proname='reap_stale_collection_tasks') <> 1 THEN
    RAISE EXCEPTION 'reap_stale_collection_tasks function missing';
  END IF;
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name='stale-orchestration-reaper'
         AND pre_query LIKE '%reap_stale_collection_tasks()%'
         AND pre_query NOT LIKE '%UPDATE business_intel.collection_tasks%') <> 1 THEN
    RAISE EXCEPTION 'reaper pre_query not rewired to the shared function';
  END IF;
END
$do$;

COMMIT;
