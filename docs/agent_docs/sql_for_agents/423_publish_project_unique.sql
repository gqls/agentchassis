-- 423: one hosted copy per publish_project — uniqueness enforced at the DB
--
-- From the council round on 21aba3f5 (editquality, medium, advisory on an
-- APPROVED verdict): b2worker's object key is hostname + path, so two sites
-- carrying the same publish_project would silently overwrite each other's
-- prefix in portfolio-sites — last publisher wins, no error anywhere. The
-- Go guards validate the VALUE's shape (bare hostname, != domain); only the
-- DB can guard uniqueness ACROSS rows, and making the bad state
-- unrepresentable beats documenting it (order-fix-candidates ruling).
--
-- Safe to apply immediately: 0 rows carry publish_project (seam is opt-in
-- OFF fleet-wide, migration 412), so the index cannot fail on existing data.
--
-- Rollback: 423_publish_project_unique_ROLLBACK.sql.

BEGIN;

CREATE UNIQUE INDEX IF NOT EXISTS idx_sites_publish_project_unique
  ON sites (publish_project)
  WHERE publish_project IS NOT NULL;

COMMENT ON INDEX idx_sites_publish_project_unique IS
  'Publish seam (DGH-008): publish_project is a hosting-side namespace (b2worker: the serving hostname prefix in portfolio-sites). Two sites sharing one would overwrite each other''s hosted copy silently — this makes that state unrepresentable. Partial: NULL (seam off) stays repeatable.';

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM pg_indexes
   WHERE tablename = 'sites' AND indexname = 'idx_sites_publish_project_unique';
  IF n <> 1 THEN
    RAISE EXCEPTION '423: idx_sites_publish_project_unique missing after create';
  END IF;
END $$;

COMMIT;
