-- ============================================================================
-- 252_code_index_ref_is_pinned_deterministically.sql
--
-- Make `code-index-refresh` index a NAMED branch, deterministically.
-- OWNER DIRECTIVE 2026-07-28: *"for now I'd like the indexer to use the current
-- branch as that's the most up to date, and I won't find a quiet moment for a
-- while to do the merge."*
--
-- ── CORRECTS 251, WHICH IS INERT ────────────────────────────────────────────
-- 251 added {"ref": "086_experience_loop"} to this task's input_data on the
-- premise that the task supplied NO ref and therefore fell back to HEAD (→ the
-- default branch, main, 2,238 commits behind). **That premise was wrong, and it
-- was wrong because I read PART of the row.** The task has a `pre_query`, and
-- the scheduler treats it as authoritative:
--
--   cmd/scheduler/main.go:216  inputData, err = mergeJSON(task.InputData, dynamicData)
--   cmd/scheduler/main.go:480  for k, v := range overlayMap { baseMap[k] = v }
--
-- The pre-query result is the OVERLAY, so it overwrites any static key of the
-- same name. Worse for 251's purpose:
--
--   cmd/scheduler/main.go:198-212  if dynamicData == nil { ...stampCompleted...; continue }
--
-- A pre-query returning NO ROWS does not fall back to the static input_data —
-- **the task does not fire at all.** So the static `ref` 251 added can never be
-- read, in either branch. It is left in place (it now agrees with what this
-- migration makes operative, so it is documentation rather than a lie) but it
-- is NOT what selects the ref. If you change the branch, change it HERE.
--
-- ── WHAT WAS THERE, AND WHY IT IS BEING REPLACED ────────────────────────────
-- The previous pre_query, recorded verbatim so it can be restored:
--
--   SELECT collected_data->'input_data'->>'ref' AS ref
--     FROM orchestration_states
--    WHERE collected_data->'input_data'->>'ref' ~ '^[0-9]{3}_'
--      AND COALESCE(owner_agent_type,'') NOT IN ('index-orchestrator','code-indexer')
--      AND created_at > now() - interval '14 days'
--    ORDER BY created_at DESC
--    LIMIT 1
--
-- It infers the branch from whatever ref the most recent NON-indexer
-- orchestration happened to carry. That is clever and it has been working — the
-- index is at the branch tip right now because of it. Two failure modes make it
-- wrong for a directive that names a branch:
--
--   1. NO ROWS ⇒ THE INDEX SILENTLY STOPS REFRESHING. Nothing errors; the task
--      stamps itself completed and returns. The freshness banner then ages out
--      on a corpus nobody is updating. It needs only 14 quiet days, or a spell
--      in which no other orchestration carries a `NNN_`-shaped ref.
--   2. It follows whichever branch ANOTHER session last mentioned. Two threads
--      working different `NNN_` branches make the indexed ref a race, decided by
--      `ORDER BY created_at DESC` — and nothing anywhere records that the corpus
--      changed identity.
--
-- Neither is hypothetical after a merge: `main` does not match `^[0-9]{3}_`, so
-- the day this branch is merged the pre-query goes dry and the refresh stops,
-- quietly, which is the worst of the two.
--
-- A constant always returns exactly one row, so the task always fires and always
-- indexes the branch that was named. It is one line to change.
--
-- ── REVERSAL TRIGGER — READ BEFORE MERGING 086_experience_loop TO main ───────
-- Change the literal below to 'main'. Do it as part of the merge, not after:
-- between the merge and the edit, this task keeps indexing a branch that has
-- stopped moving while main becomes the live tree, and the freshness banner
-- keeps reporting FRESH because the commit it names is real and recent. That is
-- bugs_open/108 defect A in its quiet form, re-armed by a stale pin instead of a
-- stale clock.
--
-- Config, so live immediately — no image, no roll.
-- ============================================================================

BEGIN;

DO $$
DECLARE
    v_pre text;
BEGIN
    SELECT pre_query INTO v_pre FROM scheduled_tasks WHERE name = 'code-index-refresh';
    IF NOT FOUND THEN
        RAISE EXCEPTION '252: no scheduled_tasks row named code-index-refresh — re-survey before forcing this';
    END IF;
    IF v_pre IS NOT NULL AND v_pre LIKE '%086_experience_loop%' THEN
        RAISE EXCEPTION '252: the pre_query already pins a literal branch — nothing to do';
    END IF;
    RAISE NOTICE '252: replacing the inferring pre_query with a deterministic pin';
END $$;

UPDATE scheduled_tasks
SET pre_query = $q$-- Owner-pinned 2026-07-28: index the working branch, not the default branch.
-- A CONSTANT, deliberately: this pre_query's result is the OVERLAY that decides
-- the ref (scheduler mergeJSON), and a pre_query returning no rows makes the
-- task skip entirely rather than fall back. One row, always, by construction.
-- CHANGE THIS LITERAL TO 'main' AS PART OF THE MERGE — see migration 252.
SELECT '086_experience_loop'::text AS ref$q$,
    updated_at = now()
WHERE name = 'code-index-refresh';

-- Prove the pre_query returns exactly one row with the right key, by RUNNING it
-- rather than by reading it. A pre_query is executable code with no compiler;
-- the fleet has already been bitten by asserting one instead of executing it.
DO $$
DECLARE
    v_pre  text;
    v_ref  text;
    v_rows int;
BEGIN
    SELECT pre_query INTO v_pre FROM scheduled_tasks WHERE name = 'code-index-refresh';
    EXECUTE 'SELECT count(*), max(ref) FROM (' || v_pre || ') s' INTO v_rows, v_ref;
    IF v_rows <> 1 OR v_ref <> '086_experience_loop' THEN
        RAISE EXCEPTION '252: the stored pre_query returned % row(s) / ref % — a pre_query that returns no rows makes the task skip SILENTLY', v_rows, v_ref;
    END IF;
    RAISE NOTICE '252: stored pre_query executes and yields ref=% (1 row)', v_ref;
END $$;

COMMIT;
