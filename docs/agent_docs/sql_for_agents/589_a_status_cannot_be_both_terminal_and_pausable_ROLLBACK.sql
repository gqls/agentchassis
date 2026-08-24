-- 589 ROLLBACK — drop the terminal/pausable CHECK constraint and restore arm 3's unguarded predicate.
--
-- ⚠ WHAT YOU ARE RESTORING IS A DATA-LOSS HAZARD. After rollback, a status row may again be
-- marked BOTH `is_terminal` and `is_pausable`, and `database-cleanup` arm 3 will delete those
-- orchestrations 24 hours after `updated_at` while arm 4 spares them — i.e. live paused,
-- human-in-the-loop work destroyed against its own "never reap this" flag, silently, with no
-- error and nothing naming the status. That is the failure the council's `guardian` seat raised
-- on 566's round and that 589 exists to close.
--
-- Roll back only if 589 broke something live. Two plausible reasons, neither yet observed:
--   * the sweep errors — the whole `pre_query` is ONE statement, so if it fails NONE of the six
--     arms run, including the agent_error_log and audit cleanups. The tell is `scheduled_tasks`
--     reporting a failed run for `database-cleanup`, or all six deletion counts at zero while
--     rows accumulate;
--   * a legitimate status genuinely needs both flags. **Prefer fixing that case over rolling
--     back**: if a status must end a workflow AND wait for ever, the two arms disagree about it
--     by construction and the design question is which arm is wrong, not how to re-permit the
--     contradiction.
--
-- Dropping ONLY the constraint (and leaving arm 3 guarded) is the safer partial rollback: run
-- just section 1 below and stop. Arm 3's `AND NOT is_pausable` is inert while no row is both,
-- so it costs nothing to leave in place.
--
-- The guard is on the POST-589 md5, so this refuses if anything landed after 589. That row is
-- edited often by other lanes — a refusal here is the guard working, not a fault.

BEGIN;

DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '589 ROLLBACK REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;
  IF md5(q) <> 'cc43ddc768991444005d228039af197f' THEN
    RAISE EXCEPTION '589 ROLLBACK REFUSED: pre_query is not 589''s output (md5 %). Something landed after 589 - read the live row and undo by hand rather than clobbering another lane''s change.', md5(q);
  END IF;
END
$do$;

-- ── 1. drop the constraint ──────────────────────────────────────────────────
ALTER TABLE orchestration_status_vocabulary
  DROP CONSTRAINT IF EXISTS chk_status_not_terminal_and_pausable;

-- ── 2. restore arm 3's unguarded predicate and the two original comments ────
UPDATE scheduled_tasks
   SET pre_query = replace(replace(replace(pre_query,
         $n3$-- 3. Clean TERMINAL orchestration_states (> 24 hours). The vocabulary decides which
    --    statuses are terminal (migration 566), EXCLUDING any marked is_pausable: a row that
    --    is both is forbidden by a CHECK constraint (migration 589), and this predicate is
    --    the second line of defence if that constraint is ever dropped.$n3$,
         '-- 3. Clean completed/failed orchestration_states (> 24 hours)'),
         'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal AND NOT is_pausable)',
         'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)'),
         $n4$-- 4. Clean stale NON-terminal, NON-pausable orchestrations (> 24 hours; this comment
    --    read "> 4 hours" until migration 589 and the code never did). Not only
    --    EXECUTING_STEP: also INITIALIZED, RUNNING and AWAITING_RESPONSES.$n4$,
         '-- 4. Clean stale orchestrations stuck in EXECUTING_STEP (> 4 hours)'),
       updated_at = now()
 WHERE name = 'database-cleanup'
   AND md5(pre_query) = 'cc43ddc768991444005d228039af197f';

UPDATE orchestration_status_vocabulary
   SET notes = replace(notes, ' Since migration 589 a status cannot be both is_terminal and is_pausable (CHECK chk_status_not_terminal_and_pausable), and database-cleanup arm 3 excludes pausable rows independently of that constraint.', ''),
       updated_at = now()
 WHERE status = 'CANCELLED';

-- ── 3. verify the undo actually undid ───────────────────────────────────────
DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF md5(q) <> '8226bb0f4dfad4bbede0f15a4badacc2' THEN
    RAISE EXCEPTION '589 ROLLBACK: pre_query did not return to its pre-589 text (md5 %)', md5(q);
  END IF;
  IF EXISTS (SELECT 1 FROM pg_constraint
              WHERE conrelid = 'orchestration_status_vocabulary'::regclass
                AND conname = 'chk_status_not_terminal_and_pausable') THEN
    RAISE EXCEPTION '589 ROLLBACK: the constraint is still present';
  END IF;
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '589 ROLLBACK: restored pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;
  RAISE NOTICE '589 ROLLBACK: constraint dropped, arm 3 unguarded again, query executes. The terminal+pausable hazard is back.';
END
$do$;

COMMIT;
