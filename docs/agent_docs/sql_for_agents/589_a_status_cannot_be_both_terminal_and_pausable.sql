-- 589 — a status cannot be both `is_terminal` and `is_pausable`, and arm 3 stops ignoring the flag.
--
-- Ships the two halves the council's `guardian` seat asked for on migration 566's round
-- (submission correlation `9d23ccd9-c16c-422d-8bf9-7b60e8b52795`, verdict APPROVED with this as
-- a MEDIUM advisory objection). 566 made `database-cleanup`'s arm 3 read the status vocabulary
-- instead of a literal pair; this closes the hole that opened underneath it.
--
-- ── THE DEFECT ──────────────────────────────────────────────────────────────
-- `orchestration_status_vocabulary` carries two booleans per status:
--   `is_terminal`  — the workflow has ended
--   `is_pausable`  — this status may legitimately wait for ever, so never reap it
--
-- The sweep asks them two SEPARATE questions, in two arms:
--   arm 3 (deleted_orchestrations): DELETE WHERE is_terminal                     -- 24h
--   arm 4 (deleted_stale):          DELETE WHERE NOT (is_terminal OR is_pausable) -- 24h
--
-- Three of the four flag combinations behave sensibly. The fourth does not:
--
--   is_terminal | is_pausable | outcome
--   ------------+-------------+-------------------------------------------------
--   t           | f           | arm 3 deletes at 24h            — ordinary finished run
--   f           | f           | arm 4 deletes at 24h            — stale/stuck cleanup
--   f           | t           | neither arm touches it          — immortal, which is what
--                             |                                   `is_pausable` MEANS
--   t           | t           | **arm 3 DELETES it; arm 4 spares it**             ← the defect
--
-- **Arm 3 never looks at `is_pausable` at all.** So marking a row both silently voids the
-- "never reap this" protection: the two arms disagree and the destructive one wins. The damage
-- would be live human-in-the-loop orchestrations deleted 24h after `updated_at`, and it would
-- look exactly like normal reaping — rows simply gone, no error, no work item, nothing naming
-- the status. The guardian's words: *"a cross-pipeline data-loss failure mode that arm 3's old
-- literal predicate could never trigger, because it only ever matched COMPLETED/FAILED by name."*
--
-- ── WHY BOTH HALVES, AND WHY THAT IS NOT BELT-AND-BRACES FOR ITS OWN SAKE ────
-- They defend at different points and fail in different directions:
--
--   A. the CHECK constraint stops the bad row being WRITTEN — the state becomes
--      unrepresentable, and the error arrives at the moment someone makes the mistake rather
--      than 24 hours later as missing data;
--   B. the arm 3 predicate stops the bad row causing DAMAGE if it ever exists — so the sweep
--      is correct regardless of what the config table holds, and a both-flagged row is then
--      merely never reaped (retained, not destroyed), which is the safe direction to fail in.
--
-- **A reviewer may reasonably say B is unreachable while A holds. That is the point and it is
-- deliberate:** A is one `ALTER TABLE` away from being dropped by any session with psql, and B
-- is what still protects live paused work on the day that happens. B costs 22 characters.
--
-- ── THE PREMISE, MEASURED, AND IT COULD HAVE COME OUT OTHERWISE ─────────────
-- `[MEASURED 2026-08-24]` The vocabulary holds **7** statuses: 3 terminal (CANCELLED,
-- COMPLETED, FAILED), **0 pausable**, and **0 rows are currently both** — so this migration is
-- INERT on today's data and is a guard against a future write, not a repair. It could have come
-- out otherwise: a non-zero count would have meant live rows were already being reaped against
-- their own pausable flag, and the `ALTER TABLE` below would then fail rather than silently
-- skip them (which is why the guard names that case explicitly).
-- `[MEASURED 2026-08-24]` There is **no CHECK constraint of any kind** on the table
-- (`pg_constraint … contype='c'` returns 0 rows), so nothing prevents the write today.
--
-- ── A DECLARED EXTRA: TWO COMMENTS IN THE SWEEP THAT WERE FALSE ─────────────
-- Stated rather than slipped in, because they are not strictly part of the fix. Both are inside
-- the arms this migration edits or sits beside, and both are provably wrong against their own
-- code — which is the exact failure mode that produced this bug family (a register entry,
-- `DBI-014`, quotes retention figures for this sweep that its own text has not matched for
-- weeks). Comments do not execute, so the risk is zero and the anchors are asserted unique:
--   * arm 3 read *"Clean completed/failed orchestration_states"* — untrue since 566, which is
--     what made it vocabulary-driven. Left alone it would contradict this migration's own edit.
--   * arm 4 read *"stuck in EXECUTING_STEP (> 4 hours)"* — the code says **24 hours** and has
--     for as long as it has read the vocabulary, and the arm covers INITIALIZED, RUNNING and
--     AWAITING_RESPONSES too, not only EXECUTING_STEP.
--
-- ── SCOPE ───────────────────────────────────────────────────────────────────
-- Normal council-gate scope, not an RFC, by the owner ruling of 2026-07-29 §1: this does not
-- add a capability or widen a guarantee — it NARROWS a shared mechanism to the behaviour its
-- own column names already promise, and it is inert on live data. It is the same seam, the same
-- arm and a smaller delta than 566, whose own round the `architecture` seat classified
-- `point_fix`. The seam is registered in the concept register in the same commit, per the
-- 2026-07-28 ruling's condition (2).
--
-- Rollback sidecar: 589_a_status_cannot_be_both_terminal_and_pausable_ROLLBACK.sql

