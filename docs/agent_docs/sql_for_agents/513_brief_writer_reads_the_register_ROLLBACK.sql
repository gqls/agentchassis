-- ROLLBACK for 513 — the brief writer stops reading the register.
-- Briefs already written keep whatever positioning they absorbed; a brief is not
-- made wrong by removing its input.
BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        default_config #- '{workflow,steps,read_register}',
        '{workflow,steps,read_specs,next_step}', '"search_web"'::jsonb, true),
      '{workflow,steps,write_brief,config,input_fields}',
      '["input_data","site_specs","search_results","scrape_results","prepared_urls"]'::jsonb, true),
    updated_at = now()
WHERE type = 'brief-writer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DELETE FROM schema_migrations WHERE filename = '513_brief_writer_reads_the_register.sql';

DO $$
DECLARE n int; nxt text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions,
       LATERAL jsonb_object_keys(default_config->'workflow'->'steps') k
   WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 8 THEN RAISE EXCEPTION 'rollback: expected 8 steps, found %', n; END IF;
  SELECT default_config #>> '{workflow,steps,read_specs,next_step}' INTO nxt
    FROM agent_definitions WHERE type='brief-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF nxt <> 'search_web' THEN RAISE EXCEPTION 'rollback: read_specs chains to %', nxt; END IF;
  RAISE NOTICE '513 rollback OK — note the prompt section is left in place; it renders empty without the input_field and is harmless';
END $$;

COMMIT;
