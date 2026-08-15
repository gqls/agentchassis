-- 422 ROLLBACK (sidecar, hand-run only). Reverses the publish reconciler:
-- stops the schedule, retires the orchestrator, restores site-publisher from
-- the snapshot 422 took, and drops the stamp table. Safe while the seam is
-- opt-in OFF; if any site carries publish_target, its hosted copies simply
-- stop reconciling (nothing is deleted from any bucket).

BEGIN;

UPDATE scheduled_tasks SET enabled = false, updated_at = now()
 WHERE name = 'site-publish-reconciler';

UPDATE agent_definitions SET is_active = false, deleted_at = now()
 WHERE type = 'publish-reconciler' AND deleted_at IS NULL;

-- Restore the pre-422 site-publisher workflow from the snapshot 422 wrote.
UPDATE agent_definitions live
   SET default_config = snap.default_config,
       updated_at = now()
  FROM (SELECT default_config
          FROM agent_definitions
         WHERE type = 'site-publisher' AND COALESCE(is_snapshot, false) = true
           AND description LIKE '%migration 422%'
         ORDER BY created_at DESC LIMIT 1) snap
 WHERE live.type = 'site-publisher'
   AND live.is_active AND COALESCE(live.is_snapshot, false) = false
   AND live.deleted_at IS NULL;

DROP TABLE IF EXISTS site_publish_checks;

COMMIT;
