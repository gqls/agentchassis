-- ROLLBACK for 534 — restore page-build-handler's `complete` step to
-- `output_fields` list mode, removing `result_mapping` and the `commit_sha`
-- exposure.
--
-- WHEN YOU WOULD RUN THIS: if the conversion changes observable behaviour for
-- an existing reader (should not — sections_saved and deploy_result are
-- identity-mapped, byte-for-byte the same). Also relevant if migration 519
-- (page-rerender) is ever rolled back: this file's mapped path
-- (deploy_result.response.commit_sha) depends on 519's shape and would
-- silently stop resolving — not a crash, just an absent field, but worth
-- rolling this one back too for cleanliness if 519 goes.
--
-- Restores the EXACT prior list; refuses if result_mapping is not the exact
-- one 534 wrote.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '534 ROLLBACK: no live page-build-handler complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '534 ROLLBACK: complete carries no result_mapping — 534 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'sections_saved', 'sections_saved',
        'deploy_result',  'deploy_result',
        'commit_sha',     'deploy_result.response.commit_sha'
    ) THEN
        RAISE EXCEPTION '534 ROLLBACK: result_mapping is %, not exactly what 534 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["sections_saved","deploy_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '534 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["sections_saved","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '534 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '534 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
