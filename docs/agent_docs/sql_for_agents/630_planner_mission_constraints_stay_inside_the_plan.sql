-- 630_planner_mission_constraints_stay_inside_the_plan.sql
--
-- OWNER RULING 2026-08-25/26 (ruling 1 follow-up, in chat: "yes please ship"). The replay
-- experiment (copy_quality_two_stage/AUDIT_prompts/EXPERIMENT_2026-08-25_about_section_replay.md)
-- showed the PLANNER's page premise is the primary source of the about-page methodology register:
-- homegarden's page was PLANNED as "About Home Garden - Editorial Approach and What We Will Not
-- Do", condensed from a prohibition-rich mission seed, despite the template's existing
-- methodology-page guard. ~6 of 23 fleet about pages carry self-limiting titles.
--
-- THE WORDING BELOW IS THE TESTED SECOND DRAFT. The first draft (register rule alone) held in only
-- 1 of 2 planner replays - the other produced "Practical UK Guidance, No Products to Sell", a
-- self-limiting subtitle, the exact named error. Adding the HARD FORMAT clause ("'About' plus the
-- site name and nothing more: no subtitle, no dash, no qualifier") held 3 of 3: every replay
-- titled the page "About | Home Garden" (results appended to the experiment doc). Register rules
-- bend; format rules hold.
-- ROLLBACK: 630_..._ROLLBACK.sql (restores from migration_backups).

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '630_planner_mission_constraints_stay_inside_the_plan', 'agent_definitions', ad.id::text,
       jsonb_build_object('default_config', ad.default_config), 'pre-630 build-site-planner default_config (post-629)'
FROM agent_definitions ad
WHERE ad.type='build-site-planner' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

DO $mig$
DECLARE
  tpl text;
  anchor text := 'serves the reader worse than the section being absent. (Owner ruling 2026-08-25.)';
  new_rule text := '

Mission constraints stay inside the plan. Where the mission or briefing lists prohibitions (nothing tested, no brands, no shop, no advice), obey them in what you plan, and pass them to the writer inside page briefs as constraints. A reader-facing page is titled and premised on what the reader gets from it. Title the about page ''About'' plus the site name and nothing more: no subtitle, no dash, no qualifier. Its premise is what the site offers the reader (''we''re hoping you can get a lot of useful tips from this site''), kept brief. The same holds for every page title: a title states the page''s subject, and a constraint never appears in one. Titling or premising a page around what the site will not do, does not sell, or cannot promise is a planning error, whatever the mission says. (Owner ruling 2026-08-25.)';
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '630: template path empty'; END IF;
  IF (length(tpl)-length(replace(tpl,anchor,'')))/length(anchor) <> 1 THEN RAISE EXCEPTION '630: anchor (629 rule tail) not found exactly once - template drifted, do not apply blind'; END IF;
  IF position('Mission constraints stay inside the plan' in tpl) > 0 THEN RAISE EXCEPTION '630: rule already present'; END IF;
  tpl := replace(tpl, anchor, anchor || new_rule);
  IF position('no subtitle, no dash, no qualifier' in tpl) = 0 THEN RAISE EXCEPTION '630: the TESTED format clause is missing - wrong rule draft'; END IF;
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config, '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(tpl))
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  RAISE NOTICE '630 OK: mission-constraints rule with the TESTED hard-format about-title clause added. Live on the next plan_site call.';
END $mig$;

COMMIT;
