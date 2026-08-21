-- 535 — asset-deployer's `complete` step exposes `commit_sha` at a CANONICAL
--       top-level key, converting `output_fields` to `result_mapping`.
--       RFC_029 §10.13 step 5's real gate (bdl/`commit_sha`). CONFIG ONLY.
--
-- ============================================================================
-- A REAL GAP IN THIS LANE'S OWN SCOPING, FOUND BY THE staged-component-build
-- LANE'S EMPIRICAL CROSS-CHECK, NOT BY THE STRUCTURAL CENSUS THIS BATCH USED
-- ============================================================================
-- Migrations 519/521/522/523/527/528/534 scoped their population by searching
-- for a step whose ACTION is literally `git_commit`. asset-deployer has no
-- such step — but its `deploy_asset` step (action `deploy_image_asset`) DOES
-- commit to git, as its own live config description says verbatim: "Download
-- from S3, optimize by purpose, commit to git". A search keyed on the action
-- NAME `git_commit` structurally cannot find this — it is a different action
-- that happens to also touch git internally.
--
-- [MEASURED 2026-08-21] confirmed at asset-deployer's OWN collected_data (not
-- inferred, not borrowed from a caller): a real completed orchestration shows
-- `deploy_result.response.data.commit_sha` populated with a genuine sha,
-- alongside `deployed: true`, `image_url`, `size_bytes` — the shape of
-- `deploy_image_asset`'s own action result, one `.response` hop down from
-- asset-deployer's own `deploy_result` output_field. Structurally identical to
-- the git-adapter reply shape every other migration in this batch maps
-- (`.response.data.commit_sha`), because it is the same underlying adapter.
--
-- asset-deployer IS a real, live bdl handler: 269 items / 156 complete in the
-- period sampled by this lane's earlier census (site_work_items.handler_agent).
--
-- ============================================================================
-- THE OTHER TWO AGENTS THE SAME CROSS-CHECK FLAGGED, AND WHY ONLY ONE OF THEM
-- IS BUILT HERE (image-build-handler is 536; tool-generator is NOT built —
-- see the note this migration's sibling commit records)
-- ============================================================================
-- image-build-handler calls THIS agent (asset-deployer) via `call_agent` and
-- is the dependent case, same shape as 534/page-build-handler → 519/
-- page-rerender — built as migration 536, guarded on THIS file being live
-- first.
--
-- Same mechanism as 519 (output_fields cannot rename a nested path to a
-- top-level key; the fix is a result_mapping CONVERSION) — not repeated here.
--
-- ROLLBACK: 535_asset_deployer_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('asset-deployer', '535_asset_deployer_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '535: expected exactly 1 live asset-deployer row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '535: asset-deployer has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '535: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '535: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '535: complete.output_fields is %, want exactly ["deploy_result"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_asset','action']
          FROM agent_definitions WHERE type='asset-deployer' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'deploy_image_asset' THEN
        RAISE EXCEPTION '535: deploy_asset no longer runs deploy_image_asset — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_asset','output_field']
          FROM agent_definitions WHERE type='asset-deployer' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'deploy_result' THEN
        RAISE EXCEPTION '535: deploy_asset''s output_field is no longer deploy_result — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'deploy_result', 'deploy_result',
               'commit_sha',    'deploy_result.response.data.commit_sha'
           ),
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

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '535 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'deploy_result' IS DISTINCT FROM 'deploy_result' THEN
        RAISE EXCEPTION '535 VERIFY: deploy_result identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'deploy_result.response.data.commit_sha' THEN
        RAISE EXCEPTION '535 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 2 THEN
        RAISE EXCEPTION '535 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'asset-deployer'
           AND collected_data #> '{deploy_result,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '535 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
    END IF;

    -- Not a cross-fleet string-match negative control (see 528's fix, same
    -- session): the real protection is structural (exactly-1-row guard above
    -- plus this UPDATE's own WHERE type='asset-deployer' clause).

    RAISE NOTICE '535 OK: asset-deployer.complete exposes deploy_result (unchanged) and commit_sha (new) via result_mapping; confirmed against a live orchestration';
END $$;

COMMIT;
