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
    il jsonb;
    cl jsonb;
BEGIN
    SELECT default_config#>'{workflow,steps}' INTO il FROM agent_definitions
    WHERE type = 'improvement-loop' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF NOT (il ? 'enrich_directory_features') THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop has no enrich_directory_features step - nothing to roll back';
    END IF;
    IF il#>>'{enrich_news_feed,next_step}' IS DISTINCT FROM 'enrich_directory_features' THEN
        RAISE EXCEPTION '432 ROLLBACK: improvement-loop enrich_news_feed.next_step is not in the 432 shape - re-check before rolling back';
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
