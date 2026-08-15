-- 412 ROLLBACK (sidecar, hand-run only — SIDECAR_RE excludes it from --apply).
-- Removes the publish-seam columns. Safe at any time while the seam is
-- default-OFF; once any row carries a non-NULL publish_target, dropping the
-- columns abandons that state — check first:
--   SELECT domain, publish_target FROM sites WHERE publish_target IS NOT NULL;

BEGIN;

ALTER TABLE sites
  DROP COLUMN IF EXISTS publish_target,
  DROP COLUMN IF EXISTS publish_project,
  DROP COLUMN IF EXISTS published_hash,
  DROP COLUMN IF EXISTS published_at;

COMMIT;
