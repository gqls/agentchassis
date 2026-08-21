-- 521 — rerender-pages's `complete` step exposes `commit_sha` at a CANONICAL
--       top-level key, converting its result contract from `output_fields`
--       (list) to `result_mapping` (explicit target<-source pairs).
--       RFC_029 §10.13 step 5's real gate (bdl/`commit_sha`) — handler half,
--       following migration 519 (page-rerender, the first and largest).
--       CONFIG ONLY — live on apply.
--
-- Same mechanism as 519 (read that file for the full explanation of why an
-- ADDED key cannot work and a MODE CONVERSION is required): `output_fields`
-- (`ResultModeFields`) uses each list entry as BOTH the extraction path and
-- the response key, so a dotted path added to the list produces a response
-- key that is the literal dotted STRING, invisible to any dot-path reader.
-- `result_mapping` (`ResultModeMapping`) is the only mode that builds an
-- explicit, independently-named target<-source map, and it is mutually
-- exclusive with `output_fields` (`ResolveResultSpec`'s precedence table
-- matches exactly one key) — so this is a CONVERSION, not an addition.
--
-- WHY rerender-pages: real bdl handler (89/7d), single direct git_commit step
-- (`deploy_js_snippets`), single path to `complete` (deploy_js_snippets ->
-- rebuild_blog_listing -> get_pages -> check_pages_exist -> ... ->
-- create_rerender_items -> mark_site_deployed -> complete) — no branching that
-- reaches `complete` without having gone through the commit, and no second
-- git-touching call in the chain, so no ambiguity about which commit is
-- canonical (contrast section-editor / webdesign-agent below, which each have
-- a second, later git-touching call and are resolved by the same principle:
-- map to the handler's OWN direct commit, not a downstream call's).
--
-- WHAT THIS CHANGES: NOTHING for existing readers. Every field currently in
-- `output_fields` is re-declared as an identity mapping (byte-identical
-- response), plus one new `commit_sha` entry sourced from `js_snippets_deployed.response.data.commit_sha`.
--
-- [MEASURED 2026-08-21] a live completed `rerender-pages` orchestration confirms the
-- path resolves: `collected_data.js_snippets_deployed.response.data.commit_sha` is
-- present (see this file's VERIFY block, which re-checks this against a real
-- row rather than trusting the config string alone).
--
-- ROLLBACK: 521_rerender_pages_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('rerender-pages', '521_rerender_pages_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'rerender-pages' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '521: expected exactly 1 live rerender-pages row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'rerender-pages' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '521: rerender-pages has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '521: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '521: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["rerender_pages","items_result","site_components_result"]'::jsonb THEN
        RAISE EXCEPTION '521: complete.output_fields is %, want exactly ["rerender_pages","items_result","site_components_result"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_js_snippets','action']
          FROM agent_definitions WHERE type='rerender-pages' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_commit' THEN
        RAISE EXCEPTION '521: deploy_js_snippets no longer runs git_commit — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_js_snippets','output_field']
          FROM agent_definitions WHERE type='rerender-pages' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'js_snippets_deployed' THEN
        RAISE EXCEPTION '521: deploy_js_snippets''s output_field is no longer js_snippets_deployed — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'rerender_pages', 'rerender_pages',
               'items_result', 'items_result',
               'site_components_result', 'site_components_result',
               'commit_sha', 'js_snippets_deployed.response.data.commit_sha'
           ),
           true),
       updated_at = NOW()
 WHERE type = 'rerender-pages'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg    jsonb;
    leaked text;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'rerender-pages' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '521 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'rerender_pages' IS DISTINCT FROM 'rerender_pages' THEN RAISE EXCEPTION '521 VERIFY: rerender_pages identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'items_result' IS DISTINCT FROM 'items_result' THEN RAISE EXCEPTION '521 VERIFY: items_result identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'site_components_result' IS DISTINCT FROM 'site_components_result' THEN RAISE EXCEPTION '521 VERIFY: site_components_result identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'js_snippets_deployed.response.data.commit_sha' THEN
        RAISE EXCEPTION '521 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 4 THEN
        RAISE EXCEPTION '521 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'rerender-pages'
           AND collected_data #> '{js_snippets_deployed,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '521 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
    END IF;

    WITH RECURSIVE steps(type, path, step) AS (
        SELECT ad.type, s.key, s.value
          FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
         WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
        UNION ALL
        SELECT p.type, p.path || '.' || s.key, s.value
          FROM steps p, LATERAL jsonb_each(p.step->'config'->'sub_workflow'->'steps') s
    )
    SELECT string_agg(type || '.' || path, ', ') INTO leaked
      FROM steps
     WHERE step->'config'->'result_mapping'->>'commit_sha' = 'js_snippets_deployed.response.data.commit_sha'
       AND type <> 'rerender-pages';
    IF leaked IS NOT NULL THEN
        RAISE EXCEPTION '521 VERIFY: the mapping leaked to steps it was not meant for: %', leaked;
    END IF;

    RAISE NOTICE '521 OK: rerender-pages.complete exposes rerender_pages, items_result, site_components_result (unchanged) and commit_sha (new) via result_mapping; confirmed against a live orchestration';
END $$;

COMMIT;
