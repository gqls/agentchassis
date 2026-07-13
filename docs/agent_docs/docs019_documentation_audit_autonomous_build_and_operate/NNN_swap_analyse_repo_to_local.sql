-- NNN_swap_analyse_repo_to_local.sql
--
-- Renumber NNN to the next number in your migration sequence (3-digit prefixes).
--
-- Swaps the diagnose-agent workflow's `analyse_repo` step from the cross-pod
-- `request_repo_analysis` (Kafka -> analyser adapter, which fetches+parses in ITS
-- OWN pod and returns metadata, leaving the diagnose-agent with no checkout) to
-- the in-process `analyse_repo_local` (fetch source to a LOCAL temp dir via the
-- shared read-only tarball fetcher, then analysis.Analyse in-process). The new
-- step leaves repo_analysis.root pointing at a real local checkout, so
-- diagnose_assemble_bundle's ReadSymbolBody can slice symbol bodies and the
-- "repo root not found at repo_analysis.root" failure stops firing.
--
-- DEPENDS ON the image carrying the new action (analyse_repo_local) — deploy that
-- first, THEN run this. The action is registered in
-- platform/orchestration/actions/registry.go (see PATCH_lift_fetcher_and_register.md)
-- and needs GITHUB_READ_TOKEN on the diagnose-agent pod (spawn_actions.go already
-- injects it for isRepoCloningAgent agents).
--
-- Step shape is an OBJECT keyed by name at default_config -> workflow -> steps
-- (verified against NNN_fix_diagnose_agent_workflow.sql), so this is a targeted
-- jsonb_set on action/config/description; next_step ("lookup_symbols") and
-- output_field ("repo_analysis") are preserved untouched. Re-running is safe.
--
-- WHAT CHANGES on the step:
--   action       request_repo_analysis  ->  analyse_repo_local
--   config       + pin_to_index_commit:true   (keeps language/owner/repo/ref fields)
--   description  reworded to the in-process fetch+analyse
-- UNCHANGED: next_step=lookup_symbols, output_field=repo_analysis, start_step.

-- BACKUP FIRST (standing rule): snapshot the row before changing it.
SELECT snapshot_agent('diagnose-agent',
  'swap analyse_repo: request_repo_analysis -> analyse_repo_local (in-process fetch+analyse, local checkout, pin index commit)');

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        jsonb_set(
          jsonb_set(
            default_config,
            '{workflow,steps,analyse_repo,action}',
            '"analyse_repo_local"'::jsonb,
            false),
          '{workflow,steps,analyse_repo,config}',
          '{
             "language": "go",
             "ref_field": "input_data.ref",
             "repo_field": "input_data.repo",
             "owner_field": "input_data.owner",
             "pin_to_index_commit": true
           }'::jsonb,
          false),
        '{workflow,steps,analyse_repo,description}',
        '"Fetch repo source to a local temp dir (tarball, read-only token) and analyse IN-PROCESS, so repo_analysis.root is a real local checkout the assembler reads bodies from; pins the code_symbols commit; self-contained (no analyser-adapter round-trip)."'::jsonb,
        false),
    updated_at = now()
WHERE type = 'diagnose-agent';

-- Verify:
--   SELECT (default_config->'workflow'->'steps'->'analyse_repo'->>'action')                       AS action,
--          (default_config->'workflow'->'steps'->'analyse_repo'->'config'->>'pin_to_index_commit') AS pin,
--          (default_config->'workflow'->'steps'->'analyse_repo'->'config'->>'repo_field')           AS repo_field,
--          (default_config->'workflow'->'steps'->'analyse_repo'->>'next_step')                      AS next_step,
--          (default_config->'workflow'->'steps'->'analyse_repo'->>'output_field')                   AS output_field
--   FROM agent_definitions WHERE type = 'diagnose-agent';
-- expect: action=analyse_repo_local, pin=true, repo_field=input_data.repo,
--         next_step=lookup_symbols, output_field=repo_analysis

-- ─────────────────────────────────────────────────────────────────────────────
-- REVERT (back to the cross-pod analyser call). Run if analyse_repo_local must be
-- rolled back (e.g. the image is pulled). Restores the exact prior step body.
-- ─────────────────────────────────────────────────────────────────────────────
-- SELECT snapshot_agent('diagnose-agent','revert analyse_repo_local -> request_repo_analysis');
-- UPDATE agent_definitions
-- SET default_config =
--       jsonb_set(
--         jsonb_set(
--           jsonb_set(
--             default_config,
--             '{workflow,steps,analyse_repo,action}',
--             '"request_repo_analysis"'::jsonb, false),
--           '{workflow,steps,analyse_repo,config}',
--           '{"language":"go","ref_field":"input_data.ref","repo_field":"input_data.repo","owner_field":"input_data.owner"}'::jsonb, false),
--         '{workflow,steps,analyse_repo,description}',
--         '"Analyse the repo at ref; awaits the analyser adapter (Output incl. root + commit_sha)"'::jsonb, false),
--     updated_at = now()
-- WHERE type = 'diagnose-agent';
