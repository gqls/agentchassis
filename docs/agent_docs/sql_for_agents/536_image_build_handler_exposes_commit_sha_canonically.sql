-- 536 — image-build-handler's `complete` step exposes `commit_sha` at a
--       CANONICAL top-level key, converting `output_fields` to
--       `result_mapping`. RFC_029 §10.13 step 5's real gate (bdl/`commit_sha`).
--       CONFIG ONLY. DEPENDS ON migration 535 (asset-deployer) — apply that
--       first; this file's guard refuses otherwise.
--
-- ============================================================================
-- DEPENDENT CASE, SAME SHAPE AS 534 (page-build-handler) → 519 (page-rerender)
-- ============================================================================
-- image-build-handler has no git-touching action of its own — it reaches one
-- by calling `asset-deployer` via `call_asset_deployer` (`call_agent`,
-- `target_role: asset_deployer`, `output_field: deploy_result`,
-- `next_step: mark_work_item_complete`). Live config description: "Call
-- asset-deployer to download from S3, optimize by purpose, commit to git."
--
-- [MEASURED 2026-08-21, a real completed image-build-handler orchestration]
-- PRE-535, `deploy_result.response` = `{deploy_result: {...raw
-- deploy_image_asset result...}, agent_type: "asset-deployer", ...call_agent
-- envelope...}` — the SAME one-response-hop-then-nested-again shape as
-- page-build-handler's pre-519 sample. Once 535 converts asset-deployer's OWN
-- `complete` step to `result_mapping` (adding `commit_sha` as a SIBLING of
-- `deploy_result`, not nested inside it), the post-535 shape gains that one
-- key at the SAME level, so image-build-handler's canonical path is
-- `deploy_result.response.commit_sha` — one level shallower than
-- asset-deployer's OWN internal path (`deploy_result.response.data.commit_sha`),
-- exactly the same relationship 534 has to 519.
--
-- image-build-handler IS a real, live bdl handler (183 items / 125 in 7 days
-- in this lane's earlier census).
--
-- ============================================================================
-- THE OTHER PATH TO `complete` — this migration does not touch it
-- ============================================================================
-- `complete`'s current output_fields = ["image_result","asset_stored",
-- "deploy_result"] — three fields from three different upstream generation/
-- storage steps, only one of which (`deploy_result`, the asset-deployer call)
-- is git-touching. `image_result` and `asset_stored` are identity-mapped
-- unchanged; this file adds only `commit_sha`.
--
-- ROLLBACK: 536_image_build_handler_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('image-build-handler', '536_image_build_handler_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '536: expected exactly 1 live image-build-handler row, found %', n;
    END IF;

    -- THE DEPENDENCY: refuse unless asset-deployer's OWN commit_sha mapping
    -- (migration 535) is live.
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'asset-deployer' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
           AND default_config #> '{workflow,steps,complete,config,result_mapping}' ? 'commit_sha'
    ) THEN
        RAISE EXCEPTION '536: asset-deployer does not yet expose commit_sha via result_mapping (migration 535) — this file depends on it; apply 535 first';
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '536: image-build-handler has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '536: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '536: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["image_result","asset_stored","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '536: complete.output_fields is %, want exactly the three-entry list — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','call_asset_deployer','action']
          FROM agent_definitions WHERE type='image-build-handler' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'call_agent' THEN
        RAISE EXCEPTION '536: call_asset_deployer no longer runs call_agent — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','call_asset_deployer','output_field']
          FROM agent_definitions WHERE type='image-build-handler' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'deploy_result' THEN
        RAISE EXCEPTION '536: call_asset_deployer''s output_field is no longer deploy_result — re-measure the mapped path before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','call_asset_deployer','config','target_role']
          FROM agent_definitions WHERE type='image-build-handler' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'asset_deployer' THEN
        RAISE EXCEPTION '536: call_asset_deployer no longer targets asset_deployer (asset-deployer) — the mapped path assumes THAT agent''s response shape specifically; re-derive if the target changed';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'image_result',  'image_result',
               'asset_stored',  'asset_stored',
               'deploy_result', 'deploy_result',
               'commit_sha',    'deploy_result.response.commit_sha'
           ),
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

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '536 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'image_result' IS DISTINCT FROM 'image_result' THEN
        RAISE EXCEPTION '536 VERIFY: image_result identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'asset_stored' IS DISTINCT FROM 'asset_stored' THEN
        RAISE EXCEPTION '536 VERIFY: asset_stored identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'deploy_result' IS DISTINCT FROM 'deploy_result' THEN
        RAISE EXCEPTION '536 VERIFY: deploy_result identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'deploy_result.response.commit_sha' THEN
        RAISE EXCEPTION '536 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 4 THEN
        RAISE EXCEPTION '536 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'image-build-handler'
           AND collected_data #> '{deploy_result,response}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE WARNING '536 VERIFY: no image-build-handler orchestration yet shows deploy_result.response.commit_sha (post-535 shape) — expected if none has called asset-deployer since 535 applied; the mapping is correct by construction (535 verified independently) but UNCONFIRMED end-to-end until a fresh completion lands. Re-check after one.';
    END IF;

    RAISE NOTICE '536 OK: image-build-handler.complete exposes image_result, asset_stored, deploy_result (unchanged) and commit_sha (new) via result_mapping';
END $$;

COMMIT;
