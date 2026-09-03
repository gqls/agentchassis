-- 734_classifier_reads_the_positioning_register_ROLLBACK.sql
--
-- Reverses 734: removes the register step, unwires it, and restores the ORIGINAL
-- four-field input allow-list.
--
-- ⚠ NOTE ON SCOPE. 734 fixed a PRE-EXISTING defect as well as adding the register:
-- `layout_taxonomy` had never been in `input_fields`, so the classifier was shown a null
-- library-tag list. This rollback restores that defect too, because a rollback that leaves
-- half of a migration in place is not a rollback. If you want to drop only the REGISTER and
-- keep the taxonomy fix, use the variant at the bottom of this file instead.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '734R REFUSED: expected exactly 1 active row, found %', n; END IF;
  PERFORM snapshot_agent('domain-research-classifier',
                         '734_..._ROLLBACK.sql: pre-rollback');
END $$;

DO $do$
DECLARE cfg jsonb; tpl text; newtpl text; n int;
  block text := $b734${{.positioning_register.block}}
{{if .site_specs.specs.site_archetype}}## Adoption Reference — STRONGEST signal$b734$;
  anchor text := $a734${{if .site_specs.specs.site_archetype}}## Adoption Reference — STRONGEST signal$a734$;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  tpl := cfg #>> '{workflow,steps,classify_and_extract,config,prompt_template}';
  n := (length(tpl) - length(replace(tpl, block, ''))) / length(block);
  IF n <> 1 THEN
    RAISE EXCEPTION '734R: the inserted block appears % times, expected 1 — another lane has edited this prompt; resolve by hand rather than guessing', n;
  END IF;
  newtpl := replace(tpl, block, anchor);

  cfg := jsonb_set(cfg, '{workflow,steps,classify_and_extract,config,prompt_template}', to_jsonb(newtpl), false);
  cfg := jsonb_set(cfg, '{workflow,steps,classify_and_extract,config,input_fields}',
                   '["input_data","search_results","scraped_data","site_specs"]'::jsonb, false);
  cfg := jsonb_set(cfg, '{workflow,steps,read_layout_taxonomy,next_step}', '"classify_and_extract"'::jsonb, false);
  cfg := cfg #- '{workflow,steps,read_positioning_register}';

  UPDATE agent_definitions SET default_config = cfg, updated_at = now()
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '734R: updated % rows, expected 1', n; END IF;
END $do$;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg #> '{workflow,steps,read_positioning_register}' IS NOT NULL THEN
    RAISE EXCEPTION '734R VERIFY: the step still exists';
  END IF;
  IF cfg #>> '{workflow,steps,read_layout_taxonomy,next_step}' IS DISTINCT FROM 'classify_and_extract' THEN
    RAISE EXCEPTION '734R VERIFY: the chain was not restored';
  END IF;
  IF position('{{.positioning_register.block}}' in (cfg #>> '{workflow,steps,classify_and_extract,config,prompt_template}')) > 0 THEN
    RAISE EXCEPTION '734R VERIFY: the prompt still references the register block';
  END IF;
END $$;

COMMIT;

-- VARIANT — drop the register, KEEP the taxonomy fix (recommended if 734(B) misbehaves but
-- the null library list was genuinely a defect). Run the transaction above with this line
-- substituted for the input_fields assignment:
--   '["input_data","search_results","scraped_data","site_specs","layout_taxonomy"]'::jsonb
