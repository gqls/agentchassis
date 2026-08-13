-- ROLLBACK for 400_thunder_provision_wait_timeout.sql
--
-- Run BY HAND, deliberately. The migration runner never executes an
-- UPPERCASE-suffixed sidecar.
--
-- Safe to run: the adapter treats a missing column as "use my compiled-in
-- default" (store.LoadConfig falls back rather than erroring), so dropping this
-- restores the previous hardcoded behaviour — including bugs_open/258 defect 2,
-- which destroys any instance slower than the compiled default to boot.
--
-- Prefer UPDATEing the value over dropping the column: the whole point of the
-- column is that the deadline is tunable without a build.

BEGIN;

ALTER TABLE thunder_config DROP COLUMN IF EXISTS provision_wait_timeout_seconds;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='thunder_config' AND column_name='provision_wait_timeout_seconds'
  ) THEN
    RAISE EXCEPTION '400 ROLLBACK: column still present';
  END IF;
END $$;

COMMIT;
