-- 200_implementer_plan_paths_are_law.sql
--
-- B4 round-4 finding (2026-07-24, orch 0fe15199): the model relocated the
-- planned file internal/tools-api/config.go to internal/tools-api/config/config.go
-- (its own package) and the deterministic allowlist rightly refused. The
-- 199-added rule 8's OWN EXAMPLE ("github.com/gqls/agentchassis/internal/
-- tools-api/config") suggested a config package directory, contradicting the
-- plan's file list — the rule that fixed imports seeded a path deviation.
-- Own-goal recorded in WRONG_CALLS-adjacent NOTES; this migration corrects the
-- example and makes plan-paths-are-law explicit.
-- ROLLBACK: snapshot below.

BEGIN;

SELECT snapshot_agent('feature-implementer', '200_implementer_plan_paths_are_law: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,stage_implement,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'stage_implement'->'config'->>'prompt_template',
           '8. The Go module is `github.com/gqls/agentchassis` — EVERY internal import is `github.com/gqls/agentchassis/<dir>/...` (e.g. `github.com/gqls/agentchassis/platform/health`, `github.com/gqls/agentchassis/internal/tools-api/config`). NEVER invent a module or org name; if you are unsure of a package''s import path, derive it from the repo-relative directory under this module.',
           '8. The Go module is `github.com/gqls/agentchassis` — EVERY internal import is `github.com/gqls/agentchassis/<directory-of-the-file>` , where the directory comes from THE PLAN''S OWN FILE PATHS. NEVER invent a module or org name. NEVER relocate, rename or re-package a planned file to make an import path nicer (e.g. if the plan says `internal/tools-api/config.go`, the file lives THERE, in the package of that directory — do NOT create `internal/tools-api/config/config.go`): the plan''s paths are a deterministic allowlist and any deviation is rejected wholesale.'
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
  IF position('NEVER relocate, rename or re-package a planned file' in rp) = 0 THEN
    RAISE EXCEPTION '200: plan-paths-are-law rule did not land';
  END IF;
  IF position('internal/tools-api/config/config.go' in rp) = 0 THEN
    RAISE EXCEPTION '200: negative example missing';
  END IF;
  IF position('github.com/gqls/agentchassis' in rp) = 0 THEN
    RAISE EXCEPTION '200: module rule lost';
  END IF;
END $$;

COMMIT;
