-- ROLLBACK for 522 — restore css-patch-agent's `complete` step to `output_fields`
-- list mode, removing `result_mapping` and the `commit_sha` exposure.
--
-- WHEN YOU WOULD RUN THIS: if the `result_mapping` conversion changes
-- observable behaviour for an EXISTING reader (it should not: every prior
-- field is identity-mapped). Restores the EXACT prior list; refuses if
-- result_mapping is not the exact one 522 wrote.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'css-patch-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '522 ROLLBACK: no live css-patch-agent complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '522 ROLLBACK: complete carries no result_mapping — 522 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'css_fix', 'css_fix',
        'css_deployed', 'css_deployed',
        'commit_sha', 'css_deployed.response.data.commit_sha'
    ) THEN
        RAISE EXCEPTION '522 ROLLBACK: result_mapping is %, not exactly what 522 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["css_fix","css_deployed"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'css-patch-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '522 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["css_fix","css_deployed"]'::jsonb THEN
        RAISE EXCEPTION '522 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '522 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
