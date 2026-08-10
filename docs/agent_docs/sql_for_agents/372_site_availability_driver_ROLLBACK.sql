-- 372_site_availability_driver_ROLLBACK.sql
-- Reverts 372_site_availability_driver (bugs_open/236, 522 half).
-- Safe at any time: the task stops firing, the agent soft-deletes, the rotation
-- stamps clear. Open site_unreachable work items are NOT touched — if any
-- exist, a real outage was detected and deleting the evidence is not a
-- rollback. Close them by hand after reading them.

BEGIN;

UPDATE scheduled_tasks SET enabled = false, updated_at = now()
 WHERE name = 'site-discovery-rotation-availability';

UPDATE agent_definitions SET is_active = false, deleted_at = now(), updated_at = now()
 WHERE type = 'availability-discovery-agent' AND deleted_at IS NULL;

DELETE FROM site_discovery_rotation WHERE agent_type = 'availability-discovery-agent';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'site-discovery-rotation-availability' AND enabled;
  IF n <> 0 THEN RAISE EXCEPTION 'task still enabled'; END IF;
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'availability-discovery-agent' AND deleted_at IS NULL;
  IF n <> 0 THEN RAISE EXCEPTION 'agent still live'; END IF;
END $$;

COMMIT;
