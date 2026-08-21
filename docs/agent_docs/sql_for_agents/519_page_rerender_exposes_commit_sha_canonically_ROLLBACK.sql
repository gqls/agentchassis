-- ROLLBACK for 519 — restore page-rerender's `complete` step to `output_fields`
-- list mode, removing `result_mapping` and the `commit_sha` exposure.
--
-- WHEN YOU WOULD RUN THIS: if the `result_mapping` conversion turns out to
-- change observable behaviour for an EXISTING reader (it should not: both
-- rendered_page and deploy_result are identity-mapped, byte-for-byte the same
-- as the output_fields list produced). If something downstream reads
-- `handler_result.response` differently after this file (e.g. expects the
-- response to be a specific instance TYPE, not just the same keys/values),
-- that is the observable to check before assuming this rollback is the fix.
--
-- Restores the EXACT prior list; refuses if result_mapping is not the exact
-- one 519 wrote (another session may have extended it on purpose).

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '519 ROLLBACK: no live page-rerender complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '519 ROLLBACK: complete carries no result_mapping — 519 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'rendered_page', 'rendered_page',
        'deploy_result', 'deploy_result',
        'commit_sha',    'deploy_result.response.data.commit_sha'
    ) THEN
        RAISE EXCEPTION '519 ROLLBACK: result_mapping is %, not exactly what 519 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["rendered_page","deploy_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'page-rerender'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '519 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["rendered_page","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '519 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '519 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
