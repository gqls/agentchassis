-- ROLLBACK for 432_wire_evaluate_directory_features_b3d.sql
--
-- Surgical inverse (NOT a restore-from-backup: both agents are other lanes'
-- machinery and may have gained unrelated edits since 432 - restoring the
-- whole default_config from the 432 backups would clobber those). Removes
-- the enrich_directory_features step from both agents and re-points the
-- edges back to their pre-432 targets. Guards refuse if the edges are not
-- in the 432 shape.
--
-- (The 432 pre-update backups in agent_definitions_backup remain available
-- for a full-row restore if a surgical inverse is ever insufficient:
-- snapshot_reason = '432_wire_evaluate_directory_features_b3d.sql: pre-update'.)

BEGIN;

DO $do$
DECLARE
    n int;
    il jsonb;
    cl jsonb;
BEGIN
    -- Exactly-one-active-row guards (added after council round 1 on corr
    -- 47785bb5, 2026-08-15: the two-active-rows landmine - four agent types
    -- carry TWO active rows and only the higher version loads. The forward
    -- migration carried these guards from the start; this file initially did
    -- not, and a SELECT INTO over two rows picks one ARBITRARILY while the
    -- UPDATE hits both. Refuse-on-ambiguity rather than version-pin: a
    -- max(version) pin would silently choose a row in exactly the state
    -- where a human should look first.)
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop does not have exactly one active row (found %) - resolve before rolling back', n;
    END IF;
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'domain-research-classifier' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '432 ROLLBACK: domain-research-classifier does not have exactly one active row (found %) - resolve before rolling back', n;
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO il FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF NOT (il ? 'enrich_directory_features') THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop has no enrich_directory_features step - nothing to roll back';
    END IF;
    IF il#>>'{enrich_news_feed,next_step}' IS DISTINCT FROM 'enrich_directory_features' THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop enrich_news_feed.next_step is not in the 432 shape - re-check before rolling back';
    END IF;
    -- Round-2 advisory (guardian seat, corr 47785bb5, 2026-08-15): pin EVERY
    -- edge this file is about to overwrite or orphan, not just the entry
    -- edge. If a third lane has re-pointed the news error path, or spliced
    -- its own step AFTER enrich_directory_features (so its outbound edges no
    -- longer point at load_audit_state), writing the pre-432 values back
    -- would silently disconnect that lane's work - refuse instead.
    IF il#>>'{enrich_news_feed,config,error_step}' IS DISTINCT FROM 'enrich_directory_features' THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop enrich_news_feed.config.error_step is not in the 432 shape (found %) - a later edit re-pointed it; re-check before rolling back', il#>>'{enrich_news_feed,config,error_step}';
    END IF;
    IF il#>>'{enrich_directory_features,next_step}' IS DISTINCT FROM 'load_audit_state' THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop enrich_directory_features.next_step is not load_audit_state (found %) - a later splice hangs off this step and would be orphaned; re-check before rolling back', il#>>'{enrich_directory_features,next_step}';
    END IF;
    IF il#>>'{enrich_directory_features,config,error_step}' IS DISTINCT FROM 'load_audit_state' THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop enrich_directory_features.config.error_step is not load_audit_state (found %) - a later edit re-pointed it; re-check before rolling back', il#>>'{enrich_directory_features,config,error_step}';
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO cl FROM agent_definitions
    WHERE type = 'domain-research-classifier' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF NOT (cl ? 'enrich_directory_features') THEN
        RAISE EXCEPTION '432 ROLLBACK: classifier has no enrich_directory_features step - nothing to roll back';
    END IF;
    IF cl#>>'{write_classification_spec,next_step}' IS DISTINCT FROM 'enrich_directory_features' THEN
        RAISE EXCEPTION '432 ROLLBACK: classifier write_classification_spec.next_step is not in the 432 shape - re-check before rolling back';
    END IF;
    -- Same round-2 advisory, classifier side: pin the step's outbound edges.
    IF cl#>>'{enrich_directory_features,next_step}' IS DISTINCT FROM 'write_content_direction_spec' THEN
        RAISE EXCEPTION '432 ROLLBACK: classifier enrich_directory_features.next_step is not write_content_direction_spec (found %) - a later splice hangs off this step and would be orphaned; re-check before rolling back', cl#>>'{enrich_directory_features,next_step}';
    END IF;
    IF cl#>>'{enrich_directory_features,config,error_step}' IS DISTINCT FROM 'write_content_direction_spec' THEN
        RAISE EXCEPTION '432 ROLLBACK: classifier enrich_directory_features.config.error_step is not write_content_direction_spec (found %) - a later edit re-pointed it; re-check before rolling back', cl#>>'{enrich_directory_features,config,error_step}';
    END IF;
END $do$;

UPDATE agent_definitions
SET default_config =
    jsonb_set(
        jsonb_set(
            default_config #- '{workflow,steps,enrich_directory_features}',
            '{workflow,steps,enrich_news_feed,next_step}',
            '"load_audit_state"'::jsonb
        ),
        '{workflow,steps,enrich_news_feed,config,error_step}',
        '"load_audit_state"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config =
    jsonb_set(
        default_config #- '{workflow,steps,enrich_directory_features}',
        '{workflow,steps,write_classification_spec,next_step}',
        '"write_content_direction_spec"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $do$
DECLARE
    il jsonb;
    cl jsonb;
BEGIN
    SELECT default_config#>'{workflow,steps}' INTO il FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF il ? 'enrich_directory_features'
       OR il#>>'{enrich_news_feed,next_step}' IS DISTINCT FROM 'load_audit_state'
       OR il#>>'{enrich_news_feed,config,error_step}' IS DISTINCT FROM 'load_audit_state' THEN
        RAISE EXCEPTION '432 ROLLBACK verify: improvement-loop not restored to the pre-432 shape';
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO cl FROM agent_definitions
    WHERE type = 'domain-research-classifier' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cl ? 'enrich_directory_features'
       OR cl#>>'{write_classification_spec,next_step}' IS DISTINCT FROM 'write_content_direction_spec' THEN
        RAISE EXCEPTION '432 ROLLBACK verify: classifier not restored to the pre-432 shape';
    END IF;
END $do$;

COMMIT;
