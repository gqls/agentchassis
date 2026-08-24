-- 580 — `database-cleanup`'s arm-1 COMMENT stops describing behaviour that no longer exists.
--
-- A defect this lane INTRODUCED, found while closing out. Migration 567 replaced arm 1's
-- retention rule but left the comment above it untouched, so the live sweep reads:
--
--     -- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)
--
-- and that is false in both halves. `resolved` no longer shortens a row's life at all, and
-- unresolved rows live 30 days ONLY if their code is in the short-retention list — everything
-- else lives 365.
--
-- ── WHY A COMMENT IS WORTH A MIGRATION HERE ─────────────────────────────────
-- Normally it would not be. This one is worth it because of WHERE it lives. `pre_query` is not
-- source you read in an editor beside the code that contradicts it; it is a text column, and an
-- operator inspecting `database-cleanup` sees the comment and the SQL together with nothing else
-- to check it against. A stale comment in a repo file is corrected by the diff below it. A stale
-- comment in a live config row is the only description of the row there is.
--
-- It is also exactly the class `bugs_open/358` is about — a record that misleads whoever reads it —
-- so leaving it while closing that lane would be poor form.
--
-- ── SCOPE: THE COMMENT AND NOTHING ELSE ─────────────────────────────────────
-- No predicate changes. The verify block asserts the retention RULE is byte-identical before and
-- after, so this cannot quietly alter behaviour while claiming to touch only prose.
--
-- Rollback sidecar: 580_database_cleanup_comment_stops_lying_about_retention_ROLLBACK.sql

BEGIN;

DO $do$
DECLARE q text; n int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '580 REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;

  -- Anchored on the stale comment, not on a whole-text md5. This row is edited by several lanes
  -- (466, 566, 567) and pinning the whole text would make this refuse for reasons that have
  -- nothing to do with the comment. The anchor must still be unambiguous.
  n := (length(q) - length(replace(q, '-- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)', '')))
       / length('-- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)');
  IF n <> 1 THEN
    RAISE EXCEPTION '580 REFUSED: expected the stale comment exactly once, found % — it was already fixed, or reworded', n;
  END IF;

  -- The thing this migration must NOT change had better be here to begin with.
  IF q NOT LIKE '%INTERVAL ''365 days''%' THEN
    RAISE EXCEPTION '580 REFUSED: arm 1 does not carry 567''s 365-day rule — this is not the row this comment belongs to';
  END IF;
END
$do$;

UPDATE scheduled_tasks
   SET pre_query = replace(pre_query,
         '-- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)',
         '-- 1. Clean agent_error_log. Retention is BY FINDING CODE since migration 567:' || chr(10) ||
         '    --    codes in the list below (failure plumbing + RFC_029 instrumentation) expire at 30 days;' || chr(10) ||
         '    --    EVERY OTHER CODE LIVES 365 DAYS, because a deliberate finding outlives the plumbing it' || chr(10) ||
         '    --    shares this table with. `resolved` does NOT shorten a row any more (it used to halve it' || chr(10) ||
         '    --    to 14 days, which was backwards: a resolved row is the finding PLUS its outcome).' || chr(10) ||
         '    --    The list and the finding-code registry are kept in step by config-key-audit' || chr(10) ||
         '    --    --finding-codes; see bugs_open/358.'),
       updated_at = now()
 WHERE name = 'database-cleanup';

DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';

  -- 1. the lie is gone and the truth is there
  IF q LIKE '%resolved errors > 14 days%' THEN
    RAISE EXCEPTION '580: the stale comment survives';
  END IF;
  IF q NOT LIKE '%Retention is BY FINDING CODE since migration 567%' THEN
    RAISE EXCEPTION '580: the replacement comment is not present';
  END IF;

  -- 2. THE POINT OF THIS BLOCK: behaviour is untouched. A "comment-only" change that moved a
  --    predicate would be far worse than the stale comment it fixed.
  IF q NOT LIKE '%INTERVAL ''365 days''%'
     OR q NOT LIKE '%split_part(error_code, '':'', 1) = ANY (ARRAY[%'
     OR q LIKE '%resolved = true AND occurred_at%'
     OR q LIKE '%INTERVAL ''14 days''%' THEN
    RAISE EXCEPTION '580: arm 1''s retention RULE changed — this migration may only touch the comment';
  END IF;

  -- 3. every other arm survives, arm 3 in its post-566 form
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orchestrations AS (%' AND q LIKE '%deleted_stale AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '580: a pre-existing database-cleanup arm was lost';
  END IF;

  -- 4. and it still executes. A comment edit that breaks the sweep is not a comment edit.
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '580: database-cleanup pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '580: applied — the sweep now describes what it actually does.';
END
$do$;

COMMIT;
