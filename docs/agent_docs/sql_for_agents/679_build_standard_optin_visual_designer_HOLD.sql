-- 679_build_standard_optin_visual_designer_HOLD.sql
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
-- visual-designer / design. Siblings: 677/678/679. Supersedes the bundled 676 (never applied).
-- Landmine honoured ("an agent-config migration keyed on type can hit TWO active rows"):
-- the DO block REFUSES unless exactly one active row exists, and snapshot_agent() records
-- the pre-state. Carrier row 675 must exist (verified below).

BEGIN;

SELECT snapshot_agent('visual-designer', '679_build_standard_optin_visual_designer_HOLD.sql: pre-optin');

DO $m$
DECLARE cfg jsonb; tpl text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '679: % active visual-designer rows, want exactly 1 — duplicate-row landmine; find which row the runtime loads before touching either', n; END IF;

  IF NOT EXISTS (SELECT 1 FROM agent_default_configs WHERE config_name='build_standard_block') THEN
    RAISE EXCEPTION '679: carrier row absent — apply 675 first';
  END IF;

  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  tpl := cfg->'workflow'->'steps'->'design'->'config'->>'prompt_template';
  IF tpl IS NULL THEN RAISE EXCEPTION '679: visual-designer design has no prompt_template'; END IF;
  IF position('{{.build_standard}}' IN tpl) > 0 THEN RAISE EXCEPTION '679: visual-designer already opted in — re-census before applying'; END IF;
  IF (length(tpl) - length(replace(tpl, 'Create visual design for: {{.business_type}}.', ''))) / length('Create visual design for: {{.business_type}}.') <> 1 THEN
    RAISE EXCEPTION '679: anchor not-exactly-once — template drifted, re-base';
  END IF;
  tpl := replace(tpl, 'Create visual design for: {{.business_type}}.', 'Create visual design for: {{.business_type}}.' || E'\n\n## {{.build_standard}}\n');
  cfg := jsonb_set(cfg, '{workflow,steps,design,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '679: update touched % rows, want 1', n; END IF;

  IF NOT EXISTS (SELECT 1 FROM agent_definitions
                  WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                    AND default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template' LIKE '%{{.build_standard}}%') THEN
    RAISE EXCEPTION '679 VERIFY: the loaded row does not carry the placeholder';
  END IF;
  RAISE NOTICE '679 verify: visual-designer opted in on its single active row.';
END $m$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('679_build_standard_optin_visual_designer_HOLD.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Build-standard opt-in for visual-designer (PLAN_2026-08-25 s3; one pipeline per file per council round b5a642b7). Held until the 8c62e9f1b roll.');

COMMIT;
