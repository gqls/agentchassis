-- 348_pageflow_swo_deploy_steps_resolve_by_identity_ROLLBACK.sql
-- Sidecar (UPPERCASE suffix — excluded from run-migrations.sh runs).
-- Restores the exact pre-348 shape of the four deploy steps: static purpose +
-- uri_field, no input_fields, no s3_uri/asset_id/domain keys.
--
-- NOTE deliberately restored AS-WAS: the pre-348 shape carries bugs_open/231
-- (static purpose shadowed by the spec default — logo deploys as hero). Rolling
-- back reinstates that defect; do it only for a stated reason, and say so in
-- schema_migrations.notes via --record-only.

BEGIN;

DO $$
DECLARE
  n int;
BEGIN
  -- Only roll back the shape 348 installed; refuse anything else.
  SELECT count(*) INTO n
  FROM agent_definitions ad,
       LATERAL (VALUES ('deploy_hero_image','hero'), ('deploy_logo_image','logo')) AS want(step_name, p),
       LATERAL (SELECT ad.default_config#>ARRAY['workflow','steps',want.step_name,'config'] AS cfg) c
  WHERE ad.type IN ('pageflow-builder','site-work-orchestrator')
    AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND c.cfg->>'purpose' = want.p || '_stored.purpose'
    AND c.cfg->'input_fields' = '["purpose","domain","asset_id"]'::jsonb;
  IF n <> 4 THEN
    RAISE EXCEPTION '348_ROLLBACK: rows are not in the 348 shape (% of 4) — nothing to roll back, or drifted since', n;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,deploy_hero_image,config}',
      ((((default_config#>'{workflow,steps,deploy_hero_image,config}')
        - 's3_uri') - 'asset_id') - 'domain') - 'input_fields'
        || '{"purpose":"hero","uri_field":"hero_result.image_uri"}'::jsonb
    ),
    updated_at = NOW()
WHERE type IN ('pageflow-builder','site-work-orchestrator')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,deploy_logo_image,config}',
      ((((default_config#>'{workflow,steps,deploy_logo_image,config}')
        - 's3_uri') - 'asset_id') - 'domain') - 'input_fields'
        || '{"purpose":"logo","uri_field":"logo_result.image_uri"}'::jsonb
    ),
    updated_at = NOW()
WHERE type IN ('pageflow-builder','site-work-orchestrator')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
  n_ok int;
BEGIN
  SELECT count(*) INTO n_ok
  FROM agent_definitions ad,
       LATERAL (VALUES
         ('deploy_hero_image','hero','hero_result.image_uri'),
         ('deploy_logo_image','logo','logo_result.image_uri')
       ) AS want(step_name, want_purpose, want_uri_field),
       LATERAL (SELECT ad.default_config#>ARRAY['workflow','steps',want.step_name,'config'] AS cfg) c
  WHERE ad.type IN ('pageflow-builder','site-work-orchestrator')
    AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND c.cfg->>'purpose'   = want.want_purpose
    AND c.cfg->>'uri_field' = want.want_uri_field
    AND c.cfg->'input_fields' IS NULL
    AND c.cfg->'s3_uri' IS NULL AND c.cfg->'asset_id' IS NULL AND c.cfg->'domain' IS NULL;
  IF n_ok <> 4 THEN
    RAISE EXCEPTION '348_ROLLBACK POST-VERIFY FAILED: % of 4 steps restored', n_ok;
  END IF;
END $$;

COMMIT;