BEGIN;

-- ── 1. refuse unless the live text and table are what this migration was written against ──
DO $do$
DECLARE q text; n int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '589 REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;
  IF md5(q) <> '8226bb0f4dfad4bbede0f15a4badacc2' THEN
    RAISE EXCEPTION '589 REFUSED: database-cleanup pre_query is not the text this migration was written against (md5 %). Another lane edits this row often - re-read the LIVE row, re-derive the three anchors and re-compute both md5s. Do NOT force.', md5(q);
  END IF;

  -- each anchor must be unambiguous; a blind replace() on a repeated fragment would rewrite
  -- a second site this migration has no opinion about
  n := (length(q) - length(replace(q, 'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)', '')))
       / length('WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)');
  IF n <> 1 THEN RAISE EXCEPTION '589 REFUSED: arm 3 predicate anchor occurs % time(s), expected 1', n; END IF;

  n := (length(q) - length(replace(q, '-- 3. Clean completed/failed orchestration_states (> 24 hours)', '')))
       / length('-- 3. Clean completed/failed orchestration_states (> 24 hours)');
  IF n <> 1 THEN RAISE EXCEPTION '589 REFUSED: arm 3 comment anchor occurs % time(s), expected 1', n; END IF;

  n := (length(q) - length(replace(q, '-- 4. Clean stale orchestrations stuck in EXECUTING_STEP (> 4 hours)', '')))
       / length('-- 4. Clean stale orchestrations stuck in EXECUTING_STEP (> 4 hours)');
  IF n <> 1 THEN RAISE EXCEPTION '589 REFUSED: arm 4 comment anchor occurs % time(s), expected 1', n; END IF;

  -- the constraint must not already exist (a re-run is a no-op, not a silent second apply)
  IF EXISTS (SELECT 1 FROM pg_constraint
              WHERE conrelid = 'orchestration_status_vocabulary'::regclass
                AND conname = 'chk_status_not_terminal_and_pausable') THEN
    RAISE EXCEPTION '589: already applied (chk_status_not_terminal_and_pausable exists)';
  END IF;

  -- and no live row may already violate it: if one does, the ALTER would fail with a message
  -- that does not say WHY it matters, and the right response is to fix the data first
  SELECT count(*) INTO n FROM orchestration_status_vocabulary WHERE is_terminal AND is_pausable;
  IF n <> 0 THEN
    RAISE EXCEPTION '589 REFUSED: % status row(s) are ALREADY both terminal and pausable. Those rows are being reaped by arm 3 against their own pausable flag RIGHT NOW. Decide per row which flag is wrong, fix the data, then re-run this.', n;
  END IF;

  RAISE NOTICE '589: live text verified, 3 anchors unique, 0 rows violate - applying.';
END
$do$;

-- ── 2a. the constraint: make the contradictory row unrepresentable ──────────
ALTER TABLE orchestration_status_vocabulary
  ADD CONSTRAINT chk_status_not_terminal_and_pausable
  CHECK (NOT (is_terminal AND is_pausable));

COMMENT ON CONSTRAINT chk_status_not_terminal_and_pausable ON orchestration_status_vocabulary IS
  'A status cannot both end a workflow and be allowed to wait for ever. database-cleanup arm 3 deletes is_terminal rows at 24h while arm 4 spares is_pausable ones, so a row marked both is deleted against its own pausable flag. Migration 589.';

-- ── 2b. the sweep: arm 3 stops ignoring is_pausable (+ the two false comments) ──
UPDATE scheduled_tasks
   SET pre_query = replace(replace(replace(pre_query,
         '-- 3. Clean completed/failed orchestration_states (> 24 hours)',
         $n3$-- 3. Clean TERMINAL orchestration_states (> 24 hours). The vocabulary decides which
    --    statuses are terminal (migration 566), EXCLUDING any marked is_pausable: a row that
    --    is both is forbidden by a CHECK constraint (migration 589), and this predicate is
    --    the second line of defence if that constraint is ever dropped.$n3$),
         'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)',
         'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal AND NOT is_pausable)'),
         '-- 4. Clean stale orchestrations stuck in EXECUTING_STEP (> 4 hours)',
         $n4$-- 4. Clean stale NON-terminal, NON-pausable orchestrations (> 24 hours; this comment
    --    read "> 4 hours" until migration 589 and the code never did). Not only
    --    EXECUTING_STEP: also INITIALIZED, RUNNING and AWAITING_RESPONSES.$n4$),
       updated_at = now()
 WHERE name = 'database-cleanup'
   AND md5(pre_query) = '8226bb0f4dfad4bbede0f15a4badacc2';

