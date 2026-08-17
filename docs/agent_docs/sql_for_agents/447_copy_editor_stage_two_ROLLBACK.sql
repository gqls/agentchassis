-- 447_copy_editor_stage_two_ROLLBACK.sql
--
-- Reverses 447 by soft-deleting the seeded row, matching how the estate retires an
-- agent definition (retire, never hard-delete: a snapshot or a historical
-- orchestration may still reference the row, and a hard DELETE turns those into
-- dangling ids).
--
-- Safe to run more than once. Asserts the row is gone rather than reporting success
-- on a no-op.

BEGIN;

UPDATE agent_definitions
   SET is_active = false,
       deleted_at = now(),
       updated_at = now()
 WHERE type = 'copy-editor'
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION 'rollback failed: % live copy-editor row(s) remain', n;
  END IF;
  RAISE NOTICE 'copy-editor retired (soft-deleted)';
END $$;

COMMIT;
