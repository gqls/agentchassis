-- 676_build_standard_optins_HOLD.sql
--
-- ⚠ _HOLD — ORDERING-CRITICAL, APPLY BY HAND ONLY AFTER THE ROLL. A template that names
-- {{.build_standard}} before the chassis binary injects it renders the literal string
-- "<no value>" into live LLM prompts (text/template default missingkey; verified at
-- RenderPromptTemplate — no Option set). The runner's SIDECAR_RE excludes _HOLD files, which
-- is why this is named so. PRECONDITION PROBE (per service, at the artefact): the running
-- agent-chassis stamp must be a descendant of the voicestyle-generalisation commit —
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <that commit> <the stamp>
-- (empty log window means scrolled, probe the binary per CLAUDE.md — never assume).
--
-- WHAT IT DOES: the three plan/design surfaces from PLAN_2026-08-25 §3 opt in to the build
-- standard — build-site-planner (plan_site), content-gap-planner (plan_gaps), visual-designer
-- (design). The writer deliberately does NOT opt in (the standard reaches it through
-- briefs/strategy; raw injection into the copy writer is the demonstration hazard this lane
-- exists to manage). Exact-anchor replaces, each asserted before and after: the corpus MOVES.

BEGIN;

DO $m$
DECLARE cfg jsonb; tpl text; n int;
BEGIN
  -- ── build-site-planner / plan_site ─────────────────────────────────────
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN RAISE EXCEPTION '676: no active build-site-planner row'; END IF;
  tpl := cfg->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template';
  IF tpl IS NULL THEN RAISE EXCEPTION '676: build-site-planner plan_site has no prompt_template'; END IF;
  IF position('{{.build_standard}}' IN tpl) > 0 THEN RAISE EXCEPTION '676: build-site-planner already opted in — re-census before applying'; END IF;
  IF (length(tpl) - length(replace(tpl, 'Plan a website for {{.input_data.domain}}.', ''))) / length('Plan a website for {{.input_data.domain}}.') <> 1 THEN
    RAISE EXCEPTION '676: build-site-planner anchor not-exactly-once — template drifted, re-base';
  END IF;
  tpl := replace(tpl, 'Plan a website for {{.input_data.domain}}.',
                      'Plan a website for {{.input_data.domain}}.' || E'\n\n## {{.build_standard}}\n');
  cfg := jsonb_set(cfg, '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '676: build-site-planner update touched % rows, want 1', n; END IF;

  -- ── content-gap-planner / plan_gaps ────────────────────────────────────
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='content-gap-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN RAISE EXCEPTION '676: no active content-gap-planner row'; END IF;
  tpl := cfg->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template';
  IF tpl IS NULL THEN RAISE EXCEPTION '676: content-gap-planner plan_gaps has no prompt_template'; END IF;
  IF position('{{.build_standard}}' IN tpl) > 0 THEN RAISE EXCEPTION '676: content-gap-planner already opted in — re-census before applying'; END IF;
  IF (length(tpl) - length(replace(tpl, 'You are a site architect planning how to fill content gaps.', ''))) / length('You are a site architect planning how to fill content gaps.') <> 1 THEN
    RAISE EXCEPTION '676: content-gap-planner anchor not-exactly-once — template drifted, re-base';
  END IF;
  tpl := replace(tpl, 'You are a site architect planning how to fill content gaps.',
                      'You are a site architect planning how to fill content gaps.' || E'\n\n## {{.build_standard}}\n');
  cfg := jsonb_set(cfg, '{workflow,steps,plan_gaps,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='content-gap-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '676: content-gap-planner update touched % rows, want 1', n; END IF;

  -- ── visual-designer / design ───────────────────────────────────────────
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN RAISE EXCEPTION '676: no active visual-designer row'; END IF;
  tpl := cfg->'workflow'->'steps'->'design'->'config'->>'prompt_template';
  IF tpl IS NULL THEN RAISE EXCEPTION '676: visual-designer design has no prompt_template'; END IF;
  IF position('{{.build_standard}}' IN tpl) > 0 THEN RAISE EXCEPTION '676: visual-designer already opted in — re-census before applying'; END IF;
  IF (length(tpl) - length(replace(tpl, 'Create visual design for: {{.business_type}}.', ''))) / length('Create visual design for: {{.business_type}}.') <> 1 THEN
    RAISE EXCEPTION '676: visual-designer anchor not-exactly-once — template drifted, re-base';
  END IF;
  tpl := replace(tpl, 'Create visual design for: {{.business_type}}.',
                      'Create visual design for: {{.business_type}}.' || E'\n\n## {{.build_standard}}\n');
  cfg := jsonb_set(cfg, '{workflow,steps,design,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='visual-designer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '676: visual-designer update touched % rows, want 1', n; END IF;
END $m$;

DO $v$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND type IN ('build-site-planner','content-gap-planner','visual-designer')
     AND default_config::text LIKE '%{{.build_standard}}%';
  IF n <> 3 THEN RAISE EXCEPTION '676 VERIFY: % of 3 targets opted in', n; END IF;
  -- the writer must NOT have gained it (the plan routes it indirectly)
  IF EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                AND default_config::text LIKE '%{{.build_standard}}%') THEN
    RAISE EXCEPTION '676 VERIFY: the writer opted in — that is not this migration''s licence';
  END IF;
  -- the carrier must exist (675 first)
  IF NOT EXISTS (SELECT 1 FROM agent_default_configs WHERE config_name='build_standard_block') THEN
    RAISE EXCEPTION '676 VERIFY: carrier row absent — apply 675 first';
  END IF;
  RAISE NOTICE '676 verify: 3 opt-ins in, writer clean, carrier present.';
END $v$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('676_build_standard_optins_HOLD.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Build-standard opt-ins for the three plan/design surfaces (PLAN_2026-08-25 s3). HELD until the voicestyle-generalisation roll; precondition probe in the header.');

COMMIT;
