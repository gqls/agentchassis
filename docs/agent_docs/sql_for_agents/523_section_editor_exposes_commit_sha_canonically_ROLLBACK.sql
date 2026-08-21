-- ROLLBACK for 523 — restore section-editor's `complete` step to `output_fields`
-- list mode, removing `result_mapping` and the `commit_sha` exposure.
--
-- WHEN YOU WOULD RUN THIS: if the `result_mapping` conversion changes
-- observable behaviour for an EXISTING reader (it should not: every prior
-- field is identity-mapped). Restores the EXACT prior list; refuses if
-- result_mapping is not the exact one 523 wrote.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'section-editor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '523 ROLLBACK: no live section-editor complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '523 ROLLBACK: complete carries no result_mapping — 523 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'edit_result', 'edit_result',
        'git_result', 'git_result',
        'deploy_result', 'deploy_result',
        'commit_sha', 'git_result.response.data.commit_sha'
    ) THEN
        RAISE EXCEPTION '523 ROLLBACK: result_mapping is %, not exactly what 523 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["edit_result","git_result","deploy_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'section-editor'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'section-editor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '523 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["edit_result","git_result","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '523 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '523 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
