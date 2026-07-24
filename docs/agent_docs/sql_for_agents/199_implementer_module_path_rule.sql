-- 199_implementer_module_path_rule.sql
--
-- B4 first-fire finding (2026-07-24, corr 278a37c3 rounds 2-3): the
-- feature-implementer's stage_implement prompt never states the repo's Go
-- module path, so generated files imported an INVENTED module
-- (github.com/resistance-app/platform/...) — which would fail the build gate
-- on the first Go stage even after the formatGeneratedGo shape fix (council
-- corr 6bf3806f). Deterministic gap: nothing in the prompt names the module.
--
-- Fix: append Hard rule 8 pinning the module path + internal-import shape.
-- Config-only, live immediately. (198 deliberately skipped: reserved by the
-- in-flight feature PR's rounds-table migration file.)
-- ROLLBACK: snapshot below.

BEGIN;

SELECT snapshot_agent('feature-implementer', '199_implementer_module_path_rule: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,stage_implement,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'stage_implement'->'config'->>'prompt_template',
           '7. The stage''s expected symbols MUST appear in your output — they are checked deterministically: {{.stage.expected_symbols}}',
           '7. The stage''s expected symbols MUST appear in your output — they are checked deterministically: {{.stage.expected_symbols}}' || chr(10) ||
           '8. The Go module is `github.com/gqls/agentchassis` — EVERY internal import is `github.com/gqls/agentchassis/<dir>/...` (e.g. `github.com/gqls/agentchassis/platform/health`, `github.com/gqls/agentchassis/internal/tools-api/config`). NEVER invent a module or org name; if you are unsure of a package''s import path, derive it from the repo-relative directory under this module.'
         )),
         true)
 WHERE type='feature-implementer'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE rp text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'stage_implement'->'config'->>'prompt_template' INTO rp
    FROM agent_definitions
   WHERE type='feature-implementer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('github.com/gqls/agentchassis' in rp) = 0 THEN
    RAISE EXCEPTION '199: module-path rule did not land';
  END IF;
  IF position('checked deterministically' in rp) = 0 THEN
    RAISE EXCEPTION '199: rule 7 disturbed';
  END IF;
  IF position('{"files": [{"path"' in rp) = 0 THEN
    RAISE EXCEPTION '199: output-shape spec disturbed';
  END IF;
END $$;

COMMIT;
