-- Rollback for 352_pages_noindex_flag.sql
-- Dropping the column also discards the per-page flip -- there is no
-- separate way to "unflip" without removing the capability, since the flip
-- IS the column's only current use. bugs_open/232.
BEGIN;
ALTER TABLE pages DROP COLUMN noindex;
COMMIT;
