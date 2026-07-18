-- 0NN_fix_implementer_ref_input_F1_2.sql — F1.2 cleanup (2026-07-18).
-- Make the fix-implementer's base branch a PER-RUN INPUT, not a stale literal.
--
-- The def hardcoded '084_site_improvements_local_ai' in THREE places — stale the
-- moment the active branch moved to 085 (an implementer run today would READ code
-- from, cut its fix/* branch FROM, and target its PR AT a dead branch):
--   read_current_files.config.ref                       (where it reads current code)
--   prepare.config.base_branch                          (the PR base)
--   create_branch.config.data_literals.from_branch      (the branch the fix is cut from)
--
-- This wires all three to input_data.base_branch (passed by 092_TRIGGER, default
-- main), with a graceful literal fallback to 'main' so an un-set run still works:
--   * read_current_files: ref_field -> input_data.base_branch (Go already supports
--     ref_field; literal ref set to 'main' as the fallback).
--   * prepare: base_branch_field -> input_data.base_branch (needs the image carrying
--     the base_branch_field Go change; literal base_branch set to 'main' fallback).
--   * create_branch: move from_branch out of data_literals into data_fields ->
--     input_data.base_branch (config-only; data_fields resolves from collected_data).
--
-- ██ SEQUENCING ██ apply AFTER an image carrying the base_branch_field change is
-- live (verify: strings /app/agent-chassis | grep -c base_branch_field in the pod).
-- On an older image prepare ignores base_branch_field and falls back to the literal
-- 'main' — safe, just not per-run for that step until the image lands.
--
-- PATCH-STYLE, idempotent. Backs up first.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_agentdef_fiximpl_F1_2_20260718 AS
  SELECT *, now() AS backed_up_at FROM agent_definitions WHERE type='fix-implementer';

-- 1. read_current_files: ref_field -> input_data.base_branch; literal ref -> main.
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,read_current_files,config,ref_field}', '"input_data.base_branch"'),
        '{workflow,steps,read_current_files,config,ref}', '"main"'),
    updated_at = now()
WHERE type='fix-implementer';

-- 2. prepare: base_branch_field -> input_data.base_branch; literal base_branch -> main.
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,prepare,config,base_branch_field}', '"input_data.base_branch"'),
        '{workflow,steps,prepare,config,base_branch}', '"main"'),
    updated_at = now()
WHERE type='fix-implementer';

-- 3. create_branch: from_branch becomes a data_field (dynamic), removed from literals.
UPDATE agent_definitions
SET default_config =
      jsonb_set(
        (default_config #- '{workflow,steps,create_branch,config,data_literals,from_branch}'),
        '{workflow,steps,create_branch,config,data_fields,from_branch}',
        '"input_data.base_branch"'),
    updated_at = now()
WHERE type='fix-implementer';

COMMIT;

-- Verify (run after):
--   SELECT default_config #>> '{workflow,steps,read_current_files,config,ref_field}',
--          default_config #>> '{workflow,steps,prepare,config,base_branch_field}',
--          default_config #>> '{workflow,steps,create_branch,config,data_fields,from_branch}',
--          (default_config::text ~ '084_site_improvements') AS still_has_stale_084
--   FROM agent_definitions WHERE type='fix-implementer';
--   -- all three = input_data.base_branch ; still_has_stale_084 = f
