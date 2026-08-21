-- ROLLBACK for 528 — restore nav-updater's `complete` step to `output_fields`
-- list mode, removing `result_mapping` and the `commit_sha` exposure.
--
-- WHEN YOU WOULD RUN THIS: if the `result_mapping` conversion changes
-- observable behaviour for an EXISTING reader (it should not: every prior
-- field is identity-mapped). Restores the EXACT prior list; refuses if
-- result_mapping is not the exact one 528 wrote.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'nav-updater' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '528 ROLLBACK: no live nav-updater complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '528 ROLLBACK: complete carries no result_mapping — 528 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'site_record', 'site_record',
        'nav_refreshed', 'nav_refreshed',
        'site_components_rendered', 'site_components_rendered',
        'rerender_pages', 'rerender_pages',
        'items_result', 'items_result',
        'commit_sha', 'js_snippets_deployed.response.data.commit_sha'
    ) THEN
        RAISE EXCEPTION '528 ROLLBACK: result_mapping is %, not exactly what 528 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["site_record","nav_refreshed","site_components_rendered","rerender_pages","items_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'nav-updater'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'nav-updater' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '528 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["site_record","nav_refreshed","site_components_rendered","rerender_pages","items_result"]'::jsonb THEN
        RAISE EXCEPTION '528 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '528 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
