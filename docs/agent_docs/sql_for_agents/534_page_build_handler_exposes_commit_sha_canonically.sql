-- 534 — page-build-handler's `complete` step exposes `commit_sha` at a
--       CANONICAL top-level key, converting `output_fields` to `result_mapping`.
--       RFC_029 §10.13 step 5's real gate (bdl/`commit_sha`) — the 6th and
--       LAST handler config in this lane's real scope. CONFIG ONLY.
--       DEPENDS ON migration 519 (page-rerender) — already APPLIED and
--       verified live; this file's own guard re-confirms the dependency.
--
-- ============================================================================
-- WHY page-build-handler IS DIFFERENT FROM THE OTHER FIVE IN THIS BATCH
-- ============================================================================
-- Every other handler (519/521/522/523/527/528) has its OWN direct
-- `git_commit` step. page-build-handler does NOT — it reaches one INDIRECTLY,
-- one hop away, by calling `page-rerender` via `deploy_page` (`call_agent`,
-- `target_role: page_renderer`, `output_field: deploy_result`,
-- `next_step: complete` directly). So page-build-handler's own `commit_sha`
-- can only exist once page-rerender's OWN response carries one — which
-- migration 519 shipped and this file's own guard reconfirms is live.
--
-- [MEASURED 2026-08-21, PRE-519 sample, orchestration cc6167b9-...]
-- page-build-handler's `deploy_result.response` (i.e. what call_agent stored
-- from calling page-rerender) was, before 519:
--     {"deploy_result": {…raw git_commit result…}, "rendered_page": {…}}
-- — this is page-rerender's OWN pre-519 `output_fields` response, one
-- `.response` hop down from page-build-handler's `deploy_result` key. Since
-- 519 converted page-rerender's OWN `complete` step to `result_mapping`
-- (adding `commit_sha` at the SAME top level as `deploy_result`/
-- `rendered_page`, changing neither), the POST-519 shape is
--     {"deploy_result": {…}, "rendered_page": {…}, "commit_sha": "<sha>"}
-- so page-build-handler's own canonical path is
-- `deploy_result.response.commit_sha` — ONE level down from ITS OWN
-- `deploy_result` output_field, then straight to `commit_sha`. NOT
-- `deploy_result.response.deploy_result.response.data.commit_sha` (the
-- pre-519 indirect path) and NOT `deploy_result.response.data.commit_sha`
-- (page-rerender's OWN internal path, one layer too shallow from here).
--
-- No live post-519 sample existed at write time (page-build-handler's last
-- run reaching deploy_page predates 519's apply by about an hour, per the
-- demand-control read below) — this migration's own VERIFY block checks
-- against whatever the freshest sample is at apply time, and the guard below
-- refuses if migration 519 is not live, so this file cannot silently drift
-- ahead of its own dependency.
--
-- ============================================================================
-- THE OTHER PATHS TO `complete` — absence is correct, unchanged
-- ============================================================================
-- Not every run reaches `deploy_page`: `mark_no_ready_sections` /
-- `mark_writer_skipped` (no sections to write) also route to `complete`
-- without ever calling page-rerender — [MEASURED 2026-08-21] the two most
-- recent page-build-handler completions took exactly this path, with NO
-- `deploy_result` key in `collected_data` at all. `commit_sha` correctly
-- resolves ABSENT there, same as `deploy_result` itself does today.
-- `complete_error` (a separate complete_workflow step, output_fields =
-- ["page_content","site_record"]) never exposed `deploy_result` and is
-- UNTOUCHED by this file.
--
-- Same mechanism as 519 (output_fields cannot rename a nested path; the fix
-- is a result_mapping CONVERSION, not an addition) — see that file for the
-- full mechanism explanation, not repeated here.
--
-- ROLLBACK: 534_page_build_handler_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('page-build-handler', '534_page_build_handler_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '534: expected exactly 1 live page-build-handler row, found %', n;
    END IF;

    -- THE DEPENDENCY: refuse unless page-rerender's OWN commit_sha mapping
    -- (migration 519) is live. Without it this file would map to a path that
    -- resolves to nothing forever, silently, on every run.
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'page-rerender' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
           AND default_config #> '{workflow,steps,complete,config,result_mapping}' ? 'commit_sha'
    ) THEN
        RAISE EXCEPTION '534: page-rerender does not yet expose commit_sha via result_mapping (migration 519) — this file depends on it; apply 519 first';
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '534: page-build-handler has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '534: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '534: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["sections_saved","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '534: complete.output_fields is %, want exactly ["sections_saved","deploy_result"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','action']
          FROM agent_definitions WHERE type='page-build-handler' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'call_agent' THEN
        RAISE EXCEPTION '534: deploy_page no longer runs call_agent — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','output_field']
          FROM agent_definitions WHERE type='page-build-handler' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'deploy_result' THEN
        RAISE EXCEPTION '534: deploy_page''s output_field is no longer deploy_result — re-measure the mapped path before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','config','target_role']
          FROM agent_definitions WHERE type='page-build-handler' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'page_renderer' THEN
        RAISE EXCEPTION '534: deploy_page no longer targets page_renderer (page-rerender) — the mapped path assumes THAT agent''s response shape specifically; re-derive if the target changed';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'sections_saved', 'sections_saved',
               'deploy_result',  'deploy_result',
               'commit_sha',     'deploy_result.response.commit_sha'
           ),
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

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '534 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'sections_saved' IS DISTINCT FROM 'sections_saved' THEN
        RAISE EXCEPTION '534 VERIFY: sections_saved identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'deploy_result' IS DISTINCT FROM 'deploy_result' THEN
        RAISE EXCEPTION '534 VERIFY: deploy_result identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'deploy_result.response.commit_sha' THEN
        RAISE EXCEPTION '534 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 3 THEN
        RAISE EXCEPTION '534 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    -- SANITY CHECK against REAL data: confirm at least one historical run's
    -- deploy_result.response actually carries a commit_sha key NOW (post-519,
    -- the shape this migration depends on) — not merely that 519's config is
    -- live, but that a real call_agent envelope has actually reflected it.
    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'page-build-handler'
           AND collected_data #> '{deploy_result,response}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE WARNING '534 VERIFY: no page-build-handler orchestration yet shows deploy_result.response.commit_sha — expected if none has called page-rerender since 519 applied; the mapping is correct by construction (519 verified independently) but UNCONFIRMED end-to-end until a fresh completion lands. Re-check after one.';
    END IF;

    RAISE NOTICE '534 OK: page-build-handler.complete exposes sections_saved, deploy_result (unchanged) and commit_sha (new) via result_mapping';
END $$;

COMMIT;
