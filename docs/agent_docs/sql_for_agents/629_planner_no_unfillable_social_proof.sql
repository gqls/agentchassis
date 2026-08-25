-- 629_planner_no_unfillable_social_proof.sql
--
-- OWNER RULING 2026-08-25 (ruling 6): "Stop planning them on sites that can't fill them."
-- Two edits to build-site-planner's plan_site template:
--   (1) the worked EXAMPLE output no longer demonstrates a testimonials section - the example is
--       the instruction ("the example is the instruction; the rule is commentary", proven twice
--       in this estate), and the example planted the slot fleet-wide;
--   (2) an explicit rule beside "Plan ONLY pages that are directly justified...".
-- The writer's rules 16-17 (fill empty testimonial slots with values/approach statements) become
-- dead letters as planned slots stop appearing; their removal follows separately.
-- ROLLBACK: 629_..._ROLLBACK.sql (restores from migration_backups).

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '629_planner_no_unfillable_social_proof', 'agent_definitions', ad.id::text,
       jsonb_build_object('default_config', ad.default_config), 'pre-629 build-site-planner default_config'
FROM agent_definitions ad
WHERE ad.type='build-site-planner' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

DO $mig$
DECLARE
  tpl text;
  ex_old text := '{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"]}, {"name": "testimonials", "facts": []}, {"name": "call-to-action", "facts": []}';
  ex_new text := '{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"]}, {"name": "call-to-action", "facts": []}';
  anchor text := 'If you don''t have evidence for a page, leave it out — fewer well-justified pages are better than padding the count.';
  new_rule text := '

Plan a testimonials, reviews, or case-study section ONLY when the inputs above contain real, named testimonial or case material for it. A site with none gets no such section: an empty slot filled with the company describing its own values serves the reader worse than the section being absent. (Owner ruling 2026-08-25.)';
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '629: template path empty'; END IF;
  IF (length(tpl)-length(replace(tpl,ex_old,'')))/length(ex_old) <> 1 THEN RAISE EXCEPTION '629: example line not found exactly once - template drifted'; END IF;
  IF (length(tpl)-length(replace(tpl,anchor,'')))/length(anchor) <> 1 THEN RAISE EXCEPTION '629: anchor not found exactly once - template drifted'; END IF;
  tpl := replace(tpl, ex_old, ex_new);
  tpl := replace(tpl, anchor, anchor || new_rule);
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config, '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(tpl))
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  RAISE NOTICE '629 OK: example no longer demonstrates a testimonials slot; explicit rule added. Live on the next plan_site call.';
END $mig$;

COMMIT;
