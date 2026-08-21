-- 519 — page-rerender's `complete` step exposes `commit_sha` at a CANONICAL
--       top-level key, converting its result contract from `output_fields`
--       (list) to `result_mapping` (explicit target<-source pairs).
--       RFC_029 §10.13 step 5's real gate (bdl/`commit_sha`) — handler half.
--       CONFIG ONLY — live on apply.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- `build-dispatch-loop` calls a work item's handler via `call_agent`, then
-- calls `complete_work_item`, which writes `result.commit_sha` FROM the
-- handler's own reply when present (`load_work_item_actions.go:937`,
-- `Optional` field). Today `commit_sha` reaches `complete_work_item` only via
-- the resolver's whole-tree search, because no handler exposes it at a
-- consistent name — bugs_open/334, and the CONTRIB/reply exchange at
-- `docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/CONTRIB_2026-08-20_from_staged_component_build_commit_sha_resolves_by_guess.md`.
-- The 315 lane's answer: standardise the SOURCE, one handler at a time, so
-- `complete_work_item` can eventually use a single explicit mapping
-- (`"commit_sha?": "handler_result.response.commit_sha"`, theirs to add once
-- every handler conforms) instead of the search.
--
-- ============================================================================
-- WHY page-rerender FIRST, and why it is NOT one of "~16 agents with a
-- git_commit step" — it is the ONE agent that actually matters most
-- ============================================================================
-- `handler_agent` is dispatched dynamically per work item
-- (`current_item.handler_agent`, `build-dispatch-loop`'s `spawn_handler` step).
-- [MEASURED 2026-08-21] the REAL population, from `site_work_items.handler_agent`
-- (a top-level column, not `spec` — the census this file is built from queries
-- the column directly): of ~34 distinct handler types ever dispatched,
-- **`page-rerender` is 4,810 of the last 7 days' items, by far the largest**,
-- and it has ONE direct `git_commit` step. Several agents this bug's earlier
-- census flagged (content-feed-orchestrator, nav-link-fixer, rerender-site,
-- site-asset-renderer, model-directory-publisher, deployer-agent, site-deployer,
-- pageflow-builder, page-rebuild, site-work-orchestrator, report-builder) DO
-- have a `git_commit` step but **0 occurrences as a live `handler_agent`
-- value** — they are never bdl's handler, so they cannot be feeding THIS
-- conflict class regardless of their own internal shape. Conversely
-- `page-build-handler` (662/7d, the SECOND largest) has NO `git_commit` step
-- of its own — it reaches one INDIRECTLY, one hop away, by calling
-- `page-rerender` via its own `deploy_page` (`call_agent`) step — so
-- page-build-handler's own fix (a later migration) is only meaningful AFTER
-- this one ships, and depends on this file's canonical key existing.
--
-- Six other real, direct handlers remain after this one: rerender-pages,
-- css-patch-agent, section-editor, webdesign-agent, nav-updater (each real but
-- an order of magnitude smaller by volume) — separate migrations, this lane's
-- next work. Everything else in the real handler_agent population (confirmed
-- by reading each one's full action list, not inferred): asset-deployer
-- (`deploy_image_asset` — S3, not git), image-build-handler (calls
-- asset-deployer + image generators, no git), diagnose-orchestrator (calls a
-- diagnoser, no git), tool-generator/tool-improver/component-creator/
-- component-template-fixer/content-gap-planner/internal-linker (all defer the
-- actual deploy to a LATER, separate `page-rerender` work item via
-- `create_rerender_item`/`create_work_item` — so THEIR `complete_work_item`
-- call correctly carries no commit_sha; the commit happens in a different
-- orchestration entirely), tool-acceptance-agent/tool-auditor/human-review/
-- required-fields-missing-handler/color-variable-fixer (no git-adjacent action
-- at all). For all of these, absence is CORRECT and this file does not touch
-- them.
--
-- ============================================================================
-- THE MECHANISM — why an ADDED key cannot work, and a MODE CONVERSION is required
-- ============================================================================
-- The live `complete` step (`workflow.steps.complete`):
--     {"output_fields": ["rendered_page", "deploy_result"]}
-- `output_fields` is `ResultModeFields` (`result_contract.go`): for each string
-- `fn` in the list, `result[fn] = ExtractNestedField(collectedData, fn)` — the
-- SAME string is both the extraction path and the response key. Adding
-- `"deploy_result.response.data.commit_sha"` as a new entry would produce a
-- response key that is that LITERAL DOTTED STRING, not `commit_sha` — invisible
-- to any dot-path reader, including the very resolver this fix serves. Routing
-- through `extract_fields` first does not help either: it always wraps its
-- result in a map (`result[target] = value`), so the extracted value arrives
-- one nesting level too deep however it is named. The only mode that builds an
-- explicit, independently-named target<-source map is `result_mapping`
-- (`ResultModeMapping`) — mutually exclusive with `output_fields`
-- (`ResolveResultSpec`'s precedence list matches exactly one key), so the fix
-- is a CONVERSION: keep every currently-exposed field as an identity mapping,
-- add `commit_sha` as the one new entry.
--
-- ============================================================================
-- THE PATH, TRACED AND MEASURED, NOT ASSUMED
-- ============================================================================
-- Workflow order: `deploy_page` (`git_commit`, `output_field: deploy_result`,
-- no `error_step` — a failed commit fails the orchestration outright, it does
-- not reach a graceful no-commit completion) -> `update_status`
-- (`update_page_status`) -> `complete`. So every run that reaches `complete`
-- has a `deploy_result` carrying a real commit.
--
-- [MEASURED 2026-08-21 12:0xZ, a live completed orchestration]
--     collected_data.deploy_result.response.data.commit_sha = "428c4730917449c3602e2605fbf3931047cace8b"
-- confirming the exact path this migration maps, at the artefact, not inferred
-- from the action's Go source alone.
--
-- `complete_skipped` (reached when there are no components to render, deploy
-- never runs) exposes only `["rendered_page"]` today and is UNCHANGED by this
-- file — absence of `commit_sha` there is correct and this file does not add
-- it, on purpose.
--
-- ============================================================================
-- WHAT THIS CHANGES, STATED PLAINLY: NOTHING FOR EXISTING READERS
-- ============================================================================
-- `result_mapping = {"rendered_page":"rendered_page","deploy_result":"deploy_result","commit_sha":"deploy_result.response.data.commit_sha"}`
-- reproduces `rendered_page` and `deploy_result` byte-for-byte (identity maps
-- of already-top-level keys) and adds exactly one new key. Any existing reader
-- of `handler_result.response.rendered_page` or `.deploy_result` sees no
-- change; a NEW reader can use `handler_result.response.commit_sha` once the
-- `complete_work_item` wire (staged_component_build's, next, after this lands)
-- reads it explicitly instead of guessing.
--
-- ROLLBACK: 519_page_rerender_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('page-rerender', '519_page_rerender_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '519: expected exactly 1 live page-rerender row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '519: page-rerender has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '519: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '519: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    -- The premise is EXACTLY this output_fields list. If it has changed, some
    -- OTHER field is exposed today that this file does not know to preserve —
    -- refuse rather than silently drop it.
    IF step->'config'->'output_fields' <> '["rendered_page","deploy_result"]'::jsonb THEN
        RAISE EXCEPTION '519: complete.output_fields is %, want exactly ["rendered_page","deploy_result"] — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    -- The path this migration maps depends on deploy_page still being the
    -- git_commit step whose output_field is deploy_result. If either has
    -- changed, the mapped source would silently resolve to nothing.
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','action']
          FROM agent_definitions WHERE type='page-rerender' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_commit' THEN
        RAISE EXCEPTION '519: deploy_page no longer runs git_commit — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','deploy_page','output_field']
          FROM agent_definitions WHERE type='page-rerender' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'deploy_result' THEN
        RAISE EXCEPTION '519: deploy_page''s output_field is no longer deploy_result — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'rendered_page', 'rendered_page',
               'deploy_result', 'deploy_result',
               'commit_sha',    'deploy_result.response.data.commit_sha'
           ),
           true),
       updated_at = NOW()
 WHERE type = 'page-rerender'
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
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '519 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'rendered_page' IS DISTINCT FROM 'rendered_page' THEN
        RAISE EXCEPTION '519 VERIFY: rendered_page identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'deploy_result' IS DISTINCT FROM 'deploy_result' THEN
        RAISE EXCEPTION '519 VERIFY: deploy_result identity mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'deploy_result.response.data.commit_sha' THEN
        RAISE EXCEPTION '519 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 3 THEN
        RAISE EXCEPTION '519 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    -- SANITY CHECK against a REAL orchestration: replay the mapping's own
    -- extraction against actual collected_data, so the check is not only that
    -- the config STRING is correct but that it RESOLVES.
    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'page-rerender'
           AND collected_data #> '{deploy_result,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '519 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
    END IF;

    -- NEGATIVE CONTROL, recursive (sub_workflow-safe): no other live step may
    -- have acquired this exact three-key mapping.
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
     WHERE step->'config'->'result_mapping'->>'commit_sha' = 'deploy_result.response.data.commit_sha'
       AND type <> 'page-rerender';
    IF leaked IS NOT NULL THEN
        RAISE EXCEPTION '519 VERIFY: the mapping leaked to steps it was not meant for: %', leaked;
    END IF;

    RAISE NOTICE '519 OK: page-rerender.complete exposes rendered_page, deploy_result (unchanged) and commit_sha (new) via result_mapping; confirmed against a live orchestration';
END $$;

COMMIT;
