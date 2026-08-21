-- 522 — css-patch-agent's `complete` step exposes `commit_sha` at a CANONICAL
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
-- WHY css-patch-agent: real bdl handler (85/7d), single direct git_commit step
-- (`deploy_css`), single path to `complete` directly (deploy_css -> complete).
-- `complete_error` (output_fields=["css_fix"] only, no css_deployed — the
-- deploy never happened) and `complete_no_css` (reached before deploy_css
-- ever runs) are UNCHANGED by this file: absence of commit_sha on those two
-- paths is correct and this migration does not touch them, on purpose.
--
-- WHAT THIS CHANGES: NOTHING for existing readers. Every field currently in
-- `output_fields` is re-declared as an identity mapping (byte-identical
-- response), plus one new `commit_sha` entry sourced from `css_deployed.response.data.commit_sha`.
--
-- [MEASURED 2026-08-21] a live completed `css-patch-agent` orchestration confirms the
-- path resolves: `collected_data.css_deployed.response.data.commit_sha` is
-- present (see this file's VERIFY block, which re-checks this against a real
-- row rather than trusting the config string alone).
--
-- ROLLBACK: 522_css_patch_agent_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('css-patch-agent', '522_css_patch_agent_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'css-patch-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '522: expected exactly 1 live css-patch-agent row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'css-patch-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '522: css-patch-agent has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '522: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '522: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["css_fix","css_deployed"]'::jsonb THEN
        RAISE EXCEPTION '522: complete.output_fields is %, want exactly ["css_fix","css_deployed"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_css','action']
          FROM agent_definitions WHERE type='css-patch-agent' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_commit' THEN
        RAISE EXCEPTION '522: deploy_css no longer runs git_commit — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_css','output_field']
          FROM agent_definitions WHERE type='css-patch-agent' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'css_deployed' THEN
        RAISE EXCEPTION '522: deploy_css''s output_field is no longer css_deployed — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'css_fix', 'css_fix',
               'css_deployed', 'css_deployed',
               'commit_sha', 'css_deployed.response.data.commit_sha'
           ),
           true),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
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
     WHERE type = 'css-patch-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '522 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'css_fix' IS DISTINCT FROM 'css_fix' THEN RAISE EXCEPTION '522 VERIFY: css_fix identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'css_deployed' IS DISTINCT FROM 'css_deployed' THEN RAISE EXCEPTION '522 VERIFY: css_deployed identity mapping missing or wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'css_deployed.response.data.commit_sha' THEN
        RAISE EXCEPTION '522 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 3 THEN
        RAISE EXCEPTION '522 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'css-patch-agent'
           AND collected_data #> '{css_deployed,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '522 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
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
     WHERE step->'config'->'result_mapping'->>'commit_sha' = 'css_deployed.response.data.commit_sha'
       AND type <> 'css-patch-agent';
    IF leaked IS NOT NULL THEN
        RAISE EXCEPTION '522 VERIFY: the mapping leaked to steps it was not meant for: %', leaked;
    END IF;

    RAISE NOTICE '522 OK: css-patch-agent.complete exposes css_fix, css_deployed (unchanged) and commit_sha (new) via result_mapping; confirmed against a live orchestration';
END $$;

COMMIT;
