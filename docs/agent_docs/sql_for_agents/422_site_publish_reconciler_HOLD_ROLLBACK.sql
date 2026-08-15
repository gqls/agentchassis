-- 422 ROLLBACK (sidecar, hand-run only). Reverses the publish reconciler:
-- stops the schedule, retires the orchestrator, restores site-publisher from
-- the agent_definitions_backup snapshot 422 took (found by snapshot_reason +
-- snapshot_taken_at DESC — backup rows copy id AND created_at from the
-- source, so those columns identify NOTHING; LANDMINES), and drops the stamp
-- table. Safe while the seam is opt-in OFF; if any site carries
-- publish_target, its hosted copies simply stop reconciling (nothing is
-- deleted from any bucket).

BEGIN;

UPDATE scheduled_tasks SET enabled = false, updated_at = now()
 WHERE name = 'site-publish-reconciler';

UPDATE agent_definitions SET is_active = false, deleted_at = now()
 WHERE type = 'publish-reconciler' AND deleted_at IS NULL;

-- Restore the pre-422 workflow from the sanctioned backup, and abort loudly
-- if the restorable snapshot is not there — a silent partial rollback is
-- worse than a failed one.
DO $restore$
DECLARE
  snap_config jsonb;
  a text;
BEGIN
  SELECT b.default_config INTO snap_config
    FROM agent_definitions_backup b
   WHERE b.type = 'site-publisher'
     AND b.snapshot_reason = '422 pre-repurpose: upload_to_s3 fossil'
   ORDER BY b.snapshot_taken_at DESC LIMIT 1;

  IF snap_config IS NULL THEN
    RAISE EXCEPTION '422 rollback: no snapshot with reason ''422 pre-repurpose: upload_to_s3 fossil'' in agent_definitions_backup — nothing safe to restore from';
  END IF;
  IF snap_config#>>'{workflow,steps,publish,action}' IS DISTINCT FROM 'upload_to_s3' THEN
    RAISE EXCEPTION '422 rollback: newest 422 snapshot does not hold the pre-change config — refusing to restore from it';
  END IF;

  UPDATE agent_definitions
     SET default_config = snap_config,
         updated_at = now()
   WHERE type = 'site-publisher'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  SELECT default_config#>>'{workflow,steps,publish,action}' INTO a
    FROM agent_definitions
   WHERE type = 'site-publisher'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF a IS DISTINCT FROM 'upload_to_s3' THEN
    RAISE EXCEPTION '422 rollback: post-restore read-back shows %, want upload_to_s3', a;
  END IF;

  -- Bookkeeping: mark the snapshot used (the estate''s partial index on
  -- unrestored snapshots excludes it from future "latest snapshot" lookups).
  UPDATE agent_definitions_backup
     SET restored_at = now()
   WHERE type = 'site-publisher'
     AND snapshot_reason = '422 pre-repurpose: upload_to_s3 fossil'
     AND restored_at IS NULL;
END $restore$;

DROP TABLE IF EXISTS site_publish_checks;

COMMIT;
