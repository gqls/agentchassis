-- ROLLBACK for 536 — restore image-build-handler's `complete` step to
-- `output_fields` list mode, removing `result_mapping` and `commit_sha`.
--
-- WHEN YOU WOULD RUN THIS: if the conversion changes observable behaviour for
-- an existing reader (should not — image_result/asset_stored/deploy_result
-- are identity-mapped).

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '536 ROLLBACK: no live image-build-handler complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '536 ROLLBACK: complete carries no result_mapping — 536 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'image_result',  'image_result',
        'asset_stored',  'asset_stored',
        'deploy_result', 'deploy_result',
        'commit_sha',    'deploy_result.response.commit_sha'
    ) THEN
        RAISE EXCEPTION '536 ROLLBACK: result_mapping is %, not exactly what 536 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["image_result","asset_stored","deploy_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'image-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '536 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["image_result","asset_stored","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '536 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '536 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
