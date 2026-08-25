-- ROLLBACK for 610 — restores the COLUMN but deliberately NOT the numbers.
--
-- The column comes back as NULL/0 everywhere. That is the correct rollback: the old
-- values were wrong in both directions (missing ~1,900 real bindings while counting
-- non-usages), so restoring them would restore a lie. If a binary genuinely needs the
-- column to exist again, DEFAULT 0 is what its INSERT expected.
\set ON_ERROR_STOP on
BEGIN;
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS usage_count integer DEFAULT 0;
COMMENT ON COLUMN content_components.usage_count IS
  'RESTORED BY ROLLBACK of migration 610 and INTENTIONALLY EMPTY. The historical values were not recoverable and were not worth recovering (bugs_closed/378). Nothing writes or reads this column.';
COMMIT;
