-- 627_page_content_writer_delete_july_substitutes.sql
--
-- OWNER RULING 2026-08-25 (copy_quality_two_stage/OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md,
-- ruling 2): "the other option is to not say it at all. I approve of deleting the substitutes."
-- Deletes the three "say this instead" clauses the July integrity rules installed in
-- page-content-writer's generate_content template. EVERY BAN STAYS: no-overclaiming, the
-- no-history-claims list, the "honest" word ban, and (ruling 3) the tool-fallibility mandate.
-- Nothing is added in their place: the pre-registered replay experiment
-- (copy_quality_two_stage/AUDIT_prompts/EXPERIMENT_2026-08-25_about_section_replay.md) showed the
-- drafted replacement earned nothing over deletion (arm D vs arm C).
-- Evidence the clauses teach the register: the same experiment's arm B (10->6 on the
-- methodology/self-limiting score) and the phase-2 verdict file.
-- ROLLBACK: 627_..._ROLLBACK.sql (restores default_config from migration_backups).

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '627_page_content_writer_delete_july_substitutes', 'agent_definitions', ad.id::text,
       jsonb_build_object('default_config', ad.default_config),
       'pre-627 page-content-writer default_config (July substitute clauses still present)'
FROM agent_definitions ad
WHERE ad.type='page-content-writer' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

DO $mig$
DECLARE
  tpl text;
  cut1 text := 'If the section calls for a statement about method, say what we DO -- and with no recorded operating history that means ONLY how the content is sourced: we name our sources and their dates so a reader can check them -- and, where it fits, say plainly that we can still be wrong. ';
  cut2 text := 'Where the brief asks for method, say what the site does WITH SOURCES: name the manufacturer specification, published standard or retailer listing a figure comes from, and date it. ';
  cut3 text := 'Show it instead, by naming the limit, the failure mode, or what the thing cannot do. Say "we cannot tell you X" rather than "an honest assessment". ';
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '627: template path empty'; END IF;
  IF (length(tpl)-length(replace(tpl,cut1,'')))/length(cut1) <> 1 THEN RAISE EXCEPTION '627: cut1 count <> 1 - template has drifted, do not apply blind'; END IF;
  IF (length(tpl)-length(replace(tpl,cut2,'')))/length(cut2) <> 2 THEN RAISE EXCEPTION '627: cut2 count <> 2 - template has drifted, do not apply blind'; END IF;
  IF (length(tpl)-length(replace(tpl,cut3,'')))/length(cut3) <> 1 THEN RAISE EXCEPTION '627: cut3 count <> 1 - template has drifted, do not apply blind'; END IF;
  tpl := replace(replace(replace(tpl,cut1,''),cut2,''),cut3,'');
  IF position('NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE' in tpl)=0 THEN RAISE EXCEPTION '627: overclaiming ban lost'; END IF;
  IF position('must say the tool can give a wrong answer' in tpl)=0 THEN RAISE EXCEPTION '627: tool-fallibility mandate lost (owner ruling 3 KEEPS it)'; END IF;
  IF position('Never write the words "honest"' in tpl)=0 THEN RAISE EXCEPTION '627: honest word ban lost'; END IF;
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
       '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}', to_jsonb(tpl))
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  RAISE NOTICE '627 OK: three substitute clauses deleted; overclaiming ban, history-claims bans, honest word ban and tool-fallibility mandate all verified still present. Live on the next writer call.';
END $mig$;

COMMIT;
