-- ROLLBACK for 535 — restore asset-deployer's `complete` step to
-- `output_fields` list mode, removing `result_mapping` and `commit_sha`.
--
-- WHEN YOU WOULD RUN THIS: if the conversion changes observable behaviour for
-- an existing reader (should not — deploy_result is identity-mapped). If
-- migration 536 (image-build-handler) depends on this, roll that back first.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '535 ROLLBACK: no live asset-deployer complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '535 ROLLBACK: complete carries no result_mapping — 535 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'deploy_result', 'deploy_result',
        'commit_sha',    'deploy_result.response.data.commit_sha'
    ) THEN
        RAISE EXCEPTION '535 ROLLBACK: result_mapping is %, not exactly what 535 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["deploy_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'asset-deployer'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '535 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '535 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '535 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
