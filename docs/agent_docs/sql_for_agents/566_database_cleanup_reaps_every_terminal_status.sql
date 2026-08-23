-- 566 — `database-cleanup` reaps EVERY terminal status, not a literal pair.
--
-- Filed from `bugs_open/354`'s lane. This is a PRE-EXISTING leak, found while costing
-- 354's candidate 1 (a new terminal orchestration status). It is worth shipping on its
-- own account and ships first, because it is losing rows today.
--
-- ── THE DEFECT ──────────────────────────────────────────────────────────────
-- `database-cleanup` has two arms that delete from `orchestration_states`:
--
--   arm 3 (deleted_orchestrations):  WHERE status IN ('COMPLETED', 'FAILED')
--   arm 4 (deleted_stale):           WHERE status NOT IN (SELECT status FROM
--                                      orchestration_status_vocabulary
--                                      WHERE is_terminal OR is_pausable)
--
-- Migration 465/466 converted arm 4 to read the vocabulary — correctly — but left arm 3
-- carrying a literal pair. So a terminal status that is NOT one of those two literals
-- falls through BOTH arms: arm 3 does not name it, and arm 4 skips it precisely BECAUSE
-- it is terminal. Such rows are never deleted by anything.
--
-- ── THE EVIDENCE, AND IT IS NOT HYPOTHETICAL ────────────────────────────────
-- `CANCELLED` is already in exactly that position — terminal since 466, named by neither
-- arm. Measured 2026-08-22 on the live table:
--
--   SELECT status, count(*), min(created_at)::date, max(updated_at)::date
--     FROM orchestration_states WHERE updated_at < now() - interval '24 hours'
--    GROUP BY 1;
--   -->  COMPLETED | 38 | 2026-08-21 | 2026-08-21   (inside the window, reaps normally)
--        CANCELLED | 24 | 2026-07-19 | 2026-07-24   (34 DAYS OLD, against a 24h norm)
--
-- Every row in the table older than 24 hours is CANCELLED. The prediction was made by
-- reading the two arms and THEN confirmed against the table, so it could have come out
-- otherwise: had arm 3 been reaching them, the 24 rows would not exist.
--
-- ── WHY THIS IS THE SAME SHAPE AS bugs_closed/294 ───────────────────────────
-- 466's own header says of the inverted formulation: *"`IN (non-terminal…)` would treat
-- it as NOT reapable — silent, immortal, i.e. exactly bug 294 again"*. That reasoning was
-- applied to arm 4 and not to arm 3, which is where the literal survived. This migration
-- finishes the job 465/466 started: after it, "which statuses are terminal" is asked of
-- the vocabulary in BOTH arms and nowhere is it enumerated.
--
-- ── A DOC CORRECTION THIS FORCES ────────────────────────────────────────────
-- `docs/agent_docs/docs024_key_docs_latest/orchestration_status_lifecycle/RUNBOOK_orchestration_status_lifecycle.md`
-- says adding a status is *"One INSERT. Forgetting it is a hard write failure at the first
-- attempt — loud by design."* That is true for a NON-terminal status and FALSE for a
-- terminal one: the FK accepts the INSERT, nothing fails loudly, and the rows simply never
-- reap. Corrected in that runbook in the same commit as this file. (That lane is closed;
-- this is a contribution to its runbook, not a claim on it.)
--
-- ── WHY replace() AND NOT A PASTED pre_query ────────────────────────────────
-- 466 pasted the whole text. Here the edit is one predicate inside a ~90-line query, and
-- re-typing the other 89 lines to change one is how a transcription error reaches a live
-- sweep. `replace()` is anchored on a fragment PROVEN to occur exactly once (asserted
-- below before the write), and the md5 guard means the input text is known exactly, so the
-- output is deterministic. Both the before-md5 and the after-md5 are asserted.
--
-- ── ADOPTED 2026-08-23, AND WHAT CHANGED IN THE ADOPTION ───────────────────
-- Written by the `bugs_open/354` lane on 2026-08-22 and left UNTRACKED and UNAPPLIED when
-- that session ended. Two other lanes (`bugs_open/358`, `bugs_open/307 [abdc1e]`) found it,
-- independently declined to adopt it, and recorded it rather than let it vanish
-- (`docs024_key_docs_latest/bugfix_358_unread_finding_codes/NOTES_unread_finding_codes.md`,
-- commit 6dd0e01a6). Adopted here at the owner's direction. The SQL below is the original
-- author's; what this session changed and checked:
--
--   * BOTH md5 LITERALS REFRESHED. Migration 567 landed between authoring and adoption, so
--     the before-md5 this file shipped with (c26ccf49…) no longer named the live text and
--     the migration would have refused — correctly. Live text is now 7f4321d4…, and the
--     after-md5 (7e9fe52d…) was recomputed IN THE DATABASE with the same replace()
--     expression section 2 runs, not by hashing a local copy. That distinction is
--     load-bearing: psql's text transfer and `length()` disagree with `md5()` about this
--     row by 3 bytes (`length()` counts CHARACTERS, `md5()` hashes BYTES, and the query
--     holds a multi-byte character), so a locally-computed after-md5 would have been wrong
--     and section 3a would have aborted the migration.
--   * THE ANCHOR AND EVERY ASSERTION PRE-FLIGHTED against the live text: the literal
--     predicate still occurs exactly once, `orchestration_status_vocabulary` occurs once
--     (so 3b's expected 2 is right), the 24-hour bound is present, and all six arms are
--     intact. Nothing in the edit had gone stale except the md5s.
--   * THE PREMISE RE-MEASURED, because a count goes stale by ADDITION and reads as current
--     for ever. `[MEASURED 2026-08-23]` the leak is unchanged and still live: 24 CANCELLED
--     rows older than 24h, oldest 2026-07-19 — now **35 days**, not the 34 the section
--     above records on 2026-08-22. The vocabulary still marks exactly 3 statuses terminal
--     (CANCELLED, COMPLETED, FAILED), so 3e's `n < 3` guard passes with no margin: a
--     FOURTH terminal status is the next thing this migration protects.
--   * THE BLAST RADIUS MEASURED RATHER THAN ASSERTED (CLAUDE.md 2026-07-28: "no collision
--     is possible is a query, not an argument"). Making these 24 rows deletable cascades
--     into four tables (`input_requests`, `pending_requests`, `agent_groups`,
--     `orchestration_requests` — all ON DELETE CASCADE) and orphans rows in
--     `awaited_requests`, which has NO foreign key. For these 24 rows every one of those
--     counts is ZERO. `[MEASURED 2026-08-23]`
--     **With a demand control, because a zero from a query that cannot return non-zero is
--     not evidence**: the same join key against COMPLETED rows returns 4,069
--     `awaited_requests` rows, so the key is right and the zero is real; the four cascade
--     tables are zero because they are EMPTY tables (0 rows each), not because the join
--     missed. And orphaning is pre-existing behaviour of this sweep for every status, not
--     something this migration introduces: 25,886 of 30,330 `awaited_requests` rows are
--     ALREADY orphaned today.
--
-- Rollback sidecar: 566_database_cleanup_reaps_every_terminal_status_ROLLBACK.sql

BEGIN;

-- ── 1. refuse unless the live text is the one this migration was written against ──
DO $do$
DECLARE q text; n int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '566 REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;
  IF md5(q) <> '7f4321d43784dd26b0a9b3ee27ec412d' THEN
    RAISE EXCEPTION '566 REFUSED: database-cleanup pre_query is not the text this migration was written against (md5 %). Another change landed first — re-read the LIVE row, re-derive the edit and re-compute both md5s.', md5(q);
  END IF;

  -- The anchor must be unambiguous. If arm 3's predicate ever appears twice, a blind
  -- replace() would rewrite both and this migration has no opinion about the second.
  n := (length(q) - length(replace(q, 'WHERE status IN (''COMPLETED'', ''FAILED'')', '')))
       / length('WHERE status IN (''COMPLETED'', ''FAILED'')');
  IF n <> 1 THEN
    RAISE EXCEPTION '566 REFUSED: expected exactly 1 occurrence of arm 3''s literal predicate, found %', n;
  END IF;

  RAISE NOTICE '566: live text verified, anchor is unique — applying.';
END
$do$;

-- ── 2. the edit ─────────────────────────────────────────────────────────────
UPDATE scheduled_tasks
   SET pre_query = replace(pre_query,
         'WHERE status IN (''COMPLETED'', ''FAILED'')',
         'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)'),
       updated_at = now()
 WHERE name = 'database-cleanup'
   AND md5(pre_query) = '7f4321d43784dd26b0a9b3ee27ec412d';

-- ── 3. verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────
DO $do$
DECLARE q text; n int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';

  -- 3a. byte-exact: the text is the one this migration intends, not merely "changed"
  IF md5(q) <> '7e9fe52d8f3d822e53fb8afd3628ccd7' THEN
    RAISE EXCEPTION '566: post-edit pre_query is not byte-exact (md5 %)', md5(q);
  END IF;

  -- 3b. the literal is gone and both arms now read the vocabulary
  IF q LIKE '%''COMPLETED'', ''FAILED''%' THEN
    RAISE EXCEPTION '566: arm 3 still carries the literal status pair';
  END IF;
  n := (length(q) - length(replace(q, 'orchestration_status_vocabulary', '')))
       / length('orchestration_status_vocabulary');
  IF n <> 2 THEN
    RAISE EXCEPTION '566: expected both arms to read the vocabulary, found % reference(s)', n;
  END IF;

  -- 3c. NEGATIVE CONTROLS: every pre-existing arm must survive. Without these, a
  --     replace() that ate more than intended would still pass 3a/3b.
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orchestrations AS (%' AND q LIKE '%deleted_stale AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '566: a pre-existing database-cleanup arm was lost';
  END IF;
  -- arm 3 must still be time-bounded; a predicate swap must not widen it to all rows
  IF q NOT LIKE '%AND updated_at < NOW() - INTERVAL ''24 hours''%' THEN
    RAISE EXCEPTION '566: arm 3 lost its 24-hour bound';
  END IF;

  -- 3d. it still parses AND executes (a syntactically valid query that errors at
  --     runtime would silently stop the whole sweep)
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '566: database-cleanup pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;

  -- 3e. the behavioural claim, asserted rather than assumed: CANCELLED is terminal and
  --     is now selected by arm 3's new predicate. This is the whole point of the change.
  IF NOT EXISTS (SELECT 1 FROM orchestration_status_vocabulary
                  WHERE status = 'CANCELLED' AND is_terminal) THEN
    RAISE EXCEPTION '566: CANCELLED is not marked terminal — the premise of this migration is wrong';
  END IF;
  SELECT count(*) INTO n FROM orchestration_status_vocabulary WHERE is_terminal;
  IF n < 3 THEN
    RAISE EXCEPTION '566: expected at least 3 terminal statuses, found %', n;
  END IF;

  RAISE NOTICE '566: both arms read the vocabulary, all six arms intact, query executes, % terminal statuses now reaped.', n;
END
$do$;

-- ── 4. record why CANCELLED sat unreaped, where the next reader will look ───
UPDATE orchestration_status_vocabulary
   SET notes = notes || ' Reaped at 24h by database-cleanup arm 3 since migration 566; before that arm 3 named only COMPLETED/FAILED literally and arm 4 skips is_terminal rows, so CANCELLED rows were never deleted (24 of them, oldest 34 days, measured 2026-08-22).',
       updated_at = now()
 WHERE status = 'CANCELLED';

COMMIT;
