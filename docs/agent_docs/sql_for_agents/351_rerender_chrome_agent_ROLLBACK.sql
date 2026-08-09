-- 351_rerender_chrome_agent_ROLLBACK.sql
--
-- Soft-deletes the rerender-chrome agent definition (bugs_open/226 stamp-only
-- chrome renderer). Config-only row; no data loss. In-flight orchestrations of
-- this type are unaffected (they carry their own workflow_plan) — this only
-- stops NEW dispatches from resolving the type.

BEGIN;

UPDATE agent_definitions
SET is_active = false, deleted_at = now()
WHERE type = 'rerender-chrome' AND deleted_at IS NULL;

DO $$
DECLARE remaining integer;
BEGIN
  SELECT count(*) INTO remaining
  FROM agent_definitions
  WHERE type = 'rerender-chrome' AND is_active AND deleted_at IS NULL;
  IF remaining <> 0 THEN
    RAISE EXCEPTION 'rollback guard: % active rerender-chrome row(s) remain', remaining;
  END IF;
END $$;

COMMIT;
