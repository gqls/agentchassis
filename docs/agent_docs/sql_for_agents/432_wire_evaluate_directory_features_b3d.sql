-- 432 - B3d: wire evaluate_directory_features (Phase B directory-kind
-- recommender, in the binary since v1.0.1301, dispatched by NOTHING until
-- this migration) into the two places that decide what a site carries:
--
--   1. improvement-loop - a new enrich_directory_features step between
--      enrich_news_feed and load_audit_state, so every audited site's
--      classification spec gains/refreshes its per-kind directory
--      recommendation each cycle. Mirrors the enrich_news_feed step's exact
--      shape (config.error_step continue-on-failure: an enrichment failure
--      must never kill the audit loop - the property bug 291's re-point
--      established, preserved through the new step).
--   2. domain-research-classifier - the same step between
--      write_classification_spec and write_content_direction_spec, so a
--      GREENFIELD build carries the flag at plan time (the content-gap
--      planner reads content_features; without this, a new finance site
--      would only learn it wants a directory on its first improvement-loop
--      cycle, after the site was already planned).
--
-- The action is a deterministic no-LLM mirror of evaluate_news_feed
-- (feed_directory_recommendation_action.go): reads the CURRENT
-- classification spec, matches the vertical against the Phase B starter
-- kinds, deep-merges content_features.<spec_key> and supersede-then-inserts.
-- No match / no spec => NO WRITE (an opt-in flag's safe default), so wiring
-- it fleet-wide cannot flip any existing site: only the finance verticals'
-- signals appear in verticalDirectoryMap at all.
--
-- Surgical jsonb_set edits, NOT whole-workflow replacement: both agents are
-- other lanes' machinery; this migration touches exactly two edges and adds
-- one step per agent, and its guards pin the LIVE edge values so any drift
-- refuses rather than clobbers (the 2026-08-15 handoff's warning: 291
-- re-pointed improvement-loop edges - the seed is history, the live row is
-- fact, and this file was written FROM the live rows).
--
-- Ordering note (improvement-loop): site_id comes from site_record.site_id
-- (ensure_site_record runs first); in the classifier it is input_data.site_id
-- (the sibling write steps' idiom). Each context keeps its own reference.
--
-- Config-only: live immediately, no image roll. Both updates in ONE
-- transaction. Zero in-flight orchestrations for either type checked at
-- authoring time; in-flight runs carry their own workflow snapshot anyway.

SELECT snapshot_agent('improvement-loop',           '432_wire_evaluate_directory_features_b3d.sql: pre-update');
SELECT snapshot_agent('domain-research-classifier', '432_wire_evaluate_directory_features_b3d.sql: pre-update');

BEGIN;

-- ── Pre-flight guards ──────────────────────────────────────────────────────
DO $do$
DECLARE
    n int;
    il jsonb;
    cl jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '432: improvement-loop does not have exactly one active row (found %)', n;
    END IF;
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'domain-research-classifier' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '432: domain-research-classifier does not have exactly one active row (found %)', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '432_wire_evaluate_directory_features_b3d.sql: pre-update'
      AND type IN ('improvement-loop','domain-research-classifier');
    IF n <> 2 THEN
        RAISE EXCEPTION '432: expected 2 pre-update backup rows, found % - snapshot_agent did not run', n;
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO il FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF il ? 'enrich_directory_features' THEN
        RAISE EXCEPTION '432: already applied - improvement-loop carries enrich_directory_features';
    END IF;
    -- Pin the LIVE 291-shaped edges this migration re-points.
    IF il#>>'{enrich_news_feed,next_step}' IS DISTINCT FROM 'load_audit_state' THEN
        RAISE EXCEPTION '432: improvement-loop enrich_news_feed.next_step is not load_audit_state (found %) - the row has drifted, re-read before applying', il#>>'{enrich_news_feed,next_step}';
    END IF;
    IF il#>>'{enrich_news_feed,config,error_step}' IS DISTINCT FROM 'load_audit_state' THEN
        RAISE EXCEPTION '432: improvement-loop enrich_news_feed.config.error_step is not load_audit_state - the row has drifted, re-read before applying';
    END IF;
    IF NOT (il ? 'load_audit_state') THEN
        RAISE EXCEPTION '432: improvement-loop has no load_audit_state step - the row has drifted, re-read before applying';
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO cl FROM agent_definitions
    WHERE type = 'domain-research-classifier' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cl ? 'enrich_directory_features' THEN
        RAISE EXCEPTION '432: already applied - domain-research-classifier carries enrich_directory_features';
    END IF;
    IF cl#>>'{write_classification_spec,next_step}' IS DISTINCT FROM 'write_content_direction_spec' THEN
        RAISE EXCEPTION '432: classifier write_classification_spec.next_step is not write_content_direction_spec (found %) - the row has drifted, re-read before applying', cl#>>'{write_classification_spec,next_step}';
    END IF;
    IF NOT (cl ? 'write_content_direction_spec') THEN
        RAISE EXCEPTION '432: classifier has no write_content_direction_spec step - the row has drifted, re-read before applying';
    END IF;
END $do$;

-- ── 1. improvement-loop: insert the step, re-point both news edges ─────────
UPDATE agent_definitions
SET default_config =
    jsonb_set(
        jsonb_set(
            jsonb_set(
                default_config,
                '{workflow,steps,enrich_directory_features}',
                $json${
                    "action": "evaluate_directory_features",
                    "config": {
                        "site_id": "site_record.site_id",
                        "error_step": "load_audit_state"
                    },
                    "next_step": "load_audit_state",
                    "description": "Refresh per-kind directory recommendation from classification spec (deterministic, no LLM; Phase B starter kinds; no match = no write)",
                    "output_field": "directory_features_enrichment"
                }$json$::jsonb
            ),
            '{workflow,steps,enrich_news_feed,next_step}',
            '"enrich_directory_features"'::jsonb
        ),
        '{workflow,steps,enrich_news_feed,config,error_step}',
        '"enrich_directory_features"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 2. domain-research-classifier: insert the step, re-point one edge ──────
UPDATE agent_definitions
SET default_config =
    jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,enrich_directory_features}',
            $json${
                "action": "evaluate_directory_features",
                "config": {
                    "site_id": "input_data.site_id",
                    "error_step": "write_content_direction_spec"
                },
                "next_step": "write_content_direction_spec",
                "description": "Directory-kind recommendation at plan time, immediately after the classification spec exists (deterministic, no LLM; a failure continues the build - enrichment must not kill a greenfield classification)",
                "output_field": "directory_features_enrichment"
            }$json$::jsonb
        ),
        '{workflow,steps,write_classification_spec,next_step}',
        '"enrich_directory_features"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Verify with DO/RAISE (IS DISTINCT FROM: a missing path is NULL and a
--    plain <> comparison can never fire) ──────────────────────────────────
DO $do$
DECLARE
    il jsonb;
    cl jsonb;
