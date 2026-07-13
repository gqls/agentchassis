-- ============================================================================
-- Phase 2E — pre-migration backup of the image-build-handler agent definition
--
-- Phase 2E adds a new branch to the workflow for variant items
-- (unfulfilled_hero_variant). The existing logo and hero paths are not
-- modified; they continue to use the same input format and steps as before.
--
-- Rollback: restore default_config from the backup row.
-- ============================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS agent_def_backups_phase2e AS
SELECT id, type, version, default_config, NOW() AS backed_up_at
FROM agent_definitions
WHERE type = 'image-build-handler'
  AND version = (SELECT MAX(version) FROM agent_definitions WHERE type = 'image-build-handler')
LIMIT 1;

-- If the backup table already had rows, this re-runs cleanly because we
-- INSERT only if the current row isn't already there.
INSERT INTO agent_def_backups_phase2e (id, type, version, default_config, backed_up_at)
SELECT id, type, version, default_config, NOW()
FROM agent_definitions
WHERE type = 'image-build-handler'
  AND version = (SELECT MAX(version) FROM agent_definitions WHERE type = 'image-build-handler')
  AND NOT EXISTS (
    SELECT 1 FROM agent_def_backups_phase2e b
    WHERE b.id = agent_definitions.id
      AND b.version = agent_definitions.version
      AND b.backed_up_at > NOW() - INTERVAL '5 minutes'
  );

SELECT id, type, version, backed_up_at,
       jsonb_pretty(default_config #> '{workflow,steps,check_item_type}') AS check_item_type
FROM agent_def_backups_phase2e
ORDER BY backed_up_at DESC
LIMIT 1;

COMMIT;


-- ============================================================================
-- RESTORE — only if reverting Phase 2E
-- ============================================================================
-- BEGIN;
-- UPDATE agent_definitions ad
-- SET default_config = b.default_config
-- FROM agent_def_backups_phase2e b
-- WHERE ad.id = b.id
--   AND ad.version = b.version
--   AND b.backed_up_at = (SELECT MAX(backed_up_at) FROM agent_def_backups_phase2e
--                        WHERE id = ad.id AND version = ad.version);
-- COMMIT;
