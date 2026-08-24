-- ROLLBACK for 598_build_site_planner_distinct_no_facts_arms.sql
-- Anchored reverse edits; aborts unless each inserted text is found exactly once.
-- Restores the bugs_open/380 defect (identical fail-open arms) — recovery only.

BEGIN;

SELECT snapshot_agent('build-site-planner', '598_ROLLBACK: pre-revert');

DO $do$
DECLARE
  tpl text; n int;
  anchor1 text := '{{end}}{{end}}{{else}}No verified facts are registered for this site — use plain string section entries and no facts keys.{{end}}{{else}}No verified facts are registered for this site — use plain string section entries and no facts keys.{{end}}';
  repl1 text := $R${{end}}{{end}}{{else}}This site's evidence register holds NO verified facts. Use the OBJECT form for EVERY section entry on EVERY page with "facts": [] (rule 17) — nothing may be stated as a verified fact, and no section may state business numbers or named-entity relationships.{{end}}{{else}}NO evidence register exists for this site: nothing about this business is verified and it has no recorded operating history. Use the OBJECT form for EVERY section entry on EVERY page with "facts": [] (rule 17). No page brief may describe testing, buying, measuring, sourcing samples or any other operating practice as something this business does — a methodology or "how we assess" page may be planned only as what the site IS and does for a reader, never as a description of practice.{{end}}$R$;
  anchor2 text := 'When no Verified Facts are listed, use plain strings only.';
  repl2 text := $S$When no Verified Facts are listed, still use the object form with "facts": [] on every section — a plain string there leaves the writer unconstrained, which is the failure this rule exists to prevent (bugs_open/380).$S$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  n := (length(tpl) - length(replace(tpl, repl1, ''))) / length(repl1);
  IF n <> 1 THEN RAISE EXCEPTION '598 ROLLBACK: new arms found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, repl2, ''))) / length(repl2);
  IF n <> 1 THEN RAISE EXCEPTION '598 ROLLBACK: rule-17 amendment found % times, expected 1', n; END IF;

  tpl := replace(replace(tpl, repl1, anchor1), repl2, anchor2);

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(tpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '598 ROLLBACK: updated % rows', n; END IF;
  RAISE NOTICE '598 ROLLBACK OK.';
END $do$;

COMMIT;
