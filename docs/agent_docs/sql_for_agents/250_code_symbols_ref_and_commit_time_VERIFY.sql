-- VERIFY 250 (bugs_open/108 defect A). Run any time; five checks.
-- Checks 1-2 must be green immediately after applying; 3-5 only after the
-- chassis image carrying the candidate-1 code has rolled AND one reindex has
-- completed (cadence, or a manual index-orchestrator dispatch).

-- 1. Columns exist with the right types.
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'code_symbols' AND column_name IN ('ref', 'commit_time')
ORDER BY column_name;
-- expect: commit_time | timestamp with time zone | YES
--         ref         | text                     | YES

-- 2. Before the first post-roll reindex: every row NULL (no invented history).
--    After it: 0 NULLs for the freshly indexed repo.
SELECT count(*) AS rows_total,
       count(ref) AS with_ref,
       count(commit_time) AS with_commit_time
FROM code_symbols;

-- 3. After the first post-roll reindex: the recorded facts, one row.
--    commit_time must equal the committer date of commit_sha (spot-check it
--    against `git show -s --format=%cI <sha>` — the sha is now the FULL 40-char
--    form, upgraded by CommitInfo from the tarball's short form).
SELECT DISTINCT ref, commit_sha, commit_time, max(updated_at) OVER () AS refreshed_at
FROM code_symbols;

-- 4. The verdict's induced check, in SQL: a fresh refresh clock over an old
--    commit must NOT be reported fresh. While origin's tip stays at the
--    2026-07-24 commit, this returns the exact live fault state the banner must
--    flag STALE (the unit test proves the branch; this proves the data drives it).
SELECT (now() - max(updated_at))    < interval '48 hours' AS refresh_clock_looks_fresh,
       (now() - max(commit_time))   > interval '48 hours' AS commit_is_stale,
       max(commit_time) IS NULL     AS commit_time_unrecorded
FROM code_symbols;
-- expect after reindex, until someone pushes: t | t | f  → banner must read STALE
-- expect after a push + reindex:              t | f | f  → banner quiet

-- 5. Negative control (assert-position discipline): a value that must NOT
--    appear. No row may carry a commit_time NEWER than its refresh time — that
--    would mean an invented date, the failure mode NULL exists to prevent.
SELECT count(*) AS impossible_rows
FROM code_symbols
WHERE commit_time IS NOT NULL AND commit_time > updated_at + interval '5 minutes';
-- expect: 0
