-- 250: bugs_open/108 defect A — record WHAT the index describes, not just when it was written.
--
-- The freshness verdict in diagnose_code_lookup_action.go used to key on
-- updated_at, which every refresh resets even when it re-fetches the same
-- pushed tip — measured live 2026-07-28: banner said "refreshed 17h ago" about
-- a commit 1,003 commits behind local HEAD. These two columns are the facts the
-- reworked verdict keys on instead:
--
--   ref         — the ref the indexer fetched (analyse_repo_local exports it in
--                 repo_analysis.ref; the cadence task's pre_query derives it).
--                 The index mirrors this ref's last PUSHED tip, never local work.
--   commit_time — the indexed commit's own committer date (resolved by
--                 reposource.GitHubSource.CommitInfo in the spawned indexer pod,
--                 which holds the read-only token; the shared chassis pod still
--                 only reads this table). Re-running the indexer can reset
--                 updated_at; it can never reset this.
--
-- Both NULLABLE, deliberately: NULL is the honest "unrecorded" state (pre-250
-- rows, or the CommitInfo call failed) and the banner renders it as a loud
-- UNKNOWN, never as fresh. No backfill: the next reindex populates every
-- surviving row, and inventing a commit_time for rows we did not fetch would
-- recreate the defect this migration removes.
--
-- Inert until the chassis image carrying the 108 candidate-1 code rolls; safe
-- to apply before it (the live binary neither reads nor writes these columns).

ALTER TABLE code_symbols ADD COLUMN IF NOT EXISTS ref text;
ALTER TABLE code_symbols ADD COLUMN IF NOT EXISTS commit_time timestamptz;

COMMENT ON COLUMN code_symbols.ref IS
  'ref the indexer fetched (e.g. 086_experience_loop) — the index mirrors this ref''s last PUSHED tip (bugs_open/108)';
COMMENT ON COLUMN code_symbols.commit_time IS
  'committer date of commit_sha — the freshness verdict keys on THIS, never on updated_at (bugs_open/108 defect A)';
