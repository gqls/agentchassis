-- 586 ROLLBACK — hand-run only, never by the migration runner.
--
-- Restores call_variant_gen to its pre-586 shape: no kind, no site_id, and the
-- three dead `default_kind` keys back where they were. Note what that MEANS
-- before running it: per-page hero images go back to being generated with no
-- kind (so, on a pre-382 adapter, by SDXL) and with no site style guide. Roll
-- back only if 586 caused a failure you have actually observed — the missing
-- site_id is a REQUIRED mapping, so the failure shape to look for is
-- call_variant_gen erroring to mark_work_item_failed on a site whose
-- ensure_site_record did not populate site_record.site_id.

BEGIN;

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        jsonb_set(
          jsonb_set(
            (default_config
              #- '{workflow,steps,call_variant_gen,config,input_mapping,kind?}')
              #- '{workflow,steps,call_variant_gen,config,input_mapping,site_id}',
            '{workflow,steps,call_variant_gen,config,default_kind}', '"hero"'),
          '{workflow,steps,call_hero_gen,config,default_kind}', '"hero"'),
        '{workflow,steps,call_logo_gen,config,default_kind}', '"logo"'),
    updated_at = now()
WHERE type='image-build-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='image-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND NOT (default_config->'workflow'->'steps'->'call_variant_gen'->'config'->'input_mapping' ? 'kind?')
     AND NOT (default_config->'workflow'->'steps'->'call_variant_gen'->'config'->'input_mapping' ? 'site_id');
  IF n <> 1 THEN
    RAISE EXCEPTION '586 ROLLBACK: expected exactly 1 row restored, found %', n;
  END IF;
END $$;

COMMIT;