BEGIN
    SELECT default_config#>'{workflow,steps}' INTO il FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF il#>>'{enrich_directory_features,action}' IS DISTINCT FROM 'evaluate_directory_features' THEN
        RAISE EXCEPTION '432 verify: improvement-loop enrich_directory_features step missing or wrong action';
    END IF;
    IF il#>>'{enrich_directory_features,config,site_id}' IS DISTINCT FROM 'site_record.site_id' THEN
        RAISE EXCEPTION '432 verify: improvement-loop step site_id reference wrong';
    END IF;
    IF il#>>'{enrich_news_feed,next_step}' IS DISTINCT FROM 'enrich_directory_features'
       OR il#>>'{enrich_news_feed,config,error_step}' IS DISTINCT FROM 'enrich_directory_features' THEN
        RAISE EXCEPTION '432 verify: improvement-loop news edges not re-pointed';
    END IF;
    IF il#>>'{enrich_directory_features,next_step}' IS DISTINCT FROM 'load_audit_state'
       OR il#>>'{enrich_directory_features,config,error_step}' IS DISTINCT FROM 'load_audit_state' THEN
        RAISE EXCEPTION '432 verify: improvement-loop new step does not continue to load_audit_state on both paths';
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO cl FROM agent_definitions
    WHERE type = 'domain-research-classifier' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cl#>>'{enrich_directory_features,action}' IS DISTINCT FROM 'evaluate_directory_features' THEN
        RAISE EXCEPTION '432 verify: classifier enrich_directory_features step missing or wrong action';
    END IF;
    IF cl#>>'{enrich_directory_features,config,site_id}' IS DISTINCT FROM 'input_data.site_id' THEN
        RAISE EXCEPTION '432 verify: classifier step site_id reference wrong';
    END IF;
    IF cl#>>'{write_classification_spec,next_step}' IS DISTINCT FROM 'enrich_directory_features' THEN
        RAISE EXCEPTION '432 verify: classifier edge not re-pointed';
    END IF;
    IF cl#>>'{enrich_directory_features,next_step}' IS DISTINCT FROM 'write_content_direction_spec'
       OR cl#>>'{enrich_directory_features,config,error_step}' IS DISTINCT FROM 'write_content_direction_spec' THEN
        RAISE EXCEPTION '432 verify: classifier new step does not continue to write_content_direction_spec on both paths';
    END IF;
END $do$;

COMMIT;

-- Post-apply (hand-run):
--   The improvement-loop wiring self-proves on its next scheduled cycle: any
--   audited FINANCE-vertical site gains content_features.<kind>_directory in
--   its current classification spec; every other site's spec is untouched
--   (no-match = no-write). Nothing to force-trigger for the classifier until
--   the next greenfield build (Phase C pilot will be the proof).
--   Check: SELECT s.domain, ss.data->'content_features' FROM site_specs ss
--          JOIN sites s ON s.id=ss.site_id
--          WHERE ss.aspect='classification' AND ss.is_current
--            AND ss.source_agent='evaluate_directory_features';
--
-- ROLLBACK: 432_wire_evaluate_directory_features_b3d_ROLLBACK.sql
