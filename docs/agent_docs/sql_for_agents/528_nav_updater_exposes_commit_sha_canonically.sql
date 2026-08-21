-- 528 — nav-updater's `complete` step exposes `commit_sha` at a CANONICAL
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
-- WHY nav-updater: real bdl handler (42/7d), single direct git_commit step
-- (`deploy_js_snippets`), and it runs BEFORE `get_pages`/`check_has_pages`/
-- `create_rerender_items` on every path to `complete` (deploy_js_snippets's
-- own next_step is get_pages), so js_snippets_deployed is already in
-- collected_data regardless of which branch check_has_pages takes. No second
-- git-touching call in this workflow — same simple shape as rerender-pages.
--
-- WHAT THIS CHANGES: NOTHING for existing readers. Every field currently in
-- `output_fields` is re-declared as an identity mapping (byte-identical
-- response), plus one new `commit_sha` entry sourced from `js_snippets_deployed.response.data.commit_sha`.
--
-- [MEASURED 2026-08-21] a live completed `nav-updater` orchestration confirms the
-- path resolves: `collected_data.js_snippets_deployed.response.data.commit_sha` is
-- present (see this file's VERIFY block, which re-checks this against a real
-- row rather than trusting the config string alone).
--
-- ROLLBACK: 528_nav_updater_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('nav-updater', '528_nav_updater_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'nav-updater' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '528: expected exactly 1 live nav-updater row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'nav-updater' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '528: nav-updater has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '528: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '528: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["site_record","nav_refreshed","site_components_rendered","rerender_pages","items_result"]'::jsonb THEN
        RAISE EXCEPTION '528: complete.output_fields is %, want exactly ["site_record","nav_refreshed","site_components_rendered","rerender_pages","items_result"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_js_snippets','action']
          FROM agent_definitions WHERE type='nav-updater' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_commit' THEN
        RAISE EXCEPTION '528: deploy_js_snippets no longer runs git_commit — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_js_snippets','output_field']
          FROM agent_definitions WHERE type='nav-updater' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'js_snippets_deployed' THEN
        RAISE EXCEPTION '528: deploy_js_snippets''s output_field is no longer js_snippets_deployed — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'site_record', 'site_record',
               'nav_refreshed', 'nav_refreshed',
               'site_components_rendered', 'site_components_rendered',
               'rerender_pages', 'rerender_pages',
               'items_result', 'items_result',
               'commit_sha', 'js_snippets_deployed.response.data.commit_sha'
           ),
           true),
       updated_at = NOW()
 WHERE type = 'nav-updater'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg    jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'nav-updater' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '528 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'site_record' IS DISTINCT FROM 'site_record' THEN RAISE EXCEPTION '528 VERIFY: site_record identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'nav_refreshed' IS DISTINCT FROM 'nav_refreshed' THEN RAISE EXCEPTION '528 VERIFY: nav_refreshed identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'site_components_rendered' IS DISTINCT FROM 'site_components_rendered' THEN RAISE EXCEPTION '528 VERIFY: site_components_rendered identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'rerender_pages' IS DISTINCT FROM 'rerender_pages' THEN RAISE EXCEPTION '528 VERIFY: rerender_pages identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'items_result' IS DISTINCT FROM 'items_result' THEN RAISE EXCEPTION '528 VERIFY: items_result identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'js_snippets_deployed.response.data.commit_sha' THEN
        RAISE EXCEPTION '528 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 6 THEN
        RAISE EXCEPTION '528 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'nav-updater'
           AND collected_data #> '{js_snippets_deployed,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '528 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
    END IF;

    -- NOT a cross-fleet string-match negative control (an earlier version of
    -- this migration template had one, and it produced a FALSE POSITIVE:
    -- rerender-pages and nav-updater both legitimately use
    -- js_snippets_deployed.response.data.commit_sha as their OWN commit_sha
    -- source, because both agents' git_commit steps happen to share that
    -- output_field name — two independently-correct migrations, not a leak.
    -- Similarly css-patch-agent and webdesign-agent both use css_deployed.
    -- The REAL protection here is structural, not a runtime check: the guard
    -- above already confirmed exactly ONE live row of this agent type exists,
    -- and the UPDATE's own WHERE clause (type = this agent, below) makes it
    -- impossible for the statement to touch any other agent's row regardless
    -- of what value ends up written. Nothing further to verify.

    RAISE NOTICE '528 OK: nav-updater.complete exposes site_record, nav_refreshed, site_components_rendered, rerender_pages, items_result (unchanged) and commit_sha (new) via result_mapping; confirmed against a live orchestration';
END $$;

COMMIT;
