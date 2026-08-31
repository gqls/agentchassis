-- 677_build_standard_optin_build_site_planner_HOLD.sql
--
-- ⚠ _HOLD — ORDERING-CRITICAL, APPLY BY HAND ONLY AFTER THE ROLL carrying the voicestyle
-- generalisation (commit 8c62e9f1b). A template naming {{.build_standard}} before the roll
-- renders literal '<no value>' into live prompts (RenderPromptTemplate does WARN loudly on that
-- literal — "TEMPLATE RENDERED WITH MISSING DATA" — but the prompt still ships). Precondition
-- probe, per service at the artefact:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor 8c62e9f1b <the stamp>
--
-- ONE pipeline per file (council round b5a642b7, guardian): this file touches ONLY
-- build-site-planner / plan_site. Siblings: 677/678/679. Supersedes the bundled 676 (never applied).
-- Landmine honoured ("an agent-config migration keyed on type can hit TWO active rows"):
-- the DO block REFUSES unless exactly one active row exists, and snapshot_agent() records
-- the pre-state. Carrier row 675 must exist (verified below).

BEGIN;

SELECT snapshot_agent('build-site-planner', '677_build_standard_optin_build_site_planner_HOLD.sql: pre-optin');

DO $m$
DECLARE cfg jsonb; tpl text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '677: % active build-site-planner rows, want exactly 1 — duplicate-row landmine; find which row the runtime loads before touching either', n; END IF;

  IF NOT EXISTS (SELECT 1 FROM agent_default_configs WHERE config_name='build_standard_block') THEN
    RAISE EXCEPTION '677: carrier row absent — apply 675 first';
  END IF;

  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  tpl := cfg->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template';
  IF tpl IS NULL THEN RAISE EXCEPTION '677: build-site-planner plan_site has no prompt_template'; END IF;
  IF position('{{.build_standard}}' IN tpl) > 0 THEN RAISE EXCEPTION '677: build-site-planner already opted in — re-census before applying'; END IF;
  IF (length(tpl) - length(replace(tpl, 'Plan a website for {{.input_data.domain}}.', ''))) / length('Plan a website for {{.input_data.domain}}.') <> 1 THEN
    RAISE EXCEPTION '677: anchor not-exactly-once — template drifted, re-base';
  END IF;
  tpl := replace(tpl, 'Plan a website for {{.input_data.domain}}.', 'Plan a website for {{.input_data.domain}}.' || E'\n\n## {{.build_standard}}\n');
  cfg := jsonb_set(cfg, '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '677: update touched % rows, want 1', n; END IF;

  IF NOT EXISTS (SELECT 1 FROM agent_definitions
                  WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                    AND default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template' LIKE '%{{.build_standard}}%') THEN
    RAISE EXCEPTION '677 VERIFY: the loaded row does not carry the placeholder';
  END IF;
  RAISE NOTICE '677 verify: build-site-planner opted in on its single active row.';
END $m$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('677_build_standard_optin_build_site_planner_HOLD.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Build-standard opt-in for build-site-planner (PLAN_2026-08-25 s3; one pipeline per file per council round b5a642b7). Held until the 8c62e9f1b roll.');

COMMIT;
