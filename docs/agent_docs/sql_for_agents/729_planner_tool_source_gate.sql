-- 729_planner_tool_source_gate.sql
--
-- bugs_open/450 (council corr 4e7497ed-62ed-4426-a814-8361754c2352): the SUPPLY half.
-- The planner names tool pages BEFORE their tools exist — tools arrive from the ~3h
-- design rotation hours-to-days later and under names the planner never saw
-- (seotools.co.uk: 0 of 7 planned names matched what tool-deployer built) — so the page
-- row is minted at plan time, links point at it, and generic producers fill it with prose
-- about a tool that is not there. The Go half (chassis, commit 5e6fee47b) adds an opt-in
-- gate to validate_site_plan: a page_type='tool' page with no live component_level='tool'
-- component on this site is held out of the plan and filed as a capability_gap naming
-- tool-builder. THIS MIGRATION IS THAT GATE'S ONE LIVE CONSUMER (RFC_022 shape: opt-in
-- key, unsafe default OFF, enumerated here):
--
--   1. sets enforce_tool_sources: true on build-site-planner's validate_plan step;
--   2. removes `tool` from the prompt's rule-3 empty-sections licence and tells the
--      planner not to plan NEW tool pages at all.
--
-- WHY THE PROMPT SAYS "DO NOT PLAN" RATHER THAN "NAME THE TOOL COMPONENT". The planner's
-- load_components step loads `component_level IN ('section','element')` — tool components
-- are NEVER in its menu, and rule 1 forbids naming a component that is not there. So
-- advertise.co.uk's plan did not name its tools by the model choosing them: the preserve
-- guard echoed its already-realised composition back into the plan. Telling the model to
-- name a tool component would be telling it to do something it cannot do.
--
-- WHY THE LICENCE IS REMOVED RATHER THAN NARROWED. A sectionless tool page does not park
-- harmlessly: bugs_open/450's websitepromotion instance parked 7 unbuilt_internal_link
-- items in a HUMAN queue (mark_no_ready_sections) plus a needs_content_page, recurring per
-- remake, on a row §7 proved no producer will ever fill. Both branches of that fork are
-- liability, so the gate holds sectioned and sectionless tool pages alike and the licence
-- has nothing left to license.
--
-- THE ARMING LICENCE IS A MEASUREMENT, and it is the thing to re-check if this misbehaves:
-- holding a planner tool stub starves nothing, because tool-deployer CREATES ITS OWN page
-- rows under its own names and nothing reads planned tool pages to decide what to build
-- (bugs_open/450 §7, measured at the rows 2026-09-03). If that is ever falsified, disarm
-- by setting the key to false — one jsonb_set, no code change.
--
-- ORDER-SAFE either way round, same argument as 720: the Go gate reads the key (old
-- binaries ignore it; a new binary without the key = today's behaviour), and the prompt
-- change alone only makes the planner more conservative.
--
-- ⚠ INTERACTION WITH 720_ROLLBACK: this migration rewrites the sentence 720 installed.
-- After 729 applies, 720_ROLLBACK's anchor no longer appears verbatim and it will refuse
-- by its own design — resolve that by hand, in that order, if you ever need to unwind
-- both. 729_ROLLBACK restores 720's exact sentence and deletes only this migration's key.
--
-- Apply: psql -f THIS FILE ONLY (never an unscoped runner --apply). Companion ROLLBACK
-- alongside. Working docs: docs024_key_docs_latest/bugfix_450_tool_page_shells/.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '729 REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '729_planner_tool_source_gate.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int; vars_before int;
  -- Anchored against the LIVE text as it stands AFTER 720 (verified 2026-09-03), never
  -- against the seed: two other lanes edit this row (718 imagery, 720 listing) and the
  -- seed has diverged from live.
  anchor_A text := $A729$3. Pages with page_type entity-page, tool, blog-index, blog-post may have empty sections arrays.$A729$;
  repl_A text := $RA729$3. Pages with page_type entity-page, blog-index, blog-post may have empty sections arrays. Do NOT plan a page with page_type tool: interactive tools are built by the tool pipeline, which creates the tool and its page together and only once the tool exists, so a planned tool page with no tool ships as prose about the tool. Tool pages the site already has are preserved automatically with their existing sections — keep those. Validation holds back tool pages whose tool does not exist for this site and records each as a capability gap naming tool-builder.$RA729$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '729: plan_site.config.prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, anchor_A, ''))) / length(anchor_A);
  IF n <> 1 THEN RAISE EXCEPTION '729: anchor A found % times, expected 1 (has 720 been applied? has another lane rewritten rule 3?)', n; END IF;

  ifs_before   := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends_before  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');
  vars_before  := (length(tpl) - length(replace(tpl, '{{.', ''))) / length('{{.');

  newtpl := replace(tpl, anchor_A, repl_A);

  IF length(newtpl) <> length(tpl) + (length(repl_A) - length(anchor_A)) THEN
    RAISE EXCEPTION '729: unexpected length delta';
  END IF;
  -- A new {{.Var}} without a matching input_fields entry renders EMPTY (the template runs
  -- under text/template's default missingkey, so a bare print emits "<no value>").
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars_before THEN
    RAISE EXCEPTION '729: replacement introduced a template variable — it would render EMPTY without input_fields';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '729: template if/else/end balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           jsonb_set(default_config,
             '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
           '{workflow,steps,validate_plan,config,enforce_tool_sources}', 'true'::jsonb, true),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '729: updated % rows, expected exactly 1', n; END IF;
END $do$;

-- Verify: licence narrowed, new rule present, this gate's flag ON — AND every neighbouring
-- lane's surface still intact. The last three are the point: three lanes edit this one row,
-- so a migration that silently ate another's sentence would look like a clean apply.
DO $$
DECLARE tpl text; flag text; listing_flag text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}',
         default_config #>> '{workflow,steps,validate_plan,config,enforce_tool_sources}',
         default_config #>> '{workflow,steps,validate_plan,config,enforce_listing_sources}'
    INTO tpl, flag, listing_flag
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('entity-page, tool, blog-index, blog-post may have empty sections arrays' in tpl) > 0 THEN
    RAISE EXCEPTION '729 VERIFY: tool is still licensed for empty sections arrays';
  END IF;
  IF position('Do NOT plan a page with page_type tool' in tpl) = 0 THEN
    RAISE EXCEPTION '729 VERIFY: new tool rule missing';
  END IF;
  IF position('records each as a capability gap naming what to enable' in tpl) = 0 THEN
    RAISE EXCEPTION '729 VERIFY: 720 listing-source rule went missing — refuse';
  END IF;
  IF position('Do NOT plan a directory page or section for a kind' in tpl) = 0 THEN
    RAISE EXCEPTION '729 VERIFY: 433 directory rule went missing — refuse';
  END IF;
  IF position('Content-carrying imagery is EXPECTED' in tpl) = 0 THEN
    RAISE EXCEPTION '729 VERIFY: 718 imagery surface went missing — refuse';
  END IF;
  IF flag IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '729 VERIFY: enforce_tool_sources not set (got %)', COALESCE(flag,'NULL');
  END IF;
  IF listing_flag IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '729 VERIFY: enforce_listing_sources is no longer true (got %) — 720 was disarmed', COALESCE(listing_flag,'NULL');
  END IF;
END $$;

COMMIT;
