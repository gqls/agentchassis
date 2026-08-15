-- ROLLBACK for 429_directory_publish_trigger_kind_aware_fan_out.sql
--
-- Restores default_config, input_contract and description for BOTH rows
-- (model-directory-publisher, model-directory-trigger) from the two
-- agent_definitions_backup rows the forward migration took via the two-arg
-- snapshot_agent() (which copies the full row, contracts included).
-- Order by snapshot_taken_at, NOT created_at - the backup keeps the SOURCE
-- row's created_at verbatim (LANDMINES).
--
-- After this, the publish leg is back to: kind-blind model-only trigger
-- gating + the hard-coded model->company->protocol publisher chain. The
-- finance kinds become structurally unpublishable again.

BEGIN;

DO $do$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '429_directory_publish_trigger_kind_aware_fan_out.sql: pre-update'
      AND type IN ('model-directory-publisher','model-directory-trigger');
    IF n <> 2 THEN
        RAISE EXCEPTION '429 ROLLBACK: expected 2 backup rows with the 429 pre-update reason, found % - nothing restored', n;
    END IF;

    -- Refuse to "roll back" a row that is not in the 429 state - that would
    -- clobber someone else's later change with the pre-429 config.
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config#>'{workflow,steps}' ? 'render_directory_json';
    IF n <> 1 THEN
        RAISE EXCEPTION '429 ROLLBACK: live publisher is not in the 429 shape (render_directory_json absent) - re-check before rolling back';
    END IF;
END $do$;

UPDATE agent_definitions live
SET default_config = bak.default_config,
    input_contract = bak.input_contract,
    description    = bak.description,
    updated_at     = NOW()
FROM (
    SELECT DISTINCT ON (type) type, default_config, input_contract, description
    FROM agent_definitions_backup
    WHERE snapshot_reason = '429_directory_publish_trigger_kind_aware_fan_out.sql: pre-update'
    ORDER BY type, snapshot_taken_at DESC
) bak
WHERE live.type = bak.type AND live.is_active
  AND COALESCE(live.is_snapshot, false) = false AND live.deleted_at IS NULL;

DO $do$
DECLARE
    pub_steps jsonb;
    trg_query text;
BEGIN
    SELECT default_config#>'{workflow,steps}' INTO pub_steps FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF NOT (pub_steps ? 'render_adoption_json') THEN
        RAISE EXCEPTION '429 ROLLBACK verify: publisher is not back to the post-411 7-step chain';
    END IF;

    SELECT default_config#>>'{workflow,steps,find_directory_sites,config,query}' INTO trg_query
    FROM agent_definitions
    WHERE type = 'model-directory-trigger' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF trg_query IS NULL OR position('LIMIT 5' IN trg_query) = 0 THEN
        RAISE EXCEPTION '429 ROLLBACK verify: trigger query is not back to the pre-429 shape';
    END IF;
END $do$;

COMMIT;
