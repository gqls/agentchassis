-- NNN_swap_indexer_to_local_analysis.sql
--
-- §7B (RUNBOOK_code_retrieval_route.md): the code-indexer's analysis step goes
-- through the analyser ADAPTER (request_repo_analysis, Kafka round-trip), which
-- resolved ref "HEAD" to commit 4c2c172 — dated 2025-07-14, a YEAR-old tree
-- (stale checkout on the adapter side; the §6D trigger envelope demonstrably
-- sent ref:"HEAD"). Result: code_symbols holds a faithful index of the July-2025
-- repo (69 files / 436 symbols vs 572 .go files today) with none of the
-- page/section/result_spec symbols the diagnosis loop needs.
--
-- FIX: swap the step to analyse_repo_local (in-process: fresh GitHub tarball at
-- ref + the same internal/analysis walker; already proven in the diagnose
-- workflow). Three compatibility corrections, each read from the deployed code:
--   1. pin_to_index_commit DEFAULTS TO TRUE in analyse_repo_local — an indexer
--      left at the default would re-fetch the OLD dominant index commit and
--      perpetuate the staleness. Set FALSE: the indexer DEFINES the commit.
--   2. analyse_repo_local returns the analysis Output fields at TOP LEVEL of
--      repo_analysis (root/files/... + commit_sha/owner/repo/ref alongside);
--      there is NO .output key. So index_symbols.analysis_field changes
--      "repo_analysis.output" -> "repo_analysis".            [DELIBERATE CHANGE]
--   3. timeout_seconds 300 -> 1800: a full-tree index (~572 files, thousands of
--      symbols) embeds each new symbol via ollama; 300s was sized for 69 files.
--                                                            [DELIBERATE CHANGE]
-- Also: complete.output_fields drops repo_analysis (the full 572-file analysis
-- JSON in the completion response is multi-MB over Kafka; index_result carries
-- the counts).                                               [DELIBERATE CHANGE]
--
-- The step KEEPS its name request_analysis (start_step untouched) — the ACTION
-- changes, the name deliberately does not. No other variable names change.
--
-- Prune note: the first run at the new commit deletes all 436 old rows
-- (commit_sha IS DISTINCT FROM new) — self-cleaning. The diagnose workflow's
-- analyse_repo_local (pin_to_index_commit:true) then picks up the NEW dominant
-- commit automatically; no diagnose-side change needed.
--
-- snapshot first (standing rule). REVERT block at the bottom restores the
-- exact pre-migration step (verbatim from the 2026-07-02 dump).

BEGIN;

SELECT snapshot_agent(
  'code-indexer',
  'swap request_analysis to analyse_repo_local (in-process; adapter resolved HEAD to stale 2025-07-14 commit); analysis_field -> repo_analysis (top-level Output shape); pin_to_index_commit=false (indexer defines the commit); timeout 1800; complete returns index_result only'
);

UPDATE agent_definitions
SET default_config =
  jsonb_set(
    jsonb_set(
      jsonb_set(
        jsonb_set(
          default_config,
          '{workflow,steps,request_analysis}',
          '{
            "action": "analyse_repo_local",
            "config": {
              "language": "go",
              "ref_field": "input_data.ref",
              "repo_field": "input_data.repo",
              "owner_field": "input_data.owner",
              "pin_to_index_commit": false
            },
            "next_step": "index_symbols",
            "description": "Fetch repo tarball at ref and analyse IN-PROCESS (no adapter round-trip). repo_analysis carries the Output fields top-level + owner/repo/ref/commit_sha. pin_to_index_commit=false: the indexer DEFINES the index commit.",
            "output_field": "repo_analysis"
          }'::jsonb
        ),
        '{workflow,steps,index_symbols,config,analysis_field}',
        '"repo_analysis"'::jsonb
      ),
      '{workflow,steps,complete,config,output_fields}',
      '["index_result"]'::jsonb
    ),
    '{timeout_seconds}',
    '1800'::jsonb
  ),
  updated_at = now()
WHERE type = 'code-indexer';

-- verify — expect: action=analyse_repo_local with pin_to_index_commit=false;
-- analysis_field=repo_analysis; output_fields=["index_result"]; timeout 1800
SELECT default_config #> '{workflow,steps,request_analysis}'                 AS request_analysis_step,
       default_config #> '{workflow,steps,index_symbols,config,analysis_field}' AS analysis_field,
       default_config #> '{workflow,steps,complete,config,output_fields}'   AS complete_outputs,
       default_config #> '{timeout_seconds}'                                AS timeout_seconds
FROM agent_definitions
WHERE type = 'code-indexer';

COMMIT;

-- ── REVERT (restores the exact pre-migration config, from the 2026-07-02 dump) ─
--
-- BEGIN;
-- SELECT snapshot_agent('code-indexer','revert to adapter-based request_repo_analysis');
-- UPDATE agent_definitions
-- SET default_config =
--   jsonb_set(
--     jsonb_set(
--       jsonb_set(
--         jsonb_set(
--           default_config,
--           '{workflow,steps,request_analysis}',
--           '{"action":"request_repo_analysis","config":{"language":"go","ref_field":"input_data.ref","repo_field":"input_data.repo","owner_field":"input_data.owner"},"next_step":"index_symbols","description":"Ask the analyser adapter to parse the repo at ref into symbols; await the reply","output_field":"repo_analysis"}'::jsonb
--         ),
--         '{workflow,steps,index_symbols,config,analysis_field}',
--         '"repo_analysis.output"'::jsonb
--       ),
--       '{workflow,steps,complete,config,output_fields}',
--       '["repo_analysis","index_result"]'::jsonb
--     ),
--     '{timeout_seconds}',
--     '300'::jsonb
--   ),
--   updated_at = now()
-- WHERE type = 'code-indexer';
-- SELECT default_config #> '{workflow,steps,request_analysis}' FROM agent_definitions WHERE type='code-indexer';
-- COMMIT;
