-- 523 — section-editor's `complete` step exposes `commit_sha` at a CANONICAL
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
-- WHY section-editor maps to `git_result`, NOT the later `deploy_result`:
-- section-editor has TWO git-touching events per run — its OWN direct commit
-- (`deploy_page`, output_field `git_result` — the section EDIT being
-- committed) and a LATER call to `deployer-agent` (`trigger_deploy`, output_field
-- `deploy_result`, described in the live config as "Trigger Cloudflare
-- deployment" — a fleet-wide deployment ping, not this item's own edit).
-- `git_result` is what "the sha of the git_commit that satisfied THIS item"
-- means for a section edit — the edit's own commit, not a downstream trigger's.
-- This is the same principle applied to webdesign-agent below (map to the
-- handler's OWN direct commit, never a nested call's) and is stated here
-- explicitly because — unlike rerender-pages/css-patch-agent/nav-updater,
-- which have only one git-touching step — the choice here is a genuine
-- judgement call, not a mechanical one. `deploy_result` (the Cloudflare
-- trigger's result) is left as a plain identity mapping, unchanged; this file
-- does not touch it or attempt to also expose ITS nested commit.
--
-- WHAT THIS CHANGES: NOTHING for existing readers. Every field currently in
-- `output_fields` is re-declared as an identity mapping (byte-identical
-- response), plus one new `commit_sha` entry sourced from `git_result.response.data.commit_sha`.
--
-- [MEASURED 2026-08-21] a live completed `section-editor` orchestration confirms the
-- path resolves: `collected_data.git_result.response.data.commit_sha` is
-- present (see this file's VERIFY block, which re-checks this against a real
-- row rather than trusting the config string alone).
--
-- ROLLBACK: 523_section_editor_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('section-editor', '523_section_editor_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'section-editor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '523: expected exactly 1 live section-editor row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'section-editor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '523: section-editor has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '523: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '523: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["edit_result","git_result","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '523: complete.output_fields is %, want exactly ["edit_result","git_result","deploy_result"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','action']
          FROM agent_definitions WHERE type='section-editor' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_commit' THEN
        RAISE EXCEPTION '523: deploy_page no longer runs git_commit — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','output_field']
          FROM agent_definitions WHERE type='section-editor' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_result' THEN
        RAISE EXCEPTION '523: deploy_page''s output_field is no longer git_result — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'edit_result', 'edit_result',
               'git_result', 'git_result',
               'deploy_result', 'deploy_result',
               'commit_sha', 'git_result.response.data.commit_sha'
           ),
           true),
       updated_at = NOW()
 WHERE type = 'section-editor'
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
     WHERE type = 'section-editor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '523 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'edit_result' IS DISTINCT FROM 'edit_result' THEN RAISE EXCEPTION '523 VERIFY: edit_result identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'git_result' IS DISTINCT FROM 'git_result' THEN RAISE EXCEPTION '523 VERIFY: git_result identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'deploy_result' IS DISTINCT FROM 'deploy_result' THEN RAISE EXCEPTION '523 VERIFY: deploy_result identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'git_result.response.data.commit_sha' THEN
        RAISE EXCEPTION '523 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 4 THEN
        RAISE EXCEPTION '523 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'section-editor'
           AND collected_data #> '{git_result,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '523 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
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
     WHERE step->'config'->'result_mapping'->>'commit_sha' = 'git_result.response.data.commit_sha'
       AND type <> 'section-editor';
    IF leaked IS NOT NULL THEN
        RAISE EXCEPTION '523 VERIFY: the mapping leaked to steps it was not meant for: %', leaked;
    END IF;

    RAISE NOTICE '523 OK: section-editor.complete exposes edit_result, git_result, deploy_result (unchanged) and commit_sha (new) via result_mapping; confirmed against a live orchestration';
END $$;

COMMIT;