-- ── 3. verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────
DO $do$
DECLARE q text; n int; ok boolean;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';

  -- 3a. byte-exact: the text is what this migration intends, not merely "changed"
  IF md5(q) <> 'cc43ddc768991444005d228039af197f' THEN
    RAISE EXCEPTION '589: post-edit pre_query is not byte-exact (md5 %)', md5(q);
  END IF;

  -- 3b. arm 3 now excludes pausable rows, and the old predicate is gone
  IF q NOT LIKE '%WHERE is_terminal AND NOT is_pausable)%' THEN
    RAISE EXCEPTION '589: arm 3 does not carry the AND NOT is_pausable exclusion';
  END IF;
  IF q LIKE '%orchestration_status_vocabulary WHERE is_terminal)%' THEN
    RAISE EXCEPTION '589: the old unguarded arm 3 predicate is still present';
  END IF;
  n := (length(q) - length(replace(q, 'orchestration_status_vocabulary', '')))
       / length('orchestration_status_vocabulary');
  IF n <> 2 THEN RAISE EXCEPTION '589: expected both arms to read the vocabulary, found %', n; END IF;

  -- 3c. NEGATIVE CONTROLS: every pre-existing arm survives, and arm 3 keeps its time bound.
  --     Without these, a replace() that ate more than intended would still pass 3a/3b.
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orchestrations AS (%' AND q LIKE '%deleted_stale AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '589: a pre-existing database-cleanup arm was lost';
  END IF;
  n := (length(q) - length(replace(q, 'AND updated_at < NOW() - INTERVAL ''24 hours''', '')))
       / length('AND updated_at < NOW() - INTERVAL ''24 hours''');
  IF n <> 2 THEN RAISE EXCEPTION '589: expected 2 orchestration arms bounded at 24h, found %', n; END IF;

  -- 3d. it still parses AND executes (a valid-looking query that errors at runtime would
  --     silently stop the WHOLE sweep, all six arms, not just this one)
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '589: database-cleanup pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;

  -- 3e. the constraint exists and says what it should
  IF NOT EXISTS (SELECT 1 FROM pg_constraint
                  WHERE conrelid = 'orchestration_status_vocabulary'::regclass
                    AND conname = 'chk_status_not_terminal_and_pausable'
                    AND pg_get_constraintdef(oid) = 'CHECK ((NOT (is_terminal AND is_pausable)))') THEN
    RAISE EXCEPTION '589: constraint missing or not the expected definition (got: %)',
      COALESCE((SELECT pg_get_constraintdef(oid) FROM pg_constraint
                 WHERE conrelid='orchestration_status_vocabulary'::regclass
                   AND conname='chk_status_not_terminal_and_pausable'), '(absent)');
  END IF;

  -- 3f. MUTATION PROOF. A constraint observed only passing cannot be told from one that
  --     cannot fire. Deliberately violate it and require the violation to be caught.
  ok := false;
  BEGIN
    UPDATE orchestration_status_vocabulary SET is_pausable = true WHERE status = 'CANCELLED';
    RAISE EXCEPTION 'MUTATION_NOT_CAUGHT';
  EXCEPTION
    WHEN check_violation THEN ok := true;            -- correct: the constraint fired
    WHEN OTHERS THEN
      IF SQLERRM = 'MUTATION_NOT_CAUGHT' THEN
        RAISE EXCEPTION '589: the CHECK constraint did NOT fire on a deliberate terminal+pausable write - it is inert';
      END IF;
      RAISE;
  END;
  IF NOT ok THEN RAISE EXCEPTION '589: mutation proof did not complete'; END IF;

  -- 3g. CONTROL for the mutation proof: a LEGAL write must still succeed, or 3f would pass
  --     simply because every write to this table fails. Marking a NON-terminal status
  --     pausable is exactly the case the constraint must permit.
  BEGIN
    UPDATE orchestration_status_vocabulary SET is_pausable = true WHERE status = 'INITIALIZED';
    RAISE EXCEPTION 'CONTROL_OK';                    -- unwind: this write must not persist
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'CONTROL_OK' THEN
      RAISE EXCEPTION '589: the constraint BLOCKS a legal write (non-terminal + pausable) - it is too wide (%)', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '589: constraint live and mutation-proven (violation caught, legal write allowed), arm 3 excludes pausable rows, all six arms intact, query executes.';
END
$do$;

-- ── 4. record it where the next reader of the table will look ───────────────
UPDATE orchestration_status_vocabulary
   SET notes = notes || ' Since migration 589 a status cannot be both is_terminal and is_pausable (CHECK chk_status_not_terminal_and_pausable), and database-cleanup arm 3 excludes pausable rows independently of that constraint.',
       updated_at = now()
 WHERE status = 'CANCELLED';

COMMIT;
