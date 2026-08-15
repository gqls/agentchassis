-- 423 ROLLBACK (sidecar, hand-run only): removes the publish_project
-- uniqueness guard. Only sensible if the seam's project semantics change
-- (e.g. a backend where sharing a project is legitimate) — dropping it
-- re-opens the silent-overwrite window the index exists to close.

BEGIN;

DROP INDEX IF EXISTS idx_sites_publish_project_unique;

COMMIT;
