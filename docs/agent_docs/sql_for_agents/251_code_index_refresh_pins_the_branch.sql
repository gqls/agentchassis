-- ============================================================================
-- 251_code_index_refresh_pins_the_branch.sql
--
-- Pin the 24h `code-index-refresh` scheduled task to an EXPLICIT ref
-- (086_experience_loop), instead of letting it resolve implicitly.
--
-- OWNER DIRECTIVE 2026-07-28: index the working branch, because that is the
-- most up-to-date code, and a merge to main is not happening for a while.
--
-- WHY THIS IS NEEDED EVEN THOUGH THE INDEX IS CURRENTLY CORRECT
--   As of 11:26 today the index is exactly at the branch tip
--   (ref='086_experience_loop', commit d98010e8b, 0 commits behind, 4,992/4,992
--   bodies). But the scheduled task's input_data is {repo, owner, language} —
--   there is NO ref key. analyse_repo_local therefore falls back to ref="HEAD"
--   (analyse_repo_local_action.go), and a GitHub tarball at HEAD is the
--   repository's DEFAULT branch, which `gh repo view` reports as **main**.
--   main is 2,238 commits behind this branch. So the next scheduled fire could
--   silently re-point the whole index at code from before 24 July, and nothing
--   would report an error — the index would simply describe a different tree.
--
--   That the index currently holds the BRANCH tip while config names no ref is
--   itself the tell: something is resolving the ref outside this row. Config
--   that produces the right answer for a reason nobody can state is not
--   configured, it is lucky. This makes it stated.
--
-- WHY A REF AND NOT A COMMIT
--   A pinned commit would freeze the index at one tree and require a migration
--   per refresh — the opposite of a 24h cadence. A ref tracks the branch's tip
--   as it moves, which is what "index the most up-to-date code" means.
--
-- REVERSAL TRIGGER — READ THIS BEFORE THE MERGE TO main
--   When 086_experience_loop is merged into main, this pin becomes WRONG in the
--   quiet direction: it will keep indexing a branch that has stopped moving
--   while main becomes the live tree, and the freshness banner will keep saying
--   FRESH because the commit it names is real and recent. Change the ref to
--   'main' (or drop the key to return to implicit HEAD resolution) as part of
--   that merge. This paragraph exists because the failure has no error message.
--
-- Field path verified against the LIVE agent row, not the seed
-- (118_code_indexer_for_analyser.sql is stale and still shows the pre-2026-07
-- request_repo_analysis wiring):
--   code-indexer.workflow.steps.request_analysis.config
--     = {"language":"go","ref_field":"input_data.ref",
--        "repo_field":"input_data.repo","owner_field":"input_data.owner",
--        "pin_to_index_commit":false}
--   ⇒ input_data.ref is the key this task must supply.
--
-- pin_to_index_commit is already false and MUST stay false: true would pin the
-- fetch to the commit already indexed, so the index could never advance.
--
-- Config, so live immediately — no image, no roll.
-- ============================================================================

BEGIN;

DO $$
DECLARE
    v_ref text;
BEGIN
    SELECT input_data->>'ref' INTO v_ref
    FROM scheduled_tasks WHERE name = 'code-index-refresh';

    IF NOT FOUND THEN
        RAISE EXCEPTION '251: no scheduled_tasks row named code-index-refresh — re-survey before forcing this';
    END IF;

    IF v_ref = '086_experience_loop' THEN
        RAISE EXCEPTION '251: the ref is already pinned to 086_experience_loop — nothing to do';
    END IF;

    RAISE NOTICE '251: ref was %, pinning to 086_experience_loop', COALESCE(v_ref, '(absent — resolved implicitly to HEAD)');
END $$;

UPDATE scheduled_tasks
SET input_data = input_data || '{"ref": "086_experience_loop"}'::jsonb,
    updated_at = now()
WHERE name = 'code-index-refresh';

-- Assert the merge, rather than trusting that `||` did what was intended: a
-- jsonb_set with a literal object here would REPLACE input_data and silently
-- drop repo/owner/language, which is a fleet landmine already recorded.
DO $$
DECLARE
    v jsonb;
BEGIN
    SELECT input_data INTO v FROM scheduled_tasks WHERE name = 'code-index-refresh';
    IF v->>'ref' IS DISTINCT FROM '086_experience_loop'
       OR v->>'repo' IS DISTINCT FROM 'agentchassis'
       OR v->>'owner' IS DISTINCT FROM 'gqls'
       OR v->>'language' IS DISTINCT FROM 'go' THEN
        RAISE EXCEPTION '251: post-update input_data is wrong (%) — the merge dropped a key', v;
    END IF;
    RAISE NOTICE '251: input_data now %', v;
END $$;

COMMIT;
