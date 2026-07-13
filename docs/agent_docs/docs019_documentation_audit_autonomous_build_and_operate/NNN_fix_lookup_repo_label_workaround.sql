-- NNN_fix_lookup_repo_label_workaround.sql
--
-- TEST-NOW ENABLER (config-only, NO rebuild). Pairs with the structural fix in
-- PATCH_code_symbols_shared_repo_label.md — apply that and REVERT this when the
-- rebuilt image ships (see the revert block at the bottom).
--
-- WHY: index_code_symbols composes the code_symbols.repo label as owner/repo
-- (repo_analysis.owner + "/" + repo_analysis.repo = "gqls/agentchassis"), but
-- lookup_code_symbols does NOT compose — it filters WHERE repo = <repo_field>, and
-- the diagnose workflow set lookup_symbols.repo_field = "repo_analysis.repo" (the
-- BARE name "agentchassis"). So the lookup queried WHERE repo='agentchassis' against
-- rows stored under 'gqls/agentchassis' -> 0 hits -> empty code_results ->
-- assemble_bundle "no scope". (Embeddings present (436/436), key/elements correct
-- (code_results / {path,symbol}); the only fault was the repo label.)
--
-- WHAT: drop lookup_symbols.repo_field (the bare-name path) and set the literal
-- config.repo = "gqls/agentchassis" (resolveRAGConfigField honours config.repo as an
-- explicit override). REPO-SPECIFIC and TEMPORARY — it hard-codes the repo into the
-- agent, which is why the real fix (lookup composing owner/repo like index, for ANY
-- repo) lives in the code patch.
--
-- Does NOT touch the read-only loop. snapshot first (standing rule).

BEGIN;

SELECT snapshot_agent(
  'diagnose-agent',
  'lookup_symbols: query by full owner/repo label (drop repo_field=repo_analysis.repo, set config.repo=gqls/agentchassis) — index stores gqls/agentchassis, lookup was filtering bare agentchassis -> 0 hits -> no scope. Temporary config-only enabler; revert when the composing-lookup image ships.'
);

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config #- '{workflow,steps,lookup_symbols,config,repo_field}',
      '{workflow,steps,lookup_symbols,config,repo}',
      '"gqls/agentchassis"'::jsonb,
      true
    ),
    updated_at = now()
WHERE type = 'diagnose-agent';

-- verify — expect {top_k, query_field, repo:"gqls/agentchassis"}, NO repo_field
SELECT default_config #> '{workflow,steps,lookup_symbols,config}' AS lookup_config
FROM agent_definitions
WHERE type = 'diagnose-agent';

COMMIT;

-- ── REVERT (run when the composing-lookup image is deployed) ──────────────────
-- Removes the hard-coded literal AND the bare repo_field, so the rebuilt
-- lookup_code_symbols composes owner/repo from repo_analysis like index does:
--
-- BEGIN;
-- SELECT snapshot_agent('diagnose-agent','revert lookup repo literal; rely on composing lookup');
-- UPDATE agent_definitions
-- SET default_config =
--       (default_config #- '{workflow,steps,lookup_symbols,config,repo}')
--                       #- '{workflow,steps,lookup_symbols,config,repo_field}',
--     updated_at = now()
-- WHERE type='diagnose-agent';
-- SELECT default_config #> '{workflow,steps,lookup_symbols,config}' FROM agent_definitions WHERE type='diagnose-agent';
-- COMMIT;
