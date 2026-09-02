-- ROLLBACK for 687: restores default_config from the backup table 687's own
-- migration created (agent_definitions_bak_687), or from the snapshot_agent
-- row it took first if that table is gone.
BEGIN;

UPDATE agent_definitions ad
SET default_config = bak.default_config,
    updated_at = NOW()
FROM agent_definitions_bak_687 bak
WHERE ad.id = bak.id
  AND ad.type = 'build-site-planner' AND ad.is_active
  AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL;

DO $$
DECLARE t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('{{toJSON .site_specs.specs.strategy}}' in t) > 0 THEN
        RAISE EXCEPTION '687 ROLLBACK: verify failed — toJSON interpolation still present after restore';
    END IF;
END $$;

COMMIT;
