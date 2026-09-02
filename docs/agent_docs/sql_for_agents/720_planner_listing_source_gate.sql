-- 720_planner_listing_source_gate.sql
--
-- bugs_open/444 (council corr c0990eb3-9f50-4e08-b578-a7e05f786945): listing pages ship
-- EMPTY of their own content type, filled with brief-echo prose — 5 pages measured at
-- served bodies across the 2026-09-02 remakes, and every naive check passes. The Go half
-- (chassis) adds an opt-in gate to validate_site_plan: a listing-family page whose item
-- source resolves to ZERO for the site is held out of the plan and filed as a
-- capability_gap naming the missing producer. This migration is the gate's one live
-- consumer (RFC_022 shape: opt-in key, unsafe default OFF, enumerated here):
--
--   1. sets enforce_listing_sources: true on build-site-planner's validate_plan step;
--   2. narrows the prompt's rule 3 — the live licence for the bug: "Pages with page_type
--      entity-directory, entity-page, tool, blog-index, blog-post may have empty sections
--      arrays" let a listing page reach the plan carrying no producer at all, and
--      contradicted 433's own "a page for a kind the site has not opted into ships empty".
--      (The seed's stronger sentence — "Plan the IDEAL site regardless" — is NOT in the
--      live row; checked 2026-09-02. Anchors are against the LIVE text, per LANDMINES:
--      anchored replace with exact-count guards, never wholesale jsonb_set.)
--
-- ORDER-SAFE either way round: the Go gate reads the key (old binaries ignore it; new
-- binary without the key = today's behaviour), and the prompt change alone only makes the
-- planner more conservative.
--
-- Apply: psql -f THIS FILE ONLY (never an unscoped runner --apply). Companion ROLLBACK
-- alongside. Working docs: docs024_key_docs_latest/bugfix_444_empty_listing_pages/.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '720 REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '720_planner_listing_source_gate.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int; vars_before int;
  anchor_A text := $A720$3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays$A720$;
  repl_A text := $RA720$3. Pages with page_type entity-page, tool, blog-index, blog-post may have empty sections arrays. An entity-directory page should name its listing component in sections. A LISTING page — news-index, entity-directory, section-index, or any page whose purpose is a list of items — may only be planned when the site's item source for it exists: a recommended news feed (or seeded sources) for a news page; an opted-in directory kind or a configured business directory for a directory page; child pages in this same plan for a section index. Do NOT plan a glossary, showcase or similar collection page unless the brief names a live producer for its items — a listing page with no item source ships as prose about itself. Validation holds back listing pages whose item source resolves to zero and records each as a capability gap naming what to enable$RA720$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '720: plan_site.config.prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, anchor_A, ''))) / length(anchor_A);
  IF n <> 1 THEN RAISE EXCEPTION '720: anchor A found % times, expected 1', n; END IF;

  ifs_before   := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends_before  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');
  vars_before  := (length(tpl) - length(replace(tpl, '{{.', ''))) / length('{{.');

  newtpl := replace(tpl, anchor_A, repl_A);

  IF length(newtpl) <> length(tpl) + (length(repl_A) - length(anchor_A)) THEN
    RAISE EXCEPTION '720: unexpected length delta';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars_before THEN
    RAISE EXCEPTION '720: replacement introduced a template variable — it would render EMPTY without input_fields';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '720: template if/else/end balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           jsonb_set(default_config,
             '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
           '{workflow,steps,validate_plan,config,enforce_listing_sources}', 'true'::jsonb, true),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '720: updated % rows, expected exactly 1', n; END IF;
END $do$;

-- Verify (DO/RAISE, induced-failure tested locally): licence narrowed, gate flag ON,
-- 433's directory rule untouched, 718's imagery surfaces untouched.
DO $$
DECLARE tpl text; flag text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}',
         default_config #>> '{workflow,steps,validate_plan,config,enforce_listing_sources}'
    INTO tpl, flag
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays' in tpl) > 0 THEN
    RAISE EXCEPTION '720 VERIFY: old rule-3 licence still present';
  END IF;
  IF position('records each as a capability gap naming what to enable' in tpl) = 0 THEN
    RAISE EXCEPTION '720 VERIFY: new listing-source rule missing';
  END IF;
  IF position('Do NOT plan a directory page or section for a kind' in tpl) = 0 THEN
    RAISE EXCEPTION '720 VERIFY: 433 directory rule went missing — refuse';
  END IF;
  IF position('Content-carrying imagery is EXPECTED' in tpl) = 0 THEN
    RAISE EXCEPTION '720 VERIFY: 718 imagery surface went missing — refuse';
  END IF;
  IF flag IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '720 VERIFY: enforce_listing_sources not set (got %)', COALESCE(flag,'NULL');
  END IF;
END $$;

COMMIT;
