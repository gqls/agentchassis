-- 729_planner_tool_source_gate_ROLLBACK.sql
--
-- Undoes 729: restores rule 3 to EXACTLY the sentence 720 left, and DELETES the
-- enforce_tool_sources key (absent = default false = the Go gate never runs).
--
-- Deletes rather than sets false, deliberately: an absent key and `false` behave
-- identically in the Go half (`config["enforce_tool_sources"].(bool)` yields false either
-- way), but an absent key leaves no trace of a gate that is no longer wanted, which is the
-- honest post-rollback state. 720's enforce_listing_sources is NOT touched.
--
-- ⚠ IF YOU ARE UNWINDING BOTH GATES, RUN THIS ONE FIRST. 729 rewrote the sentence 720
-- installed, so while 729 is applied 720_ROLLBACK's anchor does not appear verbatim and it
-- refuses by its own design. This file restores 720's text, after which 720_ROLLBACK
-- matches again.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '729 ROLLBACK REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '729_planner_tool_source_gate_ROLLBACK.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int; vars_before int;
  -- applied = what 729 installed; restore = what 720 left (byte-exact).
  applied text := $AP729$3. Pages with page_type entity-page, blog-index, blog-post may have empty sections arrays. Do NOT plan a page with page_type tool: interactive tools are built by the tool pipeline, which creates the tool and its page together and only once the tool exists, so a planned tool page with no tool ships as prose about the tool. Tool pages the site already has are preserved automatically with their existing sections — keep those. Validation holds back tool pages whose tool does not exist for this site and records each as a capability gap naming tool-builder.$AP729$;
  restore text := $RS729$3. Pages with page_type entity-page, tool, blog-index, blog-post may have empty sections arrays.$RS729$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '729 ROLLBACK: plan_site.config.prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, applied, ''))) / length(applied);
  IF n <> 1 THEN
    RAISE EXCEPTION '729 ROLLBACK: 729 text found % times, expected 1 — either 729 was never applied, or a later migration rewrote rule 3. Resolve by hand; do NOT force.', n;
  END IF;

  ifs_before   := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends_before  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');
  vars_before  := (length(tpl) - length(replace(tpl, '{{.', ''))) / length('{{.');

  newtpl := replace(tpl, applied, restore);

  IF length(newtpl) <> length(tpl) + (length(restore) - length(applied)) THEN
    RAISE EXCEPTION '729 ROLLBACK: unexpected length delta';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars_before
     OR (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '729 ROLLBACK: template token balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
             '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false)
             #- '{workflow,steps,validate_plan,config,enforce_tool_sources}',
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '729 ROLLBACK: updated % rows, expected exactly 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text; flag text; listing_flag text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}',
         default_config #>> '{workflow,steps,validate_plan,config,enforce_tool_sources}',
         default_config #>> '{workflow,steps,validate_plan,config,enforce_listing_sources}'
    INTO tpl, flag, listing_flag
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('Do NOT plan a page with page_type tool' in tpl) > 0 THEN
    RAISE EXCEPTION '729 ROLLBACK VERIFY: the tool rule is still present';
  END IF;
  IF position('entity-page, tool, blog-index, blog-post may have empty sections arrays' in tpl) = 0 THEN
    RAISE EXCEPTION '729 ROLLBACK VERIFY: 720-era rule 3 was not restored';
  END IF;
  IF flag IS NOT NULL THEN
    RAISE EXCEPTION '729 ROLLBACK VERIFY: enforce_tool_sources still present (got %)', flag;
  END IF;
  -- The whole point of the sibling-key design: unwinding this gate must not touch 444's.
  IF listing_flag IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '729 ROLLBACK VERIFY: enforce_listing_sources is no longer true (got %) — 720 was collateral damage', COALESCE(listing_flag,'NULL');
  END IF;
  IF position('records each as a capability gap naming what to enable' in tpl) = 0 THEN
    RAISE EXCEPTION '729 ROLLBACK VERIFY: 720 listing rule went missing — refuse';
  END IF;
END $$;

COMMIT;
