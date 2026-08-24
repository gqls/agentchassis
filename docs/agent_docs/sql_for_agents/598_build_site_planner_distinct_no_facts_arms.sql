-- 598_build_site_planner_distinct_no_facts_arms.sql
--
-- bugs_open/380 slice S1 (planner half). Migration 329's Verified Facts block ends in two
-- {{else}} arms carrying the IDENTICAL sentence — "No verified facts are registered for this
-- site — use plain string section entries and no facts keys." — one for a register with no
-- facts, one for no register at all. Two consequences, both live on garden-tools.uk:
--   * plain-string sections leave assigned_fact_ids NULL, so the writer's strong
--     facts_scoped arm ("state NO business numbers…") never fires on exactly the sites with
--     nothing verified. The [] path is proven end to end with no Go change: scopeItem
--     (plan_sections_action.go:1149-1188) scopes a non-nil empty list, and every carrier
--     keeps [] distinct from NULL (v3_site_actions.go:4046-4057, write_site_plan_action.go:
--     649-655, load_page_sections_from_spec_action.go:166-175). The section_facts wire is
--     live on page-build-handler [MEASURED 2026-08-24].
--   * nothing tells the planner not to BRIEF a methodology page as practice — garden-tools'
--     largest page described a product-testing operation that has never tested a product.
-- The sentence had NO stated rationale in 329 ("plain strings" was simply the pre-existing
-- behaviour), and rule 17's last sentence mandates the very plain strings the fix retires,
-- so it is edited in the same migration or the prompt contradicts itself.
--
-- Anchors: the bare sentence occurs TWICE, so the composite double-arm block is the anchor
-- (expected exactly once); rule 17's closing sentence is anchored on its own text (once).
-- No new template variables ({{.…}}); {{if}}/{{else}}/{{end}} balance asserted unchanged.
-- Owner decisions 2026-08-24: D1 (no register is minted — absence is the cold posture,
-- which is why BOTH arms must constrain) and the bugs_open/380 adoption. RFC_023: prompt
-- narrowing, no consumer's success path changes — ordinary council gate.
--
-- Coordinated with the bugs_open/381 session (migrations 591-595): their plan_site edits
-- anchor on the component-menu line and rule 18; these anchors are disjoint, both sides
-- splice on the live text, so apply order does not matter.
--
-- Apply: psql -f, then record. Companion ROLLBACK alongside.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '598 REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '598_build_site_planner_distinct_no_facts_arms.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int;
  anchor1 text := '{{end}}{{end}}{{else}}No verified facts are registered for this site — use plain string section entries and no facts keys.{{end}}{{else}}No verified facts are registered for this site — use plain string section entries and no facts keys.{{end}}';
  repl1 text := $R${{end}}{{end}}{{else}}This site's evidence register holds NO verified facts. Use the OBJECT form for EVERY section entry on EVERY page with "facts": [] (rule 17) — nothing may be stated as a verified fact, and no section may state business numbers or named-entity relationships.{{end}}{{else}}NO evidence register exists for this site: nothing about this business is verified and it has no recorded operating history. Use the OBJECT form for EVERY section entry on EVERY page with "facts": [] (rule 17). No page brief may describe testing, buying, measuring, sourcing samples or any other operating practice as something this business does — a methodology or "how we assess" page may be planned only as what the site IS and does for a reader, never as a description of practice.{{end}}$R$;
  anchor2 text := 'When no Verified Facts are listed, use plain strings only.';
  repl2 text := $S$When no Verified Facts are listed, still use the object form with "facts": [] on every section — a plain string there leaves the writer unconstrained, which is the failure this rule exists to prevent (bugs_open/380).$S$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '598: plan_site.config.prompt_template not found'; END IF;

  -- The bare sentence occurs twice, both inside this one composite block, so the
  -- composite itself is the unique anchor.
  n := (length(tpl) - length(replace(tpl, anchor1, ''))) / length(anchor1);
  IF n <> 1 THEN RAISE EXCEPTION '598: composite anchor found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor2, ''))) / length(anchor2);
  IF n <> 1 THEN RAISE EXCEPTION '598: rule-17 anchor found % times, expected 1', n; END IF;

  ifs_before   := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends_before  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');

  newtpl := replace(tpl, anchor1, repl1);
  newtpl := replace(newtpl, anchor2, repl2);

  IF length(newtpl) <> length(tpl)
       + (length(repl1) - length(anchor1))
       + (length(repl2) - length(anchor2)) THEN
    RAISE EXCEPTION '598: unexpected length delta';
  END IF;

  -- No new template variables, and structural balance unchanged.
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) <> (length(tpl) - length(replace(tpl, '{{.', ''))) THEN
    RAISE EXCEPTION '598: replacement introduced a template variable — it would render EMPTY without input_fields';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '598: template if/else/end balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '598: updated % rows, expected exactly 1', n; END IF;
END $do$;

-- Verify (DO/RAISE): both new arms present exactly once, the old sentence gone entirely.
DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  n := (length(tpl) - length(replace(tpl, 'No verified facts are registered for this site — use plain string section entries', '')))
       / length('No verified facts are registered for this site — use plain string section entries');
  IF n <> 0 THEN RAISE EXCEPTION '598 VERIFY: old sentence still present % times', n; END IF;

  IF position($V$evidence register holds NO verified facts$V$ in tpl) = 0 THEN
    RAISE EXCEPTION '598 VERIFY: empty-register arm missing';
  END IF;
  IF position($V$NO evidence register exists for this site$V$ in tpl) = 0 THEN
    RAISE EXCEPTION '598 VERIFY: no-register arm missing';
  END IF;
  IF position($V$still use the object form with "facts": []$V$ in tpl) = 0 THEN
    RAISE EXCEPTION '598 VERIFY: rule-17 amendment missing';
  END IF;
  RAISE NOTICE '598 OK: both arms distinct, object form mandated, rule 17 consistent.';
END $$;

COMMIT;
