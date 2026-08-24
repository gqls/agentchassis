-- ROLLBACK for 600_claims_audit_rotation.sql
-- Removes the rotation task and its rotation stamps. The auditor agent itself is untouched
-- (597 owns that); after this the auditor is dispatchable by hand only, as before.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  DELETE FROM scheduled_tasks WHERE name='claims-audit-rotation';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '600 ROLLBACK: deleted % scheduled_tasks rows, expected 1', n; END IF;
  DELETE FROM site_discovery_rotation WHERE agent_type='claims-auditor';
  RAISE NOTICE '600 ROLLBACK OK (rotation task and stamps removed).';
END $$;

COMMIT;
