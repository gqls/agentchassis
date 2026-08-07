-- ============================================================================
-- 332_code_index_ref_follows_the_live_working_branch.sql
--
-- Repoint `code-index-refresh` from the DEAD branch 086_experience_loop to the
-- live working branch 087_towards_multiple_domains.
--
-- This is 252's "REVERSAL TRIGGER" being pulled — but NOT to the ref 252 named.
--
-- ── WHY NOT 'main', WHICH IS WHAT 252 AND MEMORY BOTH SAY ───────────────────
-- 252's trigger reads "change the literal to 'main' … as part of the merge".
-- The merge never happened, and the branch moved on instead. Measured
-- 2026-08-07 01:5xZ against `git ls-remote --heads origin`:
--
--   ref                          remote tip    local HEAD is ahead by
--   086_experience_loop          d98010e8b     2,365   <- the current pin
--   087_towards_multiple_domains 2c3041f7d         0   <- == local HEAD
--   main                         c02d56b9a     4,594
--
-- Following 252's literal instruction would make the index ~2x MORE stale than
-- leaving it alone. The instruction was right about the mechanism and wrong
-- about the destination, because it assumed the only way off 086 was a merge.
-- The owner directive 252 itself quotes is the thing to honour: *"I'd like the
-- indexer to use the current branch as that's the most up to date"*.
--
-- ── WHAT WAS ACTUALLY BROKEN (nothing was failing) ──────────────────────────
-- The refresh has been working perfectly and indexing a corpse. Measured:
--
--   scheduled_tasks.code-index-refresh  enabled=t, every 86400s,
--                                       last_triggered 2026-08-06 18:12:42Z
--   code_symbols   4,992 symbols, ref 086_experience_loop,
--                  commit_time  2026-07-28 10:31:33Z   <- the CODE is 9d old
--                  updated_at   2026-08-06 18:14:13Z   <- the JOB ran yesterday
--
-- Recent updated_at + old commit_time is exactly the discrimination bugs_open/108
-- built: the banner keys on commit_time, so it has been honestly reporting STALE
-- this whole time. Consequence measured 2026-08-07: a `090` diagnosis run
-- returned UNVERIFIABLE with "zero hits for `package agenterrors`" — a package
-- created 2026-08-06 and therefore absent from a corpus frozen at 07-28.
--
-- ── WHY THE pre_query AND NOT input_data ────────────────────────────────────
-- The scheduler merges the pre-query result OVER the static input_data
-- (cmd/scheduler/main.go: mergeJSON, then overlayMap wins per key), so
-- input_data->>'ref' is INERT — it still reads 086_experience_loop and is left
-- alone deliberately, exactly as 252 left it. A pre_query returning NO ROWS
-- makes the task skip ENTIRELY rather than fall back (main.go:198-212), which is
-- why this stays a CONSTANT: one row, always, by construction.
--
-- ── THIS PIN WILL ROT AGAIN, AND THAT IS ACCEPTED, NOT OVERLOOKED ───────────
-- The day work moves to 088_*, this silently indexes a branch that has stopped
-- moving. 252 considered and REJECTED the self-deriving alternative for reasons
-- that still hold (it went dry silently on no rows, and it raced between
-- sessions working different NNN_ branches — nothing recorded that the corpus
-- had changed identity). So the constant is deliberate and the rot is the price.
-- **Making it dynamic-but-fail-loud is a real design question, not a drive-by** —
-- it needs an architecture round, because every council `code_checks` lookup and
-- the diagnosis loop's code tier read this corpus.
-- Until then: REPOINTING THIS IS PART OF CUTTING A NEW BRANCH. One line, here.
--
-- ── PREVIOUS VALUE, VERBATIM, SO IT CAN BE RESTORED ─────────────────────────
--   -- Owner-pinned 2026-07-28: index the working branch, not the default branch.
--   -- A CONSTANT, deliberately: this pre_query's result is the OVERLAY that decides
--   -- the ref (scheduler mergeJSON), and a pre_query returning no rows makes the
--   -- task skip entirely rather than fall back. One row, always, by construction.
--   -- CHANGE THIS LITERAL TO 'main' AS PART OF THE MERGE — see migration 252.
--   SELECT '086_experience_loop'::text AS ref
--
-- Config, so live immediately — no image, no roll. The next scheduled tick picks
-- it up; a manual reindex (083b_TRIGGER_code_indexer_v2.sh, REF explicit) makes
-- it current now. The indexer fetches the REMOTE PUSHED tip, so an unpushed
-- commit is invisible however current the ref name looks.
-- ============================================================================

BEGIN;

DO $$
DECLARE
    v_pre text;
BEGIN
    SELECT pre_query INTO v_pre FROM scheduled_tasks WHERE name = 'code-index-refresh';
    IF NOT FOUND THEN
        RAISE EXCEPTION '332: no scheduled_tasks row named code-index-refresh — re-survey before forcing this';
    END IF;

    -- Idempotence, and the runner's probe depends on this wording ('... already ...').
    IF v_pre LIKE '%087_towards_multiple_domains%' THEN
        RAISE EXCEPTION '332: the pre_query already pins 087_towards_multiple_domains — nothing to do';
    END IF;

    -- FAIL CLOSED on an unexpected pin. If another session has already moved this
    -- to something else, silently overwriting their choice is the wrong move —
    -- one tree, many sessions, and this is a fleet-wide corpus.
    IF v_pre NOT LIKE '%086_experience_loop%' THEN
        RAISE EXCEPTION '332: expected the pin to be 086_experience_loop but found pre_query=%; another session has changed it — reconcile before re-running', v_pre;
    END IF;

    RAISE NOTICE '332: repointing the code index from 086_experience_loop (dead, tip d98010e8b) to 087_towards_multiple_domains';
END $$;

UPDATE scheduled_tasks
SET pre_query = $q$-- Owner-pinned 2026-07-28, repointed 2026-08-07 (migration 332): index the LIVE
-- WORKING BRANCH, which is the most up-to-date tree. NOT 'main' — main was 4,594
-- commits behind HEAD when this was set, so pinning there is worse than useless.
-- A CONSTANT, deliberately: this pre_query's result is the OVERLAY that decides
-- the ref (scheduler mergeJSON), and a pre_query returning no rows makes the
-- task skip entirely rather than fall back. One row, always, by construction.
-- ⚠ REPOINT THIS WHEN A NEW BRANCH IS CUT — it is one line, and until it is
-- changed the index silently mirrors a branch that has stopped moving while
-- reporting a real, recent commit. That is bugs_open/108 defect A re-armed by a
-- stale pin instead of a stale clock. See migrations 252 and 332.
SELECT '087_towards_multiple_domains'::text AS ref$q$,
    updated_at = now()
WHERE name = 'code-index-refresh';

-- Prove the stored pre_query by RUNNING it, not by reading it: a pre_query is
-- executable code with no compiler, and a no-rows result disables the task
-- SILENTLY. DO/RAISE rather than a bare SELECT, because ON_ERROR_STOP ignores a
-- non-empty result set and a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    v_pre  text;
    v_ref  text;
    v_rows int;
BEGIN
    SELECT pre_query INTO v_pre FROM scheduled_tasks WHERE name = 'code-index-refresh';
    EXECUTE 'SELECT count(*), max(ref) FROM (' || v_pre || ') s' INTO v_rows, v_ref;
    IF v_rows <> 1 OR v_ref <> '087_towards_multiple_domains' THEN
        RAISE EXCEPTION '332: the stored pre_query returned % row(s) / ref % — a pre_query that returns no rows makes the task skip SILENTLY', v_rows, v_ref;
    END IF;
    RAISE NOTICE '332: stored pre_query executes and yields ref=% (1 row)', v_ref;
END $$;

COMMIT;
